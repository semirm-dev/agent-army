---
scope: universal
languages: []
extends: [code-quality]
---

> Extends `code-quality.md`. Code clarity and naming standards apply.

# AI-Assisted Development Patterns

## AI-Safe Refactoring
- **Always run tests after AI-generated changes.** Never trust without verification.
- **Review AI-generated code diff-by-diff.** Not as a whole. Each diff should make sense independently.
- **AI may introduce subtle bugs in edge cases.** Test boundary conditions specifically after AI changes.
- **AI may over-engineer.** Check that changes match the requested scope — no extra abstractions, no bonus features.
- **AI may use deprecated APIs.** Verify library versions match project dependencies before accepting suggestions.
- **Verify files are current before editing.** AI working from outdated file reads produces invalid diffs.

## Anti-Patterns
- **Blind trust.** Accepting AI suggestions without reading them. Every line must be understood.
- **Cargo-culting.** Copying AI patterns without understanding why. If you can't explain the code, don't commit it.
- **Scope creep.** Letting AI add "improvements" beyond the requested change. A bug fix is not an invitation to refactor.
- **Stale context.** AI working from outdated file reads — verify files are current before generating changes.
- **Missing verification.** Marking tasks done without running build/test. Evidence before assertions.

## Test Patterns for AI Verification
- **Descriptive assertion messages.** Include context that explains what was expected and why.
- **Deterministic tests.** No random data, no time-dependent assertions, no external dependencies in unit tests.
- **Table-driven tests with named cases.** Each case name describes the scenario being tested.
- **One assertion per test when practical.** Easier to diagnose which behavior broke.
- **Edge cases as explicit test cases.** Not hidden in helper functions or shared fixtures.

## Cross-References
- See `code-quality.md` for naming, constants, error messages, and implicit behavior rules.
- See `testing-patterns.md` for universal testing patterns referenced in "Test Patterns for AI Verification" above.
