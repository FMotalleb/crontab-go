# 02 — Config simplification & consistency

Priority: **Medium**. Mostly removes confusion for users and for future agents editing config code.

## 2.1 `shell:` / `shell_args:` top-level keys are dead

`Config.Shell`/`Config.ShellArgs` (`config/config.go:11-12`) are never consumed anywhere: grep shows the only references are the struct declaration and the viper defaults/binding in `cmd/root.go:121-126, 194-202`. The executor resolves the shell from the **process environment** merged with task env — `env["SHELL"]` / `env["SHELL_ARGS"]` (`core/cmd_connection/command/command.go:82-96`) — populated via `.env` + `godotenv`.

Consequences:
- A YAML `shell:` key silently does nothing.
- The Windows default `cmd.exe /c` set in `root.go:126` is dead too; `getShell()` (`command/command.go:84-95`) hard-codes `/bin/sh` regardless of GOOS.
- Type mismatch: `Config.ShellArgs []string` vs the executor's single `:`-joined string split via `utils.EscapedSplit`.

Suggested fix: choose one channel. Either make the executor read the config fields, or delete the fields + viper defaults + bindings (recommended — env-only already works; just document `SHELL`/`SHELL_ARGS` as env vars only).

## 2.2 json vs mapstructure tag drift

The parser emits YAML from `json` tags while viper reads by `mapstructure`. Several diverge:
- `webserver_address` (mapstructure) vs `webserver_listen_address` (json) — `config.go:15`.
- `error-limit-count` (mapstructure) vs `error-limit` (json) — `config.go:57`.

A config round-tripped through `parse` emits different keys than viper accepts. Align the two tag sets.

## 2.3 Naming inconsistency

`WebserverUsername` (`config.go:17`, lowercase "s") vs `WebServerAddress`/`WebServerPassword`/`WebServerMetrics` on the same struct. Rename to `WebServerUsername`.

## 2.4 Under-specified `data` for POST

`Task.Data any` (`config.go:74`) accepts any YAML, but the sender always JSON-marshals it (`core/task/post.go:78-85`), so you cannot POST a raw string or form body even though the schema suggests you can. Either make `data` explicitly an object and document "always JSON", or add a `content-type`/raw-body mode.

## 2.5 Task-connection semantics are confusing

- `connections[].docker` (the Docker socket/host for a task connection, `config.go:105`) overlaps conceptually with `docker.connection` on `DockerEvent` (`config.go:52`) — two config keys for the same concept (docker connection/context path).
- When no `connections` array is given, `Command` silently defaults to local (`core/task/command.go:69-76`); when given, selection is first-match by filename order — `docker_attach` before `docker_create` before `local` (see 01.3 and 03.2). `TaskConnection` has no validation, so `local: true` + `image: x` silently resolves to Docker.
- A bare `docker: "<host>"` with no container/image/local selector matches nothing and nil-panics at run time.

Suggested fix: validate `TaskConnection` (exactly one transport family), document the precedence, and reuse one "Docker connection" concept across events and task connections.

## 2.6 Logging env translation layer

`root.go:78-98` translates `LOG_*` env vars into `ZAPLOG_*` (go-tools/log's contract), mutating the process env via `os.Setenv`. `LOG_STDOUT=false` only takes effect when `LOG_FILE` is set, which the README (§Logging) implies otherwise. `-v`/`--verbose` precedence against `LOG_LEVEL` is undocumented and surprising (`SetDebugDefaults` then `FromEnv` → `LOG_LEVEL` wins). Consider one env namespace and a documented precedence.

## 2.7 `log-line-breaker` has no effect

Read from config (`core/event/logfile.go:66-68`) and used in metric labels, but the reader hard-codes `ReadString('\n')` and trims `"\r\n"` (`logfile.go:137-145`). Either honour it or drop the option.

## 2.8 Duplicate validation at startup

Every job is validated twice: `Config.Validate()` via `root.go:146` and again `job.Validate()` per job in `core/jobs/runner.go:36`. Both compile cron expressions and Docker-action regexes. Validate once.

## 2.9 Dead abstraction

`abstraction.Validatable` (`abstraction/validatable.go:6-8`) is referenced nowhere (`grep` returns zero hits). The concrete `Validate` methods happen to match its shape. Delete it, or wire it in so registries can drive validation (see 03.1).