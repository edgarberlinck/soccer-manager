include .env

GOOSE := $(shell go env GOPATH)/bin/goose
SQLC := $(shell go env GOPATH)/bin/sqlc
MIGRATIONS_DIR=internal/infrastructure/database/migrations

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

sqlc:
	$(SQLC) generate