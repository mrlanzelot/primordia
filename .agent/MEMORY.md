# Primordia Team Memory

## Decisions

- WebSocket remains JSON-based for the current milestone.
- Frontend uses direct Pixi API instead of `@pixi/react`.

## Open Questions

- Should energy decay be constant or speed-based?
- When should reproducible seeded simulation become mandatory?

## Technical Debt

- Simulation and transport are still in a single backend file.
- Protocol is full-state broadcast (no delta updates yet).

## Recent Changes

- Simplified frontend styling and configuration.
- Refactored backend update loop into named system functions.
