---
name: concurrency
description: Select concurrency models, size worker pools, prevent deadlocks and races, design backpressure, implement graceful shutdown, and coordinate distributed tasks.
scope: universal
languages: []
uses_skills: [observability]
---

# Concurrency Skill

## When to Use

Invoke this skill when:
- Introducing goroutines, threads, async tasks, or worker pools
- Designing shared-state access across concurrent units
- Adding distributed locks or leader election
- Planning graceful shutdown for a service
- Diagnosing race conditions, deadlocks, or resource leaks
- Implementing backpressure or rate limiting on producers/consumers

## Concurrency Model Decision Tree

```
What kind of work are the concurrent units doing?

  ├── CPU-bound (computation, parsing, encoding)
  │     → Fixed worker pool, sized to available cores.
  │       Avoid spawning unbounded tasks.
  │
  ├── I/O-bound (HTTP calls, DB queries, file I/O)
  │     → Larger pool sized to expected concurrent I/O operations.
  │       Use async I/O where the language supports it.
  │
  └── Mixed
        → Separate CPU-bound and I/O-bound into distinct pools.
          Prevents CPU-bound work from starving I/O waiters.
```

## Shared State Decision Tree

```
Do concurrent units need to share data?
  NO  → Message passing (channels, queues). Preferred default.
  YES ↓

Is the shared state a simple counter or flag?
  YES → Atomic operations. No locks needed.
  NO  ↓

Is the data read-heavy (>10:1 read-to-write)?
  YES → Read-write lock (RWMutex). Many readers, exclusive writers.
  NO  → Mutex with minimal lock scope. Hold for shortest duration.
```

**Default to message passing.** Only use shared memory when message passing adds unacceptable overhead.

## Deadlock Prevention Checklist

Before adding lock-based coordination:

1. [ ] All code paths acquire locks in the same global order
2. [ ] All lock acquisitions have timeouts (no indefinite waits)
3. [ ] Lock scope is minimized — no I/O or network calls under lock
4. [ ] No nested locks, or nesting is documented with ordering rationale

## Backpressure Design

```
Is the producer faster than the consumer?
  NO  → No backpressure needed. Monitor queue depth as early warning.
  YES ↓

Can the producer slow down?
  YES → Bounded queue with blocking send. Producer waits when full.
  NO  ↓

Is dropping work acceptable?
  YES → Bounded queue with load shedding. Drop oldest or reject new.
  NO  → Buffer to durable storage (disk, database) and drain async.
```

## Graceful Shutdown Sequence

Every service must shut down cleanly:

1. **Stop accepting** — close listeners, stop queue consumers
2. **Drain in-flight** — wait for active requests/tasks to complete
3. **Timeout** — force-cancel remaining work after a deadline
4. **Release resources** — close DB pools, flush logs, release file handles

Set the drain timeout shorter than the orchestrator's kill timeout (e.g., drain at 25s if container kill is at 30s).

## Distributed Coordination Decision Tree

```
Do multiple instances need to coordinate?
  NO  → In-process concurrency only. Use local locks/channels.
  YES ↓

Can the operation be made idempotent?
  YES → Prefer idempotency over distributed locks.
        Design writes as upserts. Use idempotency keys.
  NO  ↓

Is this a one-at-a-time singleton task (cron, migration)?
  YES → Leader election via coordination service or DB advisory lock.
  NO  → Distributed lock with lease-based TTL + fencing tokens.
```

## Structured Concurrency

- Scope concurrent task lifetimes to a parent. When the parent completes or fails, all child tasks are joined or canceled.
- Never fire-and-forget goroutines, threads, or async tasks without a mechanism to await or cancel them.

## Cancellation Propagation

- Pass cancellation tokens/contexts through the call chain.
- Check for cancellation before expensive operations.
- Respect cancellation in loops and retry logic.

## Observability

- Monitor active concurrent task/worker count — unbounded growth indicates a leak.
- Track queue depth and processing latency for work queues.
- Propagate trace context across concurrent boundaries.

## Pre-Implementation Checklist

Before adding concurrency to a feature, verify:

1. [ ] Identified whether work is CPU-bound, I/O-bound, or mixed
2. [ ] Selected concurrency model (message passing vs shared state)
3. [ ] Worker pool size is bounded and justified
4. [ ] All shared mutable state uses appropriate synchronization
5. [ ] Backpressure strategy is defined for producer-consumer flows
6. [ ] Graceful shutdown drains in-flight work before exiting
7. [ ] Cancellation propagates through the call chain (context/tokens)
8. [ ] Observability: active task count, queue depth, and processing latency are monitored
9. [ ] No fire-and-forget tasks — all spawned work is awaitable or cancelable
