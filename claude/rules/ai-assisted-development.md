<!-- Sync: Must stay in sync with cursor/507-ai-dev.mdc -->

# 🤖 AI-Assisted Development Patterns

## Code Organization for AI Readability
- **Small files (< 300 lines).** One responsibility per file.
- **Clear, descriptive naming:** Avoid abbreviations. `getUserById` not `getUsr`.
- **Explicit types on all function signatures:** Helps AI understand intent.
- **Flat directory structures where possible:** Deep nesting confuses context.
- **Co-locate related code:** Handler + service + types in same package.

## Comment Patterns That Help AI
- **WHY comments:** Explain business decisions, not what the code does. `// We retry 3 times because the payment gateway has transient failures` not `// retry 3 times`.
- **Invariant documentation:** Document preconditions, postconditions, and constraints. `// INVARIANT: balance must never go negative`.
- **Boundary markers:** Mark where external systems connect. `// BOUNDARY: calls payment gateway API`.
- **TODO format:** Use `// TODO(scope): description` for trackable items.
- **Do NOT comment obvious code:** `i++ // increment i` wastes tokens.

## Test Patterns for AI Verification
- **Descriptive assertion messages:** `expect(result).toBe(expected, "user should be active after verification")`.
- **Deterministic tests:** No random data, no time-dependent assertions, no external dependencies.
- **Table-driven tests with named cases:** Each case name describes the scenario.
- **One assertion per test when possible:** Easier for AI to diagnose failures.
- **Edge cases as explicit test cases:** Not hidden in helper functions.

## Prompt-Friendly Code
- **Named constants over magic numbers:** `MAX_RETRY_ATTEMPTS = 3` not `3`.
- **Explicit error messages:** `"user not found: id=%s"` not `"not found"`.
- **No implicit behavior:** Avoid init() functions, module-level side effects.
- **Self-documenting function signatures:** Parameters tell the story.
- **Use enums/constants for state:** `OrderStatus.PENDING` not `"pending"`.

## AI-Safe Refactoring
- **Always run tests after AI-generated changes:** Never trust without verification.
- **Review AI-generated code diff-by-diff:** Not as a whole.
- **AI may introduce subtle bugs in edge cases:** Test boundary conditions specifically.
- **AI may over-engineer:** Check that changes match the requested scope.
- **AI may use deprecated APIs:** Verify library versions match project dependencies.

## Anti-Patterns
- **Blind trust:** Accepting AI suggestions without reading them.
- **Cargo-culting:** Copying AI patterns without understanding why.
- **Scope creep:** Letting AI add "improvements" beyond the requested change.
- **Stale context:** AI working from outdated file reads — verify files are current.
- **Missing verification:** Marking tasks done without running build/test.
