# Project: [PROJECT_NAME]

## Build & Test
- **Build:** `[build command, e.g., go build ./..., npm run build]`
- **Test:** `[test command, e.g., go test ./... -race, npm test]`
- **Lint:** `[lint command, e.g., golangci-lint run, npx eslint .]`
- **Format:** `[format command, e.g., gofmt -w ., npx prettier --write .]`

## Architecture
- **Language(s):** [Go / TypeScript / Python / etc.]
- **Structure:** [Vertical slices / layered / monorepo / etc.]
- **Key directories:**
  - `cmd/` — [entrypoints]
  - `internal/` — [business logic]
  - `pkg/` — [shared libraries]

## Key Decisions
- [Decision 1: e.g., "Using PostgreSQL for persistence"]
- [Decision 2: e.g., "JWT for authentication, refresh tokens stored in Redis"]
- [Decision 3: e.g., "gRPC between internal services, REST for public API"]

## External Dependencies
- **Database:** [PostgreSQL / MySQL / MongoDB / etc.]
- **Cache:** [Redis / Memcached / none]
- **Message Queue:** [Kafka / RabbitMQ / none]
- **Third-party APIs:** [list any external API integrations]

## Environment Variables
| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | Database connection string | Yes |
| `PORT` | Server port | No (default: 8080) |

## Testing Strategy
- **Test runner:** `[go test / pytest / vitest / jest]`
- **Coverage tool:** `[go test -cover / pytest --cov / vitest --coverage]`
- **Minimum threshold:** `[e.g., 80% for critical paths]`
- **Integration tests:** `[describe approach, e.g., testcontainers, docker-compose test env]`

## Deployment
- **Staging URL:** `[staging URL or "N/A"]`
- **Production URL:** `[production URL or "N/A"]`
- **Deploy command:** `[e.g., git push origin main, make deploy, kubectl apply]`
- **Rollback:** `[e.g., git revert + redeploy, kubectl rollout undo]`

## Monitoring
- **Error tracking:** `[Sentry / Datadog / Rollbar / none]`
- **Dashboards:** `[Grafana / Datadog / CloudWatch / none]`
- **Alerting:** `[PagerDuty / Slack / email / none]`

## Development Workflow
- **Branch from:** `main`
- **PR target:** `main`
- **CI required:** [yes / no]
- **Deploy:** [manual / automatic on merge to main]

## Agent Overrides
- [Any project-specific agent behavior, e.g., "use testcontainers for integration tests"]
- [e.g., "py-tester should use pytest-asyncio for all async tests"]
- [e.g., "go-coder should use sqlc instead of raw SQL"]

## Sensitive Files
- [Files beyond .env that must never be committed, e.g., "config/secrets.yaml"]
- [e.g., "*.pem, *.key files in certs/ directory"]

## Notes
- [Any project-specific conventions that override or extend global rules]
- [Known tech debt or areas to be careful with]
