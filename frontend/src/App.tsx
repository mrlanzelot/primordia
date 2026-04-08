import { useEffect, useMemo, useRef, useState } from 'react';
import { HUD } from './components/HUD';
import { OrganismInspector } from './components/OrganismInspector';
import type { OrganismMsg } from './hooks/useWorldSocket';
import { useWorldSocket } from './hooks/useWorldSocket';
import { createStage, type StageBundle } from './pixi/stage';
import './App.css';

const DEFAULT_WS_URL = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/ws`;

// postSpeedRate sends the selected simulation speed multiplier to the backend runtime control endpoint.
async function postSpeedRate(rate: number): Promise<void> {
  await fetch(`/api/speed?rate=${rate}`, { method: 'POST' });
}

// postControlAction sends start/stop/restart commands to the backend runtime controls endpoint.
async function postControlAction(action: 'start' | 'stop' | 'restart'): Promise<void> {
  await fetch(`/api/control?action=${action}`, { method: 'POST' });
}

export default function App() {
  const [selectedId, setSelectedId] = useState<number | null>(null);
  const [activeRate, setActiveRate] = useState(1);
  const [simState, setSimState] = useState<'running' | 'stopped'>('running');
  const hostRef = useRef<HTMLDivElement>(null);
  const stageRef = useRef<StageBundle | null>(null);
  const selectedIdRef = useRef<number | null>(null);

  const wsURL = import.meta.env.VITE_WS_URL ?? DEFAULT_WS_URL;
  const { world, connectionState, lastPacketAt } = useWorldSocket(wsURL);

  const selected = useMemo<OrganismMsg | null>(() => {
    if (!world || selectedId === null) {
      return null;
    }
    return world.organisms.find((org) => org.id === selectedId) ?? null;
  }, [world, selectedId]);

  // Mirror selected id into a ref so ticker callbacks can read current value without re-subscribing.
  useEffect(() => {
    selectedIdRef.current = selectedId;
  }, [selectedId]);

  // Initialize Pixi stage once and wire global pointer interaction for deselection.
  useEffect(() => {
    if (!hostRef.current) {
      return;
    }

    const stage = createStage(hostRef.current, (id) => setSelectedId(id));
    stageRef.current = stage;
    (stage.app.view as HTMLCanvasElement).style.cursor = 'crosshair';

    stage.app.stage.eventMode = 'static';
    stage.app.stage.hitArea = stage.app.screen;
    stage.app.stage.on('pointerdown', () => {
      setSelectedId(null);
    });

    const onResize = () => stage.resizeToViewport();
    window.addEventListener('resize', onResize);

    stage.app.ticker.add(() => {
      const seconds = stage.app.ticker.lastTime / 1000;
      stage.foodLayer.animate(seconds);
      stage.organismLayer.animate(seconds, selectedIdRef.current);
    });

    return () => {
      window.removeEventListener('resize', onResize);
      stage.destroy();
      stageRef.current = null;
    };
  }, []);

  // Apply incoming world snapshots onto pooled Pixi layers.
  useEffect(() => {
    if (!stageRef.current || !world) {
      return;
    }

    if (selectedId !== null) {
      const stillPresent = world.organisms.some((o) => o.id === selectedId);
      if (!stillPresent) {
        setSelectedId(null);
      }
    }

    stageRef.current.foodLayer.update(world.foods);
    stageRef.current.organismLayer.update(world.organisms, selectedId);

    const selectedOrganism = selectedId === null
      ? null
      : world.organisms.find((org) => org.id === selectedId) ?? null;

    stageRef.current.senseLayer.draw(selectedOrganism);
  }, [world, selectedId]);

  const staleMs = lastPacketAt ? Date.now() - lastPacketAt : 0;
  const tick = world?.tick ?? 0;
  const population = world?.organisms.length ?? 0;
  const foodCount = world?.foods.length ?? 0;

  useEffect(() => {
    if (population === 0) {
      setSimState('stopped');
    }
  }, [population]);

  return (
    <div className="app-shell">
      <div className="grain" />
      <HUD
        tick={tick}
        population={population}
        food={foodCount}
        activeRate={activeRate}
        simState={simState}
        connectionState={connectionState}
        onRateChange={(rate) => {
          void postSpeedRate(rate).then(() => {
            setActiveRate(rate);
          });
        }}
        onControl={(action) => {
          void postControlAction(action).then(() => {
            if (action === 'start') {
              setSimState('running');
            } else if (action === 'stop') {
              setSimState('stopped');
            } else {
              setSimState('running');
            }
          });
        }}
      />

      <div className="meta-strip">Last packet {staleMs ? `${Math.round(staleMs / 1000)}s ago` : 'none'}</div>

      <OrganismInspector
        organism={selected}
        selectedId={selectedId}
        onDeselect={() => setSelectedId(null)}
      />

      <div ref={hostRef} className="canvas-host" />
    </div>
  );
}
