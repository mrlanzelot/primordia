import * as PIXI from 'pixi.js';
import type { FoodMsg } from '../hooks/useWorldSocket';

const FOOD_COLOR = 0xf5a623;
const FOOD_RADIUS = 3;

interface FoodVisual {
  g: PIXI.Graphics;
  phase: number;
}

// foodKey provides a stable bucket key for food delta-upserts from world snapshots.
function foodKey(x: number, y: number): string {
  return `${x.toFixed(2)}_${y.toFixed(2)}`;
}

export class FoodLayer {
  public readonly container: PIXI.Container;

  private readonly food = new Map<string, FoodVisual>();

  constructor() {
    this.container = new PIXI.Container();
  }

  // update applies add/remove deltas so persistent food graphics are reused across ticks.
  update(items: FoodMsg[]): void {
    const alive = new Set<string>();

    for (const item of items) {
      if (!Number.isFinite(item.x) || !Number.isFinite(item.y)) {
        continue;
      }
      const key = foodKey(item.x, item.y);
      alive.add(key);

      let visual = this.food.get(key);
      if (!visual) {
        const g = new PIXI.Graphics();
        g.beginFill(FOOD_COLOR, 0.7);
        g.drawCircle(0, 0, FOOD_RADIUS);
        g.endFill();
        g.position.set(item.x, item.y);
        this.container.addChild(g);
        visual = { g, phase: Math.random() * Math.PI * 2 };
        this.food.set(key, visual);
      }
      visual.g.position.set(item.x, item.y);
    }

    for (const [key, visual] of this.food) {
      if (alive.has(key)) {
        continue;
      }
      this.container.removeChild(visual.g);
      visual.g.destroy();
      this.food.delete(key);
    }
  }

  // animate applies the gentle per-food pulse used by the deep-ocean visual style.
  animate(seconds: number): void {
    for (const visual of this.food.values()) {
      const t = seconds * ((2 * Math.PI) / 2) + visual.phase;
      const scale = 1 + 0.15 * Math.sin(t);
      visual.g.scale.set(scale);
    }
  }

  // count returns currently rendered food entities for HUD display.
  count(): number {
    return this.food.size;
  }

  // destroy releases all pooled graphics and their container.
  destroy(): void {
    for (const visual of this.food.values()) {
      this.container.removeChild(visual.g);
      visual.g.destroy();
    }
    this.food.clear();
    this.container.destroy({ children: true });
  }
}
