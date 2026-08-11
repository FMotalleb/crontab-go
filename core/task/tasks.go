package task

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/abstraction"
	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/generator"
)

var tg = generator.New[*config.Task, abstraction.Executable]()

func Build(ctx context.Context, log *zap.Logger, cfg config.Task) (abstraction.Executable, error) {
	exe, ok := tg.Get(log, &cfg)
	if !ok {
		return nil, errors.New("no executable action matched for task")
	}
	onDone := []abstraction.Executable{}
	for _, d := range cfg.OnDone {
		h, err := Build(ctx, log, d)
		if err != nil {
			return nil, err
		}
		onDone = append(onDone, h)
	}
	exe.SetDoneHooks(ctx, onDone)
	onFail := []abstraction.Executable{}
	for _, d := range cfg.OnFail {
		h, err := Build(ctx, log, d)
		if err != nil {
			return nil, err
		}
		onFail = append(onFail, h)
	}
	exe.SetFailHooks(ctx, onFail)
	return exe, nil
}
