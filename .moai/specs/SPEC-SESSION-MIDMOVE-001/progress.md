# Progress — SPEC-SESSION-MIDMOVE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: v0.4.0 delta revision after plan-audit iter-3 (FAIL 0.83, `.moai/reports/t1279/plan-audit-iter3.md`). ND1–ND11 addressed under the one-time extension (`verdict.md` §⑦); the next audit is a delta audit over `b284d634a..<this commit>`.
- plan_time_ac_ledger_v040: research.md §R11 addendum (HEAD `bce7248ff`)
- tier: L; artifacts spec.md, plan.md, acceptance.md, design.md, research.md, progress.md; PASS threshold 0.85
- requirements: 20 (REQ-SMM-001..020); acceptance criteria: 22 (AC-SMM-001..022)
- decision_points: DP-1 hook (recommended docs-only) and DP-2 integration-window moves (recommended exempt) open for kickoff; N1 `/clear` placement and DP-3 ordering settled by the lead
- plan_time_ac_ledger: research.md §R11
- self_check: SPEC ID regex PASS; no open clarification markers
- source_of_truth: `.moai/reports/t1279/verdict.md` §①–§⑤

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## HOLD (2026-09-27, lead decision b1)

Plan closed as draft + HOLD. Do NOT enter run on this SPEC as written.

Reasons (.moai/reports/t1279/plan-audit-delta.md, auditor-model claude-opus-5-5[1m], FAIL 0.83):
- D2: the P3 continuity measurement design cannot hold. A real /clear always opens a new transcript file (369/369 observed), so "same transcript contains the /clear row" is satisfied only by compact rows.
- D1: regression. AC-002 rejects gaps that REQ-005 permits.
- D3 (ND8): forbidden probe flags pass AC-007 via brace expansion or `-p <flag>`.
- D4: AC-022 does not check that the develop merge commit is an ancestor.

State at hold: t1175 has landed on develop (merge 7fe658815). This branch did not absorb it at hold time because develop CI was red (t1175 regression under repair). It was absorbed afterwards, together with the repair, via `git merge develop` at develop 80e0fdc0e in the integration window (`git merge-base --is-ancestor 7fe658815 HEAD` exit 0). The HOLD itself is unchanged.
