---
name: go/architect
description: Go project architecture — layout selection, package decomposition, dependency direction, interface boundary design, module organization, and evolution planning.
scope: language-specific
languages: [go]
uses_rules: [go/patterns]
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

## Package Decomposition Workflow

```
Does this code represent a distinct domain concept?
  YES --> new package under internal/
  NO  --> continue

Is this code shared by 3+ packages?
  YES --> extract to internal/pkg/<name>/
  NO  --> continue

Is this code reusable outside this project?
  YES --> consider pkg/ (public) or separate module
  NO  --> continue

Is the current package >500 lines?
  YES --> Check if it has multiple responsibilities
         YES --> split
         NO  --> refactor for clarity
  NO  --> continue

Does splitting introduce import cycles?
  YES --> Rethink boundaries — cycles indicate wrong decomposition
  NO  --> safe to split
```

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

```
Does this package depend on an external service (DB, API, queue)?
  YES --> Define an interface in the consumer package.
         Implementation lives in an infra/adapter package.
  NO  --> continue

Do two domain packages need to communicate?
  YES --> Define a shared interface or event. Never import one domain into another.
  NO  --> continue

Is this for testing?
  YES --> Interface at the dependency boundary. Fake implementation for tests.
  NO  --> continue

How many methods on the interface?
  1-3 --> Good. Keep it.
  4+  --> Split into focused interfaces (Reader, Writer, Closer pattern)
```

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
