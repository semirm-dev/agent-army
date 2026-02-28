---
name: code-review-workflow
description: "Structured code review process — multi-pass review strategy, priority checklist, PR sizing guidelines, splitting strategies, review comment format, and approval criteria."
scope: universal
---

# Code Review Workflow Skill

## When to Use

Invoke this skill when:
- Reviewing a pull request (as reviewer)
- Preparing code for review (as author)
- Establishing review standards for a team
- Training new reviewers on review priorities and process

> See `rules/git-workflow.md` for branch naming, commit message format, and merge strategy.
> See `rules/security.md` for security review checklist items.

## Multi-Pass Review Strategy

Review in five ordered passes. Do not skip ahead. Each pass has a clear focus and a time budget.

### Pass 1: Understand Intent (2 min)

Before reading any code:

1. Read the PR description -- what changed and why
2. Check if the scope matches the stated intent (no scope creep)
3. Verify the PR title follows Conventional Commits format
4. Look at the file list -- do the changed files align with the described feature/fix?

```
Does the PR description explain WHY?
  NO  --> Request a description update before reviewing code
  YES --> Does the scope match the description?
    NO  --> Comment: scope exceeds stated intent, suggest splitting
    YES --> Proceed to Pass 2
```

### Pass 2: Correctness (main pass)

This is where most review time is spent. Focus on:

- **Logic errors:** Does the code do what the description says it does?
- **Edge cases:** Empty inputs, zero values, nil/null/undefined, boundary values
- **Off-by-one errors:** Loop bounds, slice ranges, pagination cursors
- **Error handling:** Are all error paths handled? Are errors wrapped with context?
- **Null/undefined paths:** Can any variable be null when the code assumes it is not?
- **Race conditions:** Shared mutable state accessed from multiple goroutines/threads/tasks?
- **Resource leaks:** Are connections, file handles, and channels properly closed?

### Pass 3: Security and Safety

Scan every changed file for:

- **Input validation:** Is all external input validated at the handler boundary?
- **SQL injection:** Are queries parameterized? Any string concatenation with user input?
- **XSS:** Is user content escaped before rendering in HTML contexts?
- **Auth checks:** Does every endpoint verify authentication and authorization?
- **Secrets exposure:** Are API keys, tokens, or credentials hardcoded or logged?
- **Data leaks:** Are internal fields (passwords, tokens, PII) excluded from API responses?
- **Dependency safety:** Do new dependencies have known vulnerabilities?

### Pass 4: Design and Maintainability

Evaluate structural quality:

- **Naming:** Are variables, functions, and types named clearly and consistently?
- **Abstractions:** Is the abstraction level appropriate? Not too early, not too duplicated (rule of three).
- **Coupling:** Do changes require touching many unrelated files? Could the boundary be cleaner?
- **Single responsibility:** Does each function/class/module do one thing?
- **Test coverage:** Are the changed code paths covered by tests? Are edge cases tested?
- **Error taxonomy:** Do errors follow the domain/infrastructure/system classification?

### Pass 5: Nits (lowest priority)

Only after Passes 1-4 are clean:

- Formatting inconsistencies not caught by linters
- Typos in comments or variable names
- Minor style preferences
- Import ordering

If there are blocking issues from Passes 2-4, skip nits entirely. Do not mix blocking feedback with style suggestions.

## Review Priority Order

```
Bugs > Security > Performance > Design > Style
```

Decision tree for each comment:

```
Is this a correctness or security issue?
  YES --> "bug:" or "security:" prefix, request changes (blocking)
  NO  ↓

Could this cause a production incident (performance, data loss)?
  YES --> "perf:" prefix, request changes (blocking)
  NO  ↓

Does this hurt long-term maintainability?
  YES --> "design:" prefix, comment with suggestion (non-blocking)
  NO  ↓

Is this a style/formatting issue?
  YES --> "nit:" prefix, optional (non-blocking)
  NO  --> "question:" prefix, ask for clarification
```

## Per-Language Red Flags

Universal red flags to watch for:

| Category | What to Check |
|----------|---------------|
| Error handling | Are all errors checked and wrapped with context? |
| Type safety | Are type assertions safe? Any use of unsafe escape hatches? |
| Null safety | Can any variable be null when the code assumes it is not? |
| Concurrency | Is shared mutable state properly protected? |
| Resource management | Are connections, handles, and channels properly closed? |
| Global state | Is there mutable global/package-level state? |
| Dependencies | Are new dependencies necessary and vetted? |

## PR Size Guidelines

> See `rules/git-workflow.md` for PR size targets and merge strategy.

### Splitting Strategies

When a PR exceeds 400 lines, suggest one of these splits:

- **By layer:** Separate database migration, backend logic, and frontend changes into stacked PRs
- **By feature:** If the PR touches multiple features, split into one PR per feature
- **Refactor-then-feature:** Extract refactoring into a standalone PR, then build the feature on top
- **Interface-first:** PR 1 defines types/interfaces, PR 2 implements them, PR 3 adds tests and integration

```
PR > 400 lines?
  NO  --> Review normally
  YES ↓

Is the size justified? (migration + schema, single large file refactor)
  YES --> Note the justification, review with extra care
  NO  ↓

Can it split by layer (DB / backend / frontend)?
  YES --> Request split by layer
  NO  ↓

Can it split by feature (multiple unrelated changes)?
  YES --> Request split by feature
  NO  ↓

Is there a refactoring mixed with new behavior?
  YES --> Request refactor-then-feature split
  NO  --> Request author to find a natural seam
```

## Review Comment Format

Prefix every review comment with a category tag. Include a suggestion with every criticism.

### Prefixes

| Prefix | Meaning | Blocking? |
|--------|---------|-----------|
| `bug:` | Correctness issue, logic error | Yes |
| `security:` | Security vulnerability or risk | Yes |
| `perf:` | Performance issue that could degrade production | Yes |
| `design:` | Structural or maintainability concern | No (unless severe) |
| `nit:` | Style, formatting, minor preference | No |
| `question:` | Clarification needed, not a request for change | No |

### Examples

- **bug:** "Off-by-one in pagination. When `offset` equals `total`, returns empty page instead of stopping. Check boundary condition."
- **security:** "User ID from request body not verified against authenticated user. Validate ownership before proceeding."
- **perf:** "Query inside loop causes N+1 calls. Batch the IDs and query once."
- **design:** "Function handles validation, payment, and notification in 120 lines. Split into focused functions."
- **nit:** "Inconsistent naming -- `userData` here but `userInfo` on line 45."
- **question:** "Why retry 5 times instead of standard 3? Known upstream issue?"

## Approval Criteria Checklist

Before approving a PR, verify all items:

1. [ ] **Tests pass:** CI is green. No skipped or flaky tests.
2. [ ] **No bugs:** No logic errors, off-by-one, or unhandled edge cases found in Pass 2.
3. [ ] **No security issues:** No vulnerabilities found in Pass 3. Input is validated. Auth is enforced.
4. [ ] **Error handling complete:** All error paths are handled. Errors are wrapped with context. Error taxonomy is followed.
5. [ ] **Scope matches description:** The PR does what it says. No unrelated changes bundled in.
6. [ ] **Breaking changes documented:** If the PR changes public APIs or DB schemas, breaking changes are noted in the PR description and commit message.
7. [ ] **Test coverage adequate:** Changed code paths have tests. Critical paths meet 80%+ coverage target. New edge cases are covered.
8. [ ] **No obvious performance regressions:** No N+1 queries, unbounded loops, or missing indexes on new queries.

If any item fails, request changes with a specific, actionable comment using the prefix format above.
