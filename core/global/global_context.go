// Package global contains global state management logics
package global

import (
	"context"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FMotalleb/crontab-go/core/concurrency"
	"github.com/FMotalleb/crontab-go/ctxutils"
)

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
	listeners := c.EventListeners()
	listeners[event] = append(listeners[event], listener)
	c.Context = context.WithValue(c.Context, ctxutils.EventListeners, listeners)
}
