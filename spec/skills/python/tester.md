---
name: python/tester
description: Select and write Python tests by choosing test type (unit, integration, property-based), designing pytest fixtures, deciding mock vs real dependencies, configuring async test infrastructure, and analyzing coverage gaps.
scope: language-specific
languages: [python]
uses_skills: [python/patterns, testing]
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

## Test Naming
- Use `test_function_name_scenario` (e.g., `test_create_user_duplicate_email`)
- Use `class TestCreateUser:` for grouping related tests

## Parametrized Tests

Use `@pytest.mark.parametrize` for data-driven tests (Python's equivalent of table-driven tests):

```python
@pytest.mark.parametrize("input_val,expected", [
    (5, 25),
    (0, 0),
    (-3, 9),
])
def test_square(input_val: int, expected: int) -> None:
    assert square(input_val) == expected
```

## Test Isolation

- Use `yield` fixtures for resource teardown (Python-specific)
- Use `tmp_path` fixture for temporary file system tests (Python-specific)
- Scope fixtures appropriately (`function`, `module`, `session`)

## CI Parallelization
- `pytest-xdist`: `-n auto` for auto-detected parallelism
- Use `--dist loadfile` to keep test files together
- Use `pytest-timeout` to catch hung tests: `pytest --timeout=10`

## Async Testing
- Use `pytest-asyncio` for async test support
- Decorate async tests with `@pytest.mark.asyncio`:

```python
import pytest

@pytest.mark.asyncio
async def test_fetch_user() -> None:
    user = await fetch_user("user-123")
    assert user.name == "Alice"
```

- Set `asyncio_mode = "auto"` in `pyproject.toml` to avoid marking every test:

```toml
[tool.pytest.ini_options]
asyncio_mode = "auto"
```

- Use `async with` in fixtures for async resource management:

```python
@pytest.fixture
async def db_session():
    async with async_session_factory() as session:
        yield session
```

## Mocking
- Use `unittest.mock.patch` sparingly. Prefer dependency injection and fake implementations.
- Use `pytest-mock`'s `mocker` fixture for cleaner mock syntax:

```python
def test_send_email(mocker) -> None:
    mock_send = mocker.patch("myapp.email.send")
    notify_user("user-123")
    mock_send.assert_called_once_with("user-123")
```

- For async mocks, use `AsyncMock`:

```python
from unittest.mock import AsyncMock

@pytest.mark.asyncio
async def test_async_service(mocker) -> None:
    mocker.patch("myapp.client.fetch", new_callable=AsyncMock, return_value={"id": 1})
    result = await get_item(1)
    assert result["id"] == 1
```

## Property-Based Testing
- Use `hypothesis` for testing parsers, validators, and serialization logic.
- Define strategies for domain types. Seed with known edge cases via `@example()`.
- Property tests should assert invariants (round-trip equality, no exceptions), not specific outputs.
