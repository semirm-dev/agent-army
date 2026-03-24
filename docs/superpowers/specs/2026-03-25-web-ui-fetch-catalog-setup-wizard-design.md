# Web UI: Fetch Catalog & Setup Wizard

**Date:** 2026-03-25
**Status:** Draft

## Context

The web UI currently has Catalog, Manifest, Sync, and Doctor pages. Two CLI features are missing from the web experience:

1. **Fetch Catalog** — the CLI's `army fetch-catalog` downloads the latest catalog from GitHub, but the web UI has no way to trigger this (the existing `POST /api/catalog/refresh` only clears the in-memory cache).
2. **Setup Wizard** — the CLI's `army setup` launches a TUI wizard for first-time configuration (destination → tech detection → plugin/skill selection → confirm). The web UI has no equivalent guided setup flow.

Adding these brings the web UI to feature parity with the CLI.

---

## Feature 1: Fetch Catalog

### What It Does

A button on the Catalog page that triggers a full remote fetch from GitHub, updates `~/.army/catalog.json`, then refreshes the page data.

### Backend

**New endpoint:** `POST /api/catalog/fetch` in `army-catalog.controller.ts`

- Calls `ArmyService.exec(['fetch-catalog', '--json'])` via `execFile` (not `exec`, to avoid shell injection)
- On success, clears the in-memory catalog cache (same as existing `refresh` logic)
- Returns the updated catalog in the response
- On failure, returns error message from CLI stderr

### Frontend

**File:** `army/web/fe/src/pages/CatalogPage.tsx`

- Add a "Fetch Latest" button with a refresh icon in the Catalog page header, next to the search bar
- On click: `POST /api/catalog/fetch` via a `useMutation` hook
- Loading state: spinner on the button, button disabled
- On success: invalidate the react-query `catalog` cache (triggers automatic refetch), show inline success toast ("Catalog updated")
- On error: show inline error message ("Fetch failed: ...")

**New API function:** `fetchCatalog()` in `army/web/fe/src/api/catalog.ts`

### Files to Modify

| File | Change |
|------|--------|
| `army/web/be/src/army/army-catalog.controller.ts` | Add `POST /api/catalog/fetch` endpoint |
| `army/web/be/src/army/army.service.ts` | Add `fetchCatalog()` method (if not reusing generic `exec`) |
| `army/web/fe/src/api/catalog.ts` | Add `fetchCatalog()` API function |
| `army/web/fe/src/pages/CatalogPage.tsx` | Add fetch button with loading/error states |

---

## Feature 2: Setup Wizard

### What It Does

A new `/setup` page with a horizontal stepper wizard that mirrors the TUI's `army setup` flow. Guides the user through destination selection, tech detection, plugin/skill selection, and saves a manifest.

### Wizard Steps

1. **Destination** — two radio-style cards: "User-level (global defaults)" and "Project-level (current project)". Each shows its manifest path.
2. **Tech Stack** (project-level only, skipped for user-level) — auto-detected technologies shown as toggleable chips. Fetched from `GET /api/detect`. User can toggle on/off to adjust recommendations.
3. **Plugins** — multi-select list of all plugins from catalog. Search input at top. Select All / Clear buttons. Recommended plugins (based on selected tech profiles) pre-checked with a ★ badge. Each row shows: checkbox, name, description, marketplace source.
4. **Skills** — same pattern as plugins. Each row shows: checkbox, name, description, repo source.
5. **Confirm** — read-only summary showing destination path, selected plugins (count + chip list), selected skills (count + chip list). "Save Manifest" button.
6. **Done** — success screen with checkmark. "Run Sync Now" button (navigates to `/sync` and triggers sync) and "Back to Catalog" button.

### Layout

Horizontal stepper at the top with numbered circles and step labels. Content area below changes per step. Back/Next buttons at the bottom of each step. Completed steps show checkmarks and are clickable to navigate back. Future steps are disabled.

### Backend

**New endpoint:** `POST /api/manifest` in `army-manifest.controller.ts`

- Accepts body: `{ destination: "user" | "project", plugins: Array<{ name, marketplace, tags }>, skills: Array<{ name, source, tags }> }`
- Builds a full manifest JSON and writes it atomically via the Go CLI
- Implementation: uses `execFile` (via `ArmyService`) to call a CLI sub-command that accepts manifest data and writes it atomically (temp-file + rename pattern)
- Returns `{ path: string }` on success

**Reused endpoints:**
- `GET /api/catalog` — provides plugin list, skill list, and tech profiles
- `GET /api/detect` — returns detected tech names for the current working directory

### Frontend

**New files:**

| File | Purpose |
|------|---------|
| `army/web/fe/src/pages/SetupPage.tsx` | Route page, renders SetupWizard |
| `army/web/fe/src/components/setup/SetupWizard.tsx` | Main wizard component: stepper + step routing + state |
| `army/web/fe/src/components/setup/StepDestination.tsx` | Step 1: destination radio cards |
| `army/web/fe/src/components/setup/StepTechStack.tsx` | Step 2: tech detection chips |
| `army/web/fe/src/components/setup/StepPlugins.tsx` | Step 3: plugin multi-select |
| `army/web/fe/src/components/setup/StepSkills.tsx` | Step 4: skill multi-select |
| `army/web/fe/src/components/setup/StepConfirm.tsx` | Step 5: review + save |
| `army/web/fe/src/components/setup/StepDone.tsx` | Step 6: success + sync offer |
| `army/web/fe/src/components/setup/Stepper.tsx` | Horizontal step indicator component |
| `army/web/fe/src/components/setup/SelectableList.tsx` | Reusable multi-select list with search (shared by Steps 3 & 4) |

**Modified files:**

| File | Change |
|------|--------|
| `army/web/fe/src/App.tsx` | Add `/setup` route |
| `army/web/fe/src/components/layout/Sidebar.tsx` | Add "Setup" nav item (top position) |
| `army/web/fe/src/api/manifest.ts` | Add `saveManifest()` function for `POST /api/manifest` |
| `army/web/be/src/army/army-manifest.controller.ts` | Add `POST /api/manifest` endpoint |
| `army/web/be/src/army/army.service.ts` | Add method to save full manifest |
| `army/web/be/src/army/dto/manifest.dto.ts` | Add DTO for save-manifest request body |

### State Management

Local React state in `SetupWizard.tsx` — no global store needed:

```typescript
interface WizardState {
  currentStep: number;          // 0-5
  destination: "user" | "project" | null;
  selectedTech: string[];       // tech profile names
  selectedPlugins: string[];    // plugin names
  selectedSkills: string[];     // skill names
}
```

### Data Flow

1. On mount: fetch catalog via `useQuery('catalog', getCatalog)`
2. Step 2 (if project-level): fetch tech via `useQuery('detect', getDetect, { enabled: destination === 'project' })`
3. Steps 3-4: derive recommended items by cross-referencing `selectedTech` with `catalog.tech_profiles[tech].plugins` and `.skills`
4. Step 5 save: `useMutation` calling `POST /api/manifest` with selections
5. Step 6 sync: navigate to `/sync?autostart=true` or call sync API directly

### Recommendation Logic (Frontend)

For each selected tech profile, collect its recommended plugins/skills from the catalog's `tech_profiles` map. Union all recommendations. Mark matching items as recommended and pre-check them when the user first enters Steps 3/4.

```
selectedTech = ["go", "react"]
recommended_plugins = union(tech_profiles.go.plugins, tech_profiles.react.plugins)
recommended_skills = union(tech_profiles.go.skills, tech_profiles.react.skills)
```

---

## Verification

1. **Build:** `cd army/web/be && npm run build` and `cd army/web/fe && npm run build` — both must succeed
2. **Manual test — Fetch Catalog:**
   - Start `army serve`, open web UI
   - Go to Catalog page, click "Fetch Latest"
   - Verify spinner, success message, catalog data refreshes
3. **Manual test — Setup Wizard:**
   - Navigate to `/setup`
   - Select user-level → verify tech step is skipped → select plugins/skills → confirm → save
   - Verify manifest written to `~/.army/manifest.json`
   - Go back, select project-level → verify tech detection runs → recommendations appear → save
   - Verify manifest written to `<cwd>/.army/manifest.json`
   - Click "Run Sync Now" → verify navigation to sync page
4. **Go tests:** `cd army && make test` — existing tests should still pass (no Go changes expected beyond possible new CLI sub-command)
