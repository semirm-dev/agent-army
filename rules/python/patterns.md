---
scope: language-specific
languages: [python]
---

> Extends `code-quality.md`. Language-agnostic standards apply.

# Python Coding Patterns
- **Type Hints:** Use type hints on all function signatures. Use `from __future__ import annotations` for forward references.
- **Type checking:** Use `mypy` (strict mode) or `pyright`. Run in CI. Fix all type errors before committing.
- **Formatting:** Use `ruff` (preferred) or `black` for formatting. Line length 88 (black default) or 120 (ruff default). Pick one and stay consistent per project.
- **Linting:** Use `ruff check` for linting. Fix all warnings before committing.
- **Imports:** Order: stdlib → third-party → local. Use `isort` or ruff's import sorting. No wildcard imports (`from x import *`).
- **Naming:** `snake_case` for functions/variables, `PascalCase` for classes, `UPPER_SNAKE_CASE` for constants. Prefix private with `_`.
- **Project structure:** Package by feature, use `src/` layout for libraries. Keep `__init__.py` minimal -- only re-export public API, no logic.
- **Virtual Environments:** Always use `uv` (preferred), `venv`, or `poetry` for dependency isolation. Never install into system Python.
- **Dependencies:** Pin versions in `requirements.txt` or use `pyproject.toml` with lock files. Use `uv lock`, `poetry lock`, or `pip-compile` to generate lock files.
- **Packaging:** Always use `pyproject.toml` for new projects. `setup.py` only for legacy compatibility.
- **Error Handling:** Use specific exception types. Never bare `except:`. Wrap with context: `raise DomainError("context") from original`.
- **Docstrings:** All public functions and classes must have docstrings. Use Google or NumPy style consistently per project.
- **Configuration:** Use environment variables via a validated config module (e.g., `pydantic-settings`).
- **Global state:** No module-level mutable state. Prefer dependency injection.
- **Dataclasses/Pydantic:** Prefer `dataclasses` or `pydantic.BaseModel` over plain dicts for structured data. Use `frozen=True` for immutable value objects.
- **Path handling:** Use `pathlib.Path` over `os.path` string manipulation.
- **Context managers:** Use `with` for resource cleanup. Implement `__enter__`/`__exit__` or use `contextlib.contextmanager` / `contextlib.asynccontextmanager`.

## Cross-References
> See `security.md` for secrets management, input validation, and injection prevention.
> See `cross-cutting.md` for error taxonomy and coverage targets.
> See `observability.md` for logging standards. Use `structlog` or `logging` with JSON formatter.

## Concurrency
> See `concurrency.md` for universal patterns (deadlocks, backpressure, shutdown).

### asyncio
- **`asyncio.gather`:** Run multiple coroutines concurrently. Use `return_exceptions=True` for partial failure handling.
- **`asyncio.TaskGroup`** (3.11+): Structured concurrency — all tasks cancel on first failure
- **`asyncio.create_task`:** Schedule coroutine execution. Always hold a reference to the task.
- **`asyncio.to_thread`:** Offload blocking/CPU-bound work to a thread pool

### Sync Primitives
- **`asyncio.Lock`:** Protect shared async state. Use `async with lock:`.
- **`asyncio.Semaphore`:** Limit concurrent access (e.g., max 10 DB connections)
- **`asyncio.Event`:** Signal between coroutines

### ThreadPoolExecutor
- Use for CPU-bound work or legacy blocking libraries
- Set `max_workers` explicitly based on workload type
- Use `asyncio.to_thread` (3.9+) instead of raw executor

### Pitfalls
- **Forgetting to await:** Unawaited coroutines silently don't execute. Enable `RuntimeWarning`.
- **Blocking the event loop:** `time.sleep()`, synchronous I/O in async context. Use `asyncio.sleep()`.
- **Task reference lost:** `create_task()` returns a task — if you don't hold a reference, it can be GC'd.

> See `testing-patterns.md` for universal testing patterns.

## Recommended Stack

### Database
> See `database.md` for universal patterns.
- **ORM (sync):** SQLAlchemy 2.0+ — Core for complex queries, ORM for CRUD operations
- **ORM (async):** Tortoise ORM — async-first ORM for FastAPI and other async frameworks
- **Migrations:** Alembic (with SQLAlchemy) or Aerich (with Tortoise ORM)

### Messaging
> See `messaging-patterns.md` for universal patterns.
- **celery:** Distributed task queue with multiple broker backends (Redis, RabbitMQ).
- **dramatiq:** Simple, reliable task processing with automatic retries.

### Observability
> See `observability.md` for universal patterns.
- **OTel:** Use `opentelemetry-instrumentation` packages for Flask, FastAPI, SQLAlchemy, requests/httpx
- **Logging:** Use `structlog` or `logging` with JSON formatter

## Performance Budgets
> See `cross-cutting.md` for performance budget targets.
