.PHONY: help dev-up dev-down migrate-up migrate-down dev-api dev-admin dev-scraper dev-settler gen-api gen-admin gen-web-client gen-admin-client build-all

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Docker
dev-up: ## Start local PostgreSQL + Redis
	docker compose up -d

dev-down: ## Stop local services
	docker compose down

# Migrations
migrate-up: ## Run all migrations up
	@for f in backend/migrations/*.up.sql; do \
		echo "Running $$f..."; \
		docker exec -i poly-predict-postgres-1 psql -U polypredict -d polypredict < $$f; \
	done

migrate-down: ## Run all migrations down (reverse order)
	@for f in $$(ls -r backend/migrations/*.down.sql); do \
		echo "Running $$f..."; \
		docker exec -i poly-predict-postgres-1 psql -U polypredict -d polypredict < $$f; \
	done

# Backend services
dev-api: ## Run API service (hot reload)
	cd backend && go run ./services/api/cmd/main.go

dev-admin: ## Run Admin service (hot reload)
	cd backend && go run ./services/admin/cmd/main.go

dev-scraper: ## Run Scraper service
	cd backend && go run ./services/scraper/cmd/main.go

dev-settler: ## Run Settler service
	cd backend && go run ./services/settler/cmd/main.go

# Code generation
gen-api: ## Generate Go types from API spec
	oapi-codegen -package handler -generate types shared/api-spec.yaml > backend/services/api/internal/handler/types_gen.go

gen-admin: ## Generate Go types from Admin spec
	oapi-codegen -package handler -generate types shared/admin-spec.yaml > backend/services/admin/internal/handler/types_gen.go

gen-web-client: ## Generate TypeScript client for web frontend
	cd frontend/web && npx openapi-typescript ../../shared/api-spec.yaml -o src/lib/api/schema.d.ts

gen-admin-client: ## Generate TypeScript client for admin frontend
	cd frontend/admin-web && npx openapi-typescript ../../shared/admin-spec.yaml -o src/lib/api/schema.d.ts

# Build
build-all: ## Build all Go services
	cd backend && go build ./services/api/cmd/main.go
	cd backend && go build ./services/admin/cmd/main.go
	cd backend && go build ./services/scraper/cmd/main.go
	cd backend && go build ./services/settler/cmd/main.go

# Frontend
dev-web: ## Run web frontend
	cd frontend/web && npm run dev

dev-admin-web: ## Run admin frontend
	cd frontend/admin-web && npm run dev
