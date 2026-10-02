# SPEC-TODO-AUTO-PICK-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-10-02T12:22:07Z   # iteration-1 repair; first plan completion was 2026-10-02T11:55:40Z
card: t1448
tier: M
plan_head: b3646de10   # repair base HEAD (the repair commit follows it); first plan was authored at 4bf547bca
plan_audit: pending iteration 2   # iteration 1: FAIL 0.73 against Tier M 0.80 (audited_sha b3646de10; .moai/reports/t1448/plan-audit-iter1.md, not committed); Tier M ceiling 2 iterations
artifact_set: spec.md, plan.md, acceptance.md, research.md, spec-compact.md, progress.md
requirements: 16       # Tier M ceiling 16
acceptance_criteria: 14  # Tier M ceiling 16
needs_clarification: 0
```

Plan-phase notes (what a reader of this record needs, nothing populated for later phases):

- Baselines the criteria measure against were observed in this tree before any run commit and land
  in the plan commits (`verification-claim-integrity.md` §2.3): `acceptance.md` ledger rows
  L1-L22, C1-C4, S1-S2, context rows G1-G2.
- Iteration-1 repair: all eighteen audit items are resolved or deferred-with-reason in
  `plan.md` § Audit-1 resolution map. The two deliberate deltas the leader may veto without ripple
  remain isolated: D-DEF (second clause of REQ-TAU-007) and D-LANE (REQ-TAU-008, now with a defined
  predicate, a dedicated refusal text and a corrected `factory fallback declare` string).
- The compensating-control claim for the decision record is retracted: no party executes a re-read
  of the line today (spec §B.5, §G); AC-TAU-009 is a regression-guard with no executing party.
- Unverified items are listed in `research.md` R10.
- The throwaway probe file `internal/cli/zz_t1448_probe_test.go` was created and deleted during
  research; it never entered a commit. Its four observations are `research.md` R2 (O1-O4). The
  repair added a scratch binary and a scratch queue **outside the tree** (setup rows S1/S2,
  `research.md` R11); nothing of them is committed.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
