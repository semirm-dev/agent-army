import { apiFetch } from './client';
import type { Catalog } from '../lib/types';

export function getCatalog(): Promise<Catalog> {
  return apiFetch<Catalog>('/catalog');
}
