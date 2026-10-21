---
id: SPEC-FACTORY-RETIRE-GUARD-001
title: "Factory run retirement guard — operator run retirement refuses a run that still holds in-progress cards"
version: "0.1.0"
status: completed
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate, internal/cli"
lifecycle: spec-anchored
tier: M
tags: "factory, run-retire, retire-guard, in-progress-cards, compare-and-set, card-t1852"
card: t1852
related_specs: [SPEC-FACTORY-RUN-RETIRE-001, SPEC-FACTORY-STALE-RUN-HEAL-001]
---

# SPEC-FACTORY-RETIRE-GUARD-001 — Factory run retirement guard

## A. History

| Date | Version | Change |
|---|---|---|
| 2026-10-17 | 0.1.0 | Initial plan-phase draft (card t1852, Tier M, Class C). Part C only of the twice-failed card line t1825 -> t1848. The closed audit decisions that concern part C are carried, not re-argued (§F). Tree `d2d9ac08fb1899440da911121ddf13eecc806a3f`; every code fact in §C was re-read at that tree. |

## B. Problem Statement

Observed 2026-10-17 (evidence case taken from the card body; the production factory database was not read for this SPEC, see §H): run `tmvgvn` was active with a dead leader and 41 in-progress cards. The operator ran `moai factory runs --retire tmvgvn`. The command retired the run and left the 41 card records inside a retired run, so lane `stage` / `complete` calls on the replacement run failed with `factory card not found`.

`moai factory runs --retire` retires any run whose owner classifies dead, whatever cards the run carries (§C.1). This SPEC is the **temporary safeguard** the card asks for: refuse, and list the cards, until card t1842 ("retire = transfer") gives retirement a way to move the cards. The refusal is expected to be revised by that card.

## C. Root Cause (code facts re-read at tree `d2d9ac08f`)

1. **Operator retirement.** `RetireRunIfDead` (`internal/homestate/factory_run_retire.go:381`) classifies every run row, refuses a non-`dead` classification through `retirable` (`:165`), and calls `retireRun` (`:407`). `retireRun` flips with `UPDATE runs SET status='retired' ... WHERE run_id=? AND status='active'` (`:414`), never reads `RowsAffected`, and appends `run.retired` unconditionally; its own comment says it "retires regardless of the records the run carries" (`:401-406`). Consequences measured by a throwaway probe (`acceptance.md` evidence ledger P1-P5, §H): a dead run holding a card is retired; a second retirement of the same run appends a second `run.retired` event; an owner restamp landing between classification and the flip is retired anyway under the new owner's stamp.
2. **Callers.** `RetireRunIfDead` has exactly two non-test callers: `moai factory runs --retire` (`internal/cli/factory_handoff_recover.go:91`) and `moai factory relaunch --from-run` (`internal/cli/factory_relaunch.go:98`). `retireRun` has one caller, `RetireRunIfDead` (`:393`).
3. **Two retirement writers exist; only one is changed here.** `grep -rn "status='retired'"` over non-test Go files returns two statements: `retireRun` (`:414`) and `retireRunUnlessLive` (`:445`), the automatic reconciler's write, called only by `ReconcileActiveRuns` (`:355`). `retireRunUnlessLive` is unchanged by this SPEC (§G).
4. **The transaction is write-locked from its first statement.** `OpenFactory` (`internal/homestate/factory.go:171`) opens through `OpenFactoryPath` (`:226`) -> `openFactoryPathBusy` (`:292`), whose DSN carries `_txlock=immediate` (`:303`) and whose pool is capped at one connection (`:313`). `BeginTx` therefore issues `BEGIN IMMEDIATE`: a second handle's write is refused (busy) while the first holds the transaction. The same one-connection pool means a read through the pool handle while a transaction is open deadlocks, so the in-transaction read must use the transaction (the hazard `runHasLiveRecords` documents, `factory_run_adopt.go:28-30`).
5. **Compare-and-set precedent.** `AdoptRun` (`factory_run_adopt.go:128`) restamps with `WHERE run_id=? AND status='active' AND lead_pid=? AND lead_process_start=?` and reads `RowsAffected`; the retirement flip adopts the same shape.
6. **Terminal card states are one shared set.** `terminalStates = {done, failed, abandoned}` (`internal/homestate/card_record.go:74`); the SQL literal `sqlTerminalStates` (`factory_run_adopt.go:21`) spells the same set for the automatic live-records guard. Every other card state, including `picked`, `needs-decision`, `blocked`, `merged-local` and `pushed`, is non-terminal.
7. **Exits the refusal may cite, read in code (none was run).** (a) `moai factory decide <card> --choice abandon --run <run>`: the verb and flags exist (`newFactoryDecideCommand`, `internal/cli/factory_card.go:3255`); `--run` is taken verbatim and need not name an active run (`resolveFactoryCardRun`, `:2646`); the transition table admits abandon from every non-terminal state (`internal/homestate/card_transition.go:172-176`, edge T25) for decider `human` (`:702-706`), which is the flag default; the verb is refused in a lane session (`factoryDecideLaneRefusal`, `factory_card.go:3219`, applied at `:3270`); a card whose lease has expired returns `ErrLeaseExpired` on the first call (`card_transition.go:315`) and needs a second call; abandonment cannot be undone. (b) `moai cc -l --factory-run <run>` joins a run only while it is `active`: `ResolveActiveRun` returns `NO_ACTIVE_FACTORY` for an explicit id whose status is not `active` (`internal/factorymsg/store.go:355`). (c) `moai cc -f --factory-run <run>` resolves through the same check first (`cc.go:203`, `enterSelectedFactoryRun` in `factory.go:468`), so it too works only while the run is `active`, and it is refused for a run led by codex (`refuseCodexLeaderRun`, `factory.go:674`). What is established is therefore narrow: the `--factory-run` flag cannot revive a non-active run. (d) **Revival routes outside that flag exist (read in code, not run here; the plan-audit iteration 1 codex leg reproduced the first, `.moai/reports/t1852/plan-audit.md`).** `RecordRun` re-activates an existing run unconditionally (`internal/homestate/runtime.go:44-46`, `ON CONFLICT(run_id) DO UPDATE ... status='active'`), and a leader start reaches it when its run id comes from a `leader-<run-id>` name label or from the run-id environment marker (`leaderRunID`, `factory_launch_helpers.go:81-88`, then `recordLeaderRunStart`, `factory_run_adopt.go:173`, then `recordFactoryRunStart`, `factory.go:695`, then `RecordRun`). `ResumeRun` (`internal/homestate/factory_run_resume.go`, upsert guarded by `WHERE runs.status!='active'`) is a second re-activating writer, used by the lane-join discovery for a verified live leader (`resumeDiscoveredRun`, `factory.go:657`). This SPEC does not offer either route as an operator exit, adds no requirement about them, and does not claim the abandon exit is the only one; they are a residual risk recorded here: a retired run that still holds cards can be made `active` again by a leader start that carries its run id, their behavior is unchanged, and the refusal and the stranded block say only what the `--factory-run` flag does.
8. **`relaunch --from-run` never reaches a retired run.** `relaunchRetireFromRun` (`factory_relaunch.go:75`) first probes the run state and returns `not active (...)` for any status other than `active` (`:88-91`); only an `active` run reaches `RetireRunIfDead`. Its outcome vocabulary is English (`retired (owner ...)`, `left active (owner ...)`, `could not be retired (...)`).

## D. Requirements (GEARS)

Terms. **Non-terminal card** (the card's "in-progress card"): a card whose state is outside `{done, failed, abandoned}` (§C.6); stated once here and used by every requirement. **Dead owner**: a run owner classified `dead` by the existing classifier, the single retirement gate `retirable`. **Stranded cards**: non-terminal cards inside a run whose status is `retired`.

- **REQ-RRG-001** (Event-detected): When `RetireRunIfDead` finds a run that is `active`, whose owner is classified dead, and that holds at least one non-terminal card, it shall refuse with a typed error (matchable with `errors.Is` / `errors.As`, distinct from the owner-not-dead error) carrying the run id and every non-terminal card (id, state, owner label), and shall change nothing: no status change, no event, no `updated_at` change. It shall read the run's non-terminal cards inside the same write-locked transaction as the status flip, so that no card write can commit between that read and the flip. A run classified live or indeterminate shall keep being refused as owner-not-dead, whether or not it holds cards (the owner gate precedes the card read).
- **REQ-RRG-002** (Ubiquitous): The retirement flip shall change a run only while its status is still `active` and its owner stamp (lead pid and process-start fingerprint) still equals the stamp the classification read, shall read the number of changed rows, and shall append the `run.retired` event only when exactly one row changed. When no row changed it shall write nothing and return an error that wraps `ErrRunOwnerNotDead` and states that the run's owner or status changed since classification. For a run classified dead through the lead peer or the boot proof, the compared stamp is the row's own `(0, '')`. This requirement binds the operator retirement path only; no other retirement writer is changed (§G).
- **REQ-RRG-003** (Event-driven): When `RetireRunIfDead` finds a run whose status is already `retired` and whose owner is classified dead, it shall return the classification and a nil error and shall perform no write, no card read and no second `run.retired` event, whatever cards the run holds; the card guard of REQ-RRG-001 applies to `active` runs only, and the stranded cards of a retired run are reported by REQ-RRG-005, never refused and never silent. A retired run whose owner is not classified dead shall keep being refused as owner-not-dead.
- **REQ-RRG-004** (Event-driven): When `moai factory runs --retire <run>` is refused under REQ-RRG-001, the command shall print the run id, the owner classification, each non-terminal card (id, state, owner label) and the exits of §C.7 — join with `moai cc -l --factory-run <run>`, adopt with `moai cc -f --factory-run <run>` (both only while the run is `active`), and end each card with `moai factory decide <card> --choice abandon --run <run>` — together with the facts that abandonment is irreversible, that the verb is refused in a lane session, and that an expired lease needs a second call; it shall exit non-zero in every session; and it shall offer no prompt, flag, environment variable or confirmation that retires the run, and read nothing from stdin (refuse-only). Operator-facing prose is Korean (UTF-8) with identifiers, states and commands verbatim.
- **REQ-RRG-005** (Event-driven): When `moai factory runs --retire <run>` returns success (a fresh retirement, or the idempotent no-op of REQ-RRG-003) and the run holds stranded cards, the command shall keep its first line `retired <run> (owner <classification>)`, shall exit zero, and shall print an informational block listing those cards (id, state, owner label) with `moai factory decide <card> --choice abandon --run <run>` per card as an available exit, together with the same three caveats REQ-RRG-004 states (abandonment is irreversible; the verb is refused in a lane session; a card whose lease has expired needs a second call), and the statement that the `--factory-run` flag selects only an active run, so it cannot join or adopt this one; the block shall not claim that any listed exit is the only one. A run with no stranded cards shall print no block. Operator-facing prose is Korean with identifiers verbatim.
- **REQ-RRG-006** (Event-driven): When `moai factory relaunch --from-run <run>` is refused under REQ-RRG-001, it shall report the run as left active because of its in-progress cards, in the form `run <run>: left active (<N> in-progress cards)`; the launch proceeds as today, and every other outcome text of the verb is unchanged.
- **REQ-RRG-007** (Unwanted behavior): No change made under this SPEC shall add a flag, prompt, environment variable or acknowledging variant that retires a run holding non-terminal cards or a run whose owner is not classified dead; `retirable` shall remain the only retirement predicate, the flag set of `moai factory runs` shall remain exactly `{retire}`, and the automatic reconciler's retirement write (`retireRunUnlessLive`, `ReconcileActiveRuns`) shall stay unchanged.

## E. Decision Points

Full rows with routing labels are in `decision-index.md`. **Dispatch-fixed scope** (not decision rows): refuse-only retirement, no force-cleanup path for a live-leader run (`retirable` stays the single gate), operator-facing prose in Korean, and the part split (part C only; §G).

| DP | Decision | Disposition |
|---|---|---|
| Q1 | How far the compare-and-set reaches: the operator path only, or also the automatic reconciler. | **Default applied: operator path only**; the reconciler's residual risk is stated in §G. |
| Q2 | Definition of "in-progress card". | **Default applied: every non-terminal state** (§C.6). |
| Q3 | Re-retiring an already-retired run, including one that still holds cards (the incident shape). | **Default applied: idempotent no-op plus an informational block** listing the stranded cards (REQ-RRG-003, REQ-RRG-005). |
| Q4 | Labeling the "no force flag" criterion, which is green on the unmodified tree. | **Default applied: regression guard**, green by design, never release-blocking (AC-RRG-008). |
| Q5 | The five existing hint strings that name `moai factory runs --retire`. | **Default applied: unchanged**; the refusal itself carries the exits (§G). |

## F. Prior audit history (closed — carried, not re-argued)

This part C was audited three times inside two larger SPECs that hit the Tier M ceiling and were split: `SPEC-FACTORY-RUN-SELECT-001` (card t1825, tree `971e2668c`, verdicts `.moai/reports/t1825/`) and `SPEC-FACTORY-RUN-SELECT-CORE-001` (card t1848, tree `b22972c55`, verdicts `.moai/reports/t1848/plan-audit.md`, `plan-audit-iter2.md`, `plan-audit-iter3.md`, ceiling record `.moai/state/audit-ceiling/SPEC-FACTORY-RUN-SELECT-CORE-001.json`). Those files live in the two sibling card worktrees and are not committed to this tree's history; they are read-only references, not authority. The t1848 iteration-3 verdict is FAIL 0.90 with a single blocking defect in the lane chooser (part A, not this SPEC); no part-C defect was open at that verdict. The part-C decisions below are closed and are implemented here as written.

| Closed defect (source) | Decision carried | Where it lands |
|---|---|---|
| t1848 D1: guard vs retired run holding cards | The card guard applies to `active` runs only; a retired run is an idempotent no-op that lists its stranded cards; the retired-run AC fixture carries non-terminal cards | REQ-RRG-001/003/005, AC-RRG-004, AC-RRG-006 |
| t1848 D5: criterion green on the unmodified tree | The "no force flag" criterion is labeled a regression guard | AC-RRG-008, Q4 |
| t1825 D1 / N1: flip not CAS'd; reconciler not CAS'd | CAS on the classified owner stamp, `RowsAffected` read; the claim is narrowed to the one writer changed | REQ-RRG-002, §C.3, §G |
| t1825 D2 / N4: AC cannot tell in-transaction from pre-transaction read | A test hook fires inside the read helper right after its `SELECT`; a second connection commits a card there | AC-RRG-002 |
| t1825 N5: CAS stamp for peer / boot basis is `(0,'')` | Stated in REQ-RRG-002; the degenerate comparison is a residual risk | §G |
| t1825 N6: abandon exit caveats | Lane-session refusal, expired-lease second call, irreversibility are stated in the refusal text | REQ-RRG-004/005 |
| t1825 N3 / N12, t1848 O1: revival claim | `--factory-run` cannot revive a non-active run; `cc -f --factory-run` works only while the run is `active` | §C.7, REQ-RRG-004/005 |
| t1825 N13: baseline equals pre-call snapshot after a competing write | Every AC that races a write takes its baseline after the competing write | AC-RRG-003 |
| t1848 O4: "refuses when the run holds cards" read as including a retired run | Disclosed: the incident shape (active run, dead owner, cards) is refused; a retired run is a no-op that reports its cards | REQ-RRG-003, this table |

## G. Out of Scope

### Out of Scope — Lane-join chooser, TTY prompts, KST rendering, per-run facts table

- The interactive lane-join run chooser, any prompt, the KST rendering of run times and the shared per-run facts query. Moved to card t1853 (`SPEC-LANE-JOIN-RUN-SELECT-001`). No requirement here covers them; this SPEC adds no prompt, so TTY detection is irrelevant to it.

### Out of Scope — Leader-start adopt prompt

- The `moai cc -f` / `moai glm -f` prompt that offers adopt, new run or cancel when a dead-owner run carries cards. Moved to card t1849. Nothing here depends on it: the refusal names the existing deliberate command `moai cc -f --factory-run <run>`.

### Out of Scope — Cross-run card-record transfer

- A command or function that moves card records, leases, approvals, dispatch rows or broker state from one run to another ("retire = transfer"). Moved to card t1842. The refusal of REQ-RRG-001 is a temporary safeguard that card is expected to revise.
- Recovering the 41 cards already stranded in the retired run `tmvgvn`; REQ-RRG-005 only reports them.

### Out of Scope — New cleanup power over live runs

- Any retirement, forced cleanup, takeover or restamp of a run whose leader is alive or indeterminate. `retirable` stays the single gate (REQ-RRG-007). Any change to the owner classifier, the boot proof or `ReconcileActiveRuns` semantics.

### Out of Scope — Compare-and-set on the automatic reconciler (residual risk)

- `retireRunUnlessLive` (`factory_run_retire.go:431`) keeps flipping on `run_id` and `status='active'` alone, ignores `RowsAffected`, and writes `run.retired` unconditionally; `ReconcileActiveRuns` appends to `Preserved` on every `false` (`:358-361`), so a lost compare-and-set would need a three-way outcome and a routing edit there. Residual risk, read in code and not run: a records-free run classified dead whose owner is restamped after classification and before the reconciler's transaction would be retired under its new owner, and two concurrent reconcilers would append two `run.retired` events. This exists today, is not widened by this SPEC, and is the alternate of Q1.
- For a run classified dead through the lead peer or the boot proof, the operator-path compare-and-set compares the row's `(0, '')` stamp, so a leader that registers as a peer without restamping the row is not caught by the comparison; `retirable` and the in-transaction card guard are the protection there.

### Out of Scope — Existing hint strings that name `--retire`

- `internal/cli/factory.go:500`, `:545`, `:687`, `internal/cli/factory_run_adopt.go:103` and `internal/factory/factory_slots.go:142` tell an operator to retire a run. They stay unchanged (Q5): an operator who follows one for a run that holds cards meets the refusal of REQ-RRG-004, which lists the exits, and nothing is retired.

## H. Verification Gaps (not observed in this plan phase)

- The evidence case (41 cards, run `tmvgvn`) comes from the card body; the production factory database was not read.
- Behavioral RED was observed only through a throwaway probe test (five cases, `acceptance.md` ledger P1-P5: three at the homestate layer, two at the cli layer; deleted, never committed). The tests named in the acceptance criteria do not exist yet; their RED-now cells are empty sweeps (ledger E1, E2). Behavioral REDs of the named tests are recorded by the run phase (`progress.md` §E.2).
- The in-transaction read, the second connection's refusal under `_txlock=immediate`, and the compare-and-set race were established by reading code (§C.4, §C.5) and by probe P3, not by the final tests.
- `moai factory decide ... --choice abandon`, `moai cc -l --factory-run` and `moai cc -f --factory-run` were read, not run.
- The Korean strings are authored but not yet read by a native reviewer beyond the author; the sync-phase humanize pass is the reviewer.
- The docs-site `advanced/factory-mode.md` (en, ko, ja, zh) and the four README files carry the sentences that describe `runs --retire`; line numbers were found by search and the en/ko pages and README.md/README.ko.md were read in part, the ja and zh files only by search output. The sync phase edits all eight.
- Cross-model audit backends were not run in the plan phase; the plan auditor runs them.
