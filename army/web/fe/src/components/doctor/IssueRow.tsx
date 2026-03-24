import { AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
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
    <div className="flex items-start gap-3 py-3 px-4 rounded-lg border">
      <Icon className={`size-4 mt-0.5 shrink-0 ${config.className}`} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-1">
          <Badge variant="outline" className="text-[10px]">
            {issue.category}
          </Badge>
          {issue.item && (
            <span className="text-xs font-medium text-muted-foreground">
              {issue.item}
            </span>
          )}
        </div>
        <p className="text-sm">{issue.description}</p>
      </div>
    </div>
  );
}
