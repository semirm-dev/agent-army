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

## Test Isolation
- Use `t.Cleanup()` for resource teardown
- Use `t.TempDir()` for temporary file system tests
- Wrap database tests in a transaction, rollback after

## CI Parallelization
- `go test -parallel N` controls per-test parallelism
- `-count=1` disables test caching
- Always use `-race` flag: `go test ./... -race`
- Set timeouts: `go test -timeout 30s`
