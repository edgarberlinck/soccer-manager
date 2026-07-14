# Soccer Manager DB-First UI Design

## Objective

Migrate auth, squad creation, dashboard, and player details to a DB-first architecture, removing local-only player classification labels from the UI and exposing full player management data.

## Requirements

- Signup and login must use backend and database as source of truth.
- Player creation must happen in backend during club bootstrap.
- Dashboard must not show internal labels (Ruim/Mediano/Bom).
- User must click a player and navigate to dedicated player page.
- Player page must show full attributes, contract, salary in EUR (reference), and match performance history.
- Preserve existing visual direction and document visual style.

## Architecture

- Frontend authenticates via `/auth/signup` and `/auth/signin`, stores JWT, and fetches protected resources.
- Backend owns club bootstrap and starter squad generation.
- New protected routes expose club roster and player details.
- New DB tables persist active contracts and per-match player stats.

## Backend Design

- Extend signup payload to include `manager_name` and club fields (`club_name`, `club_short_name`, `abbreviation`, `continent`, `country`).
- On signup, in a DB transaction:
  - create user
  - create user_meta
  - create club
  - create starter players
  - create active contracts in EUR (stored as cents)
- Return JWT on signup to avoid local temporary sessions.
- Add endpoints:
  - `GET /clubs/{clubID}/players`
  - `GET /clubs/{clubID}/players/{playerID}`

## Database Design

- Players:
  - Add `club_id`, `position`, `overall`, `potential`, `tier`.
- Contracts:
  - New `player_contracts` table with EUR salary/clauses in cents.
- Performance:
  - New `player_match_stats` table keyed by `(match_id, player_id)`.

## API Response Shape

- Roster response includes manager-facing data only (no tier labels).
- Player detail includes:
  - identity + core ratings
  - full attributes
  - active contract in EUR
  - aggregated performance summary
  - recent match performance list

## UI Design

- Replace `session.ts` local account/game state with API client + JWT storage helper.
- Dashboard:
  - load clubs and roster from API
  - remove Ruim/Mediano/Bom chips and copy
  - row/card click navigates to player route
- Dedicated player page:
  - top card with identity + salary + contract
  - attributes grid
  - KPI summary
  - match history table

## Visual Style Documentation

- Add a visual language document describing:
  - typography system
  - color tokens and theme tokens
  - surface system and card language
  - interaction and spacing patterns
  - usage guidance for future screens
