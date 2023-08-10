package processors

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"time"
)

func init() {
	Register("rabbitmq_meta_injector", NewRabbitMQMetaInjector)
}

type RabbitMQMetaInjector struct {
}

func NewRabbitMQMetaInjector(config jsonObj) Processor {
	return &RabbitMQMetaInjector{}
}

// Process implements Processor
func (*RabbitMQMetaInjector) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	logger.Debugf("InjectMetaRabbitMQPayload processing...")
	rbbtPayload := p.(*payloads.RabbitMQPayload)

	meta := make(jsonObj)
	meta["data-processor-id"] = "rabbitmq-event-data-processor"
	meta["data-processor-timestamp"] = time.Now()
	meta["data-processor-env"] = "local"

	// 이 안에서 데이터를 처리한다.
	rbbtPayload.Value["meta"] = meta

	return rbbtPayload, nil

}
