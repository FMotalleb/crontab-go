# 01 — Correctness bugs & silent failures

Priority: **High**. These are latent crashes and "false success" paths. Fixing them is low-effort and high-value.

## 1.1 GET and POST tasks panic on their first execution

`common.Executable.Execute` (`core/common/execution_wrapper.go:30`) runs the embedded `Action` interface: `Execute` → `forceRetry` → `rh.Do(ctx)` (`execution_wrapper.go:21-27`). Only `NewCommand` initialises that field — `cmd.Action = cmd` (`core/task/command.go:39`). `NewGet` (`core/task/get.go:20-37`) and `NewPost` (`core/task/post.go:22-40`) **never set `Action`**, so the embedded interface is nil and every GET/POST dispatch panics with an interface-dispatch panic.

Why it is invisible today: no test *executes* a built Get/Post (`core/task/task_test.go` only builds them and checks non-nil), and the panic happens *before* the body of `Get.Do`/`Post.Do` runs, so the `recover()` inside those methods never catches it — the process crashes from the unrecovered task goroutine.

Suggested fix:
- `get.Action = get` in `NewGet` and `post.Action = post` in `NewPost`, mirroring `command.go:39`.
- Add an execution-level test that builds and runs a Get and a Post task.

## 1.2 Per-task `recover()` converts crashes into silent success

All three `Do` methods wrap their body in `defer recover()` without ever re-panicking or returning the recovered error (`command.go:58-67`, `get.go:55-64`, `post.go:60-69`). On any panic they log a warning and fall through to `return` the zero value (`nil`). Combined with 1.1 and 1.3, a fatal failure is recorded as a successful task run: `Execute` then runs `DoDoneHooks` (`execution_wrapper.go:33`), increments the OK Prometheus metric, and the job's done-hooks fire.

Suggested fix: distinguish "recoverable" from fatal panics; either re-panic (let `main.go`'s top-level recover report it) or convert the recovered panic into a returned error so retry/hooks treat it as a failure. Never return success from a panic.

## 1.3 Nil `TaskConnection` → nil deref → (false) success

`connection.Get` returns `nil` when no connection type matches (`core/cmd_connection/connection.go:22-24`), e.g. a bare `docker: "unix://..."` with no `local`/`container`/`label`/`image` selector. `Command.Do` then dereferences it: `connection.Prepare(...)` (`core/task/command.go:85`) panics on a nil interface, and the 1.2 recover block turns that into a successful run.

Suggested fix: `Get` should return `(cmd, error)` (or a guaranteed non-nil Local fallback) instead of `nil`, and `Command.Do` should guard `if connection == nil { return errors.New(...) }`.

## 1.4 Docker emitter busy-spins on shutdown

On context cancellation `connectAndListen` returns `true` (`core/event/docker.go:181-182`), which `BuildTickChannel` (`docker.go:104-110`) treats as "reconnect". During shutdown the derived context is already cancelled, so the loop re-creates a Docker client with a cancelled context and spins in a hot loop until the process exits.

Suggested fix: return `false` on `<-ctx.Done()`; keep `true` only for the `ErrorPolReconnect` branch (`docker.go:151-153`).

## 1.5 os_credential lookups panic instead of returning errors

`lookupGID`/`lookupUIDAndGID` call `log.Panic(...)` while also returning errors (`core/os_credential/unix_credential.go:54-86`). Because `validateCredential` runs during config validation (`config/task_validator.go:103-108`), a typo in a task's `user:`/`group:` value kills the app at boot with a zap panic rather than a clean validation error.

In addition, the root-privilege check in `Validate` (`unix_credential.go:17-35`) wrongly rejects tasks that only run inside Docker containers (the host credential is never touched there — `user` is passed to Docker at `docker_create.go:85` / `docker_attach.go:81`). The root check should be scoped to `Local` connections.

Suggested fix: honor the `error` returns and surface them as validation failures; scope the root check to local execution.

## 1.6 Escape/parse panics

- `utils.EscapedSplit` panics on a trailing backslash (`core/utils/strings.go:37`).
- `docker_create.go:74-78` indexes `EscapedSplit(volume, ':')[1]` without a length check.

There is no user-facing way to recover from these. Suggest returning errors.

## 1.7 (minor) Unreachable and dead branches

- `core/common/retry.go:91` panics for values `SetDelayModifierFromString` (retry.go:57-70) already rejects — unreachable by construction.
- `validateWebserverConfig`'s second condition `cfg.WebServerAddress != "" && cfg.WebServerPort == 0` (`config/validators.go:33-39`) has a dead `!= ""` half — the first check already returned when the address is empty.