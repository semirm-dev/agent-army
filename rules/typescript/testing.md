---
name: typescript-testing
description: Vitest/Jest config, table-driven tests, async error testing, and test doubles
scope: language-specific
languages: [typescript]
uses_rules: [testing-patterns, cross-cutting, typescript/patterns]
---

# TypeScript Testing Patterns

## Test Naming
- Use `describe("FunctionName", () => { it("should do X when Y", ...) })`
- Use clear, behavioral names that describe expected outcomes

## Table-Driven Tests
- Use `it.each` for data-driven tests:

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
- Use `beforeEach`/`afterEach` for setup/teardown
- Use temp directories for file system tests
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
