package storage_providers_test

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"event-data-pipeline/pkg/pipelines"
	"event-data-pipeline/pkg/storage_providers"
	"testing"

	"fmt"
	"os"
	"sync"

	gc "gopkg.in/check.v1"
)

// go test -check.f FilesystemSuite
type FilesystemSuite struct{}

var _ = gc.Suite(&FilesystemSuite{})

// 실행 순서 : SetUpSuite -> TestWrite -> TearDownSuite
func (f *FilesystemSuite) SetUpSuite(c *gc.C) {
	os.Args = nil
	os.Setenv("EDP_ENABLE_DEBUG_LOGGING", "true") // 환경변수 셋팅
	logger.Setup()                                // 로거

	fmt.Println("Setting up suite: clearing fs directory...")
	err := os.RemoveAll("fs") // 시작할 때 fs 디텍토리가 존재하고 있다면 그걸 없애고 클린한 상태로 하겠다.
	c.Assert(err, gc.IsNil)
}
func (f *FilesystemSuite) TearDownSuite(c *gc.C) {
	fmt.Println("Tearind down the suit : clearing fs directory...")
	err := os.RemoveAll("fs") //  디렉토리 삭제하고 종료
	c.Assert(err, gc.IsNil)
}
func (f *FilesystemSuite) TestWrite(c *gc.C) {

	// filesystem config 오브젝트 생성
	fsCfg := make(jsonObj) // 설정값을 넘겨주기 위한 오브젝트 값
	fsCfg["path"] = "fs/"  // 필요한 설정값 패스를 넣어줌

	// filesystem storage provider 인스턴스 생성
	filesystem, err := storage_providers.CreateStorageProvider("filesystem", fsCfg) // StorageProvider 생성해서 받아온다.

	// 에러 체크
	c.Assert(err, gc.IsNil) // nil인지 아닌지 체크

	// 페이로드 stub 생성
	// fsPayloadStub는 페이로드 인터페이스를 만족시키는 별도의 비즈니스 로직이 없는 최소한의 요건만 만족시키는 형태로 생성해서 Write 한다.
	payload := &fsPayloadStub{"event-data-test", fmt.Sprintf("filesystem.write.test.%d", 0)}

	filesystem.Write(payload)

	dirs, err := os.ReadDir("fs/event-data-test") // 에러 있는지 확인하고
	c.Assert(err, gc.IsNil)

	for _, dir := range dirs { // 받아온 디렉토리를 내가 쓸 때 이름과 동일한 지 테스트를 한다. 테스트 후 종료 된다.
		c.Assert("filesystem.write.test.0", gc.Equals, dir.Name())
	}

}

func (f *FilesystemSuite) TestConcurrentWrite(c *gc.C) {

	// filesystem config 오브젝트 생성
	fsCfg := make(jsonObj)
	fsCfg["path"] = "fs/"

	// filesystem storage provider 인스턴스 생성
	filesystem, err := storage_providers.CreateStorageProvider("filesystem", fsCfg)

	// 에러 체크
	c.Assert(err, gc.IsNil)

	// TestConcurrentWrite : 앞에 생성하는 건 동일

	// 동시 트린잭션 개수
	requests := 10 // 동시에 요청하는 리퀘스트 개수를 설정한다.

	// WaitGroup 생성
	var wg sync.WaitGroup

	// 고루틴 생성
	//for문을 돌면서 고루틴 생성한다.
	//고루틴안에서는 페이로드를 생성해서 write 한다.
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(idx int, wg *sync.WaitGroup) {
			// 페이로드 stub 생성
			payload := &fsPayloadStub{"event-data-test-concurrent", fmt.Sprintf("filesystem.write.test.%d", idx)}
			filesystem.Write(payload)
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	// 동시에 몇개를 쓸수있는지 테스팅이 가능하다.

	dirs, err := os.ReadDir("fs/event-data-test-concurrent") // 디렉토리에 쓰고 난 후에 파일 이름들을 비교해서 테스트를 마친다.
	c.Assert(err, gc.IsNil)

	for idx, dir := range dirs {
		c.Assert(fmt.Sprintf("filesystem.write.test.%d", idx), gc.Equals, dir.Name())
	}
}

type jsonObj = map[string]interface{}

var _ payloads.Payload = new(fsPayloadStub)

type fsPayloadStub struct {
	dir      string
	filename string
}

// Clone implements payloads.Payload
func (*fsPayloadStub) Clone() payloads.Payload {
	ps := &fsPayloadStub{}
	return ps
}

// MarkAsProcessed implements payloads.Payload
func (*fsPayloadStub) MarkAsProcessed() {

}

// Out implements payloads.Payload
func (p *fsPayloadStub) Out() (string, string, []byte) {
	return p.dir, p.filename, []byte(`{}`)
}

// 테스트 케이스
var table = []struct {
	input  int
	worker int
	buffer int
}{
	{
		input: 10000, worker: 100, buffer: 100,
	},
	{
		input: 10000, worker: 100, buffer: 1000,
	},
	{
		input: 10000, worker: 100, buffer: 5000,
	},
	{
		input: 10000, worker: 100, buffer: 10000,
	},
}

// 실행하는 코드
// 1)  cd app/pkg/storage_providers
// go test -bench=. -benchtime=10s -benchmem -run=^#
func BenchmarkWrite(b *testing.B) {
	for _, v := range table { // 테스트 케이스를 순회하면서
		b.Run(fmt.Sprintf("input_size_%d_worker_size_%d_buffer_size_%d", v.input, v.worker, v.buffer), func(b *testing.B) {
			for i := 0; i < b.N; i++ { // Benchmark 라이브러리가 특정횟수만큼 반복을 하면서 테스트를 진행합니다.
				FilesystemWrite(v.input, v.worker, v.buffer) // 함수를 실행하는데 실행되는 시간과 소유된 메모리와 새로 오브젝트를 할당하는 횟수가 몇번 발생했는지 알수있다.
			}
		})
	}
}

// 이 업무 단위로 벤치마크를 테스트팅을 한다.
func FilesystemWrite(num int, worker int, buffer int) {
	// filesystem config 오브젝트 생성 => 설정값 생성
	fsCfg := make(jsonObj)
	fsCfg["path"] = "fs/"
	fsCfg["worker"] = worker
	fsCfg["buffer"] = buffer

	// fsClient storage provider 인스턴스 생성 => 파일시스템 클라이언트를 생성
	fsClient, err := storage_providers.CreateStorageProvider("filesystem", fsCfg)

	if err != nil {
		fmt.Errorf(err.Error())
	}
	sink := fsClient.(pipelines.Sink) // sink 타입으로 변환을 한다.
	for i := 0; i < num; i++ {
		// stub 페이로드 생성
		payload := &fsPayloadStub{"event-data-test", fmt.Sprintf("filesystem.write.test.%d", i)}

		sink.Drain(context.TODO(), payload) // 지정 개수만큼 실행하고 종료하는 함수
	}
}
