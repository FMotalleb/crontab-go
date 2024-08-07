package event

import "github.com/FMotalleb/crontab-go/core/global"

type WebEventListener struct {
	c chan any

	event string
}

func NewEventListener(event string) WebEventListener {
	return WebEventListener{
		c:     make(chan any),
		event: event,
	}
}

// BuildTickChannel implements abstraction.Scheduler.
func (w *WebEventListener) BuildTickChannel() <-chan any {
	global.CTX().AddEventListener(
		w.event, func() {
			w.c <- false
		},
	)
	return w.c
}
