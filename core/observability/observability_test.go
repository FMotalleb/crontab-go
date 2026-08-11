package observability

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/fmotalleb/crontab-go/config"
)

func TestParseEndpoint_SchemeSelectsTransportAndTLS(t *testing.T) {
	cases := []struct {
		name      string
		url       string
		transport transportKind
		endpoint  string
		path      string
		tls       bool
	}{
		{name: "http plaintext", url: "http://collector:4318", transport: transportHTTP, endpoint: "collector:4318", path: "", tls: false},
		{name: "https tls", url: "https://collector:4318/v1/traces", transport: transportHTTP, endpoint: "collector:4318", path: "/v1/traces", tls: true},
		{name: "grpc plaintext", url: "grpc://collector:4317", transport: transportGRPC, endpoint: "collector:4317", path: "", tls: false},
		{name: "grpcs tls", url: "grpcs://collector:4317", transport: transportGRPC, endpoint: "collector:4317", path: "", tls: true},
		{name: "bare host defaults to http plaintext", url: "collector:4318", transport: transportHTTP, endpoint: "collector:4318", path: "", tls: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ep, err := parseEndpoint(&config.ObservabilitySignal{URL: tc.url})
			assert.NoError(t, err)
			assert.Equal(t, tc.transport, ep.transport)
			assert.Equal(t, tc.endpoint, ep.endpoint)
			assert.Equal(t, tc.path, ep.path)
			assert.Equal(t, tc.tls, ep.tls)
		})
	}
}

func TestParseEndpoint_UnsupportedScheme(t *testing.T) {
	_, err := parseEndpoint(&config.ObservabilitySignal{URL: "udp://collector:4318"})
	assert.Error(t, err)
}

func TestParseEndpoint_InsecureSkipsVerify(t *testing.T) {
	ep, err := parseEndpoint(&config.ObservabilitySignal{URL: "https://collector:4318", Insecure: true})
	assert.NoError(t, err)
	assert.True(t, ep.tls)
	assert.True(t, ep.skipVerify)
}
