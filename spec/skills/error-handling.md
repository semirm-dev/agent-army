---
name: error-handling
description: Error taxonomy, per-language error creation patterns, propagation decision tree, and user-facing error guidelines following project standards.
scope: universal
languages: []
uses_rules: [cross-cutting, api-design, observability]
---

# Error Handling Skill

## When to Use

Invoke this skill when:
- Creating new error types or error handling code
- Reviewing error propagation across boundaries
- Designing user-facing error messages
- Deciding whether to wrap, translate, or bubble an error

## Error Taxonomy

Categorize every error into one of three levels:

| Level | Examples | HTTP Status | Log Level | Action |
|-------|----------|-------------|-----------|--------|
| **Domain** | Validation failure, not found, conflict, business rule violation | 4xx | DEBUG/INFO | Return to caller with actionable message |
| **Infrastructure** | Timeout, connection failure, service unavailable | 503 | WARN | Retry with backoff, return 503 with retry guidance |
| **System** | Internal bug, panic recovery, unhandled state | 500 | ERROR | Log full stack trace, alert on-call, return generic 500 |

## Per-Language Error Creation

### Go

```go
// Domain errors: sentinel errors + wrapping
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

// Always wrap with context
fmt.Errorf("auth: validate token: %w", err)

// Check errors with Is/As
if errors.Is(err, ErrNotFound) { ... }

// Custom error types for rich context
type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: %s: %s", e.Field, e.Message)
}
```

### TypeScript

```typescript
// Base error class for domain errors
class DomainError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly details?: unknown
  ) {
    super(message);
    this.name = this.constructor.name;
  }
}

// Specific domain errors
class NotFoundError extends DomainError {
  constructor(resource: string, id: string) {
    super(`${resource} not found: ${id}`, "NOT_FOUND");
  }
}

class ValidationError extends DomainError {
  constructor(field: string, message: string) {
    super(`${field}: ${message}`, "VALIDATION_FAILED", { field, message });
  }
}
```

### Python

```python
class DomainError(Exception):
    """Base class for domain errors."""
    def __init__(self, message: str, code: str = "DOMAIN_ERROR") -> None:
        super().__init__(message)
        self.code = code

class NotFoundError(DomainError):
    def __init__(self, resource: str, resource_id: str) -> None:
        super().__init__(f"{resource} not found: {resource_id}", "NOT_FOUND")

# Always chain exceptions
try:
    result = do_something()
except SomeError as e:
    raise DomainError("context: operation failed") from e
```

## Propagation Decision Tree

```
Error occurs
  |
Same layer? (e.g., service -> service)
  YES -> Bubble: re-raise/return as-is
  NO  |

Crossing boundary? (e.g., repository -> service, service -> handler)
  YES |

Is the original error meaningful to the consumer?
  YES -> Wrap: add context, preserve original
         Go: fmt.Errorf("svc: op: %w", err)
         TS: throw new ServiceError("context", { cause: err })
         Py: raise ServiceError("context") from err
  NO  -> Translate: create new error appropriate for consumer
         Go: return NewNotFoundError("user", id)
         TS: throw new NotFoundError("user", id)
         Py: raise NotFoundError("user", id)
```

## User-Facing Error Guidelines

- **Actionable messages:** Tell the user what they can do to fix it ("Email is required" not "Validation failed")
- **No stack traces:** Never expose internal error details to end users
- **Error codes:** Include machine-readable error codes for programmatic handling
- **Consistency:** Use the standard error response format from api-design rules:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Email address is required",
    "details": [{ "field": "email", "message": "must not be empty" }]
  }
}
```

## Logging / Returning / Alerting Matrix

| Error Level | Return to User | Log | Alert |
|-------------|---------------|-----|-------|
| Domain | 4xx with specific error code and message | DEBUG or INFO | No |
| Infrastructure | 503 with retry guidance | WARN with error details | If repeated (circuit breaker) |
| System | 500 with generic message | ERROR with full stack trace | Yes -- page on-call |

## Anti-Patterns

- **Swallowing errors:** Never catch and ignore. At minimum, log.
- **Stringly-typed errors:** Never match error messages with string comparison. Use error types/codes.
- **Double-logging:** Log at the boundary where you handle the error, not at every layer it passes through.
- **Generic catch-all:** Never `catch (error) { return 500 }` without distinguishing error types.
- **Exposing internals:** Never include SQL errors, file paths, or stack traces in API responses.
