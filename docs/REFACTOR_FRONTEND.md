# Primordia — Frontend Visualizer Update

## Context

You are working on **Primordia**, an evolution simulation visualizer.

- **Frontend**: React + PixiJS, fullscreen canvas, WebSocket consumer
- **Backend**: Go, broadcasts `WorldMsg` JSON over WebSocket each tick
- **Current state**: Organisms are drawn as circles. Food is not drawn. No organism inspector exists.

### WorldMsg shape (after backend refactor)

```ts
interface OrganismMsg {
  id: number
  x: number
  y: number
  a: number        // angle in radians
  e: number        // energy 0–1 normalized
  sv?: number[]    // sense vector, 21 floats — present when organism is selected
  sel?: boolean    // true when this organism is selected
}

interface FoodMsg {
  x: number
  y: number
}

interface WorldMsg {
  tick: number
  organisms: OrganismMsg[]
  foods: FoodMsg[]
}
```

---

## Aesthetic Direction

Primordia is a **primordial, bioluminescent deep-ocean world** — dark and atmospheric,
with organisms that glow faintly against an almost-black background.

- **Background**: near-black (`#050a0e`), subtle animated noise grain overlay
- **Organisms**: soft teal/cyan glow (`#00e5cc`), brightness scales with energy level
- **Food**: dim amber/gold particles (`#f5a623`), small pulsing dots
- **Sense rays**: very faint cyan lines, opacity proportional to hit strength
- **Selected organism**: brighter glow ring, distinct accent color (`#ff6b6b`)
- **UI overlay**: dark frosted glass panels, monospaced font, minimal — feels like a scientific instrument
- **Typography**: Use `DM Mono` (Google Fonts) for all UI text. Pair with `fragment` or similar for any display headings.

Do NOT use generic sans-serif fonts, white backgrounds, or flat bright color schemes.
The aesthetic should feel like watching life through a deep-sea submersible porthole.

---

## Task 1 — Draw food

Add a `FoodLayer` to the PixiJS stage.

- Represent each food item as a small circle, radius **3px**, fill `#f5a623`, alpha `0.7`
- Add a gentle CSS/PixiJS pulse animation: scale oscillates between `0.85` and `1.15` over ~2s
- Use a `Map` keyed by `"${x}_${y}"` for upsert/removal
- Food positions are stable between ticks unless eaten — do not recreate graphics every frame, only add/remove deltas

---

## Task 2 — Organism visual overhaul

Replace the current plain circle with a direction-aware organism sprite.

Each organism should render as:
1. **Body**: filled circle, radius proportional to energy (`6px` at full energy, `3px` minimum)
2. **Glow**: outer circle, same center, radius `body + 4px`, fill `#00e5cc`, alpha `0.15`
3. **Direction indicator**: a small triangle pointing in the direction of `Organism.a`, tip at `body + 2px` from center
4. **Energy tint**: interpolate fill color from dim (`#004d44`) at low energy to bright (`#00e5cc`) at full energy using the `e` field

For **selected** organisms (`sel === true`):
- Replace glow with a pulsing ring in `#ff6b6b`, alpha `0.5`
- Body stroke `#ff6b6b`, stroke width `1.5px`

---

## Task 3 — Sense ray visualization

When an organism is selected and `sv` is present, draw its sense rays.

Sense vector layout:
- Indices `0–15`: 8 rays × 2 values each — `[dist_0, type_0, dist_1, type_1, ...]`
  - `dist`: normalized 0–1 (0 = right next to organism, 1 = max range = 150 world units)
  - `type`: 0 = empty, 1 = food, 2 = organism
- Rays are cast in a **240° arc** centered on `Organism.a`
- Ray spacing: `240° / 7` between rays (indices 0–7 left to right across the arc)

Rendering:
- Draw each ray as a line from organism center in the ray's direction
- Length = `dist * 150` world units (converted to screen pixels)
- Color by type: empty = `rgba(0,229,204,0.15)`, food = `rgba(245,166,35,0.5)`, organism = `rgba(255,107,107,0.5)`
- Draw a small circle at the hit point if `type > 0`, radius `3px`, same color at `0.8` alpha
- All ray graphics live in a dedicated `SenseLayer` above the organism layer, cleared and redrawn each tick for the selected organism only

---

## Task 4 — Organism inspector panel

Add a React component `` that appears when an organism is selected.

**Layout**: fixed panel, bottom-left corner, `280px` wide, dark frosted glass style.

**Contents**:
```
┌─────────────────────────────┐
│ ORGANISM #4821              │
│ ─────────────────────────── │
│ Energy    ████████░░  78%   │
│ Age       1,204 ticks       │
│ Speed     0.34 u/tick       │
│                             │
│ SENSE INPUT                 │
│ [mini heatmap of SenseVec]  │
│                             │
│ [Deselect]                  │
└─────────────────────────────┘
```

- Energy bar: CSS progress bar, color interpolates dim→bright teal matching organism tint
- Sense heatmap: a row of 21 small colored squares representing each `sv` float (0=dark, 1=bright). Group visually: first 16 (rays) | next 3 (smell) | last 2 (self)
- Clicking anywhere on the canvas that is NOT an organism deselects
- Clicking an organism: send its ID to a React state `selectedId`, the backend does not need to be notified — the frontend filters `sv` from the broadcast

**Selection interaction (PixiJS side)**:
- Make each organism `Graphics` object interactive: `organism.eventMode = 'static'`
- On `pointerdown`, call `onSelect(id)` callback passed from React
- Cursor: `crosshair` on the canvas element

---

## Task 5 — HUD overlay refinements

Update the existing lightweight overlay:

- Font: `DM Mono`, 11px, color `rgba(255,255,255,0.5)`
- Show: `TICK {n}` · `POP {n}` · `FOOD {n}` · connection status dot (green/red)
- Position: top-right, `16px` margin, no background — just text with `text-shadow: 0 1px 4px #000`
- Add a **speed control**: a row of buttons `0.5×  1×  2×  4×` that POST to `http://localhost:8080/speed?rate={n}` (backend endpoint — add a stub if not present)

---

## File structure

Organise new code as follows:

```
frontend/src/
├── pixi/
│   ├── OrganismLayer.ts    # organism graphics, selection, energy tint
│   ├── FoodLayer.ts        # food particle graphics
│   ├── SenseLayer.ts       # ray + hit point rendering for selected organism
│   └── stage.ts            # PixiJS app init, layer composition
├── components/
│   ├── OrganismInspector.tsx
│   └── HUD.tsx
├── hooks/
│   └── useWorldSocket.ts   # WebSocket consumer, returns WorldMsg stream
└── App.tsx                 # top-level layout, state: selectedId
```

---

## Constraints

- PixiJS v7 API — do not use v8 syntax
- No new npm dependencies beyond what is already installed, except `@google/fonts` import via CSS `@import` for DM Mono
- All PixiJS graphics must use object pooling — never create a new `PIXI.Graphics` for an entity that already exists, only update its properties
- `useWorldSocket` must handle reconnect with exponential backoff (max 8s)
- The inspector panel must not cause layout reflow on the canvas — use `position: fixed`
- TypeScript strict mode — no `any` types