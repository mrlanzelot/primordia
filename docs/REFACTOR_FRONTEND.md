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
┌