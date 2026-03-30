# Primordia Coding Standards

## General

- Keep behavior changes explicit in commits and docs.
- Prefer small, named functions over long mixed-responsibility blocks.
- Avoid hidden implicit config; expose important runtime values.

## Backend (Go)

- Guard shared mutable world state with `sync.RWMutex`.
- Avoid mutating maps during iteration; collect IDs then apply deletes.
- Keep simulation constants centralized near configuration.
- Use structured, readable logs for connection and server lifecycle events.

## Frontend (React + TypeScript)

- Keep rendering loop in Pixi, UI state in React.
- Define payload types for websocket messages.
- Prefer CSS classes over heavy inline style objects.
- Keep fullscreen canvas behavior resilient on resize.

## Documentation

- Update docs in the same change when behavior or file ownership changes.
- Keep setup commands runnable exactly as written.
