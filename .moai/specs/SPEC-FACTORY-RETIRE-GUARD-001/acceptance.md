# SPEC-FACTORY-RETIRE-GUARD-001 — Acceptance Criteria

Every AC is verifiable without a live factory run and without a terminal: instruments are Go tests on `t.TempDir()` factory databases. TTY detection is irrelevant (this SPEC adds no prompt). Each AC names the layer its test lives at — `homestate` (`internal/homestate`) or `cli` (`internal/cli`) — and every assertion below is producible at that layer: the homestate tests call `RetireRunIfDead` with an injectable `ReconcileOptions.Classify`; the cli tests run the cobra command with the production options (a real dead owner from `deadOwnerIdentity`) and cannot inject a classifier. The Go-test form is a single invocation, one package per command:

```text
go test <one package> -run '^(<TestName1>|<TestName2>)$' -count=1 -v
```

The run phase invokes it with the factory environment scrubbed (`MOAI_FACTORY_WORKERS`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_ROLE`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_KANBAN_ID`, `MOAI_KANBAN_BACKEND`, `MOAI_KANBAN_SETTINGS_INJECTED` unset), because a lane-stamped environment can falsify env-reading tests. The ledger below was measured in a session that carried those variables (see Gaps).

**Swept-count rule.** A Go-test AC passes only when the `-v` output shows one `--- PASS` line per named test; `[no tests to run]`, a PASS count below the named count, or a missing name is a FAIL. The `-run` pattern anchors every alternation branch.

**Two-cell discipline.** Each AC is either (a) RED-now with a named failing command — the ledger entry showing it red on the pre-implementation tree, and why — plus a green-path cell naming the milestone that flips it, or (b) explicitly labeled a **regression guard**: green on the pre-implementation tree by design, kept to pin behavior this work must not change, and never release-blocking. The RED-now evidence of kind (a) is two-layered: the named test does not exist, so the sweep is empty (E1, E2: red for the stated reason that the criterion asserts a behavior that does not exist yet, and not vacuous, because the swept-count rule turns the empty sweep into a FAIL); and, for the behaviors that change, a throwaway probe observed the old behavior directly (P1-P5, below). The probe source files were deleted and never committed; the behavioral REDs of the named tests themselves are recorded by the run phase at each milestone's RED step in `progress.md` §E.2.

## Evidence ledger (pinned tree `d2d9ac08fb1899440da911121ddf13eecc806a3f`; the SPEC files are the only difference from it; `git status --short` was empty before the SPEC files were written)

```text
E1  command : go test ./internal/homestate -run '^(TestRetireRunIfDeadRefusesInProgressCards|TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers|TestRetireRunIfDeadCardReadHoldsWriteLock|TestRetireRunIfDeadFlipIsCASOnClassifiedStamp|TestRetireRunIfDeadAlreadyRetiredIsNoOp)$' -count=1 -v
    stdout  : testing: warning: no tests to run
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/homestate	0.322s [no tests to run]
    exit    : 0 (read from the `ok` line; a failing run prints `FAIL` and exits 1)
    verdict : 0 of 5 named tests swept -> FAIL under the swept-count rule
E2  command : go test ./internal/cli -run '^(TestFactoryRunsRetireRefusesInProgressCards|TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun|TestRelaunchFromRunLeavesRunWithCardsActive|TestFactoryRunsHasNoForceFlag)$' -count=1 -v
    stdout  : testing: warning: no tests to run
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/cli	0.882s [no tests to run]
    exit    : 0 (read from the `ok` line)
    verdict : 0 of 4 named tests swept -> FAIL
E3  command : go test ./internal/homestate -run '^(TestRetireRunIfDeadRefusesLiveAndIndeterminate|TestEveryRetirementPathDeclinesUnenumeratedClassification|TestRetireRunIfDeadAcceptsBootProof|TestRetireRunPreservesRowAndAppendsEvent)$' -count=1 -v
    stdout  : === RUN   TestRetireRunIfDeadAcceptsBootProof
              --- PASS: TestRetireRunIfDeadAcceptsBootProof (0.01s)
              === RUN   TestRetireRunPreservesRowAndAppendsEvent
              --- PASS: TestRetireRunPreservesRowAndAppendsEvent (0.32s)
              === RUN   TestEveryRetirementPathDeclinesUnenumeratedClassification
              === RUN   TestEveryRetirementPathDeclinesUnenumeratedClassification/reconciler
              === RUN   TestEveryRetirementPathDeclinesUnenumeratedClassification/migration_pass
              === RUN   TestEveryRetirementPathDeclinesUnenumeratedClassification/operator_retire
              --- PASS: TestEveryRetirementPathDeclinesUnenumeratedClassification (0.63s)
                  --- PASS: TestEveryRetirementPathDeclinesUnenumeratedClassification/reconciler (0.25s)
                  --- PASS: TestEveryRetirementPathDeclinesUnenumeratedClassification/migration_pass (0.17s)
                  --- PASS: TestEveryRetirementPathDeclinesUnenumeratedClassification/operator_retire (0.21s)
              --- PASS: TestRetireRunIfDeadRefusesLiveAndIndeterminate (0.62s)
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/homestate	1.771s
    exit    : 0
    verdict : the existing retirement-gate guards are GREEN at the pinned tree (baseline for AC-RRG-001 controls and AC-RRG-008)
E4  command : go test ./internal/cli -run '^(TestFactoryRunsCommandReportsAndRefuses|TestFactoryRelaunchFromRunRetiresDeadOwnerOnly)$' -count=1 -v
    stdout  : === RUN   TestFactoryRelaunchFromRunRetiresDeadOwnerOnly
              === RUN   TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runDead
              === RUN   TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runLive
              === RUN   TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runUnknownOwner
              === RUN   TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runRetired
              === RUN   TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runNoSuchRun
              --- PASS: TestFactoryRelaunchFromRunRetiresDeadOwnerOnly (3.74s)
                  --- PASS: TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runDead (0.60s)
                  --- PASS: TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runLive (0.60s)
                  --- PASS: TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runUnknownOwner (0.64s)
                  --- PASS: TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runRetired (0.40s)
                  --- PASS: TestFactoryRelaunchFromRunRetiresDeadOwnerOnly/runNoSuchRun (0.31s)
              === RUN   TestFactoryRunsCommandReportsAndRefuses
              --- PASS: TestFactoryRunsCommandReportsAndRefuses (1.75s)
              PASS
              ok  	github.com/modu-ai/moai-adk/internal/cli	6.191s
    exit    : 0
    verdict : the existing cli surface guards are GREEN at the pinned tree (baseline for AC-RRG-005..008)
```

**Behavioral probes (throwaway tests in `internal/homestate` and `internal/cli`, run with `-run '^TestProbeT1852'`, deleted afterwards; verbatim `t.Logf` lines, same pinned tree).**

```text
P1  dead run + one `needs-decision` card, RetireRunIfDead with Classify=dead
    PROBE guard: err=<nil> status=retired events=1                                   -> the run is retired today although it holds a card (AC-RRG-001 RED)
P2  the same card-free dead run, RetireRunIfDead called twice
    PROBE re-retire call 1: err=<nil> status=retired events=1
    PROBE re-retire call 2: err=<nil> status=retired events=2                        -> a second run.retired event today (AC-RRG-004 RED)
P3  AdoptRun(pid 4242, "adopter-start") lands inside Classify, then the flip runs
    PROBE cas: err=<nil> status=retired stamp=(4242,adopter-start) events=1          -> retired under the adopter's stamp today (AC-RRG-003 RED)
P4  `moai factory runs --retire runC`, dead owner, two cards (needs-decision, picked)
    PROBE cli retire: err=<nil> out="retired runC (owner dead)\n" status=retired      -> success, no card listing today (AC-RRG-005 RED)
P5  relaunchRetireFromRun for the same shape (runD)
    PROBE relaunch: outcome="run runD: retired (owner dead)" status=retired          -> the run is retired today (AC-RRG-007 RED)
```

Not probed, and why: AC-RRG-002 (the hook that makes the placement observable does not exist yet; RED-now is E1) and AC-RRG-006 (the retired-run-with-cards command output; its homestate half is P2, RED-now at the cli layer is E2).

**Regression guard (AC-RRG-008), stated once.** `newFactoryRunsCommand` registers exactly one flag, `retire` (a single `cmd.Flags().StringVar(&retire, "retire", ...)`, `factory_handoff_recover.go:110`), and the retirement-gate guards of E3 already pass. Both halves are green on the unmodified tree, so no RED-now cell exists for AC-RRG-008 by design; E2 lists its name only because the test file does not exist yet. The property it pins — no override flag, no second retirement predicate — is the card's prohibition on a live-leader cleanup path, which no behavior change of this work should flip. The criterion that is red today for the refusal itself is AC-RRG-005.

## AC Matrix

| ID | Requirement | Scenario (Given / When / Then) | Test (layer, milestone) | RED-now | Green path | Severity |
|---|---|---|---|---|---|---|
| AC-RRG-001 | REQ-RRG-001 | **Given** an `active` run R classified dead on its stamp (`Classify` fixed to dead), holding cards in the states `needs-decision`, `merged-local` and `picked` (all outside the lease-holding set) and one `done` card **When** `RetireRunIfDead` runs **Then** `errors.Is(err, ErrRunHasInProgressCards)` holds, `errors.Is(err, ErrRunOwnerNotDead)` does not, `errors.As` yields a `*RunInProgressCardsError` whose run id is R and whose cards are exactly the three non-terminal ones (id, state, owner label, ordered by card id), R stays `active`, zero `run.retired` events exist, and the full `runs` row (including `updated_at`) equals the row read just before the call (no competing write occurs in this test, so the pre-call baseline is valid). **Controls:** a dead run with only `done`, `failed` and `abandoned` cards, and a dead run with only worker rows, both retire with the payload `{"classification":"dead","basis":"stamp"}`; a live-owner and an indeterminate-owner run that each hold a non-terminal card are refused with `ErrRunOwnerNotDead` and **not** with the cards error. Mutant probes: a guard keyed on lease-holding states only fails the `picked` / `needs-decision` / `merged-local` cards; a guard that counts terminal cards fails the control; a guard placed before the owner gate fails the live / indeterminate control. | `TestRetireRunIfDeadRefusesInProgressCards`, `TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers` (homestate, M1) | P1 (retired today with a card); E1 sweep empty | M1 | Release-blocking |
| AC-RRG-002 | REQ-RRG-001 | **Given** a dead-owner run R with no cards, a second factory handle to the same database file opened **before** the call (`openFactoryPathBusy` on `db.Path`, short busy timeout), and the `retireAfterCardRead` hook set to insert one in-progress card for R through that second handle **When** `RetireRunIfDead` runs **Then** the hook's insert fails with a busy / locked error (the card read ran inside the write-locked transaction), R is retired, and the card is absent from `cards` afterwards. Mutant probe ("card read moved before `BeginTx`"): the hook moves with the read, the write lock is not held, the second handle's insert commits, the hook reports success, and the test fails. The hook is called from inside the card-read helper right after its `SELECT`, so a refactor that moves only the read cannot leave the hook behind. The competing card is not injected from `Classify`, which runs before `BeginTx` and could not tell the two placements apart. | `TestRetireRunIfDeadCardReadHoldsWriteLock` (homestate, M1) | E1 sweep empty (the hook does not exist) | M1 (with the measure-first step of plan §D) | Release-blocking |
| AC-RRG-003 | REQ-RRG-002 | **Given** a dead-owner run R and a `Classify` closure that, on its first call only, lands `AdoptRun` for R with the adopter's stamp (pid `4242`, start `adopter-start`, supplied by the test) after the stamp was read **When** `RetireRunIfDead` runs **Then** (a) for a card-free R: the error wraps `ErrRunOwnerNotDead` and its text states that the owner or status changed, R is `active`, its stamp equals the adopter's stamp (asserted against the values the test supplied, not against a pre-race snapshot, because the adoption is itself a write to `runs`), and zero `run.retired` events exist; (b) for an R carrying a non-terminal card (the competing adopt lands before the transaction, so the stamp check of plan §C.2 step 0 sees it): the error wraps `ErrRunOwnerNotDead` and states that the owner or status changed, `errors.Is(err, ErrRunHasInProgressCards)` is **false** (no cards error, so the CLI never lists the abandon exit for a run a live leader just adopted), R `active` under the adopter's stamp, zero events, the baseline taken after the competing write; (c) when the closure instead retires R through a nested `RetireRunIfDead` (a concurrent retirer wins; the closure returns dead without recursing on later calls) the outer call returns an error wrapping `ErrRunOwnerNotDead` and exactly **one** `run.retired` event exists, the baseline being taken after the nested retirement; (d) a run whose row stamp is `(0,'')` and which classifies dead through the lead-peer fallback (`ClearRunOwner`, then `ReconcileOptions.Fallback` supplying the identity) retires with the payload basis `peer`, and the existing `TestRetireRunIfDeadAcceptsBootProof` stays green (basis `boot`, E3). Mutant probes: a flip on `status='active'` alone (today's code) retires the adopted run and fails (a); the card read placed ahead of the stamp check returns the cards error in (b) and fails it; a retirement with neither the stamp check nor the `RowsAffected` check appends a second event and fails (c). | `TestRetireRunIfDeadFlipIsCASOnClassifiedStamp` (homestate, M1; table rows (a)-(d)) | P3 (retired under the adopter's stamp today); E1 sweep empty | M1 | Release-blocking |
| AC-RRG-004 | REQ-RRG-003 | **Given** a dead-owner run R whose status is `retired` (seeded: cards `needs-decision` and `picked` inserted, then `status` set to `retired` and one `run.retired` event inserted) **When** `RetireRunIfDead` runs **Then** it returns the classification `dead` and a nil error; the `runs` row, the `cards` rows and the `run.retired` event count all equal the snapshot taken just before the call (no competing write occurs in this test); no second event exists; the card guard did **not** fire; and `InProgressCards(R)` after the call lists exactly the two cards. **Controls:** a `retired` run whose owner classifies live is still refused with `ErrRunOwnerNotDead`; an `active` dead run holding the same two cards is refused with the cards error (the status, not the cards, distinguishes the two). Mutant probes: a card guard placed ahead of the already-retired early return refuses R and fails; no early return (today's code) appends a second event and fails; an early return placed before the owner gate returns nil for the live control and fails. | `TestRetireRunIfDeadAlreadyRetiredIsNoOp` (homestate, M1) | P2 (second event today); E1 sweep empty | M1 | Release-blocking |
| AC-RRG-005 | REQ-RRG-004, REQ-RRG-007 | **Given** run R with a dead owner (production options, no injected classifier) holding two non-terminal cards (`needs-decision` with owner label `lane-1`, `picked` with an empty label) and a command whose stdin is a reader that fails the test if read **When** `moai factory runs --retire R` runs **Then** the returned error satisfies `errors.Is(err, ErrRunHasInProgressCards)` (non-zero exit); the captured output contains the refusal header of plan §E with R and the count `2`, the owner line with `dead`, each card id with its state and owner label (`-` for the empty one), the exits verbatim — `moai cc -l --factory-run R`, `moai cc -f --factory-run R`, and `moai factory decide <card> --choice abandon --run R` once per card — and the abandon note; the header contains at least one Hangul rune (so an English catalogue cannot satisfy the Korean-prose clause); R stays `active`, zero `run.retired` events exist, and nothing was read from stdin. The same assertions hold when `MOAI_FACTORY_WORKER` is set (a lane session) and when it is unset (refusal in every session). Mutant probes: a command that prints the block and exits zero fails; one that prompts for confirmation fails the stdin guard; one that omits an exit fails the verbatim check; one that retires and then prints fails the `active` check. | `TestFactoryRunsRetireRefusesInProgressCards` (cli, M2) | P4 (success, no listing today); E2 sweep empty | M2 | Release-blocking |
| AC-RRG-006 | REQ-RRG-005 | **Given** (A) a dead-owner run R seeded `retired` holding a `needs-decision` card and a `picked` card and one `run.retired` event; (B) a `retired` dead-owner run holding only a `done` card; (C) an `active` card-free dead-owner run **When** `moai factory runs --retire <run>` runs on each **Then** (A) exits zero; the first line is exactly `retired R (owner dead)`; the stranded header of plan §E follows with R and the count `2`, both cards with state and owner label, `moai factory decide <card> --choice abandon --run R` per card, the stranded note (the `--factory-run` flag selects only an active run), and the abandon note of plan §E carrying its three facts (irreversible; refused in a lane session; an expired lease needs a second call); neither `moai cc -l` nor `moai cc -f` appears, and no text claims that the abandon exit is the only one (the output contains neither `남은 방법` nor an `only`/`유일` claim); the `runs`, `cards` and `events` tables equal the snapshot taken just before the call (no write occurs); (B) prints the first line only; (C) prints the first line only, R is `retired` and exactly one `run.retired` event exists. Mutant probes: a command that prints nothing for stranded cards fails (A); a stranded block that lists the abandon command without the abandon note fails (A); a stranded block that labels abandon the sole exit fails (A); one that prints the block for (B) or (C) fails; one that offers join or adopt for a non-active run fails (A). | `TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun` (cli, M2) | E2 sweep empty (the homestate half of the old behavior is P2) | M2 | Release-blocking |
| AC-RRG-007 | REQ-RRG-006 | **Given** an `active` run R with a dead owner holding two non-terminal cards **When** `relaunchRetireFromRun` runs for R, and again through `moai factory relaunch --provider cc --from-run R` with the launcher seam captured **Then** the outcome text contains `left active` and `2 in-progress cards`, R stays `active`, zero `run.retired` events exist, and the verb still launches `cc -l` once. The relaunch verb's other outcomes keep their text (`TestFactoryRelaunchFromRunRetiresDeadOwnerOnly`, E4, stays green). | `TestRelaunchFromRunLeavesRunWithCardsActive` (cli, M2) | P5 (retired today); E2 sweep empty | M2 | High |
| AC-RRG-008 | REQ-RRG-007 | **Regression guard (green on the unmodified tree by design; see the note after the ledger).** **Given** the `moai factory runs` command constructed fresh and enumerated before execution (cobra adds `help` at execution time) **When** its local flags are enumerated **Then** the set is exactly `{retire}` — no `--force`, `--yes`, `--discard`, `--abandon` or `--confirm*` flag exists — and the existing gate guards stay green: a live-owner or indeterminate-owner run is refused, a classification outside the enumerated three retires nothing on any path (E3). Mutant probe: adding any acknowledging flag changes the set and fails. | `TestFactoryRunsHasNoForceFlag` (cli, M2) + E3 tests unchanged | none by design: regression guard (E3 guards GREEN today) | M2 (keeps it green) | Regression guard |

Counts: 7 requirements and 8 acceptance criteria (Tier M ceilings 16 and 16). Release-blocking, each with an attached RED-now cell (P1-P5 and E1, E2): AC-RRG-001, 002, 003, 004, 005, 006 (six). High: AC-RRG-007. Regression guard: AC-RRG-008 (one).

Requirement-to-criterion map: REQ-RRG-001 -> AC-001 [REF], AC-002 [REF]; REQ-RRG-002 -> AC-003 [REF]; REQ-RRG-003 -> AC-004 [REF]; REQ-RRG-004 -> AC-005 [REF]; REQ-RRG-005 -> AC-006 [REF]; REQ-RRG-006 -> AC-007 [REF]; REQ-RRG-007 -> AC-005 [REF], AC-008 [REF].

## Edge cases

- A dead-owner run with only worker rows and no cards retires as today (AC-RRG-001 control); the workers remain, as today.
- A run retired between classification and the flip by a concurrent retirer: the compare-and-set matches zero rows and the call returns the owner-or-status-changed error with one event in total (AC-RRG-003 row (c)). A run retired *before* classification takes the no-op (AC-RRG-004).
- A restamped run that also carries a card gets the owner-changed error, not the cards error, because the stamp check precedes the card read; nothing is written and no abandon exit is shown for it (AC-RRG-003 row (b)). Only a run whose classified owner is still on the row reaches the card read and the cards refusal.
- The incident shape: an `active` run, dead leader, many cards is refused (AC-RRG-005); the same run retired earlier reports its cards with the abandon exit and exits zero (AC-RRG-006).
- `moai factory relaunch --from-run` for a `retired` run never reaches `RetireRunIfDead` (`not active (retired)`, existing behavior, E4).

## Quality gate and Definition of Done

- All 8 ACs satisfy the swept-count rule at the final tree (AC-RRG-008 as a regression guard: its test passes and the E3 guards stay green); the E3 and E4 guards and the named existing suites (`factory_run_retire_test.go`, `factory_run_boot_proof_test.go`, `factory_runs_command_test.go`, `factory_relaunch_test.go`) are green.
- `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint` clean for the touched packages; new files at or above 85% statement coverage by the named tests.
- No new force-cleanup path and no confirmation override (AC-RRG-008); every retirement this SPEC changes is a compare-and-set on the classified stamp (AC-RRG-003).
- `progress.md` §E.2 carries, per milestone, the verbatim RED output before GREEN.

## Gaps (what the criteria above do not observe)

- The ledger and probes ran with the lane environment present (`MOAI_KANBAN_ID`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_ROLE`, `MOAI_FACTORY_WORKERS`, `MOAI_FACTORY_CLEAR_POLICY`, `MOAI_FACTORY_AUTO_DISPATCH`, `MOAI_KANBAN_BACKEND`, `MOAI_KANBAN_SETTINGS_INJECTED` were set); the empty sweeps and the probes do not read those variables, but the scrubbed form was not observed.
- Exit codes were read from the runner's `ok` / `FAIL` line; the single-invocation form forbids `; echo $?`.
- The probe source files no longer exist; their output lines above are the record. The probes used `Classify` fixed to dead (homestate) and the production options (cli).
- No test of the new fixtures exists; the AC fixtures were read against `factory_run_retire_test.go`, `factory_runs_command_test.go` and `factory_relaunch_test.go` helpers, not run.
- AC-RRG-002 and AC-RRG-006 have no behavioral RED, only an empty sweep.
