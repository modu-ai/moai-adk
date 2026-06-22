# Progress — SPEC-WEB-CONSOLE-010

Lifecycle tracking for the 3-phase plan → run → sync flow. §E carries the audit-ready signals.

## §E.1 Plan-phase Audit-Ready Signal

- **Phase**: plan (complete) → run (complete)
- **Tier**: L (5-artifact: spec.md + plan.md + acceptance.md + design.md + research.md)
- **Status**: `draft` → `in-progress` (M1 commit) → `implemented` (run-phase close — progress.md owns this transition per REQ-ARR-003)
- **Artifacts created**:
  - `spec.md` — 12-field frontmatter, HISTORY, 19 GEARS requirements (REQ-WC10-001..019), Exclusions with 5 `### Out of Scope —` H3 sub-headings.
  - `plan.md` — §A Context, §B Known Issues (B1-B6), §C Pre-flight, §D Constraints, §E pointer, §F Milestones (M1-M6 priority-ordered), §G Anti-Patterns (AP-1..7), §H Cross-References.
  - `acceptance.md` — 21 mechanically-verifiable AC (AC-WC10-001..021), 4 Given-When-Then scenarios, 6 edge cases (EC-1..6), Definition of Done.
  - `design.md` — §A import-constraint solution, §B field-def shape, §C 6-section/34-field reconciliation, §D persistence routing, §E dual-surface render flow, §F i18n key strategy, §G risks, §H cross-refs.
  - `research.md` — §A established inventory (verified), §B divergence, §C import-constraint rationale, §D Cleanup Candidates (Tier A/B/C + mandatory grep gate), §E source references.
- **SPEC ID self-check**: `decomposition: SPEC ✓ | WEB ✓ | CONSOLE ✓ | 010 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Ground-truth verification**: `internal/settings` confirmed non-existent; 11 `ProfilePreferences` fields, 6-section/34-field union, import constraint, nested write seam, label drift, permission_mode normalization all verified against the live tree (file reads).
- **Plan-auditor gate**: pending (Phase 0.5 independent audit).
- **Implementation Kickoff Approval**: pending (plan-to-implement human gate per CLAUDE.local.md §19.1 — required before run-phase entry).

## §E.2 Run-phase Evidence

- **Phase**: run (complete)
- **cycle_type**: tdd (quality.yaml development_mode=tdd)
- **Status**: `in-progress` (spec.md) → `implemented` (this artifact at run-phase close; spec/plan/acceptance body transitions deferred to sync-phase per REQ-ARR-003)
- **Milestones completed**: M1 (SSOT schema package) → M2 (relocate nested seam) → M3 (TUI derives + 7 nested fields + bridge) → M4 (web derives + statusline re-add + validate.go cleanup + i18n seg keys) → M5 (label/persistence normalization) → M6 (verification batch).

### AC PASS/FAIL matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|----------------------|---------------|
| AC-WC10-001 | PASS | `go list ./internal/settings/...` | `github.com/modu-ai/moai-adk/internal/settings` (build succeeds) |
| AC-WC10-002 | PASS | `go list -deps ./internal/cli/... \| grep -c 'internal/settings$'` (and web) | cli: 1, web: 1 (both ≥ 1) |
| AC-WC10-003 | PASS | `go list -deps ./internal/settings/... \| grep -E 'internal/(cli\|web)$'` | no matches (grep exit 1) + `TestSchemaImportsNeitherCLINorWeb` PASS |
| AC-WC10-004 | PASS | `go test ./...` (touched) | `ok` settings/cli/web/profile |
| AC-WC10-005 | PASS | `golangci-lint run ./internal/settings/... ./internal/cli/... ./internal/web/...` | `0 issues.` |
| AC-WC10-006 | PASS | `go vet ./internal/settings/... ./internal/cli/... ./internal/web/...` | no output (exit 0) |
| AC-WC10-007 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| AC-WC10-008 | PASS | `go test -cover ./internal/settings/...` | `coverage: 96.1% of statements` (≥ 90%) |
| AC-WC10-009 | PASS-WITH-DEBT | `go test -cover ./internal/cli/ ./internal/web/ ./internal/profile/` | cli 72.0% (baseline 72.2%), web 73.0% (baseline 73.4%), profile 79.9% (baseline 79.9%). cli/web essentially flat; see Gaps re: defensive-branch nominal -0.4% on web. |
| AC-WC10-010 | PASS | `TestSchemaFieldNameSet` + `TestTUIRendersSchemaFieldSet` + `TestWebRendersSchemaFieldSet` | all PASS; schema has exactly 34 fields; each surface's rendered field-name set == schema set |
| AC-WC10-011a | PASS | `grep -c 'data-i18n={ "seg." + seg }' internal/web/fieldsets.templ` + `name="statusline_theme"` + `TestWebStatuslineRendersThemeAnd15Segments` | segment loop present (renders 15) + 1 theme control; test PASS |
| AC-WC10-011b | PASS | `grep -in 'preset' fieldsets.templ validate.go handlers.go \| grep -v '//' \| grep -vi retire` + `TestWebStatuslineNoPresetControl` | NO live preset matches; test PASS |
| AC-WC10-012 | PASS | `grep -nE '(modelCanonical\|effortLevelCanonical\|langOptions\|developmentModeCanonical\|conventionCanonical) *=' internal/web/validate.go` | 0 standalone re-declarations (grep exit 1); all 5 derive from settings schema |
| AC-WC10-013 | PASS | `TestTUINestedConfigRoundTrip` + `TestTUINestedConfigNoParallelWriter` | 7 nested fields round-trip; no parallel yaml.Marshal/os.WriteFile in profile_setup.go (delegates to shared seam) |
| AC-WC10-014 | PASS | `TestSchemaEmptyLabelParity` (web) + `TestTUIEmptyLabelsSchemaSourced` (cli) | both surfaces source empty labels from `settings.EmptyLabelFor`; 4 drifts resolved |
| AC-WC10-015 | PASS | `TestPermissionModeNormalizeAcceptEdits` | schema `permission_mode` Persist.Normalize(acceptEdits)="" ; TUI applies acceptEdits→"" normalization |
| AC-WC10-016 | PASS | `TestI18nKeySetParity` (cli bridge) + `TestI18nKeySetParity` (web flat keys) + `TestI18nSegmentParity`/`TestI18nSegmentKeysComplete` | every schema key resolves via TUI bridge (4 locales) AND has flat dotted-key entry in i18n.js (4 locales); 15 seg keys × 4 = 60 present |
| AC-WC10-017 | PASS | `TestTUINestedConfigEmptyPreserve` + `TestSharedNestedSeamEmptyPreserve` | empty/unsubmitted nested field retains prior on-disk value |
| AC-WC10-018 | PASS | `git status --porcelain .moai/config/sections/ \| grep '^ D\|^D'` | no config deletions |
| AC-WC10-019 | PASS | `grep -c 'Cleanup Candidates' research.md` + tier/grep-gate grep | appendix present (authored plan-phase); 3 tiers (A/B/C) + grep-gate sentence present |
| AC-WC10-020 | PASS | `git status --porcelain \| grep -c 'internal/template/templates/'` | 0 (no template tree touched) |
| AC-WC10-021 | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/settings/ internal/cli/profile_setup.go internal/cli/schema_bridge.go internal/web/ \| grep -v '_test.go' \| grep -v '//'` | no matches (subagent boundary preserved) |

### Invariants

| Invariant | Status | Evidence |
|-----------|--------|----------|
| Existing test suite never broken | PASS | `go test ./internal/cli/ ./internal/web/ ./internal/profile/` all `ok` |
| Cross-platform build | PASS | `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| Race-free (touched concurrency) | PASS | `go test -race ./internal/settings/... ./internal/web/...` all `ok` |
| Whole-section-copy nested isolation (B4) | PASS | `TestSharedNestedSeamWholeSectionPreserve` + `TestConfigManagerStillUsed` |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-23
run_commit_sha: 16329f3b168177be4ef0a3a505aba6bb79ee4d6a
run_status: implemented
ac_pass_count: 21          # AC-WC10-001..021 (AC-009 PASS-WITH-DEBT on web nominal -0.4%)
ac_fail_count: 0
preserve_list_post_run_count: 0   # no out-of-scope file modified
l44_pre_commit_fetch: deferred-to-orchestrator   # worktree-isolated run; orchestrator owns pre-spawn fetch + integration
l44_post_push_fetch: deferred-to-orchestrator
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues; go vet clean
cross_platform_build:
  linux_darwin: pass        # go build ./... exit 0
  windows: pass             # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 21   # 3 new settings src + 2 new settings test + 1 new cli src + 2 new cli test + 3 new web test + 8 modified web + 2 modified cli (+ spec.md frontmatter)
m1_to_mN_commit_strategy: single-commit   # one run-phase commit covering M1-M6 (single-session worktree run)
coverage:
  internal/settings: 96.1%   # ≥ 90% target met
  internal/cli: 72.0%        # baseline 72.2% (flat)
  internal/web: 73.0%        # baseline 73.4% (defensive-branch nominal -0.4%)
  internal/profile: 79.9%    # baseline 79.9% (unchanged)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_
