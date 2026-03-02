---
name: react/coder
description: React implementation workflow — component placement, hook design, state management, accessibility, form handling, and pre-PR verification.
scope: language-specific
uses_rules:
  - react/patterns
  - typescript/patterns
  - code-quality
  - cross-cutting
  - security
---

# React Coder Skill

## When to Use

Invoke this skill when:
- Writing new React components
- Implementing custom hooks
- Building forms with validation
- Adding data fetching to components
- Implementing accessibility features
- Creating interactive UI flows

## Implementation Workflow

```
Understand requirements
  |
  +-- Check existing components and hooks in codebase
  |     (avoid duplicating shared UI primitives)
  |
  +-- Design component API first (props interface)
  |     (props drive implementation; avoid retrofitting)
  |
  +-- Choose state management (see decision tree below)
  |
  +-- Implement component
  |
  +-- Write tests (testing-library + MSW)
  |
  +-- Verify accessibility
  |
  +-- PR
```

**Emphasize "props interface first"** — define the component contract before writing JSX.

## Component Placement Decision Tree

```
Where does this component belong?
  |
  +-- Reusable UI primitive (button, modal, card)?
  |     YES --> src/components/<Component>/
  |             Generic props, no business logic, fully controlled
  |
  +-- Feature-specific component?
  |     YES --> src/features/<feature>/components/
  |             Can use feature-specific hooks and state
  |
  +-- Route/page-level component?
  |     YES --> src/pages/<PageName>.tsx
  |             Composes feature components, handles layout
  |
  +-- Layout component (sidebar, header, footer)?
  |     YES --> src/components/layout/
  |             Accepts children, controls spatial arrangement
  |
  +-- Custom hook?
  |     --> Shared: src/hooks/use<Name>.ts
  |     --> Feature-specific: src/features/<feature>/hooks/
  |
  +-- Type definitions?
        --> Co-locate in <Component>.types.ts
            Shared types in src/types/
```

## Hook Design Decisions

```
Should this be a custom hook?
  |
  +-- Logic reused across 2+ components?
  |     YES --> Extract to custom hook
  |
  +-- Component has complex state transitions?
  |     YES --> Extract to useReducer or custom hook
  |
  +-- Data fetching logic?
  |     YES --> Wrap TanStack Query in a domain hook (useUser, useOrders)
  |
  +-- Side effect management (timers, subscriptions)?
  |     YES --> Extract to custom hook with proper cleanup
  |
  +-- Simple local state (toggle, counter)?
        NO  --> Keep useState inline in component

Hook rules:
  - Always prefix with "use"
  - One responsibility per hook
  - Return object for 3+ values: { data, isLoading, error }
  - Return tuple for 2 values: [value, setValue]
  - Never call hooks conditionally
  - Document dependencies in useEffect with WHY comments
```

## State Management Selection

```
What state am I managing?
  |
  +-- API data?
  |     --> TanStack Query hook
  |         const { data, isLoading, error } = useQuery({ queryKey, queryFn })
  |
  +-- Global UI state (theme, sidebar, notifications)?
  |     --> Zustand store with selectors
  |         const theme = useThemeStore((s) => s.theme)
  |
  +-- URL-driven (filters, pagination, tabs)?
  |     --> useSearchParams
  |         Shareable, bookmarkable, survives refresh
  |
  +-- Form state?
  |     --> Simple (<3 fields): useState per field
  |     --> Complex: React Hook Form + Zod schema
  |
  +-- Derived from other state?
  |     --> Compute inline or useMemo
  |         NEVER sync with useEffect
  |
  +-- Local UI state (hover, open/close, animation)?
        --> useState in the component
```

## Accessibility Checklist

- [ ] Interactive elements use semantic HTML (`<button>`, `<a>`, `<input>`)
- [ ] Custom interactive elements have `role`, `tabIndex`, `onKeyDown` handlers
- [ ] All images have meaningful `alt` text (decorative: `alt=""` + `aria-hidden="true"`)
- [ ] Form inputs have associated `<label>` elements (via `htmlFor` or wrapping)
- [ ] Error states use `aria-invalid` and `aria-describedby` for help text
- [ ] Dynamic content updates use `aria-live` regions
- [ ] Focus managed on route changes, modal open/close
- [ ] Color is not the sole indicator of state (pair with icons or text)
- [ ] Contrast ratio >= 4.5:1 for normal text, >= 3:1 for large text
- [ ] Keyboard navigation works (Tab, Enter, Escape, Arrow keys where appropriate)

## Dependency Addition Checklist

- [ ] Checked if existing components/hooks cover the need
- [ ] Checked bundle size impact (bundlephobia.com)
- [ ] Verified TypeScript types included or `@types/` package available
- [ ] Tree-shakeable (ESM exports)
- [ ] No duplicate functionality with existing dependencies
- [ ] Compatible with React version in project
- [ ] Actively maintained (check last publish date, open issues)

## Pre-PR Checklist

- [ ] `tsc --noEmit` clean (no type errors)
- [ ] ESLint clean (including React-specific rules)
- [ ] Tests pass with `@testing-library/react`
- [ ] Components render without console warnings
- [ ] Loading, error, and empty states handled
- [ ] Accessibility checklist passed (above)
- [ ] No inline object/array literals in JSX props
- [ ] No `useEffect` for derived state
- [ ] Bundle size impact checked (route-level code splitting where needed)
- [ ] No `any` types without documented justification
