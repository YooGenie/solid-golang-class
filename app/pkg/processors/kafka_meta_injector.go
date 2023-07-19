package processors

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"time"
)

func init() {
	Register("kafka_meta_injector", NewKafkaMetaInjector)
}

type KafkaMetaInjector struct {
}

func NewKafkaMetaInjector(config jsonObj) Processor {
	return &KafkaMetaInjector{}
}

// Process implements Processor
func (*KafkaMetaInjector) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	logger.Debugf("InjectMetaKafkaPayload processing...")
	kfkPayload := p.(*payloads.KafkaPayload)

	meta := make(jsonObj)
	meta["data-processor-id"] = "kafka-event-data-processor"
	meta["data-processor-timestamp"] = time.Now()
	meta["data-processor-env"] = "local"

	kfkPayload.Value["meta"] = meta

	return kfkPayload, nil
}
