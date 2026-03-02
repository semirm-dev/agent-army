---
scope: universal
languages: []
---

# Testing Patterns

## Test Naming
- Describe behavior, not implementation: "returns error when user not found" not "test getUserById."
- Test names should read as documentation. A failing test name alone should tell you what broke.
- Use a consistent naming convention per project. See language-specific files for format details.

## Test Organization
- One test file per source file, co-located or mirrored in a tests directory.
- Group related tests with describe blocks or subtests.
- Use setup/teardown hooks for shared fixture creation and resource cleanup.
- Order tests logically: happy path first, then edge cases, then error paths.

## Test Data
- Use factories with deterministic, fixed values. No random data in tests.
- Build minimal test data -- only fields relevant to the test case. Name factory variants after the scenario, not the data shape.

## Test Isolation
- **Database tests:** Wrap each test in a transaction, rollback after. No test should depend on another test's data.
- **File system tests:** Use temp directories. Clean up in teardown.
- **In-memory databases:** Use in-memory storage for fast unit tests when full database features are not needed.
- **Network isolation:** No real HTTP calls in unit tests. Use fakes or recorded responses.

## Mocking Philosophy
- Prefer fakes and thin interfaces over heavy mocking frameworks.
- Mock at system boundaries (HTTP, DB, message queues), not between internal modules.
- Use language-specific spy/mock tools (`vi.fn`, `unittest.mock`, `gomock`) only for call verification.

## Snapshot Testing
- Do not use snapshot tests. They break on every change and provide no useful signal.
- Assert specific values and behaviors instead.

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

## Flaky Test Prevention
- **No sleep-based synchronization.** Use polling, event-based waiting, or signaling mechanisms.
- **No network calls in unit tests.** Mock at the boundary.
- **Deterministic test data.** Use factories with fixed values, not random generators.
- **Isolated test state.** Each test creates its own data. No shared mutable fixtures.
- **Explicit timeouts.** Set test timeouts to catch hangs. CI timeouts should be stricter than local.

## Coverage
- Run coverage as part of CI, not just locally.
- Set coverage thresholds as CI gates. Fail the build if coverage drops below the threshold.
- Prefer branch coverage over line coverage. Line coverage misses untested conditional paths.
- See `cross-cutting.md` for coverage targets by code category.
