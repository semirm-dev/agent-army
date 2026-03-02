---
name: python/tester
description: Python testing workflow — test type selection, pytest patterns, fixture design, mock vs real dependency decisions, async test setup, coverage analysis, and hypothesis usage.
scope: language-specific
uses_rules:
  - python/testing
  - testing-patterns
  - cross-cutting
---

# Python Tester Skill

## When to Use

Invoke this skill when:
- Writing tests for Python code
- Fixing failing pytest tests
- Improving Python test coverage
- Setting up async test infrastructure
- Adding property-based tests

## Test Type Selection for Python

```
What are you testing?
  |
  +-- Pure function (no I/O)?
  |     YES --> Unit test with @pytest.mark.parametrize
  |     NO  |
  |
  +-- Database query (SQLAlchemy/Tortoise)?
  |     YES --> Integration test with real DB + transaction rollback fixture
  |     NO  |
  |
  +-- FastAPI/Flask handler?
  |     YES --> TestClient (FastAPI) or test_client fixture (Flask) with dependency overrides
  |     NO  |
  |
  +-- External API?
  |     YES --> Protocol + fake implementation for unit. responses or respx library for integration.
  |     NO  |
  |
  +-- Async function?
  |     YES --> @pytest.mark.asyncio with pytest-asyncio
  |     NO  |
  |
  +-- Data pipeline/transformation?
        YES --> hypothesis for property-based testing if wide input domain
```

## Mock vs Real Dependency Decision

```
What dependency are you testing against?
  |
  +-- Database?
  |     YES --> Use real DB via testcontainers-python or docker-compose. Use transaction rollback per test.
  |     NO  |
  |
  +-- External HTTP API?
  |     YES --> Define Protocol. Use fake for unit tests. Use responses/respx/httpx.MockTransport for integration.
  |     NO  |
  |
  +-- File system?
  |     YES --> Use tmp_path fixture (real FS). Avoid mocking.
  |     NO  |
  |
  +-- Time?
  |     YES --> Inject clock dependency or use freezegun for time-dependent tests.
  |     NO  |
  |
  +-- Randomness?
        YES --> Inject seed or use random.seed() in fixture for determinism.
```

## Fixture Design Workflow

- **Scope:** `function` (default, most isolated) → `module` (shared DB) → `session` (shared across all tests)
- Use `yield` fixtures for setup + teardown
- Factory fixtures return callables: `def user_factory() -> Callable[..., User]`
- Compose fixtures: a `db_session` fixture depends on `db_engine` fixture
- Never share mutable state between tests via module-level fixtures without isolation

## Parametrize Design

- Use `@pytest.mark.parametrize` for data-driven tests
- Use `pytest.param(..., id="descriptive-name")` for readable test IDs
- Combine multiple `@pytest.mark.parametrize` decorators for combinatorial testing (use sparingly)
- For complex test data, define a list of dataclass/NamedTuple instances outside the decorator

## Coverage Analysis Workflow

```
pytest --cov=src --cov-report=term-missing
pytest --cov=src --cov-report=html    # visual report
```

- Focus on uncovered branches (`--cov-branch`)
- Use `# pragma: no cover` sparingly and only with justification comment
- Exclude test files and generated code from coverage

## Pre-Merge Test Checklist

- [ ] `pytest` passes with no warnings (or warnings are expected and filtered)
- [ ] `pytest --cov` meets project threshold
- [ ] No `@pytest.mark.skip` without a tracking issue
- [ ] Async tests use `pytest-asyncio`, not manual event loop management
- [ ] No test depends on execution order or shared mutable state
- [ ] Fixtures clean up resources via `yield` teardown or context managers
- [ ] Property-based tests (hypothesis) run for sufficient examples
