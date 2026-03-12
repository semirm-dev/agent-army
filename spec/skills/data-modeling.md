---
name: data-modeling
description: Guides data modeling decisions including SQL vs NoSQL selection, entity identification, relationship patterns, normalization, index strategy, column type selection, constraint design, partitioning, multi-tenant isolation (RLS), and temporal data patterns (audit trails, SCD, event sourcing).
scope: universal
languages: []
uses_skills: [database, security]
---

# Data Modeling Skill

## When to Use

Invoke this skill when:
- Designing a new database schema for a feature or service
- Choosing between SQL and NoSQL for a data store
- Planning data relationships (1:1, 1:N, M:N)
- Adding or reviewing indexes on existing tables
- Reviewing schema design for a new feature
- Planning a database migration (adding columns, changing types, splitting tables)

## SQL vs NoSQL Decision Tree

```
What's the primary access pattern?
  |
  +-- Complex joins, aggregations, reporting?
  |     YES --> Relational DB (PostgreSQL)
  |
  +-- Simple key lookup by ID?
  |     YES --> Is the data ephemeral (cache, sessions)?
  |               YES --> Key-Value store (Redis, DynamoDB)
  |               NO  --> Do you need ACID transactions?
  |                         YES --> Relational DB (PostgreSQL)
  |                         NO  --> Key-Value or Document store
  |
  +-- Nested documents, flexible/evolving schema?
  |     YES --> Do you need multi-document transactions?
  |               YES --> Relational DB (PostgreSQL with JSONB)
  |               NO  --> Document store (MongoDB, DynamoDB)
  |
  +-- Mixed or unclear?
        --> Start with PostgreSQL (default choice)
```

### Scale Branch

```
What's the expected scale?
  |
  +-- Single region, < 10K requests/sec?
  |     --> PostgreSQL with read replicas
  |
  +-- Multi-region, > 10K requests/sec?
  |     --> Do you need strong consistency?
  |           YES --> PostgreSQL with Citus or CockroachDB
  |           NO  --> DynamoDB, Cassandra, or MongoDB with sharding
  |
  +-- Sub-millisecond latency required?
        --> Redis or Memcached (cache layer in front of primary store)
```

### ACID Decision

```
Do you need ACID transactions across multiple entities?
  |
  +-- YES --> Relational DB (PostgreSQL)
  |           Use explicit transaction boundaries.
  |
  +-- NO  --> Is eventual consistency acceptable?
                YES --> NoSQL (Document or Key-Value)
                NO  --> Relational DB (PostgreSQL)
```

**Default: PostgreSQL.** Unless you have a specific, documented reason for NoSQL (massive horizontal scale, flexible schema, sub-millisecond reads), start with a relational database.

## Normalization Level Guide

```
Is the data transactional (orders, payments, user accounts)?
  +-- YES --> Normalize to 3NF (every non-key column depends on
  |           the key, the whole key, and nothing but the key)
  +-- NO  --> Is read performance the primary concern?
                +-- YES --> Denormalize for reads (materialized views, duplicated data)
                +-- NO  --> Start 3NF, denormalize only after profiling shows bottlenecks
```

**Rule of thumb:** Normalize for writes, denormalize for reads. Start normalized. Measure. Denormalize only with evidence.

## Schema Design Workflow

Follow these steps in order for every new table or entity:

### Step 1: Identify Entities

List the distinct things you need to store. Each entity becomes a table.

- Ask: "What nouns appear in the feature requirements?"
- Ask: "Which of these have independent lifecycles?"
- Avoid: creating tables for things that are just attributes of another entity

### Step 2: Define Relationships

Map how entities relate to each other. See the Relationship Modeling section below for patterns.

### Step 3: Choose Primary Keys

```
Is this a distributed system (multiple databases, microservices)?
  |
  +-- YES --> UUID (prefer UUIDv7 for ordering)
  |
  +-- NO  --> BIGINT / SERIAL
```

### Steps 4-7: Naming, Audit Columns, Constraints, Indexes

- Add `created_at` and `updated_at` to every table. Add `deleted_at` for soft deletes.
- Apply constraints: `NOT NULL`, `UNIQUE` on natural keys, `CHECK` for value ranges, `FOREIGN KEY` for relationships, `DEFAULT` where sensible.
- Plan indexes using the Index Strategy Selection section below.

## Index Strategy Selection

- Index columns that appear in `WHERE`, `JOIN`, and `ORDER BY` clauses
- Always index foreign key columns
- **Read-heavy tables:** more indexes acceptable — cover common query patterns, consider covering indexes
- **Write-heavy tables:** fewer, targeted indexes — each index slows writes

After creating indexes, verify usage with `EXPLAIN ANALYZE` to confirm the planner uses them. Look for sequential scans on large tables, high loop counts in nested loops, and sorts spilling to disk.

## Relationship Modeling

### One-to-One (1:1)

```
Do the two entities always exist together?
  |
  +-- YES --> Embed in the same table (add columns)
  |           Simpler schema, fewer joins
  |
  +-- NO  --> Separate tables with FK + UNIQUE constraint
              The "optional" side holds the FK
```

### One-to-Many (1:N)

The "many" side holds the foreign key. Always index the FK column.

### Many-to-Many (M:N)

Use a junction (join) table. Add a composite primary key or a composite unique constraint.

### Self-Referential Relationships

For hierarchical data (categories, org charts, comment threads), use a self-referencing foreign key. For deep hierarchies: **adjacency list** (simple, shallow trees), **materialized path** (fast subtree queries), or **closure table** (fast ancestor/descendant queries, more storage).

### Polymorphic Associations (Avoid)

Avoid polymorphic FKs (`commentable_type` + `commentable_id`) — they break referential integrity and prevent FK constraint enforcement. Prefer separate nullable FK columns or separate junction tables instead.

### Soft Deletes

- Use a `deleted_at` timestamp column (nullable) instead of physical `DELETE`. A `NULL` value means the row is active.
- Add a partial index on active rows (e.g., `WHERE deleted_at IS NULL`) to keep queries fast and avoid scanning deleted records.
- Filter deleted rows at the query level or via a repository/middleware layer. Be consistent — choose one approach per project.

For migration strategies and expand-contract patterns, see the `migration-safety` skill.

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

## Pre-Ship Schema Checklist

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
