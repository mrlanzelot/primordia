# Primordia Architecture

## Runtime Overview

Primordia currently runs as two processes:

1. Go backend simulation engine
2. React + PixiJS frontend visualizer

The backend is the source of truth. The frontend consumes websocket state and draws it.

## Backend Flow

Main runtime flow:

1. Create world state (`internal/world`), uniform grid, and seed population
2. Start simulation tick loop (`TickRate`)
3. Accept websocket clients on `/ws` (`internal/server`)
4. Broadcast protocol snapshots at `BroadcastRate` (`internal/protocol`)

Simulation update is split into systems in `internal/systems`:

1. `sense`: fills `SenseVec` (raycast + smell + self sensors)
2. `movement`: movement, boundary clamp, energy decay, death
3. `eating`: proximity checks and energy gain via spatial query
4. `food`: food cap enforcement and random spawn

## Frontend Flow

1. Initialize Pixi Application fullscreen
2. Open websocket connection (`VITE_WS_URL`)
3. Parse incoming world messages
4. Upsert organism graphics by id
5. Redraw food points from current snapshot
6. Remove graphics for missing organisms
7. Show lightweight overlay metadata (population, connection)

## Concurrency and Safety

- World state is guarded by `sync.RWMutex`.
- Simulation loop updates world state under write lock.
- Broadcast loop reads state under read lock.
- Shutdown uses signal context and coordinated goroutine stop.

## Known Limitations

- Protocol is full-state broadcast, not delta-based.
- Brain system is currently a random-walk stub.
- No persistence layer is active yet.

## Next Structural Step

Phase 1 backend refactor is complete with package split:

- `backend/cmd/primordia`: entrypoint and process wiring
- `backend/internal/world`: world state and tick orchestration
- `backend/internal/systems`: movement/eating/sense/brain/food systems
- `backend/internal/spatial`: uniform grid
- `backend/internal/server`: websocket hub and handler
- `backend/internal/protocol`: frontend wire format
