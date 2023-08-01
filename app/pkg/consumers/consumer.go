package consumers

import (
	"context"
	"errors"
	"event-data-pipeline/pkg/logger"
	"fmt"
	"strings"
)

type (
	jsonObj = map[string]interface{}
	jsonArr = []interface{}
)

// consumer interface
type Consumer interface { // 이 인터페이스를 구현하는 상세 타입들은 kafka.go
	//초기 작업
	Initializer
	//읽어 오기
	Consume(ctx context.Context) error // Consume 메소드 시그니처
	// 선언이 인터페이스로 되어있지만 실제로 상세 페이지에서 어떻게 구현 되는지 따라 동작하는 행동이 달라진다.
}

// 컨슈머 팩토리 함수
type ConsumerFactory func(config jsonObj) Consumer // ConsumerFactory는 일종의 함수있다. 설정값을 받아서 Consumer 타입을 반환하는 함수이다. 저희가 원하는 카푸카 이거나 레디큐 을 실제 구현 상세 구현 타입을 반환해주는 메소드이다.
// 이걸 가능하게 해주는 것이 맨 처음에 Register 함수를 통해서 필요한 Consumer를 사전에 등록 했기 때문이다.

// 컨슈머 팩토리 저장소
var consumerFactories = make(map[string]ConsumerFactory) //이안에 consumerFactories를 담고 있는 맵이라는 객체를 확인할 수 있다.

// 컨슈머를 최초 등록하기 위한 함수
func Register(name string, factory ConsumerFactory) { // 컨슈머팩토리를 레지즈터를 통해서 등록하고 있다.
	logger.Debugf("Registering consumer factory for %s", name)
	if factory == nil {
		logger.Panicf("Consumer factory %s does not exist.", name)
	}
	_, registered := consumerFactories[name] // 등록을 해서 컨슈머팩토리에 카푸카라는 이름으로 등록을 했다. 이후에 CreateConsumer 라는 메소드를 프로그램을 구동할 때 원하는 이름 (카푸카 or 레디큐)이 될 수 있다.
	if registered {
		logger.Errorf("Consumer factory %s already registered. Ignoring.", name)
	}
	consumerFactories[name] = factory
}

// 컨슈머를 사용자의 설정값에 따라 반환하는 함수
func CreateConsumer(name string, config jsonObj) (Consumer, error) { // 컨슈머의 이름과 컨슈머를 생성할 때 필요한 설정값을 담는 jsonObj 인자로 받는다.
	//  name => 카푸카 or 레디큐 될 수 있다. 이름에 따라 컨슈머 인터페이스에 상세 구현 타입을 팩토리로 부터 가져오게 한다.
	factory, ok := consumerFactories[name] // consumerFactories 객체를 앞서 선언하고 있다.
	if !ok {
		availableConsumers := make([]string, 0)
		for k := range consumerFactories {
			availableConsumers = append(availableConsumers, k)
		}
		return nil, errors.New(fmt.Sprintf("Invalid Consumer name. Must be one of: %s", strings.Join(availableConsumers, ", ")))
	}

	// Run the factory with the configuration.
	return factory(config), nil // CreateConsumer 함수는 factory를 반환하고 넘겨 받는 설정값을 넘겨 준다.
	// factory가 무엇이 되냐 우리가 카푸카를 넘겼으면 NewKafkaConsumerClient를 반환할 것이다. 이안에서는 원하는 설정 클라이언트를 반환할 것이다.
	//client := &KafkaConsumerClient{ // 클라이언트 카푸카 설정값
	//	Consumer: kfkCnsmr,
	//	Source:   sources.NewKafkaSource(kfkCnsmr),
	//} ==> 이부분을 다른 파일에 있다.
	// 클라이언트를 반환하고 그렇게 생성된 컨슈머를 가지고 프로그래밍을 가지고 사용한다.
}
