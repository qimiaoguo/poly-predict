# Poly-Predict

Virtual prediction market competition platform where users bet credits on real Polymarket outcomes. No real money involved.

## Architecture

```
                         Internet
                            |
            +---------------+---------------+
            |                               |
    Vercel (web)                    Vercel (admin-web)
    Next.js :3000                   Next.js :3001
            |                               |
            v                               v
    api service :8080               admin service :8081
            |                               |
            +-------+-----------+-----------+
                    |           |
             Supabase DB    Redis (TBD)
                    |
        +-----------+-----------+
        |                       |
  scraper service          settler service
  (cron 30min)             (cron 5min)
        |
  Polymarket APIs
  (Gamma + CLOB)
```

### Services

| Service | Port | Description |
|---------|------|-------------|
| **api** | 8080 | User-facing: events, bets, rankings, profile. Supabase JWT auth. |
| **admin** | 8081 | Admin panel: user/event mgmt, dashboard, force-settle. Custom JWT auth. |
| **scraper** | — | Cron every 30 min. Syncs events from Polymarket Gamma/CLOB APIs. |
| **settler** | — | Cron every 5 min. Settles resolved events, distributes credits. |
| **web** | 3000 | User-facing Next.js frontend with Supabase Auth. |
| **admin-web** | 3001 | Admin Next.js frontend with custom JWT auth. |

## Tech Stack

- **Backend:** Go (Gin) × 4 services, Go workspace
- **Database:** Supabase PostgreSQL (prod) / Docker PostgreSQL (dev)
- **Auth:** Supabase Auth (users) + custom JWT (admin)
- **Frontend:** Next.js 16, Tailwind CSS, shadcn/ui
- **Data Fetching:** SWR, Zustand
- **Charts:** Recharts
- **API Contract:** OpenAPI 3.0 specs in `shared/`

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL client (`psql`)

### Setup

1. **Clone and configure environment:**

```bash
git clone git@github.com:qimiaoguo/poly-predict.git
cd poly-predict
cp .env.example .env
# Edit .env with your Supabase credentials
```

2. **Start local database:**

```bash
make dev-up
```

3. **Run migrations:**

```bash
make migrate-up
```

4. **Install frontend dependencies:**

```bash
cd frontend/web && npm install
cd ../admin-web && npm install
```

5. **Start services** (each in a separate terminal):

```bash
make dev-scraper     # Sync Polymarket events
make dev-api         # User API
make dev-admin       # Admin API
make dev-settler     # Settlement cron
make dev-web         # User frontend
make dev-admin-web   # Admin frontend
```

## Project Structure

```
poly-predict/
├── backend/
│   ├── go.work                  # Go workspace
│   ├── pkg/                     # Shared library (models, db, config, response)
│   ├── services/
│   │   ├── api/                 # User-facing API service
│   │   ├── admin/               # Admin API service
│   │   ├── scraper/             # Polymarket sync service
│   │   └── settler/             # Settlement service
│   └── migrations/              # SQL migrations (001-004)
├── frontend/
│   ├── web/                     # User Next.js app
│   └── admin-web/               # Admin Next.js app
├── shared/
│   ├── api-spec.yaml            # User API OpenAPI spec
│   └── admin-spec.yaml          # Admin API OpenAPI spec
├── docker-compose.yml           # PostgreSQL + Redis (dev)
└── Makefile                     # Unified command entry
```

## Database Schema

- **users** — balance (BIGINT credits), frozen_balance, level, XP, streaks, stats
- **events** — synced from Polymarket, JSONB outcomes/prices, status enum
- **price_history** — time-series price data per outcome
- **bets** — user bets with locked odds and potential payout
- **settlements** — idempotent settlement records (UNIQUE on event_id)
- **credit_transactions** — full audit log of every balance change
- **rankings** — materialized leaderboard data
- **admin_users** — admin accounts with bcrypt passwords

## Key Business Logic

- **Bet placement:** Atomic transaction — locks user row, verifies balance, freezes credits, records bet and audit log
- **Settlement:** Idempotent per event — checks settlement exists, locks pending bets, distributes payouts, updates streaks, recalculates rankings
- **Payout formula:** `potential_payout = amount / locked_odds`
- **Balance:** Stored as BIGINT credits (1 credit = 1 in DB). New users start with 10,000 credits.

## API Documentation

OpenAPI specs are in `shared/`:
- `api-spec.yaml` — User-facing API (12 endpoints)
- `admin-spec.yaml` — Admin API (9 endpoints)

## License

MIT
