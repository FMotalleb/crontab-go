package event

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/fmotalleb/crontab-go/abstraction"
)

// tracer is the named tracer used for event emitter spans.
var tracer = otel.Tracer("crontab-go/event")

// emitWithSpan wraps the dispatch of an event in an event.<emitter> span.
// The traced context is forwarded to listeners so every downstream task
// becomes a sub span of the emitting event.
func emitWithSpan(ed abstraction.EventDispatcher, ctx context.Context, e abstraction.Event) {
	emitter := emitterName(e)
	ctx, span := tracer.Start(ctx, "event."+emitter,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attribute.String("event.emitter", emitter)),
	)
	defer span.End()
	ed.Emit(ctx, e)
}

// emitterName returns the emitter type recorded in the event metadata.
func emitterName(e abstraction.Event) string {
	name, ok := e.GetData()["emitter"].(string)
	if !ok || name == "" {
		return "unknown"
	}
	return name
}
