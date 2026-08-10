# Proposal Implementation Tracking

| Doc | Status | Commit | Notes |
|---|---|---|---|
| 01-correctness-bugs | **Done** | `2f58ab6` | All 7 items implemented |
| 02-config-simplification | **Done** | `96c09e3`, `af2f818` | 02.1/02.7/02.9 done; rest deferred |
| 03-extensibility | Not started | — | |
| 04-concurrency-and-shutdown | **Done** | `59037da` | 04.1/04.2/04.4/04.6 done |
| 05-refactoring-and-dedup | **Done** | `c13a397` | 5.2/5.4/5.6/5.8 done |
| 06-dead-code-and-minor | Not started | — | |
| 07-opentelemetry-observability | Not started | — | Spec only; large feature |

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
