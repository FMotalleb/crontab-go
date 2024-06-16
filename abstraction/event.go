// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

// Events is an object that can be executed using a execute method and stopped using cancel method
type Events interface {
	BuildTickChannel() <-chan any
	Cancel()
}
