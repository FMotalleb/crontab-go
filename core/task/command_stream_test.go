package task

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
	connection "github.com/fmotalleb/crontab-go/core/cmd_connection"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func TestCommand_Execute_StreamsPrefixedOutput(t *testing.T) {
	ctx := context.WithValue(t.Context(), ctxutils.JobKey, "echo")
	log := zap.NewNop()

	conn, err := connection.Get(&config.TaskConnection{Local: true}, log)
	assert.NoError(t, err)
	assert.NoError(t, conn.Prepare(ctx, &config.Task{Command: "echo hello", Env: map[string]string{"SHELL": "/bin/sh"}}))
	assert.NoError(t, conn.Connect())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	prefix := outputPrefix("echo", "echo hello")
	assert.NoError(t, conn.Execute(NewPrefixWriter(&stdout, prefix), NewPrefixWriter(&stderr, prefix)))
	assert.NoError(t, conn.Disconnect())

	assert.Equal(t, "echo:"+shortHash("echo hello")+" | hello\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestCommand_Execute_StreamsStderrSeparately(t *testing.T) {
	ctx := context.WithValue(t.Context(), ctxutils.JobKey, "echo")
	log := zap.NewNop()

	conn, err := connection.Get(&config.TaskConnection{Local: true}, log)
	assert.NoError(t, err)
	assert.NoError(t, conn.Prepare(ctx, &config.Task{Command: "echo boom >&2", Env: map[string]string{"SHELL": "/bin/sh"}}))
	assert.NoError(t, conn.Connect())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	prefix := outputPrefix("echo", "echo boom >&2")
	err = conn.Execute(NewPrefixWriter(&stdout, prefix), NewPrefixWriter(&stderr, prefix))
	assert.NoError(t, err)
	assert.NoError(t, conn.Disconnect())

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "echo:"+shortHash("echo boom >&2")+" | boom\n", stderr.String())
}

func TestPrefixWriter_WritesThroughExec(t *testing.T) {
	var buf bytes.Buffer
	w := NewPrefixWriter(&buf, "svc:deadbeef | ")
	_, err := io.WriteString(w, "a\nb\n")
	assert.NoError(t, err)
	assert.Equal(t, "svc:deadbeef | a\nsvc:deadbeef | b\n", buf.String())
}
