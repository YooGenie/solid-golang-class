package processors_test

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"event-data-pipeline/pkg/pipelines"
	"event-data-pipeline/pkg/processors"
	"os"
	"testing"
)

func TestNoopProcessor_Process(t *testing.T) {
	// 테스트 케이스를 선언해 진행한다. 보통 스트럭트로 만들어서 하지만 언나니머스 스트럭을 사용한다. => 어나니머스를 선언하고 동시에 초기화를 시킨다. 슬라이스에 형테이다. 순회를 하면서 구조체에 있는 값을 읽기때문에
	testCases := []struct {
		desc          string
		processorName string
		payload       payloads.Payload
		testcase      string
	}{ // 여기 초기화값
		{
			desc:          "valid payload",
			processorName: "noop",
			payload:       &ConcretePayload{id: 1},
			testcase:      "valid",
		},
		{
			desc:          "invalid payload",
			processorName: "noop",
			payload:       nil,
			testcase:      "invalid",
		},
	}
	os.Args = nil                                 // vs에서 테스트할 때 불필요한 인자값이 넘어 때 인지할 수 없다.
	os.Setenv("EDP_ENABLE_DEBUG_LOGGING", "true") // 디버깅해보기 위해서 환경변수 설정
	logger.Setup()
	// 실제로 테스트 코드를 실행하는 부분
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// 프로세서 생성
			np, err := processors.CreateProcessor(tC.processorName, nil)
			if err != nil {
				t.Error(err)
			}
			// 스테이지 러너 생성 1개
			fifo := pipelines.FIFO(np)

			// 파라미터 생성
			ctx, cancelFunc := context.WithCancel(context.Background())

			// Stage 개수는 1개로 고정
			// 채널 개수는 Stage 개수 +1
			stageCh := make([]chan payloads.Payload, 1+1)
			// 에러채널 개수는 Stage 개수 +2
			errCh := make(chan error, 1+2)
			for i := 0; i < len(stageCh); i++ {
				stageCh[i] = make(chan payloads.Payload)
			}

			// FiFO Stage Runner 에게 넘길 파라미터 생성
			wp := &workerParams{
				stage: 0,
				inCh:  stageCh[0],
				outCh: []chan<- payloads.Payload{stageCh[1]},
				errCh: errCh,
			}
			// StageRunner 구현체 FIFO 실행
			// Goroutine 으로 실행
			go fifo.Run(ctx, wp)

			stageCh[0] <- tC.payload

			for {
				select {
				case err := <-errCh:
					if tC.testcase == "invalid" {
						t.Log(err.Error())
					} else {
						t.Errorf(err.Error())
					}
					cancelFunc()
					return
				case data := <-stageCh[1]:
					t.Log(data)
					cancelFunc()
					return
				}
			}
		})
	}
}

var _ payloads.Payload = new(ConcretePayload)

type ConcretePayload struct {
	id int
}

// Clone implements payloads.Payload
func (c *ConcretePayload) Clone() payloads.Payload {
	newCp := &ConcretePayload{}
	newCp.id = c.id
	return newCp
}

// MarkAsProcessed implements payloads.Payload
func (*ConcretePayload) MarkAsProcessed() {
	// Do nothing
}

// Out implements payloads.Payload
func (*ConcretePayload) Out() (string, string, []byte) {
	return "", "", nil
}

// processors_test 패키지에서 테스트용으로만 사용하는 struct
// StageParams 인터페이스의 구현체
type workerParams struct {
	stage int

	// Channels for the worker's input, output and errors.
	inCh  <-chan payloads.Payload
	outCh []chan<- payloads.Payload
	errCh chan<- error
}

func (p *workerParams) StageIndex() int                   { return p.stage }
func (p *workerParams) Input() <-chan payloads.Payload    { return p.inCh }
func (p *workerParams) Output() []chan<- payloads.Payload { return p.outCh }
func (p *workerParams) Error() chan<- error               { return p.errCh }
