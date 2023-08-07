package processors

// 데이터를 차리하는 부분
import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"fmt"
	"strings"
)

type jsonObj = map[string]interface{}

type ProcessorFactory func(config jsonObj) Processor // ProcessorFactory jsonObj 형식 설정 파일을 받고 Processor 인터페이스 리턴을 받는다.

var processorFactories = make(map[string]ProcessorFactory) // 맵으로 초기화를 해놨다. key는 string 타입이고 값은 ProcessorFactory 타입

// Each processor implementation must Register itself
func Register(name string, factory ProcessorFactory) { // 이름이랑 ProcessorFactory 타입으로 입력을 받는다
	logger.Debugf("Registering processor factory for %s", name)
	if factory == nil {
		logger.Panicf("Processor factory %s does not exist.", name)
	}
	_, registered := processorFactories[name] // 닐이 아니면 ProcessorFactory에 이름이 있는 없는지 확인을 하고
	// processorFactories 타입은 맵이다.
	if registered {
		logger.Errorf("Processor factory %s already registered. Ignoring.", name)
	}
	processorFactories[name] = factory // 여기에 넣어 준다.
}

// CreateProcessor is a factory method that will create the named processor
func CreateProcessor(name string, config jsonObj) (Processor, error) { // CreateProcessor 언제 사용 되는지? event_data/pipeline.go에서 사용

	factory, ok := processorFactories[name]
	if !ok {
		// Factory has not been registered.
		// Make a list of all available datastore factories for logging.
		availableProcessors := make([]string, 0)
		for k := range processorFactories {
			availableProcessors = append(availableProcessors, k)
		}
		return nil, fmt.Errorf("invalid Processor name. Must be one of: %s", strings.Join(availableProcessors, ", "))
	}

	// Run the factory with the configuration.
	return factory(config), nil
}

// 모든 프로세서는 본 인터페이스를 구현해야함.
type Processor interface {
	Process(context.Context, payloads.Payload) (payloads.Payload, error) // Process 메서드는 페이로드라는 인터페이스를 구현하는 객체를 주고 받는다.
	// 프로세스라고 하는 메소드 시그니처가 있다.
	// context.Context : 현재는 사용하고 있지 않지만 프로그램 전반적으로 실행이 될 때 프로그램 동작에 어떤 실행 이후에 취소나 그런 것들을 아울러서 담하기 위해서는 context라는 파라미터를 받아서 컨텍스트가 취소가 됐으면 이 프로세스는 중단을 한다는 것들이
	// 필요할 수 있으니까 컨텍스트 파라미터를 넘겨줬다.
	// Payload는 무조건 필요하다. Payload를 받아서 처리를 해서 반환 받아야한다. Payload를 처리하다가 어떤 문제가 발생하면 에러를 반환하도록 설계 했다.

}

// 일반 func 를 프로세서 인터페이스 타입으로 사용할 수 있도록 도와주는 ProcessorFunc 타입
type ProcessorFunc func(context.Context, payloads.Payload) (payloads.Payload, error)

// ProcessorFunc 를 Processor 인터페이스를 구현하도록 도와주는 Process 메소드
func (f ProcessorFunc) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	return f(ctx, p)
}
