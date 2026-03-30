# Primordia API

## Transport

- Protocol: WebSocket
- Endpoint: `/ws`
- Server default: `ws://127.0.0.1:8080/ws`
- Direction: server -> client stream

## Message Format (Current)

The backend broadcasts the world as JSON.

```json
{
  "orgs": {
    "1": {
      "id": 1,
      "pos": { "x": 120.2, "y": 88.6 },
      "energy": 96.3
    }
  },
  "food": {
    "50": {
      "id": 50,
      "pos": { "x": 301.1, "y": 552.4 }
    }
  }
}
```

## Entity Fields

### Organism

- `id`: numeric identifier
- `pos.x`: horizontal coordinate (0..WorldWidth)
- `pos.y`: vertical coordinate (0..WorldHeight)
- `energy`: current energy level

### Food

- `id`: numeric identifier
- `pos.x`: horizontal coordinate
- `pos.y`: vertical coordinate

## Update Rates

- Simulation update: every 30ms (`TickRate`)
- WebSocket broadcast: every 40ms (`BroadcastRate`)

## Compatibility Notes

- Frontend currently renders only `orgs`.
- `food` is transmitted but not yet visualized.

## Planned Changes

- Add message envelope (`type`, `tick`, `payload`)
- Add optional delta updates
- Add client command messages (pause/reset/spawn)
