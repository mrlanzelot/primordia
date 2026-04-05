import * as PIXI from 'pixi.js';
import type { OrganismMsg } from '../hooks/useWorldSocket';

const ARC = (4 * Math.PI) / 3;
const RAY_LENGTH = 150;

// clamp01 keeps encoded sensor distances in valid normalized range.
function clamp01(v: number): number {
  return Math.max(0, Math.min(1, v));
}

export class SenseLayer {
  public readonly container: PIXI.Container;

  private readonly pool: PIXI.Graphics[] = [];
  private active = 0;

  constructor() {
    this.container = new PIXI.Container();
  }

  // nextGraphic returns a reusable graphics object from pool or allocates one lazily.
  private nextGraphic(): PIXI.Graphics {
    if (this.active < this.pool.length) {
      const g = this.pool[this.active];
      this.active += 1;
      g.clear();
      g.visible = true;
      return g;
    }
    const g = new PIXI.Graphics();
    this.pool.push(g);
    this.container.addChild(g);
    this.active += 1;
    return g;
  }

  // clear hides previously used pooled graphics so a new draw pass can start cleanly.
  clear(): void {
    for (let i = 0; i < this.active; i += 1) {
      this.pool[i].clear();
      this.pool[i].visible = false;
    }
    this.active = 0;
  }

  // draw renders the selected organism's 8 sense rays and hit markers from sv data.
  draw(selected: OrganismMsg | null): void {
    this.clear();
    if (!selected || !selected.sv || selected.sv.length < 16) {
      return;
    }

    const start = selected.a - ARC / 2;
    const step = ARC / 7;

    for (let i = 0; i < 8; i += 1) {
      const dist = clamp01(selected.sv[i * 2] ?? 1);
      const kind = Math.round(selected.sv[i * 2 + 1] ?? 0);
      const rayAngle = start + i * step;
      const len = dist * RAY_LENGTH;
      const ex = selected.x + Math.cos(rayAngle) * len;
      const ey = selected.y + Math.sin(rayAngle) * len;

      let lineColor = 0x00e5cc;
      let lineAlpha = 0.15;
      if (kind === 1) {
        lineColor = 0xf5a623;
        lineAlpha = 0.5;
      } else if (kind === 2) {
        lineColor = 0xff6b6b;
        lineAlpha = 0.5;
      }

      const line = this.nextGraphic();
      line.lineStyle(1, lineColor, lineAlpha);
      line.moveTo(selected.x, selected.y);
      line.lineTo(ex, ey);

      if (kind > 0) {
        const hit = this.nextGraphic();
        hit.beginFill(lineColor, 0.8);
        hit.drawCircle(ex, ey, 3);
        hit.endFill();
      }
    }
  }

  // destroy releases pooled graphics resources.
  destroy(): void {
    this.clear();
    for (const g of this.pool) {
      g.destroy();
    }
    this.pool.length = 0;
    this.container.destroy({ children: true });
  }
}
