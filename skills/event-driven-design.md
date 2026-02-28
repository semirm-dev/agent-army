---
name: event-driven-design
description: "Event-driven architecture workflow — sync vs async decision tree, delivery guarantees, queue technology selection, event schema design, idempotency key selection, consumer patterns, and circuit breaker design."
scope: universal
---

# Event-Driven Design Skill

## When to Use

Invoke this skill when:
- Deciding whether communication between services should be synchronous or asynchronous
- Implementing async processing (background jobs, event pipelines, notifications)
- Designing event schemas or choosing an envelope format
- Selecting a queue technology for a new project or feature
- Setting up dead letter queues and retry strategies
- Adding idempotency to message consumers

> See `rules/messaging-patterns.md` for DLQ platform configs, exponential backoff, idempotency patterns, backpressure mechanisms, and consumer concurrency limits.

## Sync vs Async Decision Tree

```
Does the caller need an immediate response?
  |
  +-- YES --> Does the response require data from another service?
  |             |
  |             +-- YES --> Synchronous HTTP / gRPC call
  |             |
  |             +-- NO  --> Handle locally, return response
  |
  +-- NO  --> Is it fire-and-forget (no result needed)?
                |
                +-- YES --> Standard queue (SQS, RabbitMQ, BullMQ)
                |           No ordering guarantees needed.
                |
                +-- NO  --> Is message ordering required?
                              |
                              +-- NO  --> Standard queue with at-least-once delivery
                              |
                              +-- YES --> Per-entity or global ordering?
                                            |
                                            +-- Per-entity --> Kafka (partition by entity ID)
                                            |                  or SQS FIFO (group by entity ID)
                                            |
                                            +-- Global    --> Single Kafka partition
                                                               or single SQS FIFO group
```

### Follow-up: Delivery Guarantee Selection

After choosing sync vs async, determine the delivery guarantee.

```
Can the consumer tolerate duplicates?
  |
  +-- YES --> At-least-once (default). Make consumer idempotent.
  |
  +-- NO  --> Is this a financial or transactional operation?
                |
                +-- YES --> Exactly-once via transactional outbox pattern.
                |           Write event + business data in one DB transaction.
                |
                +-- NO  --> Is data loss acceptable (metrics, analytics)?
                              |
                              +-- YES --> At-most-once (fire-and-forget publish)
                              |
                              +-- NO  --> At-least-once with idempotent consumer
```

## Queue Technology Selection

> See `rules/messaging-patterns.md` for queue pattern comparison (ordering, throughput, retention, DLQ support).

### Selection Heuristic

```
Already using AWS and need simple task queue?
  YES --> SQS (Standard or FIFO)

Need complex routing (topic, header-based, priority)?
  YES --> RabbitMQ

Need event replay, audit trail, or high-throughput streaming?
  YES --> Kafka

Already have Redis and need background jobs?
  YES --> Redis-based task queue

None of the above?
  --> Start with SQS (lowest operational burden)
```

## Event Schema Design Workflow

### Step 1: Define the Envelope

Every event uses a standard envelope. This is non-negotiable.

> See `rules/messaging-patterns.md` for the envelope format (event_id, event_type, correlation_id, timestamp, version, data).

### Step 2: Name the Event

Follow the event naming conventions in `rules/messaging-patterns.md` (past tense, dot notation).

### Step 3: Version the Schema

```
Is this a backward-compatible change (adding optional fields)?
  YES --> Increment minor version, keep same event_type
          Consumers ignore unknown fields.

Is this a breaking change (removing fields, changing types)?
  YES --> Create new event_type version: order.placed.v2
          Support N-1: keep publishing v1 alongside v2
          until all consumers migrate.
```

### Step 4: Decide on Schema Registry

```
Do multiple teams produce/consume events?
  YES --> Use a schema registry (Confluent Schema Registry, AWS Glue)
          Validate schemas on publish.

Single team, small number of events?
  NO  --> Validate with shared type definitions in code
```

### Step 5: Propagate Correlation ID

- Generate `correlation_id` at the entry point (API gateway, first service).
- Pass it through every downstream event and HTTP call.
- Log `correlation_id` in all structured log entries for distributed tracing.

## Idempotency Implementation

### Step 1: Choose the Idempotency Key

```
Does the event have a natural business key?
  (e.g., order_id + action, payment_id, invoice_number)
  |
  +-- YES --> Use the natural key. It survives retries and replays.
  |
  +-- NO  --> Use event_id from the envelope. Generate as UUID at
              the producer. Never let the queue system generate it.
```

### Step 2: Implement the Check

Implement the idempotency check using your project's database or cache layer.

## Consumer Patterns

### Point-to-Point vs Pub/Sub Selection

```
How many services need to react to this event?
  |
  +-- One (task assignment) --> Point-to-point queue
  |
  +-- Multiple (broadcast) --> Pub/Sub (topic + subscriptions)
  |
  +-- One now, possibly more later --> Pub/Sub
      Start with pub/sub to avoid migration when you add consumers.
```

> See `rules/messaging-patterns.md` for queue pattern details and use cases.

### Circuit Breaker for Downstream Services

```
Monitor downstream error rate
  |
  +-- Error rate < 50% --> CLOSED (normal operation)
  |                        Process messages normally.
  |
  +-- Error rate >= 50% --> OPEN (stop consuming)
  |                         Pause consumer. Messages stay in queue.
  |                         Wait for cooldown period (30s-60s).
  |
  +-- After cooldown --> HALF-OPEN (probe)
                         Process one message.
                         Success? --> CLOSED
                         Failure? --> OPEN (restart cooldown)
```

## Pre-Implementation Checklist

Before adding event-driven communication, verify:

1. [ ] Sync vs async decision documented (why async?)
2. [ ] Queue technology selected with rationale
3. [ ] Event schema follows envelope pattern with versioning
4. [ ] DLQ configured with retry count and backoff
5. [ ] Consumer is idempotent (dedup strategy chosen)
6. [ ] Monitoring in place: queue depth, DLQ depth, consumer lag
7. [ ] Backpressure mechanism configured (prefetch, rate limit, or circuit breaker)
8. [ ] Correlation ID propagated from upstream
9. [ ] Message size under 256KB (large payloads in blob storage)
10. [ ] Graceful shutdown: consumer drains in-flight messages before exit
