---
name: Python Patterns
description: Type hints, formatting, error handling, async patterns, and project structure
scope: language-specific
languages: [python]
extends: [code-quality]
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
- **Virtual Environments:** Use `uv` for new projects. `venv` and `poetry` are acceptable in existing projects that already use them. Never install into system Python.
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
- See `security.md` for secrets management, input validation, and injection prevention.
- See `cross-cutting.md` for error taxonomy, coverage targets, and performance budget targets.
- See `observability.md` for logging standards. Use `structlog` or `logging` with JSON formatter.
- See `testing-patterns.md` for universal testing patterns.
