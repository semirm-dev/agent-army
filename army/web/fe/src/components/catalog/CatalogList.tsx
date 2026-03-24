import { useMemo } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { CatalogCard } from './CatalogCard';
import { addPlugin, addSkill } from '@/api/manifest';
import type { CatalogPlugin, CatalogSkill } from '@/lib/types';

interface CatalogListProps {
  items: (CatalogPlugin | CatalogSkill)[];
  type: 'plugin' | 'skill';
  search: string;
  selectedTags: string[];
  manifestNames: Set<string>;
}

export function CatalogList({
  items,
  type,
  search,
  selectedTags,
  manifestNames,
}: CatalogListProps) {
  const queryClient = useQueryClient();

  const addMutation = useMutation({
    mutationFn: (name: string) =>
      type === 'plugin' ? addPlugin(name) : addSkill(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['manifest'] });
    },
  });

  const filtered = useMemo(() => {
    return items.filter((item) => {
      const matchesSearch =
        !search ||
        item.name.toLowerCase().includes(search.toLowerCase()) ||
        item.description.toLowerCase().includes(search.toLowerCase());
      const matchesTags =
        selectedTags.length === 0 ||
        selectedTags.some((tag) => item.tags.includes(tag));
      return matchesSearch && matchesTags;
    });
  }, [items, search, selectedTags]);

  if (filtered.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        No {type}s found matching your criteria.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
      {filtered.map((item) => {
        const source =
          'marketplace' in item ? item.marketplace : item.source;
        return (
          <CatalogCard
            key={item.name}
            name={item.name}
            description={item.description}
            source={source}
            tags={item.tags}
            inManifest={manifestNames.has(item.name.toLowerCase())}
            onAdd={() => addMutation.mutate(item.name)}
            isAdding={
              addMutation.isPending &&
              addMutation.variables === item.name
            }
          />
        );
      })}
    </div>
  );
}
