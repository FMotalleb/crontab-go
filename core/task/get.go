package task

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Get struct {
	address string
	headers *map[string]string
	logger  *logrus.Entry
	cancel  context.CancelFunc

	retries    uint
	retryDelay time.Duration
	timeout    time.Duration
}

// Cancel implements abstraction.Executable.
func (g *Get) Cancel() {
	if g.cancel != nil {
		g.logger.Debugln("canceling get request")
		g.cancel()
	}
}

// Execute implements abstraction.Executable.
func (g *Get) Execute(ctx context.Context) (e error) {
	g.Cancel()
	// ctx := context.Background()
	ctx, g.cancel = context.WithCancel(ctx)
	client := &http.Client{}

	req, e := http.NewRequestWithContext(ctx, "GET", g.address, nil)
	g.logger.Debugln("sending get http request")
	if e != nil {
		return
	}
	for key, val := range *g.headers {
		req.Header.Add(key, val)
	}
	res, e := client.Do(req)
	g.logger.
		WithField("status", res.StatusCode).
		Infoln("received answer", res.StatusCode, res.Body)
	return
}

// func NewGet(address string, headers *map[string]string, logger logrus.Entry) abstraction.Executable {
// 	return &Get{
// 		address,
// 		headers,
// 		logger.WithFields(
// 			logrus.Fields{
// 				"url":    address,
// 				"method": "get",
// 			},
// 		),
// 		nil,
// 	}
// }
