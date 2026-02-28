---
name: testing-strategy
description: "Testing pyramid guidance, test type decision tree, what NOT to test, and contract testing."
scope: universal
---

# Testing Strategy Skill

## When to Use

Invoke this skill when:
- Planning test coverage for a new feature
- Deciding which type of test to write
- Diagnosing flaky tests
- Setting up test infrastructure
- Reviewing test strategy for a module

> See `rules/testing-patterns.md` for flaky test prevention, test data factories, mocking philosophy, and async testing patterns. See `rules/cross-cutting.md` for coverage targets.

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
  YES -> Integration test (contract test or MSW/wiremock)
  NO |

Full user journey across multiple services/pages?
  YES -> E2E test
  NO -> Unit test (default to the simplest test type)
```

## What NOT to Test

- **Generated code:** protobuf stubs, OpenAPI clients, ORM migrations (test the queries, not the generated glue)
- **Trivial getters/setters:** If the function just returns a field, don't test it
- **Third-party internals:** Don't test that Express routes correctly or that React renders -- test YOUR logic
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
