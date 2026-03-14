---
name: code-architecture
description: Guides structuring new projects and modules using vertical slice architecture, package-by-feature layouts, dependency injection, and interface boundaries. Delegates language-specific examples to language architect skills.
scope: universal
languages: []
uses_skills: [code-quality]
---

# Code Architecture Skill

## When to Use

Invoke this skill when:
- Starting a new project or feature module
- Deciding between vertical slice vs layered architecture
- Creating new packages or modules
- Reviewing dependency injection patterns
- Evaluating whether to split or keep code together

## Architecture Decision Tree

```
Is this a new project or major module?
  YES |

How many bounded contexts / features?
  1-3 -> Vertical slices (package by feature)
  4+  -> Vertical slices with shared kernel

Is there significant cross-feature logic?
  YES -> Extract shared kernel package (types, interfaces)
  NO  -> Keep features fully independent
```

## Vertical Slice vs Layered

| Aspect | Vertical Slice (Recommended) | Layered |
|--------|------------------------------|---------|
| Package by | Feature | Technical concern |
| Change scope | One package per feature change | Multiple layers per feature change |
| Coupling | Low (features are independent) | High (layers depend on each other) |
| Scaling team | Easy (teams own features) | Hard (everyone touches all layers) |
| Best for | Most projects | Tiny projects, pure CRUD |

## Package-by-Feature Layout

Organize code by feature (vertical slice), not by technical layer. Each feature directory contains its handlers, services, repositories, and tests together. See language-specific architect skills (`go/architect`, `typescript/architect`, `python/architect`, `react/architect`) for concrete directory layouts and naming conventions.

## Dependency Injection

Prefer constructor injection: define interfaces at the consumer side, inject dependencies via constructor parameters. See language-specific architect and patterns skills for idiomatic DI examples per language.

## Interface Boundary Guidelines

- **Define at consumer side:** The package that _uses_ the interface defines it, not the package that implements it
- **Keep narrow:** 2-3 methods maximum (Go). If wider, split into focused interfaces
- **No leaking:** Public APIs should not expose internal types (database models, framework types)
- **Cross-boundary DTOs:** Use dedicated types for data crossing package boundaries

## Split vs Keep Heuristic

**Keep together when:**
- Types change for the same business reason
- Functions share the same data structures
- Splitting would create circular dependencies
- The package is under 500 lines

**Split when:**
- Types change for different business reasons
- The package has multiple unrelated responsibilities
- Different parts have different deployment/scaling needs
- The package exceeds 1000 lines with distinct sections

## Module Boundary Checklist

Before creating a new package/module, verify:

1. [ ] Can you describe its purpose in one sentence without "and"?
2. [ ] Does it have a clear public API (types, functions, interfaces)?
3. [ ] Are its dependencies pointing inward (toward domain, not infrastructure)?
4. [ ] Could another team work on it independently?
5. [ ] Does it avoid duplicating types/logic from existing packages?
6. [ ] Is the name descriptive, non-generic, and non-stuttering?
