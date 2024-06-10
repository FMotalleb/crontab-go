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

func (h *Hooked) DoDoneHooks(ctx context.Context) []error {
	return executeTasks(ctx, h.doneHooks)
}

func (h *Hooked) DoFailHooks(ctx context.Context) []error {
	return executeTasks(ctx, h.failHooks)
}

func executeTasks(ctx context.Context, tasks []abstraction.Executable) []error {
	errs := []error{}
	for _, exe := range tasks {
		if err := exe.Execute(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
