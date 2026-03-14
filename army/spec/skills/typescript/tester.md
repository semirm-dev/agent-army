---
name: typescript/tester
description: Select and implement the right TypeScript test strategy — unit/integration/E2E/component test selection, MSW for API mocking, React Testing Library patterns, coverage analysis, and pre-merge test verification.
scope: language-specific
languages: [typescript]
uses_skills: [typescript/patterns, testing]
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

## Test Naming
- Use `describe("FunctionName", () => { it("should do X when Y", ...) })`
- Use clear, behavioral names that describe expected outcomes

## Table-Driven Tests

Use `it.each` for data-driven tests (TypeScript/vitest/jest pattern):

```typescript
it.each([
  { name: "positive", input: 5, want: 25 },
  { name: "zero", input: 0, want: 0 },
  { name: "negative", input: -3, want: 9 },
])("$name", ({ input, want }) => {
  expect(square(input)).toBe(want);
});
```

## Test Isolation

- Use `beforeEach`/`afterEach` for setup/teardown (vitest/jest pattern)
- Clean up fake timers in `afterEach`

## CI Parallelization
- vitest: `--pool threads` (default) for speed, `--pool forks` for isolation
- jest: `--maxWorkers=N` for parallel execution
- Set `testTimeout: 10000` for async tests in vitest/jest config

## Async Error Testing
- Test rejected promises explicitly:

```typescript
await expect(asyncFn()).rejects.toThrow(NotFoundError);
await expect(asyncFn()).rejects.toMatchObject({ code: "NOT_FOUND" });
```

- Always `await` async operations in tests. Test both resolved and rejected paths.

## Mock Patterns
- Use `vi.fn()` (vitest) or `jest.fn()` for spies and stubs:

```typescript
const mockSend = vi.fn().mockResolvedValue({ success: true });
const service = new EmailService(mockSend);
await service.notify("user-123");
expect(mockSend).toHaveBeenCalledWith("user-123");
```

- Use `vi.spyOn()` to monitor existing methods without replacing behavior:

```typescript
const spy = vi.spyOn(logger, "warn");
await processItem(invalidItem);
expect(spy).toHaveBeenCalledWith(expect.stringContaining("invalid"));
spy.mockRestore();
```

- Prefer fake implementations or thin interfaces over heavy mocking. Use `vi.fn()` / `jest.fn()` only for call verification.
