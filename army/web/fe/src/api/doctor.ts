import { apiFetch } from './client';
import type { DoctorResponse } from '../lib/types';

export function getDoctor(): Promise<DoctorResponse> {
  return apiFetch<DoctorResponse>('/doctor');
}
