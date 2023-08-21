package processors

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/payloads"
	"fmt"
	"strconv"
)

var _ Processor = new(ProcessorFunc) // Processor 인터페이스 타입을 만족시키는 것이다.
var _ ProcessorFunc = NormalizeKafkaPayload

func init() {
	Register("kafka_normalizer", NewKafkaNormalizer) // 저장소 있다면 저장소에 맞게 스키마를 일관성 있게 변경할 필요가 있다.
	// config에 kafka_normalizer 이게 있다.
}

func NewKafkaNormalizer(config jsonObj) Processor {

	return ProcessorFunc(NormalizeKafkaPayload) //스트럭이 아니고 함수이다. 그냥 함수인데 프로세서 타입으로 반환하고 있다. 가능한 이유는? 덕타이핑의 장점때문이다. ProcessorFunc 감싸서 ProcessorFunc타입으 변환했다.
	// 일반함수를 변환을 시켜서 리턴을 하며 Processor 인터페이스로 전환 된다.
	// 장점은? 스트럭으로 선언할 필요가 없어서 함수로 선언하면 된다. 일반함수로 개발을 하고 프로세서로 형변환을 해서 등록하면 런타임시 마다 사용할 수 있다.
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
