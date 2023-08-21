package payloads

import (
	"sync"
	"time"
)

var (
	// 컴파일 타임 타입 변경 체크
	_ Payload = (*KafkaPayload)(nil)

	kafkaPayloadPool = sync.Pool{
		New: func() interface{} { return new(KafkaPayload) }, //사용했던 인스턴스를 다시 반환
	}
)

// 페이로드를 구현하는 상위 모 구현하는 구현체이다.
type KafkaPayload struct {
	Topic     string                 `json:"topic,omitempty"`
	Partition float64                `json:"partition,omitempty"`
	Offset    float64                `json:"offset,omitempty"`
	Key       string                 `json:"key,omitempty"`
	Value     map[string]interface{} `json:"value,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`

	Index string `json:"index,omitempty"`
	DocID string `json:"doc_id,omitempty"`
	Data  []byte `json:"data,omitempty"`
}

// Clone implements pipeline.Payload.
func (kp *KafkaPayload) Clone() Payload { // 디카피하는 내용이다.
	newP := kafkaPayloadPool.Get().(*KafkaPayload)

	newP.Topic = kp.Topic
	newP.Partition = kp.Partition
	newP.Offset = kp.Offset
	newP.Key = kp.Key
	newP.Value = kp.Value

	newP.Index = kp.Index
	newP.DocID = kp.DocID
	newP.Data = kp.Data

	return newP
}

// Out implements Payload
func (kp *KafkaPayload) Out() (string, string, []byte) {
	return kp.Index, kp.DocID, kp.Data // 원형 형태로 output 하는 형식이다.
}

// MarkAsProcessed implements pipeline.Payload
func (p *KafkaPayload) MarkAsProcessed() { // 더이상 처리할 필요가 없는 경우

	p.Topic = ""
	p.Partition = 0
	p.Offset = 0
	p.Key = ""
	p.Value = nil
	p.Timestamp = time.Time{}

	p.DocID = ""
	p.Index = ""
	p.Data = nil

	kafkaPayloadPool.Put(p) //PayloadPool이라고 하는 메모리를 효율적으로 관리하기 위한 기법
}
