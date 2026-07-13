include .env

GOOSE := $(shell go env GOPATH)/bin/goose
SQLC := $(shell go env GOPATH)/bin/sqlc
AIR := $(shell go env GOPATH)/bin/air
MIGRATIONS_DIR=internal/infrastructure/database/migrations
APP_ENTRY=./cmd/api
SIMULATE_DEBUG_ENTRY=./cmd/simulate-debug
SIM_DEBUG_HOME ?= ./simulation/testdata/manual_debug/home_debug.json
SIM_DEBUG_AWAY ?= ./simulation/testdata/manual_debug/away_debug.json
SIM_OUTPUT ?= ./tmp/simulation-output.json

.PHONY: migrate-up migrate-down sqlc start watch install-air test test-watch test-coverage simulate-debug simulation

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

test:
	go test ./...

test-watch:
	@if [ -x "$(AIR)" ]; then \
		$(AIR) -build.cmd "go test ./..." -build.bin "/usr/bin/true"; \
	else \
		echo "air is not installed. Run: make install-air"; \
		exit 1; \
	fi

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

simulate-debug:
	go run $(SIMULATE_DEBUG_ENTRY) -out $(SIM_OUTPUT) $(SIM_DEBUG_HOME) $(SIM_DEBUG_AWAY)

simulation: simulate-debug