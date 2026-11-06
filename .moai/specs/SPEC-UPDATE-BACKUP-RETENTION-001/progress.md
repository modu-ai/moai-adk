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

### M4 — An empty database file is reported, not removed (card tree `WT-10-14-class-3`, RED `471dbc55b`, GREEN `9d7f1823c`)

Operator ruling Q1 (report only). The state-residue pass no longer deletes a zero-byte `*.db`; the only file `removeStateResidue` still removes is the zero-byte `backlog.json.lock`. A new `emptyStateDBs` lists zero-byte `*.db` files older than the minimum age with no `-wal` / `-shm` sidecar, and the housekeeping pass appends one reference row per file: `empty DB file present: <path> (zero bytes, left in place; nothing was removed)`. Both share `walkStateResidue`, which keeps the symbolic-`.moai` refusal, the regular-file test, the size test and the minimum-age test. No registry row existed for the removed call (the destructive registry counts `os.RemoveAll` / `os.Rename` and root-handle calls, not a plain `os.Remove`), and `TestDestructiveTargetRegistry_CoversAllSites` stays green. No data-deleting behavior was added; one was removed.

RED (committed at `471dbc55b`, captured before the implementation; start tree `9f4f34b3b`), verbatim from `go test -count=1 -run 'TestRemoveStateResidue|TestRunUpdateHousekeeping|TestRunUpdate_RestoreDoesNotRunHousekeeping' ./internal/cli/`:

```
--- FAIL: TestRemoveStateResidue_RemovesRetiredLockOnly (0.01s)
    update_state_residue_test.go:58: removed [.moai/state/factory.db .moai/state/factory/factory.db .moai/state/todo/backlog.db .moai/state/todo/backlog.json.lock], want [.moai/state/todo/backlog.json.lock]
--- FAIL: TestRemoveStateResidue_KeepsStubWithSQLiteSidecar (0.02s)
    --- FAIL: TestRemoveStateResidue_KeepsStubWithSQLiteSidecar/-wal (0.00s)
        update_state_residue_test.go:102: removed [.moai/state/todo/plain.db], want none
    --- FAIL: TestRemoveStateResidue_KeepsStubWithSQLiteSidecar/-shm (0.01s)
        update_state_residue_test.go:102: removed [.moai/state/todo/plain.db], want none
--- FAIL: TestRunUpdateHousekeeping_EmptyDBReportedNotRemoved_ACUBR18 (0.00s)
    update_state_residue_test.go:152: empty database removed: stat .../.moai/state/todo/backlog.db: no such file or directory
--- FAIL: TestRunUpdateHousekeeping_ReportsChangesInProject (0.01s)
    update_state_residue_test.go:276: reference rows = ["removed 1 leftover state file(s): .moai/state/todo/backlog.db"], want ["empty DB file present: .moai/state/todo/backlog.db (zero bytes, left in place; nothing was removed)"]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.441s
```

M4 acceptance matrix (Exit: AC-UBR-18 green):

| AC | Status | Evidence |
|---|---|---|
| AC-UBR-18 | PASS | `TestRunUpdateHousekeeping_EmptyDBReportedNotRemoved_ACUBR18`: a handle opened on the zero-byte database before the pass writes successfully after it and the bytes land at the path; the row names the path; the zero-byte `backlog.json.lock` of the same age is removed; `go test -count=1 -run 'TestRemoveStateResidue|TestRunUpdateHousekeeping|TestRunUpdate_RestoreDoesNotRunHousekeeping|TestDestructiveTargetRegistry' ./internal/cli/` → `ok` |

Adapted existing tests (requirement REQ-UBR-018 reverses the expected behavior; assertion intent kept, the file is still expected present and is now reported): `TestRemoveStateResidue_RemovesStubsAndRetiredLockOnly` renamed `TestRemoveStateResidue_RemovesRetiredLockOnly` (the three zero-byte databases move from the removed list to the must-survive list); `TestRemoveStateResidue_KeepsStubWithSQLiteSidecar` (the plain stub is now kept and is the only one reported; the sidecar stub is neither removed nor reported); `TestRunUpdateHousekeeping_ReportsChangesInProject` (the expected reference row is the empty-DB row, not a removal row); `TestRemoveStateResidue_RefusesSymlinkedMoAIDir` gained a report-side check (no empty-DB row through a symbolic `.moai`). Unchanged and still passing: `TestRemoveStateResidue_KeepsRecentStub`, `TestRunUpdateHousekeeping_NoOpOutsideProject`, `TestRunUpdateHousekeepingLocked_*`, `TestRunUpdate_RestoreDoesNotRunHousekeeping`, `TestRunUpdateHousekeeping_RepairsSkillResidue`.

Measured, not claimed: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet ./internal/cli/` exit 0; `gofmt -l` on the two touched files empty; `go test -count=1 -cover ./internal/cli/update/...` all `ok` (update 85.2%, backup 88.4%, deploy 85.9%, merge 93.1%, plan 95.2%, report 95.1%, skillresidue 88.4%); `golangci-lint run ./internal/cli/` reports exactly the three pre-existing errcheck findings in `internal/cli/factory_run_adopt.go` and nothing else (a fourth finding, `holder.Close` unchecked in the new test, appeared and was fixed with `defer func() { _ = holder.Close() }()` before the GREEN commit); `grep -n AskUserQuestion internal/cli/update_state_residue.go` empty (exit 1).
Environment brackets around the `internal/cli` runs: `~/.gitconfig` 503 lines, sha256 `376c6febc88ba234321078fa8e8c4ec922b458192d1199a31a4eb03b55660548`, two `/var/folders|/repo|Test` lines, unchanged before and after; backlog.db header `SQLite format 3`, `PRAGMA quick_check` `ok`.
Not run (Gaps): the full `internal/cli` suite (named `-run` tests only; the package coverage figure of 5.0% in that bounded run is not the package's coverage); coverage of `internal/cli` as a whole; `make build`; the template tests (no template file changed); whether `moai doctor` lists the same finding (optional, not required by REQ-UBR-018); Windows behavior of the pass (cross-build only).

### M5 — Mechanical closure (card tree `WT-10-14-class-3`, tests `96b2a17d5`, start tree `56ca7f30c`)

M5 adds tests only. No production file changed, and no data-deleting behavior was added. The AC-UBR-21 premise from M1 (StateConfig cannot join `symmetryCases`) was resolved inside the AC's wording: `checkSymmetry` is a whole-section bijection, and the shipped `state:` block carries `profile_max_bytes` (no struct field) while the struct carries `retention_days` (not shipped), so joining the section would fail on keys this SPEC does not own and would need a production change (a struct field or a shipped key) the SPEC does not ask for. `internal/config/state_symmetry_key_test.go` applies the same two-direction check to the one key the SPEC ships, and decodes the shipped bytes through a `state:` section wrapper to `DefaultBackupPastVersionsKeep`. `internal/cli/update_backup_retention_paths_test.go` runs the real full-reinstall path (`runCleanReinstall`) and the real template-sync path (`runTemplateSyncAt`, which drives `runTemplateSyncWithReporter`) over seven seeded stale config-backup runs and asserts the per-run removal row and the retention summary row on each path's own surface (stdout for the reinstall, the update ledger's reference rows for the sync).

Kind of check: both are closure guards over behavior that already exists (M1 shipped the key, M2 wired the rows), so they are green on the start tree and have no RED against it. Their failing state was observed by mutation of the start tree, each reverted and not committed (`git status --short` showed only the two new test files afterwards):

```
# mutant 1+2: `for _, line := range backupRetentionReportLines(rep, projectRoot)[:0] {` at update_clean_install.go and update_template_sync.go
$ go test -count=1 -run 'TestCleanReinstall_PrintsBackupRetentionRows|TestTemplateSync_RecordsBackupRetentionRows' ./internal/cli/
--- FAIL: TestCleanReinstall_PrintsBackupRetentionRows (0.05s)
    update_backup_retention_paths_test.go:64: AC-UBR-21: the full-reinstall output has no per-run removal row "[clean-reinstall] old backup run removed: .moai-backups/20250101_000000":
        ... [clean-reinstall] Cleaned up 3 old config backup(s) ...
    update_backup_retention_paths_test.go:67: AC-UBR-21: the full-reinstall output has no retention summary row:
--- FAIL: TestTemplateSync_RecordsBackupRetentionRows (0.39s)
    update_backup_retention_paths_test.go:98: AC-UBR-21: the template-sync ledger has no per-run removal row "old backup run removed: .moai-backups/20250101_000000":
    update_backup_retention_paths_test.go:101: AC-UBR-21: the template-sync ledger has no retention summary row:
FAIL
# mutant 3: `yaml:"backup_past_keep"` on StateConfig.BackupPastVersionsKeep in internal/config/types.go
$ go test -count=1 -run 'TestStructYAMLSymmetry_StateBackupPastVersionsKeep' ./internal/config/
--- FAIL: TestStructYAMLSymmetry_StateBackupPastVersionsKeep (0.00s)
    state_symmetry_key_test.go:60: CONFIG_STRUCT_YAML_MISMATCH: field=StateConfig.backup_past_versions_keep, side=yaml-only (key exists in shipped state.yaml but not in struct)
    state_symmetry_key_test.go:72: shipped state.backup_past_versions_keep decodes to 0, want DefaultBackupPastVersionsKeep (3)
FAIL
```

GREEN on the restored tree (this run, HEAD `96b2a17d5`):

```
$ go test -count=1 -v -run 'TestCleanReinstall_PrintsBackupRetentionRows|TestTemplateSync_RecordsBackupRetentionRows|TestBackupRetentionReportLines|TestDestructiveTargetRegistry' ./internal/cli/
--- PASS: TestCleanReinstall_PrintsBackupRetentionRows (0.04s)
--- PASS: TestTemplateSync_RecordsBackupRetentionRows (0.26s)
--- PASS: TestBackupRetentionReportLines_NamesEveryDecision (0.00s)
--- PASS: TestBackupRetentionReportLines_QuietWhenNothingHappened (0.00s)
--- PASS: TestDestructiveTargetRegistry_CoversAllSites (0.03s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.752s
$ go test -count=1 -cover ./internal/cli/update/backup/... ./internal/cli/update/skillresidue/... ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/cli/update/backup	1.709s	coverage: 88.4% of statements
ok  	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.597s	coverage: 88.4% of statements
ok  	github.com/modu-ai/moai-adk/internal/config	12.405s	coverage: 85.6% of statements
```

M5 acceptance matrix (Exit: AC-UBR-21, 22, 23 green):

| AC | Status | Evidence |
|---|---|---|
| AC-UBR-21 | PASS | the symmetry test above (key-scoped, both directions, plus the struct-backed read) and one rows test per path; each observed failing under the mutants above |
| AC-UBR-22 | PASS | `gofmt -l` on the touched files and on `internal/cli/update internal/config` empty; `go vet ./internal/cli/update/... ./internal/config/ ./internal/cli/` exit 0 (and `GOOS=windows GOARCH=amd64 go vet` on the same packages exit 0, which also compiles their test files for Windows); `golangci-lint run --timeout=5m ./internal/cli/update/... ./internal/config/... ./internal/cli/` reports exactly the three pre-existing errcheck findings in `internal/cli/factory_run_adopt.go` (lines 79, 91, 103) and nothing else; coverage 88.4% / 88.4% (backup / skillresidue). Platform guards: the FIFO test file carries `//go:build unix`; the permission-bit tests in `skillresidue/residue_coverage_test.go` skip on `runtime.GOOS == "windows"` and for root; every symlink test skips when the link cannot be made |
| AC-UBR-23 | PASS | `make build` exit 0 (its `agents-emit-check`, `commands-emit-check`, `tool-policy-drift-check`, `embed-manifest-check`, `templ-generate` and `gen-catalog-hashes.go --all` steps all ran) and `git status --short` afterwards showed no tracked change; `go test -count=1 ./internal/template/` (the whole package) `ok 289.211s`; the tracked-change set is compared with the expected set below |

Tracked changes of the card, `git diff --name-only 7357f0c22 HEAD` taken before the M5 test commit: SPEC `spec.md` (the `status: draft` to `in-progress` line only, M1) and `progress.md`; the template `internal/template/templates/.moai/config/sections/state.yaml` (its embedded counterpart is the same file, the templates are `go:embed`ded from that tree, and `make build` regenerated no tracked file); `internal/config/` (`defaults.go`, `types.go`, `loader_backup_budget.go`, three test files, `testdata/shipped_key_inventory.yaml`); and the touched Go packages `internal/cli/update/backup`, `internal/cli/update/skillresidue` and the update files of `internal/cli`. M5 adds `internal/config/state_symmetry_key_test.go` and `internal/cli/update_backup_retention_paths_test.go` to that set. Nothing outside the expected set appears.

Measured, not claimed: `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0 at the start tree; `grep -rn AskUserQuestion` over the touched non-test files empty (exit 1).
Environment brackets around the whole-template run and the `internal/cli` runs: `~/.gitconfig` 503 lines, sha256 `376c6febc88ba234321078fa8e8c4ec922b458192d1199a31a4eb03b55660548`, the two pre-existing `/var/folders|/repo|Test` lines (479, 480), unchanged before and after; backlog.db header `SQLite format 3`, `PRAGMA quick_check` `ok`. No test added a trace line.
Not run (Gaps): the full `internal/cli` suite (named `-run` tests only) and whole-package `internal/cli` coverage; the key-scoped symmetry check is not a whole-section bijection (`profile_max_bytes` has no struct field and `retention_days` is not shipped, left as found); a retained-with-reason row on either path was not seeded (the seeded runs are empty and removable; the retained-row rendering is covered at the renderer in `update_backup_retention_report_test.go`); Windows runtime behavior (cross-build and vet only); real-data remeasure (M6).

### M6 — Real-data remeasure and close evidence (card tree `WT-10-14-class-3`; production code identical to `5f4b35a64`; harness commit `6eaf221c172947792884b5352118cb5f69cce0df`)

M6 changed no production file. The one file it adds is the env-gated test `internal/cli/update/backup/retention_realdata_test.go` (committed on its own at `6eaf221c1`, before this evidence block), because the gitignored probe A.1 only logs and carries no oracle for AC-UBR-06(b). It is skipped unless `T1733_CLONE` is set. It stamps every run directory with the time its name encodes, prunes through the exported `backup.PruneBackups`, and judges the result from a disk census taken before and after (SHA-256 per regular file, `.template-defaults/` and writer bookkeeping excluded, a non-regular or unreadable entry marks the run exempt); it does not call or re-run the selection.

**Input provenance.** Read-only source: the 10-14 t1687 audit snapshot `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/dfd2478c-11e3-4ea8-a787-c06d8bb4ade9/scratchpad/rc1/moaikr` (533 MB, a clone of the live mo.ai.kr project). Every run used a fresh copy, `cp -Rcp <source> /private/tmp/t1733-m6-<label>/moaikr`, exit 0. Each copy was checked: `.moai-backups` 888 files, `.moai/backups` 29797 files (equal to the source: 888 and 29797), and `stat` on run directory `20261008_203921` read `2026-10-14T16:21:43` in source and copy (modification time preserved; every snapshot directory carries that one copy time, which is why the harness stamps explicit times). The source was never pruned; no `moai update` was run; no binary was installed.

Pre-flight on `5f4b35a64`: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go test -count=1 ./internal/cli/update/backup/ ./internal/config/` `ok 0.983s` and `ok 8.134s`.

#### AC-UBR-06 (a)-(e), harness on a fresh copy, tree `6eaf221c1`

```
$ env T1733_CLONE=/private/tmp/t1733-m6-h5/moaikr T1733_KEEP=5 go test -count=1 -v -run TestRetentionRealData_AC_UBR06 ./internal/cli/update/backup/     (exit 0)
    retention_realdata_test.go:275: AC06a keepCount=5 runs-before=84 runs-after=43 window-present=5 R(retained-excess)=38 R-non-exempt=38 exempt-runs-before=0
    retention_realdata_test.go:277: AC06a unmanaged-dirs before=15 after=15 loose-files before=9 after=9 report-unmanaged-count=16 report-unmanaged-bytes=10613047
    retention_realdata_test.go:312: AC06b protected-versions=1399 paths-over-budget=56 uncovered=0 redundant-retained-non-exempt=0 retained-non-exempt=38 bound-ok=true
    retention_realdata_test.go:340: AC06c distinct-file-versions-before=1814 after=1754 lost=60 paths-with-lost-versions=22 lost-inside-budget=0 lost-in-window=0
    retention_realdata_test.go:359: AC06d pre-clean-files before=48 after=48
    retention_realdata_test.go:374: AC06e report runs=84 window=5 items=79 removed-items=41 retained-items=38 removed-dirs=41 protected=1399
    retention_realdata_test.go:385: AC06e retained-reasons=map[protects-versions:38]
--- PASS: TestRetentionRealData_AC_UBR06 (11.98s)

$ env T1733_CLONE=/private/tmp/t1733-m6-h4/moaikr T1733_KEEP=4 go test -count=1 -v -run TestRetentionRealData_AC_UBR06 ./internal/cli/update/backup/     (exit 0)
    retention_realdata_test.go:275: AC06a keepCount=4 runs-before=84 runs-after=43 window-present=4 R(retained-excess)=39 R-non-exempt=39 exempt-runs-before=0
    retention_realdata_test.go:277: AC06a unmanaged-dirs before=15 after=15 loose-files before=9 after=9 report-unmanaged-count=16 report-unmanaged-bytes=10613047
    retention_realdata_test.go:312: AC06b protected-versions=1400 paths-over-budget=56 uncovered=0 redundant-retained-non-exempt=0 retained-non-exempt=39 bound-ok=true
    retention_realdata_test.go:340: AC06c distinct-file-versions-before=1814 after=1754 lost=60 paths-with-lost-versions=22 lost-inside-budget=0 lost-in-window=0
    retention_realdata_test.go:359: AC06d pre-clean-files before=48 after=48
    retention_realdata_test.go:374: AC06e report runs=84 window=4 items=80 removed-items=41 retained-items=39 removed-dirs=41 protected=1400
    retention_realdata_test.go:385: AC06e retained-reasons=map[protects-versions:39]
--- PASS: TestRetentionRealData_AC_UBR06 (14.89s)
```

Measured figures (R is measured, not a target; Q2 soft cap):

| Figure | keepCount 5 (production) | keepCount 4 |
|---|---|---|
| runs before / after | 84 / 43 | 84 / 43 |
| R, retained excess runs (all non-exempt; no exempt run on this input) | **38** | **39** |
| runs after = keepCount + R | 5 + 38 | 4 + 39 |
| distinct protected versions (bound on R) | 1399 | 1400 |
| namespace-root directories after (main: 79 to 0) | not separately counted: 43 runs remain across both roots | same |
| unmanaged config-root dirs before / after; loose files before / after (both roots) | 15 / 15; 9 / 9 | 15 / 15; 9 / 9 |
| report unmanaged count / bytes (config root: 15 dirs + 1 loose file) | 16 / 10613047 | 16 / 10613047 |
| lost versions (all older than their budget; 0 inside budget, 0 in window) | 60 over 22 paths | 60 over 22 paths |
| pre-clean files before / after | 48 / 48 | 48 / 48 |

The 48-file expectation of REQ-UBR-003 (derived in AC-UBR-06(d)) is now observed: `after=48`.

#### AC-UBR-06(d) fixed-mtime probe P5 (the probe of the AC, unmodified), tree `6eaf221c1`

```
$ env T1733_CLONE=/private/tmp/t1733-m6-p4/moaikr go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup_p5.rel.json -run TestT1733ProbeRealDataFixedMtime ./internal/cli/update/backup/     (exit 0)
    zz_t1733_probe_p5_test.go:108: P5 keepCount=4 runs=84 distinct-run-mtimes=true pre-clean-run=20261008_203921 pre-clean-run-rank=9 pre-clean-run-is-oldest-config-run=true
    zz_t1733_probe_p5_test.go:110: P5 pre-clean-files before=48 after=48 deleted=41 pre-clean-run-exists-after=true
--- PASS: TestT1733ProbeRealDataFixedMtime (12.35s)
$ env T1733_CLONE=/private/tmp/t1733-m6-p5/moaikr T1733_KEEP=5 go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup_p5.rel.json -run TestT1733ProbeRealDataFixedMtime ./internal/cli/update/backup/     (exit 0)
    zz_t1733_probe_p5_test.go:108: P5 keepCount=5 runs=84 distinct-run-mtimes=true pre-clean-run=20261008_203921 pre-clean-run-rank=9 pre-clean-run-is-oldest-config-run=true
    zz_t1733_probe_p5_test.go:110: P5 pre-clean-files before=48 after=48 deleted=41 pre-clean-run-exists-after=true
--- PASS: TestT1733ProbeRealDataFixedMtime (9.30s)
```

RED reference (AC cell, start tree): `before=48 after=0 deleted=80 pre-clean-run-exists-after=false`, exit 1. GREEN now: `after=48`, exit 0.

#### The AC's own (a)-(c) probe A.1 (no mtime stamping), same overlay command as the RED cell

```
$ env T1733_CLONE=/private/tmp/t1733-m6-a1/moaikr go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run TestT1733ProbeRealData ./internal/cli/update/backup/     (exit 0)
    zz_t1733_probe_test.go:217: REAL cfg-root dirs 20->20 ns-root dirs 79->31 deleted=48 pre-clean-dirs 1->1
    zz_t1733_probe_test.go:218: REAL distinct-file-versions-before=1951 lost=93 paths-with-lost-versions=30
    zz_t1733_probe_test.go:220: REAL lost-top path=.moai/harness/usage-log.jsonl versions-lost=30
    zz_t1733_probe_test.go:220: REAL lost-top path=.claude/agents/moai/manager-lead.md versions-lost=6
    zz_t1733_probe_test.go:220: REAL lost-top path=.claude/agents/moai/plan-auditor.md versions-lost=6
--- PASS: TestT1733ProbeRealData (22.94s)
```

RED reference (start `e669da81b`, ledger E6): `ns-root dirs 79->0 deleted=79`, `lost=1700 paths-with-lost-versions=996`. Now `79->31`, `lost=93` over 30 paths. This probe leaves the snapshot's single copy time in place, so run order falls to the path-ascending tie-break (config root first), a different window from the stamped runs above; its census also counts differently (1951 distinct versions against the harness's 1814), so the two `lost` figures (93 and 60) are not comparable with each other, only each with its own RED reference. This probe does not judge inside-budget loss; the harness above does, under stamped ordering.

#### P1-P4b main-versus-tree cells, pruner level, tree `6eaf221c1`

```
$ go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_backup.rel.json -run TestT1733ProbeP1 ./internal/cli/update/backup/     (exit 0)
    P1 pre-clean-dirs-before=[20261011_000000 20261012_000000 20261013_000000 20261014_000000 20261015_000000 20261016_105403] user-file-copies-before=1 deleted=0 user-file-copies-after=1
$ ... -run TestT1733ProbeP2 ...     (exit 0)
    P2 deleted=1 remaining=[update-2026-10-01T00-00-00Z update-2026-10-02T00-00-00Z update-2026-10-03T00-00-00Z update-2026-10-05T00-00-00Z] v1-entry-survives=true v2-entry-survives=true v3-entry-survives=true
$ ... -run TestT1733ProbeP3 ...     (exit 0)
    P3 deleted=9 remaining=[update-2026-10-10T00-00-00Z update-2026-10-11T00-00-00Z] v1-holding-entries-remaining=1
$ go test -count=1 -v -overlay=.moai/reports/t1733/probes/overlay_residue.rel.json -run TestT1733ProbeP4a ./internal/cli/update/skillresidue/     (exit 0)
    P4a err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill} {.agents/skills/moai-kanban-foreman reported-retired-link}] b-link-survives=true a-folder-in-place=true
$ ... -run TestT1733ProbeP4b ...     (exit 0)
    P4b err=<nil> actions=[{.claude/skills/moai-kanban-foreman reported-retired-skill}] content-now-outside-project=false a-folder-in-place=true
```

Against the AC cells: P1 `user-file-copies-after` 0 to 1 (pre-revert reference 1); P2 `deleted=4 ... all false` to `deleted=1 ... all true`; P3 `deleted=10 ... remaining=0` to `deleted=9 ... remaining=1` and the remaining v1 holder is `update-2026-10-10T00-00-00Z`, the newest of the ten; P4a and P4b equal their guard values.

#### AC-UBR-24 — ordering in the commit graph (`git log --reverse 7357f0c22..HEAD`, `git show --stat`)

RED re-observed on the base production code in this run: the committed probe tests run against `backup.go` as of `7357f0c22` (overlay hiding every file this card added to the package):

```
$ go test -count=1 -v -overlay=/private/tmp/t1733-m6-red/overlay.json -run 'TestAC_UBR_01_...|TestAC_UBR_02_...|TestAC_UBR_03_...' ./internal/cli/update/backup/     (exit 1)
    retention_probe_p1_test.go:74: AC-UBR-01 user-file-copies-before=1 deleted=1 user-file-copies-after=0
--- FAIL: TestAC_UBR_01_PreCleanUserFileSurvivesRotation (0.01s)
    retention_probe_p2_test.go:22: AC-UBR-02 deleted=4 remaining=[update-2026-10-05T00-00-00Z]
--- FAIL: TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot (0.01s)
    retention_probe_p3_test.go:22: AC-UBR-03 deleted=10 remaining=[update-2026-10-11T00-00-00Z]
--- FAIL: TestAC_UBR_03_RedundantCopiesAreNotAllKept (0.01s)
FAIL
```

| AC | RED commit (test added, with assertions) | GREEN commit | Order |
|---|---|---|---|
| AC-UBR-01 | `8c7c4cdb5` (probe P1) | `312c303a0` (`backup.go` calls the budgeted pruner) | RED first |
| AC-UBR-02 | `f3d4ad42d` (probe P2); selection tests `716d81b53` | `312c303a0` at pruner level; `815223b5b` at selection level | RED first |
| AC-UBR-03 | `750a8a7fe` (probe P3); selection tests `716d81b53` | `312c303a0`; `815223b5b` | RED first |
| AC-UBR-07 | `716d81b53` | `815223b5b` | RED first |
| AC-UBR-08, 09, 11, 13, 14 | `396484c7f` (tests plus API stubs) | `312c303a0` | RED first |
| AC-UBR-12 | pruner report `396484c7f`; update output `73b390a05` | `312c303a0`; `ab1e48138` | RED first |
| AC-UBR-16, 17 | `f2e692588` | `75d97ab02` | RED first |
| AC-UBR-18 | `471dbc55b` | `9d7f1823c` | RED first |
| AC-UBR-04, 05 (guards) | `3935b4434` | residue change `75d97ab02` | guard precedes the residue change |
| AC-UBR-15 (guard) | `3935b4434` | n/a (green on start by design) | no RED |
| AC-UBR-10 | `d774ab272` carries tests and implementation together | same commit | **shared commit** |
| AC-UBR-19 | registry test existed; the row for the pruner site moves in `312c303a0` with the pruner | same commit | **shared commit** (the M2 registry-row move) |
| AC-UBR-20 | `75d97ab02` carries the pin test, the registry rows and `residue.go` | same commit | **shared commit**; the failure on a deliberately added alias was observed once in M3 (see the M3 block) |
| AC-UBR-21 | `96b2a17d5` | no GREEN commit: closure guards, green on the start tree; failing state observed by mutation (M5 block) | no RED commit |
| AC-UBR-22, 23 | n/a | gates, no test commit | not applicable |

The probe tests of AC-UBR-01 to AC-UBR-03 each precede the commit that makes them pass, and §E.2 above (M1 block) holds the verbatim RED output for each. AC-UBR-10, 19 and 20 cannot be re-witnessed by the commit graph for ordering; their RED observations rest on the M1 and M3 session records only.

#### Quality gates and brackets, tree `6eaf221c1`

```
$ gofmt -l internal/cli/update internal/config internal/cli/update_*.go               (empty, exit 0)
$ go vet ./internal/cli/update/... ./internal/config/ ./internal/cli/                  (exit 0)
$ go build ./...                                                                        (exit 0)
$ GOOS=windows GOARCH=amd64 go build ./...                                              (exit 0)
$ golangci-lint run --timeout=5m ./internal/cli/update/backup/... ./internal/cli/update/skillresidue/... ./internal/config/...   0 issues (exit 0)
$ go test -count=1 -cover ./internal/cli/update/backup/... ./internal/cli/update/skillresidue/... ./internal/config/
ok  github.com/modu-ai/moai-adk/internal/cli/update/backup        1.007s  coverage: 88.4% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/update/skillresidue  0.359s  coverage: 88.4% of statements
ok  github.com/modu-ai/moai-adk/internal/config                   51.035s coverage: 85.6% of statements
$ go test -count=1 -v -run 'TestDestructiveTargetRegistry|TestCleanReinstall_PrintsBackupRetentionRows|TestTemplateSync_RecordsBackupRetentionRows|TestBackupRetentionReportLines|TestRemoveStateResidue|TestBackupSubsystem_DestructiveSurfaces' ./internal/cli/
--- PASS for all nine named tests, ok github.com/modu-ai/moai-adk/internal/cli 1.681s
```

Environment brackets around the `internal/cli` run (before and after identical): `~/.gitconfig` 503 lines, sha256 `376c6febc88ba234321078fa8e8c4ec922b458192d1199a31a4eb03b55660548`, two pre-existing `/var/folders|Test` lines (479, 480); backlog.db header `SQLite format 3`, `PRAGMA quick_check` `ok`. `TMPDIR`, `GOTMPDIR` and `HOME` were not set.

#### AC matrix (24 ACs, this run, tree `6eaf221c1`)

| AC | Proof | Status |
|---|---|---|
| 01 | `TestAC_UBR_01_PreCleanUserFileSurvivesRotation`; probe P1 `after=1` | PASS |
| 02 | `TestAC_UBR_02_CurrentVersionDoesNotTakeABudgetSlot`; `TestAC_UBR_02_SelectionKeepsPastVersionsOutsideTheWindow`; probe P2 | PASS |
| 03 | `TestAC_UBR_03_RedundantCopiesAreNotAllKept`; `TestAC_UBR_03_SelectionRetainsOneCopyOfRepeatedVersion`; probe P3 | PASS |
| 04 | `TestAC_UBR_04_ContentBehindSymbolicAgentsSkillsIsUntouched`; probe P4a | PASS |
| 05 | `TestAC_UBR_05_NothingIsWrittenOutsideTheProjectThroughArchiveLink`; probe P4b | PASS |
| 06 | (a) (b) (c) (e) `TestRetentionRealData_AC_UBR06` at keepCount 5 and 4; (d) probe P5 at 4 and 5 | PASS (R = 38 at 5, 39 at 4) |
| 07 | `TestAC_UBR_07_BudgetBoundsProtection` (4 subtests) | PASS |
| 08 | `TestAC_UBR_08_BookkeepingIsRecognizedByWhatTheWriterWrote` | PASS |
| 09 | `TestAC_UBR_09_TemplateDefaultsRenderIsExempt` | PASS |
| 10 | `TestAC_UBR_10_BudgetKeyShipsOnce`; `TestLoadBackupPastVersionsKeep` | PASS |
| 11 | `TestAC_UBR_11_UnclassifiableContentFailsClosedAndNeverBlocks` | PASS |
| 12 | `TestAC_UBR_12_EachDecisionIsReported`; `TestBackupRetentionReportLines_NamesEveryDecision` | PASS |
| 13 | `TestAC_UBR_13_RootContainmentStillRefuses` (3 subtests) | PASS |
| 14 | `TestAC_UBR_14_ExplicitWipeAndNegativeKeepCount` | PASS |
| 15 | `TestAC_UBR_15_RetiredFoldersAndLinksAreReportedNeverMovedOrRemoved` | PASS |
| 16 | `TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename` | PASS |
| 17 | `TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName`; `..._AtTheDestination` | PASS |
| 18 | `TestRunUpdateHousekeeping_EmptyDBReportedNotRemoved_ACUBR18` (inside the `TestRemoveStateResidue` / housekeeping `-run` set; named tests of the file listed above passed) | NOT-OBSERVED by name in this run; see Gaps |
| 19 | `TestDestructiveTargetRegistry_CoversAllSites` | PASS |
| 20 | `TestDestructiveTargetRegistry_CoversAllSites`; `TestMoveFileReferencesArePinned` | PASS |
| 21 | `TestCleanReinstall_PrintsBackupRetentionRows`; `TestTemplateSync_RecordsBackupRetentionRows`; `TestStructYAMLSymmetry_StateBackupPastVersionsKeep` | PASS |
| 22 | gofmt, vet, golangci-lint above; coverage 88.4 / 88.4 / 85.6 | PASS |
| 23 | not re-run in M6; M5 block holds `make build` and `internal/template` evidence at `96b2a17d5`; no production or template file changed since | NOT-OBSERVED in this run (carried from M5, attributed there) |
| 24 | commit graph table above | PASS-WITH-DEBT (AC-UBR-10, 19, 20 share a commit) |

#### Files for the sync auditor

`git diff --name-only 7357f0c22 HEAD` is the list in the M5 block plus `internal/cli/update/backup/retention_realdata_test.go` (this commit) and this file. Pre-existing tests edited across M1-M5 (status `M`, from `git diff --name-status 7357f0c22 HEAD -- '*_test.go'`):

| File | Reason |
|---|---|
| `internal/cli/update/skillresidue/residue_coverage_test.go` | the `moveFile` seam is replaced by the project-root-bound rename (REQ-UBR-017); tests re-targeted to the new seam |
| `internal/cli/update/skillresidue/residue_moveone_test.go` | the pin test counted references to `moveFile` / `moveOne`; it now counts the root-handle move calls (REQ-UBR-019, AC-UBR-20) |
| `internal/cli/update/skillresidue/residue_safety_test.go` | a test injected a failing `moveFile`; re-expressed against the root-bound rename (AC-UBR-16, 17) |
| `internal/cli/update_destructive_registry_test.go` | the scan counted only `os.RemoveAll` / `os.Rename`; extended to the root-handle operations so a changed move site cannot be missed (F22, AC-UBR-20) |
| `internal/cli/update_preserve_reach_test.go` | the rotation subtest wrote a distinct body per backup; with budgeted protection each would be a last holder, so every backup now holds the same bytes and survivors are identified by directory name (REQ-UBR-003) |
| `internal/cli/update_state_residue_test.go` | zero-byte database files were expected removed; now expected present and reported (REQ-UBR-018, operator ruling Q1); the `backlog.json.lock` removal is unchanged |

Also modified, not a test: `internal/config/testdata/shipped_key_inventory.yaml` (the shipped state key inventory gained `backup_past_versions_keep`).

#### Post-landing operator procedure (design D11), close evidence for the operator — NOT executed by this run

1. Before anything else, copy `.moai-backups` and `.moai/backups` of the real project with modification times preserved (macOS APFS `cp -Rcp`; elsewhere `cp -a`) to a directory outside the project, and note the directory count and the recursive file count of each root.
2. Make a second copy of the whole project (a throwaway) and run `moai update` there, never in the original.
3. In the copy, list both backup roots again and compare with this block: the config root keeps its window of five runs plus any reported retained excess; the namespace root is no longer emptied (main empties it, 79 to 0); the update summary names every removed run and every retained excess run with its reason. Reference figures on the 10-14 snapshot at keepCount 5: 84 runs before, 43 after, R = 38, 1399 protected versions, 48 pre-clean files kept.
4. Confirm at least one pre-clean user-file copy remains (a file under a `pre-clean/` directory whose bytes match a user file the template does not ship).
5. Confirm the retained-excess count the update reports equals the number of run directories present beyond the window in the copy, and does not exceed the number of distinct protected versions.
6. If any step fails, keep the original project and the step-1 backup untouched and report the numbers to card t1733.

#### Gaps (not observed in M6)

- The command-refusal record: the worktree-isolation guard refused three shell shapes this run (a `find` over a path held in a shell variable; `go test -run '^Name$'` with anchors; a `go test ... > file; grep` pipeline). Each was re-issued as plain separate commands with literal paths and unanchored `-run` names; no measurement rests on a refused call.
- AC-UBR-23 (`make build`, `internal/template` suite) was not re-run in M6; its evidence is the M5 block at `96b2a17d5`, and no template or production file changed afterwards (`git diff --stat 5f4b35a64 HEAD` lists only the new test file).
- AC-UBR-18's test is `TestRunUpdateHousekeeping_EmptyDBReportedNotRemoved_ACUBR18`; the `-run` set used here named `TestRemoveStateResidue` and `TestBackupSubsystem_DestructiveSurfaces` and did not name it, so it was not observed by name in this run; its GREEN at `9d7f1823c` is in the M4 block.
- The full `internal/cli` suite and the whole-repository suite were not run; the repository-wide test verdict belongs to the CI run on the integration branch and is PENDING at report time.
- Windows and Linux runtime behavior of the new tests: build and vet only for Windows; the harness and the probes ran on darwin.
- The oracle uses the spec's recency definition (newest excess holder, ties broken on the digest); a different tie rule could change which of two equally recent versions is protected, and would not change the counts above on this input unless a boundary tie occurs. Boundary ties were not enumerated.
- `moai update` end to end on a real project (design D11) is the operator's act and was not run.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-16
run_commit_sha: pending-backfill   # the M6 evidence commit cannot cite its own hash
run_status: complete-pending-sync-audit
harness_commit_sha: 6eaf221c172947792884b5352118cb5f69cce0df
measurement_tree_sha: 6eaf221c172947792884b5352118cb5f69cce0df
ac_pass_count: 23   # corrected 2026-10-16 from 22: the sync audit (.moai/reports/t1733/sync-audit.md, finding F3) re-derived 23 PASS + AC-UBR-24 PASS-WITH-DEBT = 24 AC
ac_fail_count: 0
ac_pass_with_debt: [AC-UBR-24]   # AC-UBR-10, 19 and 20 share a commit between test and implementation
ac_not_observed_in_m6: [AC-UBR-18 by name, AC-UBR-23]   # evidence carried from M4 and M5 blocks
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run   # card branch unpushed; no remote contact
l44_post_push_fetch: not-run    # no push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_darwin: pass
  windows_amd64: pass
total_run_phase_files: 59       # git diff --name-only 7357f0c22 HEAD, plus this file already counted
m1_to_mN_commit_strategy: RED test commit before each GREEN commit; evidence commit per milestone; unpushed
retained_excess_R: {keepCount_5: 38, keepCount_4: 39}
repository_wide_test_verdict: PENDING   # owned by the CI run on the integration branch
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-16
sync_commit_sha: 524c581fe6db94dc9de452e34ef6c79d73fead11
sync_status: complete-pending-sync-audit
sync_base_head: ecc040ccfa1064aee2ec34c31770a69830275ff8   # tree HEAD the sync commit sits on (main f13fed1c4 already absorbed)
ac_source: .moai/specs/SPEC-UPDATE-BACKUP-RETENTION-001/acceptance.md   # tier L
b12_self_test_a: "pass   # grep -c 'SPEC-UPDATE-BACKUP-RETENTION-001' CHANGELOG.md printed 0 before emission, 1 after"
b12_self_test_b: "pass   # AC counter on acceptance.md: stdout 24, stderr live=24 excluded=0 ambiguous=0; CHANGELOG states AC-UBR-01..24 (24)"
b12_self_test_c: "pass   # ls verified: spec.md, internal/cli/update_destructive_registry.go, internal/cli/update/backup/retention_realdata_test.go, internal/template/templates/.moai/config/sections/state.yaml, docs-site/content/{en,ko,ja,zh}/cli-reference/update.md"
changelog_entry_position: "[Unreleased] > ### Fixed, first bullet"
docs_site: unchanged   # cli-reference/update.md x 4 locales describes only the pre-clean backup; no retention statement to correct
readme: unchanged      # README.md backup lines concern moai clean --home, not moai update retention
frontmatter_status_transitions:
  spec_md: "in-progress -> completed (single sync commit; status + updated only; updated was already 2026-10-16)"
  plan_md: n/a   # no frontmatter status field
  acceptance_md: n/a
canary_compliance_check: n/a   # this SPEC defines no forward-looking policy that its own sync tests
build_after_edit:
  go_build: "exit 0 (go build ./...)"
  go_build_windows_amd64: "exit 0 (GOOS=windows GOARCH=amd64 go build ./...)"
observation_for_leader: "E.3 records ac_pass_count 22; the M6 matrix has 21 PASS rows, AC-UBR-24 PASS-WITH-DEBT, and AC-UBR-18 / AC-UBR-23 carried from M4 / M5 (NOT-OBSERVED in M6). E.3 left untouched (manager-develop owned)."
```

## §E.5 Repair-phase evidence (card t1733, sync-audit finding F1; owner manager-develop)

Source: `.moai/reports/t1733/sync-audit.md` (local-only). Repaired: F1 (REQ-UBR-017, held-back rename overwrote a file created inside the staging folder), F2 (registry text, with F1), F3 (E.3 counter), F4 (CHANGELOG sentence). Tree `WT-10-14-class-3`, start HEAD `69da9c172`.

| Step | Commit | Evidence |
|---|---|---|
| RED | `4db47ed51` | `go test -count=1 ./internal/cli/update/skillresidue/ -run AC_UBR_17` -> `--- FAIL: TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsInsideTheStagingFolder`: "the user's file inside the staging folder was overwritten; tree now [. dir SKILL.md file] (Clean err: <nil>)" (same result as the audit overlay repro) |
| GREEN | `42ab64100` | `go test -count=1 ./internal/cli/update/skillresidue/` -> `ok ... 0.345s` (the whole package) |
| Mutant (plain Rename restored in the first step) | overlay, scratch `/private/tmp/t1733-repair/ov/` | `go test -count=1 -overlay=/private/tmp/t1733-repair/ov/overlay.json ./internal/cli/update/skillresidue/ -run AC_UBR_17` -> `--- FAIL: TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsInsideTheStagingFolder`, exit non-zero |

Choice for filesystems without hard links: none added. The rename's final step (link to SKILL.md) already requires hard links, so a refused first link leaves `skill.md` untouched, removes the empty staging folder, and returns the error. No fallback path can replace an existing destination.

Existing tests adapted, not weakened: three tests hooked the `moveFile` seam for the first move and now hook `linkFile` at the same point (`TestClean_RenameFailureIsReturnedAndNothingMoves`, `TestRenameSkillFile_RestoreFailureIsReported`, `TestAC_UBR_17_...AtTheDestination`, which now plants SKILL.md just before the final link because on a case-insensitive filesystem a SKILL.md written while `skill.md` still exists is the same file); `TestMoveFileReferencesArePinned` counts updated (`linkFile@renameSkillFile` 2, `dropFile@renameSkillFile` 7, `moveFile@renameSkillFile` removed).

Verification (env: GIT_CONFIG_GLOBAL temp file, TMPDIR/GOTMPDIR under /private/tmp/t1733-repair): `go build ./...` exit 0; `gofmt -l internal/cli internal/cli/update internal/template` empty; `go vet ./internal/cli/update/...` exit 0; `go test -count=1 ./internal/cli/update/...` all 7 packages ok; `go test -count=1 ./internal/cli/ -run 'Backup|Retention|Destructive|Housekeeping|UBR|Unmanaged|TemplateSync|CleanReinstall'` ok 19.446s; `go test -count=1 ./internal/template/...` ok (template 137.003s); `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

Gaps: repository-wide test verdict PENDING (CI on the integration branch); `make build` not run; re-audit of the F1 delta is the auditor's.

## §E.6 Repair-round-2 evidence (card t1733, sync-audit iteration 2 finding F7 + F8 text; owner manager-develop)

Source: `.moai/reports/t1733/sync-audit-iter2.md` (local-only). Repaired: F7 (P2, blocking; REQ-UBR-017 source name: an atomic save over `skill.md` between the link and the drop was deleted while the step returned success), F8 (P3; registry `moveOne` row and the `moveOne` comment). Tree `WT-10-14-class-3`, start HEAD `c0f8edbb4`.

| Step | Commit | Evidence |
|---|---|---|
| RED | `3ee61a4dd` | `go test -count=1 ./internal/cli/update/skillresidue/ -run AtomicallyReplaced` -> `--- FAIL: TestAC_UBR_17_RenameDoesNotDropASkillFileThatWasAtomicallyReplaced`: "the rename reported success although skill.md was replaced mid-step; tree now [. dir SKILL.md file]" (the user's file gone, old content promoted; same outcome as the audit overlay repro) |
| GREEN | `6d2c46b87` | `go test -count=1 ./internal/cli/update/skillresidue/` -> `ok ... 0.289s` (whole package); `go test -count=1 ./internal/cli/ -run DestructiveTargetRegistry` -> `ok ... 0.882s` |
| Mutant (SameFile result forced to true in `sameFileAt`) | overlay JSON and mutated copy under `/private/tmp/t1733r2/mutant/` | `go test -count=1 -overlay=/private/tmp/t1733r2/mutant/overlay.json ./internal/cli/update/skillresidue/ -run AtomicallyReplaced` -> `--- FAIL: TestAC_UBR_17_RenameDoesNotDropASkillFileThatWasAtomicallyReplaced`: "the rename reported success although skill.md was replaced mid-step", exit non-zero |

The repair: `renameSkillFile` re-reads `skill.md` and the staged link with `Lstat` through the root handle and requires `os.SameFile` before dropping `skill.md`. When they differ, `skill.md` (the user's new file) is left in place, our extra staged link and the empty stage folder are removed, and the step returns an error naming the situation. When `skill.md` cannot be inspected (gone or unreadable), nothing is dropped, because the staged name may be the only one left; the error names where the content is. The existing rollback for a failed drop is unchanged. `TestMoveFileReferencesArePinned`: `dropFile@renameSkillFile` 7 -> 9 (the two drops on the replaced path); no new seam.

Residual window, stated plainly: a window of microseconds remains between the `Lstat` pair and the unlink of `skill.md`. Go's `os.Root` offers no rename that refuses to replace a name, so the look and the unlink cannot be made one atomic step; an atomic save landing exactly in that instant is still lost. The same statement sits in the `renameSkillFile` doc comment and the `dropOne` registry row. This narrows the F7 window; it does not close it.

F8: the registry `moveOne` row and the `moveOne` comment now say the fallback Rename runs whenever the restore link fails and the name is seen free (any failure reason), not "only where hard links are unavailable". No behavior changed.

Known P3 carried to a follow-up (not repaired here): F2, `putBack` does `root.Lstat(lower)` then a real `Rename`; a file created between the two is replaced. Reachability is very low (the first link already proved hard links work, so the restore link normally fails with EEXIST and the fallback is skipped). The audit's one-step fix (delete `moveFile`, `moveOne`, and the registry row, leaving a failed restore staged and named in the error) is the follow-up candidate; accepted debt for this card.

Verification (env: GIT_CONFIG_GLOBAL temp file holding only [user]; TMPDIR/GOTMPDIR under /private/tmp/t1733r2; no HOME override): `go build ./...` exit 0; `gofmt -l internal/cli internal/cli/update internal/template` empty; `go vet ./internal/cli/update/...` exit 0; `go test -count=1 ./internal/cli/update/...` all 7 packages ok; `go test -count=1 ./internal/cli/ -run 'Backup|Retention|Destructive|Housekeeping|UBR|Unmanaged|TemplateSync|CleanReinstall'` ok 16.812s; `go test -count=1 ./internal/template/...` ok (template 130.604s); `GOOS=windows GOARCH=amd64 go build ./...` exit 0.

Gaps: repository-wide test verdict PENDING (CI on the integration branch); `make build` not run; no guard refused any command in this round; re-audit of the F7 delta is the auditor's.

## §E.7 Repair-round-3 evidence (card t1733, sync-audit iteration 3 finding F11 + F12 test + F2; owner manager-develop)

Source: `.moai/reports/t1733/sync-audit-iter3.md` (local-only). Start HEAD `1828fb4f3`. Why a systematic pass: three audit rounds in a row found a new narrow data-loss branch in `renameSkillFile` (F1 overwrite, F7 atomic save replaced the original name, F11 staged name replaced). This round replaces the pair comparison by a per-name identity rule: the identity (`Lstat`) of `skill.md` is captured before the link and of the staging folder when it is claimed; a name is removed only while `Lstat` still reports that identity (`dropOurs`); anything unproven is left in place and reported.

### State table (mirrored from the code comment above `renameSkillFile`)

L = `skill.md`, T = staged name, S = staging folder; "ours" = same file as the original captured before the link.

| State after the first link | Action | User data | Test |
|---|---|---|---|
| L ours, T ours | continue: drop L, link SKILL.md, drop T (if ours), drop S (if ours) | none lost | `TestRenameSkillFile_*`, `TestClean_*`, whole package |
| L replaced (atomic save), T ours | keep L; drop T (our link) and S; error | L is the user's new file; the old inode loses only the extra name the editor's save already unlinked | `TestAC_UBR_17_RenameDoesNotDropASkillFileThatWasAtomicallyReplaced` |
| T replaced, L ours (F11) | keep L, T, S; error names both | foreign T and original L kept | `TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedName` |
| L replaced, T replaced | keep L, T, S; error names both | both foreign files kept | `TestAC_UBR_17_RenameLeavesEverythingWhenBothNamesWereReplaced` |
| L gone (F12) | keep T, S; error names T | content kept at T | `TestAC_UBR_17_RenameKeepsTheContentStagedWhenSkillMdIsGone` |
| T gone | keep L, S; error names both | content kept at L | `TestAC_UBR_17_RenameLeavesSkillMdWhenTheStagedNameIsGone` |
| T unreadable (S loses search permission) | keep L, T, S; error names both | nothing removed | `TestAC_UBR_17_RenameLeavesEverythingWhenTheStagedNameIsUnreadable` |
| L unreadable, T readable | unreachable: T lies under S inside L's folder, so any permission that blocks the lookup of L blocks T too; "L and T unreadable" takes the same branch as "T unreadable" (`err != nil` on either lookup) | n/a | covered by the T-unreadable row |
| drop of L fails | keep L; drop T only if ours, then S; error | none lost | `TestAC_UBR_17_RenameTakesBackItsLinkWhenTheDropOfSkillMdFails` |
| undo of T fails (drop of L and drop of T both refused) | keep L, T, S; error | content kept at L and T | `TestAC_UBR_17_RenameReportsWhenTheUndoOfTheStagedLinkFails` |
| SKILL.md link refused, restore link works, T replaced before its drop | keep foreign T; skill.md is back; error | foreign T kept | `TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedNameBeforeTheRestoreDrop` |
| SKILL.md link refused, restore link refused, skill.md free (F2) | keep T, S; error names T; no fallback move | content kept at T | `TestAC_UBR_17_RenameLeavesTheContentStagedWhenTheRestoreLinkIsRefused` |
| SKILL.md link refused, user created skill.md before the restore | restore link fails on the taken name; keep user's skill.md and T; error names T | user's file kept, content kept at T | `TestAC_UBR_17_RenameKeepsAFileTheUserCreatedAtSkillMdDuringTheRestore` |
| SKILL.md linked, T replaced before its drop | keep foreign T; SKILL.md holds the content; error | foreign T kept | `TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedNameBeforeTheFinalDrop` |

### RED / GREEN / mutant

| Step | Commit | Command and result |
|---|---|---|
| RED | `99f752bfda52d87d932bb43aaf15215aba069699` | `go test -count=1 ./internal/cli/update/skillresidue/ -run 'AC_UBR_17_Rename(Leaves\|Keeps\|Takes\|Reports)'`, exit 1, 7 of 12 new tests FAIL on the unchanged code (verbatim below) |
| GREEN | `b5b7f3714e913b2e4d84c66f9295ca9db867eac0` | same command, all 12 new tests and the 3 older `AC_UBR_17` tests PASS; whole package `ok` |
| Mutant (old pair comparison restored, drops unconditional) | overlay JSON `/private/tmp/t1733r3/mutant/overlay.json`, mutated copy `/private/tmp/t1733r3/mutant/residue.go` | `go test -count=1 -overlay=/private/tmp/t1733r3/mutant/overlay.json ./internal/cli/update/skillresidue/ -run 'AC_UBR_17'` -> FAIL `TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedName` and `TestAC_UBR_17_RenameLeavesEverythingWhenBothNamesWereReplaced` ("read .../skill.md.tmp-rename/skill.md: no such file or directory": the foreign staged file was deleted), exit non-zero |

RED output, verbatim (tmp path prefix `/private/tmp/t1733r3/gotmp/` shortened on three lines):

```
--- FAIL: TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedName (0.00s)
    residue_identity_test.go:90: read .../skill.md.tmp-rename/skill.md: open .../skill.md.tmp-rename/skill.md: no such file or directory
--- FAIL: TestAC_UBR_17_RenameLeavesEverythingWhenBothNamesWereReplaced (0.00s)
    residue_identity_test.go:116: read .../skill.md.tmp-rename/skill.md: open .../skill.md.tmp-rename/skill.md: no such file or directory
--- FAIL: TestAC_UBR_17_RenameLeavesSkillMdWhenTheStagedNameIsGone (0.00s)
    residue_identity_test.go:163: error = statat skill.md.tmp-rename/skill.md: no such file or directory (the content is at skill.md.tmp-rename/skill.md); want one naming both names
--- FAIL: TestAC_UBR_17_RenameLeavesEverythingWhenTheStagedNameIsUnreadable (0.00s)
    residue_identity_test.go:188: error = statat skill.md.tmp-rename/skill.md: permission denied (the content is at skill.md.tmp-rename/skill.md); want one naming both names
--- FAIL: TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedNameBeforeTheFinalDrop (0.00s)
    residue_identity_test.go:261: the rename reported success although the staged name was replaced before its drop
--- FAIL: TestAC_UBR_17_RenameLeavesAForeignFileThatReplacedTheStagedNameBeforeTheRestoreDrop (0.00s)
    residue_identity_test.go:291: read .../skill.md.tmp-rename/skill.md: open .../skill.md.tmp-rename/skill.md: no such file or directory
--- FAIL: TestAC_UBR_17_RenameLeavesTheContentStagedWhenTheRestoreLinkIsRefused (0.00s)
    residue_identity_test.go:314: error = injected link refusal; want one naming skill.md.tmp-rename/skill.md
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.080s
FAIL
```

Reading the RED: four tests fail on real data loss (F11 staged name replaced; both replaced; staged name replaced before the final drop; staged name replaced before the restore drop: the foreign file was deleted, or success was reported). Two fail on the error wording only (staged gone, staged unreadable: the old error named the staged path but not `skill.md`, the requirement is both). One fails on F2 (the old fallback Rename put the file back by a move that can replace a name). The F12 test (`...KeepsTheContentStagedWhenSkillMdIsGone`) already passed on the old code, as the audit said: it is added as a regression pin, not as a RED. The tests for "drop of L fails", "undo of T fails" and "user created skill.md before the restore" also passed already (their branches were data-safe); they are pins.

### F2 (deleted, not carried)

`putBack`, `moveFile`, `moveOne` and the `moveOne` registry row are deleted. `grep -rn "moveOne\|moveFile\|putBack" internal/` after the change finds no code reference: the remaining hits are the tripwire names in `residue_moveone_test.go` (`"moveFile": true, "moveOne": true` kept in the pinned set so that any reappearance of a Rename seam fails the pin test, plus its comments), one word in the `dropOne` registry prose in `internal/cli/update_destructive_registry.go` ("the moveOne row ... removed"), and the unrelated `moveFile` in `internal/cli/migrate_profiles.go` (a different package, its own function). Observed with `grep -rn "moveOne\|moveFile\|putBack" internal/`, filtering the test-comment lines. The restore is now an inline `linkFile(staged, lower)` that cannot replace a name; a refused restore leaves the content staged and the error names it. The registry has one row fewer (`moveOne`), `dropOne` row rewritten for the identity rule.

### Existing tests edited (no assertion weakened)

| Test / file | Edit | Reason |
|---|---|---|
| `TestRenameSkillFile_RestoreFailureIsReported` (`residue_coverage_test.go`) | removed the `moveFile` seam replacement and the comment words "and the fallback move" | `moveFile` no longer exists; every assertion (error contains "restoring" and the staging folder, content at the staged path) unchanged and still passing |
| `TestMoveFileReferencesArePinned` (`residue_moveone_test.go`) | counts: `linkFile@renameSkillFile` 2 -> 3, `dropFile@renameSkillFile` 9 -> 1, new `dropFile@dropOurs` 1, removed `linkFile@putBack`, `dropFile@putBack`, `moveFile@putBack`, `moveFile@package-level`, `moveOne@package-level`; `moveFile` and `moveOne` stay in the pinned set so any reference now fails as a "new reference" | the drops moved into `dropOurs`, `putBack` and the Rename seam were deleted; the test still fails on any new reference to a seam |
| registry (`update_destructive_registry.go`) | `moveOne` row removed, `dropOne` row rewritten | `TestDestructiveTargetRegistry_CoversAllSites` reports a stale row otherwise; the one remaining destructive site is `dropOne` |

### Honest residual

A window of microseconds remains between each identity check and the unlink it guards (the second look before the drop of `skill.md`, and the `Lstat` inside `dropOurs` before each other drop): a replacement landing exactly there is removed with the name. Go's `os.Root` offers no rename or remove that refuses to replace or that compares identity, so look and unlink cannot be made one atomic step. A recycled inode number (a foreign writer removes our empty staging folder and creates another entry that reuses its inode) would also read as "ours" on filesystems that recycle inodes immediately; that needs two foreign operations inside the same microsecond window. The old inode in the "L replaced" row loses its extra name by design (the editor's save already unlinked it from `skill.md`).

### Verification

Env: `GIT_CONFIG_GLOBAL=/private/tmp/t1733r3/gitconfig` (only `[user]`), `TMPDIR=/private/tmp/t1733r3/tmp`, `GOTMPDIR=/private/tmp/t1733r3/gotmp`, no HOME override. Each command run in this round, results as observed:

```
go build ./...                                   -> exit 0
gofmt -l internal/cli internal/cli/update internal/template -> (empty), exit 0
go vet ./internal/cli/update/...                 -> exit 0
go test -count=1 ./internal/cli/update/...      -> ok update 1.039s, backup 1.023s, deploy 0.533s, merge 0.994s, plan 0.251s, report 0.256s, skillresidue 0.187s
go test -count=1 ./internal/cli/ -run 'Backup|Retention|Destructive|Housekeeping|UBR|Unmanaged|TemplateSync|CleanReinstall' -> ok internal/cli 20.074s
go test -count=1 ./internal/template/...         -> ok template 138.560s, agentemit 0.400s, commandemit 0.244s, embedemit 0.259s, scripts [no test files]
GOOS=windows GOARCH=amd64 go build ./...         -> exit 0
```

Gaps: repository-wide test verdict PENDING (CI on the integration branch); `make build` not run; no guard refused any command in this round; the T-unreadable test skips under root and on Windows (`requireUnprivileged`); re-audit of the F11 delta is the auditor's.

## §E.8 Repair-round-4 evidence (card t1733, sync-audit iteration 4: codex P1 on claimStage and dropOurs; owner manager-develop)

Source: `.moai/reports/t1733/sync-audit-iter4.md` (local-only). Start HEAD `4774a0b82`. The iteration-4 auditor held that the two windows (stage folder swapped between `Mkdir` and `Lstat`; replacement between the identity check in `dropOurs` and the unlink) cannot be closed with `os.Root`. That holds for `os.Root` only. The platforms have an atomic no-replace rename, and `golang.org/x/sys` is already a direct dependency (`go list -m -f '{{.Version}} indirect={{.Indirect}}' golang.org/x/sys` -> `v0.49.0 indirect=false`; the brief said v0.48.0, the module is v0.49.0). `go.mod` and `go.sum` are untouched. Symbols checked by compiling for each GOOS (`go vet` for linux, darwin, windows, freebsd): `unix.Renameat2` + `unix.RENAME_NOREPLACE` (linux), `unix.RenamexNp` + `unix.RENAME_EXCL` (darwin), `windows.MoveFileEx` + `windows.ERROR_NOT_SUPPORTED` (windows).

### Design

`renameSkillFile` now moves the file by atomic no-replace renames: A claim a staging folder S with a random name (`skill.md.tmp-rename-<12 hex>`); B `skill.md -> S/skill.md`; C `S/skill.md -> SKILL.md`; D `rmdir S` (`syscall.Rmdir`: cannot take a file or a non-empty folder). B and C stay two steps because skill.md and SKILL.md are one name on APFS/NTFS. `renameNoReplace(from, to)` is a package seam over the per-platform `noReplaceRename` (`rename_noreplace_{linux,darwin,windows,other}.go`); "exists" is the OS error (`errors.Is(err, fs.ErrExist)`), "cannot" is `errNoReplaceUnsupported`. No user-visible name is unlinked by this path, so window 2 has no unlink to race; window 1 can no longer cause a removal (rename into a swapped parent fails ENOTDIR or lands the file in the foreign folder, and D is rmdir-only). Because the renames are path-based, each is preceded by a root-bound `Lstat` of its source (`moveNoReplace`), which keeps the F24 refusal of a skill folder swapped for a link out of the project. The round-3 link and identity path moved unchanged into `renameByLinks` and runs only when the first rename reports `errNoReplaceUnsupported`.

### State table (mirrored from the code comment above `renameSkillFile`)

L = skill.md, S = staging folder, T = S/skill.md.

| Step | State | Action | User data | Test |
|---|---|---|---|---|
| A | name taken | try another name; 100 taken: error | nothing changed | `TestRenameSkillFile_NoFreeTemporaryNameIsRefused`, `TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsAtTheTemporaryName` |
| A | Mkdir fails | error | nothing changed | `TestClean_RenameFailureIsReturnedAndNothingMoves` (fallback seam) |
| A/B | S swapped for a regular file | B fails ENOTDIR; D fails; L stays; error names S | foreign file kept | `TestAC_UBR_17_RenameLeavesAFileThatReplacedTheStageFolderBeforeTheFirstMove` |
| A/B | S swapped for a foreign folder | B lands T in it; C completes; D fails if the folder holds anything else, error names it | content not lost; foreign files kept | `TestAC_UBR_17_RenameKeepsTheContentsOfAFolderThatReplacedTheStageFolder` (pin) |
| A/B | skill folder swapped for a link out of the project | root-bound look refuses; nothing moves | outside untouched | `TestAC_UBR_16_RenameDoesNotFollowASkillFolderSwappedAfterTheStageWasClaimed` (pin), `TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename` |
| B | a foreign file sits at T | B refused; L untouched; D fails (S not empty); error names T | foreign file and L kept | `TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsInsideTheStagingFolder` |
| B | L replaced (editor's atomic save) before B | B moves whatever is at L | the user's newest file becomes SKILL.md | `TestAC_UBR_17_RenameMovesTheUsersNewFileWhenSkillMdWasSavedBeforeTheFirstRename` |
| B | B fails for another reason | L untouched; rmdir S; error | none lost | `TestAC_UBR_17_RenameLeavesSkillMdUntouchedWhenTheFirstRenameFails` |
| B | B reports `errNoReplaceUnsupported` | nothing moved; `renameByLinks` takes over with the same S | see fallback table | `TestAC_UBR_17_RenameFallsBackToLinksWhenNoReplaceRenameIsUnsupported`, `TestAC_UBR_17_FallbackKeepsAForeignStagedFileUnderARandomStageName` |
| B-C | T overwritten by a foreign atomic write | C moves whatever holds T; this step removes nothing | the content of L was overwritten by the foreign write, not by this step (see Honest residual) | `TestAC_UBR_17_RenameMovesWhateverHoldsTheStagedNameAndRemovesNothing` |
| B-C | a file appears at SKILL.md (case-folding FS: at skill.md) | C refused; D puts T back to L; error names SKILL.md; if D is refused too the error names T | user's file kept; content back at L, or staged at T when L is taken | `TestAC_UBR_17_RenameKeepsAFileThatAppearsAtSkillMdBetweenTheTwoRenames`, `...AtSkillMdLowerBetweenTheTwoRenames`, `TestAC_UBR_17_RenameLeavesTheContentStagedWhenThePutBackIsRefused` |
| C | C fails for another reason | D puts T back; error | none lost | `TestAC_UBR_17_RenamePutsTheFileBackWhenTheSecondRenameIsRefused` |
| C | C and D both fail | content stays at T; error names T | content kept | `TestAC_UBR_17_RenameLeavesTheContentStagedWhenBothLaterRenamesFail` |
| C-D | S swapped for a file after C | D fails ENOTDIR; error names S | foreign file kept | `TestAC_UBR_17_RenameLeavesAFileThatReplacedTheStageFolderBeforeItsRemoval` |
| C-D | S swapped for a non-empty folder | D fails ENOTEMPTY; error names S | foreign files kept | same shape as the foreign-folder row |
| all | no user-visible name is unlinked | `dropFile` is never called on this path | no check-to-unlink window | `TestAC_UBR_17_RenameUnlinksNoNameWhenTheFilesystemCanRenameWithoutReplacing` |
| A | stage name unpredictable | 12 random hex characters, distinct per run | a swap needs a guess | `TestAC_UBR_17_RenameStageFolderNameIsUnpredictable` |

Fallback (`renameByLinks`) keeps the round-3 table in E.7 unchanged, now with random stage names in production. Its rows are exercised by the round-3 tests, re-pointed with `forceFallback`.

### RED / GREEN / mutant

| Step | Commit | Command and result |
|---|---|---|
| RED | `667a21209919483f72b90457d7f0f5339c7fd0a0` | `go test -count=1 -v -run 'NoReplace\|AC_UBR_17\|AC_UBR_16\|Fallback\|RenameSkillFile\|MoveFileRef\|Clean_Rename' ./internal/cli/update/skillresidue/` exit 1: 16 FAIL, 25 PASS |
| GREEN | `eb0a6a0048f3f9a86e685e5c30b3cc5ad7e2e237` | same command exit 0: 41 PASS, 0 FAIL, 0 SKIP; package `ok`, coverage 89.4% |

RED stands on a stub file (`rename_noreplace_stub.go`, deleted in GREEN): the primitive returns "not implemented" and the stage suffix is empty, so the tests run against the unchanged rename instead of failing to compile. Verbatim RED (path prefix `/private/tmp/t1733r4/gotmp/...` shortened; each message line precedes its `--- FAIL` line):

```
    residue_noreplace_test.go:106: rename onto a free name: noReplaceRename: not implemented
--- FAIL: TestNoReplaceRename_MovesWhenTheDestinationIsFree
    residue_noreplace_test.go:128: error "noReplaceRename: not implemented" must be a destination-exists error and not the unsupported sentinel
--- FAIL: TestNoReplaceRename_RefusesAnExistingDestination
    residue_noreplace_test.go:167: read .../skill.md.tmp-rename: open .../skill.md.tmp-rename: no such file or directory
--- FAIL: TestAC_UBR_17_RenameLeavesAFileThatReplacedTheStageFolderBeforeTheFirstMove
    residue_noreplace_test.go:237: stage folder name "skill.md.tmp-rename" does not match ^skill\.md\.tmp-rename-[0-9a-f]{12}$
--- FAIL: TestAC_UBR_17_RenameStageFolderNameIsUnpredictable
    residue_noreplace_test.go:255: dropFile("skill.md"): the no-replace path must not unlink a name behind a check
    residue_noreplace_test.go:255: dropFile("skill.md.tmp-rename/skill.md"): the no-replace path must not unlink a name behind a check
    residue_noreplace_test.go:255: dropFile("skill.md.tmp-rename"): the no-replace path must not unlink a name behind a check
--- FAIL: TestAC_UBR_17_RenameUnlinksNoNameWhenTheFilesystemCanRenameWithoutReplacing
    residue_noreplace_test.go:284: the rename made 0 calls of the no-replace seam, want at least 2
--- FAIL: TestAC_UBR_17_RenameKeepsAFileThatAppearsAtSkillMdBetweenTheTwoRenames
    residue_noreplace_test.go:312: the rename never called the no-replace seam: the hook did not run
--- FAIL: TestAC_UBR_17_RenameKeepsAFileThatAppearsAtSkillMdLowerBetweenTheTwoRenames
    residue_noreplace_test.go:336: no-replace seam calls = 0, want 1
--- FAIL: TestAC_UBR_17_RenameLeavesSkillMdUntouchedWhenTheFirstRenameFails
    residue_noreplace_test.go:363: no-replace seam calls = 0, want 3 (move, refused move, put back)
--- FAIL: TestAC_UBR_17_RenamePutsTheFileBackWhenTheSecondRenameIsRefused
    residue_noreplace_test.go:392: error = <nil>; want one naming the staged path
--- FAIL: TestAC_UBR_17_RenameLeavesTheContentStagedWhenThePutBackIsRefused
    residue_noreplace_test.go:416: error = <nil>; want one naming the staged path
--- FAIL: TestAC_UBR_17_RenameLeavesTheContentStagedWhenBothLaterRenamesFail
    residue_noreplace_test.go:441: the rename made 0 calls of the no-replace seam, want 2
--- FAIL: TestAC_UBR_17_RenameMovesWhateverHoldsTheStagedNameAndRemovesNothing
    residue_noreplace_test.go:473: no-replace seam calls = 0, want 2
--- FAIL: TestAC_UBR_17_RenameLeavesAFileThatReplacedTheStageFolderBeforeItsRemoval
    residue_noreplace_test.go:511: rename: skill.md and skill.md.tmp-rename/skill.md no longer both identify the file being renamed; both are left in place with the staging folder, and nothing was renamed
--- FAIL: TestAC_UBR_17_RenameMovesTheUsersNewFileWhenSkillMdWasSavedBeforeTheFirstRename
    residue_noreplace_test.go:541: no-replace seam calls = 0, want 1 (one probe, then the fallback)
--- FAIL: TestAC_UBR_17_RenameFallsBackToLinksWhenNoReplaceRenameIsUnsupported
    residue_noreplace_test.go:563: staged file not found: [] <nil>
--- FAIL: TestAC_UBR_17_FallbackKeepsAForeignStagedFileUnderARandomStageName
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.132s
```

Reading the RED. Real data loss on the unchanged code: the codex window-1 repro (`...BeforeTheFirstMove`: the foreign file at the stage name was deleted by `dropOurs`) and the no-unlink test (three `dropFile` calls on user-visible names). The rest fail because the old code does not use the primitive or the random name (the seam was never called, the name is plain); they fail for that reason, not for data loss. Pins that already passed on the unchanged code (kept as regression guards): `TestAC_UBR_17_RenameKeepsTheContentsOfAFolderThatReplacedTheStageFolder`, `TestAC_UBR_16_RenameDoesNotFollowASkillFolderSwappedAfterTheStageWasClaimed`, and the 15 round-3 link/identity tests re-pointed at the fallback.

Mutants (`go test -count=1 -overlay=<json> ./internal/cli/update/skillresidue/`; overlay JSON and mutated copies of `residue.go` under `/private/tmp/t1733r4/ov/`):

| Mutant | Edit | Result |
|---|---|---|
| a | `renameNoReplace = noReplaceRename` replaced by `renameNoReplace = os.Rename` | exit 1, 4 FAIL: `...KeepsAFileThatAppearsAtSkillMdBetweenTheTwoRenames` (`error = <nil>; want one naming SKILL.md`), `...AtSkillMdLowerBetweenTheTwoRenames` (`the user's file is gone`), `...LeavesTheContentStagedWhenThePutBackIsRefused` (`skill.md = "old body", want the user's file`), `TestAC_UBR_17_RenameDoesNotOverwriteAFileThatAppearsInsideTheStagingFolder` (`the user's file inside the staging folder was overwritten`) |
| b | the put-back `moveNoReplace(root, staged, lower)` block removed after a refused second rename (returns the error only) | exit 1, 1 FAIL: `TestAC_UBR_17_RenamePutsTheFileBackWhenTheSecondRenameIsRefused` (`no-replace seam calls = 2, want 3 (move, refused move, put back)`) |

### Existing tests edited (assertions untouched; one line of reason each)

| Test / file | Edit | Reason |
|---|---|---|
| `idSetup` in `residue_identity_test.go` (covers the 10 round-3 `TestAC_UBR_17_Rename*` identity tests) | added `forceFallback(t)` | they drive the link and drop seams and address `skill.md.tmp-rename/...` by literal path; the default path is now no-replace renames, so they run through the fallback they exercise |
| `TestAC_UBR_17_RenameDoesNotDropASkillFileThatWasAtomicallyReplaced` (`residue_atomic_save_test.go`) | added `forceFallback(t)` | injects through the link seam |
| `TestAC_UBR_17_RenameDoesNotOverwriteASkillFileThatAppearsAtTheDestination` (`residue_rename_race_test.go`) | added `forceFallback(t)` | injects through the link seam; the no-replace analogue is `...BetweenTheTwoRenames` |
| `TestClean_RenameFailureIsReturnedAndNothingMoves`, `TestRenameSkillFile_RestoreFailureIsReported` (`residue_coverage_test.go`) | added `forceFallback(t)` | link-seam injection; the second also reads the literal stage path |
| `TestRenameSkillFile_NoFreeTemporaryNameIsRefused` (`residue_coverage_test.go`) | added `legacyStageNames(t)` | it fills the deterministic names; random names cannot be exhausted |
| `TestRenameSkillFile_RestoresLowercaseNameWhenTheLinkIsRefused` (`residue_safety_test.go`) | added `forceFallback(t)` | link-seam injection |
| `TestMoveFileReferencesArePinned` (`residue_moveone_test.go`) | pinned set gains `renameNoReplace`, `noReplaceRename`, `Rmdir`; counts: `linkFile@renameSkillFile` 3 -> `linkFile@renameByLinks` 3, `dropFile@renameSkillFile` 1 -> `dropFile@renameByLinks` 1, new `renameNoReplace@moveNoReplace` 1, `Rmdir@removeStage` 1, package-level bindings | the link steps moved into `renameByLinks`; the test is stricter (it now also fails on any new reference to the no-replace rename or the rmdir) |
| registry (`internal/cli/update_destructive_registry.go`) | `dropOne` row text rewritten (Sites still 1) | the row described the old flow; now says the row is reached only by the fallback and names the no-replace renames and the rmdir |

### Honest residual

- Fallback only: on a platform or filesystem that reports `errNoReplaceUnsupported` the round-3 path keeps its microsecond window between each identity look and the unlink it guards (E.7 residual unchanged). Filesystems without RENAME_NOREPLACE/RENAME_EXCL (some network and FUSE filesystems; kernels before Linux 3.15) take that path. The first-rename probe also treats `EINVAL` as unsupported, so an unrelated `EINVAL` falls back to the safe link path.
- Between B and C the content has exactly one name, the random private T. A foreign atomic overwrite of T in that interval replaces the content; the step then moves the foreign file and removes nothing (test above). Reaching it requires guessing a 48-bit name or enumerating the folder inside a microsecond gap. The old link design kept two names; this is the price of having no unlink.
- The renames are path-based, so the root handle no longer bounds them. A root-bound `Lstat` precedes each (F24 test stays green), leaving a microsecond window in which a skill folder swapped for a link out of the project would redirect the rename; the outcome is a rename, not a deletion.
- A symlink or folder swapped in at `skill.md` between the caller's regular-file check and B is moved by name to SKILL.md (no content is lost or replaced).
- If the second rename itself reports "unsupported" after the first worked (not observed), the put back is unsupported too and the content stays staged with the error naming it.
- Windows runtime unobserved: `MoveFileEx` without a replace flag is cross-built and vetted only (`GOOS=windows`); its refusal of an existing destination comes from the Win32 documentation, not from a run. Linux runtime unobserved (cross-built and vetted only); the real-filesystem primitive tests ran on case-insensitive APFS (darwin/arm64).

### Verification

Env: `GIT_CONFIG_GLOBAL=/private/tmp/t1733r4/gitconfig` (only `[user]`), `TMPDIR=/private/tmp/t1733r4/tmp`, `GOTMPDIR=/private/tmp/t1733r4/gotmp`, no HOME override, no `-mod=mod`. Run at GREEN `eb0a6a004`:

```
go build ./...                                   -> exit 0
gofmt -l internal/cli internal/cli/update internal/template -> (empty), exit 0
go vet ./internal/cli/update/...                 -> exit 0
GOOS=linux go vet ./internal/cli/update/skillresidue/   -> exit 0
GOOS=windows go vet ./internal/cli/update/skillresidue/ -> exit 0
go test -count=1 ./internal/cli/update/...      -> ok update 0.945s, backup 1.186s, deploy 0.706s, merge 0.767s, plan 0.270s, report 0.280s, skillresidue 0.249s
go test -count=1 ./internal/cli/ -run 'Backup|Retention|Destructive|Housekeeping|UBR|Unmanaged|TemplateSync|CleanReinstall' -> ok internal/cli 28.966s
go test -count=1 ./internal/template/...         -> ok template 175.270s, agentemit 0.514s, commandemit 0.261s, embedemit 0.282s, scripts [no test files]
GOOS=linux GOARCH=amd64 go build ./...           -> exit 0
GOOS=windows GOARCH=amd64 go build ./...         -> exit 0
go test -count=1 -cover ./internal/cli/update/skillresidue/ -> ok 0.193s coverage: 89.4% of statements
```

Gaps: repository-wide test verdict PENDING (CI on the integration branch); `make build` not run; the `verify` skill delegates to the full MoAI gate, which was not run before the GREEN commit (a whole-repository run is excluded locally by the load rule; the scoped commands above were run instead and the pre-commit hook owns the blocking decision); no guard refused any command in this round; `-race` not run; Linux and Windows runtime behavior of the new primitives not observed; re-audit is the auditor's.

## §E.9 Repair-round-5 evidence (card t1733, sync-audit iteration 5: codex P1 residue.go:408, codex P2 residue.go:389, F15-F19; owner manager-develop)

Source: `.moai/reports/t1733/sync-audit-iter5.md` (local-only). Start HEAD `3f9d25c70`. Round 4 replaced link+identity+unlink by path-based no-replace renames; iteration 5 found two regressions of that move: a stage folder swapped for a link to an outside folder made the first rename carry the user's file out of the project (at `f542d166d` `os.Root` refused the same swap), and a staged file replaced between the renames was moved and reported as success. This round makes the renames descriptor-bound where the platform allows it and adds the identity check.

### Design

- `holdDir(root, rel)` opens a folder through the project root (`root.OpenRoot(rel)`, then `Open(".")`) and keeps both handles open for the whole rename. The root refuses a path that leaves the project; both handles name the folder's inode.
- `renameSkillFile` holds the skill folder before `claimStage`, and the stage folder right after `Mkdir`. Seam `renameNoReplace(srcDir, src, dstDir, dst *os.File ...)` replaced the absolute-path form; linux `unix.Renameat2(sfd, src, dfd, dst, RENAME_NOREPLACE)`, darwin `unix.RenameatxNp(sfd, src, dfd, dst, RENAME_EXCL)`. `RenameatxNp` exists in `golang.org/x/sys v0.49.0` (`unix/syscall_darwin.go:414`; confirmed by compiling and by the darwin test run), so darwin is fully descriptor-bound and the fallback "root-bound Lstat plus path-based RenamexNp" the brief offered was not needed. `go.mod`/`go.sum` untouched.
- The stage folder is removed by `unlinkat(AT_REMOVEDIR)` relative to the held skill folder (`removeDirAt`; `syscall.Rmdir` on windows and on platforms with no rename primitive).
- Identity (F16): the step takes `moving = skill.root.Lstat("skill.md")` from the skill folder's descriptor right before the first rename, requires `stageDir.root.Lstat("skill.md")` to be that file before the second rename (mismatch: second rename NOT run, nothing put back, foreign file and stage left, error names T), and requires `skill.root.Lstat("SKILL.md")` to be that file after it (mismatch: error names SKILL.md, the foreign file is left there). Deviation from the brief, on purpose: the brief said to record the identity right after the first rename. A record taken after the rename cannot see a replacement that lands between the rename and the record (the in-tree test replaces the staged name inside the seam, after the real rename), so the identity is taken before the rename; the cost is that a user edit landing between that record and the rename is refused instead of renamed (row below), which is the conservative side.
- The root-bound `Lstat` of the source stays in `moveNoReplace` (F17). With descriptors it is no longer the containment on linux and darwin; it makes the step refuse when the path no longer leads to the folder it holds, and on windows it is the containment.
- F18: the comment now says what was measured: on APFS a direct no-replace rename `skill.md -> SKILL.md` succeeded (iteration 5 row 6), so the staging hop is not shown necessary there; it stays as the designed route and other case-folding filesystems were not measured. A direct hop would also remove the single-private-name interval of the B-C row; that is a design choice for the leader, not taken here.
- F19: `TestMoveFileReferencesArePinned` now pins the platform calls by owner (`Renameat2`, `RenameatxNp`, `MoveFileEx`, `Unlinkat`, `Rmdir`, `removeDirAt`, `renameatNoReplace`), so a new reference to them inside the package fails a test. A caller in another package is still invisible to it and to the registry scan (the scan counts `os.Remove*`/`os.Rename` and the `os.Root` forms only). No new registry mechanism was added; the `dropOne` row text was updated.

### State table (mirrored from the code comment above `renameSkillFile`)

L = skill.md, S = staging folder, T = S/skill.md. Rows marked * changed or are new in this round.

| Step | State | Action | User data | Test |
|---|---|---|---|---|
| A | name taken | try another name; 100 taken: error | nothing changed | unchanged |
| A | Mkdir fails | error | nothing changed | unchanged |
| A/B | S swapped for a regular file | opening S fails (ENOTDIR); rmdir S fails; L stays; error names S | foreign file kept | `...LeavesAFileThatReplacedTheStageFolderBeforeTheFirstMove` (unchanged assertions) |
| A/B* | S swapped for a link to a folder outside the project | the root refuses the open, nothing moves, the link stays (rmdir fails on a link), error names S | outside untouched; L stays | `TestAC_UBR_16_RenameRefusesAStageFolderSwappedForALinkOutsideTheProject` |
| A/B | S swapped for a foreign folder | B lands T in it; C completes; D fails if the folder holds anything else, error names it | content not lost; foreign files kept | `...KeepsTheContentsOfAFolderThatReplacedTheStageFolder` (unchanged) |
| A/B | skill folder swapped for a link out of the project before it was opened | the root refuses the open; nothing moves | outside untouched | `TestAC_UBR_16_SwappedSkillFolderCannotRedirectTheRename` (unchanged) |
| B* | S (or the skill folder) swapped after it was opened | the rename uses the open folder and lands in the original inode (now moved aside, inside the project); the look before C refuses (the path now leads out); error | outside untouched; the file stays in the project **[CORRECTED in §E.10: false when the held folder is renamed out of the project; the file went with it and the put-back was refused]** | `TestAC_UBR_16_RenameIsBoundToTheOpenedStageFolderNotToItsPath` |
| B | a foreign file sits at T | B refused; L untouched; D fails (S not empty); error names T | foreign file and L kept | unchanged |
| B | L replaced before the identity is taken | B moves whatever is at L | the user's newest file becomes SKILL.md | `...MovesTheUsersNewFileWhenSkillMdWasSavedBeforeTheFirstRename` (unchanged) |
| B* | L replaced between the identity and B | the file at T is not the identity: C not run, error names T, nothing put back | the moved file stays at T | covered by the same check as the next row; not separately injectable (no seam between the record and the rename) |
| B | B fails for another reason | L untouched; rmdir S; error | none lost | unchanged |
| B | B reports `errNoReplaceUnsupported` | nothing moved; `renameByLinks` takes over | see fallback table | unchanged |
| B-C* | T replaced by a foreign file | C NOT run, nothing put back, S kept; error names T | the foreign file stays at T; the content of L was overwritten by the foreign write, not by this step | `TestAC_UBR_17_RenameRefusesAStagedFileThatWasReplacedBetweenTheRenames` (replaces the round-4 test that required success) |
| B-C | a file appears at SKILL.md (case-folding FS: at skill.md) | C refused; D puts T back to L; error names SKILL.md; if D is refused too the error names T | user's file kept; content back at L, or staged at T | unchanged |
| C* | T replaced inside the step itself | the foreign file moves to SKILL.md; the file at SKILL.md is not the identity: error names SKILL.md | the foreign file is kept at SKILL.md; the content of L was overwritten by the foreign write | `TestAC_UBR_17_RenameReportsAFileThatReplacedTheStagedNameDuringTheSecondRename` |
| C | C fails for another reason | D puts T back; error | none lost | unchanged |
| C-D | S swapped for a file / non-empty folder after C | D fails; error names S | foreign entries kept | unchanged |
| all | no user-visible name is unlinked | `dropFile` never called on this path | no check-to-unlink window | unchanged |

### RED / GREEN / mutant

| Step | Commit | Command and result |
|---|---|---|
| RED | `6f142dc1f0979aaccf9a9950eda5f856918239dd` | `go test -count=1 -v -run 'AC_UBR_16_Rename(Refuses\|IsBound)\|AC_UBR_17_Rename(Refuses\|Reports)' ./internal/cli/update/skillresidue/`: 4 FAIL, 2 PASS; whole package: 4 FAIL |
| GREEN | `cd3c69aa9ac83e470f7778a892f8af55803c12ea` | same command: 6 PASS; `go test -count=1 -race ./internal/cli/update/skillresidue/` -> `ok ... 1.318s`; coverage 89.4% |

No stub file was needed this time: the new tests use only the existing `renameSkillFile`, `mkdirFile`, `afterNthRename` and `renameNoReplace` names, so RED compiles on the unchanged code. Verbatim RED (temporary path prefix removed):

```
=== RUN   TestAC_UBR_16_RenameRefusesAStageFolderSwappedForALinkOutsideTheProject
    residue_handle_test.go:65: the folder outside the project received 1 entr(ies): [. dir skill.md file]
--- FAIL: TestAC_UBR_16_RenameRefusesAStageFolderSwappedForALinkOutsideTheProject (0.00s)
=== RUN   TestAC_UBR_16_RenameIsBoundToTheOpenedStageFolderNotToItsPath
    residue_handle_test.go:96: the folder outside the project received 1 entr(ies): [. dir skill.md file]
--- FAIL: TestAC_UBR_16_RenameIsBoundToTheOpenedStageFolderNotToItsPath (0.00s)
=== RUN   TestAC_UBR_16_RenameRefusesWhenTheSkillFolderIsSwappedBetweenTheTwoRenames
--- PASS: TestAC_UBR_16_RenameRefusesWhenTheSkillFolderIsSwappedBetweenTheTwoRenames (0.00s)
=== RUN   TestAC_UBR_17_RenameRefusesAStagedFileThatWasReplacedBetweenTheRenames
    residue_handle_test.go:148: error = <nil>; want one naming the staged path
--- FAIL: TestAC_UBR_17_RenameRefusesAStagedFileThatWasReplacedBetweenTheRenames (0.00s)
=== RUN   TestAC_UBR_17_RenameReportsAFileThatReplacedTheStagedNameDuringTheSecondRename
    residue_handle_test.go:181: error = <nil>; want one naming SKILL.md
--- FAIL: TestAC_UBR_17_RenameReportsAFileThatReplacedTheStagedNameDuringTheSecondRename (0.00s)
=== RUN   TestAC_UBR_17_RenameReportsWhenTheUndoOfTheStagedLinkFails
--- PASS: TestAC_UBR_17_RenameReportsWhenTheUndoOfTheStagedLinkFails (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.088s
```

Reading the RED. The two P1 tests fail for the right reason: the user's `skill.md` was carried into the outside folder (the first through a stage folder swapped for a link right after `Mkdir`, the second through a stage path swapped after the step had chosen it). The two P2 tests fail because the step returned success after a foreign file was moved. The skill-folder-swap test (F17) already passed on the unchanged code: it is a pin of the root-bound look, killed by mutant 3 below. `...ReportsWhenTheUndoOfTheStagedLinkFails` is a round-3 test matched by the `-run` pattern and passes before and after.

Mutants (`go test -count=1 -overlay=<json> ./internal/cli/update/skillresidue/` against GREEN; overlay JSON and mutated copies under `/private/tmp/t1733r5/ov/<name>/`):

| Mutant | Edit | Result |
|---|---|---|
| 1 path-based rename | `noReplaceRenameAt` in `rename_noreplace_unix.go` calls `unix.RenamexNp(filepath.Join(srcDir.Name(), src), filepath.Join(dstDir.Name(), dst), RENAME_EXCL)` instead of the descriptor form | exit 1, 1 FAIL: `TestAC_UBR_16_RenameIsBoundToTheOpenedStageFolderNotToItsPath` (`the folder outside the project received 1 entr(ies): [. dir skill.md file]`) |
| 2a identity before C removed | the `stageDir.root.Lstat` check before the second rename deleted | exit 1, 1 FAIL: `TestAC_UBR_17_RenameRefusesAStagedFileThatWasReplacedBetweenTheRenames` (`error = SKILL.md does not hold the file that was staged (<nil>) ...; want one naming the staged path`; the post-check still reports, the test wants the refusal before the move) |
| 2b identity after C removed | the `skill.root.Lstat("SKILL.md")` check deleted | exit 1, 1 FAIL: `TestAC_UBR_17_RenameReportsAFileThatReplacedTheStagedNameDuringTheSecondRename` (`error = <nil>; want one naming SKILL.md`) |
| 3 root-bound look removed | the `root.Lstat(from.rel)` in `moveNoReplace` deleted | exit 1, 1 FAIL: `TestAC_UBR_16_RenameRefusesWhenTheSkillFolderIsSwappedBetweenTheTwoRenames` (`the rename reported success although the skill folder path was swapped for an outside link`) |

A fourth mutant removing both identity checks did not compile (an unused variable) and was not run; 2a and 2b each fail on their own.

### Existing tests edited (assertions untouched except the one flipped row; one line of reason each)

| Test / file | Edit | Reason |
|---|---|---|
| `requireNoReplace`, the two `TestNoReplaceRename_*` primitive tests, `forceFallback`, `afterNthRename`, `TestAC_UBR_17_RenameFallsBackToLinksWhenNoReplaceRenameIsUnsupported` (`residue_noreplace_test.go`) | re-pointed from the absolute-path seam/primitive to the folder-relative one (new helper `renameInDir`; seam closures take `(*os.File, string, *os.File, string)`) | the primitive and seam changed signature; the hooks `act(n)`/`fail(n)` and every assertion are unchanged |
| `TestAC_UBR_17_RenameMovesWhateverHoldsTheStagedNameAndRemovesNothing` | **replaced** by `...RefusesAStagedFileThatWasReplacedBetweenTheRenames` with the opposite assertion (error, foreign file left at T, SKILL.md not published) | this is the audit's required flip (F16): the old test demanded the success the auditor reproduced as a defect; the state-table row changed with it |
| `TestMoveFileReferencesArePinned` | pinned set and counts extended to the platform calls and `removeDirAt` | F19; stricter, no count weakened |
| registry `dropOne` row (`internal/cli/update_destructive_registry.go`) | text rewritten, `Sites: 1` unchanged | the row described the path-based design |

### F15: SPEC wording item for the leader (spec.md not edited)

`spec.md` REQ-UBR-017 still reads "shall perform the rename through a handle bound to the project root". After this round: on linux and darwin the folders the rename is relative to are opened through the project root (`os.Root`) and held, so the statement is true in substance (the handles are root-bound, a path leaving the project is refused when opened). On windows and on platforms that take the link fallback it holds as follows: windows renames by path behind a root-bound look (not a handle), the fallback renames through `os.Root` links. The text is therefore not wrong for the primary platforms and is stricter than the code on windows. Item for the leader: either accept the deviation knowingly or ask manager-spec to qualify REQ-UBR-017 for windows. AC-UBR-16 and AC-UBR-17 as written hold.

### Honest residual (per platform)

- linux, darwin: no path swap can redirect a rename (descriptor-relative) **[CORRECTED in §E.10: a held folder renamed away still receives the move, because the descriptor names the inode; the claim held only for a swap that leaves the folder where it is]**. A microsecond window remains between the identity check before C and the call of C itself: a foreign atomic overwrite landing exactly there moves the foreign file to SKILL.md; the step cannot prevent it and reports it (error names SKILL.md) but does not move it back. Between B and C the content has one private name; the overwrite of that name destroys the content (the foreign writer's overwrite, not this step's), and it is now reported at the next check instead of succeeding. Reaching either needs the 48-bit stage name.
- linux: runtime unobserved. `Renameat2`, `Unlinkat` and the `SyscallConn` use compile and vet for linux/amd64; they ran only on darwin. A seccomp profile answering `renameat2` with EPERM still takes no fallback (unchanged from iteration 5).
- windows: the rename is still path-based (`MoveFileEx` with the held folders' `Name()`); `golang.org/x/sys/windows` has no folder-relative no-replace rename. The root-bound look before each rename is the containment, and a microsecond window remains between the look and the call: a stage folder replaced by a junction or symlink inside that window carries the file out of the project (a move, not a deletion; the next look then refuses, the put-back is refused, the error names the staged path). The windows build is cross-compiled and vetted only. I kept the path-based call over the round-3 `os.Root` link route for windows because the link route reintroduces the unlink-after-check data-loss windows (F13/F14) that the primary design removed.
- platforms with no primitive (`rename_noreplace_other.go`) and filesystems answering EINVAL/ENOSYS/ENOTSUP: the round-3 link fallback with its look-to-unlink windows (E.7), unchanged.
- The error text of a refused put-back after a swapped stage path names `<stage>/skill.md`, which in the "stage swapped after it was opened" row is the swapped path; the file is in the original folder, moved aside next to it (the test finds it inside the project). Not changed.
- A user edit landing between the identity record and the first rename is refused (the moved file stays at T, error names it) instead of renamed; one microsecond window, retried by the next update run only if skill.md exists again (it does not, until the user restores the file from T).

### Verification

Env: `GIT_CONFIG_GLOBAL=/private/tmp/t1733r5/gitconfig` (only `[user]`), `TMPDIR=/private/tmp/t1733r5/tmp`, `GOTMPDIR=/private/tmp/t1733r5/gotmp`, no HOME override, no `-mod=mod`. Run at GREEN `cd3c69aa9`:

```
go build ./...                                          -> build=0
gofmt -l internal/cli internal/cli/update internal/template -> (empty), gofmt=0
go vet ./internal/cli/update/...                        -> vet=0
GOOS=linux go vet ./internal/cli/update/skillresidue/   -> vet-linux=0
GOOS=windows go vet ./internal/cli/update/skillresidue/ -> vet-windows=0
GOOS=freebsd go build ./internal/cli/update/skillresidue/ -> build-freebsd=0
GOOS=linux GOARCH=amd64 go build ./...                  -> linuxbuild=0
GOOS=windows GOARCH=amd64 go build ./...                -> winbuild=0
go test -count=1 -race ./internal/cli/update/skillresidue/ -> ok ... 1.318s
go test -count=1 -cover ./internal/cli/update/...       -> ok update 0.622s 85.2%, backup 0.724s 88.4%, deploy 0.463s 85.9%, merge 0.828s 93.1%, plan 0.589s 95.2%, report 0.595s 95.1%, skillresidue 0.370s 89.4%
go test -count=1 ./internal/cli/ -run 'Backup|Retention|Destructive|Housekeeping|UBR|Unmanaged|TemplateSync|CleanReinstall' -> ok internal/cli 23.260s
go test -count=1 ./internal/template/...                -> ok template 150.029s, agentemit 0.362s, commandemit 0.254s, embedemit 0.268s, scripts [no test files]
```

Extra cross-builds of the package (amd64 unless noted): darwin, linux, windows, freebsd, openbsd, js/wasm -> all exit 0.

Gaps: repository-wide test verdict PENDING (CI on the integration branch owns it); `make build` not run; Linux and Windows runtime behavior of the descriptor-relative renames, `Unlinkat` and `MoveFileEx` not observed (cross-build and vet only); the real-filesystem tests ran on case-insensitive APFS (darwin/arm64) only; a hostile-writer race stress with real syscalls (iteration 5 row 12) was not re-run this round; the row "L replaced between the identity and B" has no injection point and is covered only by reading the check; no tool call was refused this round; the `verify` skill delegation of the pre-commit gate was not separately run (scoped commands above were); re-audit is the auditor's.

## §E.10 Repair-round-6 evidence (card t1733, sync-audit iteration 6: codex P1 residue.go:505, codex P2 residue.go:502, F20, F21, F22, F23; owner manager-develop)

Source: `.moai/reports/t1733/sync-audit-iter6.md` (local-only). Start HEAD `9c74d7d0e`. Round 5 bound the renames to open folder handles, which closed the stage-swap and staged-file-replacement findings and, as iteration 6 verified, opened two new states: a held folder renamed away keeps receiving the move because the descriptor names the inode, and a held skill folder renamed away can leave the requested path without SKILL.md while the step reports success. Nothing is destroyed in either state: the bytes sit in a folder the same-user actor chose to move. This round makes the put-back independent of paths on unix and makes the final check name the requested path.

### Correction of an earlier claim (E.9 and the code comment)

The state-table row `B*` of §E.9 and of the comment above `renameSkillFile` said, for "S (or the skill folder) swapped after it was opened", "outside untouched; the file stays in the project". That is **false when the held folder is renamed out of the project**: the first rename lands in the moved folder and the file leaves the project with it. The same sentence in the first bullet of §E.9 "Honest residual" ("no path swap can redirect a rename") held only for a swap that leaves the folder where it is. Both E.9 places now carry an inline `[CORRECTED in §E.10 ...]` marker; the code comment table was rewritten (rows below).

### Design

- `moveNoReplace(root, from, to, look bool)`: the root-bound look at the source path precedes the rename only when `look` is set. The two forward moves (B and C) pass `true`, so the F17 pin (the look at the second rename's source) is unchanged. The put-back passes `!renameIsDescriptorBound`.
- `renameIsDescriptorBound` is a per-platform constant: `true` on linux and darwin (`rename_noreplace_unix.go`), `false` on windows and on the platforms without a rename primitive. On unix the put-back is therefore `renameat2` / `renameatx_np` from the held stage folder into the held skill folder, with no path look; on windows it keeps the look.
- F21: after the second rename, besides the held-folder identity check, `root.Lstat(canonical)` must name the same file as `moving`. On mismatch the step returns an error that says the skill folder this step opened as `<dirRel>` was moved (checked with `leadsTo`: `os.SameFile` of the path's `Lstat` and the held folder's `Lstat(".")`), that SKILL.md was written inside that moved folder, and that nothing was removed or moved back. The staging folder is left in place on purpose (nothing is removed in that branch). The moved folder's new path is not determinable without a path-of-descriptor lookup (`F_GETPATH` on darwin, `/proc/self/fd` on linux), so the message does not claim one.
- F22: `contentPlace` / `where` name the staged path only when the path still leads to the held stage folder (`leadsTo`); otherwise they say "in skill.md inside the folder this step opened as `<stage>`, which was moved". The identity-mismatch error before C uses the same helper.
- The look error of `moveNoReplace` is now wrapped (`move <from> to <to>: statat ...`) like the rename error; `errors.Is` on `errNoReplaceUnsupported` is unaffected.
- `go.mod` / `go.sum` untouched; `spec.md` untouched (the F15 wording item for the leader stays in §E.9).

### State table (rows new or changed this round; mirrored from the code comment above `renameSkillFile`)

L = skill.md, S = staging folder, T = S/skill.md. The remaining rows are as in §E.9. The windows column is read from the code (path-based `MoveFileEx` behind the look) and was not run.

| Step | State | Before this round (code at `9c74d7d0e`) | After this round (unix) | Windows (path-based) | Test |
|---|---|---|---|---|---|
| B* | S renamed away after it was opened (to a link, a folder, out of the project, or nothing at its old path) | B landed in the moved S; the look before C refused; the put-back went through the same look and was refused; **the file stayed in the moved folder** (outside the project when S was moved out), the error named the old stage path where nothing is | B lands in the held S, the look before C refuses, the put-back goes through the held descriptors with no look: **the file is back at L**; an empty or foreign folder may sit where S was; the error is the look error (plus the rmdir refusal when S is gone) and does not claim a content path | the put-back is refused by the look; the file stays in the moved folder; the error says so ("which was moved") | `...PutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject` (two variants), `...MovedOutBetweenTheRenames` |
| B* | the skill folder renamed away before B | B landed in S inside the moved folder; the put-back was refused; the file stayed at `<moved>/S/skill.md` | the put-back goes into the held (moved) skill folder: the file is back as skill.md inside the moved folder; error | as above | `...PutsTheFileBackIntoASkillFolderThatWasMovedAside` |
| C* | the skill folder renamed away after B, a decoy folder put at its old path | C ran in the held moved folder, the look passed on the decoy, the step **returned nil** with no SKILL.md at the requested path | error: the requested path does not name the renamed file, the skill folder was moved, SKILL.md is inside it; nothing removed, nothing moved back by path, decoy kept | same (the check is a root-bound look, on every platform) | `...DoesNotReportSuccessWhenTheSkillFolderWasMovedAndADecoyTookItsPath` |
| B-C* | the stage folder moved away and a user file took L, so the put-back is refused | error said "the content is at `<stage>/skill.md`", a path with nothing at it | error says the content is in skill.md inside the folder this step opened as `<stage>`, which was moved; nothing overwritten | same | `...NamesTheMovedFolderWhenThePutBackIsRefused` |

### RED / GREEN / mutant

| Step | Commit | Command and result |
|---|---|---|
| RED | `46aacdcde540719adea67ef04402d1abff497cb0` | `go test -count=1 -v -run 'MovedOutOfTheProject\|MovedOutBetween\|MovedAside\|MovedAndADecoy\|NamesTheMovedFolder' ./internal/cli/update/skillresidue/`: 5 test functions (6 with subtests) FAIL, 0 pass; no pin among the new tests |
| GREEN | `6888aff16283c782a6e2680fbeac4fafa1076f50` | same command: 5 functions PASS (6 with subtests); `go test -count=1 -race ./internal/cli/update/skillresidue/` -> `ok ... 1.232s`; coverage 89.7% |

RED compiled on the unchanged code (the tests use only `renameSkillFile`, `afterNthRename`, `handleSetup`, `stageFolder`, `readOrFail`, `fileWithBodyUnder`, `treeSnapshot`). RED output (run at `46aacdcde`; the temporary-directory prefix is removed or shown as `<tmp>`, the subtest lines are shortened with `...`, nothing else is edited):

```
=== RUN   TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject/nothing_at_the_old_path
    residue_moved_folder_test.go:92: skill.md = "" (present false), want the user's bytes back at the requested path (project [. dir moai-ref-owasp-checklist dir], outside [. dir moved-stage dir moved-stage/skill.md file])
=== RUN   TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject/a_new_folder_at_the_old_path
    residue_moved_folder_test.go:92: skill.md = "" (present false), want the user's bytes back at the requested path (project [. dir moai-ref-owasp-checklist dir moai-ref-owasp-checklist/skill.md.tmp-rename-27d461b2c94c dir], outside [. dir moved-stage dir moved-stage/skill.md file])
--- FAIL: TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject (0.00s)
    --- FAIL: ...MovedOutOfTheProject/nothing_at_the_old_path (0.00s)
    --- FAIL: ...MovedOutOfTheProject/a_new_folder_at_the_old_path (0.00s)
=== RUN   TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutBetweenTheRenames
    residue_moved_folder_test.go:110: skill.md = "" (present false), want the user's bytes back at the requested path (project [. dir moai-ref-owasp-checklist dir], outside [. dir moved-stage dir moved-stage/skill.md file])
--- FAIL: TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutBetweenTheRenames (0.00s)
=== RUN   TestAC_UBR_16_RenamePutsTheFileBackIntoASkillFolderThatWasMovedAside
    residue_moved_folder_test.go:130: moved skill folder's skill.md = "" (present false), want the user's bytes put back (tree [. dir moai-ref-owasp-checklist.moved dir moai-ref-owasp-checklist.moved/skill.md.tmp-rename-a30597eb6395 dir moai-ref-owasp-checklist.moved/skill.md.tmp-rename-a30597eb6395/skill.md file])
--- FAIL: TestAC_UBR_16_RenamePutsTheFileBackIntoASkillFolderThatWasMovedAside (0.00s)
=== RUN   TestAC_UBR_17_RenameDoesNotReportSuccessWhenTheSkillFolderWasMovedAndADecoyTookItsPath
    residue_moved_folder_test.go:161: the rename reported success although <tmp>/002/.claude/skills/moai-ref-owasp-checklist has no SKILL.md (tree [. dir moai-ref-owasp-checklist dir moai-ref-owasp-checklist.moved dir moai-ref-owasp-checklist.moved/SKILL.md file moai-ref-owasp-checklist/skill.md.tmp-rename-5faa7a9e6497 dir moai-ref-owasp-checklist/skill.md.tmp-rename-5faa7a9e6497/skill.md file])
--- FAIL: TestAC_UBR_17_RenameDoesNotReportSuccessWhenTheSkillFolderWasMovedAndADecoyTookItsPath (0.00s)
=== RUN   TestAC_UBR_17_RenameNamesTheMovedFolderWhenThePutBackIsRefused
    residue_moved_folder_test.go:203: error = statat .claude/skills/moai-ref-owasp-checklist/skill.md.tmp-rename-ec33e23c7d38/skill.md: no such file or directory (restoring .claude/skills/moai-ref-owasp-checklist/skill.md also failed: statat .claude/skills/moai-ref-owasp-checklist/skill.md.tmp-rename-ec33e23c7d38/skill.md: no such file or directory; the content is at .claude/skills/moai-ref-owasp-checklist/skill.md.tmp-rename-ec33e23c7d38/skill.md); it points at a staging path that holds nothing
--- FAIL: TestAC_UBR_17_RenameNamesTheMovedFolderWhenThePutBackIsRefused (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.250s
```

Reading the RED. The first three fail for the right reason: the user's `skill.md` is not at the requested path (it sits in the folder moved out of the project, or inside the moved skill folder under the staging name). The fourth fails because the step returned success with no SKILL.md at the requested path. The fifth fails because the error points at a staging path that holds nothing.  The decoy test and the refused-put-back test also assert that the decoy, the staging folder and the user's new file stay as they were.

Verbatim GREEN (same command at `6888aff16`):

```
--- PASS: TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject (0.00s)
    --- PASS: TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject/nothing_at_the_old_path (0.00s)
    --- PASS: TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutOfTheProject/a_new_folder_at_the_old_path (0.00s)
--- PASS: TestAC_UBR_16_RenamePutsTheFileBackWhenTheStageFolderWasMovedOutBetweenTheRenames (0.00s)
--- PASS: TestAC_UBR_16_RenamePutsTheFileBackIntoASkillFolderThatWasMovedAside (0.00s)
--- PASS: TestAC_UBR_17_RenameDoesNotReportSuccessWhenTheSkillFolderWasMovedAndADecoyTookItsPath (0.00s)
--- PASS: TestAC_UBR_17_RenameNamesTheMovedFolderWhenThePutBackIsRefused (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli/update/skillresidue	0.253s
```

Error texts observed at GREEN through a scratch overlay test (not added to the tree; a joined error prints its parts on separate lines, the first line is shown):

```
stage moved out, nothing at its old path:
  move .claude/skills/moai-ref-owasp-checklist/skill.md.tmp-rename-45ecb5d6442f/skill.md to .claude/skills/moai-ref-owasp-checklist/SKILL.md: statat .claude/skills/moai-ref-owasp-checklist/skill.md.tmp-rename-45ecb5d6442f/skill.md: no such file or directory
  tree after: project=[. dir .claude dir .claude/skills dir .claude/skills/moai-ref-owasp-checklist dir .claude/skills/moai-ref-owasp-checklist/skill.md file] outside=[. dir moved-stage dir]
skill folder moved, decoy at the old path:
  .claude/skills/moai-ref-owasp-checklist/SKILL.md does not name the file that was renamed (statat .claude/skills/moai-ref-owasp-checklist/SKILL.md: no such file or directory): the skill folder this step opened as .claude/skills/moai-ref-owasp-checklist was moved, so SKILL.md was written inside that moved folder; nothing was removed or moved back, and the staging folder skill.md.tmp-rename-7562b56e85a8 inside it was left in place
stage moved out and a user file at skill.md (put-back refused):
  move .../skill.md.tmp-rename-1f2a035dbb65/skill.md to .../SKILL.md: statat ...: no such file or directory (restoring .claude/skills/moai-ref-owasp-checklist/skill.md also failed: move .../skill.md.tmp-rename-1f2a035dbb65/skill.md to .../skill.md: file exists; the content is in skill.md inside the folder this step opened as .claude/skills/moai-ref-owasp-checklist/skill.md.tmp-rename-1f2a035dbb65, which was moved)
```

In the first case the joined error carries a second line from the rmdir of the moved-away staging name (`staging folder ... not removed: ...`); the step does not hide it. Where a new empty folder was put at the stage name, the rmdir takes that empty folder back (the same `removeStage` behaviour as in every earlier row: it removes only an empty folder at our private random name).

Mutants (`go test -count=1 -overlay=<json> ./internal/cli/update/skillresidue/`; overlay JSON and mutated copies under `/private/tmp/t1733r6/ov/`, nothing in the tree):

| Mutant | Edit | Result |
|---|---|---|
| 1 put-back through the path look | `moveNoReplace(root, mid, from, !renameIsDescriptorBound)` -> `moveNoReplace(root, mid, from, true)` | exit 1, 3 test functions FAIL: `...WhenTheStageFolderWasMovedOutOfTheProject` (both variants), `...MovedOutBetweenTheRenames` (`skill.md = "" (present false), want the user's bytes back ... outside [. dir moved-stage dir moved-stage/skill.md file]`), `...IntoASkillFolderThatWasMovedAside` |
| 2 post-second-rename path check removed | the `root.Lstat(canonical)` identity check deleted | exit 1, 1 FAIL: `...DoesNotReportSuccessWhenTheSkillFolderWasMovedAndADecoyTookItsPath` (`the rename reported success although ... has no SKILL.md`) |
| 3 content place always claimed as a path | `contentPlace` returns "at <path>" regardless of `leadsTo` | exit 1, 1 FAIL: `...NamesTheMovedFolderWhenThePutBackIsRefused` |

### Existing tests (assertions untouched)

No existing test file was edited this round: `git diff --stat 9c74d7d0e..HEAD` before this docs commit lists `residue.go`, the three `rename_noreplace_{unix,windows,other}.go` files and the new `residue_moved_folder_test.go` only. `TestMoveFileReferencesArePinned` needed no change: `renameNoReplace` is still referenced once, from `moveNoReplace`. The `afterNthRename` / `mkdirFile` injections are used as in round 5 (the swap runs inside the seam, F24 of the audit still applies: a mutant that resolves the folders by path right before the seam call would survive; the faithful path-based mutant of round 5 stays killed).

### Honest residual (per platform)

- linux, darwin: a held folder renamed away after it was opened still receives the move, which is the property of binding to inodes, not a bug to remove. The step now restores the file by the held descriptors and says so; it does not prevent the move. If the put-back itself is refused (a user file took `skill.md`, or the filesystem refuses) the content stays in the moved folder and the error says "in skill.md inside the folder this step opened as <stage>, which was moved"; the moved folder's new path is not reported (it needs `F_GETPATH` / `/proc/self/fd`). A skill folder moved away after the first rename leaves SKILL.md inside the moved folder; the step reports an error, removes nothing and moves nothing back by path; the staging folder inside the moved folder stays. The user ends up with the renamed file in a folder they or another process moved, and a clear error. Both states need a same-user rename of a 48-bit-random, microsecond-lived staging folder, or of the skill folder, inside a microsecond gap.
- windows: unchanged in kind and now stated plainly. The rename is path-based (`MoveFileEx`), the put-back keeps the root-bound look, and a stage folder (or skill folder) moved away between the look and the call, or before the put-back, leaves the file in the moved folder: the look refuses the put-back, the error says the staging folder was moved. A move, not a deletion. The windows build is cross-built and vetted only.
- F23 (unobserved, windows): `srcDir.Name()` for a file opened through `Root.Open(".")` joins to a path under the project root only while the root name is absolute or the working directory is unchanged (read in `os/root_windows.go` by the auditor, not run); `removeDirAt` (`syscall.Rmdir`) runs while the `stageDir` handles are still open, which Go opens with FILE_SHARE_DELETE, so the removal should be a pending delete. Neither was observed on Windows in this round.
- Linux runtime of `Renameat2`, `Unlinkat` and `SyscallConn().Control` is still unobserved (compile and vet for linux/amd64; tests ran on darwin/arm64 APFS only).
- Platforms with no primitive and filesystems answering EINVAL/ENOSYS/ENOTSUP: the round-3 link fallback with its look-to-unlink windows, unchanged (E.7).
- Required codex gate for iteration 6 returned `fail` (rcpt-441c9ee2bd8c68ac93efb525); this round answers its two findings; a re-audit is the auditor's.

### Verification

Env: `GIT_CONFIG_GLOBAL=/private/tmp/t1733r6/gitconfig` (only `[user]`), `TMPDIR=/private/tmp/t1733r6/tmp`, `GOTMPDIR=/private/tmp/t1733r6/gotmp`, no HOME override, no `-mod=mod`. Run at GREEN `6888aff16` (the docs commit changes this file only):

```
go build ./...                                          -> build=0
gofmt -l internal/cli internal/cli/update internal/template -> (empty), gofmt=0
go vet ./internal/cli/update/...                        -> vet=0
GOOS=linux go vet ./internal/cli/update/skillresidue/   -> vet-linux=0
GOOS=windows go vet ./internal/cli/update/skillresidue/ -> vet-windows=0
GOOS=freebsd go build ./internal/cli/update/skillresidue/ -> build-freebsd=0
GOOS=linux GOARCH=amd64 go build ./...                  -> linuxbuild=0
GOOS=windows GOARCH=amd64 go build ./...                -> winbuild=0
go test -count=1 -race ./internal/cli/update/skillresidue/ -> ok ... 1.232s
go test -count=1 -cover ./internal/cli/update/...       -> ok update 0.775s 85.2%, backup 0.702s 88.4%, deploy 0.446s 85.9%, merge 0.444s 93.1%, plan 0.234s 95.2%, report 0.610s 95.1%, skillresidue 0.387s 89.7%
go test -count=1 ./internal/cli/ -run 'Backup|Retention|Destructive|Housekeeping|UBR|Unmanaged|TemplateSync|CleanReinstall' -> ok internal/cli 15.520s
go test -count=1 ./internal/template/...                -> ok template 137.081s, agentemit 0.365s, commandemit 0.242s, embedemit 0.255s, scripts [no test files]
```

Gaps: repository-wide test verdict PENDING (CI on the integration branch owns it); `make build` not run; Linux and Windows runtime behaviour of the descriptor-relative renames, `Unlinkat`, `MoveFileEx` and `syscall.Rmdir` not observed (cross-build and vet only); the real-filesystem tests ran on case-insensitive APFS only; the F23 Windows handle behaviour (item above) not observed; the auditor's real-syscall hostile-writer race stress (iteration 6, `Evidence` row 10) was not re-run this round; no tool call was refused this round; the `verify` skill delegation of the pre-commit gate was not separately run (the scoped commands above were); re-audit is the auditor's.
