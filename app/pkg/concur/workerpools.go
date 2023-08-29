package concur

import (
	"encoding/json"
	"event-data-pipeline/pkg/logger"
	"time"

	"github.com/google/uuid"
)

type Task func(interface{}) (int, error) //일종의 함수 타입이다. Task는 매개변수 인터페이스이고 반환값은 인트와 에러다.

type WorkerPool struct {
	ID     string
	name   string
	size   int
	ch     chan interface{}
	signal chan bool
	task   Task
}

func NewWorkerPool(name string, ch chan interface{}, size int, task Task) *WorkerPool { // WorkerPool이라는 포인트 스트럭을 반환하는걸 알 수 있다.
	id := uuid.New()
	return &WorkerPool{
		ID:     id.String(),
		name:   name,
		size:   size, // size
		ch:     ch,   // 초기화가 된 채널
		signal: make(chan bool),
		task:   task, // Write 메소드를 전해준다는 걸 알 수 있다.
	}

	//WorkerPool이라는 일반적으로 많이 사용하는 동시성 컨트롤할 때 사용하는 보편적인 방법이다. 여러개의 워커들이 동시적으로 있어서 잡이 들어왔을 때 그걸 팬아웃한다. 파이브라인 잡에 비유하면 여러개의 워커들이 있어서 그 잡을 분배해줘서 그 안에 있는 워커들이 동시적으로 데이터를 처리할 수 있도록 하는 방법이다.
}

func (w *WorkerPool) runTask(nbr int) { // Task를 초기화를 할 때 WorkerPool 구조체 안에 넣어줬다. 그것이  w.task(data) 이다.
	for {
		select {
		case data := <-w.ch:
			start := time.Now()
			_json, _ := json.MarshalIndent(data, "", " ")
			logger.Debugf("%v [#%v] worker [%v] write data: [%s]...", w.name, nbr, w.ID, _json)
			size, err := w.task(data)
			if err != nil {
				logger.Errorf("%v [#%v] handler [%v] error: %v", w.name, nbr, w.ID, err)
			}
			logger.Debugf("%v [#%v] handler [%v] written %v in %v ms...", w.name, nbr, w.ID, size, time.Since(start).Milliseconds())
		case <-w.signal:
			logger.Infof("%v [#%v] received shutdown signal", w.name, nbr)
			return
		}
	}
	// for 안에서는 채널 리스닝을 하고 있다. for select를 통해서 동시 다발적으로 여러 채널의 리스닝을 하고 있다. 워터풀을 초기화할 때 넣어준 인터페이스 ch 채널을 통해서 데이터를 읽어 오고
	// w.task(data) 함수를 호출해서 작업이 이루어진다. 범용적으로 사용할 수 있도록 설계 되어있다. 나는 워커풀이라는 동시성 모델을 어디 적용하고 싶으면 파일시스템처럼 사용하고자 하는 채널, 워커개수, 수행해야할 테스크를
	// 넘겨줘서 초기화를 해주고 스타트를 하게 되면 워커풀에 워커들이 실행되면서 테스크를 실제 실행하게 된다. => 데이터 처리 성능을 높여줄 수 있다.
}

func (w *WorkerPool) Start() {
	for i := 0; i < w.size; i++ {
		go w.runTask(i)
	}
	// for문을 돌면서 지정된 사이즈 만큼 고루틴을 실행한다. 고루틴을 실행한다는 것은 runTask가 뭔지 봐야한다.
}

func (w *WorkerPool) Stop() {
	for i := 0; i < w.size; i++ {
		w.signal <- true
	}
	logger.Infof("%v done shutting down", w.name)
}
