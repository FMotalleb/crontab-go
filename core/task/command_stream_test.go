package task

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
	connection "github.com/fmotalleb/crontab-go/core/cmd_connection"
)

func TestCommand_Execute_StreamsPrefixedOutput(t *testing.T) {
	ctx := t.Context()
	log := zap.NewNop()

	conn, err := connection.Get(&config.TaskConnection{Local: true}, log)
	assert.NoError(t, err)
	assert.NoError(t, conn.Prepare(ctx, &config.Task{Command: "echo hello", Env: map[string]string{"SHELL": "/bin/sh"}}))
	assert.NoError(t, conn.Connect())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	prefix := executionPrefix("run-1234")
	assert.NoError(t, conn.Execute(NewPrefixWriter(&stdout, prefix), NewPrefixWriter(&stderr, prefix)))
	assert.NoError(t, conn.Disconnect())

	assert.Equal(t, "run-1234 | hello\n", stdout.String())
	assert.Equal(t, "", stderr.String())
}

func TestCommand_Execute_StreamsStderrSeparately(t *testing.T) {
	ctx := t.Context()
	log := zap.NewNop()

	conn, err := connection.Get(&config.TaskConnection{Local: true}, log)
	assert.NoError(t, err)
	assert.NoError(t, conn.Prepare(ctx, &config.Task{Command: "echo boom >&2", Env: map[string]string{"SHELL": "/bin/sh"}}))
	assert.NoError(t, conn.Connect())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	prefix := executionPrefix("run-5678")
	err = conn.Execute(NewPrefixWriter(&stdout, prefix), NewPrefixWriter(&stderr, prefix))
	assert.NoError(t, err)
	assert.NoError(t, conn.Disconnect())

	assert.Equal(t, "", stdout.String())
	assert.Equal(t, "run-5678 | boom\n", stderr.String())
}

func TestPrefixWriter_WritesThroughExec(t *testing.T) {
	var buf bytes.Buffer
	w := NewPrefixWriter(&buf, "svc:deadbeef | ")
	_, err := io.WriteString(w, "a\nb\n")
	assert.NoError(t, err)
	assert.Equal(t, "svc:deadbeef | a\nsvc:deadbeef | b\n", buf.String())
}

func TestCommand_Execute_AppliesTimeout(t *testing.T) {
	ctx := t.Context()
	log := zap.NewNop()

	cmd, ok := NewCommand(log, &config.Task{
		Command:    "sleep 5",
		Env:        map[string]string{"SHELL": "/bin/sh"},
		Timeout:    200 * time.Millisecond,
		RetryDelay: time.Second,
	})
	assert.True(t, ok)

	start := time.Now()
	err := cmd.Execute(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.True(t, elapsed < 5*time.Second, "command should have been killed by its timeout, took %s", elapsed)
}
