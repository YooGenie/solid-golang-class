package kafka

import (
	"encoding/json"
	"errors"
	"event-data-pipeline/pkg/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Admin interface {
	GetPartitions() (*PartitionsResponse, error) // 카프카에서 Partition 정보를 가져오 메소드를 정의하고 있는 Admin 인터페이스
}

// Admin Class that implements Admin
// Admin 인터페이스를 구현하는 Struct
type AdminClient struct {
	topic      string
	consumer   *kafka.Consumer
	partitions []kafka.PartitionMetadata
}

// Admin interface와 AdminClient struct 은 오로지 카프카의 파티션 정보를 가져오는 것만 관심을 가진다.
// 이것에 대한 이점은 무멋이 있을까요?
// AdminClient 를 Embedding 하는 struct 인 KafkaConsumer 타입은 초기화 작업 이후에 adminClient의 메소드를 그대로 이용할 수 있다.

func NewAdminClient(topic string, consumer *kafka.Consumer) (*AdminClient, error) {
	if topic == "" {
		return nil, errors.New("topic is empty")
	}
	if consumer == nil {
		return nil, errors.New("consumer is nil")
	}
	return &AdminClient{
		topic:    topic,
		consumer: consumer,
	}, nil
}

func (ac *AdminClient) GetPartitions() (*PartitionsResponse, error) {
	// create admin client from a consumer
	adminClient, err := kafka.NewAdminClientFromConsumer(ac.consumer)
	if err != nil {
		return nil, err
	}
	// close it on return
	defer adminClient.Close()
	md, err := adminClient.GetMetadata(&ac.topic, false, 5000)
	if err != nil {
		return nil, err
	}
	raw, _ := json.Marshal(md)
	logger.Debugf("metadata: %v", string(raw))

	//set partitions data to the admin client
	ac.partitions = md.Topics[ac.topic].Partitions

	partitionsResponse := PartitionsResponse{
		Topic: ac.topic,
	}
	for _, partition := range ac.partitions {
		partitionsResponse.Partitions = append(partitionsResponse.Partitions, float64(partition.ID))
	}
	return &partitionsResponse, err
}
