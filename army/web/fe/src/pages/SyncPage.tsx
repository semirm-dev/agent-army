import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { SyncPanel } from '@/components/sync/SyncPanel';
import { useSyncStream } from '@/hooks/use-sync-stream';

export function SyncPage() {
  const [destination, setDestination] = useState<string | undefined>(undefined);
  const { events, isRunning, startSync } = useSyncStream();
  const [searchParams, setSearchParams] = useSearchParams();
  const autostart = searchParams.get('autostart') === 'true';

  useEffect(() => {
    if (autostart && !isRunning) {
      startSync(destination);
      setSearchParams({}, { replace: true });
    }
  }, [autostart]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className="space-y-4 max-w-3xl">
      <div>
        <h2 className="text-xl font-semibold">Sync</h2>
        <p className="text-sm text-muted-foreground">
          Synchronize installed state with your manifest
        </p>
      </div>

      <div className="space-y-1">
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
        <p className="text-xs text-muted-foreground">
          Auto detects a project manifest in the current directory tree and syncs to it, otherwise falls back to the global user manifest (~/.army/manifest.json). User always syncs to the global manifest. Project always syncs to the current directory's manifest.
        </p>
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
