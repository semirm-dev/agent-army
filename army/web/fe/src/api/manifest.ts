import { apiFetch } from './client';
import type { ManifestResponse, AddRemoveResult, SaveManifestRequest, SaveManifestResponse } from '../lib/types';

export function getManifest(scope?: string): Promise<ManifestResponse> {
  const params = scope ? `?scope=${scope}` : '';
  return apiFetch<ManifestResponse>(`/manifest${params}`);
}

export function addPlugin(name: string, project = false): Promise<AddRemoveResult> {
  return apiFetch<AddRemoveResult>('/manifest/plugin', {
    method: 'POST',
    body: JSON.stringify({ name, project }),
  });
}

export function removePlugin(name: string, project = false): Promise<AddRemoveResult> {
  const params = project ? '?project=true' : '';
  return apiFetch<AddRemoveResult>(`/manifest/plugin/${encodeURIComponent(name)}${params}`, {
    method: 'DELETE',
  });
}

export function addSkill(name: string, project = false): Promise<AddRemoveResult> {
  return apiFetch<AddRemoveResult>('/manifest/skill', {
    method: 'POST',
    body: JSON.stringify({ name, project }),
  });
}

export function removeSkill(name: string, project = false): Promise<AddRemoveResult> {
  const params = project ? '?project=true' : '';
  return apiFetch<AddRemoveResult>(`/manifest/skill/${encodeURIComponent(name)}${params}`, {
    method: 'DELETE',
  });
}

export function saveManifest(req: SaveManifestRequest): Promise<SaveManifestResponse> {
  return apiFetch<SaveManifestResponse>('/manifest/save', {
    method: 'POST',
    body: JSON.stringify(req),
  });
}
