package task

import (
	"crypto/tls"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/helpers"
)

// newHTTPClient builds an http.Client, optionally skipping TLS certificate verification.
func newHTTPClient(insecure bool) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // explicit user opt-in
	}
	return &http.Client{Transport: transport}
}

// doHTTP sends the request, streams the response body to out, and returns any error.
func doHTTP(client *http.Client, req *http.Request, headers *map[string]string, out io.Writer, log *zap.Logger) error {
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
		log.Info("received response with status", zap.Int("status", res.StatusCode))
		if out != nil {
			if _, cpErr := io.Copy(out, res.Body); cpErr != nil {
				log.Warn("failed to stream response body", zap.Error(cpErr))
			}
		}
	}
	if err != nil || (res != nil && res.StatusCode >= 400) {
		log.Warn("request failed", zap.Error(err), zap.Int("status", statusCodeOf(res)))
		return err
	}
	return nil
}

func statusCodeOf(res *http.Response) int {
	if res == nil {
		return 0
	}
	return res.StatusCode
}
