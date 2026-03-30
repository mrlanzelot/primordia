# Primordia Roadmap

## Phase 1: Emergence Explorer

- Stabilize simulation loop and websocket streaming.
- Improve frontend clarity and rendering ergonomics.
- Produce baseline docs for setup, architecture, and files.

## Phase 2: Reproduction + Selection

- Add reproduction rules and mutation parameters.
- Introduce simple species trait tracking.
- Add controls for simulation speed and reset.

## Phase 3: Explainability

- Add event and lineage logs.
- Provide visual debugging overlays for behavior.
- Add replay-friendly persistence snapshots.

## Phase 4: Scale and Persistence

- Transition to delta/protobuf transport if needed.
- Add SQLite snapshot strategy.
- Split backend into packages by simulation domain.
