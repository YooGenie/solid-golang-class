package processors

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/payloads"
	"fmt"
)

var _ Processor = new(ProcessorFunc)
var _ ProcessorFunc = NormalizeRabbitMQPayload

func init() {
	Register("rabbitmq_normalizer", NewNormalizeRabbitMQPayloadProcessor)
}

func NewNormalizeRabbitMQPayloadProcessor(config jsonObj) Processor {

	return ProcessorFunc(NormalizeRabbitMQPayload)
}

func NormalizeRabbitMQPayload(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	rbbtMQPayload := p.(*payloads.RabbitMQPayload)

	// 인덱스 생성
	index := fmt.Sprintf("%s-%s", "event-data", rbbtMQPayload.Timestamp.Format("01-02-2006"))
	rbbtMQPayload.Index = index

	// 식별자 생성
	docID := fmt.Sprintf("%s.%v.%s.%s", rbbtMQPayload.Queue, rbbtMQPayload.Id, rbbtMQPayload.Email, rbbtMQPayload.Gender)
	rbbtMQPayload.DocID = docID

	// 데이터 생성
	data, err := json.Marshal(rbbtMQPayload.Value)
	if err != nil {
		return nil, err
	}
	rbbtMQPayload.Data = data
	return rbbtMQPayload, nil
}
