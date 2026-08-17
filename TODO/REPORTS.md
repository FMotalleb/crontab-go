# Reports

## Report 001 — Command timeout leaves tracing span unfinished

**Date:** 2026-08-17

**Status:** open (root cause identified, no fix applied)

### Symptom

When a local `command` task hits its configured `timeout`, the tracing spans for
that execution are never ended. The `task.command` span — and therefore the
parent `job.<name>/task.execute` and `event.<emitter>` spans — stay open
("unfinished") in the trace backend.

### How to reproduce

Any local command whose shell *forks* the real workload, with a `timeout` set,
e.g.:

```yaml
tasks:
  - command: "sleep 30"
    timeout: 1s
```

Observed with an in-memory span exporter (tracetest): after the timeout fires,
`Command.Execute` never returns and the `task.command` span end time stays zero.

Minimal raw repro (no crontab code involved):

```go
ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
defer cancel()
cmd := exec.CommandContext(ctx, "/bin/sh", "-c", "sleep 30")
cmd.Stdout = io.Discard
cmd.Stderr = io.Discard
err := cmd.Run() // never returns, even after the context deadline fires
```

The same hang occurs with a manual `Process.Kill()` — the kill happens, `Wait`
does not return.

### Root cause

`core/cmd_connection/local.go` executes commands with `exec.CommandContext`
(line 56). On timeout, `exec.CommandContext` kills **only the direct child**
with `Process.Kill()` (SIGKILL, no process group). Two os/exec behaviors then
combine to hang the task:

1. When `cmd.Stdout`/`cmd.Stderr` are `io.Writer`s (here
   `NewPrefixWriter(os.Stderr, …)` from `core/task/command.go`), `Cmd` creates
   `os.Pipe`s and copy goroutines, and **`Cmd.Wait()` blocks until those pipes
   reach EOF** — i.e. until *every* write-end file descriptor is closed.
2. `sh -c "sleep 30"` **forks** (does not `exec`) the real command, so `sleep`
   is a grandchild that inherits the stdout/stderr pipe write-ends. When the
   timeout kills `sh`, the orphaned grandchild keeps running and keeps the pipe
   write-ends open. EOF never arrives, so `Wait()` blocks forever.

Contrast proved by `sh -c "exec sleep 30"` (no forked child): the timeout kills
the direct child, pipes close, and `Run()` returns `signal: killed` promptly.
`pgrep` after the hung repro confirmed the orphaned `sleep 30` processes.

Because `Wait()` never returns, the whole chain in `core/task/command.go`
(`Do` → `Execute` → `conn.Execute`) never unwinds, so `defer span.End()`
(line 66) never runs.

### Not the retry mechanism (ruled out)

The internal retry loop (`core/common/retry.go`) is *not* the cause:

- The original hang reproduces with the default `retries: 0` (which
  `WithMaxRetries(0)` turns into a single attempt — the first `Next()` call
  immediately returns stop), so the hang happens inside the **first** `Do`,
  before the retry loop ever runs. The raw `exec.CommandContext` repro above
  has no crontab retry code at all and still hangs.
- Empirically verified: `exec sleep 30` with `timeout: 500ms`, `retries: 2`
  returns after ~1.8s (3 attempts) and **all 3 `task.command` spans are
  ended** — one per attempt, each closed by its own `defer span.End()`.
- The retry loop only ever holds the *outer* `job.*/task.execute` span open
  across attempts (each attempt gets a fresh timeout and its own span), and
  only for a bounded number of attempts.

One real (minor) interaction worth noting: because `ApplyTimeout` is applied
per-attempt inside `Do`, a task with `timeout` + `retries > 0` can run for
`(retries+1) × timeout` plus backoff delays — the configured `timeout` does not
bound the whole task. The outer span stays open that long, but still ends.

### Impact (worse than an unfinished span)

- **Leaked spans:** `task.command`, `job.<name>/task.execute`
  (`core/jobs/task_handler.go`), and `event.<emitter>` (`core/event/tracing.go`)
  all stay unfinished, skewing trace latencies and inflating open-span counts.
- **Job wedges permanently:** `executeTask` in `core/jobs/task_handler.go`
  holds the job's `ConcurrentPool` semaphore slot (`lock.Lock()`, released by
  defer after `task.Execute` returns) for the entire execution. A hung task
  leaks that slot forever; with the default `concurrency: 1` the whole job
  stops processing future events.
- **Hooks and metrics never run:** `on-done`/`on-fail` hooks and the
  ok/err Prometheus counters (incremented in `core/common/hooked.go`) are
  skipped.
- **Retry is defeated:** the per-attempt timeout lives inside `Do`, so a hung
  `Wait` also blocks the retry loop in `core/common/retry.go`.

### Scope

- Affected: local connection only (`core/cmd_connection/local.go` is the only
  `exec.CommandContext` in the codebase). Any `sh -c`-style command that spawns
  a child (i.e. virtually all real commands, plus daemons / `&` background
  processes) is at risk.
- Not affected: docker connections (`docker_create.go`, `docker_attach.go`)
  use the docker API with the timeout context, not os/exec pipes — no equivalent
  mechanism.

### Suggested fixes (not yet applied)

1. **Kill the whole process tree on timeout** (Go 1.20+): in `Prepare`, set
   `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` and provide a custom
   `cmd.Cancel` that SIGKILLs the process group, e.g.
   `syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)`.
2. **Bound the pipe drain** with `cmd.WaitDelay` (Go 1.20+) so `Wait()` cannot
   block indefinitely on the copy goroutines even if a grandchild survives.
3. **Belt-and-braces:** wrap `task.Execute(ctx)` in `executeTask`
   (`core/jobs/task_handler.go`) with a `context.WithTimeout` derived from the
   task timeout, so the concurrency slot is always released and the
   `job.*/task.execute` span always ends, regardless of connection behavior.

### Related code

- `core/cmd_connection/local.go` — `exec.CommandContext`, `Start`/`Wait`
- `core/task/command.go` — timeout via `ApplyTimeout`, `defer span.End()`
- `core/jobs/task_handler.go` — `job.*/task.execute` span, concurrency lock
- `core/common/timeout.go` — `ApplyTimeout`
- `core/event/tracing.go` — `event.<emitter>` span
