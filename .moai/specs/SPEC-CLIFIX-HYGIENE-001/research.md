# SPEC-CLIFIX-HYGIENE-001 — Research: Dead-Code Caller-Graph Inventory

Method (evidence gathered 2026-07-10 against the live tree): per symbol, `grep -rn '<symbol>' internal/cli cmd --include='*.go'` filtered to exclude `_test.go` and the defining file, so each row records cross-file PRODUCTION callers. Line anchors drift; re-run the greps at run time before deleting. Dispositions: DELETE / EXCLUDE-P0 / WIRE-DECISION / LIVE-KEEP / RE-VERIFY.

## §A Per-Symbol Inventory (REQ-HYG-001-002 input)

| Symbol (definition site) | Production callers found (outside defining file, non-test) | Disposition |
|---|---|---|
| buildGLMEnvVars (glm.go:936) | 0 — callers are test-only (glm_team_test.go:144, coverage_improvement_test.go:5788) | DELETE (together with its now-orphaned tests) |
| ttyConfirmer (branch_protection.go:34-43) | 0 — carries `nolint:unused` annotation: "SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 deferred (paired with ttyConfirmer)" | DELETE only after resolving the deferred pairing in SPEC-V3R6-CI-BASELINE-DRIFT-001 §D.1 (check before removal) |
| acquireUpdateLock (update_cleanup.go:55) | 0 today (test callers only: update_e2e_test.go:230, update_cleanup_test.go) — CRITICAL-001 REQ-CRIT-001-005 wires it into runUpdate | EXCLUDE-P0 — never delete |
| cleanStaleLock (update_cleanup.go:85) | called by acquireUpdateLock | EXCLUDE-P0 |
| scanDeprecatedPaths (update_cleanup.go:119) | LIVE — update_clean_install.go:181, :229 | LIVE-KEEP (audit's "update_cleanup.go entirely dead" claim narrows here) |
| backupDeprecatedPaths (update_cleanup.go:195) | 0 code callers — comment-only reference at update_clean_install.go:12; audit cluster-1 Major requires WIRING it into clean-install (backup before RemoveAll) | WIRE-DECISION — not in the deletion inventory; wiring is the audit-preferred outcome |
| filterSkipMarkerPaths (:148), manifest2hashBytes (:260), inspectDeprecatedPath (:277), removeDeprecatedFile (:303), classifyDeprecatedFile (:335), probeCaseSensitiveFS (:379), emitCleanupTelemetry (:411) | 0 callers outside update_cleanup.go — internal sub-cluster reachable only through the backupDeprecatedPaths/removeDeprecatedFile roots | Fate follows the backupDeprecatedPaths WIRE-DECISION: after wiring, delete only what remains unreachable |
| worktree_validation.go — WorktreePathInvalidError (+Error/Is), validateWorktreeReturn | 0 non-test callers outside the file | DELETE candidate (whole file) |
| init_layout.go — renderInitHeader, renderInitNextSteps | 0 non-test callers outside the file | DELETE candidate; RE-VERIFY first (file mtime 2026-07-07 is recent — may be pending wiring; audit also notes its next-steps hint advertises the non-existent `moai plan` CLI command) |
| cleanup_old_backups (update.go:1953) | LIVE — update.go:838 | LIVE-KEEP (removed from the deletion inventory) |

## §B Audit-Claimed, Not Yet Independently Re-Verified (RE-VERIFY at run)

These rows come from the audit report; no fresh caller-graph evidence was gathered in this plan phase. Each requires its own verification step in milestone M5 before deletion:

| Item | Audit claim | Verification method at run |
|---|---|---|
| update.go:618-717 Backup/Restore step closures | Backup closure never invoked; Restore is a no-op stub (~90 unreachable lines) | trace the steps-table execution path; add a characterization test proving non-invocation before removal |
| update.go:1596-1624 excludedDirs | always empty | grep writers of excludedDirs; confirm zero population sites |
| update.go:1069-1073 manifest load | loaded, never used | verify the loaded value has no reads |
| update.go:730-737 parameters | unused parameters | `go vet` / unused-parameter linter confirmation |

## §C Net-Delta Honesty Note

The audit's ~500-line dead-code estimate included "update_cleanup.go 전체 (~300라인)". The caller-graph evidence above shows that file is NOT entirely dead: scanDeprecatedPaths is live, the lock pair is EXCLUDE-P0, and backupDeprecatedPaths + its sub-cluster are a WIRE-DECISION. The verified deletable subset is therefore smaller than 500 lines; acceptance.md §D.5 treats −500 LOC as a target, and the actual net delta plus the final per-symbol deletion list are reported as run-phase evidence in progress.md §E.2.

## §D Cross-References

- design.md §A (seam map — deletions land before/alongside M5 decomposition per plan.md §F).
- plan.md §D constraint: no deletion of hook-registered wrappers or build-tag-gated symbols without per-GOOS build verification.
- Audit SSOT: `.moai/reports/cli-improvement-audit-20260710.html` §3 cluster 1 Minor (dead code rows), cluster 2 Minor (buildGLMEnvVars), cluster 4 Minor (worktree_validation/init_layout/ttyConfirmer).
