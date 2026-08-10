# Proposal Implementation Tracking

| Doc | Status | Commit | Notes |
|---|---|---|---|
| 01-correctness-bugs | **Done** | `2f58ab6` | All 7 items implemented |
| 02-config-simplification | **Done** | `96c09e3`, `af2f818` | 02.1/02.7/02.9 done; rest deferred |
| 03-extensibility | **Done** | `4729b14` | 3.2/3.3 done |
| 04-concurrency-and-shutdown | **Done** | `59037da` | 04.1/04.2/04.4/04.6 done |
| 05-refactoring-and-dedup | **Done** | `c13a397` | 5.2/5.4/5.6/5.8 done |
| 06-dead-code-and-minor | **Done** | `53b4663` | 6.1/6.3 done |
| 07-opentelemetry-observability | **Done** | `c8725d0`, `d1aa7cb`, `5dda3c2`, `5994db3` | 07.1-07.8 complete |

## 01-correctness-bugs — Done

- [x] 01.1 Set `Action` on Get/Post + regression tests
- [x] 01.2 Convert `recover()` to return errors
- [x] 01.3 Guard nil `TaskConnection` in `Command.Do`
- [x] 01.4 Fix Docker emitter shutdown busy-spin
- [x] 01.5 `os_credential`: error returns instead of `log.Panic`; scope root check to local
- [x] 01.6 `EscapedSplit` trailing backslash → literal; volume index guard
- [x] 01.7 Dead branch + unreachable panic removed

## 02-config-simplification — In progress

Selected items (low-risk, high-clarity wins):

- [x] 02.1 Remove dead `shell`/`shell_args` fields + viper defaults + BindEnv (`config.go:11-12`, `root.go:119-126, 192-203`)
- [x] 02.7 Make `log-line-breaker` effective or remove (`logfile.go:137` hard-codes `\n`)
- [x] 02.9 Delete dead `abstraction.Validatable` interface (`abstraction/validatable.go`)

Deferred (need discussion / broader change):
- 02.2 json/mapstructure tag drift — touches parser output semantics
- 02.3 `WebserverUsername` casing — cosmetic, low value
- 02.4 `data` field semantics — design decision
- 02.5 Connection semantics — touches proposals 03
- 02.6 Logging env translation — touches go-tools internals
- 02.8 Duplicate validation — touches proposals 03

## 04-concurrency-and-shutdown — Done

- [x] 04.1 Replace `ConcurrentPool` with buffered-channel semaphore
- [x] 04.2 Add `recover()` to task goroutines in `task_handler`
- [x] 04.4 Fix Docker shutdown busy-spin (done in 01.4)
- [x] 04.6 Add `defer cancel()` in task `Do` methods

## 05-refactoring-and-dedup — Done

- [x] 5.2 Drop redundant `Retry` from `Hooked` struct
- [x] 5.4 Extract `doHTTP` helper for GET/POST dedup
- [x] 5.6 Fix infinite busy-loops in `DockerCreate.Execute`
- [x] 5.8 Sort env keys for deterministic ordering

## 06-dead-code-and-minor — Done

- [x] 6.1 Remove dead `helpers.WarnOnErr`, `helpers.IsOk`, `ctxutils.LoggerKey`, `ctxutils.ScopeKey`
- [x] 6.3 Delete commented-out `retry_test.go`
- [x] 6.3 Rename `job_valdiator.go` → `job_validator.go`
- [x] 6.3 Remove stale legacy build tag in `windows_credential.go`
- [x] 6.3 Collapse `meta/github.go` to consts

## 03-extensibility — Done

- [x] 3.2 Deterministic priority-based selection via `RegisterWithPriority`
- [x] 3.3 Nil-safe registry: `Build`/`Get` return `(O, error)` instead of nil/panic

## 07-opentelemetry-observability — Partial

- [x] 07.1 Add `Observability`/`ObservabilitySignal` config structs + schema
- [x] 07.2 Add OTel SDK + HTTP exporter dependencies
- [x] 07.3 Create `core/observability` package with Setup/Shutdown
- [x] 07.4 Wire observability into `cmd/root.go`
- [x] 07.8 Update `schema.json` with Observability definitions
- [x] 07.5 Add tracing spans to task_handler + task Do methods
- [ ] 07.6 Bridge Prometheus metrics to OTel counters
- [ ] 07.7 Bridge zap logging to OTLP via otelzap
