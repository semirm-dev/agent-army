import { Plus, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface CatalogCardProps {
  name: string;
  description: string;
  source: string;
  tags: string[];
  inManifest: boolean;
  onAdd: () => void;
  isAdding: boolean;
}

export function CatalogCard({
  name,
  description,
  source,
  tags,
  inManifest,
  onAdd,
  isAdding,
}: CatalogCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-4 hover:border-primary/30 transition-colors">
      <div className="flex items-center gap-2 mb-2">
        <span className="size-1.5 rounded-full bg-primary shrink-0" />
        <span className="font-mono text-sm font-semibold truncate">{name}</span>
      </div>
      <p className="text-xs text-muted-foreground line-clamp-2 mb-3 leading-relaxed">
        {description}
      </p>
      <div className="flex flex-wrap gap-1.5 mb-3">
        {tags.map((tag) => (
          <span
            key={tag}
            className="px-2 py-0.5 rounded text-[10px] bg-muted border border-border text-muted-foreground"
          >
            {tag}
          </span>
        ))}
      </div>
      <div className="flex items-center justify-between">
        <span className="font-mono text-[10px] text-muted-foreground/50 truncate">{source}</span>
        {inManifest ? (
          <span className="text-[11px] text-muted-foreground/40 flex items-center gap-1">
            <Check className="size-3" /> Added
          </span>
        ) : (
          <Button
            variant="outline"
            size="sm"
            className="h-7 text-xs border-primary/50 text-primary hover:bg-primary/10 hover:text-primary"
            onClick={onAdd}
            disabled={isAdding}
          >
            <Plus className="size-3" />
            {isAdding ? 'Adding...' : 'Add'}
          </Button>
        )}
      </div>
    </div>
  );
}
