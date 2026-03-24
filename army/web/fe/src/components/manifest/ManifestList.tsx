import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ManifestItem } from './ManifestItem';
import { removePlugin, removeSkill } from '@/api/manifest';
import type { ManifestPluginItem, ManifestSkillItem } from '@/lib/types';

interface ManifestListProps {
  plugins: ManifestPluginItem[];
  skills: ManifestSkillItem[];
}

export function ManifestList({ plugins, skills }: ManifestListProps) {
  const queryClient = useQueryClient();

  const removePluginMutation = useMutation({
    mutationFn: (name: string) => removePlugin(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['manifest'] });
    },
  });

  const removeSkillMutation = useMutation({
    mutationFn: (name: string) => removeSkill(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['manifest'] });
    },
  });

  const pluginCount = plugins.filter((p) => p.installed).length;
  const skillCount = skills.filter((s) => s.installed).length;

  return (
    <div className="space-y-6">
      {plugins.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground mb-3">
            Plugins ({pluginCount}/{plugins.length} installed)
          </h3>
          <div className="space-y-2">
            {plugins.map((plugin) => (
              <ManifestItem
                key={plugin.name}
                name={plugin.name}
                source={plugin.marketplace}
                destination={plugin.destination}
                status={plugin.status}
                installed={plugin.installed}
                onRemove={() => removePluginMutation.mutate(plugin.name)}
                isRemoving={
                  removePluginMutation.isPending &&
                  removePluginMutation.variables === plugin.name
                }
              />
            ))}
          </div>
        </div>
      )}

      {skills.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-muted-foreground mb-3">
            Skills ({skillCount}/{skills.length} installed)
          </h3>
          <div className="space-y-2">
            {skills.map((skill) => (
              <ManifestItem
                key={skill.name}
                name={skill.name}
                source={skill.source}
                destination={skill.destination}
                status={skill.status}
                installed={skill.installed}
                onRemove={() => removeSkillMutation.mutate(skill.name)}
                isRemoving={
                  removeSkillMutation.isPending &&
                  removeSkillMutation.variables === skill.name
                }
              />
            ))}
          </div>
        </div>
      )}

      {plugins.length === 0 && skills.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No items in manifest. Add some from the Catalog.
        </div>
      )}
    </div>
  );
}
