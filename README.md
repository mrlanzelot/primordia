# Primordia

Primordia is a simulation project exploring how complex behavior and intelligence can emerge from simple organisms and environmental pressure.

## Goals

- Build a simulation-first architecture where behavior emerges from rules, not scripted species.
- Visualize population dynamics in real time.
- Keep the system explainable and easy to iterate on.

## Current Scope

The current milestone is Phase 1: Emergence Explorer.

- Backend simulation loop in Go
- WebSocket broadcast of world state
- Frontend visualization with React + PixiJS

## Prerequisites

- Go 1.25.0+
- Node.js 18+
- npm 9+

## Quick Start

From the repository root:

```bash
make install
make dev
```

Then open `http://localhost:5173`.

The backend runs on `:8080` and serves WebSocket updates at `/ws`.

## Build

```bash
make build
```

Backend binary output:

- `bin/primordia-engine`

## Folder Guide

- `backend/`: Go simulation engine and WebSocket server.
- `frontend/`: React + PixiJS visualization client.
- `docs/`: Project documentation (architecture, protocol, file map).
- `.agent/`: Team workflow, standards, requirements, roadmap.
- `HEARTBEAT.md`: Current sprint/objectives status.
- `Makefile`: Install, dev, and build entry points.

## Documentation Index

- `docs/ARCHITECTURE.md`: Runtime architecture and data flow.
- `docs/API.md`: Current WebSocket payload contract.
- `docs/FILES.md`: What each important file does.
- `frontend/README.md`: Frontend implementation notes.
- `.agent/spec/requirements.md`: Milestone requirements and acceptance criteria.

## Design Principles

- Prefer simple, explicit systems over hidden framework magic.
- Keep docs synchronized with code changes.
- Keep frontend rendering fast and deterministic.

