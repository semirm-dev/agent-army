---
name: react/tester
description: Orchestrates React test creation with decision trees for test type selection, mock-vs-real boundaries, MSW integration setup, hook testing patterns, accessibility automation, and coverage analysis workflows.
scope: language-specific
languages: [react]
uses_skills: [react/patterns, testing]
---

# React Tester Skill

## When to Use

Invoke this skill when:
- Writing tests for React components
- Testing custom hooks
- Setting up MSW for API mocking
- Adding accessibility tests
- Fixing failing React tests
- Improving frontend test coverage

## Test Type Selection

```
What are you testing?
  |
  +-- React component (renders UI)?
  |     YES --> Component test with @testing-library/react
  |             Use userEvent for interactions, screen queries for assertions
  |
  +-- Custom hook (shared logic)?
  |     YES --> renderHook from @testing-library/react
  |             Wrap with providers if hook depends on context
  |
  +-- Component with API calls?
  |     YES --> Component test + MSW for API mocking
  |             Never mock hooks directly — mock at the network boundary
  |
  +-- Form with validation?
  |     YES --> Component test with userEvent.type + form submission
  |             Test valid, invalid, and edge-case inputs
  |
  +-- Accessibility compliance?
  |     YES --> jest-axe / vitest-axe for automated checks
  |             Manual keyboard navigation tests for custom widgets
  |
  +-- Full user flow across pages?
        YES --> E2E with Playwright or Cypress
                Component tests are insufficient for cross-page flows
```

## Mock vs Real Decision Tree

```
What dependency are you testing against?
  |
  +-- REST / GraphQL API?
  |     YES --> MSW (setupServer for vitest/jest, setupWorker for browser)
  |             Define handlers in a shared handlers.ts file
  |             Override per-test with server.use() for error/edge cases
  |
  +-- React context / providers?
  |     YES --> Render with real providers in a test wrapper
  |             Create a renderWithProviders utility
  |
  +-- Router?
  |     YES --> MemoryRouter with initialEntries for route tests
  |
  +-- Browser APIs (localStorage, matchMedia, IntersectionObserver)?
  |     YES --> Use jsdom mocks or vi.stubGlobal()
  |             Clean up in afterEach
  |
  +-- Time (debounce, throttle, animations)?
  |     YES --> vi.useFakeTimers() + vi.advanceTimersByTime()
  |             Always vi.useRealTimers() in afterEach
  |
  +-- Child components?
        NO  --> Do NOT mock child components
                Test the composed behavior (integration over isolation)
```

## Component Test Design Workflow

```
1. Identify what the user sees and does
   (not component internals or state shape)
   |
2. Choose queries (priority order):
   getByRole > getByLabelText > getByText > getByTestId
   |
3. Simulate user interactions with userEvent
   (prefer userEvent over fireEvent — simulates real behavior)
   |
4. Assert on visible outcomes
   (text content, element presence, ARIA attributes)
   |
5. Test all states:
   - Loading (skeleton/placeholder visible)
   - Success (data rendered correctly)
   - Error (error message + retry button visible)
   - Empty (empty state message visible)
```

## Hook Test Patterns

```
Testing a custom hook?
  |
  +-- Pure computation hook (no side effects)?
  |     --> renderHook, assert on result.current
  |
  +-- Hook with state transitions?
  |     --> renderHook, call actions via act(), assert new state
  |
  +-- Hook that fetches data?
  |     --> renderHook + MSW for API mock
  |         Wrap with QueryClientProvider if using TanStack Query
  |
  +-- Hook that depends on context?
        --> renderHook with wrapper option providing the context
```

## MSW Integration Setup

```
Test suite setup:
  |
  +-- Create src/test/handlers.ts with default happy-path handlers
  |
  +-- Create src/test/server.ts:
  |     const server = setupServer(...handlers)
  |     export { server }
  |
  +-- In test setup file (vitest.setup.ts / jest.setup.ts):
  |     beforeAll(() => server.listen())
  |     afterEach(() => server.resetHandlers())
  |     afterAll(() => server.close())
  |
  +-- Per-test overrides for error cases:
        server.use(
          http.get("/api/users", () => HttpResponse.error())
        )
```

## Accessibility Test Automation

```
Automated a11y checks:
  |
  +-- Run jest-axe / vitest-axe on every component test
  |     const { container } = render(<Component />)
  |     expect(await axe(container)).toHaveNoViolations()
  |
  +-- Test keyboard navigation for custom widgets
  |     Tab to element, Enter/Space to activate, Escape to close
  |
  +-- Verify focus management
  |     Modal open: focus moves to modal
  |     Modal close: focus returns to trigger
  |     Route change: focus moves to main content
  |
  +-- Verify ARIA attributes
        aria-expanded, aria-selected, aria-invalid update correctly
```

## Coverage Analysis Workflow

```
vitest --coverage                          # or jest --coverage
vitest --coverage --reporter=html          # visual report
```

- Focus on branch coverage over line coverage
- Exclude test files, stories, and type-only files from coverage
- Set CI thresholds in vitest/jest config
- Prioritize coverage on: user-facing interactions, error states, conditional rendering

## Pre-Merge Test Checklist

- [ ] All tests pass (`vitest run` / `jest`)
- [ ] Coverage meets project threshold
- [ ] No `it.skip` / `describe.skip` without a tracking issue
- [ ] Component tests use `@testing-library` queries (not `container.querySelector`)
- [ ] User interactions use `userEvent` (not `fireEvent`)
- [ ] Async assertions use `waitFor` or `findBy*`
- [ ] API mocks use MSW (not `vi.mock` on fetch/axios)
- [ ] Mock cleanup in `afterEach` (fake timers, spies, MSW handlers)
- [ ] Accessibility checks included (`jest-axe` / `vitest-axe`)
- [ ] Loading, error, and empty states tested
- [ ] No snapshot tests unless intentionally tracking exact output

## Component Testing Patterns
- Use `@testing-library/react` for all component tests. Test behavior, not implementation.
- **User-centric queries:** Use `getByRole`, `getByLabelText`, `getByText`. Avoid `getByTestId` unless no semantic alternative.
- **Async testing:** Use `waitFor` or `findBy*` for async state changes. Never use `act()` directly unless wrapping non-RTL code.
- **Mock at boundaries:** Mock API calls (MSW), not internal hooks or components.

```tsx
render(<LoginForm onSubmit={mockSubmit} />);
await userEvent.type(screen.getByLabelText("Email"), "a@b.com");
await userEvent.click(screen.getByRole("button", { name: "Sign in" }));
expect(mockSubmit).toHaveBeenCalledWith({ email: "a@b.com" });
```

## Hook Testing Patterns
- Use `renderHook` from `@testing-library/react` for custom hook tests.
- Wrap with providers when hooks depend on context.
- Test state transitions by calling returned functions and asserting new values.

```tsx
const { result } = renderHook(() => useCounter(0));
act(() => result.current.increment());
expect(result.current.count).toBe(1);
```

## API Mocking with MSW
- Use MSW (Mock Service Worker) for API mocking. Define handlers at suite level.
- Share handlers via a `handlers.ts` file. Override per-test when needed.
- Assert on request payloads when testing mutation flows.

```typescript
const server = setupServer(
  http.get("/api/users/:id", ({ params }) => {
    return HttpResponse.json({ id: params.id, name: "Alice" });
  }),
);
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

## Accessibility Testing
- Use `jest-axe` or `vitest-axe` for automated a11y checks in component tests.
- Test keyboard navigation for custom interactive components (Tab, Enter, Escape).
- Verify focus management after modals, dialogs, and dynamic content changes.

```tsx
const { container } = render(<Navigation />);
const results = await axe(container);
expect(results).toHaveNoViolations();
```

## Error Boundary Testing
- Test that the fallback UI renders when a child component throws.
- Test the reset/retry flow via `resetErrorBoundary` to verify recovery.
- Test that `onError` fires with the error and component stack for logging.
