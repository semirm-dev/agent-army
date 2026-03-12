---
name: react/testing
description: Guides writing React component and hook tests using Testing Library user-centric queries, MSW network-level API mocking, and automated accessibility checks with jest-axe/vitest-axe.
scope: language-specific
languages: [react]
uses_skills: [react/patterns, testing]
---

# React Testing Patterns

## Component Testing
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

## Hook Testing
- Use `renderHook` from `@testing-library/react` for custom hook tests.
- Wrap with providers when hooks depend on context.
- Test state transitions by calling returned functions and asserting new values.

```tsx
const { result } = renderHook(() => useCounter(0));
act(() => result.current.increment());
expect(result.current.count).toBe(1);
```

## API Mocking
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
