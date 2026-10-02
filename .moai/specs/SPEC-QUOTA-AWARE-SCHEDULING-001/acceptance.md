# acceptance.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

This file is the verification layer: Given-When-Then criteria, each decided by one named command. The requirement layer (GEARS) is `spec.md` §C.

## §A Scope of verification

The criteria verify: the rate-limit windows in the session telemetry record and its schema, throttle, and exhaustion time; the aggregator's freshness, rollover, and fail-open rules; the `workflow.quota_gate` configuration and its template mirror; the Claude-lane hold at `moai factory next` (CLI and MCP) with its hold line and in-wait latch; the unaffected non-Claude lanes and unaffected `stage`/`complete`; the `moai factory status` quota block; the warn-only line at `moai integration acquire`; the offline/read-only property; the baseline-first commit ordering; the untouched kanban paths, and (leader verdict, 0.3.0) the steering: one shared pressure evaluation, the lane inventory, the `moai todo --auto` and `moai factory status` recommendation and no-candidate warning, no dispatch side effect, and unchanged output when pressure is off. Not verified here: kanban mode, Codex-lane slot behaviour, any real Claude Code runtime or real quota (no test starts a real `claude` or `codex`).

## §B Test-environment constraint (binds every AC)

- Every test builds its own project root with `t.TempDir()`; the MoAI home is redirected to a temporary directory. No test reads or writes the real `~/.moai` or the real `.moai/state`.
- Time is an injected clock (`factoryCardNow` exists for the factory seam, `factoryNextWaitSleep` for the wait loop); no test sleeps or reads the wall clock for a verdict. Quota records are fixture files written into a temporary `context-usage/` directory; no test calls the network.
- Lane environment in the CLI tests means the factory role marker set to the role-value constant, a lane label, and the backend variable under test; the gate predicate reads the launch-provider variable first and the kanban backend variable second (`DO-10`).
- `./internal/cli` runs only with an anchored `-run` selector naming the AC tests (`-run '^(TestQAS_AC008_...|TestQAS_AC008b_...)$'`); a whole-package run of `internal/cli` is not a verdict here (it measures the machine, not the change). `./internal/statusline` and `./internal/config` may run whole-package once at the end of M4.
- **Pass convention:** an AC passes only when every command it names exits 0 and each test's verbose output carries a `--- PASS: <name> ` line for exactly the test the command names (name followed by a space, never a prefix match). `[no tests to run]`, `no tests ran`, or a zero swept count is a Gap, never a pass (`verification-completeness.md` §1.1). Subtests are decided by their parent's PASS line plus a `-v` listing of the subtest names the AC enumerates.
- Defaults the tests assert come from the SPEC: hold 90 (five-hour) and 95 (seven-day), release margin 5, max age 30m, heartbeat 5m, exhaustion 100, hold exit status 3 (the no-card status, DO-5) — the numbers are the resolved defaults of `spec.md` §B; the thresholds, margin, and age are unmeasured defaults (DO-3, accepted as provisional by the leader) and change here and in the tests together when the leader sets other values.

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

## §D Acceptance criteria

Each entry ends with its **RED-now** cell (the pre-change observation, ledger row id in §E) and its **green path** cell (the milestone that flips it and what the passing output becomes). The structural gaps are observed now (§E); the behaviour decided only by a not-yet-written test records its verbatim RED at the start of the milestone (E8 in `progress.md` §E.2), because the test does not exist to run today — that is stated here rather than implied.

### AC-QAS-001 — the record carries exactly the supplied windows

- **Given** a statusline stdin payload with a five-hour window (62.5%, reset epoch 1790000000) and a seven-day window (41.2%, reset epoch 1790500000), **when** the builder runs against a temporary project directory, **then** the session record decodes with both windows carrying exactly those percentages and reset times; with a payload carrying only the seven-day window the five-hour window is absent from the record; with a payload carrying no `rate_limits` the record carries no window.
- Decider: `go test ./internal/statusline -run '^TestQAS_AC001_RecordCarriesSuppliedWindowsOnly$' -count=1 -v` → `--- PASS: TestQAS_AC001_RecordCarriesSuppliedWindowsOnly ` and subtests `both_windows`, `seven_day_only`, `no_rate_limits`.
- RED-now: ledger E1 (`context_usage.go` mentions `rate_limits` 0 times). Green path: M1.

### AC-QAS-002 — an old-schema record reads as "no windows"; a window-less record keeps its bytes

- **Given** record fixtures at schema version 1 and 2 (no window fields) and a window-less record built from a fixed input, **when** they are read and serialized, **then** the old fixtures decode without error and report no windows, and the window-less serialization is byte-identical to the baseline golden after normalizing only the capture time, the writer pid, and the schema version value.
- Decider: `go test ./internal/statusline -run '^(TestQAS_AC002_PreviousSchemaReadsAsNoWindows|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline)$' -count=1 -v` → two PASS lines; regression guard `go test ./internal/statusline -run '^TestReadsPreviousSchemaRecord$' -count=1 -v` → PASS (`session_telemetry_payload_test.go:121`).
- RED-now: ledger E6 (schema constant is `2`, the AC expects `3`). Green path: M1. Depends on the baseline golden `internal/statusline/testdata/qas_baseline_windowless_record.golden.json` measured before any implementation (AC-QAS-015).

### AC-QAS-003 — throttle: bucket, heartbeat, window drop, window-less unchanged

- **Given** an on-disk record carrying a five-hour window at 62.2% and an injected clock, **when** the statusline renders again, **then**: (a) 62.9% within the heartbeat (same integer bucket, same reset time) writes nothing (file bytes and modification time unchanged); (b) 63.0% writes; (c) an unchanged reading older than the heartbeat rewrites with a refreshed capture time; (d) the window disappearing from stdin writes a record without it; (e) a record that carries no window and is unchanged beyond the heartbeat is NOT rewritten (throttled exactly as before).
- Decider: `go test ./internal/statusline -run '^TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop$' -count=1 -v` → PASS with subtests (a)-(e) listed; positive controls `go test ./internal/statusline -run '^(TestWriteContextUsage_ThrottleSkipUnchanged|TestThrottleUnaffectedByModelAndEffort)$' -count=1 -v` → two PASS lines.
- RED-now: ledger E7 (`context_usage.go` mentions `heartbeat` 0 times). Green path: M1.

### AC-QAS-004 — first-observed-exhausted time is sticky until the window rolls

- **Given** a window first observed at 100% at time T with reset R, **when** later writes carry the window at 100% with reset R, **then** the exhausted time stays T; **when** the reset becomes R2, or the window is no longer carried, **then** the exhausted time is dropped; a window first seen at 99% carries no exhausted time.
- Decider: `go test ./internal/statusline -run '^TestQAS_AC004_ExhaustedAtStickyUntilRollover$' -count=1 -v` → PASS with the four subtests.
- RED-now: ledger E1. Green path: M1.

### AC-QAS-005 — freshest record per window; a reset window is reset, not stale-high

- **Given** three records in a temporary `context-usage/` directory (a fresh one carrying both windows, a fresher one carrying only the seven-day window, a 31-minute-old one at 99%), **when** the aggregator runs, **then** the five-hour reading comes from the fresh both-windows record, the seven-day reading from the fresher record, and the 31-minute-old record contributes nothing; **given** a fresh record whose window reset time is one second before now at 99%, **then** that window's state is reset (not hold, not stale).
- Decider: `go test ./internal/statusline -run '^TestQAS_AC005_AggregateFreshestPerWindowAndRollover$' -count=1 -v` → PASS with subtests `freshest_per_window`, `stale_contributes_nothing`, `reset_window_is_reset`.
- RED-now: ledger E1 (no aggregator, no window in the record). Green path: M2.

### AC-QAS-006 — absent or unreadable data fails open

- **Given** a missing record directory, an unparseable record file, only stale records, and an unparseable `workflow.yaml`, **when** the aggregator and then `moai factory next` for a Claude lane with the gate enabled and a leasable queued card run, **then** every window reads unknown and the lane leases the card exactly as with the gate disabled.
- Decider: `go test ./internal/statusline -run '^TestQAS_AC006_FailOpenOnAbsentOrUnreadable$' -count=1 -v` → PASS (four subtests); `go test ./internal/cli -run '^TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent$' -count=1 -v` → PASS (the same four fixtures, a leased card in each).
- RED-now: ledger E2 (`factory_card.go` mentions quota 0 times — no gate exists to fail open). Green path: M2 (aggregator) and M3 (lane).

### AC-QAS-007 — the configuration keys exist, with defaults mirrored in the template

- **Given** the Go defaults, the shipped template `workflow.yaml`, the shipped-key inventory, and the config cache version, **when** the template is decoded into the workflow config, **then** `workflow.quota_gate` carries `enabled: false`, `five_hour_hold_pct: 90`, `seven_day_hold_pct: 95`, `release_margin_pct: 5`, `max_age: 30m`, each equal to the Go default; the template contains no `enabled: true` for the gate; an absent key, an absent file, an unparseable file, and a value outside its valid range (a hold percentage outside 1-100, a release margin outside 0-50, a non-positive or unparseable duration) each yield the default; the five keys are classified in the inventory; and the cache schema version is greater than 10.
- Decider: `go test ./internal/config -run '^TestQAS_AC007_ConfigDefaultsMirrorTemplate$' -count=1 -v` → PASS (subtests `defaults_equal_template`, `template_ships_off`, `absent_or_unparseable_yields_default`, `out_of_range_yields_default`, `cache_schema_bumped`); `go test ./internal/config -run '^TestShippedConfigKeysHaveReaders$' -count=1 -v` → PASS (the inventory classification, `shipped_key_reader_test.go:74`); the positive control inside the new test decodes the sibling `slot_lease.default_max_duration` from the same template (the pattern of `workflow_jev_test.go:84-88`).
- RED-now: ledger E3, E4, E8 (`quota_gate` appears 0 times in the template, `types.go`, and the inventory). Green path: M2.

### AC-QAS-008 — a held Claude lane leases no new card but still receives its own assigned card

- **Given** the gate enabled, a lane environment whose launch provider is Claude, a queued card, a picked-unassigned card, a picked card with no record row, and a fresh record with the five-hour window at 92% (reset in the future), **when** `moai factory next` runs (and, separately, the `factory_next` MCP handler), **then** none of those three cards is leased, the factory record and the queue are byte-identical before and after, one hold line is printed on the error stream with standard output empty, and the CLI exits with status 3 (the no-card status); **given** the same state plus a card assigned to this lane (arm (a)), **then** that card is leased and returned exactly as with the gate disabled and no hold line is printed; a window at 89.9% (below 90) leases normally; a seven-day window at 95% holds.
- Decider: `go test ./internal/cli -run '^(TestQAS_AC008_ClaudeLaneHeldAtThreshold|TestQAS_AC008b_MCPFactoryNextHeld)$' -count=1 -v` → two PASS lines; subtests `five_hour_92_new_cards_skipped`, `five_hour_92_assigned_card_still_leased`, `five_hour_89_9`, `seven_day_95`.
- RED-now: ledger E2. Green path: M3.

### AC-QAS-009 — the hold line carries the reset time

- **Given** the hold fixtures of AC-QAS-008, **when** the hold line is captured, **then** it is exactly one line on the error stream (standard output carries nothing, whereas a true empty queue prints `no card is available` on standard output and no `quota hold:` line, both with status 3), matches `^quota hold: five_hour used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$` for one held window, carries one such segment per held window separated by `; ` when both are held, and the instant equals the fixture's reset epoch rendered in UTC.
- Decider: `go test ./internal/cli -run '^TestQAS_AC009_HoldLineCarriesResetTime$' -count=1 -v` → PASS.
- RED-now: ledger E2. Green path: M3.

### AC-QAS-010 — the wait latch releases only below the margin or at the reset

- **Given** a held Claude lane in `moai factory next --wait` and an injected sleep that advances fixtures, **when** the reading goes 92% → 88% → 84%, **then** the verb stays held at 88% (release needs below 85%) and leases at 84%; **when** the reading stays at 92% and the reset time passes, **then** the verb leases once the window reads reset; **when** the wait bound elapses while held, **then** it exits with status 3 and the stderr hold line; **when** a card is assigned to the lane while the verb waits held, **then** the next re-check leases that card.
- Decider: `go test ./internal/cli -run '^TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrReset$' -count=1 -v` → PASS (subtests `margin_release`, `reset_release`, `bound_elapsed`, `assigned_card_leased_during_wait`).
- RED-now: ledger E2. Green path: M3.

### AC-QAS-011 — non-Claude lanes are never held; stage and complete are never gated

- **Given** a 99% fresh five-hour record and the gate enabled, **when** `moai factory next` runs with the lane's backend set to gpt, glm, empty, and an unrecognised value, **then** each leases the card exactly as with the gate disabled; **given** a held Claude lane with a card already leased, **when** `moai factory stage` and `moai factory complete` run for that card, **then** both behave as before the change.
- Decider: `go test ./internal/cli -run '^(TestQAS_AC011_NonClaudeBackendsNeverHeld|TestQAS_AC011b_StageAndCompleteIgnoreQuotaHold)$' -count=1 -v` → two PASS lines; regression guards `go test ./internal/cli -run '^(TestSD_AC023_CodexNextSkipsUnadvanceableCard|TestSD_AC008_NextSelectionOrderAndOutput)$' -count=1 -v` → two PASS lines (Codex skip and selection order intact).
- RED-now: ledger E2. Green path: M3.

### AC-QAS-012 — `moai factory status` reports the quota block and changes nothing else

- **Given** fixture records and a factory record, **when** `moai factory status` runs in text and with `--json`, **then** the text carries a quota line per window with used percentage, reset time, source capture time, state (fresh, reset, or unknown), and held/not-held; the JSON carries a `quota` key with the same fields; with no window data anywhere the JSON has no `quota` key and the output is byte-identical to the baseline golden `internal/cli/testdata/qas_baseline_factory_status.golden.json`.
- Decider: `go test ./internal/cli -run '^(TestQAS_AC012_StatusReportsQuotaBlock|TestFactoryStatusShowsHolderModePriority)$' -count=1 -v` → two PASS lines (the second is the existing status guard, `factory_classify_test.go:392`).
- RED-now: ledger E2. Green path: M4. Depends on the baseline golden (AC-QAS-015).

### AC-QAS-013 — `moai integration acquire` warns and never refuses

- **Given** the gate enabled, a Claude lane environment, and a fresh window at or above its hold percentage, **when** `moai integration acquire` runs (text and `--json`), **then** the error stream carries one line naming the window and its reset time, and the lock record, the exit status, and standard output (the `--json` object included) are identical to a run with the gate disabled; with a window below its threshold, or with a non-Claude backend, the error stream carries no such line.
- Decider: `go test ./internal/cli -run '^TestQAS_AC013_AcquireWarnsNeverRefuses$' -count=1 -v` → PASS.
- RED-now: ledger E5. Green path: M4.

### AC-QAS-014 — the quota path is offline, spawn-free, and read-only

- **Given** the new aggregator and gate source files, **when** their imports are parsed and the aggregator runs on a temporary directory snapshot, **then** none imports `net`, `net/http`, or `os/exec`, and the directory tree (names, sizes, bytes) is identical before and after the run.
- Decider: `go test ./internal/statusline ./internal/cli -run '^TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly$' -count=1 -v` → PASS in each package that defines it; the swept count is the three files named in `plan.md` §D (`internal/statusline/quota.go`, `internal/cli/factory_quota.go`, `internal/cli/factory_quota_lanes.go`), and the steering code is included in the no-network, no-spawn, no-write sweep.
- RED-now: ledger E1 (the files do not exist yet). Green path: M2 and M3.

### AC-QAS-015 — the baseline lands before the implementation

- **Given** the run-phase commit graph, **when** the first commit adding the three baseline goldens (`internal/statusline/testdata/qas_baseline_windowless_record.golden.json`, `internal/cli/testdata/qas_baseline_factory_status.golden.json`, `internal/cli/testdata/qas_baseline_todo_auto.golden.txt`) is B and the first commit adding `internal/statusline/quota.go` is I, **then** B is an ancestor of I and distinct from it.
- Decider (plain commands, recorded verbatim in `progress.md` §E.2): `git log --reverse --format=%H -- internal/statusline/testdata/qas_baseline_windowless_record.golden.json internal/cli/testdata/qas_baseline_factory_status.golden.json internal/cli/testdata/qas_baseline_todo_auto.golden.txt` (first line = B); `git log --reverse --format=%H -- internal/statusline/quota.go` (first line = I); `git rev-list --count B..I` prints a number of at least 1 and `git rev-list --count I..B` prints `0`.
- RED-now: ledger E9 (no golden and no `testdata` directory exist on the pre-change statusline tree). Green path: the M0 commit adds both goldens, then every later milestone commit follows it. Why a commit-graph criterion: a baseline committed together with the implementation it measured leaves the ordering unverifiable (`verification-claim-integrity.md` §2.3).

### AC-QAS-016 — no kanban path is touched

- **Given** the card branch after its last implementation commit and `CARD_BASE` read at that moment as `git merge-base develop HEAD`, **when** the changed-file list is taken, **then** it names no file under `internal/kanban/` and neither `internal/cli/kanban.go` nor `internal/cli/kanban_settings.go`; the positive control (the same range without the pathspec) is non-empty.
- Decider (plain commands): `git merge-base develop HEAD` (= `CARD_BASE`); `git diff --name-only <CARD_BASE>..HEAD -- internal/kanban internal/cli/kanban.go internal/cli/kanban_settings.go` prints nothing; `git diff --name-only <CARD_BASE>..HEAD` prints at least the SPEC and source files (an empty control reports "unmeasurable", never "no change"). Pre-merge evaluation only (`gitflow-lane-protocol.md` §8).
- RED-now: not applicable as a red (a guard that holds at arrival); its mutant is the probe: touching `internal/cli/kanban.go` must make the first command print that path. Green path: every milestone.

### AC-QAS-017 — one pressure evaluation, four surfaces

- **Given** the gate enabled and a fixture with the five-hour window at 92% (fresh, reset in the future), **when** the lane gate (as a Claude lane), the `moai factory status` block, the `moai todo --auto` recommendation, and the `moai integration acquire` warning are each evaluated, **then** all four report pressure; with the window at 89.9%, reset, unknown, or the gate disabled, none reports pressure; the lane gate holds only a Claude lane while the other three report pressure for any caller backend.
- Decider: `go test ./internal/cli -run '^TestQAS_AC017_FourSurfacesAgreeOnPressure$' -count=1 -v` → PASS (subtests `at_92`, `at_89_9`, `reset`, `unknown`, `gate_disabled`, `caller_backend_independent`).
- RED-now: ledger E2, E5, E10. Green path: M3 (the shared evaluation) and M4-M5 (the surfaces).

### AC-QAS-018 — the lane inventory lists live non-Claude lanes only

- **Given** a temporary factory registry and kanban session records with: lane-1 alive, backend claude; lane-2 alive, backend glm; lane-3 alive, backend gpt; lane-4 registered with a dead pid, backend glm; lane-5 alive with no session record; lane-6 alive with two records of the same newest instant that disagree on backend; and, separately, an unreadable registry and an unreadable record directory, **when** the inventory is read, **then** the candidates are lane-2 (glm) and lane-3 (gpt), the unknown count is 2 (lane-5 and lane-6), a lane whose newest record is newer than an older record of a different backend takes the newest, and an unreadable registry or record store yields no candidates; the registry and record files are byte-identical before and after.
- Decider: `go test ./internal/cli -run '^TestQAS_AC018_LaneInventoryCandidates$' -count=1 -v` → PASS (subtests `candidates_and_unknown`, `newest_record_wins`, `unreadable_yields_none`, `read_only`).
- RED-now: ledger E10. Green path: M4.

### AC-QAS-019 — pressure on: `--auto` and `moai factory status` recommend the candidate lanes

- **Given** the gate enabled, pressure on (five-hour 92%), the inventory of AC-QAS-018, a two-card queue and the Jev line stubbed, **when** `moai todo --auto` runs one cycle and `moai factory status` runs in text and `--json`, **then** the `--auto` output carries, immediately before each `accept` line, exactly one line matching `^quota pressure: five_hour used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z; recommend non-Claude lane\(s\): lane-2 \(glm\), lane-3 \(gpt\)$`; the status text carries the same line under the quota block; the status JSON `quota` key carries `pressure: true`, the two candidates with label and backend, and `unknown_lanes: 2`.
- Decider: `go test ./internal/cli -run '^TestQAS_AC019_AutoAndStatusRecommendNonClaudeLanes$' -count=1 -v` → PASS (subtests `auto_line_before_each_accept`, `status_text`, `status_json`).
- RED-now: ledger E10, E11. Green path: M4.

### AC-QAS-020 — pressure on and no candidate lane: a warning, nothing else

- **Given** pressure on and an inventory with only claude lanes, only dead lanes, or no registered lane, **when** `moai todo --auto` and `moai factory status` run, **then** each prints one line matching `^quota pressure: five_hour used=[0-9]+\.[0-9]% resets_at=[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z; warning: no live non-Claude lane \(unknown backend: [0-9]+\); nothing is re-dispatched$` in place of the recommendation line, the status JSON carries `warning: "no-non-claude-lane"` and an empty candidate list, and every other line of the output (accept, directive, evidence, done) is identical to the pressure-off output of the same fixture.
- Decider: `go test ./internal/cli -run '^TestQAS_AC020_NoNonClaudeLaneWarnsOnly$' -count=1 -v` → PASS (subtests `only_claude_lanes`, `only_dead_lanes`, `no_lanes`, `rest_of_output_unchanged`).
- RED-now: ledger E10, E11. Green path: M4.

### AC-QAS-021 — the steering changes nothing

- **Given** pressure on, a queue, a factory record, a lane registry, and a session registry, **when** the recommendation is produced through `moai factory status` and through the `--auto` cycle, **then** the set of queue mutations the `--auto` cycle performs is identical to the pressure-off run of the same fixture (its own pick and done only), the factory record, the lane registry, and the session registry are byte-identical before and after, no card changes state or owner, and the steering code path starts no process, opens no network connection, and sends no message (static import check on the steering files, shared with AC-QAS-014).
- Decider: `go test ./internal/cli -run '^TestQAS_AC021_SteeringChangesNothing$' -count=1 -v` → PASS (subtests `status_is_read_only`, `auto_mutations_equal_pressure_off`, `no_card_class_read`).
- RED-now: ledger E10 (no steering code exists to compare). Green path: M4.

### AC-QAS-022 — pressure off: the output is unchanged

- **Given** the baseline goldens measured before the change (`qas_baseline_factory_status.golden.json`, `qas_baseline_todo_auto.golden.txt`) and a fixture with the gate disabled, then enabled with every window below its hold percentage, reset, or unknown, **when** `moai todo --auto` (Jev line stubbed) and `moai factory status` run, **then** each output is byte-identical to its golden.
- Decider: `go test ./internal/cli -run '^TestQAS_AC022_PressureOffOutputUnchanged$' -count=1 -v` → PASS (subtests `status_gate_disabled`, `status_below_threshold`, `auto_gate_disabled`, `auto_below_threshold`, `auto_unknown_windows`); existing guards `go test ./internal/cli -run '^(TestFactoryStatusShowsHolderModePriority|TestTodoAutoSerialCycle|TestTodoAutoCreatesNoFactoryLease|TestTodoAutoClearGuidancePerCard)$' -count=1 -v` → four PASS lines (`todo_auto_test.go:247,395,339`, `factory_classify_test.go:392`).
- RED-now: baseline goldens absent (ledger E9); the existing `--auto` tests are the pre-change measurement. Green path: M0 (goldens), M4.
- Depends on the baseline goldens (AC-QAS-015).

## §E RED-now evidence ledger

Measured in this plan phase on tree `c50da9c2f` (the card branch tip and the develop tip at the time; the same commit the installed `moai` build reports, `moai version` → `moai_cp/20260925_122548-1896-gc50da9c2f`). Each row is a single read-only invocation; the stdout is verbatim. The exit-status field is a named Gap (G1): the tool harness reports no exit status for a zero-count `grep` (it printed "completed with no output" for `grep`, `git grep -q`, and `test -e` that all select nothing), so the stdout count is the deciding signal and the exit status is not recorded rather than inferred.

| Row | Command | Stdout | Exit status | Meaning |
|---|---|---|---|---|
| E1 | `grep -c "rate_limits" internal/statusline/context_usage.go` | `0` | not observed (G1) | the record has no window today |
| E2 | `grep -c -i "quota" internal/cli/factory_card.go` | `0` | not observed (G1) | the lease seam has no quota logic |
| E3 | `grep -c "quota_gate" internal/template/templates/.moai/config/sections/workflow.yaml` | `0` | not observed (G1) | the template carries no gate key |
| E4 | `grep -c "quota_gate" internal/config/types.go` | `0` | not observed (G1) | the Go config carries no gate key |
| E5 | `grep -c -i "quota" internal/cli/integration.go` | `0` | not observed (G1) | acquire has no quota line |
| E6 | `grep -n "contextUsageSchemaVersion = " internal/statusline/context_usage.go` | `32:const contextUsageSchemaVersion = 2` | not observed (G1) | the schema version is 2; the AC expects 3 |
| E7 | `grep -c -i "heartbeat" internal/statusline/context_usage.go` | `0` | not observed (G1) | the throttle has no heartbeat |
| E8 | `grep -c "quota_gate" internal/config/testdata/shipped_key_inventory.yaml` | `0` | not observed (G1) | the inventory has no gate entries |
| E9 | `ls internal/statusline/testdata` | `ls: internal/statusline/testdata: No such file or directory` | 1 (observed; the harness printed `Exit code 1`) | the statusline baseline golden's directory does not exist yet; the golden is created in M0. The CLI side: `ls internal/cli/testdata` lists 36 entries, none named `qas_*` |
| E10 | `grep -c -i "quota" internal/cli/todo_auto.go` | `0` | not observed (G1) | the `--auto` cycle has no quota or lane logic |
| E11 | `grep -c -i "quota" internal/cli/todo.go` | `0` | not observed (G1) | the `--auto` wiring has no quota seam |

Each green path flips the count to at least 1 (E1-E5, E7-E8, E10-E11) or the value to `3` (E6). The mutant probe for the count rows: an implementation that only adds the word to a comment satisfies the count but cannot satisfy the Go test that decides the same AC, which is why every AC names a Go test as its decider and the ledger row only pins the starting observation.

## §F Edge cases

- A window present in stdin with `resets_at` already past (Claude Code normally drops it): the record carries it, the aggregator reads it as reset.
- A `rate_limits` object carrying only `spend_limit` (gateway): neither window is carried; the quota state is unknown for both.
- A percentage above 100: treated as exhausted; the hold line prints the received value.
- Two Claude sessions of different accounts writing the same project directory (`DO-2`): freshest-per-window applies; the reset times differ and `moai factory status` shows the source of each reading.
- `max_age` of `0` or an unparseable duration, or a hold percentage outside 1-100: the default applies (REQ-QAS-008); the status block reports the effective values.
- A held lane with `--wait` and a very long reset: the wait bound (15m default) ends the wait with status 3 and the stderr hold line; no new-card lease, no record change.

## §G Quality gate and Definition of Done

- All 22 deciders pass with a non-empty swept count; the existing guards named in AC-QAS-002, -003, -007, -011, -012, -022 stay green; `go vet ./internal/statusline ./internal/config ./internal/cli` and `GOOS=windows GOARCH=amd64 go build ./...` pass; the lint baseline is not worsened.
- Template-first: the template `workflow.yaml` and the local `.moai/config/sections/workflow.yaml` are edited in the same change (local enables the gate for this repository); `make build` regenerates embedded templates before the commit.
- `moai spec lint` reports no error for this SPEC directory.
- `progress.md` §E.2 carries, per milestone, the verbatim RED (E8), the verbatim GREEN output, and the commit SHAs including the AC-QAS-015 ordering evidence.
