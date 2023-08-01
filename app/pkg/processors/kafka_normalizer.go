package processors

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/payloads"
	"fmt"
	"strconv"
)

var _ Processor = new(ProcessorFunc)
var _ ProcessorFunc = NormalizeKafkaPayload

func init() {
	Register("kafka_normalizer", NewKafkaNormalizer)
}

func NewKafkaNormalizer(config jsonObj) Processor {

	return ProcessorFunc(NormalizeKafkaPayload)
}

func NormalizeKafkaPayload(ctx context.Context, p payloads.Payload) (payloads.Payload, error) { // 카푸카 메시지가 전달 되었을  특정 양식으 인덱스와 docID, 그 안에 들어갈 데이를 정규화한다.
	// 정규화된 데이터를 파일 시스템이라고 하는 스토리 프로바이로 넘겨준다.
	kfkPayload := p.(*payloads.KafkaPayload)

	// 인덱스 생성
	index := fmt.Sprintf("%s-%s", "event-data", kfkPayload.Timestamp.Format("01-02-2006"))
	kfkPayload.Index = index

	// 식별자 생성
	docID := fmt.Sprintf("%s.%s.%v.%s", kfkPayload.Key, kfkPayload.Topic, strconv.FormatFloat(kfkPayload.Partition, 'f', 0, 64), strconv.FormatFloat(kfkPayload.Offset, 'f', 0, 64))
	kfkPayload.DocID = docID

	// 데이터 생성
	data, err := json.Marshal(kfkPayload.Value)
	if err != nil {
		return nil, err
	}
	kfkPayload.Data = data
	return kfkPayload, nil
}
