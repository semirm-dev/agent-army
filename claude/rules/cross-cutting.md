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
