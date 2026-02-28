---
scope: universal
languages: []
---

# Database Patterns

## Migrations
- **Versioned, forward-only.** Never edit an already-applied migration. Create a new one.
- **Tools:** Use the migration tool appropriate for your language stack (see language-specific database files).
- **Naming:** Use timestamps: `20260225120000_create_users.sql`. Include both `up` and `down` migrations.
- **Review:** Every migration must be reviewed for data safety. DROP/ALTER on production tables requires explicit confirmation.

## Connection Pooling
- **Always pool.** Never create per-request connections.
- **Set limits:** `max_connections`, `idle_timeout`, `connection_lifetime`. Size pool based on expected concurrency.
- **Health checks:** Validate connections before use (ping or lightweight query).
- **Graceful shutdown:** Drain pool on application shutdown. Close idle connections first, wait for active queries.

## Transactions
- **Keep short.** No network calls (HTTP, gRPC) inside transactions.
- **Explicit boundaries.** Use `BEGIN`/`COMMIT`/`ROLLBACK` explicitly. Avoid auto-commit for multi-statement operations.
- **Isolation levels:** Default to READ COMMITTED. Use SERIALIZABLE only when required (e.g., financial operations). Document the choice.
- **Retry on conflict.** Serialization failures should be retried, not surfaced as user errors.

## Query Safety
- **Always parameterized.** Never string-concatenate user input into queries.
- **Query builders/ORMs:** Use parameterized execution. Verify generated SQL in development.
- **Avoid `SELECT *`.** List columns explicitly. Prevents schema change breakage and reduces data transfer.
- See also `security.md` SQL Injection Prevention for the threat model.

## Indexes
- **Index `WHERE`, `JOIN`, `ORDER BY` columns.** Review query plans with `EXPLAIN ANALYZE` for N+1 detection.
- **Composite indexes:** Order columns by selectivity (most selective first).
- **Partial indexes:** Use for filtered queries on large tables (e.g., `WHERE status = 'active'`).
- **Monitor:** Watch for unused indexes (bloat) and missing indexes (slow queries).
- **N+1 prevention:** Detect and eliminate N+1 query patterns. Use eager loading, joins, or batch queries instead of querying in loops.

## Query Plan Analysis

- Analyze query execution plans for all new queries to detect inefficiencies.
- **Red flags:** Full table scan on large tables with filter (add index), Nested Loop with high loop count (use JOIN), external sort spills (add index or increase memory budget), Rows Removed by Filter >> returned rows (missing index)
- **Common join strategies:** Nested loop (small/indexed), hash join (large equality), merge join (pre-sorted).

## Schema Conventions
- **Primary keys:** UUID for distributed systems, BIGINT/SERIAL for single-database systems.
- **Timestamps:** Use `timestamptz` for all date/time columns. Never `timestamp` without timezone.
- **Audit columns:** `created_at` and `updated_at` on every table. Set `created_at` at insert, update `updated_at` via trigger or application.
- **Soft deletes:** Use `deleted_at` timestamp instead of physical deletion when audit trail matters.
- **Naming:** `snake_case` for tables and columns. Plural table names (`users`, `orders`). Foreign keys: `<table>_id` (e.g., `user_id`).

## ORMs vs Raw SQL
- **ORMs:** Use when productivity matters (CRUD-heavy code, rapid prototyping). Good for simple queries and schema management.
- **Raw SQL:** Use for complex queries, performance-critical paths, bulk operations, and advanced database features.
- **Never mix both in the same function.** Pick one approach per operation. Mixing creates confusion about query execution and error handling.

## NoSQL Patterns

### Document Stores

- **Denormalize for reads.** Embed related data in the document when it's always read together.
- **Reference for writes.** Use references (foreign keys) when related data changes independently.
- **Pagination:** Use cursor-based pagination (`_id > last_id`). Never use skip/offset on large collections.
- **Transactions:** Use sparingly. Multi-document transactions are expensive. Design schemas to minimize cross-document updates.
- **Indexes:** Index all query fields. Use compound indexes for multi-field queries. Monitor with `explain()`.
- **Schema validation:** Enforce schema at the application layer (application-layer validators) even though the DB is schemaless.

### Key-Value Stores

- **Right data structure:** Choose the appropriate data structure for the access pattern: simple strings for cache/counters, hash maps for object fields, sets for unique collections, sorted sets for ranked data, lists for queues.
- **Always set TTL.** Every key must expire. Unbounded key spaces grow until out of memory.
- **Key naming:** `{service}:{entity}:{id}:{field}` (e.g., `auth:session:abc123`).
- **Atomic operations:** Use transactions or scripting features for multi-step operations that must be atomic.

### When to Choose NoSQL vs Relational

| Factor | Relational DB | Document Store | Key-Value |
|--------|--------------|----------------|-----------|
| **Data model** | Relational, normalized | Nested, denormalized | Flat key-value |
| **Query complexity** | Complex joins, aggregations | Single-document queries, simple aggregations | Key lookup only |
| **Consistency** | Strong (ACID) | Tunable (eventual to strong) | Eventual |
| **Scale** | Vertical + read replicas | Horizontal sharding | Horizontal sharding |
| **Best for** | Business data, transactions | Content, catalogs, user profiles | Cache, sessions, real-time |

**Default to a relational database** unless you have a specific reason for NoSQL (massive scale, flexible schema, sub-millisecond reads).
