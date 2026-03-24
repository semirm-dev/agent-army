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

### Go CLI Change

**File:** `army/cli/fetch_catalog.go`

The `fetch-catalog` command currently outputs plain text (`fmt.Println`). Add `--json` support: when `globalFlags.JSON` is set, output `{"path":"...","updated":true}` on success or `{"error":"..."}` on failure. This is required for the web backend to parse the response.

### Backend

**New endpoint:** `POST /api/catalog/fetch` in `army-catalog.controller.ts`

- Calls `ArmyService.exec(['fetch-catalog'])` — the `exec` method already appends `--json` and parses the response
- On success, clears the in-memory catalog cache (same as existing `refresh` logic)
- Returns the updated catalog in the response (re-fetches via `GET /api/catalog` internally)
- On failure, returns error message from CLI stderr

### Frontend

**File:** `army/web/fe/src/pages/CatalogPage.tsx`

- Add a "Fetch Latest" button with a `RefreshCw` (lucide) icon in the Catalog page header, next to the search bar
- On click: `POST /api/catalog/fetch` via a `useMutation` hook
- Loading state: spinner on the button, button disabled (prevents double-click)
- On success: invalidate the react-query `catalog` cache (triggers automatic refetch), show inline success toast ("Catalog updated")
- On error: show inline error message ("Fetch failed: ...")

**New API function:** `fetchCatalog()` in `army/web/fe/src/api/catalog.ts`

### Files to Modify

| File | Change |
|------|--------|
| `army/cli/fetch_catalog.go` | Add `--json` output support |
| `army/web/be/src/army/army-catalog.controller.ts` | Add `POST /api/catalog/fetch` endpoint |
| `army/web/fe/src/api/catalog.ts` | Add `fetchCatalog()` API function |
| `army/web/fe/src/pages/CatalogPage.tsx` | Add fetch button with loading/error states |

---

## Feature 2: Setup Wizard

### What It Does

A new `/setup` page with a horizontal stepper wizard that mirrors the TUI's `army setup` flow. Guides the user through destination selection, tech detection, plugin/skill selection, and saves a manifest.

### Wizard Steps

1. **Destination** — two radio-style cards: "User-level (global defaults)" and "Project-level (current project)". Each shows its manifest path. After selecting a destination, if a manifest already exists at that path, load it and pre-populate Steps 3-4 with its selections. If no manifest exists, start fresh.
2. **Tech Stack** (project-level only, skipped for user-level) — auto-detected technologies shown as toggleable chips. Fetched from `GET /api/detect`. User can toggle on/off to adjust recommendations.
3. **Plugins** — multi-select list of all plugins from catalog. Search input at top. Select All / Clear buttons. Recommended plugins (based on selected tech profiles) pre-checked with a ★ badge. Items from an existing manifest are also pre-checked. Each row shows: checkbox, name, description, marketplace source.
4. **Skills** — same pattern as plugins. Each row shows: checkbox, name, description, repo source.
5. **Confirm** — read-only summary showing destination path, selected plugins (count + chip list), selected skills (count + chip list). "Save Manifest" button.
6. **Done** — success screen with checkmark. "Run Sync Now" button (navigates to `/sync` and auto-triggers sync) and "Back to Catalog" button.

### Layout

Horizontal stepper at the top with numbered circles and step labels. Content area below changes per step. Back/Next buttons at the bottom of each step. Completed steps show checkmarks and are clickable to navigate back. Future steps are disabled.

### Go CLI Change

**New command:** `army write-manifest --destination <user|project>` in `army/cli/write_manifest.go`

Follows the existing flat hyphenated naming convention (like `fetch-catalog`).

- Reads a full manifest JSON from **stdin**
- Writes it atomically to the destination path (temp-file + rename, reusing `manifest.Save()`)
- With `--json`: outputs `{"path":"..."}` on success
- Input contract: stdin is a JSON object matching the `Manifest` type (`{ version, plugins, skills }`)
- The top-level `destination` flag determines the file path; the `destination` field on each plugin/skill entry is set to match
- Registered on the root command in `root.go` alongside other commands

### Backend

**New endpoint:** `POST /api/manifest` in `army-manifest.controller.ts`

- Accepts body: `{ destination: "user" | "project", plugins: Array<{ name, marketplace, tags }>, skills: Array<{ name, source, tags }> }`
- The top-level `destination` is stamped as the `destination` field on every plugin and skill entry in the manifest
- Builds the manifest JSON and pipes it to `army write-manifest --destination <dest>` via stdin. Uses Node's `execFile` with the `input` option (which writes to the child's stdin without needing `spawn`): `execFileAsync(bin, args, { input: jsonString })`
- Returns `{ path: string }` on success
- NestJS route: bare `@Post()` on the existing `ArmyManifestController` (no sub-path, alongside existing `@Post('plugin')` and `@Post('skill')`)

**Reused endpoints:**
- `GET /api/catalog` — provides plugin list, skill list, and tech profiles
- `GET /api/detect` — returns detected tech names for the current working directory
- `GET /api/manifest?scope=<user|project>` — loads existing manifest for pre-population

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
| `army/web/fe/src/api/detect.ts` | `getDetect()` API function for `GET /api/detect`. Returns `{ tech: string[] }` matching the backend `DetectDto` |

**Modified files:**

| File | Change |
|------|--------|
| `army/web/fe/src/App.tsx` | Add `/setup` route |
| `army/web/fe/src/components/layout/Sidebar.tsx` | Add "Setup" nav item (top position, `Wand2` icon from lucide) |
| `army/web/fe/src/api/manifest.ts` | Add `saveManifest()` function for `POST /api/manifest` |
| `army/web/fe/src/pages/SyncPage.tsx` | Read `autostart` query param; auto-trigger sync via `useEffect` when present |
| `army/web/be/src/army/army-manifest.controller.ts` | Add `POST /api/manifest` endpoint |
| `army/web/be/src/army/army.service.ts` | Add `execWithInput()` method using `execFile` with `input` option; add `saveManifest()` using it |
| `army/web/be/src/army/dto/manifest.dto.ts` | Add DTO for save-manifest request body |

**New Go files:**

| File | Purpose |
|------|---------|
| `army/cli/write_manifest.go` | `army write-manifest` command: reads manifest JSON from stdin, writes atomically |

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
2. Step 1 (after destination selected): fetch existing manifest via `useQuery('manifest', () => getManifest(destination), { enabled: !!destination })` — if it exists, pre-populate `selectedPlugins` and `selectedSkills`
3. Step 2 (if project-level): fetch tech via `useQuery('detect', getDetect, { enabled: destination === 'project' })`
4. Steps 3-4: derive recommended items by cross-referencing `selectedTech` with `catalog.tech_profiles[tech].plugins` and `.skills`
5. Step 5 save: `useMutation` calling `POST /api/manifest` with selections. On success, invalidate the react-query `manifest` cache.
6. Step 6 sync: navigate to `/sync?autostart=true`

### Recommendation Logic (Frontend)

For each selected tech profile, collect its recommended plugins/skills from the catalog's `tech_profiles` map. Union all recommendations. Mark matching items as recommended and pre-check them when the user first enters Steps 3/4.

```
selectedTech = ["go", "react"]
recommended_plugins = union(tech_profiles.go.plugins, tech_profiles.react.plugins)
recommended_skills = union(tech_profiles.go.skills, tech_profiles.react.skills)
```

---

## Verification

1. **Go build + test:** `cd army && make build && make test` — must pass (includes new `write-manifest` command)
2. **Web build:** `cd army/web/be && npm run build` and `cd army/web/fe && npm run build` — both must succeed
3. **Manual test — Fetch Catalog:**
   - Start `army serve`, open web UI
   - Go to Catalog page, click "Fetch Latest"
   - Verify spinner, success message, catalog data refreshes
   - Click again immediately — verify button is disabled during fetch
4. **Manual test — Setup Wizard:**
   - Navigate to `/setup`
   - Select user-level → verify tech step is skipped → select plugins/skills → confirm → save
   - Verify manifest written to `~/.army/manifest.json`
   - Re-enter setup → select user-level again → verify previous selections are pre-checked
   - Go back, select project-level → verify tech detection runs → recommendations appear → save
   - Verify manifest written to `<cwd>/.army/manifest.json`
   - Click "Run Sync Now" → verify navigation to sync page and sync auto-starts
5. **Go tests:** `cd army && make test` — all existing tests pass
