package global

import (
	"context"
	"sync"

	"github.com/FMotalleb/crontab-go/ctxutils"
)

var CTX = newGlobalContext()

type (
	EventListenerMap = map[string][]func()
	GlobalContext    struct {
		context.Context
		lock sync.Locker
	}
)

func newGlobalContext() *GlobalContext {
	ctx := &GlobalContext{
		context.WithValue(
			context.Background(),
			ctxutils.EventListeners,
			EventListenerMap{},
		),
		&sync.Mutex{},
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
