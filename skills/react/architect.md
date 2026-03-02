---
name: react/architect
description: React project architecture — component hierarchy design, state management selection, routing strategy, project layout, feature decomposition, and evolution planning.
scope: language-specific
languages: [react]
uses_rules:
  - react/patterns
  - typescript/patterns
  - code-quality
  - cross-cutting
---

# React Architect Skill

## When to Use

Invoke this skill when:
- Starting a new React application or component library
- Restructuring an existing React project
- Planning component hierarchy for a feature
- Choosing state management approach
- Designing routing strategy
- Evaluating feature decomposition

## Project Layout Decision Tree

```
Is this a single-page application (SPA)?
  |
  +-- With routing?
  |     YES --> src/pages/ (route components), src/components/ (shared), src/hooks/, src/lib/
  |
  +-- Dashboard / admin panel?
  |     YES --> src/features/<feature>/ (feature-based), src/components/ (shared UI)
  |
  +-- Marketing / content site?
        YES --> Consider Next.js/Remix. src/app/ (routes), src/components/, src/lib/
  |
Is this a component library?
  |
  YES --> src/components/<Component>/ with co-located index, component, types, tests, stories
  |
Is this a micro-frontend?
  YES --> Self-contained with own routing, state, and API layer. Shared types via package.
```

### SPA Reference Layout

```
src/
  components/              (shared UI components)
    <Component>/
      <Component>.tsx
      <Component>.test.tsx
      <Component>.types.ts
      index.ts
  features/                (feature modules)
    <feature>/
      components/          (feature-specific components)
      hooks/               (feature-specific hooks)
      <feature>.types.ts
      index.ts
  hooks/                   (shared custom hooks)
  lib/                     (API clients, utilities)
  pages/                   (route-level components)
  types/                   (shared type definitions)
  App.tsx
  main.tsx
```

### Component Library Reference Layout

```
src/
  components/
    <Component>/
      <Component>.tsx
      <Component>.test.tsx
      <Component>.stories.tsx
      <Component>.types.ts
      index.ts
  hooks/                   (shared hooks)
  utils/                   (internal utilities)
  index.ts                 (public API — only export consumer-facing components)
```

## Component Hierarchy Decision Tree

```
New UI feature?
  |
  +-- Is it a full page/route?
  |     YES --> Page component in src/pages/
  |             Contains layout + feature composition, minimal logic
  |
  +-- Is it a reusable UI primitive (button, input, card)?
  |     YES --> src/components/<Component>/
  |             Generic props, no business logic, fully controlled
  |
  +-- Is it feature-specific (order form, user profile card)?
  |     YES --> src/features/<feature>/components/
  |             Can contain business logic, feature-scoped state
  |
  +-- Is it a layout component (sidebar, header, grid)?
        YES --> src/components/layout/
                Accepts children, controls spatial arrangement
```

## State Management Selection

```
What kind of state is this?
  |
  +-- Server data (API responses, cached entities)?
  |     YES --> TanStack Query (React Query)
  |             Handles loading, error, caching, refetch, optimistic updates
  |
  +-- Global client state (auth user, theme, sidebar open)?
  |     YES --> Zustand store
  |             Lightweight, no boilerplate, supports selectors
  |
  +-- URL-driven state (filters, pagination, sort, search)?
  |     YES --> URL search params (useSearchParams)
  |             Shareable, bookmarkable, survives refresh
  |
  +-- Local component state (form input, toggle, hover)?
  |     YES --> useState
  |             Simplest option, no sharing needed
  |
  +-- Derived from other state?
        YES --> Compute inline or useMemo
                NEVER useEffect to sync derived state
```

## Routing Strategy Decision Tree

```
Does the app need routing?
  |
  +-- Multi-page with deep linking?
  |     YES --> React Router or TanStack Router
  |             Nested layouts, route-level code splitting with React.lazy()
  |
  +-- File-based routing (Next.js/Remix)?
  |     YES --> Follow framework conventions (app/ directory)
  |
  +-- Protected routes?
  |     YES --> Auth guard wrapper component
  |             Redirect to login, check roles before render
  |
  +-- Route-level data loading?
        YES --> Loader pattern (TanStack Router, Remix) or route-level useQuery
```

## Feature Decomposition Workflow

```
New feature request?
  |
  +-- Identify data requirements
  |     What API endpoints? What server state?
  |     --> Define query hooks (useTanStackQuery wrappers)
  |
  +-- Identify UI components
  |     What does the user see? What interactions?
  |     --> Sketch component tree (page → sections → primitives)
  |
  +-- Identify shared vs feature-specific
  |     Used by other features?
  |       YES --> src/components/ or src/hooks/
  |       NO  --> src/features/<feature>/
  |
  +-- Identify state needs
  |     Apply State Management Selection tree above
  |
  +-- Define module boundary
        What does this feature export?
        --> Only page component + types (if consumed externally)
```

## Architecture Evolution Checklist

- [ ] Each feature directory has a single clear responsibility
- [ ] No circular imports between features (enforced by ESLint `import/no-cycle`)
- [ ] Shared components are genuinely reusable (no feature-specific logic)
- [ ] State management choice matches the state category (server/client/URL/local)
- [ ] Route-level code splitting in place for all routes
- [ ] Error boundaries at route level with fallback UI
- [ ] Accessibility considerations documented per component
- [ ] No prop drilling deeper than 2 levels
- [ ] Custom hooks extract reusable logic (each hook does one thing)
- [ ] New features can be added without modifying existing feature modules
