package consumers_test

import (
	"context"
	"encoding/json"
	"event-data-pipeline/pkg/cli"
	"event-data-pipeline/pkg/config"
	"event-data-pipeline/pkg/consumers"
	"event-data-pipeline/pkg/logger"
	"os"
	"testing"
	"time"

	"github.com/alexflint/go-arg"
)

func TestRabbitMQConsumerClient_Consume(t *testing.T) {
	configPath := getCurDir() + "/test/consumers/rabbitmq_consumer_config.json"
	os.Setenv("EDP_ENABLE_DEBUG_LOGGING", "false")
	os.Setenv("EDP_CONFIG", configPath)
	os.Args = nil
	arg.MustParse(&cli.Args)
	logger.Setup()
	cfg := config.NewConfig()
	pipeCfgs := config.NewPipelineConfig(cfg.PipelineCfgsPath)

	ctx := context.TODO()
	stream := make(chan interface{})
	errCh := make(chan error)

	for _, cfg := range pipeCfgs {
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

		rabbitmqConsumer, err := consumers.CreateConsumer(cfg.Consumer.Name, cfgParams)
		if err != nil {
			t.Error(err)
		}
		err = rabbitmqConsumer.Init()
		if err != nil {
			t.Error(err)
		}
		rabbitmqConsumer.Consume(context.TODO())
	}

	for {
		select {
		case msg := <-stream:
			data, err := json.MarshalIndent(msg, "", " ")
			if err != nil {
				t.Error(err)
			}
			t.Logf("data: %s", string(data))
			return
		case err := <-errCh:
			t.Logf("err: %v", err)
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
