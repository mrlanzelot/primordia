# Primordia

Primordia is a simulation project exploring how complex behavior and intelligence can emerge from simple organisms and environmental pressure.

## Goals

- Build a simulation-first architecture where behavior emerges from rules, not scripted species.
- Visualize population dynamics in real time.
- Keep the system explainable and easy to iterate on.

## Current Scope

The current milestone is Phase 1: Emergence Explorer.

- Backend simulation loop in Go with internal package split
- WebSocket broadcast of world state
- Frontend visualization with React + PixiJS layered renderer, inspector, and HUD speed/lifecycle controls
- Search behavior blending wander, nearest-food cue, smell-gradient cue, and crowd avoidance

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

## Docker Image

The repository includes a production Docker image that packages:

- the Go engine API/WebSocket server
- the built frontend bundle (`frontend/dist`), served by the backend on `/`

The container listens on port `8080`.

Build the image locally:

```bash
make docker-build IMAGE_NAME=ghcr.io/<owner>/primordia IMAGE_TAG=latest
```

Build and push the image:

```bash
make docker-deploy IMAGE_NAME=ghcr.io/<owner>/primordia IMAGE_TAG=latest
```

Useful runtime routes:

- `/` serves the frontend UI
- `/health` and `/api/health` return health checks
- `/api/ws` websocket stream
- `/api/speed` and `/api/control` runtime controls

Run the container on a Docker host:

```bash
make docker-run IMAGE_NAME=ghcr.io/<owner>/primordia IMAGE_TAG=latest CONTAINER_NAME=primordia HOST_PORT=8080
```

### Easiest/Safest (Single Docker Host, No Registry)

If you only deploy to one host (for example `192.168.68.100`), the simplest safe flow is to stream the image over SSH.
This avoids opening Docker TCP ports and avoids registry credentials/setup.

```bash
make docker-deploy-host REMOTE_USER=root REMOTE_HOST=192.168.68.100 HOST_PORT=8080
```

This command will:

- build the image locally
- transfer it to the host via SSH
- load it into Docker on the host
- restart the `primordia` container

Or use compose:

```bash
IMAGE=ghcr.io/<owner>/primordia:latest make compose-up
```

For Proxmox deployment details (Docker VM and LXC), see `docs/DEPLOY_PROXMOX.md`.

## Folder Guide

- `backend/`: Go simulation engine and WebSocket server.
	- `cmd/primordia`: entrypoint
	- `internal/world`: world and tick lifecycle
	- `internal/systems`: movement/eating/sense/brain systems
	- `internal/spatial`: uniform spatial grid
	- `internal/protocol`: wire protocol snapshot types
	- `internal/server`: websocket hub
- `frontend/`: React + PixiJS visualization client.
	- `src/pixi`: stage + render layers (organism, food, sense)
	- `src/hooks/useWorldSocket.ts`: reconnecting websocket client
	- `src/components`: HUD + organism inspector UI
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

