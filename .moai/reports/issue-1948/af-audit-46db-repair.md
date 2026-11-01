# Repair after the audit of 46dbcce0d4, lane-2 items N42, N43, N45 part (card issue-1948)

Tree `.moai/worktrees/agentic-factory-integration`, branch `WT-agentic-factory-integration`, start HEAD `46dbcce0d4`.
Single writer. No push, no PR, no `MOAI_*_LIVE`, no run of the whole `./internal/cli/...` package, no independent audit.
`internal/agenticfactory/ledger/**` untouched. The reclaim-cap logic is untouched.

## Claim

| # | Claim | Commits |
|---|---|---|
| 1 | N42: `moai af resume <card> --epoch N --guard reclaim-cap --reason <text>` exists. From the state of `TestA3Park_ReclaimCapOnALandedCard` (three reclaims, parked, landed commit in `refs/heads/af/trunk`) the existing refusals stand without it, and with it the card unparks, is claimed, measures c1 again in its own epoch, is taken over and reaches DONE through the verbs. An empty or blank or missing reason, a wrong or missing guard, a stale epoch (both directions), an unknown card and an unparked card are refused with the ledger byte-for-byte as it was (snapshot of cards, queue, notes, evidence, leases). | RED `5529297d60`, GREEN `7deb693d57` |
| 2 | N43: the retention pass of the failure log removes a stale `.tmp-*` leftover (older than `failureLogTmpMaxAge`, one minute) at every prune, counts a fresh temp file toward the total cap and never touches it. | RED `f7558b851d`, GREEN `04a8b2de6f` |
| 3 | N45 part: the help of `af resume` is written to the behavior, and the help of `af move` (which said a reclaim-cap resume can only be made through the ledger API) now points at `resume`. The help of `af land` ("does not let the caller choose", "at most 1 MiB for the whole directory") stays true and is unchanged. | `7deb693d57` (text), test `c924c9365e` |

Commit list on top of `46dbcce0d4`, oldest first:

```
5529297d60 test(af): N42 a card parked at reclaim-cap has an operator verb out, af resume, RED
7deb693d57 feat(af): N42 af resume, the operator's exit from a reclaim-cap park, through the existing in-place move with the resume guard, GREEN
f7558b851d test(af): N43 a temp file left by a dead log write is removed and counted in the total cap, a fresh one is not touched, RED
04a8b2de6f fix(af): N43 the retention pass removes a stale temp file left by a dead log write and counts a fresh one in the total cap, GREEN
c924c9365e test(af): N45 part, the help of af resume and the pointer in af move match what the verbs do
```

(the commit adding this report comes after them; the final HEAD is reported with the hand-over)

What `af resume` does: it reads the card, refuses unless it carries a park with reason `reclaim-cap`, and calls the existing
`Store.Move` with the same stage and sub (no park) and one `ledger.Evidence` whose `Guards` is `resume-guard:reclaim-cap`.
The `--reason` goes to the `reason` column of the existing `transitions` row, `--actor` (default `operator`) to its `actor`;
no schema was added. The flags follow `af move`: `--db`, `--epoch` (required), `--actor`, `--reason`. `--guard` and `--reason`
are required flags and refuse a blank value.

## Evidence

Commands and outputs of this run, tree HEAD `c924c9365e` (before the report commit), all run from the worktree.

RED (before the verb existed), `go test -count=1 -run '^(TestAFVerbs_N42.*|TestAFCommand_VerbsRegistered)$' ./internal/cli/`:

```
--- FAIL: TestAFVerbs_N42TheResumeVerbLetsAReclaimCapParkedCardFinishOnItsLandedCommit (0.71s)
    af_resume_test.go:156: moai af resume t1 --epoch 3 --guard reclaim-cap --reason ...: unknown command "resume" for "af"
--- FAIL: TestAFCommand_VerbsRegistered (0.00s)
    af_test.go:49: no `moai af resume` verb for the operator resume of a reclaim-cap park step
```

RED for N43 (before the fix), `go test -count=1 -run '^TestFailureLog_A(Stale|Fresh)' ./internal/agenticfactory/mergeq/`:

```
--- FAIL: TestFailureLog_AStaleLeftoverTempIsRemovedAndTheDirectoryEndsWithinTheCap (0.10s)
    faillog_temp_test.go:66: the stale leftover temp is still there
    faillog_temp_test.go:69: the directory holds 309240 bytes, over its total cap 2048
--- FAIL: TestFailureLog_AStaleLeftoverTempIsRemovedEvenBelowTheCap (0.02s)
    faillog_temp_test.go:82: the stale leftover temp is still there
--- FAIL: TestFailureLog_AFreshTempIsLeftAloneAndCountedInTheTotal (0.03s)
    faillog_temp_test.go:109: the older log stayed although the fresh temp's bytes put the directory over its cap
```

Mutation contrasts (each: break the fix, run the test, see it fail, restore; the restored tree passes):

| Mutation | Result |
|---|---|
| resume: `strings.TrimSpace(reason) == ""` becomes `reason == ""` | `.../a_reason_of_blanks` fails (the blank reason resumed the card, and the later subtests fail with it) |
| resume: guard check disabled (`false && guard != ...`) | `.../the_guard_of_another_park`, `.../the_guard_fix-limit`, `.../the_guard_spelled_as_the_ledger's_key` fail (message check) |
| resume: `ev.Guards = nil` | `TestAFVerbs_N42TheResumeVerbLetsAReclaimCapParkedCardFinishOnItsLandedCommit` fails: `a transition guard has no evidence: resume-guard:reclaim-cap` |
| prune: stale removal disabled (`false && age > max`) | both `AStaleLeftover...` tests fail |
| prune: every temp removed (`age > 0`) | `AFreshTempIsLeftAloneAndCountedInTheTotal` fails |
| prune: fresh temp not added to `sum` | `AFreshTempIsLeftAloneAndCountedInTheTotal` fails |
| move help: `the separate "resume" verb` becomes `which has no verb` | `TestAFVerbs_N45TheHelpOfResumeAndMoveMatchesWhatTheVerbsDo` fails |

Final checks:

```
$ gofmt -l internal/cli internal/agenticfactory            -> (no output), exit 0
$ go build ./...                                           -> exit 0
$ go vet ./internal/agenticfactory/... ./internal/cli/     -> exit 0
$ GOOS=windows go build ./...                              -> exit 0
$ GOOS=linux go build ./...                                -> exit 0
$ go test -race -count=1 -p 4 ./internal/agenticfactory/...  -> exit 0, 21 packages "ok", no FAIL; among them
    ok  .../agenticfactory/ledger  155.270s
    ok  .../agenticfactory/mergeq  304.636s
    ok  .../agenticfactory/runner   26.389s
$ go test -race -count=1 -cover ./internal/agenticfactory/mergeq/
    ok  .../agenticfactory/mergeq  199.429s  coverage: 88.2% of statements
$ golangci-lint run ./internal/agenticfactory/mergeq/...   -> 0 issues.
$ go test -count=1 -run '^(TestAFCommand_|TestAFVerbs_|TestAFLand_|TestAFReap_|TestCardFlow_|TestAFResume|TestAFHolder|TestAFReconcile)' ./internal/cli/
    ok  github.com/modu-ai/moai-adk/internal/cli  162.178s     (61 tests matched by `go test -list`)
$ go test -count=1 -run '^(TestAFClaim_|TestAFUnknownExpired_|TestAFUnknownHolder_)' ./internal/cli/
    ok  github.com/modu-ai/moai-adk/internal/cli  1.549s
$ golangci-lint run ./internal/cli/
    internal/cli/codex_close_entry_signal_test.go:89:17: Error return value of `pw.Close` is not checked (errcheck)   1 issues
```

The last lint finding is in a file this repair did not touch.

## Baseline-attribution

- Tree: the worktree above, start `46dbcce0d4`; every figure is from this run against the commits listed, not carried from the audit.
- Judging build: `go` 1.26.8 compiled the tree in place (`go build`, `go test`, `go vet`); no installed `moai` binary and no MCP tool measured anything. The audit's tool build `b1806498fe` plays no part.
- The N42 fixture is the audit's state rebuilt as a permanent test (`afN28Landed`, move to REVIEWING, three `reap` calls through the verb with a claim between them); the third reap's `Parked: true` is asserted, not logged.

## Gaps

- Not run, by instruction: the whole `./internal/cli/...` package (only anchored selections above), push, PR, any `MOAI_*_LIVE`, an independent audit. The `./internal/cli/` tests outside the `af*` families were not run.
- `go test -race` was not run for `./internal/cli/` (the instruction asks it for `./internal/agenticfactory/...` only).
- Windows and Linux were checked by `go build` only; no test was run on either. The CLI tests skip on native Windows (`afSkipWithoutSh`); the N43 tests use no OS-specific call.
- No separate RED commit for N45: the help text change rode in the N42 GREEN commit (`7deb693d57`), so its test (`c924c9365e`) was written after the text and proven by the mutation above, not by a failing run.
- At RED, the refusal subtests of N42 passed vacuously (an unknown command is also a refusal); they gain meaning at GREEN through the `want` message check and the mutations above.
- The branch of `prune` where a stale leftover cannot be removed (it then still counts) has no test: it needs a directory that refuses `Remove` while the log write succeeds. Package coverage is 88.2 percent with it uncovered.
- N38 (repo binding in the ledger) is not done, so `af resume` makes no repository or trunk check: it touches only the ledger given by `--db` or the project default, and the land help claims nothing about binding beyond today's code.
- `golangci-lint run ./internal/cli/` was run once and shows one pre-existing finding outside the files changed here; it was not fixed (scope).
- REQ-AFL-014, SPEC files and the other open items (N4-N8, N10, N19, N20, F8, F9, F12, E6, N37, N38, N39, N40, N41, N44 rest, N46-N48) were not touched.

## Residual-risk

- `af resume` hands the ledger its guard inside an `Evidence` (the only way `Store.Move` takes a guard), so it records one passing `gate` row with subject `n/a` at the parked epoch, as the ledger's own test of this resume does. It matches no commit, so no judgment on a real commit reads it, but it is visible in an evidence listing of the card. Removing it needs a ledger API change (lane-1's tree).
- `failureLogTmpMaxAge` is one minute against file modification times: a clock step of more than a minute, or a filesystem with coarse or wrong times, could make a live temp file look stale (the write then fails at rename and is reported as a failed log write) or keep a leftover longer. A write fills and renames its temp file within milliseconds.
- A user file whose name starts with `.tmp-` in the log directory is treated as a temp file by prune (counted, and removed when old). The directory belongs to the failure log.
- The resume verb cannot tell whether the operator really looked: `--reason` is free text recorded as given.
