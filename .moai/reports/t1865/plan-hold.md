# t1865 — plan hold at the audit ceiling (lane-31)

Disposition: **hold** (`moai spec ceiling SPEC-UPDATE-OWNERSHIP-REPAIR-001 --evidence .moai/reports/t1865`, read-only: "ceiling HIT at 3 rounds — latest verdict FAIL (admitted=false) — disposition hold"; policy `hold-and-split`). No fourth round was started. Kickoff, run and landing did not begin.

## Audit rounds
| Round | Audited SHA | Score | Required codex | Blocking | Verdict file |
|---|---|---|---|---|---|
| 1 | a8ad48c08 | 0.68 | fail (rcpt-8a4bab9510ae7274bc21b2b6) | 5 (D1-D5), MP-8 failed | plan-audit-iter1.md |
| 2 | 3c334dfdd | 0.76 | fail (rcpt-71285364ed5552d8951fefe6) | 5 (N1-N5) | plan-audit-iter2.md |
| 3 | a2a7b8bb2 | 0.84 (threshold 0.85) | fail (rcpt-48fbda4048c92f444d7ba40d) | 2 (N12, N13) | plan-audit-iter3.md |

## State at the hold
- N1-N11: ten closed, N4 partly (the AC is fixed; plan.md M8 still states the whole-tree hash condition).
- Open BLOCKING: N12 (AC-OWR-21 cannot observe atomic creation of `manifest.prior.json`: a direct write to the final name passes all injections) and N13 (stale plan.md lines: M8 step 1 and M9 step 1(a) contradict the rewritten criteria and REQ-OWR-016). Both sit inside the verdict's `fix_scope` anchors.
- Optional, carried as debt: N14-N21 (prior-generation precedence, `DeployedHash` looseness, fragmented-path profile read, tree-hash mtime wording, AC-OWR-20 vs rule 7, no run dir without `.moai/config`, two minor items).
- Q2 = (가) conditional (board d-20261018T040918Z-def2); Q5 conservative reading (leader); OQ-5 narrow carve-out (leader); OQ-6 same-version re-run still skips. None of these is open as a defect.

## Split line (proposal, the owner of the decision is the leader/operator)
The SPEC's own plan names the line M1b + M8 (producing and retaining prior-manifest evidence: `deployed_version`, `manifest.prior.json`, atomic write, backup-failure => zero removals, REQ-OWR-015/-016/-017, AC-OWR-06/-15/-20/-21) versus the rest (ownership verdict, profile-reach, commands, case paths, empty-dir pruning, remaining list, idempotency/recovery). Both open BLOCKING defects (N12, N13) lie on the M1b + M8 side, so the "rest" can be re-audited without them.

## Not done / not asserted
- The lease lapsed at 2026-10-18T04:39:03Z (my renewal fell behind; the renew output said "card lease expired"). It was not recovered because the card is on hold; tree and commits are intact.
- No claim is made that a delta round would pass; rule C allows at most one delta audit limited to one correction commit and only on a ruling.
