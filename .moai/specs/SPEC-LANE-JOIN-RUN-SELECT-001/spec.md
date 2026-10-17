---
id: SPEC-LANE-JOIN-RUN-SELECT-001
title: "Lane-join run selection — an interactive chooser when two or more active factory runs survive, KST times, and a fail-closed non-interactive path"
version: "0.1.0"
status: completed
created: 2026-10-17
updated: 2026-10-18
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/homestate, internal/factorymsg"
lifecycle: spec-anchored
tier: M
tags: "factory, lane-join, run-selection, chooser, kst, card-t1853"
card: t1853
related_specs: [SPEC-FACTORY-RUN-RETIRE-001, SPEC-FACTORY-STALE-RUN-HEAL-001, SPEC-FACTORY-LANE-JOIN-SOCKET-001]
---

# SPEC-LANE-JOIN-RUN-SELECT-001 — Lane-join run selection

## A. History

| Date | Version | Change |
|---|---|---|
| 2026-10-17 | 0.1.0 | Initial plan-phase draft (card t1853, Tier M, Class C). The A-part (join-time selection) of the split of card t1848; the C-part (the retirement guard) is card t1852, SPEC-FACTORY-RETIRE-GUARD-001, and is not covered here. The predecessor SPEC-FACTORY-RUN-SELECT-CORE-001 (card t1848; plan audits FAIL 0.81, 0.89, 0.90) lives in a sibling card tree and is not committed to this tree; its closed defects D1-D5 and N1 are carried as written (§F) and its one open defect N2 is resolved here (REQ-LJR-005, REQ-LJR-006, AC-LJR-007). Tree `fb749d372e61a6b10c53091362c76b4412827427`; every code fact in §C was re-read at that tree. |

## B. Problem Statement

Observed 2026-10-17 (evidence case taken from the card body; the production factory database was not re-read for this SPEC, see §H): run `tmvgvn` was active with a dead leader and 41 in-progress cards; a second run `tn1p4e` was minted by a later leader start; `moai cc -l` then failed with `AMBIGUOUS_FACTORY`, and the operator had no way to choose a run from a list.

What the lane-join gate does at this tree (§C): with exactly one active run it joins that run implicitly; with two or more survivors it refuses with a text-only `AMBIGUOUS_FACTORY` message that prints each run's start time as stored (UTC). This SPEC changes only the two-or-more case: on a terminal it replaces the refusal with an interactive chooser that shows the facts distinguishing each run and lets the operator join one, clean one up and join the remainder, or cancel; off a terminal it keeps the refusal and renders the times in KST. The single-run implicit join is unchanged.

## C. Root Cause (code facts re-read at tree `fb749d372e61`; line numbers are coordinates at that tree and drift, the names are the contract)

1. **The shared gate.** `enterFactoryLaneRun` (`internal/cli/factory.go:597`) is the single lane-join gate. It calls `enterSelectedFactoryRun(root, explicit, true, timing)` (`:468`); on an `*AmbiguousRunsError` it wraps the error with `ambiguousJoinGuidance` (`:530`); discovery answers only record-absence or a named leader (`:613`). Its non-test callers are `cc.go:240`, `glm.go:299`, `codex_factory.go:103` and `codex_launcher.go:959` (`enterCodexRelaunchJoin`, the relaunch loop's join step).
2. **`enterSelectedFactoryRun` joins by side effect.** After a successful `ResolveActiveRun` it captures the environment (`:502`) and runs `os.Setenv(config.EnvFactoryRunID, runID)` (`:503`); it returns only a restore function and an error, never the run id. Any call that succeeds has already joined.
3. **Resolution.** `ResolveActiveRun` (`internal/factorymsg/store.go:343`): one active row returns its id with no reconciliation (`:372-377`); two or more run `ReconcileActiveRuns` (`:383`), which retires records-free dead runs (a write), and return `*AmbiguousRunsError{Owners}` when two or more survive (`:393`). `describeRunOwners` (`:415`) prints `started %s` with `RunOwner.CreatedAt` as stored (`:423`); `RecordRun` writes `time.Now().UTC().Format(time.RFC3339Nano)`, whose fractional digits vary in length.
4. **A read-only classification exists.** `(*FactoryDB).ClassifyRuns` (`internal/homestate/factory_run_retire.go:168`) delegates to `classifyRuns` (`:172`), whose body issues SELECTs, the identity probes and the boot-proof queries and no write. `factorymsg.ReconcileOptionsFor` (`internal/factorymsg/factory_run_retire.go:136`) is the production option set. `RunOwner` (`:53`) carries run id, status, lead pid, `CreatedAt`, `UpdatedAt`, classification and basis; it carries no card count and no activity time.
5. **Retirement.** `RetireRunIfDead` (`factory_run_retire.go:381`) classifies, refuses a non-`retirable` classification with `ErrRunOwnerNotDead` (`retirable`, `:165`, is the single gate), and calls `retireRun` (`:407`). Its in-progress-card guard and compare-and-set are SPEC-FACTORY-RETIRE-GUARD-001 and are not changed here.
6. **Terminal card states are one shared set.** `terminalStates = {done, failed, abandoned}` (`internal/homestate/card_record.go:74`); `sqlTerminalStates` (`factory_run_adopt.go:21`) spells it for SQL. Every other card state is non-terminal.
7. **Seams that exist.** `isInteractiveStdin` (`internal/cli/init_update_notice.go:69`) is an assignable function variable; `writerIsTerminal(w io.Writer)` (`internal/cli/glamour_style.go:137`) reports a terminal-backed file; `leaderNamesByPID` (`factory.go:551`) maps a lead pid to its registered leader name; `TestMain` (`internal/cli/main_test.go:372`) already neutralizes ambient factory state for every test. Test fixtures `seedLiveRun` (`factory_ambiguous_join_test.go`), `seedDeadRunWithLiveRecords` (`factory_run_adopt_test.go:46`), `holdLeaderName`, `discoveryTestRoot` and `clearFactoryTestEnv` exist.

## D. Requirements (GEARS)

Terms. **Non-terminal card**: a card whose state is outside `{done, failed, abandoned}` (§C.6); used by every requirement. **Survivor**: an active run that remains active after the automatic reconciliation. **Interactive**: stdin is a terminal (the `isInteractiveStdin` seam) and the prompt output stream is a terminal-backed file. **Read path**: a read of the factory store that writes no `runs` row and no `events` row and sets no environment variable.

- **REQ-LJR-001** (Ubiquitous): The factory state layer shall expose, through a read path, one fact set per active run — run id, owner classification, lead pid, start time, last lane activity time and the count of non-terminal cards. Last lane activity is the latest, by parsed time value (never by string order, because stored times are RFC 3339 with a variable-length fraction), of the run's worker heartbeat times, card heartbeat times and card update times, skipping empty and unparseable values; a run with none reports no activity. The fact set shall list only runs whose status is `active`.
- **REQ-LJR-002** (Ubiquitous): Every run start time and last-activity time that the lane chooser and the ambiguity refusal text render shall be rendered in KST (UTC+9) as `YYYY-MM-DD HH:MM:SS KST`; a value that does not parse as an RFC 3339 time shall be rendered unchanged.
- **REQ-LJR-003** (Event-driven): When a lane join (`moai cc -l`, `moai glm -l`, `moai codex -l`) is refused because two or more active runs survive reconciliation, no `--factory-run` was given, no `--leader` was named and the session is interactive, the launcher shall present each survivor with its run id, owner liveness (`live`, `dead` or `indeterminate`), leader name (`-` when the leader-name registry maps none), start time, last lane activity time (`-` when none) and non-terminal card count, in Korean prompt text; shall offer a `join` choice for every listed run; shall offer `cleanup then join` only for a run whose owner is classified dead and whose non-terminal card count is zero; shall always offer `cancel`; and shall carry no preselected or default choice (an empty line is cancel).
- **REQ-LJR-004** (Event-driven): When the operator chooses `join` for a run, the launcher shall join exactly that run as if `--factory-run <run-id>` had been given, and shall set the factory-run environment variable only then. When the operator chooses `cancel`, enters nothing, or input ends before a valid choice, the launcher shall exit non-zero with the `FACTORY_JOIN_CANCELLED` error, shall not set the variable, and the chooser's own cancel path shall perform no retirement and no further write; changes made before the chooser opened persist — the automatic reconciliation that precedes every multi-run resolution may already have retired records-free dead runs, and an operator may cancel after a completed chooser retirement.
- **REQ-LJR-005** (Event-driven): When the operator chooses `cleanup then join` for a run, the launcher shall retire that run through `RetireRunIfDead` and then re-read the active runs through the read path of REQ-LJR-001, and shall never re-read through a call that sets the factory-run environment variable. When the retirement succeeded and exactly one active run remains, the launcher shall join that run (the join half of the operator's choice); when two or more remain, it shall present again from the re-read facts, so that a run that has meanwhile become live shows `live`, a run retired meanwhile is absent, and eligibility is recomputed from the re-read facts; when none remain, it shall fail with `NO_ACTIVE_FACTORY`. The chooser's call of `RetireRunIfDead` shall be the only retirement call this SPEC adds, and no selection, flag or prompt this SPEC adds shall retire, or offer to retire, a run whose owner is not classified dead; `retirable` shall remain the only retirement predicate.
- **REQ-LJR-006** (Event-driven): When a retirement requested under REQ-LJR-005 is refused — the run gained a non-terminal card, its owner became live, or its owner or status changed — the launcher shall print the refusal and present again from facts re-read through the read path, even when only one active run remains; for that single run it shall offer `join` and `cancel` only, shall never join it automatically, and shall set the factory-run environment variable only after the operator chooses `join`.
- **REQ-LJR-007** (State-driven): While the session is not interactive, the lane join gate shall keep its existing fail-closed refusal — the `AMBIGUOUS_FACTORY` prefix, every candidate and the resolution commands — shall invoke no chooser, shall select no run, and shall render each start time in KST; an explicit `--factory-run` and a named `--leader` shall keep their current behavior on both the interactive and the non-interactive path.

## E. Decision Points

Full rows with routing labels are in `decision-index.md`. **Dispatch-fixed scope** (not decision rows): the A-part only; an interactive-only chooser with a fail-closed non-interactive path; no new cleanup power over a run whose leader is not dead; and the N2 resolution shape (re-query through a read path, `enterSelectedFactoryRun` only after an operator `join`).

| DP | Decision | Disposition |
|---|---|---|
| Q1 | Whether a re-presentation re-runs the automatic reconciler (which can retire further records-free dead runs mid-prompt) or only reads. | **Default applied: read only** — the chooser performs no write the operator did not choose (REQ-LJR-005, AC-LJR-007 sub-scenario e). |

## F. Prior audit history (closed — carried, not re-argued; N2 resolved here)

The predecessor was audited three times (FAIL 0.81 -> 0.89 -> 0.90) and hit the Tier M ceiling; the reports are in the sibling tree `.moai/worktrees/t1848/.moai/reports/t1848/` (`plan-audit.md`, `plan-audit-iter2.md`, `plan-audit-iter3.md`), read-only references, not authority here.

| Closed defect (source) | Decision carried | Where it lands |
|---|---|---|
| D1: guard versus a retired run holding cards | Belongs to the retirement guard, not to join-time selection | SPEC-FACTORY-RETIRE-GUARD-001 REQ-RRG-003; out of scope here (§G) |
| D2: cancel "writes nothing" was false | Cancel promises no retirement and no further write by the chooser's own path; earlier changes persist | REQ-LJR-004, AC-LJR-005, AC-LJR-006 |
| D3: renderer assertion could pass on a substring | A cell-level row seam; the test asserts each column cell by equality | REQ-LJR-003, AC-LJR-003 |
| D4: "fresh facts" had no source | Every presentation reads the store; none reuses a prior resolution's owners | REQ-LJR-005, REQ-LJR-006, AC-LJR-006, AC-LJR-007 |
| D5: criterion green on the unmodified tree | Such a criterion is labeled a regression guard, never release-blocking | AC-LJR-008 (guard half) |
| N1: a refused retirement leaving one survivor auto-joined it | Present again, join/cancel only, never auto-join | REQ-LJR-006, AC-LJR-007 |
| **N2 (open at the predecessor's last audit)**: its re-resolve call `enterSelectedFactoryRun(root, "", true, ...)` returns the survivor and has already run `os.Setenv` (§C.2), so "present again, no automatic join" was unsatisfiable, and it named no source for the one-row facts | The loop never re-resolves through the env-setting entry. It re-reads through the read path; `enterSelectedFactoryRun` is called only with an explicit run id, only after an operator `join` or the join half of a successful `cleanup then join` | REQ-LJR-005, REQ-LJR-006; AC-LJR-007 pins it with the environment observed at every presentation entry |
| O8-O10 (optional) | Loop-bound sentence corrected; a dedicated one-row header and legend string; the one-row table never offers cleanup | plan §C.4, §E |

## G. Out of Scope

### Out of Scope — Retirement guard and compare-and-set (card part C)

- The in-progress-card refusal of `RetireRunIfDead`, the owner-stamp compare-and-set, the `runs --retire` refusal and stranded-cards blocks, and the `relaunch --from-run` outcome text. Moved to card t1852 (SPEC-FACTORY-RETIRE-GUARD-001). This SPEC calls `RetireRunIfDead` unchanged and treats any non-nil error it returns as a refusal; no requirement here depends on that SPEC landing first (plan §J).

### Out of Scope — Leader-start adopt / new run / cancel prompt (card part B)

- The `moai cc -f` / `moai glm -f` prompt for a dead-owner run that still carries cards. Moved to card t1849.

### Out of Scope — Cross-run card-record transfer (card part D)

- A command or function that moves card records, leases or approvals from one run to another. Moved to card t1842.

### Out of Scope — Automatic selection and the single-run join

- Any automatic choice among surviving runs; fail-closed selection among live runs stays. The implicit join when exactly one run is active, the verified-leader discovery fallback and the named `--leader` path keep their behavior.
- Any prompt for a non-interactive session; REQ-LJR-007 keeps the refusal.

### Out of Scope — Lane behavior after joining

- Dispatch, lease, stage or complete behavior of a lane that has joined a run; the stale-run hook notices (SPEC-FACTORY-STALE-RUN-HEAL-001); worker rows left behind by a retired run; any change to the owner classifier, the boot proof or `ReconcileActiveRuns` semantics.

## H. Verification Gaps (not observed in this plan phase)

- The evidence case (run `tmvgvn`, 41 cards) comes from the card body; the production factory database was not read.
- No behavioral RED was observed. Every RED-now cell in `acceptance.md` is "the named test does not exist", measured by an empty sweep (ledger E1-E3); behavioral REDs are recorded by the run phase (`progress.md` §E.2).
- The claim that the read path writes no `runs` or `events` row rests on reading `classifyRuns` and `predatesBoot` (SELECTs only); the legacy-peer fallback of `ReconcileOptionsFor` opens an existing broker database through `OpenExistingWithDeadline`, and whether that open runs idempotent schema statements was not traced. The claim is scoped to the `runs` and `events` tables for that reason.
- Whether a dead leader's name survives in the leader-name registry was not observed; the name column may render `-` for dead leaders.
- The default interactive gate on a real terminal and the `cc.go` / `glm.go` call sites are not driven by any test (the gate function and `enterCodexRelaunchJoin` are); the call-site inventory is a grep.
- The docs-site `advanced/factory-mode.md` pages (en, ko, ja, zh) carry the ambiguity-refusal sentence at line 67 of each (found by search at this tree); their content was not read in full and the sync phase edits all four.
- `moai spec lint` ran under a binary built from this tree (see `progress.md` §E.1); the installed build is a strict ancestor of HEAD and was not used for the cited result.
