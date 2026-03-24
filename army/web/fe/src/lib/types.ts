// Mirrors Go types from army/internal/core/types/types.go

export interface CatalogPlugin {
  name: string;
  marketplace: string;
  description: string;
  tags: string[];
}

export interface CatalogSkill {
  name: string;
  source: string;
  description: string;
  tags: string[];
}

export interface TechProfile {
  detect: string[];
  plugins: string[];
  skills: string[];
}

export interface Catalog {
  version: number;
  updated_at: string;
  plugins: CatalogPlugin[];
  skills: CatalogSkill[];
  tech_profiles: Record<string, TechProfile>;
}

export interface ManifestPluginItem {
  name: string;
  marketplace: string;
  destination: string;
  tags: string[];
  installed: boolean;
  status: 'ok' | 'missing' | 'drift';
}

export interface ManifestSkillItem {
  name: string;
  source: string;
  destination: string;
  tags: string[];
  installed: boolean;
  status: 'ok' | 'missing' | 'drift';
}

export interface ManifestResponse {
  manifest_path: string;
  manifest_scope: 'user' | 'project';
  plugins: ManifestPluginItem[];
  skills: ManifestSkillItem[];
}

export interface DoctorIssue {
  severity: 'error' | 'warning' | 'info';
  category: string;
  description: string;
  item: string;
}

export interface DoctorResponse {
  issues: DoctorIssue[];
  summary: {
    errors: number;
    warnings: number;
  };
}

export interface SyncAction {
  type: 'install' | 'remove';
  item_type: 'plugin' | 'skill';
  name: string;
  source: string;
  destination: string;
}

export type SyncEvent =
  | { event: 'plan'; actions: SyncAction[] }
  | { event: 'action_start'; type: string; item_type: string; name: string }
  | { event: 'action_done'; type: string; item_type: string; name: string; success: boolean; error?: string }
  | { event: 'complete'; succeeded: number; failed: number }
  | { event: 'exit'; code: number }
  | { event: 'error'; message: string };

export interface AddRemoveResult {
  action: 'add' | 'remove';
  item_type: 'plugin' | 'skill';
  name: string;
  added_to_manifest?: boolean;
  removed_from_manifest?: boolean;
  installed?: boolean;
  uninstalled?: boolean;
  error: string;
}
