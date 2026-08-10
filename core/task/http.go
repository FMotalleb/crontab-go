package task

import (
	"net/http"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/helpers"
)

func doHTTP(client *http.Client, req *http.Request, headers *map[string]string, log *zap.Logger) error {
	for key, val := range *headers {
		req.Header.Add(key, val)
	}
	log.Debug("sending http request")
	res, err := client.Do(req)
	if res != nil {
		if res.Body != nil {
			defer helpers.WarnOnErrIgnored(
				log,
				res.Body.Close,
				"cannot close response body",
			)
		}
		log = log.With(zap.Int("status", res.StatusCode))
		log.Info("received response with status", zap.String("status", res.Status))
		if log.Level() >= zap.DebugLevel {
			ans, respErr := logHTTPResponse(res)
			log.Debug("fetched data", zap.String("response", ans), zap.Error(respErr))
		}
	}
	if err != nil || (res != nil && res.StatusCode >= 400) {
		log.Warn("request failed", zap.Error(err))
		return err
	}
	return nil
}
