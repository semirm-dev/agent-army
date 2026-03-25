import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Loader2, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { CatalogSearch } from '@/components/catalog/CatalogSearch';
import { CatalogList } from '@/components/catalog/CatalogList';
import { getCatalog, fetchCatalog } from '@/api/catalog';
import { getManifest } from '@/api/manifest';
import { cn } from '@/lib/utils';

export function CatalogPage() {
  const [search, setSearch] = useState('');
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [activeTab, setActiveTab] = useState<'plugins' | 'skills'>('plugins');

  const queryClient = useQueryClient();

  const fetchMutation = useMutation({
    mutationFn: fetchCatalog,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['catalog'] });
    },
  });

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
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Catalog</h2>
          <p className="text-sm text-muted-foreground">
            Browse available plugins and skills
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => fetchMutation.mutate()}
          disabled={fetchMutation.isPending}
        >
          {fetchMutation.isPending ? (
            <Loader2 className="size-3.5 animate-spin" />
          ) : (
            <RefreshCw className="size-3.5" />
          )}
          {fetchMutation.isPending ? 'Fetching...' : 'Fetch Latest'}
        </Button>
      </div>
      {fetchMutation.isSuccess && (
        <p className="text-xs text-green-500">Catalog updated</p>
      )}
      {fetchMutation.isError && (
        <p className="text-xs text-red-500">
          Fetch failed: {fetchMutation.error.message}
        </p>
      )}

      <CatalogSearch value={search} onChange={setSearch} />

      {allTags.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {allTags.map((tag) => (
            <button
              key={tag}
              className={cn(
                'px-2.5 py-1 rounded-md text-xs border transition-colors cursor-pointer',
                selectedTags.includes(tag)
                  ? 'bg-primary/15 text-primary border-primary/30'
                  : 'bg-muted border-border text-muted-foreground'
              )}
              onClick={() => toggleTag(tag)}
            >
              {tag}
            </button>
          ))}
        </div>
      )}

      <div>
        <div className="flex gap-4 border-b border-border">
          <button
            className={cn(
              'pb-2 text-sm font-medium transition-colors',
              activeTab === 'plugins'
                ? 'border-b-2 border-primary text-primary'
                : 'text-muted-foreground hover:text-foreground'
            )}
            onClick={() => setActiveTab('plugins')}
          >
            Plugins ({catalog.plugins.length})
          </button>
          <button
            className={cn(
              'pb-2 text-sm font-medium transition-colors',
              activeTab === 'skills'
                ? 'border-b-2 border-primary text-primary'
                : 'text-muted-foreground hover:text-foreground'
            )}
            onClick={() => setActiveTab('skills')}
          >
            Skills ({catalog.skills.length})
          </button>
        </div>
        <div className="mt-4">
          {activeTab === 'plugins' ? (
            <CatalogList
              items={catalog.plugins}
              type="plugin"
              search={search}
              selectedTags={selectedTags}
              manifestNames={manifestNames}
            />
          ) : (
            <CatalogList
              items={catalog.skills}
              type="skill"
              search={search}
              selectedTags={selectedTags}
              manifestNames={manifestNames}
            />
          )}
        </div>
      </div>
    </div>
  );
}
