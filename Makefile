include .env

GOOSE := $(shell go env GOPATH)/bin/goose
SQLC := $(shell go env GOPATH)/bin/sqlc
AIR := $(shell go env GOPATH)/bin/air
PNPM ?= pnpm
MIGRATIONS_DIR=internal/infrastructure/database/migrations
APP_ENTRY=./cmd/api
SIMULATE_DEBUG_ENTRY=./cmd/simulate-debug
UI_DIR=./ui
SIM_DEBUG_HOME ?= ./simulation/testdata/manual_debug/home_debug.json
SIM_DEBUG_AWAY ?= ./simulation/testdata/manual_debug/away_debug.json
SIM_OUTPUT ?= ./tmp/simulation-output.json

.PHONY: migrate-up migrate-down sqlc start build-api watch install-air test test-watch test-coverage simulate-debug simulation ui-install ui-dev ui-routes ui-build ui-test ui-preview test-all build-all create-bots create-calendar test-calendar coverage-html

migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

sqlc:
	$(SQLC) generate

start:
	go run $(APP_ENTRY)

build-api:
	go build $(APP_ENTRY)

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

coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-calendar:
	@echo "🧪 Testing calendar package..."
	go test -v -coverprofile=calendar_coverage.out ./internal/domain/calendar/...
	@echo ""
	@echo "📊 Calendar Coverage:"
	go tool cover -func=calendar_coverage.out | grep total
	@echo ""
	@echo "Detailed coverage:"
	go tool cover -func=calendar_coverage.out

simulate-debug:
	go run $(SIMULATE_DEBUG_ENTRY) -out $(SIM_OUTPUT) $(SIM_DEBUG_HOME) $(SIM_DEBUG_AWAY)

simulation: simulate-debug

create-bots:
	@echo "🤖 Creating bots..."
	go run ./cmd/create-bots/main.go

create-calendar:
	@echo "🗓️  Creating calendar..."
	go run ./cmd/create-calendar/main.go

create-calendar-help:
	@echo "📖 Calendar creation options:"
	@echo ""
	@echo "  make create-calendar                    # Create with default options (two-legs)"
	@echo "  make create-calendar-single             # Single round (no return matches)"
	@echo "  make create-calendar-shuffle            # Shuffle fixture order"
	@echo ""
	@echo "Or use directly:"
	@echo "  go run ./cmd/create-calendar/main.go -two-legs=false"
	@echo "  go run ./cmd/create-calendar/main.go -shuffle=true"
	@echo "  go run ./cmd/create-calendar/main.go -match-duration=120 -break=200"

create-calendar-single:
	@echo "🗓️  Creating calendar (single round)..."
	go run ./cmd/create-calendar/main.go -two-legs=false

create-calendar-shuffle:
	@echo "🗓️  Creating calendar (shuffled)..."
	go run ./cmd/create-calendar/main.go -shuffle=true

demo-calendar:
	@echo "🎬 Demonstração Completa do Sistema de Calendário"
	@echo "=================================================="
	@echo ""
	@echo "1️⃣  Rodando testes com coverage..."
	@make test-calendar
	@echo ""
	@echo "2️⃣  Estatísticas:"
	@go test ./internal/domain/calendar/... -coverprofile=calendar_coverage.out > /dev/null 2>&1
	@echo "   ✅ Coverage: $$(go tool cover -func=calendar_coverage.out | grep total | awk '{print $$3}')"
	@echo "   ✅ Testes: $$(go test -v ./internal/domain/calendar/... 2>&1 | grep -c 'PASS:') casos passando"
	@echo ""
	@echo "3️⃣  Para criar o calendário:"
	@echo "   make create-calendar"
	@echo ""
	@echo "4️⃣  Para visualizar:"
	@echo "   make start"
	@echo "   curl http://localhost:8080/calendar/season"
	@echo ""
	@echo "📖 Documentação completa: CALENDAR_QUICKSTART.md"

ui-install:
	$(PNPM) --dir $(UI_DIR) install

ui-dev:
	$(PNPM) --dir $(UI_DIR) run dev

ui-routes:
	$(PNPM) --dir $(UI_DIR) run generate-routes

ui-build:
	$(PNPM) --dir $(UI_DIR) run build

ui-test:
	$(PNPM) --dir $(UI_DIR) exec vitest run --passWithNoTests

ui-preview:
	$(PNPM) --dir $(UI_DIR) run preview

test-all: test ui-test

build-all: build-api ui-build