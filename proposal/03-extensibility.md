# 03 — Extensibility & flexibility

Priority: **Medium**.

## 3.1 The generator registry is good — validate through it too

`generator.Core[I,O]` (`generator/generator.go`) is a tidy first-match registry with no central dispatch table. It is used for event emitters (`core/event/events.go:14`), task builders (`core/task/tasks.go:13`), and connections (`core/cmd_connection/connection.go:12`), each self-registered via `init()`.

But config *validation* re-implements the same knowledge by hand, outside the registry:
- `JobEvent.Validate` keeps a hard-coded "exactly one active" list — `Interval != 0, Cron != "", WebEvent != "", Docker != nil, LogFile != "", OnInit` — built via `utils.List` + `utils.Fold` (`config/job_valdiator.go:145-171`). Adding an emitter requires editing the registry (fine) **and** this list (easy to forget).
- `validateActionsList` hard-codes `Get != "", Command != "", Post != ""` (`config/task_validator.go:110-133`) plus per-action field rules in `validateFields` (`task_validator.go:85-101`).

Suggested change: ask the registries "which providers match this config?" instead of maintaining parallel list literals; expose a `Match` capability at registration so compilation and validation stay in lockstep. This directly reduces the cost of adding a new emitter/task type from 5 touch points to 2-3.

## 3.2 Make first-match selection deterministic and introspectable

Selection order today is implicit Go `init()` filename order (e.g. connections resolve `docker_attach` → `docker_create` → `local`). Surprising: `local: true` + `image: x` silently picks Docker, and selection changes if a file is renamed.

Suggested changes to `generator/generator.go`:
- Add registration priority (or an explicit order) so `Local` always wins when `local: true`, then DockerAttach, then DockerCreate — independent of filename ordering (see 01.3, 02.5).
- Add `RegisterNamed(name, fn)` + a `Names()` introspection method so a `--list-emitters` / `--list-tasks` flag, or a schema generator, can enumerate providers.
- Add `TaskConnection.Validate` so unusable selectors (none set, or mutually-exclusive families set) fail at config time instead of nil-panicking at run time.

## 3.3 Nil-safe registry contract

Failure semantics are inconsistent across the three registries:
- `event.Build` logs `Warn` and returns `nil` (`core/event/events.go:16-23`).
- `task.Build` **panics** (`core/task/tasks.go:18`).
- `connection.Get` returns `nil` (`core/cmd_connection/connection.go:22`).

Callers don't guard nil uniformly (`initializer.go:45-51` appends nil emitters; `command.go:85` dereferences a nil connection). Make `Get` return `(O, error)` with "no match" as a build/validation error, and log at one consistent severity.

## 3.4 Real flexibility: `type:`-based selection + declarative emitters

Today the config field named after each implementation (`cron:`, `interval:`, `web-event:`, ...) *is* the discriminator. Two directions for more flexibility:

1. Add an explicit `type:` selector (e.g. `type: docker` + params) so a config can opt into a provider by name and future providers are addressable without first gaining a named top-level field. The named-field approach can be kept for backwards compatibility.
2. Allow per-task *hooks* to also accept the full set of task types (they already are full `Task`s, so this mostly works today).

## 3.5 Document the external-addon story

Everything registers via `init()` side effects, so an external Go module can't add an emitter/task without a blank import somewhere. At minimum, document the full touch point list for adding a provider (registry `init()`, `config` struct field, validator list, `schema.json`, `config.doc.yaml`); alternatively introduce a small plugin hook. See also 3.1, which removes most of that list.