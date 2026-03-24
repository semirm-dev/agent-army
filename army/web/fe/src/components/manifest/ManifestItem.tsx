import { Trash2 } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
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
  ok: { color: 'bg-green-500', label: 'Installed' },
  missing: { color: 'bg-red-500', label: 'Missing' },
  drift: { color: 'bg-yellow-500', label: 'Drift' },
} as const;

export function ManifestItem({
  name,
  source,
  destination,
  status,
  onRemove,
  isRemoving,
}: ManifestItemProps) {
  const statusInfo = statusConfig[status];

  return (
    <div className="flex items-center justify-between py-3 px-4 rounded-lg border">
      <div className="flex items-center gap-3 min-w-0">
        <div
          className={cn('size-2.5 rounded-full shrink-0', statusInfo.color)}
          title={statusInfo.label}
        />
        <div className="min-w-0">
          <p className="font-medium text-sm">{name}</p>
          <p className="text-xs text-muted-foreground truncate" title={source}>
            {source}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <Badge variant="outline" className="text-[10px]">
          {destination}
        </Badge>
        <Badge
          variant={status === 'ok' ? 'secondary' : status === 'missing' ? 'destructive' : 'outline'}
          className="text-[10px]"
        >
          {statusInfo.label}
        </Badge>
        <Button
          variant="ghost"
          size="icon"
          className="size-8 text-muted-foreground hover:text-destructive"
          onClick={onRemove}
          disabled={isRemoving}
        >
          <Trash2 className="size-4" />
        </Button>
      </div>
    </div>
  );
}
