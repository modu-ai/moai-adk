# SPEC-FACTORY-ATOMIC-LEASE-001 — Progress

Card t1458, Tier M. This file is the phase record. Plan-phase writes only §E.1; §E.2 and §E.3 belong to the
run phase (manager-develop) and §E.4 to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-03
- plan_start_head: 2de0a2cb613b04765a1554f86685a3b48e0be806
- artifacts: spec.md (13 requirements), plan.md (6 work milestones), acceptance.md (14 criteria), evidence/
  (eight probes and their overlays, re-executable from the repository root)
- plan_audit: iteration 1 FAIL 0.79 (Tier M threshold 0.80; audited tree `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`;
  no must-pass criterion failed, the score was driven by Testability 0.55 and Clarity 0.70); iteration 2
  not run — to be run by plan-auditor; no verdict is claimed here.

### 2026-10-03 — iteration 2 repair (spec.md 0.2.0)

Each audit defect id, where it was fixed, and how a re-auditor checks it. Measurements made in this
iteration are ledger rows L9–L18 in `acceptance.md`, on HEAD `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`
(Go files equal to the plan-start tree), judged by `moai` build `0732cc699` (L18).

| Defect | Where it was fixed | Check |
|---|---|---|
| D1 | spec REQ-FAL-006 restated as an outcome, mechanism removed from it; plan D2 states both mechanisms and why (L14); spec §A.2 O6 corrected (the iteration-1 probe used a runtime PRAGMA) and O5 cites `factory.go` 473–486 | read REQ-FAL-006 last sentences; plan D2; O6; L14 |
| D2 | acceptance AC-FAL-009 clause (ii): per-path table plus the invariant; checked against `factoryNextClaim`, `factoryNextRecordAndClaim`, `RecordPicked`, `withCardTx`. **One cell differs from the audit's table:** bare arm (b2) makes 0 promotions, not 1, because its queue item is already `picked` | AC-FAL-009 (ii) |
| D3 | spec REQ-FAL-007 names the foreign-worktree directory check as the only outside read; AC-FAL-009 clause (iii) pins the allowed set (the read half stated doctrine-only); plan D4 | REQ-FAL-007; AC-FAL-009 (iii) |
| D4 | acceptance S4 (pass condition) and `-v` on every Command; per-name PASS count N; AC-FAL-010 given a literal selector (L9, 68 names swept, run exit 0); L10 shows the new selectors sweep 0 at the pin | S4; L9; L10 |
| D5 | AC-FAL-002/003 RED cells now cite L11 (clause (i), goroutine form, positive control; `completed inside section = true` at the pin); L1/L3 M2 probes relabeled context; evidence `probe-clause-i_test.go.txt`, `overlay-clause-i.json` | L11; AC-FAL-002/003 |
| D6 | plan §7 MU7 (and MU8–MU11); plan WM1 "cross-process lane helper (what it needs)"; AC-FAL-001 clause (c) and `TestFactoryLeaseSerialCrossProcessExactlyOne`; L12 (10 of 13 breaches from two processes); spec §H DL-3; evidence `probe-xproc_test.go.txt` | AC-FAL-001; plan WM1, §7; L12 |
| D7 | spec REQ-FAL-003 adds "within the queue lock's wait budget" and cites §F R6 | REQ-FAL-003 |
| D8 | spec REQ-FAL-006 bare-form outcome (end the pass at once, error, never an empty-queue report); §H DL-1 for the leader (worst cases: rejected design 5 × 3.3 s = 16.5 s; adopted one budget plus ≤ 50 ms); AC-FAL-007 (c) single outcome with a bound; plan D2; spec §F R13 | §H DL-1; AC-FAL-007 (c) |
| D9 | plan WM1 test list adds `TestHomestateDoesNotImportKanban` (non-test files only; L16: 0 vs 1 with `-test`) and `TestFactoryLeaseSectionRejectsNestedMutate`; AC-FAL-011 reclassified release-blocking until the tests exist (L10); spec REQ-FAL-010 scoped | AC-FAL-011; plan WM1; L16 |
| D10 | plan WM1 now three commits: seam-and-stub (compile-only stubs for every symbol a test names), baseline, RED; the AC-FAL-010 baseline is taken before the RED commit; Definition of Done item 1 | plan WM1; DoD 1 |
| D11 | spec §F R3 corrected (no stale clear on acquisition; the clear is wired to three other locks only, L17), R14 (step-lock crash), R15 (leased without a worktree and its recovery, with the audit's pointer to `factory_lane_relaunch.go` ~110 corrected: that line follows a new lease and is not a recovery); plan D3 failure outcomes; spec §D new Out-of-Scope bullets | §F R3, R14, R15 |
| D12 | AC-FAL-006: deterministic overlap criterion (ordered event log) with the forced-overlap probe (L13: 60 of 60 iterations overlapped); statistics restated as ranges (unforced 5% to 50% over eight runs, forced 55% to 75% over six; "one in ten million" withdrawn); spec §A.1 M5 row | AC-FAL-006; L13; L15 |
| D13 | spec §F R4 second shape (T2 landed, T3 timed out: row `assigned`, item `picked`, verb says `raced`) and third shape (arm (c) has no compensation); REQ-FAL-006 detail wording; AC-FAL-007 (b) mid-claim stall | §F R4; AC-FAL-007 (b) |
| D14 | spec REQ-FAL-004 reworded into the positive form | REQ-FAL-004 |
| D15 | AC-FAL-007 margin stated once (500 ms, a labeled heuristic); `factoryLeaseClaimWaitCap` defined in plan D2 as deadline plus busy timeout and used by AC-FAL-007/-008 | plan D2; AC-FAL-007/-008 |
| D16 | spec O13 before O14, §F R1–R15 in order; O10 replaced by re-measured, attributed counts (L17) | spec §A.2, §F |
| D17 | every `Covers:` line begins with its AC id | acceptance.md |
| D18 | plan D1: the locked handle offers `LoadPure`, no adopting read; spec REQ-FAL-009 names the non-adopting reads; `TestLockedBacklogLoadIsPure` in AC-FAL-011 | plan D1; AC-FAL-011 |

Decisions for the leader are in spec §H (DL-1 bare-form outcome, DL-2 nominated-form outcome, DL-3
cross-process helper). Not observed in this iteration, stated so nobody reads silence as a pass: the
mutants (none can run before implementation); any behavior after the fix; Windows; a multi-lane stall;
whether anything issues the request that applies a lease expiry to a card leased without a worktree.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
