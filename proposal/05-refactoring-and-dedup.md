# 05 — Refactoring & deduplication

Priority: **Low/Medium**. Large readability win, no behaviour change intended.

## 5.1 Collapse the two hook layers

Task-level hooks (`OnDone`/`OnFail` on a task) run inside `common.Executable.Execute` (`core/common/execution_wrapper.go:30-44`). Job-level hooks (`Hooks.done`/`Hooks.failed`) run in `executeTask` (`core/jobs/task_handler.go:42-51`). Same feature, two code paths, **divergent error handling**: task hooks collect errors and `zap.Warn` them (`core/common/hooked.go:47-63`), job hooks discard them with `_ = ...`.

There is also a duplicated runner: `executeTasks` (`core/common/hooked.go:65-72`) vs the inline loop in `task_handler.go:44-51`. Both fire per task-completion, and job hooks reuse the completed task's context (including `ctxutils.TaskKey`).

Suggested fix: model a job run as one root `Executable` whose on-done/on-fail are the job hooks, so there is a single hook mechanism and a single runner (reuse `common.executeTasks`).

## 5.2 `Retry` embedded twice

`Executable` embeds both `Retry` and `Hooked` (`execution_wrapper.go:15-19`), and `Hooked` also embeds `Retry` (`hooked.go:13`). Method promotion resolves to the shallower `Executable.Retry`; the `Hooked.Retry` copy is dead weight and invites `SetMaxRetry` on the wrong receiver. Drop it from `Hooked`.

## 5.3 Duplicated emitter loop skeleton

Cron/interval/logfile/docker emitters repeat the same skeleton — `ctx, cancel := context.WithCancel(global.CTX())` + `select { emit; IncMetric; NewMetaData(...); <-ctx.Done() }` (`core/event/cron.go:77-91`, `interval.go:61-83`, `logfile.go:125-165`, `docker.go:123-184`). Extract a `loopEmitter(ctx, metric, onTrigger func() (abstraction.Event, bool))`-style helper; the genuinely unique part per emitter is only the trigger source.

## 5.4 GET vs POST share ~40 duplicated lines

Request build, header loop, `client.Do`, body-close defer, status/error handling, and the debug response dump are duplicated (`core/task/get.go:69-100` vs `post.go:76-120`). Extract `doHTTP(ctx, method, url, headers, body, log)`. Also fix the copy-pasted wrong message `"recovering command execution from a fatal error"` in the HTTP tasks (`get.go:59`, `post.go:61`) — see 01.2.

## 5.5 Docker client + log-copy duplication

`docker_attach.go` and `docker_create.go` duplicate:
- Docker client construction — `WithHost` + `WithAPIVersionNegotiation` + default-socket fallback (`docker_attach.go:92-102`, `docker_create.go:109-119`).
- The ~30-line output log-copy tail (`docker_attach.go:138-166` vs `docker_create.go:129-212`).
- `Connect()`/`Disconnect()` lifecycle (`docker_attach.go:172-174`, `docker_create.go:218-220`).

The default socket `unix:///var/run/docker.sock` is hard-coded in 3+ places (`docker_create.go:66`, `docker_attach.go:66`, `core/event/docker.go:36`). Extract `newClient(host string)` + `const defaultDockerSocket` + a shared log-copy helper. Move the `DockerConnection` defaulting out of `Prepare` (which mutates its constructor input) into the constructors (`docker_create.go:64-67`, `docker_attach.go:64-67`).

## 5.6 Infinite no-backoff busy-loops in `DockerCreate.Execute`

`docker_create.go:155-179` retries `ContainerStart` / `ContainerStats` in `for { ... }` loops with no context check and no backoff — a missing image or rejected container spins CPU hot forever. Bound with `ctx.Err()`/backoff or return the first error. Also deduplicate the double "container started" debug log at lines 167 and 181.

## 5.7 Connection selection order is fragile

Covered in 03.2 — make precedence explicit rather than relying on `init()` filename order (docker_attach < docker_create < local).

## 5.8 Non-deterministic env ordering

`envReshape` (`core/cmd_connection/command/command.go:71-80`) emits `KEY=VALUE` pairs in map order. Sort keys for reproducible env, and consider quoting/escaping values containing newlines.