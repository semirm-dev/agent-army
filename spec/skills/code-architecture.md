---
name: code-architecture
description: Guides structuring new projects and modules using vertical slice architecture, package-by-feature layouts, dependency injection, and interface boundaries across Go, TypeScript, and Python.
scope: universal
languages: []
uses_skills: [code-quality]
---

# Code Architecture Skill

## When to Use

Invoke this skill when:
- Starting a new project or feature module
- Deciding between vertical slice vs layered architecture
- Creating new packages or modules
- Reviewing dependency injection patterns
- Evaluating whether to split or keep code together

## Architecture Decision Tree

```
Is this a new project or major module?
  YES |

How many bounded contexts / features?
  1-3 -> Vertical slices (package by feature)
  4+  -> Vertical slices with shared kernel

Is there significant cross-feature logic?
  YES -> Extract shared kernel package (types, interfaces)
  NO  -> Keep features fully independent
```

## Vertical Slice vs Layered

| Aspect | Vertical Slice (Recommended) | Layered |
|--------|------------------------------|---------|
| Package by | Feature | Technical concern |
| Change scope | One package per feature change | Multiple layers per feature change |
| Coupling | Low (features are independent) | High (layers depend on each other) |
| Scaling team | Easy (teams own features) | Hard (everyone touches all layers) |
| Best for | Most projects | Tiny projects, pure CRUD |

## Package-by-Feature Patterns

### Go

```
internal/
  auth/
    handler.go      # HTTP handlers
    service.go      # Business logic
    repository.go   # Data access interface
    postgres.go     # Repository implementation
    auth_test.go    # Tests
  order/
    handler.go
    service.go
    repository.go
    postgres.go
    order_test.go
  shared/
    middleware/      # Cross-cutting HTTP middleware
    types/          # Shared domain types (IDs, pagination)
```

### TypeScript

```
src/
  features/
    auth/
      auth.controller.ts
      auth.service.ts
      auth.repository.ts
      auth.types.ts
      auth.test.ts
    order/
      order.controller.ts
      order.service.ts
      order.repository.ts
      order.types.ts
      order.test.ts
  shared/
    middleware/
    types/
```

### Python

```
src/
  auth/
    __init__.py
    router.py       # FastAPI/Flask routes
    service.py      # Business logic
    repository.py   # Data access
    models.py       # SQLAlchemy/Pydantic models
    test_service.py
  order/
    __init__.py
    router.py
    service.py
    repository.py
    models.py
    test_service.py
  shared/
    middleware/
    types.py
```

## Dependency Injection Patterns

### Go: Constructor Injection (preferred)

```go
// Define interface at consumer side
type UserRepository interface {
    GetByID(ctx context.Context, id string) (*User, error)
}

// Inject via constructor
func NewAuthService(users UserRepository, log *slog.Logger) *AuthService {
    return &AuthService{users: users, log: log}
}
```

### TypeScript: Constructor Injection

```typescript
interface UserRepository {
  getByID(id: string): Promise<User | null>;
}

class AuthService {
  constructor(
    private readonly users: UserRepository,
    private readonly logger: Logger
  ) {}
}
```

### Python: Constructor Injection

```python
class AuthService:
    def __init__(self, users: UserRepository, logger: Logger) -> None:
        self._users = users
        self._logger = logger
```

## Interface Boundary Guidelines

- **Define at consumer side:** The package that _uses_ the interface defines it, not the package that implements it
- **Keep narrow:** 2-3 methods maximum (Go). If wider, split into focused interfaces
- **No leaking:** Public APIs should not expose internal types (database models, framework types)
- **Cross-boundary DTOs:** Use dedicated types for data crossing package boundaries

## Split vs Keep Heuristic

**Keep together when:**
- Types change for the same business reason
- Functions share the same data structures
- Splitting would create circular dependencies
- The package is under 500 lines

**Split when:**
- Types change for different business reasons
- The package has multiple unrelated responsibilities
- Different parts have different deployment/scaling needs
- The package exceeds 1000 lines with distinct sections

## Module Boundary Checklist

Before creating a new package/module, verify:

1. [ ] Can you describe its purpose in one sentence without "and"?
2. [ ] Does it have a clear public API (types, functions, interfaces)?
3. [ ] Are its dependencies pointing inward (toward domain, not infrastructure)?
4. [ ] Could another team work on it independently?
5. [ ] Does it avoid duplicating types/logic from existing packages?
6. [ ] Is the name descriptive, non-generic, and non-stuttering?
