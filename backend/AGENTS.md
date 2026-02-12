# Backend AGENTS.md

## Build & Test

```bash
# Build a single service (output to bin/)
bash -c 'cd backend/services/api && go build -o ../../../bin/api ./cmd/main.go'

# Run all tests for a service
bash -c 'cd backend/services/api && go test ./...'

# Run a single test
bash -c 'cd backend/services/api && go test ./internal/service/ -run TestFindOutcomeIndex'
```

## Go Workspace Structure

Go workspace at `backend/go.work` links shared `pkg/` and 4 services. Each service has its own `go.mod` with `replace github.com/poly-predict/backend/pkg => ../../pkg`. Import shared code as `github.com/poly-predict/backend/pkg/...`.

- **`pkg/`** — Shared code: `model/` (data types), `db/` (connection), `config/` (env loading), `response/` (unified JSON responses)
- **`services/{api,admin,scraper,settler}/`** — Each has `cmd/main.go` entrypoint and `internal/` package with handler → service → repository layers
- **`migrations/`** — Numbered SQL migrations (001-004)

Key libraries: Gin (HTTP), pgx/v5 (PostgreSQL), zerolog (logging), robfig/cron/v3 (scheduling)

## Backend Patterns

### Response Envelope

All HTTP responses use `pkg/response`: `Success(c, data)`, `Created(c, data)`, `Error(c, status, message)`, `ValidationError(c, message)`, `Paginated(c, data, total, page, pageSize)`. Responses wrap data in `{ "success": bool, "data": ..., "error": ... }`.

### Auth Middleware

- **API service:** `RequireAuth()` and `OptionalAuth()` middleware extract Supabase JWT. Access user ID in handlers via `c.GetString("user_id")`.
- **Admin service:** Custom JWT + bcrypt login. Access admin ID via `c.GetString("admin_id")`.

### Database Conventions

- Balance stored as BIGINT cents: 10,000 credits = 1,000,000 in DB. New users start with 10,000 credits.
- Event outcomes and prices stored as JSONB arrays (e.g., `["Yes","No"]` and `["0.65","0.35"]`)
- Atomic operations use transactions with `SELECT ... FOR UPDATE` row locking to prevent race conditions
- Payout formula: `potential_payout = amount / locked_odds`

## Testing

- **All Go code changes must include corresponding test cases**
- Table-driven tests with `testing` package, files next to code (`foo.go` → `foo_test.go`)
- Handler tests: `httptest` + `gin.CreateTestContext`
- Repository tests: interfaces/mocks, no real DB in unit tests
- Cover happy path and error/edge cases

## Environment

Backend `.env` lives in this directory (`backend/.env`). Makefile runs services with `cd backend &&`, so `godotenv.Load()` picks it up automatically. See `.env.example` for required variables.

## Code Style

- Standard `gofmt`, no unused imports
