package processors

import (
	"context"
	"errors"
	"event-data-pipeline/pkg/payloads"
)

type Validator struct {
}

func NewValidator() *Validator {

	return &Validator{}
}

func (Validator) Validate(ctx context.Context, p payloads.Payload) error {
	if p == nil {
		return errors.New("payload is nil")
	}
	// 데이터 유효성 체크
	// 유효성 통과하지 못하는 경우 에러 반환
	return nil
}
