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

## Development Workflow
- **Branch from:** `main`
- **PR target:** `main`
- **CI required:** [yes / no]
- **Deploy:** [manual / automatic on merge to main]

## Notes
- [Any project-specific conventions that override or extend global rules]
- [Known tech debt or areas to be careful with]
