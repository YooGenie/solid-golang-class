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

type KafkaMetaInjector struct { // 적어도 카푸카 데이터가 들어왔을 때 일종의 메타데이터를 넣어줘서 나중에 데이터를 분석하는 용도로 사용하고 싶다.=>
}

func NewKafkaMetaInjector(config jsonObj) Processor {
	return &KafkaMetaInjector{}
}

// Process implements Processor
func (*KafkaMetaInjector) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	logger.Debugf("InjectMetaKafkaPayload processing...")
	kfkPayload := p.(*payloads.KafkaPayload)

	meta := make(jsonObj)                                    // 메타데이터 기존에 없는 데이터를 넣고 싶을 때
	meta["data-processor-id"] = "kafka-event-data-processor" // 어느 데이터 프로세서가 프로세싱을 했고
	meta["data-processor-timestamp"] = time.Now()            // 언제 프로세싱을 했고
	meta["data-processor-env"] = "local"                     // 그 환경이 어떤 환경인지

	// 그 외에도 비즈니스 로직에 따라 기능이 늘어날 필요가 있다. 기능을 계속 추가 해야할 때 하나의 프로세싱에 모든 로직을 다 구현할 수 있지만 그러면 확장성도 떨어지고 재사용성도 떨어지고 기존 코드의 영향을 받을 수 있다.
	// 만약에 별도의 코드로 구현한 이후 임베딩을 하 디커플링도 할 수 있고 기존 코드에 영향도 최소화할 수 있고 코드 변경을 했을 때 조립하듯 코드를 사용할 수 있다. 구현을 해 놓고 사용할 프로세서들을 설정에서 주입해서 런타임시 사용을 할 수 있다.

	kfkPayload.Value["meta"] = meta

	return kfkPayload, nil
}
