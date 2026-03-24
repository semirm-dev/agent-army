import { X } from 'lucide-react';
import { cn } from '@/lib/utils';

interface ManifestItemProps {
  name: string;
  source: string;
  destination: string;
  status: 'ok' | 'missing' | 'drift';
  installed: boolean;
  onRemove: () => void;
  isRemoving: boolean;
}

const statusConfig = {
  ok: { color: 'bg-green-500', label: '' },
  missing: { color: 'bg-red-500', label: 'missing' },
  drift: { color: 'bg-yellow-500', label: 'drift' },
} as const;

export function ManifestItem({
  name,
  source,
  destination,
  status,
  onRemove,
  isRemoving,
}: ManifestItemProps) {
  const cfg = statusConfig[status];

  return (
    <div className="flex items-center justify-between py-2 px-3 rounded-md border border-border bg-card hover:border-primary/20 transition-colors">
      <div className="flex items-center gap-2.5 min-w-0">
        <span className={cn('size-1.5 rounded-full shrink-0', cfg.color)} />
        <span className="font-mono text-[13px] font-medium truncate">{name}</span>
        <span className="font-mono text-[10px] text-muted-foreground/40 truncate hidden sm:inline">{source}</span>
        {cfg.label && (
          <span className={cn(
            'text-[10px] font-mono',
            status === 'missing' ? 'text-red-400' : 'text-yellow-400'
          )}>
            {cfg.label}
          </span>
        )}
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <span className="px-2 py-0.5 rounded text-[10px] bg-muted border border-border text-muted-foreground">
          {destination}
        </span>
        <button
          className="size-6 rounded flex items-center justify-center text-muted-foreground/40 hover:text-red-400 hover:bg-red-400/10 transition-colors disabled:opacity-50"
          onClick={onRemove}
          disabled={isRemoving}
        >
          <X className="size-3" />
        </button>
      </div>
    </div>
  );
}
