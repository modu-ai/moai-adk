# Progress — SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit>_

plan_status: audit-ready
plan_complete_at: 2026-08-03
tier: L
artifact_count: 6 (spec.md, plan.md, acceptance.md, design.md, research.md, progress.md)
era: V3R6
depends_on:
  - SPEC-UPDATE-YAML-PRESERVE-001
iter1_audit:
  verdict: FAIL
  score: 0.68
  threshold: 0.85
  defects:
    - D1 BLOCKING — Tier L artifact contract unmet → authored design.md + research.md
    - D2 SHOULD-FIX/blocking — AC-TBS-010 inverted intermediate assertion for snapshot.go → repaired to file-existence form
    - D3 SHOULD-FIX/blocking — quality.yaml AC cross-ref (AC-TBS-006 → AC-TBS-013) → corrected in spec.md + plan.md (D8 + M4)
    - D4 SHOULD-FIX/blocking — runUpdateRestore third restore-completion site → dispositioned in plan.md §B + Decision D4 + §C pre-flight check 5 + M3 wiring
    - D5 MINOR — basePath line citation (restore.go:107-108 → restore.go:118) → corrected in spec.md + plan.md
    - D6 MINOR — REQ-TBS-001/002 subject/trigger mismatch → rephrased subject to "the moai snapshot subsystem"
iter2_audit:
  verdict: PASS
  score: 1.00
  threshold: 0.85

## §E.2 Run-phase Evidence

### Pre-flight (M0)

- Prior SPEC merged: `3ced5b152 fix(SPEC-UPDATE-YAML-PRESERVE-001): ... (#1313)` — confirmed.
- Baseline backup suite green: `ok github.com/modu-ai/moai-adk/internal/cli/update/backup 3.197s`.
- M0 coverage baseline: **89.6%** of statements.
- Clean-step walk radius verified — `grep "filepath.Walk\|os.RemoveAll\|os.Remove" internal/cli/update_cleanup.go internal/cli/update_clean_install.go` shows no `.moai/cache/` deletion site; `grep "cache\|CacheSubdir"` returns no matches.
- Restore-completion sites: 3 production sites confirmed (`update_template_sync.go:420`, `update_clean_install.go:451`, `update_restore.go:53`).
- Snapshot location gitignored: `.gitignore:273 .moai/cache/` confirmed; `git check-ignore` exits 0.

### M1 — Snapshot package skeleton (RED → GREEN)

- RED (compile failure): `undefined: WriteSnapshot`, `SnapshotDir`, `HasSnapshot` (snapshot_test.go references symbols not yet defined).
- Implementation files created: `internal/cli/update/backup/snapshot.go` (SnapshotSubdir const, SnapshotDir, WriteSnapshot, HasSnapshot).
- Test file: `internal/cli/update/backup/snapshot_test.go`.
- GREEN: `go test ./internal/cli/update/backup/ -run 'TestWriteSnapshot|TestHasSnapshot|TestSnapshotDir|TestSnapshot_Carries' -count=1` → 6 PASS.

### M2 — SaveTemplateBase base-loader (RED → GREEN)

- RED (compile failure): `undefined: SaveTemplateBase`.
- Implementation files created: `internal/cli/update/backup/base_loader.go` (SaveTemplateBase — prefers snapshot, falls back to SaveTemplateDefaults).
- Edit: `internal/cli/update/backup/backup.go` — `BackupMoaiConfig` switched from `SaveTemplateDefaults(templateDefaultsDir)` to `SaveTemplateBase(templateDefaultsDir, projectRoot)`.
- Test file: `internal/cli/update/backup/base_loader_test.go`.
- GREEN: `go test ./internal/cli/update/backup/ -run 'TestSaveTemplateBase' -count=1` → 4 PASS. `go build ./...` clean.

### M3 — Write-time hook points + clean-step survival

- New helper: `internal/cli/update_snapshot_hook.go` — `writeTemplateSnapshotBestEffort(projectRoot, errOut)` (best-effort, swallows errors per REQ-TBS-014).
- Wired `writeTemplateSnapshotBestEffort` at all 4 Decision D4 trigger sites:
  - `internal/cli/init.go` (before `return nil`) — REQ-TBS-001.
  - `internal/cli/update_template_sync.go` (after RestoreMoaiConfig) — REQ-TBS-002 site 1.
  - `internal/cli/update_clean_install.go` (after RestoreMoaiConfig) — REQ-TBS-002 site 2.
  - `internal/cli/update_restore.go` (after RestoreFromBackupDir) — REQ-TBS-002 site 3 (lockout-escape).
- Test files: `internal/cli/update/backup/snapshot_survival_test.go`, `internal/cli/update_snapshot_hook_test.go`.
- GREEN: survival + wiring tests pass. `go build ./...` clean.

### M4 — Provenance + quality.yaml correctness tests

- Test file: `internal/cli/update/backup/snapshot_provenance_test.go`.
- Tests: TestMerge_BaseFromSnapshot_NotFromEmbedded (load-bearing, AC-TBS-010 target), TestMerge_AdoptsNewTemplateValue_WhenSnapshotMatchesLocal (AC-TBS-011), TestMerge_PreservesUserCustomization_WhenSnapshotDiffersFromLocal (AC-TBS-012), TestMerge_QualityYaml_Real3Way_WithSnapshot (AC-TBS-013), TestRestore_2WayFallback_UnparseableBase (AC-TBS-009), TestMergeYAML3Way_SignatureUnchanged (AC-TBS-008), TestBackupMoaiConfig_PrefersSnapshot (AC-TBS-006a).
- Design note: provenance tests use NON-system fields (test_coverage_target, enforce_quality), NOT `version`, because `version`/`template_version` are systemFields (node_merge.go) that always take the NEW value regardless of base — a provenance test on `version` is vacuous.
- GREEN: all provenance tests pass against the correct (snapshot-sourced) base.

### M5 — Falsifiability verification (AC-TBS-010, the non-vacuity gate)

Procedure: revert ONLY the `backup.go` edit (SaveTemplateBase → SaveTemplateDefaults) so the wrong-base path is active while snapshot.go/base_loader.go remain intact. The provenance test `TestMerge_BaseFromSnapshot_NotFromEmbedded` is self-contained (uses only BackupMoaiConfig + MergeYAML3Way + os/filepath, no references to snapshot.go/base_loader.go symbols), so it compiles and runs against the wrong base.

RED observation (wrong-base, SaveTemplateDefaults active):
```
$ go test ./internal/cli/update/backup/ -run TestMerge_BaseFromSnapshot_NotFromEmbedded -count=1 -v
--- FAIL: TestMerge_BaseFromSnapshot_NotFromEmbedded (0.01s)
    snapshot_provenance_test.go:132: TestMerge_BaseFromSnapshot_NotFromEmbedded (old==base → adopt NEW): key "test_coverage_target" = "80", want "85"
FAIL
```
The wrong base (embedded-raw `{{.TestCoverageTarget}}` placeholder) misread the rendered `80` as a user edit (`old != base`) and preserved the stale `80` instead of adopting the NEW `85`. Non-vacuous.

GREEN observation (correct base restored, SaveTemplateBase active):
```
$ go test ./internal/cli/update/backup/ -run TestMerge_BaseFromSnapshot_NotFromEmbedded -count=1 -v
--- PASS: TestMerge_BaseFromSnapshot_NotFromEmbedded (0.04s)
PASS
```

### M6 — Verification sweep

- Full backup suite: `ok github.com/modu-ai/moai-adk/internal/cli/update/backup 0.437s`.
- Coverage: **89.8%** (>= M0 baseline 89.6%). New file coverage: WriteSnapshot 91.4%, HasSnapshot 100%, SaveTemplateBase 84.6%.
- Cross-platform build: `go build ./...` OK; `GOOS=windows GOARCH=amd64 go build ./...` OK; `GOOS=linux GOARCH=amd64 go build ./...` OK.
- Lint: `golangci-lint run ./internal/cli/update/backup/... ./internal/cli/...` → 0 issues.
- Subagent boundary grep (AC-TBS-018): 0 matches for AskUserQuestion/mcp__askuser in internal/cli/update/backup/ (excluding tests/comments).
- Template neutrality (AC-TBS-019): 0 matches for SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT in internal/template/templates/.
- MergeYAML3Way signature (AC-TBS-008): `func MergeYAML3Way(newData, oldData, baseData []byte) ([]byte, error)` byte-identical.
- 2-way fallback (AC-TBS-009): `MergeYAMLDeep` present at restore.go:139.
- Snapshot scope (AC-TBS-016): snapshot.go walk root references defs.SectionsSubdir only.
- Gitignore (AC-TBS-005): `.gitignore:273 .moai/cache/`; `git check-ignore .moai/cache/template-snapshot/sections/system.yaml` exits 0.
- No new dependency (AC-TBS-020): no new require lines in go.mod/go.sum.
- Prior SPEC golden preservation suite (AC-TBS-022): 6 TestPreserve tests PASS, no regression.
- Full project suite (`go test ./... -count=1`): all packages `ok`, 0 FAIL.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-08-03
run_commit_sha: "64fef3362"
run_status: green
ac_pass_count: 23
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a (Route A main-direct disabled for this repo per repo-local-pr-policy.md; PR-mandatory)
l44_post_push_fetch: pending (post-commit)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
  linux_amd64: pass
total_run_phase_files: 10
  - internal/cli/update/backup/snapshot.go (new)
  - internal/cli/update/backup/snapshot_test.go (new)
  - internal/cli/update/backup/base_loader.go (new)
  - internal/cli/update/backup/base_loader_test.go (new)
  - internal/cli/update/backup/snapshot_provenance_test.go (new)
  - internal/cli/update/backup/snapshot_survival_test.go (new)
  - internal/cli/update/backup/backup.go (modified — SaveTemplateBase switch)
  - internal/cli/update_snapshot_hook.go (new)
  - internal/cli/update_snapshot_hook_test.go (new)
  - internal/cli/init.go (modified — WriteSnapshot hook)
  - internal/cli/update_template_sync.go (modified — WriteSnapshot hook)
  - internal/cli/update_clean_install.go (modified — WriteSnapshot hook)
  - internal/cli/update_restore.go (modified — WriteSnapshot hook)
m1_to_mN_commit_strategy: single-PR squash (Route B Tier L per repo-local-pr-policy.md)

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

sync_commit_sha: "pending-backfill-sync"
