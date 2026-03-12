---
name: testing
description: Plan test coverage using the testing pyramid, apply language-agnostic test structure, isolation, mocking boundaries, async testing, flaky prevention, coverage strategy, CI pipeline integration, and property-based and performance testing across unit, integration, and e2e tests.
scope: universal
languages: []
uses_skills: [cross-cutting]
---

# Testing Patterns

## When to Use

Invoke this skill when:
- Planning test coverage for a new feature
- Deciding which type of test to write
- Diagnosing flaky tests
- Setting up test infrastructure
- Reviewing test strategy for a module
- Planning CI test pipeline (parallelization, splitting, quarantine)
- Deciding whether to add performance or property-based tests

## Testing Pyramid

```
        /\
       /E2E\          ~10% -- Full user flows
      /------\
     / Integr. \      ~20% -- DB, API, service boundaries
    /------------\
   /    Unit      \   ~70% -- Pure logic, transformations
  /----------------\
```

Adjust ratios based on project type:
- **API/backend heavy:** 70% unit, 25% integration, 5% E2E
- **Frontend heavy:** 50% unit, 30% integration, 20% E2E (component tests count as integration)
- **Data pipeline:** 60% unit, 35% integration, 5% E2E

## Test Type Decision Tree

```
What are you testing?
  |
Pure logic (no I/O, no side effects)?
  YES -> Unit test
  NO |

Database query or migration?
  YES -> Integration test (real DB, transaction rollback)
  NO |

API endpoint (HTTP handler)?
  YES -> Integration test (test server, mocked dependencies)
  NO |

External service interaction?
  YES -> Integration test (contract test or HTTP mocking library)
  NO |

Full user journey across multiple services/pages?
  YES -> E2E test
  NO -> Unit test (default to the simplest test type)
```

## What NOT to Test

- **Generated code:** protobuf stubs, OpenAPI clients, ORM migrations (test the queries, not the generated glue)
- **Trivial getters/setters:** If the function just returns a field, don't test it
- **Third-party internals:** Don't test that your framework routes correctly or that your UI library renders -- test YOUR logic
- **Configuration wiring:** Test that the app starts, not every config combination
- **Private functions:** Test through the public API. If a private function is complex enough to need its own test, it should probably be a public function in a smaller package

## Test Naming
- Describe behavior, not implementation: "returns error when user not found" not "test getUserById."
- Test names should read as documentation. A failing test name alone should tell you what broke.
- Use a consistent naming convention per project.

## Test Organization
- One test file per source file, co-located or mirrored in a tests directory.
- Group related tests with describe blocks or subtests.
- Use setup/teardown hooks for shared fixture creation and resource cleanup.
- Order tests logically: happy path first, then edge cases, then error paths.

## Test Data Management

### Strategy Selection

```
What kind of test needs data?
  |
  +-- Unit test?
  |     --> In-memory factories with minimal fields.
  |         Build only what the test asserts on.
  |
  +-- Integration test (database)?
  |     --> Use factories that insert into the real database.
  |         Wrap each test in a transaction and rollback.
  |         Never depend on seed data from another test.
  |
  +-- E2E test?
        --> Use a known seed dataset loaded before the suite.
            Reset to seed state between tests (truncate + re-seed or snapshot restore).
            Keep the seed dataset small and version-controlled.
```

### Factory Design Principles

- **Minimal defaults:** Factories produce valid objects with the fewest fields possible. Override only what the test cares about.
- **Named variants:** `factory.expiredUser()` not `factory.user({ expired: true, expiresAt: ... })`. Encode intent in the name.
- **No randomness:** Deterministic values only. Random data hides bugs and causes flaky tests.
- **Composable:** Factories can call other factories for nested objects. `factory.orderWithItems()` creates both order and line items.

## Test Isolation
- **Database tests:** Wrap each test in a transaction, rollback after. No test should depend on another test's data.
- **File system tests:** Use temp directories. Clean up in teardown.
- **In-memory databases:** Use in-memory storage for fast unit tests when full database features are not needed.
- **Network isolation:** No real HTTP calls in unit tests. Use fakes or recorded responses.

## Mocking Philosophy
- Prefer fakes and thin interfaces over heavy mocking frameworks.
- Mock at system boundaries (HTTP, DB, message queues), not between internal modules.
- Use spy/mock tools only for call verification, not for replacing internal module behavior.

## Async Testing
- Always await async operations. Unawaited assertions silently pass and hide real failures.
- Test both resolved (success) and rejected (error) paths for every async operation.
- Set explicit timeouts on async tests to catch hangs rather than waiting indefinitely.
- Clean up async resources (timers, subscriptions, listeners) in teardown to prevent leaks between tests.

## Error Path Testing
- Every happy path needs a corresponding error path test.
- Test that errors contain useful context: type, message, and machine-readable code where applicable.
- Test error propagation. Verify errors are wrapped with context, not swallowed or replaced with generic messages.
- Test boundary validation: missing fields, wrong types, out-of-range values, empty inputs.

## Snapshot Testing
- Avoid snapshot tests for rendered UI output -- large diffs with no diagnostic signal. Use snapshot tests only for serialized contracts (API response shapes, CLI output, config schemas) where the exact shape is the specification.

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
- Tools: Pact, HTTP mocking libraries (framework-specific interceptors, WireMock), schema validation

## Flaky Test Prevention

- **No sleep-based synchronization.** Use polling, event-based waiting, or signaling mechanisms.
- **No network calls in unit tests.** Mock at the boundary.
- **Deterministic test data.** Use factories with fixed values, not random generators.
- **Isolated test state.** Each test creates its own data. No shared mutable fixtures.
- **Explicit timeouts.** Set test timeouts to catch hangs. CI timeouts should be stricter than local.

### Flaky Test Quarantine

When a test fails intermittently:

1. Mark it as quarantined (skip in CI gate, run in a separate non-blocking job)
2. File a ticket with the flaky test name, failure frequency, and last failure log
3. Fix within one sprint — quarantined tests that linger erode trust in the suite
4. After fixing, remove quarantine and monitor for one week

## Coverage
- **Prefer branch coverage over line coverage.** Line coverage misses untested conditional paths.
- Run coverage as part of CI, not just locally.
- Set coverage thresholds as CI gates. Fail the build if coverage drops below the threshold.
- **Critical paths** (auth, payments, data mutations): 80%+ line coverage.
- **Utilities and shared libraries:** 90%+ line coverage.
- **Generated code** (protobuf, OpenAPI stubs): No coverage requirement.

## CI Integration

### Test Pipeline Ordering

```
Fast feedback first:

1. Lint + type check        (seconds)
2. Unit tests               (seconds to low minutes)
3. Integration tests        (minutes)
4. E2E tests                (minutes)
5. Performance benchmarks   (optional, on PR or nightly)
```

### Parallelization Decision

```
Are tests independent (no shared state)?
  YES --> Run in parallel across CI workers.
          Split by file or test suite for even distribution.
  NO  --> Fix the shared state first.
          Then parallelize.
```

## Property-Based Testing

### When to Use

```
Is the function a pure transformation (input -> output, no side effects)?
  YES --> Does it have a wide input domain (strings, numbers, collections)?
            YES --> Property-based testing is a good fit.
            NO  --> Standard unit tests are sufficient.
  NO  --> Standard unit tests. Property-based testing adds complexity to stateful code.
```

### Good Candidates
- Serialization/deserialization roundtrips (encode then decode equals original)
- Sort functions (output is ordered, same length, same elements)
- Parsers (valid input always parses, invalid input always fails)
- Idempotent operations (applying twice equals applying once)
- Mathematical properties (commutativity, associativity, distributivity)

### Bad Candidates
- Tests that require specific assertions on specific outputs
- Code with complex side effects or external dependencies
- UI rendering or layout logic

## Performance Testing

### Decision Tree

```
Is this a user-facing API or critical path?
  |
  +-- YES --> Does it have a performance budget?
  |             YES --> Write a benchmark that asserts the budget.
  |                     Run in CI (nightly or per-PR for critical paths).
  |             NO  --> Establish a baseline, then set a budget.
  |
  +-- NO  --> Is there a known performance concern (large data, complex computation)?
                YES --> Write a benchmark. Compare before/after on optimization PRs.
                NO  --> Skip performance tests. Add them when a problem appears.
```

### Load Testing Guidance

```
Is the service deployed to production (or staging)?
  YES --> Run load tests against staging with production-like data volume.
          Measure: throughput, latency percentiles (p50, p95, p99), error rate.
          Compare against performance budgets.
  NO  --> Defer load testing until a staging environment exists.
          Use unit-level benchmarks for hot paths in the meantime.
```

## Pre-Ship Test Checklist

1. [ ] Unit tests cover all business logic branches
2. [ ] Integration tests cover database queries and external service contracts
3. [ ] E2E tests cover critical user journeys (login, core workflow, payment)
4. [ ] No flaky tests in quarantine for more than one sprint
5. [ ] CI pipeline runs tests in the correct order (fast feedback first)
6. [ ] Coverage meets project thresholds
7. [ ] Performance-critical paths have benchmarks with documented baselines
