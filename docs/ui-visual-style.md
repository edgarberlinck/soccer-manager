# UI Visual Style Guide

## Visual Direction

- Theme: night-stadium dashboard with neon pitch accents.
- Mood: tactical, energetic, readable for long management sessions.
- Surfaces use layered dark greens with luminous lime highlights.

## Typography

- Display family: Orbitron for major titles (`.title-xl`, `.title-lg`, `.title-md`).
- UI/body family: Rajdhani for labels, body copy, forms, and chips.
- Eyebrow style: uppercase, wide tracking, accent color (`.eyebrow`).

## Color System

- Core tokens are defined in `ui/src/styles.css` under `:root` and light theme overrides.
- Key tokens:
  - `--sea-ink`: primary text
  - `--text-muted`: supporting text
  - `--line`: border/separator system
  - `--accent` and `--accent-2`: CTA and highlight gradients
  - `--bg-1`, `--bg-2`, `--bg-3`: atmospheric layered backgrounds

## Surfaces and Components

- Main containers use `.pitch-card` and `.dashboard-hero`.
- Card language:
  - Soft bordered, rounded, layered gradients.
  - Subtle lift on hover (`.feature-tile`, `.player-card`).
- Action hierarchy:
  - Primary action: `.btn-primary` gradient button.
  - Secondary action: `.btn-ghost` outlined translucent button.
- Data display:
  - Compact metrics via `.stat-chip`.

## Motion and Interaction

- Hover transforms are subtle (`translateY(-1px/-2px)`), paired with border brighten.
- Use transitions around 180ms for responsiveness without visual noise.

## Layout and Spacing

- Global width constrained by `.page-wrap`.
- Vertical rhythm controlled by `.page-gutter` and section spacing (`mt-*`).
- Mobile-first with border radius and gutters adjusted under 640px.

## Accessibility Notes

- Keep high text contrast against dark layered surfaces.
- Preserve semantic headings and button usage for interactive cards.
- Keep focus styles visible on all form controls.

## Future Screen Guidance

- Reuse tokenized colors and card classes instead of adding one-off palettes.
- Prefer extending existing components (`pitch-card`, `stat-chip`) for consistency.
- New manager screens should preserve tactical/football language without exposing internal-only labels to users.
