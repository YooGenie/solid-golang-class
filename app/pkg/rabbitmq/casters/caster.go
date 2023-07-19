package casters

import (
	"event-data-pipeline/pkg/logger"
	"fmt"
	"strings"

	"github.com/streadway/amqp"
)

const (
	CASTER_SUFFIX = "_caster"
)

type Caster interface {
	// Process operates on the input payload and returns back a new payload
	// to be forwarded to the next pipeline stage. Processors may also opt
	// to prevent the payload from reaching the rest of the pipeline by
	// returning a nil payload value instead.
	Cast(meta jsonObj, message amqp.Delivery) (jsonObj, error)
}

type CasterFunc func(meta jsonObj, message amqp.Delivery) (jsonObj, error)

func (f CasterFunc) Cast(meta jsonObj, message amqp.Delivery) (jsonObj, error) {
	return f(meta, message)
}

type jsonObj = map[string]interface{}

type CasterFactory func() Caster

var casterFactories = make(map[string]CasterFactory)

// Each Caster implementation must Register itself
func Register(name string, factory CasterFactory) {
	logger.Debugf("Registering caster factory for %s", name)
	if factory == nil {
		logger.Panicf("caster factory %s does not exist.", name)
	}
	_, registered := casterFactories[name]
	if registered {
		logger.Errorf("caster factory %s already registered. Ignoring.", name)
	}
	casterFactories[name] = factory
}

// CreateProcessor is a factory method that will create the named processor
func CreateCaster(name string) (Caster, error) {

	factory, ok := casterFactories[name]
	if !ok {
		// Factory has not been registered.
		// Make a list of all available datastore factories for logging.
		availableCasters := make([]string, 0)
		for k := range casterFactories {
			availableCasters = append(availableCasters, k)
		}
		return nil, fmt.Errorf("invalid caster name. Must be one of: %s", strings.Join(availableCasters, ", "))
	}

	// Run the factory with the configuration.
	return factory(), nil
}
