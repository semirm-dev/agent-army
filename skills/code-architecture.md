---
name: code-architecture
description: "Architecture decision patterns — vertical slices vs layered, package-by-feature principles, dependency injection guidelines, interface boundaries, and split-vs-keep heuristics."
scope: universal
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
  YES ↓

How many bounded contexts / features?
  1-3 → Vertical slices (package by feature)
  4+  → Vertical slices with shared kernel

Is there significant cross-feature logic?
  YES → Extract shared kernel package (types, interfaces)
  NO  → Keep features fully independent
```

## Vertical Slice vs Layered

| Aspect | Vertical Slice (Recommended) | Layered |
|--------|------------------------------|---------|
| Package by | Feature | Technical concern |
| Change scope | One package per feature change | Multiple layers per feature change |
| Coupling | Low (features are independent) | High (layers depend on each other) |
| Scaling team | Easy (teams own features) | Hard (everyone touches all layers) |
| Best for | Most projects | Tiny projects, pure CRUD |

## Package-by-Feature Patterns

Organize code by feature, not by technical layer. Each feature package contains its handler, service, repository, types, and tests. Cross-cutting concerns (middleware, shared types) go in a shared package. Adapt directory structure to your project's language and framework conventions.

## Dependency Injection Patterns

- **Constructor injection** is the default pattern. Pass dependencies as constructor/factory parameters.
- **Define interfaces at the consumer side.** The package that uses the interface defines it, not the implementor.
- **Avoid service locators and global registries.** They hide dependencies and make testing harder.
- **For complex dependency graphs,** consider compile-time DI (code generation) or a lightweight DI container. Avoid reflection-based DI in production.

Adapt DI approach to your language's idiomatic patterns.

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
