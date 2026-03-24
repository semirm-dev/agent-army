import { apiStreamUrl } from './client';

export function createSyncUrl(destination?: string): { url: string; body: string } {
  return {
    url: apiStreamUrl('/sync'),
    body: JSON.stringify({ destination }),
  };
}

export function startSyncFetch(
  destination?: string,
  signal?: AbortSignal
): Promise<Response> {
  const { url, body } = createSyncUrl(destination);
  return fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
    signal,
  });
}
