---
name: infrastructure
description: Container best practices, image scanning, secrets handling, CI/CD pipelines, and deployment strategies
scope: universal
languages: []
---

# Container & CI/CD Patterns

## Container Best Practices
- **Multi-stage builds:** Separate build and runtime stages. The final image should contain only runtime dependencies, not build tools or source code.
- **Non-root execution:** Run application processes as a non-root user in production containers. Principle of least privilege applies to container runtime.
- **Minimal base images:** Include only what the application needs to run. Fewer packages mean fewer vulnerabilities and smaller attack surface.
- **Pin base image versions** by digest or specific tag for reproducibility. Floating tags (`latest`) cause non-deterministic builds.
- **Layer ordering:** Place frequently-changing layers (application code) last in the build to maximize cache reuse.
- **HEALTHCHECK instruction:** Include a `HEALTHCHECK` in Dockerfiles so the orchestrator can detect unhealthy containers. Use the application's health endpoint, not a generic TCP check.
- **`.dockerignore`:** Maintain a `.dockerignore` file to exclude build artifacts, test files, `.git`, and local config from the build context. Prevents image bloat and unintended secret leakage.
- **Resource limits:** Always set CPU and memory limits in deployment manifests. Containers without limits can starve co-located workloads.
- **Log to stdout/stderr.** Never write logs to files inside containers. Let the orchestrator (Docker, Kubernetes) handle log collection and routing.

## Secrets in Containers
- **Never bake secrets into images.** No ENV with credentials, no COPY of .env files, no ARG for sensitive values.
- **Runtime injection:** Pass secrets via environment variables from a secret manager (Vault, AWS Secrets Manager, GCP Secret Manager) or orchestrator secrets (Kubernetes Secrets, Docker Secrets).
- **Build-time secrets:** Use Docker BuildKit secret mounts (`--mount=type=secret`) for secrets needed only during build (private registry tokens, license keys). These are not persisted in image layers.

## Image Scanning
- **Scan images in CI** before pushing to registry. Fail the pipeline on critical or high-severity CVEs.
- **Re-scan periodically.** New CVEs are discovered after image build. Schedule weekly scans of deployed images.
- **Base image updates:** Monitor base image updates and rebuild when security patches are available.

## CI/CD Pipeline
- **Pipeline order:** lint, build, test, security scan, deploy. Tests must pass before deploy — no manual bypass mechanism.
- **Image tagging:** Tag images with the commit SHA. Never rely on `latest` for deployment — it is ambiguous and not rollback-friendly.
- **Dependency caching:** Cache dependency downloads (modules, packages) across pipeline runs to reduce build times.
- **Artifact signing:** Sign container images and verify signatures before deployment to prevent supply chain tampering.

## Deployment Strategy
- **Default to rolling updates.** Replace instances incrementally, verifying health at each step. Zero-downtime for stateless services.
- **Blue/green deployments:** Use for high-risk releases or when instant rollback is required. Run both versions simultaneously, switch traffic atomically.
- **Canary deployments:** Route a small percentage of traffic (1-5%) to the new version first. Monitor error rate and latency before full rollout.
- **Rollback:** Every deployment must have a documented rollback path. For container deployments, rollback means redeploying the previous image tag (commit SHA). Verify rollback works before relying on it.
- **Post-deploy verification:** Run smoke tests or synthetic checks against the new deployment. Verify health check endpoints return healthy before routing production traffic.
- **Schema migrations and application deploys are separate steps.** Apply backward-compatible migrations before deploying new code. Never couple a breaking migration with the deploy that requires it.
