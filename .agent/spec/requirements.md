# Primordia Requirements

## Phase 1: Emergence Explorer

### Objective

Provide a functioning baseline simulation where organisms move, lose energy, consume food, and can be observed in real time.

### Functional Requirements

1. Backend maintains a world with organisms and food entities.
2. Simulation updates at a fixed tick cadence.
3. Organisms move, consume energy, and die at zero energy.
4. Food spawns up to a configured cap.
5. Organisms near food gain energy and remove that food.
6. Backend broadcasts world state via websocket.
7. Frontend renders organism positions with visible state feedback.

### Non-Functional Requirements

1. Code should prioritize readability and explicit behavior.
2. Shared world state must be concurrency-safe.
3. Setup and run workflow must be documented and reproducible.

### Acceptance Criteria

1. `make dev` starts backend and frontend successfully.
2. Frontend shows changing population while simulation runs.
3. WebSocket disconnects are handled without frontend crash.
4. Repository docs explain where core logic lives.
