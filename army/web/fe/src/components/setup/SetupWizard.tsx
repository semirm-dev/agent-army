import { useState, useCallback, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { Stepper } from './Stepper';
import { StepDestination } from './StepDestination';
import { StepTechStack } from './StepTechStack';
import { StepPlugins } from './StepPlugins';
import { StepSkills } from './StepSkills';
import { StepConfirm } from './StepConfirm';
import { StepDone } from './StepDone';
import { getCatalog } from '@/api/catalog';
import { getManifest, saveManifest } from '@/api/manifest';

const ALL_STEPS = ['Destination', 'Tech', 'Plugins', 'Skills', 'Confirm', 'Done'];

export interface WizardState {
  destination: 'user' | 'project' | null;
  selectedTech: string[];
  selectedPlugins: string[];
  selectedSkills: string[];
}

export function SetupWizard() {
  const [currentStep, setCurrentStep] = useState(0);
  const [state, setState] = useState<WizardState>({
    destination: null,
    selectedTech: [],
    selectedPlugins: [],
    selectedSkills: [],
  });
  const [hasAutoSelected, setHasAutoSelected] = useState(false);
  const [hasPrePopulated, setHasPrePopulated] = useState(false);
  const [savedPath, setSavedPath] = useState('');

  const queryClient = useQueryClient();

  const catalogQuery = useQuery({
    queryKey: ['catalog'],
    queryFn: getCatalog,
  });

  const existingManifestQuery = useQuery({
    queryKey: ['manifest', state.destination],
    queryFn: () => getManifest(state.destination!),
    enabled: !!state.destination,
  });

  // Steps visible depend on destination
  const steps = state.destination === 'project'
    ? ALL_STEPS
    : ALL_STEPS.filter((s) => s !== 'Tech');

  const activeStepName = steps[currentStep];

  const next = useCallback(() => setCurrentStep((s) => Math.min(s + 1, steps.length - 1)), [steps.length]);
  const back = useCallback(() => setCurrentStep((s) => Math.max(s - 1, 0)), []);

  // Pre-populate from existing manifest
  useEffect(() => {
    if (existingManifestQuery.data && !hasPrePopulated) {
      const m = existingManifestQuery.data;
      if (m.plugins.length > 0 || m.skills.length > 0) {
        setState((s) => ({
          ...s,
          selectedPlugins: [...new Set([...s.selectedPlugins, ...m.plugins.map((p) => p.name)])],
          selectedSkills: [...new Set([...s.selectedSkills, ...m.skills.map((sk) => sk.name)])],
        }));
      }
      setHasPrePopulated(true);
    }
  }, [existingManifestQuery.data, hasPrePopulated]);

  // Auto-select recommended items when entering Plugins step
  useEffect(() => {
    if (activeStepName === 'Plugins' && !hasAutoSelected && catalogQuery.data) {
      const catalog = catalogQuery.data;
      const recPlugins = new Set<string>();
      const recSkills = new Set<string>();
      for (const tech of state.selectedTech) {
        const profile = catalog.tech_profiles[tech];
        if (profile) {
          for (const p of profile.plugins) recPlugins.add(p);
          for (const s of profile.skills) recSkills.add(s);
        }
      }
      if (recPlugins.size > 0 || recSkills.size > 0) {
        setState((s) => ({
          ...s,
          selectedPlugins: [...new Set([...s.selectedPlugins, ...recPlugins])],
          selectedSkills: [...new Set([...s.selectedSkills, ...recSkills])],
        }));
      }
      setHasAutoSelected(true);
    }
  }, [activeStepName, hasAutoSelected, catalogQuery.data, state.selectedTech]);

  const saveMutation = useMutation({
    mutationFn: saveManifest,
    onSuccess: (data) => {
      setSavedPath(data.path);
      queryClient.invalidateQueries({ queryKey: ['manifest'] });
      next();
    },
  });

  const handleSave = () => {
    if (!state.destination || !catalogQuery.data) return;
    const catalog = catalogQuery.data;
    saveMutation.mutate({
      destination: state.destination,
      plugins: state.selectedPlugins.map((name) => {
        const p = catalog.plugins.find((cp) => cp.name === name);
        return { name, marketplace: p?.marketplace ?? '', tags: p?.tags ?? [] };
      }),
      skills: state.selectedSkills.map((name) => {
        const s = catalog.skills.find((cs) => cs.name === name);
        return { name, source: s?.source ?? '', tags: s?.tags ?? [] };
      }),
    });
  };

  const handleDestinationChange = (dest: 'user' | 'project') => {
    setState({ destination: dest, selectedTech: [], selectedPlugins: [], selectedSkills: [] });
    setHasAutoSelected(false);
    setHasPrePopulated(false);
  };

  if (catalogQuery.isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (catalogQuery.isError || !catalogQuery.data) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500 text-sm">Failed to load catalog</p>
      </div>
    );
  }

  const catalog = catalogQuery.data;

  return (
    <div className="space-y-6">
      <Stepper steps={steps} currentStep={currentStep} onStepClick={setCurrentStep} />
      <div className="min-h-[400px]">
        {activeStepName === 'Destination' && (
          <StepDestination
            value={state.destination}
            onChange={handleDestinationChange}
            onNext={next}
          />
        )}
        {activeStepName === 'Tech' && (
          <StepTechStack
            catalog={catalog}
            selectedTech={state.selectedTech}
            onChange={(tech) => setState((s) => ({ ...s, selectedTech: tech }))}
            onNext={next}
            onBack={back}
          />
        )}
        {activeStepName === 'Plugins' && (
          <StepPlugins
            catalog={catalog}
            selectedTech={state.selectedTech}
            selectedPlugins={state.selectedPlugins}
            onChange={(plugins) => setState((s) => ({ ...s, selectedPlugins: plugins }))}
            onNext={next}
            onBack={back}
          />
        )}
        {activeStepName === 'Skills' && (
          <StepSkills
            catalog={catalog}
            selectedTech={state.selectedTech}
            selectedSkills={state.selectedSkills}
            onChange={(skills) => setState((s) => ({ ...s, selectedSkills: skills }))}
            onNext={next}
            onBack={back}
          />
        )}
        {activeStepName === 'Confirm' && state.destination && (
          <StepConfirm
            destination={state.destination}
            selectedPlugins={state.selectedPlugins}
            selectedSkills={state.selectedSkills}
            onSave={handleSave}
            onBack={back}
            isSaving={saveMutation.isPending}
            error={saveMutation.isError ? saveMutation.error.message : null}
          />
        )}
        {activeStepName === 'Done' && (
          <StepDone
            savedPath={savedPath}
            pluginCount={state.selectedPlugins.length}
            skillCount={state.selectedSkills.length}
          />
        )}
      </div>
    </div>
  );
}
