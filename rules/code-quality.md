---
name: Code Quality
description: Code clarity, structure, naming, comments, and linting standards
scope: universal
languages: []
---

# Code Quality Patterns

## Code Clarity
- **Small files (< 300 lines).** One responsibility per file.
- **Clear, descriptive naming.** Avoid abbreviations — names should be self-documenting.
- **Explicit types on all function signatures.** Parameters and return types tell the story.
- **Named constants over magic numbers.** Give every literal a meaningful name.
- **Enums or constants for state values.** Never represent state as raw strings or integers.
- **Explicit error messages.** Include contextual information that helps diagnose the failure.
- **No implicit behavior.** Avoid module-level side effects and hidden initialization.
- **Flat directory structures.** Co-locate related code — keep handler, service, and types together.

## Code Structure
- **Function length.** Functions should be < 30 lines. If longer, extract sub-functions.
- **Nesting depth.** Maximum 3 levels of nesting. Use early returns or extract functions to flatten.
- **DRY threshold.** Extract shared logic only when a pattern repeats 3+ times. Prefer 3 similar lines over a premature abstraction.
- **Single responsibility.** Each function does one thing. Each file has one purpose.
- **Parameter count.** Functions with > 3 parameters should accept an options or config object instead.

## Comments
- **WHY comments.** Explain business decisions, not what the code does.
- **Document invariants.** State preconditions, postconditions, and constraints explicitly.
- **TODO format.** Use `// TODO(scope): description` for trackable items.
- **No obvious comments.** Do not narrate what the code already says clearly.

## Linting & Formatting
- **Automate formatting.** Enforce consistent style via automated tools — no manual style debates.
- **Pre-commit or CI gates.** Run linter and formatter before code reaches the main branch.
- **Zero warnings in committed code.** Treat linter warnings as errors in CI.
