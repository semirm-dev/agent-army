---
name: observability-setup
description: "Observability implementation workflow -- maturity decision tree, log level guide, health endpoint patterns, metrics RED/USE selection, and alert design."
scope: universal
uses_rules:
  - observability
  - cross-cutting
---

# Observability Setup Skill

## When to Use

Invoke this skill when:
- Setting up structured logging in a new service
- Adding metrics to an existing application
- Implementing distributed tracing across services
- Creating health check endpoints (/healthz, /readyz)
- Integrating OpenTelemetry into a project
- Debugging production issues and improving signal quality
- Designing alerting rules or dashboards

> See `rules/observability.md` for structured logging field requirements, OTel SDK setup per language, context propagation, collector architecture, and metrics naming conventions.

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

## Structured Logging Setup

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

> See `rules/observability.md` for common log level mistakes and structured logging field requirements.

Use your language's structured logging library.

## Health Endpoint Patterns

> See `rules/observability.md` for liveness vs readiness definitions, endpoint responsibilities, and probe behavior.

### Health Endpoint Checklist

1. [ ] `/healthz` returns 200 with no dependency checks
2. [ ] `/readyz` checks all required dependencies with a timeout (2-5s)
3. [ ] Neither endpoint requires authentication
4. [ ] Both return JSON with a `status` field
5. [ ] Kubernetes/orchestrator probes are configured to use the correct endpoint
6. [ ] Readiness check does not cascade (checking only direct dependencies)

## Metrics Selection Guide

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

## Alert Design Principles

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

> See `rules/observability.md` for required alert fields (severity, runbook, response time, threshold justification).

> See `rules/observability.md` for alert severity level definitions.

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
