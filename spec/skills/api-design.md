---
name: api-design
description: Drives API design decisions including REST vs GraphQL vs RPC selection, resource modeling, versioning, pagination, error envelope design, caching headers, deprecation lifecycle, batch operations, and pre-ship checklists for new or evolving service interfaces.
scope: universal
languages: []
uses_skills: [security]
---

# API Design Patterns

## When to Use

Invoke this skill when:
- Designing a new API endpoint or service interface
- Choosing between REST, GraphQL, and RPC for a service
- Planning pagination, versioning, or error response formats
- Scaffolding error response envelopes
- Reviewing API consistency across endpoints
- Adding batch operations to an existing API

## API Style Selection

```
What are the primary consumers?
  |
  +-- Multiple client types with varying data needs (web, mobile, third-party)?
  |     YES --> Do clients frequently over-fetch or under-fetch?
  |               YES --> GraphQL (see GraphQL Patterns below)
  |               NO  --> REST (default, well-understood, tooling-rich)
  |
  +-- Internal services only (backend-to-backend)?
  |     YES --> Is low latency or streaming critical?
  |               YES --> RPC / binary protocol (see RPC Patterns below)
  |               NO  --> REST (simpler debugging, wider tooling support)
  |
  +-- Public developer API?
        YES --> REST (lowest barrier to adoption, best documentation tooling)
```

**Default to REST.** Choose GraphQL or RPC only when you have a documented reason.

## REST Resource Design

### Resource Modeling Decision

```
Is this a noun (thing) or a verb (action)?
  |
  +-- Noun (user, order, invoice)
  |     --> Standard CRUD resource: /api/v1/{resource}
  |         GET (list), POST (create), GET /:id, PUT /:id, PATCH /:id, DELETE /:id
  |
  +-- Verb / action (approve, cancel, export)
        --> Is it a state transition on an existing resource?
              YES --> PATCH /api/v1/{resource}/:id with status field
                      or POST /api/v1/{resource}/:id/{action} for complex transitions
              NO  --> POST /api/v1/{action-noun} (treat the action result as a resource)
```

### Nested vs Flat Resources

```
Does the child resource make sense without the parent?
  |
  +-- NO  (order line items, comment replies)
  |     --> Nest: /orders/:id/items
  |         Limit nesting to 2 levels maximum.
  |
  +-- YES (a user's orders, but orders also have their own identity)
        --> Flat with filter: /orders?user_id=123
            Provide the nested convenience route only if the primary access pattern is always through the parent.
```

### REST Conventions

- **HTTP Methods:** GET (read), POST (create), PUT (full replace), PATCH (partial update), DELETE (remove). Be strict about semantics.
- **Status Codes:** 200 (ok), 201 (created), 204 (no content), 400 (bad request), 401 (unauthorized), 403 (forbidden), 404 (not found), 409 (conflict), 422 (unprocessable), 429 (rate limited), 500 (internal error), 503 (unavailable).
- **Naming:** Use plural nouns for resources (`/users`, `/orders`). Use kebab-case for multi-word paths. Nest logically (`/users/{id}/orders`).
- **Request Validation:** Validate at the handler boundary. Return 400 with specific field errors. Never trust client input past the handler layer.
- **Idempotency:** POST endpoints that create resources should support idempotency keys. PUT and DELETE must be idempotent by definition.
- **Content-Type enforcement:** Require `Content-Type` header on all requests with a body. Reject requests with unsupported media types (415).
- **Documentation:** Maintain a machine-readable API spec (e.g., OpenAPI) for all public APIs. Generate from code annotations where possible. Keep spec in sync with implementation.

## Versioning Strategy

```
Who consumes this API?
  |
  +-- External / public consumers?
  |     --> URL path versioning: /api/v1/...
  |         Explicit, visible, easy to document and route.
  |
  +-- Internal services only?
  |     --> Is the API behind a gateway?
  |           YES --> Header versioning (Accept-Version or custom header)
  |                   Gateway can route by version.
  |           NO  --> URL path versioning (simpler, no header coordination)
  |
  +-- Single client (BFF pattern)?
        --> No versioning needed. Client and server deploy together.
            Add versioning when a second consumer appears.
```

## Pagination Decision Tree

```
How large is the dataset?
  |
  +-- Small and stable (<1000 items, rarely changes)?
  |     --> Offset pagination (page + limit)
  |         Simple, supports "jump to page N".
  |         Acceptable trade-off: inconsistent results on concurrent writes.
  |
  +-- Large, growing, or frequently mutated?
        --> Cursor-based pagination
            Return nextCursor + hasMore in every list response.
            |
            Is ordering important beyond insertion order?
              YES --> Encode sort key + unique tiebreaker in cursor
              NO  --> Use primary key as cursor
```

## Error Response Design

Use a consistent envelope across all endpoints:
```json
{ "error": { "code": "VALIDATION_FAILED", "message": "Human-readable message", "details": [] } }
```

### Error Detail Granularity

```
Who is the consumer?
  |
  +-- External / public API?
  |     --> Return machine-readable code + human message + field-level details
  |         Never expose stack traces, internal IDs, or infrastructure details.
  |
  +-- Internal service?
  |     --> Include request_id and upstream error chain for tracing.
  |         Mask PII in error details.
  |
  +-- Frontend (same team)?
        --> Return field-level validation errors keyed by field name
            for direct form binding. Include a top-level message for
            non-field errors (auth, rate limit).
```

## Caching Headers

- **Cache-Control:** Set explicit caching directives on every response. Use `no-store` for authenticated/dynamic data, `max-age` for static resources.
- **ETag / Conditional Requests:** Support `ETag` and `If-None-Match` for read endpoints to reduce bandwidth and enable conditional updates (`If-Match` for optimistic concurrency).
- **Vary:** Include `Vary` header when response differs by request header (e.g., `Accept`, `Authorization`).

## Rate Limit Headers

- Include `X-RateLimit-Limit`, `X-RateLimit-Remaining`, and `X-RateLimit-Reset` (or `RateLimit` headers per IETF draft) on rate-limited endpoints so clients can self-throttle.

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

- **Endpoint design:** Accept an array of items. Process all, return per-item results so the caller knows which items succeeded and which failed.
- **Transaction boundaries:** Decide per use case: all-or-nothing (single transaction) vs partial success (per-item transactions).
- **Size limits.** Enforce a maximum batch size to protect server resources. Document the limit.
- **Rate limiting:** Count batch operations by item count, not request count.
- **Progress reporting:** For large batches, consider async processing with status polling.

## GraphQL Patterns

### Schema Design
- **Schema-first:** Define the schema before implementing resolvers. The schema is the contract.
- **Thin resolvers:** Business logic belongs in service layer, not resolvers.
- **N+1 prevention:** Use DataLoader for batching and caching within a single request.

### Pagination
- **Relay-style cursor pagination:** Use Connection/Edge/PageInfo types for list endpoints.
- **No offset pagination** in GraphQL — cursors are more efficient and stable.

### Error Handling
- **Machine-readable errors:** Return errors with `extensions.code` for client-side handling.
- **Partial success:** GraphQL can return both `data` and `errors` — use this for partial failures.

### Auth & Security
- **Auth in middleware:** Authenticate in context, authorize in resolvers or schema directives (@auth, @hasRole).
- **Query safety:** Enforce depth limiting, complexity scoring, and query timeout. Reject expensive queries.

### Subscriptions
- **Lifecycle management:** Handle subscribe, unsubscribe, and connection keepalive.
- **Auth on subscribe:** Validate credentials on initial subscription, not just on connection.

## RPC / Binary Protocol Patterns

### Method Semantics
- **Unary:** Standard request-response. Default choice for most operations.
- **Server streaming:** Use for large result sets, real-time feeds, or long-running operations.
- **Client streaming:** Use for file uploads or batched writes.
- **Bidirectional streaming:** Use for chat, real-time collaboration, or multiplexed channels.

### Error Model
- **Use canonical error codes.** Map to gRPC status codes or equivalent.
- **Structured error details.** Include machine-readable details for client retry logic.
- **Deadline propagation:** Always propagate timeouts. Set a default deadline on every call.

## Pre-Ship API Checklist

1. [ ] Every endpoint uses the standard error envelope
2. [ ] List endpoints return paginated responses with `hasMore` and cursor/offset
3. [ ] API version is explicit (URL path or header, per strategy above)
4. [ ] POST endpoints that create resources support idempotency keys
5. [ ] Machine-readable API spec (OpenAPI or equivalent) is up to date
6. [ ] Rate limit headers are present on public endpoints
