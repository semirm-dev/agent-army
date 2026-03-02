---
name: go-testing
description: Table-driven tests, race detection, benchmarks, and Go test helpers
scope: language-specific
languages: [go]
uses_rules: [testing-patterns, cross-cutting, go/patterns]
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

## Fuzz Testing
- Use `testing.F` for fuzz tests on parsers, validators, and serialization logic.
- Seed the corpus with known edge cases via `f.Add()`.
- Fuzz tests should assert invariants (no panic, round-trip equality), not specific outputs.

## Benchmarks
- Use `func BenchmarkX(b *testing.B)` for performance-sensitive code.
- Run with `go test -bench=. -benchmem` to include memory allocations.
- Compare before/after with `benchstat`. Require statistically significant results.
- Never benchmark in CI by default -- run on dedicated hardware or use relative comparison.

## CI Parallelization
- `go test -parallel N` controls per-test parallelism
- `-count=1` disables test caching
- Always use `-race` flag: `go test ./... -race`
- Set timeouts: `go test -timeout 30s`
