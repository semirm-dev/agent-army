---
scope: language-specific
languages: [go]
---

> Extends `code-quality.md`. Language-agnostic standards apply.

# Go Coding Patterns
- **Linting:** Use `golangci-lint` with project config. Fix all warnings before committing.
- **Packages:** Avoid "stuttering." Use `auth.Service` instead of `auth.AuthService`.
- **Error Handling:** ALWAYS wrap errors with context: `fmt.Errorf("domain: operation: %w", err)`.
  - Use `errors.Is` and `errors.As` for checking error types.
- **Interfaces:** "Accept interfaces, return concrete types." Keep interfaces small (2-3 methods max).
- **Project structure:** Follow vertical-slices architecture (feature + hexagonal/clean), package by feature. Follow Golang best practices.
- **Naming:** Use `MixedCaps` (Acronyms like `ID`, `HTTP`, `URL` should be consistent case).
- **Context:** Always pass `context.Context` as the first parameter to blocking/IO operations.
- **Panics:** Never use `panic()` for normal error paths. Reserve for truly unrecoverable situations.
- **Configuration:** Use environment variables, config files, or functional options.
- **Godoc:** All exported types, functions, and methods must have a godoc comment starting with the identifier name.
- **Dependencies:** Use `go get` to add/update dependencies. Run `go mod tidy` after changes. Never manually edit `go.mod` or `go.sum`.
- **init():** Avoid `init()` functions -- they make testing difficult and create hidden dependencies. Document if truly unavoidable.
- **Global state:** Avoid package-level `var` for mutable state. Prefer dependency injection.
- **Type assertions:** Always use the two-value form: `v, ok := x.(Type)`. Never use single-value form that panics.
- **Generics:** Use generics for type-safe collections and utilities; prefer interfaces for domain logic.
- **defer:** Use `defer` for resource cleanup. Be aware of loop and closure pitfalls (e.g., `defer` in a loop defers until function exit, not iteration end).

## Cross-References
> See `security.md` for secrets management, input validation, and injection prevention.
> See `cross-cutting.md` for error taxonomy and coverage targets.
> See `observability.md` for logging standards. Use `log/slog` for structured logging.

## Concurrency
> See `concurrency.md` for universal patterns (deadlocks, backpressure, shutdown).

### Goroutine Lifecycle
- Always pass `context.Context` for cancellation
- Use `errgroup.Group` for coordinated goroutine management with error propagation
- Use `sync.WaitGroup` only when errors aren't needed
- Every goroutine must have a clear shutdown path — never fire-and-forget

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
- **Closure capture in loop (pre-Go 1.22):** `go func() { use(v) }()` captures the variable, not the value. Pass as parameter: `go func(v T) { use(v) }(v)`. Go 1.22+ scopes loop variables per iteration.
- **Defer in loop:** `defer` runs at function exit, not loop iteration end. Use an inner function.

For testing patterns, see `testing-patterns.md`.

## Recommended Stack

### Database
> See `database.md` for universal patterns.
- **Query generation:** sqlc — type-safe SQL query generation from raw SQL
- **Driver/Pooling:** pgx/pgxpool — high-performance PostgreSQL driver with built-in connection pooling
- **Migrations:** golang-migrate — versioned, forward-only database migrations

### Messaging
> See `messaging-patterns.md` for universal patterns.
- **asynq:** Redis-based task queue. Good for background jobs with retry and scheduling.
- **watermill:** Message router supporting multiple backends (Kafka, RabbitMQ, NATS, Google Pub/Sub).

### Observability
> See `observability.md` for universal patterns.
- **OTel:** Use `go.opentelemetry.io/contrib/instrumentation/` packages (net/http, gRPC, database/sql)
- **Logging:** Use `log/slog` for structured logging

## Performance Budgets
> See `cross-cutting.md` for performance budget targets.
