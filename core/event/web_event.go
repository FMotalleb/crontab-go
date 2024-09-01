package event

import "github.com/FMotalleb/crontab-go/core/global"

type WebEventListener struct {
	event string
}

func NewEventListener(event string) WebEventListener {
	return WebEventListener{
		event: event,
	}
}

// BuildTickChannel implements abstraction.Scheduler.
func (w *WebEventListener) BuildTickChannel() <-chan []string {
	notifyChan := make(chan []string)
	global.CTX().AddEventListener(
		w.event, func() {
			notifyChan <- []string{"web", w.event}
		},
	)
	return notifyChan
}
