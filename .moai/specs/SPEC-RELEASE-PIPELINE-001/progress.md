# progress.md — SPEC-RELEASE-PIPELINE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-16T19:56:40Z
plan_artifacts: spec.md, plan.md, acceptance.md, decision-index.md (Tier M + decision gate on)
spec_id: SPEC-RELEASE-PIPELINE-001
card: t1536
baseline_head: 0bd1da4973d1d9539172f90587ca1ee84c8d9426

Plan-phase evidence: item-5 jq defect reproduced RED at the baseline head
(acceptance.md §B EV-3, exit 5) with fixed-form divergence (EV-4); parity
guard baseline green (EV-1); re-anchor table for all five card items in
plan.md §A.1 (item 1 and half of item 3 already landed on main).

Iter-1 repair (audit verdict FAIL 0.69 → this revision, spec v0.2.0): all
six blocking defects + MP-8 disposed — D1 both required-name SSoT surfaces
(gtmpl + required-checks.yml) named; D2 dry-run dispatch wiring added to
M5; D3 REQ-RPC-012 pins both sides' cancel policy; D4 REQ-RPC-009
redefined + M4 upload extension, AC-009 release-blocking; D5 REQ-RPC-001
fail-open narrowed to the release consumer; D6/MP-8 RED-now cells
EV-7/EV-8/EV-9 measured at c8ede48f90fbfe4c6873f300c91243f03928d254
(grep -c observed 0/0/0, exit 1 each).

Iter-2 repair (audit verdict FAIL 0.69, `.moai/reports/t1536/plan-audit-iter2.md`
→ this revision, spec v0.3.0): D7 (critical) — boolean→"1" conversion specified
at the MOAI_CLEANUP_DRY_RUN wiring point (`${{ inputs.dry_run && '1' || '0' }}`);
TestCleanupDryRunDispatchReachable asserts zero gh release delete under the
wired "true" (mock-gh); §D.2 carries the value contract. D8 — ref-independent
literal concurrency group name on both workflows + group-key-invariance
assertion in TestDrafterCleanupConcurrency; REQ-RPC-012 names ref-independence.
D9 — .github/required-checks.yml restated as the single required-name SSoT;
the gtmpl + template mirror reclassified shape-only PRESERVE (§D-11, M2 steps
1+4, AC-RPC-004). O5 — retry-wrapper anchors corrected to the seven measured
ci.yml sites (:222/:370/:422/:473/:525/:578/:635, verified this run).

Hold-and-split revision (iter-3 FAIL 0.69 at the Tier M ceiling →
operator-approved split → this revision, spec v0.4.0, 2026-10-22):

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-22T12:26:32Z
plan_tree: bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7
scope: M1 (verify-and-preserve) + M2 + M4 + M5 (with D10); M3 + D11 moved to a follow-up card
requirements: 11 live (REQ-RPC-001..004, 007..013); REQ-RPC-005/006 [RETIRED]
acceptance_criteria: 11 live (9 release-blocking, 2 regression-guard); AC-RPC-005/006 [RETIRED]
authorized_audit_rounds: 1 (operator-approved, scope = this revision incl. D10)

Changes: card item 3 removed (REQ/AC-005/006, EV-2, E5, M3; identifiers retired
in place); REQ-RPC-013 / AC-RPC-013 added for D10 (EV-10 RED: `gh release list`
rejects `targetCommitish`, exit 1; EV-12 green-path: `gh release view` supplies
it); every RED-now cell re-measured at plan_tree (acceptance.md §B, document-level
pin); line-citation drift table in plan.md §A.2 (release-pr-multi-os.yml moved
by t1892 sharding; required-checks.yml gained `release_gate` by t1858; ci.yml,
codeql.yml, release-drafter*.yml unchanged). Decision-index: Q4 re-opened and
reversed to retained names (release_gate subset invariant), Q7 (fetch source)
and Q8 (event-class gating scope) added and default-applied. AC-RPC-004
reclassified regression-guard (no RED-now cell exists on the real tree).

## Binding run conditions

Copied from the PASS-WITH-DEBT verdict `.moai/reports/t1536/plan-audit-iter7.md` (audited b0f6783e434f, score 0.93, codex receipt rcpt-8e4e368d127d6d4389f9b43f); both sync verdict owners re-read these, and an undisposed condition caps the sync verdict at FAIL. Kickoff 2026-10-23: autonomous audit-decider approve (`moai factory decide t1536 --gate kickoff --choice approve --decider audit`, card v12), under the operator ceiling exception recorded in §G.

- debt: N1 dispose_in=run REQ-RPC-002 / AC-RPC-002 fail-open text lacks the "run not itself cancelled" qualifier stated in REQ-RPC-003
- debt: N3 dispose_in=run AC-RPC-003 cell (i) (real failure, marker skipped) pair relies on the landed same-name resolution; state it in (i)
- debt: N4 dispose_in=run acceptance.md §D.0 "verdict step without fail-open" probe cites AC-RPC-004 as a red source that no longer reads red

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~10-14 files (`.github/workflows/*.yml`, `.github/required-checks.yml` consumers, Go guard tests under `internal/template/`, `internal/template/testdata/`); domains 2 (CI workflows, Go guard tests); file mix YAML + Go + jq; concurrency benefit LOW (coding-heavy, M2 -> M5 -> M4 -> M1 ordered, M4 depends on M2's job names).
- Mode evaluation: direct not selected (semantic change); serial selected; fanout not selected (coding-heavy, <3 domains); sweep not selected (not mechanical-uniform); agent-team not requested.
- Decision: serial
- Justification: coding-heavy work with ordered milestones stays sequential per Anthropic's coding-task parallelism caveat; one manager-develop (cycle_type=tdd, effort high for Class C) carries M2, M5, M4, M1 in plan order, one writer in this tree.
- Kickoff: autonomous audit-decider approve 2026-10-23 (card v12) on the PASS-WITH-DEBT verdict plan-audit-iter7.md, under the operator ceiling exception recorded in §G; Binding run conditions above.

## §G Override and Refusal Record

- 2026-10-22T18:41:00Z SPEC-RELEASE-PIPELINE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md
- 2026-10-23 SPEC-RELEASE-PIPELINE-001 operator-decision (relayed by the leader session 3936623d to lane-9 on 2026-10-23, recorded verbatim at the leader's request; the lane did not observe the operator directly): "운영자 결정 2026-10-23: plan-audit iter6 PASS-WITH-DEBT 0.93(rcpt-78c759130e822df7b345659e)을 수용, REQ-ACE-006 천장(6회>=2) 예외 승인, 리더 세션 3936623d 경유."
- 2026-10-22T19:20:15Z SPEC-RELEASE-PIPELINE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md
- 2026-10-23 SPEC-RELEASE-PIPELINE-001 round-7 note (instruction relayed by the leader session, which relayed the operator decision above): 7th round = hash-binding re-audit to enforce the operator exception, scope limited to the relabel delta 71b227b088..8e097c9d84 with a regression glance only; the verdict must record the engine plan-artifact hash.
- 2026-10-22T19:32:29Z SPEC-RELEASE-PIPELINE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter7.md
- 2026-10-22T19:35:50Z SPEC-RELEASE-PIPELINE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter7.md
- 2026-10-22T19:36:02Z SPEC-RELEASE-PIPELINE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter7.md
