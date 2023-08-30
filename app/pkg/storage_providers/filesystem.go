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
type FsCfg struct { //로드를 하기 위한 구조체와 구조체안에 필드를 선 => 그 필드에서 로딩히고자는 태그를 설정한다.
	Path   string `json:"path,omitempty"`
	Worker int    `json:"worker,omitempty"`
	Buffer int    `json:"buffer,omitempty"`
}

type FilesystemClient struct {
	RootDir string
	file    fs.File
	count   int
	mu      sync.Mutex // 여러개 고루틴 실행되면 읽기, 쓰기 작업이 동시에 일어날 때 데이터 레이스 현상이 발생할 수 있다. 2개의 고루틴이 하나의 데이터를 접근을 해서 두명이 다 쓰려고 한다. 데이터 컬옵션?
	//매번 실행할 때마다 결과값이 달라질 수 있다. 그런 현상을 막도록하기 위해서 일종에 작업을 하기 위한 자기만에 공간을 독립적으로 확보하는 용도로 뮤텍스를 사용한다.
	workers     *concur.WorkerPool
	inCh        chan interface{}
	rateLimiter *rate.Limiter
}

func NewFilesystemClient(config jsonObj) StorageProvider { // NewFilesystemClient가 StorageProvider 팩토리로 등록할 때 사용하는 함수이다.
	var fsc FsCfg //설정값을 로드
	// 바이트로 변환
	cfgByte, _ := json.Marshal(config) // config가 들어오면 얘는 map 타입이기 때문에 바이트로 전환

	// 설정파일 Struct 으로 Load
	json.Unmarshal(cfgByte, &fsc) //바이트를 Struct로 전환

	fc := &FilesystemClient{
		RootDir:     fsc.Path,
		inCh:        make(chan interface{}, fsc.Buffer),
		count:       0,
		rateLimiter: ratelimit.NewRateLimiter(ratelimit.RateLimit{Limit: 10, Burst: 0}),
	} // 원하는 값을 뽑아서 초기화를 한다.

	// 워커의 개수도 지정해준다.
	numWorkers := 1
	if fsc.Worker > 0 {
		numWorkers = fsc.Worker
	}

	// 일련에 초기화 작업이 있다.

	fc.workers = concur.NewWorkerPool("filesystem-workers", fc.inCh, numWorkers, fc.Write) //WorkerPool를 생성해준다.
	// inCh는 input 채널를 통해서 데이터를 전송할 때 drain 할때 FilesystemClient안에서 내부 전송 시스템이 된다
	// fc.Write 메소드가 실제로는 Write 관련된 비즈니스를 구현하는 로직이다. Write를 통해서 덕타이핑이 만족한다.
	fc.workers.Start() // 워커는 실행하고 있다.

	return fc
}

// 제공된 config 파일 대로 fs/ 라는 경로 저장이 될 것이다.
func (f *FilesystemClient) Write(payload interface{}) (int, error) { // 저장하는 컴포넌트 => 파일 시스템에 저장하는 코드
	if payload != nil {

		// 페이로 타입으로 형변환을 하고 아웃이라는 메소드를 통해서 3가지 반환값을 받는다.
		index, docID, data := payload.(payloads.Payload).Out() // Out를 실제 사용하는 곳
		// out을 통해서 원하는 데이터를 반환하고 저장하는 로직을 탄다.

		f.mu.Lock() //  방에 들어가서 문을 잠그겠다. 그때 락 사용
		// Write 메소드가 리턴 할 때 Unlock
		defer f.mu.Unlock()                     // 메소드가 작업이 끝나면 언락을 하겠다. // defer :실행하는 시점이 언제이냐? 리턴하고 종료될 때 이 함수를 꼭 실행하겠다.
		f.file = fs.NewFile(index, docID, data) // 파일을 쓰기 위한 작업을 한다.
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

			os.MkdirAll(fmt.Sprintf("%s/%s", f.RootDir, f.file.SubDir), 0775) // 디렉토리를 만들고 거기 정해진 형태대로 데이터를 쓰기 작업을 한다.
			err := ioutil.WriteFile(fmt.Sprintf("%s/%s/%s", f.RootDir, f.file.SubDir, f.file.Name), f.file.Data, 0775)

			// 성공적인 쓰기에 리턴.
			if err == nil {
				return 1, nil
			}

			// 실패했을 때 리트라이를 한다.
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
	f.inCh <- p // input 채널이 하는 역할은 Drain을 하는 역할이다. 처리가 완료 된 데이터를 받아서 내부로 있는 채널로 전송하는 역할을 한다.
	return nil
}
