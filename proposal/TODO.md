# Proposal Implementation Tracking

| Doc | Status | Commit | Notes |
|---|---|---|---|
| 01-correctness-bugs | **Done** | `2f58ab6` | All 7 items implemented |
| 02-config-simplification | **Done** | `96c09e3`, `af2f818` | 02.1/02.7/02.9 done; rest deferred |
| 03-extensibility | Not started | — | |
| 04-concurrency-and-shutdown | Not started | — | |
| 05-refactoring-and-dedup | Not started | — | |
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
