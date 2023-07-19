package concur

import (
	"encoding/json"
	"event-data-pipeline/pkg/logger"
	"time"

	"github.com/google/uuid"
)

type Task func(interface{}) (int, error)

type WorkerPool struct {
	ID     string
	name   string
	size   int
	ch     chan interface{}
	signal chan bool
	task   Task
}

func NewWorkerPool(name string, ch chan interface{}, size int, task Task) *WorkerPool {
	id := uuid.New()
	return &WorkerPool{
		ID:     id.String(),
		name:   name,
		size:   size,
		ch:     ch,
		signal: make(chan bool),
		task:   task,
	}
}

func (w *WorkerPool) runTask(nbr int) {
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
}

func (w *WorkerPool) Start() {
	for i := 0; i < w.size; i++ {
		go w.runTask(i)
	}
}

func (w *WorkerPool) Stop() {
	for i := 0; i < w.size; i++ {
		w.signal <- true
	}
	logger.Infof("%v done shutting down", w.name)
}
