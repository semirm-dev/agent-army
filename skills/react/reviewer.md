---
name: react/reviewer
description: React code review workflow — review order, component anti-patterns, hook correctness, accessibility audit, performance red flags, and security checklist.
scope: language-specific
uses_rules:
  - react/patterns
  - react/testing
  - typescript/patterns
  - code-quality
  - cross-cutting
  - security
---

# React Reviewer Skill

## When to Use

Invoke this skill when:
- Reviewing React PRs
- Conducting code audits on React codebases
- Performing pre-merge checks on frontend code
- Evaluating component library changes

## Review Order

1. **Hooks** — rules of hooks, dependency arrays, custom hook design
2. **State** — correct state category (server/client/URL/local), no derived-state-in-effect
3. **Accessibility** — semantic HTML, ARIA, keyboard nav, focus management
4. **Performance** — unnecessary re-renders, bundle impact, code splitting
5. **Security** — XSS prevention, input sanitization, secret handling
6. **Style** — naming, file structure, prop types (defer to linter for formatting)

## React Anti-Pattern Checklist

- [ ] `useEffect` used to sync derived state (compute inline or `useMemo` instead)
- [ ] Missing or incorrect `useEffect` dependency array
- [ ] Conditional hook call (hooks must always be called in the same order)
- [ ] Inline object/array literal in JSX props (causes unnecessary re-renders)
- [ ] `useState` + `useEffect` for server data (use TanStack Query instead)
- [ ] Context used for frequently-changing state (use Zustand or local state)
- [ ] Prop drilling deeper than 2 levels (restructure with composition)
- [ ] Class component (use functional component + hooks)
- [ ] `index` used as list key (use stable unique ID)
- [ ] Direct DOM manipulation (`document.querySelector`) instead of refs
- [ ] Raw HTML injection without sanitization (use DOMPurify)

## Hook Correctness Checklist

- [ ] All hooks called at the top level of the component (not inside conditions, loops, or callbacks)
- [ ] `useEffect` dependency array includes all referenced values
- [ ] `useEffect` cleanup function handles subscriptions, timers, and AbortController
- [ ] `useMemo` / `useCallback` used only when necessary (measured perf issue or memoized child)
- [ ] Custom hooks return consistent shape (`{ data, isLoading, error }` or `[value, setter]`)
- [ ] `useRef` used for mutable values that don't trigger re-renders (timers, previous values)
- [ ] No stale closure bugs in event handlers or effects

## Accessibility Review

```
Semantic HTML used for interactive elements?
  NO  --> Flag: replace <div onClick> with <button>, <a>, etc.

All images have alt text?
  NO  --> Flag: add meaningful alt or alt="" + aria-hidden for decorative

Form inputs have associated labels?
  NO  --> Flag: add <label htmlFor> or wrap input in label

Dynamic content updates announced to screen readers?
  NO  --> Flag: add aria-live region for async updates

Focus managed on route change / modal open?
  NO  --> Flag: add focus management (autoFocus or programmatic focus)

Color used as sole indicator?
  YES --> Flag: pair with icon, text, or pattern
```

## Performance Red Flags

```
Large dependency added for small feature?
  YES --> Check bundle size, consider lighter alternative or tree-shaking

Component re-renders on every parent render?
  YES --> Check for inline objects/arrays in props
          Check if React.memo() is warranted (profile first)

List with >100 items rendered without virtualization?
  YES --> Use react-window or tanstack-virtual

Route-level component not lazy-loaded?
  YES --> Use React.lazy() + Suspense

Heavy computation in render path?
  YES --> Move to useMemo with correct dependencies

API calls triggered on every render?
  YES --> Missing dependency array in useEffect, or use TanStack Query
```

## Security Review Checklist

- [ ] No raw HTML injection with unsanitized user content (use DOMPurify if HTML is required)
- [ ] User input validated with schema library (Zod) at form/API boundary
- [ ] No secrets, API keys, or tokens in client-side code
- [ ] External links use `rel="noopener noreferrer"` with `target="_blank"`
- [ ] No dynamic code evaluation (eval, Function constructor) with user input
- [ ] Auth tokens stored in HTTP-only cookies (not localStorage)
- [ ] CORS configured with specific origins on the backend
- [ ] Dependencies audited (`npm audit`)
