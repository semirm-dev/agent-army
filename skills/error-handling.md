---
name: error-handling
description: "Error taxonomy decisions, propagation patterns, and user-facing message design"
scope: universal
uses_rules:
  - cross-cutting
---

# Error Handling

> Invoke when creating error types, reviewing error propagation, or designing user-facing error messages.

## Error Taxonomy Decision Tree

```
Is the error expected in normal operation?
├── YES → Domain Error (4xx)
│   ├── Input invalid? → Validation Error (400/422)
│   ├── Resource missing? → Not Found (404)
│   ├── Business rule violated? → Conflict/Forbidden (409/403)
│   └── Rate exceeded? → Too Many Requests (429)
└── NO
    ├── External dependency failure? → Infrastructure Error (503)
    │   ├── Timeout? → Retry with backoff
    │   ├── Connection refused? → Circuit breaker
    │   └── Rate limited by upstream? → Back off, propagate 503
    └── Internal bug? → System Error (500)
        ├── Log at ERROR with full stack trace
        ├── Page on-call in production
        └── Return generic message to client
```

Follow error taxonomy definitions in `rules/cross-cutting.md`.

## Error Propagation by Language

### Go
- Wrap with context: `fmt.Errorf("service: operation: %w", err)`
- Check with `errors.Is()` and `errors.As()` — never compare error strings
- Define domain error types with sentinel errors or custom types implementing `error`
- Return errors explicitly — never panic for expected failures

### TypeScript
- Define typed error classes extending `Error` with a `code` field
- Never throw plain strings — always throw typed errors
- Use `instanceof` for error type checking
- Async errors: always catch and wrap with context before re-throwing

### Python
- Chain exceptions: `raise DomainError("context") from original_error`
- Define domain exceptions inheriting from a base project exception
- Use specific exception types — never bare `except:`
- Include machine-readable error codes alongside human messages

## User-Facing Error Messages

### Principles
- **Actionable:** Tell the user what to do, not just what went wrong.
- **Specific:** "Email is already registered" not "Validation failed."
- **Safe:** Never expose stack traces, internal IDs, or system details to end users.
- **Consistent format:** Use the error envelope from `rules/api-design.md`.

### Message Template
```
{summary of what happened} + {what the user can do about it}
```

Examples:
- "This email is already registered. Try logging in or use a different email."
- "Payment processing is temporarily unavailable. Please try again in a few minutes."
- "File exceeds the 10MB limit. Choose a smaller file or compress it before uploading."
