# 04 — Concurrency, leaks & shutdown

Priority: **High** for the leak; Medium for the rest.

## 4.1 `ConcurrentPool.Unlock` leaks a goroutine per release

`Unlock` (`core/concurrency/concurrent_pool.go:53-59`) sends on an **unbuffered** `changeChan` from a fire-and-forget goroutine:

```go
p.decrease()
go func() { p.changeChan <- false }()
```

When no goroutine is currently blocked in `Lock`'s `for range p.changeChan` (the common, uncontended case), the send blocks forever → **one leaked goroutine per unlock**, unbounded over a long-running process. There is also a lost-wakeup window between `decrease()` and the channel send (the slot can be re-acquired by a new `Lock` before the signal fires).

Simplest correct replacement: a buffered-channel semaphore.

```go
type ConcurrentPool struct{ sem chan struct{} }

func NewConcurrentPool(c uint) (*ConcurrentPool, error) { /* c==0 → error; sem = make(chan struct{}, c) */ }
func (p *ConcurrentPool) Lock()   { p.sem <- struct{}{} }
func (p *ConcurrentPool) Unlock() { <-p.sem }
```

No goroutines, no lost wakeup, satisfies `sync.Locker`. Optionally add `LockContext(ctx)` (select on `ctx.Done()`) so waiters are interruptible on shutdown. Also clean up the misleading names (`available` is really total capacity; `changeChan chan interface{}` should be `chan struct{}`).

## 4.2 Unmanaged task goroutines

`core/jobs/task_handler.go:24-27` spawns one goroutine per task per event with no `recover()` and no `WaitGroup`. A panic in any one of them kills the whole process — `main.go`'s top-level recover and the per-task `recover()` in `Do` (see 01.2) do not cover goroutines spawned here. Add per-goroutine `recover()` and a `WaitGroup` (or a drain on `global.CTX().Done()`).

## 4.3 Goroutine storm per event

For N tasks per event, N goroutines are created even though the `ConcurrentPool` only bounds *executing* tasks. Under cron ticks on a job with many tasks this is spike-y. Consider a fixed worker pool sized by `job.Concurrency` and dispatching events to it instead of one goroutine per task.

## 4.4 Docker shutdown busy-spin

Already detailed in 01.4 — `connectAndListen` must return `false` on `<-ctx.Done()` (stop), keeping `true` only for the reconnect policy. `core/event/docker.go:181-182` → `104-110`.

## 4.5 Debouncer spawns a goroutine per trailing-event window

The job debouncer (`debouncer.NewDebouncedSignal` in `core/jobs/runner.go:40-42`) spawns a goroutine per debounce window that only exits when the window idles out. Under a high event rate this accumulates goroutines. Lower concern if debounce is rarely used; worth documenting.

## 4.6 Cancellation is created but never called

`ApplyTimeout` (`core/common/timeout.go:16-21`) derives a context + cancel func that is stored via `SetCancel` (`command.go:82-83`, `get.go:66-67`, `post.go:73-74`) but **never invoked** on completion — `Cancel()` appears only in tests (`core/common/cancelable_test.go`), and no `defer cancel()` exists. One cancel func leaks per attempt. `Command.Do` also re-applies `SetCancel` per connection, so `Cancel()` would only affect the last connection.

Suggested fix: `defer cancel()` in each `Do`, and wire `abstraction.Executable.Cancel()` to the in-flight request/command so an external cancel can actually abort work.

## 4.7 Shutdown is not graceful

`cmd/root.go` blocks on `<-global.CTX().Done()` and returns immediately — in-flight task goroutines are abandoned, hooks may not run, and hook chains consume the pool slot while running (`task_handler.go:38-51`), extending the window. A `WaitGroup` + drain phase would make shutdown deterministic.