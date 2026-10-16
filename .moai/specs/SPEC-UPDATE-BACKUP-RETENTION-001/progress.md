# Progress — SPEC-UPDATE-BACKUP-RETENTION-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-16
author: manager-spec
tier: L
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md]
requirements: 20
acceptance_criteria: 24
open_clarifications: 0   # operator answered Q1-Q3 on 2026-10-16; markers removed
decision_gate: on        # decision-index.md Q1-Q3 RESOLVED (operator via brain), Q4 closed, Q5 DEFAULT-APPLIED, Q6 closed by Q3
tree_sha: e669da81b3f865488dd40f840623bed69839f459   # code identical at 3b133bedf (SPEC commit)
revision: 0.2.0 (after plan-audit FAIL, receipt rcpt-23e014c572cca1b636f63244)
```

Revision 0.2.0 artifacts are uncommitted at this signal; the orchestrator commits them. No audit was run and no verdict file written by this author.

## Revision log (plan-audit iteration 1, D1-D14)

| Finding | Disposition |
|---|---|
| D1 (MP-7, three markers) | Closed. Markers deleted from plan.md §B; Q1-Q3 answers recorded in decision-index.md with date 2026-10-16, verdict source "operator via brain"; Q1 (report only), Q2 (soft cap), Q3 (no reintroduction) folded into REQ-UBR-010/015/016/018 and AC-UBR-04/05/06/15/18. |
| D2 (REQ-005 vs 008 vs 010, design D2 step 4, AC-11) | Closed. Single disposition: an unclassifiable (FIFO-only, unreadable) run is an exempt run, retained unconditionally and counted as a holder; REQ-005 carries the exception, REQ-010 says "plus the exempt runs", design D2 step 4 never drops exempt runs, AC-UBR-11 asserts the FIFO-only case. |
| D3 (AC-06(b) equality with a non-unique cover) | Closed. (b) asserts coverage, irredundancy and the bound "retained non-exempt excess runs at most the number of distinct protected versions"; no equality; design D2 gives the {abc},{a},{b},{c} counter-shape. |
| D4 ((a) cannot fail; mutant note wrong) | Closed. (a) replaced by "runs present = keepCount + R, R named in the report"; mutant note corrected (keep-everything fails (b) irredundancy); the card's "at or under cap" wording replaced in REQ-UBR-010 rationale by the soft-cap form. |
| D5 (Q2 evidence ordering) | Closed. Q2 is ruled and Q4 closed; no decision depends on the figure; the oracle can compute R from the pre-prune census at any time (design D2); AC-UBR-06 prints it as a measured output; the "before run-phase entry" obligation for the figure is deleted. |
| D6 (Q6 misclassified, false rationale) | Closed. Q6 reclassified product-level, DEFAULT-APPLIED removed; the default (keep current non-refusing report-only behavior, measured E4) recorded as consistent with the 3.2 principle; closed by the Q3 ruling. |
| D7 (AC-UBR-20 given clause vs M2 Exit) | Closed. Split: AC-UBR-19 (CleanupOldBackups row) bound to M2 Exit; AC-UBR-20 (rename site row) bound to M3 Exit. |
| D8 (bookkeeping shape imitation) | Deferred by design, recorded: design D4 residual and spec risk 7; provenance is a larger design choice; no unsafe fixture asserted. |
| D9 (RED-now commands not self-contained) | Closed. Overlay files rewritten with tree-relative paths (`overlay_*.rel.json`), commands in research.md/acceptance.md have no placeholder; the placement steps and the overlay JSON text are in research.md Appendix A.3; probes-only-log caveat stated; commands re-executed in this run. |
| D10 (AC-UBR-19 without Exit) | Closed. M4 has an Exit line naming AC-UBR-18 (the F21 criterion, renumbered). |
| D11 (Q5 default rationale cites reverted code) | Closed. Rationale rewritten: no current behavior exists; rule applied is the smaller user-visible surface. |
| D12 (REQ-020/021 verification-layer) | Kept as requirements (renumbered REQ-UBR-019/020); harmless, tier ceiling respected. |
| D13 (unlisted `related_specs` field) | Closed. Field removed; the three SPEC ids moved to HISTORY. |
| D14 (AC-UBR-23 judgment wording) | Closed. The expected tracked-change set is named in AC-UBR-23. |

Additional folds from the leader brief: baseline on main added (research.md, spec A.2, plan A.1); reference behavior `8e43f064a^` in design D0; main-versus-tree RED/GREEN cells in AC-UBR-01 to AC-UBR-06; post-landing operator verification in design D11; counts 20 requirements, 24 acceptance criteria (ceilings 25/25).

## §E.2 Run-phase Evidence

### M1 — Retention contract and shipped key (card tree `WT-10-14-class-3`, code HEAD `815223b5b4493ee0b101d2587101bba8386c7014`)

Commit order (the commit graph is the ordering witness, `verification-claim-integrity.md` §2.3):

| Commit | Content |
|---|---|
| `8c7c4cdb5` | RED probe P1 (AC-UBR-01) + shared helpers; spec.md `draft -> in-progress` |
| `f3d4ad42d` | RED probe P2 (AC-UBR-02) |
| `750a8a7fe` | RED probe P3 (AC-UBR-03) |
| `d774ab272` | shipped key `state.backup_past_versions_keep`, struct field, single default, inventory entry, AC-UBR-10 tests |
| `716d81b53` | RED selection-step tests + vocabulary + decide-nothing stub (AC-UBR-02/03/07 at the selection level) |
| `815223b5b` | selection step implemented |

Verbatim RED output, captured before the fix (pruner-level probes, start tree code identical to `e669da81b`; command `go test -count=1 -v -run '^TestAC_UBR_0N_' ./internal/cli/update/backup/`, exit code 1 each):

```
=== RUN   TestAC_UBR_01_PreCleanUserFileSurvivesRotation
    retention_probe_p1_test.go:74: AC-UBR-01 user-file-copies-before=1 deleted=1 user-file-copies-after=0
    retention_probe_p1_test.go:76: AC-UBR-01: no copy of the user file remains after the prune (before=1 after=0); the pruner kept only the newest runs
--- FAIL: TestAC_UBR_01_PreCleanUserFileSurvivesRotation (0.06s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.015s
```

```
=== RUN   TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot
    retention_probe_p2_test.go:22: AC-UBR-02 deleted=4 remaining=[update-2026-10-05T00-00-00Z]
    retention_probe_p2_test.go:28: AC-UBR-02: the run holding v1 (update-2026-10-01T00-00-00Z) was removed; want it retained (budget 3, window 1)
    retention_probe_p2_test.go:28: AC-UBR-02: the run holding v2 (update-2026-10-02T00-00-00Z) was removed; want it retained (budget 3, window 1)
    retention_probe_p2_test.go:28: AC-UBR-02: the run holding v3 (update-2026-10-03T00-00-00Z) was removed; want it retained (budget 3, window 1)
    retention_probe_p2_test.go:40: AC-UBR-02: deleted=4, want 1 (only the older v4 run)
--- FAIL: TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot (0.01s)
FAIL
```

```
=== RUN   TestAC_UBR_03_RedundantCopiesAreNotAllKept
    retention_probe_p3_test.go:22: AC-UBR-03 deleted=10 remaining=[update-2026-10-11T00-00-00Z]
    retention_probe_p3_test.go:26: AC-UBR-03: the newest v1 holder (update-2026-10-10T00-00-00Z) was removed; want it retained as the sole v1 copy
    retention_probe_p3_test.go:35: AC-UBR-03: 0 runs holding v1 remain, want exactly 1
    retention_probe_p3_test.go:42: AC-UBR-03: deleted=10, want 9
--- FAIL: TestAC_UBR_03_RedundantCopiesAreNotAllKept (0.02s)
FAIL
```

(The package-level `FAIL github.com/modu-ai/moai-adk/internal/cli/update/backup <time>` trailer lines are omitted from the P2 and P3 blocks; every test line above is verbatim.)

RED for AC-UBR-10 (`go test -count=1 -v -run '^TestAC_UBR_10_' ./internal/config/`, before the key existed):

```
=== RUN   TestAC_UBR_10_BudgetKeyShipsOnce/template_ships_key_with_default
    state_backup_past_keep_ship_test.go:40: AC-UBR-10: the shipped state.yaml has no state.backup_past_versions_keep key
=== RUN   TestAC_UBR_10_BudgetKeyShipsOnce/default_declared_in_one_place
    state_backup_past_keep_ship_test.go:70: AC-UBR-10: the past-versions budget default is declared in [], want exactly [internal/config/defaults.go]
--- FAIL: TestAC_UBR_10_BudgetKeyShipsOnce (1.61s)
FAIL	github.com/modu-ai/moai-adk/internal/config	2.020s
```

RED for the selection step (against the decide-nothing stub, commit `716d81b53`): ten tests failed
(`TestAC_UBR_02_SelectionKeepsPastVersionsOutsideTheWindow`, `TestAC_UBR_03_SelectionRetainsOneCopyOfRepeatedVersion`,
`TestAC_UBR_07_BudgetBoundsProtection`, `TestSelection_CurrentVersionIsNotAPastVersion`,
`TestSelection_CoverPrefersTheRunHoldingMostVersions`, `TestSelection_IrredundancyPassDropsACoveredRun`,
`TestSelection_AnIrredundantCoverOfEitherSizePassesTheOracle`, `TestSelection_ExemptRunsAreRetainedAndCountAsHolders`,
`TestSelection_KeepCountEdges`, `TestSelection_RandomCensusesSatisfyTheOracle`).

GREEN at `815223b5b` (this run, this tree):

```
$ go test -count=1 -v -run 'TestAC_UBR_02_Selection|TestAC_UBR_03_Selection|TestAC_UBR_07|TestSelection_' ./internal/cli/update/backup/
--- PASS: TestAC_UBR_02_SelectionKeepsPastVersionsOutsideTheWindow (0.00s)
--- PASS: TestAC_UBR_03_SelectionRetainsOneCopyOfRepeatedVersion (0.00s)
--- PASS: TestAC_UBR_07_BudgetBoundsProtection (0.00s)
--- PASS: TestSelection_CurrentVersionIsNotAPastVersion / CoverPrefersTheRunHoldingMostVersions / IrredundancyPassDropsACoveredRun / AnIrredundantCoverOfEitherSizePassesTheOracle / ExemptRunsAreRetainedAndCountAsHolders / KeepCountEdges / RandomCensusesSatisfyTheOracle
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.247s
$ go test -count=1 -v -run 'TestAC_UBR_10|TestNewDefaultStateConfigCarriesBackup|TestStateBackupPastVersionsKeepLoadsFromYAML' ./internal/config/
--- PASS: TestAC_UBR_10_BudgetKeyShipsOnce (4.28s)
--- PASS: TestNewDefaultStateConfigCarriesBackupPastVersionsKeep (0.00s)
--- PASS: TestStateBackupPastVersionsKeepLoadsFromYAML (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/config	4.772s
```

M1 acceptance matrix (Exit: AC-UBR-07 and AC-UBR-10 green; AC-UBR-02 and AC-UBR-03 green against the selection step):

| AC | Status | Evidence |
|---|---|---|
| AC-UBR-02 (selection step) | PASS | `TestAC_UBR_02_SelectionKeepsPastVersionsOutsideTheWindow` |
| AC-UBR-03 (selection step) | PASS | `TestAC_UBR_03_SelectionRetainsOneCopyOfRepeatedVersion` |
| AC-UBR-07 | PASS | `TestAC_UBR_07_BudgetBoundsProtection` (budget 3, 1, and 0 / -4 reading as 3) |
| AC-UBR-10 | PASS | `TestAC_UBR_10_BudgetKeyShipsOnce`, `TestNewDefaultStateConfigCarriesBackupPastVersionsKeep`, `TestStateBackupPastVersionsKeepLoadsFromYAML` |
| AC-UBR-01, 02, 03 (pruner level) | RED by design until M2 | `TestAC_UBR_01/02/03_*` fail against `CleanupOldBackups`, which M1 does not wire to the selection |

Measured, not claimed: `go test -cover` backup package 88.5% (with the three probes failing), config package 85.6%;
`golangci-lint run ./internal/cli/update/backup/... ./internal/config/...` exit 0, `0 issues.`;
`go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0;
`go test -count=1 ./internal/template/` ok (415.168s, run on the tree after the template edit);
`go test -count=1 ./internal/cli/update/...`: `update`, `deploy`, `merge`, `plan`, `report`, `skillresidue` ok; `backup` FAIL on exactly the three pruner-level probes above and nothing else;
`go run ./internal/template/scripts/gen-catalog-hashes.go --all` (the catalog step of `make build`) left no tracked change (`git status --short` empty).
Not run (Gaps): `make build` in full (the catalog-hash step above was run on its own); `go test -run TestDestructiveTargetRegistry_CoversAllSites ./internal/cli/` (an attempt with an isolated `GIT_CONFIG_GLOBAL` was refused by the worktree-session guard; M1 adds no `os.Remove*` / `os.Rename` / syscall call site, shown by a grep of the M1 diff, and the registry row for `CleanupOldBackups` is M2's AC-UBR-19).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_ — M1 of M6 is complete; the signal is written at run close.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
