package event

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	attrs := []attribute.KeyValue{
		attribute.String("event.emitter", emitter),
	}
	attrs = append(attrs, eventParams(e)...)
	ctx, span := tracer.Start(ctx, "event."+emitter,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attrs...),
	)
	defer span.End()
	ed.Emit(ctx, e)
	span.SetStatus(codes.Ok, "")
}

// emitterName returns the emitter type recorded in the event metadata.
func emitterName(e abstraction.Event) string {
	name, ok := e.GetData()["emitter"].(string)
	if !ok || name == "" {
		return "unknown"
	}
	return name
}

// eventParams extracts known parameters from the event data map as span attributes.
func eventParams(e abstraction.Event) []attribute.KeyValue {
	data := e.GetData()
	attrs := make([]attribute.KeyValue, 0, len(data))
	for k, v := range data {
		if k == "emitter" {
			continue
		}
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String("event."+k, val))
		case int:
			attrs = append(attrs, attribute.Int64("event."+k, int64(val)))
		case int64:
			attrs = append(attrs, attribute.Int64("event."+k, val))
		case float64:
			attrs = append(attrs, attribute.Float64("event."+k, val))
		case bool:
			attrs = append(attrs, attribute.Bool("event."+k, val))
		default:
			attrs = append(attrs, attribute.String("event."+k, fmt.Sprintf("%v", val)))
		}
	}
	return attrs
}
