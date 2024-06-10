// Package abstraction must contain only interfaces and abstract layers of modules
package abstraction

import (
	"context"
)

// Executable is an object that can be executed using a execute method and stopped using cancel method
type Executable interface {
	Execute(context.Context) error
	SetDoneHooks([]Executable)
	SetFailHooks([]Executable)
	Cancel()
}
