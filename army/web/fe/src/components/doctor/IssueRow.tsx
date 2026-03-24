import { AlertCircle, AlertTriangle, Info } from 'lucide-react';
import type { DoctorIssue } from '@/lib/types';

interface IssueRowProps {
  issue: DoctorIssue;
}

const severityConfig = {
  error: { icon: AlertCircle, className: 'text-red-500' },
  warning: { icon: AlertTriangle, className: 'text-yellow-500' },
  info: { icon: Info, className: 'text-blue-500' },
} as const;

export function IssueRow({ issue }: IssueRowProps) {
  const config = severityConfig[issue.severity];
  const Icon = config.icon;

  return (
    <div className="flex items-start gap-3 py-2.5 px-3 rounded-md border border-border bg-card hover:border-primary/20 transition-colors">
      <Icon className={`size-3.5 mt-0.5 shrink-0 ${config.className}`} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-0.5">
          <span className="font-mono text-[10px] px-1.5 py-0.5 rounded bg-muted border border-border text-muted-foreground">
            {issue.category}
          </span>
          {issue.item && (
            <span className="font-mono text-xs font-medium">
              {issue.item}
            </span>
          )}
        </div>
        <p className="text-xs text-muted-foreground">{issue.description}</p>
      </div>
    </div>
  );
}
