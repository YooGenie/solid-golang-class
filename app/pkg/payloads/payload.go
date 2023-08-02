package payloads

// payloads 패키지 밑에 정의가 되어 있는 상위 모듈이다. 인터페이스로 되어있다.
// 하는 일: 여러 개의 프로세서가 돌아가고 있을 때 1번 프로세스가 작업을 마치고 그 다음 프로세서한테 데이터를 넘겨주거나 혹은 동시에 하나의 인스턴스를 바라보고
// 작업을 해야할 경우(만약에) 데이터베이스라고 한다. 여러개의 프로세서가 동시에 하나의 객체에 접근해서 데이터를 변조해야하는 작업을 할 때 클론이라는 딥 카피를 위한 메소드 시그니처 이다.
// 아예 새로운 복제본을 만들어서 그게 json오브젝트여서 복사를 해서 다음 프로세서에 넘겨줘서 그 다음 진행을 할 수 있도록 한다. 그런 용도로 클론을 만들었다.
type Payload interface {
	// Clone returns a new Payload that is a deep-copy of the original.
	Clone() Payload

	Out() (string, string, []byte)

	// MarkAsProcessed is invoked by the pipeline when the Payload either
	// reaches the pipeline sink or it gets discarded by one of the
	// pipeline stages.
	MarkAsProcessed()
}
