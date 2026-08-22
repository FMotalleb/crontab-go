package event

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/maniartech/signals"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/fmotalleb/crontab-go/abstraction"
)

func setupTracerProvider(t *testing.T) *tracetest.InMemoryExporter {
	t.Helper()
	prev := otel.GetTracerProvider()
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		otel.SetTracerProvider(prev)
	})
	return exporter
}

func spanByName(spans tracetest.SpanStubs, name string) *tracetest.SpanStub {
	for i := range spans {
		if spans[i].Name == name {
			return &spans[i]
		}
	}
	return nil
}

func TestEmitterName(t *testing.T) {
	assert.Equal(t, "interval", emitterName(NewMetaData("interval", map[string]any{})))
	assert.Equal(t, "unknown", emitterName(NewMetaData("", map[string]any{})))
}

func TestSpanDispatcher_TaskIsSubSpanOfEvent(t *testing.T) {
	exporter := setupTracerProvider(t)

	base := signals.NewSync[abstraction.Event]()
	base.AddListener(func(ctx context.Context, _ abstraction.Event) {
		_, span := otel.Tracer("job").Start(ctx, "job.echo/task.execute")
		span.End()
	})

	sd := NewSpanDispatcher(base, attribute.KeyValue{})
	sd.Emit(t.Context(), NewMetaData("interval", map[string]any{}))

	spans := exporter.GetSpans()
	eventSpan := spanByName(spans, "event.interval")
	taskSpan := spanByName(spans, "job.echo/task.execute")
	assert.NotEqual(t, nil, eventSpan)
	assert.NotEqual(t, nil, taskSpan, "task span must be created from the event context")
	assert.Equal(t, eventSpan.SpanContext.SpanID(), taskSpan.Parent.SpanID())
}

func TestSpanDispatcher_DebounceAttribute(t *testing.T) {
	exporter := setupTracerProvider(t)

	base := signals.NewSync[abstraction.Event]()
	base.AddListener(func(ctx context.Context, _ abstraction.Event) {})

	sd := NewSpanDispatcher(base, attribute.String("event.debounce", "5s"))
	sd.Emit(t.Context(), NewMetaData("cron", map[string]any{"schedule": "* * * * *"}))

	spans := exporter.GetSpans()
	eventSpan := spanByName(spans, "event.cron")
	assert.NotEqual(t, nil, eventSpan)
	found := false
	for _, a := range eventSpan.Attributes {
		if a.Key == "event.debounce" && a.Value.AsString() == "5s" {
			found = true
			break
		}
	}
	assert.True(t, found, "event.debounce attribute must be present")
}

func TestSpanDispatcher_NoDebounceAttributeWhenZero(t *testing.T) {
	exporter := setupTracerProvider(t)

	base := signals.NewSync[abstraction.Event]()
	base.AddListener(func(ctx context.Context, _ abstraction.Event) {})

	sd := NewSpanDispatcher(base, attribute.KeyValue{})
	sd.Emit(t.Context(), NewMetaData("cron", map[string]any{}))

	spans := exporter.GetSpans()
	eventSpan := spanByName(spans, "event.cron")
	assert.NotEqual(t, nil, eventSpan)
	// When debounce is not set, the attribute should not be present.
	for _, a := range eventSpan.Attributes {
		assert.NotEqual(t, "event.debounce", string(a.Key))
	}
}
