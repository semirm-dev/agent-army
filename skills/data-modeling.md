---
name: data-modeling
description: "Data modeling workflow — SQL vs NoSQL decision tree, normalization guidance, schema design steps, index strategy selection, relationship modeling, and zero-downtime migration strategy."
scope: universal
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

> See `rules/database.md` for schema naming conventions, index design, composite ordering, migration safety rules, and anti-patterns.

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
  |
  +-- YES --> Normalize to 3NF
  |           - Eliminate repeating groups (1NF)
  |           - Remove partial dependencies (2NF)
  |           - Remove transitive dependencies (3NF)
  |           - Every non-key column depends on the key,
  |             the whole key, and nothing but the key
  |
  +-- NO  --> Is read performance the primary concern?
                |
                +-- YES --> Denormalize for reads
                |           - Duplicate data into read-optimized tables
                |           - Use materialized views for aggregations
                |           - Accept write complexity for read speed
                |
                +-- NO  --> Balanced: 3NF with strategic denormalization
                            - Start normalized
                            - Denormalize only after profiling shows bottlenecks
                            - Document every denormalization decision
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

> See `rules/database.md` "Schema Conventions" for primary key selection guidelines.

### Step 4: Add Audit Columns

Every table gets `created_at` and `updated_at` columns. Add `deleted_at` for soft deletes when audit trail is required.

> See `rules/database.md` "Schema Conventions" for audit column definitions and conventions.

### Step 5: Apply Naming Conventions

Follow the naming conventions defined in `rules/database.md` "Schema Conventions".

### Step 6: Add Constraints

Apply constraints to enforce data integrity at the database level:

- `NOT NULL` on all columns unless NULL has explicit business meaning
- `UNIQUE` on natural keys (email, username, external IDs)
- `CHECK` for value ranges and valid states
- `FOREIGN KEY` for all relationships (with appropriate `ON DELETE` action)
- `DEFAULT` for columns with sensible defaults

### Step 7: Plan Indexes

See the Index Strategy Selection section below.

## Index Strategy Selection

```
Which columns appear in WHERE clauses?
  --> Index those columns

Which columns appear in JOIN conditions?
  --> Index foreign key columns (always)

Which columns appear in ORDER BY?
  --> Consider index if combined with WHERE filter

Is this a read-heavy table?
  |
  +-- YES --> More indexes are acceptable
  |           - Cover common query patterns
  |           - Consider covering indexes (INCLUDE columns)
  |
  +-- NO (write-heavy) --> Fewer, targeted indexes
                           - Each index slows writes
                           - Only index columns used in critical queries
```

### Index Verification

Always verify index usage after creation:

```sql
EXPLAIN ANALYZE SELECT ... ;  -- Check that the index is actually used
```

Red flags: sequential scan on large tables, nested loops with high loop count, external sort spills.

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

```sql
-- Separate tables: user always exists, profile is optional
CREATE TABLE users (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_profiles (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    bio        TEXT,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### One-to-Many (1:N)

The "many" side holds the foreign key. Always index the FK column.

```sql
CREATE TABLE orders (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total      NUMERIC(10, 2) NOT NULL CHECK (total >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
```

### Many-to-Many (M:N)

Use a junction (join) table. Add a composite primary key or a composite unique constraint.

```sql
CREATE TABLE user_roles (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id    BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_id)
);
```

### Self-Referential Relationships

For hierarchical data (categories, org charts, comment threads):

```sql
CREATE TABLE categories (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    parent_id  BIGINT REFERENCES categories(id) ON DELETE SET NULL,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_categories_parent_id ON categories (parent_id);
```

For deep hierarchies that need efficient querying, consider:
- **Adjacency list** (above) -- simple, good for shallow trees
- **Materialized path** (`path TEXT: '/1/5/12/'`) -- fast subtree queries
- **Closure table** -- fast ancestor/descendant queries, more storage

### Polymorphic Associations (Avoid)

```
Do you need a single table referencing multiple entity types?
  |
  +-- AVOID --> Polymorphic FKs (commentable_type + commentable_id)
  |             - No FK constraint enforcement
  |             - Complex queries
  |             - Breaks referential integrity
  |
  +-- PREFER --> Separate FK columns (nullable) or separate junction tables
                 - Each FK has proper constraints
                 - Database enforces integrity
```

```sql
-- BAD: polymorphic
-- commentable_type TEXT, commentable_id BIGINT (no FK possible)

-- GOOD: separate nullable FKs
CREATE TABLE comments (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    post_id    BIGINT REFERENCES posts(id) ON DELETE CASCADE,
    article_id BIGINT REFERENCES articles(id) ON DELETE CASCADE,
    body       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (
        (post_id IS NOT NULL AND article_id IS NULL) OR
        (post_id IS NULL AND article_id IS NOT NULL)
    )
);
```

### Soft Deletes

> See `rules/database.md` for soft delete conventions and query patterns.

When using soft deletes, add a partial index for active records:

```sql
CREATE INDEX idx_users_active ON users (email) WHERE deleted_at IS NULL;
```

**Important:** Every query must include `WHERE deleted_at IS NULL` unless explicitly querying deleted records. Consider a database view for convenience.

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
