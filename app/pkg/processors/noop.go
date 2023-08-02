package processors

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
)

// 컴파일 타임 Type Casting 체크
var _ Processor = new(NoopProcessor) // var 라는 예약어를 사용해서 변수 선언을 하고 사용하지 않기 때문에 _ 앞에 써준다.
// 내가 체크하고 싶은 캐스팅 타입으로 선언을 한다.
// new 라는 키워드 통해서 빈어 있는 NoopProcessor 객체를 생성한다.
// func (n *NoopProcessor) Process(ctx context.Context, p payloads.Payload, data int) (payloads.Payload, error) {}
// 입력 값을 변경하면 컴파일 타입에서 에러가 난다. Processor 인터페이스에서는 data int 이게 없는데 이걸 리턴하는 타입은 	return &NoopProcessor{*validator}이 만족되지 않는다는 의미이다.
// Process(context.Context, payloads.Payload) (payloads.Payload, error)
// 이런 컴파일 타입이 체크가 가능하다.

// 프로세서 등록
func init() {
	Register("noop", NewNoopProcessor)
	// 컨슈머에서 레지스트 과정이 있 프로세서 패키지 동일 방법에서 프로세서를 등록했다.
	// 런타임때 필요한 프로세서만 가져와서 쓸수 있도록 구성해놨다.

}

// 프로세서 타입 정의
// NoopProcessor : 어떤 작업을 딱히 하지 않는 프로세서
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

// NewNoopProcessor 함수는 Processor 인터페이스 타입으로 반환한다.
// 실제 인스턴트를 초기화할 때는 NoopProcessor 스트럭 타입으로 초기화를 해서 리턴을 한다.
// 상위 모듈 Processor 타입을 만족시키는 구현체 있기 때문에 리턴 시킬수 있다.
// 이 구현한 Processor 인터페이스 만족시킨다는 것을 var _ Processor = new(NoopProcessor)
// 컴파일 타임에 체크를 하고 싶다.

// 프로세서 인스턴스의 Process 메소드 구현으로 Duck Typing
func (n *NoopProcessor) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	logger.Debugf("Processing %v", p) // 디버그로 이 데이터를 Processing하고 있다.
	// 확장된 Validate 메소드 사용
	err := n.Validate(ctx, p) // 로그 찍고 Validate에 Validate 로직을 탄다.
	if err != nil {
		return nil, err
	}
	return p, nil
}

// 별도의 로직 없 껍데기만 있는 걸 실행할 예정
func (n *NoopProcessor) Validate(ctx context.Context, p payloads.Payload) error { // NoopProcessor에서 Validate를 따로 정의 함
	// 기존의 Validtor라는 메소드를 그냥 사용하여 기존 삽입하고 있는 Validate 유효성 검증 로직을 그대로 사용한다.
	// 별도에 NoopProcessor에 Validate 로직을 구현 함으로 기존 코드 확장 했다.
	n.Validator.Validate(ctx, p)
	return nil
}
