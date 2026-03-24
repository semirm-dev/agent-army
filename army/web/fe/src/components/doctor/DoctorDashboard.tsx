import { CheckCircle2, AlertCircle, AlertTriangle } from 'lucide-react';
import { Card, CardContent } from '@/components/ui/card';
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
      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <AlertCircle className="size-5 text-red-500" />
            <div>
              <p className="text-2xl font-bold">{summary.errors}</p>
              <p className="text-xs text-muted-foreground">Errors</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <AlertTriangle className="size-5 text-yellow-500" />
            <div>
              <p className="text-2xl font-bold">{summary.warnings}</p>
              <p className="text-xs text-muted-foreground">Warnings</p>
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="flex items-center gap-3 pt-6">
            <CheckCircle2 className="size-5 text-green-500" />
            <div>
              <p className="text-2xl font-bold">
                {isHealthy ? 'OK' : issues.filter((i) => i.severity === 'info').length}
              </p>
              <p className="text-xs text-muted-foreground">
                {isHealthy ? 'All Healthy' : 'Info'}
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      {isHealthy ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <CheckCircle2 className="size-12 text-green-500 mb-3" />
            <p className="text-lg font-medium">All Healthy</p>
            <p className="text-sm text-muted-foreground">
              No issues detected. Everything is in sync.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          <h3 className="text-sm font-medium text-muted-foreground">
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
