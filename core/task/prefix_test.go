package task

import (
	"bytes"
	"io"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestPrefixWriter_PrefixesEveryLine(t *testing.T) {
	var buf bytes.Buffer
	w := NewPrefixWriter(&buf, "svc")
	_, err := io.WriteString(w, "first\nsecond line\n\nthird")
	assert.NoError(t, err)
	assert.Equal(t, "svc | first\nsvc | second line\nsvc | \nsvc | third\n", buf.String())
}

func TestNewExecutionID_IsUniqueHex(t *testing.T) {
	a := newExecutionID()
	b := newExecutionID()
	assert.NotEqual(t, a, b)
	assert.Equal(t, 16, len(a))
}
