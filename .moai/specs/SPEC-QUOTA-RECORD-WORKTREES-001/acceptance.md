# acceptance.md — SPEC-QUOTA-RECORD-WORKTREES-001 (card t1442)

This file is the verification layer: Given-When-Then criteria, each decided by one named command. The requirement layer (GEARS) is `spec.md` §C. Test files named here are plan choices; the run phase may rename a file but not an AC's test name.

## §A Scope of verification

The criteria verify: a reading that exists only in a linked worktree's record directory is seen (three layouts plus a metadata entry whose name differs from its directory, and the on-disk layout git itself writes); freshest-wins across directories (a stale primary copy of a session id against a fresh worktree copy, the reverse, per window, tie, not-max); the single-directory entry point and every old `TestQAS_` test unchanged; unusable worktree entries contributing nothing; the entry bound and its configuration key `workflow.quota_gate.max_scan_dirs` (default, range, template, local twin, inventory, cache schema, and the production seam applying it); the offline, spawn-free, read-only, non-recursive property of the new code; nothing read while the gate is disabled; the lane gate, status block, and acquire warning seeing a worktree-only reading with unchanged output form; the no-git-metadata root reading its own directory alone; the measurement record and its commit ordering; and the untouched writer, thresholds, schema, and predecessor files. Not verified here: a real Claude Code statusline writing windows (the installed binary writes schema 2, no windows), real quota consumption, kanban mode, and any codex or GLM lane.

## §B Test-environment constraint (binds every AC)

- Every test builds its own roots with `t.TempDir()` (a primary root, and worktree directories under the same or a second temporary directory); the MoAI home is redirected to a temporary directory. No test reads or writes the real `~/.moai` or a real `.moai/state`. Metadata entries are written by the test as `<root>/.git/worktrees/<entry>/gitdir` files, except the one subtest that runs the installed `git` in a temporary repository (`real_git_worktree_layout`, which clears `GIT_*` environment variables with `t.Setenv`, skips with a named reason only when `git` is absent from `PATH`, and is a Gap in that case).
- Time is an injected clock (`factoryCardNow`, `factoryNextWaitSleep` for the factory seam; explicit `now` arguments for the aggregator). Record files are fixtures with `os.Chtimes` set per case; no test sleeps or reads the wall clock for a verdict.
- `./internal/cli` runs only with an anchored `-run` naming the AC tests; a whole-package run is not a verdict (it measures the machine, `gitflow-lane-protocol.md` §8). `./internal/statusline` may run whole-package once at the end. Heavy runs take a `moai slot` lease.
- Pass convention: an AC passes only when every command it names exits 0 and each test's verbose output carries a `--- PASS: <name> ` line for exactly the test the command names (name then a space). `[no tests to run]`, `no tests ran`, or a zero swept count is a Gap, never a pass (`verification-completeness.md` §1.1); a skipped subtest is a Gap named in the AC, not a pass. At M1 each anchored selector is run once against the pre-implementation tree and must print `[no tests to run]` — the intended RED of a test that does not exist yet.
- Defaults the tests assert come from the predecessor: hold 90 / 95, margin 5, max age 30m, clock-skew tolerance 5m; this SPEC adds one configuration key, `workflow.quota_gate.max_scan_dirs` (default 128, range 1-1024, unmeasured; AC-QWR-013), and one compiled constant, the 4 KiB `gitdir` read bound.

## §C Traceability

| REQ | AC |
|---|---|
| REQ-QWR-001 | AC-QWR-001, AC-QWR-002, AC-QWR-003 |
| REQ-QWR-002 | AC-QWR-001 |
| REQ-QWR-003 | AC-QWR-004 |
| REQ-QWR-004 | AC-QWR-005, AC-QWR-013 |
| REQ-QWR-005 | AC-QWR-006 |
| REQ-QWR-006 | AC-QWR-007 |
| REQ-QWR-007 | AC-QWR-008, AC-QWR-009 |
| REQ-QWR-008 | AC-QWR-010 |
| REQ-QWR-009 | AC-QWR-011 |
| REQ-QWR-010 | AC-QWR-003, AC-QWR-012 |
| REQ-QWR-011 | AC-QWR-013 |

## §D Acceptance criteria

Each entry ends with its **RED-now** cell (ledger row id in §E) and its **green path** cell (the milestone that flips it and what the passing output becomes).

### AC-QWR-001 — a reading that exists only in a linked worktree's directory is seen

- **Given** a temporary primary root whose own `context-usage` directory holds no fresh record, git metadata entries `<root>/.git/worktrees/<entry>/gitdir`, and one fresh record (five-hour 62.5%, reset in the future) in the worktree directory the entry names, **when** the evaluation's multi-directory reading runs at a fixed `now`, **then** the five-hour window reads fresh at 62.5%, in each layout: `moai_worktrees_layout` (`<root>/.moai/worktrees/<name>`), `claude_worktrees_layout` (`<root>/.claude/worktrees/<name>`), `outside_root_layout` (the worktree directory under a second temporary directory), `entry_name_differs_from_directory` (entry `a1` whose `gitdir` names directory `b1`), and `real_git_worktree_layout` (a temporary repository, a worktree made by the installed `git worktree add`, a fresh record in that worktree's directory, found through the real metadata git wrote).
- Decider: `go test ./internal/statusline -run '^TestQWR_AC001_WorktreeOnlyRecordIsSeen$' -count=1 -v` → `--- PASS: TestQWR_AC001_WorktreeOnlyRecordIsSeen ` and the five subtests listed.
- RED-now: ledger E1, E2. Green path: M2 (the enumerator and the multi-directory reading).

### AC-QWR-002 — freshest wins across directories (not max, not min)

- **Given** a primary directory and one worktree directory, **when** the multi-directory reading runs at `now`, **then**: `stale_primary_copy_fresh_worktree_copy` (the same session id; primary copy captured now-20m at 95%, worktree copy captured now-2m at 60%) reads 60%; `newer_primary_copy_wins` (the reverse placement: primary now-2m at 60%, worktree now-20m at 95%) reads 60%; `not_max_across_dirs` (older record in one directory at 97%, newer in the other at 70%) reads 70%; `not_min_across_dirs` (older at 40%, newer at 80%) reads 80%; `per_window_across_dirs` (a newer record carrying only the seven-day window in one directory does not hide an older fresh record's five-hour window in the other); `tie_keeps_primary` (equal capture time, different percentages: the primary directory's record wins because it is listed first); `skew_and_age_apply_to_worktree_records` (a worktree record captured 5m1s ahead of `now`, or 30m1s behind, contributes nothing).
- Decider: `go test ./internal/statusline -run '^TestQWR_AC002_FreshestAcrossDirectories$' -count=1 -v` → PASS with the seven subtests listed. Mutants pinned: first-directory-wins, last-directory-wins, maximum over directories, minimum over directories, a worktree record exempt from the skew or age rule.
- RED-now: ledger E1. Green path: M2.

### AC-QWR-003 — the single-directory entry point and the old tests are unchanged

- **Given** the existing single-directory reading and the old test suites, **when** the multi-directory reading is added by refactoring the shared scan loop, **then** `single_dir_equals_multi_dir_of_one` (on the fixtures of the predecessor's AC-QAS-005 shape, `AggregateQuota(dir, now, age)` equals the multi-directory reading of that one directory); `signature_unchanged` (a compile-time assignment of the single-directory function to `func(string, time.Time, time.Duration) QuotaAggregate` and of the cli seam variable to `func(string, time.Time, time.Duration) statusline.QuotaAggregate`); `capture_age_independent_of_mtime` (file mtime at `now`, `CapturedAt` 30m1s earlier → unknown; pins the line the refactor moves, audit finding F3 of the predecessor); and every old `TestQAS_` decider stays green with no edit to its file.
- Decider: `go test ./internal/statusline -run '^TestQWR_AC003_SingleDirectoryUnchanged$' -count=1 -v` → PASS (three subtests); `go test ./internal/statusline ./internal/config -run '^(TestQAS_AC001_RecordCarriesSuppliedWindowsOnly|TestQAS_AC002_PreviousSchemaReadsAsNoWindows|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline|TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop|TestQAS_AC004_ExhaustedAtStickyUntilRollover|TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover|TestQAS_AC006_FailOpenOnAbsentOrUnreadable|TestQAS_AC007_ConfigDefaultsMirrorTemplate|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$' -count=1 -v` → nine top-level `--- PASS: <name> ` lines (the predecessor's audit recorded 9 top-level and 44 subtest lines for these two packages), zero FAIL, zero SKIP; `go test ./internal/cli -run '^(TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent|TestQAS_AC008_ClaudeLaneHeldAtThreshold|TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown|TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData|TestQAS_AC017_SharedPressureEvaluationAndSurfaces|TestQAS_AC022_GateOffOrPressureOffOutputUnchanged)$' -count=1 -v` (scrubbed anchored form, under a `moai slot` lease) → six PASS lines.
- RED-now: ledger E1 for the new test; the old tests are regression guards by nature (green at arrival). Green path: M2.

### AC-QWR-004 — an unusable worktree entry contributes nothing and breaks nothing

- **Given** a fixture holding one valid entry with a fresh record (the positive control, read in every subtest) plus one broken entry, **when** the multi-directory reading runs, **then** the valid entry's reading is returned, with no panic and no error, for each broken entry: `gitdir_unreadable` (mode 0; skipped with a named reason on Windows and when running as root — a Gap), `gitdir_empty`, `gitdir_garbage_not_a_path`, `gitdir_relative_path`, `gitdir_oversized` (a 1 MiB file; only the first 4 KiB is read), `target_missing_pruned` (the named `.git` path no longer exists), `record_dir_absent`, `record_dir_is_a_file`, `record_dir_symlink_loop` (a `context-usage` link pointing at itself), `locked_entry_is_still_read` (a `locked` marker beside `gitdir`; its worktree's record IS read), and `worktrees_dir_unreadable` (`<root>/.git/worktrees` is a file → only the primary directory is read).
- Decider: `go test ./internal/statusline -run '^TestQWR_AC004_UnusableEntriesContributeNothing$' -count=1 -v` → PASS with the eleven subtests listed (a skipped subtest is named in the report).
- RED-now: ledger E1. Green path: M2 (the entry handling).

### AC-QWR-005 — at most the given number of entries are read (statusline half of REQ-QWR-004)

- **Given** 130 metadata entries named `wt-000` … `wt-129` (sorted order is name order), each naming a worktree directory with a distinct fresh record whose percentage equals its index plus one, **when** the multi-directory reading runs with the bound argument 128, **then** the record of `wt-127` (the 128th entry) is visible, the records of `wt-128` and `wt-129` are not (the winner is chosen among the first 128 only); **when** it runs with the bound 3, **then** only `wt-000`..`wt-002` are read (the winner is the highest-index record of those three) and with the bound 1 only `wt-000`; and a `gitdir` file larger than 4 KiB is read for at most 4 KiB (a fixture whose first line is a valid path followed by 1 MiB of filler is accepted, one whose valid path begins after byte 4096 is skipped).
- Decider: `go test ./internal/statusline -run '^TestQWR_AC005_EnumerationBound$' -count=1 -v` → PASS (subtests `bound_128_entry_128_visible`, `bound_128_entry_129_and_130_not_read`, `bound_3_reads_three`, `bound_1_reads_one`, `gitdir_read_is_bounded`).
- RED-now: ledger E1. Green path: M2. Mutant pinned: no bound (all 130 read), a hard-coded 128 that ignores the argument, an off-by-one bound (127 or 129 for argument 128, 2 or 4 for argument 3).

### AC-QWR-006 — the new code is offline, spawn-free, read-only, and non-recursive

- **Given** the non-test files matching `internal/statusline/quota*.go` and `internal/cli/factory_quota*.go`, **when** their imports and identifiers are parsed and the multi-directory reading runs on a temporary directory tree, **then** the statusline swept set contains at least two files (`quota.go`, `quota_dirs.go`), none of the swept files imports `net`, `net/http`, or `os/exec`, none references `filepath.Walk`, `filepath.WalkDir`, `os.WriteFile`, `os.Create`, `os.Remove`, `os.RemoveAll`, `os.Mkdir`, or `os.MkdirAll` (a positive-control string containing each is flagged), and the bytes, names, and sizes of the primary directory, every worktree directory, and the metadata tree are identical before and after the run.
- Decider: `go test ./internal/statusline -run '^TestQWR_AC006_OfflineSpawnFreeReadOnlyNonRecursive$' -count=1 -v` → PASS; the predecessor's `go test ./internal/statusline ./internal/cli -run '^TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly$' -count=1 -v` → PASS in each package (its glob picks up `quota_dirs.go` automatically). A swept count below two fails the new test.
- RED-now: ledger E1, E2. Green path: M2.

### AC-QWR-007 — with the gate disabled nothing is read

- **Given** a root with metadata entries and worktree directories holding fresh high records, and `workflow.quota_gate.enabled: false`, **when** the evaluation, the lane gate, and the status block run, **then** the aggregator seam is called zero times (a counting replacement of `factoryQuotaAggregate`), the status output has no quota block, and with the gate enabled the same fixture calls the seam exactly once per evaluation (the positive control).
- Decider: `go test ./internal/cli -run '^TestQWR_AC007_GateDisabledReadsNothing$' -count=1 -v` → PASS (subtests `disabled_zero_seam_calls`, `enabled_one_seam_call`). Static complement: the enumerator is reachable only through the seam value; `grep -rn "QuotaStateDirs" internal/cli` over non-test files names one call site.
- RED-now: ledger E3. Green path: M3.

### AC-QWR-008 — the surfaces see a worktree-only reading, with unchanged output form

- **Given** the gate enabled, a Claude lane environment, the real aggregator (the seam not replaced), a root whose only fresh record (five-hour 92%, reset in the future) sits in a linked worktree's directory, and the fixtures of the predecessor's `qasFixture`, **when** `moai factory next` runs (CLI and the `factory_next` MCP tool), `moai factory status` runs in text and `--json`, and `moai integration acquire` runs, **then** `next_holds_on_worktree_only_92` (the hold line `quota hold: five_hour used=92.0% resets_at=…Z` on the error stream, nothing on standard output, exit status 3, no factory record changed); `assigned_card_still_leased` (arm (a) exempt, as in the predecessor); `leases_when_no_record_anywhere` (the same fixture without the worktree record leases a card); `status_block_pressure` (the quota block reports pressure); `acquire_warning_for_claude_caller` (one `quota warning:` line on the error stream, the lock record and standard output equal a gate-disabled run); `wait_recheck_sees_worktree_update` (under `--wait`, a worktree record that falls to 80% between passes releases the lane on the second pass, driven by the injected sleep).
- Decider: `go test ./internal/cli -run '^TestQWR_AC008_SurfacesSeeWorktreeReading$' -count=1 -v` (scrubbed anchored form, under a `moai slot` lease) → PASS with the six subtests listed.
- RED-now: ledger E3, plus the observed failing assertion of the first subtest on the unmodified code recorded verbatim at the start of M3 (`want hold line, got a leased card`). Green path: M3.

### AC-QWR-009 — the same reading renders identically wherever it was found

- **Given** two fixtures that carry byte-identical record files, one in the primary directory, one in a linked worktree's directory, **when** `moai factory status` (text and `--json`), the hold line, and the acquire warning are produced for each, **then** the outputs are byte-identical between the two fixtures (no source directory, no directory count, no new key or line), and the predecessor's goldens and tests stay green: `go test ./internal/cli -run '^(TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData|TestQAS_AC022_GateOffOrPressureOffOutputUnchanged)$' -count=1 -v` → two PASS lines.
- Decider: `go test ./internal/cli -run '^TestQWR_AC009_WorktreeSourceRendersIdentically$' -count=1 -v` → PASS (subtests `status_text`, `status_json`, `hold_line`, `acquire_warning`).
- RED-now: ledger E3. Green path: M3.

### AC-QWR-010 — a root without a git metadata directory reads its own directory alone

- **Given** a root with a fresh high record in what would be a worktree directory but whose git metadata is absent in each of three shapes — `no_git_entry` (no `.git`), `git_is_a_file` (a `.git` file reading `gitdir: …`), `no_worktrees_dir` (`.git` is a directory with no `worktrees` inside) — **when** the evaluation runs, **then** the result equals the single-directory reading of the root's own directory (a record there is seen; the would-be worktree record is not), exactly as before the change.
- Decider: `go test ./internal/statusline -run '^TestQWR_AC010_NoGitMetadataReadsRootOnly$' -count=1 -v` → PASS (three subtests); cli complement in `go test ./internal/cli -run '^TestQWR_AC010b_CliRootWithoutGitMetadata$' -count=1 -v` → PASS (the predecessor's fixtures, whose roots carry no `.git`, still lease).
- RED-now: ledger E1, E3. Green path: M2, M3.

### AC-QWR-011 — the real-lane measurement is recorded, in its own commit, before the implementation (one-shot, pre-merge)

- **Given** the run-phase commit graph, `CARD_BASE` read as `git merge-base develop HEAD`, B the first commit in `<CARD_BASE>..HEAD` touching `.moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/progress.md` with the evidence cell `QWR-M0` present in its content, and I the first commit in `<CARD_BASE>..HEAD` touching an implementation file (any file under `internal/` that is not a `_test.go` file and not under `testdata/`), **then** B exists, `QWR-M0` carries the before and after listings, the dates, the command measured, and the stated limits of §C, and B is a strict ancestor of I.
- Decider (plain commands, recorded verbatim in `progress.md` §E.2): `git merge-base develop HEAD`; `git log --reverse --format=%H <CARD_BASE>..HEAD -- .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/progress.md` (first line = B, non-empty); `git show <B>:.moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001/progress.md` carries `QWR-M0` (read, not piped); `git log --reverse --format=%H <CARD_BASE>..HEAD -- internal ':(exclude)*_test.go' ':(exclude)*/testdata/*'` (first line = I, non-empty); `git merge-base --is-ancestor <B> <I>` exit 0; `git rev-list --count <I>..<B>` prints `0`.
- RED-now: ledger E4 (neither the directory nor the cell exists at the parent). Green path: M0 commits the cell; every later commit follows it. Why a commit-graph criterion: a measurement committed together with the implementation it precedes leaves the ordering unverifiable (`verification-claim-integrity.md` §2.3). Pre-merge evaluation only (`gitflow-lane-protocol.md` §8): after the card merges `CARD_BASE` equals the card tip and the range is empty.

### AC-QWR-013 — the directory bound is the configuration key `workflow.quota_gate.max_scan_dirs`

- **Given** the Go defaults, the shipped template `workflow.yaml`, the local `.moai/config/sections/workflow.yaml`, the shipped-key inventory, and the config cache version, **when** the template is decoded and `config.LoadQuotaGate` runs on temporary roots, **then** (config package, `TestQWR_AC013_MaxScanDirsConfigKey`): `default_equals_template` (the Go default and the template value are both 128, and the template carries the key inside `quota_gate`); `template_comment_says_unmeasured` (the comment block above the template gate mentions `max_scan_dirs` and the word `unmeasured`, a positive control being the existing predecessor assertion on the same block); `local_twin_carries_key`; `absent_or_unparseable_yields_default` (absent key, absent file, unparseable file → 128); `out_of_range_yields_default` (0, -1, 1025, and a string value → 128; 1, 3, 128, and 1024 are honoured); `other_keys_unchanged` (the five predecessor keys keep their defaults and ranges when `max_scan_dirs` is out of range or set); `inventory_row_with_reader` (the inventory carries `workflow.quota_gate.max_scan_dirs` as class W with evidence `reader`); `cache_schema_bumped` (the cache schema version is 12); and (cli package, `TestQWR_AC013b_ProductionSeamAppliesConfiguredBound`) a root whose local `workflow.yaml` sets `max_scan_dirs: 3` with five worktree entries each holding a distinct fresh record reads only the first three through the production seam (the higher-index two are not seen), the same root with the key absent reads all five, and the seam call count per evaluation stays one.
- Decider: `go test ./internal/config -run '^TestQWR_AC013_MaxScanDirsConfigKey$' -count=1 -v` → PASS with the eight subtests listed; `go test ./internal/config -run '^(TestShippedConfigKeysHaveReaders|TestQAS_AC007_ConfigDefaultsMirrorTemplate)$' -count=1 -v` → two PASS lines (the predecessor's guards stay green with the new field); `go test ./internal/cli -run '^TestQWR_AC013b_ProductionSeamAppliesConfiguredBound$' -count=1 -v` (scrubbed anchored form, under a `moai slot` lease) → PASS.
- RED-now: ledger E5, E6. Green path: M2a (config key) and M3 (the production seam). Mutants pinned: a hard-coded bound ignoring the key, a range check that accepts 0 or 1025, a missing cache bump, a template without the key (the shipped-key guard).

### AC-QWR-012 — the writer, the other thresholds, the record schema, and the predecessor files are untouched (one-shot, pre-merge)

- **Given** the card branch after its last implementation commit and `CARD_BASE` read at that moment as `git merge-base develop HEAD`, **then** (1) `git diff --name-only <CARD_BASE>..HEAD -- internal/stateanchor internal/statusline/state_anchor.go internal/statusline/context_usage.go .moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001 internal/statusline/quota_test.go internal/cli/factory_quota_test.go internal/config/workflow_quota_gate_test.go` prints nothing; (2) `git diff --name-only <CARD_BASE>..HEAD -- internal/config` prints only paths from the allowed set `internal/config/types.go`, `internal/config/defaults.go`, `internal/config/loader_quota_gate.go`, `internal/config/cache.go`, `internal/config/testdata/shipped_key_inventory.yaml`, and the new test file `internal/config/quota_gate_scan_dirs_test.go`; (3) the `git diff -U0 <CARD_BASE>..HEAD -- internal/config/defaults.go` and the same for `internal/config/loader_quota_gate.go`, printed through `grep -c '^-[^-]'`, count zero removed lines (the predecessor defaults and ranges are intact; the key is added, nothing is rewritten), with the positive control that the same `git diff` printed through `grep -c '^+[^+]'` counts at least one added line in each file; with the positive control that `git diff --name-only <CARD_BASE>..HEAD` is non-empty and names both `internal/statusline/quota_dirs.go` and `internal/config/types.go` (an empty control reports "unmeasurable", never "no change").
- Decider (plain commands): `git merge-base develop HEAD`; the three `git diff` forms above, each plain and single-invocation. Pre-merge evaluation only.
- RED-now: not applicable as a red — a guard that holds at arrival (regression-guard, §E.2). Its mutants are the probes: editing any named path makes the first command print it. Green path: every milestone.

## §E RED-now evidence

### §E.1 Ledger

Measured in this plan phase on tree `284e09c44023598affe486f17701717ca173e6ca` (HEAD of `WT-quota-read-worktree-records`; clean tree at measurement). Each row is one single invocation, its stdout verbatim, its exit status as the harness reported it (`0` where the harness printed no failure line, `1` where it printed `Exit code 1`).

| Row | Command | Stdout | Exit status | Meaning (why red) |
|---|---|---|---|---|
| E1 | `go test -count=1 -run '^TestQWR_AC001_WorktreeOnlyRecordIsSeen$' ./internal/statusline` | `ok  	github.com/modu-ai/moai-adk/internal/statusline	0.542s [no tests to run]` | 0 | the new decider does not exist; the swept count is zero, which the pass convention reads as a Gap, never a pass (the same selector form stands for AC-QWR-002..006 and -010) |
| E2 | `ls internal/statusline/quota_dirs.go` | `ls: internal/statusline/quota_dirs.go: No such file or directory` (stderr) | 1 | the enumeration file does not exist |
| E3 | `ls internal/cli/factory_quota_worktrees_test.go` | `ls: internal/cli/factory_quota_worktrees_test.go: No such file or directory` (stderr) | 1 | the cli deciders (AC-QWR-007..010, and -013b by the same absence of new cli test files) do not exist |
| E4 | `ls .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001` | `ls: .moai/specs/SPEC-QUOTA-RECORD-WORKTREES-001: No such file or directory` (stderr) | 1 | at the parent of the plan commit neither the SPEC directory nor the measurement cell exists |
| E5 | `go test -count=1 -run '^TestQWR_AC013_MaxScanDirsConfigKey$' ./internal/config` | `ok  	github.com/modu-ai/moai-adk/internal/config	0.389s [no tests to run]` | 0 | the config decider does not exist; swept count zero, a Gap, never a pass. Measured at HEAD `8da5e1f39db1317f970a1e25831f22d8437f6f01`, whose `internal/` tree is byte-identical to `284e09c44` (only this SPEC directory differs) |
| E6 | `ls internal/config/quota_gate_scan_dirs_test.go` | `ls: internal/config/quota_gate_scan_dirs_test.go: No such file or directory` (stderr) | 1 | the config test file does not exist; same tree as E5 |

Each green path flips E1 to `--- PASS: <name> ` lines with a non-empty swept count, and E2, E3, E4 to existing files (E4 after the M0 commit also carries `QWR-M0`). The defect itself — a fresh reading that sits where the gate does not look — is a **historical observation** (spec.md §A: session `2da35a68-…` in four directories, 4 versus 19 fresh files, predecessor audit F1) that cannot be re-executed as a RED on this tree without a live session; it is recorded as context, not as a release-blocking RED. The repair is accepted on AC-QWR-001, -002, and -008, whose deciders construct the defect shape (a fresh reading only in a worktree directory) in a fixture.

### §E.2 Classification (verification-completeness §2.1)

An AC is **release-blocking** when its RED-now cell carries the four elements — a single-invocation read-only command, its verbatim stdout, its exit status, and the pinned tree SHA — and the starting observation can be re-executed on the pre-change tree. A criterion whose RED cannot be re-executed is a **regression-guard**.

| AC | Class | RED-now rows |
|---|---|---|
| AC-QWR-001, -002, -003 (new test), -004, -005, -006 | release-blocking | E1, E2 |
| AC-QWR-007, -008, -009 | release-blocking | E3 |
| AC-QWR-010 | release-blocking | E1, E3 |
| AC-QWR-011 | release-blocking | E4 |
| AC-QWR-013 | release-blocking | E5, E6 |
| AC-QWR-012 | regression-guard (holds at arrival; no red to observe) | — |

The old `TestQAS_` deciders named inside AC-QWR-003 and -009 are regression-guards by nature. Each test-decided AC also records, at the start of its milestone in `progress.md` §E.2, the verbatim failing output of its test written RED against the unmodified code (the "E8" convention of the predecessor): the RED must be red for the stated reason — the assertion that a worktree-only reading is seen fails on the old code, not a compile error or a fixture mistake.

### §E.3 Continued firing (verification-completeness §1.3)

How a reader learns that a check stopped firing, per check class, without asking:

- **Go-test deciders.** CI runs the full suite on every push (`.github/workflows/ci.yml:229`, `go test -json -coverprofile=coverage.out -covermode=atomic ./... > test-stream.json`, read in this plan phase), so each decider appears by name in that stream; a renamed or deleted decider is a missing name there, and the pass convention (§B) turns a selector that matches nothing into a Gap. The comparison of the AC roster against the stream is a manual read, not an automatic one — stated rather than implied.
- **The on-disk layout assumption (the failure that is silent by construction).** If git changed the layout of `<common>/worktrees/<entry>/gitdir`, the enumeration would silently return the primary directory alone: nothing fails and nothing is read, the exact "nothing failed, and there was nothing there to fail" shape. `real_git_worktree_layout` in AC-QWR-001 creates a worktree with the `git` on the CI runner and asserts the enumeration finds it, so a layout change turns red on the next push instead of leaving the gate blind; it skips (a named Gap) only when `git` is absent from `PATH`.
- **Sweep (AC-QWR-006).** Sweeps globs, not a named list, and fails below a floor of two statusline files, so a vanished or renamed `quota_dirs.go` shrinks the sweep visibly and a new `quota*.go` file joins it automatically.
- **Commit ordering and scope (AC-QWR-011, -012).** One-shot, pre-merge evaluations recorded once in `progress.md` §E.2 at the close of the run phase; vacuous after the card merges (the merge-base moves). Not recurring checks and not claimed to fire later.
- **Ledger rows (E1-E4).** Starting observations only; their continued life is the Go test of the same AC.

## §F Edge cases

- The same session id fresh in two directories (a lane that moved between trees inside `max_age`): the later capture time wins; equal capture times keep the primary directory's copy.
- A worktree listed in the metadata whose directory was removed with `rm -rf` (not `git worktree remove`): the `gitdir` target is missing → skipped, no error.
- A worktree created with `moai cc -w <abs-path>` outside the repository root: read through its `gitdir` (AC-QWR-001 `outside_root_layout`).
- A worktree whose `.moai/state` does not exist yet (no session has rendered there): skipped (AC-QWR-004 `record_dir_absent`).
- Windows paths in the `gitdir` file (`C:/…`): read with the platform's absolute-path rule; the Unix-style absolute-path fixtures skip with a named reason on Windows (a Gap there, not a pass).
- A gate that is enabled while the repository has more linked worktrees than `max_scan_dirs` (128 by default): entries beyond the bound are not read (fail open; the template comment tells the operator to raise the key).
- `max_scan_dirs` set to a value outside 1-1024, to a string, or to a mistyped block: the default 128 applies for that key (the other keys are resolved on their own, as before).

## §G Quality gate and Definition of Done

- All 13 deciders pass with a non-empty swept count; the old `TestQAS_` deciders named in AC-QWR-003, -009, and -013 stay green without an edit; `go vet ./internal/statusline ./internal/config ./internal/cli` and `GOOS=windows GOARCH=amd64 go build ./...` pass; the lint baseline is not worsened.
- Template-first (CLAUDE.local.md): the key is added to the Go defaults, the shipped template `workflow.yaml`, and the local `.moai/config/sections/workflow.yaml` in the same change, `make build` regenerates the embedded templates before the commit, and the config cache schema is bumped 11 to 12 with a comment line in the established style (AC-QWR-013 checks all of it). A sync-phase CHANGELOG entry goes through the usual sync commit.
- `moai spec lint` reports no error for this SPEC directory.
- `progress.md` §E.2 carries the M0 measurement cell `QWR-M0` (its own commit), per milestone the verbatim RED and GREEN output, the one timing observation of the evaluation, and the AC-QWR-011 and -012 evidence.
- A scoped re-audit of the delta (sync-auditor, scoped to this SPEC's changed files) is requested by the sync phase; its verdict file is read before the card advances (`kanban-dispatch.md` § Completion is read, never trusted).
