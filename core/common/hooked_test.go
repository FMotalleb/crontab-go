package common

import (
	"context"
	"errors"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/FMotalleb/crontab-go/abstraction"
)

type mockExecutable struct {
	Cancelable
	Timeout
	Hooked

	err error
}

func (m *mockExecutable) Execute(ctx context.Context) error {
	return m.err
}

var exe abstraction.Executable = &mockExecutable{}

func TestSetDoneHooks(t *testing.T) {
	h := &Hooked{}
	hooks := []abstraction.Executable{}
	h.SetDoneHooks(hooks)
	assert.Equal(t, hooks, h.doneHooks)
}

func TestSetFailHooks(t *testing.T) {
	h := &Hooked{}
	failHooks := []abstraction.Executable{&mockExecutable{}}
	h.SetFailHooks(failHooks)
	assert.Equal(t, failHooks, h.failHooks)
}

func TestDoDoneHooks_NoErrors(t *testing.T) {
	h := &Hooked{
		doneHooks: []abstraction.Executable{&mockExecutable{}},
	}
	errs := h.DoDoneHooks(context.Background())
	assert.Zero(t, errs)
}

func TestDoFailHooks_WithErrors(t *testing.T) {
	h := &Hooked{
		failHooks: []abstraction.Executable{&mockExecutable{err: errors.New("fail")}},
	}
	errs := h.DoFailHooks(context.Background())
	assert.Equal(t, len(errs), 1)
	assert.EqualError(t, errs[0], "fail")
}

func TestExecuteTasks_MultipleErrors(t *testing.T) {
	tasks := []abstraction.Executable{
		&mockExecutable{err: errors.New("error1")},
		&mockExecutable{err: errors.New("error2")},
	}
	errs := executeTasks(context.Background(), tasks)
	assert.Equal(t, len(errs), 2)
	assert.EqualError(t, errs[0], "error1")
	assert.EqualError(t, errs[1], "error2")
}
