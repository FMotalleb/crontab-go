package global

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/fmotalleb/go-tools/concurrency"

	"github.com/fmotalleb/crontab-go/abstraction"
)

const (
	OKMetricName  = "done_tasks"
	OKMetricHelp  = "Amount of done tasks (with ok status)"
	ErrMetricName = "failed_tasks"
	ErrMetricHelp = "Amount of failed tasks"

	namespace = "crontab_go"
)

type Metrics = map[string]*prometheus.CounterVec

var collectors = concurrency.NewLockedValue(make(Metrics, 0))

func IncMetric(name string, help string, labels prometheus.Labels) {
	collectors.Operate(
		func(m Metrics) Metrics {
			if vec, ok := m[name]; ok {
				if olderVec, err := vec.GetMetricWith(labels); err == nil {
					olderVec.Inc()
				} else {
					vec.With(labels).Inc()
				}
				otelInc(namespace+"."+name, labels)
				return m
			}
			keys := make([]string, 0, len(labels))
			for key := range labels {
				keys = append(keys, key)
			}
			vec := promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      name,
					Help:      help,
				},
				keys,
			)
			counter := vec.With(labels)
			counter.Inc()
			m[name] = vec
			otelInc(namespace+"."+name, labels)
			return m
		})
}

func RegisterCounter(name string, help string, labels prometheus.Labels) {
	collectors.Operate(
		func(m Metrics) Metrics {
			if vec, ok := m[name]; ok {
				if _, err := vec.GetMetricWith(labels); err != nil {
					vec.With(labels).Add(0)
				}
				return m
			}
			keys := make([]string, 0, len(labels))
			for key := range labels {
				keys = append(keys, key)
			}
			vec := promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: namespace,
					Name:      name,
					Help:      help,
				},
				keys,
			)
			counter := vec.With(labels)
			counter.Add(0)
			m[name] = vec
			return m
		})
}

type otelCounterEntry struct {
	counter metric.Int64Counter
}

var otelCounters = concurrency.NewLockedValue(make(map[string]*otelCounterEntry, 0))

func otelInc(name string, labels prometheus.Labels) {
	otelCounters.Operate(
		func(m map[string]*otelCounterEntry) map[string]*otelCounterEntry {
			if c, ok := m[name]; ok {
				c.counter.Add(context.Background(), 1, metric.WithAttributes(labelsToAttrs(labels)...))
				return m
			}
			meter := otel.GetMeterProvider().Meter("crontab-go")
			counter, err := meter.Int64Counter(name)
			if err != nil {
				return m
			}
			counter.Add(context.Background(), 1, metric.WithAttributes(labelsToAttrs(labels)...))
			m[name] = &otelCounterEntry{counter: counter}
			return m
		})
}

func labelsToAttrs(labels prometheus.Labels) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(labels))
	for k, v := range labels {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}

func CountSignals(signal abstraction.EventDispatcher, name string, help string, labels prometheus.Labels) {
	signal.AddListener(func(_ context.Context, _ abstraction.Event) {
		IncMetric(name, help, labels)
	})
}
