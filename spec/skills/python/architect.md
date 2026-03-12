---
name: python/architect
description: Design Python project architecture including layout selection (web API, CLI, library), module decomposition, dependency direction enforcement, sync vs async architecture decisions, and dependency injection patterns.
scope: language-specific
languages: [python]
uses_skills: [python/patterns]
---

# Python Architect Skill

## When to Use

Invoke this skill when:
- Starting a new Python service or library
- Restructuring an existing Python project
- Planning module layout
- Choosing between sync and async architecture
- Evaluating package boundaries

## Project Layout Decision Tree

```
Is this an application (web service, CLI, worker)?
  |
  +-- Web API?
  |     YES --> Framework-specific structure (FastAPI: app/, routers/, services/, models/)
  |
  +-- CLI tool?
  |     YES --> src/<package>/cli.py entry point, src/<package>/commands/
  |
  +-- Worker/pipeline?
        YES --> src/<package>/workers/, src/<package>/tasks/
  |
Is this a library?
  |
  +-- Single purpose?
  |     YES --> src/<package>/ with __init__.py as public API
  |
  +-- Multi-purpose?
        YES --> Sub-packages under src/<package>/
```

### Reference Layout

```
src/
  <project>/
    <domain>/
      __init__.py
      routes.py       (or handlers.py)
      service.py
      repository.py
      models.py        (DB models)
      schemas.py       (API schemas / Pydantic)
    common/
      config.py
      exceptions.py
    main.py
pyproject.toml
tests/
  <domain>/
    test_service.py
    test_routes.py
```

## Module Decomposition Workflow

```
Does this code represent a distinct domain concept?
  YES --> new package under src/<project>/
  NO  --> continue

Is this code shared by 3+ modules?
  YES --> extract to common/ or lib/
  NO  --> continue

Is the current module >300 lines?
  YES --> Check if it has multiple responsibilities
         YES --> split
         NO  --> refactor for clarity
  NO  --> continue

Does splitting introduce circular imports?
  YES --> Use dependency injection or move shared types to a types module
  NO  --> safe to split
```

## Dependency Direction Rules

```
main.py → domain packages → (database, external services)
         ↓
domain/A/ → domain/B/ (AVOID! use shared types or events)
         ↓
common/ ← (shared utilities, imported by domain packages)

Rules:
- Entry point (main.py) wires everything together
- Domain packages do NOT import each other
- Cross-domain communication through shared schemas/events or an orchestrator
- common/ has zero dependencies on domain packages
- External dependencies (DB sessions, HTTP clients) injected via FastAPI Depends or constructor
```

## Async Architecture Decision

```
Is the application I/O-heavy (many concurrent HTTP requests, DB queries)?
  YES --> Async framework (FastAPI with async handlers, async SQLAlchemy)
  NO  --> continue

Is it CPU-heavy (data processing, computation)?
  YES --> Sync + multiprocessing or task queue (Celery, Dramatiq)
  NO  --> continue

Mixed?
  YES --> Async for I/O, offload CPU work to thread pool (asyncio.to_thread()) or task queue
  NO  --> continue

Library code?
  YES --> Provide async interface if consumers are async.
         Use asyncio.to_thread() to wrap sync internals.
```

## Dependency Injection Pattern

- FastAPI: use `Depends()` for handler-level injection
- Non-FastAPI: use constructor injection (`Service(repo=repo, client=client)`)
- Avoid service locator patterns or global registries
- Wire dependencies in `main.py` or a `container.py` composition root
- Use Protocols for type-safe dependency interfaces (not ABC unless you need enforcement)

## Architecture Evolution Checklist

- [ ] Each module has a single clear responsibility
- [ ] No circular imports (use `ruff` or `import-linter` to enforce)
- [ ] `src/` layout with `pyproject.toml` for proper packaging
- [ ] Domain modules do not import each other directly
- [ ] External services behind Protocol interfaces for testability
- [ ] Configuration loaded via validated config module, not scattered `os.getenv()`
- [ ] `__init__.py` files only re-export the public API, no logic
- [ ] New domains can be added without modifying existing modules
