---
name: testing-strategy
description: Testing pyramid guidance, decision trees for test type selection, flaky test prevention, and test data patterns. References cross-cutting.md coverage targets.
---

# Testing Strategy Skill

## When to Use

Invoke this skill when:
- Planning test coverage for a new feature
- Deciding which type of test to write
- Diagnosing flaky tests
- Setting up test infrastructure
- Reviewing test strategy for a module

## Testing Pyramid

```
        ╱╲
       ╱E2E╲          ~10% — Full user flows
      ╱──────╲
     ╱ Integr. ╲      ~20% — DB, API, service boundaries
    ╱────────────╲
   ╱    Unit      ╲   ~70% — Pure logic, transformations
  ╱────────────────╲
```

Adjust ratios based on project type:
- **API/backend heavy:** 70% unit, 25% integration, 5% E2E
- **Frontend heavy:** 50% unit, 30% integration, 20% E2E (component tests count as integration)
- **Data pipeline:** 60% unit, 35% integration, 5% E2E

## Test Type Decision Tree

```
What are you testing?
  ↓
Pure logic (no I/O, no side effects)?
  YES → Unit test
  NO ↓

Database query or migration?
  YES → Integration test (real DB, transaction rollback)
  NO ↓

API endpoint (HTTP handler)?
  YES → Integration test (test server, mocked dependencies)
  NO ↓

External service interaction?
  YES → Integration test (contract test or MSW/wiremock)
  NO ↓

Full user journey across multiple services/pages?
  YES → E2E test
  NO → Unit test (default to the simplest test type)
```

## What NOT to Test

- **Generated code:** protobuf stubs, OpenAPI clients, ORM migrations (test the queries, not the generated glue)
- **Trivial getters/setters:** If the function just returns a field, don't test it
- **Third-party internals:** Don't test that Express routes correctly or that React renders — test YOUR logic
- **Configuration wiring:** Test that the app starts, not every config combination
- **Private functions:** Test through the public API. If a private function is complex enough to need its own test, it should probably be a public function in a smaller package

## Contract Testing

Use contract tests at service boundaries:

**When:**
- Service A calls Service B's API
- Frontend calls backend API
- Any cross-team or cross-service dependency

**How:**
- Define expected request/response shapes as contracts
- Provider verifies it satisfies contracts
- Consumer verifies it correctly handles provider responses
- Tools: Pact, MSW (frontend), WireMock, schema validation

## Flaky Test Prevention

| Cause | Fix |
|-------|-----|
| Time-dependent assertions | Use fake clocks (`vi.useFakeTimers()`, `clock` package) |
| Network calls in unit tests | Mock at boundary (MSW, fake HTTP client) |
| Shared mutable test state | Isolate: fresh DB per test, `beforeEach` reset |
| Race conditions | Use `-race` flag (Go), avoid shared state |
| Non-deterministic data | Use factories/builders with fixed seeds |
| Sleep/wait for async | Use polling/waitFor, not `time.Sleep`/`setTimeout` |
| File system dependencies | Use temp directories with cleanup |
| Port conflicts | Use random ports or test-assigned ports |

## Test Data Patterns

### Factories / Builders (preferred)

```go
// Go
func NewTestUser(opts ...func(*User)) *User {
    u := &User{ID: uuid.New(), Name: "Test User", Email: "test@example.com"}
    for _, opt := range opts {
        opt(u)
    }
    return u
}

// Usage
user := NewTestUser(func(u *User) { u.Name = "Custom Name" })
```

```typescript
// TypeScript
function createTestUser(overrides?: Partial<User>): User {
  return {
    id: crypto.randomUUID(),
    name: "Test User",
    email: "test@example.com",
    ...overrides,
  };
}
```

```python
# Python
def create_test_user(**overrides: Any) -> User:
    defaults = {"id": uuid4(), "name": "Test User", "email": "test@example.com"}
    return User(**{**defaults, **overrides})
```

### Anti-Patterns
- **Shared fixtures:** Mutable test data shared across tests causes coupling
- **Raw SQL inserts:** Use factories, not hand-written INSERT statements
- **Snapshot-based:** Snapshot tests break on any change and provide no useful signal
- **Copy-paste setup:** Extract into fixtures/factories/helpers

## Coverage Targets

From cross-cutting standards:
- **Critical paths** (auth, payments, data mutations): 80%+ line coverage
- **Utilities and shared libraries:** 90%+ line coverage
- **Generated code:** No coverage requirement
- **Integration tests:** Cover all API endpoints and external service interactions — not counted toward line coverage

## CI Parallelization

| Language | Tool | Flag |
|----------|------|------|
| Go | `go test` | `-parallel N`, `-count=1` for no cache |
| TypeScript | vitest | `--threads`, `--pool forks` |
| Python | pytest-xdist | `-n auto` |
