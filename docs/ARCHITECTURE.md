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
2. `movement`: movement, boundary clamp, smell-guided steering, organism crowd-avoidance, energy decay, death
3. `eating`: proximity checks and energy gain via spatial query
4. `food`: food cap enforcement and random spawn

## Frontend Flow

1. Initialize Pixi Application fullscreen
2. Open websocket connection (`VITE_WS_URL`)
3. Parse incoming world messages via reconnecting socket hook
4. Upsert food and organism graphics via pooled layer objects
5. Draw sense rays only for locally selected organism
6. Update HUD and fixed inspector panel (no canvas layout reflow)
7. Handle local organism selection/deselection with Pixi pointer events
8. Send speed-control POST requests to `/speed?rate={n}`
9. Send lifecycle control POST requests to `/control?action=start|stop|restart`

## Concurrency and Safety

- World state is guarded by `sync.RWMutex`.
- Simulation loop updates world state under write lock.
- Broadcast loop reads state under read lock.
- Shutdown uses signal context and coordinated goroutine stop.

## Known Limitations

- Protocol is full-state broadcast, not delta-based.
- Brain system is currently a random-walk stub.
- No persistence layer is active yet.

## Runtime Behaviors

- Simulation auto-pauses when population reaches zero.
- Search movement blends wander target + nearest-food cue + smell gradient cue.
- Close organism proximity adds avoidance steering and temporary slowdown to reduce clustering.

## Next Structural Step

Phase 1 backend refactor is complete with package split:

- `backend/cmd/primordia`: entrypoint and process wiring
- `backend/internal/world`: world state and tick orchestration
- `backend/internal/systems`: movement/eating/sense/brain/food systems
- `backend/internal/spatial`: uniform grid
- `backend/internal/server`: websocket hub and handler
- `backend/internal/protocol`: frontend wire format
