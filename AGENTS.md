# Poly-Predict: Project Conventions

## Architecture
- Microservices: 4 Go backend services + 2 Next.js frontends
- Database: Supabase PostgreSQL (prod) / Docker PostgreSQL (dev)
- Auth: Supabase Auth (user-facing) + custom JWT (admin)

## Go Conventions
- Go workspace with `go.work` linking `pkg/` and 4 services
- Shared code in `backend/pkg/` — models, db, config, response helpers
- Each service has its own `go.mod` and `internal/` package
- Use `gin-gonic/gin` for HTTP services (api, admin)
- Use `robfig/cron/v3` for background services (scraper, settler)
- Use `jackc/pgx/v5` for PostgreSQL
- Use `rs/zerolog` for structured logging
- Balance stored as BIGINT (cents): 10,000 credits = 1,000,000

## Frontend Conventions
- Next.js 15 with App Router
- Tailwind CSS + shadcn/ui components
- TypeScript strict mode
- API clients generated from OpenAPI specs
- SWR for data fetching, Zustand for state management

## Code Style
- Go: standard `gofmt`, no unused imports
- TypeScript: ESLint + Prettier defaults
- SQL migrations: numbered, idempotent where possible
- Environment variables: loaded via `godotenv`, prefixed by service name

## Testing
- Go: table-driven tests, `testing` package
- Frontend: Jest + React Testing Library

## Error Handling
- Go services return unified JSON responses via `pkg/response`
- HTTP errors use standard status codes with error detail body
- All database operations use transactions where atomicity is required
