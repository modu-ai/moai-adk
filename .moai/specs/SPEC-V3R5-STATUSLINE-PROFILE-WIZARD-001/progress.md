---
id: SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001
title: "Run-phase Progress — profile setup wizard statusline preset/segments"
version: "0.2.0"
created: 2026-05-22
updated: 2026-05-22
---

# Run-Phase Progress — SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001

## Summary

Tier S Quick Win — backend (preferences + sync) was already complete; only frontend wizard input UI was missing. Implementation completed orchestrator-direct (manager-develop Agent delegation failed with `WorktreeCreate hook returned a path that is not a directory: {}` — Claude Code runtime autonomous L1 isolation regression, identical to `cleanupBogusRootDir` lesson pattern; per CLAUDE.md §16 fallback to orchestrator-direct).

## Milestone Status

| Milestone | Status | Commits | Notes |
|-----------|--------|---------|-------|
| M1 translations + helper | DONE | (this commit) | 23 keys × 4 locales (92 cells) + `statuslinePresetCanonical` + `statuslineAllSegments` + `isCanonicalStatuslinePreset` + `normalizeStatuslinePreset` |
| M2 wizard Display + Segments group | DONE | (this commit) | Display group: + Preset Select (4 options). Section 5: MultiSelect group with `WithHideFunc(preset != "custom")` (R-SPW-001 mitigation A worked — no form-split fallback needed) |
| M3 regression tests | DONE | (this commit) | `TestNormalizeStatuslinePreset` (7 cases covering EC-SPW-001/002/003) + `TestStatuslineAllSegments_CardinalityAndOrder` + `TestProfileSetupTranslations_PresetSegments` (4 locale × 23 cells = 92 cells verified) |
| M4 final verification + commit | DONE | (this commit) | 7/7 ACs PASS / cross-platform PASS / lint NEW=0 in profile_setup*.go |

## AC PASS/FAIL Matrix

| AC | Status | Verification Command | Actual |
|----|--------|---------------------|--------|
| AC-SPW-001 | PASS | `grep -n 'StatuslinePresetTitle' internal/cli/profile_setup.go` | `352: Title(t.StatuslinePresetTitle)` |
| AC-SPW-002 | PASS | `grep -n 'huh.NewMultiSelect' internal/cli/profile_setup.go` | `376: huh.NewMultiSelect[string]()` + 50 NewOption (preset 4 + 15 segments included) |
| AC-SPW-003 | PASS | `go test -run TestProfileSetupTranslations_PresetSegments ./internal/cli/...` | PASS |
| AC-SPW-004 | PASS | `go test -run TestNormalizeStatuslinePreset ./internal/cli/...` | PASS |
| AC-SPW-005 | PASS | `go test -count=1 ./internal/cli/...` | PASS (cli pkg + sub-packages all ok) |
| AC-SPW-006 | PASS | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` + linux | darwin exit 0 / windows exit 0 / linux exit 0 |
| AC-SPW-007 | PASS | `git diff 02a0a8304 -- internal/profile/sync.go internal/profile/preferences.go` | 0 lines (backend byte-identical preserved) |

## Plan-Auditor SHOULD Findings — Run-Phase Absorption

- **S1** (stale SHA) — resolved inline before commit (plan.md line 245 `9e14c180...` → `02a0a8304...`).
- **S2** (LOC inconsistency) — resolved inline before commit (spec.md HISTORY `~130` → `~200-280`).
- **S3** (git diff reference-point) — AC-SPW-007 used explicit `git diff 02a0a8304` (pre-SPEC HEAD) instead of `git diff main`. PASS.
- **S4** (nil-map default rendering) — `internal/cli/profile_setup.go:212-227` documents explicit choice: nil map → all-15 segments enabled (matches statusline.yaml baseline, NOT 11-segment `defaultStatuslineSegments()`). Comment explicitly states rationale. Run-phase absorbed.
- **S5** (PRESET-FIX-001 dangling reference) — resolved inline before commit (spec.md "proposed, not yet filed").

## Out-of-Scope Baseline Failures (Pre-Existing, NOT This SPEC's Responsibility)

Full repo `go test ./...` shows these pre-existing baseline failures (verified absent in `internal/cli/...` scope where this SPEC operates):

- `internal/hook/` 2 (TestHookWrapper_*) — hook subsystem, unrelated
- `internal/lsp/subprocess/` 1 (TestSupervisor_NonZeroExit) — LSP, unrelated
- `internal/skills/` 2 (TestTemplateMirrorParity / TestSubSkillLOCCeiling) — template drift (memory project_v3r6_template_mirror_drift_audit_2026_05_22 documents 25-file drift)
- `internal/statusline/` 1 (TestRenderPRSegment_Absence) — statusline renderer baseline (memory project_v3r5_statusline_fmc_001_run_complete documents)
- `internal/template/` 3 (TestLateBranchTemplateMirror / TestRuleTemplateMirrorDrift / TestSkillsContainPlanAuditGateMarkers) — template drift

All pre-existing per memory `project_v3r5_statusline_fmc_001_run_complete` and `project_v3r6_template_mirror_drift_audit_2026_05_22`. Out of scope per spec.md §4.1.

## Coverage

```
internal/cli       67.2% of statements
internal/cli/pr    91.7%
internal/cli/wizard 37.1%
internal/cli/worktree 84.2%
```

(internal/cli baseline stable — new code paths exercised by 3 new tests.)

## Files Modified

- `internal/cli/profile_setup.go` — +50 LOC (helper + canonical slice + Preset Select + Segments MultiSelect group + segments map conversion)
- `internal/cli/profile_setup_translations.go` — +96 LOC (struct 23 fields + 4 locale × 23 entries = 92 cells)
- `internal/cli/profile_setup_test.go` — +43 LOC (TestNormalizeStatuslinePreset + TestStatuslineAllSegments_CardinalityAndOrder)
- `internal/cli/profile_setup_translations_test.go` — +49 LOC (TestProfileSetupTranslations_PresetSegments)

**Total LOC delta**: ~238 (estimated in plan.md §3.1 was ~200-280, on-target).

## Branch State

- Pre-run-phase HEAD: `56a618587` (after plan commit `b5b3354ed` + research commit `56a618587`)
- Post-run-phase HEAD: (this commit SHA — to be filled by commit step)
- Branch: `main` (Late-Branch policy, REQ-LB-005)
- Push status: deferred to sync-phase per REQ-LB-005

## Sync-Phase Handoff

Ready for `/moai sync SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001` (manager-docs):

- README/CHANGELOG update for v2.20.0-rc1 release notes
- Optional: docs-site `advanced/statusline.md` update mentioning preset/segments configurability via `moai profile setup`
- PR creation (squash) with batch sync of accumulated SPECs since last push
