package task

import (
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
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
