# SPEC-HARNESS-RETENTION-HARDEN-001 — Acceptance Criteria (card t1432)

Verification layer for `spec.md` §C. Every criterion is `Given … When … Then …` and binary-testable. Test names below are handles for the run phase, not requirements on identifiers. "RED anchor" is the observed failure that must be recorded in `.moai/reports/t1432/red-baseline.md` in a commit that precedes the fix commit (`verification-claim-integrity.md` §2.3); "pin" means the criterion is green at baseline and guards existing behaviour. Test commands run env-scrubbed in one invocation, e.g. `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -run '^<Name>$' ./internal/harness/`. Permission-based tests skip when the effective uid is 0 and on Windows; FIFO-based tests skip on Windows.

## Traceability

| AC | REQ | Item | Premise label | RED anchor |
|---|---|---|---|---|
| AC-HRH-001 | REQ-HRH-001 | F4 | OBSERVED | `PROBE-F4 ... victim_changed=true`; test fails on `1e2151a38` |
| AC-HRH-002 | REQ-HRH-001 | F4 | OBSERVED | same probe; state path still a symlink after the prune on baseline |
| AC-HRH-003 | REQ-HRH-002 | F7 | OBSERVED | `retention: prune state open failed: ... permission denied` on both calls (probe), `stale_still_in_log=true` |
| AC-HRH-004 | REQ-HRH-003 | F7 boundary | READ | pin (green at baseline); mutation check in M0 |
| AC-HRH-005 | REQ-HRH-004 | event loss | OBSERVED | `PROBE-LOSS late_event_present=false fresh_present=true` |
| AC-HRH-006 | REQ-HRH-004 | event loss | READ | pin; boundary-rule case fails on a naive tail copy |
| AC-HRH-007 | REQ-HRH-005 | event loss | READ | pin (constraint check) |
| AC-HRH-008 | REQ-HRH-006 | event loss | READ | baseline grep: `residual window` 0, old sentence 1 |
| AC-HRH-009 | REQ-HRH-007 | F5 | not reproduced, not measured | baseline grep: `not reproduced, not measured` 0 |
| AC-HRH-010 | REQ-HRH-008 | F6 | not reproduced, not measured | baseline grep: `lock waiters block` 0, `5 s hook timeout` 0 |
| AC-HRH-011 | REQ-HRH-009 | F5, F6 | READ | pin (constraint check) |
| AC-HRH-012 | REQ-HRH-010 | N1 | AUDIT-REPORTED, re-observed in M0 | mutant mD: `panic: test timed out after 40s` |
| AC-HRH-013 | REQ-HRH-011 | N2 / mB | AUDIT-REPORTED, re-observed in M0 | mutant mB survives the existing suite (`ok`) |
| AC-HRH-014 | REQ-HRH-012 | N2 / mC | AUDIT-REPORTED, re-observed in M0 | mutant mC survives the existing suite (`ok`) |
| AC-HRH-015 | REQ-HRH-001..004, -010 | all | READ | pin (existing suite green, unmodified) |

## Acceptance scenarios

### AC-HRH-001 — A symlinked state path never reaches its target (REQ-HRH-001)

- **Given** a log with one stale event, a 64-byte victim file, and a symbolic link at `<log>.prune-state` pointing at the victim,
- **When** `PruneStaleEntries(30)` runs,
- **Then** the victim file's bytes are identical to before the call.
- Verify: `go test -count=1 -run '^TestPruneStateSymlinkTargetUntouched$' ./internal/harness/` exits 0. RED: the same test exits non-zero on the unmodified code, output recorded in M0.

### AC-HRH-002 — The prune continues with a regular state file (REQ-HRH-001)

- **Given** the setup of AC-HRH-001,
- **When** `PruneStaleEntries(30)` runs,
- **Then** the call returns nil, the stale event is gone from the log and present in the archive, the path `<log>.prune-state` is a regular file (not a symlink) whose content parses as a fresh stamp, and the victim is unchanged.
- Verify: `go test -count=1 -run '^TestPruneStateSymlinkReplacedAndPruneCompletes$' ./internal/harness/`. RED anchor as AC-HRH-001.

### AC-HRH-003 — An unwritable state file is replaced and retention continues (REQ-HRH-002)

- **Given** a log with one stale event and a pre-existing state file of mode 0400 owned by the current user in a writable directory,
- **When** the first `PruneStaleEntries(30)` runs, and then a second instance runs one day later against a log that has gained a new stale event,
- **Then** both calls return nil, both stale events are archived and gone from the log, and the state file is a regular file the current user can open read-write.
- Verify: `go test -count=1 -run '^TestPruneStateUnwritableFileReplaced$' ./internal/harness/`. RED: on baseline both calls return the `permission denied` open error and the stale event stays in the log (probe output recorded in M0).

### AC-HRH-004 — An unreplaceable state path still skips safely (REQ-HRH-003, pin)

- **Given** (a) the state path is a directory, and (b) the state file is mode 0400 inside a directory made non-writable (mode 0555),
- **When** `PruneStaleEntries(30)` runs in each case,
- **Then** each call returns a non-nil error, the log is byte-identical to before, no archive directory exists, and `RecordEvent` through an observer carrying that retention still succeeds.
- Verify: the existing `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` (case a) stays green unmodified, plus one new case for (b). Pin: green at baseline; M0 records the baseline result and a mutation run in which the heal removes a directory is shown to fail case (a).

### AC-HRH-005 — A late event survives the prune (REQ-HRH-004)

- **Given** a log with one stale and one fresh event and the archive path made a FIFO so the pruner blocks in its archive step after it has read the log,
- **When** a second writer appends `late-event` through the append helper while the pruner is blocked, and the pruner is then released,
- **Then** the pruner returns nil, the log contains `fresh` and then `late-event` (in that order, `late-event` verbatim), does not contain the stale event, and the archive contains the stale event.
- Verify: `go test -count=1 -run '^TestPruneCarriesLateEvents$' ./internal/harness/` (non-Windows). RED: baseline output `late_event_present=false` (probe), recorded in M0.

### AC-HRH-006 — Output without late events is unchanged; the tail boundary is a line boundary (REQ-HRH-004, pin plus new case)

- **Given** (a) a log nothing is appended to during the prune, and (b) a log whose appended tail begins with a complete line and ends without a trailing newline at the moment of the prefix measurement,
- **When** the prune runs,
- **Then** in (a) the replacement log equals what the baseline pruner writes for the same input (the existing `TestPruneKeepsUnparsedLinesVerbatim` and `TestPruneNothingStaleLeavesLogUntouched` pass unmodified), and in (b) the unterminated line appears exactly once, whole, in the replacement.
- Verify: `go test -count=1 -run '^(TestPruneKeepsUnparsedLinesVerbatim|TestPruneNothingStaleLeavesLogUntouched|TestPruneTailBoundaryIsALineBoundary)$' ./internal/harness/`.

### AC-HRH-007 — The append path gains nothing (REQ-HRH-005, constraint)

- **Given** the final tree,
- **When** the observer source is compared to the base,
- **Then** `internal/harness/observer.go` is byte-identical to `1e2151a38`.
- Verify: `git diff --quiet 1e2151a38 -- internal/harness/observer.go` exits 0 (displayed exit code in `verdict.md`).

### AC-HRH-008 — The residual window is disclosed, the false claim removed (REQ-HRH-006)

- **Given** the final `internal/harness/retention.go`,
- **When** it is searched for the sentinel phrases,
- **Then** `residual window` occurs at least once and `events other hooks append in that window are lost` occurs 0 times.
- Verify: `grep -c -F "residual window" internal/harness/retention.go` prints a number ≥ 1 and `grep -c -F "events other hooks append in that window are lost" internal/harness/retention.go` prints 0. RED: baseline prints 0 and 1 (observed at plan authoring; re-observed M0).

### AC-HRH-009 — F5 is disclosed with its evidence label (REQ-HRH-007)

- **Given** the final `retention.go`,
- **Then** the pruner's source documentation contains `not reproduced, not measured` in the sentence that states the Windows limitation (no cross-process exclusion; a burst finding no fresh stamp may prune concurrently once per interval).
- Verify: `grep -n -F "not reproduced, not measured" internal/harness/retention.go` lists at least two lines (one for F5, one for F6, AC-HRH-010) and the F5 line also contains the word `Windows`; reviewer reads the sentence. Verification level: documentation (no Windows runtime). RED: baseline count 0.

### AC-HRH-010 — F6 is recorded with the 5 s fact (REQ-HRH-008)

- **Given** the final `retention.go`,
- **Then** it contains `lock waiters block`, `5 s hook timeout`, and a statement that the event is appended before the wait begins and that the earlier design question did not carry the hook-timeout fact, with the waiter delay labelled `not reproduced, not measured`.
- Verify: `grep -c -F "lock waiters block" internal/harness/retention.go` ≥ 1 and `grep -c -F "5 s hook timeout" internal/harness/retention.go` ≥ 1. RED: baseline prints 0 and 0.

### AC-HRH-011 — The shared lock package is untouched and Windows still builds (REQ-HRH-009, constraint)

- **Given** the final tree,
- **Then** every file under `internal/lockfile` is byte-identical to the base, and `internal/harness` and `internal/lockfile` build and vet for Windows.
- Verify: `git diff --quiet 1e2151a38 -- internal/lockfile` exits 0; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exits 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exits 0. Verification level: build and vet only (no Windows runtime; the CI release matrix compiles the Windows target, which is the only further evidence available).

### AC-HRH-012 — A pruner that never archives fails the killed-pruner test quickly (REQ-HRH-010)

- **Given** mutant mD (the pruner stamps and then returns without pruning) applied as a temporary overlay,
- **When** `TestPruneStamp_StampExistsBeforeTheWork` runs with `-timeout 40s`,
- **Then** it exits non-zero in at most 15 seconds wall time, the output names the missing archive step and does not contain `panic: test timed out`; and with no mutation the same test passes.
- Verify: wall time of `go test -count=1 -timeout 40s -run '^TestPruneStamp_StampExistsBeforeTheWork$' ./internal/harness/` under the overlay (measured with `time` or an outer `timeout 30`), plus the unmutated run exiting 0. RED: baseline under mD ends in `panic: test timed out after 40s` (AUDIT-REPORTED; re-observed in M0 and pasted).

### AC-HRH-013 — Mutant mB (no truncation of a longer stamp) is killed (REQ-HRH-011)

- **Given** a state file holding an older stamp with nine fractional digits and a prune whose clock formats with fewer bytes,
- **When** the prune records its stamp,
- **Then** the state file content equals exactly the new stamp (no trailing bytes) and parses as fresh; with `Truncate(0)` removed from `writeStamp` the test fails.
- Verify: `go test -count=1 -run '^TestPruneStampShorterOverLongerIsExact$' ./internal/harness/` exits 0 unmutated and non-zero under mB. RED: mB survives the existing suite (`ok`), recorded in M0.

### AC-HRH-014 — Mutant mC (stamp-write error ignored) is killed (REQ-HRH-012)

- **Given** a pruner whose stamp writer is replaced by one that always fails,
- **When** `PruneStaleEntries(30)` runs on a log with a stale event,
- **Then** it returns a non-nil error, the log is byte-identical, and no archive directory is created; with the stamp-write error ignored in `pruneExclusive` the test fails.
- Verify: `go test -count=1 -run '^TestPruneStampWriteFailureSkipsPrune$' ./internal/harness/` exits 0 unmutated and non-zero under mC. RED: mC survives the existing suite (`ok`), recorded in M0. If D5 option B is selected this AC is withdrawn by amendment.

### AC-HRH-015 — The existing retention suite stays green and unedited, apart from N1 (REQ-HRH-001..004, -010)

- **Given** the final tree,
- **Then** every pre-existing `TestPrune*` function in `retention_*_test.go` passes, and no pre-existing test file other than `retention_killed_test.go` is modified.
- Verify: `go test -race -count=1 ./internal/harness/` exits 0; `git diff --name-only 1e2151a38 -- 'internal/harness/*_test.go'` lists only `retention_killed_test.go` among files that existed at the base.

## Edge cases

- The effective uid is 0, or the platform is Windows: permission and FIFO tests skip; the build/vet criterion (AC-HRH-011) still runs.
- The state path is a symbolic link to a directory or to a path that does not exist: the link is removed (never followed) and the prune continues; the target is untouched.
- Two processes meet the same permission fault at the same instant: at most one duplicate prune (spec §F); the stamp then bounds repeats.
- A late line is a malformed JSON line: it is carried verbatim like any other tail bytes (no classification of the tail).
- Late events older than the retention cutoff: carried, kept in the log, and archived by a later interval's prune.

## Quality gate criteria

- `go vet ./internal/harness/ ./internal/lockfile/` exit 0; the project linter at the CI-pinned version reports no new finding in `retention.go` or the touched tests (state the judging build next to the tree HEAD, `verification-claim-integrity.md` §2.2).
- `internal/harness` package coverage stays at or above the 85 percent TRUST 5 floor (the t1425 audit measured 87.5 percent); the new branches (heal, tail-carry, seam) are covered by AC-HRH-001..006 and -014.
- `go test -race` and `-count=20` on the killed-pruner and concurrent-process tests exit 0.

## Definition of Done

1. `red-baseline.md` is committed in a commit that precedes every commit touching `internal/harness/retention.go` (witnessed by `git log --reverse --format=%h -- <path>`).
2. AC-HRH-001..015 each PASS with the command and verbatim output recorded in `verdict.md`; any criterion whose RED was not observed is listed as a Gap.
3. `internal/lockfile` and `observer.go` are byte-identical to `1e2151a38`.
4. The residual window, F5 and F6 are disclosed in the source documentation with the labels in `spec.md` §A.
5. No criterion claims a Windows runtime observation; Windows evidence is stated as build and vet only.
6. The leader has been told which §B options remain operator-held and unselected.
