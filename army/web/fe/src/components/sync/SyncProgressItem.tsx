import { Loader2, CheckCircle2, XCircle } from 'lucide-react';
import { Badge } from '@/components/ui/badge';

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
    <div className="flex items-center gap-3 py-2 px-3 rounded-md border">
      <div className="shrink-0">
        {status === 'running' && (
          <Loader2 className="size-4 animate-spin text-blue-500" />
        )}
        {status === 'success' && (
          <CheckCircle2 className="size-4 text-green-500" />
        )}
        {status === 'failed' && (
          <XCircle className="size-4 text-red-500" />
        )}
        {status === 'pending' && (
          <div className="size-4 rounded-full border-2 border-muted-foreground/30" />
        )}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{name}</span>
          <Badge variant="outline" className="text-[10px]">
            {type}
          </Badge>
          <Badge variant="secondary" className="text-[10px]">
            {itemType}
          </Badge>
        </div>
        {error && (
          <p className="text-xs text-red-500 mt-1">{error}</p>
        )}
      </div>
    </div>
  );
}
