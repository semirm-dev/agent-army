---
name: caching-patterns
description: Cache-aside strategy, write policies, invalidation, key design, and failure handling
scope: universal
languages: []
---

# Caching Patterns

## Cache-Aside (Lazy Loading)
- **Read path:** Check cache; on miss, read from source of truth, write to cache, return.
- **Write path:** Write to source of truth, then invalidate cache. Never update cache directly on write.
- **TTL required:** Every cache entry must have a TTL. No indefinite caching.

## Write Strategies

| Strategy | When to Use | Trade-off |
|----------|-------------|-----------|
| **Write-through** | Data consistency critical, low write volume | Higher write latency, always consistent |
| **Write-behind** | High write volume, eventual consistency acceptable | Lower latency, risk of data loss |
| **Cache-aside** | Default choice. Read-heavy workloads | Stale reads during TTL window |

## Cache Invalidation
- **TTL-based:** Default strategy. Set TTL based on data volatility (user profiles: 15min, config: 5min, sessions: match session timeout).
- **Event-based:** Invalidate on write events. Use pub/sub or message queues for distributed invalidation.
- **Tag-based:** Group related keys with tags. Invalidate all keys with a tag on bulk changes.
- **Never rely on manual invalidation alone.** Always pair with TTL as a safety net.

## Key Design
- **Format:** `{service}:{entity}:{id}` (e.g., `auth:user:550e8400`, `catalog:product:sku-123`).
- **Namespacing:** Prefix with service name to avoid collisions in shared cache instances.
- **Avoid large keys:** Keep key names under 128 bytes.
- **Versioning:** Include a version when schema changes: `v2:user:550e8400`.

## Cache Failure Handling
- Design for graceful degradation when cache is unavailable.
- Fall back to source of truth -- accept higher latency over service failure.
- Never let cache failures cascade into application outages.
- Use circuit breakers on cache connections to avoid blocking on unresponsive backends.

## Observability
- Monitor cache hit/miss ratio -- target above 80% for production caches.
- Track cache latency (p50, p95, p99) alongside source-of-truth latency for comparison.
- Alert on sudden hit rate drops -- may indicate invalidation bugs or traffic pattern shifts.
- Log cache eviction rate -- high eviction signals an undersized cache or missing TTL tuning.

## Anti-Patterns
- **No TTL:** Every entry must expire. Unbounded caches grow until out of memory.
- **Cache as primary store:** Cache is ephemeral. Always have a source of truth.
- **N+1 cache calls:** Batch multi-key lookups instead of individual get calls in loops.
- **Caching errors:** Never cache error responses or null results without a short TTL (< 30s).
- **Cache stampede:** When a popular key expires, many requests simultaneously rebuild it, overwhelming the source of truth. Mitigate with lock-on-rebuild (only one request rebuilds, others wait), stale-while-revalidate (serve stale data while one request rebuilds in background), or probabilistic early expiration.
- **Over-caching:** Do not cache data that changes every request or is cheap to compute.
