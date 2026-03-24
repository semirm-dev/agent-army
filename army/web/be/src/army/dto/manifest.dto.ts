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
