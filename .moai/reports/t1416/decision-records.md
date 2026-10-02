# t1416 decision records

decision record: decided_by=lane-6 orchestrator (card t1416, Claude Sonnet 5.5) evidence_refs=.moai/reports/t1416/plan-audit-iter3.md (PASS-WITH-DEBT 0.89 audited at 3d44379b8), .moai/specs/SPEC-CC-ULTRACODE-TOGGLE-001/progress.md §E.1 (plan_status audit-ready), .moai/reports/t1416/plan-audit.md (iter1 FAIL 0.75), .moai/reports/t1416/plan-audit-iter2.md (iter2 FAIL 0.86) ladder_path=gate row "plan→run Kickoff" (.claude/rules/moai/workflow/auto-semantics.md §9.1, AUTONOMOUS form)

## §9.1 conditions, each observed

- Independent plan-audit verdict is not FAIL or INCONCLUSIVE: iteration 3 returned PASS-WITH-DEBT, score 0.89 against the Tier M threshold 0.80, all MP-1..MP-9 pass. Read as a pass with recorded debt (the hard blocks are FAIL and INCONCLUSIVE).
- Plan phase records audit-ready status: progress.md §E.1 `plan_status: audit-ready`.
- Plan-artifact hashes unchanged since the verdict: `git diff --stat 3d44379b8 HEAD -- .moai/specs` printed nothing (audited commit 3d44379b8).
- No blocker open: the iteration 3 defects are D1 (major) and D2 (minor) as pin-coverage gaps plus D3-D8 minor/optional; the auditor offered accepting D1/D2 as debt or adding two pins.
- Keep-set check: no environment-impossible, operator-held or irreversible external-shared operation sits between here and the run phase; integration and push stay later gates.

## Choice taken and why

Accept the debt; do not revise the SPEC again. Iteration 3 is the third of three allowed plan-audits, and a lane does not extend the audit ceiling (memory: lane must not extend audit ceiling). Adding the two pins would change the plan-artifact hash and force a fourth audit.

## Debt carried into run and sync (to be read in the diff, not pinned)

- D1: each edited docs-site workflows row must state that ultracode is a toggle and that it leaves the effort level unchanged (REQ-005); no AC-004 pin checks these.
- D2: each edited docs-site workflows row must keep the same column count (REQ-005); suggested check `^\| \`/effort ultracode\` \|[^|]*\|$`.
- D3: pin W1 returns 0 on correct ko/ja/zh text under `LC_ALL=C` and 1 under UTF-8; run its commands under a UTF-8 locale and say which locale was used.
- D4: spec.md L50 still cites the ledger at `0e7b6af5b` while acceptance.md uses `ff7b64f6e`; HISTORY lists v0.1.2 before v0.1.1. Cosmetic, left as is.
- Known residual from iteration 3: a coupling worded without the token `xhigh` passes the count pins and is left to diff review.

## Explicit wait

wait record: waiting_on=leader (integration window designation and merge path for card t1416) reason=the card was dispatched directly and has no factory record, so `moai factory complete` cannot record or integrate it, and the lane-integration lesson says not to improvise a merge; the evidence packet is `.moai/reports/t1416/verdict.md` recheck=next awaken: run the lane stall watchdog first, then read `moai integration status` and the leader's reply; resume the merge only on a designation, never on silence
