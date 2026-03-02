---
name: react-testing
description: Testing Library patterns, user-centric queries, MSW mocking, and component tests
scope: language-specific
languages: [react]
extends: [testing-patterns]
uses_rules: [cross-cutting, react/patterns, typescript/patterns]
---

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
