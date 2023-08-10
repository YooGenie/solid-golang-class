package processors_test

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"event-data-pipeline/pkg/pipelines"
	"event-data-pipeline/pkg/processors"
	"os"
	"testing"
)

type jsonObj = map[string]interface{}

func TestKafkaDefaultProcessor_Process(t *testing.T) {
	testCases := []struct { // 테스트 데이터를 설정한다.=> 값들을 초기화한다.
		desc          string
		processorName string
		payload       payloads.Payload
		testcase      string
	}{
		{
			desc:          "valid payload",
			processorName: "kafka_default",
			payload:       &payloads.KafkaPayload{Value: make(jsonObj)},
			testcase:      "valid",
		},
		{
			desc:          "invalid payload",
			processorName: "kafka_default",
			payload:       &payloads.KafkaPayload{Value: nil},
			testcase:      "invalid",
		},
	}
	os.Args = nil
	os.Setenv("EDP_ENABLE_DEBUG_LOGGING", "true")
	logger.Setup()
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// 프로세서 생성
			np, err := processors.CreateProcessor(tC.processorName, nil)
			if err != nil {
				t.Error(err)
			}
			// 스테이지 러너 생성
			fifo := pipelines.FIFO(np)

			// 파라미터 생성
			ctx, cancelFunc := context.WithCancel(context.Background()) // 컨텍스트와 캔슬함수 받아와서 선언하고

			// Stage 개수는 1개로 고정
			// 채널 개수는 Stage 개수 +1
			stageCh := make([]chan payloads.Payload, 1+1)
			// 에러채널 개수는 Stage 개수 +2
			errCh := make(chan error, 1+2) // 에러 채널도 파라미터에 넣어주기 위 값중에 하나이다.
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
			go fifo.Run(ctx, wp) // 피포 런을 해서 인풋으로부터 데이터를 받아오고 아웃풋으로 그 다음 스테이지에 넘겨주는 코드를 실행한다.

			stageCh[0] <- tC.payload // 스테이지 0번째 채널에 임의 페이로드를 테스트 코드를 통해서 받아서 넣어 준다.
			// => 피포런에서 인풋 채널로부터 데이터를 읽어와서 프로세서 로직을 태우난 뒤에 에러 채널이나 그 다음 스테이지 채널로 데이터를 아웃풋 해준다.

			noread := 0
			for {
				select {
				case err := <-errCh: // 에러 채널 데이터 아웃풋
					if tC.testcase == "invalid" {
						t.Log(err.Error())
					} else {
						t.Errorf(err.Error())
					}
					cancelFunc()
					return
				case msg := <-stageCh[1]: // 그 다음 스테이지 채널에 데이터 아웃풋 => 데이터를 읽어와서 실행한다.
					data, _ := json.MarshalIndent(msg, "", " ")
					t.Log(string(data))
					cancelFunc()
					return
				default:
					noread++
					if noread > 5 {
						cancelFunc()
						t.Log("no output for invalid payload")
					}
				}
			}
		})
	}
}
