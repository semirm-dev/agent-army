import { useState, useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { CatalogSearch } from '@/components/catalog/CatalogSearch';
import { CatalogList } from '@/components/catalog/CatalogList';
import { getCatalog } from '@/api/catalog';
import { getManifest } from '@/api/manifest';
import { cn } from '@/lib/utils';

export function CatalogPage() {
  const [search, setSearch] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);

  const catalogQuery = useQuery({
    queryKey: ['catalog'],
    queryFn: getCatalog,
  });

  const manifestQuery = useQuery({
    queryKey: ['manifest'],
    queryFn: () => getManifest(),
  });

  const manifestNames = useMemo(() => {
    const names = new Set<string>();
    if (manifestQuery.data) {
      for (const p of manifestQuery.data.plugins) {
        names.add(p.name.toLowerCase());
      }
      for (const s of manifestQuery.data.skills) {
        names.add(s.name.toLowerCase());
      }
    }
    return names;
  }, [manifestQuery.data]);

  const allTags = useMemo(() => {
    if (!catalogQuery.data) return [];
    const tags = new Set<string>();
    for (const p of catalogQuery.data.plugins) {
      for (const t of p.tags) tags.add(t);
    }
    for (const s of catalogQuery.data.skills) {
      for (const t of s.tags) tags.add(t);
    }
    return Array.from(tags).sort();
  }, [catalogQuery.data]);

  const toggleTag = (tag: string) => {
    setSelectedTags((prev) =>
      prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]
    );
  };

  if (catalogQuery.isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (catalogQuery.isError) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500 text-sm">
          Failed to load catalog: {catalogQuery.error.message}
        </p>
      </div>
    );
  }

  const catalog = catalogQuery.data;
  if (!catalog) return null;

  return (
    <div className="space-y-4 max-w-5xl">
      <div>
        <h2 className="text-2xl font-bold">Catalog</h2>
        <p className="text-sm text-muted-foreground">
          Browse available plugins and skills
        </p>
      </div>

      <CatalogSearch value={search} onChange={setSearch} />

      {allTags.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {allTags.map((tag) => (
            <Badge
              key={tag}
              variant={selectedTags.includes(tag) ? 'default' : 'outline'}
              className={cn('cursor-pointer text-xs', selectedTags.includes(tag) && 'bg-primary')}
              onClick={() => toggleTag(tag)}
            >
              {tag}
            </Badge>
          ))}
        </div>
      )}

      <Tabs defaultValue="plugins">
        <TabsList>
          <TabsTrigger value="plugins">
            Plugins ({catalog.plugins.length})
          </TabsTrigger>
          <TabsTrigger value="skills">
            Skills ({catalog.skills.length})
          </TabsTrigger>
        </TabsList>
        <TabsContent value="plugins" className="mt-4">
          <CatalogList
            items={catalog.plugins}
            type="plugin"
            search={search}
            selectedTags={selectedTags}
            manifestNames={manifestNames}
          />
        </TabsContent>
        <TabsContent value="skills" className="mt-4">
          <CatalogList
            items={catalog.skills}
            type="skill"
            search={search}
            selectedTags={selectedTags}
            manifestNames={manifestNames}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
