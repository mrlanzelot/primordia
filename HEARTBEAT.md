# Primordia Heartbeat

**Current Phase:** Phase 1 - Emergence Explorer
**Active Sprint:** 0.1.0 - The Spark of Life (Basic Movement & Energy)
**Status:** 🟢 Stabilizing & Documenting

---

## 🎯 Current Objectives
1. [x] **Backend:** Implement ECS World and basic 'Organism' Entity.
2. [x] **Simulation:** Simple 'Foraging' System (Energy goes down over time; up when touching food).
3. [x] **Network:** Setup Gorilla WebSocket hub to broadcast Entity positions.
4. [x] **Frontend:** Basic Canvas/PixiJS layer to render Entities as dots.
5. [x] **Docs:** Baseline setup, architecture, API, and file-role documentation.

## 🚧 Blockers / Notes
- **Decision Needed:** Should energy loss be constant or based on movement speed?
- **Technical Debt:** Protocol is still full-state websocket broadcast (no delta compression).

## 📈 Latest Evolutionary Milestone
- *None yet. Waiting for first successful 'World' generation.*

## 🛠 Active Agents
- **Architect:** Planning ECS Component signatures.
- **Scientist:** Defining "Metabolic Cost" formulas.