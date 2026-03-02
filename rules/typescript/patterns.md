---
name: typescript/patterns
description: Strict mode, type safety, naming, imports, error handling, and async patterns
scope: language-specific
languages: [typescript]
uses_rules: [code-quality, security, cross-cutting, observability, testing-patterns]
---

# TypeScript Coding Patterns

## Naming and Structure
- **Naming:** `camelCase` for variables/functions, `PascalCase` for types/classes/components, `UPPER_SNAKE_CASE` for constants.
- **Exports:** Use named exports, not default exports. Barrel files limited to one level.
- **Imports:** Order: Node built-ins → external packages → internal (absolute) → relative. Separate groups with blank lines. No circular imports.

## Type Safety
- **Strict Mode:** All projects must use `strict: true` in tsconfig.json. No exceptions.
- **No `any`:** Never use `any`. Use `unknown` and narrow with type guards. Only exception: third-party interop where types are unavailable.
- **No non-null assertions:** Avoid the `!` operator. Use proper null checks or optional chaining.
- **Explicit return types:** All exported functions must have explicit return types.

## Error Handling
- Define typed error classes for domain errors. Never throw plain strings. Validate external input at boundaries.
- **Async:** Always use async/await over raw promises. Never mix callbacks and promises.

## Dependencies and Tooling
- **Linting:** Use ESLint with strict TypeScript rules. Fix all warnings before committing.
- **Formatting:** Use Prettier (or Biome). Enforce via pre-commit hook or CI.
- **Configuration:** Access env vars through a validated config module, never directly via `process.env` in business logic.

## Type Patterns
- **Discriminated unions** for state modeling: `type Result<T, E> = { ok: true; data: T } | { ok: false; error: E }`. Use a literal discriminant field.
- **Branded types** for domain IDs: `type UserId = string & { __brand: "UserId" }`. Prevents mixing IDs across domains.
- **`satisfies` operator** over type annotations when you want type checking without widening the inferred type.
- **Avoid `enum`:** Use `as const` objects or union types instead. Enums have runtime overhead and quirky reverse-mapping behavior. Pattern: `const Status = { Active: "active", Inactive: "inactive" } as const; type Status = (typeof Status)[keyof typeof Status];`
- **Generic constraints:** Prefer `<T extends Base>` over unconstrained generics. Keeps type inference useful at call sites.
- **Utility types:** Use `Pick`, `Omit`, `Partial`, `Required` to derive types from existing ones instead of duplicating shapes.
- **`readonly` by default:** Mark arrays as `readonly T[]` and object properties as `readonly` unless mutation is intentional.
- **`using` keyword** (TS 5.2+): Use `using` and `Disposable`/`AsyncDisposable` for deterministic resource cleanup (connections, file handles, locks). Replaces manual try/finally patterns.
- **Runtime validation:** Use schema validation libraries (e.g., Zod, Valibot) at system boundaries (API responses, form input, environment config). TypeScript types are erased at runtime -- untrusted data must be validated, not just typed.
