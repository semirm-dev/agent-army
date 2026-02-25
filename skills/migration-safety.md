---
name: migration-safety
description: Structured checklist for reviewing database migration safety. Use when writing or reviewing migrations that ALTER or DROP tables, columns, or indexes.
---

# Migration Safety Checklist

Use this skill when writing or reviewing database migrations. Walk through each section in order.

## When to Use

- Writing a migration that ALTERs existing tables
- Dropping columns, tables, or indexes
- Renaming columns or tables
- Adding NOT NULL constraints to existing columns
- Any migration touching production data

## 1. Backward Compatibility

Before deploying the migration, verify the **current running application code** still works:

- [ ] Can the current app version read/write correctly with the new schema?
- [ ] If renaming a column: is there a dual-write period with both old and new column names?
- [ ] If dropping a column: has all application code already stopped referencing it?
- [ ] If adding NOT NULL: do all existing rows have a valid value (or a DEFAULT)?
- [ ] If changing a type: is the new type a superset of the old type?

**If any answer is "no":** Split into multiple migrations deployed across multiple releases.

## 2. Dual-Schema Operation

For non-trivial schema changes, use the expand-contract pattern:

1. **Expand:** Add the new column/table alongside the old one
2. **Migrate data:** Backfill the new column from the old one
3. **Switch:** Update application code to use the new column
4. **Contract:** Drop the old column in a later migration

This avoids downtime and allows rollback at each step.

## 3. Down Migration Verification

- [ ] Down migration exists
- [ ] Down migration reverses the up migration cleanly
- [ ] Down migration preserves data where possible (not just `DROP TABLE`)
- [ ] Round-trip tested: `up → down → up` produces identical schema
- [ ] If data was transformed in `up`, down migration can reconstruct original data (or documents data loss)

## 4. Lock Time Estimation

Large tables can lock for extended periods during ALTER:

- [ ] Estimate row count of affected table(s)
- [ ] For tables >1M rows: use online DDL tools (`pt-online-schema-change`, `gh-ost`, PostgreSQL concurrent index creation)
- [ ] For `ADD INDEX`: use `CREATE INDEX CONCURRENTLY` (PostgreSQL) or equivalent
- [ ] For `ALTER TABLE ... ADD COLUMN` with DEFAULT: check if DB version supports instant ADD (PostgreSQL 11+, MySQL 8.0+)
- [ ] Document expected lock time in migration file comments

## 5. Data Safety

- [ ] No data is permanently lost (or loss is explicitly confirmed and documented)
- [ ] Backup plan documented (which tables, when was last backup)
- [ ] For data transformations: dry-run on staging with production-like data
- [ ] Rollback tested on staging before production deployment

## 6. Pre-Deploy Verification

Run these before deploying to production:

```
-- Check migration applies cleanly
migrate up (on staging with production-like data)

-- Check round-trip
migrate down
migrate up

-- Check application works with new schema
run integration tests against migrated database

-- Check query plans haven't regressed
EXPLAIN ANALYZE on critical queries
```

## Decision Tree

```
Is this a new table/column only?
  YES → Safe to deploy directly
  NO ↓

Does it DROP or ALTER existing columns/tables?
  YES ↓
  NO → Review indexes and constraints only

Does running application code reference the dropped/altered object?
  YES → Split into expand-contract migrations
  NO ↓

Is the affected table >1M rows?
  YES → Use online DDL tools, estimate lock time
  NO → Deploy with monitoring, have rollback ready
```
