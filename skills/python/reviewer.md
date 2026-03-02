---
name: python/reviewer
description: Python code review workflow — review order, async audit, type safety verification, performance red flags, security checklist, and anti-pattern detection.
scope: language-specific
languages: [python]
uses_rules:
  - python/patterns
  - python/testing
  - code-quality
  - cross-cutting
  - security
---

# Python Reviewer Skill

## When to Use

Invoke this skill when:
- Reviewing Python PRs
- Conducting code audits on Python codebases
- Performing pre-merge checks on Python code
- Evaluating Python library choices

## Review Order

1. **Structure** — module organization, import order, `__init__.py` cleanliness
2. **Type Safety** — type hints on all signatures, mypy/pyright compliance, no `# type: ignore` without reason
3. **Correctness** — logic, error handling chains, resource cleanup (`with` statements)
4. **Async** — async/sync boundary correctness, no blocking in async, proper TaskGroup usage
5. **Performance** — unnecessary list copies, O(n²) in disguise, N+1 queries
6. **Security** — input validation, SQL injection, secret handling
7. **Style** — naming, docstrings (defer to ruff/black for formatting)

## Python Anti-Pattern Checklist

- [ ] Bare `except:` or `except Exception:` without re-raise or specific handling
- [ ] Mutable default arguments (`def f(items=[])`)
- [ ] `import *` (wildcard imports)
- [ ] Blocking I/O call inside an `async` function
- [ ] Module-level mutable state (global dicts, lists used as singletons)
- [ ] Shell invocation with user-controlled input (command injection risk)
- [ ] Manual resource management instead of context managers (`with`)
- [ ] String formatting with `%` or `.format()` for SQL (SQL injection risk)
- [ ] `type: ignore` without an explanation comment
- [ ] Catching and silently swallowing exceptions (empty except body)

## Async Review Checklist

- [ ] No `time.sleep()` in async code (use `asyncio.sleep()`)
- [ ] No blocking I/O (file, network, DB) without `asyncio.to_thread()`
- [ ] `asyncio.TaskGroup` used for concurrent tasks (not bare `create_task`)
- [ ] Timeouts set on all async I/O operations
- [ ] Graceful shutdown handles task cancellation

## Performance Red Flags

```
List comprehension creating large intermediate list when generator would suffice?
  YES --> Use generator expression (parentheses) or yield

in check on a list where a set would be O(1)?
  YES --> Convert to set for membership tests

String concatenation in a loop instead of "".join()?
  YES --> Use str.join() or io.StringIO

DataFrame operation in a loop instead of vectorized operation?
  YES --> Prefer pandas vectorized methods

ORM query in a loop (N+1)?
  YES --> Use select_related/prefetch_related or batch query
```

## Security Review Checklist

- [ ] SQL queries parameterized (never f-strings or `.format()` for SQL)
- [ ] No dynamic code execution or unsafe deserialization on untrusted data
- [ ] User input validated at handler boundary
- [ ] Secrets loaded from environment/secret manager, not hardcoded
- [ ] File paths from user input sanitized
- [ ] Dependencies pinned and audited (`pip audit`)
