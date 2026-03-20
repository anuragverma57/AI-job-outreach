# Load env vars from .env
include .env
export

# --- Docker ---
.PHONY: up down logs restart ps

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose restart

ps:
	docker compose ps

# --- Database Migrations ---
MIGRATE_URL = "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"
MIGRATIONS_DIR = api-gateway/migrations

.PHONY: migrate-up migrate-down migrate-create migrate-force

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database $(MIGRATE_URL) up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database $(MIGRATE_URL) down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

migrate-force:
	@read -p "Version to force: " version; \
	migrate -path $(MIGRATIONS_DIR) -database $(MIGRATE_URL) force $$version

# --- API Gateway ---
.PHONY: run-api build-api

run-api:
	cd api-gateway && GOTOOLCHAIN=local go run cmd/server/main.go

build-api:
	cd api-gateway && GOTOOLCHAIN=local go build -o bin/server cmd/server/main.go

# --- AI Service ---
.PHONY: setup-ai run-ai

setup-ai:
	cd ai-service && python3 -m venv venv && ./venv/bin/pip install -r requirements.txt

run-ai:
	cd ai-service && ./venv/bin/uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload

# Full dev startup: Docker + migrations + both services
dev: up
	@echo "Waiting for Postgres to be ready..."
	@sleep 2
	@$(MAKE) migrate-up
	@echo "Starting AI Service in background..."
	@$(MAKE) run-ai &
	@sleep 2
	@echo "Starting API Gateway..."
	@$(MAKE) run-api
