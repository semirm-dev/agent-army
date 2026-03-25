import { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { SelectableList, type SelectableItem } from './SelectableList';
import type { Catalog } from '@/lib/types';

interface StepSkillsProps {
  catalog: Catalog;
  selectedTech: string[];
  selectedSkills: string[];
  onChange: (skills: string[]) => void;
  onNext: () => void;
  onBack: () => void;
}

export function StepSkills({ catalog, selectedTech, selectedSkills, onChange, onNext, onBack }: StepSkillsProps) {
  const recommended = useMemo(() => {
    const names = new Set<string>();
    for (const tech of selectedTech) {
      const profile = catalog.tech_profiles[tech];
      if (profile) {
        for (const s of profile.skills) names.add(s);
      }
    }
    return names;
  }, [catalog, selectedTech]);

  const items: SelectableItem[] = useMemo(
    () =>
      catalog.skills.map((s) => ({
        name: s.name,
        description: s.description,
        source: s.source,
        recommended: recommended.has(s.name),
      })),
    [catalog.skills, recommended]
  );

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Skills</h3>
      <p className="text-sm text-muted-foreground mb-4">
        Select skills to include. Stars indicate recommendations for your tech stack.
      </p>
      <SelectableList items={items} selected={selectedSkills} onChange={onChange} />
      <div className="flex justify-between mt-6">
        <Button variant="outline" onClick={onBack}>Back</Button>
        <Button onClick={onNext}>Next</Button>
      </div>
    </div>
  );
}
