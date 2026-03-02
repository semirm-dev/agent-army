---
name: messaging-patterns
description: Messaging workflow — queue pattern selection, idempotency design, DLQ setup, event schema conventions, delivery guarantee trade-offs, and transactional outbox guidance.
scope: universal
languages: []
uses_rules:
  - messaging-patterns
  - cross-cutting
  - observability
---

# Messaging Patterns Skill

## When to Use

Invoke this skill when:
- Adding asynchronous communication between services
- Choosing between queue patterns (point-to-point, pub/sub)
- Designing event schemas or versioning events
- Implementing idempotent consumers
- Setting up dead letter queues and retry policies
- Evaluating delivery guarantees for a message flow

## Queue Pattern Decision Tree

```
How many consumers need to process each message?

  ├── Exactly one (work distribution)
  │     → Point-to-point queue.
  │       Use for: task processing, email sending, order fulfillment.
  │
  ├── Multiple (event broadcasting)
  │     → Pub/Sub with per-consumer subscriptions.
  │       Use for: user signup → email + analytics + audit.
  │
  └── One consumer, but synchronous response needed
        → Request-Reply over queue.
          Prefer direct HTTP/gRPC unless decoupling is required.
```

**Default to pub/sub** for event-driven architecture. Use point-to-point for work queues.

## Delivery Guarantee Selection

```
What happens if a message is lost?

  ├── Nothing meaningful (metrics, analytics, non-critical notifications)
  │     → At-most-once. Fire and forget. Lowest cost.
  │
  ├── Business impact, but retrying is safe
  │     → At-least-once + idempotent consumer. Default choice.
  │
  └── Financial or compliance impact, duplicates are unacceptable
        → Exactly-once via transactional outbox pattern.
          Write event + business state in the same DB transaction.
```

**Default to at-least-once** with idempotent consumers.

## Idempotency Design

Every consumer must handle redelivered messages safely.

```
Does the operation have a natural idempotency key?
  YES → Use it (e.g., order_id + action, payment_id).
  NO  → Use the message_id from the event envelope.

Implementation:
  1. Check processed-messages store for the idempotency key
  2. If found → skip (already processed), ack the message
  3. If not found → process, then record the key with TTL
```

Prefer naturally idempotent operations:
- `SET` over `INCREMENT`
- `UPSERT` over `INSERT`
- Absolute values over relative deltas

## Event Schema Conventions

Every event must include an envelope:

```
{
  "event_id":       "uuid",
  "event_type":     "entity.action_past_tense",
  "correlation_id": "uuid (propagated across workflow)",
  "timestamp":      "ISO 8601 UTC",
  "version":        1,
  "data":           { ... payload ... }
}
```

- **Naming:** Past tense — `user.created`, `order.shipped`, `payment.failed`
- **Versioning:** Bump `version` on breaking schema changes. Consumers must handle N-1.
- **Size:** Keep messages under 256KB. Store large payloads in blob storage and pass a reference.

## DLQ and Retry Configuration

```
Can the failure be retried (transient error, timeout)?
  YES → Retry with exponential backoff: 1s, 5s, 25s (3-5 attempts).
  NO  ↓

Is the failure a permanent error (bad data, validation)?
  YES → Route directly to DLQ. No retries.
  NO  → Route to DLQ after max retries exhausted.
```

DLQ requirements:
- Every queue must have a paired DLQ
- Alert when DLQ depth > 0
- Build replay tooling to reprocess DLQ messages after fixes
- Set message TTL on DLQ to prevent unbounded growth

## Backpressure for Consumers

```
Is the consumer falling behind (queue depth growing)?
  ├── Consumer is slow → Scale consumers horizontally
  ├── Downstream is slow → Circuit breaker, stop consuming until healthy
  └── Burst traffic → Set prefetch limit to control per-consumer concurrency
```

## Pre-Implementation Checklist

Before adding messaging to a feature, verify:

1. [ ] Selected queue pattern (point-to-point vs pub/sub) with rationale
2. [ ] Defined delivery guarantee (at-most-once / at-least-once / exactly-once)
3. [ ] Consumer is idempotent with documented idempotency key
4. [ ] Event schema follows envelope conventions with versioning
5. [ ] DLQ is configured with max retries and alerting
6. [ ] Processing timeout is set and shorter than visibility timeout
7. [ ] Backpressure strategy is defined (prefetch limit, circuit breaker)
8. [ ] Correlation ID propagates for distributed tracing
9. [ ] Message size is under 256KB; large payloads use blob storage references
