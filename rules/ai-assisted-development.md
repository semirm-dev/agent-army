---
name: ai-assisted-development
description: Safe practices for AI-generated code, verification patterns, and anti-patterns
scope: universal
languages: []
---

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
- **Missing verification.** Marking tasks done without running build/test. Evidence before assertions.

## Testing After AI Changes
- **Re-run full test suite** after any AI-generated change, even if the change looks trivial.
- **Add regression tests** for the specific behavior the AI modified. AI may silently alter edge cases.
- **Diff-test before and after.** If a refactoring should be behavior-preserving, add a characterization test first, then apply the AI change.
- **Check test quality.** AI-generated tests may assert implementation details or use tautological assertions (testing mocks, not behavior).
- **Verify edge cases independently.** AI tends to handle happy paths well but miss boundary conditions, off-by-one errors, and empty/null inputs.
