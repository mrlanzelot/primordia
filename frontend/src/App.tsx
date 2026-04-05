import { useEffect, useRef, useState } from 'react';
import * as PIXI from 'pixi.js';
import './App.css';

const BG_COLOR = 0x050505;
const HIGH_ENERGY_COLOR = 0x3b82f6;
const MID_ENERGY_COLOR = 0xfacc15;
const LOW_ENERGY_COLOR = 0xef4444;
const FOOD_COLOR = 0xfacc15;
const INITIAL_ENERGY = 80;
const ORGANISM_RADIUS = 12;
const FOOD_RADIUS = 5;
const WORLD_WIDTH = 1000;
const WORLD_HEIGHT = 1000;

type ConnectionState = 'connecting' | 'connected' | 'closed' | 'error';
interface OrganismMsg { id?: number; x?: number; y?: number; a?: number; e?: number; sv?: number[] }
interface FoodMsg { x?: number; y?: number }
interface WorldMessage { tick?: number; organisms?: OrganismMsg[]; foods?: FoodMsg[] }

const stateColor = (energy?: number) => {
  const pct = ((energy ?? 0) / INITIAL_ENERGY) * 100;
  if (pct > 67) return HIGH_ENERGY_COLOR;
  if (pct < 33) return LOW_ENERGY_COLOR;
  return MID_ENERGY_COLOR;
};

export default function App() {
  const [pop, setPop] = useState(0);
  const [connectionState, setConnectionState] = useState<ConnectionState>('connecting');
  const [lastSeen, setLastSeen] = useState(0);
  const [debug, setDebug] = useState('waiting');
  const pixiContainer = useRef<HTMLDivElement>(null);
  const appRef = useRef<PIXI.Application | null>(null);
  const worldRef = useRef<PIXI.Container | null>(null);
  const orgsRef = useRef<Map<number, PIXI.Graphics>>(new Map());
  const foodLayerRef = useRef<PIXI.Container | null>(null);

  useEffect(() => {
    const app = new PIXI.Application({ width: window.innerWidth, height: window.innerHeight, backgroundColor: BG_COLOR, antialias: true });
    if (pixiContainer.current) pixiContainer.current.appendChild(app.view as HTMLCanvasElement);
    appRef.current = app;

    const world = new PIXI.Container();
    worldRef.current = world;
    app.stage.addChild(world);

    const foodLayer = new PIXI.Container();
    foodLayerRef.current = foodLayer;
    world.addChild(foodLayer);

    const drawCamera = () => {
      const scale = Math.min(window.innerWidth / WORLD_WIDTH, window.innerHeight / WORLD_HEIGHT);
      world.scale.set(scale, scale);
      world.x = (window.innerWidth - WORLD_WIDTH * scale) / 2;
      world.y = (window.innerHeight - WORLD_HEIGHT * scale) / 2;
    };
    drawCamera();

    const center = new PIXI.Graphics();
    center.lineStyle(1, 0x222222, 0.9);
    center.moveTo(WORLD_WIDTH / 2, 0);
    center.lineTo(WORLD_WIDTH / 2, WORLD_HEIGHT);
    center.moveTo(0, WORLD_HEIGHT / 2);
    center.lineTo(WORLD_WIDTH, WORLD_HEIGHT / 2);
    world.addChild(center);

    const marker = new PIXI.Graphics();
    marker.beginFill(0xff00ff);
    marker.drawRect(WORLD_WIDTH / 2 - 10, WORLD_HEIGHT / 2 - 10, 20, 20);
    marker.endFill();
    world.addChild(marker);

    const ws = new WebSocket(import.meta.env.VITE_WS_URL ?? 'ws://127.0.0.1:8080/ws');
    const handleResize = () => {
      appRef.current?.renderer.resize(window.innerWidth, window.innerHeight);
      drawCamera();
    };
    window.addEventListener('resize', handleResize);
    ws.onopen = () => setConnectionState('connected');
    ws.onerror = () => setConnectionState('error');
    ws.onclose = () => setConnectionState('closed');
    ws.onmessage = (e) => {
      const data = JSON.parse(e.data) as WorldMessage;
      const orgs = data.organisms || [];
      const food = data.foods || [];
      setLastSeen(Date.now());
      setPop(orgs.length);
      const firstOrg = orgs[0];
      const fx = firstOrg?.x;
      const fy = firstOrg?.y;
      setDebug(firstOrg && typeof fx === 'number' && typeof fy === 'number'
        ? `first org: ${Math.round(fx)},${Math.round(fy)} scale:${world.scale.x.toFixed(2)} size:${window.innerWidth}x${window.innerHeight}`
        : `first org: missing scale:${world.scale.x.toFixed(2)} size:${window.innerWidth}x${window.innerHeight}`);
      if (foodLayerRef.current) {
        foodLayerRef.current.removeChildren().forEach((child) => child.destroy());
        food.forEach((item) => {
          const x = item.x;
          const y = item.y;
          if (typeof x !== 'number' || typeof y !== 'number') return;
          const graphic = new PIXI.Graphics();
          graphic.beginFill(FOOD_COLOR);
          graphic.drawCircle(x, y, FOOD_RADIUS);
          graphic.endFill();
          foodLayerRef.current?.addChild(graphic);
        });
      }
      orgs.forEach((org) => {
        const id = org.id;
        const x = org.x;
        const y = org.y;
        if (typeof x !== 'number' || typeof y !== 'number') return;
        if (typeof id !== 'number') return;
        let graphic = orgsRef.current.get(id);
        if (!graphic) { graphic = new PIXI.Graphics(); world.addChild(graphic); orgsRef.current.set(id, graphic); }
        graphic.clear(); graphic.beginFill(stateColor(org.e)); graphic.drawCircle(x, y, ORGANISM_RADIUS); graphic.endFill();
      });
      const orgIDs = new Set(orgs.map((org) => org.id).filter((id): id is number => typeof id === 'number'));
      orgsRef.current.forEach((val, key) => {
        if (!orgIDs.has(key)) {
          world.removeChild(val);
          orgsRef.current.delete(key);
        }
      });
    };
    return () => {
      window.removeEventListener('resize', handleResize);
      ws.close();
      orgsRef.current.clear();
      foodLayerRef.current?.removeChildren().forEach((child) => child.destroy());
      app.destroy(true, true);
    };
  }, []);

  const staleMs = lastSeen ? Date.now() - lastSeen : 0;
  return <div className="app-shell"><div className="overlay"><h1>PRIMORDIA ENGINE</h1><p>Population: {pop}</p><p>Connection: {connectionState}</p><p>Last packet: {staleMs ? `${Math.round(staleMs / 1000)}s ago` : 'none'}</p><p>{debug}</p></div><div ref={pixiContainer} /></div>;
}
