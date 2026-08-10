# crontab-go — Codebase Review & Change Proposals

This directory documents code-review findings and concrete change/refactor proposals for the `crontab-go` codebase. Nothing here is implemented — every document is a proposal with suggested changes, rationale, and precise `file:line` anchors so the work can be picked up directly.

| Document | Scope | Priority |
|---|---|---|
| [01-correctness-bugs.md](01-correctness-bugs.md) | Latent crashes and silent-failure paths | High — fix first |
| [02-config-simplification.md](02-config-simplification.md) | Config schema simplification and consistency | Medium |
| [03-extensibility.md](03-extensibility.md) | Registry-driven extensibility and flexibility | Medium |
| [04-concurrency-and-shutdown.md](04-concurrency-and-shutdown.md) | Goroutine leaks, shutdown behavior, panic safety | High |
| [05-refactoring-and-dedup.md](05-refactoring-and-dedup.md) | Code deduplication and structural refactors | Low/Medium |
| [06-dead-code-and-minor.md](06-dead-code-and-minor.md) | Dead code and minor cleanups | Low |

## Top recommendations (highest impact)

1. **Fix GET/POST tasks crashing on first execution** — `NewGet`/`NewPost` never set `common.Executable.Action`, so every GET/POST task panics at its first tick (01.1). One-line fix each plus a regression test.
2. **Stop treating fatal panics as success** — the per-task `recover()` blocks swallow nil-interface and nil-connection panics and return `nil`, running done-hooks for work that never happened (01.2, 01.3).
3. **Fix the Docker event shutdown busy-spin** — `connectAndListen` returns `true` on context cancellation, so the emitter reconnects in a tight loop until the process dies (01.4).
4. **Remove the `ConcurrentPool.Unlock` goroutine leak** — replace with a buffered-channel semaphore (04.1).
5. **Collapse / simplify the config schema** — drop dead `shell: / shell_args:` keys, fix `json`/`mapstructure` tag drift, and derive event/task validation from the registries instead of hand-maintained lists (02, 03).

The single most valuable change is **01.1**, which unblocks the two default task types (`get` and `post`).