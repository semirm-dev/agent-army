---
name: go/tester
description: Go testing workflow — test type selection, table-driven test design, mock vs real dependency decisions, testcontainers usage, coverage analysis, and benchmark methodology.
scope: language-specific
languages: [go]
uses_rules: [go/testing]
---

# Go Tester Skill

## When to Use

Invoke this skill when:
- Writing tests for Go code
- Fixing failing Go tests
- Improving Go test coverage
- Setting up integration tests with testcontainers
- Writing benchmarks

## Test Type Selection for Go

```
What are you testing?
  |
  +-- Pure function (no I/O)?
  |     YES --> Unit test with table-driven []struct + t.Run()
  |     NO  |
  |
  +-- Database query?
  |     YES --> Integration test with testcontainers-go (real DB) + transaction rollback
  |     NO  |
  |
  +-- HTTP handler?
  |     YES --> httptest.NewServer or httptest.NewRecorder with injected dependencies
  |     NO  |
  |
  +-- gRPC service?
  |     YES --> bufconn for in-memory gRPC testing
  |     NO  |
  |
  +-- External API?
  |     YES --> Interface + fake implementation for unit tests, contract test for integration
  |     NO  |
  |
  +-- CLI command?
        YES --> Test the Run() function with captured stdout/stderr
```

## Mock vs Real Dependency Decision

```
What dependency are you testing against?
  |
  +-- Database?
  |     YES --> Use real DB via testcontainers. Mock only if startup cost is prohibitive.
  |     NO  |
  |
  +-- External HTTP API?
  |     YES --> Define interface at consumer. Use fake for unit tests. Use httptest.Server for integration.
  |     NO  |
  |
  +-- File system?
  |     YES --> Use t.TempDir() (real FS). Avoid mocking.
  |     NO  |
  |
  +-- Time-dependent?
  |     YES --> Inject a Clock interface. Use fake clock in tests.
  |     NO  |
  |
  +-- Logger?
        YES --> Inject *slog.Logger. Use slog.New(slog.NewTextHandler(io.Discard, nil)) in tests.
```

## Table-Driven Test Design Workflow

1. Identify the function under test and its input/output contract
2. List cases: happy path, edge cases (zero values, nil, empty), error cases
3. Structure as `[]struct{ name string; input T; want T; wantErr bool }`
4. Each test case must be independent — no shared mutable state between cases
5. Use `t.Run(tt.name, ...)` for sub-test isolation
6. Use `t.Parallel()` when tests are independent

## Test Helper Design

- **Helpers that create test fixtures** → return the fixture, accept `testing.TB`
- **Helpers that assert** → call `t.Helper()` first so failure points to caller
- **Helpers that set up and tear down** → use `t.Cleanup()` for automatic teardown
- **Golden file tests** → `testdata/` directory, use `go test -update` flag pattern

## Coverage Analysis Workflow

```
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out        # summary by function
go tool cover -html=coverage.out        # visual HTML report
```

- Focus on uncovered branches, not percentage
- 100% coverage is not a goal — cover business logic and error paths
- Skip generated code and trivial getters from coverage analysis

## Pre-Merge Test Checklist

- [ ] `go test ./... -race` passes (race detector enabled)
- [ ] `go test ./... -count=1` passes (no cached results)
- [ ] No `t.Skip()` without a tracking issue
- [ ] Integration tests use `testcontainers-go` or build tags, not mock-heavy substitutes
- [ ] No test depends on execution order or global state
- [ ] Benchmark baselines documented for performance-critical paths
