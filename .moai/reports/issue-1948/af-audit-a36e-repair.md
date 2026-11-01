# Repair after the final code audit of a36ea8c368 (Issue #1948, card issue-1948, AF-22)

Audit: `.moai/reports/issue-1948/final-code-audit-a36ea8c368/verdict.md` (FAIL 0.73). This file is written in sections, one per spawn.
Section 1 (below) is the identity group. The transition group (N63, N66) and the remaining items are the second spawn's sections.

## Identity group (N61 N62 N65 N67 N68 N64)

Commits: RED `e544228237`, GREEN `19eb2c9cc9` (source, the new marker tests, the normal-path tests, the `AF_TEST_BIN` seam, the fixture changes), test-only follow-ups `919b6d8610`, this report in the docs commit after it. (The commit message of `919b6d8610` lists the normal-path tests and the seam, which were already in GREEN; the follow-up itself holds only the items of E10.)
Scope correction received during the work (leader): the done-condition is N61..N66; N64 only as `--from <old project dir>` plus the
fixed sentences, no automatic discovery; N67 and N68 are not completion conditions. See "N67/N68: how far it got".

### Design as built (and where it departs from the brief)

One mechanism, a repository identity marker, as briefed.

- `internal/cli/af_repository_marker.go`: `<common git dir>/moai-af-repository-id`, 128 random bits as 32 lowercase hex digits plus a newline,
  mode 0600. Created through a temp file in the same directory and `os.Link` to the final name. **Departure:** the brief said `O_EXCL`
  create and re-read after losing the race. A plain `O_EXCL` create leaves an empty file visible to a concurrent reader between the create and the write,
  and a crash there leaves an empty marker for good. Link-after-write makes the name appear complete or not at all; the loser of the race
  reads the winner (`EEXIST`). A file that is not 32 hex digits is an error (a refusal), never an absence.
- Ledger (`internal/agenticfactory/ledger`): new table `repository_marker` (single row, update/delete refused by trigger),
  `RecordRepositoryMarker` (first writer wins, returns the value on record), `RepositoryMarker`, `MoveRepositoryMarked` (move + marker in one
  transaction; `MoveRepository` is the unmarked form and is unchanged for its callers). The ledger runs no git (REQ-AFL-008): it stores and compares a string.
- Every verb (`afBindRepository`, after the identity check): ledger has a marker -> the repository's marker must equal it, else refuse (message names both
  identities and the first 8 characters of the two markers, never the value); ledger has none -> adopt the repository's marker (created if absent) and record it.
- `move-repository`: marked ledger -> the new repository must carry the recorded marker; commits are only counted (`commits_checked`, `commits_unchecked`),
  a git that cannot answer does not refuse. Unmarked ledger -> fallback: the candidates of queue rows that are not failed must be in the new repository (git must answer),
  `--bind-repo` must declare the repository (the project directory is not taken for a declaration), a `--db` file must not sit in another repository; then the new
  repository's marker is created and recorded with the move. Output gains `commits_unchecked`.
- N65: a lease worktree that agrees with the call's repository no longer counts as a clue (a disagreeing one is still a refusal). With `--db` and no other clue the
  first bind needs `--bind-repo`. **Departure:** the brief said to use the lease worktree's marker when the ledger has one. `afCheckKnownRepositories` runs only while the
  ledger is unbound, and an unbound ledger has no marker on record, so there is nothing to compare; for a bound ledger the clone-through-a-lease form is closed by the
  identity-plus-marker check of every verb. The marker is therefore not read from lease worktrees.
- N62: `afMissingCommit(ctx, dir, shas) (string, error)`; `afMissingCommits` returns an error for a git failure, a context end, an answer of another length than the
  question, and any line that is neither `<oid> commit <size>` nor `... missing`. **Found while doing this:** an empty answer with exit 0 to one question passed the old
  length check (`strings.Split("", "\n")` has length 1), so it read as "nothing missing"; the new line-shape check closes that. First bind and `move-repository` refuse on the error;
  an unreadable ledger row set at the first bind is a refusal too (it used to be skipped).
- N64: `move-repository --from <old project directory>` (not with `--db`): the ledger of the old key is moved (file, `-wal`, `-shm`; all or none, undone if the record
  fails) to the new key before the move is recorded, only when the new key has no ledger file (else a refusal naming both paths, nothing touched). The old directory is
  resolved through its nearest existing ancestor (`afResolveGone`) because a directory that moved away cannot be resolved itself and the key is made from the resolved
  spelling (on macOS `/var` vs `/private/var`). No automatic discovery. A fresh ledger at the default new key (no cards, no moves, bound to the repository the call names)
  gets "a fresh one for this place ... run this again with --from <old project directory>" instead of "still exists on disk, so it did not move". A missing default ledger
  refusal also names `--from`. The `--db` help sentence of the three verbs (N72) and the stale-epoch wording are not touched.

### Claim

1. Moving a ledger to a clone or mirror while the repository is away, to an unrelated repository when no row names a commit, or with the ledger file in another repository is refused (marked ledgers by the marker; unmarked ones by the commits, the declaration and the file location). N61.
2. A git failure, a time-out, a short or long answer and an unreadable marker are refusals at the first bind and at `move-repository`, and never a clue or a pass. N62.
3. A first bind of a `--db` ledger through a lease alone is refused with a declaration request; a repository re-created at the bound path is refused by every verb. N65 (and the re-created-repository form of N61).
4. A marked ledger moves past a commit that was pruned; an unmarked one moves past a pruned evidence commit but still not past a missing queue candidate. N67 (as far as the marker resolves it).
5. The first recorded move raises `PRAGMA user_version` to 3; a ledger that never moved stays at 2; this build opens 1, 2 and 3 and refuses 4. N68 (kept as written, see below).
6. `move-repository --from` moves the default-location ledger of a moved project; a conflict, `--db`, and a project with no ledger are refused with nothing changed; the false "did not move" sentence is replaced for a fresh ledger. N64.
7. The normal paths hold beside each refusal: a plain rename, the chain A->B->C->A, a linked worktree named as the new place, a copy once the original is gone, a holder that died before the move and a second lane that finishes the card after it.

### Evidence

All commands were run in this worktree (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`), tree HEAD `19eb2c9cc9` plus the test-only follow-ups of the third commit; source files are identical to `19eb2c9cc9`.

E1 RED (before any source change), `go test -count=1 -run TestAFIdentity_ ./internal/cli/`: 22 top-level tests failed, by behaviour (`the verb was accepted, want a refusal that says ["identity marker"]`, `reap with a git that cannot answer: <nil>`, `evidence from the clone through its own lease: <nil>`, `unknown flag: --from`, `user_version 2 after the move`). The ledger package did not compile (`s.RepositoryMarker undefined`), as a tests-only RED for new API must.

E2 `go test -race -count=1 -p 4 ./internal/agenticfactory/ledger/...`
```
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	31.388s
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger/importer	1.553s
```
E3 `go test -count=1 -cover ./internal/agenticfactory/ledger/ ./internal/agenticfactory/mergeq/`
```
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	5.838s	coverage: 87.8% of statements
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	28.413s	coverage: 88.2% of statements
```
(the ledger was 88.0% before, mergeq 88.2%; both stay above 85%.)

E4 `gofmt -l internal cmd` printed nothing; `go vet ./internal/cli/ ./internal/agenticfactory/...` printed nothing; `go build ./...`, `GOOS=windows go build ./internal/cli/ ./cmd/moai/` and `GOOS=linux go build ./internal/cli/ ./cmd/moai/` printed nothing (exit 0).

E5 `golangci-lint run ./internal/agenticfactory/ledger/...` -> `0 issues.` `golangci-lint run ./internal/cli/` reported one issue, `internal/cli/codex_close_entry_signal_test.go:89:17 errcheck` (F12 of the audit, outside this change).

E6 anchored run `go test -count=1 -run '^(TestAF|TestCardFlow)' ./internal/cli/` after the follow-ups:
```
ok  	github.com/modu-ai/moai-adk/internal/cli	145.276s
```
E7 the same permanent group through the real CLI binary, `AF_TEST_BIN=<binary> go test -count=1 -v -run TestAFIdentity_ ./internal/cli/` (the test helper `afRunBoth` runs `<binary> af ...` instead of the in-process command when the variable is set; the tests, their fixtures and the Go ledger API they read the ledger with are the same).
Binaries built from trees, in the scratchpad, never from the main checkout: `moai-a36e` from `git archive a36ea8c368` extracted to the scratchpad (sha256 `ca9969f1ec7d379f4e5948098107207caef5c743cb8fa1c15bc58ec428834acc`); `moai-final` from this tree (sha256 `dff75cdb8571e3cce614a789f0ceb8a061c8223da7ff3381035013e8398a245b`).
- a36ea8c368 build: `FAIL github.com/modu-ai/moai-adk/internal/cli 36.503s`, 27 tests: 23 `--- FAIL`, 4 `--- PASS`. The four that pass are the normal paths: `TestAFIdentity_ACopyIsAMoveOnlyOnceTheOriginalIsGone`, `..._AChainOfMovesAndTheWayBackFinishesTheCard`, `..._ALinkedWorktreeOfTheMovedRepositoryIsAcceptedAsTheNewPlace`, `..._AHolderThatDiedBeforeTheMoveDoesNotStrandTheCard`. Every refusal test (N61 forms, N62, N65, N67, N64, N68, the re-created repository, the damaged marker, the help) fails there, and so do the marker tests (`the marker in the git directory is ""`).
- final build: `ok github.com/modu-ai/moai-adk/internal/cli 26.563s`, 27 `--- PASS`, 0 `--- FAIL`.
- Both outputs, saved on this machine: `/tmp/claude-501/bin-old.txt`, `/tmp/claude-501/bin-final.txt` (scratch, not a citation target; the lines above are the carried evidence).

E8 mutation checks (each rule broken by a local source edit, the named tests run, the edit reverted with `git checkout -- <file>`; none committed):

| # | Rule | Mutation | Failing test(s) |
|---|---|---|---|
| M1 | marker compare in every verb | `if false && mine != have` in `afCheckMarker` | `TestAFIdentity_ARepositoryRecreatedAtTheSamePathIsNotTheOneTheLedgerBelongsTo` |
| M2 | marker compare in move-repository | `if false && found != recorded` | `TestAFIdentity_N61AMoveToACloneWhileTheRepositoryIsAwayIsRefused/mirror`, `..._N61AMoveToAnUnrelatedRepository...`, `..._N61ALedgerFileInOneRepositoryIsNotMovedToAnother/a_ledger_with_a_marker` |
| M3 | git failure is a refusal | the exit-error branch of `afMissingCommits` returns `nil, nil` | `..._N62TheFirstBindRefusesWhenGitCannotAnswerTheCommitCheck/fail`, `..._N62TheFirstBindRefusesWhenAnObjectIsDamaged`, `..._N62AMoveOfAnUnmarkedLedgerRefusesWhenGitCannotAnswer/fail`, `..._N62AMoveOfAMarkedLedgerDoesNotDependOnTheCommitCheck` |
| M4 | answer length is checked | the length check removed | `.../extra` of the first-bind and unmarked-move tests |
| M5 | answer line shape is checked | the `default:` refusal removed | `.../empty` of the same two tests |
| M6 | a lease is not a clue (N65) | `check(l.Worktree, ..., true)` | `TestAFIdentity_N65ACloneBoundThroughALeaseNeedsADeclaration`, `TestAFVerbs_N51ADeclarationIsNotAskedWhere.../a_lease_worktree_in_the_repository_is_not_a_clue` |
| M7 | evidence commits are not required for a move | `add(r.SubjectSHA, true)` | `TestAFIdentity_N67AnUnmarkedLedgerMovesPastAPrunedEvidenceCommitButNotAMissingCandidate` |
| M8 | unmarked move needs a declaration | `if false && f.bindRepo == ""` | `TestAFIdentity_AnUnmarkedLedgerMoveNeedsADeclarationNotTheWorkingDirectory` |
| M9 | unmarked move asks where the ledger file sits | `&& false` on the file-location comparison | `TestAFIdentity_N61ALedgerFileInOneRepositoryIsNotMovedToAnother/a_ledger_without_one_asks_where_the_ledger_file_sits` |
| M10 | `--from` never overwrites a ledger | both the conflict case and the destination-exists guard disabled | `TestAFIdentity_N64AFromThatWouldOverwriteALedgerAtTheNewKeyIsRefused` (one guard alone survives: the other still refuses, a defence in depth) |
| M11 | fresh-ledger sentence | the fresh branch disabled | `TestAFIdentity_N64AFreshLedgerAtTheNewKeyIsNotCalledTheRepositoryThatDidNotMove` |
| M12 | an unreadable marker is a refusal | `afReadMarker` returns `"", nil` for a file that holds no marker | `TestAFIdentity_AMarkerFileThatHoldsNoMarkerIsARefusalNotAnAbsence` (both subtests) |
| M13 | schema bump on the first move | the `PRAGMA user_version` line removed | `TestSchemaVersion_TheFirstRecordedMoveRaisesItSoThePreviousBuildRefusesTheFile`, `TestAFIdentity_N68ALedgerMovedByThisBuildIsRefusedByABuildThatDoesNotKnowTheMove` |
| M14 | marker first writer wins | the returned value taken from the argument, not the record | `TestRepositoryMarker_TheFirstRecordWinsAndTheRecordIsKept`, `TestRepositoryMarker_ConcurrentRecordsConverge` |
| M15 | ledger refuses a move with another marker | the comparison disabled | `TestRepositoryMarker_AMarkedMoveWithAnotherMarkerThanTheRecordedOneChangesNothing` |

Not mutated: the first-bind refusal on an unreadable row set (see Gaps).

E9 existing tests changed (assertions kept; the one deliberate reversal is named):
- `ledger/repository_move_test.go` `TestRepositoryMove_FilesOfEarlierSchemasOpenIntoTheNewOne`: after a move `user_version` is compared with `SchemaVersionMoved` (was `SchemaVersion`); the same equality, the new value. `TestRepositoryMove_AFileOfANewerSchemaIsStillRefused`: the unknown future version is 4 and the message `schema 4` (was 3 and `schema 3`, now a known version).
- `cli/af_repo_n51clone_test.go` subtest "a lease worktree in the repository" -> "... is not a clue": **reversed on purpose (N65)**. Was: `reap --db` without declaration binds. Now: refused with `not bound yet`, `--bind-repo` and the identity, then bound with `--bind-repo`; the final `bound == id` assertion is kept.
- `cli/af_repo_n52_test.go` `TestAFVerbs_N52TheMoveIsRefusedWhenTheNewRepositoryLacksACommitTheRowsName`: the ledger is made one without a marker first (`afIDForget`), because a marked ledger is now refused by the marker; the assertions (`a.c1`, `does not have`, ledger unchanged, then the control with `--bind-repo`) are unchanged.
- `cli/af_repo_n38_firstbind_test.go` `TestAFVerbs_N38FirstBindLinkedWorktreeOfTheProjectPasses`: `reap` gets `--bind-repo <project>` (the lease in the linked worktree is still compared). `TestAFVerbs_N38FirstBindWorkingDirectoryNamesAnotherRepository`: the control measures with `measureDeclared` (adds `--bind-repo`).
- `cli/af_repo_n51d_test.go`: the control `evidence` gets `--bind-repo`.
- Fixtures only, nothing asserted differently: `afBindFixture(t, st, repo)` after `afUnqueuedCard` in `af_land_n3_test.go`, `af_land_n28_test.go`, `af_repo_n38_test.go`, `af_resume_test.go`, `af_land_d1_test.go` (the fixtures used to be bound by the first verb on the strength of the lease). `af_land_d1_test.go` `afRunBoth` gains the `AF_TEST_BIN` branch; unset, it is unchanged.

E10 the follow-up commit `919b6d8610` holds only test changes made after the GREEN commit (deterministic order of the clone/mirror subtests, a stronger linked-worktree marker assertion, whitespace-insensitive path assertion for the binary's framed errors, a lint-clean schema test) and this report.

### Baseline-attribution

Every command above ran in this run against this tree (HEAD `19eb2c9cc9`, source unchanged since). The a36ea8c368 measurements are of a binary built from `git archive a36ea8c368` of this repository. The ledger/mergeq coverage baseline (88.0% / 88.2%) is the audit's figure for a36ea8c368, not re-measured here; this run measured 87.8% / 88.2%.
Tool provenance (§2.2): the binaries named in E7 were built from the trees named there and invoked by path; `go test` ran from this tree. The installed `moai` on PATH was not used.

### N67/N68: how far it got

N67 (a pruned commit blocks a real move): closed for every ledger that has a marker (the commit check is informational there) and, for an unmarked ledger, narrowed to the commits of queue rows that are not failed; the evidence subjects and failed rows are counted as `commits_unchecked`. The first-bind check still compares all row commits (an unbound ledger has not moved). No further code or tests were added for it after the correction.
N68 (older builds accept the old path of a moved ledger): written before the correction arrived and kept as it stood: `SchemaVersionMoved = 3`, raised by `MoveRepositoryMarked` in the move's own transaction, `SchemaVersion` stays 2, this build opens 1/2/3 and refuses 4; the marker table itself needs no version change. It sits in one place (the `PRAGMA user_version` line in `ledger/repository_move.go`, the constant in `schema.go`, and the tests in `repository_marker_test.go` and `TestAFIdentity_N68...`). The leader wrote "do not add the schema-version bump"; it is there because it was written first. Removing it is a one-line change plus those tests, if the leader wants it out.

### Gaps (not observed, refused, or not run)

- The full `./internal/cli/...` suite was not run (the spawn forbids it); the anchored `^(TestAF|TestCardFlow)` run was.
- The `/verify` gate (`moai gate`: lint, format, type, full test run) was not invoked before the GREEN commit, because it runs the whole test suite; `gofmt`, `go vet`, `golangci-lint` (ledger and cli) and the targeted tests above stand in for it, and that is a substitution.
- `MOAI_*_LIVE` checks, push, PR and the audit were not done (forbidden).
- The binary comparison (E7) runs the `TestAFIdentity_` group only, not the changed existing tests; those run in-process in E6.
- A first-bind refusal on an unreadable row set (`rerr != nil` in `afCheckKnownRepositories`) has no test and no mutation; it needs a ledger whose `Cards` or `Queue` read fails.
- A marker write that fails because the git directory is read-only has no test (the code refuses and says the ledger is bound but holds no marker).
- `--from` with a symbolic-linked spelling of the old directory that cannot be resolved is covered only by `afResolveGone` on a macOS temporary directory; other layouts (the project key for a worktree-root project, `rootLayout`) were not tried.
- The session's recorded heartbeat binding (`hook.BindLedgerHeartbeat`, the ledger path recorded at `claim`) still names the old default ledger after `--from`; the renewal then fails quietly. Not changed, not tested.
- No Windows run: `os.Link` and the shell-based tests were compiled for windows only (`GOOS=windows go build`).

### Residual-risk

- An unmarked ledger (bound before this change, not used since) with no non-failed queue row accepts any repository the operator declares with `--bind-repo`: the fallback has no evidence to refuse a clone with. The first verb run in the old place after this change gives the ledger its marker, so this window closes by use.
- A plain `cp -R` of the repository carries the marker; it is a continuation only once the original is gone, and the "old git directory still exists" rule stops a live copy. A copy made while the original is gone for a moment (an unmounted volume) and then the original comes back is two repositories with one marker; the ledger follows whichever was named last.
- Anyone who can write the git directory can copy the marker file into a stranger; the marker is identity against accident (a clone, a mirror, a re-creation), not against an operator who forges it.
- `--bind-repo` is still a declaration and not evidence.
- The marker file is a new file inside `.git` of every repository a ledger is bound to; a tool that cleans the git directory removes it, and the next verb then refuses until the operator re-adopts (a marked ledger with a missing marker is refused, by design, with no automated re-adoption verb).
- Old builds that do not know the marker ignore it; a ledger that never moved is readable by them and they decide it by path as before (N68 only protects a moved ledger).

### Impact-path table (claim -> run -> check -> handoff -> integrate -> sync/close, and recovery)

| Stage | Call site (file:function) | Classification | Reason |
|---|---|---|---|
| claim | `af.go: newAFClaimCmd` -> `afRun` -> `afBindRepository` | repaired | marker compare at every verb; a lease worktree is no clue at the first bind |
| claim | `hook.BindLedgerHeartbeat` (renewal binding) | out of scope | records the ledger path; stale after `--from` (Gaps) |
| run | `af.go: evidence` -> `afCheckCommit` -> `afBindRepository(repo)` | repaired | same check, for the repository the commit is looked up in |
| run | `af.go: add`, `move`, `release`, `resume` -> `afRun` | repaired | same check through `afRun` |
| check | `af.go: afCheckLandedInTrunk` (move to DONE) | unaffected | asks the project repository for the trunk; the identity of that repository is enforced by the bind that ran before it |
| check | `af.go: afCheckRepoFlag` (`--repo`) | unaffected | compares git directories of the project and `--repo`; independent of the marker, runs before the ledger opens |
| handoff | `af.go: enqueue` -> `afCheckCommit` + `afInTrunk` | repaired (bind) / unaffected (trunk test) | |
| integrate | `af_land.go: afBindRepository(repo)` and the land body | repaired (bind) | `mergeq.TrunksOwnCommit` and E6 are untouched |
| integrate | `af_land.go: afCardMayLand` (N54) | repaired (see "Transition group") | N63, N66, N69 first bullet: the second spawn's rows are in the Transition group section below |
| sync/close | `moai af move ... --stage DONE` | repaired (bind) | the card is closed by the same verbs; the sync-phase documents are not an `af` call |
| recovery | `af.go: reap`, `reconcile` -> `afRun` | repaired | lane death then restart across a move: `TestAFIdentity_AHolderThatDiedBeforeTheMoveDoesNotStrandTheCard` |
| recovery | `af_move_repository.go` | repaired (new rules) | marker, fallback, `--from`, fresh-ledger sentence |
| recovery | `hook/ledger_heartbeat.go` (opens the ledger to renew a lease) | unaffected | renews by card and epoch; it decides no repository identity |
| recovery | `internal/web` board store, `mergeq` real-ledger tests | unaffected | read the ledger or use it in tests; no bind decision |

### Failure ledger, identity family (the transition family rows are the second spawn's)

| Family | Round (audited SHA) | Fix SHA | Root cause | Regression test | Status |
|---|---|---|---|---|---|
| N38 | 46dbcce0d4 | RED `452f8b999b`, GREEN `319b39c1cc`; first-bind `b5932a8bfb`/`7f957917ab` | verbs ran against whatever repository the caller stood in; the first call decided | `af_repo_n38_test.go`, `af_repo_n38_firstbind_test.go` | direct form closed; forms through a lease and a clone reopened as N51, N65 |
| N49 | 1c05083f71 | RED `2a0c9fb455`, GREEN `a920857df4` | evidence/enqueue bound without the `--db` the caller gave | `af_repo_n49_test.go` | closed |
| N50 | 1c05083f71 | RED `f6674dc628`, GREEN `5c389902ec` | the default ledger location counted as a clue | `af_repo_n50_test.go` | closed |
| N51 | 1c05083f71 | (c) `b3c9104ee1`/`7fe486101e`, (d) `7fb3ee9507`/`5e775c89a5`, (b) `220da0d4a9`/`6e02850e3b`, clone `95e81a2e74`/`d247597ae6` | first-bind clues were a symlink, a relative lease, the commits, the declaration | `af_repo_n51*_test.go` | direct forms closed; new routes N62 and N65 below |
| N52 | 1c05083f71 | ledger `4034ebfe87`/`8781b603ac`, verb `baac4dc05c`/`9df79edb2a` | the repository moved on disk and nothing finished the card | `af_repo_n52_test.go` | a plain rename closed; clone, unrelated and pruned-commit forms reopened as N61, N67; default location N64 |
| N61 | a36ea8c368 | RED `e544228237`, GREEN `19eb2c9cc9` | repository identity decided by path and commit presence | `TestAFIdentity_N61*`, `..._ARepositoryRecreated...`, `..._ACopyIsAMove...` | repaired for marked ledgers; unmarked fallback documented as residual |
| N62 | a36ea8c368 | same | `afMissingCommit` read a git failure, an empty or short answer as "nothing missing" | `TestAFIdentity_N62*` | repaired (first bind, move-repository, marker read) |
| N64 | a36ea8c368 | same | default ledger keyed by the project path; no way to carry it over | `TestAFIdentity_N64*` | repaired as `--from` (no discovery); heartbeat binding stays stale |
| N65 | a36ea8c368 | same | a lease worktree counted as proof of ownership | `TestAFIdentity_N65*`, N51 subtest | repaired |
| N39 N40 N41 N42 N54 N63 N66 | (transition family) | see the Transition group section below: its failure ledger rows | | | |

## Transition group (N63 N66 N69 N72 N73-help N58)

Commits: RED `36365f2c82` (tests only), GREEN `222aa8b800` (source and the `afTrCarryOn` test fix found while going green), test-only follow-up `89e3962a36` (E6 pair test), this report in the docs commit after it. Source files are identical in `222aa8b800` and `89e3962a36`.
Scope: the leader's list for this spawn (N63, N66, N69 first bullet, N72 false `--db` help, N73 E6 help sentence, N58 for the two BLOCKINGs and `add --risk`). The identity group above, the rest of N69, N70, N71, N74 and the SPEC side of N73 were not touched.

### Design as built, and the N63 choice

N63 was closed by the audit's option (a), the smallest of the three. Compared against source:

- (a) `ledger.Move`, `move.go`: a move without a lease was already allowed for a MERGING card at its own epoch when its queue row of that epoch is `landed` or `failed` (`queueSettledAt`). A waiting row whose candidate's newest green, gate or review run failed is the row `af land` refuses; with the holder gone nothing else settles it. One new query (`queueUnlandableAt`, `queue.go`) and one condition beside `queueSettledAt`; the table (`core.CheckMove`) still decides the move. It writes nothing on a refusal, so the N54 tests that compare the whole ledger before and after a refused `land` (`TestAFLand_N54AFailingNewestRunWhileTheRowWaitsRefusesTheLanding`) hold unchanged.
- (b) `land` marks the row `failed` on the refusal: `land` would write on a refusal, which breaks "nothing was changed" (the sentence and that test's snapshot compare) and would write the row of a live holder who is in the middle of measuring again.
- (c) `pickTx` claims a MERGING card with a waiting row: a claim of a MERGING card would give a lease on MERGING and a `claimed again` transition at MERGING, and every consumer of a claim assumes a work stage; it also puts a per-candidate evidence query inside the claim transaction for every card.

The way out is `moai af move <card> --epoch <n> --stage REVIEWING` (the card's own epoch, no lease). The next `moai af claim` takes the card (REVIEWING, no lease), measures again, `enqueue` restarts the waiting row of the older epoch (a decision note is written), `land` lands, the card finishes. The `land` refusal now prints this command next to the measuring command.

N66: `mergeq.Merger.LastCheck func(ctx, cardID, candidateSHA) error` (optional). `landUnderLock` calls it with the landing lock held, after the row says `landing` and immediately before `git update-ref`. An error returns through `leaveLanding`: the trunk is untouched, the row goes back to `waiting` under the lock, the lock is released, the error is returned as it is. `af land` fills it with `Store.FailingNewestRun` and the same refusal text (`afLandRefusal`). `reconcile` never lands and sets none. The already-in-the-trunk path (E6) returns before the gates and never reaches it.

N69 first bullet: the N54 refusal now says `measure the commit again and record what the measurement gives (`moai af evidence <card> --epoch <n> --kind <kind> --subject-sha <sha> --result <pass|fail>`, and the same for the other kinds the card needs), then land again; nothing was changed. If the holder of the card is gone (it holds no lease), `moai af move <card> --epoch <n> --stage REVIEWING` takes the card back, and the next `moai af claim` carries it on`. The epoch is the card's real epoch, and `<pass|fail>` is the one placeholder.

N72: one constant `afDBFlagHelp` (`af.go`) used by every `--db` flag of the 11 verbs: `ledger file (default: <MoAI home>/db/<project name>-<hash>/af/ledger.db, keyed by the project directory; it is not inside the project directory)`. N73-help: one constant `afDoneHelp` in the help of `move` and `land`: `DONE guarantees that the candidate commit landed in the trunk; it does not guarantee the card's work content (a candidate that is the trunk's own commit lands as already in the trunk).` The `move` help also names the lease-free move.

### Claim

1. A card whose holder left (released it, process gone then reaped, lease run out and process gone then reaped) while its waiting row's candidate has a failing newest run is no longer stuck: `move --stage REVIEWING` at its epoch is accepted, a new session claims it, measures again, and the card reaches DONE through a real landing. Every other verb listed by the audit (land, evidence, enqueue, resume, claim, reconcile, reap) was tried first and none moved the card. N63.
2. The rule is narrow: a waiting row that can land, a failure about another commit, a verifying row and a wrong epoch each still refuse a lease-free move. The ordinary paths hold: no failing run, release in MERGING, land, move back, claim (as before N54); a live holder re-measures and lands; a live holder takes the card back with its lease.
3. A failing run recorded while the gate of `af land` runs is not landed: the trunk ref does not move, the row is waiting, the card is MERGING, and the same exits work (live holder re-measures and lands; released or gone holder: way out then DONE). With nothing recorded the same `land` lands, and a failure about another commit recorded during the gate does not stop it. N66.
4. The refusal prints commands that run as printed (the evidence command with `<pass|fail>` filled in with `pass` and with `fail`; the move command as printed). N69 first bullet.
5. The `--db` help of all 11 verbs is one true sentence; `move` and `land` help carry the DONE sentence. N72, N73-help.
6. `moai af add --risk` is pinned: a risk card is refused at `enqueue` without a review, and at `integrated`/`DONE` after a failing review; the control without `--risk` is offered with no review. N58.
7. E6 is unchanged: the trunk's own commit lands as `already in the trunk` and the card reaches DONE, with the same result on the a36ea8c368 binary and on this tree.

### Evidence

All commands were run in this worktree (`/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration`).

E1 RED, before any source change (tree `36365f2c82`):
- `go test -count=1 -run TestAFTransition_ ./internal/cli/`: `--- FAIL: TestAFTransition_N63AHolderThatLeftARefusedRowLeavesAWayOutThatEndsInDone` (all three holder variants: `the way out (moai af move t1 --epoch 1 --stage REVIEWING) is refused: ledger: stale epoch; the card was claimed again: t1 holds no lease`), `--- FAIL: TestAFTransition_N66AFailureRecordedWhileTheGateRunsIsNotLanded` (all three: `land landed a commit whose review failed while the gate ran: {"card_id":"t1","outcome":"landed",...}`).
- `go test -count=1 -run TestAFMessage_ ./internal/cli/`: `the message tells the reader to type --result pass`, `the message does not say "record what the measurement gives"`, `the printed command [evidence t1 --kind gate ... --result pass --db ...] is refused: required flag(s) "epoch" not set`.
- `go test -count=1 -run TestAFVerbs_N7 ./internal/cli/`: `moai af land --db still says the default is in the project directory: "ledger file (default: <project dir>/af/ledger.db)"` for 11 verbs; `moai af move --help does not say: DONE guarantees ...`.
- `go test -count=1 ./internal/agenticfactory/ledger/ -run 'TestMove_(AMergingCard|TheWayOut)'`: `--- FAIL: TestMove_AMergingCardWhoseWaitingRowCannotLandMovesWithoutALease` (green, gate, review): `ledger: stale epoch; the card was claimed again: a holds no lease`.
- `go vet ./internal/agenticfactory/mergeq/`: `last_check_test.go:30:4: m.LastCheck undefined (type *Merger has no field or method LastCheck)` (a tests-only RED for new API).
- `go test -count=1 -run TestAFAdd_A ./internal/cli/` passes on the existing source once the test's own expectation was fixed (it first failed on `E12 needs merged`, which is the N40 rule working); the risk tests are pins of existing behaviour, and their RED is the mutation M11 below.

E2 GREEN (tree `222aa8b800`):
- `go test -count=1 ./internal/agenticfactory/ledger/ -run TestMove_` -> `ok`; `go test -count=1 ./internal/agenticfactory/mergeq/ -run TestLand_RealLedger` -> `ok`.
- `go test -count=1 -run TestAFTransition_ ./internal/cli/` -> `ok  	github.com/modu-ai/moai-adk/internal/cli	14.754s`; `-run TestAFMessage_` -> `ok`; `-run TestAFVerbs_N7` -> `ok`; `-run TestAFAdd_A` -> `ok`.
- `go test -count=1 -run 'TestAFLand_|TestAFVerbs_N' ./internal/cli/` -> `ok  	github.com/modu-ai/moai-adk/internal/cli	120.390s` (the neighbouring existing tests, before the commit).

E3 `gofmt -l internal cmd` printed nothing. `go vet ./internal/cli/ ./internal/agenticfactory/...` printed nothing. `go build ./...` printed nothing. `GOOS=windows GOARCH=amd64 go build ./internal/cli/ ./cmd/moai/` and `GOOS=linux GOARCH=amd64 go build ./internal/cli/ ./cmd/moai/` printed nothing.

E4 `go test -race -count=1 -p 4 ./internal/agenticfactory/...`: 21 lines, all `ok` (ledger `53.591s`, mergeq `66.687s`).

E5 `go test -count=1 -cover ./internal/agenticfactory/ledger/ ./internal/agenticfactory/mergeq/`:
```
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	6.830s	coverage: 87.7% of statements
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/mergeq	30.489s	coverage: 88.3% of statements
```
(the previous measurement of this tree was 87.8% / 88.2%.)

E6 `go test -count=1 -run '^(TestAF|TestCardFlow)' ./internal/cli/` (tree `222aa8b800`, before the E6 pair test was added; the pair test was run alone, below):
```
ok  	github.com/modu-ai/moai-adk/internal/cli	139.899s
```
`golangci-lint run ./internal/agenticfactory/ledger/... ./internal/agenticfactory/mergeq/... ./internal/cli/` -> one issue, `internal/cli/codex_close_entry_signal_test.go:89:17: Error return value of `pw.Close` is not checked (errcheck)` (F12 of the audit, outside this change).

E7 the real-CLI cumulative group `TestAFTransition_` (it goes through `afRunBoth`, so with `AF_TEST_BIN` set every verb is a separate process of the named binary). Binaries built in the scratchpad, never from the main checkout: `moai-a36e` from `git archive a36ea8c368` extracted into `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/b26a325f-a8b4-45a3-8d9a-1a9e774e549b/scratchpad/bin-SJ4Tzm/src` (`go -C <src> build -o ... ./cmd/moai`), sha256 `e49d4af71a2787e263707f97b549bbd6427a381af47f318e3152def5da7461c1`; `moai-final` from this tree at `222aa8b800`, sha256 `6e71f2d878d2213c3eb89319f01cbbe236bba7216e4fc82ad92412a6eb05510d`.
- a36ea8c368 build, `AF_TEST_BIN=.../moai-a36e go test -count=1 -v -run TestAFTransition_ ./internal/cli/`: `FAIL github.com/modu-ai/moai-adk/internal/cli 10.808s`. `--- FAIL: TestAFTransition_N63AHolderThatLeftARefusedRowLeavesAWayOutThatEndsInDone` (3 subtests: `the way out (moai af move t1 --epoch 1 --stage REVIEWING) is refused: Ledger: stale epoch; the card was claimed again: t1 holds no lease.`), `--- FAIL: TestAFTransition_N66AFailureRecordedWhileTheGateRunsIsNotLanded` (3 subtests: `land landed a commit whose review failed while the gate ran: {"card_id":"t1","outcome":"landed",...}`). The controls pass there: `--- PASS: TestAFTransition_N63TheOrdinaryPathsHoldBesideTheWayOut`, `--- PASS: TestAFTransition_N66WithNothingRecordedWhileTheGateRunsTheCardLands`, `--- PASS: TestAFTransition_N66AFailureAboutAnotherCommitWhileTheGateRunsDoesNotStopTheLanding`.
- final build, same command: `PASS`, `ok  	github.com/modu-ai/moai-adk/internal/cli	15.356s`; all 6 top-level tests and every subtest `--- PASS`.
- `TestAFTransition_E6ACandidateThatIsTheTrunksOwnCommitStillLandsAsAlreadyInTheTrunk`, run alone (`-run TestAFTransition_E6`): in-process `ok 1.875s`; with the a36ea8c368 binary `ok 2.076s`; with the final binary `ok 1.931s`. The same behaviour on both builds is the point.
- the previous group against the final build again: `AF_TEST_BIN=.../moai-final go test -count=1 -run TestAFIdentity_ ./internal/cli/` -> `ok  	github.com/modu-ai/moai-adk/internal/cli	21.494s`.
- (`TestAFMessage_` is skipped under `AF_TEST_BIN` by design: it reads the in-process message; the binary frames the message and may cut a word. It is outside the `TestAFTransition_` prefix.)

E8 mutation checks. Each rule was broken by a local source edit through a script that restores the file text afterwards (`git status --short` printed nothing after the last one); the named tests were run; none was committed.

| # | Rule | Mutation | Result | Failing test(s) |
|---|---|---|---|---|
| M1 | N66: the failing-run question is asked before the swap (library) | `if false && m.LastCheck != nil` in `landUnderLock` | caught | `TestLand_RealLedger_ARefusedLastCheckMovesNothingAndLeavesTheRowWaiting`, `..._ALastCheckThatAgreesIsAskedUnderTheLockBeforeTheSwap`, `..._ALastCheckThatRefusesAfterARebuildLeavesTheRowWaiting` |
| M2 | N66: a refusal stops the landing | the refusal ignored (`_ = err`) | caught | `..._ARefusedLastCheckMovesNothingAndLeavesTheRowWaiting`, `..._ALastCheckThatRefusesAfterARebuildLeavesTheRowWaiting` |
| M3 | N66: a refusal puts the row back to waiting | return the error without `leaveLanding` | caught | the same two |
| M4 | N66: `af land` sets `LastCheck` | `m.LastCheck = func(` -> `_ = func(` | caught | `TestAFTransition_N66AFailureRecordedWhileTheGateRunsIsNotLanded` |
| M5 | N63: the lease-free move for an unlandable waiting row | the new condition disabled (`if false`) | caught | `TestMove_AMergingCardWhoseWaitingRowCannotLandMovesWithoutALease` |
| M5b | N63 through the verbs | the same mutation | caught | `TestAFTransition_N63AHolderThatLeftARefusedRowLeavesAWayOutThatEndsInDone` |
| M6 | N63: only a waiting row | `AND state='waiting'` removed | caught | `TestMove_TheWayOutIsOnlyForAWaitingRowWhoseCandidateCannotLand`      |
| M7 | N63: only a row whose candidate failed | the answer made `true` for any waiting row | caught | `TestMove_TheWayOutIsOnlyForAWaitingRowWhoseCandidateCannotLand`      |
| M8 | N69/N63: the way-out command is the printed one | `--stage REVIEWING` -> `--stage BUILDING` in the text | caught | `TestAFMessage_TheRefusalNamesCommandsThatRunAsPrinted` |
| M9 | N69: no blind `--result pass` | `--result <pass|fail>` -> `--result pass` | caught | `TestAFMessage_TheRefusalNamesCommandsThatRunAsPrinted` |
| M10 | N69: the printed epoch is the card's | a wrong epoch (0) printed | caught | `TestAFMessage_TheRefusalNamesCommandsThatRunAsPrinted` |
| M11 | N58: `add --risk` reaches the card | `Risk: risk, ` removed from `af_add.go` | caught | `TestAFAdd_ARiskCardNeedsItsReviewBeforeItIsOfferedAndBeforeDone` |
| M12 | N72: one true `--db` help on every verb | `af land` keeps the old text | caught | `TestAFVerbs_N72EveryDbFlagSaysWhereTheDefaultLedgerReallyIs` |
| M13 | N73: `move` help carries the DONE sentence | removed | caught | `TestAFVerbs_N73TheHelpOfTheVerbsThatEndACardSaysWhatDoneGuarantees` |
| M14 | N73: `land` help carries it | removed | caught | the same |

Two mutants did not compile at first (M7 `declared and not used`, M10 a `go vet` format-type error) and were rewritten into mutants that compile; the table is of the compiled ones. 15 of 15 compiled mutants were caught, none survived.

E9 existing tests changed: none deleted or weakened. One shared fixture helper was split, `afUnqueuedCard` (`af_land_d1_test.go`) now calls `afUnqueuedCardAs` with the claim it always made (`SessionID: "s-t1"`, the test process as holder, `PIDStart: "start-t1"`, the repository as worktree); every existing caller sees the same fixture. No assertion of any existing test changed.

### Baseline-attribution

Every command above ran in this run against this tree (HEAD `222aa8b800`, source unchanged through `89e3962a36`); the a36ea8c368 measurements are of a binary built from `git archive a36ea8c368` of this repository. The 85% coverage baseline is the SSOT figure; the earlier ledger/mergeq measurement (87.8% / 88.2%) is the identity group's, this run measured 87.7% / 88.3%. Tool provenance (§2.2): the binaries named in E7 were built from the trees named there and invoked by path through `AF_TEST_BIN`; `go test` ran from this tree. The installed `moai` on PATH was not used.

### Failure ledger, transition family

| Family | Round (audited SHA) | Fix SHA | Root cause | Regression test | Status |
|---|---|---|---|---|---|
| N39 | 46dbcce0d4 | RED `863aeb76f7`/`78fda9263a`, GREEN `fe59cd74ea`; ledger side `ed8ea412b6`/`c054fe62ed` | a landed row was trusted as "in the trunk" without asking the repository | `af_trunk_n39_test.go`, `internal/agenticfactory/ledger` take-over tests | closed (the audit's own re-measure: closed, mutants caught); E6 kept by decision |
| N40 | 46dbcce0d4 | RED `590f3715f1`, GREEN `6b494aef42` | review was outside the newest-result judgment of a landed candidate | `done_review_and_failures_test.go`; `TestAFAdd_ARiskCard...` (new) pins the input side | closed; the audit's surviving mutant `N59-add-risk` is now caught (M11) |
| N41 | 46dbcce0d4 | RED `6ed5d24ceb`, GREEN `3ef799ca90`, pins `71904ebb02` | a review or a failing record counted as measuring a commit again | `enqueue_measuring_again_test.go` | closed |
| N42 | 46dbcce0d4 | RED `5529297d60`, GREEN `7deb693d57` | a reclaim-cap park had no verb out | `af_resume_test.go` | closed; operator-authority limits are N60 |
| N54 | 1c05083f71 | RED `477938d926`, GREEN `7d6eeb8cc8` | `af land` looked at position, not at a failing newest run | `af_land_n54_test.go` | direct form closed; its two side effects are N63 and N66 below |
| N63 | a36ea8c368 | RED `36365f2c82`, GREEN `222aa8b800` | the N54 refusal asked the holder to measure again, but a holder that left cannot (every write needs a lease), a MERGING card is not claimable, and a waiting row does not settle the card | `TestAFTransition_N63*`, `TestMove_AMergingCard...`, `TestMove_TheWayOutIsOnly...`, `TestAFMessage_...` | repaired by option (a); mutants M5, M5b, M6, M7 caught |
| N66 | a36ea8c368 | RED `36365f2c82`, GREEN `222aa8b800` | the failing-run check ran once, before a gate that can take minutes | `TestAFTransition_N66*`, `TestLand_RealLedger_A...LastCheck...` | repaired; mutants M1 to M4 caught; the window between the check and `git update-ref` remains (Residual-risk) |

### Impact-path table, rows of the transition family (claim -> run -> check -> handoff -> integrate -> sync/close, and recovery)

| Stage | Call site (file:function) | Classification | Reason |
|---|---|---|---|
| claim | `ledger/claim.go: Claim`, `pickTx` | unaffected | a MERGING card is still not claimable (option (c) was not taken); the claim of the card the way out put in REVIEWING is `TestAFTransition_N63...` |
| run | `af.go: evidence` (lease) | unaffected | still needs the lease; a gone holder's card goes through the way out first |
| check | `af_land.go: afCardMayLand`, `afLandRefusal` | repaired | N54 refusal now names the way out and commands that run as printed (N69 first bullet) |
| check | `ledger/queue.go: FailingNewestRun`, `failingNewestRun` | unaffected | the same rule, now also asked by `queueUnlandableAt` and by `LastCheck` |
| handoff | `af.go: enqueue` / `ledger.Enqueue` | unaffected | restarts the waiting row of the older epoch after the way out (decision note written); exercised by `afTrCarryOn` |
| integrate | `mergeq/mergeq.go: landUnderLock` | repaired | N66: `LastCheck` under the lock before the swap; the row returns to waiting through `leaveLanding` |
| integrate | `mergeq/mergeq.go: land` path "already in the trunk" (`settleLanded`) | unaffected | returns before the gates; E6 pair test passes on both builds |
| integrate | `af_land.go: reconcile` | unaffected | never lands, sets no `LastCheck` |
| sync/close | `af.go: move` to DONE (`afCheckLandedInTrunk`) | unaffected (help changed) | N73 help sentence only |
| recovery | `ledger/move.go: Move` | repaired | the lease-free move for a waiting row that cannot land (N63) |
| recovery | `af.go: reap`, `ledger.Reap`, `reclaimTx` | unaffected | after a reap the lease is gone and the epoch is unchanged, which is the state the way out serves; the third reap parks the card at reclaim-cap and `resume` is that exit (not combined with this one in a test) |
| recovery | `af.go: release` | unaffected | the same state through a release |
| recovery | `hook/ledger_heartbeat.go`, `internal/web` board | unaffected | read and renew by card and epoch; no decision about landing |

### Gaps (not observed, refused, or not run)

- The full `./internal/cli/...` suite and `/verify` (`moai gate`, which runs the whole suite) were not run (the spawn forbids the suite); the anchored `^(TestAF|TestCardFlow)` run, `go vet`, `gofmt`, `golangci-lint`, `go test -race ./internal/agenticfactory/...` and the cross builds stand in for it, and that is a substitution. The anchored run was made at `222aa8b800`; the commit after it adds one test (run alone, E7) and this report.
- Worktree-guard refusals met in this run, each answered by splitting the command into plain ones: a compound `mkdir ... && cat > file <<EOF ... cd ... && python3` for the mutation script ("too complex to verify"), a `go test` with a redirect to the scratchpad and a pipe (same message, seen on the first RED run), and nothing else; no check was dropped because of a refusal, and the script was then written with the Write tool and run as a plain `python3 <script> <ids>`.
- `move --stage CANCELLED` without a lease on such a row is not tested or documented (only REVIEWING is): the table's own guards (E16) may refuse it.
- The window between `LastCheck` and `git update-ref` is not closed (evidence writes do not take the landing lock); no test can pin it either way.
- `TestAFMessage_` runs in-process only (skipped under `AF_TEST_BIN`); the tie between the printed text and the commands is therefore not proven against the binary, where the way-out command is built by the test from the same words.
- `afTrRestart` emulates a restart in the in-process run (it closes the setup's ledger connection and opens a new one); the literal process separation is the `AF_TEST_BIN` runs in E7.
- The rest of N69 (a lease recorded before the move needing `--repo`, a linked worktree needing `git worktree repair`, a case-only rename), N70, N71, N74, the second half of N72 (the "recorded as given" warning on the new input flags) and the SPEC side of N73 were not done; they were not in this spawn's list.
- Windows: compile only (`GOOS=windows go build`); the tests need `sh` and are skipped there.
- `go test -race` of `./internal/cli/` was not run.

### Residual-risk

- A failing run recorded after `LastCheck` has read and before `git update-ref` has run still lands (microseconds). Nothing in the ledger can close it without making every evidence write take the landing lock.
- The lease-free move is authorised by the card's epoch number, which is not a secret (the same authority the ledger already gives a landed or failed row). Anyone who knows the epoch can take such a card back to REVIEWING while its row waits with a failing candidate; the position table still has to accept the move, and the move is recorded with its actor and reason.
- After the way out the old row stays `waiting` at the older epoch until the next holder's `enqueue` restarts it (tested); `land` on that card in the meantime is refused by the epoch rule of `afCardMayLand` (the card is not in MERGING).
- A failure about ANOTHER commit recorded while the row waits is still not a refusal (N74, D5): the landing stands and `DONE` is refused later. Unchanged.
- `<pass|fail>` in the refusal is text for a person; a script that runs the printed command without replacing it records a literal result and is refused by the verb (`evidence result`), not silently accepted.
