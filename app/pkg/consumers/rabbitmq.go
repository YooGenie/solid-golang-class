package consumers

import (
	"context"
	"event-data-pipeline/pkg/rabbitmq"
	"event-data-pipeline/pkg/sources"
)

// compile type assertion check
var _ Consumer = new(RabbitMQConsumerClient)
var _ ConsumerFactory = NewRabbitMQConsumerClient

// ConsumerFactory 에 rabbitmq 컨슈머를 등록
func init() {
	Register("rabbitmq", NewRabbitMQConsumerClient)
}

type RabbitMQConsumerClient struct {
	rabbitmq.Consumer
	sources.Source
}

func NewRabbitMQConsumerClient(config jsonObj) Consumer {

	consumer := rabbitmq.NewRabbitMQConsumer(config)
	source := sources.NewRabbitMQSource(consumer)
	client := &RabbitMQConsumerClient{
		Consumer: consumer,
		Source:   source,
	}
	return client
}

// Init implements Consumer
func (rc *RabbitMQConsumerClient) Init() error {
	err := rc.CreateConsumer()
	if err != nil {
		return err
	}
	err = rc.QueueDeclare()
	if err != nil {
		return err
	}

	err = rc.InitDeliveryChannel()
	if err != nil {
		return err
	}

	return nil
}

// Consume implements Consumer
func (rc *RabbitMQConsumerClient) Consume(ctx context.Context) error {
	go rc.Read(ctx)
	return nil
}
