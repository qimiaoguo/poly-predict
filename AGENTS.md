# AGENTS.md

This file provides guidance to AI coding agents when working with code in this repository.

## Shell Quirk

zsh has broken `cd` due to gvm hooks. Always use `bash -c 'cd /path && command'` pattern for Go commands. Writing `go.mod` files directly instead of using `go mod init` from zsh also avoids wrong-directory issues.

## Architecture

Microservices: 4 Go backend services + 2 Next.js frontends, all sharing one PostgreSQL database.

- **api** (:8080) — User-facing: events, bets, rankings, profile. Supabase JWT auth.
- **admin** (:8081) — Admin panel: user/event mgmt, dashboard, force-settle. Custom JWT + bcrypt auth.
- **scraper** (cron 30m) — Syncs events from Polymarket Gamma/CLOB APIs. No HTTP server.
- **settler** (cron 5m) — Settles resolved events, distributes credits. No HTTP server.
- **web** (:3000) — User-facing Next.js frontend with Supabase Auth.
- **admin-web** (:3001) — Admin Next.js frontend with custom JWT auth.

## Build & Run Commands

```bash
# Run services in dev mode (from project root)
make dev-api          # API service :8080
make dev-admin        # Admin service :8081
make dev-scraper      # Scraper cron
make dev-settler      # Settler cron
make dev-web          # User frontend :3000
make dev-admin-web    # Admin frontend :3001
make build-all        # Build all services
```

### Infrastructure

```bash
make dev-up        # Start local PostgreSQL + Redis (Docker)
make dev-down      # Stop
make migrate-up    # Run all SQL migrations
make migrate-down  # Reverse all migrations
```

### Code Generation (from OpenAPI specs)

```bash
make gen-api            # Go types from api-spec.yaml
make gen-admin          # Go types from admin-spec.yaml
make gen-web-client     # TypeScript client for web
make gen-admin-client   # TypeScript client for admin-web
```

## API Spec (Single Source of Truth)

OpenAPI specs in `docs/api-reference/user-api.yaml` and `docs/api-reference/admin-api.yaml` are the **single source of truth** for all API contracts.

- **Before** changing any endpoint, request/response fields, or route paths, **read the spec first**
- Before writing frontend API calls, **verify field names, query param names, and HTTP methods** against the spec
- Backend routes and response structures **must match** the spec exactly
- Frontend API calls and types **must match** the spec exactly
- When spec and code disagree, fix the code (or propose a spec change first)

## Documentation (`docs/`)

Mintlify docs live in `docs/`. OpenAPI specs live at `docs/api-reference/user-api.yaml` and `docs/api-reference/admin-api.yaml` — these are the canonical copies used by code generation, frontend clients, and Mintlify.

- When **adding/removing API endpoints**, update both the spec file and navigation entries in `docs/mint.json`
- When **changing features or flows** (betting, settlement, rankings, admin), update the relevant guide in `docs/guides/`
- When **changing architecture** (new service, DB schema change, new cron job), update `docs/architecture.mdx`
- Preview locally: `make dev-docs` (starts Mintlify at :3333)

## Testing

- **All code changes (backend and frontend) must include corresponding test cases**
- Backend: Go `testing` package, table-driven tests, files next to code (`foo.go` → `foo_test.go`)
- Frontend: Vitest + React Testing Library, files next to source (`foo.tsx` → `foo.test.tsx`)

## Key Conventions

- Balance stored as BIGINT credits (1 credit = 1 in DB). New users start with 10,000 credits
- Environment variables loaded via `godotenv` (see `backend/.env.example`)
- All database operations use transactions where atomicity is required
