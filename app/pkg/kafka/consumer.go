package kafka

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var _ Consumer = new(KafkaConsumer)

type Consumer interface {
	CreateConsumer() error
	CreateAdminConsumer() error
	GetPartitions() error
	Read(ctx context.Context) error // 이거 실제 구현체는 KafkaConsumer 스트럭 이라는 걸 알 수 있다.
	AssignPartition(partition int) error
	Poll(ctx context.Context)
	Stream() chan interface{}
	PutPaylod(p payloads.Payload) error
	GetPaylod() payloads.Payload
}

type KafkaConsumer struct {
	// topic to consume from
	topic string

	//configuration to create confluent-kafka-go consumer
	configMap *kafka.ConfigMap

	//confluent kafka go consumer
	kafkaConsumer *kafka.Consumer

	//adminClient to get partitions
	adminClient *AdminClient // 초기화 후에 adminClient의 메소드를 그대로 이용 가능
	//adminClient 만 변경 되면 partition를 가져오는 메소드는 KafkaConsumer 별도로 책임 질 필요가 없다. => 코드의 유지보수성이 높아진다.

	//partitions response
	partitions *PartitionsResponse

	ctx context.Context

	stream chan interface{}

	errCh chan error

	payload payloads.Payload
}

func NewKafkaConsumer(config jsonObj) *KafkaConsumer {
	topic, ok := config["topic"].(string)
	if !ok {
		logger.Panicf("no topic provided")
	}
	//context, stream, errch 추출
	ctx, stream, errch := extractPipeParams(config)

	//consumerOptions
	kfkCnsmrCfg, ok := config["consumerOptions"].(map[string]interface{})
	if !ok {
		logger.Panicf("no consumer options provided")
	}

	// load Consumer Options to kafka.ConfigMap
	cfgMapData, _ := json.Marshal(kfkCnsmrCfg)
	var kcm kafka.ConfigMap
	json.Unmarshal(cfgMapData, &kcm)

	// create a new KafkaConsumer with configMap fed in
	kafkaConsumer := &KafkaConsumer{
		topic:     topic,
		configMap: &kcm,
		ctx:       ctx,
		stream:    stream,
		errCh:     errch,
	}

	return kafkaConsumer

}

func extractPipeParams(config jsonObj) (context.Context, chan interface{}, chan error) {
	pipeParams, ok := config["pipeParams"].(map[string]interface{})
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

// create KafkaClient instance
func (kc *KafkaConsumer) CreateConsumer() error {

	if kc == nil {
		kc = &KafkaConsumer{configMap: kc.configMap}
	}

	var err error
	// create kafka consumer instance
	kc.kafkaConsumer, err = kafka.NewConsumer(kc.configMap)

	if err != nil {
		return err
	}
	logger.Debugf("Check in consumer creation: %s", kc.kafkaConsumer)
	return nil

}

func (kc *KafkaConsumer) CreateAdminConsumer() error {
	var err error
	kc.adminClient, err = NewAdminClient(kc.topic, kc.kafkaConsumer)
	if err != nil {
		return err
	}
	return nil
}

func (kc *KafkaConsumer) GetPartitions() error {
	var err error
	// get partitions from admin client
	kc.partitions, err = kc.adminClient.GetPartitions()
	if err != nil {
		return err
	}
	return nil
}

func (kc *KafkaConsumer) Read(ctx context.Context) error { // 카푸카와 커넥션을 맺은 다음 데이터를 읽어 오는 비즈니스 로직이 구현 되어 있다.
	// 파티션 별로 카프카 컨슈머 생성
	for _, p := range kc.partitions.Partitions {
		// KafkaConsumer 인스턴스 복사
		ckc := kc.Copy()
		// KafkaConsumer 내부 실제 컨슈머 생성
		ckc.CreateConsumer()
		// 데이터를 읽어오기 위한 파티션에 할당
		ckc.AssignPartition(int(p))
		// 실제 데이터를 읽어오는 고루틴 생성
		go ckc.Poll(ctx) // 비동기식으로 컨슈머별로 데이터를 읽어오는 고루틴을 실행함
	}
	return nil
}

// Copy KafkaConsumer instance
func (kc *KafkaConsumer) Copy() *KafkaConsumer {
	return &KafkaConsumer{
		topic:     kc.topic,
		configMap: kc.configMap,
		stream:    kc.stream,
		errCh:     kc.errCh,
	}
}

func (kc *KafkaConsumer) AssignPartition(partition int) error {

	var partitions []kafka.TopicPartition

	tp := NewTopicPartition(kc.topic, partition)
	partitions = append(partitions, *tp)

	err := kc.kafkaConsumer.Assign(partitions)
	if err != nil {
		return err
	}
	return err
}

func (kc *KafkaConsumer) Poll(ctx context.Context) { //실제로 데이터를 읽어오는 코드가 있다.
	cast := func(msg *kafka.Message) map[string]interface{} {
		var record = make(map[string]interface{})

		// dereference to put plain string
		record["topic"] = *msg.TopicPartition.Topic
		record["partition"] = float64(msg.TopicPartition.Partition)
		record["offset"] = float64(msg.TopicPartition.Offset)

		record["key"] = string(msg.Key)

		var valObj map[string]interface{}

		err := json.Unmarshal(msg.Value, &valObj)
		if err != nil {
			logger.Errorf("error in casting value object: %v", err)
		}
		record["value"] = valObj
		record["timestamp"] = msg.Timestamp
		return record
	}
	for {
		select {
		case <-ctx.Done():
			logger.Infof("shutting down consumer read")
			return
		default:
			ev := kc.kafkaConsumer.Poll(100)
			switch e := ev.(type) {
			case *kafka.Message:
				ConsumerReadTotal.Inc()
				record := cast(e)
				kc.stream <- record
				data, _ := json.MarshalIndent(record, "", " ")
				logger.Debugf("%s", string(data))
			case kafka.Error:
				logger.Errorf("Error: %v: %v", e.Code(), e)
				kc.errCh <- e
			case kafka.PartitionEOF:
				logger.Infof("[PartitionEOF][Consumer: %s][Topic: %v][Partition: %v][Offset: %d][Message: %v]", kc.kafkaConsumer.String(), *e.Topic, e.Partition, e.Offset, fmt.Sprintf("\"%s\"", e.Error.Error()))
			default:
				time.Sleep(1 * time.Second)
				logger.Infof("polling...")
			}
		}
	}
}

// GetPaylod implements Consumer
func (kc *KafkaConsumer) GetPaylod() payloads.Payload {
	return kc.payload
}

// PutPaylod implements Consumer
func (kc *KafkaConsumer) PutPaylod(p payloads.Payload) error {
	kc.payload = p
	return nil
}

// Stream implements Consumer
func (kc *KafkaConsumer) Stream() chan interface{} {
	return kc.stream
}
