---
name: cross-cutting
description: Error taxonomy, coverage targets, dependency policy, performance budgets, and data lifecycle
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

## Testing Coverage Targets
- **Critical paths** (auth, payments, data mutations): 80%+ line coverage.
- **Utilities and shared libraries:** 90%+ line coverage.
- **Generated code** (protobuf, OpenAPI stubs): No coverage requirement.
- **Integration tests:** Cover all API endpoints and external service interactions. Not counted toward line coverage targets.

## Dependency Update Policy
- **Security patches:** Apply immediately. No waiting for a release cycle.
- **Minor versions:** Review and update monthly. Check changelogs for breaking behavior changes despite semver.
- **Major versions:** Evaluate breaking changes, plan migration, test in isolation before upgrading. Create a dedicated branch for major upgrades.
- **Audit:** Run dependency audit tools as part of CI. Block merges on critical vulnerabilities.

## Performance Budget Targets
- **API endpoints:** p95 response time < 200ms for reads, < 500ms for writes. Measure at the handler boundary, excluding network.
- **Database queries:** p95 < 50ms for indexed lookups, < 200ms for complex joins. Verify with query plan analysis.
- **Startup time:** Service healthy within 10s of container start. Measure from process start to first successful health check.
- **Web LCP:** Largest Contentful Paint < 2.5s on 4G connection. Measure with Lighthouse CI.
- **Web INP:** Interaction to Next Paint < 200ms. Profile with Chrome DevTools Performance panel.
- **Bundle size:** JavaScript bundle < 200KB gzipped for initial load. Use code splitting for routes.

## SBOM Requirement
- **Production deployments** must include a Software Bill of Materials (SBOM) in CycloneDX or SPDX format.
- Generate SBOM as part of the CI/CD build stage, before the deploy stage.
- Store SBOM artifacts alongside release artifacts and verify against known vulnerability databases before deploying to production.

## Service Level Objectives
- Define SLOs for availability and latency per service. SLOs are business decisions that drive alerting and observability implementation.

## Data Lifecycle
- Classify data by sensitivity: public, internal, confidential, restricted.
- Encrypt confidential and restricted data at rest and in transit.
- Define retention policies per data class -- never retain data indefinitely without justification.
- Anonymize or pseudonymize PII in non-production environments.
- Log access to restricted data for audit purposes.
