package pipelines

import (
	"context"
	"event-data-pipeline/pkg/processors"
	"sync"
)

type fixedWorkerPool struct {
	fifos []StageRunner
}

// FixedWorkerPool returns a StageRunner that spins up a pool containing
// numWorkers to process incoming payloads in parallel and emit their outputs
// to the next stage.
// FixedWorkerPool 여러개의 고루틴이 있어서 워커가 대비를 하고 있다가 잡이 들어오면 분산으로 처리한다.

func FixedWorkerPool(proc processors.Processor, numWorkers int) StageRunner {
	if numWorkers <= 0 {
		panic("FixedWorkerPool: numWorkers must be > 0")
	}

	fifos := make([]StageRunner, numWorkers)
	for i := 0; i < numWorkers; i++ {
		fifos[i] = FIFO(proc)
	}

	return &fixedWorkerPool{fifos: fifos}
}

// Run implements StageRunner.
func (p *fixedWorkerPool) Run(ctx context.Context, params StageParams) {
	var wg sync.WaitGroup

	// Spin up each worker in the pool and wait for them to exit
	for i := 0; i < len(p.fifos); i++ {
		wg.Add(1)
		go func(fifoIndex int) {
			p.fifos[fifoIndex].Run(ctx, params)
			wg.Done()
		}(i)
	}

	wg.Wait()
}
