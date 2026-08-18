# Command Timeout / Cancel Logic — Analysis Report

> Analysis of `ApplyTimeout`, `SetTimeout`, `SetCancel`, `Cancel` as used by command tasks.
> Requested: review only — no code changes implemented for this logic (2026-08-18).

## Where the logic lives

- `core/common/timeout.go` — `Timeout` struct, `SetTimeout`, `ApplyTimeout`.
- `core/common/cancalable.go` — `Cancelable` struct, `SetCancel`, `Cancel`.
- `core/task/command.go` — `Command` embeds both; `NewCommand` calls `cmd.SetTimeout(task.Timeout)`; `Command.Do` currently does **not** apply the timeout.

```go
// common.Timeout
func (t *Timeout) SetTimeout(timeout time.Duration) { t.timeout = timeout }
func (t *Timeout) ApplyTimeout(ctx context.Context) (context.Context, func()) {
    if t.timeout != 0 {
        return context.WithTimeout(ctx, t.timeout)
    }
    return context.WithCancel(ctx)
}

// common.Cancelable
func (c *Cancelable) SetCancel(cancel func()) { c.cancel = cancel }
func (c *Cancelable) Cancel() {
    if c.cancel != nil {
        c.cancel()
    }
}
```

## Findings

### 1. `Cancelable` / `SetCancel` / `Cancel` is currently dead — intended for future use
- `Cancel()` is required by `abstraction.Executable`, but **no production code ever calls it** (only `cancelable_test.go` does).
- The func stored via `SetCancel` is therefore never invoked outside tests. It only exists to satisfy the interface.
- **Do not remove it.** It is currently dead and documented here as **meant for future use** (e.g. a signal-triggered task stop mechanism). The timeout fix below deliberately does **not** call `SetCancel` to keep the cancel machinery untouched.

### 2. `SetTimeout` vs `ApplyTimeout` split is fragile / redundant
- `SetTimeout` mutates a field on an embedded struct that is **reused across executions** (tasks are built once in `initTasks` and `Execute`d for every event). Nothing prevents mutation between executions.
- The pair is only coherent because the constructor sets it once and `Do` reads it once. There is no getter and no clear ownership of the field.

### 3. `ApplyTimeout` fallback `context.WithCancel` is odd
- When `timeout == 0` (unset/default), `ApplyTimeout` still derives a fresh context and returns a cancel that is only consumed by `defer cancel()`.
- So an execution with no timeout allocates a derived context for no observable benefit, and `SetCancel` always stores a cancel — even one that (if ever invoked) would only cancel a context the deferred `cancel()` already tore down.
- If an external `Cancel()` were added later, it would cancel the *latest* execution's context even when the task is idle or belongs to a different concurrent run (tasks are shared; concurrency is bounded by `ConcurrentPool`).

### 4. Timeout is per-attempt, not per whole task execution
- `Execute` → `ExecuteRetry` → `forceRetry` → `Do`. Each retry attempt calls `Do` again and gets a **fresh** `ApplyTimeout`.
- Result: `timeout` bounds a single attempt; total wall time can exceed `timeout` by the sum of all attempts.
- Overlaps with the separate `retry-timeout` (`WithMaxDuration`) knob — two timeouts governing different scopes, easy to confuse in config docs.

### 5. Pre-existing bug: command tasks never applied the timeout — FIXED
- `get.go` and `post.go` call `ApplyTimeout` inside `Do` and bind the HTTP request to the derived context.
- `command.go` **never called `ApplyTimeout`** — `exec.CommandContext` was given the raw, unbounded `ctx`. So `timeout` was validated, stored, and traced (`task.timeout` attribute) but never enforced for commands. A hung command ran forever.
- **Fix applied** (2026-08-18): `Command.Do` now derives `localCtx, cancel := c.ApplyTimeout(ctx)` (`defer cancel()`) and passes `localCtx` into `executeConnection`, so `exec.CommandContext` is bound to the deadline. `SetCancel` is intentionally **not** called here to respect the untouched cancel machinery.
- Regression test added: `TestCommand_Execute_AppliesTimeout` (`core/task/command_stream_test.go`) asserts a `sleep 5` task with a 200ms timeout errors out well before 5s.

### 6. Orphaned grandchildren hold the output pipe after the kill — FIXED
- After the timeout kills the direct child (`sh -c sleep 5`), an orphaned `sleep` grandchild keeps the stdout/stderr pipe write-end open. `cmd.Wait` therefore blocks until the orphan exits (5s), making the timeout appear ineffective from the task's point of view even though the command was killed at the deadline.
- **Fix applied** (2026-08-18): `core/cmd_connection/local.go` sets `l.cmd.WaitDelay = time.Second`. When the context deadline kills the process, the `watchCtx`-owned WaitDelay timer closes the descriptors and unblocks `Wait` ~1s later, so the task returns in ~deadline + 1s instead of waiting on the orphan. (Note: the orphan itself is not killed — it is left to finish in the background.)

### 7. Retry is enabled by default — changed to disabled-by-default (2026-08-18)
- Even with no retry parameters set, `ExecuteRetry` always built the backoff machinery (`retry.Do` + `buildBackoff`), and `buildBackoff` constructed `retry.NewExponential(0)` which panics (`base must be greater than 0`) whenever `retryDelay` was explicitly 0 (e.g. in tests). `WithMaxRetries(0)` prevented actual retries, but the machinery still looked/behaved enabled.
- **Change applied**: `ExecuteRetry` now returns `fn(ctx)` immediately when `maxRetries == 0` (no `retries` param set), skipping the retry/trace machinery entirely. Retry is disabled by default and only engages once at least one retry parameter (i.e. `retries`) is configured.
- Also hardened `buildBackoff`: default backoff to Exponential (matches `SetDelayModifierFromString`) instead of leaving it nil, and default `retryDelay` to 1s if 0 so enabling retries without a delay cannot panic.
- Tests added: `core/common/retry_test.go` (`TestExecuteRetryDisabledByDefault`, `TestExecuteRetryEnabledWhenRetriesSet`). Note: `retry.Do` only retries errors wrapped via `retry.RetryableError`, so the enabled-case test wraps its error.

## Open questions
- Is `Cancelable` / `Cancel()` meant to support a future stop mechanism (e.g. signal-triggered task cancellation)? If not, drop it from the `Executable` interface and the three task structs.
- Should `timeout` be per-attempt (current get/post behavior) or per whole execution including retries? If the latter, the retry loop should carry one deadline across attempts.

## Status
- [x] Apply the `ApplyTimeout` fix in `Command.Do` (see finding 5).
- [x] Bound post-kill pipe-drain wait with `WaitDelay` (see finding 6).
- [ ] Decide fate of `Cancelable`/`Cancel` — currently dead, kept for future use, do not remove.
- [x] Add a command-timeout regression test (`TestCommand_Execute_AppliesTimeout`).
- [x] Make retry disabled by default until a retry parameter is set (see finding 7).
