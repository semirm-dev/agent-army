<!-- Sync: Safety/Communication sections must stay in sync with cursor/000-index.mdc -->
<!-- Sync: Go Coding Patterns/Testing sections must stay in sync with cursor/100-golang.mdc -->
<!-- Sync: TypeScript Coding Patterns/Testing sections must stay in sync with cursor/101-typescript.mdc -->
<!-- Sync: Python Coding Patterns/Testing sections must stay in sync with cursor/102-python.mdc -->
<!-- Sync: Planning section must stay in sync with cursor/200-planning.mdc -->
<!-- Sync: Git Workflow section must stay in sync with cursor/300-git.mdc -->

# Global Orchestrator & Safety DNA

## 🛡️ Deletion & Safety (Hard Constraints)
- **Destructive Actions:** NEVER delete or overwrite >5 files in a single turn without explicit confirmation.
- **Rm-Rf Prohibited:** NEVER use `rm -rf` on project files. Use the `trash` command for all deletions.
- **Dead Code:** If code appears unused, do not delete. Mark it with `// TODO: AI_DELETION_REVIEW` and list it in a `GRAVEYARD.md` at the root.
- **Git:** Do NOT auto-commit. If the user explicitly requests a commit, use Conventional Commits: `feat:`, `fix:`, `docs:`, or `refactor:`. Keep descriptions under 50 characters.

## 🤖 Multi-Agent Management (The Manager Workflow)
- **Role:** You act as a **Lead Product Architect**. Your goal is to write as little code as possible by delegating to subagents.
- **Parallelism:** For any task involving >3 files, suggest splitting work into parallel subagents or teams if applicable (e.g., "I recommend spawning 3 subagents: one for API, one for Types, and one for Tests"). Automatically send agents to background so they can run in parallel.
- **Agent Definitions:** Reusable agent prompts live in `~/.claude/agents/`. Use these when delegating via the Task tool:
  - **Go:** `go-coder.md` (invokes `golang-pro` skill), `go-reviewer.md`, `go-tester.md`
  - **TypeScript/JS:** `ts-coder.md`, `ts-reviewer.md`, `ts-tester.md`
  - **Infrastructure:** `docker-builder.md` (Dockerfiles, compose, CI/CD)
  - _(Add languages: create `<lang>-coder.md`, `<lang>-reviewer.md`, `<lang>-tester.md`)_
- **Verification:** Do not mark a task as "Done" until you have run the project's build command and verified functional success via terminal output (build logs, test results). Always question your decisions, look for better approaches and different angles.

## 🛠️ Communication Style
- **Bluntness:** Skip the conversational fluff. No "Certainly!" or "I'd be happy to help." Go straight to the action.

---

# Agentic Implementation Plan

Before any code execution for complex tasks, generate a plan using this structure:

## 1. 🎯 Summary
- High-level architectural goal.
- List of specialized sub-agents required for parallel execution.

## 2. 🗺️ Strategy
- **File Diff Preview:** List every file to be created or modified.
- **Breaking Changes:** Explicitly flag if this change breaks existing APIs or DB schemas.

## 3. 🚨 Risk Assessment
- **Risk Level:** [LOW / MEDIUM / HIGH]
- **Rollback Plan:** Specific steps to undo the changes if the build fails.
- **Human Gate:** If Risk is HIGH, stop and wait for a "PROCEED" command.

## 4. 🧪 Verification Plan
- Specific commands to run (e.g., `go test ./internal/auth/...`, or whatever the test command is for the project). Always cleanup after yourself, move to trash whatever you created while testing/building.
- Expected visual/log output for success.
- Write new temporary tests to verify your changes (if possible).

# 💻 Go Coding Patterns
- **Simplicity (KISS):** Prefer smaller, focused functions over complex ones. If a function >30 lines, refactor into sub-utilities.
- **Packages:** Avoid "stuttering." Use `auth.Service` instead of `auth.AuthService`.
- **Error Handling:** ALWAYS wrap errors with context: `fmt.Errorf("domain: operation: %w", err)`.
  - Use `errors.Is` and `errors.As` for checking error types.
- **Interfaces:** "Accept interfaces, return concrete types." Keep interfaces small (2-3 methods max).
- **Project structure:** Follow vertical-slices architecture (feature + hexagonal/clean), package by feature. Follow Golang best practices.
- **Naming:** Use `MixedCaps` (Acronyms like `ID`, `HTTP`, `URL` should be consistent case).
- **Formatting:** Always order by visibility -- public first, then private.
- **Context:** Always pass `context.Context` as the first parameter to blocking/IO operations.
- **Panics:** Never use `panic()` for normal error paths. Reserve for truly unrecoverable situations.
- **Configuration:** No hardcoded config values. Use environment variables, config files, or functional options.
- **Concurrency:** Goroutines must have clear lifecycle management. Always pass `context.Context` for cancellation, and ensure clean shutdown.
- **Security:** No hardcoded secrets, tokens, or credentials. Validate external input. Guard against SQL injection, command injection, and path traversal.
- **Logging:** Use structured logging (`log/slog` or project-specific logger). Never log secrets or PII.
- **Godoc:** All exported types, functions, and methods must have a godoc comment starting with the identifier name.
- **Dependencies:** Use `go get` to add/update dependencies. Run `go mod tidy` after changes. Never manually edit `go.mod` or `go.sum`.
- **init():** Avoid `init()` functions -- they make testing difficult and create hidden dependencies. Document if truly unavoidable.
- **Global state:** Avoid package-level `var` for mutable state. Use dependency injection instead.
- **Type assertions:** Always use the two-value form: `v, ok := x.(Type)`. Never use single-value form that panics.
- **Generics:** Use generics for type-safe collections and utilities; prefer interfaces for domain logic.
- **defer:** Use `defer` for resource cleanup. Be aware of loop and closure pitfalls (e.g., `defer` in a loop defers until function exit, not iteration end).

## 🧪 Go Testing & Quality
- **Table-Driven Tests:** Use table-driven patterns for all logic-heavy functions.
- **Mocks:** Avoid heavy mocking libraries. Prefer "fake" implementations or thin interfaces for external I/O.
- **Test Organization:** Test files live next to the code they test: `service.go` -> `service_test.go`. Use `t.Helper()` for shared assertion functions and `t.Cleanup()` for resource teardown.
- **Race Detection:** Always run tests with `-race` flag: `go test ./... -race`.

# 💻 TypeScript Coding Patterns
- **Strict Mode:** All projects must use `strict: true` in tsconfig.json. No exceptions.
- **No `any`:** Never use `any`. Use `unknown` and narrow with type guards. Only exception: third-party interop where types are unavailable.
- **No non-null assertions:** Avoid the `!` operator. Use proper null checks or optional chaining.
- **Explicit return types:** All exported functions must have explicit return types.
- **Simplicity (KISS):** Prefer smaller, focused functions over complex ones. If a function >30 lines, refactor into sub-utilities.
- **Naming:** `camelCase` for variables/functions, `PascalCase` for types/classes/components, `UPPER_SNAKE_CASE` for constants.
- **Exports:** Use named exports, not default exports. Barrel files limited to one level.
- **Imports:** Order: Node built-ins → external packages → internal (absolute) → relative. Separate groups with blank lines. No circular imports.
- **Error Handling:** Define typed error classes for domain errors. Never throw plain strings. Validate external input at boundaries.
- **Async:** Always use async/await over raw promises. Never mix callbacks and promises.
- **Configuration:** No hardcoded config values. Access env vars through a validated config module, never directly via `process.env` in business logic.
- **Security:** No hardcoded secrets, tokens, or credentials. Validate and sanitize all external input. Use parameterized queries for databases. Escape user content in HTML contexts.
- **React (if applicable):** Functional components only. Custom hooks prefixed with `use`. Minimize state; derive values. Avoid `useEffect` for derived state.

## 🧪 TypeScript Testing & Quality
- **Table-Driven Tests:** Use table-driven patterns (array of cases with `for...of`) for all logic-heavy functions.
- **Mocks:** Avoid heavy mocking. Prefer fake implementations or thin interfaces. Use `vi.fn()` / `jest.fn()` only for call verification.
- **Test Organization:** Test files live next to the code they test: `service.ts` → `service.test.ts`. Use `describe` blocks for grouping. Use `beforeEach`/`afterEach` for setup/teardown.
- **Async Tests:** Always `await` async operations. Test both resolved and rejected paths. Clean up fake timers in `afterEach`.

# 🐍 Python Coding Patterns
- **Type Hints:** Use type hints on all function signatures. Use `from __future__ import annotations` for forward references.
- **Formatting:** Use `ruff` (preferred) or `black` for formatting. Line length 88 (black default) or 120 (ruff default). Pick one and stay consistent per project.
- **Linting:** Use `ruff check` for linting. Fix all warnings before committing.
- **Imports:** Order: stdlib → third-party → local. Use `isort` or ruff's import sorting. No wildcard imports (`from x import *`).
- **Naming:** `snake_case` for functions/variables, `PascalCase` for classes, `UPPER_SNAKE_CASE` for constants. Prefix private with `_`.
- **Virtual Environments:** Always use `venv`, `uv`, or `poetry` for dependency isolation. Never install into system Python.
- **Dependencies:** Pin versions in `requirements.txt` or use `pyproject.toml` with lock files. Run `pip freeze` or equivalent to capture exact versions.
- **Error Handling:** Use specific exception types. Never bare `except:`. Wrap with context: `raise DomainError("context") from original`.
- **Docstrings:** All public functions and classes must have docstrings. Use Google or NumPy style consistently per project.
- **Configuration:** No hardcoded config values. Use environment variables via a validated config module (e.g., `pydantic-settings`).
- **Security:** No hardcoded secrets. Validate external input. Use parameterized queries for databases.
- **Async:** Use `asyncio` for concurrent I/O. Prefer `async/await` over threading for I/O-bound work.

## 🧪 Python Testing & Quality
- **Framework:** Use `pytest` for all testing. No `unittest` unless the project already uses it.
- **Table-Driven Tests:** Use `@pytest.mark.parametrize` for data-driven tests.
- **Fixtures:** Use `pytest` fixtures for setup/teardown. Scope fixtures appropriately (`function`, `module`, `session`).
- **Mocks:** Use `unittest.mock.patch` sparingly. Prefer dependency injection and fake implementations.
- **Test Organization:** Test files live next to code: `service.py` → `test_service.py` (or in a `tests/` directory mirroring the source structure).
- **Coverage:** Run with `pytest --cov`. See coverage targets in Cross-Cutting Standards below.

---

# 🔀 Git Workflow Conventions
- **Branch Naming:** Use prefixes: `feat/`, `fix/`, `refactor/`, `docs/`, `test/`, `chore/`. Example: `feat/user-auth`, `fix/login-redirect`.
- **Commit Messages:** Follow Conventional Commits. Subject line under 50 characters, imperative mood. Blank line, then body explaining WHY (not what).
  - Format: `type(scope): description`
  - Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `ci`, `perf`
  - Example: `feat(auth): add JWT refresh token rotation`
- **PR Size:** Aim for <400 lines changed. Split large features into stacked PRs.
- **PR Description:** Include: summary (what + why), test plan, breaking changes (if any).
- **Merge Strategy:** Default to squash-and-merge for feature branches. Use merge commits for long-lived branches. Never force-push shared branches.
- **Commit Hygiene:** Each commit should compile and pass tests independently. No "WIP" or "fix typo" commits in PRs -- squash before merge.

---

# 🌐 API Design Patterns
- **Error Response Format:** Use consistent structure across all endpoints:
  ```json
  { "error": { "code": "VALIDATION_FAILED", "message": "Human-readable message", "details": [] } }
  ```
- **Pagination:** Use cursor-based pagination for large datasets (offset-based is acceptable for small, stable datasets). Always return `hasMore` and `nextCursor`.
- **Versioning:** Use URL path versioning (`/api/v1/`) for public APIs. Use header versioning (`Accept: application/vnd.api.v2+json`) only if URL versioning is impractical.
- **Request Validation:** Validate at the handler boundary. Return 400 with specific field errors. Never trust client input past the handler layer.
- **HTTP Methods:** GET (read), POST (create), PUT (full replace), PATCH (partial update), DELETE (remove). Be strict about semantics.
- **Status Codes:** 200 (ok), 201 (created), 204 (no content), 400 (bad request), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict), 422 (unprocessable), 500 (internal error).
- **Naming:** Use plural nouns for resources (`/users`, `/orders`). Use kebab-case for multi-word paths. Nest logically (`/users/{id}/orders`).
- **Idempotency:** POST endpoints that create resources should support idempotency keys. PUT and DELETE must be idempotent.

---

# 📊 Observability & Infrastructure Patterns
- **Health Checks:** Expose `/healthz` (liveness) and `/readyz` (readiness). Liveness: process is running. Readiness: dependencies are connected.
- **Structured Logging:** Always log as structured JSON. Include fields: `timestamp`, `level`, `message`, `request_id`, `user_id` (when available), `duration_ms` (for operations).
- **Log Levels:** `DEBUG` (dev only), `INFO` (normal operations), `WARN` (recoverable issues), `ERROR` (failures requiring attention). Never log at ERROR for expected conditions.
- **Metrics Naming:** Use `<namespace>_<subsystem>_<name>_<unit>` pattern. Examples: `app_http_requests_total`, `app_db_query_duration_seconds`.
- **Tracing:** Propagate trace context (`traceparent` header) across service boundaries. Log the `trace_id` in all log entries for correlation.
- **Dockerfile Best Practices:**
  - Multi-stage builds: separate build and runtime stages.
  - Run as non-root user (`USER nonroot:nonroot`).
  - Minimal base image (`distroless`, `alpine`, or `scratch` for Go).
  - Pin base image versions by digest, not just tag.
  - Copy only necessary files (use `.dockerignore`).
  - Place frequently-changing layers last for cache efficiency.
- **CI/CD Pipeline Structure:**
  - Stages: lint → build → test → security scan → deploy.
  - Tests must pass before deploy. No manual "skip test" overrides.
  - Use caching for dependencies (go mod cache, node_modules, pip cache).
  - Tag images with git SHA, not `latest`.

---

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