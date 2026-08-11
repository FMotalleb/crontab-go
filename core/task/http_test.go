package task

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"
)

func TestNewHTTPClient_DefaultVerifiesTLS(t *testing.T) {
	c := newHTTPClient(false)
	tr, ok := c.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.True(t, tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify)
}

func TestNewHTTPClient_InsecureSkipsVerify(t *testing.T) {
	c := newHTTPClient(true)
	tr, ok := c.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.True(t, tr.TLSClientConfig != nil)
	assert.True(t, tr.TLSClientConfig.InsecureSkipVerify)
}

func TestDoHTTP_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "oops", http.StatusInternalServerError)
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	assert.NoError(t, err)
	headers := map[string]string{}
	err = doHTTP(srv.Client(), req, &headers, io.Discard, zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

type failingBodyReader struct{}

func (failingBodyReader) Read([]byte) (int, error) { return 0, errors.New("stream interrupted") }

type failingBodyTransport struct{}

func (failingBodyTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(failingBodyReader{}),
	}, nil
}

func TestDoHTTP_FailingResponseBody(t *testing.T) {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.test/", nil)
	assert.NoError(t, err)
	headers := map[string]string{}
	client := &http.Client{Transport: failingBodyTransport{}}
	err = doHTTP(client, req, &headers, io.Discard, zap.NewNop())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "stream interrupted")
}
