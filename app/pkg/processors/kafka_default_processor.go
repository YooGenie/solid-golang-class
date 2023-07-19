package processors

import (
	"context"
	"errors"
	"event-data-pipeline/pkg/payloads"
)

var _ Processor = new(KafkaDefaultProcessor)

type KafkaDefaultProcessor struct {
	Validator
	KafkaMetaInjector
}

func init() {
	Register("kafka_default", NewKafkaDefaultProcessor)
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
