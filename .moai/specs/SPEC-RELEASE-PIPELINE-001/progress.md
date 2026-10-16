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

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
