// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

type EventGenerator interface {
	BuildTickChannel() EventChannel
}

type (
	EventChannel     = <-chan Event
	EventEmitChannel = chan Event
)

type Event interface {
	GetData() map[string]any
}
