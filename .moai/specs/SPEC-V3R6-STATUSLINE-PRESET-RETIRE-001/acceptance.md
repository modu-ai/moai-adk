---
id: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001
title: "Acceptance — Retire statusline preset system + remove web-console statusline panel"
version: "0.2.0"
status: draft
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "acceptance"
lifecycle: spec-anchored
tags: "acceptance, statusline, preset, mode, retire"
tier: M
---

# Acceptance Criteria — SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001

## §A. Verification Philosophy

Every AC is **observable**: it is verified by a specific command (grep, build,
test, byte-diff) whose output can be pasted verbatim into `progress.md §E.2`.
No AC relies on a summary claim or a non-reproducible assertion.

Severity model:
- **MUST** — blocks merge. Failure means the retirement is incomplete or a
  regression was introduced.
- **SHOULD** — does not block merge, but tracks polish/debt. Must be
  documented in `progress.md §E.2` with rationale if deferred.

---

## §B. AC ID Convention

- `AC-SPR-{NNN}` — acceptance criterion IDs, scoped to this SPEC.
- Numeric order groups by area:
  - 001-010: structural removal (Go code — preset + mode + wrapper)
  - 011-014: web console removal (fieldset + handler + validation + caller)
  - 015-018: wizard cleanup (preset helpers + mode Select + translations + theme-preserve)
  - 019-021: template + docs (preset + mode + parity)
  - 022-024: non-regression (PRESERVE behavior)
  - 025-027: new MUST ACs for mode-axis removal (D3), fieldset-caller (D2), lowercase wrapper (D1)
  - 028-030: SHOULD (polish)

---

## §C. Given-When-Then Template

Each AC follows:
```
Given <initial state>
When <action/command>
Then <observable outcome>
Verification: <exact command(s) to run>
```

---

## §D. Acceptance Criteria Matrix

### §D.1 Structural removal — Go code (MUST)

#### AC-SPR-001 (MUST) — `Preset` field removed from `StatuslineConfig`

```
Given the retire is complete
When the struct definition is inspected
Then `pkg/models.StatuslineConfig` has exactly two fields: Segments, Theme
Verification: grep -n 'Preset' pkg/models/config.go
  Expected: 0 matches in the StatuslineConfig struct region (L197-202)
```

#### AC-SPR-002 (MUST) — `PresetToSegments` function AND lowercase wrapper deleted (D1)

```
Given the retire is complete
When the statusline package + CLI are grepped for BOTH capital and lowercase symbols
Then no `PresetToSegments` symbol (capital) AND no `presetToSegments` wrapper
(lowercase, in internal/cli/update.go:2736) exist anywhere in internal/ or pkg/
Verification: grep -rn 'PresetToSegments\|presetToSegments' internal/ pkg/
  Expected: 0 matches
Note: iter-0 AC grepped only the capital symbol and missed the lowercase wrapper
at internal/cli/update.go:2732-2738. This AC now catches both cases.
```

#### AC-SPR-003 (MUST) — `CanonicalSegments` preserved

```
Given the retire is complete
When the statusline package is grepped
Then `CanonicalSegments` variable still exists and still lists 15 keys
Verification: grep -n 'CanonicalSegments' internal/statusline/preset.go
  Expected: ≥1 match (the variable declaration); verify the 15-key list intact
  Cross-check: grep -rn 'statusline.CanonicalSegments\|CanonicalSegments' internal/ pkg/
  Expected: call sites in cli/profile/web still resolve
```

#### AC-SPR-004 (MUST) — `compact`/`minimal` preset strings gone from preset.go

```
Given the retire is complete
When the preset.go file is grepped for the named-preset tokens
Then no `compact` or `minimal` string literals remain
Verification: grep -n 'compact\|minimal' internal/statusline/preset.go
  Expected: 0 matches
```

#### AC-SPR-005 (MUST) — `StatuslinePreset` AND `StatuslineMode` fields removed from `ProfilePreferences` (D3)

```
Given the retire is complete
When the preferences struct is inspected
Then `internal/profile.ProfilePreferences` has no `StatuslinePreset` field (L46)
AND no `StatuslineMode` field (L45). The mode preferences axis is retired
in lockstep with preset (runtime mode: already removed by V3R5).
Verification: grep -n 'StatuslinePreset\|StatuslineMode' internal/profile/preferences.go
  Expected: 0 matches
```

#### AC-SPR-006 (MUST) — `statuslineFileConfig.Preset` field removed

```
Given the retire is complete
When the CLI statusline config struct is inspected
Then `internal/cli.statuslineFileConfig` has no `Preset` field and the raw
YAML binding has no `preset:` key
Verification: grep -n 'Preset\|preset:' internal/cli/statusline.go
  Expected: 0 matches
```

#### AC-SPR-007 (MUST) — Profile sync writes segments + theme only (incl. `statuslineData` struct — D5)

```
Given the retire is complete
When `syncStatusline` + the `statuslineData` YAML struct are grepped
Then no `StatuslinePreset` reference, no `PresetToSegments` call, AND no `Preset`
field on the `statuslineData` struct (L80-91, which mirrors models.StatuslineConfig)
Verification: grep -n 'Preset\|presetToSegments\|PresetToSegments' internal/profile/sync.go
  Expected: 0 matches
Note: D5 folded the statuslineData.Preset struct field into this AC's scope.
```

#### AC-SPR-008 (MUST) — CLI render path is segments-only

```
Given the retire is complete
When the CLI statusline render branch is inspected
Then the `if Preset != "" && Preset != "full"` branch is gone; only
`segmentConfig = statuslineCfg.Segments` (nil-safe) remains
Verification: grep -n 'Preset != \|presetToSegments(' internal/cli/statusline.go
  Expected: 0 matches
```

#### AC-SPR-009 (MUST) — Preset-default-to-`full` branch removed from sync

```
Given the retire is complete
When `syncStatusline` default logic is inspected
Then the line `current.Statusline.Preset = "full"` is gone
Verification: grep -n 'Preset = "full"\|Preset = full' internal/profile/sync.go
  Expected: 0 matches
```

#### AC-SPR-010 (MUST) — No `Preset` token anywhere in 5 core packages

```
Given the retire is complete
When the 5 affected packages are grepped
Then no `Preset` token (case-sensitive) appears in non-test Go files of
internal/statusline, internal/cli/statusline.go, internal/profile, pkg/models,
internal/web (excluding vendored/generated noise)
Verification:
  grep -rn 'Preset' internal/statusline/ internal/cli/statusline.go \
    internal/profile/ pkg/models/config.go internal/web/handlers.go \
    internal/web/validate.go internal/web/fieldsets.templ
  Expected: 0 matches
```

### §D.2 Web console removal (MUST)

#### AC-SPR-011 (MUST) — Statusline fieldset AND its caller removed (D2)

```
Given the retire is complete
When fieldsets.templ + root.templ + their generated mirrors are grepped
Then no statusline `<fieldset>` block (fieldsetStatusline component in fieldsets.templ),
no `@fieldsetStatusline(view)` caller in root.templ, and no generated
`fieldsetStatusline` references in fieldsets_templ.go or root_templ.go remain
Verification:
  grep -rn 'fieldsetStatusline\|id="statusline_preset"\|id="statusline_theme"\|custom-segments\|data-i18n="sec.statusline' \
    internal/web/fieldsets.templ internal/web/root.templ \
    internal/web/fieldsets_templ.go internal/web/root_templ.go
  Expected: 0 matches
Note: D2 found that root.templ:108 invokes @fieldsetStatusline(view) and
root_templ.go:175 renders it. Removing only the component (fieldsets.templ)
without the caller breaks templ generate. This AC now covers BOTH.
```

#### AC-SPR-012 (MUST) — Web handler does not bind statusline form values

```
Given the retire is complete
When handlers.go is grepped
Then no `r.PostFormValue("statusline_preset")`,
no `r.PostFormValue("statusline_theme")`,
no `r.PostFormValue("segment_*")`,
no `if prefs.StatuslinePreset == "custom"` branch remains
Verification:
  grep -n 'statusline_preset\|statusline_theme\|segment_\|StatuslinePreset' internal/web/handlers.go
  Expected: 0 matches
```

#### AC-SPR-013 (MUST) — Web view model + validation cleaned

```
Given the retire is complete
When validate.go + the view model struct are grepped
Then `statuslinePresetCanonical`, `StatuslinePresets` view field, and the
`statusline_preset` validation rule are all gone
Verification:
  grep -n 'statuslinePresetCanonical\|StatuslinePresets\|statusline_preset' internal/web/validate.go internal/web/handlers.go
  Expected: 0 matches
```

### §D.3 Wizard cleanup (MUST)

#### AC-SPR-014 (MUST) — Wizard preset AND mode canonical lists + helpers deleted (D3)

```
Given the retire is complete
When profile_setup.go is grepped
Then `statuslinePresetCanonical`, `isCanonicalStatuslinePreset`,
`normalizeStatuslinePreset` (preset helpers) AND `statuslineModeCanonical`,
`defaultStatuslineMode`, `isCanonicalStatuslineMode`,
`normalizeStatuslineModeRaw`, `normalizeStatuslineMode` (mode helpers) are
ALL gone. The mode helpers are retired in lockstep with preset helpers.
Verification:
  grep -n 'statuslinePresetCanonical\|isCanonicalStatuslinePreset\|normalizeStatuslinePreset\|statuslineModeCanonical\|defaultStatuslineMode\|isCanonicalStatuslineMode\|normalizeStatuslineModeRaw\|normalizeStatuslineMode' internal/cli/profile_setup.go
  Expected: 0 matches
```

#### AC-SPR-015 (MUST) — Wizard segment section is unconditional

```
Given the retire is complete
When the wizard segment-step gating is inspected
Then no `statuslinePreset != "custom"` hide-condition remains; the segment
step is always shown when reached
Verification:
  grep -n 'statuslinePreset != "custom"\|statuslinePreset == "custom"\|StatuslinePreset:' internal/cli/profile_setup.go
  Expected: 0 matches
```

#### AC-SPR-016 (MUST) — Wizard theme Select preserved

```
Given the retire is complete
When the wizard theme step is inspected
Then `statuslineThemeCanonical` and the theme Select widget are intact
Verification:
  grep -n 'statuslineThemeCanonical\|StatuslineThemeTitle\|StatuslineTheme' internal/cli/profile_setup.go
  Expected: ≥1 match (the canonical list + the theme step survive)
```

### §D.4 Template + docs (MUST)

#### AC-SPR-017 (MUST) — Template statusline.yaml has no preset key

```
Given the retire is complete
When the template config file is grepped
Then no `preset:` line and no preset explanation comment remains
Verification:
  grep -n 'preset:\|Preset name\|Write-time shorthand' \
    internal/template/templates/.moai/config/sections/statusline.yaml
  Expected: 0 matches
```

#### AC-SPR-018 (MUST) — 4-locale docs parity on preset AND mode removal (D7)

```
Given the retire is complete
When each of the 4 statusline.md pages is grepped
Then no `preset:` line (L213) AND no `mode:` line (L211) remain in any of
en/ko/ja/zh
Verification:
  for loc in en ko ja zh; do
    echo "== $loc =="
    grep -n 'preset:\|mode:' docs-site/content/$loc/advanced/statusline.md
  done
  Expected: 0 matches per locale for both patterns

Cross-locale parity gate:
  bash scripts/docs-i18n-check.sh
  Expected: exit 0 (PASS)
```

### §D.4a Mode-axis removal (D3) — additional MUST ACs

#### AC-SPR-025 (MUST) — Wizard mode Select widget + migration notice removed

```
Given the retire is complete
When profile_setup.go is grepped for the mode Select widget and its supporting logic
Then the mode huh.NewSelect block (Title StatuslineModeTitle, options
ModeDefault/ModeFull, Value &statuslineMode), the rawStatuslineMode capture
(L267), the statuslineMode normalization (L270), the MigrationNoticeStatuslineMode
print (L337-338), the StatuslineMode: prefs literal binding (L556), and the
SummaryStatuslineMode summary row (L620-622) are ALL gone
Verification:
  grep -n 'StatuslineModeTitle\|StatuslineModeDesc\|ModeDefault\|ModeFull\|rawStatuslineMode\|statuslineMode\|MigrationNoticeStatuslineMode\|SummaryStatuslineMode\|StatuslineMode:' internal/cli/profile_setup.go
  Expected: 0 matches
```

#### AC-SPR-026 (MUST) — Mode i18n keys removed from wizard translations (D6)

```
Given the retire is complete
When profile_setup_translations.go is grepped across all 4 locales
Then the orphaned mode i18n keys (StatuslineModeTitle, StatuslineModeDesc,
ModeDefault, ModeCompact, ModeFull, ModeVerbose, ModeMinimal,
SummaryStatuslineMode, MigrationNoticeStatuslineMode) AND the orphaned
preset i18n keys (StatuslinePresetTitle, StatuslinePresetDesc, PresetFull,
PresetCompact, PresetMinimal, PresetCustom) are all gone. Theme i18n keys
(StatuslineThemeTitle/Desc + canonical labels) survive.
Verification:
  grep -n 'StatuslineModeTitle\|StatuslineModeDesc\|ModeDefault\|ModeCompact\|ModeFull\|ModeVerbose\|ModeMinimal\|SummaryStatuslineMode\|MigrationNoticeStatuslineMode\|StatuslinePresetTitle\|StatuslinePresetDesc\|PresetFull\|PresetCompact\|PresetMinimal\|PresetCustom' internal/cli/profile_setup_translations.go
  Expected: 0 matches
Cross-check (theme keys preserved):
  grep -n 'StatuslineThemeTitle\|StatuslineThemeDesc\|statuslineThemeCanonical' internal/cli/profile_setup_translations.go
  Expected: ≥1 match per locale (theme keys survive)
```

#### AC-SPR-027 (MUST) — D4 test files compile-clean post-retire

```
Given the retire is complete
When the 7 test files assigned to M2/M4 (D4 + D1 collateral) are compiled
Then each file either (a) compiles cleanly with its preset/mode references
removed, or (b) is deleted entirely. No compile-breaking reference to a
removed symbol survives.
Verification:
  go build ./...
  go test -run xxx_noop ./internal/profile/... ./internal/cli/...
  Expected: exit 0 (compile-clean; the -run xxx_noop skips actual test
  execution, only verifying compilation)
Files in scope:
  M2: internal/profile/preferences_test.go,
      internal/profile/sync_test.go,
      internal/profile/statusline_segments_test.go (likely deleted),
      internal/cli/wizard_config_test.go (preset tests removed),
      internal/cli/update_test.go (TestPresetToSegments removed).
  M4: internal/cli/profile_setup_test.go (mode/preset assertions removed),
      internal/cli/profile_setup_translations_test.go
      (mode/preset key rows removed).
```

#### AC-SPR-028 (MUST) — `internal/statusline` Builder API preserved (D3 non-regression)

```
Given the retire is complete
When the statusline package Builder API symbols are grepped
Then the StatuslineMode type, its constants (ModeDefault/ModeFull/ModeMinimal/
ModeCompact/ModeVerbose), and NormalizeMode in internal/statusline/types.go +
builder.go + renderer.go are INTACT. Only the prefs/wizard/docs mode surface
was retired; the in-package Builder type is preserved and fed ModeDefault at
construction (cli/statusline.go:62-66).
Verification:
  grep -n 'type StatuslineMode\|ModeDefault\|ModeFull\|ModeMinimal\|ModeCompact\|ModeVerbose\|func NormalizeMode' internal/statusline/types.go
  Expected: ≥1 match per symbol (the Builder API survives)
  grep -n 'Mode:.*statusline.ModeDefault\|Mode:.*ModeDefault' internal/cli/statusline.go
  Expected: ≥1 match (Builder fed ModeDefault)
```

### §D.5 Non-regression — PRESERVE behavior (MUST)

#### AC-SPR-019 (MUST) — Segment rendering byte-parity

```
Given the retire is complete and the existing renderer test suite is intact
When `go test ./internal/statusline/...` is run
Then every existing segment-rendering test passes WITHOUT modification to
its expected-output values (only test fixtures referencing preset config may
be edited; expected render strings are unchanged)
Verification:
  go test ./internal/statusline/... -v -count=1
  Expected: PASS; no expected-string modification in:
    - renderer_test.go expectation values
    - git_test.go, usage_test.go, task_test.go, memory_test.go, metrics_test.go
    - theme_test.go, gradient_test.go
```

#### AC-SPR-020 (MUST) — Legacy `preset:` key silently ignored

```
Given a statusline.yaml that still contains a legacy `preset: compact` line
plus a valid `segments:` block
When the CLI loads the config
Then no error, no warning on stderr; the segments block is consulted as-is
Verification (characterization test, added in M2 or M3):
  Create a temp statusline.yaml with both `preset: compact` and an explicit
  segments block; load via the CLI config loader; assert (a) no error,
  (b) the segments block values are returned verbatim, (c) the preset value
  is not reflected anywhere in the resulting config.
  Test name: TestLoadStatuslineConfig_LegacyPresetIgnored
  Expected: PASS
```

#### AC-SPR-021 (MUST) — Full test suite green + race clean

```
Given the retire is complete
When the full test suite + race detector are run
Then `go test ./...` exits 0 and `go test -race` on the 4 affected packages
exits 0
Verification:
  go test ./...
  go test -race ./internal/statusline/... ./internal/profile/... \
    ./internal/web/... ./internal/cli/...
  Expected: both exit 0
```

### §D.6 SHOULD (polish)

#### AC-SPR-022 (SHOULD) — Wizard Section 4 title adjusted

```
Given the retire is complete
When the wizard Section 4 (Display) is inspected
Then if the section title still references "preset", it is renamed to reflect
the new segments-only content (e.g., "Statusline segments + theme")
Verification:
  grep -n 'PresetTitle\|PresetDesc\|preset' internal/cli/profile_setup.go
  Expected: 0 matches OR matches only in commented-out historical context
  (If the i18n keys PresetTitle/PresetDesc are no longer referenced, they
  SHOULD also be removed from the i18n catalog. Track as SHOULD.)
```

#### AC-SPR-023 (SHOULD) — `preset.go` renamed to `segments.go`

```
Given the retire is complete
When the statusline package is listed
Then if the file now contains only CanonicalSegments, it MAY be renamed to
`segments.go` for honesty. This is aesthetic; not blocking.
Verification (if rename applied):
  ls internal/statusline/ | grep -E 'preset.go|segments.go'
  Expected: only segments.go (preset.go gone) OR preset.go unchanged
  (either outcome is acceptable for SHOULD)
```

#### AC-SPR-024 (SHOULD) — i18n catalog cleaned of orphaned statusline-preset keys

```
Given the retire is complete
When the i18n message catalog is grepped
Then orphaned keys `PresetFull`, `PresetCompact`, `PresetMinimal`,
`PresetCustom`, `StatuslinePresetTitle`, `StatuslinePresetDesc` (and any
`sec.statusline.*` / `count.statusline` keys rendered only by the removed
fieldset) are removed or documented as retained-for-backward-compat
Verification:
  grep -rn 'PresetFull\|PresetCompact\|PresetMinimal\|PresetCustom\|StatuslinePresetTitle' internal/cli/ internal/web/
  Expected: 0 references in active code (catalog entries may remain with a
  retention comment)
```

---

## §E. Traceability Matrix (REQ ↔ AC)

| REQ | AC(s) | Severity |
|-----|-------|----------|
| REQ-SPR-001 (struct field removed, incl. statuslineData mirror — D5) | AC-SPR-001, AC-SPR-007 | MUST |
| REQ-SPR-002 (ProfilePreferences: StatuslinePreset AND StatuslineMode removed — D3) | AC-SPR-005 | MUST |
| REQ-SPR-003 (CLI config struct cleaned) | AC-SPR-006, AC-SPR-008 | MUST |
| REQ-SPR-004 (PresetToSegments AND lowercase wrapper deleted — D1) | AC-SPR-002, AC-SPR-004 | MUST |
| REQ-SPR-005 (wizard preset + mode canonical lists/helpers gone — D3) | AC-SPR-014 | MUST |
| REQ-SPR-006 (web canonical list + validation gone) | AC-SPR-013 | MUST |
| REQ-SPR-007 (sync writes segments+theme only, incl. statuslineData — D5) | AC-SPR-007, AC-SPR-009 | MUST |
| REQ-SPR-008 (segments verbatim / preserve-existing) | AC-SPR-007 (implicit via test) | MUST |
| REQ-SPR-009 (default-segments all-enabled fallback) | AC-SPR-020 (characterization) | MUST |
| REQ-SPR-010 (fieldset removed) | AC-SPR-011 | MUST |
| REQ-SPR-011 (handler does not bind) | AC-SPR-012 | MUST |
| REQ-SPR-012 (no render/i18n regression) | AC-SPR-011, AC-SPR-019 | MUST |
| REQ-SPR-013 (wizard preset section gone) | AC-SPR-014, AC-SPR-015 | MUST |
| REQ-SPR-014 (wizard theme preserved) | AC-SPR-016 | MUST |
| REQ-SPR-015 (segment build unconditional) | AC-SPR-015 | MUST |
| REQ-SPR-016 (template preset key removed) | AC-SPR-017 | MUST |
| REQ-SPR-017 (4-locale docs: preset AND mode line removed — D7) | AC-SPR-018 | MUST |
| REQ-SPR-018 (moai update preserves user segments) | AC-SPR-020 (implicit) | MUST |
| REQ-SPR-019 (segment rendering byte-parity) | AC-SPR-019 | MUST |
| REQ-SPR-020 (theme byte-parity) | AC-SPR-019 | MUST |
| REQ-SPR-021 (legacy preset silently ignored) | AC-SPR-020 | MUST |
| REQ-SPR-022 (fieldset caller @fieldsetStatusline removed — D2) | AC-SPR-011 | MUST |
| REQ-SPR-023 (wizard mode Select widget + migration notice removed — D3) | AC-SPR-025 | MUST |
| REQ-SPR-024 (mode + preset i18n keys removed — D6) | AC-SPR-026 | MUST |
| (D3 non-regression) Builder API preserved | AC-SPR-028 | MUST |
| (D4) test files compile-clean | AC-SPR-027 | MUST |
| (polish) wizard title | AC-SPR-022 | SHOULD |
| (polish) file rename | AC-SPR-023 | SHOULD |
| (polish) i18n catalog (now covered by AC-SPR-026 MUST) | AC-SPR-024 | SHOULD |

Total: **25 MUST + 3 SHOULD = 28 ACs** (iter-0 had 18 MUST + 3 SHOULD = 21 ACs;
iter-1 remediation added AC-SPR-025/026/027/028 as MUST + expanded
AC-SPR-002/005/007/011/014/018 to cover D1/D2/D3/D5/D7). REQ coverage 100%
(24/24 requirements mapped to ≥1 AC; the 3 SHOULD ACs map to spec.md §F open
questions / §E exclusions where applicable). Note: AC-SPR-024 (SHOULD i18n
catalog) is now largely subsumed by AC-SPR-026 (MUST mode+preset i18n removal);
it remains as SHOULD for any residual orphaned catalog entries not covered by
the MUST AC.

---

## §F. Quality Gates (Definition of Done)

A milestone is "done" when ALL of:

1. **The ACs in scope for that milestone** are verified PASS with verbatim
   command output pasted into `progress.md §E.2`.
2. **`go build ./...`** is green on darwin/amd64.
3. **`GOOS=linux GOARCH=amd64 go build ./...`** is green.
4. **`GOOS=windows GOARCH=amd64 go build ./...`** is green.
5. **`go vet ./...`** is clean.
6. **No new lint findings** on the modified packages (`golangci-lint run`).
7. **No `AskUserQuestion` / `mcp__askuser` references** introduced in modified
   files (subagent boundary preserved).
8. **Template neutrality audit PASS** (if `internal/template/templates/` was
   touched).
9. **docs-site i18n check PASS** (if docs were touched).

The SPEC is "complete" when ALL milestones M1-M6 are done AND:
- Full `go test ./...` green.
- Full `go test -race ./...` green on affected packages.
- `progress.md §E.2` populated with verbatim evidence.
- `progress.md §E.3` audit-ready signal emitted.
- Sync-phase (manager-docs) and Mx-phase close complete.

---

## §G. Edge Cases (must be handled or explicitly documented)

1. **Empty statusline.yaml** (file exists, only `statusline: {}`) — loader
   returns nil/zero config; runtime falls back to all-enabled segments. Must
   not error.
2. **statusline.yaml with `preset:` only, no `segments:`** (legacy install) —
   loader silently ignores preset; runtime falls back to all-enabled. Document
   in CHANGELOG as a one-time reconfiguration expectation.
3. **statusline.yaml with both `preset:` and `segments:`** — segments wins
   (preset ignored). Characterization test (AC-SPR-020) covers this.
4. **Corrupt statusline.yaml** (invalid YAML) — loader returns nil silently
   (existing behavior). No change.
5. **Web console rendered after upgrade** — the statusline fieldset is absent;
   adjacent fieldsets must not visually break. Manual smoke-test in M3.
6. **Wizard invoked on a profile with a legacy `statusline_preset` in
   preferences.yaml** — the legacy field is ignored by the new
   `ProfilePreferences` struct (Go YAML unmarshal tolerates unknown keys by
   default). No migration step.

---

## §H. Forward-Looking Checks (post-merge sanity, not blocking)

1. **Release CHANGELOG** entry documenting the retirement + the one-time
   reconfiguration expectation for users who relied on `compact`/`minimal`.
2. **README** statusline section (if it mentions presets) updated.
3. **One release later**, consider removing the silent-ignore tolerance for
   `preset:` in the YAML loader (emit a deprecation warning first). Track as
   a follow-up SPEC, not this one.

---

## §I. Closure Gates

This SPEC is ready for `implemented → completed` transition when:

- [ ] All 25 MUST ACs verified PASS with pasted evidence
- [ ] 3 SHOULD ACs either PASS or documented as accepted-debt
- [ ] sync-phase complete (CHANGELOG, frontmatter `implemented`)
- [ ] sync-auditor PASS (4-dimension score ≥ threshold)
- [ ] Mx-phase audit-ready signal emitted (§E.5 of progress.md)
- [ ] No MUST-FIX findings from `moai spec lint`

---

## Out of Scope (acceptance-phase)

The following are explicitly out of scope for acceptance verification:

### 1. Out of Scope — Builder API non-regression (D3)

- The `internal/statusline.StatuslineMode` type, its constants
  (`ModeDefault`/`ModeFull`/`ModeMinimal`/`ModeCompact`/`ModeVerbose`), and
  `NormalizeMode` are PRESERVED as part of the Builder API (AC-SPR-028 verifies
  their survival). Only the prefs/wizard/docs mode surface is retired. No AC
  in this file requires removal of these in-package Builder symbols.

### 2. Out of Scope — segment rendering internals

- AC-SPR-019 verifies byte-parity of segment rendering but does NOT require
  any modification to `renderer.go` render branches, `git.go`, `usage.go`,
  `task.go`, `memory.go`, `metrics.go`, `model_cache.go`, `version.go`,
  `theme.go`, `gradient.go`. These files are explicitly out of scope
  (spec.md §E.6, §A.3).

### 3. Out of Scope — post-merge sanity

- The forward-looking checks in §H (CHANGELOG, README, future deprecation
  warning) are post-merge sanity items, NOT acceptance gates. They do not
  block the `implemented → completed` transition.
