package processors

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
)

// 컴파일 타임 Type Casting 체크
var _ Processor = new(NoopProcessor)

// 프로세서 등록
func init() {
	Register("noop", NewNoopProcessor)
}

// 프로세서 타입 정의
type NoopProcessor struct { // NoopProcessor 타입 안에서 데이터를 처리하는 단계에서 임베딩하는 벨리데이터에 벨리데이트라고 하는 메소드를 사용항 수 있다.
	// 이런 경우 기존의 코드를 사용할 수 있지만 벨리데이터에 벨리데이트라고 하는 메소드 안은 조작할 수 없다. 조작할 수 없는 것에서 폐쇄라고 한다.
	// 동시에 개방되어있다.  NoopProcessor는 벨리데이터라는 객체 타입을 임베딩하고 있기 때문에 확장해서 사용할 수 있다.
	// 기존의 벨리데이터는 조작할 수 없다.
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
	// 확장된 Validate 메소드 사용
	err := n.Validate(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (n *NoopProcessor) Validate(ctx context.Context, p payloads.Payload) error { // NoopProcessor에서 Validate를 따로 정의 함
	// 기존의 Validtor라는 메소드를 그냥 사용하여 기존 삽입하고 있는 Validate 유효성 검증 로직을 그대로 사용한다.
	// 별도에 NoopProcessor에 Validate 로직을 구현 함으로 기존 코드 확장 했다.
	n.Validator.Validate(ctx, p)
	return nil
}
