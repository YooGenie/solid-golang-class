package storage_providers_test

import (
	"encoding/json"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"event-data-pipeline/pkg/storage_providers"
	"fmt"
	"os"
	"sync"

	gc "gopkg.in/check.v1"
)

// go test -check.f ESSuite
type ESSuite struct{}

var _ = gc.Suite(&ESSuite{})

func (e *ESSuite) SetUpSuite(c *gc.C) {
	os.Args = nil
	os.Setenv("EDP_ENABLE_DEBUG_LOGGING", "false")
	logger.Setup()

}
func (e *ESSuite) TearDownSuite(c *gc.C) {

}

func (f *ESSuite) TestWrite(c *gc.C) {
	cfgObj := make(jsonObj)
	addresses := &[]interface{}{"http://localhost:9200"}
	cfgObj["addresses"] = addresses

	// ES storage provider 인스턴스 생성
	es, err := storage_providers.CreateStorageProvider("elasticsearch", cfgObj)

	// 에러 체크
	c.Assert(err, gc.IsNil)

	data := make(map[string]interface{})
	data["id"] = 0
	json, _ := json.Marshal(data)
	// 페이로드 stub 생성
	count := 1000
	for i := 0; i < count; i++ {
		payload := &esPayloadStub{"event-data-test", fmt.Sprintf("es.write.test.%d", i), json}
		written, err := es.Write(payload)
		c.Assert(written, gc.NotNil)
		c.Assert(err, gc.IsNil)
	}

}

func (f *ESSuite) TestConcurrentWrite(c *gc.C) {

	cfgObj := make(jsonObj)
	addresses := &[]interface{}{"http://localhost:9200"}
	cfgObj["addresses"] = addresses

	// ES storage provider 인스턴스 생성
	es, err := storage_providers.CreateStorageProvider("elasticsearch", cfgObj)

	// 에러 체크
	c.Assert(err, gc.IsNil)

	// 동시 트린잭션 개수
	requests := 1000

	// WaitGroup 생성
	var wg sync.WaitGroup

	// 고루틴 생성
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			data := make(map[string]interface{})
			data["id"] = idx
			json, _ := json.Marshal(data)
			// 페이로드 stub 생성
			payload := &esPayloadStub{"event-data-test-concurrent", fmt.Sprintf("es.write.test.%d", idx), json}
			c.Logf("payload: %v", payload)
			_, err := es.Write(payload)
			c.Assert(err, gc.IsNil)
		}(i)
	}
	c.Logf("waiting...")
	wg.Wait()
	c.Logf("completed...")
}

var _ payloads.Payload = new(esPayloadStub)

type esPayloadStub struct {
	index string
	docID string
	data  []byte
}

// Clone implements payloads.Payload
func (p *esPayloadStub) Clone() payloads.Payload {
	ps := &esPayloadStub{p.index, p.docID, p.data}
	return ps
}

// MarkAsProcessed implements payloads.Payload
func (*esPayloadStub) MarkAsProcessed() {

}

// Out implements payloads.Payload
func (p *esPayloadStub) Out() (string, string, []byte) {
	return p.index, p.docID, p.data
}
