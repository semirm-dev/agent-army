<!-- Sync: Must stay in sync with cursor/505-caching.mdc -->

# 🗄️ Caching Patterns

## Cache-Aside (Lazy Loading)
- **Read path:** Check cache → if miss, read from DB → write to cache → return.
- **Write path:** Write to DB → invalidate cache. Never update cache directly on write.
- **TTL required:** Every cache entry must have a TTL. No indefinite caching.

### Per-Language Examples

#### Go (Redis)

```go
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    key := fmt.Sprintf("user:%s", id)

    // Check cache
    cached, err := s.redis.Get(ctx, key).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }

    // Cache miss — read from DB
    user, err := s.repo.GetUser(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user: get: %w", err)
    }

    // Write to cache
    data, _ := json.Marshal(user)
    s.redis.Set(ctx, key, data, 15*time.Minute)

    return user, nil
}
```

#### TypeScript (ioredis)

```typescript
async function getUser(id: string): Promise<User> {
  const key = `user:${id}`;

  const cached = await redis.get(key);
  if (cached) {
    return JSON.parse(cached) as User;
  }

  const user = await userRepo.findById(id);
  await redis.set(key, JSON.stringify(user), "EX", 900);

  return user;
}
```

#### Python (redis-py)

```python
async def get_user(user_id: str) -> User:
    key = f"user:{user_id}"

    cached = await redis.get(key)
    if cached:
        return User.model_validate_json(cached)

    user = await user_repo.get(user_id)
    await redis.set(key, user.model_dump_json(), ex=900)

    return user
```

## Write Strategies

| Strategy | When to Use | Trade-off |
|----------|-------------|-----------|
| **Write-through** | Data consistency critical, low write volume | Higher write latency, always consistent |
| **Write-behind** | High write volume, eventual consistency acceptable | Lower latency, risk of data loss |
| **Cache-aside** | Default choice. Read-heavy workloads | Stale reads during TTL window |

## Cache Invalidation

- **TTL-based:** Default strategy. Set TTL based on data volatility (user profiles: 15min, config: 5min, sessions: match session timeout).
- **Event-based:** Invalidate on write events. Use pub/sub or message queues for distributed invalidation.
- **Tag-based:** Group related keys with tags. Invalidate all keys with a tag on bulk changes (e.g., `tag:tenant:123`).
- **Never rely on manual invalidation alone.** Always pair with TTL as a safety net.

## Key Design

- **Format:** `{service}:{entity}:{id}` (e.g., `auth:user:550e8400`, `catalog:product:sku-123`).
- **Namespacing:** Prefix with service name to avoid collisions in shared Redis instances.
- **Avoid large keys:** Keep key names under 128 bytes.
- **Versioning:** Include a version when schema changes: `v2:user:550e8400`.

## Distributed Cache Considerations

- **Redis Cluster:** Use hash tags `{user}:profile`, `{user}:sessions` to co-locate related keys on the same shard.
- **Stampede prevention:** Use probabilistic early expiration or distributed locks to prevent thundering herd on popular keys.
- **Serialization:** Use JSON for debuggability, MessagePack/Protobuf for performance. Be consistent per entity type.
- **Connection pooling:** Always pool Redis connections. Set max connections based on concurrency.

## Anti-Patterns

- **No TTL:** Every entry must expire. Unbounded caches grow until OOM.
- **Cache as primary store:** Cache is ephemeral. Always have a source of truth (DB).
- **N+1 cache calls:** Batch multi-key lookups with `MGET` / pipeline instead of individual `GET` calls in loops.
- **Caching errors:** Never cache error responses or null results without a short TTL (< 30s).
- **Over-caching:** Don't cache data that changes every request or is cheap to compute.
