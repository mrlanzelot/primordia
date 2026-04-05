import * as PIXI from 'pixi.js';
import type { FederatedPointerEvent } from 'pixi.js';
import type { OrganismMsg } from '../hooks/useWorldSocket';

const BASE_ORG_COLOR = 0x004d44;
const BRIGHT_ORG_COLOR = 0x00e5cc;
const SELECTED_COLOR = 0xff6b6b;
const MIN_BODY_RADIUS = 3;
const MAX_BODY_RADIUS = 6;

interface OrganismVisual {
  root: PIXI.Container;
  glow: PIXI.Graphics;
  body: PIXI.Graphics;
  direction: PIXI.Graphics;
  ring: PIXI.Graphics;
}

// clamp01 bounds normalized render inputs.
function clamp01(v: number): number {
  return Math.max(0, Math.min(1, v));
}

// lerp computes a linear interpolation between two scalar values.
function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}

// mixColor blends two RGB colors by normalized factor t.
function mixColor(a: number, b: number, t: number): number {
  const ra = (a >> 16) & 0xff;
  const ga = (a >> 8) & 0xff;
  const ba = a & 0xff;

  const rb = (b >> 16) & 0xff;
  const gb = (b >> 8) & 0xff;
  const bb = b & 0xff;

  const r = Math.round(lerp(ra, rb, t));
  const g = Math.round(lerp(ga, gb, t));
  const bl = Math.round(lerp(ba, bb, t));
  return (r << 16) | (g << 8) | bl;
}

// drawSprite refreshes one organism's visual parts from energy and selection state.
function drawSprite(visual: OrganismVisual, energy: number, selected: boolean): void {
  const e = clamp01(energy);
  const bodyR = lerp(MIN_BODY_RADIUS, MAX_BODY_RADIUS, e);
  const bodyColor = mixColor(BASE_ORG_COLOR, BRIGHT_ORG_COLOR, e);

  visual.glow.clear();
  visual.body.clear();
  visual.direction.clear();
  visual.ring.clear();

  if (!selected) {
    visual.glow.beginFill(BRIGHT_ORG_COLOR, 0.15);
    visual.glow.drawCircle(0, 0, bodyR + 4);
    visual.glow.endFill();
  }

  if (selected) {
    visual.ring.lineStyle(1.5, SELECTED_COLOR, 0.5);
    visual.ring.drawCircle(0, 0, bodyR + 4.5);
    visual.body.lineStyle(1.5, SELECTED_COLOR, 1);
  }

  visual.body.beginFill(bodyColor, 1);
  visual.body.drawCircle(0, 0, bodyR);
  visual.body.endFill();

  const tip = bodyR + 2;
  const side = Math.max(1.8, bodyR * 0.6);
  const tail = Math.max(1.5, bodyR * 0.45);
  visual.direction.beginFill(BRIGHT_ORG_COLOR, 0.95);
  visual.direction.moveTo(tip, 0);
  visual.direction.lineTo(tail, side);
  visual.direction.lineTo(tail, -side);
  visual.direction.closePath();
  visual.direction.endFill();
}

export class OrganismLayer {
  public readonly container: PIXI.Container;

  private readonly visuals = new Map<number, OrganismVisual>();
  private readonly onSelect: (id: number) => void;

  constructor(onSelect: (id: number) => void) {
    this.onSelect = onSelect;
    this.container = new PIXI.Container();
  }

  // update keeps organism graphics in sync with latest snapshot while reusing pooled objects.
  update(organisms: OrganismMsg[], selectedId: number | null): void {
    const alive = new Set<number>();

    for (const org of organisms) {
      if (
        !Number.isFinite(org.id) ||
        !Number.isFinite(org.x) ||
        !Number.isFinite(org.y) ||
        !Number.isFinite(org.a) ||
        !Number.isFinite(org.e)
      ) {
        continue;
      }

      alive.add(org.id);
      let visual = this.visuals.get(org.id);
      if (!visual) {
        const root = new PIXI.Container();
        root.eventMode = 'static';
        root.cursor = 'pointer';

        const glow = new PIXI.Graphics();
        const body = new PIXI.Graphics();
        const direction = new PIXI.Graphics();
        const ring = new PIXI.Graphics();

        root.addChild(glow, body, direction, ring);
        root.on('pointerdown', (event: FederatedPointerEvent) => {
          event.stopPropagation();
          this.onSelect(org.id);
        });

        this.container.addChild(root);
        visual = { root, glow, body, direction, ring };
        this.visuals.set(org.id, visual);
      }

      const selected = selectedId === org.id;
      drawSprite(visual, org.e, selected);
      visual.root.position.set(org.x, org.y);
      visual.root.rotation = org.a;
    }

    for (const [id, visual] of this.visuals) {
      if (alive.has(id)) {
        continue;
      }
      this.container.removeChild(visual.root);
      visual.root.destroy({ children: true });
      this.visuals.delete(id);
    }
  }

  // animate applies selection pulse effects without rebuilding geometry each frame.
  animate(seconds: number, selectedId: number | null): void {
    if (selectedId === null) {
      return;
    }
    const visual = this.visuals.get(selectedId);
    if (!visual) {
      return;
    }
    const pulse = 1 + 0.06 * Math.sin(seconds * 2 * Math.PI * 0.7 + selectedId * 0.11);
    visual.ring.scale.set(pulse);
    visual.ring.alpha = 0.6 + 0.2 * Math.sin(seconds * 2 * Math.PI * 0.7 + selectedId * 0.11);
  }

  // destroy disposes all organism visuals and clears pooled state.
  destroy(): void {
    for (const visual of this.visuals.values()) {
      this.container.removeChild(visual.root);
      visual.root.destroy({ children: true });
    }
    this.visuals.clear();
    this.container.destroy({ children: true });
  }
}
