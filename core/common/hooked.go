package common

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/core/global"
)

type Hooked struct {
	metaName  string
	doneHooks []abstraction.Executable
	failHooks []abstraction.Executable
}

func (h *Hooked) SetMetaName(metaName string) {
	h.metaName = metaName
}

func (h *Hooked) SetDoneHooks(ctx context.Context, hooks []abstraction.Executable) {
	global.CTX().MetricCounter(
		ctx,
		"done_tasks",
		"Amount of done tasks (with ok status)",
		prometheus.Labels{"task_type": h.metaName},
	).Operate(
		func(f float64) float64 {
			return f + 1
		},
	)
	h.doneHooks = hooks
}

func (h *Hooked) SetFailHooks(ctx context.Context, failHooks []abstraction.Executable) {
	global.CTX().MetricCounter(
		ctx,
		"done_tasks",
		"Amount of done tasks (with ok status)",
		prometheus.Labels{"task_type": h.metaName},
	).Operate(
		func(f float64) float64 {
			return f + 1
		},
	)
	h.failHooks = failHooks
}

func (h *Hooked) DoDoneHooks(ctx context.Context) []error {
	global.CTX().MetricCounter(
		ctx,
		"failed_tasks",
		"Amount of failed tasks",
		prometheus.Labels{"task_type": h.metaName},
	).Operate(
		func(f float64) float64 {
			return f + 1
		},
	)
	return executeTasks(ctx, h.doneHooks)
}

func (h *Hooked) DoFailHooks(ctx context.Context) []error {
	global.CTX().MetricCounter(
		ctx,
		"failed_tasks",
		"Amount of failed tasks",
		prometheus.Labels{"task_type": h.metaName},
	).Operate(
		func(f float64) float64 {
			return f + 1
		},
	)
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
