package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"event-data-pipeline/pkg/rabbitmq/casters"
	"time"

	"github.com/streadway/amqp"
)

var _ Consumer = new(RabbitMQConsumer)

type Consumer interface {
	CreateConsumer() error
	QueueDeclare() error
	Read(ctx context.Context) error
	Connect() error
	CreateChannel() error
	Delete() error
	ReConnect() error
	InitDeliveryChannel() error

	//Source 구현체에서 필요한 인터페이스
	Stream() chan interface{}
	PutPaylod(p payloads.Payload) error
	GetPaylod() payloads.Payload
}

type RabbitMQConsumer struct {
	config         *RabbitMQConsumerConfig
	conn           *amqp.Connection
	ConnMaxRetries int
	ch             *amqp.Channel
	q              amqp.Queue
	message        <-chan amqp.Delivery
	ctx            context.Context
	stream         chan interface{}
	errCh          chan error
	caster         casters.Caster
	payload        payloads.Payload
}

// Read implements Consumer
func (rc *RabbitMQConsumer) Read(ctx context.Context) error {

	cast := func(msg amqp.Delivery) (jsonObj, error) {
		meta := make(jsonObj)
		meta["queue"] = rc.config.QueueName
		record, err := rc.caster.Cast(meta, msg)
		if err != nil {
			return nil, err
		}
		return record, nil
	}

	// Check RabbitMQ Connection Error
	if rc.conn == nil {
		err := rc.ReConnect()
		if err != nil {
			logger.Errorf("error in reconnecting: %v", err)
		}
	}

	for {
		select {
		case <-rc.ctx.Done():
			logger.Debugf("Context cancelled, shutting down...")
			return rc.ctx.Err()
		case msg := <-rc.message:
			record, err := cast(msg)
			if err != nil {
				logger.Errorf(err.Error())
			}
			data, _ := json.Marshal(record)
			logger.Debugf("rabbitmq message :%s", string(data))
			rc.stream <- record
		default:
			logger.Debugf("no message coming in...")
			time.Sleep(1 * time.Second)
		}
	}
}

func NewRabbitMQConsumer(config jsonObj) *RabbitMQConsumer {
	//context, stream, errch 추출
	ctx, stream, errch := extractPipeParams(config)

	//consumerCfg
	rbbtmqCnsmrCfg, ok := config["consumerCfg"].(jsonObj)
	if !ok {
		logger.Panicf("no consumer options provided")
	}
	var cfg RabbitMQConsumerConfig
	cfgData, err := json.Marshal(rbbtmqCnsmrCfg)
	if err != nil {
		logger.Panicf("error in mashalling rabbitmq configuration: %v", err)
		return nil
	}

	err = json.Unmarshal(cfgData, &cfg)
	if err != nil {
		logger.Panicf("error in loading rabbitmq configuration: %v", err)
		return nil
	}
	casterName := cfg.QueueName + casters.CASTER_SUFFIX
	castFunc, err := casters.CreateCaster(casterName)
	if err != nil {
		logger.Panicf("error in loading caster function: %v", err)
		return nil
	}

	c := &RabbitMQConsumer{
		config: &cfg,
		ctx:    ctx,
		stream: stream,
		errCh:  errch,
		caster: castFunc,
	}
	return c
}

func (c *RabbitMQConsumer) CreateConsumer() error {

	// best practice is to reuse connections and channels
	err := c.Connect()
	if err != nil {
		return err
	}
	err = c.CreateChannel()
	if err != nil {
		return err
	}
	logger.Debugf("Check in consumer creation: %v", c.ch)

	return nil
}

func (c *RabbitMQConsumer) Connect() error {
	c.ConnMaxRetries = -1
	retry := 0
	for {
		conn, err := amqp.Dial(c.config.Host)
		if conn != nil {
			logger.Infof("rabbitmq connection made")
			c.conn = conn
			break
		}
		retry++
		if c.ConnMaxRetries >= 0 && retry > c.ConnMaxRetries {
			return err
		}
		time.Sleep(1 * time.Second)
		logger.Infof("retry to connect to rabbitmq")
	}
	//Listen to NotifyClose
	go func() {
		<-c.conn.NotifyClose(make(chan *amqp.Error))
		c.errCh <- errors.New("rabbitmq connection closed")
	}()
	return nil
}

func (c *RabbitMQConsumer) CreateChannel() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	c.ch = ch
	return nil
}

func (c *RabbitMQConsumer) Delete() error {
	logger.Debugf("deleting rabbit mq consumer connection: %s and channel: %s", c.conn, c.ch)
	if c.ch != nil {
		err := c.ch.Close()
		if err != nil {
			return err
		}
	}
	if c.conn != nil && !c.conn.IsClosed() {
		err := c.conn.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
func (c *RabbitMQConsumer) ReConnect() error {
	logger.Debugf("Reconnecting")
	err := c.Connect()
	if err != nil {
		return err
	}
	err = c.CreateChannel()
	if err != nil {
		return err
	}

	err = c.InitDeliveryChannel()
	if err != nil {
		return err
	}
	return nil

}

func (c *RabbitMQConsumer) QueueDeclare() error {
	q, err := c.ch.QueueDeclare(
		c.config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	c.q = q
	if err != nil {
		return err
	}
	return nil
}

func (c *RabbitMQConsumer) InitDeliveryChannel() error {
	msg, err := c.ch.Consume(
		c.config.QueueName, // queue
		"edp-consumer",     // consumer
		true,               // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return err
	}
	c.message = msg
	return nil
}

// GetPaylod implements Consumer
func (rc *RabbitMQConsumer) GetPaylod() payloads.Payload {
	if rc.payload == nil {
		return nil
	}
	return rc.payload
}

// PutPaylod implements Consumer
func (rc *RabbitMQConsumer) PutPaylod(p payloads.Payload) error {
	if p == nil {
		return errors.New("paylod is nil")
	}
	rc.payload = p
	return nil
}

// Stream implements Consumer
func (rc *RabbitMQConsumer) Stream() chan interface{} {
	return rc.stream
}

func extractPipeParams(config jsonObj) (context.Context, chan interface{}, chan error) {
	pipeParams, ok := config["pipeParams"].(jsonObj)
	if !ok {
		logger.Panicf("no pipeParams provided")
	}

	ctx, ok := pipeParams["context"].(context.Context)
	if !ok {
		logger.Panicf("no topic provided")
	}

	stream, ok := pipeParams["stream"].(chan interface{})
	if !ok {
		logger.Panicf("no stream provided")
	}

	errch, ok := pipeParams["errch"].(chan error)
	if !ok {
		logger.Panicf("no errch provided")
	}
	return ctx, stream, errch
}
