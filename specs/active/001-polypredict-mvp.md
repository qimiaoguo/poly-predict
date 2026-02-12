# 001: Poly-Predict MVP

## Overview
Prediction competition platform where users bet virtual credits on real Polymarket outcomes.

## Status: In Progress

## Phase Checklist

### Phase 1: Scaffolding + Database
- [x] Go workspace (go.work) with pkg + 4 service modules
- [x] docker-compose.yml (PostgreSQL + Redis)
- [x] SQL migrations (001-004)
- [x] Shared pkg (model, db, config, response)
- [x] Makefile
- [x] AGENTS.md
- [x] OpenAPI specs (api-spec.yaml, admin-spec.yaml)

### Phase 2: Scraper + API Core
- [ ] Scraper: Polymarket Gamma client
- [ ] Scraper: CLOB client
- [ ] Scraper: Syncer logic (UPSERT events, price history)
- [ ] Scraper: Cron scheduler (30 min)
- [ ] API: Gin setup + Supabase JWT middleware
- [ ] API: Event handlers (list/detail/prices/categories)
- [ ] API: Auto-create user on first auth

### Phase 3: Betting + Settlement
- [ ] API: Bet placement (atomic transaction)
- [ ] Settler: Resolution detection
- [ ] Settler: Idempotent settlement
- [ ] Settler: Credit distribution
- [ ] Settler: Streak updates

### Phase 4: User Frontend (web)
- [ ] Next.js 15 + Tailwind + shadcn/ui setup
- [ ] Supabase Auth integration
- [ ] Events list page
- [ ] Event detail + bet panel
- [ ] Profile page
- [ ] Leaderboard page
- [ ] Odds chart (recharts)

### Phase 5: Admin Service + Frontend
- [ ] Admin Gin service + custom JWT auth
- [ ] Admin handlers (dashboard, users, events, settle)
- [ ] Admin Next.js frontend
- [ ] Admin login, dashboard, user/event management

### Phase 6: Rankings + Gamification
- [ ] Ranking calculation in settler
- [ ] Leaderboard API + frontend
- [ ] Achievement system
- [ ] Level progression, streak bonuses

### Phase 7: Polish + Deploy
- [ ] Vercel deployment (frontends)
- [ ] Railway/Fly.io deployment (Go services)
- [ ] Health checks, structured logging
- [ ] Rate limiting, CORS, input validation
