package task

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Post struct {
	address string
	headers *map[string]string
	data    *map[string]any
	logger  *logrus.Entry
	cancel  context.CancelFunc

	retries    uint
	retryDelay time.Duration
	timeout    time.Duration
}

// Cancel implements abstraction.Executable.
func (g *Post) Cancel() {
	if g.cancel != nil {
		g.logger.Debugln("canceling get request")
		g.cancel()
	}
}

// Execute implements abstraction.Executable.
func (g *Post) Execute(ctx context.Context) (e error) {
	g.Cancel()

	ctx, g.cancel = context.WithCancel(ctx)
	client := &http.Client{}
	data, e := json.Marshal(g.data)
	body := bytes.NewReader(data)
	req, e := http.NewRequestWithContext(ctx, "POST", g.address, body)
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

// func NewPost(address string, headers *map[string]string, data *map[string]any, logger logrus.Entry) abstraction.Executable {
// 	return &Post{
// 		address,
// 		headers,
// 		data,
// 		logger.WithFields(
// 			logrus.Fields{
// 				"url":    address,
// 				"method": "post",
// 			},
// 		),
// 		nil,
// 	}
// }
