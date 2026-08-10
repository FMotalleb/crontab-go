// Package observability provides OpenTelemetry setup for tracing, metrics, and logging.
package observability

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/fmotalleb/crontab-go/config"
)

// ShutdownFunc is called to flush pending telemetry data.
type ShutdownFunc func(context.Context) error

// SetupResult holds the shutdown function and an optional otelzap core for logging bridge.
type SetupResult struct {
	Shutdown ShutdownFunc
	LogCore  zapcore.Core
}

// Setup initializes OpenTelemetry providers from config.
// Returns a shutdown function and optional otelzap core; provider startup is non-fatal (logs warning on failure).
func Setup(ctx context.Context, cfg *config.Observability, logger *zap.Logger) (SetupResult, error) {
	if cfg == nil {
		return SetupResult{Shutdown: noopShutdown}, nil
	}

	svcName := cfg.ServiceName
	if svcName == "" {
		svcName = "crontab-go"
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(svcName),
		),
	)
	if err != nil {
		return SetupResult{Shutdown: noopShutdown}, fmt.Errorf("create otel resource: %w", err)
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	var (
		shutdownFuncs []ShutdownFunc
		logCore       zapcore.Core
	)

	if cfg.Tracing != nil && cfg.Tracing.URL != "" {
		sd, setupErr := setupTracing(ctx, cfg.Tracing, res, logger)
		if setupErr != nil {
			logger.Warn("tracing setup failed, tracing disabled", zap.Error(setupErr))
		} else {
			shutdownFuncs = append(shutdownFuncs, sd)
		}
	}

	if cfg.Metrics != nil && cfg.Metrics.URL != "" {
		sd, setupErr := setupMetrics(ctx, cfg.Metrics, res, logger)
		if setupErr != nil {
			logger.Warn("metrics setup failed, metrics disabled", zap.Error(setupErr))
		} else {
			shutdownFuncs = append(shutdownFuncs, sd)
		}
	}

	if cfg.Log != nil && cfg.Log.URL != "" {
		sd, core, setupErr := setupLogExport(ctx, cfg.Log, res, svcName, logger)
		if setupErr != nil {
			logger.Warn("log export setup failed, OTLP logging disabled", zap.Error(setupErr))
		} else {
			shutdownFuncs = append(shutdownFuncs, sd)
			logCore = core
		}
	}

	return SetupResult{
		Shutdown: combinedShutdown(shutdownFuncs),
		LogCore:  logCore,
	}, nil
}

func combinedShutdown(funcs []ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		var firstErr error
		for _, fn := range funcs {
			if err := fn(ctx); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}
}

func noopShutdown(_ context.Context) error { return nil }

// Tracer returns a named tracer from the global provider.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

// StartSpan starts a span from context. No-op if OTel is not configured.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("crontab-go").Start(ctx, name, opts...)
}

// SpanAttr is a convenience for attribute.String.
func SpanAttr(key, val string) attribute.KeyValue {
	return attribute.String(key, val)
}

func parseEndpoint(rawURL string) (host, path string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, ""
	}
	host = u.Host
	path = u.Path
	if host == "" {
		host = rawURL
	}
	return host, path
}

func isInsecure(sig *config.ObservabilitySignal) bool {
	if sig.Insecure {
		return true
	}
	return strings.HasPrefix(sig.URL, "http://") || strings.HasPrefix(sig.URL, "grpc://")
}

func setupTracing(ctx context.Context, sig *config.ObservabilitySignal, res *resource.Resource, logger *zap.Logger) (ShutdownFunc, error) {
	host, path := parseEndpoint(sig.URL)
	if path == "" {
		path = "/v1/traces"
	}
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(host),
		otlptracehttp.WithURLPath(path),
	}
	if sig.Headers != nil {
		opts = append(opts, otlptracehttp.WithHeaders(sig.Headers))
	}
	if isInsecure(sig) {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exp, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	logger.Info("OTLP tracing enabled", zap.String("url", sig.URL))
	return tp.Shutdown, nil
}

func setupMetrics(ctx context.Context, sig *config.ObservabilitySignal, res *resource.Resource, logger *zap.Logger) (ShutdownFunc, error) {
	host, path := parseEndpoint(sig.URL)
	if path == "" {
		path = "/v1/metrics"
	}
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(host),
		otlpmetrichttp.WithURLPath(path),
	}
	if sig.Headers != nil {
		opts = append(opts, otlpmetrichttp.WithHeaders(sig.Headers))
	}
	if isInsecure(sig) {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	exp, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("metric exporter: %w", err)
	}

	interval := sig.Interval
	if interval == 0 {
		interval = 60_000_000_000
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exp, metric.WithInterval(interval))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)
	logger.Info("OTLP metrics enabled", zap.String("url", sig.URL))
	return mp.Shutdown, nil
}

func setupLogExport(ctx context.Context, sig *config.ObservabilitySignal, res *resource.Resource, svcName string, logger *zap.Logger) (ShutdownFunc, zapcore.Core, error) {
	host, path := parseEndpoint(sig.URL)
	if path == "" {
		path = "/v1/logs"
	}
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(host),
		otlploghttp.WithURLPath(path),
	}
	if sig.Headers != nil {
		opts = append(opts, otlploghttp.WithHeaders(sig.Headers))
	}
	if isInsecure(sig) {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	exp, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("log exporter: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(exp)),
		log.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	otelzapCore := otelzap.NewCore(svcName,
		otelzap.WithLoggerProvider(lp),
	)
	logger.Info("OTLP logging enabled", zap.String("url", sig.URL))
	return lp.Shutdown, otelzapCore, nil
}
