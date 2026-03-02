---
scope: language-specific
languages: [react]
extends: [code-quality, typescript/patterns]
---

> Extends `code-quality.md`. React components use TypeScript; see `typescript/patterns.md` for TS-specific standards.

# React & Frontend Patterns

## Component Structure
- **One component per file.** Co-locate styles, tests, and types with the component.
- **File naming:** `ComponentName.tsx` for components, `useHookName.ts` for hooks, `ComponentName.test.tsx` for tests.
- **Functional components only.** No class components. Use hooks for all state and lifecycle.
- **Props:** Destructure in function signature. Define prop types as a named interface above the component.

## State Management
- **Server state:** Preferred: TanStack Query (React Query) for server data. Follow project conventions if alternatives are already in use. Never manually manage loading/error/data states for API calls.
- **Client state:** Preferred: Zustand for global client state. Follow project conventions if alternatives are already in use. Use `useState` for local component state. Use Context only for truly global, rarely-changing values (theme, locale, auth user).
- **Derived state:** Compute from existing state inline or with `useMemo`. Never use `useEffect` to sync derived state — this is the most common React anti-pattern.
- **URL state:** Use URL search params for filterable/shareable UI state (filters, pagination, sort order).

## Data Fetching
- **Hooks pattern:** Every data fetch should return `{ data, isLoading, error }`.
- **Loading states:** Show skeletons or placeholders, not spinners. Avoid layout shift.
- **Error states:** Show actionable error messages. Include retry buttons.
- **Optimistic updates:** Use for low-risk mutations (toggle, like). Rollback on failure.

## Component Composition
- **Composition over prop drilling.** Use children, render props, or compound component pattern.
- **Context:** Only for truly global state (auth, theme, locale). Never for prop drilling avoidance — use composition instead.
- **Custom hooks:** Extract reusable logic into `use*` hooks. Each hook should do one thing.
- **Avoid prop drilling >2 levels.** Restructure components or use composition pattern.

## Performance
- **Avoid premature optimization.** Profile first with React DevTools Profiler.
- **`useMemo` / `useCallback`:** Only use when there's a measured performance issue or when passing to memoized children.
- **Lazy loading:** Use `React.lazy()` + `Suspense` for route-level code splitting.
- **List rendering:** Use stable, unique keys (never array index). Virtualize long lists (>100 items) with `react-window` or `tanstack-virtual`.

- **Accessibility:** Target WCAG 2.1 AA compliance.

## Error Boundaries
- **Use `react-error-boundary` library** (not class components). Maintains "functional components only" rule.
- **Wrap at route level:** Each route should have an error boundary with a fallback UI.
- **Fallback UI:** Show actionable error messages with a "Try again" button using `resetErrorBoundary`.
- **Logging:** Use `onError` prop to report errors to your logging service.

## Cross-References
> See `code-quality.md` for universal code quality standards.
> See `typescript/patterns.md` for TypeScript-specific standards used by React components.
> See `cross-cutting.md` for error taxonomy, coverage targets, and performance budget targets (LCP, bundle size, INP).
> See `security.md` for secrets management, input validation, and injection prevention.
> See `testing-patterns.md` for universal testing standards.

