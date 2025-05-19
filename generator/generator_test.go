package generator_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/generator"
	mocklogger "github.com/FMotalleb/crontab-go/logger/mock_logger"
)

type result struct {
	Value string
}

func TestCore_Get_WithAllNilGenerators(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	core := generator.New[int, *result]()
	// No generators added

	out := core.Get(log, 42)
	assert.Equal(t, out, (*result)(nil))
}

func TestCore_Get_FirstNonNilReturned(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	core := generator.New[int, *result]()

	core.Register(func(log *logrus.Entry, input int) *result {
		return nil
	})
	core.Register(func(log *logrus.Entry, input int) *result {
		if input > 0 {
			return &result{Value: "success"}
		}
		return nil
	})
	var empty *result
	out := core.Get(log, 1)
	assert.NotEqual(t, out, empty)
	assert.Equal(t, out.Value, "success")
}

func TestCore_Get_NilReturned(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	core := generator.New[int, *result]()

	core.Register(func(log *logrus.Entry, input int) *result {
		return nil
	})

	out := core.Get(log, 1)
	assert.Equal(t, out, nil)
}

func TestCore_Get_PanicIfNilReceiver(t *testing.T) {
	logger, _ := mocklogger.HijackOutput(logrus.New())
	log := logrus.NewEntry(logger)

	var core *generator.Core[int, *result] = nil

	assert.Panics(t, func() {
		_ = core.Get(log, 99)
	})
}
