---
name: python-patterns
description: Type hints, formatting, error handling, async patterns, and project structure
scope: language-specific
languages: [python]
uses_rules: [code-quality, security, cross-cutting, observability, testing-patterns]
---

# Python Coding Patterns

## Naming and Structure
- **Naming:** `snake_case` for functions/variables, `PascalCase` for classes, `UPPER_SNAKE_CASE` for constants. Prefix private with `_`.
- **Project structure:** Package by feature, use `src/` layout for libraries. Keep `__init__.py` minimal -- only re-export public API, no logic.
- **Imports:** Order: stdlib → third-party → local. Use `isort` or ruff's import sorting. No wildcard imports (`from x import *`).
- **Docstrings:** All public functions and classes must have docstrings. Use Google or NumPy style consistently per project.

## Type Safety
- **Type Hints:** Use type hints on all function signatures. Use `from __future__ import annotations` for forward references.
- **Type checking:** Use `mypy` (strict mode) or `pyright`. Run in CI. Fix all type errors before committing.
- **Dataclasses/Pydantic:** Prefer `dataclasses` or `pydantic.BaseModel` over plain dicts for structured data. Use `frozen=True` for immutable value objects.

## Error Handling
- Use specific exception types. Never bare `except:`. Wrap with context: `raise DomainError("context") from original`.
- **Context managers:** Use `with` for resource cleanup. Implement `__enter__`/`__exit__` or use `contextlib.contextmanager` / `contextlib.asynccontextmanager`.

## Dependencies and Tooling
- **Virtual Environments:** Use `uv` for new projects. `venv` and `poetry` are acceptable in existing projects that already use them. Never install into system Python.
- **Dependencies:** Pin versions in `requirements.txt` or use `pyproject.toml` with lock files. Use `uv lock`, `poetry lock`, or `pip-compile` to generate lock files.
- **Packaging:** Always use `pyproject.toml` for new projects. `setup.py` only for legacy compatibility.
- **Formatting:** Use `ruff` (preferred) or `black` for formatting. Line length 88 (black default) or 120 (ruff default). Pick one and stay consistent per project.
- **Linting:** Use `ruff check` for linting. Fix all warnings before committing.

## Configuration and State
- **Configuration:** Use environment variables via a validated config module (e.g., `pydantic-settings`).
- **Global state:** No module-level mutable state. Prefer dependency injection.
- **Path handling:** Use `pathlib.Path` over `os.path` string manipulation.
