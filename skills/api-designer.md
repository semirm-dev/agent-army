---
name: api-designer
description: REST/gRPC API design patterns, error format scaffolding, and pagination helpers following project API design standards.
---

# API Designer Skill

## When to Use

Invoke this skill when:
- Designing new API endpoints
- Scaffolding error response formats
- Implementing pagination
- Reviewing API design for consistency

## Endpoint Design Checklist

For each new endpoint, verify:

1. **Resource naming:** Plural nouns, kebab-case (`/users`, `/order-items`)
2. **HTTP method:** GET (read), POST (create), PUT (replace), PATCH (update), DELETE (remove)
3. **Status codes:** Correct for the operation (see table below)
4. **Request validation:** At handler boundary, return 400 with field errors
5. **Error response:** Uses standard error format
6. **Pagination:** Cursor-based for large datasets
7. **Idempotency:** POST has idempotency key, PUT/DELETE are idempotent
8. **Versioning:** URL path (`/api/v1/`) for public APIs

## Error Response Format

All endpoints must return errors in this structure:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "Human-readable description of what went wrong",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address"
      }
    ]
  }
}
```

**Error codes by category:**
- `VALIDATION_FAILED` — 400: Input validation errors
- `UNAUTHORIZED` — 401: Missing or invalid authentication
- `FORBIDDEN` — 403: Authenticated but not authorized
- `NOT_FOUND` — 404: Resource does not exist
- `CONFLICT` — 409: Resource state conflict (duplicate, version mismatch)
- `RATE_LIMITED` — 429: Too many requests
- `INTERNAL_ERROR` — 500: Unexpected server error

## Pagination Pattern

### Cursor-based (recommended for large datasets)

**Request:**
```
GET /api/v1/users?limit=20&cursor=eyJpZCI6MTIzfQ==
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "hasMore": true,
    "nextCursor": "eyJpZCI6MTQzfQ=="
  }
}
```

### Offset-based (acceptable for small, stable datasets)

**Request:**
```
GET /api/v1/categories?limit=20&offset=40
```

**Response:**
```json
{
  "data": [...],
  "pagination": {
    "total": 85,
    "limit": 20,
    "offset": 40
  }
}
```

## Status Code Reference

| Code | Meaning | When to Use |
|------|---------|-------------|
| 200 | OK | Successful GET, PUT, PATCH, DELETE |
| 201 | Created | Successful POST that creates a resource |
| 204 | No Content | Successful DELETE with no response body |
| 400 | Bad Request | Validation errors, malformed request |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Authenticated but lacks permission |
| 404 | Not Found | Resource does not exist |
| 409 | Conflict | Duplicate resource, version conflict |
| 422 | Unprocessable | Syntactically valid but semantically wrong |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Error | Unexpected server error |

## Scaffold Workflow

When designing a new API:

1. Define resources and their relationships
2. Map CRUD operations to endpoints
3. Define request/response schemas
4. Implement error handling with standard format
5. Add pagination for list endpoints
6. Add rate limiting for public endpoints
7. Document with OpenAPI/Swagger
