package task

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

type Get struct {
	address string
	headers *map[string]string
	log     *logrus.Entry
	cancel  context.CancelFunc

	retries    uint
	retryDelay time.Duration
	timeout    time.Duration

	doneHooks []abstraction.Executable
	failHooks []abstraction.Executable
}

// SetDoneHooks implements abstraction.Executable.
func (g *Get) SetDoneHooks(done []abstraction.Executable) {
	g.doneHooks = done
}

// SetFailHooks implements abstraction.Executable.
func (g *Get) SetFailHooks(fail []abstraction.Executable) {
	g.failHooks = fail
}

// Cancel implements abstraction.Executable.
func (g *Get) Cancel() {
	if g.cancel != nil {
		g.log.Debugln("canceling get request")
		g.cancel()
	}
}

// Execute implements abstraction.Executable.
func (g *Get) Execute(ctx context.Context) (e error) {
	r := getRetry(ctx)
	log := g.log.WithField("retry", r)
	if getRetry(ctx) > g.retries {
		log.Warn("maximum retry reached")
		runTasks(g.failHooks)
		return fmt.Errorf("maximum retries reached")
	}
	if r != 0 {
		log.Debugln("waiting", g.retryDelay, "before executing the next iteration after last fail")
		time.Sleep(g.retryDelay)
	}
	ctx = increaseRetry(ctx)
	// ctx := context.Background()
	var localCtx context.Context
	if g.timeout != 0 {
		localCtx, g.cancel = context.WithTimeout(ctx, g.timeout)
	} else {
		localCtx, g.cancel = context.WithCancel(ctx)
	}
	client := &http.Client{}

	req, e := http.NewRequestWithContext(localCtx, "GET", g.address, nil)
	log.Debugln("sending get http request")
	if e != nil {
		return
	}
	for key, val := range *g.headers {
		req.Header.Add(key, val)
	}

	res, e := client.Do(req)
	if res != nil {
		log = log.WithField("status", res.StatusCode)
		log.Infoln("received response with status: ", res.Status)
		if log.Logger.IsLevelEnabled(logrus.DebugLevel) {
			logData := logResponse(res)
			log.Debugln(logData()...)
		}

	}
	if e != nil {
		log = log.WithError(e)
	}

	if e != nil || res.StatusCode >= 400 {
		log.Warnln("request failed")
		return g.Execute(ctx)
	}
	runTasks(g.doneHooks)
	return
}

func NewGet(task *config.Task, logger *logrus.Entry) abstraction.Executable {
	return &Get{
		address:    task.Get,
		headers:    &task.Headers,
		retries:    task.Retries,
		retryDelay: task.RetryDelay,
		timeout:    task.Timeout,
		log: logger.WithFields(
			logrus.Fields{
				"url":    task.Get,
				"method": "get",
			},
		),
	}
}
