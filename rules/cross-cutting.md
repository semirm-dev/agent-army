---
name: cross-cutting
description: Error taxonomy, dependency policy, performance budgets, and data lifecycle
scope: universal
languages: []
---

# Cross-Cutting Standards

Standards that span multiple domains.

## Error Taxonomy
Categorize all errors into three levels:
- **Domain Errors:** Validation failures, not-found, conflict, business rule violations. These are expected and handled. Return appropriate 4xx status codes.
- **Infrastructure Errors:** Timeouts, connection failures, service unavailable. These are retryable. Log at WARN, return 503 with retry guidance.
- **System Errors:** Internal bugs, panic recovery, unhandled states. These are unexpected. Log at ERROR with full stack trace, return 500. Page on-call if in production.

## Dependency Update Policy
- **Security patches:** Apply immediately. No waiting for a release cycle.
- **Minor versions:** Review and update monthly. Check changelogs for breaking behavior changes despite semver.
- **Major versions:** Evaluate breaking changes, plan migration, test in isolation before upgrading. Create a dedicated branch for major upgrades.
- **Audit:** Run dependency audit tools as part of CI. Block merges on critical vulnerabilities.

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

## SBOM Requirement
- **Production deployments** must include a Software Bill of Materials (SBOM) in CycloneDX or SPDX format.
- Generate SBOM as part of the CI/CD build stage, before the deploy stage.
- Store SBOM artifacts alongside release artifacts and verify against known vulnerability databases before deploying to production.

## Data Lifecycle
- Classify data by sensitivity: public, internal, confidential, restricted. Apply controls proportional to classification.
- **Retention policies:** Define per data class. Automate enforcement -- scheduled jobs to purge or archive expired data. Never retain data indefinitely without justification.
- **PII handling:** Identify all fields containing personally identifiable information. Support deletion and anonymization requests (GDPR right-to-erasure, CCPA). Track PII across replicas, caches, backups, and logs.
- **Data deletion:** Soft-delete first (recoverable), hard-delete after retention window. Verify deletion propagates to derived stores (caches, search indexes, analytics pipelines).
- **Audit trail:** Log data access and mutations for confidential and restricted data. Include who, what, when, and from where.
