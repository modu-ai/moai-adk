# Decision Index — SPEC-FACTORY-LEASE-BUSY-001

Authored at plan close (`interview.decision_gate: on`, `.moai/config/sections/interview.yaml`).
One row per decision surfaced during assembly that no interview settled. Rows route DECIDED /
POLICY-COVERED / EVIDENCE-NEEDED / FOUNDER; authority anchors cite committed artifacts only.

---

### Q1: Which refusal surface do the two non-race claim-failure classes get, and is exit 4 shared?

Label: FOUNDER
Class: product-level
Authority anchor: none verifiable in the committed register — SPEC-FACTORY-ATOMIC-LEASE-001's
HISTORY pins the timing bounds, not the refusal token surface, so the row cannot route DECIDED or
POLICY-COVERED.
Why unresolved: the fix changes the shipped `factory next` refusal line that machine consumers and
operators read today (token `raced`, exit 4, lane-observed 2026-10-21); how far the visible surface
may move is the operator's call, not a derivation.
Default: keep exit code 4 for every refusal class and distinguish record-contention and
claim-deadline classes by reason token only (rule: the option with the smaller user-visible surface
— exit-code-keyed consumers keep working)
Alternate: a distinct exit code for the claim-deadline class
Operator verdict:

### Q2: What latency budget does `factory status` target?

Label: EVIDENCE-NEEDED
Authority anchor: none exists yet — no measurement of the status read path is recorded anywhere in
the committed tree (the card evidence is a lane observation only).
Why unresolved: an absolute budget would be an unobserved claim; plan.md §F M4 produces the
baseline (N ≥ 20, load ≤ core count, p50/p95) and acceptance.md AC-FLB-009 is written
baseline-relative until that data exists.
Operator verdict:

### Q3: Which development methodology does the run phase use?

Label: POLICY-COVERED
Authority anchor: `.moai/config/sections/quality.yaml` § `constitution.development_mode` (value
`tdd`; file verified tracked in this tree via `git ls-files`).
Why unresolved: nothing — settled by the operator setting; recorded so the kickoff review sees the
routing (TDD, RED-GREEN-REFACTOR).
Operator verdict:

### Q4: Is the queue-lock-held `raced` refusal (`factory_card.go` 1527-1528, `TestFactoryLeaseQueueLockStallBounded` and its subtests) reclassified in this SPEC?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the register decides the identical question — the card's defect evidence
names the record-busy path only, and no operator setting or prior SPEC row fixes the queue-lock
site's token.
Why unresolved: the site shares the `raced` token but its detail already names the queue lock;
widening the reclassification to it is a scope judgment beyond the card's stated defect.
Default: no — the queue-lock refusal stays as authored (rule: the option that preserves current
behavior)
Alternate: include the queue-lock site in M2's reclassification
Operator verdict: DEFAULT-APPLIED 2026-10-21T03:30:33Z manager-spec (card t1916, plan close)
