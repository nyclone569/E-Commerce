SHELL := /bin/bash
SQLC_VERSION := v1.29.0
GOOSE_VERSION := v3.27.3
PNPM_VERSION := 11.23.0
DATABASE_URL ?= postgres://aurora:aurora@localhost:5432/aurora?sslmode=disable

.PHONY: dev down logs migrate-up migrate-status sqlc backend-check backend-integration frontend-install frontend-check check images

dev:
	docker compose up -d postgres
	@until docker compose exec -T postgres pg_isready -U "$${POSTGRES_USER:-aurora}" -d "$${POSTGRES_DB:-aurora}" >/dev/null 2>&1; do sleep 1; done
	$(MAKE) migrate-up
	docker compose up --build backend frontend

down:
	docker compose down

logs:
	docker compose logs -f backend frontend postgres

migrate-up:
	@cd backend && go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir db/migrations postgres "$(DATABASE_URL)" up

migrate-status:
	@cd backend && go run github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION) -dir db/migrations postgres "$(DATABASE_URL)" status

sqlc:
	cd backend && go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

backend-check:
	cd backend && test -z "$$(gofmt -l .)" && go vet ./... && go test ./... && go test -race ./...

backend-integration:
	cd backend && go test -tags=integration -v ./tests/integration

frontend-install:
	npx --yes pnpm@$(PNPM_VERSION) --dir frontend install --frozen-lockfile

frontend-check:
	npx --yes pnpm@$(PNPM_VERSION) --dir frontend lint
	npx --yes pnpm@$(PNPM_VERSION) --dir frontend typecheck
	npx --yes pnpm@$(PNPM_VERSION) --dir frontend test
	npx --yes pnpm@$(PNPM_VERSION) --dir frontend build

check: backend-check backend-integration frontend-check

images:
	test -n "$(GIT_SHA)"
	docker build --build-arg APP_VERSION=$(GIT_SHA) -t aurora-shop-backend:$(GIT_SHA) backend
	docker build --build-arg APP_VERSION=$(GIT_SHA) -t aurora-shop-frontend:$(GIT_SHA) frontend
