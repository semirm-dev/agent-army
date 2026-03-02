---
name: python/coder
description: Python implementation workflow — code placement, error handling patterns, async decisions, type safety enforcement, dependency management, and pre-PR verification.
scope: language-specific
uses_rules:
  - python/patterns
  - code-quality
  - cross-cutting
  - security
---

# Python Coder Skill

## When to Use

Invoke this skill when:
- Writing new Python features
- Implementing API handlers
- Building async services
- Creating CLI tools
- Writing data pipelines

## Implementation Workflow

```
Understand requirements
  |
  +-- Check existing patterns in codebase
  |     (routes, services, repositories, models)
  |
  +-- Design types first (dataclasses / Pydantic models)
  |     (models drive implementation; avoid retrofitting)
  |
  +-- Implement
  |
  +-- Write tests
  |
  +-- Run lint
  |
  +-- PR
```

**Emphasize "models first"** — define data shapes before logic.

## Code Placement Decision Tree

```
Where does this code belong?
  |
  +-- New domain/feature?
  |     YES --> src/<project>/<domain>/ package
  |
  +-- Shared utility?
  |     --> src/<project>/common/ or src/<project>/lib/
  |
  +-- API handler?
  |     --> src/<project>/<domain>/routes.py or handlers.py
  |
  +-- Business logic?
  |     --> src/<project>/<domain>/service.py
  |
  +-- Data access?
  |     --> src/<project>/<domain>/repository.py
  |
  +-- Models/schemas?
  |     --> src/<project>/<domain>/models.py (DB)
  |         src/<project>/<domain>/schemas.py (API)
  |
  +-- CLI entry point?
        --> src/<project>/cli/
```

## Sync vs Async Decision Tree

```
Is this a web framework that supports async (FastAPI, Starlette)?
  |
  +-- YES --> Use async
  |
  +-- Does the code do I/O (HTTP, DB, file)?
  |     YES --> Does it need concurrency?
  |               YES --> async
  |               NO  --> sync is fine
  |
  +-- Is this library code?
  |     --> Offer both sync and async interfaces, or pick one and document
  |
  +-- Mixing sync in async code?
        --> Use asyncio.to_thread() to offload blocking calls
            Never call blocking I/O in async functions directly
```

## Error Handling Workflow

```
What kind of error is this?
  |
  +-- Known business error?
  |     --> Define custom exception class inheriting from base domain exception
  |
  +-- Wrapping underlying error?
  |     --> raise DomainError("context") from original
  |
  +-- At API boundary?
  |     --> Catch domain exceptions in handler
  |         Map to HTTP status codes
  |         Never expose tracebacks to clients
  |
  +-- General rule
        --> Never bare except:
            Always catch specific types
```

## Dependency Addition Checklist

- [ ] Checked if stdlib covers the need
- [ ] Verified package on PyPI (maintenance, downloads, license)
- [ ] Added via `uv add` / `poetry add` / `pip install` + pinned in requirements
- [ ] Regenerated lock file
- [ ] Type stubs available or `py.typed` marker present
- [ ] No version conflicts with existing dependencies

## Pre-PR Checklist

- [ ] `ruff check .` clean (or project's linter)
- [ ] `ruff format --check .` passes (or project's formatter)
- [ ] `mypy .` or `pyright` clean (strict mode)
- [ ] `pytest` passes
- [ ] All public functions have type hints and docstrings
- [ ] No `# type: ignore` without explanation comment
- [ ] Virtual environment dependencies match lock file
