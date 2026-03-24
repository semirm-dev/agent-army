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
