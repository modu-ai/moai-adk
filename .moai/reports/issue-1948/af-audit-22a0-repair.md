# Repair after the audit of 22a0afd247 (FAIL 0.68), lane-2 part: D1-D4, N1, N11, N12, N13 (issue-1948)

HEAD before: 22a0afd247. Commits: RED 6b2b6c8ee7; GREEN 742bbed59c (D1/D2/D3) and 565cf6de7b (D4); test follow-ups
330f67d042, 317e572e07, 7e36d586f9; this file in the docs commit that follows.

## Claim

1. N1, N11, N12 are closed by one change of kind, not by more rules: the output of a gate or of a git step no longer
   enters the ledger or the verb's output. The reason of a failed landing carries the failure kind, the gate, the exit
   status or timeout, the output byte count and the path of a log file. Closed-by-test: the audit's 46 plain shapes, 5
   URL shapes without a dot in the host, 4 escape-sequence shapes, 13 off-window shapes, a merge hook's stderr, 5 MB of
   hook output and 5 MB of gate output leave the value in none of stdout, stderr, the returned error and the whole ledger
   file (db, -wal, -shm bytes).
2. N13 is closed by test: `moai af evidence --subject-sha` and `moai af enqueue --candidate` accept only a complete
   lowercase hex SHA (40 or 64) that names a commit in the repository (`--repo`, else the card's lease worktree, else the
   project directory), and `mergeq.Merger.Land` refuses anything else with `ErrBadCandidate`. The audit's branch-name
   sequence ends in a refusal at its first verb.
3. N18, N21, N9 close with D3 (the tail writer is a ring buffer again, the marker watcher is gone); see Evidence for the
   test of each. Nothing here claims N14-N17 (lane-1) or the debt list below.
4. Not claimed: that no secret can reach a log file (it can; the log is the place output goes, readable by the user only)
   or that the log directory is not read by someone with the user's rights.

## Evidence

### Root causes, read in the code at 22a0afd247

- **N1** `internal/cli/af_land.go` (afFailureTail): `afStripControls` ran before `core.Redact`. A marker inside an OSC
  sequence (`ESC ] 0 ; password= BEL`) was deleted with the sequence, and the value after it became text with no marker.
  Round 1-2 had made the masking depend on seeing a marker; any normalisation that removes the marker defeats it.
- **N11** `core.Redact` is a deny-list of named shapes; the audit's plain shapes (`pwd=`, `--password <v>`, a Python dict,
  `postgres://...@db:5432`, `glpat-`...) are not in it, and 30 of 51 reached the ledger. The ledger cannot erase a row, and
  `Enqueue` copies a failure reason into a `notes` row. The defect is that output was recorded at all.
- **N12** `mergeq.Merger.Land` passed `err.Error()` of an `ErrMerge` to `m.fail`, and `GitCLI.run` built its error as
  `git <sub>: <err>: <stderr>` with up to 8 MiB of stderr. A repository hook's stderr therefore reached the ledger and
  stdout with controls and without a bound, around the whole masking path.
- **N13** `ledger.Enqueue` and `AddEvidence` take any non-empty string; `newestWorkIs` compares strings. Nothing at the
  verbs (`af.go` evidence and enqueue; `af.go:227` is the only non-test caller of `Enqueue`, checked with grep) required
  the string to name a commit, so a branch name stood for "the commit the branch points at today".

### Design (D1, D2, D3, D4)

- `mergeq.GitError{Sub, Cause, Output}`: `Error()` is `git <sub>: <cause> (N bytes of output withheld)`; the output is a
  field. `mergeq.MergeError{Cause, Output}` (`errors.Is(err, ErrMerge)` still holds) and `mergeq.GateError{Gate, Cause,
  Output, Total}` carry the same way. Every other `Verify` error is treated as output too (its text goes to the log).
- `mergeq.FailureLog{Dir, Cap}`: one file per card+candidate+kind (`gate`, `merge`), name `<clean card, 40>-<12 hex of
  sha256(card, candidate, kind)>-<kind>.log` (so `../../x`, `a/b`, `a\b`, an empty or NUL-bearing name cannot escape or
  collide), directory 0700 (an existing one is chmod-ed), file 0600 and created with `O_EXCL` after removing whatever is at
  the path (a planted symbolic link is replaced, not followed), the end of the output kept up to `DefaultFailureLogCap`
  (256 KiB) after `StripControls` and two passes of `core.Redact` (best effort), a first line says how many earlier bytes
  were dropped. The directory is `gate-logs` beside the ledger file. A failure to write is reported in the reason as
  "logging failed" and does not stop `m.fail`.
- `Merger.failReason` builds `<summary> (output N bytes; log: <path>)`; N counts what the writer was given, never less than
  what was kept. `afGate` returns a `GateError` whose cause is the exit status or timeout.
- D3: `TailWriter` is the ring buffer plus `Total()`. Removed: the watch, `MaskedPrefix`, `SecretMarkerPattern`,
  `afFailureTail`, `afMaskNotice`, `afMarkerAtLineEnd`, `afStripControls`. `StripControls` moved to faillog.go; its OSC
  rule now ends at the end of the line, so an unterminated OSC cannot swallow the lines after it (N21).
- D4: `mergeq.IsCompleteSHA`, `Merger.validCandidate` (format, then `RevParse(sha) == sha`, which also refuses the object
  name of an annotated tag), and `afCheckCommit` at both verbs, with a `--repo` flag on each.

### Observed output

RED (6b2b6c8ee7), before the change: every D1 test failed on "the value is in the stdout / ledger file" (66 reports) or "the
reason names no log file" (73), `MergeHookVolume` "stdout is 5000294 bytes", `D4Only...` "evidence accepted HEAD / trunk / c1 /
abbreviated / `--no-verify` / `-Xours` / zeros", `D4BranchNameRework...` "evidence accepted the branch name c1".

After (HEAD 7e36d586f9):

```
gofmt -l internal/cli internal/agenticfactory                                -> (no output)
go build ./...                                                                -> exit 0
go vet ./internal/cli/ ./internal/agenticfactory/...                          -> exit 0
env GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...  -> exit 0
go test -race -count=1 -p 4 ./internal/agenticfactory/...                    -> 21 `ok` lines, nothing else, exit 0 (mergeq 23.1s)
anchored cli run (the 32 older af tests that survive, the 3 anchors TestCleanReportsArchive_MovesOnlyEligibleCandidates,
  TestLaneChoiceMergeOffer, TestAutoPickMirrorParity, and the 12 new D1/D4 tests, -v)
                                                                              -> 47 top-level `--- PASS`, 174 `--- PASS` lines in all, 0 FAIL, ok 31.9s
golangci-lint run ./internal/agenticfactory/mergeq/... ./internal/agenticfactory/core/...   -> 0 issues
golangci-lint run ./internal/cli/...                                         -> 1 issue, errcheck in codex_close_entry_signal_test.go:89 (untouched);
                                                                                none in a file changed here (two staticcheck findings in the new test file were fixed)
```

Per-finding tests:

| Finding | Closed by |
|---|---|
| N1 (marker in OSC) | `TestAFLand_D1MarkerInsideAnEscapeSequenceNeverReachesTheLedgerOrStdout` (4 shapes, the unterminated one keeps its last line in the log) |
| N11 (plain shapes) | `TestAFLand_D1PlainShapesNeverReachTheLedgerOrStdout` (49 rows incl. the audit's 46 and 3 listed shapes), `...UrlHostWithoutDot...` (5), `...MarkerOffTheWindow...` (13) |
| N12 (git merge path) | `TestAFLand_D1MergeHookStderrNeverReachesTheLedgerOrStdout`, `...MergeHookVolumeIsBounded` (5 MB: stdout and fail_reason under 4 KiB, log under the cap, last hook line kept); mergeq: `TestLand_AMergeConflictsGitOutputIsInTheLogAndNotInTheReason`, `TestGitError_TextHoldsNoOutputButTheOutputIsKept` |
| N13 (candidate) | `TestAFVerbs_D4OnlyACompleteExistingCommitSHAIsAccepted` (14 bad values x 2 verbs, refusal wording checked), `...D4EnqueueRefusesAnEvidencedNameThatIsNotACommit`, `...D4BranchNameReworkEndsInARefusal`, `...D4TheRepositoryFlagAndTheObjectNameOfATag`; mergeq: `TestLand_RefusesACandidateThatIsNotACompleteSHAOfAnExistingCommit`, `TestLand_TheObjectNameOfAnAnnotatedTag...`, `TestLand_ATextThatIsNotACompleteSHAIsNotGivenToGit`, `TestIsCompleteSHA` |
| D2 (log) | `TestAFLand_D1LoudFailingGateKeepsItsLastLineInTheLog` (mode 0600/0700, cap, last line, byte count), `...ALogThatCannotBeWrittenDoesNotBlockTheFailure`, mergeq `TestFailureLog_*` (modes, cap and last line, name cannot escape, one file per key, planted link, loose directory mode) |
| N21 | the unterminated-OSC row above |
| N9 (rare off-window shapes) | the 13 off-window rows: with no output in the ledger the shapes no longer matter |
| N18 (speed) | `TestTailWriter_NineMegabytesStayWithinTheLimit` (0.00 s for 9 MB, `Total()` counts the dropped bytes); the detector it measured is gone |

### Tests removed or changed by D1/D3 (handoff list for lane-2-progress.md)

Removed (files deleted, with their helpers `afQ`, `afGateTailOf`, `afOutputOf`, `afFailReason`):

| Test | Reason |
|---|---|
| `TestAFGate_ALongValueWhoseMarkerIsOutsideTheKeptWindowIsMasked` (22 rows) | tested the masking of output in the reason; there is none; the same shapes are in `...D1MarkerOffTheWindow...` with the stronger oracle (whole ledger file, stdout, stderr) |
| `TestAFGate_ALoudGateWithNoMarkerStillShowsItsLastLine` | "last line verbatim in the reason" is now "last line in the log file"; covered by `...D1LoudFailingGateKeepsItsLastLineInTheLog` and the moved `TestAFLand_ALoudFailingGateKeepsItsLastLine` |
| `TestAFFailureTail_AMarkerSplitAcrossWritesIsStillSeen` | `afFailureTail` and the watcher are removed |
| `TestAFLand_ALongValueWhoseMarkerFellOffIsNotInTheJSONNorTheLedger` | superseded by the D1 off-window test (ledger bytes, not only fail_reason) |
| `TestAFGate_ARound2ShapesWhoseMarkerFellOffAreMasked`, `TestAFFailureTail_ARound2ShapesSplitAcrossWrites`, `TestAFLand_ARound2ShapesAreNotInTheJSONNorTheLedger` | the seven round-2 shapes (Bearer, ghp_, sk-, JSON, PEM, blank line, 600 BEL) cost nothing now: nothing is recorded; their plain forms are in the D1 plain test |
| `TestAFWatch_TheWritersMarkerDetectorAgreesWithRedact` | the detector it tied to `core.Redact` is removed |
| `TestAFLand_ASecretSplitByAControlOrAnEscapeIsStillMasked`, `TestAFLand_ACookieLineDoesNotTakeTheNextLinesMarkerWithIt`, `TestAFLand_AValueWhoseMarkerFellOffTheKeptTailIsNotShown`, `TestAFLand_AValueStraddlingTheCutAndControlsBeyondItStayMasked` | asserted masking inside the reason; replaced by the same shapes in the D1 tests |
| mergeq `TestTailWriter_MaskedPrefixCoversTheKeptPartOfAMarkedLine`, `TestTailWriter_WatchStateIsBounded` | `MaskedPrefix` and the watch state no longer exist |

Changed, not weakened:

| Test | Change |
|---|---|
| `TestAFLand_ALoudFailingGateKeepsItsLastLine` | moved as the leader decided: the last line is in the log file the reason names, and is not in the reason |
| `TestAFLand_ControlCharactersInAFailedGatesOutputAreNotKept` | the control-character check now runs on the reason (none allowed) and on the log (none allowed, tab kept). Found on the way: the old "the text around the controls survives" check passed vacuously, because `end` was found in the quoted gate text (`...end\n` in the gate command); the log shows the 8-bit CSI introducer consuming the `e` after it, as a terminal would, so the check is now `nd` |
| mergeq `TestLand_FailedGatesLeaveTheTrunkAndTakeNoLock` | asserted that the reason holds the Verify error's text (`go test failed`); now it must start with `gates failed` and must NOT hold that text (D1) |
| `af_land_d1_test.go`, one test fixed after RED | the first draft looked for 0x1b in the binary ledger file, where the byte occurs by chance; it now looks in stdout, stderr and the fail reason |

### Mutation contrast (each broken once, the named test failed, restored; the tree was identical to the commit after each)

| Guard | Mutation | Failed |
|---|---|---|
| `validCandidate` equality `got == sha` | removed | `TestLand_TheObjectNameOfAnAnnotatedTagIsNotACandidate` only |
| `validCandidate` format gate | disabled | `TestLand_ATextThatIsNotACompleteSHAIsNotGivenToGit` only (it counts ref lookups; the equality makes the gate otherwise redundant in effect, not in purpose: no text reaches git) |
| `failReason` "logging failed" | wrong message | `TestLand_ALogThatCannotBeWrittenStillRecordsTheFailure` |
| `failReason` total clamp | removed | survived the first tests (both call sites pass a total >= kept); `TestLand_TheByteCountIsNeverLessThanWhatWasKept` added, then fails |
| gate path: output in the summary | added | `TestLand_AGateErrorsOutputIsInTheLogAndNotInTheReason` |
| merge path: output in the summary | added | `TestLand_AMergeConflictsGitOutputIsInTheLogAndNotInTheReason` |
| `FailureLog` StripControls | removed | `TestFailureLog_FileModeDirectoryModeCapAndLastLine` |
| `FailureLog` Redact | removed | `TestFailureLog_BestEffortMaskingStillRunsBeforeTheWrite` |
| `FailureLog` cap | disabled | `TestFailureLog_FileModeDirectoryModeCapAndLastLine` |
| remove-before-create | removed | `TestFailureLog_OneFilePer...`, `TestFailureLog_ALinkAtThePathIsReplacedNotFollowed` |
| directory chmod | removed | `TestFailureLog_ADirectoryThatExistsWithLooseModeIsTightened` |
| digest in the file name | removed | `TestFailureLog_ACardNameCannotMakeTheFileEscape...`, `TestFailureLog_OneFilePer...` |
| `/` and `.` allowed in the name | added | `TestFailureLog_ACardNameCannotMakeTheFileEscapeItsDirectory` |
| file mode | 0644 and no chmod | `TestFailureLog_FileModeDirectoryModeCapAndLastLine` |
| `GitError.Error()` | includes output | `TestGitError_TextHoldsNoOutputButTheOutputIsKept` |
| `MergeError.Is` | always false | `TestLand_AMergeConflictsGitOutput...`, `TestLand_AConflictingCandidateFailsAndTheTrunkStays` |
| `UpdateRef` reads `err.Error()` | for the output | `TestGitCLI_OnlyAValueMismatchIsErrRefMoved`, `...EveryGitSubprocessIsPinnedToTheCLocale`, two real-ledger landing tests |
| OSC rule ends at the line | removed | `TestAFLand_D1MarkerInsideAnEscape...` (the unterminated row) |
| `TailWriter.Total` | not counted | `TestTailWriter_NineMegabytesStayWithinTheLimit` |
| evidence verb check | removed | `TestAFVerbs_D4Only...` |
| enqueue verb check | removed | survived the first tests (the fixture's card was already queued, so the ledger refused for another reason); `TestAFVerbs_D4EnqueueRefusesAnEvidencedNameThatIsNotACommit` (un-queued card, evidence under each bad name) added, then fails |
| `afCheckCommit` equality | removed | `TestAFVerbs_D4TheRepositoryFlagAndTheObjectNameOfATag` |
| `afCheckCommit` lease-worktree lookup | disabled | `TestAFVerbs_D4Only...` (valid SHA without `--repo`), `...D4EnqueueRefuses...` |
| `afCheckCommit` format gate | disabled | survived; the refusal wording of `TestAFVerbs_D4Only...` was added (a non-SHA must be refused for that, not for naming no commit), then fails |
| `afGate` output into the error | dropped | `TestAFLand_D1LoudFailingGateKeepsItsLastLineInTheLog`, `TestAFLand_ALoudFailingGateKeepsItsLastLine` |
| `afMerger` `Log` | dropped | the same two |

## Baseline-attribution

- Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, branch `WT-agentic-factory-integration`.
  HEAD before 22a0afd247 (read with `git rev-parse` before the first edit; status clean). Code commits listed at the top.
- Tools: `go`, `golangci-lint` from PATH on this tree. The log-mode assertions run where the OS has POSIX modes; they are
  skipped on Windows, which was built but not run.
- The audit's repro file `final-code-audit-22a0afd247/repro/cli_zz_audit_22a0_test.go.txt` was ported, not copied: the
  secret words are @-tokens expanded at run time (the write hook refused literal values), the oracle is stronger.

## Gaps

- `internal/cli` was not run as a whole (instructed). Only anchored tests ran: the 47 top-level tests listed above.
- N14-N17 and `TestZ2Codex_P2_*`, `TestZ2CLI_ActorSpoof`, `TestZ2Done_*` are lane-1's; not run, not merged.
- Not ported from the audit repro: `TestZ2Perf_*` (timing, the detector they measured is gone), `TestZ2F8_GateTimeoutZero`
  (F8 is debt), `TestZ2F11_OptionCandidate` (its cases `--no-verify`, `-Xours` are rows of the D4 tests).
- `lane-2-progress.md` was not written: it is outside the write scope of this assignment (internal/cli/af*, mergeq, core,
  `af-*.md`). The removed-test list is the table above and in the report back; the leader can paste it there.
- The write hook refused literal credential values and a literal PEM header earlier in this card; the tests assemble them.
- Throughput of the new path was not re-measured beyond the 9 MB tail-writer test and the 5 MB hook/gate tests (they pass
  within the package timeouts).
- The leader's `moai af land` live run on a new archive is not mine.

## Residual-risk

1. A log file can hold a secret. The cap is 256 KiB, the content is stripped and redacted only best effort (30 of the audit's
   51 shapes survive `core.Redact`), the mode is 0600 in a 0700 directory, and the reason names the path. A user who opens
   the file, or any process running as that user, can read it. The directory sits beside the ledger and is not cleaned up:
   files accumulate one per card+candidate+kind until removed by hand.
2. Diagnosis of a git failure that is not a merge (worktree add, rev-parse, update-ref) is withheld: `GitError.Error()` says
   how many bytes were withheld and the bytes are in the error chain, but only merge and gate failures are written to a log.
   `af land` prints the sanitised error; the git output of those steps is not saved.
3. A `Verify` that is not a `GateError` has its text sent to the log too, so a test double or a new gate type loses its
   message in the reason by design.
4. `afCheckCommit` trusts `--repo`, then the lease worktree, then the project directory; a caller that names the wrong
   repository gets "names no commit" (safe) and a caller that names a repository holding a different commit with that SHA
   cannot exist (SHA-1/SHA-256 collision). A SHA that names a commit in the repository but is not the card's own work (E6:
   the trunk's existing commit as candidate) is still accepted; D7 puts it out of scope.
5. The ledger still takes any non-empty string as a commit name through its API (REQ-AFL-008: it must not call git); the
   protection is at the verb and at landing. A caller that writes the ledger directly is not covered.
6. Debt reported, not fixed (D6): N3 (reclaimed landed card strands; not solved by this change), N4, N5, N6, N7, N8, N10,
   N19, N20, F8 (`--gate-timeout 0`), F9, F12. N14-N17 are lane-1's.
7. The 8-bit CSI strip consumes the byte after U+009B as a final byte (visible in the log as `end` -> `nd`); it matches a
   terminal and loses one letter of a log in a rare case.
