# Run-phase evidence, milestones M1 to M3 (card t1432, SPEC-HARNESS-RETENTION-HARDEN-001)

Platform for every run below: darwin, uid 501, go1.26.8. Linux and Windows runtime: not observed. Windows evidence is `GOOS=windows go build` and `go vet` only. Tests were run with the kanban variables unset in the same invocation. Baseline commit T = `ac40cf3bf` (observed RED).

## M1 — tail-carry and log terminator (REQ-HRH-006, -007, -008; AC-HRH-007, -008, -009)

Tree measured: HEAD `ac40cf3bf` plus the uncommitted M1 edit of `internal/harness/retention.go` (committed as the M1 commit that carries this section).

Claim 1: AC-HRH-007 and AC-HRH-008 now pass and the quiescent pins stay green.

- Command: `go test -count=1 -v -run '^(TestPruneCarriesLateEvents|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval|TestPruneKeepsUnparsedLinesVerbatim|TestPruneNothingStaleLeavesLogUntouched)$' ./internal/harness/`
- Output (the deciding lines, exit 0):

```
--- PASS: TestPruneNothingStaleLeavesLogUntouched (0.00s)
--- PASS: TestPruneKeepsUnparsedLinesVerbatim (0.00s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/wrong-field-type (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/truncated-json (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/plain-text (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/json-array (0.01s)
    --- PASS: TestPruneKeepsUnparsedLinesVerbatim/padded-text (0.01s)
--- PASS: TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval (0.01s)
--- PASS: TestPruneCarriesLateEvents (0.31s)
--- PASS: TestPruneTailPartialLineCarriedAndTerminated (0.31s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.945s
```

Claim 2: the whole package passes except the two tests that are M2-owned RED by design (AC-HRH-001, AC-HRH-002).

- Command: `go test -count=1 -v ./internal/harness/` redirected to a scratch file: exit 1, 271 `--- PASS` lines, exactly two `--- FAIL` lines: `TestPruneStateUnwritableFileReplaced` and `TestPruneStateSymlinkReplacedTargetUntouched` (the two M0 RED tests that M2 flips).
- Command: `go test -count=1 -skip '^(TestPruneStateSymlinkReplacedTargetUntouched|TestPruneStateUnwritableFileReplaced)$' ./internal/harness/`
- Output: `ok  	github.com/modu-ai/moai-adk/internal/harness	1.218s`, exit 0.
- Gap: the instruction to commit M1 "when the whole package passes" cannot be met literally, because the two M0 REDs belong to M2 (`plan.md` M2 step 5). The M1 commit is gated on the package minus those two named tests.

Claim 3: vet, gofmt and Windows compile are clean after M1.

- `go vet ./internal/harness/` exit 0; `gofmt -l internal/harness/` empty; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.

Mutation runs (mutants built from the CURRENT `retention.go` as scratch copies outside the tree, applied with `go test -overlay`; each removed afterwards, `git status --short` shows only `M internal/harness/retention.go`). Command shape: `go test -count=1 -overlay <overlay> -run '^(TestPruneCarriesLateEvents|TestPruneTailPartialLineCarriedAndTerminated|TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval)$' ./internal/harness/`

| Mutant | Killed by | Exit |
|---|---|---|
| verbatim tail copy with no terminator (the `if tail[len(tail)-1] != '\n'` block removed) | `TestPruneTailPartialLineCarriedAndTerminated` (`replacement log does not end with a newline`; the next append glued onto the fragment) and `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` (`rewritten log does not end with a newline`) | 1 |
| classifying the final unterminated line (the unterminated-line guard in `scanTerminatedLines` removed) | `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` (`stale-final count after first prune = 0, want 1`) | 1 |
| dropping the tail copy (the `appendLogTail` call removed) | `TestPruneCarriesLateEvents` (`late-event count = 0, want 1`), `TestPruneTailPartialLineCarriedAndTerminated` (`partial fragment occurs 0 times, want 1`), `TestPruneStaleFinalLineWithoutNewlineWaitsOneInterval` | 1 |

Gaps (M1): whether a concurrent `O_APPEND` write can be seen half-complete was not observed (the boundary rule is tested deterministically). The residual window between the final tail reading and the rename is not closed and not measured. The `-race` run is recorded once after M3.

## M2 — state-path hygiene: ownership, heal safety, warning (REQ-HRH-001..005; AC-HRH-001..006)

Tree measured: HEAD `c1cc3fe67` (the M1 commit) plus the uncommitted M2 edits (committed as the M2 commit that carries this section). New files: `internal/harness/retention_owner_unix.go` (`//go:build !windows`), `internal/harness/retention_owner_windows.go` (`//go:build windows` twin, same symbol `entryOwnedByCurrentUser`), `internal/harness/retention_owner_test.go` (`//go:build !windows`). Edited: `internal/harness/retention.go` (the `ownerCheck` field, `openStateFile`, `healStateEntry`, `removeStateEntryIfUnchanged`), `internal/harness/retention_statepath_test.go` (one added assertion, see Disagreements).

Claim 1: AC-HRH-001 and AC-HRH-002 flipped, and every M2 test passes.

- Command: `go test -count=1 -v -run '^(TestPruneStateSymlinkReplacedTargetUntouched|TestPruneStateUnwritableFileReplaced|TestPruneStateUnreplaceableInReadOnlyDirSkips|TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds|TestPruneStateForeignOwnedLeftUntouchedAndWarns|TestOwnerCheckDefault|TestHealDoesNotRemoveAFreshStateFile)$' ./internal/harness/` (output filtered to result lines), exit 0:

```
--- PASS: TestPruneStateForeignOwnedLeftUntouchedAndWarns (0.05s)
    --- PASS: TestPruneStateForeignOwnedLeftUntouchedAndWarns/symlink (0.04s)
    --- PASS: TestPruneStateForeignOwnedLeftUntouchedAndWarns/unwritable-file (0.00s)
--- PASS: TestOwnerCheckDefault (0.00s)
--- PASS: TestPruneStateUnreplaceableInReadOnlyDirSkips (0.01s)
--- PASS: TestHealDoesNotRemoveAFreshStateFile (0.00s)
--- PASS: TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds (0.00s)
--- PASS: TestPruneStateSymlinkReplacedTargetUntouched (0.02s)
--- PASS: TestPruneStateUnwritableFileReplaced (0.02s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.911s
```

No `--- SKIP` line for any of the seven (uid 501, so the root skips did not fire; `TestOwnerCheckDefault` step (e) found `/var` as a foreign-owned symbolic link).

Claim 2: the whole package passes.

- Command: `go test -count=1 -v ./internal/harness/` redirected to a scratch file: exit 0, 276 `--- PASS` lines, zero `--- FAIL`, one `--- SKIP` (`TestPruneHelperProcess`, a pre-existing helper process test, not part of this card).
- `go vet ./internal/harness/` exit 0; `gofmt -l internal/harness/` empty; `golangci-lint run --timeout=3m ./internal/harness/` prints `0 issues.` (golangci-lint v2.1.6 built with go1.26.8; not asserted to be the CI-pinned build).
- `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.

Mutation runs (each mutant built from the CURRENT `retention.go` or `retention_owner_unix.go` as a scratch copy outside the tree and applied with `go test -count=1 -overlay <json> -run '<test>' ./internal/harness/`; all nine exit 1; removed afterwards, `git status --short` shows only the intended M2 files):

| Criterion | Mutant | Killed by | Deciding failing output |
|---|---|---|---|
| AC-HRH-004 | owner result ignored (`if false && !r.ownerCheck(...)`) | `TestPruneStateForeignOwnedLeftUntouchedAndWarns` | `want an error naming <state path>, got <nil>`; `state-path entry changed`; `want exactly one warning line ... got ""` |
| AC-HRH-004 | removal before the ownership check | same | `state-path entry changed: err=lstat ...: no such file or directory` |
| AC-HRH-004 | warning omitted | same | `want exactly one warning line naming <state path>, got ""` (both subtests) |
| AC-HRH-005 | check always "owned" | `TestOwnerCheckDefault` | `/ must not count as owned for a non-root user`; `the constructor must install the real owner check: / counted as owned`; `/var is a symbolic link owned by another user and must not count as owned` |
| AC-HRH-005 | owner read follows the link (`os.Stat`) | `TestOwnerCheckDefault` | `a user-owned link to / must count as owned: the link's own owner decides, not its target's` |
| AC-HRH-005 | every symbolic link treated as owned | `TestOwnerCheckDefault` | `/var is a symbolic link owned by another user and must not count as owned` |
| AC-HRH-006 | unconditional removal | `TestHealDoesNotRemoveAFreshStateFile` | `removal of a changed entry: removed=true err=<nil>, want false and nil`; `the fresh state file was removed or replaced` |
| AC-HRH-003 | heal that removes a directory | `TestPruneStamp_StateNotOpenableSkipsPruneAndRecordSucceeds` (case a) | `expected an error when the state file cannot be opened` |
| AC-HRH-003 | heal that ignores the removal failure | `TestPruneStateUnreplaceableInReadOnlyDirSkips` (case b) | `error does not wrap a permission error: retention: prune state entry <state path> changed on every inspection; prune skipped` |

Disagreements between the SPEC text and reality (the SPEC is not edited here):

1. AC-HRH-003's mutant "a heal that ignores the removal failure" is not killed by the case-b test as the SPEC states it (a non-nil error and byte-identical state): with the removal error ignored the loop still ends in an error after three inspections and leaves everything untouched. The test was strengthened by one assertion in `retention_statepath_test.go` (the returned error must satisfy `errors.Is(err, fs.ErrPermission)`, i.e. carry the cause); the unmodified code passes it, the mutant fails it, whether base code still passes the strengthened assertion was not re-observed (READ only: the base error wraps the open's permission error).
2. `plan.md` M1 asks to commit only when the whole package passes; at M1 the two M0 tests of AC-HRH-001 and AC-HRH-002 are RED by design until M2, so the M1 gate was the package minus those two named tests (recorded under M1 above).

Gaps (M2): the time-of-check gap between the inspection and the open of a regular file is closed for writes by the post-open identity check and is not pinned by a test; the heal burst of N processes is not exercised and not measured; the removal itself is conditional on an unchanged entry but the gap between that last re-inspection and `os.Remove` is not closed; whether a hook's stderr reaches a user was not measured; Linux and Windows runtime not observed (Windows twin answers "not owned" for every entry, verified by build and vet only).

## M3 — stamp-write pin and the locked-phase seam (REQ-HRH-014, -015; AC-HRH-013, -014)

Tree measured: HEAD `f52dd1b1c` (the M2 commit) plus the uncommitted M3 edits (committed as the M3 commit that carries this section). The locked phase of `pruneExclusive` is extracted into `(*Retention).pruneLocked(sf *os.File, retentionDays int) error` (stamp re-check with a fresh clock, stamp write, prune, orphan sweep); `pruneExclusive` opens the file through `openStateFile`, locks it and calls it. New file: `internal/harness/retention_stampwrite_test.go` (`//go:build !windows`).

Claim 1: AC-HRH-013 and AC-HRH-014 pass unmutated.

- Command: `go test -count=1 -v -run '^(TestPruneStampWriteFailureSkipsPrune|TestPruneStampShorterOverLongerIsExact)$' ./internal/harness/`, exit 0:

```
--- PASS: TestPruneStampWriteFailureSkipsPrune (0.00s)
--- PASS: TestPruneStampShorterOverLongerIsExact (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/harness	0.723s
```

Claim 2: the read-only-handle seam works on darwin: the stamp write fails (the unmutated test passes with a non-nil error), so the D5 fallback (option B, a field) was not needed. Linux behaviour of a read-only handle is not observed.

Claim 3: the whole package passes, with vet, gofmt, race and Windows compile.

- `go test -count=1 -v ./internal/harness/` redirected to a scratch file: exit 0, 277 `--- PASS`, zero `--- FAIL`, one `--- SKIP` (`TestPruneHelperProcess`, pre-existing); last lines `PASS` and `ok  	github.com/modu-ai/moai-adk/internal/harness	1.366s`.
- `go test -race -count=1 ./internal/harness/` prints `ok  	github.com/modu-ai/moai-adk/internal/harness	7.759s`, exit 0.
- `go vet ./internal/harness/` exit 0; `gofmt -l internal/harness/` empty; `GOOS=windows go build ./internal/harness/ ./internal/lockfile/` exit 0; `GOOS=windows go vet ./internal/harness/ ./internal/lockfile/` exit 0.

Mutation runs (scratch copies outside the tree, applied with `go test -count=1 -overlay <json> -run '<test>' ./internal/harness/`; both exit 1; removed afterwards):

| Mutant | Killed by | Deciding failing output |
|---|---|---|
| mB: the `Truncate(0)` block in `writeStamp` removed | `TestPruneStampShorterOverLongerIsExact` (AC-HRH-013) | `state file = "2026-10-02T00:00:00Z123456789Z", want exactly "2026-10-02T00:00:00Z"`; `state file does not parse as a fresh stamp` |
| mC: the stamp-write error ignored (`_ = writeStamp(sf, now)` in `pruneLocked`) | `TestPruneStampWriteFailureSkipsPrune` (AC-HRH-014) | `pruneLocked returned nil, want the stamp-write error`; `log changed although the stamp could not be written`; `archive directory exists although the stamp could not be written: <nil>` |

mD (AC-HRH-012) belongs to M4 and was not run here.

Boundary checks (AC-HRH-009, AC-HRH-011), measured against the tree HEAD `f52dd1b1c` before the M3 commit; the M3 edit touches none of the guarded functions; a re-run of the history checks on the committed M3 tip is reported in the completion message, not here, because this file is part of that commit:

- `git diff --quiet 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go` exit 0 (no output).
- The four-function `git log --format=%h -s -L ...` over `ac40cf3bf..HEAD` and over `1e2151a38..HEAD` (option order as `acceptance.md` E-028 prescribes) prints nothing, exit 0. Positive control, same command over `fe211e9c9~1..fe211e9c9`, prints `fe211e9c9`, exit 0, so the form fires on a range that holds a change.
- `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` lists `c1cc3fe67` and `f52dd1b1c` (M1 and M2), and `git merge-base --is-ancestor ac40cf3bf` exits 0 for each: the baseline commit T is an ancestor of every production commit.

Gaps (M3): Linux behaviour of a read-only handle's truncate is not observed (CI on ubuntu-latest exercises it first); `go test ./internal/lockfile/` and the remaining packages were not run here (scope is `./internal/harness/`); coverage was not re-measured; mD, M4 (N1), and M5 (disclosure comments) are outside this run. Windows runtime not observed.
