import { Button } from '@/components/ui/button';
import { Loader2 } from 'lucide-react';

interface StepConfirmProps {
  destination: 'user' | 'project';
  selectedPlugins: string[];
  selectedSkills: string[];
  onSave: () => void;
  onBack: () => void;
  isSaving: boolean;
  error: string | null;
}

export function StepConfirm({
  destination,
  selectedPlugins,
  selectedSkills,
  onSave,
  onBack,
  isSaving,
  error,
}: StepConfirmProps) {
  const destLabel = destination === 'user'
    ? '~/.army/manifest.json'
    : '<cwd>/.army/manifest.json';

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Review Your Selections</h3>
      <p className="text-sm text-muted-foreground mb-6">Confirm before saving.</p>

      <div className="flex flex-col gap-4 max-w-md">
        <div className="rounded-lg border border-border bg-card p-4">
          <p className="text-[11px] text-muted-foreground uppercase tracking-wider mb-1.5">Destination</p>
          <p className="text-sm">
            {destination === 'user' ? 'User-level' : 'Project-level'}{' '}
            <span className="font-mono text-xs text-muted-foreground">{destLabel}</span>
          </p>
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <p className="text-[11px] text-muted-foreground uppercase tracking-wider mb-1.5">
            Plugins ({selectedPlugins.length})
          </p>
          <div className="flex flex-wrap gap-1.5">
            {selectedPlugins.length === 0 ? (
              <span className="text-xs text-muted-foreground">None selected</span>
            ) : (
              selectedPlugins.map((name) => (
                <span key={name} className="px-2 py-0.5 rounded text-xs bg-primary/10 text-primary">{name}</span>
              ))
            )}
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <p className="text-[11px] text-muted-foreground uppercase tracking-wider mb-1.5">
            Skills ({selectedSkills.length})
          </p>
          <div className="flex flex-wrap gap-1.5">
            {selectedSkills.length === 0 ? (
              <span className="text-xs text-muted-foreground">None selected</span>
            ) : (
              selectedSkills.map((name) => (
                <span key={name} className="px-2 py-0.5 rounded text-xs bg-primary/10 text-primary">{name}</span>
              ))
            )}
          </div>
        </div>
      </div>

      {error && <p className="text-xs text-red-500 mt-3">{error}</p>}

      <div className="flex justify-between mt-8">
        <Button variant="outline" onClick={onBack} disabled={isSaving}>Back</Button>
        <Button onClick={onSave} disabled={isSaving}>
          {isSaving && <Loader2 className="size-3.5 animate-spin" />}
          {isSaving ? 'Saving...' : 'Save Manifest'}
        </Button>
      </div>
    </div>
  );
}
