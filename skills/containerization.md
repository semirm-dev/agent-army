---
name: containerization
description: Container and CI/CD workflow — Dockerfile design decisions, Docker Compose for development, CI/CD caching strategies, deployment gates, image security, and health check configuration.
scope: universal
languages: []
uses_rules:
  - infrastructure
  - observability
  - security
  - cross-cutting
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

Every multi-stage build follows the same principle: the build stage has everything needed to compile/bundle, the runtime stage has only what's needed to run.

1. **Build stage:** Start from a full SDK/toolchain image. Copy source and dependency manifests. Install dependencies, then copy source and run the build command.
2. **Runtime stage:** Start from a minimal base image (scratch, distroless, alpine, or slim). Copy only the built artifact from the build stage. Set the entrypoint.
3. **Layer ordering:** Copy dependency manifests and install dependencies before copying source code. This maximizes Docker layer cache hits — dependencies change less often than source.

Adapt the specific base images, build commands, and artifact paths to your language's toolchain.

### .dockerignore

Every project with a Dockerfile needs a `.dockerignore` to prevent sending unnecessary context. Exclude version control directories, environment files, dependency caches, IDE configuration, build artifacts, and Docker-related files. Keep only what the build needs.

## Docker Compose for Dev

### Key Patterns

- **Volume mounts for hot reload:** Mount source code into the container so changes are reflected without rebuilding. Use framework-specific file watchers.
- **Anonymous volumes for dependencies:** Use anonymous volumes for dependency directories (without host path) to prevent host OS dependencies from overriding container dependencies. Relevant when native modules differ by platform.
- **depends_on with healthcheck:** Use `condition: service_healthy` instead of bare `depends_on`. Bare depends_on only waits for the container to start, not for the service inside it to be ready.
- **Environment variables:** Use `.env.development` files for dev config. Never commit `.env` files with real credentials. Provide a `.env.example` with placeholder values.
- **Named volumes for persistence:** Use named volumes for database data so it survives container restarts. Use volume removal only when you want a clean slate.

## CI/CD Pipeline Design

### Caching Strategies

Cache dependency directories keyed on lockfile hash for faster CI builds:

| What to Cache | Cache Key |
|---------------|-----------|
| Dependency directory | Lockfile hash (e.g., hash of `go.sum`, `package-lock.json`, `uv.lock`) |
| Build output / artifacts | Source hash or commit SHA |
| Docker layer cache | Export/import build cache or use registry cache |

Consult your language's CI documentation for the specific cache paths.

### Deployment Gates

Before deploying to production, verify:

1. All tests pass (unit, integration, E2E where applicable)
2. Security scan reports no critical or high vulnerabilities
3. Image is signed or verified (cosign, Notary)
4. Smoke tests pass in staging
5. Rollback path is confirmed (previous known-good image tag)

## Image Security Checklist

Before shipping a container image, verify:

1. [ ] **Read-only filesystem:** Where possible, run the container with a read-only root filesystem. Use tmpfs mounts for writable paths the application needs.
2. [ ] **No unnecessary capabilities:** Drop all Linux capabilities and add back only what is needed.
3. [ ] **SBOM generated:** Software Bill of Materials produced for the image in CycloneDX or SPDX format. Store SBOM artifacts alongside release artifacts. Verify against known vulnerability databases before deploying to production.
4. [ ] **.dockerignore present:** Build context excludes version control, environment files, and other unnecessary files.

## Health Check Configuration

### Container Health Check

Configure health checks with these parameters:

- **interval:** Time between checks (30s is a reasonable default)
- **timeout:** Maximum time for a single check (keep short -- 3-5s)
- **start-period:** Grace period during startup before failures count. Set this to at least your application's startup time.
- **retries:** Number of consecutive failures before marking unhealthy

Use a lightweight HTTP check against your application's health endpoint. For minimal images without HTTP clients, build a tiny health check binary into the image.

### Orchestrator Probes

Configure three probe types for container orchestrators:

- **Liveness (`/healthz`):** Is the process alive and not deadlocked? Failure triggers a restart. Keep this check cheap -- do not call external dependencies.
- **Readiness (`/readyz`):** Can this instance serve traffic? Check database connectivity, cache availability, and other critical dependencies. Failure removes the instance from service endpoints.
- **Startup probe:** Use for slow-starting applications (JVM warmup, large model loading, migration on boot). The startup probe runs first. Liveness and readiness probes do not start until the startup probe succeeds.

### Health Endpoint Implementation

- `/healthz` returns 200 if the process is running.
- `/readyz` returns 200 if all dependencies are connected, 503 otherwise.
- Include dependency status in the response body for debugging, but do not expose sensitive connection details.
