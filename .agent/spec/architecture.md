# Primordia Architecture: ECS + WebSocket Reactive Loop

## 1. Backend: The Go Simulation Engine
We use an **Entity Component System (ECS)** pattern to ensure high performance and decoupled logic.

### Core ECS Concepts
- **Entities:** Simple IDs (e.g., `uint32`).
- **Components:** Pure data structures (e.g., `Position{X, Y}`, `Energy{Value}`, `NeuralNet{Weights}`).
- **Systems:** Functions that iterate over entities with specific components (e.g., `MovementSystem`, `MetabolismSystem`).

### Concurrency Model
- **The Tick:** The simulation runs in a fixed-frequency loop (e.g., 20Hz).
- **Parallelism:** Systems that don't depend on each other (e.g., `Vision` vs `Hunger`) run in separate Goroutines.

---

## 2. Communication: WebSockets (WS)
- **Protocol:** JSON for now (switch to Protobuf if bandwidth becomes a bottleneck).
- **Flow:** 1. Server calculates Tick.
    2. Server diffs the state (only send changed positions).
    3. Server broadcasts "Delta Update" to all connected UI clients.

---

## 3. Storage: SQLite (WAL Mode)
- **Strategy:** Every $N$ ticks, the state of the world is "snapshotted" to SQLite.
- **Goal:** Allow "Time Travel" (rewinding the simulation to see where a species branched off).
- **Optimization:** Use `PRAGMA journal_mode=WAL;` and `PRAGMA synchronous=NORMAL;`.

---

## 4. Frontend: The Visualization Layer
- **Tech:** React + PixiJS (for WebGL-accelerated rendering).
- **State:** The frontend is "dumb"—it just interpolates between the data points sent via WebSocket.