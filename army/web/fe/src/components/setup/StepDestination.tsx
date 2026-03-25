import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface StepDestinationProps {
  value: 'user' | 'project' | null;
  onChange: (dest: 'user' | 'project') => void;
  onNext: () => void;
}

export function StepDestination({ value, onChange, onNext }: StepDestinationProps) {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Where should your manifest be saved?</h3>
      <p className="text-sm text-muted-foreground mb-6">
        This determines whether selections apply globally or to this project only.
      </p>

      <div className="flex flex-col gap-3 max-w-md">
        <button
          onClick={() => onChange('user')}
          className={cn(
            'text-left rounded-lg border-2 p-4 transition-colors',
            value === 'user'
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/30'
          )}
        >
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'size-4 rounded-full border-2 flex items-center justify-center',
                value === 'user' ? 'border-primary' : 'border-muted-foreground/40'
              )}
            >
              {value === 'user' && <span className="size-2 rounded-full bg-primary" />}
            </span>
            <span className="font-medium text-sm">User-level (global defaults)</span>
          </div>
          <p className="text-xs text-muted-foreground mt-1 ml-6">
            ~/.army/manifest.json — applies to all projects
          </p>
        </button>

        <button
          onClick={() => onChange('project')}
          className={cn(
            'text-left rounded-lg border-2 p-4 transition-colors',
            value === 'project'
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/30'
          )}
        >
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'size-4 rounded-full border-2 flex items-center justify-center',
                value === 'project' ? 'border-primary' : 'border-muted-foreground/40'
              )}
            >
              {value === 'project' && <span className="size-2 rounded-full bg-primary" />}
            </span>
            <span className="font-medium text-sm">Project-level (current project)</span>
          </div>
          <p className="text-xs text-muted-foreground mt-1 ml-6">
            &lt;cwd&gt;/.army/manifest.json — scoped to this project
          </p>
        </button>
      </div>

      <div className="flex justify-end mt-8">
        <Button onClick={onNext} disabled={!value}>
          Next
        </Button>
      </div>
    </div>
  );
}
