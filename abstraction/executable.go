// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

// Executable is an object that can be executed using a execute method and stopped using cancel method
type Executable interface {
	execute() error
	cancel()
}
