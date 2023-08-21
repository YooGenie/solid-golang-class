package cmd

import (
	"event-data-pipeline/cmd/event_data"
	"event-data-pipeline/cmd/server"
	"event-data-pipeline/pkg/config"
	"event-data-pipeline/pkg/logger"
	"log"
	_ "net/http/pprof"
	"runtime/debug"
	"time"
)

type (
	jsonObj = map[string]interface{}
	jsonArr = []interface{}
)

// Run is the entrypoint for running pipeline
func Run(cfg config.Config, http *server.HttpServer) { //서비스를 직접 구동하는 메소드

	// Run Http Server
	http.Serve() // 서버 띄우고

	// Force garbage collection
	go GarbageCollector()

	// EventDataPipeline 타입의 인스턴스를 생성합니다.
	edp, err := event_data.NewEventDataPipeline(cfg) // EventDataPipeline 실제 생성하는 부분이 여기 있 런을 하기 전에 생성하는 과정이다.

	// 파이프라인 인스턴스 생성에 실패할 경우
	// 프로그램 동작을 멈춥니다.
	if err != nil {
		log.Panicf(err.Error())
	}

	// 설정 값 유효 성을 통과하지 못할 경우
	// 프로그램 동작을 멈춥니다.
	err = edp.ValidateConfigs()
	if err != nil {
		log.Panicf(err.Error())
	}

	// 파이프라인 프로세스를 구동하는 메소드
	err = edp.Run()
	if err != nil {
		logger.Errorf(err.Error())
	}

	logger.Infof("shutting down service.")
}

func GarbageCollector() {
	gcTimer := time.NewTicker(1 * time.Second)

	for _ = range gcTimer.C {
		debug.FreeOSMemory()
	}
}
