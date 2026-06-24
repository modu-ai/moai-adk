---
id: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
title: "Plan — Retire statusline preset system + remove web-console statusline panel"
version: "0.2.0"
status: draft
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "plan"
lifecycle: spec-anchored
tags: "plan, statusline, preset, mode, retire"
tier: M
---

# Plan — SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001

## §A. Context

This plan operationalizes the retirement described in `spec.md`. The work is a
multi-package removal (Go code) + template edit + 4-locale docs edit + test
refactor. Cycle type: **ddd** (this is predominantly existing-code refactoring
with behavior preservation — characterization tests guard segment rendering
byte-parity; the removal paths themselves are mechanical).

Two user decisions are locked:
1. Full retirement (no `compact`/`minimal`/`full` named presets survive).
2. Entire web-console statusline panel removed (not just the preset dropdown).

### §A.1 Investigation confidence

All target file:line locations in `spec.md §A.2` were verified by direct
`grep`/`Read` during plan-phase. Refinements vs. the delegation prompt and
iter-1 audit findings (D1-D7, all re-verified against actual code):

- **R1**: `StatuslineConfig` struct lives in `pkg/models/config.go:197-202`,
  not `internal/config/types.go` (the latter embeds it as
  `models.StatuslineConfig` at L28 and L1096 — those embedding sites do not
  need editing; only the struct definition does).
- **R2**: `CanonicalSegments` (preset.go L8-14) is co-located with the
  to-be-removed `PresetToSegments` function. The variable is SSOT and MUST be
  preserved; only the function body L16-69 is removed.
- **D1 (iter-1)**: a lowercase `presetToSegments` wrapper exists at
  `internal/cli/update.go:2732-2738` (delegates to `statusline.PresetToSegments`).
  Used by `cli/statusline.go:56`. MUST be deleted alongside the capital symbol.
  Test collateral: `internal/cli/wizard_config_test.go:407-480` +
  `internal/cli/update_test.go:2721-2779` (`TestPresetToSegments`).
- **D2 (iter-1)**: `internal/web/root.templ:108` invokes
  `@fieldsetStatusline(view)` (generated mirror `root_templ.go:175`). Removing
  the fieldset component from `fieldsets.templ` without removing this caller
  breaks `templ generate`. Both MUST be removed in the same milestone (M3).
- **D3 (iter-1, internal contradiction resolved)**: the runtime `mode:` YAML
  surface was already removed by SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001
  (`cli/statusline.go:62-66` documents: mode fixed to `ModeDefault`). The
  residual mode preferences surface (preferences.go:45 + wizard + docs) is
  retired in lockstep. The `internal/statusline.StatuslineMode` type in
  `types.go`/`builder.go`/`renderer.go` is PRESERVED (Builder API). Decision:
  **retire mode preferences axis** (not preserve) — cleaner per user intent.
- **D4 (iter-1)**: 5 test files with compile-breaking references + 2 D1
  collateral test files assigned to owning milestones (see §B Known Issues).
- **D5 (iter-1)**: `internal/profile/sync.go:80-91 statuslineData` struct has
  a `Preset` field mirroring `models.StatuslineConfig`. Folded into M2
  preset-axis removal.
- **D6 (iter-1)**: `Mode*` i18n keys in `profile_setup_translations.go` (4
  locales) folded into M4 wizard translations cleanup.
- **D7 (iter-1)**: docs `mode:` line at L211 (4 locales) folded into M5
  alongside `preset:` line at L213.

No other drift between prompt/audit and codebase was found.

### §A.2 Build/test commands (canonical)

```bash
# Build (must remain green after every milestone)
go build ./...

# Test (full suite — HARD per CLAUDE.local.md §6)
go test ./...

# Race + vet for the affected packages
go test -race ./internal/statusline/... ./internal/profile/... ./internal/web/... ./internal/cli/...
go vet ./...

# Regenerate web templates after editing fieldsets.templ
# (run from repo root; the project uses templ generate)
templ generate ./internal/web/...
go build ./internal/web/...

# 4-locale docs parity check
bash scripts/docs-i18n-check.sh

# Template neutrality CI guard (only if internal/template/templates/ touched)
go test ./internal/template/... -run TestTemplateNeutralityAudit
```

---

## §B. Known Issues (pre-run)

| ID | Issue | Resolution path |
|----|-------|-----------------|
| K1 | `internal/web/coverage_test.go:117` references `StatuslinePreset: "weird"` — will fail to compile once the field is removed | M3 — refactor or delete the test case |
| K2 | `internal/web/statusline_conditional_test.go` (whole file) tests the preset-based conditional rendering — entire file becomes obsolete | M3 — delete the file (its assertions are about behavior we are removing) |
| K3 | `internal/web/statusline_empty_option_test.go` tests the preset `<select>` empty-option rendering — file becomes obsolete | M3 — delete the file |
| K4 | `internal/statusline/preset.go` file may end up containing only `CanonicalSegments` — filename honesty | M2 — optional rename to `segments.go`. Default: leave filename to minimize churn. |
| K5 | `internal/cli/profile_setup.go` Section 4 (Display) loses its preset widget — form group may need re-numbering or section re-title | M4 — verify wizard renders end-to-end; adjust section title/description if needed |
| K6 | `internal/profile/sync.go` L114-115 default-preset-to-`full` and L139-142 preset-expansion switch — both branches deleted; verify the `StatuslineSegments != nil` path still writes verbatim | M2 — characterization test confirms segments map is persisted unchanged |
| K7 | `docs-site/content/*/advanced/statusline.md:213` — only the `preset:` line is edited; verify the surrounding YAML example still parses | M5 — visual diff per locale, run docs build if available |
| K8 (D1) | `internal/cli/update.go:2732-2738` — lowercase `presetToSegments` wrapper delegates to capital `PresetToSegments`. MUST be deleted in M2 alongside the capital symbol. | M2 — delete wrapper; test collateral in `wizard_config_test.go:407-480` + `update_test.go:2721-2779` deleted/refactored |
| K9 (D2) | `internal/web/root.templ:108` — `@fieldsetStatusline(view)` caller. Removing the fieldset component WITHOUT removing this caller breaks `templ generate` (`root_templ.go:175` dangles). | M3 — remove caller from root.templ:108 in the SAME milestone as the fieldset component removal; regenerate root_templ.go |
| K10 (D3) | `internal/profile/preferences.go:45` — `StatuslineMode` field. Runtime `mode:` already removed by V3R5; residual prefs/wizard/docs mode surface retired in lockstep. | M2 (field) + M4 (wizard Select/helpers/translations) + M5 (docs mode: line) |
| K11 (D4) | 5 test files + 2 D1 collateral test files with compile-breaking references to removed symbols, unassigned in iter-0 plan. | M2: preferences_test.go, sync_test.go, statusline_segments_test.go, wizard_config_test.go, update_test.go. M4: profile_setup_test.go, profile_setup_translations_test.go. |
| K12 (D5) | `internal/profile/sync.go:80-91` — `statuslineData.Preset` struct field mirrors `models.StatuslineConfig`. Folded into M2 preset-axis removal. | M2 — remove `Preset` field from `statuslineData` struct |
| K13 (D6) | `internal/cli/profile_setup_translations.go` — `Mode*` i18n keys (StatuslineModeTitle/Desc, ModeDefault/Compact/Full/Verbose/Minimal, SummaryStatuslineMode, MigrationNoticeStatuslineMode) across 4 locales. | M4 — remove orphaned mode i18n keys alongside preset i18n keys |
| K14 (D7) | `docs-site/content/*/advanced/statusline.md:211` — `mode:` line survives alongside `preset:` at L213. | M5 — remove both `mode:` (L211) and `preset:` (L213) lines in all 4 locales |
| K15 (D3 Builder API) | `internal/statusline/types.go` — `StatuslineMode` type + constants + `NormalizeMode` are part of the Builder API and MUST be PRESERVED (`cli/statusline.go:62-66` feeds `ModeDefault`). | OUT OF SCOPE — do NOT touch `internal/statusline/types.go`/`builder.go`/`renderer.go` Mode symbols. Only the prefs/wizard/docs mode surface is retired. |

---

## Out of Scope (plan-phase)

The following are explicitly out of scope for this plan (cross-reference
`spec.md §E` for the full SPEC-level exclusions):

### 1. Out of Scope — implementation approach

- **DO NOT** introduce worktree isolation for this SPEC unless the user
  explicitly opts in at run-phase entry (per user policy 2026-05-17,
  L2/L3 worktree is opt-in). Default: main checkout + feature branch.
- **DO NOT** run a separate `templ generate` pass outside M3 — the web
  template regeneration is a single atomic step inside M3 (fieldset removal +
  caller removal happen together to avoid a broken intermediate state).
- **DO NOT** add new characterization tests for preset-expansion behavior —
  that behavior is being removed; only segments-verbatim + legacy-silent-ignore
  paths get new tests (AC-SPR-020).

### 2. Out of Scope — files NOT to touch

- `internal/statusline/types.go`, `builder.go`, `renderer.go` — the Builder
  API `StatuslineMode` type, constants, and `NormalizeMode` are PRESERVED
  (K15). Only the prefs/wizard/docs mode surface is retired.
- `internal/statusline/git.go`, `usage.go`, `task.go`, `memory.go`,
  `metrics.go`, `model_cache.go`, `version.go`, `theme.go`, `gradient.go` —
  segment rendering internals are untouched (zero behavior regression, spec.md
  §E.6).
- `.moai/config/sections/statusline.yaml` (local, non-template) — this is a
  user-owned runtime file, not the template. Only the template at
  `internal/template/templates/.moai/config/sections/statusline.yaml` is edited.

### 3. Out of Scope — non-goals

- **No new public API** — this is a removal SPEC (spec.md §C.3).
- **No migration command** — `moai statusline migrate` is explicitly excluded
  (spec.md §E.2). The loader's silent-ignore (REQ-SPR-021) is the only
  migration story.
- **No web UI replacement** for the removed statusline panel (spec.md §E.8).

---

## §C. Pre-flight Checks (before M1)

- [ ] `git status` clean (no uncommitted changes that would conflict)
- [ ] `go build ./...` green (baseline)
- [ ] `go test ./...` green (baseline — capture pre-retire test count + coverage for affected packages)
- [ ] Branch created (or worktree if L3 `--worktree` was used at plan init)
- [ ] All target file:line locations in `spec.md §A.2` re-verified (code may have shifted since plan-phase)

---

## §D. Constraints (carried from spec.md §C)

1. Zero segment-rendering behavior regression (byte-parity).
2. Legacy `preset:` YAML key silently ignored (no migration command).
3. No new public API.
4. Template neutrality preserved (`statusline.yaml` edit free of internal IDs).
5. 4-locale docs parity via `scripts/docs-i18n-check.sh`.
6. Test isolation via `t.TempDir()`.

---

## §E. Self-Verification (manager-develop §E.1-§E.7 alignment)

The run-phase agent is responsible for populating `progress.md §E.2/§E.3` with
verbatim command output. The plan-phase expectation:

- **E1 (AC PASS/FAIL matrix)**: all 25 MUST ACs PASS, 3 SHOULD ACs PASS or
  documented as deferred with rationale.
- **E2 (cross-platform build)**: `go build ./...` PASS on darwin/amd64 (dev
  host); `GOOS=linux GOARCH=amd64 go build ./...` PASS; `GOOS=windows
  GOARCH=amd64 go build ./...` PASS.
- **E3 (coverage)**: affected packages (`internal/statusline`,
  `internal/profile`, `internal/web`, `internal/cli`, `pkg/models`) MUST NOT
  drop below their pre-retire coverage baseline. Target: net zero or net
  positive delta (removal of dead-preset code paths may slightly raise
  coverage).
- **E4 (subagent-boundary grep)**: `grep -rn 'AskUserQuestion\|mcp__askuser'`
  on modified files returns 0 (no new violations).
- **E5 (lint)**: `golangci-lint run --timeout=2m` clean on modified packages.
- **E6 (push state)**: commits pushed to feature branch; PR opened (Tier M →
  manager-git handles per Tier-based routing).
- **E7 (spec-lint)**: `moai spec lint SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001`
  returns no MUST-FIX findings on the SPEC artifacts.

---

## §F. Milestones

### M1 — Research / baseline capture (orchestrator-direct or read-only)

**Scope**: re-verify all target locations; capture pre-retire test + coverage
baseline for the 5 affected packages; confirm no drift since plan-phase.

**Deliverables**:
- Baseline test output (pass count, fail count, coverage %) for
  `internal/statusline`, `internal/profile`, `internal/web`, `internal/cli`,
  `pkg/models`.
- Re-confirmation of every file:line in `spec.md §A.2`.
- Capture of 3 known-issue test files (K1/K2/K3) — their current pass status.

**Exit criteria**: baseline captured in `progress.md §E.2` (M1 evidence row);
all file:line locations still valid OR documented drift.

### M2 — Go code removal (cycle_type=ddd, manager-develop)

**Scope**: the mechanical removal across `pkg/models`, `internal/statusline`,
`internal/cli/statusline.go` + `internal/cli/update.go` (D1 wrapper),
`internal/profile` (including `StatuslineMode` field — D3). NO web changes yet
(those are M3) to keep the compile-error surface bounded.

**Ordered edits** (dependency-aware — struct fields first, then consumers):

1. `pkg/models/config.go:199` — delete `Preset` field from `StatuslineConfig`.
   (Mode field is already absent — V3R5 removed it.)
2. `internal/statusline/preset.go:16-69` — delete `PresetToSegments` function
   body. Preserve `CanonicalSegments` (L8-14). (Optional: rename file to
   `segments.go` per K4 — decide here.)
3. **D1** — `internal/cli/update.go:2732-2738` — delete the lowercase
   `presetToSegments` wrapper function (delegates to capital
   `PresetToSegments`). Both symbols MUST be gone so
   `grep -rn 'PresetToSegments\|presetToSegments' internal/ pkg/` returns 0.
4. `internal/cli/statusline.go:52-57` — replace preset fallback branch with
   segments-only logic:
   ```go
   if statuslineCfg != nil {
       segmentConfig = statuslineCfg.Segments  // nil-safe; all-enabled fallback in Builder
       themeName = statuslineCfg.Theme
   }
   ```
   (The `Mode: statusline.ModeDefault` assignment at L66 is PRESERVED — Builder API.)
5. `internal/cli/statusline.go:123-127, 146, 157` — remove `Preset` field from
   `statuslineFileConfig` and the raw YAML `preset:` binding.
6. **D3 (field)** — `internal/profile/preferences.go:45-46` — delete BOTH the
   `StatuslinePreset` field (L46) AND the `StatuslineMode` field (L45).
   `StatuslineSegments` (L47) and `StatuslineTheme` (L48) stay.
7. **D5** — `internal/profile/sync.go:80-91` — remove the `Preset` field from
   the `statuslineData` struct (mirrors `models.StatuslineConfig`).
8. `internal/profile/sync.go:114-142` — collapse to segments-only write logic
   per REQ-SPR-007/008/009:
   ```go
   // default segments when on-disk has none
   if current.Statusline.Segments == nil {
       current.Statusline.Segments = defaultStatuslineSegments()
   }
   // submitted segments win verbatim; otherwise preserve existing
   if prefs.StatuslineSegments != nil {
       current.Statusline.Segments = prefs.StatuslineSegments
   }
   // theme (unchanged path)
   if prefs.StatuslineTheme != "" {
       current.Statusline.Theme = prefs.StatuslineTheme
   }
   ```
9. **D4 (profile tests)** — retire/refactor 3 profile test files + 2 D1
   collateral CLI test files whose referenced symbols are now gone:
   - `internal/profile/preferences_test.go` — remove `StatuslinePreset:`
     references (L78, 105-106, 177, 458, 474-475).
   - `internal/profile/sync_test.go` — remove `StatuslinePreset:` prefs binding
     + `wrapper.Statusline.Preset` assertions (L197, 209, 229-230, 262-263,
     278, 299).
   - `internal/profile/statusline_segments_test.go` — the entire file tests
     preset-expansion (`TestSyncStatuslinePresetExpand`,
     `TestStatuslinePresetEffectiveReloadRoundTrip` L60-138). Delete the file
     (preset-expansion behavior is removed; no replacement test needed —
     segments-verbatim path is covered by the new M2 characterization test).
   - `internal/cli/wizard_config_test.go:407-480` — delete the
     `TestPresetToSegments_*` functions (L409-490) that test the lowercase
     wrapper.
   - `internal/cli/update_test.go:2721-2779` — delete `TestPresetToSegments`
     that exercises the deleted wrapper.
10. Compile-fix any remaining references surfaced by `go build ./...`.

**Deliverables**: `go build ./...` green; `go test ./internal/statusline/...
./internal/profile/... ./internal/cli/... ./pkg/models/...` green.

**Exit criteria**: Go packages compile; `grep -rn 'PresetToSegments\|presetToSegments' internal/ pkg/` returns 0; `grep -n 'StatuslinePreset\|StatuslineMode' internal/profile/preferences.go` returns 0; wizard test files (profile_setup_test.go, profile_setup_translations_test.go) may still fail — those are M4 scope.

### M3 — Web console removal (cycle_type=ddd, manager-develop)

**Scope**: remove the entire statusline fieldset from the web UI + retire the
3 known-issue web tests.

**Ordered edits**:

1. `internal/web/fieldsets.templ:120-153` — delete the entire statusline
   `<fieldset>` block (the `fieldsetStatusline` templ component, from
   `// fieldsetStatusline renders ...` comment + `templ fieldsetStatusline(view pageView) {`
   opening to its closing `}`).
2. **D2** — `internal/web/root.templ:108` — delete the
   `@fieldsetStatusline(view)` invocation line. This caller MUST be removed in
   the SAME milestone as the component deletion (step 1) or `templ generate`
   breaks (`root_templ.go:175` would reference the deleted component).
3. Regenerate: `templ generate ./internal/web/...` (produces updated
   `fieldsets_templ.go` AND `root_templ.go` — both must regenerate cleanly).
4. `internal/web/handlers.go:373` — delete `StatuslinePreset:` line from the
   prefs binding.
5. `internal/web/handlers.go:374-380` — delete the entire
   `if prefs.StatuslinePreset == "custom" { ... }` segment-binding branch.
6. `internal/web/handlers.go:29, 82` — delete `StatuslinePresets` field from
   the view model and the `statuslinePresetCanonical` assignment.
7. `internal/web/validate.go:43-44` — delete `statuslinePresetCanonical` var.
8. `internal/web/validate.go:134-136` — delete the `statusline_preset`
   validation rule block.
9. `internal/web/statusline_conditional_test.go` — **delete entire file** (K2).
10. `internal/web/statusline_empty_option_test.go` — **delete entire file** (K3).
11. `internal/web/coverage_test.go:117` — remove the `StatuslinePreset: "weird"`
    field from the test fixture (K1). The test case itself may stay if it
    targets other fields.
12. `internal/web/restyle_test.go:309` — remove the comment referencing
    `fieldsetStatusline` (cosmetic; verify the test itself does not depend on
    the removed component).
13. Compile-fix any remaining web-package references.

**Deliverables**: `go build ./internal/web/...` green; `go test ./internal/web/...`
green (with the 2 retired test files gone + coverage_test fixture updated).

**Exit criteria**: web console renders without the statusline fieldset; no
`StatuslinePreset` / `statuslinePreset` / `fieldsetStatusline` token anywhere
in `internal/web/`; `templ generate` exits 0.

### M4 — Wizard cleanup (cycle_type=ddd, manager-develop)

**Scope**: `internal/cli/profile_setup.go` — remove preset Select section +
mode Select section + their canonical lists/helpers + make segment section
unconditional. Also retire wizard translations (preset + mode i18n keys) and
2 D4 wizard test files.

**Ordered edits**:

1. `internal/cli/profile_setup.go:81-84` — delete
   `statuslinePresetCanonical` var.
2. **D3 (wizard helpers)** — `internal/cli/profile_setup.go:18-20, 27, 32-68`
   — delete the mode canonical list + helpers: `statuslineModeCanonical`,
   `defaultStatuslineMode`, `isCanonicalStatuslineMode`,
   `normalizeStatuslineModeRaw`, `normalizeStatuslineMode`. AND delete the
   preset helpers: `isCanonicalStatuslinePreset`, `normalizeStatuslinePreset`
   (L96-116).
3. `internal/cli/profile_setup.go:272` — delete the
   `normalizeStatuslinePreset(existingPrefs.StatuslinePreset)` call.
4. **D3 (wizard mode capture)** — `internal/cli/profile_setup.go:267, 270` —
   delete `rawStatuslineMode` capture + `statuslineMode := normalizeStatuslineMode(...)`
   call.
5. **D3 (wizard mode migration notice)** —
   `internal/cli/profile_setup.go:337-338` — delete the
   `MigrationNoticeStatuslineMode` print block.
6. `internal/cli/profile_setup.go:427-460` — delete the preset Select widget
   block (the `huh.NewSelect(...)` for preset). Keep the theme Select.
7. **D3 (wizard mode Select)** — `internal/cli/profile_setup.go:428-437` —
   delete the mode Select widget (`huh.NewSelect` with `StatuslineModeTitle`/
   `StatuslineModeDesc`/`ModeDefault`/`ModeFull` options,
   `Value(&statuslineMode)`). (Note: this range overlaps with the preset
   Select at L427-460; verify both are removed and the theme Select at the
   adjacent position survives.)
8. `internal/cli/profile_setup.go:486` — delete the
   `return statuslinePreset != "custom"` hide-condition on the segment
   section (segment section becomes always-shown).
9. `internal/cli/profile_setup.go:530-557` — make the segment-map build
   unconditional (remove the `if statuslinePreset == "custom" { ... }` wrapper;
   always build and emit segments).
10. `internal/cli/profile_setup.go:557` — remove `StatuslinePreset:` from the
    returned `ProfilePreferences` literal.
11. **D3 (wizard mode prefs literal)** —
    `internal/cli/profile_setup.go:556` — remove `StatuslineMode: statuslineMode,`
    from the returned `ProfilePreferences` literal.
12. **D3 (wizard mode summary)** — `internal/cli/profile_setup.go:620-622` —
    delete the `SummaryStatuslineMode` summary table row.
13. **D6 (wizard translations)** —
    `internal/cli/profile_setup_translations.go` — remove orphaned mode i18n
    keys across all 4 locales: `StatuslineModeTitle`, `StatuslineModeDesc`,
    `ModeDefault`, `ModeCompact`, `ModeFull`, `ModeVerbose`, `ModeMinimal`,
    `SummaryStatuslineMode`, `MigrationNoticeStatuslineMode`. AND remove
    orphaned preset i18n keys: `StatuslinePresetTitle`, `StatuslinePresetDesc`,
    `PresetFull`, `PresetCompact`, `PresetMinimal`, `PresetCustom`. Theme keys
    (`StatuslineThemeTitle`/`Desc` + canonical labels) stay (REQ-SPR-014).
14. **D4 (wizard tests)** — retire/refactor 2 wizard test files:
    - `internal/cli/profile_setup_test.go` — remove the mode i18n key checks
      (L38-59: StatuslineModeTitle/Desc, ModeDefault/Compact/Full/Verbose/Minimal)
      AND the `TestNormalizeStatuslinePreset` function (L65-87). Keep any
      theme/segment assertions that survive.
    - `internal/cli/profile_setup_translations_test.go` — remove the
      `MigrationNoticeStatuslineMode` assertions (L158-166) AND the
      `TestProfileSetupTranslations_PresetSegments` function's preset-key
      rows (L222-241: StatuslinePresetTitle/Desc, PresetFull/Compact/Minimal/Custom).
      Keep theme-key rows.
15. Compile-fix; run the existing wizard test (add one if absent —
    characterization test that the wizard produces a segment map without
    prompting for a preset or mode).

**Deliverables**: `go build ./internal/cli/...` green; `go test ./internal/cli/...`
green.

**Exit criteria**: wizard no longer references preset OR mode; segment section
always shown; theme Select preserved; no `StatuslinePreset`/`StatuslineMode`/
`statuslinePreset`/`statuslineMode`/`Mode*` token in `profile_setup.go` +
`profile_setup_translations.go` (except theme keys + Builder API references in
`internal/statusline/` which are preserved).

### M5 — Template + docs-site (orchestrator-direct or manager-docs)

**Scope**: template `statusline.yaml` edit + 4-locale docs edit.

**Ordered edits**:

1. `internal/template/templates/.moai/config/sections/statusline.yaml` —
   delete L2-6 (the 4-line preset comment block + the `preset: "full"` line).
   Leave `theme:` and `segments:` blocks intact. (Template has no `mode:` line —
   already absent; no mode edit needed here.)
2. **D7** — `docs-site/content/en/advanced/statusline.md:211` — remove the
   `mode: default  # default | full` line. AND L213 — remove the
   `preset: custom  # full | minimal | custom` line. Adjust surrounding YAML
   example for coherence (the statusline block becomes theme + segments only).
3. Repeat for `ko`, `ja`, `zh` — identical conceptual edit (both `mode:` at
   L211 and `preset:` at L213 removed in all 4 locales).
4. Run `bash scripts/docs-i18n-check.sh` — MUST PASS.
5. Run `go test ./internal/template/... -run TestTemplateNeutralityAudit` —
   MUST PASS (the edit must not introduce internal IDs/dates/SPEC tokens).

**Deliverables**: template + 4 docs pages edited; i18n check PASS; neutrality
audit PASS.

**Exit criteria**: no `preset:` token in the template file; no `preset:` AND
no `mode:` token in any of the 4 statusline.md pages; parity check green.

### M6 — Final verification + sync prep

**Scope**: full-suite verification, AC matrix sign-off, prepare for sync-phase.

**Steps**:

1. `go build ./...` (all platforms per §A.2 commands).
2. `go test ./...` (full suite).
3. `go test -race ./internal/statusline/... ./internal/profile/... ./internal/web/... ./internal/cli/...`.
4. `go vet ./...`.
5. `golangci-lint run --timeout=2m`.
6. Grep audit (all MUST return 0 matches):
   - `grep -rn 'PresetToSegments\|presetToSegments' internal/ pkg/` → 0 (both capital and lowercase — D1)
   - `grep -rn 'StatuslinePreset\|statuslinePreset' internal/ pkg/` → 0
   - `grep -rn 'statuslinePresetCanonical\|isCanonicalStatuslinePreset\|normalizeStatuslinePreset' internal/` → 0
   - `grep -rn 'StatuslineMode\|statuslineMode' internal/cli/profile_setup.go internal/profile/preferences.go internal/cli/profile_setup_translations.go` → 0 (D3 — mode preferences axis retired; `internal/statusline/` Builder API excluded)
   - `grep -rn 'statuslineModeCanonical\|isCanonicalStatuslineMode\|normalizeStatuslineMode\|defaultStatuslineMode\|ModeDefault\|ModeFull\|ModeCompact\|ModeVerbose\|ModeMinimal\|StatuslineModeTitle\|StatuslineModeDesc\|MigrationNoticeStatuslineMode\|SummaryStatuslineMode' internal/cli/profile_setup.go internal/cli/profile_setup_translations.go` → 0 (D3/D6)
   - `grep -n 'Preset' pkg/models/config.go internal/profile/preferences.go internal/cli/statusline.go internal/profile/sync.go` → 0 (D5 sync.go statuslineData included)
   - `grep -n 'compact\|minimal' internal/statusline/preset.go` → 0
   - `grep -n 'preset:\|mode:' internal/template/templates/.moai/config/sections/statusline.yaml` → 0 (template already has no mode:; both checked)
   - `for loc in en ko ja zh; do grep -n 'preset:\|mode:' docs-site/content/$loc/advanced/statusline.md; done` → 0 per locale (D7)
   - `grep -rn 'fieldsetStatusline\|id="statusline_preset"' internal/web/fieldsets.templ internal/web/root.templ internal/web/fieldsets_templ.go internal/web/root_templ.go` → 0 (D2 — component AND caller both gone)
7. Populated `progress.md §E.2` with verifiable command output.
8. Hand-off to manager-docs for sync-phase (CHANGELOG, README if applicable,
   frontmatter status transition).

**Exit criteria**: all gates green; AC matrix in acceptance.md verified; SPEC
ready for sync-phase.

---

## §G. Anti-Patterns (avoid)

- **AP-1 Hand-editing `fieldsets_templ.go`** — the generated file. Always edit
  `fieldsets.templ` and run `templ generate`. Hand-edits will be overwritten
  on the next regeneration and cause drift.
- **AP-2 Removing `CanonicalSegments`** — it is the SSOT imported across 4
  packages. Only `PresetToSegments` (the function) is removed.
- **AP-3 Introducing a "default preset" concept** — the retirement is a
  flattening. No new named-group concept replaces presets (spec.md §E.1).
- **AP-4 Editing only 3 of 4 docs locales** — the i18n check will fail. All 4
  in one commit.
- **AP-5 Skipping the cross-platform build check** — the `StatuslineConfig`
  struct change ripples into generated code; verify linux + windows build.
- **AP-6 Deleting tests that still apply** — the 3 web tests (K1/K2/K3) ARE
  obsolete, but `coverage_test.go:117` may have other valuable cases; only
  remove the `StatuslinePreset` field reference, not the whole test function.
- **AP-7 Touching segment rendering** — `renderer.go` render branches, git.go,
  usage.go, etc. are OUT OF SCOPE (spec.md §A.3, §E.6). If a compile error
  surfaces in renderer.go because it accepted a preset arg, surface it as a
  blocker — do NOT silently refactor rendering.
- **AP-8 Forgetting the legacy YAML silent-ignore** — REQ-SPR-021 MUST be
  verified: a `statusline.yaml` with a stale `preset: compact` line MUST load
  without error.

---

## §H. Cross-References

- `spec.md` §A.2 (target locations) / §B (requirements) / §C (constraints) /
  §E (exclusions)
- `acceptance.md` §D (AC matrix — the verifiable exit contract)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` §E
  (self-verification deliverables E1-E7)
- CLAUDE.local.md §6 (test isolation), §25 (template neutrality)
- Tier M routing: sync-phase → manager-docs; PR → manager-git
