package processors

import (
	"context"
	"errors"
	"event-data-pipeline/pkg/payloads"
)

var _ Processor = new(KafkaDefaultProcessor)

// 여기 핵심은 여러가지 기능을 할 수 있는 프로세서 혹은 함수들을 만들어 놓고 카프카가 되든 래빗엠큐 되든 데이터를 소비하는 소스에 따라서 카프카 안에서도 어떤 토픽, 래빗엠큐에 어떤 슈에 따라서 소비하는 데이터 스키마가 달라질 수 있다.
// 달라짐에 따라서 코드를 매번 구현을 해야하고 구현하는 방식도 빠르게 개발을 할 수 있어야한다. 상호 영향을 덜 받아야한다. (기존 코드에 영향을 덜 준다) 확장이 가능해야한다.
// 이런걸 따져 봤을 때 스트럭임베이라는 것으로 해소할 수 있다. => 예시가 KafkaDefault 이다.
type KafkaDefaultProcessor struct {
	Validator
	KafkaMetaInjector // 별도의 코드 구현 => 디커플링도 할 수 있 여기에계속
}

func init() {
	Register("kafka_default", NewKafkaDefaultProcessor) // kafka_default 프로세서를 사용하겠다. 런타임시 생성해서 사용하겠다.
}

func NewKafkaDefaultProcessor(config jsonObj) Processor {
	return &KafkaDefaultProcessor{
		Validator{},
		KafkaMetaInjector{},
	}
}

func (k *KafkaDefaultProcessor) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	//Validator method
	err := k.Validate(ctx, p)
	if err != nil {
		return nil, err
	}
	//KafkaMetaInject method forwarding
	p, err = k.KafkaMetaInjector.Process(ctx, p)
	if err != nil {
		return nil, err
	}
	// prometheus metrics counter
	KafkaProcessTotal.Inc()
	return p, nil
}
func (k *KafkaDefaultProcessor) Validate(ctx context.Context, p payloads.Payload) error {

	// Embedded Validator 사용
	err := k.Validator.Validate(ctx, p)
	if err != nil {
		return err
	}
	kp := p.(*payloads.KafkaPayload)
	if kp.Value == nil {
		return errors.New("value is nil")
	}
	return nil
}

//여기에 계속 Process가 추가 될 수 있다. => 그런 방식들이 조립하는 방식으로 할 수 있다. => 스트럭 임베딩을 통해서
