# progress.md — SPEC-QUOTA-RECORD-WORKTREES-001 (card t1442)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete_at: 2026-10-02
plan_status: audit-ready  # artifacts authored by manager-spec; the independent plan-audit has not run
tier: M
artifacts: [spec.md, plan.md, acceptance.md, decision-index.md, progress.md]
requirements: 12
acceptance_criteria: 13
open_decisions: []  # none; D4 (decision-index.md Q2) and D2 (Q5) carry the leader's verdict, ACCEPTED provisionally (D4 values unmeasured, bound exposed as workflow.quota_gate.max_scan_dirs); D1, D5, D6 resolved by the decision oracle
spec_version: "0.5.0"  # plan-audit iterations 1 (FAIL 0.75) and 2 (FAIL 0.83) closed by amendment, dispositions in plan.md §J
planned_at_head: 284e09c44023598affe486f17701717ca173e6ca
```

## §E.2 Run-phase Evidence

### QWR-M0 — real-lane record-location measurement (baseline, own commit, precedes every implementation commit)

Protocol: plan.md §C. Measured by the lane session itself (lane-9, session `2da35a68-1196-4183-b6e5-a50fc9b6d901`), tree `.moai/worktrees/t1442` at HEAD `7a4a89548` (the Kickoff commit), 2026-10-02.

What this establishes: record locations and modification times only. It cannot show the gate reading any record (see "What it cannot establish" in plan.md §C).

Times: T0 `2026-10-02T19:51:57+0900`, T1 `2026-10-02T19:52:19+0900`, T2 `2026-10-02T19:52:23+0900`.

Installed build (judging build for every `moai` command below), `moai version`:

```
 v3.2.0-rc.25   moai_cp/20260925_122548-1952-g802a72235   built 2026-10-02T08:00:14Z
```

- `git merge-base --is-ancestor 802a72235 HEAD` printed nothing, harness reported no failure status (exit 0): the installed build is an ancestor of this tree.
- `git ls-tree --name-only 802a72235 internal/cli/factory_quota.go internal/statusline/quota.go` printed nothing: the installed build contains neither quota file, so no pressure evaluation exists in it.
- Lane anchor at the measured command: `git rev-parse --show-toplevel` printed `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1442`, a card worktree, not the parent checkout.

Copies of this session's record (`find <primary>/.moai/worktrees <primary>/.claude/worktrees -maxdepth 6 -path "*/.moai/state/context-usage/<SID>.json"` printed three, plus the primary directory's own copy checked by path):

```
-rw-r--r--@ 1 goos  staff  316 Oct  2 18:28 <primary>/.moai/state/context-usage/<SID>.json
-rw-r--r--@ 1 goos  staff  316 Oct  2 18:26 <primary>/.claude/worktrees/develop/.moai/state/context-usage/<SID>.json
-rw-r--r--@ 1 goos  staff  316 Oct  2 18:21 <primary>/.moai/worktrees/t1347/.moai/state/context-usage/<SID>.json
-rw-r--r--@ 1 goos  staff  314 Oct  2 19:51 <primary>/.moai/worktrees/t1442/.moai/state/context-usage/<SID>.json
```

The same four-line listing was printed again at T1 and at T2 with identical mtimes (no render fell inside the 26 s interval for the other three; the t1442 copy shows 19:51 at all three readings, minute resolution).

Freshest copy (`.moai/worktrees/t1442/.moai/state/context-usage/<SID>.json`, read after T0):

```
"schema_version": 2, "writer_pid": 336, "captured_at": "2026-10-02T19:51:59.052742+09:00", "raw_pct": 30, "stage": "none", "band": "large", "model": "Sonnet 5.5", "effort": "high"
```

No window field exists (schema 2).

Measured command: `moai factory status` (always; no `moai factory next` was run, no lease was due). Output (17 rows, one showing an already-expired lease for another lane, none changed by this call; verbatim rows begin `t587 run=tl4rkl state=completed (legacy) …` and end `t810 run=tm9i7y state=picked stage=- version=1 owner=- lease=none …`) is recorded here as run, and is not read as evidence about the gate.

Statements:
- (i) At the moment of the call the lane's own record was in the card worktree directory (`t1442`, 19:51), age under 1 minute against T0; the primary copy (18:28) was about 83 minutes old.
- (ii) Three stale copies of the same session id remain in other directories (18:28 primary, 18:26 develop, 18:21 t1347), 83 to 90 minutes behind the freshest copy: the shape the gate's freshest-capture-wins rule has to cope with.
- (iii) NOT established: the lane was anchored at a card worktree when it ran the command, so this observation does not show a parent-anchored lane's freshest record sitting in the primary directory (plan.md §C claim (iii) needs a parent-anchored lane; DP1 in plan.md §C is the only observation of that shape).

Gaps: end-to-end gate observation (the installed build has no quota code; post-install follow-up recorded for §E.3); `moai factory next` not run; mtimes are minute-resolution; whether the 26 s window contained a statusline render for the other directories is not observable from this data.

### M1 — stubs and RED tests (code+tests commit `0bb2048beb0d223ba2edfec65d3cafb1c900c9ef`, which also flips spec.md to `in-progress`)

Measured by the M1 milestone worker in the card tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1442`, branch `WT-quota-read-worktree-records`, 2026-10-02. Parent of the code commit: `60cb309ba` (QWR-M0 commit B). Judging build for every Go measurement below: `go version go1.26.8 darwin/arm64`, run through `go -C <tree> test` on this tree (no installed `moai` build is cited as evidence). Heavy runs (`./internal/cli`, `./internal/config`) held the `moai slot` lease `go-test-cli` and used the scrubbed single compound form `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go -C <tree> test …`.

#### Zero-sweep, before any test file was written (acceptance.md §B)

```
$ go -C <tree> test -count=1 -run '^(TestQWR_AC001_WorktreeOnlyRecordIsSeen|TestQWR_AC002_FreshestAcrossDirectories|TestQWR_AC003_SingleDirectoryUnchanged|TestQWR_AC004_UnusableEntriesContributeNothing|TestQWR_AC005_EnumerationBound|TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive|TestQWR_AC010_NoGitMetadataReadsRootOnly)$' ./internal/statusline
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.515s [no tests to run]
$ (scrubbed, lease held) go -C <tree> test -count=1 -run '^(TestQWR_AC007_GateDisabledReadsNothing|TestQWR_AC008_SurfacesSeeWorktreeReading|TestQWR_AC009_WorktreeSourceRendersIdentically|TestQWR_AC010b_CliRootWithoutGitMetadata|TestQWR_AC013b_ProductionSeamAppliesConfiguredBound)$' ./internal/cli
ok  	github.com/modu-ai/moai-adk/internal/cli	1.345s [no tests to run]
$ (scrubbed, lease held) go -C <tree> test -count=1 -run '^TestQWR_AC013_MaxScanDirsConfigKey$' ./internal/config
ok  	github.com/modu-ai/moai-adk/internal/config	0.373s [no tests to run]
```

#### The stubs (behaviour-neutral)

`internal/statusline/quota_dirs.go` (new): `QuotaStateDirs(root string, maxDirs int) []string` returns the primary state directory only; `AggregateQuotaDirs(stateDirs []string, now time.Time, maxAge time.Duration) QuotaAggregate` reads only its first directory through `AggregateQuota` (an empty list reads as unknown). `internal/config/types.go`: `MaxScanDirs int \`yaml:"max_scan_dirs"\`` on `QuotaGateConfig`; `internal/config/loader_quota_gate.go`: `MaxScanDirs int` on `QuotaGateSettings` (zero value; no default, no range, `resolveQuotaGate` untouched). `internal/cli/factory_quota.go`: `var factoryQuotaStateDirs = statusline.QuotaStateDirs` (referenced by no production code). `git diff 60cb309ba --stat` before the commit: 4 modified tracked files, 13 insertions, 1 deletion (spec.md `status:` line included).

#### E8 — verbatim RED output (after the stubs; failing assertions, never a compile error)

Statusline, `go -C <tree> test -count=1 -v -run '^(TestQWR_AC001_WorktreeOnlyRecordIsSeen|…|TestQWR_AC010_NoGitMetadataReadsRootOnly)$' ./internal/statusline` → exit 1, `FAIL github.com/modu-ai/moai-adk/internal/statusline 1.330s`. Verdict lines:

```
--- FAIL: TestQWR_AC001_WorktreeOnlyRecordIsSeen (0.23s)
    --- FAIL: TestQWR_AC001_WorktreeOnlyRecordIsSeen/moai_worktrees_layout (0.01s)
    --- FAIL: TestQWR_AC001_WorktreeOnlyRecordIsSeen/claude_worktrees_layout (0.01s)
    --- FAIL: TestQWR_AC001_WorktreeOnlyRecordIsSeen/outside_root_layout (0.01s)
    --- FAIL: TestQWR_AC001_WorktreeOnlyRecordIsSeen/entry_name_differs_from_directory (0.00s)
    --- FAIL: TestQWR_AC001_WorktreeOnlyRecordIsSeen/real_git_worktree_layout (0.21s)
--- FAIL: TestQWR_AC002_FreshestAcrossDirectories (0.05s)
    --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/stale_primary_copy_fresh_worktree_copy (0.01s)
    --- PASS: TestQWR_AC002_FreshestAcrossDirectories/newer_primary_copy_wins (0.01s)
    --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/not_max_across_dirs (0.00s)
    --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/not_min_across_dirs (0.01s)
    --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/per_window_across_dirs (0.01s)
    --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/tie_keeps_primary (0.00s)
    --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/skew_and_age_apply_to_worktree_records (0.01s)
        --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/skew_and_age_apply_to_worktree_records/skew (0.00s)
        --- FAIL: TestQWR_AC002_FreshestAcrossDirectories/skew_and_age_apply_to_worktree_records/age (0.01s)
--- PASS: TestQWR_AC003_SingleDirectoryUnchanged (0.02s)
    --- PASS: TestQWR_AC003_SingleDirectoryUnchanged/single_dir_equals_multi_dir_of_one (0.02s)
    --- PASS: TestQWR_AC003_SingleDirectoryUnchanged/signature_unchanged (0.00s)
    --- PASS: TestQWR_AC003_SingleDirectoryUnchanged/capture_age_independent_of_mtime (0.00s)
--- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing (0.07s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/gitdir_unreadable (0.01s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/gitdir_empty (0.00s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/gitdir_garbage_not_a_path (0.00s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/gitdir_relative_path (0.00s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/gitdir_oversized (0.01s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/gitdir_first_line_is_the_path (0.01s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/target_missing_pruned (0.01s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/record_dir_absent (0.00s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/record_dir_is_a_file (0.00s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/record_dir_symlink_loop (0.01s)
    --- FAIL: TestQWR_AC004_UnusableEntriesContributeNothing/locked_entry_is_still_read (0.01s)
    --- PASS: TestQWR_AC004_UnusableEntriesContributeNothing/worktrees_dir_unreadable (0.00s)
--- FAIL: TestQWR_AC005_EnumerationBound (0.46s)
    --- FAIL: TestQWR_AC005_EnumerationBound/bound_128_entry_128_visible (0.00s)
    --- FAIL: TestQWR_AC005_EnumerationBound/bound_128_entry_129_and_130_not_read (0.00s)
    --- FAIL: TestQWR_AC005_EnumerationBound/bound_3_reads_three (0.00s)
    --- FAIL: TestQWR_AC005_EnumerationBound/bound_1_reads_one (0.00s)
    --- FAIL: TestQWR_AC005_EnumerationBound/unusable_entries_consume_the_bound (0.01s)
    --- FAIL: TestQWR_AC005_EnumerationBound/gitdir_read_is_bounded (0.01s)
--- FAIL: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive (0.03s)
    --- PASS: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive/swept_set_names_both_files (0.00s)
    --- PASS: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive/no_network_or_process_imports (0.00s)
    --- PASS: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive/no_write_walk_or_mutation_selectors (0.00s)
    --- PASS: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive/new_functions_declared_in_swept_files_only (0.02s)
    --- PASS: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive/calls_resolve_to_the_two_files_or_the_reader (0.01s)
    --- FAIL: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive/run_leaves_every_tree_byte_identical (0.01s)
--- PASS: TestQWR_AC010_NoGitMetadataReadsRootOnly (0.02s)
    --- PASS: TestQWR_AC010_NoGitMetadataReadsRootOnly/no_git_entry (0.01s)
    --- PASS: TestQWR_AC010_NoGitMetadataReadsRootOnly/git_is_a_file (0.01s)
    --- PASS: TestQWR_AC010_NoGitMetadataReadsRootOnly/no_worktrees_dir (0.01s)
```

Failing assertions (statusline), in the order of the subtests above (`<R>` abbreviates the printed zero reading `(reading {State:unknown UsedPercentage:0 ResetsAt:0 CapturedAt:0001-01-01 00:00:00 +0000 UTC ExhaustedAt:0001-01-01 00:00:00 +0000 UTC})`; every other character is verbatim):

```
quota_dirs_test.go:137: five_hour: state = "unknown", want "fresh" <R>      (AC001 moai_worktrees_layout)
quota_dirs_test.go:141: five_hour: state = "unknown", want "fresh" <R>      (AC001 claude_worktrees_layout)
quota_dirs_test.go:145: five_hour: state = "unknown", want "fresh" <R>      (AC001 outside_root_layout)
quota_dirs_test.go:149: five_hour: state = "unknown", want "fresh" <R>      (AC001 entry_name_differs_from_directory)
quota_dirs_test.go:181: five_hour: state = "unknown", want "fresh" <R>      (AC001 real_git_worktree_layout)
quota_dirs_test.go:202: five_hour: used = 95, want 60                        (AC002 stale_primary_copy_fresh_worktree_copy)
quota_dirs_test.go:214: five_hour: used = 97, want 70                        (AC002 not_max_across_dirs)
quota_dirs_test.go:220: five_hour: used = 40, want 80                        (AC002 not_min_across_dirs)
quota_dirs_test.go:229: five_hour: state = "unknown", want "fresh" <R>      (AC002 per_window_across_dirs)
quota_dirs_test.go:243: seven_day: state = "unknown", want "fresh" <R>      (AC002 tie_keeps_primary)
quota_dirs_test.go:254: five_hour: state = "unknown", want "fresh" <R>      (AC002 skew)
quota_dirs_test.go:264: seven_day: state = "unknown", want "fresh" <R>      (AC002 age)
quota_dirs_test.go:396..459: five_hour: state = "unknown", want "fresh" <R>  (AC004, eleven subtests, one line each: lines 396, 405, 408, 411, 414, 421, 430, 433, 436, 447, 459)
quota_dirs_test.go:499: bound 128: state = "unknown", want "fresh" <R>      (AC005 bound_128_entry_128_visible)
quota_dirs_test.go:500: bound 127: state = "unknown", want "fresh" <R>
quota_dirs_test.go:504: bound 128: state = "unknown", want "fresh" <R>      (AC005 bound_128_entry_129_and_130_not_read)
quota_dirs_test.go:510: bound 3: state = "unknown", want "fresh" <R>        (AC005 bound_3_reads_three)
quota_dirs_test.go:513: bound 1: state = "unknown", want "fresh" <R>        (AC005 bound_1_reads_one)
quota_dirs_test.go:524: bound 3: state = "unknown", want "fresh" <R>        (AC005 unusable_entries_consume_the_bound)
quota_dirs_test.go:525: bound 4: state = "unknown", want "fresh" <R>
quota_dirs_test.go:538: five_hour: state = "unknown", want "fresh" <R>      (AC005 gitdir_read_is_bounded)
quota_dirs_test.go:735: five_hour (the run must reach the worktree directories): used = 40, want 62   (AC006 run_leaves_every_tree_byte_identical)
```

Config, `unset … && go -C <tree> test -count=1 -v -run '^TestQWR_AC013_MaxScanDirsConfigKey$' ./internal/config` → exit 1, `FAIL github.com/modu-ai/moai-adk/internal/config 0.384s`; all eight subtests FAIL (`default_equals_template`, `template_comment_says_unmeasured`, `local_twin_carries_key`, `absent_or_unparseable_yields_default`, `out_of_range_numeric_yields_default`, `type_mismatch_defaults_the_whole_block_gate_off`, `inventory_row_with_reader`, `cache_schema_bumped`). Failing assertions, verbatim:

```
quota_gate_scan_dirs_test.go:46: Go default QuotaGate.MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:49: template workflow.quota_gate.max_scan_dirs = 0, want 128
quota_gate_scan_dirs_test.go:53: the template quota_gate block does not carry the key max_scan_dirs:
quota_gate_scan_dirs_test.go:69: no template comment line names max_scan_dirs
quota_gate_scan_dirs_test.go:84: local workflow.quota_gate.max_scan_dirs = 0, want 128
quota_gate_scan_dirs_test.go:87: the local quota_gate block does not carry the key max_scan_dirs:
quota_gate_scan_dirs_test.go:103: no_file: LoadQuotaGate().MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:103: unparseable_file: LoadQuotaGate().MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:103: key_absent: LoadQuotaGate().MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:107: DefaultQuotaGate().MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:124: max_scan_dirs 0: MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:124: max_scan_dirs -1: MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:124: max_scan_dirs 1025: MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:124: max_scan_dirs 1: MaxScanDirs = 0, want 1
quota_gate_scan_dirs_test.go:124: max_scan_dirs 3: MaxScanDirs = 0, want 3
quota_gate_scan_dirs_test.go:124: max_scan_dirs 128: MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:124: max_scan_dirs 1024: MaxScanDirs = 0, want 1024
quota_gate_scan_dirs_test.go:146: string_value: LoadQuotaGate = {Enabled:false FiveHourHoldPct:90 SevenDayHoldPct:95 ReleaseMarginPct:5 MaxAge:30m0s MaxScanDirs:0}, want the whole-block defaults {Enabled:false FiveHourHoldPct:90 SevenDayHoldPct:95 ReleaseMarginPct:5 MaxAge:30m0s MaxScanDirs:128}
quota_gate_scan_dirs_test.go:146: int_overflow: LoadQuotaGate = {Enabled:false FiveHourHoldPct:90 SevenDayHoldPct:95 ReleaseMarginPct:5 MaxAge:30m0s MaxScanDirs:0}, want the whole-block defaults {Enabled:false FiveHourHoldPct:90 SevenDayHoldPct:95 ReleaseMarginPct:5 MaxAge:30m0s MaxScanDirs:128}
quota_gate_scan_dirs_test.go:181: the shipped-key inventory has no row for workflow.quota_gate.max_scan_dirs
quota_gate_scan_dirs_test.go:187: configCacheSchemaVersion = 11, want 12 — Workflow.QuotaGate gained MaxScanDirs
```

Observation (plan.md U21, now observed at M1): in the `type_mismatch` subtest the string value and the integer literal `99999999999999999999` both left `Enabled:false` and `FiveHourHoldPct:90` in the result although the fixture wrote `enabled: true` and `five_hour_hold_pct: 80`, i.e. the YAML decode failed and the whole block fell back to the defaults (the failing field is only `MaxScanDirs:0` against the asserted 128). A first draft of the test also tried the real number `2.5`: it did NOT fail the decode (`Enabled:true FiveHourHoldPct:80` came back), so yaml.v3 accepts a non-integral number for an `int` field; the case was removed from the test because no AC names it, and spec.md §F / acceptance.md §F "any non-integer" is therefore too broad (string and overflow are observed; `2.5` is not a decode failure).

Cli, `unset … && go -C <tree> test -count=1 -v -run '^(TestQWR_AC007_GateDisabledReadsNothing|TestQWR_AC008_SurfacesSeeWorktreeReading|TestQWR_AC009_WorktreeSourceRendersIdentically|TestQWR_AC010b_CliRootWithoutGitMetadata|TestQWR_AC013b_ProductionSeamAppliesConfiguredBound)$' ./internal/cli` → exit 1, `FAIL github.com/modu-ai/moai-adk/internal/cli 46.088s`. Verdict lines:

```
--- FAIL: TestQWR_AC007_GateDisabledReadsNothing (1.25s)
    --- PASS: TestQWR_AC007_GateDisabledReadsNothing/disabled_zero_calls (0.74s)
    --- FAIL: TestQWR_AC007_GateDisabledReadsNothing/enabled_one_call_each (0.20s)
    --- PASS: TestQWR_AC007_GateDisabledReadsNothing/enumerator_referenced_only_by_the_variable (0.05s)
    --- PASS: TestQWR_AC007_GateDisabledReadsNothing/missing_dir_fixture_untouched (0.26s)
--- FAIL: TestQWR_AC008_SurfacesSeeWorktreeReading (18.40s)
    --- FAIL: TestQWR_AC008_SurfacesSeeWorktreeReading/next_holds_on_worktree_only_92 (3.98s)
    --- FAIL: TestQWR_AC008_SurfacesSeeWorktreeReading/mcp_next_holds_on_worktree_only_92 (2.48s)
    --- PASS: TestQWR_AC008_SurfacesSeeWorktreeReading/assigned_card_still_leased (3.07s)
    --- PASS: TestQWR_AC008_SurfacesSeeWorktreeReading/leases_when_no_record_anywhere (2.67s)
    --- FAIL: TestQWR_AC008_SurfacesSeeWorktreeReading/status_block_pressure (2.89s)
    --- FAIL: TestQWR_AC008_SurfacesSeeWorktreeReading/acquire_warning_for_claude_caller (0.88s)
    --- FAIL: TestQWR_AC008_SurfacesSeeWorktreeReading/wait_recheck_sees_worktree_update (2.43s)
--- FAIL: TestQWR_AC009_WorktreeSourceRendersIdentically (19.46s)
    --- FAIL: TestQWR_AC009_WorktreeSourceRendersIdentically/status_text (6.58s)
    --- FAIL: TestQWR_AC009_WorktreeSourceRendersIdentically/status_json (6.80s)
    --- FAIL: TestQWR_AC009_WorktreeSourceRendersIdentically/hold_line (5.47s)
    --- FAIL: TestQWR_AC009_WorktreeSourceRendersIdentically/acquire_warning (0.62s)
--- PASS: TestQWR_AC010b_CliRootWithoutGitMetadata (5.16s)
    --- PASS: TestQWR_AC010b_CliRootWithoutGitMetadata/unregistered_directory_is_not_read (2.99s)
    --- PASS: TestQWR_AC010b_CliRootWithoutGitMetadata/roots_own_record_is_read (2.17s)
--- FAIL: TestQWR_AC013b_ProductionSeamAppliesConfiguredBound (0.66s)
    --- FAIL: TestQWR_AC013b_ProductionSeamAppliesConfiguredBound/bound_3_reads_three (0.00s)
    --- FAIL: TestQWR_AC013b_ProductionSeamAppliesConfiguredBound/key_absent_reads_all_five (0.00s)
```

Failing assertions (cli), verbatim (the `WARN config sections directory not found` log lines the run interleaves are omitted):

```
factory_quota_worktrees_test.go:246: evaluation: seam calls = aggregator 1, enumerator 0, want 1 and 1
factory_quota_worktrees_test.go:246: lane gate: seam calls = aggregator 1, enumerator 0, want 1 and 1
factory_quota_worktrees_test.go:246: status block: seam calls = aggregator 1, enumerator 0, want 1 and 1
factory_quota_worktrees_test.go:339: want hold line, got a leased card (stdout "moai: worktree commits under global git identity F1 Test <f1@example.invalid> (read-only from ~/.gitconfig)\nt2 stage=- worktree=t2\nt2\tunknown\t\t\tpicked\t\tfactory card 2\n", stderr "note: open pull requests unavailable (exit status 1); link column left empty, landed check still ran\nnote: the landed check against origin/main could not answer for t2; those cards report unknown rather than no-link, because an unanswerable query is not evidence of not-landed\n")
factory_quota_worktrees_test.go:362: held result text = "t2 stage=- worktree=t2\nt2\tunknown\t\t\tpicked\t\tfactory card 2", want exactly the hold line "quota hold: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z"
factory_quota_worktrees_test.go:397: status --json carries no quota block for a worktree-only 92% reading:   (followed by the card-only JSON object, `"unavailable": []` as its last key)
factory_quota_worktrees_test.go:420: stderr quota lines = [], want exactly one matching ^quota warning: [a-z_]+ used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z(; [a-z_]+ used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z)* \(warn-only; the integration window is still taken\)$
factory_quota_worktrees_test.go:456: wait slept 0 times, want 1 (held at 92 on the first pass, released at 80 on the second)
factory_quota_worktrees_test.go:502: status text differs between the primary-directory and worktree-directory fixtures:
        primary:  "t2 run=run-cli state=picked stage=- version=1 owner=- lease=none gate=- prefer=- after=- contract=none mode=parallelizable priority=normal\nquota five_hour: state=fresh used=92.0% resets_at=2026-09-26T12:00:00Z captured_at=2026-09-26T09:00:00Z holds_claude_lane=yes\nquota seven_day: state=fresh used=96.0% resets_at=2026-09-29T09:00:00Z captured_at=2026-09-26T09:00:00Z holds_claude_lane=yes\nquota pressure: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z; seven_day used=96.0% resets_at=2026-09-29T09:00:00Z; warning: no live non-Claude lane (unknown backend: 1); nothing is re-dispatched\n"
        worktree: "t2 run=run-cli state=picked stage=- version=1 owner=- lease=none gate=- prefer=- after=- contract=none mode=parallelizable priority=normal\n"
factory_quota_worktrees_test.go:510: status --json differs between the primary-directory and worktree-directory fixtures:   (excerpt: both strings share the identical card block; the primary one continues `…\"unavailable\": [],\n  \"quota\": {\n    \"pressure\": true,\n    \"windows\": [ … two fresh windows … ],\n    \"candidates\": [],\n    \"unknown_lanes\": 1,\n    \"warning\": \"no-non-claude-lane\"\n  }\n}\n`, the worktree one ends `…\"unavailable\": []\n}\n`)
factory_quota_worktrees_test.go:519: next output (stdout|stderr|error) differs between the primary-directory and worktree-directory fixtures:
        primary:  "|quota hold: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z; seven_day used=96.0% resets_at=2026-09-29T09:00:00Z\n|factory next: quota hold: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z; seven_day used=96.0% resets_at=2026-09-29T09:00:00Z"
        worktree: "moai: worktree commits under global git identity F1 Test <f1@example.invalid> (read-only from ~/.gitconfig)\nt2 stage=- worktree=t2\nt2\tunknown\t\t\tpicked\t\tfactory card 2\n|note: open pull requests unavailable (exit status 1); link column left empty, landed check still ran\nnote: the landed check against origin/main could not answer for t2; those cards report unknown rather than no-link, because an unanswerable query is not evidence of not-landed\n|<nil>"
factory_quota_worktrees_test.go:529: acquire output (stdout|stderr|lock) differs between the primary-directory and worktree-directory fixtures:
        primary:  "release-integration window acquired by sess-qas on release/v9.9.9\n|quota warning: five_hour used=92.0% resets_at=2026-09-26T12:00:00Z; seven_day used=96.0% resets_at=2026-09-29T09:00:00Z (warn-only; the integration window is still taken)\n|{\"session_id\":\"sess-qas\",\"session_name\":\"lane-1\",\"pid\":11196,\"pid_source\":\"session-owner\",\"branch\":\"release/v9.9.9\",\"worktree\":\"\",\"acquired_at\":\"\",\"branch_source\":\"flag\"}"
        worktree: "release-integration window acquired by sess-qas on release/v9.9.9\n||{\"session_id\":\"sess-qas\",\"session_name\":\"lane-1\",\"pid\":11196,\"pid_source\":\"session-owner\",\"branch\":\"release/v9.9.9\",\"worktree\":\"\",\"acquired_at\":\"\",\"branch_source\":\"flag\"}"
factory_quota_worktrees_test.go:609: seam calls per evaluation = aggregator 1, enumerator 0, want 1 and 1
factory_quota_worktrees_test.go:609: five_hour reading = unknown 0, want fresh 70
factory_quota_worktrees_test.go:613: seam calls per evaluation = aggregator 1, enumerator 0, want 1 and 1
factory_quota_worktrees_test.go:613: five_hour reading = unknown 0, want fresh 80
```

#### TDD classification (`.claude/rules/moai/workflow/tdd-result-contract.md`) and which tests are GREEN on the stubs

- EXPECTED_RED (assertion about the missing behaviour): AC-QWR-001 (all 5 subtests), AC-QWR-002 (6 of 7 subtests; `newer_primary_copy_wins` is GREEN on the stub), AC-QWR-004 (11 of 12 subtests; `worktrees_dir_unreadable` is GREEN), AC-QWR-005 (all 6), AC-QWR-006 `run_leaves_every_tree_byte_identical` (the positive control that the run reaches the worktree directories), AC-QWR-007 `enabled_one_call_each`, AC-QWR-008 (5 of 7: `next_holds_on_worktree_only_92`, `mcp_next_holds_on_worktree_only_92`, `status_block_pressure`, `acquire_warning_for_claude_caller`, `wait_recheck_sees_worktree_update`), AC-QWR-009 (all 4), AC-QWR-013 (all 8 config subtests), AC-QWR-013b (both subtests).
- GREEN on the stubs by design (regression guards; they cannot be made to fail on an assertion with behaviour-neutral stubs without weakening them): AC-QWR-003 (all 3 subtests, the single-directory entry point equals a one-element multi-directory reading), AC-QWR-006 static subtests (5 of 6), AC-QWR-007 `disabled_zero_calls`, `enumerator_referenced_only_by_the_variable`, `missing_dir_fixture_untouched`, AC-QWR-008 `assigned_card_still_leased` and `leases_when_no_record_anywhere`, AC-QWR-010 (3 statusline subtests), AC-QWR-010b (both cli subtests), AC-QWR-002 `newer_primary_copy_wins`, AC-QWR-004 `worktrees_dir_unreadable`. None of these is claimed as RED.
- No test was classified REGRESSION_FAILURE or TOOL_FAILURE. `BenchmarkQWR_RealRoot` compiles and prints `--- SKIP: BenchmarkQWR_RealRoot` / `MOAI_QWR_BENCH_ROOT is unset: no repository root to measure` under `-run '^$' -bench BenchmarkQWR_RealRoot -benchtime 1x -v`.

#### Predecessor guards (unedited, green on the stubs)

```
$ (scrubbed, lease held) go -C <tree> test ./internal/statusline ./internal/config -run '^(TestQAS_AC001_RecordCarriesSuppliedWindowsOnly|TestQAS_AC002_PreviousSchemaReadsAsNoWindows|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline|TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop|TestQAS_AC004_ExhaustedAtStickyUntilRollover|TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover|TestQAS_AC006_FailOpenOnAbsentOrUnreadable|TestQAS_AC007_ConfigDefaultsMirrorTemplate|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$' -count=1 -v
ok  	github.com/modu-ai/moai-adk/internal/statusline	0.692s
ok  	github.com/modu-ai/moai-adk/internal/config	0.195s
```
nine top-level `--- PASS: TestQAS_…` lines (`AC001`, `AC014`, `AC002`, `AC002b`, `AC003`, `AC006`, `AC004`, `AC005` in statusline, `AC007` in config) and 44 subtest `--- PASS` lines, no FAIL, no SKIP.

```
$ (scrubbed, lease held) go -C <tree> test ./internal/cli -run '^(TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent|TestQAS_AC008_ClaudeLaneHeldAtThreshold|TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown|TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData|TestQAS_AC017_SharedPressureEvaluationAndSurfaces|TestQAS_AC022_GateOffOrPressureOffOutputUnchanged|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$' -count=1 -v
--- PASS: TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData (12.72s)
--- PASS: TestQAS_AC022_GateOffOrPressureOffOutputUnchanged (12.60s)
--- PASS: TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent (11.51s)
--- PASS: TestQAS_AC008_ClaudeLaneHeldAtThreshold (18.20s)
--- PASS: TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown (16.01s)
--- PASS: TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly (0.27s)
--- PASS: TestQAS_AC017_SharedPressureEvaluationAndSurfaces (11.72s)
ok  	github.com/modu-ai/moai-adk/internal/cli	84.075s
```

```
$ (scrubbed, lease held) go -C <tree> test -count=1 -v -run '^(TestShippedConfigKeysHaveReaders|TestQAS_AC007_ConfigDefaultsMirrorTemplate)$' ./internal/config
--- PASS: TestShippedConfigKeysHaveReaders (6.91s)
--- PASS: TestQAS_AC007_ConfigDefaultsMirrorTemplate (0.03s)   (nine subtests PASS)
ok  	github.com/modu-ai/moai-adk/internal/config	7.108s
```

Whole-package statusline run, once at the end: `go -C <tree> test -count=1 -v ./internal/statusline` → exit 1, exactly five failing top-level tests (`TestQWR_AC001`, `AC002`, `AC004`, `AC005`, `AC006`), 342 top-level `--- PASS` lines, no other FAIL.

#### Other M1 checks

- `go -C <tree> vet ./internal/statusline ./internal/config ./internal/cli` → exit 0, no output. `GOOS=windows GOARCH=amd64 go -C <tree> build ./...` → exit 0, no output. `GOOS=windows GOARCH=amd64 go -C <tree> vet ./internal/statusline` → exit 0 (the statusline test file compiles for windows).
- `go -C <tree> run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./internal/statusline/... ./internal/config/...` → `0 issues.`; the same for `./internal/cli/...` → `0 issues.` (a first run reported three ST1023 findings in the statusline test file, fixed before the commit).
- `grep -rn AskUserQuestion` over the four changed non-test Go files printed nothing.
- PRESERVE (AC-QWR-012 shape, working tree against `60cb309ba`): `git diff --name-only 60cb309ba -- internal/stateanchor internal/statusline/state_anchor.go internal/statusline/context_usage.go .moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001 internal/statusline/quota_test.go internal/cli/factory_quota_test.go internal/config/workflow_quota_gate_test.go` printed nothing; positive control `git diff --name-only 60cb309ba` printed `.moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/spec.md`, `internal/cli/factory_quota.go`, `internal/config/loader_quota_gate.go`, `internal/config/types.go` (non-empty, so the empty first result is a measured absence). The two config files touched are inside the AC-QWR-012 allowed set.

#### Observation for the leader (an M2 plan defect, found at M1, no file touched)

plan.md §G U13 expected the predecessor's `TestQAS_AC007_ConfigDefaultsMirrorTemplate` to survive M2. It will not: its subtests `defaults_equal_template`, `loader_reads_a_configured_block`, `absent_or_unparseable_yields_default`, and `out_of_range_yields_default` compare the whole `QuotaGateSettings` struct (`DefaultQuotaGate() != want`, `LoadQuotaGate(root) != want`) against literals that carry no `MaxScanDirs`, so they stay green only while `MaxScanDirs` is zero (as in the M1 stub) and turn red the moment M2 makes `DefaultQuotaGate()` carry 128 — which AC-QWR-013 requires. REQ-QWR-010 and AC-QWR-003/-012 forbid editing that file. M2 therefore needs a leader decision before any config default is written (for example: the bound is read through a separate accessor instead of a field of `QuotaGateSettings`, which `TestQWR_AC013` and the cli seam would then follow; or an explicit, approved edit of the predecessor test).

#### Gaps

- The cli and config zero-sweep selectors were issued in one assistant turn and so may have overlapped in time (the brief says never run two Go test builds at once); both printed `[no tests to run]` and neither is a verdict.
- The statusline, config, and cli RED runs were each run once after the final test text; the statusline selector was run twice (a lint fix to one subtest in between), the second run is the one quoted.
- `real_git_worktree_layout` ran (git present, not skipped); `gitdir_unreadable` ran (not root, not Windows). Windows-only skips were not exercised on Windows (a Gap there).
- The benchmark switch `MOAI_QWR_BENCH_ROOT` is a name chosen by the worker (neither acceptance.md nor plan.md names the variable); it is test-only and never read by production code.
- Not measured at M1: coverage of the changed packages (E3), the template/`make build` path (M2), any behaviour of the real bodies.

#### Explicit wait (lane-9, recorded 2026-10-02)

Reason: the M2 plan defect above needs a decision; the lane holds the Kickoff gate's evidence (plan-audit PASS 0.87 at hashes unchanged since audited_sha bb6b9925e) and a plan-artifact revision would invalidate it. Asked: the leader (session `leader`), by cross-session message, with options A (separate accessor outside `QuotaGateSettings`, needs spec.md REQ-QWR-011 and acceptance.md AC-QWR-013 revision and a re-audit, the 3-iteration ceiling is spent), B (approved edit of the four predecessor subtests), C (struct default 0, read-site default 128). Lane recommendation: A. Recheck point: the next turn after a leader reply arrives, or the next awaken, whichever comes first, starting with the stall watchdog. No M2 file is touched while waiting.

Resolved the same day by the leader (cross-session message): option A. See §J "Run-phase decision R1".

### M1 rework + M2 — accessor stub and RED tests (commit `e350a8139`), then the key release (commit `7ff94b33f`)

Measured in the card tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1442`, branch `WT-quota-read-worktree-records`, 2026-10-02, parent of the first commit `349bf40ab`. Judging build: `go version go1.26.8 darwin/arm64`, run through `go -C <tree> test` on this tree (no installed `moai` build is cited as evidence; `make build` produced `bin/moai` from commit `e350a8139` plus the uncommitted M2 working tree, used only for the embed check below). Config runs held the `moai slot` lease `go-test-cli` and used the scrubbed single compound form `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go -C <tree> test …`; the lease was released after the last config run. A shell exit code cannot be echoed in this worktree (the guard refuses `; echo $?`), so a passing run is evidenced by the printed `ok` / `PASS` lines and a failing run by the tool-reported `Exit code 1` plus `FAIL`.

#### What the M1 rework changed (commit `e350a8139`, `test(...): M1 rework accessor stub and RED tests`)

`QuotaGateSettings.MaxScanDirs` removed from `internal/config/loader_quota_gate.go` (the M1 stub field of `0bb2048be`); the signature-only accessor stub `LoadQuotaScanBound(projectRoot string) int` added (returns 0, no default, no range); `internal/config/quota_gate_scan_dirs_test.go` reworked so every AC-QWR-013 subtest reads the accessor (the whole-block type-mismatch case compares `LoadQuotaGate(root)` with `DefaultQuotaGate()`). The YAML decode field `QuotaGateConfig.MaxScanDirs` (types.go) stays. No predecessor test file is edited. `go vet ./internal/config` clean and `golangci-lint v2.1.6 run ./internal/config/...` → `0 issues.` on the stub tree. Spec status was already flipped at `0bb2048be`; neither this commit nor the key-release commit carries the `Authored-By-Agent` trailer.

#### E8 — verbatim RED (accessor stub, before any default or accessor body), AC-QWR-013

`unset … && go -C <tree> test -count=1 -v -run '^TestQWR_AC013_MaxScanDirsConfigKey$' ./internal/config` → tool-reported `Exit code 1`, `FAIL github.com/modu-ai/moai-adk/internal/config 0.476s`; eight subtests ran and all eight FAIL (swept 8, not `[no tests to run]`). Classification (tdd-result-contract): EXPECTED_RED, assertions about the missing behaviour, never a compile error. Failing assertions, verbatim:

```
quota_gate_scan_dirs_test.go:49: LoadQuotaScanBound on a root with no workflow.yaml = 0, want 128
quota_gate_scan_dirs_test.go:53: Go default seeded into the YAML struct QuotaGate.MaxScanDirs = 0, want 128
quota_gate_scan_dirs_test.go:56: template workflow.quota_gate.max_scan_dirs = 0, want 128
quota_gate_scan_dirs_test.go:65: the template quota_gate block does not carry the key max_scan_dirs:   (followed by the five-line block: enabled false, 90, 95, 5, max_age 30m)
quota_gate_scan_dirs_test.go:81: no template comment line names max_scan_dirs
quota_gate_scan_dirs_test.go:96: local workflow.quota_gate.max_scan_dirs = 0, want 128
quota_gate_scan_dirs_test.go:99: the local quota_gate block does not carry the key max_scan_dirs:   (followed by the five-line block)
quota_gate_scan_dirs_test.go:115: no_file: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:115: unparseable_file: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:115: key_absent: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:132: max_scan_dirs 0: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:132: max_scan_dirs -1: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:132: max_scan_dirs 1025: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:132: max_scan_dirs 1: LoadQuotaScanBound = 0, want 1
quota_gate_scan_dirs_test.go:132: max_scan_dirs 3: LoadQuotaScanBound = 0, want 3
quota_gate_scan_dirs_test.go:132: max_scan_dirs 128: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:132: max_scan_dirs 1024: LoadQuotaScanBound = 0, want 1024
quota_gate_scan_dirs_test.go:161: int_overflow: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:161: string_value: LoadQuotaScanBound = 0, want 128
quota_gate_scan_dirs_test.go:196: the shipped-key inventory has no row for workflow.quota_gate.max_scan_dirs
quota_gate_scan_dirs_test.go:202: configCacheSchemaVersion = 11, want 12 — Workflow.QuotaGate gained MaxScanDirs
```

Observation: in `type_mismatch_defaults_the_whole_block_gate_off` the whole-block half (`LoadQuotaGate(root)` equals `DefaultQuotaGate()` for the string value and the overflowing integer literal, gate off, no bound in the settings) already held at RED; only the accessor half failed. That is the intended split: the predecessor's behaviour is pinned as unchanged and the accessor's 128 is the new behaviour.

#### Predecessor guards after the stub-field removal and BEFORE any default changed (stub tree, commit `e350a8139`)

`unset … && go -C <tree> test -count=1 -v -run '^(TestQAS_AC007_ConfigDefaultsMirrorTemplate|TestShippedConfigKeysHaveReaders)$' ./internal/config` (the only `TestQAS_` test in `internal/config` is `TestQAS_AC007_ConfigDefaultsMirrorTemplate`, read by grep of `^func TestQAS_`) → `--- PASS: TestShippedConfigKeysHaveReaders (22.23s)` with its four subtests PASS, `--- PASS: TestQAS_AC007_ConfigDefaultsMirrorTemplate (0.16s)` with its nine subtests PASS (`defaults_equal_template`, `template_ships_off`, `template_comment_says_unmeasured`, `local_twin_enabled`, `loader_reads_a_configured_block`, `absent_or_unparseable_yields_default`, `out_of_range_yields_default`, `max_age_below_twice_heartbeat_yields_default`, `cache_schema_bumped`), `ok  github.com/modu-ai/moai-adk/internal/config 22.643s`. Two top-level tests swept.

#### GREEN (key release, commit `7ff94b33f`, `feat(...): M2 scan bound accessor and key`)

Everything of M2 is in this one commit (decision D-R6: the Go seed and the template and local-twin key cannot be separated): `DefaultQuotaGateMaxScanDirs = 128` in `defaults.go` beside the `DefaultQuotaGate*` constants and seeded into the YAML struct in `NewDefaultWorkflowConfig`; the range constants `quotaGateScanDirsMin = 1` and `quotaGateScanDirsMax = 1024` and the accessor body in `loader_quota_gate.go` (same `loadYAMLFile` and defaults-seeded `workflowFileWrapper` as `LoadQuotaGate`); `configCacheSchemaVersion` 11 to 12 with a comment line; the inventory row `workflow.quota_gate.max_scan_dirs` (class W, evidence reader) between `max_age` and `release_margin_pct`; the key `max_scan_dirs: 128` plus a comment (one comment line names the key and says `unmeasured`, a second says a project with more worktrees than the key needs a larger value) in the shipped template block; the key and a comment in the local twin. `git diff 60cb309ba -- internal/config/loader_quota_gate.go` shows only additions (the range constants and the accessor): `QuotaGateSettings`, `DefaultQuotaGate`, `resolveQuotaGate`, and `LoadQuotaGate` are byte-identical to the predecessor.

```
$ unset … && go -C <tree> test -count=1 -v -run '^TestQWR_AC013_MaxScanDirsConfigKey$' ./internal/config
--- PASS: TestQWR_AC013_MaxScanDirsConfigKey (0.10s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/default_equals_template (0.03s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/template_comment_says_unmeasured (0.00s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/local_twin_carries_key (0.00s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/absent_or_unparseable_yields_default (0.01s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/out_of_range_numeric_yields_default (0.03s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/type_mismatch_defaults_the_whole_block_gate_off (0.02s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/inventory_row_with_reader (0.01s)
    --- PASS: TestQWR_AC013_MaxScanDirsConfigKey/cache_schema_bumped (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/config	0.807s
```

Swept: 1 top-level test, 8 subtests (the RED run printed the same eight names). The cli half of AC-QWR-013 (`TestQWR_AC013b_ProductionSeamAppliesConfiguredBound`) is an M4 item and was not run here.

```
$ unset … && go -C <tree> test -count=1 -v -run '^(TestQAS_AC007_ConfigDefaultsMirrorTemplate|TestShippedConfigKeysHaveReaders)$' ./internal/config
--- PASS: TestShippedConfigKeysHaveReaders (15.92s)   (four subtests PASS)
--- PASS: TestQAS_AC007_ConfigDefaultsMirrorTemplate (0.13s)   (the same nine subtests PASS, including defaults_equal_template and cache_schema_bumped)
ok  	github.com/modu-ai/moai-adk/internal/config	16.260s
```

`max_scan_dirs` does not appear among the keys the guard's diagnostic lists as dead or unbound (grep of the saved output printed nothing), so the inventory row's reader (`LoadQuotaScanBound`) resolved. Whole package: `unset … && go -C <tree> test -count=1 ./internal/config/` → `ok  github.com/modu-ai/moai-adk/internal/config 15.720s` (non-verbose run, so only the package verdict line was observed: no test of the package failed). Template neutrality: `go -C <tree> test -count=1 -v -run '^TestTemplateNoInternalContentLeak$' ./internal/template` → `--- PASS: TestTemplateNoInternalContentLeak (3.54s)`, `ok  github.com/modu-ai/moai-adk/internal/template 4.049s`.

#### `make build` and the embedded template

`make -C <tree> build` (`/usr/bin/make`, no Xcode-licence refusal) ran its read-only agent-emit check, `templ generate`, `gen-catalog-hashes --all` (printed `catalog.yaml updated successfully`, but `git status --short` afterwards listed no catalog change), then `go build … -o bin/moai`. Templates are embedded straight from `internal/template/templates/` (`//go:embed all:templates`), so no embedded file is regenerated into the tree; the evidence that the build embeds the new key is on the binary: `grep -a -c "max_scan_dirs: 128" bin/moai` → `1`, and the control `grep -a -c "max_scan_dirs: 129" bin/moai` → `0`. `bin/moai` is gitignored (`git check-ignore bin/moai` printed it). `git status --short` after the build listed exactly the six files the key-release commit stages.

#### M2 data point (D-R2): what the accessor returns for non-integral numbers (observation, no assertion, no AC names it)

A throwaway test (written, run once under the lease, then deleted; never committed) wrote `max_scan_dirs: <v>` beside `enabled: true` and `five_hour_hold_pct: 80` and printed the accessor and the settings:

```
OBS max_scan_dirs 2.5: LoadQuotaScanBound=2 LoadQuotaGate.Enabled=true FiveHourHoldPct=80
OBS max_scan_dirs 2.0: LoadQuotaScanBound=2 LoadQuotaGate.Enabled=true FiveHourHoldPct=80
OBS max_scan_dirs 0.5: LoadQuotaScanBound=128 LoadQuotaGate.Enabled=true FiveHourHoldPct=80
OBS max_scan_dirs 1024.5: LoadQuotaScanBound=1024 LoadQuotaGate.Enabled=true FiveHourHoldPct=80
OBS max_scan_dirs 1e3: LoadQuotaScanBound=1000 LoadQuotaGate.Enabled=true FiveHourHoldPct=80
```

Reading: the decoder does not fail on a real number and assigns the integer part (2.5 gives 2, 1024.5 gives 1024, 0.5 gives 0 which the range check turns into 128); the other keys of the block are honoured. This is observed on this tree with this decoder only; it is not a stated guarantee and no test pins it.

#### Other checks

- `go -C <tree> vet ./internal/config` → no output. `golangci-lint v2.1.6 run ./internal/config/...` → `0 issues.` (once on the stub tree and once on the key-release tree). `GOOS=windows GOARCH=amd64 go -C <tree> build ./...` → no output (the plain `go build ./...` ran inside `make build`).
- PRESERVE: `git diff --name-only 10783cb50 -- internal/config/workflow_quota_gate_test.go internal/cli/factory_quota_test.go internal/statusline/quota_test.go` printed nothing; positive control `git diff --name-only 10783cb50 -- internal/config/loader_quota_gate.go internal/config/workflow_quota_gate_test.go` printed `internal/config/loader_quota_gate.go` (a changed file shows, so the empty first result is a measured absence).
- Only `internal/config` Go files, the inventory fixture, the template `workflow.yaml`, and the local twin changed; no `internal/cli` or `internal/statusline` file was touched, so their lint and vet were not run.

#### Gaps

- `gofmt -l internal/config` lists `internal/config/slice.go`, a file this work did not touch (absent from `git status`); not investigated and not part of this card.
- The cli seam (`TestQWR_AC013b_...`, `TestQWR_AC007..009`, `TestQAS_` cli tests) and the statusline tests were not run: M3/M4 own them and this work changed no code they read except through the accessor, which no production caller uses yet. The 2.5 observation is a data point from one decoder on one machine.
- Whether `moai doctor --check "Agent Emit Embed"` / `make embed-check` is clean was not run (no agent definition changed).
- The throwaway observation test was deleted before the commit; its code is not in the tree, only its printed output above.

### M3 + M4 — statusline multi-directory reading (commit `8ea00c4bb17ad018ad35c79207479332b8e390ee`), then the gate seam (commit `45dc4e1012fe988fa8803ffb048625ddb327327e`)

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1442`, branch `WT-quota-read-worktree-records`, started from `f8fac44de`. Judging build for every Go measurement below: `go version go1.26.8 darwin/arm64` (tests are compiled from this tree by `go test`; lint is `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`). RED for these ACs is the M1 record above (assertion-level failures on the behaviour-neutral stubs) and is not repeated; this section records GREEN. Selector outputs were redirected to files and read; the Bash tool reports a non-zero exit as an error, and none was reported for any command below (a silent success is the observed `rc` 0 here).

#### M3 (code: `internal/statusline/quota_dirs.go` bodies, `internal/statusline/quota.go` refactor)

`QuotaStateDirs` returns the primary state directory, then, when `<root>/.git` is a directory with a readable `worktrees` directory, the state directory of each registered worktree: `os.ReadDir` (sorted by name), at most `maxDirs` entries examined whether usable or not, the `gitdir` file read for at most 4096 bytes through `io.ReadFull`, first line only, absolute-path check, `os.Stat` of the named `.git` path (a removed worktree contributes nothing), directory = `<dir of that path>/.moai/state`. `AggregateQuotaDirs` loops `scanQuotaDir` over the directories into one shared pair of `winner`s. The per-directory scan loop was moved out of `AggregateQuota` into `scanQuotaDir` in `quota.go` with its pre-filter lines unchanged in text; `AggregateQuota` is now `return AggregateQuotaDirs([]string{stateDir}, now, maxAge)`.

- Decider (AC-QWR-001..006, -010): `go -C <tree> test ./internal/statusline -run '^TestQWR_' -count=1 -v` → `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/statusline	1.232s`; top-level `--- PASS` lines: 7 (AC001, AC002, AC003, AC004, AC005, AC006, AC010); subtest `--- PASS` lines: 44; `--- FAIL`/`--- SKIP` lines: 0. Verbatim top-level lines: `--- PASS: TestQWR_AC001_WorktreeOnlyRecordIsSeen (0.20s)`, `--- PASS: TestQWR_AC002_FreshestAcrossDirectories (0.04s)`, `--- PASS: TestQWR_AC003_SingleDirectoryUnchanged (0.02s)`, `--- PASS: TestQWR_AC004_UnusableEntriesContributeNothing (0.13s)`, `--- PASS: TestQWR_AC005_EnumerationBound (0.35s)`, `--- PASS: TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive (0.02s)`, `--- PASS: TestQWR_AC010_NoGitMetadataReadsRootOnly (0.01s)`.
- AC-QWR-001: five subtests PASS, including `real_git_worktree_layout` (`--- PASS: .../real_git_worktree_layout (0.17s)`, run on the installed git, not skipped). AC-QWR-002: seven subtests PASS plus the nested `skew` and `age`. AC-QWR-003: `single_dir_equals_multi_dir_of_one`, `signature_unchanged`, `capture_age_independent_of_mtime` PASS. AC-QWR-004: all twelve subtests PASS with none skipped (`gitdir_unreadable` ran, so the run was not root and not Windows). AC-QWR-005: the six subtests PASS. AC-QWR-006: six subtests PASS; swept set = the two non-test `quota*.go` files `quota.go` and `quota_dirs.go`; the positive controls inside the test (13 forbidden selectors flagged, the one planted cross-file call flagged) held, because the test would fail otherwise. AC-QWR-010 (statusline half): `no_git_entry`, `git_is_a_file`, `no_worktrees_dir` PASS.
- Predecessor set: `go -C <tree> test ./internal/statusline -run '^TestQAS_' -count=1 -v` → `ok  	github.com/modu-ai/moai-adk/internal/statusline	0.629s`; 8 top-level `--- PASS` (AC001, AC002, AC002b, AC003, AC004, AC005, AC006, AC014), 43 subtest `--- PASS`, 0 `--- FAIL`/`--- SKIP`. (`TestQAS_AC007_ConfigDefaultsMirrorTemplate` lives in `internal/config`, which M3/M4 did not touch, and was not re-run in this part.)
- Whole statusline package, once: `go -C <tree> test ./internal/statusline -count=1` → `ok  	github.com/modu-ai/moai-adk/internal/statusline	36.741s` (non-verbose: only the package verdict was observed).
- `gofmt -l internal/statusline` printed nothing; `go -C <tree> vet ./internal/statusline` printed nothing.

#### M4 (code: `internal/cli/factory_quota.go`, the production value of the seam `factoryQuotaAggregate`)

The seam keeps its type (`func(string, time.Time, time.Duration) statusline.QuotaAggregate`, asserted at compile time by the new test file and by the unedited predecessor helpers that assign to it). The production value cleans the state-directory path, takes `root = Dir(Dir(path))`, and when `filepath.Join(root, ".moai", "state")` equals the cleaned path calls `factoryQuotaStateDirs(root, config.LoadQuotaScanBound(root))` and `statusline.AggregateQuotaDirs`; any other shape falls through to `statusline.AggregateQuota(stateDir, ...)` (one directory, fail open). `factoryQuotaEvaluate` and its MX anchor are unchanged.

- Check A, is the path-shape derivation safe (plan.md §D2 / U8)? `grep -rn "factoryQuotaAggregate" internal/cli/` shows one production call site, `factory_quota.go:99` (now at a later line), which passes `filepath.Join(root, ".moai", "state")`; the other references are the predecessor test helpers and this SPEC's test file, which replace the variable. Result: safe for the one production caller. The fallback branch (a differently shaped path) is by code only: no test feeds the seam a differently shaped path (Gap below).
- Check B, is the second `workflow.yaml` parse per evaluation acceptable? The two parses read the same 15902-byte file (`wc -c .moai/config/sections/workflow.yaml`) through `loadYAMLFile`, which is a `os.ReadFile` plus `yaml.Unmarshal` with no cache. A throwaway benchmark (written, run once, deleted before the commit; never committed): `go -C <tree> test ./internal/config -run NoSuchTestXyz -bench BenchmarkTmpScanBoundParse -benchtime 200x -count=1` → `BenchmarkTmpScanBoundParse/LoadQuotaGate-16 200 1234646 ns/op`, `BenchmarkTmpScanBoundParse/LoadQuotaScanBound-16 200 2588364 ns/op`, on `Apple M4 Max` with `uptime` at the time `load averages: 49.57 77.76 88.62` (a heavily loaded machine, so these are upper figures, one run each, the two numbers differ by 2x for the same work: noise, not a ratio). Reading: about 1-3 ms per parse, once per evaluation, which runs at most once per `--wait` re-check (5 s) per waiting lane and only while the gate is enabled; the scan it adds is on the order of a thousand file stats. Judged acceptable, so the D2 fallback (a type change of the seam) was not taken and no predecessor file was edited. This is a single noisy measurement, not a benchmark gate.
- Decider (AC-QWR-007, -008, -009, -010b, -013b), under `moai slot acquire --resource go-test-cli --max-duration 20m` (output `slot go-test-cli acquired by 2da35a68-1196-4183-b6e5-a50fc9b6d901 until 2026-10-02T12:09:40Z`), scrubbed single compound form `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go -C <tree> test ./internal/cli -run '^(TestQWR_AC007_GateDisabledReadsNothing|TestQWR_AC008_SurfacesSeeWorktreeReading|TestQWR_AC009_WorktreeSourceRendersIdentically|TestQWR_AC010b_CliRootWithoutGitMetadata|TestQWR_AC013b_ProductionSeamAppliesConfiguredBound)$' -count=1 -v -timeout 30m` → `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/cli	67.759s`; 5 top-level `--- PASS`, 19 subtest `--- PASS`, 0 FAIL/SKIP. Verbatim: `--- PASS: TestQWR_AC007_GateDisabledReadsNothing (2.27s)` (subtests `disabled_zero_calls`, `enabled_one_call_each`, `enumerator_referenced_only_by_the_variable`, `missing_dir_fixture_untouched`), `--- PASS: TestQWR_AC008_SurfacesSeeWorktreeReading (28.47s)` (seven subtests including `next_holds_on_worktree_only_92`, `mcp_next_holds_on_worktree_only_92`, `wait_recheck_sees_worktree_update`), `--- PASS: TestQWR_AC009_WorktreeSourceRendersIdentically (26.14s)` (`status_text`, `status_json`, `hold_line`, `acquire_warning`), `--- PASS: TestQWR_AC010b_CliRootWithoutGitMetadata (8.67s)` (`unregistered_directory_is_not_read`, `roots_own_record_is_read`), `--- PASS: TestQWR_AC013b_ProductionSeamAppliesConfiguredBound (0.89s)` (`bound_3_reads_three`, `key_absent_reads_all_five`).
- Predecessor cli guards (AC-QWR-003 six, AC-QWR-009 two, and the cli `TestQAS_AC014`), same lease, same scrubbed form, one selector naming `TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent`, `TestQAS_AC008_ClaudeLaneHeldAtThreshold`, `TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown`, `TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData`, `TestQAS_AC017_SharedPressureEvaluationAndSurfaces`, `TestQAS_AC022_GateOffOrPressureOffOutputUnchanged`, `TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly` → `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/cli	108.425s`; 7 top-level `--- PASS`, 48 subtest `--- PASS`, 0 FAIL/SKIP. AC-QWR-009's old-behaviour comparison therefore holds: `TestQAS_AC012_...` and `TestQAS_AC022_...` (gate off / pressure off output unchanged, status block) are green unedited, and `TestQWR_AC009_...` shows the worktree-sourced rendering byte-identical to the primary-sourced one on all four surfaces. AC-QWR-010b's old-behaviour comparison: `TestQAS_AC006b_...` (leases when no quota data) green plus `roots_own_record_is_read` and `unregistered_directory_is_not_read` above.
- Lease released before reporting: `moai slot release --resource go-test-cli` → `slot go-test-cli released (was 2da35a68-1196-4183-b6e5-a50fc9b6d901)`.

#### Other checks

- `go -C <tree> vet ./internal/statusline ./internal/cli ./internal/config` → no output. `gofmt -l` on the three changed files → no output. `go -C <tree> run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./internal/statusline/... ./internal/cli/... ./internal/config/...` → `0 issues.`. `GOOS=windows GOARCH=amd64 go -C <tree> build ./...` → no output.
- PRESERVE: `git -C <tree> diff --name-only f8fac44de -- internal/statusline/quota_test.go internal/statusline/state_anchor.go internal/statusline/context_usage.go internal/cli/factory_quota_test.go internal/config/workflow_quota_gate_test.go internal/stateanchor` → no output; positive control, the same form with `internal/statusline/quota.go internal/statusline/quota_dirs.go internal/cli/factory_quota.go` added printed exactly those three paths.
- E4: `grep -n "AskUserQuestion" internal/statusline/quota.go internal/statusline/quota_dirs.go internal/cli/factory_quota.go` → no output (no match).
- New non-test code is `os`, `io`, `bytes`, `strings`, `path/filepath` only: no `net`, no `os/exec`, no `filepath.Walk*`, no mutating `os.*` selector (the AC-QWR-006 sweep passed on the committed tree).

#### Gaps

- The fallback branch of the production seam (a state-directory path that is not `<root>/.moai/state`) is not pinned by any test: AC-QWR-013b and AC-QWR-007 call the seam through `factoryQuotaEvaluate` only, which always passes the conventional shape. It is a four-line branch read in the diff.
- `real_git_worktree_layout` ran on the installed git of this machine only; Windows `gitdir` content (`C:/...`) was not exercised (plan.md U2); `TestQWR_AC004/gitdir_unreadable` ran as non-root on macOS, so the Windows and root skips of that subtest were not exercised.
- The second-parse cost in Check B is one noisy sample on a machine at load average about 50-90; the 128-directory default bound and the per-call cost of the whole multi-directory reading remain unmeasured (the M5 benchmark data point is not yet taken).
- The whole `./internal/cli` and `./internal/config` packages were not run (only the anchored selectors above); the shipped-key guard and `TestQAS_AC007_ConfigDefaultsMirrorTemplate` were not re-run because `internal/config` was not touched in this part.
- The end-to-end observation of a real lane's gate reading real worktree records is still the post-install follow-up of plan.md §C (no binary built from this tree was installed or run).

### M5 — closure evidence (AC-QWR-011, AC-QWR-012, benchmark data point, closure checks, final-tree decider re-run)

Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1442`, branch `WT-quota-read-worktree-records`, HEAD at measurement `ce2cd365ca96f61c9b3cd4c3fd4082d546ce7610`, 2026-10-02, tree clean (`git status --short` printed nothing). No production code and no plan artifact was changed in this milestone. Judging build for every Go measurement: `go version go1.26.8 darwin/arm64` (tests compile from this tree; lint is `golangci-lint v2.1.6` through `go run ...@v2.1.6`). The `moai` command used for lint is the installed build `v3.2.0-rc.25 moai_cp/20260925_122548-1952-g802a72235 built 2026-10-02T08:00:14Z`, an ancestor of this tree HEAD that predates all the quota code (QWR-M0 above). The Bash tool displays a failure line for a non-zero exit and none for exit 0 (observed: `Exit code 3` for the refused slot acquire below); a command that printed "no output" and showed no failure line is recorded as exit 0 by that display, not by an echoed `$?` (the guard refuses `; echo $?`).

#### AC-QWR-011 — baseline-first ordering (commands as written in acceptance.md, no pipes)

```
$ git merge-base develop HEAD
284e09c44023598affe486f17701717ca173e6ca            (CARD_BASE)
$ git log --reverse -S'QWR-M0' --format=%H 284e09c44023598affe486f17701717ca173e6ca..HEAD -- .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/progress.md
60cb309ba8e6c2bce563ebd8ccdeaba3bd057123            (first line = B)
7ff152b2bc9ab432b9f134df1dd6166968187f35            (a later commit whose M1 RED text also adds the marker: the pinned "not the first to carry it" mutant; B is the FIRST line)
$ git log --reverse -S'QWR-M0' --format=%H 284e09c44023598affe486f17701717ca173e6ca..HEAD -- .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/plan.md     (positive control, non-empty)
8da5e1f39db1317f970a1e25831f22d8437f6f01
b92f4bd8f25067a5d00865b9482bc48c2e0fd402
bb6b9925ecf50b69ab9f54d0aa3e548bd41016e0
$ git log --reverse --format=%H 284e09c44023598affe486f17701717ca173e6ca..HEAD -- internal ':(exclude)*_test.go' ':(exclude)*/testdata/*'
0bb2048beb0d223ba2edfec65d3cafb1c900c9ef            (first line = I)
e350a81392c19bcd0555ec0ec75b64b0866d5c77
7ff94b33f79ec389a8f226cde4fbe018c0f7c68b
8ea00c4bb17ad018ad35c79207479332b8e390ee
45dc4e1012fe988fa8803ffb048625ddb327327e
$ git show 60cb309ba8e6c2bce563ebd8ccdeaba3bd057123:.moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/progress.md
(read in full, not piped: the `### QWR-M0 — real-lane record-location measurement` heading, the before/after listing of the four record copies with mtimes, dates and times T0/T1/T2, the measured command `moai factory status`, statements (i)-(iii), and the gaps)
$ git merge-base --is-ancestor 60cb309ba8e6c2bce563ebd8ccdeaba3bd057123 0bb2048beb0d223ba2edfec65d3cafb1c900c9ef
(no output, no failure line displayed: exit 0)
$ git rev-list --count 0bb2048beb0d223ba2edfec65d3cafb1c900c9ef..60cb309ba8e6c2bce563ebd8ccdeaba3bd057123
0
$ git rev-list --count 60cb309ba8e6c2bce563ebd8ccdeaba3bd057123..0bb2048beb0d223ba2edfec65d3cafb1c900c9ef      (strictness)
1
$ git show --stat --format=%h 0bb2048beb0d223ba2edfec65d3cafb1c900c9ef
(8 files: factory_quota.go, factory_quota_worktrees_test.go, loader_quota_gate.go, quota_gate_scan_dirs_test.go, types.go, quota_dirs.go, quota_dirs_test.go, and spec.md frontmatter; 1697 insertions, 1 deletion)
```

Reading: B (`60cb309ba`) exists, is none of the plan-phase commits (`8da5e1f39`, `2c6ef5077` carry no marker in progress.md; the progress.md form printed no plan commit), carries the QWR-M0 cell, and is a strict ancestor of I (`0bb2048be`): the ancestor command exited 0, `I..B` prints 0, `B..I` prints 1 (at least 1). The same-commit mutant (B..I = 0), the marker-after-implementation mutant (I..B at least 1) and the not-first-carrier mutant (B is the first line of the list) are all excluded by these outputs. AC-QWR-011: PASS. Pre-merge evaluation only (after the card merges the range is empty).

#### AC-QWR-012 — preserve, the writer, the other thresholds, the schema (commands as written, no pipes)

```
$ git diff --name-only 284e09c44023598affe486f17701717ca173e6ca..HEAD -- internal/stateanchor internal/statusline/state_anchor.go internal/statusline/context_usage.go .moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001 internal/statusline/quota_test.go internal/cli/factory_quota_test.go internal/config/workflow_quota_gate_test.go
(no output: no preserved path changed)
$ git diff --name-only 284e09c44023598affe486f17701717ca173e6ca..HEAD -- internal/config
internal/config/cache.go
internal/config/defaults.go
internal/config/loader_quota_gate.go
internal/config/quota_gate_scan_dirs_test.go
internal/config/testdata/shipped_key_inventory.yaml
internal/config/types.go
$ git diff --numstat 284e09c44023598affe486f17701717ca173e6ca..HEAD -- internal/config/defaults.go internal/config/loader_quota_gate.go
7	0	internal/config/defaults.go
27	0	internal/config/loader_quota_gate.go
$ git diff --name-only 284e09c44023598affe486f17701717ca173e6ca..HEAD      (positive control, non-empty)
(18 paths; names both internal/statusline/quota_dirs.go and internal/config/types.go, plus internal/statusline/quota.go, internal/cli/factory_quota.go, the template and local workflow.yaml, and the SPEC directory)
```

Reading: (1) the preserve list printed nothing, with the control non-empty and naming both required paths, so the empty result is a measured absence. (2) The six config paths are exactly the five allowed files plus the new test file `quota_gate_scan_dirs_test.go`; no other `internal/config` path changed. (3) The two numstat lines have removed-lines column 0 and added-lines column at least 1. The additions (`git diff -U1` read): `DefaultQuotaGateMaxScanDirs = 128` and the seed line `MaxScanDirs: DefaultQuotaGateMaxScanDirs` in `defaults.go`; the range constants `quotaGateScanDirsMin = 1` / `quotaGateScanDirsMax = 1024` and the accessor `LoadQuotaScanBound` in `loader_quota_gate.go`. The five predecessor quota_gate keys (`enabled`, `five_hour_hold_pct`, `seven_day_hold_pct`, `release_margin_pct`, `max_age`), their defaults (off, 90, 95, 5, 30m) and ranges are untouched: no `DefaultQuotaGate*` constant, no `quotaGate*` range constant, no `QuotaGateSettings`, `DefaultQuotaGate`, `resolveQuotaGate` or `LoadQuotaGate` line appears as a removed or changed line. AC-QWR-012: PASS. Pre-merge evaluation only.

#### Timing data point (D4, NOT a gate): `BenchmarkQWR_RealRoot`, one run

Command (the env switch `MOAI_QWR_BENCH_ROOT` is read from the test file constant `qwrEnvBenchRoot`, set in the same command): `env MOAI_QWR_BENCH_ROOT=/Users/goos/MoAI/moai-adk-go go -C <tree> test ./internal/statusline -run '^$' -bench 'BenchmarkQWR_RealRoot' -benchtime 20x -count=1`. Nothing else of mine ran during it.

```
uptime before: 20:57  up 16 days,  9:09, 33 users, load averages: 78.41 104.36 99.46
cpu: Apple M4 Max
BenchmarkQWR_RealRoot/multi_dir_bound_128-16      20   29208033 ns/op
BenchmarkQWR_RealRoot/single_primary_dir-16       20    7477219 ns/op
PASS
ok  	github.com/modu-ai/moai-adk/internal/statusline	1.253s
uptime after:  20:58  up 16 days,  9:09, 33 users, load averages: 82.30 104.34 99.50
```

(The first "after" reading, 20:57 / 80.06, was issued in the same tool batch as the benchmark and may have run alongside it, so the clean one above was taken afterwards.) At the time the primary repository had 38 entries under `.git/worktrees` (`ls` of that directory, 40 lines less the two dot entries), so the bound of 128 was not reached. Reading: one run, 20 iterations, on a machine at load average about 80 (an M4 Max whose 16 cores were oversubscribed by other sessions): about 29.2 ms per multi-directory call against about 7.5 ms per single-directory call on the primary directory, a ratio of about 3.9. This is a single noisy data point on a loaded machine. It is not a pass or a fail, not a gate, and not a basis for the 128 default or for any cost claim; the M4 Check B figures (a throwaway `workflow.yaml` parse benchmark at load about 50-90) are equally one noisy sample.

#### Closure checks (scoped), final tree

- `go -C <tree> vet ./internal/statusline ./internal/config ./internal/cli` → no output, no failure line (exit 0).
- `GOOS=windows GOARCH=amd64 go -C <tree> build ./...` → no output, no failure line (exit 0).
- `go -C <tree> run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./internal/statusline/... ./internal/config/... ./internal/cli/...` → `0 issues.`
- `moai spec lint .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/spec.md` and the same with `--strict` (installed build named above) → both `0 error(s), 2 warning(s)`, no failure line. Both warnings are `MovingRefUnpinned` on `progress.md` lines 232 and 243, inside the M1 RED section: they sit on verbatim captured test output (a quoted stdout line that mentions the text `origin/main`), not on a claim of this SPEC; those lines are earlier record and are not edited here. Under `--strict` the warnings did not become errors. This is a gap in the sense that the lint is not warning-free; it is not an error.
- `go -C <tree> test ./internal/spec -run 'TestACCounterFullCorpusMatchesBaseline' -count=1 -v` → `--- PASS: TestACCounterFullCorpusMatchesBaseline (12.56s)`, `ok  github.com/modu-ai/moai-adk/internal/spec 12.965s`; the test reports 37 acceptance files absent from the snapshot "reported, not failed", among them this SPEC's `acceptance.md` with `COUNT 15` (the file has 13 `### AC-QWR-` headings by `grep -c`; the counter's 15 was not reconciled, and the test treats absent-from-snapshot files as non-failing).
- E4: `grep -rn AskUserQuestion` over the changed non-test files (`internal/statusline/quota.go`, `internal/statusline/quota_dirs.go`, `internal/cli/factory_quota.go`, `internal/config/loader_quota_gate.go`, `defaults.go`, `types.go`, `cache.go`) → no output (no match; the harness does not display a failure line for a zero-match grep, so this is the printed absence).
- New non-test code uses the standard library only (`os`, `io`, `bytes`, `strings`, `path/filepath`); the AC-QWR-006 sweep inside the statusline run below passed.

#### Final-tree decider re-run (HEAD `ce2cd365c`, scrubbed single compound `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go -C <tree> test ...`, `-count=1 -v`, output redirected to a file and read)

Statusline `-run '^(TestQWR_|TestQAS_)'` (no lease; light) → `ok  github.com/modu-ai/moai-adk/internal/statusline 1.423s`; top-level `--- PASS`: 15 (seven TestQWR_: AC001, AC002, AC003, AC004, AC005, AC006, AC010; eight TestQAS_: AC001, AC002, AC002b, AC003, AC004, AC005, AC006, AC014), subtests `--- PASS`: 77, `--- FAIL`/`--- SKIP`: 0.

Coverage of the changed statusline code under that same selector (`-coverprofile`, `go tool cover -func`): `quota.go` `AggregateQuota` 100.0%, `scanQuotaDir` 94.4%, `offer` 100.0%, `reading` 100.0%; `quota_dirs.go` `QuotaStateDirs` 100.0%, `worktreeStateDir` 93.8%, `AggregateQuotaDirs` 100.0%. (The package total of that run, 32.1%, is the selector's reach over the whole package and is not a coverage claim for the package.)

Lease: `moai slot acquire --resource go-test-cli --max-duration 20m` first printed nothing and exited 3 (`Exit code 3`); `moai slot status --resource go-test-cli` showed another session (`60da3105-d660-42ae-a371-4ceb8a76c327`, pid 40942) holding it until `2026-10-02T12:09:22Z`. A retry loop acquired it after that bound passed, full output read: `slot go-test-cli acquired by 2da35a68-1196-4183-b6e5-a50fc9b6d901 until 2026-10-02T12:29:37Z` / `  displaced: 60da3105-d660-42ae-a371-4ceb8a76c327 (pid 40942), reason expired, held since 2026-10-02T11:54:22Z` (the previous holder's declared bound had expired; no `--force` was used).

Config `-run '^(TestQWR_AC013_MaxScanDirsConfigKey|TestQAS_AC007_ConfigDefaultsMirrorTemplate|TestShippedConfigKeysHaveReaders)$'` → `--- PASS: TestQWR_AC013_MaxScanDirsConfigKey (0.03s)`, `--- PASS: TestShippedConfigKeysHaveReaders (4.62s)`, `--- PASS: TestQAS_AC007_ConfigDefaultsMirrorTemplate (0.03s)`, `ok  github.com/modu-ai/moai-adk/internal/config 5.040s`; 3 top-level, 21 subtests PASS (8 + 4 + 9), 0 FAIL/SKIP.

Cli `-run '^(TestQWR_.*|TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent|TestQAS_AC008_ClaudeLaneHeldAtThreshold|TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown|TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData|TestQAS_AC017_SharedPressureEvaluationAndSurfaces|TestQAS_AC022_GateOffOrPressureOffOutputUnchanged|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$' -timeout 30m` → `ok  github.com/modu-ai/moai-adk/internal/cli 173.540s`; 12 top-level `--- PASS` (five TestQWR_: AC007 `2.13s`, AC008 `16.38s`, AC009 `21.60s`, AC010b `7.01s`, AC013b `0.62s`; seven TestQAS_: AC006b, AC008, AC010, AC012, AC014, AC017, AC022), 55 subtests PASS, 0 FAIL/SKIP.

Lease released before this record was written: `moai slot release --resource go-test-cli` → `slot go-test-cli released (was 2da35a68-1196-4183-b6e5-a50fc9b6d901)`.

AC matrix on the final tree: AC-QWR-001..006 and -010 (statusline) PASS; -007, -008, -009, -010b (cli), -013b PASS; -013 (config) PASS; -011 and -012 PASS by the one-shot commands above. 13 of 13 PASS, none skipped.

#### Gaps (M5)

- The end-to-end observation of the gate reading a real lane's records (an installed build from a tree containing both SPECs, a lane under a worktree, the freshest record in a worktree directory) was not made: no binary built from this tree was installed or run. Post-install follow-up, not an acceptance criterion.
- The 128 default, the 4 KiB `gitdir` read and the per-call cost are unmeasured as design values; the one benchmark above is a single noisy sample at load about 80, not a measurement of cost.
- The fallback branch of the production seam (a state-directory path that is not `<root>/.moai/state`) is pinned by no test (by code only).
- `real_git_worktree_layout` ran on this machine's git only; the Windows `gitdir` shape (`C:/...`) was not exercised, and the Windows and root skips of `TestQWR_AC004/gitdir_unreadable` were not exercised.
- The whole `./internal/cli` and `./internal/config` packages were not run (anchored selectors only, per the load discipline); CI on the pushed develop head carries the full-suite verdict.
- `moai spec lint` shows the two `MovingRefUnpinned` warnings noted above (not errors); the AC counter reports 15 for acceptance.md against 13 headings.
- Exit status of commands that printed nothing is inferred from the harness's failure-line display, not from an echoed `$?`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-02
run_status: audit-ready
cycle_type: tdd
head_at_signal: ce2cd365ca96f61c9b3cd4c3fd4082d546ce7610   # the M5 evidence commit follows it
commits:
  m0_baseline_B: 60cb309ba8e6c2bce563ebd8ccdeaba3bd057123
  first_implementation_I: 0bb2048beb0d223ba2edfec65d3cafb1c900c9ef   # M1 stubs and RED tests, flips spec.md to in-progress
  m1_rework: e350a81392c19bcd0555ec0ec75b64b0866d5c77
  m2: 7ff94b33f79ec389a8f226cde4fbe018c0f7c68b
  m3: 8ea00c4bb17ad018ad35c79207479332b8e390ee
  m4: 45dc4e1012fe988fa8803ffb048625ddb327327e
  m5_evidence: 80d491599   # a commit cannot cite its own hash; the real SHA is in the run report
acceptance_criteria: {total: 13, pass: 13, skipped: 0}
open_gaps:
  - end-to-end real-lane gate observation after installing a build from this tree (post-install follow-up, not an AC)
  - 128 default, 4 KiB gitdir read and per-call cost unmeasured (one noisy benchmark sample only)
  - fallback path of the production seam unpinned by any test
  - Windows gitdir shape unexercised
  - moai spec lint: 2 MovingRefUnpinned warnings on quoted M1 output (progress.md lines 232 and 243), 0 errors
minor_plan_debt:
  absorbed_in_run: [D-N9 (the cli test counters delegate to the values in force), D-R1 (the accessor is only reached through the seam, which AC-QWR-007 shows is not called while the gate is off), D-R2 (M2 recorded the 2.5 observation), D-R6 (Go seed and template key landed in one commit 7ff94b33f)]
  remaining_plan_text_for_sync_or_manager_spec: [D-N10 (the benchmark env name MOAI_QWR_BENCH_ROOT is a test-file constant; acceptance.md section B still says no test reads the real tree), D-N11, D-N12, D-N13 (no --auto steering subtest exists), D-N14, D-N15 (measured: 1846 insertions and 13 deletions under internal/ against the Tier M 300-1000 LOC guidance, 1663 of the insertions in three test files (615 + 205 + 843)), D-R3, D-R4, D-R5, D-R7]
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-02
sync_status: audit-ready
sync_commit_sha: pending-backfill   # a commit cannot cite its own hash; backfilled in the following commit
changelog_path: CHANGELOG.md   # [Unreleased] ### Added, first entry
ac_source: .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/acceptance.md   # tier M
ac_count: 13   # live "### AC-QWR-" headings in acceptance.md
open_gaps:
  - end-to-end real-lane gate observation after installing a build from this tree (post-install follow-up, not an AC)
  - 128 default, 4 KiB gitdir read and per-call cost unmeasured (one noisy benchmark sample only)
  - fallback path of the production seam unpinned by any test
  - Windows gitdir shape unexercised
  - plan debt D-N10..D-N15 and D-R3, D-R4, D-R5, D-R7 not absorbed in the run phase
  - 1846 inserted lines under internal/ exceed the Tier M LOC guidance (1663 of them in test files)
```

## §F Phase 4 Mode Selection

### Plan→run Kickoff gate (autonomous form, auto-semantics §9.1)

| Criterion | Observed | Result |
|-----------|----------|--------|
| Independent plan-audit verdict | PASS 0.87 (iteration 3 of max 3; iteration 1 FAIL 0.75, iteration 2 FAIL 0.83) — `.moai/reports/t1442/plan-audit-iter3.md` | met |
| Score vs Tier M threshold | 0.87 >= 0.80 | met |
| Plan phase audit-ready | §E.1 `plan_status: audit-ready` | met |
| Plan-artifact hashes unchanged since the verdict | audited_sha bb6b9925ecf50b69ab9f54d0aa3e548bd41016e0; `git diff --stat bb6b9925e HEAD` empty at 2026-10-02T10:49:49Z; sha256 spec.md e131417655ea253de2b1a016935ce2a2a309cd676e9bc23b64b0cae87f73536f, plan.md 6664717c142873d0af579265f2495c897b9abd7aaa9ccde3d7ffcd8679f6619c, acceptance.md 1910525753c3854f9b11199fac23c182b62a3cf65acba241f45bf792b282c8f2 | met |
| Open blocker | none (iteration 3 findings are minor only: D-N9..D-N15) | met |
| Keep-set case (environment-impossible, operator-held, irreversible external-shared) | none; local-tree work, no push, no PR | not applicable |

decision record: decided_by=claude-code lane-9 orchestrator (factory lane, Kickoff autonomous transition) evidence_refs=.moai/reports/t1442/plan-audit-iter3.md(verdict=PASS score=0.87 audited_sha=bb6b9925e),plan-artifact-hashes(unchanged) ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1)
recorded: 2026-10-02T10:49:49Z (the decision board is not writable from a lane session; this progress record is the durable copy)

### Mode selection

- Decision: serial
- Inputs: tier M; scope about 6 Go files plus tests across internal/statusline, internal/config, internal/cli; 1 domain family (quota gate); mixed Go source and YAML/markdown; concurrency benefit LOW (coding-heavy).
- Mode evaluation: direct not selected (non-trivial multi-file change); serial selected (default for coding-heavy work, one leaf worker per milestone); fanout not selected (coding-heavy, not research); sweep not selected (not a uniform mechanical transform).
- Boundary case: none.

### Known plan debt handed to the run phase (iteration 3 minor findings, not blockers)

D-N9..D-N15 are minor wording and cross-reference items listed in `.moai/reports/t1442/plan-audit-iter3.md`; run workers read that report and must not edit plan artifacts (a body change goes back to manager-spec through the orchestrator).

## §J Decision Log

| # | Decision | Source | Confidence / verdict |
|---|----------|--------|----------------------|
| Q1 / D1 | Read the primary directory plus every linked worktree enumerated from `<git-common-dir>/worktrees/*/gitdir` | decision oracle (Jev) | option A, confidence 1.00 |
| Q2 / D4 | At most 128 directories in name order, gitdir read up to 4 KiB, all unmeasured; bound exposed as `workflow.quota_gate.max_scan_dirs` | oracle leaned 64/newest-first at confidence 0.22 (under the 0.5 gate); leader verdict | leader accepted the plan default provisionally and required the config key |
| Q3 / D5 | No output change; predecessor goldens and tests untouched | decision oracle (Jev) | confidence 0.97 |
| Q4 / D6 | Measure around read-only `moai factory status` always; around `moai factory next` only when a lease is genuinely due anyway | decision oracle (Jev) | confidence 0.83 |
| Q5 / D2 | Keep the seam type and derive the root from the path shape; if that proves unsafe, switch to a type change and report to the leader | oracle chose keep at confidence 0.42 (under the 0.5 gate); leader verdict | leader accepted provisionally |

### Run-phase decision R1 (leader verdict, 2026-10-02, found at M1)

Conflict: AC-QWR-013 (REQ-QWR-011) needs `DefaultQuotaGate()` to carry `MaxScanDirs` 128, while four subtests of the predecessor's `TestQAS_AC007_ConfigDefaultsMirrorTemplate` compare the whole `QuotaGateSettings` struct to literals without it, and REQ-QWR-010 / AC-QWR-003 / AC-QWR-012 forbid editing that file. Options offered: A separate accessor outside `QuotaGateSettings`; B approved edit of the four predecessor subtests; C struct default 0 with a read-site default.

Verdict (leader, via cross-session message, not an operator gate): **A**. Predecessor tests stay unedited; the scan bound is read through a separate accessor outside `QuotaGateSettings`; key `workflow.quota_gate.max_scan_dirs`, default 128, range 1-1024 unchanged. Audit: the 3-iteration ceiling is spent, and the leader approves exactly ONE delta plan-audit, scoped to the revised REQ-QWR-011, AC-QWR-013, the plan.md §G U13 correction, and the "non-integer value" wording correction (a real number such as 2.5 does not fail the decode; a string and an overflowing integer literal do, U21 observed). Not a full re-audit. A FAIL delta gets no further extension: report to the leader. Sequence: manager-spec revises spec/plan/acceptance (acceptance revision carries the AC snapshot in the same commit), plan-auditor delta, then M2 on PASS family. Verdict file under `.moai/reports/t1442/`, decision recorded here.

R1 applied: manager-spec commit `10783cb50` (spec.md 0.6.0, plan.md, acceptance.md, decision-index.md; REQ 12 and AC 13 unchanged). Delta plan-audit (the one leader-approved extension, scope limited to the R1 revision): **PASS 0.86** against Tier M 0.80, auditor model claude-sonnet-5-5, audited_sha `10783cb50d1ffc122d2839adba7b22181c8e9337`, report `.moai/reports/t1442/plan-audit-delta1.md` (local-only). Plan-artifact sha256 now: spec.md 015949c78c739c661015db0d3bfd8210d23dbc3c175d4bd202d534ff4edb60b4, plan.md 30c60916c358f2559ca1c9d0a44c81efc83949ba8c49d89779bf850616fe92d6, acceptance.md bcb04ec37f4110b576ebf9c7ad5db5de987fe31236b1e6e136b187466a3a336e (recomputed by the lane at HEAD `10783cb50`, equal to the auditor's). No blocking defect; seven minor findings D-R1..D-R7 handed to the run phase as known plan debt (run workers must not edit plan artifacts): D-R1 the second parse happens only when the gate is enabled; D-R2 the 2.5 claim rests on a removed draft test, so M2 observes it; D-R3 plan.md line 89 credits `7ff152b2b` with the stubs but that commit is progress.md-only; D-R4 REQ-QWR-011 is a very long sentence; D-R5 AC-013 RED-now has no accessor-absence proxy; **D-R6 the Go seed and the template key must land in ONE commit or defaults_equal_template goes red in between**; D-R7 a stale "rest of that file was not read" hedge in acceptance.md. The Kickoff gate (§F) stands on the iteration-3 PASS plus this delta PASS.

Plan-audit history: iteration 1 FAIL 0.75, iteration 2 FAIL 0.83, iteration 3 PASS 0.87, delta 1 (after R1) PASS 0.86. Defect lists were relayed to manager-spec with the decisions applied each time (commits b92f4bd8f, bb6b9925e).
