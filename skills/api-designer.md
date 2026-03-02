---
name: api-designer
description: API style selection, REST resource design, versioning strategy, pagination decision tree, error format guidance, GraphQL patterns, and RPC streaming.
scope: universal
languages: []
uses_rules: [api-design, cross-cutting, security]
---

# API Designer

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

### Pre-Ship API Checklist

1. [ ] Every endpoint uses the standard error envelope
2. [ ] List endpoints return paginated responses with `hasMore` and cursor/offset
3. [ ] API version is explicit (URL path or header, per strategy above)
4. [ ] POST endpoints that create resources support idempotency keys
5. [ ] Machine-readable API spec (OpenAPI or equivalent) is up to date
6. [ ] Rate limit headers are present on public endpoints

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

## Batch Implementation Guidance

- **Endpoint design:** Accept an array of items. Process all, return per-item results.
- **Transaction boundaries:** Decide per use case: all-or-nothing (single transaction) vs partial success (per-item transactions).
- **Rate limiting:** Count batch operations by item count, not request count.
- **Progress reporting:** For large batches, consider async processing with status polling.
