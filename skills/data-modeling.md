---
name: data-modeling
description: Data modeling workflow — SQL vs NoSQL decision tree, normalization guidance, schema design steps, index strategy selection, relationship modeling, and zero-downtime migration strategy.
scope: universal
languages: []
uses_rules:
  - database
  - cross-cutting
  - security
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

After creating indexes, verify usage with query plan analysis. For `EXPLAIN ANALYZE` workflow, red flag patterns, and fixes, see the `performance-audit` skill.

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

## Zero-Downtime Migration Strategy

For changes that cannot be applied in a single step:

```
Step 1: Add new column (nullable, no constraint)
         --> Deploy migration
         --> Application ignores new column

Step 2: Backfill data
         --> Run in batches to avoid long locks
         --> Verify data integrity

Step 3: Add constraint (NOT NULL, CHECK, etc.)
         --> Deploy migration
         --> Application starts using new column

Step 4: Remove old column (if replacing)
         --> Deploy application code that no longer reads old column
         --> Deploy migration to drop old column
```

Use your project's migration tool for executing these steps.
