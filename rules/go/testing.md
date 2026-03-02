---
name: go/testing
description: Table-driven tests, race detection, benchmarks, and Go test helpers
scope: language-specific
languages: [go]
uses_rules: [go/patterns]
---

# Go Testing Patterns

## Test Naming
- Use `TestFunctionName_scenario` (e.g., `TestCreateUser_duplicateEmail`)
- Use `t.Run()` for subtests within table-driven tests

## Table-Driven Tests
- Use `[]struct` with `t.Run()` subtests:

```go
tests := []struct {
    name  string
    input int
    want  int
}{
    {"positive", 5, 25},
    {"zero", 0, 0},
    {"negative", -3, 9},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        if got := Square(tt.input); got != tt.want {
            t.Errorf("Square(%d) = %d, want %d", tt.input, got, tt.want)
        }
    })
}
```

## Test Helpers
- Use `t.Helper()` in test helper functions so failure locations point to the caller, not the helper.
- Helper functions should accept `testing.TB` to work with both tests and benchmarks.

## Test Isolation
- Use `t.Cleanup()` for resource teardown
- Use `t.TempDir()` for temporary file system tests
- Wrap database tests in a transaction, rollback after
- Call `t.Parallel()` in subtests when tests are independent. Capture loop variables in table-driven tests (Go < 1.22) to avoid data races.

## Fuzz Testing
- Use `testing.F` for fuzz tests on parsers, validators, and serialization logic.
- Seed the corpus with known edge cases via `f.Add()`.
- Fuzz tests should assert invariants (no panic, round-trip equality), not specific outputs.

## Benchmarks
- Use `func BenchmarkX(b *testing.B)` for performance-sensitive code.
- Run with `go test -bench=. -benchmem` to include memory allocations.
- Compare before/after with `benchstat`. Require statistically significant results.
- Never benchmark in CI by default -- run on dedicated hardware or use relative comparison.

## Integration Tests
- Use `testing.Short()` to skip slow integration tests during local development: `if testing.Short() { t.Skip("skipping integration test") }`. Run full suite in CI with `go test ./...` (no `-short` flag).
- Use `testcontainers-go` to spin up real dependencies (databases, caches, queues) in Docker for integration tests. Prefer real dependencies over mocks for integration-level verification.

## CI Parallelization
- `go test -parallel N` controls per-test parallelism
- `-count=1` disables test caching
- Always use `-race` flag: `go test ./... -race`
- Set timeouts: `go test -timeout 30s`
