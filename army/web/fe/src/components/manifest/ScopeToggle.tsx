import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

interface ScopeToggleProps {
  scope: 'user' | 'project';
  onChange: (scope: 'user' | 'project') => void;
}

export function ScopeToggle({ scope, onChange }: ScopeToggleProps) {
  return (
    <div className="inline-flex rounded-lg border p-0.5">
      <Button
        variant="ghost"
        size="sm"
        className={cn(
          'rounded-md px-3',
          scope === 'user' && 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground'
        )}
        onClick={() => onChange('user')}
      >
        User
      </Button>
      <Button
        variant="ghost"
        size="sm"
        className={cn(
          'rounded-md px-3',
          scope === 'project' && 'bg-primary text-primary-foreground hover:bg-primary/90 hover:text-primary-foreground'
        )}
        onClick={() => onChange('project')}
      >
        Project
      </Button>
    </div>
  );
}
