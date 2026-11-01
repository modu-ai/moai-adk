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

---

# Follow-up: N38 verb side, N39 verb side, N45 remainder (card issue-1948)

Start HEAD of this follow-up `0afd4c4e7e` (the `--no-ff` merge of lane-1's `bb17e7f6eb`). Same single writer, same tree. No push, no PR, no `MOAI_*_LIVE`,
no run of the whole `./internal/cli/...` package, no independent audit. `internal/agenticfactory/ledger/**` untouched (its API was sufficient). No SPEC file touched.

## Claim

1. N38: every `moai af` verb binds the ledger to the repository it will use and refuses another, with a message naming both, a non-zero exit and the ledger unchanged.
   Identity = absolute, symlink-resolved git common directory (`afGitCommonPath`), so a linked worktree and a symlinked spelling of the project are the same repository.
   Which repository each verb uses, in the existing precedence (`--repo`, else the card's lease worktree, else the project directory):

   | Verb | Repository bound / compared | Where |
   |---|---|---|
   | claim, move, release, reap, resume | project directory (`CLAUDE_PROJECT_DIR`, else the working directory); a project directory that is no git repository binds nothing (these verbs run no git) | `afRun` |
   | evidence, enqueue | the project directory (`afRun`) AND `--repo`, else the lease worktree, else the project directory (`afCheckCommit` -> `afCommitRepo`); not being in a git repository is a refusal here | `afRun`, `afCheckCommit` |
   | land, reconcile | `--repo`, else the project directory (what `afMerger` uses); not being in a git repository is a refusal | `afMergeRun` |

   `afCheckTrunk` and `afCheckRepoFlag` (N26) run unchanged before the ledger is opened. Verbs found: `claim`, `move`, `evidence`, `enqueue`, `release`, `reap`, `land`, `reconcile`, `resume` (nine; `newAFCmd` registers exactly these).
2. N39: `af enqueue` asks the repository `git merge-base --is-ancestor <candidate> refs/heads/af/trunk` (`mergeq.GitCLI.IsAncestor`: exit 0 yes, 1 no, anything else an error -> refusal, nothing changed) and passes `ledger.OnTrunk(result)` to `Enqueue`.
   `af move --stage DONE` finds the card's queue row; when it is `landed` it asks the same question about the row's candidate (in the project directory's repository) and refuses when it is not in the trunk, or when git cannot answer. Exit: withdraw (REVIEWING), release, the next holder measures and offers the commit, the row starts again as `waiting`, `af land` lands it, then DONE. A landed row whose commit IS in the trunk is still taken over (N3 behaviour kept).
3. N45 remainder: the `Long` of all nine verbs ends with the same binding paragraph (`afBindHelp`), `af land` no longer says "so the verb does not let the caller choose" (that was true of `--repo` only), `enqueue` and `move` say what they ask the trunk, `claim` says `--worktree` is not checked at claim, `--repo` help says a repository other than the ledger's is refused. This supersedes claim 3 of the section above for the land sentence.

Non-test callers (grep over `internal` and `cmd`, `.go` files without `_test.go`):

- `Store.Enqueue`: one, `internal/cli/af.go` (the enqueue verb). `mergeq` calls none.
- Transitions to DONE: `Store.Move` has two non-test callers, both in `internal/cli/af.go`: the move verb (guarded) and the resume verb (an in-place move at the same stage, never DONE). `core.StageDone` is otherwise only read (`monitor/flow`, `monitor/pr`, `reconcile`, `web`). Inside the ledger it appears in `move.go` (lease cleanup), `claim.go` and `queue.go` (reads).
- Openers of the ledger: `internal/cli/af.go` (`afRun`), `internal/cli/af_land.go` (`afMergeRun`), `internal/agenticfactory/mergeq/ledgerqueue.go` (a queue adapter over an open store), and `internal/hook/ledger_heartbeat.go` (see Gaps).

Commits on top of `0afd4c4e7e`, oldest first (RED test commit, then GREEN):

```
452f8b999b test(af): N38 RED ...
863aeb76f7 test(af): N39 RED, enqueue hands the ledger whether the candidate is in refs/heads/af/trunk ...
78fda9263a test(af): N39 RED, the exit of a card refused DONE because the trunk lost its commit ...
319b39c1cc fix(af): N38 GREEN ...
fe59cd74ea fix(af): N39 GREEN ...
2933943321 test(af): N45 RED ...
a338051e60 fix(af): N45 GREEN ...
```

(the commit adding this section follows; the final HEAD is reported with the hand-over)

## Evidence

Run from the worktree root, HEAD `a338051e60` (code identical to the final HEAD; the report commit changes no code).

RED observations (before each GREEN, same tree plus the new test file):

- N38 (`go test -count=1 -run '^TestAFVerbs_N38' ./internal/cli/`): `FAIL` for `LandFromASecondCloneOfTheProjectIsRefused`, `EveryVerbIsRefusedWhenTheProjectDirectoryNamesAnotherRepository/{claim,move,evidence,enqueue,release,reap,land,reconcile,resume}`, the same nine under `...WorkingDirectoryIsAnotherRepository`, `ReconcileFromAnotherRepositoryIsRefused`, `ALeaseWorktreeInAnotherRepositoryIsRefusedWhereItIsUsed`, `TheFirstVerbBindsTheLedger...`; `TheSameRepositoryByAnotherRouteIsAccepted` and `AVerbThatRunsNoGit...` passed (they guard over-refusal).
- N39: failing: `ALandedRowWhoseCommitIsInTheDedicatedTrunkIsTakenOver` (`the row is waiting, want landed (taken over)`: the ledger's new default restarts a landed row without `OnTrunk`), `EnqueueRefusesWhenTheRepositoryCannotAnswer...` (`the offer was accepted with no trunk to ask`), `TheMoveToDoneAsksWhetherTheLandedCommitIsInTheTrunk` (`DONE was accepted with the landed commit outside the dedicated trunk`), `ACardRefusedDoneBecauseTheTrunkLostItsCommitCanBeLandedAgain` (setup: DONE accepted). The same run showed the merged tree breaking `TestAFVerbs_N3AReclaimedCardFinishesOnTheLandedCommitOnceItMeasuredItAgain` and `TestAFVerbs_N42TheResumeVerbLetsAReclaimCapParkedCardFinishOnItsLandedCommit` (E12 needs merged), because lane-1's ledger no longer takes a landed row over without `OnTrunk`; N39 GREEN repairs both. `TestAFVerbs_N39ARowLandedOnAnotherTrunk...` and `...ATrunkThatLostTheLandedCommit...` PASS at RED too (the ledger default already restarts the row); they are guards against an over-eager `OnTrunk(true)` and are proven by mutant M6.
- N45: `FAIL` for the three help tests, e.g. `moai af reap --help does not say "is bound to one repository"`.

GREEN: `go test -count=1 -v -run '^TestAFVerbs_(N39|N3A|N42)' ./internal/cli/` -> `ok`, and `-run '^TestAFVerbs_N38'` -> `ok`, `-run '^TestAFVerbs_N45'` -> `ok`.

Anchored AF run of the CLI package, `go test -count=1 -p 4 -v -run '^(TestAF|TestCardFlow)' ./internal/cli/` (counts of `--- PASS` by test-name prefix; `go test -list 'TestAF|TestA2|TestA3|TestCardFlow'` lists 82 tests, no `TestA2` or `TestA3`, there is no such prefix in this tree):

```
   2 --- PASS: TestAFClaim        1 --- PASS: TestAFCommand       3 --- PASS: TestAFHolder
  34 --- PASS: TestAFLand         2 --- PASS: TestAFReap          2 --- PASS: TestAFReconcile
   1 --- PASS: TestAFUnknownExpired   1 --- PASS: TestAFUnknownHolder
  35 --- PASS: TestAFVerbs        1 --- PASS: TestCardFlow
   1 ok  	github.com/modu-ai/moai-adk/internal/cli	100.592s
```

82 passed, 0 `--- FAIL`, 0 `--- SKIP` (82 = the listed count). The N3 tests (`af_land_n3_test.go`) are unchanged and green.

Other commands (all exit 0): `gofmt -l internal/agenticfactory internal/cli` printed nothing; `go build ./...`; `go vet ./internal/agenticfactory/... ./internal/cli/`; `GOOS=windows go build ./...`; `GOOS=linux go build ./...`; `golangci-lint run ./internal/agenticfactory/...` -> `0 issues.`; `golangci-lint run ./internal/cli/` -> one finding, `internal/cli/codex_close_entry_signal_test.go:89:17: Error return value of pw.Close is not checked (errcheck)`, in a file this work did not touch.

`go test -race -count=1 -p 4 ./internal/agenticfactory/...`: 21 packages `ok` (ledger 80.124s, mergeq 101.584s, core 104.645s, others shorter; no failure). `go test -race -count=1 -cover ./internal/agenticfactory/mergeq/` -> `ok  github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	127.361s	coverage: 88.2% of statements`.

Mutation contrasts (each mutant applied to the committed code, the named tests run, the file restored with `git checkout -- <file>`; `git status --short` clean after each):

| # | Mutant | Failing tests (observed) |
|---|---|---|
| M1 | `afRun` does not bind | `EveryVerb...ProjectDirectory...` and `...WorkingDirectory...` subtests claim, move, evidence, enqueue, release, reap, resume |
| M2 | `afMergeRun` does not bind | `LandFromASecondClone...`, `ReconcileFromAnotherRepository...`, subtests land and reconcile of both EveryVerb tests |
| M3 | `afCommitRepo` ignores the lease worktree | `ALeaseWorktreeInAnotherRepositoryIsRefusedWhereItIsUsed` |
| M4 | no `EvalSymlinks` | `TheFirstVerbBindsTheLedger...` and every EveryVerb subtest (`/var` vs `/private/var`) |
| M5 | `afRun` bind is not optional for a non-repository | `AVerbThatRunsNoGit...` plus `TestAFReap_*` and `TestAFClaim_*` |
| M6 | `OnTrunk(true \|\| onTrunk)` | `N39ARowLandedOnAnotherTrunk...`, `N39ATrunkThatLostTheLandedCommit...`, `N39ACardRefusedDone...` |
| M7 | `OnTrunk(false && onTrunk)` | `N39ALandedRowWhoseCommitIsInTheDedicatedTrunk...`, `N3AReclaimedCardFinishesOnTheLandedCommit...`, `N42...FinishOnItsLandedCommit` |
| M8 | `afInTrunk` turns a git error into "not in trunk" | `N39EnqueueRefusesWhenTheRepositoryCannotAnswer...` |
| M9 | DONE guard disabled | `N39ACardRefusedDone...`, `N39TheMoveToDoneAsks...` |
| M10 | `reap` help loses the binding paragraph | `N45EveryVerbThatOpensTheLedgerSaysItIsBoundToOneRepository` |

All ten caught.

## Baseline-attribution

- Tree: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, start `0afd4c4e7e`, code HEAD `a338051e60`; every number above is from this run on this tree, none carried from the audit or from lane-1's report (lane-1's 87.7 percent ledger coverage was not re-measured here, the ledger package was run only inside the race run, without `-cover`).
- Judging build: `go` 1.26.8 compiled the tree in place; no installed `moai` binary or MCP tool measured anything (the running MCP server is `b1806498fe`, unrelated). `golangci-lint` is the installed binary of the shell path (its commit is not a repository build).
- The audit's repro tests were read from `.moai/reports/issue-1948/final-code-audit-46dbcce0d4/repro/`; the permanent tests are rewritten from them (new names, real assertions), none of the audit files was run.

## Gaps

- Refused or not run, by instruction: the whole `./internal/cli/...` package (only the anchored `^(TestAF|TestCardFlow)` selection plus narrower anchored subsets), push, PR, any `MOAI_*_LIVE`, independent audit. `go test -race` was not run for `./internal/cli/`. Tests outside the AF families in `internal/cli` were not run.
- Two tool calls in this session were refused by the worktree guard (a compound python-in-shell call containing git words, and a `sed -i` edit in a plain command); both were redone with the Edit tool and no result depends on them. A redirect to the scratchpad of a go test run was refused as a cross-tree write; the background-task output files were read instead.
- Windows and Linux were checked by `go build` only; the CLI tests skip on native Windows (`afSkipWithoutSh`). The identity helper's `EvalSymlinks` and case behaviour on Windows and on case-insensitive volumes was not tested.
- `internal/hook/ledger_heartbeat.go` opens the ledger (`ledger.Open`) and renews a lease (`Heartbeat`) with no repository comparison. It is not a `moai af` verb, writes only the lease expiry of the card named in the session's binding, and was left alone (scope); a hook reading a binding written for another repository is not refused.
- `claim --worktree <path>` accepts a path in another repository and a non-repository path; the check happens where the worktree is used (evidence, enqueue), not at claim. Stated in the claim help.
- The DONE guard asks the project directory's repository (what `afRun` bound), not the card's lease worktree; the two are the same repository once bound, but the lease worktree path itself is not read at DONE.
- The guard acts on the one queue row of the card when it is `landed`. A card finished with no landed row (the ledger decides that case) is not asked about the trunk.
- A same-epoch landed row cannot be offered again by the epoch that landed it (the ledger says "has landed"), so the exit for a trunk that lost the commit passes through withdraw, release and a new claim; it is tested end to end, it is not a one-step repair.
- Not measured: two processes binding the same fresh ledger from two different repositories at the same instant (SQLite serialises the writer; lane-1 tested two goroutines only).
- Coverage of `internal/cli` was not measured.
- REQ-AFL-014's `configurable` wording and every item named open by the leader (N4-N8, N10, N19, N20, F8, F9, F12, E6, N37, N44 rest, N46-N48) were not touched.

## Residual-risk

- The first writer decides the binding. A wrong repository named at the first verb (an `evidence` with the lease worktree of another repository, a first `claim` from the wrong directory) is bound and every later correct call is refused; the fix is an operator edit of the ledger file. Moving the repository on disk changes its git directory path and later calls refuse until the ledger is rebound by hand.
- Where the project directory is not inside a git repository, the verbs that run no git (claim, move, release, reap, resume) carry on without binding; a ledger can therefore stay unbound until the first evidence, enqueue, land or reconcile.
- A first verb run in a directory that is itself a different repository from the one the ledger was meant for (an operator typo in `CLAUDE_PROJECT_DIR`) binds the wrong one; the refusal message names both repositories so the mistake is visible, but nothing here undoes it.
- The trunk question is asked of the working repository at the moment of the call. A trunk ref moved after `enqueue` and before `land` is the landing's business (the Merger rebuilds against the current trunk), not this check's.
- DONE after the guard rests on the same equality of commit ids the ledger uses; it proves the commit is an ancestor of `refs/heads/af/trunk`, not that the gate that approved it still holds (E6, open).
