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

### M2 — Pruner integration and reporting (card tree `WT-10-14-class-3`, code HEAD `ab1e48138`)

Commit order:

| Commit | Content |
|---|---|
| `396484c7f` | RED: pruner-level tests for AC-UBR-08/09/11/12/13/14 and the budget reader, against decide-nothing stubs |
| `312c303a0` | `PruneBackups` (census, selection, whole-run removal, report), `CleanupOldBackups` forwards to it, `LoadBackupPastVersionsKeep`, registry row moved to the new function, one fixture adapted (see Gaps) |
| `73b390a05` | RED: update-output renderer test against a render-nothing stub |
| `ab1e48138` | renderer, both callers read the key and print the rows, inventory entry R to W reader |

Verbatim RED before the pruner (command `go test -count=1 -v -run 'TestAC_UBR_(08|09|11|12|13|14)|TestPrune_' ./internal/cli/update/backup/` at `396484c7f`, exit 1):

```
--- FAIL: TestAC_UBR_08_BookkeepingIsRecognizedByWhatTheWriterWrote (0.01s)
    retention_prune_test.go:85: AC-UBR-08: run holding the user .complete file = {Run: Dirs:[] Action: Reason: Protects:0 Detail:} (found=false), want retained for protects-versions
--- FAIL: TestAC_UBR_09_TemplateDefaultsRenderIsExempt (0.02s)
--- FAIL: TestAC_UBR_12_EachDecisionIsReported (0.01s)
    retention_prune_test.go:133: AC-UBR-12: window=0 runs=0, want 1 and 4
--- FAIL: TestAC_UBR_13_RootContainmentStillRefuses (0.06s)
    retention_prune_test.go:196: AC-UBR-13: RootRefused=false, want true
--- FAIL: TestAC_UBR_14_ExplicitWipeAndNegativeKeepCount (0.01s)
    retention_prune_test.go:231: AC-UBR-14: Removed=0, want 3
--- FAIL: TestPrune_RunSpansBothRootsAsOneUnit (0.00s)
--- FAIL: TestAC_UBR_11_UnclassifiableContentFailsClosedAndNeverBlocks (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.529s
```

Verbatim RED of the registry criterion (AC-UBR-19; command `go test -count=1 -v -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/`, with the pruner moved and the row not yet moved, exit 1):

```
update_destructive_registry_test.go:57: unregistered destructive site: internal/cli/update/backup/retention_prune.go PruneBackups has 1 os.RemoveAll/os.Rename call site(s) but no registry row
update_destructive_registry_test.go:68: stale registry row: internal/cli/update/backup/backup.go PruneBackups is registered but the source has no destructive call site there
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.03s)
```

Verbatim RED of the update-output renderer (at `73b390a05`, exit 1): `--- FAIL: TestBackupRetentionReportLines_NamesEveryDecision` with 14 lines of the form `AC-UBR-12: the report lines do not contain "removed":` and `--- PASS: TestBackupRetentionReportLines_QuietWhenNothingHappened`.

Verbatim RED of the budget reader (`go test -count=1 -v -run TestLoadBackupPastVersionsKeep ./internal/config/`, exit 1): `LoadBackupPastVersionsKeep = 0, want 3` and `= 0, want 5` for the six subtests.

GREEN at `ab1e48138` (this run, this tree):

```
$ go test -count=1 -v -run '^TestAC_UBR_(01|02|03|08|09|11|12|13|14)_' ./internal/cli/update/backup/
--- PASS: TestAC_UBR_01_PreCleanUserFileSurvivesRotation (0.01s)
--- PASS: TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot (0.01s)
--- PASS: TestAC_UBR_03_RedundantCopiesAreNotAllKept (0.01s)
--- PASS: TestAC_UBR_08_BookkeepingIsRecognizedByWhatTheWriterWrote (0.00s)
--- PASS: TestAC_UBR_09_TemplateDefaultsRenderIsExempt (0.00s)
--- PASS: TestAC_UBR_12_EachDecisionIsReported (0.00s)
--- PASS: TestAC_UBR_13_RootContainmentStillRefuses (0.01s)
--- PASS: TestAC_UBR_14_ExplicitWipeAndNegativeKeepCount (0.00s)
--- PASS: TestAC_UBR_11_UnclassifiableContentFailsClosedAndNeverBlocks (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	0.417s
$ go test -count=1 -v -run '^(TestBackupRetentionReportLines_|TestDestructiveTargetRegistry_CoversAllSites|TestCleanReinstall_PrunesOldConfigBackups)' ./internal/cli/
--- PASS: TestBackupRetentionReportLines_NamesEveryDecision (0.00s)
--- PASS: TestBackupRetentionReportLines_QuietWhenNothingHappened (0.00s)
--- PASS: TestCleanReinstall_PrunesOldConfigBackups (0.06s)
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.07s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.191s
```

M2 acceptance matrix (Exit: AC-UBR-01, 08, 09, 11, 12, 13, 14, 19 green):

| AC | Status | Evidence |
|---|---|---|
| AC-UBR-01 (pruner level) | PASS | `TestAC_UBR_01_PreCleanUserFileSurvivesRotation` (red at M1, green here) |
| AC-UBR-02, AC-UBR-03 (pruner level) | PASS | `TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot`, `TestAC_UBR_03_RedundantCopiesAreNotAllKept` |
| AC-UBR-08 | PASS | `TestAC_UBR_08_BookkeepingIsRecognizedByWhatTheWriterWrote` |
| AC-UBR-09 | PASS | `TestAC_UBR_09_TemplateDefaultsRenderIsExempt` |
| AC-UBR-11 | PASS | `TestAC_UBR_11_UnclassifiableContentFailsClosedAndNeverBlocks` (unix; FIFO and chmod 0) |
| AC-UBR-12 | PASS | `TestAC_UBR_12_EachDecisionIsReported` (pruner report), `TestBackupRetentionReportLines_NamesEveryDecision` (update output) |
| AC-UBR-13 | PASS | `TestAC_UBR_13_RootContainmentStillRefuses` (three root shapes) plus the existing containment tests |
| AC-UBR-14 | PASS | `TestAC_UBR_14_ExplicitWipeAndNegativeKeepCount` |
| AC-UBR-19 | PASS | `TestDestructiveTargetRegistry_CoversAllSites`; the row names `PruneBackups` in `retention_prune.go`; no removal or rename call site was added |

Measured, not claimed: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test -count=1 -cover` backup package 88.4%, config package 85.6%; `go test -count=1 ./internal/cli/update/...` all `ok` (skillresidue 76.3%, unchanged, M3's package); `go vet` clean; `gofmt -l` empty; `golangci-lint run ./internal/cli/update/backup/... ./internal/config/... ./internal/cli/` reports 3 errcheck findings, all in `internal/cli/factory_run_adopt.go` (not touched by this card).
Not run (Gaps): the full `internal/cli` suite (named tests only: the 12 `CleanupOldBackups`-bearing tests, the update-flow test files that touch backup roots, the registry test, the renderer tests); `make build` (no template file changed in M2); a mutant probe of the cover step beyond the M1 oracle tests; real-data remeasure (M6).
Adapted existing test: the rotation subtest of `TestBackupSubsystem_DestructiveSurfaces` (`internal/cli/update_preserve_reach_test.go`) wrote a distinct body into each backup, which budgeted retention (REQ-UBR-003) keeps by design; the fixture now writes one body so the subtest still asserts which backups survive by name. Flagged for the audit.

### M3 — Residue passes stay report-only; rename bound to the root (card tree `WT-10-14-class-3`, code HEAD `75d97ab02`)

Commit order (the commit graph is the ordering witness, `verification-claim-integrity.md` §2.3):

| Commit | Content |
|---|---|
| `3935b4434` | guards AC-UBR-04, AC-UBR-05, AC-UBR-15 (green on the starting tree by design; added before any residue change) |
| `f2e692588` | RED: AC-UBR-16 (swapped skill folder) and AC-UBR-17 (file at the temporary name), against the path-based rename |
| `75d97ab02` | root-bound rename, registry rows `moveOne` and `dropOne`, destructive-target scan extended to `os.Root` handles, pin test rebuilt, adapted tests |

Guards (class standard, no RED by design, operator ruling Q3). The cited probe commands on the repaired tree, `go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_residue.rel.json -run '^TestT1733ProbeP4a$' ./internal/cli/update/skillresidue/` and the `P4b` twin, exit 0 each:

```
    zz_t1733_probe_test.go:48: P4a err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill} {.agents/skills/moai-kanban-foreman reported-retired-link}] b-link-survives=true a-folder-in-place=true
    zz_t1733_probe_test.go:73: P4b err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill}] content-now-outside-project=false a-folder-in-place=true
```

Both lines equal the guard values of acceptance.md. The committed tests assert more than the probes log: the full tree snapshot of A and B is unchanged and every action is a report row. A one-off mutant that removed the reported link made `TestAC_UBR_04_*` and `TestAC_UBR_15_*` fail (project B's link gone; tree changed), and was reverted before the guard commit.

Verbatim RED before the fix (command `go test -count=1 -v -run 'TestAC_UBR_16_|TestAC_UBR_17_' ./internal/cli/update/skillresidue/` at `f2e692588`, exit 1; on this machine's case-insensitive temporary directory, the first run of AC-UBR-16 skipped itself on a `caseSensitiveDir` call that the test did not need, so that call was removed before the RED commit and the output below is the second run):

```
=== RUN   TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename
=== RUN   TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename/Clean
    residue_rename_race_test.go:74: the outside folder was changed through the swapped component: [SKILL.md]
=== RUN   TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename/ApplyDeferredRenames
    residue_rename_race_test.go:74: the outside folder was changed through the swapped component: [SKILL.md]
--- FAIL: TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename (0.01s)
    --- FAIL: TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename/Clean (0.00s)
    --- FAIL: TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename/ApplyDeferredRenames (0.00s)
=== RUN   TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName
    residue_rename_race_test.go:111: the user file's bytes are gone; names now [SKILL.md] (Clean err: <nil>)
--- FAIL: TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.261s
FAIL
```

Verbatim RED of the registry criterion (AC-UBR-20; command `go test -count=1 -v -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/`, rows `moveOne` and `dropOne` written and the scan not yet extended, exit 1). The scan counted `os.RemoveAll` and `os.Rename` only, so the new root-handle sites were invisible: the F22 shape.

```
    update_destructive_registry_test.go:68: stale registry row: internal/cli/update/skillresidue/residue.go dropOne is registered but the source has no destructive call site there
    update_destructive_registry_test.go:68: stale registry row: internal/cli/update/skillresidue/residue.go moveOne is registered but the source has no destructive call site there
    update_destructive_registry_test.go:87: scanned 15 (file, function) pair(s) across internal/cli/update and internal/cli/update*.go; registry has 17 row(s)
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.02s)
```

Falsification of the guards, observed once on the repaired tree (a scratch non-test file `zz_injected.go` in the package with an unregistered `root.Remove` function and a package-level alias `var zzAlias = dropFile`; the file was removed afterwards and is not in any commit):

```
    update_destructive_registry_test.go:64: unregistered destructive site: internal/cli/update/skillresidue/zz_injected.go zzInjectedRemoval has 1 os.RemoveAll/os.Rename call site(s) but no registry row
--- FAIL: TestDestructiveTargetRegistry_CoversAllSites (0.02s)
--- FAIL: TestMoveFileReferencesArePinned (0.00s)
    residue_moveone_test.go:88: dropFile@package-level references = 2, want 1
```

GREEN at `75d97ab02` (this run, this tree):

```
$ go test -count=1 -v -run '^(TestAC_UBR_04_|TestAC_UBR_05_|TestAC_UBR_15_|TestAC_UBR_16_|TestAC_UBR_17_|TestMoveFileReferencesArePinned)' ./internal/cli/update/skillresidue/
--- PASS: TestMoveFileReferencesArePinned (0.00s)
--- PASS: TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename (0.01s)
    --- PASS: TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename/Clean (0.00s)
    --- PASS: TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename/ApplyDeferredRenames (0.00s)
--- PASS: TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName (0.00s)
--- PASS: TestAC_UBR_17_RenameDoesNotOverwriteASkillFileThatAppearsAtTheDestination (0.00s)
--- PASS: TestAC_UBR_04_ContentBehindSymbolicAgentsSkillsIsUntouched (0.01s)
--- PASS: TestAC_UBR_05_NothingIsWrittenOutsideTheProjectThroughArchiveLink (0.00s)
--- PASS: TestAC_UBR_15_RetiredFoldersAndLinksAreReportedNeverMovedOrRemoved (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.296s
$ go test -count=1 -v -run '^TestDestructiveTargetRegistry_CoversAllSites$' ./internal/cli/
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.03s)
```

M3 acceptance matrix (Exit: AC-UBR-04, 05, 15, 16, 17, 20 green):

| AC | Status | Evidence |
|---|---|---|
| AC-UBR-04 | PASS (regression guard) | `TestAC_UBR_04_ContentBehindSymbolicAgentsSkillsIsUntouched`; cited probe `TestT1733ProbeP4a` line above |
| AC-UBR-05 | PASS (regression guard) | `TestAC_UBR_05_NothingIsWrittenOutsideTheProjectThroughArchiveLink`; cited probe `TestT1733ProbeP4b` line above |
| AC-UBR-15 | PASS | `TestAC_UBR_15_RetiredFoldersAndLinksAreReportedNeverMovedOrRemoved` (two runs, tree snapshot unchanged, no `.moai/archive`) |
| AC-UBR-16 | PASS | `TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename` (`Clean` and `ApplyDeferredRenames`) |
| AC-UBR-17 | PASS | `TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName`, `TestAC_UBR_17_RenameDoesNotOverwriteASkillFileThatAppearsAtTheDestination` |
| AC-UBR-20 | PASS | `TestDestructiveTargetRegistry_CoversAllSites` (rows `moveOne`, `dropOne`; 17 rows, 27 sites) and `TestMoveFileReferencesArePinned`; falsification observed above |

Design note for the audit (a deviation of mechanism, not of intent). D5 describes a temporary name reserved by exclusive creation of a file and moved onto. A move (`rename(2)`) replaces whatever is at its destination, so a file that appears at the claimed name between the claim and the move is overwritten and AC-UBR-17 cannot hold for it. The rename therefore claims a staging folder with an exclusive `Mkdir`, moves the lowercase file into it, links it to `SKILL.md` (a link fails when the name exists) and drops only names whose content has another name. A failed step leaves a staging folder `skill.md.tmp-rename` (formerly a file of that name) holding the content, and the error names it. No data-deleting behavior was added: `dropOne` removes an empty folder the step created, or a name of a file that has another name.
Adapted existing tests (requirement REQ-UBR-017 changes the rename's steps and signature; assertions kept, seams re-pointed): `TestClean_RenameFailureIsReturnedAndNothingMoves` (new seam signature; adds that no staging folder is left), `TestRenameSkillFile_NoFreeTemporaryNameIsRefused` and `TestRenameSkillFile_UninspectableDirectoryIsAnError` (renamed from `TestFreeName_...` because `freeName` no longer exists; call the new signature), `TestRenameSkillFile_RestoreFailureIsReported` (injects the link refusal and a failed move back, asserts the error names the staging path and the content is there), `TestRenameSkillFile_RestoresLowercaseNameWhenTheLinkIsRefused` (renamed from `...WhenSecondMoveFails`: the second step is now a link), `TestMoveFileReferencesArePinned` (names the four steps and four seams). `update_destructive_registry_test.go` gained `rootDestructiveFuncs` and `rootHandleNames`, and counts calls outside function declarations separately; the existing 16 rows count the same sites as before.

Measured, not claimed: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go vet ./internal/cli/update/skillresidue/` exit 0; `go test -count=1 -cover ./internal/cli/update/...` all `ok`, skillresidue 88.4% (76.3% before M3), backup 88.4%; `go vet ./internal/cli/update/skillresidue/ ./internal/cli/` exit 0; `gofmt -l` empty; `golangci-lint run ./internal/cli/update/skillresidue/... ./internal/cli/` reports the three pre-existing errcheck findings in `internal/cli/factory_run_adopt.go` and nothing else; 13 `internal/cli` tests run by name (the registry test, the two skill-rename ordering tests, `TestRunUpdateHousekeeping_*`, `TestRunUpdateHousekeepingLocked_*`, `TestRemoveStateResidue_*`, `TestRunUpdate_RestoreDoesNotRunHousekeeping`) all pass.
Environment brackets around the `internal/cli` runs: `~/.gitconfig` 503 lines, sha256 `376c6febc88ba234321078fa8e8c4ec922b458192d1199a31a4eb03b55660548`, two `/var/folders|/repo|Test` lines (the same two lines as at the start), unchanged by the runs; backlog.db `PRAGMA quick_check` `ok`, header `SQLite format 3`. At the start of M3 the file read 502 lines with sha256 `b3f9680a...965b`; it had changed by the time of the first `internal/cli` run (modification time 09:31) without any `internal/cli` test of this card having run, so a concurrent session is the writer.
Not run (Gaps): the behavior of `os.Root` `Mkdir`, `Rename`, `Link` and `Remove` on Windows (cross-build and vet only; a filesystem without hard links makes the rename fail closed with the content back under its own name, by `putBack`, and that path is exercised only through an injected link refusal); the case-sensitive-filesystem variants of the rename tests on this machine (its temporary directory folds case, so `TestClean_ReportsButKeepsBothSkillFilesPresent` skips, and the destination test of AC-UBR-17 asserts that no byte is lost rather than that `skill.md` is back under its name); the full `internal/cli` suite; `make build`; the template tests (no template file changed in M3).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_ — M3 of M6 is complete; the signal is written at run close.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
