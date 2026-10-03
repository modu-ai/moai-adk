# decision-index.md — SPEC-FACTORY-DECISION-AUTO-001

Decisions surfaced during assembly that the card text does not settle. Rows follow the current
manager-spec contract (Detect → Explain → Ask; no embedded preferred answer). The card t1481 text
and the leader's dispatch are not in the committed authority register, so they anchor nothing here.

### Q1: Should lanes append to the decision board, or only the leader?

Label: FOUNDER
Authority anchor: —
Why unresolved: The directive says "leader-writable, lane-readable", but lanes also produce wait
records, DEFAULT-APPLIED rows, and ceiling records. design.md §1.4 keeps lane writes on the card's
progress record (doctrine §14 already permits the card evidence path). The alternative lets lanes
append a narrow set of kinds scoped to their own leased card, which makes the short recheck
(REQ-FDA-020) read one store instead of two.
Operator verdict:

### Q2: May a FOUNDER row carry a Default marker, given the [HARD] "no embedded preferred answer" clause?

Label: FOUNDER
Authority anchor: —
Why unresolved: `manager-spec.md:110` forbids an embedded recommendation in either recommendation
mode (SPEC-DECISION-AUTHORITY-001). REQ-FDA-018 introduces a Default selected by a stated
reversibility rule. Whether a rule-derived fallback is a recommendation under that clause is a
constitutional reading this SPEC cannot settle for itself.
Operator verdict:

### Q3: What exactly is "product-level"?

Label: FOUNDER
Authority anchor: —
Why unresolved: REQ-FDA-019 blocks Kickoff only on product-level rows. design.md §6 proposes a
four-item definition (user-visible surface outside the SPEC's own, data/state compatibility,
keep-set, pricing/licensing). A wider definition blocks more Kickoffs; a narrower one lets more
defaults apply silently.
Operator verdict:

### Q4: Is the short recheck cadence 5 minutes?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: A fire after the prompt-cache window rewrites the prefix once (cache-aware
execution). A 5-minute cadence sits at the default TTL boundary; whether 4, 5, or 10 minutes
minimizes total cost against leader-reply latency has not been measured.
Operator verdict:

### Q5: Does the bind cache remove the observed "context deadline exceeded" notices?

Label: EVIDENCE-NEEDED
Authority anchor: —
Why unresolved: The benefit is argued from code order (research F6); the miss rate under load and
the share of misses caused by the already-bound path were not measured.
Operator verdict:

### Q6: May a lane invoke `factory decide` at all?

Label: FOUNDER
Authority anchor: —
Why unresolved: REQ-SD-016 (SPEC-FACTORY-SELF-DISPATCH-001) refuses every lane call. REQ-FDA-017
admits exactly one shape (own leased card, kickoff approve, decider `audit`). The alternative keeps
the refusal and has the lane record a board-free "audit-approved" progress line that the leader
session applies — which reintroduces a leader wait.
Operator verdict:

### Q7: Is "hold-and-split" the only second-hit disposition in v1?

Label: FOUNDER
Authority anchor: —
Why unresolved: The leader's observed rulings were "hold or split". The policy enum ships one
value; a second value (e.g. accept as PASS-WITH-DEBT when blocking findings are 0) would let a
second hit proceed without the leader.
Operator verdict:
