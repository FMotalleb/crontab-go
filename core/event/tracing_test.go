package event

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/maniartech/signals"
	"go.opentelemetry.io/otel"
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

func TestEmitWithSpan_TaskIsSubSpanOfEvent(t *testing.T) {
	exporter := setupTracerProvider(t)

	dispatcher := signals.NewSync[abstraction.Event]()
	dispatcher.AddListener(func(ctx context.Context, _ abstraction.Event) {
		_, span := otel.Tracer("job").Start(ctx, "job.echo/task.execute")
		span.End()
	})

	emitWithSpan(dispatcher, context.Background(), NewMetaData("interval", map[string]any{}))

	spans := exporter.GetSpans()
	eventSpan := spanByName(spans, "event.interval")
	taskSpan := spanByName(spans, "job.echo/task.execute")
	assert.NotEqual(t, nil, eventSpan)
	assert.NotEqual(t, nil, taskSpan, "task span must be created from the event context")
	assert.Equal(t, eventSpan.SpanContext.SpanID(), taskSpan.Parent.SpanID())
}
