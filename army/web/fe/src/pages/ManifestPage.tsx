import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Loader2, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ScopeToggle } from '@/components/manifest/ScopeToggle';
import { ManifestList } from '@/components/manifest/ManifestList';
import { getManifest } from '@/api/manifest';

export function ManifestPage() {
  const [scope, setScope] = useState<'user' | 'project'>('user');
  const queryClient = useQueryClient();

  const manifestQuery = useQuery({
    queryKey: ['manifest', scope],
    queryFn: () => getManifest(scope),
  });

  const handleRefresh = () => {
    queryClient.invalidateQueries({ queryKey: ['manifest', scope] });
  };

  return (
    <div className="space-y-4 max-w-3xl">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Manifest</h2>
          <p className="text-sm text-muted-foreground">
            Manage your selected plugins and skills
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={manifestQuery.isFetching}
          >
            {manifestQuery.isFetching ? (
              <Loader2 className="size-3.5 animate-spin" />
            ) : (
              <RefreshCw className="size-3.5" />
            )}
            {manifestQuery.isFetching ? 'Updating...' : 'Update'}
          </Button>
          <ScopeToggle scope={scope} onChange={setScope} />
        </div>
      </div>

      {manifestQuery.data && (
        <p className="text-xs text-muted-foreground font-mono">
          {manifestQuery.data.manifest_path}
        </p>
      )}

      {manifestQuery.isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : manifestQuery.isError ? (
        <div className="text-center py-12">
          <p className="text-red-500 text-sm">
            Failed to load manifest: {manifestQuery.error.message}
          </p>
        </div>
      ) : manifestQuery.data ? (
        <ManifestList
          plugins={manifestQuery.data.plugins}
          skills={manifestQuery.data.skills}
        />
      ) : null}
    </div>
  );
}
