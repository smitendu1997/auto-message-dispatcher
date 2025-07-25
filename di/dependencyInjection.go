package di

import (
	"reflect"
	"sync"

	"github.com/smitendu1997/auto-message-dispatcher/logger"
)

// Container provides dependency injection capabilities
type Container struct {
	services  sync.Map
	factories sync.Map
	resolving sync.Map
}

// Factory defines a factory function for creating service instances
type Factory func(*Container) interface{}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{}
}

// Register adds a concrete implementation of an interface
func (c *Container) Register(interfaceType interface{}, implementation interface{}) {
	t := reflect.TypeOf(interfaceType).Elem()
	c.services.Store(t, implementation)
	logger.Info("DI: Registered", t.String())
}

// RegisterFactory adds a factory function for creating service instances
func (c *Container) RegisterFactory(interfaceType interface{}, factory Factory) {
	t := reflect.TypeOf(interfaceType).Elem()
	c.factories.Store(t, factory)
	logger.Info("DI: Registered factory for", t.String())
}

// Resolve returns the implementation for a given interface
func (c *Container) Resolve(interfaceType interface{}) interface{} {
	t := reflect.TypeOf(interfaceType).Elem()

	// Check if already resolved
	if service, ok := c.services.Load(t); ok {
		return service
	}

	// Check for circular dependencies
	if _, resolving := c.resolving.LoadOrStore(t, true); resolving {
		logger.Error("DI: Circular dependency detected for", t.String())
		return nil
	}
	defer c.resolving.Delete(t)

	// Create new instance if we have a factory
	if factoryVal, ok := c.factories.Load(t); ok {
		factory := factoryVal.(Factory)
		result := factory(c)
		if result != nil {
			c.services.Store(t, result)
		}
		return result
	}

	logger.Error("DI: Failed to resolve", t.String())
	return nil
}

// IsRegistered checks if an implementation or factory exists for an interface
func (c *Container) IsRegistered(interfaceType interface{}) bool {
	t := reflect.TypeOf(interfaceType).Elem()

	// Check if we have a concrete implementation
	if _, exists := c.services.Load(t); exists {
		return true
	}

	// Check if we have a factory
	if _, exists := c.factories.Load(t); exists {
		return true
	}

	return false
}
