package common

import (
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestTimeoutSet(t *testing.T) {
	timeout := &Timeout{}
	timeout.SetTimeout(time.Hour)
	assert.Equal(t, timeout.timeout, time.Hour)
}

func TestTimeoutApply(t *testing.T) {
	timeout := &Timeout{}
	now := time.Now()
	timeout.SetTimeout(time.Hour)
	ctx := t.Context()
	ctx, cancel := timeout.ApplyTimeout(ctx)
	deadline, _ := ctx.Deadline()
	assert.Equal(t, time.Hour.Milliseconds(), deadline.UnixMilli()-now.UnixMilli())
	assert.NotEqual(t, nil, cancel)
}

func TestNoTimeoutApply(t *testing.T) {
	timeout := &Timeout{}
	ctx := t.Context()
	ctx, cancel := timeout.ApplyTimeout(ctx)
	_, ok := ctx.Deadline()
	assert.Equal(t, false, ok)
	assert.NotEqual(t, nil, cancel)
}
