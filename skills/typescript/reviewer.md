---
name: typescript/reviewer
description: TypeScript code review workflow — review order, type safety audit, async verification, performance red flags, security checklist, and anti-pattern detection.
scope: language-specific
uses_rules:
  - typescript/patterns
  - typescript/testing
  - code-quality
  - cross-cutting
  - security
---

# TypeScript Reviewer Skill

## When to Use

Invoke this skill when:
- Reviewing TypeScript PRs
- Conducting code audits on TypeScript/JavaScript codebases
- Performing pre-merge checks on TypeScript/JavaScript code
- Evaluating npm package choices

## Review Order

1. **Type Safety** — no `any`, explicit return types on exports, discriminated unions over type assertions
2. **Structure** — module boundaries, barrel files limited to one level, no circular imports
3. **Correctness** — logic, error handling, null/undefined checks, exhaustive switch/if
4. **Async** — proper await, no floating promises, error handling in async paths
5. **Performance** — bundle impact, unnecessary re-renders (React), O(n²) patterns
6. **Security** — input validation at boundaries, XSS prevention, secret handling
7. **Style** — naming, formatting (defer to Prettier/Biome)

## TypeScript Anti-Pattern Checklist

- [ ] `any` type used without documented justification
- [ ] Non-null assertion (`!`) without documented justification
- [ ] `enum` used instead of `as const` object + union type
- [ ] Type assertion (`as`) used to bypass type checking
- [ ] `@ts-ignore` / `@ts-expect-error` without explanation
- [ ] Floating promise (async call without `await` or `.catch()`)
- [ ] `catch(e)` without narrowing `unknown` type before use
- [ ] Default export (prefer named exports)
- [ ] Barrel file re-exporting from deeply nested paths
- [ ] `process.env` accessed directly in business logic (not via config module)

## React-Specific Review (if applicable)

- [ ] Hooks follow rules of hooks (no conditional hooks)
- [ ] `useEffect` dependencies array is complete and correct
- [ ] Expensive computations wrapped in `useMemo`
- [ ] Event handlers wrapped in `useCallback` when passed as props
- [ ] Component files under 200 lines (extract sub-components)
- [ ] No inline object/array literals in JSX props (causes re-renders)

## Performance Red Flags

```
Large dependency added for a small feature?
  YES --> Check bundle size impact (bundle analyzer)

Unoptimized re-renders in React (missing memo/useMemo/useCallback)?
  YES --> Add memoization for expensive components and callbacks

Array.find() or Array.filter() inside a loop?
  YES --> Use Map/Set for O(1) lookups

Dynamic import() missing for route-level code splitting?
  YES --> Use React.lazy() or dynamic import for routes

Large JSON parsed synchronously on main thread?
  YES --> Consider streaming or worker-based parsing
```

## Security Review Checklist

- [ ] User input validated with schema library (Zod/Valibot) at API boundary
- [ ] No raw HTML injection without sanitization
- [ ] No dynamic code evaluation with user input
- [ ] Secrets not in client-side code or committed to repo
- [ ] HTTP clients have timeouts
- [ ] CORS configured with specific origins (not `*` in production)
- [ ] Dependencies audited (`npm audit`)
