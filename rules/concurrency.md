---
name: concurrency
description: Race condition prevention, deadlock avoidance, backpressure, and graceful shutdown
scope: universal
languages: []
---

# Concurrency Patterns

## Race Condition Prevention
- **Minimize shared mutable state.** Prefer message passing (channels, queues) over shared memory.
- **Immutable data:** At message-passing boundaries, pass copies or immutable values — not mutable references — between concurrent units.
- **Atomic operations:** Use language-provided atomics for simple counters/flags.

## Deadlock Avoidance
- **Consistent lock ordering:** Always acquire locks in the same order across all code paths.
- **Timeout on locks:** Never wait indefinitely. Use context timeouts or `tryLock` patterns.
- **Minimize lock scope:** Hold locks for the shortest possible duration.

## Backpressure
- **Bounded queues:** Never use unbounded buffers. Set capacity limits.
- **Rate limiting:** Limit producers to match consumer throughput.
- **Load shedding:** Drop or reject work when queues are full, rather than accumulating.

## Graceful Shutdown
1. **Stop accepting new work** (close listeners, stop consumers)
2. **Drain in-flight work** (wait for active requests/tasks to complete)
3. **Timeout** (force-cancel remaining work after deadline)
4. **Release resources** (database connections, file handles, network sockets, message queue connections). Flush write buffers and pending logs.

## Observability
- Monitor active concurrent task/worker count — unbounded growth indicates a leak.
- Track queue depth and processing latency for work queues.
- Propagate trace context across concurrent boundaries.
