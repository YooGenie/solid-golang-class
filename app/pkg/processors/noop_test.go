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
			// 프로세서 생성 =>
			np, err := processors.CreateProcessor(tC.processorName, nil)
			if err != nil {
				t.Error(err)
			}
			// 스테이지 러너 생성 1개
			fifo := pipelines.FIFO(np) // FIFO라는  스테이지 러너에 넣어 준다. FIFO 런을 할 때 FIFO 객체를 생성한다.

			// 파라미터 생성
			ctx, cancelFunc := context.WithCancel(context.Background())

			// Stage 개수는 1개로 고정
			// 채널 개수는 Stage 개수 +1
			stageCh := make([]chan payloads.Payload, 1+1) // 여러개의 채널을 생성한다. => 이유는? 프로세서가 1개가 아니라 여러개니까 여러개의 스테이지가 있을 수 있다.
			//1+1를 생성해주는데 stage에 개수는 하나잖아요. 채널은 2개이다. 만약에 채널이 하나면 컨슈머에서 프로세서에서 가는건 아무 문제가 없다. 그런데 나중에 스토리지 프로바이저로 데이터를 전송해줄  채널이 1개 더 필요하다.

			// 에러채널 개수는 Stage 개수 +2
			errCh := make(chan error, 1+2)
			for i := 0; i < len(stageCh); i++ {
				stageCh[i] = make(chan payloads.Payload) // 데이터를 주고 받을 때 채널을 사용한다. => 동시성이 발생했을 때 고에서는 고루틴을 사용해서 별도의 작업을 동시적으로 할 수 있는데
				// 그런 경우 데이터를 어떻게 주고 받을까 고에서 사용하는 기법중에 하나가 채널을 통해서 데이터를 주고 받는다. Payload의 타입 채널을 가지고 데이터를 주고 받는다.
				//
			}

			// FiFO Stage Runner 에게 넘길 파라미터 생성
			wp := &workerParams{ // workerParams라는 것을 통해서 이것도 인터페이스 구현체이다.
				stage: 0,
				inCh:  stageCh[0], // 파라미터를 구성할 때 그 다음 채널을 넣어 줘야한다.
				outCh: []chan<- payloads.Payload{stageCh[1]},
				errCh: errCh,
			}
			// workerParams 초기화를 할 때 몇번 째 스테이지인지 사용할 채널이 무엇인지 작성해서 파라미터를 생성한다.
			// 여기까지 프로세서는 생성하고 프로세서의 임의의 파라미터를 테스트 케이스로 부터 받아와서 전송하는 작업

			// StageRunner 구현체 FIFO 실행
			// Goroutine 으로 실행
			go fifo.Run(ctx, wp)
			// fifo는 Run이라는 메소드 시그니처를 가지고 있는 StageRunner 구현체 중에 하나이다.
			// 여기서 런을 한다.

			stageCh[0] <- tC.payload

			for { // 다중 채널로부터 데이터를 받을 때 사용하는 문법 => 순서와 상관없이 채널로부터 데이터가 들오는걸 실행한다. 데이터가 들어오면
				select {
				case err := <-errCh: // 에러가 발생했을 때 처리하는 부분
					if tC.testcase == "invalid" { // invalid 케이스이 그냥 로그를 찍는다.
						t.Log(err.Error())
					} else {
						t.Errorf(err.Error()) // invalid 아닌 경우 에러 찍기
					}
					cancelFunc()
					return
				case data := <-stageCh[1]: //
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
	inCh  <-chan payloads.Payload   // 컨슈머에서 채널은 in ( <-chan : 데이터를 쓰는 용도)
	outCh []chan<- payloads.Payload // 채널에서 프로세서는 out ( []chan<- : 데이터를 받는 용도)
	errCh chan<- error
}

func (p *workerParams) StageIndex() int                   { return p.stage }
func (p *workerParams) Input() <-chan payloads.Payload    { return p.inCh }
func (p *workerParams) Output() []chan<- payloads.Payload { return p.outCh }
func (p *workerParams) Error() chan<- error               { return p.errCh }
