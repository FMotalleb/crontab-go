package task

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"

	"github.com/fmotalleb/crontab-go/ctxutils"
)

// prefixWriter prefixes every new line with a fixed prefix before forwarding to dst.
// It streams line-by-line without buffering the whole output in memory.
type prefixWriter struct {
	mu        sync.Mutex
	dst       io.Writer
	prefix    []byte
	lineStart bool
}

// NewPrefixWriter returns an io.Writer that prefixes every line with exactly the given prefix.
func NewPrefixWriter(dst io.Writer, prefix string) io.Writer {
	return &prefixWriter{
		dst:       dst,
		prefix:    []byte(prefix),
		lineStart: true,
	}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	start := 0
	for i, b := range p {
		if b == '\n' {
			if err := w.writeChunk(p[start : i+1]); err != nil {
				return i + 1, err
			}
			start = i + 1
		}
	}
	if start < len(p) {
		if err := w.writeChunk(p[start:]); err != nil {
			return start, err
		}
	}
	return len(p), nil
}

func (w *prefixWriter) writeChunk(chunk []byte) error {
	buf := bytes.NewBuffer(nil)
	if len(chunk) > 0 && w.lineStart {
		buf.Write(w.prefix)
	}
	buf.Write(chunk)
	if len(chunk) > 0 {
		w.lineStart = chunk[len(chunk)-1] == '\n'
	}
	_, err := w.dst.Write(buf.Bytes())
	return err
}

// shortHash returns a compact hex digest of the given string, enough to tell runs apart.
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:4])
}

// outputPrefix builds the stdout/stderr line prefix for a task output.
// It uses the job name from the context and a hash of the task's main parameter
// to mimic docker style logging: "<job-name>:<hash> | <line>".
func outputPrefix(jobName, param string) string {
	return jobName + ":" + shortHash(param) + " | "
}

// jobName returns the job name carried in the context, defaulting to an empty string.
func jobName(ctx context.Context) string {
	name, _ := ctx.Value(ctxutils.JobKey).(string)
	return name
}
