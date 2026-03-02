---
scope: language-specific
languages: [react]
extends: [testing-patterns]
---

> Extends `testing-patterns.md`. See parent for universal patterns (naming, isolation, flaky prevention).

# React Testing Patterns

## Component Testing
- Use `@testing-library/react` for all component tests. Test behavior, not implementation.
- **User-centric queries:** Use `getByRole`, `getByLabelText`, `getByText`. Avoid `getByTestId` unless no semantic alternative.
- **Async testing:** Use `waitFor` or `findBy*` for async state changes. Never use `act()` directly unless wrapping non-RTL code.
- **Mock at boundaries:** Mock API calls (MSW), not internal hooks or components.

## Hook Testing
- Use `renderHook` from `@testing-library/react` for custom hook tests.
- Wrap with providers when hooks depend on context.
- Test state transitions by calling returned functions and asserting new values.

## API Mocking
- Use MSW (Mock Service Worker) for API mocking. Define handlers at suite level.
- Share handlers via a `handlers.ts` file. Override per-test when needed.
- Assert on request payloads when testing mutation flows.

## Cross-References
> See `cross-cutting.md` for coverage targets and error taxonomy.
> See `react/patterns.md` for component structure, state management, and error boundary patterns.
> See `typescript/patterns.md` for TypeScript-specific standards used in React components.
