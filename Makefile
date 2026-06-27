include .env

GOOSE := $(shell go env GOPATH)/bin/goose
SQLC := $(shell go env GOPATH)/bin/sqlc
AIR := $(shell go env GOPATH)/bin/air
MIGRATIONS_DIR=internal/infrastructure/database/migrations
APP_ENTRY=./cmd/api

.PHONY: migrate-up migrate-down sqlc start watch install-air

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

sqlc:
	$(SQLC) generate

start:
	go run $(APP_ENTRY)

install-air:
	go install github.com/air-verse/air@latest

watch:
	@if [ -x "$(AIR)" ]; then \
		$(AIR) -build.cmd "go build -o ./tmp/main $(APP_ENTRY)" -build.bin "./tmp/main"; \
	else \
		echo "air is not installed. Run: make install-air"; \
		exit 1; \
	fi