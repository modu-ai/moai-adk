# decision-index.md — SPEC-FACTORY-DECISION-AUTO-001

Decisions surfaced during assembly and plan audit. All were decided by the factory leader on
2026-10-03 under mission contract `07d28c4b` (operator directive the same day: implement now,
include in v3.2.0). The label `LEADER-DECIDED` records that provenance, as the leader instructed. It
is outside the four-label vocabulary the current manager-spec contract allows (`DECIDED`,
`POLICY-COVERED`, `EVIDENCE-NEEDED`, `FOUNDER`) until REQ-FDA-005 lands. Its anchor is the mission
contract plus the leader's ruling, which the committed register does not yet admit. Each row keeps
its original Detect → Explain text. REQ numbers below follow spec.md v0.3.0. Where a later row
narrows an earlier one, the later row governs (Q19 narrows Q10; Q20/Q21 make Q16 mechanical).

### Q1: Should lanes append to the decision board, or only the leader?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q1
Why unresolved: The directive said "leader-writable, lane-readable", but lanes also produce wait
records, DEFAULT-APPLIED rows, and ceiling records.
Operator verdict: Board writes are leader-only. Lanes record waits in their card progress record and
read the board. Rationale: one ruling source with one writer class. Reflected in REQ-FDA-001 and
design.md §1.4.

### Q2: May a FOUNDER row carry a Default marker, given the [HARD] "no embedded preferred answer" clause?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q2
Why unresolved: `manager-spec.md:110` forbids an embedded recommendation in either recommendation
mode.
Operator verdict: A Default chosen by a fixed, published rule is a policy application, not an
embedded recommendation. The [HARD] clause applies to judgment calls only, and its text is amended
template-first. Rationale: a rule anyone can re-apply embeds no judgment. Reflected in REQ-FDA-017.

### Q3: What exactly is "product-level"?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q3
Why unresolved: The boundary that blocks Kickoff was undefined.
Operator verdict: A change to a shipped command's default user-visible behavior, removal of a
user-facing feature, or a change to a template default; everything else is not product-level.
Rationale: a closed, checkable list. Reflected in REQ-FDA-018.

### Q4: Is the short recheck cadence 5 minutes?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q4
Why unresolved: Cache cost against reply latency was unmeasured.
Operator verdict: 5 minutes, as a one-shot cron re-armed per wait. A run-M0 cache and cost check may
lengthen it, never below 5. Reflected in REQ-FDA-019 and plan M0(b).

### Q5: Does the bind cache remove the observed "context deadline exceeded" notices?

Label: LEADER-DECIDED (EVIDENCE-NEEDED in run M0)
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q5
Why unresolved: The benefit was argued from code order.
Operator verdict: Run M0 measures the degraded-notice rate under load with and without the cache.
Acceptance: an already-bound session receives no degraded notice. Reflected in REQ-FDA-020 and
plan M0(a).

### Q6: May a lane invoke `factory decide` at all?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q6
Why unresolved: REQ-SD-016 refuses every lane call.
Operator verdict: Yes, exactly `factory decide <own-card> --decider audit` (kickoff approve). Every
other lane decide stays refused. Reflected in REQ-FDA-016.

### Q7: Is "hold-and-split" the only final-hit disposition in v1?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 Q7
Why unresolved: A second policy value would let a final hit proceed without the leader.
Operator verdict: Yes, hold plus split proposal only (the single narrow exception is Q16).
Reflected in REQ-FDA-010/012.

### Q8: Does the shared predicate apply to sync verdicts too? (iter1 D2)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D2
Why unresolved: T7 and T13 share `guardVerdictPass`; plan thresholds and the debts schema would alter
sync verdicts.
Operator verdict: The predicate takes the phase. Plan threshold and debts schema apply to plan-audit
verdicts (T7, Kickoff). Sync thresholds apply to sync verdicts (T13). Sync PASS-WITH-DEBT semantics
are unchanged except Q14. Reflected in REQ-FDA-009.

### Q9: Which fields does the predicate check? (iter1 D3)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D3
Why unresolved: "must-pass all" had no machine field; the code sites checked different subsets.
Operator verdict: Add a must-pass field to the verdict block. Score-below-threshold and
must-pass-failed fixtures are both refused. The card-transition guard checks the same fields as
`decide.go`. Reflected in REQ-FDA-006/009 and AC-FDA-009.

### Q10: What else must the `audit` decider re-check? (iter1 D4, D11)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D4
Why unresolved: The decider checked only the predicate and the hash, so a lane could pass an open
product-level verdict.
Operator verdict: Additionally require audit-ready status, no open blocker or operator hold, and no
open product-level FOUNDER row. Any of these present means refuse and leave the human Kickoff path.
Keep-set stays human. The `audited_sha` binding (D11) is named too. Reflected in REQ-FDA-014.

### Q11: May a cache hit skip the run-state probe? (iter1 D5)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D5
Why unresolved: Skipping the probe would serve a stale binding after run retirement.
Operator verdict: Never. A hit skips only `factorymsg.Open` and the peer query, and only when the
probe reports the same live run. A retired or changed run invalidates the cache. New REQ + AC: retire
the run, then the next prompt rebinds. Reflected in REQ-FDA-020/021.

### Q12: How is delta eligibility decided? (iter1 D6)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D6
Why unresolved: "confined to the required-fix text" was a judgment, and the check compared past
state.
Operator verdict: Mechanically. The verdict carries `fix_scope` (file plus anchor list). The delta is
eligible only if the diff between the two audited SHAs touches nothing outside `fix_scope` (plus
progress, decision-index, and reports) and the REQ/AC id sets are identical. Reflected in
REQ-FDA-011.

### Q13: Who applies the ceiling policy, and what is the "second hit"? (iter1 D7)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D7
Why unresolved: Only lane behavior was specified, and "second time" was ambiguous for rounds > 1.
Operator verdict: The policy applies to every session. A non-lane orchestrator applies the auto
delta itself and, on the final hit, writes the hold record and informs the user (the question channel
only to offer an override). The final hit is the iteration reaching ceiling + `auto_delta_rounds`;
default `auto_delta_rounds` = 1. Reflected in REQ-FDA-010/012.

### Q14: Who re-reads binding run conditions, and do they block? (iter1 D8)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D8
Why unresolved: Only the fallback sync-auditor re-read them, and blocking was unstated.
Operator verdict: Both `sync-audit-4dim` and sync-auditor re-read them. An undisposed condition is a
must-pass failure that caps the sync verdict at FAIL. Reflected in REQ-FDA-008.

### Q15: What ends a wait? (iter1 D10)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 D10
Why unresolved: Any later same-card record would have ended the wait.
Operator verdict: Only a board record whose `resolves` field names that wait id. Reflected in
REQ-FDA-002/019.

### Q16: Is there any exception to "final ceiling hit → hold"?

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 (t1458 exception, encoded)
Why unresolved: On t1458 the leader granted an exception: the card was a prerequisite of an
operator-mandated release scope (t1480 → v3.2.0), the single remaining blocking defect was an
auditor-confirmed acceptance-criterion wording defect, and the auditor said a re-read of the named
hunks suffices.
Operator verdict: Encode it as a mechanically decidable exception with all three conditions required:
release-blocking (on the dependency path of a card in an operator-approved release scope), one
blocking finding of class `ac-wording`, and auditor-listed `reread_hunks`. It also needs a leader
`card:` board record, and grants a hunk-limited fix plus re-read confirmation. Reflected in
REQ-FDA-013.

### Q17: Whose authority admits `factory decide <card> --decider audit`, and how does run get a lease? (iter2 N1)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N1
Why unresolved: A kickoff card holds no lease (`internal/homestate/fr_lease_test.go:148-162`), so
"the card its own lease holds" could never match.
Operator verdict: Authority is the card record owner (the lane label persisted on the row through
T7), not the lease. T8a moves kickoff→run and re-leases the card to that same owner atomically in one
transition, binding the lease as the normal lease path does. Reflected in REQ-FDA-015/016.

### Q18: Is decision-index.md covered by the audited plan-artifact hash? (iter2 N2)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N2
Why unresolved: The digest inputs exclude decision-index.md (P25), so a post-audit reclassification
of a product-level row could pass the decider.
Operator verdict: decision-index.md is part of the plan-artifact hash. Any change after the audited SHA
invalidates the audit (decider refuses on hash mismatch). It is removed from REQ-FDA-011's
delta-exempt list. Consequence recorded in REQ-FDA-018: DEFAULT-APPLIED is filled at plan close,
before the audit, so the audited hash covers it.
Note from manager-spec: this placement is the author's reading of how N2 and N3 compose. The leader
can overrule it.

### Q19: Which empty FOUNDER rows block the audit decider? (iter2 N3)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N3
Why unresolved: An implementation-level row with no Default and an empty verdict passed the decider.
Operator verdict: Any FOUNDER row with an empty verdict, regardless of class. Only rows with a recorded
verdict or DEFAULT-APPLIED pass. Reflected in REQ-FDA-014.

### Q20: How is "release-blocking" decided mechanically? (iter2 N4)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N4
Why unresolved: No relation kinds and no release-scope field were named.
Operator verdict: The leader writes a standing board record of kind `release-scope` listing card ids
for a release. A card is release-blocking iff it is listed there, or reachable from a listed card
through `relate` edges of kind `depends` or `blocks` (the todo relate vocabulary only). No contract
schema change. Reflected in REQ-FDA-002/013.

### Q21: What fields and exit does the exception use? (iter2 N5)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N5
Why unresolved: `defect_class` and `reread_hunks` were consumed but never emitted, and the re-read exit
was undefined.
Operator verdict: Add `defect_class`, `reread_hunks` (file + anchor list) and `blocking_count` to the
REQ-FDA-006 verdict block, and P13 greps those names. The re-read confirmation is a full verdict block
with `scope: reread` that must pass the phase predicate. The hold is released by a board record whose
`resolves` names the hold id. The hunk limit is checked by diffing the two audited SHAs against
`reread_hunks` (nothing else changed except progress and reports). Reflected in REQ-FDA-006/013.

### Q22: Are AC-FDA-019/020 release-blocking? (iter2 N6)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N6
Why unresolved: They were labelled RG, but they assert new behavior.
Operator verdict: Release-blocking, each with a RED probe of today's behavior: P28 (wait records carry
no id) and P29 (only the recurring carrier) for 019; P30 (Open on every bind) for 020. M0 values move to
measurement notes.

### Q23: Does T13 change? (iter2 N7)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N7
Why unresolved: REQ-FDA-009 named undefined "sync thresholds" at T13.
Operator verdict: T13 behavior is unchanged except the D8 binding-condition must-pass, which reaches
T13 through the sync verdict label. Reflected in REQ-FDA-009 and AC-FDA-009.

### Q24: Which rows may carry DEFAULT-APPLIED and still pass the decider? (iter3 N8)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N8
Why unresolved: Once DEFAULT-APPLIED was filled before the audit and covered by the hash, a
product-level, Class-less, or Default-less row filled DEFAULT-APPLIED would pass every mechanical
check.
Operator verdict: DEFAULT-APPLIED counts only on a row whose Class is implementation-level and which
carries a `Default:` line. A product-level row, a row with no Class line, or a row with no Default
line holding DEFAULT-APPLIED makes the decider refuse. Three refusal fixtures go in AC-FDA-014, one
also in AC-FDA-018. Reflected in REQ-FDA-014.

### Q25: Is the `audit` decider an operator-held question? (iter3 N9)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N9
Why unresolved: SPEC-FACTORY-RECORD-001 records the operator's decider set as
`human | llm | llm+jev`, and `audit` lies outside it. REQ-FR-019 refuses any decider other than
`human`.
Operator verdict: It is not an operator-held question. The operator's 2026-10-03 directive to
automate factory decisions and `auto-semantics.md` §9 ("factory decide: kickoff approve/reject —
AUTONOMOUS") already classify Kickoff as autonomous. The `human` decider path stays available, and
keep-set cases still require it. SPEC-FACTORY-RECORD-001 is added to related_specs, and Amendments
rows are planned on REQ-FR-004 (new T8a edge; AC-005 accepted-pair count) and REQ-FR-019 (decider
`audit` accepted only on T8a; an `audit` approval goes to run with the lease). AC-FDA-015 asserts the
new count. Reflected in REQ-FDA-015/016.

### Q26: Does a recorded operator verdict pass the decider? (iter3 N10)

Label: LEADER-DECIDED
Authority anchor: mission contract 07d28c4b — leader ruling 2026-10-03 N10
Why unresolved: REQ-FDA-014 (a recorded verdict passes) and REQ-FDA-018 (a recorded verdict goes
only through the human path) contradicted each other.
Operator verdict: Keep Q19. Narrow REQ-FDA-018 to "a verdict recorded after the audited SHA goes only
through the human path". A verdict recorded before the audited SHA is covered by the hash and passes.
