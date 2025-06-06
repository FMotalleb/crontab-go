package task

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
	"github.com/FMotalleb/crontab-go/core/common"
	"github.com/FMotalleb/crontab-go/helpers"
)

func init() {
	tg.Register(NewGet)
}

func NewGet(logger *logrus.Entry, task *config.Task) (abstraction.Executable, bool) {
	if task.Get == "" {
		return nil, false
	}
	get := &Get{
		address: task.Get,
		headers: &task.Headers,
		log: logger.WithFields(
			logrus.Fields{
				"url":    task.Get,
				"method": "get",
			},
		),
	}
	get.SetMaxRetry(task.Retries)
	get.SetRetryDelay(task.RetryDelay)
	get.SetTimeout(task.Timeout)
	get.SetMetaName("get: " + task.Get)
	return get, true
}

type Get struct {
	common.Hooked
	common.Cancelable
	common.Retry
	common.Timeout

	address string
	headers *map[string]string
	log     *logrus.Entry
}

// Execute implements abstraction.Executable.
func (g *Get) Execute(ctx context.Context) (e error) {
	r := common.GetRetry(ctx)
	log := g.log.WithField("retry", r)
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

	if err := g.WaitForRetry(ctx); err != nil {
		g.DoFailHooks(ctx)
		return err
	}
	ctx = common.IncreaseRetry(ctx)

	localCtx, cancel := g.ApplyTimeout(ctx)
	g.SetCancel(cancel)

	client := &http.Client{}
	req, err := http.NewRequestWithContext(localCtx, http.MethodGet, g.address, nil)
	log.Debugln("sending get http request")
	if err != nil {
		log.
			WithError(err).
			Warnln("cannot create the request (pre-send)")
		return g.Execute(ctx)
	}
	for key, val := range *g.headers {
		req.Header.Add(key, val)
	}
	res, err := client.Do(req)
	if res != nil {
		if res.Body != nil {
			defer helpers.WarnOnErrIgnored(
				log,
				res.Body.Close,
				"cannot close response body: %s",
			)
		}
		log = log.WithField("status", res.StatusCode)
		log.Infoln("received response with status: ", res.Status)
		if log.Logger.IsLevelEnabled(logrus.DebugLevel) {
			logData := logHTTPResponse(res)
			log.Debugln(
				logData()...,
			)
		}
	}
	if err != nil || res.StatusCode >= 400 {
		log.
			WithError(err).
			Warnln("request failed")

		return g.Execute(ctx)
	}
	g.DoDoneHooks(ctx)
	return nil
}
