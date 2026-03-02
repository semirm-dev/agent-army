---
name: typescript-patterns
description: Strict mode, type safety, naming, imports, error handling, and async patterns
scope: language-specific
languages: [typescript]
extends: [code-quality]
uses_rules: [security, cross-cutting, testing-patterns]
---

# TypeScript Coding Patterns
- **Strict Mode:** All projects must use `strict: true` in tsconfig.json. No exceptions.
- **No `any`:** Never use `any`. Use `unknown` and narrow with type guards. Only exception: third-party interop where types are unavailable.
- **No non-null assertions:** Avoid the `!` operator. Use proper null checks or optional chaining.
- **Explicit return types:** All exported functions must have explicit return types.
- **Naming:** `camelCase` for variables/functions, `PascalCase` for types/classes/components, `UPPER_SNAKE_CASE` for constants.
- **Exports:** Use named exports, not default exports. Barrel files limited to one level.
- **Imports:** Order: Node built-ins → external packages → internal (absolute) → relative. Separate groups with blank lines. No circular imports.
- **Error Handling:** Define typed error classes for domain errors. Never throw plain strings. Validate external input at boundaries.
- **Async:** Always use async/await over raw promises. Never mix callbacks and promises.
- **Configuration:** Access env vars through a validated config module, never directly via `process.env` in business logic.
- **Linting:** Use ESLint with strict TypeScript rules. Fix all warnings before committing.
- **Formatting:** Use Prettier (or Biome). Enforce via pre-commit hook or CI.
