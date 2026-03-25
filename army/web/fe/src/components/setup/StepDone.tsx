import { useNavigate } from 'react-router-dom';
import { Check } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface StepDoneProps {
  savedPath: string;
  pluginCount: number;
  skillCount: number;
}

export function StepDone({ savedPath, pluginCount, skillCount }: StepDoneProps) {
  const navigate = useNavigate();

  return (
    <div className="text-center py-8">
      <div className="size-12 rounded-full bg-green-500/10 text-green-500 flex items-center justify-center mx-auto mb-4">
        <Check className="size-6" />
      </div>
      <h3 className="text-lg font-semibold text-green-500 mb-1">Manifest Saved!</h3>
      <p className="text-sm text-muted-foreground mb-6">
        {pluginCount} plugins and {skillCount} skills saved to{' '}
        <span className="font-mono text-xs">{savedPath}</span>
      </p>
      <div className="flex gap-3 justify-center">
        <Button onClick={() => navigate('/sync?autostart=true')}>Run Sync Now</Button>
        <Button variant="outline" onClick={() => navigate('/catalog')}>Back to Catalog</Button>
      </div>
    </div>
  );
}
