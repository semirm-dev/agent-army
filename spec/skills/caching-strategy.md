---
name: caching-strategy
description: Cache decision trees, strategy selection, TTL guidance, invalidation picker, and stampede prevention.
scope: universal
languages: []
uses_rules: [caching-patterns, observability, cross-cutting, security]
---

# Caching Strategy Skill

## When to Use

Invoke this skill when:
- Adding caching to a service or endpoint
- Choosing between caching strategies (cache-aside, write-through, write-behind)
- Designing cache key schemas
- Planning cache invalidation for a feature
- Debugging stale data or cache-related bugs
- Preventing cache stampede on high-traffic keys

## Cache-or-Not Decision Tree

Most data should not be cached. Start from "no cache" and justify adding one.

```
Is the data read-heavy (>10:1 read-to-write ratio)?
  NO  → Do not cache. Optimize the source query instead.
  YES ↓

Is it expensive to compute or fetch (>50ms, external API, complex join)?
  NO  → Do not cache. Cheap lookups don't benefit from caching overhead.
  YES ↓

How stale can the data be?
  ├── Real-time (0s tolerance) → Do not cache, or use very short TTL (<5s)
  │                               with event-based invalidation.
  ├── Seconds (1-30s)          → Short TTL + event-based invalidation.
  │                               Good for: dashboards, live feeds.
  ├── Minutes (1-15min)        → Standard cache-aside with TTL.
  │                               Good for: user profiles, product details.
  └── Hours+ (1hr-24hr)        → Aggressive caching with long TTL.
                                  Good for: static config, catalogs, translations.

Still unsure? Ask: "What happens if a user sees data that is N seconds old?"
  If the answer is "nothing bad" → cache it.
  If the answer is "incorrect behavior" → do not cache, or use event invalidation.
```

### Do Not Cache

- Data that changes on every request (nonces, CSRF tokens)
- Data that is cheap to compute (simple lookups on indexed columns)
- Error responses or null results (unless short TTL <30s)
- Security-sensitive data where staleness causes authorization bugs

## Strategy Selection

```
What is the dominant access pattern?
  ├── Read-heavy, writes are infrequent
  │     → Cache-aside (default)
  │
  ├── Reads and writes, consistency is critical
  │     → Write-through
  │
  └── High write volume, eventual consistency acceptable
        → Write-behind
```

**Default to cache-aside.** Only move to write-through or write-behind when you have a measured need.

## TTL Selection Guide

| Data Type | Recommended TTL | Rationale |
|-----------|----------------|-----------|
| User sessions | Match session timeout | Stale sessions break auth |
| Auth tokens / permissions | 1-5 min | Security-sensitive, must reflect revocations |
| User profiles | 15 min | Changes are infrequent, staleness is tolerable |
| Application config | 5 min | Rarely changes, but updates should propagate quickly |
| Product catalog | 1 hr | Stable data, price changes handled by event invalidation |
| Feature flags | 5 min | Balance between freshness and load reduction |
| Static content / translations | 24 hr+ | Rarely changes, deploy-triggered invalidation |
| Search results | 1-5 min | Depends on index freshness requirements |
| Rate limit counters | Match rate limit window | Must be accurate within the window |

**Rules of thumb:**
- When in doubt, shorter TTL is safer. You can always increase it after measuring.
- Never set TTL to infinity. Every entry must expire.
- Pair TTL with event-based invalidation for data where freshness matters.

## Invalidation Strategy Picker

Always use TTL as a safety net, even when using event-based or tag-based invalidation.

```
How critical is data freshness?
  ├── Mission-critical (auth, permissions, inventory counts)
  │     → Event-based invalidation + short TTL (1-5 min backup)
  │       Publish invalidation event on every write.
  │       TTL catches missed events.
  │
  ├── Best-effort (user profiles, product details)
  │     → TTL-only
  │       Set TTL based on acceptable staleness.
  │       Simplest approach, no extra infrastructure.
  │
  └── Bulk changes (tenant config, catalog imports, feature flags)
        → Tag-based invalidation + TTL
          Tag keys: tag:tenant:123, tag:catalog:winter-2026
          Invalidate all keys with a tag on bulk operations.
```

| Strategy | Complexity | Freshness | Infrastructure |
|----------|------------|-----------|----------------|
| TTL-only | Low | Eventual (within TTL) | None beyond cache |
| Event-based | Medium | Near real-time | Pub/sub or message queue |
| Tag-based | Medium | On-demand for bulk | Tag tracking in cache |
| Event + TTL | Medium-High | Near real-time + safety net | Pub/sub + cache |

## Stampede Prevention

A cache stampede occurs when a popular key expires and many concurrent requests all miss the cache simultaneously, overwhelming the source of truth.

```
Is this a high-traffic key (>100 requests/second)?
  NO  → Standard cache-aside is fine. Stampede risk is negligible.
  YES ↓

Can the source handle a brief spike?
  YES → Probabilistic early expiration (simplest)
  NO  ↓

Is the data expensive to recompute (>500ms)?
  YES → Distributed lock (one request recomputes, others wait)
  NO  → Request coalescing (deduplicate in-flight requests)
```

## Pre-Implementation Checklist

Before adding caching to a feature, verify:

1. [ ] Confirmed the data qualifies for caching (decision tree above)
2. [ ] Selected a caching strategy with documented rationale
3. [ ] Defined key schema following key design conventions
4. [ ] Set TTL values based on the TTL selection guide
5. [ ] Planned invalidation strategy (TTL, event-based, or tag-based)
6. [ ] Assessed stampede risk for high-traffic keys
7. [ ] Added cache hit/miss metrics for observability
8. [ ] Tested graceful degradation when cache is unavailable
9. [ ] Documented cache decisions in the service README or ADR
