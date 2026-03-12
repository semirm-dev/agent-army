---
name: git-workflow
description: Standardizes Git branch naming, Conventional Commits message format, PR size and review conventions, merge/rebase strategy, and semantic version release tagging
scope: universal
languages: []
---

# Git Workflow Conventions
- **Branch Naming:** Use prefixes: `feat/`, `fix/`, `refactor/`, `docs/`, `test/`, `chore/`. Example: `feat/user-auth`, `fix/login-redirect`.
- **Commit Messages:** Follow Conventional Commits. Subject line under 50 characters, imperative mood. Blank line, then body explaining WHY (not what).
  - Format: `type(scope): description`
  - Types: `feat` (new feature), `fix` (bug fix), `docs` (documentation only), `refactor` (no behavior change), `test` (test-only changes), `chore` (build/tooling), `ci` (CI/CD pipeline changes), `perf` (performance improvement)
  - Example: `feat(auth): add JWT refresh token rotation`
  - Breaking changes: Add a `BREAKING CHANGE:` footer in the commit body describing what changed and migration steps.
- **PR Size:** Aim for <400 lines changed. Split large features into stacked PRs.
- **PR Description:** Include: summary (what + why), test plan, breaking changes (if any).
- **Merge Strategy:** Default to squash-and-merge for feature branches. Use merge commits for long-lived branches. Never force-push shared branches.
- **Commit Hygiene:** Each commit should compile and pass tests independently. No "WIP" or "fix typo" commits in PRs -- squash before merge.
- **Merge Conflicts:** The branch author resolves merge conflicts. Re-request review after conflict resolution if the resolution touched non-trivial logic.
- **Rebase Policy:** Rebase only local, unshared feature branches. Never force-push shared or published branches.
- **Branch Protection:** Require code review approval and passing CI before merge to main/trunk. No direct pushes to protected branches.
- **Release Tagging:** Use semantic versioning tags (`vMAJOR.MINOR.PATCH`). Tag from main after merge. Use annotated tags with a changelog summary in the tag message.

## Dependency Update Policy
- **Security patches:** Apply immediately. No waiting for a release cycle.
- **Minor versions:** Review and update monthly. Check changelogs for breaking behavior changes despite semver.
- **Major versions:** Evaluate breaking changes, plan migration, test in isolation before upgrading. Create a dedicated branch for major upgrades.
