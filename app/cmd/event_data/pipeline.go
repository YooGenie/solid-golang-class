package event_data

//pipeline.go에는 파이프라인을 구동하는 데 필요한 코드들이다.

import (
	"context"
	"encoding/json"
	"errors"
	"event-data-pipeline/pkg/config"
	"event-data-pipeline/pkg/consumers"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/pipelines"
	"event-data-pipeline/pkg/processors"
	"event-data-pipeline/pkg/sources"
	"event-data-pipeline/pkg/storage_providers"

	"sync"
)

type EventDataPipeline struct { // 이타입으로 생성을 해서 구동을 하는 로직이다.
	p        *pipelines.Pipeline
	cfgsPath string
	cfgs     []*config.PipelineCfg
}

func NewEventDataPipeline(cfg config.Config) (*EventDataPipeline, error) { // EventDataPipeline 스트럭 생성 부분
	ec := &EventDataPipeline{}

	// 설정 정보 경로 값 인스턴스에 저장.
	ec.cfgsPath = cfg.PipelineCfgsPath

	var err error
	// 제공된 경로로 부터 설정 정보를 읽어옵니다.
	ec.cfgs = config.NewPipelineConfig(cfg.PipelineCfgsPath)
	if ec.cfgs == nil {
		logger.Errorf("loaded configuration is nil")
		err = errors.New("loaded configuration is nil")
	}
	return ec, err
}

func (e *EventDataPipeline) SetCollectorRuntimeConfig(confs []*config.PipelineCfg) {
	e.cfgs = confs
}

func (e *EventDataPipeline) ValidateConfigs() error {
	// 인스턴스가 제로값인 경우 에러를 반환.
	if e == nil {
		logger.Errorf("%t is %v", e, e)
		return errors.New("EventDataPipeline instance is nil")
	}

	// 메모리에 로드된 설정 정보를 출력.
	if e.cfgs != nil {
		logger.Debugf("Loading EventDataPipeline Configs from memory: %s", ObjectToJsonString(e.cfgs))
	}

	// 메모리 상 설정 값이 비어있는 경우
	// 파일로부터 다시 읽기를 시도
	if e.cfgs == nil {
		e.cfgs = config.NewPipelineConfig(e.cfgsPath)
		logger.Infof("Loading EventDataPipeline Configs from file : %s", ObjectToJsonString(e.cfgs))
	}
	if e.cfgs == nil {
		return errors.New("did not pass configs validation.")
	}
	return nil
}

// 파이프라인을 구동하는 메소드
func (e *EventDataPipeline) Run() error { // 실제 런을 할 때 필요한 작업들을 여기서 한다.

	// Goroutine 실행 후 대기를 위한 WaiterGroup
	var wg sync.WaitGroup

	// 설정파일을 모두 담은 오브젝트
	cfgParams := make(jsonObj)

	// Context, Stream, Error Channel 을 전달하기 위한 오브젝트
	pipeParams := make(jsonObj)

	// Graceful Shutdown 을 위한 Context, CancelFunction
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	Put("context", pipeParams, ctx)

	// 컨슈머 읽기 채널
	stream := make(chan interface{})
	Put("stream", pipeParams, stream)

	// 컨슈머 에러 채널
	errCh := make(chan error)
	Put("errch", pipeParams, errCh)

	// 여기까지 런할때까지 필요한 파라미터를 다 셋팅을 한다.

	// loop through PipelineConfigs
	for _, cfg := range e.cfgs { // 로드한 설정 파일을 순회하면서
		wg.Add(1)

		//채널, 컨텍스트 삽입
		cfgParams["pipeParams"] = pipeParams

		//컨슈머 설정 값
		cfgParams["consumerCfg"] = cfg.Consumer.Config

		// 이벤트 기반 데이터를 소비하는 컨슈머 생성 => 의존성 주입을 통해 의존관계 역전을 한다.
		consumer, err := consumers.CreateConsumer(cfg.Consumer.Name, cfgParams) // consumers 패키지 안에 있는 CreateConsumer 함수를 사용해서 이름과 설정정보를 가진 파라미를 넘겨주면 다양한 컨슈머 타입을 생성할 수 있다.
		// 런타임에 이 코드가 실행 되면 사용자가 이름과 설정값을 넘기는 것에 따라 동적으 원하는 컨슈머 타입을 만들어서 프로그램을 구성할 수 있다.
		if err != nil {
			logger.Errorf("%v", err)
			return err
		}

		// 컨슈머 최초 작업 실행
		err = consumer.Init()

		if err != nil {
			logger.Errorf("%v", err)
			return err
		}

		logger.Debugf("%v consumer created", consumer)

		// 컨슈머로 부터 데이터를 받아 처리하는 0개 이상의 프로세서 슬라이스 초기화
		proccers := make([]processors.Processor, len(cfg.Processors)) //프로세서 생성 시작
		for i, p := range cfg.Processors {

			// 부여된 설정값 대로 프로세서 생성
			processor, err := processors.CreateProcessor(p.Name, p.Config) // CreateProcessor 언제 사용되냐? 런타임때 설정값대로 프로세서가 생성된다.
			// 만약 noop 시용한다면 Name에 noop프로세서가 들어갈 것 이다.
			if err != nil {
				return err
			}
			// 스테이지 러너에 생성된 프로세서를 등록
			proccers[i] = processor
		}

		//프로세서 생성 끝

		// 스테이지 러너 슬라이스 초기화
		stageRunners := make([]pipelines.StageRunner, len(proccers))

		// TODO: 설정 값에 따라 FIFO, WorkerPools 등 처리 방법을 선택
		for i, p := range proccers {
			fifo := pipelines.FIFO(p)
			stageRunners[i] = fifo
		}

		// 스토리지 프로바이더 생성
		storageProviders := make([]storage_providers.StorageProvider, len(cfg.Storages))
		for i, s := range cfg.Storages {
			logger.Debugf("storage[%d]: %v", i, s.Type)
			storageProviders[i], err = storage_providers.CreateStorageProvider(s.Type, s.Config)
			if err != nil {
				logger.Errorf("%v", err)
				return err
			}
		}

		// 파이프라인 초기화
		e.p = pipelines.New(stageRunners...)

		// 컨슈머 읽기 고루틴
		go consumer.Consume(ctx)

		// 파이프라인 프로세스 구동
		e.p.Process(&wg, ctx, consumer.(sources.Source), storageProviders, errCh)
	}
	wg.Wait()
	logger.Infof("shutting down the event data pipeline...")
	return nil
}

func Put(key string, obj jsonObj, data interface{}) {
	obj[key] = data
}

func (e *EventDataPipeline) GetCollectorRuntimeConfig() []*config.PipelineCfg {
	return e.cfgs
}

func ObjectToJsonString(obj interface{}) string {
	b, err := json.Marshal(obj)
	if err != nil {
		logger.Panicf("%v", err)
	}
	return string(b)
}
