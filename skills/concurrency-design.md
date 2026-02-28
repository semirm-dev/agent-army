---
name: concurrency-design
description: "Concurrency pattern selection — unified decision trees for parallel work, shared state protection, graceful shutdown, backpressure design, and pre-commit review."
scope: universal
---

# Concurrency Design Skill

## When to Use

Invoke this skill when:
- Adding concurrent processing to a feature (parallel calls, background jobs, fan-out/fan-in)
- Choosing concurrency primitives for a new operation
- Protecting shared state with locks or message passing
- Implementing graceful shutdown for a service or worker
- Diagnosing deadlocks, race conditions, or task leaks
- Reviewing concurrency safety of existing code

> See `rules/concurrency.md` for lock scope rules, deadlock avoidance patterns, and language-specific concurrency primitives.

## Concurrency Pattern Selection

```
Need concurrent work?
  |
  v
Need error propagation from concurrent tasks?
  YES --> Use a coordinated task group with error collection
  NO  |
      v
  Need to wait for all tasks to finish?
    YES --> Use a wait-for-all pattern
    NO  |
        v
    Need communication between tasks?
      YES --> Use message passing (channels, queues)
      NO  |
          v
      Is the work I/O-bound?
        YES --> Use async concurrency (non-blocking I/O)
        NO  |
            v
        CPU-bound work?
          YES --> Use process-level parallelism (bypasses single-thread limits)
          NO  --> Single sequential execution
```

Select concurrency primitives based on your language's idiomatic patterns and the decision tree above.

## Shared State Protection

### Decision Flow

```
Concurrent code accesses shared data?
  |
  v
Can you eliminate the shared state?
  YES --> Restructure: pass copies, use return values, message passing
  NO  |
      v
  Is it a simple counter or flag?
    YES --> Use atomics
    NO  |
        v
    Read-heavy workload (reads >> writes)?
      YES --> Read-write lock (multiple readers, exclusive writer)
      NO  |
          v
      Mixed read/write?
        YES --> Exclusive lock (mutex)
        NO  --> Channel or queue (producer-consumer pattern)
```

## Graceful Shutdown

Follow the 4-step shutdown sequence from `rules/concurrency.md`: stop accepting → drain in-flight → timeout → release resources.

### Shutdown Checklist

- [ ] Signal handler registered for graceful shutdown
- [ ] New requests receive appropriate error during drain period
- [ ] Hard deadline set for drain (typically 30s)
- [ ] Force-cancelled work is logged
- [ ] All resource cleanup verified (connection pools, file handles, log buffers)

## Backpressure Design

```
Producers faster than consumers?
  |
  v
Use bounded queues / buffered channels
  |
  v
Queue full?
  YES |
      v
  Can the producer wait?
    YES --> Block or backoff (preferred for internal systems)
    NO  --> Drop / reject with error (preferred for external APIs, return 429/503)
  NO  --> Continue processing
```

> See `rules/concurrency.md` for queue sizing guidelines and backpressure primitives.

## Pre-Commit Concurrency Review

Before merging concurrent code, verify:

- [ ] Every task has a clear shutdown path (cancellation, signal, or completion)
- [ ] Shared mutable state is protected (lock, atomic, or eliminated)
- [ ] Locks are acquired in consistent order across all code paths
- [ ] No I/O operations are performed while holding a lock
- [ ] Bounded queues/channels are used (no unbounded buffers)
- [ ] Tests run with race detection enabled
- [ ] Graceful shutdown is implemented and tested
- [ ] Error propagation from concurrent work is handled (not silently dropped)
