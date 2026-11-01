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

---

# Follow-up 2: N38 first bind (leader's re-measurement, card issue-1948)

Start HEAD `54e7161211`. Same single writer, same tree. No push, no PR, no `MOAI_*_LIVE`, no run of the whole `./internal/cli/...` package, no independent audit. `internal/agenticfactory/ledger/**` untouched (its API was sufficient: `Repository`, `BindRepository`, `Leases`). No SPEC file touched.

## Claim

1. Before a verb binds a ledger that has no binding yet, it compares the repository it would bind (same identity: symlink-resolved git common directory) with what the ledger already knows, and refuses a disagreement with a message naming both, a non-zero exit and the ledger unchanged and still unbound (`afCheckKnownRepositories`, called by `afBindRepository` only while `st.Repository` is empty). What the ledger knows: (a) the repository of every lease worktree in `st.Leases`, a worktree that is gone, is no repository or that git cannot answer for being skipped; (b) the repository the `--db` file sits in (its directory), when it is inside one.
2. A bound ledger behaves as before (compared with the binding). Same-repository linked worktrees and symlinked spellings pass. When the ledger knows nothing (no lease with a worktree that is a repository, a file outside any repository) the first call decides, as before: a documented residual risk, pinned by a test.
3. `afRun` and `afMergeRun` pass the ledger path; `afCheckCommit` passes "" (the path is not at hand there; `afRun` has already tried to bind the project directory, and the lease worktrees are still compared). The `afBindHelp` paragraph of all nine verbs says it.

## Evidence

Commits on top of `54e7161211`, oldest first (RED, RED, GREEN):

```
b5932a8bfb test(af): N38 first-bind RED, a ledger that is not bound yet is refused a repository that disagrees ...
ea9b27f010 test(af): N38 first-bind RED, the help of the verbs says what an unbound ledger is compared with ...
7f957917ab fix(af): N38 first-bind GREEN, before a verb binds an unbound ledger it compares the repository ...
```
(the commit adding this section follows; the final HEAD is reported with the hand-over)

RED (`go test -count=1 -run '^TestAFVerbs_N38FirstBind' ./internal/cli/`, tree before the fix): `FAIL` for `WorkingDirectoryNamesAnotherRepository`, `EnvNamesAnotherRepositoryRefusesEveryVerb/*` (9), `WorkingDirectoryRefusesEveryVerb/*` (9), `LandFromASecondCloneOfTheProjectIsRefused`, `ReconcileFromAnotherRepositoryIsRefused`, `LeaseWorktreeNamesAnotherRepositoryRefusesEveryVerb/*` (9), `LedgerFileInsideAnotherRepositoryIsRefused`; `WithNothingKnownBindsAsBefore` and `LinkedWorktreeOfTheProjectPasses` passed (guards against over-refusal). The help test failed for all nine verbs.

GREEN, `-run '^TestAFVerbs_N38FirstBind'` -> `ok`. Every test in the next list `--- PASS`, run with `-v`:

- Port of the audit's `TestA3Repo_WorkingDirectoryNamesAnotherRepository`: `TestAFVerbs_N38FirstBindWorkingDirectoryNamesAnotherRepository`. Evidence green, evidence gate, enqueue, land, finish (MERGING then DONE) are all refused, each with the ledger snapshot unchanged and `Repository()==""`; the card is not DONE; the project repository does not hold `x`; the queue is empty; the control from the project binds the project.
- The five tests of the §2 N38 row exist as permanent tests from the last round under these names (the audit's `TestA3Repo_*` names do not exist in the tree). Verbatim result lines of the run:
  ```
  --- PASS: TestAFVerbs_N38LandFromASecondCloneOfTheProjectIsRefused (1.04s)
  --- PASS: TestAFVerbs_N38EveryVerbIsRefusedWhenTheProjectDirectoryNamesAnotherRepository (0.94s)
  --- PASS: TestAFVerbs_N38EveryVerbIsRefusedWhenTheWorkingDirectoryIsAnotherRepository (0.66s)
  --- PASS: TestAFVerbs_N38ReconcileFromAnotherRepositoryIsRefused (0.65s)
  --- PASS: TestAFVerbs_N38ALeaseWorktreeInAnotherRepositoryIsRefusedWhereItIsUsed (0.62s)
  ```
  Those five bind the ledger first through a verb in the project (`afN38Card`), so they pin the bound-ledger refusal. Their first-bind counterparts, added now on an UNBOUND ledger (card claimed and rows written through the store): `TestAFVerbs_N38FirstBindEnvNamesAnotherRepositoryRefusesEveryVerb`, `...WorkingDirectoryRefusesEveryVerb`, `...LandFromASecondCloneOfTheProjectIsRefused`, `...ReconcileFromAnotherRepositoryIsRefused`, `...LeaseWorktreeNamesAnotherRepositoryRefusesEveryVerb` (all nine verbs each where the original does).
- (3) `TestAFVerbs_N38FirstBindLedgerFileInsideAnotherRepositoryIsRefused`: file in repository A, project B -> `reap` refused naming both, still unbound; project A binds A; a linked worktree of A as project binds A.
- (4) `TestAFVerbs_N38FirstBindWithNothingKnownBindsAsBefore`: no lease / lease worktree gone / lease worktree no repository, file outside any repository, project another repository -> binds it (residual risk, explicit).
- (5) `TestAFVerbs_N38FirstBindLinkedWorktreeOfTheProjectPasses`: lease in a linked worktree, first call from the project, the linked worktree and a symlink to the project, each on a fresh unbound ledger -> binds the project's git directory.
- Help: `TestAFVerbs_N38FirstBindHelpSaysWhatAnUnboundLedgerIsComparedWith` (all nine verbs).

Mutation contrasts (break, observe `FAIL`, restore; `git status --short` was empty afterwards):

| Mutant | Failing |
|---|---|
| lease clue skipped (`\|\| true` on the worktree test) | `WorkingDirectoryNamesAnotherRepository`, `EnvNames...`, `WorkingDirectoryRefuses...`, `LandFromASecondClone...`, `ReconcileFromAnother...`, `LeaseWorktreeNames...` |
| ledger-file clue skipped (`&& false`) | `LedgerFileInsideAnotherRepositoryIsRefused` only |
| an unanswerable clue refused instead of skipped (`err == nil && got == id`) | `WithNothingKnownBindsAsBefore` (3 subtests), `LinkedWorktreeOfTheProjectPasses` (3), `WorkingDirectoryNamesAnotherRepository` (its control) |
| help sentence reworded | the help test |

Verification, run from the worktree root on the final code (the report commit changes no code):

```
gofmt -l internal/agenticfactory internal/cli         -> (no output)
go build ./...                                        -> rc=0
go vet ./internal/agenticfactory/... ./internal/cli/  -> rc=0
GOOS=windows go build ./...                           -> rc=0
GOOS=linux go build ./...                             -> rc=0
go test -race -count=1 -p 4 ./internal/agenticfactory/...   -> all 21 packages ok (ledger 52.330s, mergeq 68.440s)
go test -count=1 -p 4 -run '^(TestAF|TestCardFlow)' ./internal/cli/  -> ok  github.com/modu-ai/moai-adk/internal/cli  86.180s
go test -race -count=1 -cover ./internal/agenticfactory/mergeq/  -> ok  .../mergeq  57.497s  coverage: 88.2% of statements
golangci-lint run ./internal/agenticfactory/... ./internal/cli/   -> 1 finding: internal/cli/codex_close_entry_signal_test.go:89:17 errcheck (pw.Close), a file this work did not touch
```
`go test -list 'TestAFVerbs_N38'` lists 18 tests (10 new, 8 from the last round); the anchored AF run covers all of them. The unmodified N3, N28, N39, N42, N38 and repair tests are in that run and green.

## Baseline-attribution

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, start `54e7161211`, code HEAD `7f957917ab`. Every number and line above is from this run on this tree. Judging build: `go` compiled the tree in place; no installed `moai` binary measured anything. The audit's repro file was read, not run.

## Gaps

- Refused or not run, by instruction: the whole `./internal/cli/...` package (only the anchored `^(TestAF|TestCardFlow)` and narrower anchored subsets), push, PR, any `MOAI_*_LIVE`, independent audit. `go test -race` was not run for `./internal/cli/`. `internal/cli` coverage was not measured.
- Refused by the worktree guard in this session: one `sed -i` in a plain command (redone with Edit), one compound `go test` call with a redirect into `/tmp`, and one compound `cat >>` append of this section (all redone as plain commands or with Edit); no result depends on them.
- Windows and Linux by `go build` only; the CLI tests skip on native Windows. Identity behaviour of `EvalSymlinks` on Windows and case-insensitive volumes is untested.
- `afCheckCommit` has no ledger path, so the ledger-file clue is checked only where `afRun` and `afMergeRun` bind. In the one route where `afRun` could not bind (project directory is no repository) and `afCheckCommit` binds first, only the lease clue applies; `--repo` is refused earlier there (`afCheckRepoFlag` needs a project repository), so a lease-less ledger whose file sits in repository A, with a `--repo`-less `evidence` from a non-repository project directory, is the uncovered corner.
- `internal/hook/ledger_heartbeat.go` still opens the ledger with no repository comparison (not a `moai af` verb; unchanged).
- Not measured: two processes first-binding the same fresh ledger from two repositories at the same instant. The known-identity read and the `BindRepository` write are not one transaction; the write stays atomic and refuses a different identity, but the pre-check can let a call through that a just-written lease would have refused.
- The unbound fixtures write evidence, queue rows and claims through the store, never through a verb, to leave the ledger unbound.
- The audit's own `TestA3Repo_*` file was not copied into the tree and not run; its scenarios are covered by the permanent tests named above.

## Residual-risk

- A ledger that knows nothing (no lease worktree that is a repository, file outside any repository: the usual state of the default `<home state>/af/ledger.db` before the first claim) is still bound by the first call; a wrong first call (an operator typo in `CLAUDE_PROJECT_DIR`) binds the wrong repository, later refusals name both repositories, and nothing undoes it.
- A lease whose worktree is in another repository now makes EVERY verb refuse at first bind (claim, release and reap included) until the ledger is edited by hand; that is the reading of the brief, and it means a mis-claimed worktree cannot be cleaned up through the verbs while the ledger is unbound.
- A ledger file in a directory that happens to be a repository (the default path is not) is taken as a clue and can refuse a legitimate project that is a different repository.
- Identity is the symlink-resolved git directory path, so a repository moved on disk reads as another repository.

---

# Follow-up 3: repair after the audit of 1c05083f71 (N49, N50, N54, N51 b/c/d and clone form, N59)

Scope change by the operator via the leader: the accepted-risk record (R20) is NOT written into the SPEC; SPEC edits are only corrections of sentences the code made false. N52 (repository-move recovery verb) is NOT started: it needs lane-1's ledger API, which is not merged.

## Claim

1. N49: evidence and enqueue from a lease worktree are refused before the first bind when the `--db` ledger file sits in another repository.
2. N50: the ledger file is a first-bind clue only for a ledger given with `--db`; a default ledger under a `MOAI_HOME` that is a repository binds the project.
3. N54: `af land` refuses while the newest green, gate or review run of the candidate failed, names the kind and the way out, and lands after it is measured again; a passing gate of another commit does not block (D5).
4. N51 (b) rows naming commits the caller's repository lacks, (c) symlinked `--db`, (d) relative lease worktree, and the clone form (`--bind-repo`) are refused at the first bind with the ledger unchanged and unbound.
5. N59: `moai af add <id>` puts a READY card into the ledger; a card added only by the verb runs claim, evidence, enqueue, land, DONE through the verbs.
6. Not claimed: N52 (not started), a PASS of any audit, E6 (unchanged).

## Evidence

Commits since `1c05083f71` (RED then GREEN per item; HEAD at the time of writing is the docs commit `693f4a35b3`, the report commit follows it):

```
2a0c9fb455 test N49 RED   a920857df4 fix N49 GREEN
f6674dc628 test N50 RED   5c389902ec fix N50 GREEN
477938d926 test N54 RED   7d6eeb8cc8 fix N54 GREEN (+ Store.FailingNewestRun)
b3c9104ee1 test N51c RED  7fe486101e fix N51c GREEN
7fb3ee9507 test N51d RED  5e775c89a5 fix N51d GREEN
220da0d4a9 test N51b RED  6e02850e3b fix N51b GREEN
95e81a2e74 test N51 clone RED   d247597ae6 fix N51 clone GREEN (two existing tests changed)
3034faeade test N59 RED   e15419a87c feat N59 GREEN
693f4a35b3 docs(spec) corrections only
```

RED output, verbatim where it decided (each run before the fix, by behavior and not by compile error, except the N59 and clone controls that need the new verb or flag):

```
N49   evidence green from W: accepted, output {"card_id":"t1","kind":"green","result":"pass",...}
      enqueue from W: accepted ...   the queue must be empty: [{CardID:t1 ... State:waiting ...}]
N54   landed with a failing newest gate run: {"card_id":"t1","outcome":"landed",...}   (4 subtests FAIL)
N51c  reap through the link from O: accepted, output {"reclaimed":[],"unknown_expired":[]}
N51d  claim --worktree "sub/dir" left a lease: [{CardID:t1 ... Worktree:. ...}]
      reap with a lease worktree recorded as .: accepted, output {"reclaimed":[{"CardID":"t1",...}]...}
N51b  reap from a repository without the commits: accepted, output {"reclaimed":[],"unknown_expired":[]}
N51cl claim declaring the clone: unknown flag: --bind-repo
N59   moai af add t1 ...: unknown command "add" for "af"
N50   [reap] with the default ledger under a home repository: af: refused: this ledger is not bound yet, but it already belongs to the repository .../001/.git (the ledger file .../.moai/db/002-92f0daf0/af/ledger.db sits in it) ...
```

GREEN (this tree, code HEAD `e15419a87c`; the docs commit changes no code):

```
gofmt -l internal/agenticfactory internal/cli                  -> (no output)
go build ./...                                                 -> exit 0
go vet ./internal/agenticfactory/... ./internal/cli/           -> exit 0
go test -race -count=1 -p 4 ./internal/agenticfactory/...      -> 21 packages ok (ledger 50.311s, mergeq 72.327s)
go test -count=1 -v -run '^(TestAF|TestCardFlow)' ./internal/cli/  -> 107 top-level "--- PASS", 0 SKIP, 0 FAIL; ok
```

New permanent tests, each run by anchored name and passing: `TestAFVerbs_N49CommitCheckAsksTheLedgerFileBeforeTheFirstBind`, `TestAFVerbs_N50TheDefaultLedgerLocationIsNoFileClue`, `TestAFVerbs_N50HelpSaysTheFileClueIsForADbFlagOnly`, `TestAFLand_N54AFailingNewestRunWhileTheRowWaitsRefusesTheLanding` (4 subtests), `TestAFLand_N54APassingGateOfAnotherCommitDoesNotBlockTheLanding`, `TestAFVerbs_N51cASymlinkedLedgerFileIsAskedWhereItReallyIs`, `TestAFVerbs_N51dClaimRefusesARelativeWorktree`, `TestAFVerbs_N51dARelativeLeaseWorktreeAlreadyRecordedIsNotTrustedAtTheFirstBind`, `TestAFVerbs_N51bRowsNameCommitsTheCallerDoesNotHave`, `TestAFVerbs_N51bOnlyCompleteShasAreCluesFromTheRows`, `TestAFVerbs_N51WrongFirstCallFromASecondCloneNeedsADeclaration`, `TestAFVerbs_N51ADeclarationIsNotAskedWhereAnythingElseSaysWhereTheLedgerBelongs`, `TestAFAdd_ACardAddedByTheVerbRunsFromClaimToDone`, `TestAFAdd_TheMinimalCallAndTheRefusalsOfTheLedger`, `TestAFAdd_TheHelpSaysWhatTheVerbTakes`, and in the ledger package `TestFailingNewestRun_NamesTheKindWhoseNewestRunFailed` (9 subtests; the one permitted ledger edit, a read-only `Store.FailingNewestRun` over the existing `noFailingNewestRun` rule).

Mutation contrasts (fix broken, test seen failing, fix restored, test passing again): N49 (argument back to `""`), N50 (`path` instead of the flag), N54 (`kind != "" && false`; 4 subtests fail), N51c (symlink resolution disabled), N51d (claim check off; lease check off, each caught by its own test), N51b (`missing` never returned; the SHA filter dropped, caught by the rows test), clone form (refusal disabled; declared-mismatch check disabled), N59 (title default off).

Existing tests changed (each strengthens or follows a rule the operator changed; none weakened, none skipped):

1. `internal/cli/af_flow_test.go` `TestCardFlow_ClaimToDone_ThroughMergeQueue`: the first `claim` on a fresh `--db` outside any repository now passes `--bind-repo <repo>`, because that ledger knows nothing (clone form).
2. `internal/cli/af_repo_n38_firstbind_test.go`: `TestAFVerbs_N38FirstBindWithNothingKnownBindsAsBefore` renamed `...NeedsADeclaration`; it pinned the residual the operator now closed. It now asserts the refusal without a declaration (naming `--bind-repo` and the repository, ledger unbound and unchanged) and the bind with `--bind-repo`. The header comment was updated to match.
(No other pre-existing test was touched. One of my own new tests, `af_add_test.go`, was corrected after its RED commit: the "another repository" check now passes `--db`, because the default location is one ledger per project.)

The N51d "justify the smallest place" decision: `claim` (`internal/cli/af.go` `newAFClaimCmd`, the only verb that records a lease worktree) refuses a non-absolute `--worktree`; a relative worktree already recorded by an older build or a direct store write is refused in `afCheckKnownRepositories` at the first bind, because `git -C <relative>` runs in the caller's directory and always agrees with the caller. A bound ledger needs no third check: its comparison is with the bound identity.

The clone form decision: no existing flag declares a repository. `--repo` exists only on evidence, enqueue, land and reconcile (not on claim, move, release, reap, resume, which the repro's first verb is) and `afCheckRepoFlag` (`internal/cli/af.go`) compares it with the project directory's repository, so it declares nothing the project directory has not said. A smaller method does not exist: nothing in an unbound `--db` ledger outside any repository says where it belongs. The new flag is `--bind-repo <path>` on every verb that opens the ledger (`afFlags.bind`, `afMergeFlags.bind`), checked against the repository the call would bind, used for the first bind only, documented in `afBindHelp` and the flag help. The default ledger location needs none.

## Baseline-attribution

Tree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`, start `1c05083f71`, code HEAD `e15419a87c`. Every command and output above was run in this session against this tree. The judging build is the `go test` build of the tree itself (no installed `moai` binary was used).

## Gaps

- Not run, by instruction: the whole `./internal/cli/...` package (only the anchored `^(TestAF|TestCardFlow)` and the anchored new-test list), push, PR, any `MOAI_*_LIVE`, independent audit. `golangci-lint` and `GOOS=windows`/`linux` builds were not re-run in this round.
- The worktree guard refused three compound shell calls (a `cd ... &&` chain with a heredoc, and two chained python heredocs); they were redone as plain calls or with the Edit tool. One scratch file was written under `/tmp` and removed at once.
- N52 (repository-move recovery verb): not started (pending lane-1's ledger API). E6 unchanged: `TestA2Done_TrunksOwnCommitAsCandidate` is an audit file that is not in this tree and was not run; N54 can only refuse a landing when a newest run failed, so a candidate already in the trunk with passing runs lands as before (not measured against that test).
- The hook heartbeat recorder still opens the ledger with no repository comparison (unchanged).
- Windows behaviour (`EvalSymlinks`, `filepath.IsAbs`) and case-insensitive volumes (N56) untested.
- Two processes first-binding one fresh ledger at the same instant: still not one transaction (unchanged).
- The row-commit clue reads evidence and queue rows with several store calls and one `git cat-file --batch-check`; its cost on a very large ledger was not measured.

## Residual-risk

- `--bind-repo` is a declaration, not a proof: a caller in the wrong clone who types it binds the wrong repository. It turns a silent first call into a deliberate one; it does not remove the risk. The default location is not asked at all, so a project whose default ledger first-binds from a wrong clone binds the clone (the default location is per project key, so a second clone usually has its own ledger).
- A first bind is refused when ANY complete SHA in the rows is missing from the caller's repository, including a commit that was legitimately pruned; `--bind-repo` does not lift that refusal (the rows clue is checked whether or not a declaration was given), so the way out is a call from a repository that has the commits. This affects only a ledger that is still unbound.
- N55 (a refused verb still binds), N56 (case variants), N58 (surviving mutants) and N52 are unchanged. N57 sentences not made false by this round (the REQ-AFL-014 `land` sentence) are untouched.
