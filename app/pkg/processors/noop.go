package processors

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
)

// 컴파일 타임 Type Casting 체크
var _ Processor = new(NoopProcessor)

//프로세서 등록
func init() {
	Register("noop", NewNoopProcessor)
}

// 프로세서 타입 정의
type NoopProcessor struct {
	//struct embedding
	Validator
}

// 프로세서 인스턴스 생성
func NewNoopProcessor(config jsonObj) Processor {
	validator := NewValidator()
	return &NoopProcessor{*validator}
}

// 프로세서 인스턴스의 Process 메소드 구현으로 Duck Typing
func (n *NoopProcessor) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	logger.Debugf("Processing %v", p)
	err := n.Validate(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
