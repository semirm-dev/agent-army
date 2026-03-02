---
name: go-reviewer
description: Go code review workflow — review order, concurrency audit, error handling verification, performance red flags, security checklist, and anti-pattern detection.
scope: language-specific
uses_rules:
  - go/patterns
  - go/testing
  - code-quality
  - cross-cutting
  - security
---

# Go Reviewer Skill

## When to Use

Invoke this skill when:
- Reviewing Go PRs
- Conducting code audits on Go codebases
- Performing pre-merge checks on Go code
- Evaluating Go library choices

## Review Order

1. **Structure** — package boundaries, file organization, exported API surface
2. **Correctness** — logic, error handling chains, nil checks, type assertions
3. **Concurrency** — goroutine lifecycle, context propagation, race conditions, channel usage
4. **Performance** — allocation hotspots, unnecessary copies, N+1 queries
5. **Security** — input validation, SQL injection, secret handling
6. **Style** — naming, godoc, code formatting (defer to linter for style)

## Go Anti-Pattern Checklist

- [ ] `panic()` used in non-init, non-test code
- [ ] `interface{}` / `any` without type assertion safety
- [ ] Error returned but not checked (shadow `_`)
- [ ] Goroutine launched without lifecycle management (fire-and-forget)
- [ ] `context.Background()` used mid-call-chain instead of propagating parent context
- [ ] `sync.Mutex` held across I/O or network calls
- [ ] Large struct passed by value instead of pointer
- [ ] `init()` function with side effects
- [ ] Channel without clear ownership (who closes it?)
- [ ] `defer` inside a loop (defers until function exit, not iteration)

## Concurrency Review Checklist

- [ ] Every goroutine has a cancellation path (context or done channel)
- [ ] `sync.WaitGroup` or `errgroup.Group` tracks all spawned goroutines
- [ ] Shared mutable state protected by mutex or accessed via channels
- [ ] No unbounded goroutine spawning (use worker pool or semaphore)
- [ ] `select` statements include a `ctx.Done()` case

## Performance Red Flags

```
Large slice appended in a loop without pre-allocation?
  YES --> Suggest make([]T, 0, expectedCap)

fmt.Sprintf in hot path?
  YES --> Suggest string builder or direct concatenation

Struct with pointer fields copied frequently?
  YES --> Consider pointer receiver or pool

reflect usage in hot path?
  YES --> Flag for review — likely avoidable

Database query inside a loop?
  YES --> N+1 pattern — batch or join
```

## Security Review Checklist

- [ ] SQL queries use parameterized statements (never string concatenation)
- [ ] User input validated at handler boundary before reaching service layer
- [ ] Secrets not hardcoded or logged
- [ ] HTTP clients have timeouts set
- [ ] TLS verification not disabled outside of tests
- [ ] File paths from user input sanitized (no path traversal)
