# Primordia Frontend

The frontend is a lightweight React shell around a PixiJS renderer.

## Responsibilities

- Maintain UI metadata (population count, connection status).
- Connect to backend websocket feed.
- Render organisms each tick as circles in a fullscreen canvas.

## Run Frontend Only

From repository root:

```bash
cd frontend
npm install
npm run dev
```

Default URL: `http://localhost:5173`

## Environment

Use `VITE_WS_URL` to control websocket endpoint.

Example:

```bash
VITE_WS_URL=ws://127.0.0.1:8080/ws npm run dev
```

## Key Files

- `src/App.tsx`: Pixi setup, websocket lifecycle, render updates.
- `src/App.css`: Overlay and shell visuals.
- `src/index.css`: Global reset/fullscreen root sizing.
- `vite.config.ts`: Vite setup.

## Data Contract (Current)

Expected message shape:

```json
{
  "orgs": {
    "1": { "pos": { "x": 100, "y": 200 }, "energy": 84.2 }
  },
  "food": {
    "101": { "id": 101, "pos": { "x": 300, "y": 400 } }
  }
}
```

Only `orgs` is currently rendered in the UI.
