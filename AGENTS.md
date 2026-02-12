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

OpenAPI specs in `shared/api-spec.yaml` and `shared/admin-spec.yaml` are the **single source of truth** for all API contracts.

- **Before** changing any endpoint, request/response fields, or route paths, **update the spec first**
- Backend routes and response structures **must match** the spec exactly
- Frontend API calls and types **must match** the spec exactly
- When spec and code disagree, fix the code (or propose a spec change first)

## Key Conventions

- Balance stored as BIGINT cents: 10,000 credits = 1,000,000 in DB
- Environment variables loaded via `godotenv` (see `backend/.env.example`)
- All database operations use transactions where atomicity is required
