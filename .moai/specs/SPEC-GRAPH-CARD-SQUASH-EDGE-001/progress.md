# Progress — SPEC-GRAPH-CARD-SQUASH-EDGE-001

status: draft
card: t1559 (lane-3, self-dispatch run tmhxo0)
spec: SPEC-GRAPH-CARD-SQUASH-EDGE-001

## Plan-phase record

- 2026-10-07: Tier M artifact set authored (spec.md / plan.md / acceptance.md / this skeleton)
  at plan base df0c417e9. Baseline measurements recorded in spec.md §1 and plan.md §C (HEAD
  single-parent verified via `git rev-list --parents -1`; walk breadth 13127 vs 2664;
  attribution form-2b probe 1 match; `walkCardMerges` count 4; matcher-free layer 0 matches).
- 2026-10-07 (plan-audit r1): **FAIL** — 5 blocking (D1 scope AC impossible, D2 measured-cost
  premise false, D3 sha^1 guard unexercised, D4 plan command setup failure, D5 AC evidence
  form). Verdict `.moai/reports/t1559/plan-audit-verdict.md` (audited_sha 08de1c43c, receipt
  rcpt-e8ccfe80a979d3060d61cf30, score 0.90, iter 1/2).
- 2026-10-07 (D4 lane resolution): the D4 setup failure was the embed-manifest break whose heal
  had already landed on origin/main (f97edcc55 via PR #1762, plus #1783/#1784). Lane merged
  origin/main into the card branch (merge 5fb7baf88); measured post-absorb in this run:
  `go build ./internal/template/` exit 0, `go test -count=1 ./internal/graph/` ok 107.741s.
  No separate repair card, no SPEC scope extension — the auditor's two offered options were
  both moot at decision time.
- 2026-10-07 (repair r1): manager-spec applied 16 fix_scope hunks — D1 pathspec restriction,
  D2 §3 measured rewrite with the LANE DECISION recorded (intermediate-commit attribution
  ACCEPTED), D3 card-attributed root contrast group, D5 -v + rationale correction, base re-pinned
  5fb7baf88, plan §C.1 absorbed-state precondition. Lane spot-verified the hunks on disk
  (stale-text sweep 0 hits; `-- internal/`, `-v`, card-t1561 contrast all present). Delta
  re-audit (iter 2/2, reread_hunks) dispatched by the lane.

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-audit — populated after the plan-auditor verdict lands>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
