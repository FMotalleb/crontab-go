package task

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
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

// newExecutionID returns a random identifier unique to each task execution.
func newExecutionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// executionPrefix builds the per-line prefix for a task's output stream.
func executionPrefix(id string) string {
	return id + " | "
}
