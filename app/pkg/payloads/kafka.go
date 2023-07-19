package payloads

import (
	"sync"
	"time"
)

var (
	// 컴파일 타임 타입 변경 체크
	_ Payload = (*KafkaPayload)(nil)

	kafkaPayloadPool = sync.Pool{
		New: func() interface{} { return new(KafkaPayload) },
	}
)

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
func (kp *KafkaPayload) Clone() Payload {
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
	return kp.Index, kp.DocID, kp.Data
}

// MarkAsProcessed implements pipeline.Payload
func (p *KafkaPayload) MarkAsProcessed() {

	p.Topic = ""
	p.Partition = 0
	p.Offset = 0
	p.Key = ""
	p.Value = nil
	p.Timestamp = time.Time{}

	p.DocID = ""
	p.Index = ""
	p.Data = nil

	kafkaPayloadPool.Put(p)
}
