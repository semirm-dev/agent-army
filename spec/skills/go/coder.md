---
name: go/coder
description: Go implementation workflow — code placement, error handling patterns, interface design decisions, dependency management, and pre-PR verification.
scope: language-specific
languages: [go]
uses_rules: [go/patterns]
---

# Go Coder Skill

## When to Use

Invoke this skill when:
- Writing new Go features
- Implementing handlers/services
- Adding API endpoints
- Creating CLI commands
- Building workers/consumers

## Implementation Workflow

```
Understand requirements
  |
  +-- Check existing patterns in codebase
  |     (handlers, services, repositories, error types)
  |
  +-- Design types/interfaces first
  |     (types drive implementation; avoid retrofitting)
  |
  +-- Implement
  |
  +-- Write tests
  |
  +-- Run lint/vet
  |
  +-- PR
```

**Emphasize "types first"** — define structs, interfaces, and error types before writing logic.

## Code Placement Decision Tree

```
Where does this code belong?
  |
  +-- Is it a new domain/feature?
  |     YES --> New package under internal/<domain>/
  |
  +-- Is it shared utility?
  |     YES --> Does it have >2 consumers?
  |               YES --> internal/pkg/<name>/ or pkg/<name>/
  |               NO  --> Keep local to the consumer package
  |
  +-- Is it an HTTP handler?
  |     --> internal/<domain>/handler.go
  |
  +-- Is it business logic?
  |     --> internal/<domain>/service.go
  |
  +-- Is it a data access layer?
  |     --> internal/<domain>/repository.go
  |
  +-- Is it a CLI entry point?
        --> cmd/<name>/main.go
```

## Error Handling Workflow

```
What kind of error is this?
  |
  +-- Known business error (not found, conflict, validation)?
  |     --> Define sentinel error in the package
  |         var ErrNotFound = errors.New("user: not found")
  |
  +-- Wrapping an underlying error?
  |     --> fmt.Errorf("context: %w", err)
  |
  +-- Should the caller distinguish this error?
  |     YES --> Use typed error or sentinel
  |     NO  --> Wrap with context only
  |
  +-- At an API boundary?
        --> Map domain errors to HTTP status codes
            Never expose internal errors or stack traces to clients
```

## Interface Design Decision

```
Do you need to decouple from an implementation?
  |
  +-- YES
  |     --> Define interface at the consumer (not provider)
  |         Consumer declares what it needs; provider implements.
  |
  +-- Is the interface > 3 methods?
  |     YES --> Split into smaller, focused interfaces
  |
  +-- Is this for testing?
        --> Define interface at the test boundary
            Accept interfaces in constructors; pass concrete types in production
```

## Dependency Addition Checklist

- [ ] Checked if stdlib or existing dependency covers the need
- [ ] Verified license compatibility
- [ ] Checked maintenance status (last commit, open issues)
- [ ] Ran `go get` then `go mod tidy`
- [ ] No `replace` directives for production dependencies

## Pre-PR Checklist

- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run` clean
- [ ] `go test ./... -race` passes
- [ ] Exported functions have godoc comments
- [ ] Error messages include context (package:operation format)
- [ ] No `any` types without documented justification
- [ ] Context propagated through all I/O calls
