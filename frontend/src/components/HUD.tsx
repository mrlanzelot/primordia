interface HUDProps {
  tick: number;
  population: number;
  food: number;
  activeRate: number;
  simState: 'running' | 'stopped';
  connectionState: 'connecting' | 'connected' | 'closed' | 'error';
  onRateChange: (rate: number) => void;
  onControl: (action: 'start' | 'stop' | 'restart') => void;
}

const RATES = [0.5, 1, 2, 4] as const;

// HUD shows compact telemetry plus simulation speed and lifecycle controls.
export function HUD({
  tick,
  population,
  food,
  activeRate,
  simState,
  connectionState,
  onRateChange,
  onControl,
}: HUDProps) {
  const online = connectionState === 'connected';

  return (
    <div className="hud">
      <p>
        TICK {tick} · POP {population} · FOOD {food}{' '}
        <span className={`status-dot ${online ? 'online' : 'offline'}`} aria-label={connectionState} />
      </p>
      <div className="speed-controls">
        {RATES.map((rate) => (
          <button
            key={rate}
            type="button"
            className={activeRate === rate ? 'active' : undefined}
            onClick={() => onRateChange(rate)}
          >
            {rate}x
          </button>
        ))}
      </div>
      <div className="control-controls">
        <button
          type="button"
          className={simState === 'running' ? 'active' : undefined}
          onClick={() => onControl('start')}
        >
          Start
        </button>
        <button
          type="button"
          className={simState === 'stopped' ? 'active stop-active' : undefined}
          onClick={() => onControl('stop')}
        >
          Stop
        </button>
        <button type="button" onClick={() => onControl('restart')}>
          Restart
        </button>
      </div>
    </div>
  );
}
