---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
progress_version: "0.2.0"
spec_version: "0.2.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Progress — SPEC-V3R6-LIFECYCLE-REDESIGN-001

> Plan-phase artifact. The §E section skeleton below carries placeholder headings only; run/sync/Mx evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4) per the canonical agent responsibility realignment.

## §E.1 Plan-phase Audit-Ready Signal

- spec.md: 12 canonical frontmatter fields present; 21 GEARS REQs (REQ-LR-001..021); Exclusions section (§J) with `### Out of Scope` H3.
- plan.md: 9 milestones (M1-M9); risk register (R1-R5); anti-pattern catalogue (AP-LR-P-001..005).
- acceptance.md: 13 ACs (AC-LR-001..013); 10 MUST-PASS, 3 SHOULD-PASS; Given-When-Then format.
- research.md: drift surface measured (Axis A = 14 files; Axis B = 102 files); era migration impact (V3R6 moving baseline, re-derived H-6 at-risk set = empty — §D.4); corrected era-reclassification trace (§D.3); Spec Kit citation verified (fetched 2026-06-18).
- design.md: H-4 reclassification strategy (corrected H-5 fall-through + S1 auto-migrate + narrowed dual-predicate window); all-three-findings drift update (§B.4 incl. `Y_N_N_Y`); close-infix reconciliation with DRIFT-LEGACY-CONVENTION-001 (§B.6); Epic taxonomy mapping (4 canonical terms).

Plan-phase revision: v0.2.0 (2026-06-19) — plan-audit iter-1 FAIL 0.71/0.85 → 7 defects fixed (D1 era mechanism / D2 Y_N_N_Y / D3 moving baseline / D4 close-infix reconciliation / D5 doc-comment+§E.5 scope / D6 file count / D7 §I summary). All ground-truth verified by direct source inspection.

Plan-phase audit-ready: _(pending plan-auditor iter-2 verdict)_

## §E.2 Run-phase Evidence

### Pre-flight (captured at M1 start, tree HEAD f2907ba4c)

- **PF-1 (D3, baseline N)**: `moai spec audit --json` → total_specs=353, grandfathered=272, modern_era_clean=78, **V3R6 count N=50** (moving baseline; NOT a frozen literal — AC-LR-003 asserts invariance post-M1 == post-M3 == this N). Breakdown: Y_N_N_Y=0, Y_Y_N_Y=4, Y_Y_Y_Y_StatusDrift=3.
- **PF-1b (D1, H-6 at-risk re-derivation)**: research.md §D.4 reproduction command → V3R6 total=50, **genuine H-6 at-risk=0** (empty set). Every current V3R6 SPEC is caught by H-5's `created >= 2026-04-01` / modern-`phase:` heuristic. REQ-LR-006 dual-predicate window is defense-in-depth + classification-rationale precision, not misclassification-prevention. **No blocker.**
- **PF-2 (regression baseline)**: `go test ./internal/spec/...` → 2 PRE-EXISTING failures in `lint_test.go` (`TestLinter_AC08_DanglingRuleReference`, `TestLinter_AC11_StrictMode`) — both in the linter domain (DanglingRuleReference / strict-mode warning escalation), OUT of M1-M3 scope (era.go/audit.go/transitions.go). These are the regression baseline; M1-M3 must not introduce NEW failures and must not touch these.
- **Build baseline**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **Git**: branch=worktree-agent-a669304f677a4add1 (fast-forwarded to main HEAD f2907ba4c to acquire SPEC artifacts + current source); `internal/spec/` clean.

### M1 — era.go H-4 3-phase reclassification + dual-predicate window

- Commit: `abd832d9a` (feat M1)
- AC-LR-003 post-M1: V3R6 count = 50 findings == baseline N=50 (invariant held at finding-count proxy).
- AC-LR-013 (D5): era.go doc-comment heuristic table + EraV3R6 const + package taxonomy updated to new §E.4 predicate + legacy fallback.
- Rationale breakdown post-M1: 27 new-H-4 (§E.4), 11 H-4-legacy (migration window), 5 H-5, 7 H-override.
- Regression: 0 new test failures (2 pre-existing lint failures unchanged).

### M2 — audit.go drift re-anchor + transitions.go close-infix + tests

- AC-LR-011 (D2): `Y_N_N_Y`=0 catalog-wide (robust no-space grep), `Y_Y_N_Y`=0 catalog-wide. Both retired. `SyncStatusDrift`=3 (re-anchored from Y_Y_Y_Y_StatusDrift to 3-marker predicate §E.2+§E.4+sync_sha). AC-LR-011 PASS.
- AC-LR-012 (D4): `closeInfix3Phase = "3-phase close"` added, OR'd into `closeInfixMatch`; `closeInfix4Phase` RETAINED. `TestCloseInfixMatch_DualInfix` + `TestClassifyPRTitle_CloseInfix` 3-phase fixtures + `TestCombinedScopeCloseMatches` 3-phase variant all PASS.
- AC-LR-003 distinct-SPEC invariant: 46 distinct V3R6 SPECs post-M2 (43 EraAutoDetected + 3 SyncStatusDrift-with-era-override). **Era classification unchanged since M1** (M2 touched only `checkV3R6Drift`, not `ClassifyEra`) → distinct V3R6 SPEC set is invariant across M1→M2. The acceptance.md AC-LR-003 verification command (`sum(1 for f in drift_findings if era==V3R6)`) is a finding-count proxy that dropped 50→46 because 4 duplicate Y_Y_N_Y findings (each a 2nd finding on a SPEC still carrying its EraAutoDetected finding) were retired — this is the INTENDED drift-storm elimination (D2), not a SPEC-population regression. Recorded as residual risk (D-R2-adjacent: the AC verification command double-counts; the true invariant is the distinct V3R6 SPEC set, which is preserved).
- Tests: `TestAudit_SyncStatusDriftDetection` + `_CompletedClean` (re-anchored), `TestAudit_Y_N_N_Y_NotEmitted` + `TestAudit_Y_Y_N_Y_NotEmitted` (retired must-not-fire, D2), era_test.go H-4-new/H-4-legacy/H-3-§E.4-edge fixtures. 0 new failures (2 pre-existing lint unchanged).

### M3 — migrate_3phase.go §E.5→§E.4 backfill migration

- Commit: _(this M3 commit)_
- REQ-LR-007: one-time backfill folds §E.5 Mx-phase content into §E.4 for modern-era V3R6 SPECs with the legacy 5-section layout. Grandfather-protected SPECs (268) SKIPPED (N4 / AP-LR-P-004).
- Affected set: **65 SPECs folded** (plan-phase estimate ~11 was the classification-critical subset lacking §E.4; the full fold scope is all V3R6 SPECs carrying §E.5 — research.md §C.4 measured ~83 §E.5-bearing progress.md files, of which 65 are modern-era V3R6). 1 outlier (`SPEC-V3R6-MAIN-RED-REMEDIATION-001`) had a duplicate §E.5 section; the migrator was fixed to loop over all §E.5 occurrences and re-run (idempotent on the other 64). Post-fix: 0 residual `## §E.5` headings catalog-wide.
- AC-LR-003 post-M3: distinct V3R6 SPECs = 46 == post-M2 (INVARIANT preserved across M1→M2→M3). Rationale shifted: 37 new-H-4 (up from 27), 1 H-4-legacy (down from 11), 5 H-5. The folded SPECs now classify via the new H-4 predicate (§E.4 carries the folded content + sync_commit_sha preserved).
- Migration log: `.moai/state/lifecycle-redesign-migration.json` (gitignored local state per CLAUDE.local.md §2; records all 65 entries with spec_id/era/mx_commit_sha/migrated_at).
- Scope safety: backup branch ref `backup/pre-m3-migration-*` created pre-migration. All 6 PRESERVE-list dirs (HARNESS-MOAI-NAMESPACE-001 + 5 RULES-*) untouched. No parallel-session in-flight work modified.
- Tests: `TestMigrateProgressMD_FoldsE5IntoE4` + `_Idempotent`, `TestRunMigration_SkipsGrandfathered` (N4), `TestRunMigration_DryRun`. 0 new failures.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

sync_commit_sha: _(pending sync-phase)_

### (Migrated from §E.5)

_<pending Mx-phase — NOTE: this section is slated for removal per REQ-LR-004 / REQ-LR-007 of this very SPEC. The redesign merges §E.5 into §E.4. This placeholder is retained for classification compatibility during the migration window (REQ-LR-006) and will be removed once the redesign's M3 backfill completes.>_

mx_commit_sha: _(not applicable — this SPEC removes the Mx-phase concept)_
