.PHONY: dev down migrate-up migrate-down migrate-status seed backend-init frontend-init

# Docker
dev:
	docker compose up --build

down:
	docker compose down -v

# Migrations (run inside backend container)
migrate-up:
	docker compose exec backend goose -dir /app/migrations up

migrate-down:
	docker compose exec backend goose -dir /app/migrations down

migrate-status:
	docker compose exec backend goose -dir /app/migrations status

# Seed data
seed:
	docker compose exec backend /app/seed

# Backend init (run once to generate sqlc code)
backend-init:
	cd backend && go generate ./...

# Frontend init
frontend-init:
	cd frontend && npm install

# Create a new migration
migrate-create:
	@read -p "Migration name: " name; \
	goose -dir backend/migrations create $$name sql
