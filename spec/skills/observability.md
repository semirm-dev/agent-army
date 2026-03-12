---
name: observability
description: Guide structured logging, metrics (RED/USE methods), distributed tracing, health checks, alerting design, SLO-based monitoring, and observability maturity progression for production services.
scope: universal
languages: []
uses_skills: [cross-cutting]
---

# Observability Patterns

## When to Use

Invoke this skill when:
- Setting up structured logging in a new service
- Adding metrics to an existing application
- Implementing distributed tracing across services
- Creating health check endpoints (/healthz, /readyz)
- Integrating OpenTelemetry into a project
- Debugging production issues and improving signal quality
- Designing alerting rules or dashboards

## Observability Maturity Decision Tree

```
What observability does the service have today?
  |
  +--> No structured logging?
  |      YES --> START HERE: Level 0 -> Level 1
  |               Set up structured logging (see section below)
  |               Estimated effort: 1-2 hours
  |      NO  |
  |          v
  +--> No health endpoints or metrics?
  |      YES --> Level 1 -> Level 2
  |               Add /healthz, /readyz, and basic metrics
  |               Estimated effort: 2-4 hours
  |      NO  |
  |          v
  +--> No distributed tracing?
  |      YES --> Level 2 -> Level 3
  |               Integrate OTel SDK, add context propagation
  |               Estimated effort: 4-8 hours
  |      NO  |
  |          v
  +--> Level 3: Full Observability
         Focus on: alert tuning, SLO-based monitoring,
         dashboard quality, sampling strategy
```

### Maturity Summary

| Level | Capability | Key Deliverables |
|-------|-----------|-----------------|
| 0 | No observability | Unstructured print statements, no health checks |
| 1 | Structured logging | JSON logs with request_id, proper log levels |
| 2 | Logging + metrics + health | /healthz, /readyz, RED metrics, dashboards |
| 3 | Full observability | Distributed tracing, OTel collector, correlated signals |

## Health Checks
- Expose a **liveness** endpoint (`/healthz`): confirms the process is running and not deadlocked. Returns 200 with no dependency checks.
- Expose a **readiness** endpoint (`/readyz`): confirms dependencies (database, cache, queues) are connected and responsive. Checks critical dependencies with a timeout.
- Liveness failures trigger restarts. Readiness failures remove the instance from load balancing.
- Neither endpoint requires authentication.

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

### Log Level Decision Guide

```
Should someone be paged or take immediate action?
  YES --> ERROR
  NO  |
      v
Did the system compensate (retry, fallback, degrade)?
  YES --> WARN
  NO  |
      v
Is this a normal business operation?
  YES --> INFO
  NO  --> DEBUG (dev only, never enable in production by default)
```

## Metrics

- **Naming pattern:** `<namespace>_<subsystem>_<name>_<unit>` (e.g., `app_http_requests_total`, `app_db_query_duration_seconds`).
- **Standard metrics:** request count, error rate, latency histogram, active connections, queue depth (if applicable).
- **Cardinality warning:** Avoid high-cardinality label values (user IDs, request IDs, full URLs). Use bounded categories (status codes, endpoint names, error types).
- Use histograms for latency, not averages. Averages hide tail latency.

### RED Method (Request-Driven Services)

For every service endpoint, track:

| Signal | Metric Type | Example |
|--------|------------|---------|
| **R**ate | Counter | `app_http_requests_total` |
| **E**rrors | Counter | `app_http_errors_total` (or label on requests) |
| **D**uration | Histogram | `app_http_request_duration_seconds` |

### USE Method (Resource-Oriented)

For infrastructure resources (CPU, memory, disk, connections):

| Signal | What It Measures | Example |
|--------|-----------------|---------|
| **U**tilization | % of resource in use | `app_db_pool_utilization_ratio` |
| **S**aturation | Queued/waiting work | `app_db_pool_waiting_connections` |
| **E**rrors | Resource failures | `app_db_connection_errors_total` |

### Which Method to Use

```
Is this a service endpoint (API, handler)?
  YES --> RED method
  NO  |
      v
Is this an infrastructure resource (DB pool, CPU, queue)?
  YES --> USE method
  NO  --> Start with RED -- it covers most cases
```

## Distributed Tracing
- **Context propagation:** Use W3C TraceContext (`traceparent`, `tracestate`) across all service boundaries. Log `trace_id` in every log entry.
- **Span naming:** `{service}.{operation}` in verb form (e.g., `auth.validateToken`, `orders.processPayment`).
- **Span attributes:** Add business context (entity IDs, tenant, operation type). Never add PII as span attributes.
- **Sampling:** Sample 100% of errors and slow requests. Configurable percentage for success (start at 10%). Use a collector to batch, filter, and route telemetry -- never send directly from application to backend in production.

## Resource Attributes
- Every trace, metric, and log must carry: `service.name`, `service.version`, `deployment.environment`.

## Alerting Strategy

### Alert on Symptoms, Not Causes

```
BAD:  Alert when CPU > 80%
      (CPU can spike without user impact)

GOOD: Alert when p99 latency > 500ms for 5 minutes
      (directly measures user-facing degradation)

BAD:  Alert when disk usage > 90%
      (no immediate user impact)

GOOD: Alert when write error rate > 1% for 2 minutes
      (measures the symptom of disk exhaustion)
```

Exception: alert on resource exhaustion when it is the direct cause of user-facing impact (disk full causing write failures).

- **Every alert must have:** severity level, link to a runbook, and expected response time.
- **Tiered severity:** Critical (pages on-call, immediate), Warning (next business day), Info (dashboard only).

### Alert Fatigue Prevention

Signs of alert fatigue:
- Alerts fire daily and nobody investigates
- Team has a habit of silencing alerts without action
- On-call rotation is dreaded because of noise

Fixes:
- Delete alerts that fire without action. An unactionable alert is worse than no alert.
- Raise thresholds to eliminate false positives. Err on the side of fewer, high-signal alerts.
- Use alert windows (sustained condition over time, not instantaneous spikes).
- Group related alerts to avoid notification storms.
- Review alert quality monthly. Track alert-to-action ratio.

### Alert Design Checklist

1. [ ] Alert measures a user-facing symptom, not an internal metric
2. [ ] Threshold is based on SLO or measured baseline, not a guess
3. [ ] Alert has a runbook link with investigation and mitigation steps
4. [ ] Alert uses a time window (not instant) to avoid flapping
5. [ ] Alert has a clear severity level and response expectation
6. [ ] Someone has verified the alert fires correctly (test in staging)

## Service Level Indicators & Objectives
- **Define SLIs per service:** availability (success rate), latency (p50/p95/p99), error rate. Derived from metrics, not logs.
- **Set SLO targets:** e.g., 99.9% availability, p95 latency < 200ms. Base on user expectations and business requirements.
- **Error budget:** Track remaining error budget. When budget is exhausted, prioritize reliability over features.
