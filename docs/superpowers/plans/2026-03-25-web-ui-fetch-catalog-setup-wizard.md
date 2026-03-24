# Web UI: Fetch Catalog & Setup Wizard — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add "Fetch Catalog" button to the web Catalog page and a full Setup Wizard at `/setup` that mirrors the TUI wizard.

**Architecture:** Two features sharing the same web stack (NestJS backend calling Go CLI via execFile, React frontend with TanStack Query). Feature 1 is a small addition to an existing page. Feature 2 is a new page with stepper wizard, reusing existing API endpoints plus one new `POST /api/manifest` and one new Go CLI command `write-manifest`.

**Tech Stack:** Go (Cobra CLI), NestJS (TypeScript), React 19, TanStack Query, Tailwind CSS, lucide-react icons

**Spec:** `docs/superpowers/specs/2026-03-25-web-ui-fetch-catalog-setup-wizard-design.md`

---

## File Structure

### New Go files
| File | Responsibility |
|------|---------------|
| `army/cli/write_manifest.go` | `army write-manifest --destination <user\|project>` command — reads manifest JSON from stdin, writes atomically |

### New Frontend files
| File | Responsibility |
|------|---------------|
| `army/web/fe/src/pages/SetupPage.tsx` | Route page, renders SetupWizard |
| `army/web/fe/src/components/setup/SetupWizard.tsx` | Main wizard: stepper, step routing, state management |
| `army/web/fe/src/components/setup/Stepper.tsx` | Horizontal numbered step indicator |
| `army/web/fe/src/components/setup/StepDestination.tsx` | Step 1: user/project radio cards |
| `army/web/fe/src/components/setup/StepTechStack.tsx` | Step 2: toggleable tech chips |
| `army/web/fe/src/components/setup/StepPlugins.tsx` | Step 3: plugin multi-select |
| `army/web/fe/src/components/setup/StepSkills.tsx` | Step 4: skill multi-select |
| `army/web/fe/src/components/setup/SelectableList.tsx` | Shared multi-select list with search (used by Steps 3 & 4) |
| `army/web/fe/src/components/setup/StepConfirm.tsx` | Step 5: summary + save |
| `army/web/fe/src/components/setup/StepDone.tsx` | Step 6: success + sync offer |
### Modified files
| File | Change |
|------|--------|
| `army/cli/fetch_catalog.go` | Add `--json` output support |
| `army/cli/root.go` | Register `write-manifest` command |
| `army/web/be/src/army/army-catalog.controller.ts` | Add `POST /catalog/fetch` endpoint |
| `army/web/be/src/army/army-manifest.controller.ts` | Add `POST /manifest/save` endpoint |
| `army/web/be/src/army/army.service.ts` | Add `execWithInput()` method |
| `army/web/fe/src/api/catalog.ts` | Add `fetchCatalog()` |
| `army/web/fe/src/api/manifest.ts` | Add `saveManifest()` |
| `army/web/fe/src/lib/types.ts` | Add `SaveManifestRequest`, `SaveManifestResponse` types |
| `army/web/fe/src/pages/CatalogPage.tsx` | Add fetch button |
| `army/web/fe/src/pages/SyncPage.tsx` | Add `autostart` query param support |
| `army/web/fe/src/App.tsx` | Add `/setup` route |
| `army/web/fe/src/components/layout/Sidebar.tsx` | Add Setup nav item, update version to 0.4.0 |
| `VERSION` | Bump to `0.4.0` |

---

## Task 1: Go — Add `--json` to `fetch-catalog`

**Files:**
- Modify: `army/cli/fetch_catalog.go`

- [ ] **Step 1: Add JSON output to fetch-catalog command**

In `army/cli/fetch_catalog.go`, after the catalog is validated and written to disk (line 66), add a JSON output path. The data is already validated and parsed by `validateCatalog`, so we parse it into `types.Catalog` and encode to stdout:

```go
// Replace lines 66-67 in fetch_catalog.go with:
if globalFlags.JSON {
    return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
        "path":    path,
        "updated": true,
    })
}

fmt.Printf("Catalog saved: %s\n", path)
return nil
```

- [ ] **Step 2: Build and verify**

Run: `cd army && go build ./cmd/army/`
Expected: builds cleanly

Run: `cd army && go test ./... -race -count=1`
Expected: all tests pass

- [ ] **Step 3: Commit**

```bash
git add army/cli/fetch_catalog.go
git commit -m "feat: add --json output support to fetch-catalog command"
```

---

## Task 2: Go — Add `write-manifest` command

**Files:**
- Create: `army/cli/write_manifest.go`
- Modify: `army/cli/root.go`

- [ ] **Step 1: Create write_manifest.go**

Create `army/cli/write_manifest.go`:

```go
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/smahovkic/agent-army/army/internal/core/manifest"
	"github.com/smahovkic/agent-army/army/internal/core/types"
	"github.com/spf13/cobra"
)

func newWriteManifestCmd() *cobra.Command {
	var destination string

	cmd := &cobra.Command{
		Use:   "write-manifest",
		Short: "Write a complete manifest from JSON on stdin",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			var m types.Manifest
			if err := json.Unmarshal(data, &m); err != nil {
				return fmt.Errorf("parsing manifest JSON: %w", err)
			}
			if m.Version == 0 {
				m.Version = 1
			}

			// Stamp destination on all items
			for i := range m.Plugins {
				m.Plugins[i].Destination = destination
			}
			for i := range m.Skills {
				m.Skills[i].Destination = destination
			}

			// Determine path
			var savePath string
			if destination == "project" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				savePath = filepath.Join(cwd, ".army", "manifest.json")
			} else {
				p, err := manifest.DefaultPath()
				if err != nil {
					return err
				}
				savePath = p
			}

			if err := manifest.Save(savePath, &m); err != nil {
				return fmt.Errorf("saving manifest: %w", err)
			}

			if globalFlags.JSON {
				return json.NewEncoder(os.Stdout).Encode(map[string]string{
					"path": savePath,
				})
			}

			fmt.Printf("Manifest saved: %s\n", savePath)
			return nil
		},
	}

	cmd.Flags().StringVar(&destination, "destination", "user", "Manifest destination: user or project")
	return cmd
}
```

- [ ] **Step 2: Register in root.go**

In `army/cli/root.go`, add `newWriteManifestCmd()` to `cmd.AddCommand(...)` block (after `newFetchCatalogCmd()`):

```go
newWriteManifestCmd(),
```

- [ ] **Step 3: Build and verify**

Run: `cd army && go build ./cmd/army/`
Expected: builds cleanly

Run: `cd army && go test ./... -race -count=1`
Expected: all tests pass

Test manually:
```bash
echo '{"version":1,"plugins":[],"skills":[]}' | ./army/army write-manifest --destination user --json
```
Expected: `{"path":"/Users/.../.army/manifest.json"}`

- [ ] **Step 4: Commit**

```bash
git add army/cli/write_manifest.go army/cli/root.go
git commit -m "feat: add write-manifest command for web UI manifest saving"
```

---

## Task 3: Backend — Add `execWithInput` and catalog fetch endpoint

**Files:**
- Modify: `army/web/be/src/army/army.service.ts`
- Modify: `army/web/be/src/army/army-catalog.controller.ts`

- [ ] **Step 1: Add `execWithInput` to ArmyService**

In `army/web/be/src/army/army.service.ts`, add after the existing `execStream` method (line 32). Note: the existing codebase already uses `execFile` (not `exec`) which is the safe approach:

```typescript
async execWithInput<T = unknown>(args: string[], input: string): Promise<T> {
  const { stdout } = await execFileAsync(this.bin, [...args, '--json'], {
    cwd: this.cwd,
    env: { ...process.env },
    maxBuffer: 10 * 1024 * 1024,
    input,
  });
  return JSON.parse(stdout) as T;
}
```

- [ ] **Step 2: Add fetch endpoint to catalog controller**

In `army/web/be/src/army/army-catalog.controller.ts`, add a new method after `refreshCatalog()`:

```typescript
@Post('fetch')
async fetchCatalog() {
  await this.army.exec(['fetch-catalog']);
  this.cache = null;
  this.cache = await this.army.exec(['catalog']);
  return this.cache;
}
```

- [ ] **Step 3: Build backend**

Run: `cd army/web/be && npm run build`
Expected: builds cleanly

- [ ] **Step 4: Commit**

```bash
git add army/web/be/src/army/army.service.ts army/web/be/src/army/army-catalog.controller.ts
git commit -m "feat: add execWithInput method and catalog fetch endpoint"
```

---

## Task 4: Backend — Add `POST /manifest` endpoint

**Files:**
- Modify: `army/web/be/src/army/army-manifest.controller.ts`

- [ ] **Step 1: Add save manifest endpoint**

In `army/web/be/src/army/army-manifest.controller.ts`, add a new method at the end of the class. This uses a bare `@Post()` alongside the existing `@Post('plugin')` and `@Post('skill')`:

```typescript
// Must appear AFTER @Post('plugin') and @Post('skill') to avoid route conflicts
@Post('save')
async saveManifest(
  @Body() body: {
    destination: 'user' | 'project';
    plugins: Array<{ name: string; marketplace: string; tags: string[] }>;
    skills: Array<{ name: string; source: string; tags: string[] }>;
  },
) {
  const manifest = {
    version: 1,
    plugins: body.plugins,
    skills: body.skills,
  };
  return this.army.execWithInput(
    ['write-manifest', '--destination', body.destination],
    JSON.stringify(manifest),
  );
}
```

- [ ] **Step 2: Build backend**

Run: `cd army/web/be && npm run build`
Expected: builds cleanly

- [ ] **Step 3: Commit**

```bash
git add army/web/be/src/army/army-manifest.controller.ts
git commit -m "feat: add POST /manifest endpoint for full manifest save"
```

---

## Task 5: Frontend — Add Fetch Catalog button

**Files:**
- Modify: `army/web/fe/src/api/catalog.ts`
- Modify: `army/web/fe/src/pages/CatalogPage.tsx`

- [ ] **Step 1: Add fetchCatalog API function**

In `army/web/fe/src/api/catalog.ts`, add:

```typescript
export function fetchCatalog(): Promise<Catalog> {
  return apiFetch<Catalog>('/catalog/fetch', { method: 'POST' });
}
```

- [ ] **Step 2: Add fetch button to CatalogPage**

In `army/web/fe/src/pages/CatalogPage.tsx`:

Add to imports:
```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Loader2, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { fetchCatalog } from '@/api/catalog';
```

Inside the `CatalogPage` function, before the `return`, add:
```typescript
const queryClient = useQueryClient();

const fetchMutation = useMutation({
  mutationFn: fetchCatalog,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['catalog'] });
  },
});
```

Replace the header `<div>` (lines 79-84) with:
```tsx
<div className="flex items-center justify-between">
  <div>
    <h2 className="text-xl font-semibold">Catalog</h2>
    <p className="text-sm text-muted-foreground">
      Browse available plugins and skills
    </p>
  </div>
  <Button
    variant="outline"
    size="sm"
    onClick={() => fetchMutation.mutate()}
    disabled={fetchMutation.isPending}
  >
    {fetchMutation.isPending ? (
      <Loader2 className="size-3.5 animate-spin" />
    ) : (
      <RefreshCw className="size-3.5" />
    )}
    {fetchMutation.isPending ? 'Fetching...' : 'Fetch Latest'}
  </Button>
</div>
{fetchMutation.isSuccess && (
  <p className="text-xs text-green-500">Catalog updated</p>
)}
{fetchMutation.isError && (
  <p className="text-xs text-red-500">
    Fetch failed: {fetchMutation.error.message}
  </p>
)}
```

- [ ] **Step 3: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 4: Commit**

```bash
git add army/web/fe/src/api/catalog.ts army/web/fe/src/pages/CatalogPage.tsx
git commit -m "feat: add Fetch Latest button to Catalog page"
```

---

## Task 6: Frontend — Add types and saveManifest API

**Files:**
- Modify: `army/web/fe/src/lib/types.ts`
- Modify: `army/web/fe/src/api/manifest.ts`

Note: The `GET /api/detect` endpoint returns system config info (paths, existence flags), not detected tech names. The TUI's tech detection uses the Go `detector` package directly. For the web wizard, we show all tech profiles from the catalog for manual selection. A detect API file is not needed.

- [ ] **Step 1: Add new types**

In `army/web/fe/src/lib/types.ts`, add at the end:

```typescript
export interface SaveManifestRequest {
  destination: 'user' | 'project';
  plugins: Array<{ name: string; marketplace: string; tags: string[] }>;
  skills: Array<{ name: string; source: string; tags: string[] }>;
}

export interface SaveManifestResponse {
  path: string;
}
```

- [ ] **Step 2: Add saveManifest to manifest API**

In `army/web/fe/src/api/manifest.ts`, update the import line to include the new types:

```typescript
import type { ManifestResponse, AddRemoveResult, SaveManifestRequest, SaveManifestResponse } from '../lib/types';
```

Add at the end of the file:
```typescript
export function saveManifest(req: SaveManifestRequest): Promise<SaveManifestResponse> {
  return apiFetch<SaveManifestResponse>('/manifest/save', {
    method: 'POST',
    body: JSON.stringify(req),
  });
}
```

- [ ] **Step 3: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 4: Commit**

```bash
git add army/web/fe/src/lib/types.ts army/web/fe/src/api/manifest.ts
git commit -m "feat: add saveManifest API and new types"
```

---

## Task 7: Frontend — Setup Wizard scaffold (route, sidebar, stepper)

**Files:**
- Create: `army/web/fe/src/pages/SetupPage.tsx`
- Create: `army/web/fe/src/components/setup/Stepper.tsx`
- Create: `army/web/fe/src/components/setup/SetupWizard.tsx`
- Create: `army/web/fe/src/components/setup/StepDestination.tsx`
- Modify: `army/web/fe/src/App.tsx`
- Modify: `army/web/fe/src/components/layout/Sidebar.tsx`

- [ ] **Step 1: Create Stepper component**

Create `army/web/fe/src/components/setup/Stepper.tsx`:

```tsx
import { Check } from 'lucide-react';
import { cn } from '@/lib/utils';

interface StepperProps {
  steps: string[];
  currentStep: number;
  onStepClick: (step: number) => void;
}

export function Stepper({ steps, currentStep, onStepClick }: StepperProps) {
  return (
    <div className="flex items-center gap-2">
      {steps.map((label, i) => {
        const isCompleted = i < currentStep;
        const isCurrent = i === currentStep;
        return (
          <div key={label} className="flex items-center gap-2">
            {i > 0 && (
              <div
                className={cn(
                  'h-px w-6',
                  i <= currentStep ? 'bg-primary' : 'bg-border'
                )}
              />
            )}
            <button
              onClick={() => isCompleted && onStepClick(i)}
              disabled={!isCompleted}
              className={cn(
                'flex items-center gap-1.5 text-xs font-medium transition-colors',
                isCompleted && 'cursor-pointer text-primary hover:text-primary/80',
                isCurrent && 'text-primary',
                !isCompleted && !isCurrent && 'text-muted-foreground cursor-default'
              )}
            >
              <span
                className={cn(
                  'size-6 rounded-full flex items-center justify-center text-[11px] font-bold shrink-0',
                  isCompleted && 'bg-primary text-primary-foreground',
                  isCurrent && 'bg-primary text-primary-foreground',
                  !isCompleted && !isCurrent && 'bg-muted text-muted-foreground'
                )}
              >
                {isCompleted ? <Check className="size-3" /> : i + 1}
              </span>
              <span className="hidden sm:inline">{label}</span>
            </button>
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 2: Create StepDestination**

Create `army/web/fe/src/components/setup/StepDestination.tsx`:

```tsx
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';

interface StepDestinationProps {
  value: 'user' | 'project' | null;
  onChange: (dest: 'user' | 'project') => void;
  onNext: () => void;
}

export function StepDestination({ value, onChange, onNext }: StepDestinationProps) {
  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Where should your manifest be saved?</h3>
      <p className="text-sm text-muted-foreground mb-6">
        This determines whether selections apply globally or to this project only.
      </p>

      <div className="flex flex-col gap-3 max-w-md">
        <button
          onClick={() => onChange('user')}
          className={cn(
            'text-left rounded-lg border-2 p-4 transition-colors',
            value === 'user'
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/30'
          )}
        >
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'size-4 rounded-full border-2 flex items-center justify-center',
                value === 'user' ? 'border-primary' : 'border-muted-foreground/40'
              )}
            >
              {value === 'user' && <span className="size-2 rounded-full bg-primary" />}
            </span>
            <span className="font-medium text-sm">User-level (global defaults)</span>
          </div>
          <p className="text-xs text-muted-foreground mt-1 ml-6">
            ~/.army/manifest.json — applies to all projects
          </p>
        </button>

        <button
          onClick={() => onChange('project')}
          className={cn(
            'text-left rounded-lg border-2 p-4 transition-colors',
            value === 'project'
              ? 'border-primary bg-primary/5'
              : 'border-border hover:border-primary/30'
          )}
        >
          <div className="flex items-center gap-2">
            <span
              className={cn(
                'size-4 rounded-full border-2 flex items-center justify-center',
                value === 'project' ? 'border-primary' : 'border-muted-foreground/40'
              )}
            >
              {value === 'project' && <span className="size-2 rounded-full bg-primary" />}
            </span>
            <span className="font-medium text-sm">Project-level (current project)</span>
          </div>
          <p className="text-xs text-muted-foreground mt-1 ml-6">
            &lt;cwd&gt;/.army/manifest.json — scoped to this project
          </p>
        </button>
      </div>

      <div className="flex justify-end mt-8">
        <Button onClick={onNext} disabled={!value}>
          Next →
        </Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create SetupWizard**

Create `army/web/fe/src/components/setup/SetupWizard.tsx`:

```tsx
import { useState, useCallback, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { Stepper } from './Stepper';
import { StepDestination } from './StepDestination';
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
        {/* Tech, Plugins, Skills, Confirm, Done steps wired in Tasks 8-10 */}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Create SetupPage**

Create `army/web/fe/src/pages/SetupPage.tsx`:

```tsx
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
```

- [ ] **Step 5: Add route to App.tsx**

In `army/web/fe/src/App.tsx`, add import:
```typescript
import { SetupPage } from './pages/SetupPage';
```

Add route after the `/` redirect route:
```tsx
<Route path="/setup" element={<SetupPage />} />
```

- [ ] **Step 6: Add Setup to Sidebar**

In `army/web/fe/src/components/layout/Sidebar.tsx`:

Add `Wand2` to the lucide imports:
```typescript
import { Package, ClipboardList, RefreshCw, Stethoscope, Sun, Moon, Wand2 } from 'lucide-react';
```

Add Setup as the first nav item:
```typescript
const navItems = [
  { path: '/setup', label: 'Setup', icon: Wand2 },
  { path: '/catalog', label: 'Catalog', icon: Package },
  { path: '/manifest', label: 'Manifest', icon: ClipboardList },
  { path: '/sync', label: 'Sync', icon: RefreshCw },
  { path: '/doctor', label: 'Doctor', icon: Stethoscope },
];
```

Also update the hardcoded version in the footer from `v0.3.0` to `v0.4.0`.

- [ ] **Step 7: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 8: Commit**

```bash
git add army/web/fe/src/pages/SetupPage.tsx army/web/fe/src/components/setup/ army/web/fe/src/App.tsx army/web/fe/src/components/layout/Sidebar.tsx
git commit -m "feat: add Setup page scaffold with stepper and destination step"
```

---

## Task 8: Frontend — Tech Stack step

**Files:**
- Create: `army/web/fe/src/components/setup/StepTechStack.tsx`
- Modify: `army/web/fe/src/components/setup/SetupWizard.tsx`

- [ ] **Step 1: Create StepTechStack**

Create `army/web/fe/src/components/setup/StepTechStack.tsx`:

```tsx
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import type { Catalog } from '@/lib/types';

interface StepTechStackProps {
  catalog: Catalog;
  selectedTech: string[];
  onChange: (tech: string[]) => void;
  onNext: () => void;
  onBack: () => void;
}

export function StepTechStack({ catalog, selectedTech, onChange, onNext, onBack }: StepTechStackProps) {
  const allTechNames = Object.keys(catalog.tech_profiles).sort();

  const toggle = (name: string) => {
    onChange(
      selectedTech.includes(name)
        ? selectedTech.filter((t) => t !== name)
        : [...selectedTech, name]
    );
  };

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Tech Stack</h3>
      <p className="text-sm text-muted-foreground mb-6">
        Select technologies used in this project to get tailored recommendations.
      </p>

      <div className="flex flex-wrap gap-2">
        {allTechNames.map((name) => {
          const isSelected = selectedTech.includes(name);
          return (
            <button
              key={name}
              onClick={() => toggle(name)}
              className={cn(
                'px-3 py-1.5 rounded-md text-xs font-medium border transition-colors',
                isSelected
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border text-muted-foreground hover:border-primary/30'
              )}
            >
              {isSelected ? '✓ ' : ''}{name}
            </button>
          );
        })}
      </div>

      <div className="flex justify-between mt-8">
        <Button variant="outline" onClick={onBack}>← Back</Button>
        <Button onClick={onNext}>Next →</Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Wire into SetupWizard**

In `army/web/fe/src/components/setup/SetupWizard.tsx`, add import:
```typescript
import { StepTechStack } from './StepTechStack';
```

Add rendering inside the `min-h-[400px]` div, after the Destination block:
```tsx
{activeStepName === 'Tech' && (
  <StepTechStack
    catalog={catalog}
    selectedTech={state.selectedTech}
    onChange={(tech) => setState((s) => ({ ...s, selectedTech: tech }))}
    onNext={next}
    onBack={back}
  />
)}
```

- [ ] **Step 3: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 4: Commit**

```bash
git add army/web/fe/src/components/setup/StepTechStack.tsx army/web/fe/src/components/setup/SetupWizard.tsx
git commit -m "feat: add tech stack step to setup wizard"
```

---

## Task 9: Frontend — SelectableList + Plugins/Skills steps

**Files:**
- Create: `army/web/fe/src/components/setup/SelectableList.tsx`
- Create: `army/web/fe/src/components/setup/StepPlugins.tsx`
- Create: `army/web/fe/src/components/setup/StepSkills.tsx`
- Modify: `army/web/fe/src/components/setup/SetupWizard.tsx`

- [ ] **Step 1: Create SelectableList**

Create `army/web/fe/src/components/setup/SelectableList.tsx`:

```tsx
import { useState, useMemo } from 'react';
import { Search } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface SelectableItem {
  name: string;
  description: string;
  source: string;
  recommended: boolean;
}

interface SelectableListProps {
  items: SelectableItem[];
  selected: string[];
  onChange: (selected: string[]) => void;
}

export function SelectableList({ items, selected, onChange }: SelectableListProps) {
  const [search, setSearch] = useState('');

  const filtered = useMemo(() => {
    if (!search) return items;
    const lower = search.toLowerCase();
    return items.filter(
      (item) =>
        item.name.toLowerCase().includes(lower) ||
        item.description.toLowerCase().includes(lower)
    );
  }, [items, search]);

  const toggle = (name: string) => {
    onChange(
      selected.includes(name)
        ? selected.filter((n) => n !== name)
        : [...selected, name]
    );
  };

  const selectAll = () => onChange(items.map((i) => i.name));
  const clearAll = () => onChange([]);

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full h-8 pl-8 pr-3 text-xs rounded-md border border-border bg-card focus:outline-none focus:ring-2 focus:ring-ring/50"
          />
        </div>
        <button onClick={selectAll} className="text-[11px] text-muted-foreground hover:text-foreground px-2 py-1 border border-border rounded-md">
          Select All
        </button>
        <button onClick={clearAll} className="text-[11px] text-muted-foreground hover:text-foreground px-2 py-1 border border-border rounded-md">
          Clear
        </button>
      </div>
      <div className="flex flex-col gap-1.5 max-h-[360px] overflow-y-auto">
        {filtered.map((item) => {
          const isSelected = selected.includes(item.name);
          return (
            <button
              key={item.name}
              onClick={() => toggle(item.name)}
              className={cn(
                'flex items-center gap-3 px-3 py-2.5 rounded-lg border text-left transition-colors',
                isSelected
                  ? 'border-primary bg-primary/5'
                  : 'border-border hover:border-primary/30'
              )}
            >
              <span className={cn(
                'size-4 rounded border flex items-center justify-center shrink-0 text-[10px]',
                isSelected
                  ? 'bg-primary border-primary text-primary-foreground'
                  : 'border-muted-foreground/40'
              )}>
                {isSelected && '✓'}
              </span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-1.5">
                  <span className="font-mono text-sm font-semibold truncate">{item.name}</span>
                  {item.recommended && (
                    <span className="text-primary text-xs">★</span>
                  )}
                </div>
                <p className="text-xs text-muted-foreground truncate">{item.description}</p>
              </div>
              <span className="font-mono text-[10px] text-muted-foreground/50 shrink-0">{item.source}</span>
            </button>
          );
        })}
      </div>
      <p className="text-xs text-muted-foreground">{selected.length} selected</p>
    </div>
  );
}
```

- [ ] **Step 2: Create StepPlugins**

Create `army/web/fe/src/components/setup/StepPlugins.tsx`:

```tsx
import { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { SelectableList, type SelectableItem } from './SelectableList';
import type { Catalog } from '@/lib/types';

interface StepPluginsProps {
  catalog: Catalog;
  selectedTech: string[];
  selectedPlugins: string[];
  onChange: (plugins: string[]) => void;
  onNext: () => void;
  onBack: () => void;
}

export function StepPlugins({ catalog, selectedTech, selectedPlugins, onChange, onNext, onBack }: StepPluginsProps) {
  const recommended = useMemo(() => {
    const names = new Set<string>();
    for (const tech of selectedTech) {
      const profile = catalog.tech_profiles[tech];
      if (profile) {
        for (const p of profile.plugins) names.add(p);
      }
    }
    return names;
  }, [catalog, selectedTech]);

  const items: SelectableItem[] = useMemo(
    () =>
      catalog.plugins.map((p) => ({
        name: p.name,
        description: p.description,
        source: p.marketplace,
        recommended: recommended.has(p.name),
      })),
    [catalog.plugins, recommended]
  );

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Plugins</h3>
      <p className="text-sm text-muted-foreground mb-4">
        Select plugins to include. ★ = recommended for your tech stack.
      </p>
      <SelectableList items={items} selected={selectedPlugins} onChange={onChange} />
      <div className="flex justify-between mt-6">
        <Button variant="outline" onClick={onBack}>← Back</Button>
        <Button onClick={onNext}>Next →</Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create StepSkills**

Create `army/web/fe/src/components/setup/StepSkills.tsx`:

```tsx
import { useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { SelectableList, type SelectableItem } from './SelectableList';
import type { Catalog } from '@/lib/types';

interface StepSkillsProps {
  catalog: Catalog;
  selectedTech: string[];
  selectedSkills: string[];
  onChange: (skills: string[]) => void;
  onNext: () => void;
  onBack: () => void;
}

export function StepSkills({ catalog, selectedTech, selectedSkills, onChange, onNext, onBack }: StepSkillsProps) {
  const recommended = useMemo(() => {
    const names = new Set<string>();
    for (const tech of selectedTech) {
      const profile = catalog.tech_profiles[tech];
      if (profile) {
        for (const s of profile.skills) names.add(s);
      }
    }
    return names;
  }, [catalog, selectedTech]);

  const items: SelectableItem[] = useMemo(
    () =>
      catalog.skills.map((s) => ({
        name: s.name,
        description: s.description,
        source: s.source,
        recommended: recommended.has(s.name),
      })),
    [catalog.skills, recommended]
  );

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Skills</h3>
      <p className="text-sm text-muted-foreground mb-4">
        Select skills to include. ★ = recommended for your tech stack.
      </p>
      <SelectableList items={items} selected={selectedSkills} onChange={onChange} />
      <div className="flex justify-between mt-6">
        <Button variant="outline" onClick={onBack}>← Back</Button>
        <Button onClick={onNext}>Next →</Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Wire Steps 3-4 into SetupWizard**

In `army/web/fe/src/components/setup/SetupWizard.tsx`, add imports:
```typescript
import { StepPlugins } from './StepPlugins';
import { StepSkills } from './StepSkills';
```

Add rendering blocks inside the `min-h-[400px]` div:
```tsx
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
```

- [ ] **Step 5: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 6: Commit**

```bash
git add army/web/fe/src/components/setup/SelectableList.tsx army/web/fe/src/components/setup/StepPlugins.tsx army/web/fe/src/components/setup/StepSkills.tsx army/web/fe/src/components/setup/SetupWizard.tsx
git commit -m "feat: add selectable list, plugins and skills steps to wizard"
```

---

## Task 10: Frontend — Confirm + Done steps

**Files:**
- Create: `army/web/fe/src/components/setup/StepConfirm.tsx`
- Create: `army/web/fe/src/components/setup/StepDone.tsx`
- Modify: `army/web/fe/src/components/setup/SetupWizard.tsx`

- [ ] **Step 1: Create StepConfirm**

Create `army/web/fe/src/components/setup/StepConfirm.tsx`:

```tsx
import { Button } from '@/components/ui/button';
import { Loader2 } from 'lucide-react';

interface StepConfirmProps {
  destination: 'user' | 'project';
  selectedPlugins: string[];
  selectedSkills: string[];
  onSave: () => void;
  onBack: () => void;
  isSaving: boolean;
  error: string | null;
}

export function StepConfirm({
  destination,
  selectedPlugins,
  selectedSkills,
  onSave,
  onBack,
  isSaving,
  error,
}: StepConfirmProps) {
  const destLabel = destination === 'user'
    ? '~/.army/manifest.json'
    : '<cwd>/.army/manifest.json';

  return (
    <div>
      <h3 className="text-lg font-semibold mb-1">Review Your Selections</h3>
      <p className="text-sm text-muted-foreground mb-6">Confirm before saving.</p>

      <div className="flex flex-col gap-4 max-w-md">
        <div className="rounded-lg border border-border bg-card p-4">
          <p className="text-[11px] text-muted-foreground uppercase tracking-wider mb-1.5">Destination</p>
          <p className="text-sm">
            {destination === 'user' ? 'User-level' : 'Project-level'}{' '}
            <span className="font-mono text-xs text-muted-foreground">{destLabel}</span>
          </p>
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <p className="text-[11px] text-muted-foreground uppercase tracking-wider mb-1.5">
            Plugins ({selectedPlugins.length})
          </p>
          <div className="flex flex-wrap gap-1.5">
            {selectedPlugins.length === 0 ? (
              <span className="text-xs text-muted-foreground">None selected</span>
            ) : (
              selectedPlugins.map((name) => (
                <span key={name} className="px-2 py-0.5 rounded text-xs bg-primary/10 text-primary">{name}</span>
              ))
            )}
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-4">
          <p className="text-[11px] text-muted-foreground uppercase tracking-wider mb-1.5">
            Skills ({selectedSkills.length})
          </p>
          <div className="flex flex-wrap gap-1.5">
            {selectedSkills.length === 0 ? (
              <span className="text-xs text-muted-foreground">None selected</span>
            ) : (
              selectedSkills.map((name) => (
                <span key={name} className="px-2 py-0.5 rounded text-xs bg-primary/10 text-primary">{name}</span>
              ))
            )}
          </div>
        </div>
      </div>

      {error && <p className="text-xs text-red-500 mt-3">{error}</p>}

      <div className="flex justify-between mt-8">
        <Button variant="outline" onClick={onBack} disabled={isSaving}>← Back</Button>
        <Button onClick={onSave} disabled={isSaving}>
          {isSaving && <Loader2 className="size-3.5 animate-spin" />}
          {isSaving ? 'Saving...' : 'Save Manifest ✓'}
        </Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create StepDone**

Create `army/web/fe/src/components/setup/StepDone.tsx`:

```tsx
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
        <Button onClick={() => navigate('/sync?autostart=true')}>Run Sync Now →</Button>
        <Button variant="outline" onClick={() => navigate('/catalog')}>Back to Catalog</Button>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Wire Confirm + Done into SetupWizard**

In `army/web/fe/src/components/setup/SetupWizard.tsx`, add imports:
```typescript
import { StepConfirm } from './StepConfirm';
import { StepDone } from './StepDone';
```

Add rendering blocks inside the `min-h-[400px]` div:
```tsx
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
```

- [ ] **Step 4: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 5: Commit**

```bash
git add army/web/fe/src/components/setup/StepConfirm.tsx army/web/fe/src/components/setup/StepDone.tsx army/web/fe/src/components/setup/SetupWizard.tsx
git commit -m "feat: add confirm and done steps to setup wizard"
```

---

## Task 11: Frontend — SyncPage autostart support

**Files:**
- Modify: `army/web/fe/src/pages/SyncPage.tsx`

- [ ] **Step 1: Add autostart query param support**

In `army/web/fe/src/pages/SyncPage.tsx`, add imports:
```typescript
import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
```

Inside SyncPage, after the hook calls, add:
```typescript
const [searchParams, setSearchParams] = useSearchParams();
const autostart = searchParams.get('autostart') === 'true';

useEffect(() => {
  if (autostart && !isRunning) {
    startSync(destination);
    setSearchParams({}, { replace: true });
  }
}, [autostart]); // eslint-disable-line react-hooks/exhaustive-deps
```

- [ ] **Step 2: Build frontend**

Run: `cd army/web/fe && npm run build`
Expected: builds cleanly

- [ ] **Step 3: Commit**

```bash
git add army/web/fe/src/pages/SyncPage.tsx
git commit -m "feat: add autostart query param support to SyncPage"
```

---

## Task 12: Version bump + final build + test

**Files:**
- Modify: `VERSION`

- [ ] **Step 1: Bump version**

Update `VERSION` to `0.4.0`.

- [ ] **Step 2: Full Go build + test**

Run: `cd army && make build && make test`
Expected: all pass

- [ ] **Step 3: Full web build**

Run: `cd army/web/be && npm run build && cd ../fe && npm run build`
Expected: both build cleanly

- [ ] **Step 4: Commit**

```bash
git add VERSION
git commit -m "chore: bump version to 0.4.0"
```

---

## Verification Checklist

After all tasks are complete:

1. `cd army && make build && make test` — all pass
2. `cd army/web/be && npm run build` — builds cleanly
3. `cd army/web/fe && npm run build` — builds cleanly
4. Start `army serve`, open web UI
5. **Fetch Catalog**: Go to Catalog page → click "Fetch Latest" → spinner → success message → catalog refreshes
6. **Setup Wizard**: Navigate to `/setup` → walk through all steps → save → "Run Sync Now" triggers sync
7. **Setup re-entry**: Re-open `/setup` → select same destination → verify existing selections are pre-populated
