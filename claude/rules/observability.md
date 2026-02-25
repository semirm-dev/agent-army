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

## Per-Language Logging Examples

### Go (log/slog)

```go
slog.Info("request handled",
    "request_id", reqID,
    "duration_ms", dur.Milliseconds(),
    "status", statusCode,
)

slog.Error("payment failed",
    "request_id", reqID,
    "user_id", userID,
    "error", err,
)
```

### TypeScript (structured logger)

```typescript
logger.info({
  message: "request handled",
  requestId,
  durationMs,
  status: statusCode,
});

logger.error({
  message: "payment failed",
  requestId,
  userId,
  error: err.message,
});
```

### Python (structlog)

```python
import structlog

log = structlog.get_logger()

log.info("request_handled", request_id=req_id, duration_ms=dur, status=status_code)

log.error("payment_failed", request_id=req_id, user_id=user_id, error=str(err))
```
