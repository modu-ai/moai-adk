# progress.md — SPEC-FACTORY-MANAGED-TUI-001

> Phase record. Only `§E.1` is filled by the plan phase (manager-spec); `§E.2` and `§E.3` belong to the run phase (manager-develop), `§E.4` to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: revised-for-delta-audit (0.3.0; iteration 3, the leader-granted final delta audit scoped to D1-D5 and D7)
plan_complete_at: 2026-10-03
artifacts: spec.md (REQ 14) · plan.md · acceptance.md (AC 16) · design.md (D-1..D-10) · progress.md
tier: M (see plan.md §A; LOC total at the ceiling, tier-up path stated)
measured_tree: 2b9e4a4d0
open_clarifications: 0 NEEDS-CLARIFICATION markers; 6 leader/operator questions in the plan-phase report
card: t1408

### Ownership boundaries (leader constraint, t1440 lane-17)

Edits only in `managed_codex_factory.go`, `managed_factory_session.go`, new `managed_codex_tui.go` / `managed_codex_tui_test.go`, the fixture `testdata/codex-0.160.0/resume-help.txt`, `internal/config/defaults.go`, `envkeys.go`, one minimal separate bullet edit in the operator document, and CHANGELOG. Not touched: `codex_launcher.go`, `managed_operator_input.go`, `managed_card_child_test.go`, `store.go`.

### Plan-audit iteration 1 and its disposition (iteration 2 revision, version 0.2.0)

Iteration 1: FAIL 0.63 at `7287e64f3` (local report `.moai/reports/t1408/plan-audit-iter1.md`, not committed).

| Defect | Disposition |
|---|---|
| D1 REQ-MT-008 vs AC-MT-007(iii) | Fixed: REQ-MT-008 now carries the outstanding-`turn/start` exception; AC-MT-007 (iii) and known debt 9 use the same wording. |
| D2 false no-overlap with t1459 | Fixed: design.md D-9 rewritten from the t1459 draft (leading `ctx`, `ctx.Done()` in the select, interruption check, 14 occurrences in 6 files here vs 10 in 5 there); what each landing order owes is listed; "either order works" replaced by verified vs unverified. spec.md §D reworded. |
| D3 M1 not compile-able RED | Fixed: plan.md M1 is stub commit S plus RED commit R; RED = named failing assertion per AC; REQ-MT-014 and AC-MT-014 amended. |
| D4 M2 gate needs M4 | Fixed: minimal TUI wait/stop path moved into M2; M4 keeps status mapping, server-death monitor, blocked-call release; `Exit:` lines added for every milestone. |
| D5 AC-MT-010 nonexistent test | Fixed: AC-MT-010 is the four existing tests only; `TestManagedTUINeverReachedWithoutOptIn` is the new AC-MT-016. |
| D6 missing RED-now cells | Reclassified, not faked: test-based ACs are adoption-deferred; M2 cannot start until a delta check confirms every cell in `red-baseline.md` (acceptance.md header and §1.2). |
| D7 EXCL-syscall | Fixed: clause added to spec.md C.4, §D, §E and the pseudo-terminal exclusion (wording only). |
| D8 log-sink lifetime | Fixed: sink cleared at TUI reap; teardown lines go to the terminal; REQ-MT-006, D-6, D-8, AC-MT-005 (`post_reap_line_on_terminal`) agree. |
| D9 stderr after TUI start failure | Fixed: REQ-MT-004 exception; D-6; AC-MT-003 `tui_start_fails` checks it. |
| D10 stale-busy steering | Fixed by changing the rule: busy never clears by time, warn lines only; known debt 13; AC-MT-006 `TestManagedCodexBusyLongTurnKeepsDeferring`; mutant mu16. |
| D11 timing seams | Fixed: D-10 declares three overridable durations; D-7 amended. |
| D12 AC-MT-013 | Fixed: base counts for every row, same-line anchors, CHANGELOG rows, bare `t1459` row replaced, mutant mu17. |
| D13 "resolved" | Fixed: REQ-MT-013 and the DoD make "resolved" conditional on AC-MT-015 steps 1, 2, 3, 6. |
| D14 blocked `DeliverTurn` | Fixed: AC-MT-008 subtests `blocked_turn_start_released`, `blocked_wait_turn_released`; mutant mu15. |
| D15 §F vs D-9 | Fixed: §F says one line in `Start`; D-9 table gains the TUI-exit channel and the monitor goroutine rows. |
| D16 P10 undercount | Fixed: P10 names the three Codex writers and the out-of-scope `:179`. |
| D17 bundled REQ-MT-007, cross-refs | Cross-references fixed (D-8, §A.1); REQ-MT-007 not split (REQ count and AC traceability stay stable; optional). |
| D18 unnamed subtests, P6-false, AC-MT-014 placeholders | Fixed: AC-MT-003 subtests named; `frames_not_delivered` added; AC-MT-014 states one row per fix commit. |
| D19 ownership list drift | Fixed: CHANGELOG and the help-text fixture in plan.md and here. |

### Plan-audit iteration 2 (FAIL 0.86) and the delta revision 0.3.0 (iteration 3 = final delta audit)

Local report `.moai/reports/t1408/plan-audit-iter2.md` (not committed). Leader decisions baked in: AC-MT-015 is operator-held (keep-set), not release-blocking, recorded as "operator confirmation pending, no terminal designated" in the sync-phase deliverable; t1408 lands before t1459 (decided, no hedging); the opt-out `MOAI_FACTORY_MANAGED_TUI` stays, default ON inside the managed gate.

| Defect | Disposition |
|---|---|
| D1 blocked write not releasable by a channel | Fixed: design.md D-8 "Releasing a blocked call" — the wait goroutine closes the WebSocket connection on TUI exit; a per-write deadline is the secondary device; `Close` stays single-goroutine; AC-MT-008 `blocked_turn_start_released` blocks the write and proves it. D-9 row and plan.md M4 follow. |
| D2 CN-4 CONFLICT from the H1 | Fixed: the AC range is removed from the acceptance.md H1. Exit/AC ordering re-read by hand: M1 exit AC-MT-014 (M1 part), M2 001-005 and 016, M3 006-007, M4 008-009, M5 010-012 (and the closing part of 014), M6 013; no acceptance clause orders a criterion across those exits. |
| D3 M1 gate unmechanized | Fixed: the orchestrator or the M1 lane runs the stated three-line command (acceptance.md §1.1) and records it in §E.2; AC-MT-014 split into an M1-evaluable part and a closing part re-run at M5. |
| D4 AC-MT-014 claims more than it checks | Fixed: 11 rows, 11 or more `--- FAIL: Test` lines, zero `undefined:`/`build failed`; "tree SHA" is S's commit SHA with the R tests applied. |
| D5 E1/E4 flips | Fixed: E1 gets `--exclude='*_test.go'`; E4 is noted to flip at S. |
| D7 plan.md §F | Fixed: "AC-MT-001..014 and 016 (015 operator-held)". |
| D6 and other optional | Not touched (D6 residual stays a Gap: the sync audit reads the anchored lines). |

## §E.2 Run-phase Evidence

_pending run phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync phase_
