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
	return ctx
}

var ctx = newGlobalContext()

type (
	EventListenerMap = map[string][]func()
	GlobalContext    struct {
		context.Context
		lock          sync.Locker
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
		lock:          &sync.Mutex{},
		countersValue: make(map[string]*concurrency.LockedValue[float64]),
		counters:      make(map[string]prometheus.CounterFunc),
	}
	return ctx
}

func (ctx *GlobalContext) EventListeners() EventListenerMap {
	listeners := ctx.Value(ctxutils.EventListeners)
	return listeners.(EventListenerMap)
}

func (ctx *GlobalContext) AddEventListener(event string, listener func()) {
	ctx.lock.Lock()
	defer ctx.lock.Unlock()
	listeners := ctx.EventListeners()
	listeners[event] = append(listeners[event], listener)
	ctx.Context = context.WithValue(ctx.Context, ctxutils.EventListeners, listeners)
}
