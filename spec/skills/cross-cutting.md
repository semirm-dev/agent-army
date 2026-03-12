---
name: cross-cutting
description: Defines system-wide error classification (domain, infrastructure, system), dependency update policy, and backend/frontend performance budget targets
scope: universal
languages: []
---

# Cross-Cutting Standards

For error classification, see the `error-handling` skill. For performance budgets, see the `performance-audit` skill.

## Dependency Update Policy
- **Security patches:** Apply immediately. No waiting for a release cycle.
- **Minor versions:** Review and update monthly. Check changelogs for breaking behavior changes despite semver.
- **Major versions:** Evaluate breaking changes, plan migration, test in isolation before upgrading. Create a dedicated branch for major upgrades.

