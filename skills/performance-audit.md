---
name: performance-audit
description: "Performance investigation workflow — profiling cycle, decision trees, query plan analysis, bundle auditing, benchmark methodology, and reporting template."
scope: universal
---

# Performance Audit Skill

## When to Use

Invoke this skill when:
- Investigating slow API endpoints or degraded response times
- Optimizing database queries flagged by monitoring or logs
- Auditing frontend bundle size before a release
- Profiling memory or CPU usage in production or staging
- Running a pre-launch performance review
- Investigating a performance regression between releases

> See `rules/cross-cutting.md` for performance budget targets and `rules/database.md` for query plan analysis.

## Performance Budget Reference

> See `rules/cross-cutting.md` "Performance Budget Targets" for API reads/writes, DB query, frontend LCP, JS bundle, and service startup budgets.

If a target is missed, investigate before shipping. Document exceptions with justification.

## "What's Slow?" Decision Tree

Start here when a performance issue is reported.

```
What is slow?
  |
  +-- API response time?
  |     |
  |     +-- Is the handler doing DB queries?
  |     |     YES --> Query plan analysis (see below)
  |     |     NO  --> Profile the handler (CPU, I/O waits)
  |     |
  |     +-- Is the handler calling external services?
  |           YES --> Check external service latency, add timeouts
  |           NO  --> Profile CPU-bound logic in the handler
  |
  +-- Page load / rendering?
  |     |
  |     +-- Large JS bundle?
  |     |     YES --> Bundle audit (see below)
  |     |     NO  |
  |     |         +-- Slow API calls from client?
  |     |         |     YES --> Trace back to API investigation
  |     |         |     NO  --> Check render performance (framework profiler, layout thrashing)
  |     |
  |     +-- Large images or assets?
  |           YES --> Optimize images, add lazy loading, use CDN
  |           NO  --> Check LCP element, investigate critical rendering path
  |
  +-- Database query?
  |     |
  |     +-- Run EXPLAIN ANALYZE (see Query Plan Analysis below)
  |     +-- Check for N+1 patterns (query count per request)
  |     +-- Check connection pool saturation
  |
  +-- Application startup?
  |     |
  |     +-- Heavy initialization?
  |     |     YES --> Defer non-critical init, use lazy loading
  |     |     NO  |
  |     |         +-- Slow dependency connections?
  |     |               YES --> Add connection timeouts, parallelize init
  |     |               NO  --> Profile startup sequence
  |     |
  |     +-- Large dependency tree?
  |           YES --> Audit dependencies, remove unused
  |           NO  --> Profile module loading time
  |
  +-- Memory usage?
        |
        +-- Gradual growth (leak)?
        |     YES --> Heap profiling over time, check for retained references
        |     NO  |
        |         +-- Spike on specific operations?
        |               YES --> Profile that operation's allocations
        |               NO  --> Check baseline memory vs available resources
        |
        +-- High GC pressure?
              YES --> Reduce allocations, pool objects, profile allocation sites
              NO  --> Check if memory budget is reasonable for workload
```

## Profiling Workflow

Follow this cycle for every performance investigation. Do not skip steps.

```
1. IDENTIFY         2. MEASURE          3. OPTIMIZE         4. VERIFY
   What is slow?       Profile it          Apply fix           Re-measure
   (metrics, logs,     (tooling below)     (one change         (same conditions,
    user reports)                           at a time)          compare results)
        |                   |                   |                   |
        +-------------------+-------------------+-------------------+
                            |
                    Repeat until budget met
```

**Rules:**
- Change one thing at a time. Multiple changes make it impossible to attribute improvement.
- Measure under realistic conditions (production-like data volume, concurrent load).
- Record baseline numbers before optimizing. No baseline means no proof of improvement.

Use your language's standard profiling tools (CPU profiler, memory profiler, tracing).

## Query Plan Analysis Guide

Use this workflow for any query that misses the database performance budget.

### Workflow

1. Run `EXPLAIN ANALYZE` on the query with production-like data volume
2. Check the plan for red flags (table below)
3. Apply the fix
4. Re-run `EXPLAIN ANALYZE` to confirm improvement
5. Benchmark under concurrent load (single-query timing can be misleading)

> See also `rules/database.md` for query plan analysis fundamentals.

### Red Flags and Fixes

| Red Flag | What It Means | Fix |
|----------|---------------|-----|
| Seq Scan on large table with filter | Missing index on WHERE column | Add index on filtered column(s) |
| Nested Loop with high loop count | Join iterating too many rows | Add index on join column, consider hash join |
| Sort with external merge / disk | Sort spilling to disk | Add index matching ORDER BY, increase `work_mem` |
| Rows Removed by Filter >> rows returned | Index exists but not selective enough | Add composite index, use partial index |
| Hash Join with high bucket count | Large hash table in memory | Check join conditions, add index for nested loop path |
| Bitmap Heap Scan with many recheck rows | Index returns too many candidates | Narrow the query filter, use composite index |

> See `rules/database.md` for `EXPLAIN ANALYZE` workflow and red flag patterns.

## Frontend Bundle Audit

### Tools

Use your platform's bundle analysis tools to visualize bundle composition and set CI gates on bundle size. Common categories:
- **Bundle visualizer:** Shows what's in the bundle and how large each module is
- **Source map explorer:** Treemap from source maps to identify large modules
- **Size checker:** Check package size before adding dependencies
- **CI gate:** Automated performance scoring and size limits in CI

### Common Wins

| Problem | Solution | Expected Impact |
|---------|----------|-----------------|
| Large vendor bundle | Code split by route using lazy loading | 30-60% reduction in initial load |
| Unused library code | Tree shaking (ensure ESM imports) | 10-40% per library |
| Large utility libraries | Replace with smaller, modular alternatives | 20-80KB savings |
| Unoptimized images | Modern formats (WebP/AVIF), responsive sizes, lazy loading | 50-80% image size reduction |
| Duplicate dependencies | Deduplicate with package manager tools | 5-20% reduction |
| No compression | Enable gzip/brotli on server or CDN | 60-80% transfer size reduction |

### Measurement Checklist

1. [ ] Run bundle analyzer -- identify largest chunks
2. [ ] Check for duplicate packages in the bundle
3. [ ] Verify tree shaking is working (no dead code in output)
4. [ ] Run Lighthouse CI -- record LCP, FCP, TBT, CLS scores
5. [ ] Compare gzipped size against the 200KB budget
6. [ ] Test on throttled connection (4G profile in DevTools)

## Before/After Benchmarking

### Rules for Valid Benchmarks

- **Isolate the variable.** Change one thing between before and after runs.
- **Consistent environment.** Same machine, same load, same data volume. No other processes competing for resources.
- **Sufficient iterations.** Run enough iterations to get stable results. Discard warm-up runs.
- **Compare percentiles.** Report p50, p95, and p99 -- averages hide tail latency.
- **Statistical significance.** Use tools that compute confidence intervals. A 2% improvement within noise is not real.

Use your language's standard benchmarking tools and follow the rules above for valid results.

### Reporting Template

Document benchmark results in this format for PR descriptions and ADRs:

```
## Performance Impact

Endpoint/Function: POST /api/v1/orders
Change: Added composite index on (user_id, status, created_at)

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| p50    | 145ms  | 32ms  | -78%   |
| p95    | 312ms  | 58ms  | -81%   |
| p99    | 890ms  | 120ms | -87%   |

Iterations: 1000 requests, 10 concurrent users
Environment: staging (4 CPU, 8GB RAM, PostgreSQL 16)
```
