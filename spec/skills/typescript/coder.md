---
name: typescript/coder
description: Drive TypeScript feature implementation end-to-end — types-first design, code placement decisions (controller/service/schema/component), typed error handling, dependency vetting, and pre-PR verification checklist.
scope: language-specific
languages: [typescript]
uses_skills: [typescript/patterns]
---

# TypeScript Coder Skill

## When to Use

Invoke this skill when:
- Writing new TypeScript features
- Implementing API handlers
- Building React components
- Creating Node.js services
- Adding CLI tools

## Implementation Workflow

```
Understand requirements
  |
  +-- Check existing patterns in codebase
  |     (controllers, services, types, schemas)
  |
  +-- Design types first (interfaces, discriminated unions)
  |     (types drive implementation; avoid retrofitting)
  |
  +-- Implement
  |
  +-- Write tests
  |
  +-- Run lint
  |
  +-- PR
```

**Emphasize "types first"** — define the shape of data before writing logic.

## Code Placement Decision Tree

```
Where does this code belong?
  |
  +-- New domain/feature?
  |     YES --> src/<domain>/ directory
  |
  +-- Shared utility?
  |     --> src/lib/ or src/utils/ (only if used by 3+ modules)
  |
  +-- API handler?
  |     --> src/<domain>/<domain>.controller.ts or <domain>.handler.ts
  |
  +-- Business logic?
  |     --> src/<domain>/<domain>.service.ts
  |
  +-- Types/interfaces?
  |     --> src/<domain>/<domain>.types.ts
  |
  +-- Validation schemas?
  |     --> src/<domain>/<domain>.schema.ts
  |
  +-- React component?
        --> src/components/<ComponentName>/
            Co-locate index, component, types, and test files
```

## Type Design Decision Tree

```
How should this be modeled?
  |
  +-- State with variants?
  |     --> Discriminated union with literal discriminant field
  |
  +-- Need to prevent ID mixups?
  |     --> Branded type: type UserId = string & { __brand: "UserId" }
  |
  +-- Deriving from existing type?
  |     --> Use Pick, Omit, Partial, Required — don't duplicate shapes
  |
  +-- Enum-like values?
  |     --> as const object + derived union type (avoid enum)
  |
  +-- External data (API response, form input, env vars)?
        --> Runtime validation with Zod/Valibot at boundary
```

## Error Handling Workflow

```
What kind of error is this?
  |
  +-- Known business error?
  |     --> Define typed error class extending base DomainError
  |
  +-- At API boundary?
  |     --> Catch domain errors in handler middleware
  |         Map to HTTP status codes
  |
  +-- General rules
        --> Never throw plain strings
        --> Never use any in catch blocks
        --> Always type the error: catch (error: unknown)
            Narrow before use (instanceof, type guards)
```

## Dependency Addition Checklist

- [ ] Checked if native APIs or existing deps cover the need
- [ ] Checked bundle size impact (bundlephobia.com)
- [ ] Verified TypeScript types included or `@types/` package available
- [ ] Added via `npm install` / `pnpm add`
- [ ] Tree-shakeable (ESM exports)
- [ ] No duplicate functionality with existing dependencies

## Pre-PR Checklist

- [ ] `tsc --noEmit` clean (no type errors)
- [ ] ESLint clean
- [ ] Prettier/Biome formatted
- [ ] Tests pass (`vitest` / `jest`)
- [ ] No `any` types without documented justification
- [ ] All exported functions have explicit return types
- [ ] No `!` non-null assertions without documented justification
- [ ] Bundle size impact checked for frontend changes
