package task

import (
	"bytes"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"
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

func TestNewExecutionID_IsUniqueHex(t *testing.T) {
	a := newExecutionID()
	b := newExecutionID()
	assert.NotEqual(t, a, b)
	assert.Equal(t, 16, len(a))
}

func TestExecutionPrefix(t *testing.T) {
	assert.Equal(t, "deadbeef | ", executionPrefix("deadbeef"))
}
