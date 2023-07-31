package consumers

type Initializer interface { // 인터페이스 분리 원칙을 적용
	Init() error
}

// Initializer 인터페이스를 통해서 초기에 컨슈머마다 해야할 작업들을 실행하게 된다. Init이라는 인터페이스 시그니처는 구현 상세 타입에서 구현하고 있다는 걸 알 수 있다. => func (kc *KafkaConsumerClient) Init() error {} 여기
