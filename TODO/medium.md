# Medium Severity Tasks

> From project review (2026-08-18).

## 1. Docker event regex validation checks the wrong field

- **File**: `config/job_validator.go`
- **Severity**: medium
- **Impact**: invalid image / label regexes pass validation and panic at runtime in `regexp.MustCompile`.

### Evidence

In `dockerValidation` (`job_validator.go:188-194`) the fold lambda ignores the iterated value (`_ string`) and compiles `s.Docker.Name` on every pass instead of each item in the list:

```go
err := utils.Fold(checkList, nil, func(initial error, _ string) error {
    if initial != nil {
        return initial
    }
    _, err := regexp.Compile(s.Docker.Name) // BUG: should compile v (the current item)
    return err
})
```

Consequences:
- `s.Docker.Image` and every label-value regex are never validated.
- The name check is repeated once per list item.

A bad image/label pattern passes config validation and later panics in `regexp.MustCompile` at `core/event/docker.go:92-95` and `reshapeLabelMatcher` (`docker.go:227`).

### Fix

```go
_, err := regexp.Compile(v) // compile the current item
```

### Acceptance criteria

- A job with an invalid `docker.image` regex is rejected at config validation with a clear error.
- A job with an invalid label-value regex is rejected.
- No runtime panic from `regexp.MustCompile` on user-supplied patterns.

### Todo

- [x] Fix the fold lambda to compile the current item `v`
- [x] Add a unit test for invalid image and label regexes
- [ ] Run `make ci`

## 2. `populateVars` mutates a context-shared map (data race)

- **File**: `core/task/helper.go`
- **Severity**: medium
- **Impact**: concurrent task executions of the same job share the `Vars` map, causing a data race and cross-execution variable leakage.

### Evidence

`helper.go:20` aliases the map already stored in the context, then writes into it:

```go
varTable := old               // aliases the shared map
for k, v := range task.Vars {
    varTable[k], err = template.EvaluateTemplate(v, varTable) // mutates shared map
}
```

`task_handler.go:34` spawns task goroutines concurrently (bounded only by the job `concurrency` limit), so multiple executions can mutate the same `Vars` map at once.

This is not caught in CI because `go test -race` only runs when `CGO_ENABLED != 0` (`Makefile:64-67`).

### Fix

```go
varTable := maps.Clone(old) // copy before mutating
```

### Acceptance criteria

- `go test -race ./...` (CGO enabled) reports no races for a job with `concurrency > 1`.
- Vars set by one execution do not leak into a concurrent sibling execution.

### Todo

- [x] Clone the vars map in `populateVars` before mutation
- [x] Add a race-focused test with concurrent task executions
- [x] Verify with `go test -race` and run `make ci`