package processors

import (
	"context"
	"event-data-pipeline/pkg/logger"
	"event-data-pipeline/pkg/payloads"
	"fmt"
	"strings"
)

type jsonObj = map[string]interface{}

type ProcessorFactory func(config jsonObj) Processor

var processorFactories = make(map[string]ProcessorFactory)

// Each processor implementation must Register itself
func Register(name string, factory ProcessorFactory) {
	logger.Debugf("Registering processor factory for %s", name)
	if factory == nil {
		logger.Panicf("Processor factory %s does not exist.", name)
	}
	_, registered := processorFactories[name]
	if registered {
		logger.Errorf("Processor factory %s already registered. Ignoring.", name)
	}
	processorFactories[name] = factory
}

// CreateProcessor is a factory method that will create the named processor
func CreateProcessor(name string, config jsonObj) (Processor, error) {

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
	Process(context.Context, payloads.Payload) (payloads.Payload, error)
}

// 일반 func 를 프로세서 인터페이스 타입으로 사용할 수 있도록 도와주는 ProcessorFunc 타입
type ProcessorFunc func(context.Context, payloads.Payload) (payloads.Payload, error)

// ProcessorFunc 를 Processor 인터페이스를 구현하도록 도와주는 Process 메소드
func (f ProcessorFunc) Process(ctx context.Context, p payloads.Payload) (payloads.Payload, error) {
	return f(ctx, p)
}
