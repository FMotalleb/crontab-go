# Low Severity Tasks

> From project review (2026-08-18).

## 1. Remove debug `/foo` endpoint

- **File**: `core/webserver/webserver.go:93-98`
- **Severity**: low
- **Description**: the webserver ships a hardcoded `GET /foo` → `"bar"` route that looks like leftover test/debug code. If it is not an intentional health check, remove it; if it is, document it as such.
- **Todo**:
  - [x] Confirm intent (debug leftover vs health check)
  - [x] Remove or document the `/foo` route — removed

## 2. Remove dead code `NewErrMetaData`

- **File**: `core/event/events.go:42`
- **Severity**: low
- **Description**: `NewErrMetaData` is referenced only by its own test (`core/event/events_test.go`), never by production code. Either remove it (and its test) or keep it only if there is a plan to use it.
- **Todo**:
  - [x] Remove `NewErrMetaData` and its test

## 3. Fix stale goimports `local-prefixes`

- **File**: `.golangci.yml:79-80`
- **Severity**: low
- **Description**: `local-prefixes` still points at `github.com/fmotalleb/junction` instead of `github.com/fmotalleb/crontab-go` (already flagged in AGENTS.md). As a result first-party imports are not grouped separately and `goimports` output is inconsistent with `.vscode/settings.json`.
- **Todo**:
  - [x] Change `local-prefixes` to `github.com/fmotalleb/crontab-go`
  - [ ] Re-run `make lint` and commit any import regroupings

## 4. Use mutex-synced context accessor in cron emitter

- **File**: `core/event/cron.go:77`
- **Severity**: low
- **Description**: `BuildTickChannel` uses `global.CTX().Context` (raw embedded field) while every other emitter uses the mutex-synced `global.CTX()` accessor. Replace for consistency and correct synchronization.
- **Todo**:
  - [x] Replace `global.CTX().Context` with `global.CTX()`

## 5. Process / test coverage follow-ups

- **Severity**: low (non-blocking)
- **Description**:
  - No tests for `jobs`, `webserver`, `global`, docker connections, or `os_credential` (confirmed in AGENTS.md). Two of the fixed bugs live in untested code — add coverage where practical.
  - `go test -race` only runs when `CGO_ENABLED != 0` (`Makefile:64-67`); consider always racing or making the CI matrix run it explicitly.
  - README coverage badge is stale; update or remove.
- **Todo**:
  - [ ] Add unit tests for the docker connection compiler and webserver endpoints
  - [ ] Decide whether to always run `-race` in CI
  - [ ] Refresh or drop the stale coverage badge