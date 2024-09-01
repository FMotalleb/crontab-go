package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/common"
	"github.com/FMotalleb/crontab-go/helpers"
)

type Post struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout

	address string
	headers *map[string]string
	data    *any
	log     *logrus.Entry
}

// Execute implements abstraction.Executable.
func (p *Post) Execute(ctx context.Context) (e error) {
	r := common.GetRetry(ctx)
	log := p.log.WithField("retry", r)
	defer func() {
		err := recover()
		if err != nil {
			if err, ok := err.(error); ok {
				log = log.WithError(err)
				e = err
			}
			log.Warnf("recovering command execution from a fatal error: %s", err)
		}
	}()
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
	var dataReader *bytes.Reader
	if p.data != nil {
		data, err := json.Marshal(p.data)
		if err != nil {
			log.
				WithError(err).
				Warnln("cannot marshal the given body (pre-send)")
			return p.Execute(ctx)
		}
		dataReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(localCtx, "POST", p.address, dataReader)
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
		if res.Body != nil {
			defer helpers.WarnOnErrIgnored
				log,
				res.Body.Close,
				"cannot close response body: %s",
			)
		}
		log = log.WithField("status", res.StatusCode)
		log.Infoln("received response with status: ", res.Status)
		if log.Logger.IsLevelEnabled(logrus.DebugLevel) {
			logData := logHTTPResponse(res)
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
	post.SetMetaName(fmt.Sprintf("post: %s", task.Post))
	return post
}
