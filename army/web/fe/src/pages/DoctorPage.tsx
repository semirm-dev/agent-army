import { useQuery } from '@tanstack/react-query';
import { Loader2, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { DoctorDashboard } from '@/components/doctor/DoctorDashboard';
import { getDoctor } from '@/api/doctor';

export function DoctorPage() {
  const doctorQuery = useQuery({
    queryKey: ['doctor'],
    queryFn: getDoctor,
  });

  return (
    <div className="space-y-4 max-w-3xl">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold">Doctor</h2>
          <p className="text-sm text-muted-foreground">
            Health check for your plugins and skills
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => doctorQuery.refetch()}
          disabled={doctorQuery.isFetching}
        >
          <RefreshCw className={`size-4 ${doctorQuery.isFetching ? 'animate-spin' : ''}`} />
          Re-check
        </Button>
      </div>

      {doctorQuery.isLoading ? (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="size-6 animate-spin text-muted-foreground" />
        </div>
      ) : doctorQuery.isError ? (
        <div className="text-center py-12">
          <p className="text-red-500 text-sm">
            Failed to run health check: {doctorQuery.error.message}
          </p>
        </div>
      ) : doctorQuery.data ? (
        <DoctorDashboard data={doctorQuery.data} />
      ) : null}
    </div>
  );
}
