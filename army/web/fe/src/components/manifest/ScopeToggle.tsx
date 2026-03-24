import { cn } from '@/lib/utils';

interface ScopeToggleProps {
  scope: 'user' | 'project';
  onChange: (scope: 'user' | 'project') => void;
}

export function ScopeToggle({ scope, onChange }: ScopeToggleProps) {
  return (
    <div className="inline-flex rounded-lg border border-border bg-card overflow-hidden">
      <button
        className={cn(
          'px-3 py-1.5 text-xs font-medium transition-colors',
          scope === 'user'
            ? 'bg-primary/15 text-primary'
            : 'text-muted-foreground hover:text-foreground'
        )}
        onClick={() => onChange('user')}
      >
        User
      </button>
      <button
        className={cn(
          'px-3 py-1.5 text-xs font-medium transition-colors border-l border-border',
          scope === 'project'
            ? 'bg-primary/15 text-primary'
            : 'text-muted-foreground hover:text-foreground'
        )}
        onClick={() => onChange('project')}
      >
        Project
      </button>
    </div>
  );
}
