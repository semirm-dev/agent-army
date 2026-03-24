import { useState } from 'react';
import { SyncPanel } from '@/components/sync/SyncPanel';
import { useSyncStream } from '@/hooks/use-sync-stream';

export function SyncPage() {
  const [destination, setDestination] = useState<string | undefined>(undefined);
  const { events, isRunning, startSync } = useSyncStream();

  return (
    <div className="space-y-4 max-w-3xl">
      <div>
        <h2 className="text-xl font-semibold">Sync</h2>
        <p className="text-sm text-muted-foreground">
          Synchronize installed state with your manifest
        </p>
      </div>

      <div className="flex items-center gap-2">
        <label className="text-sm text-muted-foreground">Destination:</label>
        <select
          className="bg-card border border-border rounded-md px-2.5 py-1.5 text-xs text-muted-foreground"
          value={destination || ''}
          onChange={(e) => setDestination(e.target.value || undefined)}
        >
          <option value="">Auto</option>
          <option value="user">User</option>
          <option value="project">Project</option>
        </select>
      </div>

      <SyncPanel
        events={events}
        isRunning={isRunning}
        onSync={startSync}
        destination={destination}
      />
    </div>
  );
}
