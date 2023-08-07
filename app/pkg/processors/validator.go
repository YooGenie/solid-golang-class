package processors

import (
	"context"
	"errors"
	"event-data-pipeline/pkg/payloads"
)

// Open Close Principle : 개방 폐쇄 법칙 예제
// - 데이터를 처리하는 프로세서가 있다고 했을 때 Processor 마다 데이터 유효성을 검증하는 용도의 메소드가 필요하다.
// Validator 타입의 struct 을 선언하고 Validator 안에서는 특정한 데이터를 유효성을 체크하는 로직이 있다.
// 이런 Validator 를 Processor 라는 다른 타입 임베딩함으로서 공통적으로 사용할 수 있다.
type Validator struct {
}

func NewValidator() *Validator {

	return &Validator{}
}

func (Validator) Validate(ctx context.Context, p payloads.Payload) error {
	if p == nil { // 페이로드를 받아서 닐이냐 아니 라는 로직을 가지고 있다.
		return errors.New("payload is nil")
	}
	// 데이터 유효성 체크
	// 유효성 통과하지 못하는 경우 에러 반환
	return nil
}

// Validate 로직을 쓰고 싶을 때 굳이 별도에 프로세스에 처음부터 Validate 로직을 짜는게 아니라 기존 코드를 재사용할 수 있는지 살펴보자.
