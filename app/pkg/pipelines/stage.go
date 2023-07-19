package pipelines

import (
	"context"
	"event-data-pipeline/pkg/payloads"
)

type StageRunner interface {
	Run(context.Context, StageParams)
}

type StageParams interface {
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
