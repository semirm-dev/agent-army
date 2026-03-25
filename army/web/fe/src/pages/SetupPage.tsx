import { SetupWizard } from '@/components/setup/SetupWizard';

export function SetupPage() {
  return (
    <div className="space-y-4 max-w-3xl">
      <div>
        <h2 className="text-xl font-semibold">Setup</h2>
        <p className="text-sm text-muted-foreground">
          Configure your plugins and skills
        </p>
      </div>
      <SetupWizard />
    </div>
  );
}
