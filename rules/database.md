---
name: database
description: Migrations, connection pooling, transactions, query safety, indexes, schema conventions, backup recovery, and observability
scope: universal
languages: []
---

# Database Patterns

## Migrations
- **Versioned, forward-only.** Never edit an already-applied migration. Create a new one.
- **Tools:** Use the migration tool appropriate for your language stack.
- **Naming:** Use timestamps: `20260225120000_create_users.sql`. Include both `up` and `down` migrations.
- **Review:** Every migration must be reviewed for data safety. DROP/ALTER on production tables requires explicit confirmation.

## Connection Pooling
- **Always pool.** Never create per-request connections.
- **Set limits:** `max_connections`, `idle_timeout`, `connection_lifetime`. Size pool based on expected concurrency.
- **Health checks:** Validate connections before use (ping or lightweight query).
- **Graceful shutdown:** Drain pool on application shutdown. Close idle connections first, wait for active queries.

## Connection Security
- **Credentials:** Store connection strings via secrets management. Never hardcode in application config or environment-specific files committed to source control.
- **Encrypt connections:** Require SSL/TLS for all database connections. Reject plaintext connections in production. Verify server certificates.

## Transactions
- **Keep short.** No network calls (HTTP, gRPC) inside transactions.
- **Explicit boundaries.** Use `BEGIN`/`COMMIT`/`ROLLBACK` explicitly. Avoid auto-commit for multi-statement operations.
- **Isolation levels:** Default to READ COMMITTED. Use SERIALIZABLE only when required (e.g., financial operations). Document the choice.
- **Retry on conflict.** Serialization failures should be retried, not surfaced as user errors.

## Query Safety
- **Always parameterized.** Never string-concatenate user input into queries.
- **Query builders/ORMs:** Use parameterized execution. Verify generated SQL in development.
- **Avoid `SELECT *`.** List columns explicitly. Prevents schema change breakage and reduces data transfer.
- **Query timeouts:** Set statement-level timeouts on all queries. A runaway query without a timeout can exhaust the connection pool.

## Indexes
- **Index `WHERE`, `JOIN`, `ORDER BY` columns.** Review query plans with `EXPLAIN ANALYZE` for N+1 detection.
- **Composite indexes:** Order columns by selectivity (most selective first).
- **Partial indexes:** Use for filtered queries on large tables (e.g., `WHERE status = 'active'`).
- **Monitor:** Watch for unused indexes (bloat) and missing indexes (slow queries).
- **N+1 prevention:** Detect and eliminate N+1 query patterns. Use eager loading, joins, or batch queries instead of querying in loops.

## Schema Conventions
- **Primary keys:** UUID for distributed systems, BIGINT/SERIAL for single-database systems.
- **Timestamps:** Use `timestamptz` for all date/time columns. Never `timestamp` without timezone.
- **Audit columns:** `created_at` and `updated_at` on every table. Set `created_at` at insert, update `updated_at` via trigger or application.
- **Soft deletes:** Use `deleted_at` timestamp instead of physical deletion when audit trail matters.
- **Naming:** `snake_case` for tables and columns. Plural table names (`users`, `orders`). Foreign keys: `<table>_id` (e.g., `user_id`).

## Backup & Recovery
- **Automated backups:** Schedule regular backups (daily minimum for production). Use the database's native backup tooling (e.g., `pg_dump`, managed service snapshots).
- **Point-in-time recovery:** Enable WAL archiving or equivalent continuous backup for production databases. Know your Recovery Point Objective (RPO).
- **Test restores regularly.** A backup that has never been restored is not a backup. Run restore drills quarterly at minimum.
- **Store backups off-site.** Keep copies in a separate region or account from the primary database. Encrypt backups at rest.
- **Retention:** Keep daily backups for 7 days, weekly for 4 weeks, monthly for 12 months. Adjust based on compliance requirements.

## Read Replicas
- Route read-only queries to replicas when available. Write queries always go to primary.
- Account for replication lag in read-after-write scenarios. Read from primary when consistency matters.

## Observability
- Monitor slow query logs. Set a threshold (e.g., >100ms) and alert.
- Track connection pool utilization (active, idle, waiting). Alert when pool is near max.
- Monitor replication lag on read replicas. Route reads to primary if lag exceeds threshold.
- Log query execution plans for queries that exceed performance budgets.

## Technology Choices
- **ORMs vs Raw SQL:** Pick one approach per entity. Never mix ORM and raw SQL queries for the same table.
- **Default to PostgreSQL** unless you have a specific reason for NoSQL (massive scale, flexible schema, sub-millisecond reads).
