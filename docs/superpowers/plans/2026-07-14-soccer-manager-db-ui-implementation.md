# Soccer Manager DB UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make auth, club bootstrap, roster, and player details fully DB-backed and expose a dedicated player page in UI.

**Architecture:** Extend auth signup transaction to bootstrap club + players + contracts, add club player APIs, and migrate frontend from localStorage game state to JWT + API fetching. Preserve visual language while documenting it.

**Tech Stack:** Go, Chi, PostgreSQL, Goose, sqlc, TanStack Router/Start, TypeScript, Tailwind.

## Global Constraints

- Keep API protected endpoints behind JWT middleware.
- Do not expose internal player tier labels in UI responses.
- Salary must be represented as EUR for reference (stored in cents).
- Preserve current UI visual language while documenting tokens/components.

---

### Task 1: Database schema for squad ownership and player intelligence

**Files:**

- Modify: `internal/infrastructure/database/migrations/`
- Modify: `internal/infrastructure/database/queries/player.sql`
- Modify: `internal/infrastructure/database/queries/club.sql`
- Modify: `internal/infrastructure/database/queries/auth.sql`
- Regenerate: `internal/infrastructure/database/generated/*.go`

- [ ] Add migration for player ownership, contracts, and match stats.
- [ ] Add sqlc queries for roster listing, player detail, active contract, and performance aggregates.
- [ ] Regenerate sqlc code.
- [ ] Run focused DB/query tests.

### Task 2: Auth signup and backend squad bootstrap

**Files:**

- Modify: `internal/api/auth_handler.go`
- Modify: `internal/api/club_handler.go`
- Modify: `internal/api/router.go`
- Create/Modify: `internal/domain/player/*` or helper in API package

- [ ] Extend signup request/validation to include manager and club metadata.
- [ ] Implement transactional signup + club + starter squad + contracts.
- [ ] Return JWT from signup for immediate authenticated UX.
- [ ] Keep signin DB-backed and compatible.

### Task 3: Player API routes

**Files:**

- Modify: `internal/api/club_handler.go`
- Modify: `internal/api/router.go`

- [ ] Add `GET /clubs/{clubID}/players`.
- [ ] Add `GET /clubs/{clubID}/players/{playerID}`.
- [ ] Enforce ownership checks using authenticated `user_id`.
- [ ] Hide internal tier in responses.

### Task 4: Frontend auth/session migration and dashboard rewrite

**Files:**

- Modify: `ui/src/lib/session.ts`
- Modify: `ui/src/routes/signup.tsx`
- Modify: `ui/src/routes/login.tsx`
- Modify: `ui/src/routes/dashboard.tsx`
- Create: `ui/src/routes/dashboard.player.$playerId.tsx` (or equivalent file-route)

- [ ] Replace localStorage account model with token + API client.
- [ ] Wire signup/login to backend.
- [ ] Load dashboard from API and remove Ruim/Mediano/Bom labels.
- [ ] Add navigation to dedicated player route and render full profile/performance.

### Task 5: Visual language documentation and verification

**Files:**

- Create: `docs/ui-visual-style.md`

- [ ] Document current visual language and reusable design rules.
- [ ] Run backend tests and frontend checks.
- [ ] Fix regressions and finalize.
