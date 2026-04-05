import * as PIXI from 'pixi.js';
import { FoodLayer } from './FoodLayer';
import { OrganismLayer } from './OrganismLayer';
import { SenseLayer } from './SenseLayer';

const BG_COLOR = 0x050a0e;
const WORLD_WIDTH = 1000;
const WORLD_HEIGHT = 1000;

export interface StageBundle {
  app: PIXI.Application;
  world: PIXI.Container;
  foodLayer: FoodLayer;
  organismLayer: OrganismLayer;
  senseLayer: SenseLayer;
  resizeToViewport: () => void;
  destroy: () => void;
}

// createStage wires Pixi app, world camera container, and render layers into one bundle.
export function createStage(host: HTMLElement, onSelect: (id: number) => void): StageBundle {
  const app = new PIXI.Application({
    width: window.innerWidth,
    height: window.innerHeight,
    backgroundColor: BG_COLOR,
    antialias: true,
  });

  host.appendChild(app.view as HTMLCanvasElement);

  const world = new PIXI.Container();
  app.stage.addChild(world);

  const foodLayer = new FoodLayer();
  const organismLayer = new OrganismLayer(onSelect);
  const senseLayer = new SenseLayer();

  world.addChild(foodLayer.container);
  world.addChild(organismLayer.container);
  world.addChild(senseLayer.container);

  const resizeToViewport = () => {
    app.renderer.resize(window.innerWidth, window.innerHeight);
    const scale = Math.min(window.innerWidth / WORLD_WIDTH, window.innerHeight / WORLD_HEIGHT);
    world.scale.set(scale, scale);
    world.x = (window.innerWidth - WORLD_WIDTH * scale) / 2;
    world.y = (window.innerHeight - WORLD_HEIGHT * scale) / 2;
  };

  resizeToViewport();

  const destroy = () => {
    foodLayer.destroy();
    organismLayer.destroy();
    senseLayer.destroy();
    app.destroy(true, true);
  };

  return { app, world, foodLayer, organismLayer, senseLayer, resizeToViewport, destroy };
}
