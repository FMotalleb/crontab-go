package task

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/common"
)

type Post struct {
	*common.Hooked
	*common.Cancelable
	*common.Retry
	*common.Timeout

	address string
	headers *map[string]string
	data    *any
	log     *logrus.Entry
}

// Execute implements abstraction.Executable.
func (p *Post) Execute(ctx context.Context) error {
	r := common.GetRetry(ctx)
	log := p.log.WithField("retry", r)
	err := p.WaitForRetry(ctx)
	if err != nil {
		p.DoFailHooks(ctx)
		return err
	}
	ctx = common.IncreaseRetry(ctx)

	var localCtx context.Context
	var cancel context.CancelFunc
	localCtx, cancel = p.ApplyTimeout(ctx)
	p.SetCancel(cancel)

	client := &http.Client{}
	data, err := json.Marshal(p.data)
	if err != nil {
		log.
			WithError(err).
			Warnln("cannot marshal the given body (pre-send)")
		return p.Execute(ctx)
	}

	req, err := http.NewRequestWithContext(localCtx, "POST", p.address, bytes.NewReader(data))
	log.Debugln("sending get http request")
	if err != nil {
		log.
			WithError(err).
			Warnln("cannot create the request (pre-send)")
		return p.Execute(ctx)
	}

	for key, val := range *p.headers {
		req.Header.Add(key, val)
	}

	res, err := client.Do(req)
	if res != nil {
		log = log.WithField("status", res.StatusCode)
		log.Infoln("received response with status: ", res.Status)
		if log.Logger.IsLevelEnabled(logrus.DebugLevel) {
			logData := LogHTTPResponse(res)
			log.Debugln(logData()...)
		}
	}

	if err != nil || res.StatusCode >= 400 {
		log.
			WithError(err).
			Warnln("request failed")
		return p.Execute(ctx)
	}

	p.DoDoneHooks(ctx)
	return nil
}

func NewPost(task *config.Task, logger *logrus.Entry) abstraction.Executable {
	post := &Post{
		address: task.Post,
		headers: &task.Headers,
		data:    &task.Data,
		log: logger.WithFields(
			logrus.Fields{
				"url":    task.Post,
				"method": "post",
			},
		),
	}
	post.SetMaxRetry(task.Retries)
	post.SetRetryDelay(task.RetryDelay)
	post.SetTimeout(task.Timeout)
	return post
}
