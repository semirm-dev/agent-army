import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import type { Catalog } from '@/lib/types';

interface StepTechStackProps {
  catalog: Catalog;
  selectedTech: string[];
  onChange: (tech: string[]) => void;
  onNext: () => void;
  onBack: () => void;
}

export function StepTechStack({ catalog, selectedTech, onChange, onNext, onBack }: StepTechStackProps) {
  const allTechNames = Object.keys(catalog.tech_profiles).sort();

  const toggle = (name: string) => {
    onChange(
      selectedTech.includes(name)
        ? selectedTech.filter((t) => t !== name)
        : [...selectedTech, name]
    );
  };

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Tech Stack</h3>
      <p className="text-sm text-muted-foreground mb-6">
        Select technologies used in this project to get tailored recommendations.
      </p>

      <div className="flex flex-wrap gap-2">
        {allTechNames.map((name) => {
          const isSelected = selectedTech.includes(name);
          return (
            <button
              key={name}
              onClick={() => toggle(name)}
              className={cn(
                'px-3 py-1.5 rounded-md text-xs font-medium border transition-colors',
                isSelected
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border text-muted-foreground hover:border-primary/30'
              )}
            >
              {isSelected ? '\u2713 ' : ''}{name}
            </button>
          );
        })}
      </div>

      <div className="flex justify-between mt-8">
        <Button variant="outline" onClick={onBack}>Back</Button>
        <Button onClick={onNext}>Next</Button>
      </div>
    </div>
  );
}
