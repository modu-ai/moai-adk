# Acceptance Criteria — SPEC-WEB-CONSOLE-010

> Every criterion is mechanically verifiable (a command, grep, test, or build with a deterministic pass/fail). Verbatim command + expected output is the load-bearing artifact (verification-claim-integrity §3.2). Paths and line anchors are plan-phase references; run-phase re-derives them against the live tree.

## §A Quality gate criteria

| ID | Criterion | Mechanical verification | Pass condition |
|----|-----------|-------------------------|----------------|
| AC-WC10-001 | New SSOT package exists | `go list ./internal/settings/...` | package(s) listed, build succeeds |
| AC-WC10-002 | SSOT imported by both surfaces | `go list -deps ./internal/cli/... \| grep -c 'internal/settings'` AND same for `./internal/web/...` | both ≥ 1 |
| AC-WC10-003 | No reverse import (cli ⇎ web stays clean; settings imports neither) | `go list -deps ./internal/settings/... \| grep -E 'internal/(cli\|web)$'` | no matches |
| AC-WC10-004 | All tests pass | `go test ./...` | `ok` for all packages, 0 FAIL |
| AC-WC10-005 | Lint clean | `golangci-lint run --timeout=2m` | 0 issues |
| AC-WC10-006 | Vet clean | `go vet ./...` | no output |
| AC-WC10-007 | Cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| AC-WC10-008 | New package coverage | `go test -cover ./internal/settings/...` | ≥ 90% |
| AC-WC10-009 | Touched-package coverage non-regressing | `go test -cover ./internal/cli/... ./internal/web/... ./internal/profile/...` | each ≥ prior baseline (captured §C pre-flight) |

## §B Functional criteria

| ID | Criterion | Mechanical verification | Pass condition |
|----|-----------|-------------------------|----------------|
| AC-WC10-010 | Both surfaces render all 34 canonical fields | `TestSchemaFieldNameSet` asserts the schema enumerates exactly the 34-field name set across 6 sections; `TestTUIRendersSchemaFieldSet` + `TestWebRendersSchemaFieldSet` assert the set of field NAMES each surface renders equals the schema's 34-field-name set | schema field-name set has 34 entries; each surface's rendered field-name set == the schema set (set equality, not merely count==34) |
| AC-WC10-011a | Web statusline fieldset re-added with theme + 15 segments | `grep -c 'data-i18n="seg\.' internal/web/fieldsets.templ` (or the schema-driven segment loop) | 15 segment controls + 1 theme control present |
| AC-WC10-011b | Web statusline has NO preset selector | `grep -in 'preset' internal/web/fieldsets.templ internal/web/validate.go internal/web/handlers.go` excluding comment lines that document the retirement | no live `preset` form control / option list / validation rule |
| AC-WC10-012 | No duplicated canonical option list in web | `grep -nE '(modelCanonical\|effortLevelCanonical\|langOptions\|developmentModeCanonical\|conventionCanonical) *=' internal/web/validate.go` — all 5 currently-re-declared lists must become schema-sourced (REQ-WC10-004) | 0 standalone re-declarations (all 5 lists derive from the schema, not hand-declared in `internal/web/validate.go`) |
| AC-WC10-013 | TUI persists the 7 nested fields via the shared nested seam | `TestTUINestedConfigRoundTrip` writes all 7 nested fields through the TUI persistence path and reads them back from `quality.yaml` / `git-convention.yaml`; grep confirms `internal/cli/profile_setup.go` calls the shared nested write seam (the relocated `writeProjectNestedConfig` equivalent) | 7 nested fields round-trip; no parallel `yaml.Marshal`/`os.WriteFile` in `profile_setup.go` |
| AC-WC10-014 | Per-field canonical empty-option label single-sourced | `TestSchemaEmptyLabelParity` asserts the schema returns one empty-option label per field AND both surfaces render that exact label (no "(not set)"/"(unset)" or "(project default)"/"(unchanged)" divergence) | identical label strings on both surfaces per field |
| AC-WC10-015 | permission_mode normalization preserved | `TestPermissionModeNormalizeAcceptEdits` submits `acceptEdits` via each surface → persisted profile YAML has empty `permission_mode` (no redundant override) | `permission_mode` absent/empty after acceptEdits submit |
| AC-WC10-016 | i18n key set shared | `TestI18nKeySetParity` asserts every schema i18n key resolves through the TUI bridge resolver into a non-empty `getProfileText(locale).<field>` value for all 4 locales AND has a matching flat `f.`/`seg.`/`sec.`/`count.`-prefixed entry in `window.MOAI_I18N` (`assets/i18n.js`) for all 4 locales | every schema key resolves in BOTH stores for en/ko/ja/zh (TUI via the named bridge; web via the flat dotted key) |
| AC-WC10-017 | Empty=preserve on nested fields | `TestTUINestedConfigEmptyPreserve` submits a nested field empty via the TUI path → on-disk value unchanged | unsubmitted nested field retains prior value |

## §C Scope-boundary criteria (negative / no-deletion)

| ID | Criterion | Mechanical verification | Pass condition |
|----|-----------|-------------------------|----------------|
| AC-WC10-018 | No config field/section/file deleted | `git diff --stat` over `.moai/config/sections/` + a grep that no Tier B/C key was removed from any `*.yaml` | 0 deletions of config keys/sections/files |
| AC-WC10-019 | Cleanup Candidates appendix present with grep gate | `grep -c 'Cleanup Candidates' .moai/specs/SPEC-WEB-CONSOLE-010/research.md` AND the appendix states the mandatory `.claude/{skills,rules,agents,workflows}` grep gate before any removal | appendix present; 3 tiers (A/B/C) documented; grep-gate sentence present |
| AC-WC10-020 | No template tree touched | `git diff --name-only \| grep -c 'internal/template/templates/'` | 0 |
| AC-WC10-021 | Subagent boundary preserved | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/settings/ internal/cli/profile_setup.go internal/web/ \| grep -v '_test.go' \| grep -v '^[^:]*:[0-9]*:[ \t]*//'` | no matches |

## §D Given-When-Then scenarios

### Scenario 1 — TUI nested project-config field round-trips through the shared seam

- **Given** a MoAI project with `.moai/config/sections/quality.yaml` and `git-convention.yaml` present, and the `internal/settings` schema declaring the 7 nested fields with persistence targets.
- **When** a user runs `moai profile setup`, sets `quality.test_coverage_target = 92` and `git_convention.auto_detection.confidence_threshold = 0.8`, leaves all other nested fields untouched, and completes the wizard.
- **Then** `quality.yaml` shows `test_coverage_target: 92`, `git-convention.yaml` shows `confidence_threshold: 0.8`, every other field in both sections is byte-unchanged from before, and the persistence went through the shared nested write seam (no `yaml.Marshal`/`os.WriteFile` in `profile_setup.go`).

### Scenario 2 — Web statusline re-added without the retired preset

- **Given** the web console rendering from the shared schema, and `SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001` having retired the `preset` control.
- **When** the console page is rendered (`GET /`) and inspected.
- **Then** the Statusline section is present with a theme `<select>` and 15 segment toggle controls sourced from the schema, AND no `preset` form control, `preset` option list, or `preset` validation rule appears anywhere in `fieldsets.templ` / `validate.go` / `handlers.go` (live, non-comment).

### Scenario 3 — Single-sourced option list, no drift

- **Given** the `internal/settings` schema holding the model option list exactly once.
- **When** the model option list is needed by the TUI Select widget AND the web `<select>` AND `internal/web/validate.go` validation.
- **Then** all three derive from the same schema list; `internal/web/validate.go` no longer declares its own `modelCanonical`; changing the schema list changes all three consumers with no hand-sync.

### Scenario 4 — Canonical empty-option label parity

- **Given** the schema declaring one empty-option label per field (e.g. `model` → "(project default)").
- **When** the TUI renders the model Select and the web renders the model `<select>`.
- **Then** both render the identical empty-option label for `model`, and likewise for language, effort_level, and git_convention — eliminating the four documented label drifts.

## §E Edge cases

- **EC-1** — Outside a MoAI project (no `.moai` dir): the TUI's nested project-config persistence is a no-op (matches the existing `persistProjectConfig` guard at `profile_setup.go:236-242`); the wizard still completes and writes the profile store.
- **EC-2** — Empty submission on a nested field: empty = preserve (REQ-WC10-012); the on-disk value is unchanged.
- **EC-3** — `permission_mode = acceptEdits`: normalized to empty so no redundant override is persisted (REQ-WC10-014), on BOTH surfaces.
- **EC-4** — Non-numeric nested input (e.g. `test_coverage_target = "abc"` on the web): the existing `ParseErrs` type-conversion guard rejects it; the schema-derived validation does not regress this guard.
- **EC-5** — Legacy `preset:` / `statusline_mode:` key in an existing `statusline.yaml` or `preferences.yaml`: silently ignored on unmarshal (unknown YAML keys do not error) — the re-add does not resurrect a `preset` write path.
- **EC-6** — `templ generate` not re-run after `.templ` edit: `go build` fails or renders stale; M4 must re-generate. (Caught by AC-WC10-004/007.)

## §F Definition of Done

- [ ] AC-WC10-001..021 all PASS with cited verbatim command output.
- [ ] `internal/settings` package created, imported by both surfaces, importing neither (AC-WC10-002/003).
- [ ] Both surfaces render all 34 canonical fields; web statusline = theme + 15 segments, NO preset (AC-WC10-010/011a/011b).
- [ ] `internal/web/validate.go` has no duplicated canonical option list (AC-WC10-012).
- [ ] TUI persists the 7 nested fields through the shared seam (AC-WC10-013).
- [ ] Per-field canonical empty-option labels single-sourced + identical across surfaces (AC-WC10-014).
- [ ] i18n key set shared; each store keeps its own format (AC-WC10-016).
- [ ] NO config deleted; Cleanup Candidates appendix present with grep gate (AC-WC10-018/019).
- [ ] No template tree touched; subagent boundary preserved (AC-WC10-020/021).
- [ ] `go test ./...`, `go vet ./...`, `golangci-lint run`, `GOOS=windows GOARCH=amd64 go build ./...` all green.
- [ ] `progress.md` §E run-phase evidence populated by manager-develop.
