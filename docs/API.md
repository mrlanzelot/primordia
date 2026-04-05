# Primordia API

## Transport

- Protocol: WebSocket
- Endpoint: `/ws`
- Server default: `ws://127.0.0.1:8080/ws`
- Direction: server -> client stream

## Message Format (Current)

The backend broadcasts a world snapshot as JSON.

```json
{
  "tick": 210,
  "organisms": [
    {
      "id": 1,
      "x": 120.2,
      "y": 88.6,
      "a": 0.79,
      "e": 96.3,
      "sv": [0.12, 1, 0.34, 0, 1, 0]
    }
  ],
  "foods": [
    { "x": 301.1, "y": 552.4 }
  ]
}
```

## Entity Fields

### Organism

- `id`: numeric identifier
- `x`: horizontal coordinate (0..WorldWidth)
- `y`: vertical coordinate (0..WorldHeight)
- `a`: heading in radians
- `e`: current energy level
- `sv`: optional 21-value sense vector

### Food

- `x`: horizontal coordinate
- `y`: vertical coordinate

### World

- `tick`: simulation tick counter
- `organisms`: current organism snapshot list
- `foods`: current food snapshot list

## Update Rates

- Simulation update: every 30ms (`TickRate`)
- WebSocket broadcast: every 40ms (`BroadcastRate`)

## Compatibility Notes

- Frontend renders both `organisms` and `foods`.
- `sv` is included for organism sensing/inspection and can be omitted for lightweight clients.

## Planned Changes

- Integrate brain outputs into movement system (Phase 2)
- Add selection state and client-side inspection overlays
- Add optional delta updates for lower bandwidth
