# Primordia Frontend

The frontend is a lightweight React shell around a PixiJS renderer.

## Responsibilities

- Maintain UI metadata (population count, connection status).
- Connect to backend websocket feed.
- Render organisms, food, and selected-organism sense rays each tick in a fullscreen canvas.
- Provide selection and local inspection without backend selection RPC.

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

- `src/App.tsx`: top-level layout and selection state.
- `src/hooks/useWorldSocket.ts`: websocket stream + reconnect with exponential backoff.
- `src/pixi/stage.ts`: Pixi app init and layer composition.
- `src/pixi/OrganismLayer.ts`: organism pooling, interaction, and energy-aware visuals.
- `src/pixi/FoodLayer.ts`: food pooling and pulse animation.
- `src/pixi/SenseLayer.ts`: selected-organism sense ray rendering.
- `src/components/HUD.tsx`: top-right telemetry + speed controls.
- `src/components/OrganismInspector.tsx`: fixed inspector panel and sense heatmap.
- `src/App.css`: deep-ocean visual styling, grain overlay, HUD, and inspector.
- `src/index.css`: Global reset/fullscreen root sizing.
- `vite.config.ts`: Vite setup.

## Data Contract (Current)

Expected message shape:

```json
{
  "tick": 42,
  "organisms": [
    { "id": 1, "x": 100, "y": 200, "a": 0.31, "e": 0.84, "sv": [0.4, 1], "sel": false }
  ],
  "foods": [
    { "x": 300, "y": 400 }
  ]
}
```

Organism `sv` (sense vector) may be omitted in lightweight payloads.

## Speed Control Endpoint

The HUD speed buttons send POST requests to:

```bash
http://localhost:8080/speed?rate={n}
```

Current backend behavior is a stub endpoint (acknowledges the value, no runtime tick-rate mutation yet).
