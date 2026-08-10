# 06 — Dead code & minor cleanups

Priority: **Low**.

## 6.1 Confirmed dead code

- `helpers.WarnOnErr` (`helpers/err_handlers.go:6-12`) — zero call sites (only `WarnOnErrIgnored` is used).
- `helpers.IsOk` (`helpers/is_ok.go`) — zero call sites; also a reflect-based generic with the usual zero-value pitfalls.
- `abstraction.Validatable` (`abstraction/validatable.go:6-8`) — type declaration only, unreferenced.
- `ctxutils.LoggerKey` / `ctxutils.ScopeKey` (`ctxutils/keys.go`) — never read; the logger actually rides go-tools' own `zap-logger` context key, so two logger-in-context mechanisms coexist. `JobKey`/`TaskKey`/`EventData`/`Environments`/`Vars` are well-used.
- `Config.Shell` / `Config.ShellArgs` as reachable config — dead from the executor's perspective; see 02.1.
- `core/common/retry.go:91` panic — unreachable by construction; see 01.7.

## 6.2 `Cancelable` infrastructure is unused in production

`SetCancel` stores a func (`core/common/cancalable.go:7-15`), but `Cancel()` is exercised only in tests (`cancelable_test.go`). Either wire it into task execution (abort on external cancel) or remove it — see 04.6.

## 6.3 Minor cleanups

- `core/common/retry_test.go` is fully commented out — regenerate or delete (already flagged as a dead file in AGENTS.md).
- Filename typo `config/job_valdiator.go` → should be `job_validator.go` (rename only; multiple files reference package `config`, not the filename).
- Stale legacy `// +build windows` line at `core/os_credential/windows_credential.go:2` (the modern `//go:build` tag is present).
- `meta/github.go` is two hard-coded strings + `fmt.Sprintf` with a single caller (`main.go:32`) — collapse to consts.
- README Codecov badge is stale — no CI step uploads coverage (badge is decorative).
- Error-handling policy in `cmd/root.go` is mixed: `panicOnErr` (`root.go:113-117`) panics, `warnOnErr` (`root.go:107-111`) `fmt.Printf`s without a newline, and `Execute` `os.Exit(1)`s discarding cobra's returned error (which cobra already prints). Pick one consistent policy.
- `WarnOnErrIgnored` usage is inconsistent where it already exists (e.g. `docker_attach.go:154-156` ignores the close error vs `docker_create.go:196-202` uses `WarnOnErrIgnored`) — standardize on `WarnOnErrIgnored`.