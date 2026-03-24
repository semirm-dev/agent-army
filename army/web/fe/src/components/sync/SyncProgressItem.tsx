import { cn } from '@/lib/utils';

interface SyncProgressItemProps {
  type: string;
  itemType: string;
  name: string;
  status: 'pending' | 'running' | 'success' | 'failed';
  error?: string;
}

export function SyncProgressItem({
  type,
  itemType,
  name,
  status,
  error,
}: SyncProgressItemProps) {
  return (
    <div className={cn('flex items-center gap-2 py-0.5 font-mono text-xs', status === 'pending' && 'opacity-30')}>
      <span className="w-4 text-center">
        {status === 'success' && <span className="text-green-400">&#10003;</span>}
        {status === 'failed' && <span className="text-red-400">&#10007;</span>}
        {status === 'running' && <span className="text-yellow-400 animate-pulse">&#9611;</span>}
        {status === 'pending' && <span className="text-muted-foreground">&#9675;</span>}
      </span>
      <span className="text-muted-foreground">{type}</span>
      <span className="text-muted-foreground/60">{itemType}</span>
      <span className="text-foreground">{name}</span>
      {error && <span className="text-red-400 ml-2">&mdash; {error}</span>}
    </div>
  );
}
