package storage_providers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"event-data-pipeline/pkg/concur"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"event-data-pipeline/pkg/ratelimit"
	"fmt"
	"sync"
	"time"

	spes "event-data-pipeline/pkg/storage_providers/es"

	es "github.com/elastic/go-elasticsearch/v8"
	"golang.org/x/time/rate"
)

// var _ pipelines.Sink = new(ElasticSearchClient)

type ElasticSearchClientConfig struct {
	RateLimit  ratelimit.RateLimit `json:"rate_limit,omitempty"`
	MaxRetries int                 `json:"max_retries,omitempty"`
	Delay      int                 `json:"delay,omitempty"`
	Worker     int                 `json:"worker,omitempty"`
}
type ElasticSearchClient struct {
	client       *es.Client
	DocumentType string
	Refresh      bool

	buf   bytes.Buffer
	mu    sync.Mutex
	count int

	workers *concur.WorkerPool
	inCh    chan interface{}

	rateLimiter *rate.Limiter
	maxRetries  int
	delay       int
}

func init() {
	Register("elasticsearch", NewElasticSearchClient)
}

func NewElasticSearchClient(config jsonObj) StorageProvider {
	// 바이트 변환
	cfgData, _ := json.Marshal(config)
	var escConf ElasticSearchClientConfig
	var esConf es.Config
	json.Unmarshal(cfgData, &escConf)
	json.Unmarshal(cfgData, &esConf)

	es, err := es.NewClient(esConf)
	if err != nil {
		logger.Fatalf("error in creating elasticsearch client: %s", err)
	}

	transport, _ := json.Marshal(es.Transport)
	logger.Debugf("Elasticsearch transport: %s", string(transport))

	ec := &ElasticSearchClient{
		client: es,

		inCh:        make(chan interface{}),
		rateLimiter: ratelimit.NewRateLimiter(escConf.RateLimit),
		maxRetries:  escConf.MaxRetries,
		delay:       escConf.Delay,
	}

	numWorkers := 1
	if escConf.Worker > 0 {
		numWorkers = escConf.Worker
	}

	ec.workers = concur.NewWorkerPool("elasticsearch-workers", ec.inCh, numWorkers, ec.Write)
	ec.workers.Start()

	return ec
}

func (e *ElasticSearchClient) Drain(ctx context.Context, p payloads.Payload) error {
	e.inCh <- p
	return nil
}

func (e *ElasticSearchClient) Write(payload interface{}) (int, error) {
	if payload == nil {
		return 0, errors.New("payload is nil")
	}
	// 페이로드 가져오기
	index, docID, data := payload.(payloads.Payload).Out()
	if index == "" || docID == "" || len(data) == 0 {
		return 0, errors.New("payload is nil")
	}
	// 락 가져오기
	e.mu.Lock()
	defer e.mu.Unlock()

	// documentID 메타정보 오브젝트 생성
	meta := []byte(fmt.Sprintf(`{ "index" : { "_id" : "%v" } }%s`, docID, "\n"))

	// bulk write 을 위한 개행
	data = append(data, "\n"...)

	// 카운터
	e.count++

	// 메타, 데이타 오브젝트 사이즈 버퍼 할당
	e.buf.Grow(len(meta) + len(data))

	// 메타 정보 쓰기
	e.buf.Write(meta)

	// 데이터 정보 쓰기
	e.buf.Write(data)

	// 1000개 일때 벌크 쓰기
	if e.count >= 1000 { // 카운트가 1000개일 때 벌크를 해달라
		// 로컬 데이터 카피
		buf := e.buf.Bytes()
		// 벌크라이트
		logger.Debugf("trigger bulk write : %d", e.count)
		written, err := e.bulkWrite(index, buf)
		//too many request error가 발생할 수 있다.
		// 엘레스틱서치가 어떤 스팩, 메모리 자원으로 구성되어있어서 그거에 따라 다를 수 있지만 보통은 단일건이 아니라 벌크 라이크로 구현을안한다.
		if err != nil {
			return 0, nil
		}
		// 버퍼 초기화
		e.buf.Reset()
		// 카운트 초기화
		e.count = 0
		return written, nil
	}
	return 0, nil
}

func (e *ElasticSearchClient) bulkWrite(index string, data []byte) (int, error) {

	logger.Debugf("writing data: %s", string(data))
	retry := 0
	reader := bytes.NewReader(data)

	for {
		ctx := context.Background()
		startWait := time.Now()
		// rate limiting ...
		e.rateLimiter.Wait(ctx)
		logger.Debugf("rate limited for %f seconds", time.Since(startWait).Seconds())

		// Observe write duration in seoncds and set histogram and guage metric

		res, err := e.client.Bulk(reader,
			e.client.Bulk.WithIndex(index),
		)
		// 에러가 발생했거나, 결과값이 없는 경우
		if err != nil {
			logger.Errorf("error in bulk writing : %s", err.Error())
			retry++
			if e.maxRetries >= 0 && retry > e.maxRetries {
				err := fmt.Errorf("retry[%d] exceeded max retries[%d]", retry, e.maxRetries)
				logger.Errorf("error in bulk writing : %s", err.Error())
				return 0, err
			}
			time.Sleep(time.Duration(time.Duration(e.delay) * time.Second))
			logger.Infof("retrying[%d/%d]", retry, e.maxRetries)
			continue
		}
		defer res.Body.Close()

		numIndexed := 0
		numErrors := 0
		//응답에 에러가 없는 경우
		if !res.IsError() {
			var blk *spes.BulkResponse
			err := json.NewDecoder(res.Body).Decode(&blk)
			if err != nil {
				logger.Errorf("Failure to to parse response body: %s", err)
				return 0, err
			}
			for _, d := range blk.Items {
				// 201 코드 이상의 경우
				if d.Index.Status > 201 {
					// ... increment the error counter ...
					//
					// ... and print the response status and error information ...
					logger.Errorf("Error: [%d]: %s: %s: %s: %s",
						d.Index.Status,
						d.Index.Error.Type,
						d.Index.Error.Reason,
						d.Index.Error.Cause.Type,
						d.Index.Error.Cause.Reason,
					)
					// 201 코드 이하의 경우 성공 처리
				} else {
					logger.Debugf("Success: ID[%s] Result[%s] Status[%d] ",
						d.Index.ID,
						d.Index.Result,
						d.Index.Status)
					numIndexed++
				}
			}
			//prometheus metrics counter
			esWriteTotal.Add(float64(numIndexed))
			return numIndexed, nil
			// 응답에 에러가 있는 경우
		} else {
			var bodyObj jsonObj
			if err := json.NewDecoder(res.Body).Decode(&bodyObj); err != nil {
				logger.Errorf("Failure to to parse response body: %s", err)
			} else {
				logger.Printf("Error: [%d] %s: %s",
					res.StatusCode,
					bodyObj["error"].(jsonObj)["type"],
					bodyObj["error"].(jsonObj)["reason"],
				)
			}
			return numErrors, nil
		}
	}
}
