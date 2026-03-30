import { useEffect, useRef, useState } from 'react';
import * as PIXI from 'pixi.js';
import './App.css';

const BG_COLOR = 0x050505;
const HEALTHY_COLOR = 0x4ade80;
const LOW_ENERGY_COLOR = 0xf87171;
const ENERGY_THRESHOLD = 50;
const ORGANISM_RADIUS = 6;

type ConnectionState = 'connecting' | 'connected' | 'closed' | 'error';

interface Organism {
  pos: {
    x: number;
    y: number;
  };
  energy: number;
}

interface WorldMessage {
  orgs?: Record<string, Organism>;
}

export default function App() {
  const [pop, setPop] = useState(0);
  const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');
  const pixiContainer = useRef<HTMLDivElement>(null);
  const appRef = useRef<PIXI.Application | null>(null);
  const orgsRef = useRef<Map<number, PIXI.Graphics>>(new Map());

  useEffect(() => {
    const app = new PIXI.Application({
      width: window.innerWidth,
      height: window.innerHeight,
      backgroundColor: BG_COLOR,
      antialias: true,
    });

    if (pixiContainer.current) {
      pixiContainer.current.appendChild(app.view as HTMLCanvasElement);
    }

    appRef.current = app;
    const wsUrl = import.meta.env.VITE_WS_URL ?? 'ws://127.0.0.1:8080/ws';
    const ws = new WebSocket(wsUrl);

    const handleResize = () => {
      const current = appRef.current;
      if (!current) {
        return;
      }
      current.renderer.resize(window.innerWidth, window.innerHeight);
    };
    window.addEventListener('resize', handleResize);

    ws.onopen = () => {
      setConnectionState('connected');
    };

    ws.onerror = () => {
      setConnectionState('error');
    };

    ws.onclose = () => {
      setConnectionState('closed');
    };

    ws.onmessage = (e) => {
      const data = JSON.parse(e.data) as WorldMessage;
      const orgs = data.orgs || {};
      setPop(Object.keys(orgs).length);

      Object.entries(orgs).forEach(([idStr, org]) => {
        const id = Number(idStr);
        let graphic = orgsRef.current.get(id);

        if (!graphic) {
          graphic = new PIXI.Graphics();
          app.stage.addChild(graphic);
          orgsRef.current.set(id, graphic);
        }

        graphic.clear();
        graphic.beginFill(org.energy > ENERGY_THRESHOLD ? HEALTHY_COLOR : LOW_ENERGY_COLOR);
        graphic.drawCircle(org.pos.x, org.pos.y, ORGANISM_RADIUS);
        graphic.endFill();
      });

      orgsRef.current.forEach((val, key) => {
        if (!orgs[String(key)]) {
          app.stage.removeChild(val);
          orgsRef.current.delete(key);
        }
      });
    };

    return () => {
      window.removeEventListener('resize', handleResize);
      ws.close();
      orgsRef.current.clear();
      app.destroy(true, true);
    };
  }, []);

  return (
    <div className="app-shell">
      <div className="overlay">
        <h1>PRIMORDIA ENGINE</h1>
        <p>Population: {pop}</p>
        <p>Connection: {connectionState}</p>
      </div>
      <div ref={pixiContainer} />
    </div>
  );
}