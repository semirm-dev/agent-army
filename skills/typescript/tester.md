---
name: typescript/tester
description: TypeScript testing workflow — test type selection, vitest/jest patterns, mock design, React component testing, async test patterns, coverage analysis, and CI integration.
scope: language-specific
uses_rules:
  - typescript/testing
  - testing-patterns
  - cross-cutting
---

# TypeScript Tester Skill

## When to Use

Invoke this skill when:
- Writing tests for TypeScript code
- Fixing failing vitest/jest tests
- Improving TypeScript test coverage
- Setting up React component tests
- Configuring test infrastructure

## Test Type Selection for TypeScript

```
What are you testing?
  |
  +-- Pure function (no I/O)?
  |     YES --> Unit test with it.each (vitest/jest)
  |     NO  |
  |
  +-- API handler (Express/Fastify/Hono)?
  |     YES --> Integration test with supertest or framework test utilities
  |     NO  |
  |
  +-- React component?
  |     YES --> Component test with @testing-library/react + vitest/jest
  |     NO  |
  |
  +-- React hook?
  |     YES --> Hook test with renderHook from @testing-library/react
  |     NO  |
  |
  +-- External API interaction?
  |     YES --> MSW (Mock Service Worker) for both unit and integration
  |     NO  |
  |
  +-- Full user flow?
        YES --> E2E with Playwright or Cypress
```

## Mock vs Real Dependency Decision

```
What dependency are you testing against?
  |
  +-- External HTTP API?
  |     YES --> MSW (setupServer for Node tests, setupWorker for browser). Avoid vi.fn() for HTTP mocks.
  |     NO  |
  |
  +-- Database?
  |     YES --> Use testcontainers or real DB in Docker. Use transaction rollback per test.
  |     NO  |
  |
  +-- Browser APIs?
  |     YES --> jsdom (default in vitest/jest). Use Playwright for real browser tests.
  |     NO  |
  |
  +-- Time?
  |     YES --> vi.useFakeTimers() / jest.useFakeTimers(). Always restore in afterEach.
  |     NO  |
  |
  +-- Modules?
        YES --> vi.mock() / jest.mock() only for true external boundaries. Prefer dependency injection.
```

## React Component Test Design

- Test user behavior, not implementation details
- Use `screen.getByRole()` over `getByTestId()` (accessibility-first queries)
- Use `userEvent` over `fireEvent` (simulates real user interaction)
- Test loading, success, and error states
- Mock API calls with MSW, not component props

## Async Test Patterns

- Always `await` async operations in tests
- Use `waitFor` for assertions on async state changes in React tests
- Test both resolved and rejected promise paths
- Use `vi.useFakeTimers()` for debounce/throttle tests — advance with `vi.advanceTimersByTime()`
- Clean up fake timers in `afterEach`

## Coverage Analysis Workflow

```
vitest --coverage                          # or jest --coverage
vitest --coverage --reporter=html          # visual report
```

- Focus on branch coverage, not line coverage
- Use `/* v8 ignore next */` sparingly with justification
- Exclude test files, generated code, and type-only files from coverage
- Set CI thresholds in vitest/jest config

## Pre-Merge Test Checklist

- [ ] All tests pass (`vitest run` / `jest`)
- [ ] Coverage meets project threshold
- [ ] No `it.skip` / `describe.skip` without a tracking issue
- [ ] No `any` in test code (tests should be typed too)
- [ ] React tests use `@testing-library` queries (not `container.querySelector`)
- [ ] Async tests properly await all operations
- [ ] Mock cleanup in `afterEach` (fake timers, spies, MSW handlers)
- [ ] No snapshot tests unless intentionally tracking exact output
