import { useMemo } from 'react';
import { Play } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { SyncProgressItem } from './SyncProgressItem';
import type { SyncEvent, SyncAction } from '@/lib/types';

interface SyncPanelProps {
  events: SyncEvent[];
  isRunning: boolean;
  onSync: (destination?: string) => void;
  destination?: string;
}

interface ActionStatus {
  action: SyncAction;
  status: 'pending' | 'running' | 'success' | 'failed';
  error?: string;
}

export function SyncPanel({ events, isRunning, onSync, destination }: SyncPanelProps) {
  const actionStatuses = useMemo(() => {
    const statuses = new Map<string, ActionStatus>();

    for (const ev of events) {
      if (ev.event === 'plan') {
        for (const action of ev.actions) {
          statuses.set(action.name, { action, status: 'pending' });
        }
      } else if (ev.event === 'action_start') {
        const existing = statuses.get(ev.name);
        if (existing) {
          existing.status = 'running';
        }
      } else if (ev.event === 'action_done') {
        const existing = statuses.get(ev.name);
        if (existing) {
          existing.status = ev.success ? 'success' : 'failed';
          existing.error = ev.error;
        }
      }
    }

    return Array.from(statuses.values());
  }, [events]);

  const completeEvent = events.find((e) => e.event === 'complete');
  const errorEvent = events.find((e) => e.event === 'error');

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Button
          onClick={() => onSync(destination)}
          disabled={isRunning}
          size="lg"
          className="bg-primary text-primary-foreground hover:bg-primary/90"
        >
          <Play className="size-4" />
          {isRunning ? 'Syncing...' : 'Sync Now'}
        </Button>
      </div>

      {actionStatuses.length > 0 && (
        <div className="bg-card border border-border rounded-lg font-mono">
          <div className="px-4 py-2.5 border-b border-border">
            <span className="text-xs text-muted-foreground/60">$ army sync --json --yes</span>
          </div>
          <div className="p-4 space-y-0.5">
            {actionStatuses.map((as) => (
              <SyncProgressItem
                key={as.action.name}
                type={as.action.type}
                itemType={as.action.item_type}
                name={as.action.name}
                status={as.status}
                error={as.error}
              />
            ))}
          </div>
        </div>
      )}

      {completeEvent && completeEvent.event === 'complete' && (
        <div className="bg-card border border-border rounded-lg px-4 py-3 font-mono text-xs">
          <span className="text-green-400">✓</span>
          <span className="text-muted-foreground ml-2">
            Sync complete: {completeEvent.succeeded} succeeded
            {completeEvent.failed > 0 && `, ${completeEvent.failed} failed`}
          </span>
        </div>
      )}

      {errorEvent && errorEvent.event === 'error' && (
        <div className="bg-card border border-red-500/30 rounded-lg px-4 py-3 font-mono text-xs">
          <span className="text-red-400">✗</span>
          <span className="text-red-400 ml-2">{errorEvent.message}</span>
        </div>
      )}
    </div>
  );
}
