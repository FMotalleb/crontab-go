package event

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/maniartech/signals"

	"github.com/fmotalleb/crontab-go/abstraction"
)

// SpanDispatcher wraps a base Signal[Event] and creates an
// event.<emitter> span every time an event is actually dispatched
// (i.e. when Emit reaches the underlying signal).  Because the
// debouncer sits above this layer, the span is only created for
// events that are really sent to listeners — never for events that
// are absorbed by the debounce window.
type SpanDispatcher struct {
	next     signals.Signal[abstraction.Event]
	debounce attribute.KeyValue // optional; zero value means omitted
}

// NewSpanDispatcher returns a dispatcher that wraps sig and records
// debounce as a span attribute when debounce > 0.
func NewSpanDispatcher(sig signals.Signal[abstraction.Event], debounce attribute.KeyValue) *SpanDispatcher {
	return &SpanDispatcher{
		next:     sig,
		debounce: debounce,
	}
}

// Emit creates an event span and dispatches to the underlying signal.
// The tracer is resolved from the current global TracerProvider on
// every call so that tests and runtime provider swaps are honoured.
func (s *SpanDispatcher) Emit(ctx context.Context, payload abstraction.Event) {
	emitter := emitterName(payload)
	attrs := []attribute.KeyValue{
		attribute.String("event.emitter", emitter),
	}
	attrs = append(attrs, eventParams(payload)...)
	if s.debounce.Key != "" {
		attrs = append(attrs, s.debounce)
	}
	tracer := otel.GetTracerProvider().Tracer("crontab-go/event")
	ctx, span := tracer.Start(ctx, "event."+emitter,
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(attrs...),
	)
	defer span.End()
	s.next.Emit(ctx, payload)
	span.SetStatus(codes.Ok, "")
}

// AddListener delegates to the underlying signal.
func (s *SpanDispatcher) AddListener(handler signals.SignalListener[abstraction.Event], key ...string) int {
	return s.next.AddListener(handler, key...)
}

// RemoveListener delegates to the underlying signal.
func (s *SpanDispatcher) RemoveListener(key string) int {
	return s.next.RemoveListener(key)
}

// IsEmpty delegates to the underlying signal.
func (s *SpanDispatcher) IsEmpty() bool {
	return s.next.IsEmpty()
}

// Len delegates to the underlying signal.
func (s *SpanDispatcher) Len() int {
	return s.next.Len()
}

// Reset delegates to the underlying signal.
func (s *SpanDispatcher) Reset() {
	s.next.Reset()
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
