<!-- Sync: Must stay in sync with cursor/503-concurrency.mdc -->

# ⚡ Concurrency Patterns

## Go

### Goroutine Lifecycle
- Always pass `context.Context` for cancellation
- Use `errgroup.Group` for coordinated goroutine management with error propagation
- Use `sync.WaitGroup` only when errors aren't needed
- Every goroutine must have a clear shutdown path — never fire-and-forget

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error {
    return processItems(ctx, items)
})
if err := g.Wait(); err != nil {
    return fmt.Errorf("processing: %w", err)
}
```

### Channel Patterns
- **Fan-out/fan-in:** Distribute work across N goroutines, merge results into one channel
- **Pipeline:** Chain stages where each reads from input channel, writes to output channel
- **Done channel:** Use `ctx.Done()` (preferred) or a dedicated `done` channel for cancellation
- Always close channels from the sender side, never the receiver

### Sync Primitives
- **`sync.Mutex`:** Protect shared state. Keep critical sections small.
- **`sync.RWMutex`:** Use when reads vastly outnumber writes
- **`sync.Once`:** Safe one-time initialization (prefer over `init()`)
- **`sync.Pool`:** Reuse temporary objects to reduce GC pressure (buffers, encoders)

### Pitfalls
- **Goroutine leak:** Every goroutine must terminate. Use context cancellation + select.
- **Race on map:** Maps are not safe for concurrent access. Use `sync.Map` or `sync.Mutex`.
- **Closure capture in loop:** `go func() { use(v) }()` captures the variable, not the value. Pass as parameter: `go func(v T) { use(v) }(v)`
- **Defer in loop:** `defer` runs at function exit, not loop iteration end. Use an inner function.

## TypeScript

### Promise Patterns
- **`Promise.all`:** Fail-fast — rejects if any promise rejects. Use when all must succeed.
- **`Promise.allSettled`:** Waits for all, returns status for each. Use when partial success is acceptable.
- **`Promise.race`:** Returns first to settle. Use for timeouts: `Promise.race([work(), timeout(5000)])`

### Worker Threads
- Use `worker_threads` for CPU-intensive operations (image processing, compression)
- Keep the event loop free — never block with synchronous computation >50ms
- Transfer `ArrayBuffer` instead of copying for large data

### Queue Patterns
- Use BullMQ/Bull for persistent job queues with retry and backoff
- Separate producers and consumers for scalability
- Implement dead-letter queues for failed jobs

### Cancellation
- Use `AbortController` + `AbortSignal` for cancellable operations
- Pass signal to `fetch`, custom async functions, and timers
- Check `signal.aborted` in long-running loops

```typescript
const controller = new AbortController();
const { signal } = controller;

await fetch(url, { signal });

// Cancel after timeout
setTimeout(() => controller.abort(), 5000);
```

### Event Loop
- Never block with `while(true)`, synchronous file I/O, or heavy computation
- Use `setImmediate()` to yield to the event loop in long loops
- Profile with `--inspect` and Chrome DevTools for event loop delays

## Python

### asyncio
- **`asyncio.gather`:** Run multiple coroutines concurrently. Use `return_exceptions=True` for partial failure handling.
- **`asyncio.TaskGroup`** (3.11+): Structured concurrency — all tasks cancel on first failure
- **`asyncio.create_task`:** Schedule coroutine execution. Always hold a reference to the task.
- **`asyncio.to_thread`:** Offload blocking/CPU-bound work to a thread pool

```python
async with asyncio.TaskGroup() as tg:
    task1 = tg.create_task(fetch_users())
    task2 = tg.create_task(fetch_orders())
# Both complete or both cancel on error

# Offload blocking I/O
result = await asyncio.to_thread(blocking_function, arg)
```

### Sync Primitives
- **`asyncio.Lock`:** Protect shared async state. Use `async with lock:`.
- **`asyncio.Semaphore`:** Limit concurrent access (e.g., max 10 DB connections)
- **`asyncio.Event`:** Signal between coroutines

### ThreadPoolExecutor
- Use for CPU-bound work or legacy blocking libraries
- Set `max_workers` explicitly based on workload type
- Use `asyncio.to_thread` (3.9+) instead of raw executor

### Pitfalls
- **Forgetting to await:** Unawaited coroutines silently don't execute. Enable `RuntimeWarning`.
- **Blocking the event loop:** `time.sleep()`, synchronous I/O in async context. Use `asyncio.sleep()`.
- **Task reference lost:** `create_task()` returns a task — if you don't hold a reference, it can be GC'd.

## Universal Patterns

### Race Condition Prevention
- **Minimize shared mutable state.** Prefer message passing (channels, queues) over shared memory.
- **Immutable data:** Pass copies, not references, when sharing data between concurrent units.
- **Atomic operations:** Use language-provided atomics for simple counters/flags.

### Deadlock Avoidance
- **Consistent lock ordering:** Always acquire locks in the same order across all code paths.
- **Timeout on locks:** Never wait indefinitely. Use context timeouts or `tryLock` patterns.
- **Minimize lock scope:** Hold locks for the shortest possible duration.

### Backpressure
- **Bounded queues:** Never use unbounded buffers. Set capacity limits.
- **Rate limiting:** Limit producers to match consumer throughput.
- **Load shedding:** Drop or reject work when queues are full, rather than accumulating.

### Graceful Shutdown
1. **Stop accepting new work** (close listeners, stop consumers)
2. **Drain in-flight work** (wait for active requests/tasks to complete)
3. **Timeout** (force-cancel remaining work after deadline)
4. **Release resources** (close connections, flush buffers)
