# High Severity Tasks

> From project review (2026-08-18).

## 1. Command task `timeout` is configured but never applied

- **File**: `core/task/command.go`
- **Severity**: high
- **Impact**: the `timeout` config field is silently ignored for command tasks, so a long-running or hung command executes indefinitely.

### Evidence

`NewCommand` sets the value (`command.go:43`) and embeds `common.Timeout` (`command.go:52`), but `Command.Do` never calls `ApplyTimeout`. Compare the HTTP tasks:

- `core/task/get.go:93` — `localCtx, cancel := g.ApplyTimeout(ctx)`
- `core/task/post.go:98` — `localCtx, cancel := p.ApplyTimeout(ctx)`

`command.go` instead passes the raw, unbounded `ctx` into `executeConnection` → `cmdConn.Prepare` → `exec.CommandContext`. The embedded `Cancelable` is also inert here (`SetCancel` is never called for commands).

### Fix

```go
// inside Command.Do, before running connections:
localCtx, cancel := c.ApplyTimeout(ctx)
defer cancel()
c.SetCancel(cancel)
// pass localCtx to executeConnection
```

### Acceptance criteria

- A command task with `timeout` set is killed once the deadline is exceeded.
- `timeout: 0` (unset) keeps current unbounded behavior.
- `go test -race ./...` (with CGO enabled) passes.

### Todo

- [x] Apply timeout context in `Command.Do`
- [ ] Wire `SetCancel` for command tasks (intentionally NOT done — cancel machinery is dead, meant for future use; see `TODO/command.md`)
- [x] Add a test asserting a command exceeding its `timeout` returns an error
- [x] Bound the post-kill pipe-drain wait (`WaitDelay`) so a timed-out command returns promptly even when a shell leaves an orphaned child holding the output pipe
- [x] Run `make ci`

### Additional notes (2026-08-18)

- The command timeout fix surfaced a second issue: killing the direct child (`sh`) leaves orphaned grandchildren (e.g. `sleep`) holding the stdout/stderr pipes, so `cmd.Wait` blocked until they exited. Fixed by setting `l.cmd.WaitDelay = time.Second` in `core/cmd_connection/local.go`.
- Retry is now disabled by default unless a retry parameter is set — see `TODO/command.md` and `core/common/retry.go`.