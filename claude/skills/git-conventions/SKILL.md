---
name: git-conventions
description: Enforce branch naming, commit message format, and PR template generation following project Git workflow standards.
---

# Git Conventions Skill

## When to Use

Invoke this skill when:
- Creating a new branch
- Writing a commit message
- Creating a pull request
- Reviewing git history for convention compliance

## Branch Naming

Validate branch names follow the pattern: `<type>/<description>`

**Valid prefixes:** `feat/`, `fix/`, `refactor/`, `docs/`, `test/`, `chore/`

**Examples:**
- `feat/user-auth` — new feature
- `fix/login-redirect` — bug fix
- `refactor/payment-module` — code restructuring
- `docs/api-endpoints` — documentation
- `test/auth-integration` — test additions
- `chore/update-deps` — maintenance

**Validation command:**
```bash
branch=$(git branch --show-current)
if ! echo "$branch" | grep -qE '^(feat|fix|refactor|docs|test|chore)/'; then
  echo "ERROR: Branch '$branch' does not follow naming convention."
  echo "Expected: <type>/<description> where type is feat|fix|refactor|docs|test|chore"
fi
```

## Commit Message Format

Follow Conventional Commits:

```
type(scope): description

Body explaining WHY (not what).
```

**Rules:**
- Subject line under 50 characters
- Imperative mood ("add feature" not "added feature")
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `perf`
- Scope is optional but recommended for larger projects
- Body separated by blank line, explains motivation

**Examples:**
- `feat(auth): add JWT refresh token rotation`
- `fix(api): handle null response from payment gateway`
- `refactor(db): extract connection pooling into shared module`

## PR Template

When creating a PR, use this structure:

```markdown
## Summary
[1-3 sentences: what changed and why]

## Changes
- [Bullet list of key changes]

## Test Plan
- [ ] [How to verify each change]

## Breaking Changes
[List any breaking changes, or "None"]
```

## Workflow

1. Before creating a branch: validate the name against conventions
2. Before committing: validate the message format
3. Before creating a PR: generate description from commit history
4. When reviewing: check all commits follow conventions
