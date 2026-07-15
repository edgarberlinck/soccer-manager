# Soccer Manager API + UI

Full-stack football manager project with tick-based match simulation, event scheduling, JWT authentication, PostgreSQL persistence, and a React frontend.

## 1. Project Vision

This project models a football manager experience around three core pillars:

1. Club operations with real user accounts (signup, signin, club ownership, squad management).
1. Match and training simulation with domain rules isolated from HTTP concerns.
1. A living world approach powered by a tick-driven calendar and a concurrent scheduler.

In practice, a user creates an account, receives a club with a starter squad, manages players, and can evolve into richer simulation workflows over time.

## 2. What Is Already Implemented

1. JWT authentication with session endpoint (`/auth/me`).
1. Signup flow with automatic club and starter squad bootstrap.
1. Login-time self-healing: if a user has no club, backend creates club + squad during `signin`.
1. Club and player endpoints (list + full player detail).
1. Tick simulation engine isolated in `simulation/`.
1. Scheduler for pending trainings and in-progress match ticks.
1. Season calendar by ticks with match batch planning.
1. Frontend dashboard with squad view and complete player profile screen.

## 3. Technical Stack

### Backend

1. Go 1.26.4
1. Chi (HTTP router)
1. JWT (`github.com/golang-jwt/jwt/v5`)
1. PostgreSQL via `pgx`
1. SQLC for typed query generation
1. Goose for migrations
1. Cron (`robfig/cron`) for cron-based tick mode

### Frontend

1. React 19 + TypeScript
1. TanStack Router / TanStack Start
1. Vite + Nitro
1. Tailwind CSS v4
1. Vitest

## 4. Architecture and Key Decisions

### 4.1 Layered Separation

1. `internal/domain/*`: pure business/domain rules (calendar, training, etc).
1. `engine/`: orchestration facade over domain and simulation packages.
1. `simulation/`: match simulation and debug workflows.
1. `internal/api/`: HTTP layer (auth, clubs, CORS, middleware).
1. `internal/infrastructure/database/*`: SQLC queries, migrations, repository helpers.
1. `internal/infrastructure/scheduler/`: tick loop and concurrency orchestration.

Why: evolve business logic independently from transport and framework details.

### 4.2 SQL-First Persistence

1. Explicit SQL in `internal/infrastructure/database/queries/*.sql`.
1. Typed generated code through SQLC.
1. Versioned schema evolution through Goose migrations.

Why: predictable schema evolution, explicit query behavior, and stronger DB control.

### 4.3 Resilience at Entry Point

1. `SignIn` calls `ensureUserHasClubAndSquad`.
1. If user data is partially inconsistent (user exists, no club), backend auto-recovers.

Why: fix critical consistency issues on the backend path itself, not only in UI flows.

### 4.4 Tick-Based Simulation Runtime

1. Tick mode can run by interval (`SIMULATION_TICK_SECONDS`) or cron (`SIMULATION_TICK_CRON`).
1. Scheduler processes tasks through a configurable worker pool.
1. Debug simulation outputs structured JSON in `tmp/simulation-output.json`.

Why: deterministic timeline, reproducibility, and controllable concurrency under load.

### 4.5 Environment-Driven CORS

1. Dedicated CORS middleware in backend.
1. Allowed origins controlled by `CORS_ALLOWED_ORIGINS`.

Why: avoid preflight failures while keeping origin policy explicit per environment.

## 5. Main Folder Structure

```text
cmd/                    Entrypoints (api, simulate-debug)
engine/                 Game engine facade
internal/api/           HTTP handlers, middleware, router
internal/config/        Environment loading and validation
internal/domain/        Pure domain logic (calendar, training, etc)
internal/infrastructure/database/
  migrations/           Schema evolution (Goose)
  queries/              SQL sources (SQLC)
  generated/            SQLC-generated typed code
internal/infrastructure/scheduler/
simulation/             Match simulation engine and tests
ui/                     React/TanStack frontend
docs/                   Supplementary documentation
```

## 6. Prerequisites

1. Go 1.26.4
1. Node.js 20+
1. pnpm
1. Docker + Docker Compose (recommended for local DB)
1. Go tools:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## 7. Environment Setup

### 7.1 Backend

1. Copy environment template:

```bash
cp .env.example .env
```

1. Adjust values for your local setup. Core variables:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/soccer_manager?sslmode=disable
APP_BASE_URL=http://localhost:8080
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173
AUTH_JWT_SECRET=replace-with-a-strong-secret
AUTH_JWT_EXPIRATION_MINUTES=60
AUTH_VERIFY_TOKEN_TTL_MINUTES=1440
```

Note: if you run the API on `:8000`, keep `PORT`, `APP_BASE_URL`, and UI API base aligned.

### 7.2 Frontend

1. Configure UI environment:

```bash
cp ui/.env.example ui/.env
```

1. Set API URL:

```env
VITE_API_BASE_URL=http://localhost:8000
```

Use `http://localhost:8080` if backend runs on the default `.env.example` port.

## 8. Running Locally

### 8.1 Start Database (Docker)

```bash
docker compose up -d
```

Compose starts PostgreSQL 17 at `localhost:55432` with:

1. DB: `game`
1. User: `game`
1. Password: `game`

### 8.2 Apply Migrations

1. Ensure `DATABASE_URL` points to your active DB.
1. Run:

```bash
make migrate-up
```

### 8.3 Generate SQLC Code (when SQL changes)

```bash
make sqlc
```

### 8.4 Start API

```bash
make start
```

### 8.5 Start UI

In another terminal:

```bash
make ui-install
make ui-dev
```

UI default URL: `http://localhost:3000`

## 9. Functional End-to-End Flow

1. Create account at `/signup`.
1. Sign in at `/login`.
1. Dashboard loads authenticated user's club and squad.
1. Click a player to open full profile.
1. If user has no club, `signin` auto-recovers by creating club + squad.

## 10. Most Used Commands

### Backend

```bash
make start          # run API
make watch          # hot reload with air (if installed)
make build-api      # build API binary
make test           # go test ./...
make test-coverage  # coverage summary
```

### Database

```bash
make migrate-up
make migrate-down
make sqlc
```

### Frontend

```bash
make ui-install
make ui-dev
make ui-routes
make ui-build
make ui-test
make ui-preview
```

### Full project

```bash
make test-all
make build-all
```

## 11. Manual Simulation and Debug

Run debug simulation from JSON squads and persist artifact:

```bash
make simulate-debug
```

Or run with fixed seed:

```bash
go run ./cmd/simulate-debug -seed 21 ./simulation/testdata/manual_debug/home_debug.json ./simulation/testdata/manual_debug/away_debug.json
```

Primary outputs:

1. Tick-by-tick console snapshots.
1. `tmp/simulation-output.json` with score, snapshots, calendar, and performance summary.

## 12. Main API Endpoints

### Auth

1. `POST /auth/signup`
1. `POST /auth/signin`
1. `GET /auth/me` (JWT required)
1. `GET /auth/verify?token=...`

### Clubs and players (JWT required)

1. `GET /clubs`
1. `POST /clubs/ensure`
1. `POST /clubs`
1. `POST /clubs/{clubID}/ensure-squad`
1. `GET /clubs/{clubID}/players`
1. `GET /clubs/{clubID}/players/{playerID}`

Note: `GET /health` is currently inside an authenticated group and therefore requires JWT.

## 13. Testing and Quality Gates

Recommended checks before committing:

```bash
go test ./...
make ui-build
```

Focused simulation/scheduler checks:

```bash
go test ./simulation ./cmd/simulate-debug ./internal/domain/calendar ./internal/infrastructure/scheduler
```

Detailed simulation validation guide:

1. See `docs/simulation-test-plan.md`.

## 14. Troubleshooting

### URL changes but page does not switch on player click

1. Confirm dashboard parent route renders child route via `Outlet`.
1. Run `make ui-build` to confirm route tree consistency.

### `failed to fetch` during login/signup

1. Verify `VITE_API_BASE_URL` in `ui/.env`.
1. Verify actual backend port (`PORT`).
1. Restart UI after `.env` changes.

### CORS / preflight issues

1. Update `CORS_ALLOWED_ORIGINS` in backend environment.
1. Ensure frontend origin is included.

### Migration or bootstrap errors

1. Run `make migrate-up`.
1. Verify `DATABASE_URL` targets the expected DB instance.

## 15. Suggested Technical Roadmap

1. Add deeper scheduler observability (metrics/tracing).
1. Expand simulation regression testing by deterministic seeds.
1. Expose season/calendar endpoints for frontend consumption.
1. Implement richer transfer-market and club-economy flows.
1. Expand auth policy (refresh tokens, revocation strategy).

## 16. Internal References

1. `docs/simulation-test-plan.md`
1. `docs/superpowers/specs/2026-07-14-training-system-design.md`
1. `docs/superpowers/plans/2026-07-14-training-system-implementation.md`

## 17. License

Define the repository license (MIT, Apache-2.0, etc.) according to team decision.
