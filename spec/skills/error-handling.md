---
name: error-handling
description: Classify errors by domain/infrastructure/system level, decide wrap vs translate at boundaries, and format user-facing error responses. Delegates per-language error creation to language patterns skills.
scope: universal
languages: []
uses_skills: [api-design, observability]
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

## Error Creation

Define domain error types with machine-readable codes and human-readable messages. See language-specific patterns and coder skills (`go/patterns`, `typescript/patterns`, `python/patterns`) for idiomatic error creation examples per language.

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
  YES -> Wrap: add context, preserve original (use language-idiomatic wrapping)
  NO  -> Translate: create new error appropriate for consumer
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
