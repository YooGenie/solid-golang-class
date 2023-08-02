package pipelines

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/processors"

	"golang.org/x/xerrors"
)

type fifo struct {
	proc processors.Processor
}

// FIFO returns a StageRunner that processes incoming payloads in a first-in
// first-out fashion. Each input is passed to the specified processor and its
// output is emitted to the next stage.
func FIFO(proc processors.Processor) StageRunner {
	return fifo{proc: proc}
}

// Run implements StageRunner.
func (r fifo) Run(ctx context.Context, params StageParams) {
	defer func() {
		logger.Debugf("shutting down fifo run...")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case payloadIn, ok := <-params.Input():
			if !ok {
				return
			}
			if payloadIn == nil {
				return
			}
			clone := payloadIn.Clone()

			payloadOut, err := r.proc.Process(ctx, clone)
			if err != nil {
				wrappedErr := xerrors.Errorf("pipeline stage %d: %w", params.StageIndex(), err)
				maybeEmitError(wrappedErr, params.Error())
				return
			}

			// If the processor did not output a payload for the
			// next stage there is nothing we need to do.
			if payloadOut == nil {
				payloadIn.MarkAsProcessed() // 처리가 끝나서 필요없는 경우가 있다. 리턴된 페이로드가 없을 때 더이상 진행할 필요가 없으니까 그 다음 프로세서에 넘길 필요는 경우가 발생해서 이 코드를 선언해놨다.
			}

			// broadcast output to all output channels
			for _, outCh := range params.Output() {
				p := payloadOut.Clone()
				select {
				case outCh <- p:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
