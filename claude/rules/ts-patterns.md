<!-- Sync: Must stay in sync with cursor/101-typescript.mdc -->

# 💻 TypeScript Coding Patterns
- **Strict Mode:** All projects must use `strict: true` in tsconfig.json. No exceptions.
- **No `any`:** Never use `any`. Use `unknown` and narrow with type guards. Only exception: third-party interop where types are unavailable.
- **No non-null assertions:** Avoid the `!` operator. Use proper null checks or optional chaining.
- **Explicit return types:** All exported functions must have explicit return types.
- **Simplicity (KISS):** Prefer smaller, focused functions over complex ones. If a function >30 lines, refactor into sub-utilities.
- **Naming:** `camelCase` for variables/functions, `PascalCase` for types/classes/components, `UPPER_SNAKE_CASE` for constants.
- **Exports:** Use named exports, not default exports. Barrel files limited to one level.
- **Imports:** Order: Node built-ins → external packages → internal (absolute) → relative. Separate groups with blank lines. No circular imports.
- **Error Handling:** Define typed error classes for domain errors. Never throw plain strings. Validate external input at boundaries.
- **Async:** Always use async/await over raw promises. Never mix callbacks and promises.
- **Configuration:** No hardcoded config values. Access env vars through a validated config module, never directly via `process.env` in business logic.
- **Security:** No hardcoded secrets, tokens, or credentials. Validate and sanitize all external input. Use parameterized queries for databases. Escape user content in HTML contexts.
- **React (if applicable):** Functional components only. Custom hooks prefixed with `use`. Minimize state; derive values. Avoid `useEffect` for derived state.

## 🧪 TypeScript Testing & Quality
- **Table-Driven Tests:** Use table-driven patterns (array of cases with `for...of`) for all logic-heavy functions.
- **Mocks:** Avoid heavy mocking. Prefer fake implementations or thin interfaces. Use `vi.fn()` / `jest.fn()` only for call verification.
- **Test Organization:** Test files live next to the code they test: `service.ts` → `service.test.ts`. Use `describe` blocks for grouping. Use `beforeEach`/`afterEach` for setup/teardown.
- **Async Tests:** Always `await` async operations. Test both resolved and rejected paths. Clean up fake timers in `afterEach`.
- **Vitest Configuration:** Use `pool: 'forks'` for isolation in CI, `pool: 'threads'` for speed in local dev. Set `testTimeout: 10000` for async tests.
- **Async Error Testing:** Test rejected promises explicitly:
  ```typescript
  await expect(asyncFn()).rejects.toThrow(NotFoundError);
  await expect(asyncFn()).rejects.toMatchObject({ code: "NOT_FOUND" });
  ```
- **Test Doubles:** Prefer `vi.fn()` (vitest) or `jest.fn()` for spies. Use `vi.spyOn()` for monitoring existing methods without replacing behavior.
- **Snapshot Avoidance:** Do not use snapshot tests. They break on every change and provide no useful signal. Assert specific values instead.
