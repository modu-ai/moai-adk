# SPEC-FACTORY-RETIRE-GUARD-001 — Implementation Plan

## A. Context

Card t1852 (Tier M, Class C; part C only of the card line t1825 -> t1848). Tree `d2d9ac08fb1899440da911121ddf13eecc806a3f`, branch `WT-factory-retire-guard`. Requirements, root cause and the closed audit history are in `spec.md`. Parts A, B and D are other cards (`spec.md` §G).

**Tier M, and why not Tier S.** The change itself is under 300 lines of production code, but the files it touches exceed Tier S's "under 5 files": 5 production files (two of them new), 2 new test files, 8 sync-phase documents (§I). Tier S also carries a plan-audit ceiling of 1 (`harness.plan_audit_tier_ceilings`), which leaves no repair round; Tier M has 2.

Files at this tree (names are the contract; line numbers drift with edits):

| File | Role in this SPEC |
|---|---|
| `internal/homestate/factory_run_retire.go` | `RetireRunIfDead` (:381) gains the already-retired early return; `retireRun` (:407) takes the classified `RunOwner`, reads the cards inside its transaction, and flips with a stamp compare-and-set plus a `RowsAffected` check. `retireRunUnlessLive` (:431), `ReconcileActiveRuns` (:339), `retirable` (:165) are **not changed** |
| `internal/homestate/factory_run_cards.go` (new) | `InProgressCard`, `rowsQueryer`, `inProgressCards`, `(*FactoryDB).InProgressCards`, `ErrRunHasInProgressCards`, `RunInProgressCardsError`, the test hook `retireAfterCardRead` |
| `internal/homestate/card_record.go`, `factory_run_adopt.go` | `terminalStates` and `sqlTerminalStates` — read only; the non-terminal definition is reused, not copied |
| `internal/cli/factory_handoff_recover.go` | `newFactoryRunsCommand` `--retire`: the refusal block, the stranded-cards block, the flag help text |
| `internal/cli/factory_retire_guard_text.go` (new) | the Korean catalogue (§E) and the two block renderers |
| `internal/cli/factory_relaunch.go` | `relaunchRetireFromRun`: the `left active (N in-progress cards)` outcome |
| Tests (new) | `internal/homestate/factory_run_retire_guard_test.go`, `internal/cli/factory_runs_retire_guard_test.go` |
| `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md`, `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` | sync-phase obligation (§I) |

**Callers of `RetireRunIfDead`** (non-test, `grep -rn "RetireRunIfDead" internal --include='*.go'`): `factory_handoff_recover.go:91` and `factory_relaunch.go:98`. **Callers of `retireRun`:** `RetireRunIfDead` only (`:393`). The run phase re-runs both greps; a third `RetireRunIfDead` caller needs review.

Existing tests that must stay green (all read; four homestate and two cli ones were run, ledger E3 and E4): `internal/homestate/factory_run_retire_test.go` (`TestRetireRunIfDeadRefusesLiveAndIndeterminate`, `TestEveryRetirementPathDeclinesUnenumeratedClassification`, `TestRetireRunPreservesRowAndAppendsEvent`, the reconcile tests), `internal/homestate/factory_run_boot_proof_test.go` (`TestRetireRunIfDeadAcceptsBootProof`, which seeds `(0,'')` stamps with no cards and still retires), `internal/cli/factory_runs_command_test.go` (`TestFactoryRunsCommandReportsAndRefuses`), `internal/cli/factory_relaunch_test.go` (`TestFactoryRelaunchFromRunRetiresDeadOwnerOnly`: its `runRetired` case never reaches `RetireRunIfDead`, §C.8 of the spec). Read in this plan: no existing test retires a run that carries a non-terminal card through `RetireRunIfDead`, so REQ-RRG-001 changes none of their assertions.

## B. Decisions, ordered by likelihood of change (reversibility order)

1. **Refusal prose language and exits** (§E) — operator-facing, cheapest to change; a catalogue edit plus the constants the tests assert on.
2. **Q3 — the stranded-cards block for a retired run** (default applied). Reversible by dropping one CLI call.
3. **Q1 — how far the compare-and-set reaches** (default applied: operator path only). Reversible by a follow-up that extends the same check to `retireRunUnlessLive`.
4. **Q2 — the in-progress definition** (default applied: non-terminal, one shared set).
5. Mechanical steps last: the `rowsQueryer` interface, the hook, the relaunch string.

## C. Design

### C.1 Homestate (`internal/homestate`)

- `type InProgressCard struct { CardID, State, OwnerLabel string }`.
- `type rowsQueryer interface { QueryContext(ctx, query string, args ...any) (*sql.Rows, error) }` — new, because the existing `queryRower` (`card_record.go`) exposes only `QueryRowContext`; both `*sql.DB` and `*sql.Tx` satisfy it.
- `inProgressCards(ctx, q rowsQueryer, runID) ([]InProgressCard, error)` — `SELECT card_id,state,owner_label FROM cards WHERE run_id=? AND state NOT IN ` + `sqlTerminalStates` + ` ORDER BY card_id`. The hook `retireAfterCardRead` (a package variable, nil in production) is called as the **last statement of this helper**, after its rows are drained and closed and immediately before it returns, so a refactor that moves only the read above `BeginTx` moves the hook with it and the hook then runs without the write lock. The single-connection pool deadlocks if a transaction is open and the pool handle is used (the hazard `runHasLiveRecords` documents), so the retirement passes its `tx`.
- `(*FactoryDB).InProgressCards(ctx, runID)` — the pool-handle wrapper for read-only callers (the CLI's stranded-cards block). The hook is nil outside the one test that sets it.
- `var ErrRunHasInProgressCards = errors.New("factory run holds in-progress cards")`; `type RunInProgressCardsError struct { RunID string; Cards []InProgressCard }` with `Error()` (run id, count, every card as `id (state, owner)`; English, like the sibling errors) and `Is(target) bool` so `errors.Is(err, ErrRunHasInProgressCards)` and `errors.As` both work. It does **not** wrap `ErrRunOwnerNotDead`: the two refusals mean different things and the relaunch verb distinguishes them.

### C.2 The operator retirement (`factory_run_retire.go`)

`RetireRunIfDead` keeps its order: classify; refuse a non-`retirable` classification with `ErrRunOwnerNotDead` (this precedes any card read, so a live or indeterminate run holding cards still gets the owner-not-dead refusal); **new:** a row whose `Status` is already `retired` returns `(classification, nil)` with no write, no event and **no card read** (REQ-RRG-003) — the early return deliberately precedes the card guard, which therefore applies to `active` runs only; otherwise call `retireRun` with the classified `RunOwner`.

`retireRun(ctx, owner RunOwner, c, basis)`, inside its existing `BeginTx` transaction, in this order:

0. `SELECT status, lead_pid, lead_process_start FROM runs WHERE run_id=?` through the transaction. If the status is no longer `active`, or the stamp differs from the stamp `classifyRuns` scanned (for a peer- or boot-basis run that is `(0,'')`), return `fmt.Errorf("%w: run %s changed owner or status since classification", ErrRunOwnerNotDead, runID)` and commit nothing. This is the owner-changed error, and it is reached **before** any card read.
1. `cards, err := inProgressCards(ctx, tx, owner.RunID)` (the hook fires inside it); a non-empty result returns `*RunInProgressCardsError` and commits nothing (the deferred rollback ends the transaction). Only a run whose classified owner stamp is still on the row reaches this step.
2. `UPDATE runs SET status='retired', updated_at=? WHERE run_id=? AND status='active' AND lead_pid=? AND lead_process_start=?` with the same classified stamp, then `RowsAffected`. Step 0 and this `WHERE` clause are the same comparison; inside the write-locked transaction nothing can change between them, so the clause is the second, defensive statement of it and the `RowsAffected` check is what keeps the event honest.
3. `!= 1` -> the same owner-changed error as step 0, commit nothing.
4. Exactly one row -> append the `run.retired` event (payload unchanged: `{"classification":...,"basis":...}`), commit.

The order inside the transaction is therefore **status -> stamp -> cards -> flip -> event**.

**Why the flip is a real compare-and-set and the card read is inside the lock** (spec §C.4, §C.5): `OpenFactory` -> `OpenFactoryPath` -> `openFactoryPathBusy` builds the DSN with `_txlock=immediate` (`factory.go:303`) and caps the pool at one connection (`:313`), so `BeginTx` issues `BEGIN IMMEDIATE` and holds the write lock to commit or rollback; a second handle's write waits out its busy timeout and fails. The race this SPEC closes is the one between `ClassifyRuns` (outside the transaction) and `BeginTx`: an `AdoptRun` — itself a compare-and-set on the dead stamp — that commits in that window changes the stamp, and step 0 then returns the owner-changed error.

**Ordering note (named so no reader assumes otherwise).** The stamp comparison precedes the card read. A run that was restamped — for example adopted by a live leader — gets the owner-changed error whether or not it carries cards: nothing is written, no cards error is produced, and the CLI never prints the abandon exit for it. So the cards error means "the classified owner is still on the row". Two limits: for a peer- or boot-basis run the compared stamp is the row's `(0,'')`, so a leader that registers as a peer without restamping the row is not seen by the comparison (spec §G); and the comparison is as fresh as the transaction, not as the operator's last look, so the abandon exit a refusal lists is still an irreversible command the operator chooses to run.

**What is deliberately unchanged:** `retireRunUnlessLive`, `ReconcileActiveRuns`, `runHasLiveRecords`, `retirable`, the classifier, the boot proof. The `Preserved` routing and every test that pins it are untouched.

### C.3 Operator retire surface (`factory_handoff_recover.go`, `factory_retire_guard_text.go`)

- `--retire` refused under REQ-RRG-001 (`errors.As(err, &*RunInProgressCardsError)`): write the refusal block (§E) to `cmd.OutOrStdout()` and return the typed error unchanged (non-zero exit, `errors.Is` still works). No prompt, flag, environment variable or confirmation exists; the command reads nothing from stdin. The classification printed is the one `RetireRunIfDead` returned alongside the error.
- `--retire` returning nil: print the existing first line `retired <run> (owner <classification>)`, then read `db.InProgressCards(ctx, run)` through the pool handle (after the retirement call has returned, so no transaction is open) and, when non-empty, print the stranded block (§E). A failed read returns its error (non-zero exit): an unreadable list is never silence. A run with no stranded cards prints the first line only, byte-identical to today.
- The flag help string `Retire the named run — refused unless its owner classifies dead` gains `and it holds no in-progress cards`. The long doc comment above `newFactoryRunsCommand` is updated; "There is no interactive prompt" stays true.

### C.4 Relaunch (`factory_relaunch.go`)

`relaunchRetireFromRun`: add a case, checked before the `ErrRunOwnerNotDead` case, for `errors.As(err, &cardsErr)` returning `run %s: left active (%d in-progress cards)`; every other outcome string is unchanged and the verb launches regardless, as today. A run that is not `active` never reaches this code (spec §C.8).

## D. Milestones (priority order; no time estimates)

Each milestone is RED-first: the named tests are written and observed failing for the stated reason before the code, then turned green. Ordered by reversibility: the data-model and typed-error interface first, the user-facing surface second.

### M1 — Homestate guard, atomic card read, stamp compare-and-set, retired no-op (Priority High; REQ-RRG-001..003)

- Files: new `internal/homestate/factory_run_cards.go` and `factory_run_retire_guard_test.go`; edit `internal/homestate/factory_run_retire.go`.
- Tests: `TestRetireRunIfDeadRefusesInProgressCards`, `TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers` (AC-RRG-001); `TestRetireRunIfDeadCardReadHoldsWriteLock` (AC-RRG-002); `TestRetireRunIfDeadFlipIsCASOnClassifiedStamp` (AC-RRG-003); `TestRetireRunIfDeadAlreadyRetiredIsNoOp` (AC-RRG-004).
- **Measure-first step for AC-RRG-002:** before relying on the pre-opened second handle, run the test once against a deliberately wrong placement (the read moved above `BeginTx`, the hook with it) and record that the second handle's insert then commits and the test fails; then restore the placement. The second handle is opened **before** `RetireRunIfDead` is called (`openFactoryPathBusy` runs DDL on open and may block while the first transaction holds the lock) with a short busy timeout.
- Done when: those pass with the swept-count rule met; the named existing homestate tests stay green; `GOOS=windows GOARCH=amd64 go build ./...` exits 0.

### M2 — Operator retire surface and relaunch outcome (Priority High; REQ-RRG-004..007)

- Files: edit `internal/cli/factory_handoff_recover.go`, `internal/cli/factory_relaunch.go`; new `internal/cli/factory_retire_guard_text.go`, `internal/cli/factory_runs_retire_guard_test.go`.
- Tests: `TestFactoryRunsRetireRefusesInProgressCards` (AC-RRG-005); `TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun` (AC-RRG-006); `TestRelaunchFromRunLeavesRunWithCardsActive` (AC-RRG-007); `TestFactoryRunsHasNoForceFlag` (AC-RRG-008, regression guard).
- Done when: those pass; `TestFactoryRunsCommandReportsAndRefuses` and `TestFactoryRelaunchFromRunRetiresDeadOwnerOnly` stay green; the catalogue source contains no `\u` escape (`grep -n '\\u' internal/cli/factory_retire_guard_text.go` prints nothing).

## E. Korean string catalogue (native written register, UTF-8, one constant block so tests assert on the constants; identifiers, states and commands verbatim)

Retirement is rendered `은퇴` to match the docs-site Korean page. A particle that depends on the final sound of a variable (run id, number) is avoided by wording, not guessed.

| Constant | Text |
|---|---|
| refusal header | `run %s 은퇴를 거절했습니다. 진행 중 카드 %d개가 남아 있습니다.` |
| owner line | `리더 상태: %s` |
| cards header | `진행 중 카드 (카드 id, 상태, 소유자):` |
| card line | `  %s  %s  %s` (an empty owner label renders `-`) |
| exits header | `은퇴 대신 쓸 수 있는 방법 (run이 활성인 동안):` |
| exit: join | `  레인으로 합류: moai cc -l --factory-run %s` |
| exit: adopt | `  리더로 이어받기: moai cc -f --factory-run %s (codex가 이끄는 run은 거절됩니다)` |
| exit: abandon | `  카드 종료: moai factory decide %s --choice abandon --run %s` |
| abandon note | `카드 종료는 되돌릴 수 없고, 레인 세션에서는 거절되며, 임대가 만료된 카드는 한 번 더 호출해야 종료됩니다.` |
| stranded header | `이미 은퇴한 run %s에 진행 중 카드 %d개가 남아 있습니다.` |
| stranded exit | `  카드 종료: moai factory decide %s --choice abandon --run %s` |
| stranded note | `--factory-run 플래그는 활성 run만 고르므로 이 run에는 합류하거나 이어받을 수 없습니다 (NO_ACTIVE_FACTORY).` |
| stranded abandon note | the `abandon note` row above, printed once after the stranded exits (the three caveats: irreversible; refused in a lane session; an expired lease needs a second call) |

The first line `retired <run> (owner <classification>)` and the relaunch outcome `run <run>: left active (<N> in-progress cards)` stay English: both extend vocabulary that existing tests already pin. The typed error's `Error()` text is English (library layer). This departs from the sibling SPEC t1848, which kept the whole `--retire` surface English; the dispatch for this card requires Korean operator prose, so the departure is recorded in `spec.md` §E rather than silently made.

## F. Self-verification (manager-develop reports, Tier M template §E)

- Scoped runs only: `go test -cover` for `internal/homestate` and `internal/cli` limited to the new test names plus the named existing tests. Do not run `go test ./...` or the full `internal/cli` suite locally; CI runs the full suite.
- `GOOS=windows GOARCH=amd64 go build ./...`.
- Single-gate smoke (a smoke check, not an acceptance proxy): `grep -rn "retirable(" internal --include='*.go'` filtered to non-test files; baseline measured at the run-phase start, expected unchanged. The single-gate and no-new-force-path properties are asserted behaviorally by AC-RRG-005 and AC-RRG-008.
- Caller inventory (§A) re-run after M2; `grep -rn "status='retired'" internal --include='*.go'` filtered to non-test files still lists exactly the two writers.
- Test isolation: factory databases live under the home state directory; fixtures follow `runOwnerSandbox` (`HOME`, `MOAI_HOME` scrubbed into `t.TempDir()`), and M1 confirms no test writes the real `~/.moai` database (a prior incident had tests writing the real backlog database).
- Template-First: no file under `internal/template/templates/` is touched (`grep -rln 'runs --retire' internal/template` printed nothing at this tree), so `make build` is not required by this SPEC; if run-phase edits reach a template, run `make build` and commit the regenerated embed.

## G. Risks

- **Re-retiring semantics (Q3).** A retired run retired again used to print success and write a duplicate event; now it prints success, writes nothing, and lists its stranded cards. A caller counting `run.retired` events would see one fewer; no caller found does (read, not run).
- **Lost compare-and-set reads as "owner not dead".** The error wraps `ErrRunOwnerNotDead`, so `relaunchRetireFromRun` maps a run another call just retired to `left active (owner ...)`. Cosmetic.
- **Hints that send an operator at `--retire`** (spec §G, Q5) now end in a refusal when the run holds cards. The refusal lists the exits, so nothing is lost; it is one extra step.
- **Reconciler gap.** Unchanged by this SPEC and stated as a residual risk in `spec.md` §G.
- **The refusal is temporary.** Card t1842 is expected to turn the refusal into a transfer; the docs say so.

## H. Anti-patterns to avoid

- A second retirement predicate (anything other than `retirable`), or a retire flip without the classified-stamp condition and the `RowsAffected` check (REQ-RRG-002, REQ-RRG-007).
- Reading the cards outside the retiring write-locked transaction, or through the pool handle while the transaction is open (REQ-RRG-001; deadlock).
- Injecting the competing card from `ReconcileOptions.Classify`: `Classify` runs before `BeginTx`, so it cannot tell an in-transaction read from a pre-transaction one (closed audit finding, `spec.md` §F); the in-read hook is the instrument.
- Placing the card guard ahead of the already-retired early return: it would refuse the idempotent retry the incident shape needs (REQ-RRG-003).
- A prompt, flag, environment variable or confirmation that retires a run with cards (REQ-RRG-004, REQ-RRG-007).
- Touching `retireRunUnlessLive` or the leader-start path in this SPEC (`spec.md` §G).

## I. Sync-phase obligations (no acceptance criterion pins these; see `spec.md` §H)

- `docs-site/content/{en,ko,ja,zh}/advanced/factory-mode.md` line 67 ends with "A run whose owner is dead can be retired with `moai factory runs --retire <run-id>`" (en wording): all four locales are updated in the same change (the oss-docs 4-locale same-PR obligation) to say that the command refuses a run that still holds in-progress cards, lists them with the exits, and that this is a temporary safeguard until card t1842 lands.
- `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` line 70 carries "`moai factory runs --retire <run-id>` retires one named run by hand. It refuses if the owner is not actually dead." (en wording): the same four-locale update. The sibling SPEC t1848 listed only the docs-site pages; the four README files were found by `grep -rln 'runs --retire' README.md README.ko.md README.ja.md README.zh.md docs-site/content`.
- Template-First: `grep -rln 'runs --retire' .claude .agents internal/template CLAUDE.md AGENTS.md AGENTS.local.md` printed nothing at this tree, so no template or rule file describes the command; the sync phase re-runs the grep.
- The sync phase reads the ja and zh pages and README files before editing (only the en/ko ones were read in part in this plan), and runs the Korean strings of §E through the native-idiom pass.

## J. Cross-references

SPEC-FACTORY-RUN-RETIRE-001 (the single retirement gate, REQ-005 and REQ-008), SPEC-FACTORY-STALE-RUN-HEAL-001 (the relaunch verb's `--from-run`), `.claude/rules/moai/development/verification-completeness.md` §2 (two-cell adoption), `.claude/rules/moai/core/verification-claim-integrity.md` §2.2 (tool provenance), cards t1853 (lane-join chooser), t1849 (leader-start prompt) and t1842 (retire = transfer).
