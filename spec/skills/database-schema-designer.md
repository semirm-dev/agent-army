---
name: database-schema-designer
description: Guides column type selection, constraint design, partitioning, multi-tenant isolation (RLS), and temporal data patterns (audit trails, SCD, event sourcing) when implementing or reviewing physical database schemas.
scope: universal
languages: []
uses_skills: [database, security]
---

# Database Schema Designer

## When to Use

Invoke this skill when:
- Choosing column types for a new table (text types, numeric precision, timestamps)
- Designing CHECK constraints, exclusion constraints, or composite UNIQUE constraints
- Evaluating multi-tenant schema strategies
- Deciding whether and how to partition a table
- Implementing temporal data patterns (audit trails, slowly changing dimensions)
- Performing a final review before shipping schema changes

**Not this skill:** For entity identification, relationships, normalization, or SQL vs NoSQL decisions, use the `data-modeling` skill instead.

## Column Type Selection Guide

### Text Types

| Need | PostgreSQL | MySQL | Notes |
|------|-----------|-------|-------|
| Short strings (name, email) | `varchar(N)` | `varchar(N)` | Always set a max length |
| Long text (description, body) | `text` | `text` | No length limit, same performance as varchar in PG |
| Fixed-length codes | `char(N)` | `char(N)` | Country codes, currency codes |
| Structured data | `jsonb` | `json` | Use `jsonb` in PG for indexing; avoid for frequently queried fields |

### Numeric Types

| Need | Type | Notes |
|------|------|-------|
| Money / financial | `numeric(19,4)` or `bigint` (store cents) | Never use `float`/`double` for money |
| Counters / quantities | `integer` or `bigint` | Use `bigint` if value can exceed 2B |
| Percentages / ratios | `numeric(5,4)` | Stores 0.0000 to 1.0000 |
| Measurements | `double precision` | Acceptable for non-financial measurements |

### Timestamp Types

| Need | Type | Notes |
|------|------|-------|
| Points in time | `timestamptz` | Always use timezone-aware. Never `timestamp` |
| Dates only | `date` | Birthdays, deadlines |
| Durations | `interval` | PostgreSQL only |
| Created/updated tracking | `timestamptz NOT NULL DEFAULT now()` | Set via trigger or application |

### Identity Types

| Need | Type | Notes |
|------|------|-------|
| Single-DB primary key | `bigint GENERATED ALWAYS AS IDENTITY` | Preferred over `serial` |
| Distributed primary key | `uuid` (prefer UUIDv7) | Sortable, no coordination needed |
| External references | `text` or `varchar(N)` | Store external IDs as-is |

## Constraint Design Patterns

### CHECK Constraints

Use CHECK for value invariants that the application should never violate:

```sql
-- Positive amounts
ALTER TABLE orders ADD CONSTRAINT orders_amount_positive CHECK (amount > 0);

-- Enum-like values (prefer this over DB enums for easier migration)
ALTER TABLE orders ADD CONSTRAINT orders_status_valid
  CHECK (status IN ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled'));

-- Range constraints
ALTER TABLE products ADD CONSTRAINT products_price_range CHECK (price BETWEEN 0.01 AND 999999.99);

-- Cross-column constraints
ALTER TABLE events ADD CONSTRAINT events_dates_valid CHECK (end_date > start_date);
```

### Composite UNIQUE Constraints

```sql
-- Natural keys
ALTER TABLE subscriptions ADD CONSTRAINT subscriptions_user_plan_unique UNIQUE (user_id, plan_id);

-- Partial unique (PostgreSQL) — unique only among active records
CREATE UNIQUE INDEX idx_users_email_active ON users (email) WHERE deleted_at IS NULL;
```

### Exclusion Constraints (PostgreSQL)

```sql
-- Prevent overlapping date ranges
ALTER TABLE bookings ADD CONSTRAINT bookings_no_overlap
  EXCLUDE USING gist (room_id WITH =, tstzrange(check_in, check_out) WITH &&);
```

## Multi-Tenant Schema Strategies

```
What's the tenant isolation requirement?
  |
  +-- Strict regulatory isolation (healthcare, finance)?
  |     --> Database-per-tenant
  |         Pro: complete isolation, easy backup/restore per tenant
  |         Con: connection overhead, complex migrations (run on every DB)
  |
  +-- Moderate isolation, shared infrastructure?
  |     --> Schema-per-tenant (PostgreSQL schemas)
  |         Pro: good isolation, shared connection pool
  |         Con: migration complexity, schema drift risk
  |
  +-- Standard SaaS, cost-efficient?
        --> Shared schema + Row-Level Security (RLS)
            Pro: simple migrations, efficient resource usage
            Con: requires RLS discipline, risk of data leakage if RLS misconfigured

Default: Shared schema + RLS for most SaaS applications.
```

### RLS Setup Pattern (PostgreSQL)

```sql
-- Enable RLS
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Policy: users see only their tenant's data
CREATE POLICY tenant_isolation ON orders
  USING (tenant_id = current_setting('app.tenant_id')::uuid);

-- Set tenant context per request (in application middleware)
SET LOCAL app.tenant_id = 'tenant-uuid-here';
```

## Partitioning Strategy Decision Tree

```
Is the table expected to exceed 100M rows?
  NO → Don't partition (overhead not worth it)
  YES ↓

What's the primary access pattern?
  |
  +-- Time-series queries (logs, events, metrics)?
  |     --> Range partition by timestamp
  |         Partition interval: match retention policy (daily, monthly)
  |
  +-- Queries always filter by a category (tenant, region, status)?
  |     --> List partition by category
  |         One partition per distinct value
  |
  +-- Even distribution needed, no natural partition key?
        --> Hash partition
            Number of partitions: start with 8-16, must be power of 2 for rebalancing
```

### Partition Maintenance

- Automate partition creation (cron job or pg_partman)
- Drop old partitions instead of DELETE for instant cleanup
- Monitor partition sizes — rebalance if skewed >3x

## Temporal Data Patterns

### Audit Trail (Append-Only History)

```sql
CREATE TABLE order_history (
    id bigint GENERATED ALWAYS AS IDENTITY,
    order_id uuid NOT NULL REFERENCES orders(id),
    changed_by uuid NOT NULL,
    changed_at timestamptz NOT NULL DEFAULT now(),
    operation text NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    old_values jsonb,
    new_values jsonb
);

CREATE INDEX idx_order_history_order_id ON order_history (order_id, changed_at DESC);
```

### Slowly Changing Dimension (SCD Type 2)

```sql
CREATE TABLE product_prices (
    id bigint GENERATED ALWAYS AS IDENTITY,
    product_id uuid NOT NULL,
    price numeric(19,4) NOT NULL,
    valid_from timestamptz NOT NULL DEFAULT now(),
    valid_to timestamptz,  -- NULL = currently active
    CONSTRAINT product_prices_dates CHECK (valid_to IS NULL OR valid_to > valid_from)
);

-- Current price query
SELECT * FROM product_prices WHERE product_id = $1 AND valid_to IS NULL;

-- Price at a point in time
SELECT * FROM product_prices
WHERE product_id = $1 AND valid_from <= $2 AND (valid_to IS NULL OR valid_to > $2);
```

### Event Sourcing Table

```sql
CREATE TABLE events (
    id bigint GENERATED ALWAYS AS IDENTITY,
    aggregate_id uuid NOT NULL,
    aggregate_type text NOT NULL,
    event_type text NOT NULL,
    event_data jsonb NOT NULL,
    metadata jsonb,
    version integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT events_version_unique UNIQUE (aggregate_id, version)
);

CREATE INDEX idx_events_aggregate ON events (aggregate_id, version);
```

## Pre-Ship Checklist

Before deploying schema changes to production:

1. [ ] Every column has an explicit type (no implicit defaults)
2. [ ] Every `NOT NULL` column has a `DEFAULT` or is populated on insert
3. [ ] Every foreign key column has an index
4. [ ] `timestamptz` used for all time columns (not `timestamp`)
5. [ ] Primary keys are appropriate type (UUID for distributed, BIGINT for single-DB)
6. [ ] CHECK constraints protect business invariants
7. [ ] Sensitive columns identified and encryption/masking strategy documented
8. [ ] `created_at` and `updated_at` on every table
9. [ ] Indexes cover all common query patterns (verified with EXPLAIN ANALYZE)
10. [ ] Migration is backward-compatible with running application code
