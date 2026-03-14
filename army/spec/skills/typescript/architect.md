---
name: typescript/architect
description: Design TypeScript project structure — backend/frontend/library layout selection, module decomposition, dependency direction rules, type architecture (shared vs domain-scoped), barrel file strategy, and circular import prevention.
scope: language-specific
languages: [typescript]
uses_skills: [typescript/patterns, code-architecture]
---

# TypeScript Architect Skill

## When to Use

Invoke this skill when:
- Starting a new TypeScript service or library
- Restructuring an existing TypeScript project
- Planning module layout
- Designing type hierarchies
- Evaluating package boundaries

## Project Layout Decision Tree

```
Is this a backend service (Node.js)?
  |
  +-- API server?
  |     YES --> src/<domain>/ with controller, service, repository, types per domain
  |
  +-- Worker/consumer?
  |     YES --> src/workers/, src/jobs/
  |
  +-- Monorepo with shared packages?
        YES --> packages/<name>/src/ per package
  |
Is this a frontend (React)?
  |
  +-- SPA?
  |     YES --> src/components/, src/pages/, src/hooks/, src/lib/, src/types/
  |
  +-- Component library?
        YES --> src/components/<Component>/ with co-located index, component, types, tests, stories
  |
Is this a library?
  YES --> src/ with entry point src/index.ts as public API
```

### Backend Reference Layout

```
src/
  <domain>/
    <domain>.controller.ts
    <domain>.service.ts
    <domain>.repository.ts
    <domain>.types.ts
    <domain>.schema.ts     (Zod/Valibot validation)
    __tests__/
  lib/                     (shared utilities)
  config/                  (validated env config)
  index.ts                 (entry point)
```

### Frontend Reference Layout

```
src/
  components/
    <Component>/
      <Component>.tsx
      <Component>.test.tsx
      <Component>.types.ts
      index.ts
  pages/                   (route-level components)
  hooks/                   (shared custom hooks)
  lib/                     (API clients, utilities)
  types/                   (shared type definitions)
```

## Module Decomposition

Apply the split-vs-keep heuristics from the `code-architecture` skill. TypeScript-specific additions:

- Module size threshold: >300 lines suggests splitting
- Shared types used across domains go in `src/types/`
- Circular imports: move shared types to a types module, use barrel re-exports carefully

## Dependency Direction Rules

Follow the dependency direction principles from the `code-architecture` skill. TypeScript-specific wiring:

- Entry point wires everything together
- Domain modules do NOT import each other
- Cross-domain communication through shared types or event emitter
- `lib/` has zero dependencies on domain modules
- External dependencies (DB clients, HTTP) injected via constructor or factory

## Barrel File Strategy

```
Is this a domain directory?
  YES --> index.ts re-exports only the public API (controller, types)
  NO  --> continue

Is this a component directory?
  YES --> index.ts re-exports the component and its types
  NO  --> continue

Nesting depth >1?
  YES --> NO barrel files for deeply nested re-exports (causes bundle bloat and circular deps)
  NO  --> single-level barrel OK

Is this src/index.ts (library entry)?
  YES --> Export only the documented public API
```

## Type Architecture Decision

```
State with multiple variants?
  YES --> Discriminated union with literal discriminant
  NO  --> continue

External data at boundary?
  YES --> Zod/Valibot schema → infer TypeScript type from schema
  NO  --> continue

Shared across modules?
  YES --> Define in src/types/, import everywhere
  NO  --> continue

Internal to module?
  YES --> Define in <domain>.types.ts, export only what consumers need
  NO  --> continue

Preventing ID mixups?
  YES --> Branded types
```

## Architecture Evolution Checklist

- [ ] Each module directory has a single clear responsibility
- [ ] No circular imports (enforced by ESLint `import/no-cycle` or bundler analysis)
- [ ] `tsconfig.json` paths configured for clean absolute imports
- [ ] Domain modules do not import each other directly
- [ ] External services behind interfaces/types for testability
- [ ] Configuration loaded via validated config module (Zod + env)
- [ ] Barrel files limited to one level of re-export
- [ ] New domains can be added without modifying existing modules
- [ ] Strict mode enabled, no `any` leakage across module boundaries
