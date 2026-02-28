---
scope: language-specific
languages: [typescript]
---

> Extends `code-quality.md`. Language-agnostic standards apply.

# TypeScript Coding Patterns
- **Strict Mode:** All projects must use `strict: true` in tsconfig.json. No exceptions.
- **No `any`:** Never use `any`. Use `unknown` and narrow with type guards. Only exception: third-party interop where types are unavailable.
- **No non-null assertions:** Avoid the `!` operator. Use proper null checks or optional chaining.
- **Explicit return types:** All exported functions must have explicit return types.
- **Naming:** `camelCase` for variables/functions, `PascalCase` for types/classes/components, `UPPER_SNAKE_CASE` for constants.
- **Exports:** Use named exports, not default exports. Barrel files limited to one level.
- **Imports:** Order: Node built-ins → external packages → internal (absolute) → relative. Separate groups with blank lines. No circular imports.
- **Error Handling:** Define typed error classes for domain errors. Never throw plain strings. Validate external input at boundaries.
- **Async:** Always use async/await over raw promises. Never mix callbacks and promises.
- **Configuration:** Access env vars through a validated config module, never directly via `process.env` in business logic.
- **Linting:** Use ESLint with strict TypeScript rules. Fix all warnings before committing.
- **Formatting:** Use Prettier (or Biome). Enforce via pre-commit hook or CI.

## Cross-References
> See `security.md` for secrets management, input validation, and injection prevention.
> See `cross-cutting.md` for error taxonomy and coverage targets.

## Concurrency
> See `concurrency.md` for universal patterns (deadlocks, backpressure, shutdown).

### Promise Patterns
- **`Promise.all`:** Fail-fast — rejects if any promise rejects. Use when all must succeed.
- **`Promise.allSettled`:** Waits for all, returns status for each. Use when partial success is acceptable.
- **`Promise.race`:** Returns first to settle. Use for timeouts: `Promise.race([work(), timeout(5000)])`
- **`Promise.any`:** Returns first to fulfill (ignores rejections). Use for redundant requests: first successful health check, fastest mirror.

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

### Event Loop
- Never block with `while(true)`, synchronous file I/O, or heavy computation
- Use `setImmediate()` (Node.js only) or `queueMicrotask()` to yield to the event loop in long loops
- Profile with `--inspect` and Chrome DevTools for event loop delays

For testing patterns, see `testing-patterns.md`.

## Recommended Stack

### Database
> See `database.md` for universal patterns.
- **ORM (schema-first):** Prisma — schema-first ORM with `prisma.$transaction()` for transactional operations
- **ORM (SQL-like):** Drizzle — SQL-like query builder with `drizzle-kit` for migrations

### Messaging
> See `messaging-patterns.md` for universal patterns.
- **BullMQ:** Redis-based queue with priorities, scheduling, and rate limiting.
- **kafkajs:** Apache Kafka client for event streaming.

### Observability
> See `observability.md` for universal patterns.
- **OTel:** Use `@opentelemetry/auto-instrumentations-node` for Express, Fastify, pg, redis
- **Logging:** Use `pino` (recommended for Node.js) or another structured logger with JSON output

## Performance Budgets
> See `cross-cutting.md` for performance budget targets.
