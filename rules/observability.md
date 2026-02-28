---
scope: universal
languages: []
---

# Observability Patterns

## Health Checks
- Expose a **liveness** endpoint: confirms the process is running and not deadlocked.
- Expose a **readiness** endpoint: confirms dependencies (database, cache, queues) are connected and responsive.
- Liveness failures trigger restarts. Readiness failures remove the instance from load balancing.

## Structured Logging
- Always log as structured key-value pairs (JSON or equivalent). Include: `timestamp`, `level`, `message`, `request_id`, `user_id` (when available), `duration_ms` (for operations).
- **Log at boundaries:** incoming requests, outgoing calls, and state transitions. These are the highest-signal log points.
- **Log levels:**
  - `ERROR` -- Unhandled failures requiring human action. If no one needs to act, it is not an error.
  - `WARN` -- Degraded but recoverable. The system compensated (retried, fell back) but the condition should be investigated.
  - `INFO` -- Normal business operations. Request served, job completed, config loaded.
  - `DEBUG` -- Development diagnostics only. Never enable in production by default.
- Never log at ERROR for expected conditions (validation failures, not-found, rate-limited requests). These are domain errors, not system failures.
- Never log secrets, tokens, passwords, or PII. Mask or redact sensitive fields. See `security.md`.

## Metrics
- **Naming pattern:** `<namespace>_<subsystem>_<name>_<unit>` (e.g., `app_http_requests_total`, `app_db_query_duration_seconds`).
- **Standard metrics every service should expose:** request count, error rate, latency histogram, active connections, queue depth (if applicable).
- **Cardinality warning:** Avoid high-cardinality label values -- user IDs, request IDs, full URLs, email addresses. Each unique combination of labels creates a new time series. High cardinality explodes storage cost and degrades query performance. Use bounded categories (status codes, endpoint names, error types) as labels instead.
- Use histograms for latency, not averages. Averages hide tail latency.

## Distributed Tracing
- **Context propagation:** Use W3C TraceContext (`traceparent`, `tracestate` headers) across all service boundaries -- HTTP, messaging, and async workflows. Log `trace_id` in every log entry for correlation.
- **Span naming:** `{service}.{operation}` in verb form (e.g., `auth.validateToken`, `orders.processPayment`).
- **Span attributes:** Add business context (entity IDs, tenant, operation type). Never add PII (email, name, IP address) as span attributes.
- **Sampling strategy:** Sample 100% of errors and slow requests (exceeding latency budget). Sample a configurable percentage of successful requests -- start at 10%, adjust based on volume and cost. Head-based sampling is simpler; tail-based captures more interesting traces but costs more.
- **Telemetry collector:** Use a collector process to batch, filter, sample, and route telemetry. Never send directly from application to observability backend in production -- the collector absorbs back-pressure and decouples the application from the backend.

## Resource Attributes
- Every trace, metric, and log must carry: `service.name`, `service.version`, `deployment.environment`.
- These attributes enable filtering, grouping, and correlation across all telemetry signals.

## Alerting Strategy
- **Alert on symptoms, not causes.** Alert on error rate, latency, and availability -- not CPU or memory, unless resource exhaustion is the direct user-facing impact.
- **Every alert must have:** severity level, link to a runbook, and expected response time.
- **Tiered severity:** Critical (pages on-call, immediate response), Warning (next business day), Info (dashboard only, no notification).
- **Alert fatigue is a failure mode.** If an alert fires regularly and no one acts on it, delete it or tune the threshold. Unactionable alerts train teams to ignore all alerts.
- See `cross-cutting.md` for SLO definitions and alert-on-SLO-violation guidance.
- See `infrastructure.md` for container and CI/CD patterns.
