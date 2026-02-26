<!-- Sync: Must stay in sync with cursor/502-cross-cutting.mdc -->

# 🏗️ Cross-Cutting Standards

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
- **Audit:** Run `go mod verify` / `npm audit` / `pip audit` as part of CI. Block merges on critical vulnerabilities.

## Settings Notes
- **`skipDangerousModePermissionPrompt: true`** is intentionally enabled in `settings.json`. This skips the confirmation dialog when switching to dangerous/unrestricted mode. Rationale: the safety constraints in this file (no rm-rf, no auto-commit, deletion limits) provide guardrails at the rule level, and the plan-first default mode adds an additional gate. The prompt was adding friction to legitimate mode switches without meaningful safety benefit given the existing constraints. If you prefer the extra gate, set this to `false`.

## Performance Budget Targets
- **API endpoints:** p95 response time < 200ms for reads, < 500ms for writes. Measure at the handler boundary (excluding network).
- **Web pages (LCP):** Largest Contentful Paint < 2.5s on 4G connection. Measure with Lighthouse CI.
- **Bundle size:** JavaScript bundle < 200KB gzipped for initial load. Use code splitting for routes.
- **Database queries:** p95 < 50ms for indexed lookups, < 200ms for complex joins. Use `EXPLAIN ANALYZE` to verify.
- **Startup time:** Service healthy within 10s of container start. Measure from `docker run` to first successful health check.

## SBOM Requirement
- **Production deployments** must include a Software Bill of Materials (SBOM) in CycloneDX or SPDX format.
- Generate SBOM as part of the CI/CD build stage, before the deploy stage.
- Store SBOM artifacts alongside release artifacts (container registry, release page).
- Verify SBOM against known vulnerability databases before deploying to production.
