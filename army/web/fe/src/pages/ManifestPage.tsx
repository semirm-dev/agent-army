import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { Separator } from '@/components/ui/separator';
import { ScopeToggle } from '@/components/manifest/ScopeToggle';
import { ManifestList } from '@/components/manifest/ManifestList';
import { getManifest } from '@/api/manifest';

export function ManifestPage() {
  const [scope, setScope] = useState<'user' | 'project'>('user');

  const manifestQuery = useQuery({
    queryKey: ['manifest', scope],
    queryFn: () => getManifest(scope),
  });

  return (
    <div className="space-y-4 max-w-3xl">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Manifest</h2>
          <p className="text-sm text-muted-foreground">
            Manage your selected plugins and skills
          </p>
        </div>
        <ScopeToggle scope={scope} onChange={setScope} />
      </div>

      {manifestQuery.data && (
        <p className="text-xs text-muted-foreground">
          {manifestQuery.data.manifest_path}
        </p>
      )}

      <Separator />

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
