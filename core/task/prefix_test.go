package task

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/fmotalleb/crontab-go/ctxutils"
)

func TestPrefixWriter_PrefixesEveryLine(t *testing.T) {
	var buf bytes.Buffer
	w := NewPrefixWriter(&buf, "svc | ")
	_, err := io.WriteString(w, "first\nsecond line\n\nthird")
	assert.NoError(t, err)
	assert.Equal(t, "svc | first\nsvc | second line\nsvc | \nsvc | third", buf.String())
}

func TestPrefixWriter_SplitsAcrossWrites(t *testing.T) {
	var buf bytes.Buffer
	w := NewPrefixWriter(&buf, "svc | ")
	for _, chunk := range []string{"hel", "lo\nwo", "rld\n"} {
		_, err := io.WriteString(w, chunk)
		assert.NoError(t, err)
	}
	assert.Equal(t, "svc | hello\nsvc | world\n", buf.String())
}

func TestOutputPrefix_DockerStyle(t *testing.T) {
	got := outputPrefix("service-1", "echo hi")
	assert.Equal(t, "service-1:"+shortHash("echo hi")+" | ", got)
}

func TestJobName_FromContext(t *testing.T) {
	ctx := context.WithValue(t.Context(), ctxutils.JobKey, "my-job")
	assert.Equal(t, "my-job", jobName(ctx))
	assert.Equal(t, "", jobName(context.Background()))
}
