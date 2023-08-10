package processors

import (
	"context"
	"errors"
	"event-data-pipeline/pkg/payloads"
)

// 컴파일 타임 인터페이스 타입 체크
var _ Processor = new(RabbitMQDefaultProcessor)

func init() {
	Register("rabbitmq_default", NewRabbitMQDefaultProcessor)

}

type RabbitMQDefaultProcessor struct {
	Validator
	RabbitMQMetaInjector // 카프카와 똑같은 방식 => 메타 데이터를 넣어 줄 수 있다.
}

func NewRabbitMQDefaultProcessor(config jsonObj) Processor {
	p := &RabbitMQDefaultProcessor{
		Validator{},
		RabbitMQMetaInjector{},
	}
	return p
}

func (r *RabbitMQDefaultProcessor) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	err := r.Validate(ctx, p)
	if err != nil {
		return nil, err
	}
	p, err = r.RabbitMQMetaInjector.Process(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *RabbitMQDefaultProcessor) Validate(ctx context.Context, p payloads.Payload) error {
	err := r.Validator.Validate(ctx, p)
	if err != nil {
		return err
	}
	rp := p.(*payloads.RabbitMQPayload)
	if rp.Value == nil {
		return errors.New("value is nil")
	}
	return nil
}
