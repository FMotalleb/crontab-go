# 07 — OpenTelemetry observability (tracing + metrics + logging)

Priority: **Medium**. Gap-closing feature — today the app exposes only Prometheus counters over the webserver; there is no tracing and no external log/metric push.

## Goal

Add OpenTelemetry (OTel) support covering all three signals:

- **Tracing** with *correct* spans — proper parent/child relationships, status, attributes, and trace-context propagation for `web-event`.
- **Metrics** pushed to a collector (OTLP push model) alongside the existing Prometheus `/metrics` endpoint.
- **Logging** exported to a collector, correlated with traces (`trace_id` / `span_id`).

Transport requirements:

- Both **HTTP** and **gRPC** OTLP exporters.
- Both **secure (TLS)** and **insecure (plaintext)** variants of each.
- **Custom HTTP headers per signal**, sufficient for `Authorization: Bearer <token>` or `X-OTLP-Auth-Token`.
- Driven by `OTEL_{TRACING,METRICS,LOG}_{URL,HEADERS}` env vars (plus identical knobs in YAML config).

## Current state (grounding)

- **Metrics**: Prometheus counters behind a central facade — `global.RegisterCounter` / `global.IncMetric` (`core/global/metrics.go:27-83`, namespace `crontab_go`). Counters: `done_tasks`, `failed_tasks` (`metrics.go:15-18`) plus per-emitter event counters (`cron`, `interval`, `init`, `web_event`, `docker`, `log_file`). Exposed via `promhttp` at `/metrics` (`core/webserver/webserver.go:105-114`) when `WEBSERVER_METRICS=true`.
- **Logging**: zap built by go-tools/log from env (`core/global/global_context.go:36-52`, `log.WithNewEnvLogger`); `LOG_*` → `ZAPLOG_*` translation in `cmd/root.go:78-98`.
- **Tracing**: none. `go.opentelemetry.io/otel` v1.44.0 is already an *indirect* dep (`go.mod:246-251`, via docker/echo) but the SDK and all exporters are absent.

## Dependency additions (direct)

All pinned to the already-present `go.opentelemetry.io/otel` v1.44.0 line (run `go mod tidy` after adding):

- `go.opentelemetry.io/otel/sdk` + `sdk/trace`, `sdk/metric`, `sdk/log`
- `exporters/otlp/otlptrace/otlptracehttp` + `otlptracegrpc`
- `exporters/otlp/otlpmetric/otlpmetrichttp` + `otlpmetricgrpc`
- `exporters/otlp/otlplog/otlploghttp` + `otlploggrpc`
- `go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp` (already indirect — HTTP client/server spans)
- `go.opentelemetry.io/contrib/bridges/otelzap` (zap → OTLP log bridge)

## Env contract (fixed)

Primary contract, exactly as requested (mirrors the `LOG_*`/`WEBSERVER_*` viper-binding style in `cmd/root.go:192-203`):

| Env var | Purpose |
|---|---|
| `OTEL_TRACING_URL` / `OTEL_METRICS_URL` / `OTEL_LOG_URL` | OTLP endpoint per signal (e.g. `http://collector:4318`, `grpcs://collector:4317`) |
| `OTEL_TRACING_HEADERS` / `OTEL_METRICS_HEADERS` / `OTEL_LOG_HEADERS` | Custom headers as a **JSON object**, e.g. `{"Authorization":"Bearer abc","X-OTLP-Auth-Token":"x"}` |
| `OTEL_TRACING_PROTOCOL` / `OTEL_METRICS_PROTOCOL` / `OTEL_LOG_PROTOCOL` | `http` or `grpc` (string) |
| `OTEL_TRACING_INSECURE` / `OTEL_METRICS_INSECURE` / `OTEL_LOG_INSECURE` | `true` = plaintext; `false` (default) = TLS |
| `OTEL_SERVICE_NAME` | Resource `service.name` (default `crontab-go`) |

Headers must be parsed as JSON in env form (a single string cannot express a map). Secrets (`Authorization`) must never be logged or added to span attributes.

## Config

New YAML block (mirrors the env contract 1:1; all optional — absent signal = disabled):

```yaml
observability:
  service-name: "crontab-go"
  tracing:
    protocol: http        # http | grpc
    url: "http://collector:4318"
    insecure: true        # true=plaintext, false=TLS
    headers:              # custom headers; needed for Authorization
      Authorization: "Bearer ..."
  metrics:
    protocol: grpc
    url: "grpcs://collector:4317"
    insecure: false
    interval: 60s         # push cadence for the periodic reader
    headers:
      Authorization: "Bearer ..."
  log:
    protocol: http
    url: "https://collector:443/v1/logs"
    headers:
      Authorization: "Bearer ..."
```

- New structs in `config/config.go`: `Observability struct { ServiceName string; Tracing, Metrics, Log *ObservabilitySignal }` and `ObservabilitySignal struct { Protocol, URL string; Insecure bool; Interval time.Duration; Headers map[string]string }`; validation: `protocol ∈ {http, grpc}`, `insecure` stays false-default, per-signal `url` required to enable that signal.
- viper `BindEnv` in `cmd/root.go` maps each `OTEL_{SIGNAL}_{VAR}` to its `observability.*` field; `interval` defaults to 60s.
- **Precedence**: YAML config > `OTEL_{SIGNAL}_*` env > standard OTLP env (`OTEL_EXPORTER_OTLP_*`, `OTEL_RESOURCE_ATTRIBUTES`) > disabled. Standard vars make a good fallback (exporters offer `FromEnv()`); request-specific vars win.
- Update `config.doc.yaml`, `config.example.yaml`, and hand-maintained `schema.json`.

## Exporter factory

New package `core/observability` with one option-building path shared by all three signals:

- **http** (`otlptracehttp` / `otlpmetrichttp` / `otlploghttp` `.New(ctx, ...)`):
  - split the full URL into host:port (`WithEndpoint`) + path (`WithURLPath`; defaults `/v1/traces`, `/v1/metrics`, `/v1/logs`);
  - `WithInsecure()` when `insecure:true` or scheme is `http://`; otherwise keep TLS (default), optionally `WithTLSClientConfig`.
- **grpc** (`otlptracegrpc` / `otlpmetricgrpc` / `otlploggrpc` `.New(ctx, ...)`):
  - `WithInsecure()` when `insecure:true` or scheme is `grpc://`; otherwise `WithTLSCredentials(credentials.NewTLS(...))` (default) for `grpcs://`.
- Headers: `WithHeaders(map[string]string)` — supported by all three signal exporters on both transports.

This yields the required 2x2 matrix per signal (transport x security), one utility, no per-signal drift.

## Lifecycle / plumbing

```go
// core/observability/setup.go (new)
func Setup(ctx context.Context, cfg *config.Observability) (shutdown func(context.Context) error, err error)
```

- Build one `resource` from `service.name` (+ standard sdk/service-version attributes); set propagator `propagation.TraceContext{}` (+ `Baggage{}`).
- Create enabled providers: `sdktrace.NewTracerProvider(WithBatcher(exporter))`, `sdkmetric.NewMeterProvider(WithReader(NewPeriodicReader(exporter, WithInterval(cfg.Metrics.Interval))))`, `sdklog.NewLoggerProvider(WithProcessor(NewBatchProcessor(exporter)))`.
- `otel.SetTracerProvider/MeterProvider/LoggerProvider`; stash the providers in the global context via `global.Put[T]` (same pattern as the `*cron.Cron` in `cmd/root.go:46-47`).
- Call from `cmd/root.go` `Run` right after config load; `defer shutdown(...)` before blocking on `<-global.CTX().Done()`, with a bounded timeout (~10s) so in-flight batches flush.
- Provider startup must be **non-fatal**: a down collector logs a warning and disables that signal instead of aborting boot (matches the soft-failure style of `core/event/events.go:20-22`).

## Tracing — span design (the "correct spans" part)

Span tree per emitted event:

```
job.<job-name>                       root span, per event emission
│  attributes: job.name, job.concurrency, event.emitter
├── event.<emitter>                  the emitter tick (event.cron / event.interval /
│                                    event.docker / event.log_file / ...)
└── task.<type>                      one per (event × task)
     │  attributes: task.name, emitter
     ├── conn.<local|docker-create|docker-attach>   (command tasks only)
     ├── http.<get|post>             via otelhttp transport (auto semconv attrs)
     └── hook.on-done / hook.on-fail (task + job hooks)
```

Correctness rules (non-negotiable for "spans must be correct"):

1. **Start/end pairing**: every span gets `defer span.End()`; never hand-rolled. Forever-loops (cron/interval/logfile/docker) emit one span per tick, ended each iteration.
2. **Parent/child across goroutines**: the event signal is `signals.NewSync` (`core/jobs/runner.go:39`), so the listener runs in the *emitter* goroutine, and `taskHandler` spawns one goroutine per task (`core/jobs/task_handler.go:24-27`). Correctness requires: create the `job.<name>` span in the listener, then create each `task.<type>` span **inside `executeTask`** via `otel.Start(ctx, ...)` where `ctx` is the job-span context already threaded through `ctxInternal` (`ctxutils.EventData` / `ctxutils.TaskKey` chain at `task_handler.go:25,40`). Never end the parent before siblings start; the batch exporter buffers, but keep task spans inside the goroutine so hierarchy is unambiguous.
3. **Trace-context propagation for `web-event`** — the distributed case that matters most: in `core/webserver/endpoint/event_dispatch.go:19-35`, extract the incoming trace
   `ctx := otel.GetTextMapPropagator().Extract(c.Request().Context(), propagation.HeaderCarrier(c.Request().Header))`
   and carry it into the listener so an upstream `traceparent` becomes the parent of the job span. The other emitters (cron/interval/docker/logfile) start fresh root spans — there is no incoming context. Optionally wrap the echo handler with `otelhttp.NewMiddleware("crontab-go.webserver")` so the HTTP server span is exported too.
4. **Status and exceptions**: only the *final* outcome decides status — `span.RecordError(err)` + `span.SetStatus(codes.Error, ...)` after retries are exhausted; transient retry attempts must not fail the span. Record retries as attributes (`retry.attempt`, `retry.count`) or `AddEvent("retry", ...)`.
5. **Attributes**: use semantic conventions and always include job/task/emitter names and the emitter type; HTTP details come free from `otelhttp` (`http.request.method`, `url.full`, `http.response.status_code`); Docker events: action, container name/image/labels.
6. **Logger correlation**: `otelzap` copies `trace_id`/`span_id` into log records from the span in ctx, so the app must keep logging through `log.Of(ctx)` with the *traced* context (the ctx that flows from `global.CTX()` must be the one spans were started on).

## Metrics — push design

- `MeterProvider` with `PeriodicReader` (interval from config, default 60s) pushing to the OTLP metric exporter.
- Rather than a parallel API, extend the existing facade: `RegisterCounter` / `IncMetric` (`core/global/metrics.go:27-83`) also maintain an OTel `Int64Counter` keyed by the same (name, labels). One change point covers all existing counters (`done_tasks`, `failed_tasks`, every emitter counter). Label maps map 1:1 to OTel attributes (cardinality is low, so safe).
- Keep the Prometheus `/metrics` endpoint working; the two paths stay independent so collection robustness never blocks pushing.

## Logging — push design

- Bridge with `otelzap.NewCore("crontab-go")` (accepts `WithMinLevel`, `WithStackTrace`), tee'd onto the existing zap core via `zapcore.NewTee(existing, otelzapCore)`. Logger construction is centralized in `core/global/global_context.go:36-52`, so this is a localized change.
- zap stays the primary/console sink; OTLP is a mirror. Flush is owned by the log exporter's batch processor, flushed on `shutdown`.
- Every existing log line that carries a traced ctx becomes a correlated log record.

## Implementation steps (ordered)

1. Add direct deps (`go mod tidy`).
2. `config`: `Observability` / `ObservabilitySignal` structs + validation; update `config.doc.yaml`, `config.example.yaml`, `schema.json`.
3. `core/observability`: header-JSON parsing, option builders, exporter factory (http/grpc x insecure/secure), `Setup`/`Shutdown`.
4. `cmd/root.go`: load the `observability` block, `BindEnv` for all `OTEL_*` vars, call `Setup`, defer shutdown.
5. Tracing instruments: emitters (span per tick), `taskHandler`/`executeTask` (job span + task children), `Command`/`Get`/`Post` `Do` (child spans), hooks, `web-event` propagate, `otelhttp` transport on GET/POST clients.
6. Metrics: bridge `IncMetric`/`RegisterCounter` to OTel counters.
7. Logging: `otelzap` tee in `core/global`.
8. Tests + docs (below).

## Testing & verification

- **Unit**: exporter option-builder matrix — (protocol http|grpc) x (insecure true|false) x (headers set|unset) produces the expected options; env→config mapping; header-JSON parsing; `Authorization` never appears in spans/logs (mock exporter).
- **Span-correctness test**: a mock `sdktrace.SpanExporter` asserts the emitted hierarchy (`job → task → conn/http`), correct parent IDs, and that a fails-after-retries task has `codes.Error` while a transient-failure-then-success task does not.
- **Propagation test**: `web-event` handler extracts a fabricated `traceparent`; resulting spans share the incoming `trace_id`.
- **E2E**: run `otelcol-contrib` locally; point a dev config at it and verify traces, pushed metrics, and correlated logs from an interval job; repeat for grpc + `insecure:false`; verify shutdown flushes.
- **CI**: `make ci` must still pass on the new files (misspell, golangci-lint v2, gofumpt); keep functions within `funlen` (<=100 lines) by extracting helpers.

## Risks / edge cases

- **Goroutine lifetime vs span lifetime** (rule 2): orphaned children if the job span is ended early — mitigated by creating task spans inside `executeTask` + a shutdown drain.
- **Endless emitters**: one span per tick, always ended; never start without end.
- **Collector outage**: exporters retry with built-in backoff; keep startup and runtime non-fatal (disable the signal with a warning).
- **Header secrecy**: never log parsed headers; document that `Authorization` is sent only to the configured endpoint.
- **Two metric paths** (Prometheus + OTLP): keep them independent; expect small duplicate-series cost.
- **Panic safety**: follow the codebase norm — observability code must not crash a task goroutine (see proposal 01/04 for the existing recover pitfalls).