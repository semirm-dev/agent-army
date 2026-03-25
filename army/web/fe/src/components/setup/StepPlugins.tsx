import { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { SelectableList, type SelectableItem } from './SelectableList';
import type { Catalog } from '@/lib/types';

interface StepPluginsProps {
  catalog: Catalog;
  selectedTech: string[];
  selectedPlugins: string[];
  onChange: (plugins: string[]) => void;
  onNext: () => void;
  onBack: () => void;
}

export function StepPlugins({ catalog, selectedTech, selectedPlugins, onChange, onNext, onBack }: StepPluginsProps) {
  const recommended = useMemo(() => {
    const names = new Set<string>();
    for (const tech of selectedTech) {
      const profile = catalog.tech_profiles[tech];
      if (profile) {
        for (const p of profile.plugins) names.add(p);
      }
    }
    return names;
  }, [catalog, selectedTech]);

  const items: SelectableItem[] = useMemo(
    () =>
      catalog.plugins.map((p) => ({
        name: p.name,
        description: p.description,
        source: p.marketplace,
        recommended: recommended.has(p.name),
      })),
    [catalog.plugins, recommended]
  );

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Plugins</h3>
      <p className="text-sm text-muted-foreground mb-4">
        Select plugins to include. Stars indicate recommendations for your tech stack.
      </p>
      <SelectableList items={items} selected={selectedPlugins} onChange={onChange} />
      <div className="flex justify-between mt-6">
        <Button variant="outline" onClick={onBack}>Back</Button>
        <Button onClick={onNext}>Next</Button>
      </div>
    </div>
  );
}
