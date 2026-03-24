import { IssueRow } from './IssueRow';
import type { DoctorResponse } from '@/lib/types';

interface DoctorDashboardProps {
  data: DoctorResponse;
}

export function DoctorDashboard({ data }: DoctorDashboardProps) {
  const { issues, summary } = data;
  const isHealthy = summary.errors === 0 && summary.warnings === 0;

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-3 gap-3">
        <div className="rounded-lg border border-border bg-card p-4 text-center">
          <p className="text-2xl font-bold text-red-400">{summary.errors}</p>
          <p className="text-[10px] tracking-wider text-muted-foreground uppercase mt-1">Errors</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 text-center">
          <p className="text-2xl font-bold text-yellow-400">{summary.warnings}</p>
          <p className="text-[10px] tracking-wider text-muted-foreground uppercase mt-1">Warnings</p>
        </div>
        <div className="rounded-lg border border-border bg-card p-4 text-center">
          <p className="text-2xl font-bold text-primary">
            {isHealthy ? 'OK' : issues.filter((i) => i.severity === 'info').length}
          </p>
          <p className="text-[10px] tracking-wider text-muted-foreground uppercase mt-1">
            {isHealthy ? 'Healthy' : 'Info'}
          </p>
        </div>
      </div>

      {isHealthy ? (
        <div className="rounded-lg border border-border bg-card py-12 text-center">
          <p className="font-mono text-primary text-sm">All systems operational</p>
          <p className="text-xs text-muted-foreground mt-1">
            No issues detected. Everything is in sync.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          <h3 className="text-[11px] font-medium tracking-wider text-muted-foreground uppercase">
            Issues ({issues.length})
          </h3>
          {issues.map((issue, i) => (
            <IssueRow key={`${issue.category}-${issue.item}-${i}`} issue={issue} />
          ))}
        </div>
      )}
    </div>
  );
}
