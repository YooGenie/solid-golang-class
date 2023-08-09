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
func FIFO(proc processors.Processor) StageRunner { // fifo 들어온 순서대로 대기열을 처리하는 StageRunner 구현체
	return fifo{proc: proc}
}

// Run implements StageRunner.
func (r fifo) Run(ctx context.Context, params StageParams) {
	defer func() {
		logger.Debugf("shutting down fifo run...")
	}()
	for {
		select {
		case <-ctx.Done(): // 컨텍스에 리스닝을 하고 있다. 컨텍스트가 언제든 취소가 되면 리턴을 해서 프로그램을 종료를 하도록
			return
		case payloadIn, ok := <-params.Input(): // StageParams 인터페이스를 구현한 구현체를 넘겨준다. => go fifo.Run(ctx, wp) / params에서 input를 한다. 인풋을 하게 되면 채널로 부터 읽어 들이기 위한 채널 데이터를 읽어와서
			//Input : 프로세서가 채널을 통해서 데이터를 읽어온다.
			if !ok { // 읽어왔는데 오케이가 아니면 리턴하고
				return
			}
			if payloadIn == nil { // payloadIn가 닐이면 리턴을 한다

				return
			}
			// 이 로직을 통과하면 payloadIn에서 복사를 한다. => 디 카피를 해서 하나의 복사본을 만든다.
			clone := payloadIn.Clone()

			payloadOut, err := r.proc.Process(ctx, clone) // 실제 프로세서를 실행하는 로직은 여기다. 주입한 프로세서가 된다. 주입하는 과정은 fifo StageRunner 객체를 생성할 때 생성된 프로세서 객체를 넣어줬다. 그 안에 들어간 프로세서를 실행하게 되는것이다.
			// payloadOut 실행 결과에 따라서
			if err != nil { // 에러가 있으면 출력 처리를 한다.
				wrappedErr := xerrors.Errorf("pipeline stage %d: %w", params.StageIndex(), err)
				maybeEmitError(wrappedErr, params.Error())
				return
			}

			// 에러가 없으면
			// If the processor did not output a payload for the
			// next stage there is nothing we need to do.
			if payloadOut == nil { // payloadOut 결과값이 없으면 그 다음으로 넘어간다.
				payloadIn.MarkAsProcessed() // 처리가 끝나서 필요없는 경우가 있다. 리턴된 페이로드가 없을 때 더이상 진행할 필요가 없으니까 그 다음 프로세서에 넘길 필요는 경우가 발생해서 이 코드를 선언해놨다.
				continue
			}

			// 결과값이 있으면 payloadOut를 Output으로 넘긴다. =>
			// broadcast output to all output channels
			for _, outCh := range params.Output() { // 컨슈머는 아웃을 해서 채널에 넣고 프로세스는 채널안에 데이터를 in한다. 프로세스는 out 해서 채널에 널고 스토리지 프로바이저는 데이터를 in한다.
				// 프로세서가 output으로 데이터를 넘겨준다.
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
