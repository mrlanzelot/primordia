import { useEffect, useRef, useState } from 'react';

export interface OrganismMsg {
  id: number;
  x: number;
  y: number;
  a: number;
  e: number;
  sv?: number[];
  sel?: boolean;
  age?: number;
}

export interface FoodMsg {
  x: number;
  y: number;
}

export interface WorldMsg {
  tick: number;
  organisms: OrganismMsg[];
  foods: FoodMsg[];
}

interface UseWorldSocketResult {
  world: WorldMsg | null;
  connectionState: 'connecting' | 'connected' | 'closed' | 'error';
  lastPacketAt: number;
}

const MAX_BACKOFF_MS = 8000;

// useWorldSocket maintains a reconnecting websocket stream and exposes latest world state.
export function useWorldSocket(url: string): UseWorldSocketResult {
  const [world, setWorld] = useState<WorldMsg | null>(null);
  const [connectionState, setConnectionState] = useState<'connecting' | 'connected' | 'closed' | 'error'>('connecting');
  const [lastPacketAt, setLastPacketAt] = useState(0);
  const closedByUnmount = useRef(false);

  useEffect(() => {
    closedByUnmount.current = false;
    let ws: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    let attempt = 0;

    // clearReconnect cancels any scheduled reconnect before creating a fresh one.
    const clearReconnect = () => {
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
    };

    // connect opens a websocket and wires retry logic with exponential backoff.
    const connect = () => {
      clearReconnect();
      setConnectionState('connecting');
      ws = new WebSocket(url);

      ws.onopen = () => {
        attempt = 0;
        setConnectionState('connected');
      };

      ws.onmessage = (event) => {
        try {
          const parsed = JSON.parse(event.data) as WorldMsg;
          setWorld(parsed);
          setLastPacketAt(Date.now());
        } catch {
          setConnectionState('error');
        }
      };

      ws.onerror = () => {
        setConnectionState('error');
      };

      ws.onclose = () => {
        if (closedByUnmount.current) {
          setConnectionState('closed');
          return;
        }
        setConnectionState('closed');
        const backoff = Math.min(MAX_BACKOFF_MS, 500 * Math.pow(2, attempt));
        attempt += 1;
        reconnectTimer = window.setTimeout(connect, backoff);
      };
    };

    connect();

    return () => {
      closedByUnmount.current = true;
      clearReconnect();
      ws?.close();
    };
  }, [url]);

  return { world, connectionState, lastPacketAt };
}
