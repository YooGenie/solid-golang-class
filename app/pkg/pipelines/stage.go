package pipelines

import (
	"context"
	"event-data-pipeline/pkg/payloads"
)

type StageRunner interface {
	Run(context.Context, StageParams) //프로세서를 실제 실행할 때 사용되는 감싸고 있는 모듈이다.
}

// 테스트용에서 사요
type StageParams interface { // pipelines 패키지 아래 있는 인터페이스/ workerParams 구현체의 상위 모듈이다. 데이터를 주고 받는 걸로 선언했다.
	// StageIndex returns the position of this stage in the pipeline.
	StageIndex() int

	// Input returns a channel for reading the input payloads for a stage.
	Input() <-chan payloads.Payload

	// Output returns a channel for writing the output payloads for a stage.
	Output() []chan<- payloads.Payload

	// Error returns a channel for writing errors that were encountered by
	// a stage while processing payloads.
	Error() chan<- error
}
