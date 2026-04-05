import { useEffect, useState } from 'react';
import type { OrganismMsg } from '../hooks/useWorldSocket';

interface OrganismInspectorProps {
  organism: OrganismMsg | null;
  selectedId: number | null;
  onDeselect: () => void;
}

// clamp01 keeps inspector-derived normalized values in display-safe bounds.
function clamp01(v: number): number {
  return Math.max(0, Math.min(1, v));
}

// energyBarColor maps normalized energy to the same dim-to-bright teal ramp as organisms.
function energyBarColor(energy: number): string {
  const e = clamp01(energy);
  const low = [0, 77, 68];
  const high = [0, 229, 204];
  const rgb = low.map((v, idx) => Math.round(v + (high[idx] - v) * e));
  return `rgb(${rgb[0]}, ${rgb[1]}, ${rgb[2]})`;
}

function wrapAngle(angle: number): number {
  let out = angle;
  while (out > Math.PI) {
    out -= Math.PI * 2;
  }
  while (out < -Math.PI) {
    out += Math.PI * 2;
  }
  return out;
}

function speedNeedleRotation(speed: number): number {
  const normalized = clamp01(speed);
  return -110 + normalized * 220;
}

function ageCycle(age: number): number {
  return ((age % 600) + 600) % 600;
}

function ageMaturity(age: number): number {
  return clamp01(age / 6000);
}

function describeRelativeDirection(angle: number): string {
  if (angle < -2.45) {
    return 'behind';
  }
  if (angle < -1.35) {
    return 'back-left';
  }
  if (angle < -0.45) {
    return 'left';
  }
  if (angle < -0.1) {
    return 'ahead-left';
  }
  if (angle <= 0.1) {
    return 'straight ahead';
  }
  if (angle <= 0.45) {
    return 'ahead-right';
  }
  if (angle <= 1.35) {
    return 'right';
  }
  if (angle <= 2.45) {
    return 'back-right';
  }
  return 'behind';
}

function classifyRayType(type: number): { label: string; className: string } {
  if (type === 1) {
    return { label: 'Food', className: 'vision-food' };
  }
  if (type === 2) {
    return { label: 'Organism', className: 'vision-organism' };
  }
  return { label: 'Clear', className: 'vision-empty' };
}

// OrganismInspector renders selected-organism diagnostics and sense heatmap details.
export function OrganismInspector({ organism, selectedId, onDeselect }: OrganismInspectorProps) {
  const speedRaw = clamp01(organism?.sv?.[20] ?? 0);
  const [displaySpeed, setDisplaySpeed] = useState(speedRaw);

  useEffect(() => {
    setDisplaySpeed(speedRaw);
  }, [selectedId]);

  useEffect(() => {
    setDisplaySpeed((prev) => prev * 0.82 + speedRaw * 0.18);
  }, [speedRaw]);

  if (!organism || selectedId === null) {
    return null;
  }

  const energy = clamp01(organism.e);
  const speed = clamp01(displaySpeed);
  const smell = clamp01(organism.sv?.[16] ?? 0);
  const age = organism.age ?? 0;
  const sv = organism.sv ?? [];
  const smellX = sv[17] ?? 0;
  const smellY = sv[18] ?? 0;
  const smellAngle = wrapAngle(Math.atan2(smellY, smellX) - organism.a);
  const smellDirection = smell > 0.12 ? describeRelativeDirection(smellAngle) : 'none';
  const smellRotationDeg = smell > 0.12 ? (smellAngle * 180) / Math.PI : 0;
  const speedRotation = speedNeedleRotation(speed);
  const smellVisualOpacity = 0.18 + smell * 0.82;
  const ageCyclePercent = Math.round((ageCycle(age) / 600) * 100);
  const maturity = ageMaturity(age);

  const rays = Array.from({ length: 8 }).map((_, idx) => {
    const dist = clamp01(sv[idx * 2] ?? 1);
    const type = Math.round(sv[idx * 2 + 1] ?? 0);
    const localAngle = -((4 * Math.PI) / 3) / 2 + (((4 * Math.PI) / 3) / 7) * idx;
    const direction = describeRelativeDirection(localAngle);
    return { idx, dist, type, direction, strength: Math.max(0.06, 1 - dist), ...classifyRayType(type) };
  });

  return (
    <aside className="inspector" aria-live="polite">
      <h2>ORGANISM #{selectedId}</h2>
      <div className="energy-row">
        <span>Energy</span>
        <div className="energy-track" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(energy * 100)}>
          <div className="energy-fill" style={{ width: `${energy * 100}%`, backgroundColor: energyBarColor(energy) }} />
        </div>
        <span>{Math.round(energy * 100)}%</span>
      </div>

      <div className="status-strip" aria-label="Organism status overview">
        <div className="status-card age-card">
          <span className="status-label">Age Pulse</span>
          <div className="age-meter" role="meter" aria-valuemin={0} aria-valuemax={100} aria-valuenow={ageCyclePercent}>
            <div className="age-meter-fill" style={{ width: `${ageCyclePercent}%` }} />
          </div>
          <span className="status-value">Cycle {ageCyclePercent}%</span>
          <div className="age-maturity-track" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(maturity * 100)}>
            <div className="age-maturity-fill" style={{ width: `${Math.round(maturity * 100)}%` }} />
          </div>
          <span className="status-value">Maturity {Math.round(maturity * 100)}%</span>
        </div>
        <div className="status-card speedometer-card">
          <span className="status-label">Speed</span>
          <div className="speedometer" role="meter" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(speed * 100)}>
            <div className="speedometer-arc" />
            <div className="speedometer-needle" style={{ transform: `translateX(-50%) rotate(${speedRotation}deg)` }} />
          </div>
        </div>
      </div>

      <p className="sense-title">SENSE INPUT</p>
      <p className="sense-subtitle">These rows show what the organism currently detects about the world and itself.</p>
      <div className="sense-legend" aria-label="Sense ray legend">
        <span className="sense-legend-item">
          <span className="sense-legend-dot sense-empty" />
          EMPTY
        </span>
        <span className="sense-legend-item">
          <span className="sense-legend-dot sense-food" />
          FOOD
        </span>
        <span className="sense-legend-item">
          <span className="sense-legend-dot sense-organism" />
          ORGANISM
        </span>
      </div>

      <div className="sense-section">
        <div className="sense-row-header">
          <span className="sense-row-title">Ray Sensors</span>
          <span className="sense-row-desc">Eight vision rays across the forward arc, from left to right.</span>
        </div>
        <div className="vision-axis">
          <span>LEFT</span>
          <span>CENTER</span>
          <span>RIGHT</span>
        </div>
        <div className="vision-strip">
          {rays.map((ray) => (
            <div key={ray.idx} className={`vision-card ${ray.className}`} aria-label={`${ray.label} ${ray.direction}`}>
              <div className="vision-distance-bar" style={{ height: `${Math.round(ray.strength * 100)}%` }} />
            </div>
          ))}
        </div>
      </div>

      <div className="sense-section">
        <div className="sense-row-header">
          <span className="sense-row-title">Smell</span>
          <span className="sense-row-desc">Direction and strength of the local food scent gradient.</span>
        </div>
        <div className="smell-panel">
          <div className="smell-compass" aria-label="Smell direction compass">
            <div className="smell-compass-ring" />
            <div className="smell-compass-glow" style={{ opacity: smellVisualOpacity }} />
            <div className="smell-compass-center" />
            <div className="smell-arrow" style={{ transform: `translate(-50%, -100%) rotate(${smellRotationDeg}deg)`, opacity: smellVisualOpacity }} />
          </div>
          <div className="smell-info" aria-label={`Food scent ${smellDirection}, strength ${Math.round(smell * 100)} percent`}>
            <div className="smell-track" role="progressbar" aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(smell * 100)}>
              <div className="smell-fill" style={{ width: `${smell * 100}%` }} />
            </div>
            <div className="smell-strength-dots" aria-hidden="true">
              {Array.from({ length: 10 }).map((_, idx) => (
                <span key={idx} className={idx < Math.round(smell * 10) ? 'dot active' : 'dot'} />
              ))}
            </div>
          </div>
        </div>
      </div>

      <button type="button" className="deselect-btn" onClick={onDeselect}>
        Deselect
      </button>
    </aside>
  );
}
