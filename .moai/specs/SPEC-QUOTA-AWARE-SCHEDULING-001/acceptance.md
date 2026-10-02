# acceptance.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

This file is the verification layer: Given-When-Then criteria, each decided by one named command. The requirement layer (GEARS) is `spec.md` §C. Revised at 0.5.0 to close plan-audit iteration 1 (`.moai/reports/t1347/plan-audit-iter1.md`, defects D1-D24).

## §A Scope of verification

The criteria verify: the rate-limit windows in the session telemetry record and its schema, throttle, and exhaustion time; the aggregator's freshness, boundary, skew, rollover, and fail-open rules; the `workflow.quota_gate` configuration, its template mirror and its local twin; the Claude-lane hold at `moai factory next` (CLI and MCP) with its hold line and in-wait latch; the unaffected non-Claude lanes and unaffected `stage`/`complete`; the `moai factory status` quota block (shown only when the gate is enabled and a window has data); the warn-only line at `moai integration acquire`; the offline, spawn-free, read-only property; the baseline-first commit ordering; the untouched kanban-mode paths; the lane's backend recorded at claim; and the steering: one shared pressure evaluation, a genuinely read-only lane inventory read from the registry, the `moai todo --auto` and `moai factory status` recommendation and no-candidate warning, no dispatch side effect, and unchanged output when the gate is disabled or pressure is off. Not verified here: kanban mode, Codex-lane slot behaviour, any real Claude Code runtime or real quota (no test starts a real `claude` or `codex`).

## §B Test-environment constraint (binds every AC)

- Every test builds its own project root with `t.TempDir()`; the MoAI home is redirected to a temporary directory. No test reads or writes the real `~/.moai` or the real `.moai/state`.
- Time is an injected clock (`factoryCardNow` exists for the factory seam, `factoryNextWaitSleep` for the wait loop); no test sleeps or reads the wall clock for a verdict. Quota records are fixture files written into a temporary `context-usage/` directory; no test calls the network.
- Lane environment in the CLI tests means the factory role marker set to the role-value constant, a lane label, and the backend variable under test; the gate predicate reads the launch-provider variable first and the kanban backend variable second (`DO-10`).
- `./internal/cli` runs only with an anchored `-run` selector naming the AC tests (`-run '^(TestQAS_AC008_...|TestQAS_AC008b_...)$'`); a whole-package run of `internal/cli` is not a verdict here (it measures the machine, not the change). `./internal/statusline` and `./internal/config` may run whole-package once at the end of M6. The claim tests in `internal/kanban` run with an anchored `-run` as well (`./internal/kanban` is a minutes-long package).
- **Pass convention:** an AC passes only when every command it names exits 0 and each test's verbose output carries a `--- PASS: <name> ` line for exactly the test the command names (name followed by a space, never a prefix match). `[no tests to run]`, `no tests ran`, or a zero swept count is a Gap, never a pass (`verification-completeness.md` §1.1). Subtests are decided by their parent's PASS line plus a `-v` listing of the subtest names the AC enumerates. At M0 each anchored selector is run once against the pre-implementation tree and must print `[no tests to run]` — that is the intended RED of a test that does not exist yet, recorded in `progress.md` §E.2.
- Defaults the tests assert come from the SPEC: hold 90 (five-hour) and 95 (seven-day), release margin 5, max age 30m, heartbeat 5m, clock-skew tolerance 5m, exhaustion 100, hold exit status 3 (the no-card status, DO-5). The thresholds, margin, and age are **unmeasured defaults** (DO-3, accepted as provisional by the leader); the heartbeat, skew tolerance, and exhaustion percentage are compiled constants, also unmeasured. All change here and in the tests together when the leader sets other values.
- Boundary convention: a reading holds when it is at or above its hold percentage; a held lane is released when the reading is below the hold percentage minus the release margin; a record is fresh while its age is at most the max age.

## §C Traceability

| REQ | AC |
|---|---|
| REQ-QAS-001 | AC-QAS-001 |
| REQ-QAS-002 | AC-QAS-002 |
| REQ-QAS-003 | AC-QAS-003 |
| REQ-QAS-004 | AC-QAS-004 |
| REQ-QAS-005 | AC-QAS-005 |
| REQ-QAS-006 | AC-QAS-006 |
| REQ-QAS-007 | AC-QAS-014 |
| REQ-QAS-008 | AC-QAS-007 |
| REQ-QAS-009 | AC-QAS-008 |
| REQ-QAS-010 | AC-QAS-009 |
| REQ-QAS-011 | AC-QAS-010 |
| REQ-QAS-012 | AC-QAS-011 |
| REQ-QAS-013 | AC-QAS-012 |
| REQ-QAS-014 | AC-QAS-013 |
| REQ-QAS-015 | AC-QAS-016 |
| REQ-QAS-016 | AC-QAS-015 |
| REQ-QAS-017 | AC-QAS-017 |
| REQ-QAS-018 | AC-QAS-018 |
| REQ-QAS-019 | AC-QAS-019 |
| REQ-QAS-020 | AC-QAS-020 |
| REQ-QAS-021 | AC-QAS-021 |
| REQ-QAS-022 | AC-QAS-022 |
| REQ-QAS-023 | AC-QAS-023 |

## §D Acceptance criteria

Each entry ends with its **RED-now** cell (the pre-change observation, ledger row id in §E) and its **green path** cell (the milestone that flips it and what the passing output becomes). §E.2 classifies every AC as release-blocking or regression-guard. The behaviour decided only by a not-yet-written test records its verbatim RED at the start of the milestone (E8 in `progress.md` §E.2), because the test does not exist to run today.

### AC-QAS-001 — the record carries exactly the supplied windows

- **Given** a statusline stdin payload with a five-hour window (62.5%, reset epoch 1790000000) and a seven-day window (41.2%, reset epoch 1790500000), **when** the builder runs against a temporary project directory, **then** the session record decodes with both windows carrying exactly those percentages and reset times; with a payload carrying only the seven-day window the five-hour window is absent from the record; with a payload carrying no `rate_limits` the record carries no window.
- Decider: `go test ./internal/statusline -run '^TestQAS_AC001_RecordCarriesSuppliedWindowsOnly$' -count=1 -v` → `--- PASS: TestQAS_AC001_RecordCarriesSuppliedWindowsOnly ` and subtests `both_windows`, `seven_day_only`, `no_rate_limits`.
- RED-now: ledger E1. Green path: M1.

### AC-QAS-002 — an old-schema record reads as "no windows"; a window-less record keeps its bytes

- **Given** record fixtures at schema version 1 and 2 (no window fields) and a window-less record built from a fixed input, **when** they are read and serialized, **then** the old fixtures decode without error and report no windows, and the window-less serialization is byte-identical to the baseline golden after normalizing only the capture time, the writer pid, and the schema version value.
- Decider: `go test ./internal/statusline -run '^(TestQAS_AC002_PreviousSchemaReadsAsNoWindows|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline)$' -count=1 -v` → two PASS lines; regression guard `go test ./internal/statusline -run '^TestReadsPreviousSchemaRecord$' -count=1 -v` → PASS (`session_telemetry_payload_test.go:121`).
- RED-now: ledger E6 (schema constant is `2`, the AC expects `3`) and E9 (the baseline golden's directory does not exist). Green path: M1. Depends on the baseline golden `internal/statusline/testdata/qas_baseline_windowless_record.golden.json` measured before any implementation (AC-QAS-015).

### AC-QAS-003 — throttle: truncated bucket, heartbeat, window drop, window-less unchanged

- **Given** an on-disk record carrying a five-hour window at 62.2% and an injected clock, **when** the statusline renders again, **then**: (a) 62.9% within the heartbeat (the same truncated integer, 62, and the same reset time) writes nothing (file bytes and modification time unchanged); (b) 63.0% writes; (c) an unchanged reading older than the heartbeat rewrites with a refreshed capture time; (d) the window disappearing from stdin writes a record without it; (e) a record that carries no window and is unchanged beyond the heartbeat is NOT rewritten (throttled exactly as before).
- Decider: `go test ./internal/statusline -run '^TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop$' -count=1 -v` → PASS with subtests (a)-(e) listed; positive controls `go test ./internal/statusline -run '^(TestWriteContextUsage_ThrottleSkipUnchanged|TestThrottleUnaffectedByModelAndEffort)$' -count=1 -v` → two PASS lines.
- RED-now: ledger E7. Green path: M1.

### AC-QAS-004 — first-observed-exhausted time is sticky until the window rolls

- **Given** a window first observed at 100% at time T with reset R, **when** later writes carry the window at 100% with reset R, **then** the exhausted time stays T; **when** the reset becomes R2, or the window is no longer carried, **then** the exhausted time is dropped; a window first seen at 99% carries no exhausted time. (Where this time is read: the status quota block, AC-QAS-012.)
- Decider: `go test ./internal/statusline -run '^TestQAS_AC004_ExhaustedAtStickyUntilRollover$' -count=1 -v` → PASS with the four subtests.
- RED-now: ledger E1. Green path: M1.

### AC-QAS-005 — freshest wins (not max); boundaries; skew; reset is reset, not stale-high

- **Given** records in a temporary `context-usage/` directory and the clock at T, **when** the aggregator runs, **then**:
  - `freshest_not_max_five_hour`: an older fresh record (captured T-20m) carries the five-hour window at 95%, a newer fresh record (T-5m) carries it at 60% → the reading is 60%;
  - `freshest_not_max_seven_day`: an older fresh record (T-25m) carries the seven-day window at 97%, a newer one (T-2m) carries it at 70% → the reading is 70%;
  - `per_window`: a newer record carrying only the seven-day window does not hide an older fresh record's five-hour reading;
  - `age_exactly_max`: a record captured exactly 30m before T is fresh; `age_max_plus_1s`: a record captured 30m and 1s before T contributes nothing (unknown);
  - `future_capture_within_tolerance`: a record captured 5m after T (the clock-skew tolerance) is fresh; `future_capture_beyond_tolerance`: a record captured 5m and 1s after T is unknown, whatever its percentage;
  - `reset_window_is_reset`: a fresh record whose window reset time is one second before T at 99% gives the state reset (not hold, not stale).
- Decider: `go test ./internal/statusline -run '^TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover$' -count=1 -v` → PASS with the seven subtests listed.
- RED-now: ledger E1 and E13 (no aggregator file). Green path: M2. The mutants this pins: maximum-over-fresh-records (rejected D2 option B), `<` for `<=` at the age boundary, no skew guard.

### AC-QAS-006 — absent or unreadable data fails open; an unreadable configuration yields the default max age

- **Given** a missing record directory, an unparseable record file, only stale records, and an unparseable `workflow.yaml`, **when** the aggregator and then `moai factory next` for a Claude lane with the gate enabled and a leasable queued card run, **then** every window reads unknown where the data is absent, stale, or unparseable, the unparseable configuration leaves the aggregator on the default max age (30m) and the gate off (the shipped default), and the lane leases the card exactly as with the gate disabled.
- Decider: `go test ./internal/statusline -run '^TestQAS_AC006_FailOpenOnAbsentOrUnreadable$' -count=1 -v` → PASS (four subtests); `go test ./internal/cli -run '^TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent$' -count=1 -v` → PASS (the same four fixtures, a leased card in each).
- RED-now: ledger E2 (no gate exists to fail open) and E13. Green path: M2 (aggregator) and M3 (lane).

### AC-QAS-007 — the configuration keys exist, with defaults mirrored in the template, documented as unmeasured, and enabled locally

- **Given** the Go defaults, the shipped template `workflow.yaml`, the local `.moai/config/sections/workflow.yaml`, the shipped-key inventory, and the config cache version, **when** the template is decoded into the workflow config, **then** `workflow.quota_gate` carries `enabled: false`, `five_hour_hold_pct: 90`, `seven_day_hold_pct: 95`, `release_margin_pct: 5`, `max_age: 30m`, each equal to the Go default; the template contains no `enabled: true` for the gate; the comment block the template carries above the gate contains the word `unmeasured`; the local file decodes with `quota_gate.enabled: true`; an absent key, an absent file, an unparseable file, and a value outside its valid range (a hold percentage outside 1-100, a release margin outside 0-50, a non-positive or unparseable duration, a `max_age` below twice the 5-minute heartbeat) each yield the default; the five keys are classified in the inventory; and the cache schema version is greater than 10.
- Decider: `go test ./internal/config -run '^TestQAS_AC007_ConfigDefaultsMirrorTemplate$' -count=1 -v` → PASS (subtests `defaults_equal_template`, `template_ships_off`, `template_comment_says_unmeasured`, `local_twin_enabled`, `absent_or_unparseable_yields_default`, `out_of_range_yields_default`, `max_age_below_twice_heartbeat_yields_default`, `cache_schema_bumped`); `go test ./internal/config -run '^TestShippedConfigKeysHaveReaders$' -count=1 -v` → PASS (`shipped_key_reader_test.go:74`); the positive control inside the new test decodes the sibling `slot_lease.default_max_duration` from the same template (the pattern of `workflow_jev_test.go:84-88`).
- RED-now: ledger E3, E4, E8. Green path: M2.

### AC-QAS-008 — a held Claude lane leases no new card but still receives its own assigned card; MCP form pinned

- **Given** the gate enabled, a lane environment whose launch provider is Claude, a queued card, a picked-unassigned card, a picked card with no record row, and a fresh record with the five-hour window at 92% (reset in the future), **when** `moai factory next` runs, **then** none of those three cards is leased, the factory record and the queue are byte-identical before and after, one hold line is printed on the error stream with standard output empty, and the CLI exits with status 3; **given** the same state plus a card assigned to this lane (arm (a)), **then** that card is leased and returned exactly as with the gate disabled and no hold line is printed. **Boundaries:** the five-hour window at exactly 90.0% holds and at 89.9% leases normally; the seven-day window at exactly 95.0% holds and at 94.9% leases normally. **MCP form (`AC-QAS-008b`):** **when** the `factory_next` MCP handler runs in the held state, **then** its result text equals the hold line (one line beginning `quota hold:`), the result is not an error result, and the text does not contain `no card is available`; with a card assigned to the lane the result is the unchanged leased-card text.
- Decider: `go test ./internal/cli -run '^(TestQAS_AC008_ClaudeLaneHeldAtThreshold|TestQAS_AC008b_MCPFactoryNextHeld)$' -count=1 -v` → two PASS lines; subtests `five_hour_92_new_cards_skipped`, `five_hour_92_assigned_card_still_leased`, `five_hour_exactly_90_holds`, `five_hour_89_9_leases`, `seven_day_exactly_95_holds`, `seven_day_94_9_leases` (and for the MCP test `hold_text_is_the_hold_line`, `not_an_error_result`, `no_no_card_text`, `assigned_card_unchanged`).
- RED-now: ledger E2. Green path: M3. Mutants pinned: `>` for `>=` on either window, an MCP handler that returns `no card is available` while holding.

### AC-QAS-009 — the hold line carries the reset time

- **Given** the hold fixtures of AC-QAS-008, **when** the hold line is captured, **then** it is exactly one line on the error stream (standard output carries nothing, whereas a true empty queue prints `no card is available` on standard output and no `quota hold:` line, both with status 3), matches `^quota hold: five_hour used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$` for one held window, carries one such segment per held window separated by `; ` when both are held, and the instant equals the fixture's reset epoch rendered in UTC.
- Decider: `go test ./internal/cli -run '^TestQAS_AC009_HoldLineCarriesResetTime$' -count=1 -v` → PASS.
- RED-now: ledger E2. Green path: M3.

### AC-QAS-010 — the wait latch releases only below the margin, at the reset, or when the reading ages out

- **Given** a held Claude lane in `moai factory next --wait` and an injected sleep that advances fixtures and the clock, **when** the reading goes 92% → 88% → 85.0% → 84.9%, **then** the verb stays held at 88% and at exactly 85.0% (release needs *below* 85) and leases at 84.9%; **when** the reading stays at 92% and the reset time passes, **then** the verb leases once the window reads reset; **when** the held record was first read at 20m old and the wait advances until its age is 30m and 1s while the reset time is still in the future, **then** the reading is unknown and the verb is released and leases the queued card (`unknown_mid_wait_releases`: the in-wait hold lasts only while the reading is fresh); **when** the wait bound elapses while held and still fresh, **then** it exits with status 3 and the stderr hold line; **when** a card is assigned to the lane while the verb waits held, **then** the next re-check leases that card.
- Decider: `go test ./internal/cli -run '^TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown$' -count=1 -v` → PASS (subtests `margin_release`, `exactly_85_stays_held`, `reset_release`, `unknown_mid_wait_releases`, `bound_elapsed`, `assigned_card_leased_during_wait`).
- RED-now: ledger E2. Green path: M3. Mutants pinned: no latch, release at 88, release at 85.0, a latch that survives unknown.

### AC-QAS-011 — non-Claude lanes are never held; stage and complete are never gated

- **Given** a 99% fresh five-hour record and the gate enabled, **when** `moai factory next` runs with the lane's backend set to gpt, glm, empty, and an unrecognised value, **then** each leases the card exactly as with the gate disabled; **given** a held Claude lane with a card already leased, **when** `moai factory stage` and `moai factory complete` run for that card, **then** both behave as before the change.
- Decider: `go test ./internal/cli -run '^(TestQAS_AC011_NonClaudeBackendsNeverHeld|TestQAS_AC011b_StageAndCompleteIgnoreQuotaHold)$' -count=1 -v` → two PASS lines; regression guards `go test ./internal/cli -run '^(TestSD_AC023_CodexNextSkipsUnadvanceableCard|TestSD_AC008_NextSelectionOrderAndOutput)$' -count=1 -v` → two PASS lines (Codex skip and selection order intact).
- RED-now: ledger E2. Green path: M3.

### AC-QAS-012 — `moai factory status` shows the quota block only when the gate is enabled and a window has data

- **Given** fixture records and a factory record, **when** `moai factory status` runs in text and with `--json`, **then**: with the gate enabled and a window having data, the text carries a quota line per window with used percentage, reset time, source capture time, state (fresh, reset, or unknown), held/not-held for a Claude lane, and the recorded first-exhausted time where one exists; the JSON carries a `quota` key with the same fields; with the gate enabled, data present, and pressure off, the output is exactly the baseline golden plus the quota block (the golden's lines and keys all present and unchanged, no steering, hold, or warning line); with the gate disabled — data present or not — and with the gate enabled but no window data anywhere, the output has no quota block and is byte-identical to the baseline golden `internal/cli/testdata/qas_baseline_factory_status.golden.json`.
- Decider: `go test ./internal/cli -run '^(TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData|TestFactoryStatusShowsHolderModePriority)$' -count=1 -v` → two PASS lines; subtests `enabled_with_data_adds_block_only`, `gate_disabled_with_data_equals_golden`, `enabled_no_data_equals_golden`, `exhausted_at_shown_text_and_json`, `fields_per_window`.
- RED-now: ledger E2 and E9. Green path: M5. Depends on the baseline golden (AC-QAS-015).

### AC-QAS-013 — `moai integration acquire` warns and never blocks

- **Given** the gate enabled, a Claude lane environment, and a fresh window at or above its hold percentage, **when** `moai integration acquire` runs (text and `--json`), **then** the error stream carries one line naming the window and its reset time, and the lock record, the exit status, and standard output (the `--json` object included) are identical to a run with the gate disabled — acquiring the window is never blocked, refused, or delayed; with a window below its threshold, with the gate disabled, or with a non-Claude backend, the error stream carries no such line.
- Decider: `go test ./internal/cli -run '^TestQAS_AC013_AcquireWarnsNeverRefuses$' -count=1 -v` → PASS.
- RED-now: ledger E5. Green path: M6.

### AC-QAS-014 — the quota path is offline, spawn-free, and read-only (swept by glob, with a minimum count)

- **Given** the non-test files matching the globs `internal/statusline/quota*.go` and `internal/cli/factory_quota*.go`, **when** their imports are parsed and the aggregator and the lane inventory run on a temporary directory snapshot, **then** the swept set contains at least three files (`internal/statusline/quota.go`, `internal/cli/factory_quota.go`, `internal/cli/factory_quota_lanes.go`), none imports `net`, `net/http`, or `os/exec`, and the directory tree (names, sizes, bytes) of the record directory is identical before and after the aggregator run. (The lane inventory's own read-only property, which concerns the registry, is AC-QAS-018.)
- Decider: `go test ./internal/statusline ./internal/cli -run '^TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly$' -count=1 -v` → PASS in each package that defines it; a swept file count below 3 fails the test (a vanished or renamed file cannot shrink the sweep silently).
- RED-now: ledger E13. Green path: M2, M3, and M5 (each package's test is written RED in M2 or M3 and the glob picks up `factory_quota_lanes.go` when M5 adds it).

### AC-QAS-015 — the baseline lands before the implementation (one-shot, pre-merge)

- **Given** the run-phase commit graph, `CARD_BASE` read as `git merge-base develop HEAD`, three baseline goldens (`internal/statusline/testdata/qas_baseline_windowless_record.golden.json`, `internal/cli/testdata/qas_baseline_factory_status.golden.json`, `internal/cli/testdata/qas_baseline_todo_auto.golden.txt`), B_i the first commit adding golden i, and I the first commit in `<CARD_BASE>..HEAD` touching any implementation file — any file under `internal/` that is not a `_test.go` file and not under a `testdata/` directory, including at minimum `internal/statusline/context_usage.go` — **then** each B_i exists and is a strict ancestor of I.
- Decider (plain commands, recorded verbatim in `progress.md` §E.2): `git merge-base develop HEAD`; for each golden path `git log --reverse --format=%H <CARD_BASE>..HEAD -- <golden-path>` (first line = B_i, non-empty); `git log --reverse --format=%H <CARD_BASE>..HEAD -- internal ':(exclude)*_test.go' ':(exclude)*/testdata/*'` (first line = I, non-empty); then for each B_i `git rev-list --count B_i..I` prints at least `1` and `git rev-list --count I..B_i` prints `0`.
- RED-now: ledger E9. Green path: the M0 commit adds all three goldens and touches no implementation file; every later milestone commit follows it. Why a commit-graph criterion: a baseline committed together with the implementation it measured leaves the ordering unverifiable (`verification-claim-integrity.md` §2.3). Mutants pinned: goldens committed after M1's `context_usage.go` edit, one golden early and two late.

### AC-QAS-016 — no kanban-mode path is touched; the claim path is the one shared edit (one-shot, pre-merge)

- **Given** the card branch after its last implementation commit and `CARD_BASE` read at that moment as `git merge-base develop HEAD`, **when** the changed-file list and the hunks are taken, **then** (1) the only files it names under `internal/kanban/` are `factory_slots.go` and `factory_slots_test.go`, and it names neither `internal/cli/kanban.go` nor `internal/cli/kanban_settings.go`; and (2) every hunk of the shared launcher files falls inside the factory-lane claim call lines named in `plan.md` §D — by old-file line numbers: `internal/cli/cc.go` line 250, `internal/cli/glm.go` line 294, `internal/cli/codex_launcher.go` line 932, `internal/cli/codex_factory.go` lines 149-150, `internal/cli/factory.go` lines 790-792 — so no hunk lies in the kanban leader or companion branches of `cc.go` and `glm.go`.
- Decider (plain commands): `git merge-base develop HEAD` (= `CARD_BASE`); `git diff --name-only <CARD_BASE>..HEAD -- internal/kanban internal/cli/kanban.go internal/cli/kanban_settings.go` prints nothing but `internal/kanban/factory_slots.go` and `internal/kanban/factory_slots_test.go`; for each of the five launcher files `git diff -U0 <CARD_BASE>..HEAD -- <file>` printed through `grep '^@@'` gives hunk headers whose old-side start (`-a,b`) lies inside the allowed lines above, with a positive control that each of the five files shows at least one hunk and that `git diff --name-only <CARD_BASE>..HEAD` is non-empty (an empty control reports "unmeasurable", never "no change"). Pre-merge evaluation only (`gitflow-lane-protocol.md` §8): after the card merges, `CARD_BASE` equals the card tip and the range is empty.
- RED-now: not applicable as a red; a guard that holds at arrival (regression-guard, §E.2). Its mutants are the probes: editing `internal/cli/kanban.go`, any other file under `internal/kanban/`, or a hunk in the kanban branches of `cc.go` / `glm.go` must make a command above print that path or header. Green path: every milestone.

### AC-QAS-017 — one pressure evaluation, caller-independent; the surfaces apply their own caller rule

- **Given** the gate enabled and a fixture with the five-hour window at 92% (fresh, reset in the future), **when** the shared pressure function is evaluated and each surface is exercised, **then** (`shared_function_ignores_caller`) the shared function returns pressure on for every caller backend (claude, glm, gpt, empty) and returns the same held window and reading each time; (`at_92`) the lane gate holds a Claude lane, the `moai factory status` block and the `moai todo --auto` line report pressure, and the acquire warning is printed for a Claude caller; (`acquire_requires_claude_caller`) the acquire warning is NOT printed for a non-Claude caller while the shared function still returns pressure (REQ-QAS-014 and AC-QAS-013); (`at_89_9`, `reset`, `unknown`, `gate_disabled`) with the window at 89.9%, reset, unknown, or the gate disabled the shared function and every surface report no pressure.
- Decider: `go test ./internal/cli -run '^TestQAS_AC017_SharedPressureEvaluationAndSurfaces$' -count=1 -v` → PASS (subtests as named).
- RED-now: ledger E2, E5, E10. Green path: M3 (the shared function) and M5-M6 (the surfaces).

### AC-QAS-018 — the lane inventory lists live non-Claude lanes from the registry through a genuinely read-only open

- **Given** a temporary factory registry written by the real schema with: lane-1 alive, backend `claude`; lane-2 alive, backend `glm`; lane-3 alive, backend `gpt`; lane-4 registered with a dead pid, backend `glm`; lane-5 alive with an empty backend (a row written before REQ-QAS-023, inserted with the pre-change statement); lane-6 alive with an unrecognised backend value; session records present that name a conflicting backend for lane-5; and, separately, an absent registry file, an unreadable registry file, and a registry in a read-only directory, **when** the inventory is read, **then** the candidates are lane-2 (`glm`) and lane-3 (`gpt`), the unknown count is 2 (lane-5 and lane-6), lane-1 is excluded, an empty or unrecognised backend is never a candidate whatever any session record says, and an absent, unreadable, or un-openable registry yields no candidates and no error exit; the registry's main database file bytes (SHA-256), and a row dump of the `workers`, `cards`, and `meta` tables, are identical before and after (so no schema statement, no `legacy_workers_imported` marker, and no row was written); no project or home directory is created that did not exist (the absent-file case leaves the directory listing unchanged); the read opens the database read-only and runs no `EnsureProjectLayout`, `OpenFactoryPath`, or `ImportLegacyWorkers`. The SQLite `-wal`/`-shm` sidecar files a read-only reader of a WAL database may create are not part of the assertion (measured behaviour, plan.md §D).
- Decider: `go test ./internal/cli -run '^TestQAS_AC018_LaneInventoryCandidates$' -count=1 -v` → PASS (subtests `candidates_and_unknown`, `legacy_empty_is_unknown`, `session_record_is_not_read`, `absent_registry_no_candidates_no_create`, `unreadable_yields_none`, `read_only_main_file_and_rows_unchanged`).
- RED-now: ledger E12 (the backend is never written) and E13 (no inventory file). Green path: M5 (after the claim write of M4).

### AC-QAS-019 — pressure on: `--auto` and `moai factory status` recommend the candidate lanes

- **Given** the gate enabled, pressure on (five-hour 92%), the registry of AC-QAS-018, a two-card queue and the Jev line stubbed, **when** `moai todo --auto` runs one cycle and `moai factory status` runs in text and `--json`, **then** the `--auto` output carries, immediately before each `accept` line, exactly one line matching `^quota pressure: five_hour used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z; recommend non-Claude lane\(s\): lane-2 \(glm\), lane-3 \(gpt\)$`; the status text carries the same line under the quota block; the status JSON `quota` key carries `pressure: true`, the two candidates with label and backend, and `unknown_lanes: 2`.
- Decider: `go test ./internal/cli -run '^TestQAS_AC019_AutoAndStatusRecommendNonClaudeLanes$' -count=1 -v` → PASS (subtests `auto_line_before_each_accept`, `status_text`, `status_json`).
- RED-now: ledger E10, E11. Green path: M5.

### AC-QAS-020 — pressure on and no candidate lane: a warning, nothing else

- **Given** pressure on and an inventory with only claude lanes, only dead lanes, or no registered lane, **when** `moai todo --auto` and `moai factory status` run, **then** each prints one line matching `^quota pressure: five_hour used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z; warning: no live non-Claude lane \(unknown backend: [0-9]+\); nothing is re-dispatched$` in place of the recommendation line, the status JSON carries `warning: "no-non-claude-lane"` and an empty candidate list, and every other line of the output (accept, directive, evidence, done) is identical to the pressure-off output of the same fixture.
- Decider: `go test ./internal/cli -run '^TestQAS_AC020_NoNonClaudeLaneWarnsOnly$' -count=1 -v` → PASS (subtests `only_claude_lanes`, `only_dead_lanes`, `no_lanes`, `rest_of_output_unchanged`).
- RED-now: ledger E10, E11. Green path: M5.

### AC-QAS-021 — the steering changes nothing

- **Given** pressure on, a queue, a factory record, a lane registry, and a session registry, **when** the recommendation is produced through `moai factory status` and through the `--auto` cycle, **then** the set of queue mutations the `--auto` cycle performs is identical to the pressure-off run of the same fixture (its own pick and done only), the registry's main database file bytes and the row dump of `workers`, `cards`, and `meta` are identical before and after, the session registry is byte-identical, no card changes state or owner, and the files matching the globs of AC-QAS-014 import neither `net`, `net/http`, nor `os/exec` (the static check is AC-QAS-014's; the steering wiring edits in `todo_auto.go`, `todo.go`, `factory_card.go` and `integration.go` are outside that sweep and are covered by the behavioural assertions here).
- Decider: `go test ./internal/cli -run '^TestQAS_AC021_SteeringChangesNothing$' -count=1 -v` → PASS (subtests `status_is_read_only`, `auto_mutations_equal_pressure_off`, `no_card_class_read`).
- RED-now: ledger E10, E13. Green path: M5.

### AC-QAS-022 — gate disabled or pressure off: the output is unchanged

- **Given** the baseline goldens measured before the change (`qas_baseline_factory_status.golden.json`, `qas_baseline_todo_auto.golden.txt`) and fixtures that carry quota data (windows below their hold percentage, reset, or unknown) as well as fixtures that carry none, **when** the gate is disabled and when it is enabled with pressure off, **then**: `moai todo --auto` (Jev line stubbed) is byte-identical to its golden in every combination; `moai factory status` with the gate disabled — data present — is byte-identical to its golden; `moai factory next` and `moai integration acquire` produce the same stdout, stderr, and exit status with the gate enabled and pressure off as with the gate disabled; and with the gate enabled, data present, and pressure off, `moai factory status` adds only the informational quota block (AC-QAS-012) — no steering, hold, or warning line appears in any surface.
- Decider: `go test ./internal/cli -run '^TestQAS_AC022_GateOffOrPressureOffOutputUnchanged$' -count=1 -v` → PASS (subtests `status_gate_disabled_with_data`, `auto_gate_disabled`, `auto_below_threshold`, `auto_unknown_windows`, `next_pressure_off_equals_gate_disabled`, `acquire_pressure_off_equals_gate_disabled`, `no_steering_line_when_pressure_off`); existing guards `go test ./internal/cli -run '^(TestFactoryStatusShowsHolderModePriority|TestTodoAutoSerialCycle|TestTodoAutoCreatesNoFactoryLease|TestTodoAutoClearGuidancePerCard)$' -count=1 -v` → four PASS lines (`todo_auto_test.go:247,395,339`, `factory_classify_test.go:392`).
- RED-now: ledger E9 (baseline goldens absent); the existing `--auto` tests are the pre-change measurement. Green path: M0 (goldens), M5, M6. Depends on the baseline goldens (AC-QAS-015).

### AC-QAS-023 — the lane claim records the backend in its own insert and nothing else changes

- **Given** a temporary project root and a factory registry, **when** a lane claims its label through the Claude lane path (backend `claude`), the GLM lane path (`glm`), the Codex lane-loop path (`gpt`), and the Codex factory-entry path whose launcher token is `codex`, **then** each registry row carries the backend `claude`, `glm`, `gpt`, and `gpt` respectively. The registry used for these cases carries two test triggers: `BEFORE UPDATE ON workers` raising an error (so a claim that inserts the row and then updates the backend fails) and `AFTER INSERT ON workers WHEN NEW.backend = ''` raising an error (so the insert statement itself must carry the backend); both claims pass. **When** a claim is made through the existing entry points that carry no backend, in a registry without the insert trigger, **then** the row's backend is the empty value; **when** the Codex launcher updates the row's pid after the child starts, **then** the backend and every other column except pid and heartbeat are unchanged; **then** the registry's `schema_version` meta value is still `5` after every case and the factory DDL carries no new `ALTER TABLE workers` statement.
- Decider: `go test ./internal/kanban -run '^TestQAS_AC023_ClaimRecordsBackend$' -count=1 -v` → PASS (subtests `claim_with_backend_in_the_insert`, `empty_without_backend`, `codex_token_normalized`, `concurrent_claims_each_carry_backend`); `go test ./internal/cli -run '^TestQAS_AC023b_LaunchersPassTheirBackendToTheClaim$' -count=1 -v` → PASS (subtests `cc_claude`, `glm_glm`, `codex_loop_gpt`, `codex_factory_entry_gpt`, `codex_pid_update_preserves_backend`, built on the launchers' existing seams: the binary lookup and `factoryProcessAlive`); regression guards `go test ./internal/kanban -run '^(TestClaimFactoryLaneWithinBounds|TestClaimFactoryLaneWithinConcurrentOneSlot|TestClaimFactoryWorkerNameConcurrentClaimsAreUnique|TestClaimFactoryWorkerNameCanonicalOnly|TestFactoryFreeSlots|TestPruneFactoryDeadClaims)$' -count=1 -v` → six PASS lines (`factory_slots_test.go:17,75,109,141,231`, `factory_worker_label_test.go:91`); the schema assertion reads `meta.schema_version` from the registry and searches the DDL text for `ALTER TABLE workers` with a positive control on `ALTER TABLE runs`.
- RED-now: ledger E12. Green path: M4.

## §E RED-now evidence

### §E.1 Ledger

Measured in this plan phase on tree `245242438320d4f570aecbdec9ed6f524a4a4a99` (HEAD of `WT-quota-aware-scheduling`; its code under `internal/` is byte-identical to `c50da9c2f`, the tree the earlier measurement and the installed `moai` build — `moai version` → `gc50da9c2f` — name; the commits between them change only this SPEC directory). Each row is one invocation of the form `<command>; echo "rc=$?"`; the stdout is verbatim and the exit status is the observed `rc=` value, re-measured by the author in this revision (not copied from the audit). Zero-count `grep` exits 1; a printing `grep` that matches exits 0.

| Row | Command | Stdout | Exit status | Meaning (why red) |
|---|---|---|---|---|
| E1 | `grep -c "rate_limits" internal/statusline/context_usage.go` | `0` | 1 | the record has no window today |
| E2 | `grep -c -i "quota" internal/cli/factory_card.go` | `0` | 1 | the lease seam has no quota logic |
| E3 | `grep -c "quota_gate" internal/template/templates/.moai/config/sections/workflow.yaml` | `0` | 1 | the template carries no gate key |
| E4 | `grep -c "quota_gate" internal/config/types.go` | `0` | 1 | the Go config carries no gate key |
| E5 | `grep -c -i "quota" internal/cli/integration.go` | `0` | 1 | acquire has no quota line |
| E6 | `grep -n "contextUsageSchemaVersion = " internal/statusline/context_usage.go` | `32:const contextUsageSchemaVersion = 2` | 0 | the schema version is 2; the AC expects 3 (a matching grep, so exit 0 is the expected red) |
| E7 | `grep -c -i "heartbeat" internal/statusline/context_usage.go` | `0` | 1 | the throttle has no heartbeat |
| E8 | `grep -c "quota_gate" internal/config/testdata/shipped_key_inventory.yaml` | `0` | 1 | the inventory has no gate entries |
| E9 | `ls internal/statusline/testdata` | `ls: internal/statusline/testdata: No such file or directory` | 1 | the statusline baseline golden's directory does not exist yet. CLI side: `ls internal/cli/testdata \| wc -l` prints `37` and `ls internal/cli/testdata \| grep -c qas_` prints `0` with rc=1 — no `qas_*` golden exists |
| E10 | `grep -c -i "quota" internal/cli/todo_auto.go` | `0` | 1 | the `--auto` cycle has no quota or lane logic |
| E11 | `grep -c -i "quota" internal/cli/todo.go` | `0` | 1 | the `--auto` wiring has no quota seam |
| E12 | `grep -c "INSERT INTO workers(.*backend" internal/kanban/factory_slots.go` | `0` | 1 | no lane claim insert names the backend column (column-order independent; the earlier pattern pinned one order) |
| E13 | `ls internal/statusline/quota.go internal/cli/factory_quota.go internal/cli/factory_quota_lanes.go` | three `No such file or directory` lines | 1 | the three swept quota files do not exist yet |

Each green path flips the count to at least 1 (E1-E5, E7, E8, E10-E12), the value to `3` (E6), or the file to existing (E9, E13). The mutant probe for the count rows: an implementation that only adds the word to a comment satisfies the count but cannot satisfy the Go test that decides the same AC, which is why every AC names a Go test as its decider and the ledger row only pins the starting observation. The earlier claim that the harness reports no exit status for a zero-count `grep` (the old G1) was wrong: the exit status is observable with the `; echo "rc=$?"` form and is recorded above.

### §E.2 Classification (verification-completeness §2.1)

An AC is **release-blocking** when its RED-now cell carries the four elements — a single-invocation read-only command, its verbatim stdout, its exit status, and the pinned tree SHA (the document-level pin above) — and the starting observation can be re-executed on the pre-change tree. An AC whose RED cannot be re-executed is a **regression-guard**.

| AC | Class | RED-now rows |
|---|---|---|
| AC-QAS-001, -004 | release-blocking | E1 |
| AC-QAS-002 | release-blocking | E6, E9 |
| AC-QAS-003 | release-blocking | E7 |
| AC-QAS-005 | release-blocking | E1, E13 |
| AC-QAS-006 | release-blocking | E2, E13 |
| AC-QAS-007 | release-blocking | E3, E4, E8 |
| AC-QAS-008 to -011 | release-blocking | E2 |
| AC-QAS-012 | release-blocking | E2, E9 |
| AC-QAS-013 | release-blocking | E5 |
| AC-QAS-014 | release-blocking | E13 |
| AC-QAS-015 | release-blocking | E9 |
| AC-QAS-016 | regression-guard (holds at arrival; no red to observe) | — |
| AC-QAS-017 | release-blocking | E2, E5, E10 |
| AC-QAS-018 | release-blocking | E12, E13 |
| AC-QAS-019, -020 | release-blocking | E10, E11 |
| AC-QAS-021 | release-blocking | E10, E13 |
| AC-QAS-022 | release-blocking | E9 |
| AC-QAS-023 | release-blocking | E12 |

The existing tests named as guards inside an AC (for example `TestReadsPreviousSchemaRecord`) are regression-guards by nature.

### §E.3 Continued firing (verification-completeness §1.3)

How a reader learns that a check stopped firing, per check class, without asking:

- **Go-test deciders.** CI runs the full suite on every push (`.github/workflows/ci.yml:229`, `go test -json … ./...`, uploading the per-test event stream), so each AC decider appears by name in that stream; a renamed or deleted decider is a missing name there, and the pass convention (§B) turns a selector that matches nothing into a Gap rather than a pass. The diff of the AC roster against the stream is a manual read, not an automatic one — stated rather than implied.
- **Offline / spawn-free sweep (AC-QAS-014).** It sweeps globs, not a named list, and fails below a minimum file count of three, so a vanished or renamed quota file shrinks the sweep visibly and a new `quota*.go` file joins it automatically.
- **Baseline ordering and kanban scope (AC-QAS-015, -016).** One-shot, pre-merge evaluations: they are recorded once in `progress.md` §E.2 at the close of the run phase and are vacuous after the card merges (the merge-base moves). They are not recurring checks and are not claimed to fire later.
- **Ledger rows (E1-E13).** Starting observations only, never recurring checks; their continued life is the Go test of the same AC.

## §F Edge cases

- A window present in stdin with `resets_at` already past (Claude Code normally drops it): the record carries it, the aggregator reads it as reset.
- A `rate_limits` object carrying only `spend_limit` (gateway): neither window is carried; the quota state is unknown for both.
- A percentage above 100: treated as exhausted; the hold line prints the received value.
- Two Claude sessions of different accounts writing the same project directory (`DO-2`): freshest-per-window applies; the reset times differ and `moai factory status` shows the source of each reading.
- `max_age` of `0`, an unparseable duration, a value below twice the heartbeat, or a hold percentage outside 1-100: the default applies (REQ-QAS-008); the status block reports the effective values.
- A held lane with `--wait` and a very long reset: the wait bound (15m default) ends the wait with status 3 and the stderr hold line while the reading is fresh; no new-card lease, no record change.
- A record from a GLM or Codex session that nevertheless carries a `rate_limits` object would be read as Claude quota (accepted risk, plan.md §H, U9).

## §G Quality gate and Definition of Done

- All 23 deciders pass with a non-empty swept count; the existing guards named in AC-QAS-002, -003, -007, -011, -012, -022, -023 stay green; `go vet ./internal/statusline ./internal/config ./internal/cli ./internal/kanban` and `GOOS=windows GOARCH=amd64 go build ./...` pass; the lint baseline is not worsened.
- Template-first: the template `workflow.yaml` and the local `.moai/config/sections/workflow.yaml` are edited in the same change (local enables the gate for this repository; AC-QAS-007 checks both); `make build` regenerates embedded templates before the commit.
- `moai spec lint` reports no error for this SPEC directory.
- `progress.md` §E.2 carries, per milestone, the verbatim RED (E8), the verbatim GREEN output, and the commit SHAs including the AC-QAS-015 and AC-QAS-016 evidence.
