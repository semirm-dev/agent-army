import { useMemo } from 'react';
import { Play, CheckCircle2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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
        <Button onClick={() => onSync(destination)} disabled={isRunning} size="lg">
          <Play className="size-4" />
          {isRunning ? 'Syncing...' : 'Sync Now'}
        </Button>
      </div>

      {actionStatuses.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Progress</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
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
          </CardContent>
        </Card>
      )}

      {completeEvent && completeEvent.event === 'complete' && (
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <CheckCircle2 className="size-5 text-green-500" />
            <span className="text-sm">
              Sync complete: {completeEvent.succeeded} succeeded
              {completeEvent.failed > 0 && `, ${completeEvent.failed} failed`}
            </span>
          </CardContent>
        </Card>
      )}

      {errorEvent && errorEvent.event === 'error' && (
        <Card className="border-red-200">
          <CardContent className="pt-6">
            <p className="text-sm text-red-500">{errorEvent.message}</p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
