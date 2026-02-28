---
name: containerization
description: "Container and CI/CD workflow — Dockerfile design decisions, Docker Compose for development, CI/CD caching strategies, deployment gates, image security, and health check configuration."
scope: universal
---

# Containerization Skill

## When to Use

Invoke this skill when:
- Containerizing an application (writing a new Dockerfile)
- Reviewing or optimizing an existing Dockerfile
- Setting up Docker Compose for local development
- Designing or modifying a CI/CD pipeline
- Auditing container images for security issues
- Configuring health checks for containers or orchestrators

> See `rules/observability.md` for container best practices (non-root, pinned images, minimal base, multi-stage builds), CI/CD pipeline ordering, and image tagging.

## Dockerfile Design Decision Tree

```
What are you containerizing?
  |
  +-- Compiles to a static binary (no runtime needed)?
  |     |
  |     +-- Needs C libraries or system dependencies?
  |           YES --> Multi-stage: build image → distroless/base
  |           NO  --> Multi-stage: build image → scratch (or distroless/static)
  |
  +-- Requires a language runtime (interpreter, VM)?
  |     |
  |     +-- Is it a server application?
  |     |     YES --> Multi-stage: full image (build) → slim image (run)
  |     |
  |     +-- Is it a static frontend (SPA, static site)?
  |           YES --> Multi-stage: build image → web server (nginx:alpine)
  |
  +-- Has build-time dependencies that differ from runtime?
        YES --> Multi-stage: separate build and runtime stages
        NO  --> Single stage with slim base image

Do you need shell access in the production container?
  YES --> Use distroless or slim (not scratch). Useful for debugging.
  NO  --> Use scratch or distroless. Smallest attack surface.
```

## Multi-Stage Build Patterns

Adapt multi-stage build patterns to your language's build toolchain and runtime requirements.

### .dockerignore Essentials

Every project with a Dockerfile needs a `.dockerignore` to prevent sending unnecessary context:

```
.git
.github
.env
.env.*
node_modules
__pycache__
*.pyc
dist
build
coverage
.vscode
.idea
*.md
!README.md
docker-compose*.yml
Dockerfile*
.dockerignore
```

## Docker Compose for Dev

### Service Definition Patterns

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      target: build  # Stop at build stage for dev (includes dev deps)
    ports:
      - "3000:3000"
    volumes:
      - .:/src         # Hot reload via volume mount
      - /src/deps  # Prevent host dependencies from overriding container dependencies
    env_file:
      - .env.development
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
      POSTGRES_DB: app_dev
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dev"]
      interval: 5s
      timeout: 3s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5

volumes:
  pgdata:
```

### Key Patterns

- **Volume mounts for hot reload:** Mount source code into the container so changes are reflected without rebuilding. Use framework-specific file watchers (nodemon, air, uvicorn --reload).
- **Anonymous volumes for dependencies:** Use anonymous volumes for dependency directories (without host path) to prevent host OS dependencies from overriding container dependencies. Relevant when native modules differ by platform.
- **depends_on with healthcheck:** Use `condition: service_healthy` instead of bare `depends_on`. Bare depends_on only waits for the container to start, not for the service inside it to be ready.
- **Environment variables:** Use `.env.development` files for dev config. Never commit `.env` files with real credentials. Provide a `.env.example` with placeholder values.
- **Named volumes for persistence:** Use named volumes (`pgdata`) for database data so it survives `docker compose down`. Use `docker compose down -v` only when you want a clean slate.

## CI/CD Pipeline Design

### Caching Strategies

Cache dependency directories keyed on lockfile hash for faster CI builds:

| What to Cache | Cache Key |
|---------------|-----------|
| Dependency directory | Lockfile hash (e.g., hash of `go.sum`, `package-lock.json`, `uv.lock`) |
| Build output / artifacts | Source hash or commit SHA |
| Docker layer cache | Export/import build cache or use registry cache |

> Consult your language's CI documentation for the specific cache paths.

### Deployment Gates

Before deploying to production, verify:

1. All tests pass (unit, integration, E2E where applicable)
2. Security scan reports no critical or high vulnerabilities
3. Image is signed or verified (cosign, Notary)
4. Smoke tests pass in staging
5. Rollback path is confirmed (previous known-good image tag)

## Image Security Checklist

Before shipping a container image, verify:

1. [ ] **Read-only filesystem:** Where possible, run the container with a read-only root filesystem (`--read-only` flag or security context). Use tmpfs mounts for writable paths the application needs.
2. [ ] **No unnecessary capabilities:** Drop all Linux capabilities and add back only what is needed. Default: `--cap-drop=ALL`.
3. [ ] **SBOM generated:** Software Bill of Materials produced for the image in CycloneDX or SPDX format. Store SBOM artifacts alongside release artifacts. Verify against known vulnerability databases before deploying to production.
4. [ ] **.dockerignore present:** Build context excludes `.git`, `.env`, `node_modules`, and other unnecessary files.

## Health Check Configuration

### Dockerfile HEALTHCHECK Directive

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD ["wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:3000/healthz"] \
  || exit 1
```

- `interval`: Time between checks (30s is a reasonable default).
- `timeout`: Maximum time for a single check (keep short -- 3-5s).
- `start-period`: Grace period during startup before failures count. Set this to at least your application's startup time.
- `retries`: Number of consecutive failures before marking unhealthy.

Use `wget` or `curl` depending on what is available in the base image. For scratch/distroless images, build a tiny health check binary into the image.

### Kubernetes Probes

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 15
  timeoutSeconds: 3
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /readyz
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 3
  failureThreshold: 3

startupProbe:
  httpGet:
    path: /healthz
    port: 3000
  periodSeconds: 5
  failureThreshold: 30  # 30 * 5s = 150s max startup time
```

- **Liveness (`/healthz`):** Is the process alive and not deadlocked? Failure triggers a restart. Keep this check cheap -- do not call external dependencies.
- **Readiness (`/readyz`):** Can this instance serve traffic? Check database connectivity, cache availability, and other critical dependencies. Failure removes the pod from service endpoints.
- **Startup probe:** Use for slow-starting applications (JVM warmup, large model loading, migration on boot). The startup probe runs first. Liveness and readiness probes do not start until the startup probe succeeds.

### Health Endpoint Implementation

- `/healthz` returns 200 if the process is running.
- `/readyz` returns 200 if all dependencies are connected, 503 otherwise.
- Include dependency status in the response body for debugging, but do not expose sensitive connection details.
