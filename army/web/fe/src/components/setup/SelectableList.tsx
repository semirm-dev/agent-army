import { useState, useMemo } from 'react';
import { Search } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface SelectableItem {
  name: string;
  description: string;
  source: string;
  recommended: boolean;
}

interface SelectableListProps {
  items: SelectableItem[];
  selected: string[];
  onChange: (selected: string[]) => void;
}

export function SelectableList({ items, selected, onChange }: SelectableListProps) {
  const [search, setSearch] = useState('');

  const filtered = useMemo(() => {
    if (!search) return items;
    const lower = search.toLowerCase();
    return items.filter(
      (item) =>
        item.name.toLowerCase().includes(lower) ||
        item.description.toLowerCase().includes(lower)
    );
  }, [items, search]);

  const toggle = (name: string) => {
    onChange(
      selected.includes(name)
        ? selected.filter((n) => n !== name)
        : [...selected, name]
    );
  };

  const selectAll = () => onChange(items.map((i) => i.name));
  const clearAll = () => onChange([]);

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-8 pl-8 pr-3 text-xs rounded-md border border-border bg-card focus:outline-none focus:ring-2 focus:ring-ring/50"
          />
        </div>
        <button onClick={selectAll} className="text-[11px] text-muted-foreground hover:text-foreground px-2 py-1 border border-border rounded-md">
          Select All
        </button>
        <button onClick={clearAll} className="text-[11px] text-muted-foreground hover:text-foreground px-2 py-1 border border-border rounded-md">
          Clear
        </button>
      </div>
      <div className="flex flex-col gap-1.5 max-h-[360px] overflow-y-auto">
        {filtered.map((item) => {
          const isSelected = selected.includes(item.name);
          return (
            <button
              key={item.name}
              onClick={() => toggle(item.name)}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-lg border text-left transition-colors',
                isSelected
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:border-primary/30'
              )}
            >
              <span className={cn(
                'size-4 rounded border flex items-center justify-center shrink-0 text-[10px]',
                isSelected
                  ? 'bg-primary border-primary text-primary-foreground'
                  : 'border-muted-foreground/40'
              )}>
                {isSelected && '\u2713'}
              </span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-1.5">
                  <span className="font-mono text-sm font-semibold truncate">{item.name}</span>
                  {item.recommended && (
                    <span className="text-primary text-xs">\u2605</span>
                  )}
                </div>
                <p className="text-xs text-muted-foreground truncate">{item.description}</p>
              </div>
              <span className="font-mono text-[10px] text-muted-foreground/50 shrink-0">{item.source}</span>
            </button>
          );
        })}
      </div>
      <p className="text-xs text-muted-foreground">{selected.length} selected</p>
    </div>
  );
}
