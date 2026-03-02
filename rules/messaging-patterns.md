---
name: Messaging Patterns
description: Queue patterns, idempotent consumers, DLQs, event schemas, and delivery guarantees
scope: universal
languages: []
---

# Messaging & Event-Driven Patterns

## Queue Patterns

| Pattern | Use Case | Example |
|---------|----------|---------|
| **Point-to-point** | Task distribution to workers | Order processing, email sending |
| **Pub/Sub** | Event broadcasting to multiple consumers | User signup → email + analytics + audit |
| **Request-Reply** | Synchronous-style over async transport | RPC over message queue |

**Default to pub/sub** for event-driven architecture. Use point-to-point for work queues.

## Idempotent Consumers
- **Every consumer must be idempotent.** Messages can be delivered more than once.
- **Idempotency key:** Use `message_id` or a natural business key (e.g., `order_id + action`).
- **Implementation:** Check if the message was already processed before executing. Use a processed-messages store with TTL.
- **Natural idempotency:** Prefer operations that are naturally idempotent (SET over INCREMENT, upsert over insert).

## Dead Letter Queues (DLQ)
- **Every queue must have a DLQ.** Failed messages go to DLQ after max retries.
- **Max retries:** 3-5 attempts with exponential backoff (1s, 5s, 25s).
- **DLQ monitoring:** Alert when DLQ depth > 0. Include dashboards for DLQ message age and count.
- **Replay strategy:** Build tooling to replay DLQ messages back to the source queue after fixing the issue.

## Event Schema Design
- **Envelope:** Include `event_id`, `event_type`, `correlation_id`, `timestamp`, `version`, `data` in every event.
- **Versioning:** Include `version` field. Support at least N-1 versions. Use schema registry for contract validation.
- **Correlation ID:** Propagate `correlation_id` across all events in a workflow for distributed tracing.
- **Event naming:** Use past tense (`user.created`, `order.shipped`), not imperative (`create.user`).

## Ordering & Delivery Guarantees

| Guarantee | Cost | When to Use |
|-----------|------|-------------|
| **At-most-once** | Lowest | Metrics, analytics, non-critical notifications |
| **At-least-once** | Medium (default) | Most business events. Requires idempotent consumers |
| **Exactly-once** | Highest | Financial transactions. Use transactional outbox pattern |

- **Default to at-least-once** with idempotent consumers.
- **Ordering:** Use partition/shard keys when message order matters within an entity.
- **Transactional outbox:** Write event to an outbox table in the same DB transaction as the business operation, then publish asynchronously.

## Message Processing Timeout
- **Set explicit processing timeouts on consumers.** No consumer should process a message indefinitely.
- **Timeout < visibility timeout:** Processing timeout must be shorter than the queue's visibility timeout to allow automatic redelivery.
- **Escalation:** Log and route to DLQ when processing consistently exceeds the timeout threshold.

## Backpressure & Rate Limiting
- **Consumer rate limiting:** Limit processing rate to protect downstream services.
- **Queue depth monitoring:** Alert when queue depth exceeds threshold. Scale consumers horizontally.
- **Circuit breaker:** Stop consuming when downstream is unhealthy. Resume after health check passes.
- **Prefetch limit:** Set prefetch count to control how many messages a consumer processes concurrently.

## Anti-Patterns
- **Fire-and-forget without DLQ:** Every async operation must have failure handling. Silent drops lose data.
- **Unbounded queues:** Always set max queue size or TTL on messages. Prevent memory exhaustion.
- **Giant messages:** Keep messages small (< 256KB). Store large payloads in blob storage and pass a reference.
- **Synchronous patterns over queues:** Don't use request-reply when a direct HTTP call is simpler and sufficient.
- **No schema validation:** Validate event schema at producer and consumer boundaries.
