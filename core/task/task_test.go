package task_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"

	"github.com/fmotalleb/crontab-go/config"
	"github.com/fmotalleb/crontab-go/core/task"
	"github.com/fmotalleb/crontab-go/ctxutils"
)

func TestCompileTask_NonExistingTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{}
	_, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.Error(t, err)
}

func TestCompileTask_GetTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Get: "test",
	}
	exe, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, exe)
}

func TestCompileTask_CommandTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Command: "test",
	}
	exe, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_PostTask(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Post: "test",
	}
	exe, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, exe, nil)
}

func TestCompileTask_WithHooks(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Command: "test",
		OnDone: []config.Task{
			{
				Command: "test",
			},
		},
		OnFail: []config.Task{
			{
				Command: "test",
			},
		},
	}
	exe, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, exe, nil)
}

func TestGetTask_Execute_NoPanic(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Get:        "http://localhost:1/nonexistent",
		Timeout:    100 * time.Millisecond,
		Retries:    1,
		RetryDelay: 10 * time.Millisecond,
	}
	exe, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, exe)
	err = exe.Execute(ctx)
	assert.NotEqual(t, nil, err)
}

func TestPostTask_Execute_NoPanic(t *testing.T) {
	ctx := t.Context()
	ctx = context.WithValue(ctx, ctxutils.JobKey, "test_job")
	taskConfig := config.Task{
		Post:       "http://localhost:1/nonexistent",
		Timeout:    100 * time.Millisecond,
		Retries:    1,
		RetryDelay: 10 * time.Millisecond,
	}
	exe, err := task.Build(ctx, zap.NewNop(), taskConfig)
	assert.NoError(t, err)
	assert.NotEqual(t, nil, exe)
	err = exe.Execute(ctx)
	assert.NotEqual(t, nil, err)
}
