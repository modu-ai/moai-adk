---
id: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
title: "Retire the statusline preset system (full/compact/minimal) and remove the web-console statusline panel"
version: "0.2.0"
status: in-progress
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/statusline;internal/cli;internal/profile;internal/web;pkg/models;internal/template;docs-site"
lifecycle: spec-anchored
tags: "statusline, preset, mode, retire, web-console, config-cleanup"
tier: M
---

## §A. Overview

The statusline preset system (`full` / `compact` / `minimal` / `custom`) is a
write-time shorthand that expands into the canonical 15-key segment map at save
time. At runtime, only the segment map is consulted — the preset field is never
re-read after segments materialize. This makes the preset a redundant layer that
adds surface area without runtime value:

- Three named presets (`full`, `compact`, `minimal`) collapse to one of two
  outcomes: "every segment on" or "a curated subset". The curated subsets are
  one-time conveniences that could live as documented segment configs.
- The `custom` preset is the only mode that carries real information (an
  explicit segment map). Keeping the preset field forces every code path
  (CLI render, profile sync, web save, wizard) to branch on `== "custom"`.
- The web console exposes a preset `<select>` whose only effect is to toggle
  the disabled state of the segment checkbox grid. Removing the preset leaves
  the segment toggles as the single, honest configuration surface.

This SPEC retires the preset system entirely. After this SPEC, the segment map
is the only statusline configuration lever besides theme.

### §A.1 Confirmed User Decisions (plan-phase input)

1. **RETIRE the preset system** — keep ONLY segment-based configuration
   (the `custom` concept becomes the default and only preset).
2. **REMOVE the statusline configuration panel from the moai web console.**
   The entire statusline fieldset (preset select + theme select + segment
   checkboxes) is removed, not just the preset dropdown.
3. **RETIRE the statusline `mode` preferences axis** — the runtime `mode:`
   YAML surface was already removed by SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001
   (the `mode` field is inert: every `StatuslineMode` variant collapses to
   `ModeDefault` via `NormalizeMode`). However, three residual surfaces of the
   mode axis survive in the **preferences / wizard / docs layers** and MUST be
   retired in lockstep with the preset retirement so the user-facing
   configuration surface becomes segments + theme only:
   - `ProfilePreferences.StatuslineMode` field (`internal/profile/preferences.go:45`)
   - wizard mode Select widget + canonical list + helpers
     (`internal/cli/profile_setup.go:18-68, 267-270, 337-338, 428-437, 556, 620-622`)
   - docs `mode:` example line (4 locales) + template/wizard i18n `Mode*` keys
   The `internal/statusline.StatuslineMode` type + constants + `NormalizeMode`
   in `internal/statusline/types.go`/`builder.go`/`renderer.go` are **preserved**
   as part of the Builder API (HARD-1 in `cli/statusline.go:62-66`): the Builder
   `Config.Mode` field is fed `ModeDefault` at construction. Only the
   **preferences/wizard/docs mode surface** is retired — the Builder API stays.

### §A.2 In Scope — Target Locations (investigated, not described)

Investigation findings (file:line — content summary). Three removal axes are
enumerated: (1) **preset axis** — the named-preset shorthand; (2) **mode
preferences axis** — the residual `StatuslineMode` surface in prefs/wizard/docs
(runtime `mode:` already removed by V3R5); (3) **web-console statusline panel**.

#### Preset axis — removal targets

| Surface | File | Lines | Content |
|---------|------|-------|---------|
| Preset expansion SSOT (capital) | `internal/statusline/preset.go` | 16-69 | `PresetToSegments()` function (canonical `compact`/`minimal`/`custom`/default branch). `CanonicalSegments` (L8-14) is OUT OF SCOPE — preserved as SSOT. |
| **Preset expansion wrapper (lowercase)** — D1 | `internal/cli/update.go` | 2732-2738 | `presetToSegments(preset, custom)` wrapper that delegates to `statusline.PresetToSegments`. Used by `cli/statusline.go:56` preset fallback. MUST be deleted alongside the capital `PresetToSegments` — `grep -rn 'PresetToSegments\|presetToSegments'` MUST return 0. |
| CLI render branch | `internal/cli/statusline.go` | 52-57 | `if Segments != nil { ... } else if Preset != "" && Preset != "full" { presetToSegments(...) }` — preset fallback path. |
| CLI config struct | `internal/cli/statusline.go` | 123-127, 146, 157 | `statuslineFileConfig.Preset` field + raw YAML `preset:` binding. |
| Models struct | `pkg/models/config.go` | 197-202 | `StatuslineConfig.Preset` field (note: `Mode` field is already absent — V3R5 removed it). |
| Profile preferences (preset) | `internal/profile/preferences.go` | 46 | `StatuslinePreset string` field. |
| Profile sync (write-effective) | `internal/profile/sync.go` | 114-142 | preset default-to-`full` + `StatuslineSegments != nil` / `StatuslinePreset != "custom"` switch. |
| Profile sync YAML struct — D5 | `internal/profile/sync.go` | 80-91 | `statuslineData.Preset` field (mirrors `models.StatuslineConfig`); folded into the preset-axis removal scope. |
| Wizard preset surface | `internal/cli/profile_setup.go` | 81-84, 96-116, 272, 427-460 (preset Select portion only), 486, 530-557 | `statuslinePresetCanonical`, `isCanonicalStatuslinePreset`, `normalizeStatuslinePreset`, preset Select section, segment section gating, build-segment-map branch. |
| Wizard translations (preset) | `internal/cli/profile_setup_translations.go` | 76-81, 201-208 (+ mirrors ko/ja/zh) | `StatuslinePresetTitle`/`Desc` + `PresetFull`/`PresetCompact`/`PresetMinimal`/`PresetCustom` keys across 4 locales. |
| Web handler binding | `internal/web/handlers.go` | 373 | `StatuslinePreset: r.PostFormValue("statusline_preset")` |
| Web handler custom branch | `internal/web/handlers.go` | 374-380 | `if prefs.StatuslinePreset == "custom" { ... bind segments ... }` |
| Web canonical list | `internal/web/validate.go` | 44 | `statuslinePresetCanonical = []string{"full","compact","minimal","custom"}` |
| Web validation rule | `internal/web/validate.go` | 134-136 | `if p.StatuslinePreset != "" && !inList(...)` |
| Template config | `internal/template/templates/.moai/config/sections/statusline.yaml` | 2-6 | `preset: "full"` line + the 4-line preset explanation comment block above it. (Template has no `mode:` line — already absent.) |
| Docs-site preset line (4-locale) | `docs-site/content/{en,ko,ja,zh}/advanced/statusline.md` | 213 (per locale) | `preset: custom  # full | minimal | custom` YAML example line. |

#### Mode preferences axis — removal targets (D3)

The runtime `mode:` YAML surface was already removed by
SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001 (`cli/statusline.go:62-66` documents
this: `mode` is fixed to `ModeDefault`). The **residual preferences / wizard /
docs surface** below MUST be retired so the user-facing config becomes
segments + theme only. The `internal/statusline.StatuslineMode` type and its
constants in `types.go`/`builder.go`/`renderer.go` are PRESERVED (Builder API).

| Surface | File | Lines | Content |
|---------|------|-------|---------|
| Profile preferences (mode) | `internal/profile/preferences.go` | 45 | `StatuslineMode string` field. |
| Wizard mode canonical + helpers | `internal/cli/profile_setup.go` | 18-20, 27, 32-68 | `statuslineModeCanonical`, `defaultStatuslineMode`, `isCanonicalStatuslineMode`, `normalizeStatuslineModeRaw`, `normalizeStatuslineMode`. |
| Wizard mode migration notice | `internal/cli/profile_setup.go` | 267, 270, 337-338 | `rawStatuslineMode` capture + `MigrationNoticeStatuslineMode` print. |
| Wizard mode Select widget | `internal/cli/profile_setup.go` | 428-437 | `huh.NewSelect(...)` for mode — Title `StatuslineModeTitle`, options `ModeDefault`/`ModeFull`, `Value(&statuslineMode)`. |
| Wizard mode prefs literal binding | `internal/cli/profile_setup.go` | 556 | `StatuslineMode: statuslineMode,` in returned `ProfilePreferences` literal. |
| Wizard mode summary line | `internal/cli/profile_setup.go` | 620-622 | `SummaryStatuslineMode` summary table row. |
| Wizard translations (mode) — D6 | `internal/cli/profile_setup_translations.go` | 65-73, 121, 128, 194-200, 237, 243 (+ mirrors ko/ja/zh) | `StatuslineModeTitle`/`Desc`, `ModeDefault`/`ModeCompact`/`ModeFull`/`ModeVerbose`/`ModeMinimal`, `SummaryStatuslineMode`, `MigrationNoticeStatuslineMode` across 4 locales. |
| Docs-site mode line (4-locale) — D7 | `docs-site/content/{en,ko,ja,zh}/advanced/statusline.md` | 211 (per locale) | `mode: default  # default | full` YAML example line. |

#### Web-console statusline panel — removal targets

| Surface | File | Lines | Content |
|---------|------|-------|---------|
| Web fieldset (REMOVE entire block) | `internal/web/fieldsets.templ` | 120-153 | statusline `<fieldset>` (legend + preset select + theme select + segment grid) — `fieldsetStatusline` templ component. |
| **Web fieldset caller** — D2 | `internal/web/root.templ` | 108 | `@fieldsetStatusline(view)` invocation. Removing the fieldset component from `fieldsets.templ` without removing this caller breaks `templ generate` (`root_templ.go:175` references the deleted component). MUST be removed in the same milestone. |
| Web generated mirror (fieldset) | `internal/web/fieldsets_templ.go` | 431-520 (approx) | generated from fieldsets.templ — regenerate via `templ generate`. |
| Web generated mirror (caller) | `internal/web/root_templ.go` | 175 | generated from root.templ — regenerated alongside root.templ edit. |
| Web tests | `internal/web/statusline_conditional_test.go`, `internal/web/statusline_empty_option_test.go`, `coverage_test.go:117`, `restyle_test.go:309` (comment ref) | whole files / preset refs | retire or refactor to match new surface. |

#### Test files — removal / refactor targets (D4)

The following test files contain compile-breaking references to symbols removed
by this SPEC and MUST be assigned to the milestone that removes the referenced
symbol:

| Test file | References | Owning milestone |
|-----------|-----------|------------------|
| `internal/profile/preferences_test.go` | `StatuslinePreset` field on `ProfilePreferences` (L78, 105-106, 177, 458, 474-475) | M2 (removes the field from preferences.go) |
| `internal/profile/sync_test.go` | `StatuslinePreset` prefs binding + `wrapper.Statusline.Preset` model field (L197, 209, 229-230, 262-263, 278, 299) | M2 (removes sync.go preset logic + models.Preset) |
| `internal/profile/statusline_segments_test.go` | `ProfilePreferences.StatuslinePreset` + `statusline.PresetToSegments` expansion (L60-138) | M2 (removes PresetToSegments + prefs field) |
| `internal/cli/wizard_config_test.go` | lowercase `presetToSegments` wrapper (L407-480) — D1 collateral | M2 (removes update.go wrapper) |
| `internal/cli/update_test.go` | lowercase `presetToSegments` wrapper (`TestPresetToSegments` L2721, L2779) — D1 collateral | M2 (removes update.go wrapper) |
| `internal/cli/profile_setup_test.go` | wizard i18n mode keys + `normalizeStatuslinePreset` (L38-59, 65-87) | M4 (removes wizard mode Select + preset helpers) |
| `internal/cli/profile_setup_translations_test.go` | `MigrationNoticeStatuslineMode` + preset i18n keys (L158-166, 222-241) | M4 (removes wizard translations) |

### §A.3 Out of Scope — PRESERVE (do NOT touch)

The following are explicitly preserved. The retirement MUST NOT alter their
behavior:

- **Segment rendering logic** — each of the 15 segments (git/model/task/memory/
  usage/etc.) keeps its rendering code path, theme integration, and graceful
  no-output handling for inactive states.
- **Theme system** — `catppuccin-mocha` / `catppuccin-latte` theming is
  untouched. The `theme` field of `StatuslineConfig` and its CLI/profile/web
  plumbing stays (the web-console removal removes the UI affordance, but the
  underlying theme config remains valid via `statusline.yaml` and the wizard
  must continue to offer theme selection — see §B.4 wizard scope note).
- **`status_line.sh` entrypoint** — the wrapper script behavior (`exec moai
  statusline < input`) is unchanged.
- **`CanonicalSegments`** — the 15-key canonical list in `preset.go` is the
  SSOT that every other package imports; it stays even after the function is
  removed. The file MAY be renamed to `segments.go` if `preset.go` becomes
  empty (aesthetic, not required).
- **Fallback behavior** — when `segments` is absent from `statusline.yaml`,
  the runtime falls back to all-enabled. This default is preserved.
- **`internal/statusline.StatuslineMode` Builder API** — the `StatuslineMode`
  type, its constants (`ModeDefault`/`ModeFull`/`ModeMinimal`/`ModeCompact`/
  `ModeVerbose`), and `NormalizeMode` in `internal/statusline/types.go`/
  `builder.go`/`renderer.go` are PRESERVED as part of the Builder API. The
  Builder `Config.Mode` field is fed `ModeDefault` at construction
  (`cli/statusline.go:62-66`). Only the **preferences / wizard / docs mode
  surface** is retired (§A.2 mode axis); the in-package Builder type stays
  because `NormalizeMode` collapses every variant to `ModeDefault` and removing
  the type would ripple into `builder.go`/`renderer.go` plus their test suites
  without changing any observable behavior.

---

## §B. Requirements (GEARS)

### §B.1 Core removal requirements

**REQ-SPR-001** (Ubiquitous): The `pkg/models.StatuslineConfig` struct SHALL
expose exactly two fields, `Segments` and `Theme`. The `Preset` field SHALL be
removed from the struct definition and from every struct that mirrors its
shape (`internal/profile/sync.go statuslineData`, `internal/cli.statuslineFileConfig`).
The `Mode` field is already absent (V3R5 removed it); no mode change is needed
at the models layer.

**REQ-SPR-002** (Ubiquitous): The `internal/profile.ProfilePreferences` struct
SHALL expose `StatuslineSegments` and `StatuslineTheme` only. Both the
`StatuslinePreset` field (L46) AND the `StatuslineMode` field (L45) SHALL be
removed — the mode preferences axis is retired in lockstep with the preset
axis (§A.1 decision 3).

**REQ-SPR-003** (Ubiquitous): The `internal/cli.statuslineFileConfig` struct
and its raw YAML binding SHALL omit the `preset` key. The CLI render path
SHALL consult only `Segments` (with all-enabled fallback when absent) and
`Theme`.

**REQ-SPR-004** (Ubiquitous): The `internal/statusline.PresetToSegments`
function SHALL be deleted. The lowercase wrapper `internal/cli/update.go:
presetToSegments` (L2732-2738) that delegates to it SHALL ALSO be deleted —
both the capital and lowercase symbols MUST be gone so that
`grep -rn 'PresetToSegments\|presetToSegments' internal/ pkg/` returns 0
matches. The file `internal/statusline/preset.go` SHALL retain the
`CanonicalSegments` variable (the SSOT for the 15-key list).

**REQ-SPR-005** (Ubiquitous): The following wizard symbols in
`internal/cli/profile_setup.go` SHALL be deleted: the preset canonical list
`statuslinePresetCanonical`, `isCanonicalStatuslinePreset`,
`normalizeStatuslinePreset`; AND the mode canonical list
`statuslineModeCanonical`, `defaultStatuslineMode`,
`isCanonicalStatuslineMode`, `normalizeStatuslineModeRaw`,
`normalizeStatuslineMode`. The wizard SHALL NOT render a preset Select widget
AND SHALL NOT render a mode Select widget.

**REQ-SPR-006** (Ubiquitous): The `internal/web.statuslinePresetCanonical`
variable and the `StatuslinePresets` field on the web view model SHALL be
deleted. The `validateStatusline` (or equivalent) validation rule for
`statusline_preset` SHALL be deleted.

### §B.2 Web console fieldset-caller removal

**REQ-SPR-022** (Ubiquitous): The `@fieldsetStatusline(view)` invocation in
`internal/web/root.templ` (L108) SHALL be removed in the same milestone that
removes the `fieldsetStatusline` templ component from `internal/web/fieldsets.templ`.
Both `fieldsets_templ.go` and `root_templ.go` SHALL be regenerated via
`templ generate` so no dangling reference to the deleted component survives.
A standalone removal of the component without removing the caller is
PROHIBITED — it breaks `templ generate`.

### §B.3 Profile sync behavior (post-retire)

**REQ-SPR-007** (Ubiquitous): The `internal/profile.syncStatusline` function
SHALL write `Segments` and `Theme` only. The preset-default-to-`full` branch
(L114-115) and the write-effective preset-expansion switch (L139-142) SHALL be
removed.

**REQ-SPR-008** (State-driven): **While** `prefs.StatuslineSegments` is non-nil,
`syncStatusline` SHALL persist the submitted segment map verbatim. **While**
`prefs.StatuslineSegments` is nil, `syncStatusline` SHALL preserve the existing
on-disk segments unchanged (no preset expansion, no default rewrite).

**REQ-SPR-009** (Ubiquitous): The default-segments fallback (when on-disk
`statusline.yaml` has no `segments` block) SHALL remain all-15-enabled. This is
the only default path; no preset participates in defaulting.

### §B.4 Web console removal

**REQ-SPR-010** (Ubiquitous): The `internal/web/fieldsets.templ` statusline
`<fieldset>` block (the entire Statusline section: preset select, theme select,
and segment checkbox grid) SHALL be removed. The fieldset SHALL NOT appear in
the rendered web console.

**REQ-SPR-011** (Ubiquitous): The `internal/web/handlers.go` POST handler SHALL
NOT bind `statusline_preset`, `statusline_theme`, or `statusline_segments*`
form values. The `prefs.StatuslinePreset` assignment and the
`if prefs.StatuslinePreset == "custom"` segment-binding branch SHALL be
removed.

**REQ-SPR-012** (Capability gate): **Where** the web console route renders
preferences, the statusline fieldset absence SHALL NOT cause a render error,
an i18n key-missing warning, or a layout regression in adjacent fieldsets.

### §B.5 Wizard scope

**REQ-SPR-013** (Ubiquitous): The `internal/cli/profile_setup.go` wizard SHALL
remove the preset Select section (Section 4 Display — preset portion) and the
`statuslinePreset == "custom"` gate on the segment section. The segment
multi-select section SHALL become unconditional (always shown when statusline
configuration is reached in the wizard flow).

**REQ-SPR-014** (Ubiquitous): The wizard SHALL continue to offer theme
selection. The theme Select widget and its canonical list
(`statuslineThemeCanonical`) are preserved.

**REQ-SPR-015** (State-driven): **While** the wizard reaches the statusline
segment step, it SHALL always build and emit the segment map (no preset
branch). The previously conditional segment build (L530-557) SHALL become
unconditional.

**REQ-SPR-023** (Ubiquitous): The `internal/cli/profile_setup.go` wizard SHALL
remove the mode Select widget (L428-437: `huh.NewSelect` with
`StatuslineModeTitle`/`StatuslineModeDesc`/`ModeDefault`/`ModeFull` options,
`Value(&statuslineMode)`), the `rawStatuslineMode` capture (L267), the
`statuslineMode` normalization call (L270), the `MigrationNoticeStatuslineMode`
print (L337-338), the `StatuslineMode: statuslineMode,` prefs literal binding
(L556), and the `SummaryStatuslineMode` summary row (L620-622). After removal,
the wizard SHALL NOT prompt for mode and SHALL NOT emit a mode migration
notice.

### §B.6 Template and docs

**REQ-SPR-016** (Ubiquitous): The template file
`internal/template/templates/.moai/config/sections/statusline.yaml` SHALL omit
the `preset:` key. The 4-line preset explanation comment block SHALL be
removed. The `segments:` and `theme:` blocks SHALL remain.

**REQ-SPR-017** (Ubiquitous): Each of the four docs-site statusline pages
(`docs-site/content/{en,ko,ja,zh}/advanced/statusline.md`) SHALL remove BOTH
the `preset:` line (L213) AND the `mode:` line (L211) from the YAML example.
The locale pages SHALL remain byte-parity on the preset-removal AND
mode-removal edit (same conceptual edit across all four).

**REQ-SPR-018** (Capability gate): **Where** a user upgrades an existing
project via `moai update`, the upgrade SHALL NOT delete a user's customized
`segments:` block in their local `statusline.yaml`. The template sync logic
SHALL treat `statusline.yaml` as a user-owned file (preserve user segments,
do not overwrite).

**REQ-SPR-024** (Ubiquitous): The `internal/cli/profile_setup_translations.go`
file SHALL remove the orphaned mode i18n keys across all 4 locales:
`StatuslineModeTitle`, `StatuslineModeDesc`, `ModeDefault`, `ModeCompact`,
`ModeFull`, `ModeVerbose`, `ModeMinimal`, `SummaryStatuslineMode`,
`MigrationNoticeStatuslineMode`. The orphaned preset i18n keys
(`StatuslinePresetTitle`, `StatuslinePresetDesc`, `PresetFull`, `PresetCompact`,
`PresetMinimal`, `PresetCustom`) SHALL also be removed. The theme i18n keys
(`StatuslineThemeTitle`/`Desc`/`StatuslineThemeCanonical` labels) are preserved
(REQ-SPR-014).

### §B.7 Non-regression (PRESERVE behavior)

**REQ-SPR-019** (Ubiquitous): The 15 segment rendering paths (model, context,
output_style, claude_version, moai_version, session_time, effort_thinking,
usage_5h, usage_7d, directory, git_status, git_branch, worktree, task, pr)
SHALL produce byte-identical output for any given stdin JSON and segment
toggle state, compared to pre-retire behavior.

**REQ-SPR-020** (Ubiquitous): The theme application (foreground/background
color codes for `catppuccin-mocha` and `catppuccin-latte`) SHALL produce
byte-identical output compared to pre-retire behavior.

**REQ-SPR-021** (State-driven): **While** `statusline.yaml` contains a
`preset:` key from a pre-retire install, the loader SHALL ignore the key
silently (no error, no warning to stderr). The `segments:` block, if present,
SHALL be consulted as the source of truth; if `segments:` is absent, the
all-enabled fallback SHALL apply.

---

## §C. Constraints

1. **Zero behavior regression on segment rendering** — output bytes for any
   given (stdin JSON, segment toggle state, theme) tuple MUST be identical
   pre- and post-retire. The retirement removes configuration surface, not
   rendering capability.
2. **Backward compatibility on `statusline.yaml`** — existing installs may
   carry a `preset:` line. The loader MUST tolerate and ignore it silently
   (REQ-SPR-021). No migration step is required for end users.
3. **No new public API** — this is a removal SPEC. No new exported functions,
   types, or config fields are introduced.
4. **Template neutrality preserved** — the template edit
   (`statusline.yaml`) MUST remain free of internal MoAI-ADK identifiers
   (SPEC IDs, REQ tokens, internal dates) per CLAUDE.local.md §25.
5. **4-locale docs parity** — the docs-site edit MUST apply to all four
   locales (en/ko/ja/zh) with matching conceptual scope. The
   `scripts/docs-i18n-check.sh` harness MUST continue to PASS.
6. **Test isolation** — all new/modified tests MUST use `t.TempDir()` for
   filesystem fixtures per CLAUDE.local.md §6. No test writes to the live
   project's `.moai/config/`.

---

## §D. Acceptance Criteria Summary

The acceptance criteria are enumerated in `acceptance.md`. The summary count:

- **MUST ACs**: 25 (grep-verifiable structural removals + behavior parity + mode-axis removal + fieldset-caller removal)
- **SHOULD ACs**: 3 (docs-site polish, wizard flow cleanup niceties)
- **Total**: 28

Key grep-verifiable ACs (full text in acceptance.md):
- `grep -rn 'compact\|minimal' internal/statusline/preset.go` → 0 matches
- `grep -rn 'PresetToSegments\|presetToSegments' internal/ pkg/` → 0 matches (both capital and lowercase)
- `grep -n 'Preset' pkg/models/config.go internal/profile/preferences.go internal/cli/statusline.go` → 0 matches
- `grep -n 'Preset\|StatuslineMode\|statuslineMode' internal/profile/preferences.go` → 0 matches (both preset AND mode fields gone)
- `grep -rn 'StatuslinePreset\|statuslinePreset\|StatuslineMode\|statuslineMode' internal/cli/profile_setup.go internal/cli/profile_setup_translations.go` → 0 matches (wizard + i18n cleaned)
- `grep -n 'preset:\|mode:' internal/template/templates/.moai/config/sections/statusline.yaml` → 0 matches (template already has no mode: line; preset: removed)
- `grep -n 'id="statusline_preset"\|fieldsetStatusline' internal/web/fieldsets.templ internal/web/root.templ` → 0 matches (fieldset + caller both gone)
- 4-locale `preset:` AND `mode:` line removal parity via `scripts/docs-i18n-check.sh`
- Segment rendering byte-parity via existing renderer tests (zero modifications to renderer_test.go expectation values)

---

## §E. Exclusions (What NOT to Build)

1. **DO NOT** introduce a new "default segment profile" or named-segment-group
   concept to replace presets. The retirement is a flattening, not a rename.
   Users who previously relied on `compact` or `minimal` will configure their
   segment toggles once; no migration helper is provided.
2. **DO NOT** add a migration command (`moai statusline migrate`) that
   expands legacy presets into segment maps. The loader's silent-ignore
   behavior (REQ-SPR-021) is the only migration story.
3. **DO NOT** touch the `status_line.sh.tmpl` wrapper script. It has no
   preset references and its behavior is unchanged.
4. **DO NOT** touch the theme system (`theme.go`, `gradient.go`,
   `catppuccin-mocha`/`catppuccin-latte` palettes). Theme stays.
5. **DO NOT** remove `CanonicalSegments` from `preset.go`. It is the SSOT
   imported by CLI, profile, and web packages.
6. **DO NOT** modify segment rendering internals (`renderer.go` render
   branches, `git.go`, `usage.go`, `task.go`, `memory.go`, `metrics.go`,
   `model_cache.go`, `version.go`). The only `renderer.go` edits, if any,
   are removal of preset-related parameters from constructor signatures —
   and only if the constructor currently accepts a preset argument
   (investigation shows `New(opts)` takes `Options`, not preset; verify in
   run-phase).
7. **DO NOT** remove the statusline CLI subcommand (`moai statusline`).
   The subcommand stays; only its internal config-loading path changes.
8. **DO NOT** introduce a new web UI affordance to replace the removed
   statusline panel. Users who want to reconfigure statusline post-retire
   edit `statusline.yaml` directly or use the wizard.

---

## §F. Risks and Open Questions

### §F.1 Risks

| ID | Risk | Likelihood | Impact | Mitigation |
|----|------|-----------|--------|------------|
| R1 | Existing user `statusline.yaml` carries `preset: compact` with NO `segments:` block — after retire, loader falls back to all-enabled, changing their visible statusline | Medium | Medium (user-visible) | Silent-ignore (REQ-SPR-021) + docs note in the v3.x release changelog explaining the one-time reconfiguration. Acceptance: this is a documented breaking config-surface change, not a bug. |
| R2 | `fieldsets_templ.go` regeneration drift — hand-editing the generated file instead of regenerating from `fieldsets.templ` | Low | High (compile error) | Run the canonical `templ generate` (or project equivalent) after editing `fieldsets.templ`. Verify via `go build ./internal/web/...`. |
| R3 | `profile_setup.go` wizard flow breaks because Section 4 (Display) loses its preset Select but the surrounding form group expects N children | Medium | Medium | REQ-SPR-013/015 make segment section unconditional. Verify the wizard renders end-to-end via the existing wizard test (or add one if absent). |
| R4 | docs-site 4-locale parity check fails because one locale's `preset:` line is edited differently | Low | Low | `scripts/docs-i18n-check.sh` is the gate. Edit all four in one commit. |
| R5 | `internal/web/coverage_test.go:117` uses `StatuslinePreset: "weird"` — removing the field breaks the test compile | High (certain) | Low (test-only) | Retire the test case or refactor it to target the post-retire surface. In scope per §A.2. |

### §F.2 Open Questions (resolve before/during run-phase, NOT user-blocking)

1. **Should `preset.go` be renamed to `segments.go`?** The file would contain
   only `CanonicalSegments` after `PresetToSegments` removal. Aesthetic only;
   no behavior change. Default: leave the filename to minimize git churn, OR
   rename for honesty. Decide in M2.
2. **Does the wizard Section 4 (Display) still make sense with only theme +
   segments?** It may collapse into a single "Statusline" section. UX polish;
   not blocking. Decide in M4.
3. **Should the `StatuslineConfig.Preset` field removal use `yaml:"-"` for a
   release to ease the transition, or hard-delete?** Default: hard-delete
   (REQ-SPR-001). The silent-ignore behavior (REQ-SPR-021) covers legacy
   YAML files at the loader layer, not the struct layer.

---

## §G. History

- 2026-06-17: SPEC created (plan-phase, draft v0.1.0). Two user decisions
  confirmed (full retirement + web console panel removal). Tier M, era V3R6.
  All target file:line locations verified by direct grep/read investigation —
  descriptions in delegation prompt matched actual code, with two refinements
  captured: (a) `CanonicalSegments` is co-located in `preset.go` and MUST be
  preserved; (b) the `StatuslineConfig` struct lives in `pkg/models/config.go`
  (not `internal/config/types.go` as the prompt's wording suggested —
  `internal/config/types.go` embeds it via `models.StatuslineConfig`).
- 2026-06-17: **v0.2.0** — plan-auditor iter-1 FAIL (0.62, Tier M threshold
  0.80) remediation. 7 defects addressed (D1-D4 BLOCKING, D5-D7 SHOULD-FIX):
  - D1: enumerated lowercase `presetToSegments` wrapper in
    `internal/cli/update.go:2732-2738` as a removal target (REQ-SPR-004 +
    AC-SPR-002 grep now catches both cases).
  - D2: added REQ-SPR-022 for `@fieldsetStatusline(view)` caller removal in
    `internal/web/root.templ:108` (generated mirror `root_templ.go:175`); the
    caller MUST be removed in the same milestone as the fieldset component or
    `templ generate` breaks.
  - D3 (internal contradiction resolution): decision = **RETIRE the mode
    preferences axis** (not preserve). The runtime `mode:` surface was already
    removed by SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001 (`mode` is inert —
    every variant collapses to `ModeDefault`); the residual mode surface in
    preferences/wizard/docs is retired in lockstep with preset retirement. The
    `internal/statusline.StatuslineMode` Builder API type/constants are
    PRESERVED (Builder `Config.Mode` fed `ModeDefault`). REQ-SPR-002 now
    removes both `StatuslinePreset` AND `StatuslineMode` from
    `ProfilePreferences`; REQ-SPR-005 adds mode canonical/helpers deletion;
    new REQ-SPR-023 covers wizard mode Select widget removal; REQ-SPR-017
    adds docs `mode:` line removal; REQ-SPR-024 adds mode i18n keys cleanup.
  - D4: 5 test files assigned to owning milestones
    (preferences_test.go/sync_test.go/statusline_segments_test.go → M2;
    profile_setup_test.go/profile_setup_translations_test.go → M4) + 2 D1
    collateral test files (wizard_config_test.go/update_test.go → M2).
  - D5: `statuslineData.Preset` struct field folded into preset-axis removal
    scope (REQ-SPR-001).
  - D6: `Mode*` i18n keys cleanup folded into REQ-SPR-024.
  - D7: docs `mode:` line removal folded into REQ-SPR-017.
  AC count rose from 21 → 27 (24 MUST + 3 SHOULD); REQ count rose from 21 → 24.

---

## §H. Cross-References

- `.moai/specs/SPEC-STATUSLINE-001/` — original statusline system (historical)
- `.moai/specs/SPEC-V3R5-STATUSLINE-FULL-MODE-CLEANUP-001/` — prior mode cleanup
- `.moai/specs/SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001/` — wizard preset section origin (REQ-SPW-001/002)
- `internal/statusline/CLAUDE.md` (if present) — statusline package conventions
- `internal/cli/CLAUDE.md` — CLI subagent boundary + cobra registration rules
- CLAUDE.local.md §25 — template internal-content isolation (applies to `statusline.yaml` edit)
- CLAUDE.local.md §6 — test isolation (applies to any new test fixtures)
