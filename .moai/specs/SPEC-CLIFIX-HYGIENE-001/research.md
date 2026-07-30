# SPEC-CLIFIX-HYGIENE-001 — Research: Dead-Code Caller-Graph Inventory

Method (evidence re-verified 2026-07-30 against the live worktree HEAD = origin/main, post P0-P3 merge): per symbol, `grep -rn '<symbol>' internal/cli cmd --include='*.go'` filtered to exclude `_test.go` and the defining file, so each row records cross-file PRODUCTION callers. Line anchors drift; re-run the greps at run time before deleting. Dispositions: DELETE / EXCLUDE-P0 / EXCLUDE-KEEP / WIRE-DECISION / DROPPED.

## §A Per-Symbol Inventory (REQ-HYG-001-002 input)

| Symbol (definition site — re-verify at run) | Production callers found (outside defining file, non-test) | Disposition |
|---|---|---|
| `buildGLMEnvVars` (`internal/cli/glm.go:917`) | 0 — callers are test-only (`glm_team_test.go`, `coverage_improvement_test.go`) | DELETE (together with its now-orphaned tests) |
| `ttyConfirmer` (`internal/cli/branch_protection.go:39-43`) | 0 — carries `nolint:unused` annotation: "SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (paired with ttyConfirmer)" | DELETE only after resolving the deferred pairing in SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 (check before removal) |
| `acquireUpdateLock` (`internal/cli/update_cleanup.go:55`) | LIVE — called from `internal/cli/update.go:264` (P0 lock wiring landed) plus test callers | EXCLUDE-P0 / EXCLUDE-KEEP — never delete |
| `cleanStaleLock` (`internal/cli/update_cleanup.go:92`) | called by `acquireUpdateLock` (`update_cleanup.go:59`) | EXCLUDE-P0 / EXCLUDE-KEEP |
| `scanDeprecatedPaths` (`internal/cli/update_cleanup.go:119`) | LIVE — `update_clean_install.go` (lines 181, 229 per v0.1.0; re-verify at run) | EXCLUDE-KEEP (audit's "update_cleanup.go entirely dead" claim narrows here) |
| `backupDeprecatedPaths` (`internal/cli/update_cleanup.go:195`) | 0 code callers — comment-only reference at `update_clean_install.go:12`; audit cluster-1 Major requires WIRING it into clean-install (backup before RemoveAll) | WIRE-DECISION — not in the deletion inventory; wiring is the audit-preferred outcome |
| `filterSkipMarkerPaths`, `manifest2hashBytes`, `inspectDeprecatedPath`, `removeDeprecatedFile`, `classifyDeprecatedFile`, `probeCaseSensitiveFS`, `emitCleanupTelemetry` (internal sub-cluster in `update_cleanup.go`) | 0 callers outside `update_cleanup.go` — reachable only through the `backupDeprecatedPaths`/`removeDeprecatedFile` roots | Fate follows the `backupDeprecatedPaths` WIRE-DECISION: after wiring, delete only what remains unreachable |
| `worktree_validation.go` — `WorktreePathInvalidError` (+Error/Is), `validateWorktreeReturn` | 0 non-test callers outside the file | DELETE candidate (whole file) |
| `cleanup_old_backups` (in `update.go`, re-verify line at run) | LIVE — called from the Backup step in `update.go` (re-verify line at run) | EXCLUDE-KEEP (removed from the deletion inventory) |

### v0.1.0 rows DROPPED in v0.2.0 (no longer actionable)

| v0.1.0 row | Why dropped (v0.2.0 evidence) |
|---|---|
| `init_layout.go` — `renderInitHeader`, `renderInitNextSteps` | file does not exist (`find internal -name 'init_layout*'` → 0 results, 2026-07-30). Either P0-P3 deleted it or it was renamed; either way the row is stale. |
| `update.go:618-717` Backup/Restore step closures | the step-closure area has shifted (Backup step lives around `update.go:777-940` now); the audit's "unreachable" claim requires fresh trace at run time. Not in the v0.2.0 deletion inventory unless re-confirmed unreachable. |
| `update.go:1596-1624` `excludedDirs` (always empty) | symbol absent from current `update.go` (`grep excludedDirs internal/cli/update.go` → 0 matches). Already removed. |
| `update.go:1069-1073` manifest load (loaded, never used) | symbol absent (`grep 'loadManifest\|manifest load' internal/cli/update.go` → 0 matches). Already removed. |
| `update.go:730-737` parameters (unused) | line range requires fresh `go vet`/unused-parameter confirmation at run time; not a standalone deletion candidate. |
| v0.1.0 REQ-HYG-001-007 / AC-HYG-001-007 target: `deepMerge3Way` at `update.go:2396-2452` | past EOF (file is 1,905 lines); merge machinery extracted to `internal/merge/`; `deepMergeMap` at `strategies.go:360` preserves user-added keys (line 388). Defect already fixed — see spec.md HISTORY v0.2.0. |
| v0.1.0 REQ-HYG-001-008 / AC-HYG-001-008 target: token write at `update.go:2641-2651` (0644) | past EOF; superseded by F1 security fix at `update.go:1375-1438` — tokens intentionally NOT persisted. Requirement already satisfied by design — see spec.md HISTORY v0.2.0. |

## §B Audit-Claimed, Not Yet Independently Re-Verified (RE-VERIFY at run)

These items survive into v0.2.0 but still require a fresh verification step in milestone M5 before deletion. Each is tagged with its current-tree status:

| Item | Audit claim | Current-tree status (2026-07-30) | Verification method at run |
|---|---|---|---|
| `update.go` Backup/Restore step reachability | Backup closure never invoked; Restore is a no-op stub | step-closure area has shifted to ~`update.go:777-940`; the F1 security redesign and P0 lock wiring both touched this region | trace the steps-table execution path; add a characterization test proving non-invocation before removal |
| `backupDeprecatedPaths` wiring | audit cluster-1 Major requires wiring it into clean-install | still 0 production callers; WIRE-DECISION unchanged | decide at Implementation Kickoff: wire (preferred) vs delete the sub-cluster |

## §C Net-Delta Honesty Note

The v0.1.0 audit's ~500-line dead-code estimate rested on "update_cleanup.go 전체 (~300라인)" being dead. The v0.2.0 caller-graph evidence re-confirmed on 2026-07-30 shows that file is NOT entirely dead: `scanDeprecatedPaths` is LIVE (clean-install callers), the lock pair is P0-wired and LIVE, and `backupDeprecatedPaths` + its sub-cluster are a WIRE-DECISION. The verified-deletable subset is:

- `buildGLMEnvVars` and its orphaned tests (~30 lines)
- `worktree_validation.go` whole file (~100 lines)
- `ttyConfirmer` (~10 lines) — only after the SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred-pairing is resolved

Realistic net delta: on the order of −150 to −250 lines, contingent on the `ttyConfirmer` deferred-pairing resolution and the `backupDeprecatedPaths` WIRE-DECISION outcome. acceptance.md §D.5 carries this honest figure (replacing the v0.1.0 "≈ −500 lines" target, which was optimistic even back then — `init_layout.go` is now confirmed absent and `worktree_validation.go` is the only confirmed whole-file delete).

## §D Cross-References

- design.md §A (seam map — deletions land before/alongside M5 decomposition per plan.md §F).
- plan.md §D constraint: no deletion of hook-registered wrappers or build-tag-gated symbols without per-GOOS build verification.
- Audit SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 cluster 1 Minor (dead code rows), cluster 2 Minor (`buildGLMEnvVars`), cluster 4 Minor (`worktree_validation`/`ttyConfirmer`).
- v0.2.0 rescope driver: `.moai/reports/plan-audit/SPEC-CLIFIX-HYGIENE-001-review-1.md` (iter-1 FAIL 0.58, D1-D11).
