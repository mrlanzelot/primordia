import { useEffect, useRef, useState } from 'react';
import * as PIXI from 'pixi.js';
import './App.css';

const BG_COLOR = 0x050505;
const HEALTHY_COLOR = 0x4ade80;
const LOW_ENERGY_COLOR = 0xf87171;
const FOOD_COLOR = 0xfacc15;
const SEARCH_COLOR = 0x60a5fa;
const EXPLOIT_COLOR = 0xa78bfa;
const REORIENT_COLOR = 0xf97316;
const ENERGY_THRESHOLD = 50;
const ORGANISM_RADIUS = 12;
const FOOD_RADIUS = 5;
const WORLD_WIDTH = 1000;
const WORLD_HEIGHT = 1000;

type ConnectionState = 'connecting' | 'connected' | 'closed' | 'error';
type CellState = 'search' | 'exploit_circle' | 'reorient' | string;
interface Organism { pos: { x: number; y: number }; energy: number; state?: CellState }
interface Food { pos: { x: number; y: number } }
interface WorldMessage { orgs?: Record<string, Organism>; food?: Record<string, Food> }

const stateColor = (state?: CellState, energy?: number) => {
  if (state === 'exploit_circle') return EXPLOIT_COLOR;
  if (state === 'reorient') return REORIENT_COLOR;
  if (state === 'search') return SEARCH_COLOR;
  return (energy ?? 0) > ENERGY_THRESHOLD ? HEALTHY_COLOR : LOW_ENERGY_COLOR;
};

export default function App() {
  const [pop, setPop] = useState(0);
  const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');
  const [lastSeen, setLastSeen] = useState(0);
  const [debug, setDebug] = useState('waiting');
  const pixiContainer = useRef<HTMLDivElement>(null);
  const appRef = useRef<PIXI.Application | null>(null);
  const orgsRef = useRef<Map<number, PIXI.Graphics>>(new Map());
  const foodRef = useRef<Map<number, PIXI.Graphics>>(new Map());

  useEffect(() => {
    const app = new PIXI.Application({ width: window.innerWidth, height: window.innerHeight, backgroundColor: BG_COLOR, antialias: true });
    if (pixiContainer.current) pixiContainer.current.appendChild(app.view as HTMLCanvasElement);
    appRef.current = app;

    const center = new PIXI.Graphics();
    center.lineStyle(1, 0x222222, 0.9);
    center.moveTo(window.innerWidth / 2, 0);
    center.lineTo(window.innerWidth / 2, window.innerHeight);
    center.moveTo(0, window.innerHeight / 2);
    center.lineTo(window.innerWidth, window.innerHeight / 2);
    app.stage.addChild(center);

    const marker = new PIXI.Graphics();
    marker.beginFill(0xff00ff);
    marker.drawRect(window.innerWidth / 2 - 10, window.innerHeight / 2 - 10, 20, 20);
    marker.endFill();
    app.stage.addChild(marker);

    const ws = new WebSocket(import.meta.env.VITE_WS_URL ?? 'ws://127.0.0.1:8080/ws');
    const handleResize = () => appRef.current?.renderer.resize(window.innerWidth, window.innerHeight);
    window.addEventListener('resize', handleResize);
    ws.onopen = () => setConnectionState('connected');
    ws.onerror = () => setConnectionState('error');
    ws.onclose = () => setConnectionState('closed');
    ws.onmessage = (e) => {
      const data = JSON.parse(e.data) as WorldMessage;
      const orgs = data.orgs || {};
      const food = data.food || {};
      setLastSeen(Date.now());
      setPop(Object.keys(orgs).length);
      const firstOrg = Object.values(orgs)[0];
      if (firstOrg) setDebug(`first org: ${Math.round(firstOrg.pos.x)},${Math.round(firstOrg.pos.y)} scale:${app.stage.scale.x.toFixed(2)} size:${window.innerWidth}x${window.innerHeight}`);
      Object.entries(food).forEach(([idStr, item]) => {
        const id = Number(idStr);
        let graphic = foodRef.current.get(id);
        if (!graphic) { graphic = new PIXI.Graphics(); app.stage.addChild(graphic); foodRef.current.set(id, graphic); }
        graphic.clear(); graphic.beginFill(FOOD_COLOR); graphic.drawCircle(item.pos.x, item.pos.y, FOOD_RADIUS); graphic.endFill();
      });
      Object.entries(orgs).forEach(([idStr, org]) => {
        const id = Number(idStr);
        let graphic = orgsRef.current.get(id);
        if (!graphic) { graphic = new PIXI.Graphics(); app.stage.addChild(graphic); orgsRef.current.set(id, graphic); }
        graphic.clear(); graphic.beginFill(stateColor(org.state, org.energy)); graphic.drawCircle(org.pos.x, org.pos.y, ORGANISM_RADIUS); graphic.endFill();
      });
      orgsRef.current.forEach((val, key) => { if (!orgs[String(key)]) { app.stage.removeChild(val); orgsRef.current.delete(key); } });
      foodRef.current.forEach((val, key) => { if (!food[String(key)]) { app.stage.removeChild(val); foodRef.current.delete(key); } });
    };
    return () => { window.removeEventListener('resize', handleResize); ws.close(); orgsRef.current.clear(); foodRef.current.clear(); app.destroy(true, true); };
  }, []);

  const staleMs = lastSeen ? Date.now() - lastSeen : 0;
  return <div className="app-shell"><div className="overlay"><h1>PRIMORDIA ENGINE</h1><p>Population: {pop}</p><p>Connection: {connectionState}</p><p>Last packet: {staleMs ? `${Math.round(staleMs / 1000)}s ago` : 'none'}</p><p>{debug}</p></div><div ref={pixiContainer} /></div>;
}
