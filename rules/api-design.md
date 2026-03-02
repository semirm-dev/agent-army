---
scope: universal
languages: []
---

# API Design Patterns

## Error Response Format

Use a consistent envelope across all endpoints:
```json
{ "error": { "code": "VALIDATION_FAILED", "message": "Human-readable message", "details": [] } }
```

## REST Conventions

- **HTTP Methods:** GET (read), POST (create), PUT (full replace), PATCH (partial update), DELETE (remove). Be strict about semantics.
- **Status Codes:** 200 (ok), 201 (created), 204 (no content), 400 (bad request), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict), 422 (unprocessable), 429 (rate limited), 500 (internal error), 503 (unavailable).
- **Naming:** Use plural nouns for resources (`/users`, `/orders`). Use kebab-case for multi-word paths. Nest logically (`/users/{id}/orders`).
- **Pagination:** Use cursor-based pagination for large datasets (offset-based is acceptable for small, stable datasets). Always return `hasMore` and `nextCursor`.
- **Versioning:** Use URL path versioning (`/api/v1/`) for public APIs. Use header versioning only if URL versioning is impractical.
- **Request Validation:** Validate at the handler boundary. Return 400 with specific field errors. Never trust client input past the handler layer.
- **Idempotency:** POST endpoints that create resources should support idempotency keys. PUT and DELETE must be idempotent by definition.
- **Documentation:** Maintain a machine-readable API spec (e.g., OpenAPI) for all public APIs. Generate from code annotations where possible. Keep spec in sync with implementation -- stale docs are worse than no docs.

## Caching Headers

- **Cache-Control:** Set explicit caching directives on every response. Use `no-store` for authenticated/dynamic data, `max-age` for static resources.
- **ETag / Conditional Requests:** Support `ETag` and `If-None-Match` for read endpoints to reduce bandwidth and enable conditional updates (`If-Match` for optimistic concurrency).
- **Vary:** Include `Vary` header when response differs by request header (e.g., `Accept`, `Authorization`).

## Deprecation Strategy

- **Announce early.** Add `Deprecated` response header and `sunset` date before removing any endpoint.
- **Overlap period.** Run old and new versions concurrently for at least one release cycle.
- **Migration path.** Every deprecation notice must include the replacement endpoint or migration steps.
- **Monitor usage.** Track calls to deprecated endpoints. Do not remove until traffic drops to near zero.

## Backwards Compatibility

- **Additive changes only.** New fields, new endpoints, and new optional parameters are safe. Removing or renaming fields is a breaking change.
- **Never change the meaning** of an existing field or status code.
- **Use a new version** when a breaking change is unavoidable. See Versioning above.

## Batch Operations

- **Provide batch endpoints** for operations clients frequently call in loops (bulk create, bulk delete).
- **Partial success.** Return per-item results so the caller knows which items succeeded and which failed.
- **Size limits.** Enforce a maximum batch size to protect server resources. Document the limit.

## RPC / Binary Protocols
- **Schema versioning:** Always backward-compatible. Use proto field numbering or Avro schema registry.
- **Deadline propagation:** Propagate timeouts across service boundaries. Set a default deadline on every call.

## Cross-References

- **Caching:** See `caching-patterns.md` for server-side caching strategies behind API endpoints.
- **Security:** See `security.md` for authentication, authorization, CORS, and rate limiting.
