package storage_providers

import (
	"context"
	"encoding/json"
	"errors"
	"event-data-pipeline/pkg/concur"
	"event-data-pipeline/pkg/fs"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"

	"event-data-pipeline/pkg/ratelimit"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

var _ StorageProvider = new(FilesystemClient)

func init() {
	Register("filesystem", NewFilesystemClient)
}

// Filesystem Config includes storage settings for filesystem
type FsCfg struct {
	Path   string `json:"path,omitempty"`
	Worker int    `json:"worker,omitempty"`
	Buffer int    `json:"buffer,omitempty"`
}

type FilesystemClient struct {
	RootDir     string
	file        fs.File
	count       int
	mu          sync.Mutex
	workers     *concur.WorkerPool
	inCh        chan interface{}
	rateLimiter *rate.Limiter
}

func NewFilesystemClient(config jsonObj) StorageProvider {
	var fsc FsCfg
	// 바이트로 변환
	cfgByte, _ := json.Marshal(config)

	// 설정파일 Struct 으로 Load
	json.Unmarshal(cfgByte, &fsc)

	fc := &FilesystemClient{
		RootDir:     fsc.Path,
		inCh:        make(chan interface{}, fsc.Buffer),
		count:       0,
		rateLimiter: ratelimit.NewRateLimiter(ratelimit.RateLimit{Limit: 10, Burst: 0}),
	}
	numWorkers := 1
	if fsc.Worker > 0 {
		numWorkers = fsc.Worker
	}

	fc.workers = concur.NewWorkerPool("filesystem-workers", fc.inCh, numWorkers, fc.Write)
	fc.workers.Start()

	return fc
}

// 제공된 config 파일 대로 fs/ 라는 경로 저장이 될 것이다.
func (f *FilesystemClient) Write(payload interface{}) (int, error) {
	if payload != nil {

		index, docID, data := payload.(payloads.Payload).Out()

		f.mu.Lock()
		// Write 메소드가 리턴 할 때 Unlock
		defer f.mu.Unlock()
		f.file = fs.NewFile(index, docID, data)
		f.count += 1

		ctx := context.Background()
		retry := 0

		for {
			startWait := time.Now()
			f.rateLimiter.Wait(ctx)
			logger.Debugf("rate limited for %f seconds", time.Since(startWait).Seconds())

			limit := 1
			defer func() {
				if err := recover(); err != nil {
					logger.Println("Write to file failed:", err)
				}
			}()

			os.MkdirAll(fmt.Sprintf("%s/%s", f.RootDir, f.file.SubDir), 0775)
			err := ioutil.WriteFile(fmt.Sprintf("%s/%s/%s", f.RootDir, f.file.SubDir, f.file.Name), f.file.Data, 0775)

			// 성공적인 쓰기에 리턴.
			if err == nil {
				return 1, nil
			}

			retry++
			if limit >= 0 && retry >= limit {
				return 0, err
			}
			time.Sleep(time.Duration(5) * time.Second)
		}
	}
	return 0, errors.New("payload is nil")
}

// Drain implements pipelines.Sink
func (f *FilesystemClient) Drain(ctx context.Context, p payloads.Payload) error {
	f.inCh <- p
	return nil
}
