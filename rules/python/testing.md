---
name: python/testing
description: Pytest conventions, parametrize, fixtures, and async test support
scope: language-specific
languages: [python]
uses_rules: [python/patterns]
---

# Python Testing Patterns

## Test Naming
- Use `test_function_name_scenario` (e.g., `test_create_user_duplicate_email`)
- Use `class TestCreateUser:` for grouping related tests

## Table-Driven Tests
- Use `@pytest.mark.parametrize` for data-driven tests:

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
- Use `yield` fixtures for resource teardown
- Use `tmp_path` fixture for temporary file system tests
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
