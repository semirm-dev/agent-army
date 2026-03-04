---
name: cross-cutting
description: Error taxonomy, dependency policy, and performance budgets
scope: universal
languages: []
---

# Cross-Cutting Standards

## Error Taxonomy
Categorize all errors into three levels:
- **Domain Errors:** Validation failures, not-found, conflict, business rule violations. These are expected and handled. Return appropriate 4xx status codes.
- **Infrastructure Errors:** Timeouts, connection failures, service unavailable. These are retryable. Log at WARN, return 503 with retry guidance.
- **System Errors:** Internal bugs, panic recovery, unhandled states. These are unexpected. Log at ERROR with full stack trace, return 500. Page on-call if in production.

## Dependency Update Policy
- **Security patches:** Apply immediately. No waiting for a release cycle.
- **Minor versions:** Review and update monthly. Check changelogs for breaking behavior changes despite semver.
- **Major versions:** Evaluate breaking changes, plan migration, test in isolation before upgrading. Create a dedicated branch for major upgrades.

## Performance Budget Targets
Default targets. Adjust per-project based on domain requirements.

### Backend
- **API endpoints:** p95 < 200ms reads, < 500ms writes. Measure at handler boundary.
- **Database queries:** p95 < 50ms indexed lookups, < 200ms complex joins. Verify with query plan analysis.
- **Startup time:** Service healthy within 10s of container start.

### Frontend
- **LCP:** < 2.5s on 4G. Measure with Lighthouse CI.
- **INP:** < 200ms. Profile with Chrome DevTools.
- **Bundle size:** < 200KB gzipped initial load. Code-split routes.
