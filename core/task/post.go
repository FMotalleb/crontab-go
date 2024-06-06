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
}

// Cancel implements abstraction.Executable.
func (g *Post) Cancel() {
	if g.cancel != nil {
		g.log.Debugln("canceling get request")
		g.cancel()
	}
}

// Execute implements abstraction.Executable.
func (g *Post) Execute(ctx context.Context) (e error) {
	r := getRetry(ctx)
	log := g.log.WithField("retry", r)
	if getRetry(ctx) > g.retries {
		log.Warn("maximum retry reached")
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
	data, _ := json.Marshal(g.data)

	req, e := http.NewRequestWithContext(localCtx, "POST", g.address, bytes.NewReader(data))
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
