---
name: containerization
description: Build production-ready containers and CI/CD pipelines covering Dockerfile multi-stage builds, Docker Compose dev setup, image security hardening, deployment strategies, and health check configuration.
scope: universal
languages: []
uses_skills: [observability, security]
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

### Pipeline Order
Lint → build → test → security scan → deploy. Tests must pass before deploy — no manual bypass mechanism.

### Image Tagging
Tag images with the commit SHA. Never rely on `latest` for deployment — it is ambiguous and not rollback-friendly.

### Artifact Signing
Sign container images and verify signatures before deployment to prevent supply chain tampering. Use cosign or Notary for signing.

## Image Security Checklist

Before shipping a container image, verify:

1. [ ] **Read-only filesystem:** Where possible, run the container with a read-only root filesystem. Use tmpfs mounts for writable paths the application needs.
2. [ ] **No unnecessary capabilities:** Drop all Linux capabilities and add back only what is needed.
3. [ ] **SBOM generated:** Software Bill of Materials produced for the image in CycloneDX or SPDX format. Store SBOM artifacts alongside release artifacts. Verify against known vulnerability databases before deploying to production.
4. [ ] **.dockerignore present:** Build context excludes version control, environment files, and other unnecessary files.

## Image Lifecycle
- **Scan images in CI** before pushing to registry. Fail the pipeline on critical or high-severity CVEs.
- **Re-scan periodically.** New CVEs are discovered after image build. Schedule weekly scans of deployed images.
- **Base image updates:** Monitor base image updates and rebuild when security patches are available.
- **Pin base image versions** by digest or specific tag for reproducibility. Floating tags (`latest`) cause non-deterministic builds.

## Secrets in Containers
- **Never bake secrets into images.** No ENV with credentials, no COPY of .env files, no ARG for sensitive values.
- **Runtime injection:** Pass secrets via environment variables from a secret manager (Vault, AWS Secrets Manager, GCP Secret Manager) or orchestrator secrets (Kubernetes Secrets, Docker Secrets).
- **Build-time secrets:** Use Docker BuildKit secret mounts (`--mount=type=secret`) for secrets needed only during build (private registry tokens, license keys). These are not persisted in image layers.

## Resource Limits
Always set CPU and memory limits in deployment manifests. Containers without limits can starve co-located workloads.

## Deployment Strategy
- **Default to rolling updates.** Replace instances incrementally, verifying health at each step. Zero-downtime for stateless services.
- **Blue/green deployments:** Use for high-risk releases or when instant rollback is required. Run both versions simultaneously, switch traffic atomically.
- **Canary deployments:** Route a small percentage of traffic (1-5%) to the new version first. Monitor error rate and latency before full rollout.
- **Rollback:** Every deployment must have a documented rollback path. For container deployments, rollback means redeploying the previous image tag (commit SHA). Verify rollback works before relying on it.
- **Post-deploy verification:** Run smoke tests or synthetic checks against the new deployment. Verify health check endpoints return healthy before routing production traffic.
- **Schema migrations and application deploys are separate steps.** Apply backward-compatible migrations before deploying new code. Never couple a breaking migration with the deploy that requires it.

## Environment Parity
- **Keep dev, staging, and production as similar as possible.** Same base images, same configuration structure, same database engine.
- **Differences between environments should be limited to:** credentials, resource sizing, and feature flags. Not architecture.

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
