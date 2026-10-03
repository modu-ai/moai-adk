# decision-index.md — SPEC-FACTORY-DECISION-AUTO-001

Decisions surfaced during assembly. All seven were decided by the factory leader on 2026-10-03 under
mission contract `07d28c4b` (operator directive the same day: implement now, include in v3.2.0).
The label `LEADER-DECIDED` records that provenance as the leader instructed. It sits outside the
four-label vocabulary the current manager-spec contract allows (`DECIDED`, `POLICY-COVERED`,
`EVIDENCE-NEEDED`, `FOUNDER`) until REQ-FDA-007 lands, and its anchor is the mission contract plus the
leader's ruling, which the committed register does not yet admit. Every row keeps its original
Detect → Explain text.

### Q1: Should lanes append to the decision board, or only the leader?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q1
Why unresolved: The directive said "leader-writable, lane-readable", but lanes also produce wait
records, DEFAULT-APPLIED rows, and ceiling records.
Operator verdict: Board writes are leader-only. Lanes record waits (and DEFAULT-APPLIED and ceiling
records) in their card progress record and read the board. Rationale: the board stays one ruling
source with one writer class; doctrine §14 already accepts the card evidence path for waits.
Reflected in REQ-FDA-003 and design.md §1.4.

### Q2: May a FOUNDER row carry a Default marker, given the [HARD] "no embedded preferred answer" clause?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q2
Why unresolved: `manager-spec.md:110` forbids an embedded recommendation in either recommendation
mode.
Operator verdict: A Default chosen by a fixed, published rule (for example "the option that preserves
current behavior") is a policy application, not an embedded recommendation. The [HARD] clause applies
to judgment calls only, and its text is amended to say so, template source first. Rationale: a rule
anyone can re-apply leaves no judgment embedded in the row. Reflected in REQ-FDA-018, AC-FDA-017, and
plan M5.

### Q3: What exactly is "product-level"?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q3
Why unresolved: REQ-FDA-019 blocks Kickoff only on product-level rows; the boundary was undefined.
Operator verdict: Product-level means a change to a shipped command's default user-visible behavior,
removal of a user-facing feature, or a change to a template default. Everything else is not
product-level. Rationale: the closed three-item list is checkable and covers what users of a shipped
release would notice. Reflected in REQ-FDA-019 and design.md §6.

### Q4: Is the short recheck cadence 5 minutes?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q4
Why unresolved: A fire after the prompt-cache window rewrites the prefix once; total cost against
reply latency has not been measured.
Operator verdict: 5 minutes, as a one-shot cron re-armed for each wait. An EVIDENCE-NEEDED cache and
cost check in run M0 may lengthen it, never below 5. Rationale: a one-shot carrier cannot outlive its
wait, and the measured check decides only whether to slow down. Reflected in REQ-FDA-020, AC-FDA-022,
and plan M0.

### Q5: Does the bind cache remove the observed "context deadline exceeded" notices?

Label: LEADER-DECIDED (EVIDENCE-NEEDED in run M0)
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q5
Why unresolved: The benefit was argued from code order; the miss rate under load was not measured.
Operator verdict: Run M0 measures the degraded-notice rate under load with and without the bind
cache. Acceptance: an already-bound session receives no degraded notice. Rationale: this states the
observable outcome rather than a rate threshold. Reflected in REQ-FDA-021, AC-FDA-020, and plan M0.

### Q6: May a lane invoke `factory decide` at all?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q6
Why unresolved: REQ-SD-016 (SPEC-FACTORY-SELF-DISPATCH-001) refuses every lane call.
Operator verdict: Yes. REQ-SD-016's blanket lane refusal is narrowed to allow exactly
`factory decide <own-card> --decider audit` (kickoff approve). Every other lane decide stays refused.
Rationale: the audit decider rests on verdict-file evidence plus a hash recompute, not on lane
judgment. Reflected in REQ-FDA-017 and AC-FDA-016.

### Q7: Is "hold-and-split" the only second-hit disposition in v1?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q7
Why unresolved: A second policy value would let a second hit proceed without the leader.
Operator verdict: In v1 the second-ceiling disposition is hold plus a split proposal only. Rationale:
this matches the observed rulings (holds on t1356 and one earlier card); card creation stays with the
leader. Reflected in REQ-FDA-012/014.
