// Package global contains global state management logics
package global

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

func ctxKey(prefix string, key string) ctxutils.ContextKey {
	return ctxutils.ContextKey(fmt.Sprintf("%s:%s", prefix, key))
}

func CTX() *GlobalContext {
	return c
}

var c = newGlobalContext()

type (
	EventListenerMap = map[string][]func()
	GlobalContext    struct {
		context.Context
		lock          *sync.RWMutex
		countersValue map[string]*concurrency.LockedValue[float64]
		counters      map[string]prometheus.CounterFunc
	}
)

func newGlobalContext() *GlobalContext {
	ctx := &GlobalContext{
		Context: context.WithValue(
			context.Background(),
			ctxutils.EventListeners,
			EventListenerMap{},
		),
		lock:          new(sync.RWMutex),
		countersValue: make(map[string]*concurrency.LockedValue[float64]),
		counters:      make(map[string]prometheus.CounterFunc),
	}
	return ctx
}

func (c *GlobalContext) EventListeners() EventListenerMap {
	c.lock.RLock()
	defer c.lock.RUnlock()
	listeners := c.Value(ctxutils.EventListeners)
	return listeners.(EventListenerMap)
}

func (c *GlobalContext) AddEventListener(event string, listener func()) {
	c.lock.Lock()
	defer c.lock.Unlock()
	listeners := c.Value(ctxutils.EventListeners).(EventListenerMap)
	listeners[event] = append(listeners[event], listener)
	c.Context = context.WithValue(c.Context, ctxutils.EventListeners, listeners)
}

func getTypename[T any](item T) string {
	return reflect.TypeOf(item).String()
}

func Put[T any](item T) {
	name := getTypename(item)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Context = context.WithValue(c.Context, ctxKey("typed", name), item)
}

func Get[T any]() T {
	var zero T // Default zero value for type T
	name := reflect.TypeOf(zero).String()
	println(name)
	value := c.Value(ctxKey("typed", name))
	if value == nil {
		return zero
	}

	// Type assertion to ensure the value is of type T
	castedValue, ok := value.(T)
	if !ok {
		return zero
	}
	return castedValue
}
