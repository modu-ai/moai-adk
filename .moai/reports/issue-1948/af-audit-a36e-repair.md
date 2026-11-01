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
| integrate | `af_land.go: afCardMayLand` (N54) | out of scope | the transition group (N63, N66) belongs to the second spawn |
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
| N39 N40 N41 N42 N54 N63 N66 | (transition family) | (second spawn) | (second spawn) | (second spawn) | (second spawn) |
