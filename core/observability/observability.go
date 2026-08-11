// Package observability provides OpenTelemetry setup for tracing, metrics, and logging.
package observability

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
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
	"google.golang.org/grpc/credentials"

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
	resAttrs := make([]attribute.KeyValue, 0, len(cfg.Attributes)+1)
	resAttrs = append(resAttrs, semconv.ServiceNameKey.String(svcName))
	for key, value := range cfg.Attributes {
		resAttrs = append(resAttrs, attribute.String(key, value))
	}
	defRes := resource.Default()
	res, err := resource.Merge(
		defRes,
		resource.NewWithAttributes(
			defRes.SchemaURL(),
			resAttrs...,
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

type transportKind int

const (
	transportHTTP transportKind = iota
	transportGRPC
)

// signalEndpoint describes how to reach an OTLP signal endpoint parsed from its URL scheme.
type signalEndpoint struct {
	transport  transportKind
	endpoint   string // host:port
	path       string // HTTP URL path; empty for gRPC
	tls        bool   // https:// or grpcs:// negotiate TLS
	skipVerify bool   // insecure: true skips TLS certificate verification
}

func parseEndpoint(sig *config.ObservabilitySignal) (signalEndpoint, error) {
	raw := sig.URL
	ep := signalEndpoint{transport: transportHTTP, skipVerify: sig.Insecure}
	if !strings.Contains(raw, "://") {
		// Bare host:port, defaults to plaintext HTTP.
		ep.endpoint = raw
		return ep, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return signalEndpoint{}, fmt.Errorf("parse OTLP endpoint %q: %w", raw, err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http":
		ep.transport, ep.tls = transportHTTP, false
	case "https":
		ep.transport, ep.tls = transportHTTP, true
	case "grpc":
		ep.transport, ep.tls = transportGRPC, false
	case "grpcs":
		ep.transport, ep.tls = transportGRPC, true
	default:
		return signalEndpoint{}, fmt.Errorf("unsupported OTLP URL scheme %q (use http, https, grpc or grpcs)", u.Scheme)
	}
	if u.Host != "" {
		ep.endpoint = u.Host
	} else {
		ep.endpoint = strings.TrimPrefix(raw, u.Scheme+":")
	}
	ep.path = u.Path
	return ep, nil
}

// skipVerifyTLS allows explicit user opt-in to ignore certificate verification.
//
//nolint:gosec // skip verification is an explicit configuration choice
func skipVerifyTLS() *tls.Config {
	return &tls.Config{InsecureSkipVerify: true} //nolint:gosec // explicit user opt-in
}

func traceHTTPOpts(ep signalEndpoint, headers map[string]string) []otlptracehttp.Option {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(ep.endpoint),
		otlptracehttp.WithURLPath(pathOrDefault(ep.path, "/v1/traces")),
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlptracehttp.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlptracehttp.WithTLSClientConfig(skipVerifyTLS()))
	}
	return opts
}

func traceGRPCOpts(ep signalEndpoint, headers map[string]string) []otlptracegrpc.Option {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(ep.endpoint),
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlptracegrpc.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(skipVerifyTLS())))
	default:
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(nil)))
	}
	return opts
}

func pathOrDefault(path, def string) string {
	if path == "" {
		return def
	}
	return path
}

func newTraceExporter(ctx context.Context, ep signalEndpoint, headers map[string]string) (sdktrace.SpanExporter, error) {
	if ep.transport == transportGRPC {
		return otlptracegrpc.New(ctx, traceGRPCOpts(ep, headers)...)
	}
	return otlptracehttp.New(ctx, traceHTTPOpts(ep, headers)...)
}

func setupTracing(ctx context.Context, sig *config.ObservabilitySignal, res *resource.Resource, logger *zap.Logger) (ShutdownFunc, error) {
	ep, err := parseEndpoint(sig)
	if err != nil {
		return nil, err
	}
	exp, err := newTraceExporter(ctx, ep, sig.Headers)
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

func metricHTTPOpts(ep signalEndpoint, headers map[string]string) []otlpmetrichttp.Option {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(ep.endpoint),
		otlpmetrichttp.WithURLPath(pathOrDefault(ep.path, "/v1/metrics")),
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlpmetrichttp.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(skipVerifyTLS()))
	}
	return opts
}

func metricGRPCOpts(ep signalEndpoint, headers map[string]string) []otlpmetricgrpc.Option {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(ep.endpoint),
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(skipVerifyTLS())))
	default:
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(nil)))
	}
	return opts
}

func newMetricExporter(ctx context.Context, ep signalEndpoint, headers map[string]string) (metric.Exporter, error) {
	if ep.transport == transportGRPC {
		return otlpmetricgrpc.New(ctx, metricGRPCOpts(ep, headers)...)
	}
	return otlpmetrichttp.New(ctx, metricHTTPOpts(ep, headers)...)
}

func setupMetrics(ctx context.Context, sig *config.ObservabilitySignal, res *resource.Resource, logger *zap.Logger) (ShutdownFunc, error) {
	ep, err := parseEndpoint(sig)
	if err != nil {
		return nil, err
	}
	exp, err := newMetricExporter(ctx, ep, sig.Headers)
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

func logHTTPOpts(ep signalEndpoint, headers map[string]string) []otlploghttp.Option {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(ep.endpoint),
		otlploghttp.WithURLPath(pathOrDefault(ep.path, "/v1/logs")),
	}
	if len(headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlploghttp.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlploghttp.WithTLSClientConfig(skipVerifyTLS()))
	}
	return opts
}

func logGRPCOpts(ep signalEndpoint, headers map[string]string) []otlploggrpc.Option {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(ep.endpoint),
	}
	if len(headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(headers))
	}
	switch {
	case !ep.tls:
		opts = append(opts, otlploggrpc.WithInsecure())
	case ep.skipVerify:
		opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(skipVerifyTLS())))
	default:
		opts = append(opts, otlploggrpc.WithTLSCredentials(credentials.NewTLS(nil)))
	}
	return opts
}

func newLogExporter(ctx context.Context, ep signalEndpoint, headers map[string]string) (log.Exporter, error) {
	if ep.transport == transportGRPC {
		return otlploggrpc.New(ctx, logGRPCOpts(ep, headers)...)
	}
	return otlploghttp.New(ctx, logHTTPOpts(ep, headers)...)
}

func setupLogExport(ctx context.Context, sig *config.ObservabilitySignal, res *resource.Resource, svcName string, logger *zap.Logger) (ShutdownFunc, zapcore.Core, error) {
	ep, err := parseEndpoint(sig)
	if err != nil {
		return nil, nil, err
	}
	exp, err := newLogExporter(ctx, ep, sig.Headers)
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
