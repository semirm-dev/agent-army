---
name: api-designer
description: "GraphQL patterns, RPC streaming, batch implementation, and API design decisions"
scope: universal
uses_rules:
  - api-design
---

# API Designer

> Invoke when designing new API endpoints, scaffolding error formats, or reviewing API consistency.

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
