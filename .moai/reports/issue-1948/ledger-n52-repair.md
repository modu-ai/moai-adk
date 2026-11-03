# Ledger repair for audit finding N52 (repository moved on disk), issue-1948 / AF-18

Spec: SPEC-AGENTIC-FACTORY-LOCAL-001. Parent tree: `1c05083f71f185d0a272013fe8f0d420f3a14714`. Ledger side only; the CLI verb that calls the new API is lane-2's.

## 1. Claim

1. The ledger can record that the repository moved on disk. The record is append-only (`repository_move`, with UPDATE and DELETE refusal triggers); the first-binding row `repository` and its triggers are unchanged.
2. The current identity is the newest move's new identity, or the first binding when there is no move. `Repository()` returns it, `BindRepository(current)` is a no-op, `BindRepository(older identity)` is refused with `ErrWrongRepository`.
3. A move must start from the current identity (stale or racing callers are refused with `ErrWrongRepository`, so the history cannot fork); a move on an unbound ledger is refused; a move to the same identity and any blank value are refused with `ErrInvalid`. Chosen and tested: same-identity is a refusal, not a no-op. A move back to an earlier identity is an ordinary move.
4. No git is called (the identities are the caller's strings). `SchemaVersion` stays 2: the table is created with `CREATE TABLE IF NOT EXISTS` in the schema DDL that runs on every Open, so v1 files and v2 files lacking the table open into the new schema, as N38 did for `repository`. A newer schema than the binary knows is still refused.
5. Package coverage is 87.9% (gate 85.0%); vet, gofmt and golangci-lint are clean on the ledger package.

Exact signatures (package `internal/agenticfactory/ledger`):

```go
type RepositoryMove struct {
	Seq    int64
	From   string
	To     string
	Reason string
	At     time.Time // the database clock (nowSQL), parsed with the package's timeLayout
}
func (s *Store) MoveRepository(ctx context.Context, from, to, reason string) error
func (s *Store) RepositoryMoves(ctx context.Context) ([]RepositoryMove, error) // oldest first; nil when none
```

Errors: blank `from`/`to`/`reason` and `from == to` wrap `ErrInvalid`; a `from` that is not the current identity, or any move on an unbound ledger, wraps `ErrWrongRepository` (message names the current and the offered identity, prefix `ledger: `); engine failures come back as `ledger: ...`. `Repository()` and `BindRepository()` signatures are unchanged.

Commits (parent 1c05083f71):

| Step | SHA |
|---|---|
| RED | `4034ebfe87eceadd6a6206fd8ae6bcdfae13b12a` |
| GREEN | `8781b603ac04323ba0935f1354056d93c523fdc4` |
| Report + progress line | the commit after GREEN that adds this file (its SHA is in the hand-back summary, a file cannot name its own commit) |

Files: `repository_move.go` (new: type, two methods, `currentRepository`), `repository.go` (Bind and Repository read `currentRepository`), `schema.go` (+10 lines DDL), `repository_move_test.go` (new, 17 tests). `repository_identity_test.go` is untouched (not in `git diff 1c05083f71 HEAD`).

## 2. Evidence

RED (stub `MoveRepository` returning an error and `RepositoryMoves` returning nil, i.e. the parent's behavior, table absent):

```
$ go test -count=1 -run 'TestRepositoryMove_' ./internal/agenticfactory/ledger/ | grep -E '^(--- |FAIL|ok|PASS)'
--- FAIL: TestRepositoryMove_RecordsTheMoveAndTheCurrentIdentityFollowsIt
--- FAIL: TestRepositoryMove_TheMoveSurvivesAReopen
--- FAIL: TestRepositoryMove_SeveralMovesChainInOrderAndTheNewestIsCurrent
--- FAIL: TestRepositoryMove_MovingBackToAnEarlierIdentityIsAMoveLikeAnyOther
--- FAIL: TestRepositoryMove_BindRepositoryFollowsTheCurrentIdentity
--- FAIL: TestRepositoryMove_AMoveMustStartFromTheCurrentIdentity
--- FAIL: TestRepositoryMove_AStaleMoveIsRefusedAndTheHistoryDoesNotFork
--- FAIL: TestRepositoryMove_AnUnboundLedgerHasNothingToMoveFrom
--- FAIL: TestRepositoryMove_MovingToTheSameIdentityIsRefused
--- FAIL: TestRepositoryMove_BlankValuesAreInvalid
--- FAIL: TestRepositoryMove_TwoMovesFromTheSameIdentityLeaveExactlyOneWinner
--- FAIL: TestRepositoryMove_TheRecordCannotBeUpdatedOrDeletedByAnyoneHoldingTheFile
--- FAIL: TestRepositoryMove_TheTableRefusesBlankOrUnchangedRows
--- FAIL: TestRepositoryMove_FilesOfEarlierSchemasOpenIntoTheNewOne
--- FAIL: TestRepositoryMove_FailurePathsAreLedgerErrors
FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	0.386s
```

15 of the 17 new tests fail on RED. The other two pass on RED by design because they pin behavior that must not change: `AnUnmovedLedgerHasNoMoves` and `AFileOfANewerSchemaIsStillRefused`. A whole-package RED run showed no failing test outside `TestRepositoryMove_`.

GREEN, first attempt (no retry was needed; attempt cap 3 not used):

```
$ go test -race -count=1 -cover ./internal/agenticfactory/ledger/
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ledger	33.553s	coverage: 87.9% of statements
```

```
$ go vet ./internal/agenticfactory/ledger/...      (no output)
$ gofmt -l internal/agenticfactory/ledger/          (no output)
$ golangci-lint run ./internal/agenticfactory/ledger/...
0 issues.
```

The 17 new tests, all `--- PASS` (`go test -count=1 -v -run 'TestRepositoryMove_'`, `ok ... 0.680s`): RecordsTheMove..., AnUnmovedLedgerHasNoMoves, TheMoveSurvivesAReopen, SeveralMovesChain..., MovingBackToAnEarlierIdentity..., BindRepositoryFollowsTheCurrentIdentity, AMoveMustStartFromTheCurrentIdentity, AStaleMoveIsRefused..., AnUnboundLedgerHasNothingToMoveFrom, MovingToTheSameIdentityIsRefused, BlankValuesAreInvalid, TwoMovesFromTheSameIdentityLeaveExactlyOneWinner, TheRecordCannotBeUpdatedOrDeleted..., TheTableRefusesBlankOrUnchangedRows, FilesOfEarlierSchemasOpenIntoTheNewOne (v1; v2 without the table; a recorded move survives the migration of the other tables), AFileOfANewerSchemaIsStillRefused (user_version=3, error names "schema 3"), FailurePathsAreLedgerErrors (move table gone; binding table gone; column renamed; engine refuses the write via a BEFORE INSERT trigger; unreadable time; closed ledger).

Mutation checks, each applied as a `go test -overlay` copy (copies and overlay JSON live in the session scratchpad; the tree was never edited). Command shape: `go test -count=1 -overlay <mN>.json -run 'TestRepositor' ./internal/agenticfactory/ledger/`; counts are top-level `--- FAIL` lines.

| # | Mutation | Result | Caught by |
|---|---|---|---|
| M1 | drop the from-must-equal-current check (`case current != from` -> `case false`) | FAIL, 3 | AMoveMustStartFromTheCurrentIdentity, AStaleMoveIsRefused..., TwoMovesFromTheSameIdentity... |
| M2 | current identity ignores moves (`SELECT identity FROM repository WHERE id=1`) | FAIL, 13 | 2 `TestRepository_` and 11 `TestRepositoryMove_` tests |
| M3 | drop the two append-only triggers from the DDL | FAIL, 1 | TheRecordCannotBeUpdatedOrDeletedByAnyoneHoldingTheFile |
| M4 | BindRepository also accepts the first-bound identity after a move | FAIL, 1 | BindRepositoryFollowsTheCurrentIdentity |
| M5 | drop the `from == to` check | FAIL, 1 | MovingToTheSameIdentityIsRefused (the table CHECK still stops the row, but not as `ErrInvalid`) |
| M6 | `RepositoryMoves` newest first | FAIL, 1 | SeveralMovesChainInOrderAndTheNewestIsCurrent |
| M7 | drop the unbound-ledger case | FAIL, 1 | AnUnboundLedgerHasNothingToMoveFrom |

No survivor, so no extra test was added for one.

## 3. Baseline-attribution

- Tree: worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/af-core-claim-runner`, branch `WT-agentic-factory-core-claim-runner`. Parent `1c05083f71f185d0a272013fe8f0d420f3a14714`; RED `4034ebfe87`; GREEN `8781b603ac04323ba0935f1354056d93c523fdc4`.
- Every command in section 2 was run in this session against this tree; the final test/vet/gofmt/lint runs were made on the GREEN tree (`git status --short` then showed only the pre-existing uncommitted `18-progress.md`).
- Coverage figure: `go test -race -count=1 -cover ./internal/agenticfactory/ledger/` printed `coverage: 87.9% of statements`. The parent's coverage was not re-measured in this run; no before/after delta is claimed.
- Tool provenance: `go`, `gofmt`, `golangci-lint` were invoked through the shell path (installed builds); the Go toolchain compiled this tree, the linter's own build commit was not recorded (see Gaps).
- N38 approach followed: commit `023f20c7db` put the `repository` table into the idempotent schema DDL without a version bump; `repository_move` does the same.

## 4. Gaps

- `internal/cli/`, `mergeq/`, `core/` and the importer were not touched and not run: nothing outside `internal/agenticfactory/ledger/` was tested. The verb side (calling `MoveRepository`, computing identities, reporting the move) is lane-2's and does not exist yet; this change does not by itself unblock a card on a moved repository.
- Only the ledger package was linted/vetted/tested; the repository-wide test verdict belongs to CI on the integration branch and is PENDING.
- A looped mutation driver (`for ... go test -overlay`) was refused by the worktree guard ("a value computed at runtime ... too complex to verify"). I did not substitute anything unmeasured: I wrote the seven overlay JSON files with the Write tool and ran seven plain `go test` commands with literal paths. Mutation copies were produced with `sed` and `grep -v` (single compound commands, accepted).
- The mutation runs used `-run 'TestRepositor'` (the repository tests only), not the whole package and not `-race`; the whole package ran only in the final GREEN run. A mutation that harmed some other part of the package would not show in those counts.
- `go vet` was run as `go vet ... | tail`, so its exit status was masked by the pipe; the empty output is what was observed.
- `golangci-lint` build commit not recorded; no `moai` binary was involved.
- No independent audit, no GLM, no codex, no push (all forbidden by the brief). The two checks that cannot fail on RED by construction are named above.
- Concurrency: `TwoMovesFromTheSameIdentity...` ran under `-race` once in the final run and passed; the guarantee rests on the package's immediate write transactions (`_txlock=immediate`), not on a separate stress run.

## 5. Residual-risk

- A wrong first bind stays wrong, and a wrong `to` passed by a caller is recorded as given: the record is only as good as the identity string the caller computed. `MoveRepository` cannot check that `to` is really the repository; it runs no git (REQ-AFL-008). Two spellings of one path are two identities.
- A move needs the caller to pass the correct current identity. A caller that reads it with `Repository()` just before moving still races another caller; the loser is refused (tested), which is the intended outcome, but the CLI must surface that refusal.
- Anyone with file access can drop the triggers (`DROP TRIGGER`) or the tables, or write rows by hand; the triggers refuse rewrite and delete through the engine only (same limit as the existing history tables, I7). The table CHECKs also stop a blank or unchanged row inserted directly.
- Moves are not tied to any card or session, and `reason` is free text; there is no authorization step in the ledger, so who may record a move is the verb's decision (lane-2).
- A move back to an earlier path (A -> B -> A) is allowed, so the history can contain a cycle; the current identity is still the newest row.
- Downgrade: an older binary opening a file with moves does not know the `repository_move` table and reads the first binding as current, so it would refuse the moved path again. The schema version is unchanged by design, so the older binary does not refuse the file outright.
