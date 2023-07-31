package consumers

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/sources"

	"event-data-pipeline/pkg/kafka"
)

// compile type assertion check
var _ Consumer = new(KafkaConsumerClient)
var _ ConsumerFactory = NewKafkaConsumerClient

// ConsumerFactory 에 kafka 컨슈머를 등록
func init() { // init 함수를 통해서 메인 메소드를 구동하기 전에 사전에 메모리 상에 사용하고자 하는 컨슈머를 구현하고 있는 카푸카 KafkaConsumerClient를 등록하는 것이다.
	Register("kafka", NewKafkaConsumerClient) //컨슈머팩토리를 레지즈터를 통해서 등록하고 있다.
}

type KafkaClientConfig struct {
	ClientName      string  `json:"client_name,omitempty"`
	Topic           string  `json:"topic,omitempty"`
	ConsumerOptions jsonObj `json:"consumer_options,omitempty"`
}

// Consumer interface 구현체
type KafkaConsumerClient struct { // 이 안에는 kafka.Consumer, sources.Source라는 인터페이스를 임베딩하고 있다.
	kafka.Consumer
	sources.Source
}

func NewKafkaConsumerClient(config jsonObj) Consumer { //Consumer 타입을 반환하고 있는 컨슈머팩토이다. 컨슈머팩토리를 레지즈터를 통해서 등록하고 있다.

	// KafkaClientConfig로 값을 담기 위한 오브젝트
	consumerCfgObj, ok := config["consumerCfg"].(jsonObj)
	if !ok {
		logger.Panicf("no consumer configuration provided")
	}

	// Read config into KafkaClientConfig struct
	var kcCfg KafkaClientConfig
	cfgData, err := json.Marshal(consumerCfgObj)
	if err != nil {
		logger.Errorf(err.Error())
	}
	json.Unmarshal(cfgData, &kcCfg)

	kfkCnsmrCfg := make(jsonObj)
	kfkCnsmrCfg["topic"] = kcCfg.Topic
	kfkCnsmrCfg["consumerOptions"] = kcCfg.ConsumerOptions
	kfkCnsmrCfg["pipeParams"] = config["pipeParams"]
	kfkCnsmr := kafka.NewKafkaConsumer(kfkCnsmrCfg)

	// create a new Consumer concrete type - KafkaConsumerClient
	client := &KafkaConsumerClient{ // 클라이언트 카푸카 설정값
		Consumer: kfkCnsmr,
		Source:   sources.NewKafkaSource(kfkCnsmr),
	}

	return client

}

// Init implements Consumer
func (kc *KafkaConsumerClient) Init() error { // 카푸카의 경우
	var err error

	err = kc.CreateConsumer() // 컨슈머를 생성하고
	if err != nil {
		return err
	}
	err = kc.CreateAdminConsumer() //어드민 컨슈머를 생성하고
	if err != nil {
		return err
	}
	err = kc.GetPartitions() // 파티션 정보를 가져온다.
	if err != nil {
		return err
	}
	return nil
}

// Consumer 인터페이스 구현
func (kc *KafkaConsumerClient) Consume(ctx context.Context) error { //앞에서 정의 되어 있는 Consume 인터페이스의 메소드를 구현하고 있는 상세 구현 타입 메소드는
	//여기에 선언 되어있다.
	// KafkaConsumerClient 라는 상세 타입 있다.
	err := kc.Read(ctx) // read라는 메소드를 카푸카 패키지에서 가져와서 사용하는 걸 알 수 있다.
	if err != nil {
		return err
	}
	return nil
}
