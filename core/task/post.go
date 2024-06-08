package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

type Post struct {
	address string
	headers *map[string]string
	data    *any
	log     *logrus.Entry
	cancel  context.CancelFunc

	retries    uint
	retryDelay time.Duration
	timeout    time.Duration

	doneHooks []abstraction.Executable
	failHooks []abstraction.Executable
}

// SetDoneHooks implements abstraction.Executable.
func (p *Post) SetDoneHooks(done []abstraction.Executable) {
	p.doneHooks = done
}

// SetFailHooks implements abstraction.Executable.
func (p *Post) SetFailHooks(fail []abstraction.Executable) {
	p.failHooks = fail
}

// Cancel implements abstraction.Executable.
func (p *Post) Cancel() {
	if p.cancel != nil {
		p.log.Debugln("canceling get request")
		p.cancel()
	}
}

// Execute implements abstraction.Executable.
func (p *Post) Execute(ctx context.Context) (e error) {
	r := getRetry(ctx)
	log := p.log.WithField("retry", r)
	if getRetry(ctx) > p.retries {
		log.Warn("maximum retry reached")
		runTasks(p.failHooks)
		return fmt.Errorf("maximum retries reached")
	}
	if r != 0 {
		log.Debugln("waiting", p.retryDelay, "before executing the next iteration after last fail")
		time.Sleep(p.retryDelay)
	}
	ctx = increaseRetry(ctx)
	// ctx := context.Background()
	var localCtx context.Context
	if p.timeout != 0 {
		localCtx, p.cancel = context.WithTimeout(ctx, p.timeout)
	} else {
		localCtx, p.cancel = context.WithCancel(ctx)
	}
	client := &http.Client{}
	data, _ := json.Marshal(p.data)

	req, e := http.NewRequestWithContext(localCtx, "POST", p.address, bytes.NewReader(data))
	log.Debugln("sending get http request")
	if e != nil {
		return
	}
	for key, val := range *p.headers {
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
		return p.Execute(ctx)
	}

	runTasks(p.doneHooks)
	return
}

func NewPost(task *config.Task, logger *logrus.Entry) abstraction.Executable {
	return &Post{
		address:    task.Post,
		headers:    &task.Headers,
		data:       &task.Data,
		retries:    task.Retries,
		retryDelay: task.RetryDelay,
		timeout:    task.Timeout,
		log: logger.WithFields(
			logrus.Fields{
				"url":    task.Post,
				"method": "post",
			},
		),
	}
}
