---
name: go/architect
description: Designs Go project structure from scratch or restructures existing codebases — covering layout selection, package decomposition, dependency direction rules, interface boundaries, and architecture evolution.
scope: language-specific
languages: [go]
uses_skills: [go/patterns, code-architecture]
---

# Go Architect Skill

## When to Use

Invoke this skill when:
- Starting a new Go service or library
- Restructuring an existing Go project
- Planning package layout
- Reviewing architecture decisions
- Evaluating module boundaries

## Project Layout Decision Tree

```
Is this an application (service, CLI, worker)?
  |
  +-- Single binary?
  |     YES --> Standard layout: cmd/<name>/main.go, internal/, pkg/ (if needed)
  |
  +-- Multiple binaries (server + CLI + worker)?
        YES --> cmd/server/, cmd/cli/, cmd/worker/, shared internal/
  |
Is this a library?
  |
  +-- Single purpose?
  |     YES --> Flat layout, package root is the API
  |
  +-- Multi-purpose?
        YES --> Sub-packages under root, each with focused responsibility
```

### Reference Layout

```
cmd/
  server/main.go
  worker/main.go
internal/
  <domain>/
    handler.go
    service.go
    repository.go
    model.go
    <domain>_test.go
  pkg/          (shared internal utilities)
```

## Package Decomposition

Apply the split-vs-keep heuristics from the `code-architecture` skill. Go-specific additions:

- `internal/` prevents external packages from depending on implementation details
- Extract to `internal/pkg/<name>/` when shared by 3+ packages
- Consider `pkg/` (public) or separate module for code reusable outside this project
- Import cycles indicate wrong decomposition — `go vet` catches these

## Dependency Direction Rules

```
cmd/ → internal/ → (database, external services)
      ↓
internal/domain/ → internal/domain/ (NO! avoid cross-domain imports)
      ↓
internal/pkg/ ← (shared utilities, imported by domain packages)

Rules:
- cmd/ depends on internal/, never the reverse
- Domain packages do NOT import each other directly
- Cross-domain communication goes through interfaces or an orchestrator
- internal/pkg/ has zero dependencies on domain packages
- External dependencies (DB, HTTP clients) live behind interfaces in domain packages
```

## Interface Boundary Design

Apply the interface boundary guidelines from the `code-architecture` skill. Go-specific additions:

- Define interfaces in the consumer package, not the implementor
- Keep interfaces to 1-3 methods (Go convention). Split wider interfaces (Reader, Writer, Closer pattern)
- External dependencies (DB, API, queue) live behind interfaces in domain packages
- For cross-domain communication, use shared interfaces or events — never import one domain into another

## Wire-Up Pattern

- `cmd/main.go` is the composition root — it creates concrete implementations and wires them together
- Use constructor injection: `NewService(repo Repository, logger *slog.Logger)`
- Avoid DI frameworks in Go — explicit wiring is idiomatic and debuggable
- Group related constructors in `cmd/` by domain for readability

## Architecture Evolution Checklist

- [ ] Each package has a single clear responsibility
- [ ] No import cycles (`go vet` would catch these)
- [ ] `internal/` prevents external packages from depending on implementation details
- [ ] `cmd/` is thin — delegates to packages immediately
- [ ] Cross-domain communication uses interfaces, not direct imports
- [ ] Database/external service details hidden behind interfaces
- [ ] Package API surface is minimal — only export what consumers need
- [ ] New domains can be added without modifying existing packages
