// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

type EventGenerator interface {
	BuildTickChannel() EventChannel
}

type (
	Event        = []string
	EventChannel = <-chan Event
)
