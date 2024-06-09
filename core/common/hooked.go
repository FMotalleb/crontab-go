package common

import (
	"context"

	"github.com/FMotalleb/crontab-go/abstraction"
)

type Hooked struct {
	doneHooks []abstraction.Executable
	failHooks []abstraction.Executable
}

func (h *Hooked) SetDoneHooks(hooks []abstraction.Executable) {
	h.doneHooks = hooks
}

func (h *Hooked) SetFailHooks(failHooks []abstraction.Executable) {
	h.failHooks = failHooks
}

func (h *Hooked) DoDoneHooks(ctx context.Context) {
	for _, exe := range h.doneHooks {
		exe.Execute(ctx)
	}
}

func (h *Hooked) DoFailHooks(ctx context.Context) {
	for _, exe := range h.failHooks {
		exe.Execute(ctx)
	}
}
