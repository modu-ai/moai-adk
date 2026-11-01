# Repair after the audit of db0a69698a: lane-2 result (card issue-1948)

Tree: `WT-agentic-factory-integration`, started at `db0a69698a`, code complete at `f8310bac68`. Items: N22, N23, N29, N30,
N31, N33, N34 (queue.go excluded), N35, N36, and the added N26. N24, N25, N27, N28, N32, N3 and the `queue.go:118`
comment belong to lane-1 and are not touched here.

## Claim

1. N22, N23, N29, N30, N31, N33, N34, N35, N36 and N26 are each fixed by a RED commit (a test that fails on the old code) followed by a GREEN commit.
2. The failure log is written only after the directory is a real directory, the temporary file and the directory carry no extended ACL (macOS), and it is renamed into place; a failure at any of those steps drops the output and the reason says "logging failed".
3. The reason, the verb's stdout, the notes and the landed note carry a short gate id (`gate-` and 12 hex digits of the SHA-256 of the gate text), never the gate text; the text is the first line of the failure log.
4. `moai af land` and `moai af reconcile` accept only `--trunk refs/heads/af/trunk` (`mergeq.DefaultTrunk`); `evidence`, `enqueue`, `land` and `reconcile` accept `--repo` only for a path whose git common directory is the project directory's. Both refusals happen before the ledger is opened.
5. The `mergeq.Merger` library keeps its `Trunk` field; the verbs are the trust boundary.

## Commits (RED then GREEN)

| Item | RED | GREEN |
|---|---|---|
| N36 Redact before StripControls | 864f186cd4 | 77cf310264 |
| N31 cap again after the masking | 302762dd5a | ea8a89e76b |
| N31 one total cap, oldest first | 9293f59f2c | af42ea749a |
| N35 temp file and rename | 0211c5a6fc | 1b4d7fd941 |
| N23 log directory is a link | 9161a9bdf7 | f9fe906024 |
| N22 inherited ACL (macOS) | 6c7138e550 | a5eaea9bb5 |
| N33 Reconcile checks candidates | 8ce08a1d22 | cac331f63a |
| N30 git step output to the log, worktree add cleanup | c54ac95f6f | 143bbada4e |
| N29 gate id instead of gate text | e7772e4c71 | 428ad93449 |
| N31 lane bound tightened, verb test with growing output | (test-only, mutation shown below) | db4f7daa2b |
| N34 help, comment | d735101f59 | 88760764b6 |
| N26 fixture move (no assertion changed, passes on the old code) | bc98b808e6 | |
| N26 verbs refuse other trunk and other repo | 56d7bd261a | d516daa947 |
| N26 help wording | 538d72f87d | d516daa947 |
| Masking cost: cut right after the first pass | | f8310bac68 |

## What changed, by item

- **N22**: `faillog_acl_darwin.go` runs `/bin/chmod -N` (absolute path) on the log directory, before anything is made in it, and on the empty temporary file, before any byte is written. `faillog_acl_other.go` is a no-op. The `FailureLog` doc comment no longer calls the file mode a guarantee; it lists what is set and what is not covered. If clearing fails, `Write` returns an error, the temporary file is removed, and `failReason` says "logging failed".
- **N23**: `prepareDir` does `os.Lstat` on the last path component; a link or a non-directory is refused and nothing is changed. Parent components are out of scope (comment). A `chmod` failure is now an error, not ignored.
- **N29**: `afGateID` in `af_land.go`. `GateError.Gate` and `Merger.Gate` hold the id; `Merger.GateText` holds the text and goes only to the log, as a `gate: ...` first line (masked, at most 2 KiB). `--command`, `--reason`, `--path`, `--actor`, `--session` stay as given and their help says so. `TestAFLand_ALandedNoteSaysWhichGateVerifiedIt` asserted the gate text in the landed note; that is the thing N29 removes, so it now asserts the id and the absence of the text (the check "names the gate" is kept).
- **N30**: `Land` wraps the old body (`land`); any `*GitError` that reaches the wrapper is written to the log under kind `git` and the returned error names the file (`... (log: <path>)`) or says `(logging failed)`. The error text still holds no git output. `PrepareCandidate` discards a failed `worktree add` with `Discard` (remove directory and `worktree prune`).
- **N31**: the cap is applied before the masking (bounds the work) and again after it (bounds the file); one total cap for the directory (`DefaultFailureLogTotalCap`, 1 MiB, `TotalCap` to change), oldest `*.log` first, the file just written never removed. `afD1LogCapUpper` is now the cap plus 4 KiB (was twice the cap).
- **N33**: `Reconcile` calls `validCandidate` for each verifying or landing row before touching it; a bad row is reported as an error, the other rows are still judged.
- **N34**: the stale help sentence, the broken comment, the missing log location, the "secrets can remain" statement and the "recorded as given" statement. The `--trunk`/`--repo` wording is N26's (below), not a warning.
- **N35**: unique temporary file in the log directory, `os.Rename` onto the name.
- **N36**: `Redact`, `StripControls`, `Redact`, `Redact` (the old order after the first masking is kept).
- **N26**: `afCheckTrunk` and `afCheckRepoFlag` in `af.go`, called from `afMergeRun` (land, reconcile) and from the `evidence` and `enqueue` handlers before `afRun` opens the ledger. `afCheckRepoFlag` runs `git -C <dir> rev-parse --git-common-dir` for the project directory and for `--repo` and compares the two directories with `os.SameFile`; a project directory outside any repository refuses every `--repo`. A flag left out is not checked (the verbs use the lease worktree or the project directory as before).

## Constraint for the N19 SPEC amendment

**The trunk name cannot be changed.** `moai af land` and `moai af reconcile` take `--trunk` only as `refs/heads/af/trunk`; there is no configuration, flag or environment variable that selects another trunk. The audit records the means of setting the trunk as open (REQ-AFL-014, OD-35 (l)); with this change the means is: there is none, the trunk is the one dedicated ref. A project that wants another trunk needs a SPEC change and a code change at the verbs (`afCheckTrunk`). Likewise `--repo` is bound to the project directory's git common directory. CLI test fixtures were moved from `refs/heads/trunk` to `refs/heads/af/trunk` and the project directory (`CLAUDE_PROJECT_DIR`) is now the repository in the fixtures (`afQueuedCard`, `afUnqueuedCard`, `TestCardFlow_ClaimToDone_ThroughMergeQueue`); two lists of invalid candidate names carry `af/trunk` where they carried `trunk` (a real branch name is still a real branch name).

## Evidence (verbatim tails of what I ran in this tree)

```
gofmt -l internal/cli internal/agenticfactory         -> (no output)
go build ./...                                        -> exit 0
go vet ./internal/cli/ ./internal/agenticfactory/...  -> exit 0
env GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...  -> exit 0
env GOOS=linux   GOARCH=amd64 go build ./internal/cli/ ./internal/agenticfactory/...  -> exit 0
env GOOS=windows GOARCH=amd64 go vet ./internal/agenticfactory/mergeq/ ./internal/cli/ -> exit 0  (test files compile)
env GOOS=linux   GOARCH=amd64 go vet ./internal/agenticfactory/mergeq/ ./internal/cli/ -> exit 0
go test -race -count=1 -p 4 ./internal/agenticfactory/...  -> exit 0, 21 "ok" lines, no other line
go test -count=1 -v -run '^(TestAF|TestCardFlow|TestCleanReportsArchive_MovesOnlyEligibleCandidates|TestLaneChoiceMergeOffer|TestAutoPickMirrorParity)' ./internal/cli/
  -> ok  internal/cli  97.054s   top-level PASS 57  FAIL 0  SKIP 0
golangci-lint run ./internal/agenticfactory/mergeq/... ./internal/cli/
  -> internal/cli/codex_close_entry_signal_test.go:89:17: Error return value of `pw.Close` is not checked (errcheck)   1 issues
```

The one lint issue is in a file this card does not touch (`git log` of this repair has no commit on it); it is the audit's F12, pre-existing.

New and existing anchored CLI tests, each `--- PASS` in the last run: `TestAFLand_ALogDirectoryThatIsALinkIsRefusedAndItsTargetIsLeftAlone`, `TestAFLand_TheLogFileEndsWithoutAnInheritedACL`, `TestAFLand_TheGateTextIsNotInTheReasonTheStdoutOrTheLedgerButTheLogHoldsIt` (9 shapes), `TestAFLand_TheLandedNoteCarriesTheGateIDAndNotTheGateText`, `TestAFLand_TheLogStaysUnderItsCapWhenTheMaskingMakesTheOutputLonger`, `TestAFLand_TheHelpSaysWhereTheLogsAreThatTheyCanHoldSecretsAndWhatIsRecordedAsGiven`, `TestAFVerbs_TheFreeInputFlagsSayThatTheyAreRecordedAsGiven`, `TestAFVerbs_TheHelpSaysTheTrunkCannotBeChangedAndWhichRepositoriesAreAccepted`, `TestAFLand_ATrunkOtherThanTheDedicatedOneIsRefusedBeforeLandingAndTheLedgerStaysAsItWas`, `TestAFVerbs_ARepoThatIsNotTheProjectsRepositoryIsRefusedBeforeAnyChange`, and the existing `TestAFLand_D1*` (8) and `TestAFVerbs_D4*` (4).

New mergeq tests (in `faillog_repair_test.go`, `faillog_acl_darwin_test.go`): a marker inside an OSC keeps no value; the cap holds after masking; one total cap with oldest-first removal (seeded files with fixed mtimes); concurrent writers of one key (8 x 100) all succeed and never mix; a link or a plain file as the log directory is refused with its target untouched; an ACL that cannot be cleared drops the output and the reason says logging failed (file and directory cases); inherited ACL cleared (macOS); `Reconcile` with six bad candidate names in both states; a failing git step (worktree add and update-ref) is in the log and not in the error; a failed worktree add leaves no registered worktree and no scratch directory.

### RED evidence (old code, same tests)

- N36: `osc_marker_bel: the value is in the log: "hunter2super\n"` (and `osc_marker_st`).
- N31 cap: `the log holds 20518 bytes, over the cap 8192`; verb level (mutation below): `log: 655420 bytes (bound 266240)`.
- N31 total: `the directory holds 1286189 bytes in 6 files, over the total cap 1048576`.
- N35: `183 of 800 writes failed, the first with open .../t1-e77684e96481-gate.log: file exists`.
- N23: `the link's target holds [- keep.txt - t1-e77684e96481-gate.log]`.
- N22: `.../gate-logs/t1-e77684e96481-gate.log has an ACL: [" 0: group:everyone inherited allow read,execute,readattr,readextattr,readsecurity"]`.
- N33: `verifying row with candidate "trunk": error <nil>, want ErrBadCandidate` and `actions [{CardID:bad From:verifying To:landed ...}]`.
- N30: `the error names no git log file: "git worktree: exit status 7 (106 bytes of output withheld)"`; `a worktree stays registered ... prunable`.
- N29: the value in stdout, the ledger file and a note for the shapes the masking misses (seen in the output: `pass_assign`, `mysql_p`, `flag_space`, `token_flag`, `url_no_dot`, `curl_u`; the run's output was cut after those, so the count for the other three shapes is not recorded here); no gate id in any reason.
- N26: `land --trunk "refs/heads/c1": error <nil> (stdout {"card_id":"t1","outcome":"landed","reason":"already in the trunk"...` and `evidence --repo <other repo>` accepted.

## Mutation contrast

| Mutation | Result |
|---|---|
| remove the cap applied after the masking (`tailOf(mask(body), limit)` -> `mask(body), 0`) | killed by `TestFailureLog_TheCapHoldsAfterTheMaskingMadeTheTextLonger` (20518 > 8192) and by the verb test (655420 > 266240, the audit's number) |
| `stripACL` returns nil | killed by `TestFailureLog_TheLogAndItsDirectoryEndWithoutAnInheritedACL`, `...WhenTheACLCannotBeCleared...` (file case) and `TestAFLand_TheLogFileEndsWithoutAnInheritedACL` |
| `afCheckTrunk` never refuses | killed by the trunk test (`land --trunk "refs/heads/c1"` landed) |
| `afCheckRepoFlag` never refuses | killed by the repo test (evidence accepted) |
| `afMergeRun` skips the repo check | killed by the repo test (land and reconcile rows) |
| `enqueue` skips the repo check | killed by the repo test (enqueue rows) |
| the fixes of N22/N23/N29/N30/N31/N33/N34/N35/N36 | each RED run above is the mutation: the new test fails on the old code |

Not run as separate mutants: clearing the ACL of the directory only, or of the file only (either alone leaves the file without an ACL, so the two clearings are redundant by design; the test would pass with either).

## Baseline-attribution

Every command above was run in `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration` against the tree at the commits named, with `go1.26.8 darwin/arm64` and `git 2.54.0`; no moai binary or MCP tool judged anything. No `MOAI_*_LIVE` variable was set. `./internal/cli` was never run in full; only anchored `-run` patterns.

## Gaps

- **Linux and Windows ACLs are not measured.** They build and their tests compile (`go build`, `go vet`); the ACL tests are darwin-only by file name or `runtime.GOOS`. `clearExtendedACL` is a no-op there and the `FailureLog` comment says that nothing is claimed.
- No other account read the log; the ACL claim rests on `ls -lde` showing no entries, as the audit's did.
- The full `internal/cli` suite was not run (instructed).
- The lane-1 items (N24, N25, N27, N28, N32, N3) and the `queue.go:118` comment were not touched; whether they hold in the merged tree is the merge step's to measure.
- `rev-parse --git-common-dir` was run only with git 2.54.0; an older git that prints it differently was not tried.
- The masking time on hostile inputs was measured once on this machine while other sessions were running (below).
- The refusals of `--repo` were tested with a second repository, a non-repository directory, and a missing path; a project directory that is itself not in a repository was not tested as a separate case.
- No tool call was refused by the worktree guard apart from heredoc and compound command lines, which I replaced with the Write and Edit tools; the one stray file I created outside the tree (`/tmp/claude-501/af-race.log`) was removed.

## Residual-risk

- **Masking time.** The log write runs the masking three times (N36 added one). After the cut-right-after-the-first-pass change the worst of eleven hostile 256 KiB inputs took 3.0 s (`newlines_after_key`) and `key_eq_lines` fell from 4.2 s to 1.1 s; the times vary by a factor of two between runs here. The audit measured 0.6 s for the old code on its machine. The write happens after the gate, outside the gate timeout.
- A writer killed between creating the temporary file and renaming it leaves a `.tmp-*` file (redacted text, mode 0600) that nothing removes; the retention pass looks at `*.log` only.
- The retention pass is best effort: a file that cannot be removed is skipped, so the directory can stay over the total cap.
- The log can still hold secrets (the help says so); any process of the same user can read it. The gate text is in the log (masked best effort), so a secret written into a gate command may sit there.
- `--repo` and `--trunk` now refuse values existing scripts may pass. No script, doc or template in this tree outside `internal/cli/af*.go` mentions `--trunk` (grep), so none was found; callers outside this tree are not known.
- The project directory is resolved from `CLAUDE_PROJECT_DIR` or the working directory; a project directory that is not inside the repository the operator means makes every explicit `--repo` fail (the default, flag left out, still works).
- `go.mod`-level or CI differences between darwin and the Linux runner were not measured.
