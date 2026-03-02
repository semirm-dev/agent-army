---
name: observability
description: Health checks, structured logging, metrics, distributed tracing, and alerting
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
- **Log at boundaries:** incoming requests, outgoing calls, and state transitions.
- **Log levels:**
  - `ERROR` -- Unhandled failures requiring human action. If no one needs to act, it is not an error.
  - `WARN` -- Degraded but recoverable. The system compensated but the condition should be investigated.
  - `INFO` -- Normal business operations. Request served, job completed, config loaded.
  - `DEBUG` -- Development diagnostics only. Never enable in production by default.
- Never log at ERROR for expected conditions (validation failures, not-found, rate-limited requests).
- Never log secrets, tokens, passwords, or PII. Mask or redact sensitive fields.
- **Log sampling:** For high-throughput paths, sample repetitive log entries to prevent log volume from overwhelming storage and budgets.

## Metrics
- **Naming pattern:** `<namespace>_<subsystem>_<name>_<unit>` (e.g., `app_http_requests_total`, `app_db_query_duration_seconds`).
- **Standard metrics:** request count, error rate, latency histogram, active connections, queue depth (if applicable).
- **Cardinality warning:** Avoid high-cardinality label values (user IDs, request IDs, full URLs). Use bounded categories (status codes, endpoint names, error types).
- Use histograms for latency, not averages. Averages hide tail latency.

## Distributed Tracing
- **Context propagation:** Use W3C TraceContext (`traceparent`, `tracestate`) across all service boundaries. Log `trace_id` in every log entry.
- **Span naming:** `{service}.{operation}` in verb form (e.g., `auth.validateToken`, `orders.processPayment`).
- **Span attributes:** Add business context (entity IDs, tenant, operation type). Never add PII as span attributes.
- **Sampling:** Sample 100% of errors and slow requests. Configurable percentage for success (start at 10%). Use a collector to batch, filter, and route telemetry -- never send directly from application to backend in production.

## Resource Attributes
- Every trace, metric, and log must carry: `service.name`, `service.version`, `deployment.environment`.

## Alerting Strategy
- **Alert on symptoms, not causes.** Alert on error rate, latency, and availability -- not CPU or memory unless directly user-facing.
- **Every alert must have:** severity level, link to a runbook, and expected response time.
- **Tiered severity:** Critical (pages on-call, immediate), Warning (next business day), Info (dashboard only).
- **Alert fatigue is a failure mode.** If an alert fires regularly with no action taken, delete it or tune the threshold.

## Service Level Indicators & Objectives
- **Define SLIs per service:** availability (success rate), latency (p50/p95/p99), error rate. Derived from metrics, not logs.
- **Set SLO targets:** e.g., 99.9% availability, p95 latency < 200ms. Base on user expectations and business requirements.
- **Error budget:** Track remaining error budget. When budget is exhausted, prioritize reliability over features.
