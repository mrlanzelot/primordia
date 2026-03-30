# Primordia Architecture

## Runtime Overview

Primordia currently runs as two processes:

1. Go backend simulation engine
2. React + PixiJS frontend visualizer

The backend is the source of truth. The frontend consumes websocket state and draws it.

## Backend Flow

Main runtime flow:

1. Create world maps (`Organisms`, `Food`)
2. Seed initial population
3. Start simulation tick loop (`TickRate`)
4. Accept websocket clients on `/ws`
5. Broadcast serialized world at `BroadcastRate`

Simulation update is split into three systems:

1. `updateOrganisms`: movement, boundary clamp, energy decay, death
2. `spawnFood`: food cap enforcement and random spawn
3. `applyEating`: proximity checks and energy gain

## Frontend Flow

1. Initialize Pixi Application fullscreen
2. Open websocket connection (`VITE_WS_URL`)
3. Parse incoming world messages
4. Upsert organism graphics by id
5. Remove graphics for missing organisms
6. Show lightweight overlay metadata (population, connection)

## Concurrency and Safety

- World state is guarded by `sync.RWMutex`.
- Simulation loop updates world state under write lock.
- Broadcast loop reads state under read lock.
- Shutdown uses signal context and coordinated goroutine stop.

## Known Limitations

- Backend is still in one file (`backend/main.go`).
- Protocol is full-state broadcast, not delta-based.
- Rendering does not yet draw food or trajectories.
- No persistence layer is active yet.

## Next Structural Step

Split backend into packages:

- `backend/sim`: world + systems
- `backend/net`: websocket server/protocol
- `backend/cmd/engine`: entrypoint
