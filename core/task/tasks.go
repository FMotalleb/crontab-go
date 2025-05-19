package task

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/generator"
)

var tg = generator.New[*config.Task, abstraction.Executable]()

func Build(ctx context.Context, log *logrus.Entry, cfg *config.Task) abstraction.Executable {
	exe := tg.Get(log, cfg)
	if exe == nil {
		log.Panic("did not received any executable action from given task", cfg)
	}
	onDone := []abstraction.Executable{}
	for _, d := range cfg.OnDone {
		onDone = append(onDone, Build(ctx, log, &d))
	}
	exe.SetDoneHooks(ctx, onDone)
	onFail := []abstraction.Executable{}
	for _, d := range cfg.OnFail {
		onFail = append(onFail, Build(ctx, log, &d))
	}
	exe.SetFailHooks(ctx, onFail)
	return exe
}
