package consumers_test

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/cli"
	"event-data-pipeline/pkg/config"
	"event-data-pipeline/pkg/consumers"
	"event-data-pipeline/pkg/logger"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/alexflint/go-arg"
)

type jsonObj = map[string]interface{}

func TestKafkaConsumerClient_Consume(t *testing.T) {
	configPath := getCurDir() + "/test/consumers/kafka_consumer_config.json" // 테스트로 사용한 json 파일 => 설정 값을 json 형태로 구현
	os.Setenv("EDP_ENABLE_DEBUG_LOGGING", "true")                            // 필요한 환경변수를 코드로 설정
	os.Setenv("EDP_CONFIG", configPath)
	os.Args = nil                                              //vs 코드에 디버깅 툴을 사용할 때 불필요한 인자값이 넘겨서 오작동할 수 있어서 초기화를 해줬다.
	arg.MustParse(&cli.Args)                                   // 설정한 환경변수로 로딩을 하기 위해서 파싱을 한다.
	logger.Setup()                                             // 로거 셋팅
	cfg := config.NewConfig()                                  // config 패키지에서 newconfig를 가져오는 이유는 맨 처음 프로그램을 구하기 위해서 제방 설정을 가져온다.
	pipeCfgs := config.NewPipelineConfig(cfg.PipelineCfgsPath) // 파이프라인을 구동하기 위한 파이브라인의 컨피그패를 넘긴다.

	ctx := context.TODO()            // 사용하고자 하는 context
	stream := make(chan interface{}) // 데이터를 읽어오는 채널
	errCh := make(chan error)        // 에러 채널

	for _, cfg := range pipeCfgs { // pipeCfgs 변수를 for문을 돌면서 필요한 context와 에러채널, 스트림 채널들을 consumerCfg와 pipeParams 객체로 넣어 준다.
		cfgParams := make(jsonObj)
		pipeParams := make(jsonObj)

		// 컨텍스트
		pipeParams["context"] = ctx
		pipeParams["stream"] = stream

		// 컨슈머 에러 채널
		errCh := make(chan error)
		pipeParams["errch"] = errCh

		cfgParams["pipeParams"] = pipeParams
		cfgParams["consumerCfg"] = cfg.Consumer.Config

		kafkaConsumer, err := consumers.CreateConsumer(cfg.Consumer.Name, cfgParams) // cfgParams 객체 타입으로 CreateConsumer에 파라미터로 넘겨준다.
		// cfg.Consumer.Name => config.json 들어가면 카푸카인지 레디큐인지 따라서 생성된 컨슈머 상세 타입이 달라지게 된다.
		if err != nil {
			t.Error(err)
		}
		err = kafkaConsumer.Init() // 생성된 컨슈머를 init 과정을 통해서 초기작업을 함
		if err != nil {
			t.Error(err)
		}
		kafkaConsumer.Consume(context.TODO())
	}

	for {
		select {
		case data := <-stream:
			_json, _ := json.MarshalIndent(data, "", " ")
			t.Logf("data: %v", string(_json))
			return
		case err := <-errCh:
			t.Logf("err: %v", err)
		default:
			time.Sleep(1 * time.Second)
		}
	}

}

func getCurDir() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	return dir
}
