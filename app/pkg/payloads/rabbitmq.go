package payloads

import (
	"sync"
	"time"
)

var (
	// 컴파일 타임 타입 변경 체크
	_ Payload = (*RabbitMQPayload)(nil)

	rabbitMQPayloadPool = sync.Pool{
		New: func() interface{} { return new(RabbitMQPayload) },
	}
)

type RabbitMQPayload struct {
	Id        int    `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	Gender    string `json:"gender,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`

	Queue     string                 `json:"queue,omitempty"`
	Value     map[string]interface{} `json:"value,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`

	Index string `json:"index,omitempty"`
	DocID string `json:"doc_id,omitempty"`
	Data  []byte `json:"data,omitempty"`
}

// Clone implements pipeline.Payload.
func (kp *RabbitMQPayload) Clone() Payload {
	newP := rabbitMQPayloadPool.Get().(*RabbitMQPayload)
	newP.Id = kp.Id
	newP.Email = kp.Email
	newP.Gender = kp.Gender
	newP.FirstName = kp.FirstName
	newP.LastName = kp.LastName
	newP.Queue = kp.Queue
	newP.Value = kp.Value
	newP.Timestamp = kp.Timestamp
	newP.Index = kp.Index
	newP.DocID = kp.DocID
	newP.Data = kp.Data

	return newP
}

// Out implements Payload
func (kp *RabbitMQPayload) Out() (string, string, []byte) {
	return kp.Index, kp.DocID, kp.Data
}

// MarkAsProcessed implements pipeline.Payload
func (p *RabbitMQPayload) MarkAsProcessed() {
	p.Id = 0
	p.Email = ""
	p.Gender = ""
	p.LastName = ""
	p.FirstName = ""
	p.Queue = ""
	p.Value = nil
	p.Timestamp = time.Time{}

	p.DocID = ""
	p.Index = ""
	p.Data = nil

	rabbitMQPayloadPool.Put(p)
}
