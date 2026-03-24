import { useState, useCallback, useRef } from 'react';
import type { SyncEvent } from '../lib/types';

const API_BASE = import.meta.env.VITE_API_URL || '/api';

export function useSyncStream() {
  const [events, setEvents] = useState<SyncEvent[]>([]);
  const [isRunning, setIsRunning] = useState(false);
  const abortRef = useRef<AbortController | null>(null);

  const startSync = useCallback(async (destination?: string) => {
    // Abort any previous sync
    abortRef.current?.abort();

    const controller = new AbortController();
    abortRef.current = controller;

    setEvents([]);
    setIsRunning(true);

    try {
      const res = await fetch(`${API_BASE}/sync`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ destination }),
        signal: controller.signal,
      });

      if (!res.ok || !res.body) {
        throw new Error(`Sync failed: ${res.status}`);
      }

      const reader = res.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          const trimmed = line.trim();
          if (trimmed.startsWith('data: ')) {
            try {
              const event = JSON.parse(trimmed.slice(6)) as SyncEvent;
              setEvents((prev) => [...prev, event]);
            } catch {
              // skip malformed lines
            }
          }
        }
      }
    } catch (err) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setEvents((prev) => [
          ...prev,
          { event: 'error', message: err.message } as SyncEvent,
        ]);
      }
    } finally {
      setIsRunning(false);
      abortRef.current = null;
    }
  }, []);

  const cancelSync = useCallback(() => {
    abortRef.current?.abort();
  }, []);

  return { events, isRunning, startSync, cancelSync };
}
