# Primordia File Guide

This guide explains what each important file is responsible for.

## Root

- `README.md`: project entrypoint, setup, and documentation index.
- `HEARTBEAT.md`: current sprint and active objectives.
- `Makefile`: install/dev/build commands.
- `package.json`: root-level dependency metadata.

## Backend

- `backend/main.go`: simulation world, update systems, websocket endpoint, and process lifecycle.
- `backend/go.mod`: Go module and backend dependencies.

## Frontend

- `frontend/src/main.tsx`: React application bootstrap.
- `frontend/src/App.tsx`: websocket handling and Pixi rendering loop.
- `frontend/src/App.css`: app shell and overlay styles.
- `frontend/src/index.css`: global reset and viewport sizing.
- `frontend/index.html`: host document and app mount point.
- `frontend/vite.config.ts`: Vite dev/build config.
- `frontend/package.json`: frontend scripts and dependencies.
- `frontend/README.md`: frontend architecture notes.

## Docs

- `docs/ARCHITECTURE.md`: system architecture and data flow.
- `docs/API.md`: websocket protocol details.
- `docs/FILES.md`: this file map.

## Agent Workflow

- `.agent/BOOT.md`: startup protocol for team role.
- `.agent/AGENTS.md`: role registry and scopes.
- `.agent/spec/architecture.md`: architecture target intent.
- `.agent/spec/requirements.md`: scope and acceptance criteria.
- `.agent/MEMORY.md`: decisions, open questions, technical debt.
- `.agent/coding_standards.md`: coding constraints by layer.
- `.agent/roadmap.md`: phase roadmap.
