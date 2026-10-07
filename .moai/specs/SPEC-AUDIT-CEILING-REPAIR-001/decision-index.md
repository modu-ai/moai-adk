# decision-index.md — SPEC-AUDIT-CEILING-REPAIR-001

Decision gate: `interview.decision_gate: on` (`.moai/config/sections/interview.yaml` §interview).

## Ceiling note (plan-audit-3 — status, not a decision)

Iteration 3/3 reached — the Tier M audit ceiling is a final hit (verdict
`.moai/reports/t1560/plan-audit-3.md`, FAIL 0.75, receipt
rcpt-a6fb7c9d03b341c573bf2fb4). A leader ruling on an iteration-4 delta
round is PENDING — noted here as pending, NOT decided; closure rides the
ceiling policy's delta route or the leader's recorded exception.

In-lane plan phase (card t1560): the operator's dispatch settled the WHAT (two
defects, reproduction-first, record-format constraints, consistency reads);
the rows below are the decisions surfaced during assembly that the dispatch
did not settle verbatim. No row is a recommendation — each carries a Default
selected by the published default rule only.

### Q1: Which record carriers carry the admitted debt inventory — both lines persistOutcome writes, or the §G durable record only?

Label: FOUNDER
Class: implementation-level (the operator's dispatch already ordered the
record to carry the inventory — only the carrier scope of the same fix call
is decided here; no command behavior change beyond the mandated one)
Authority anchor: none — SPEC-AUDIT-CEILING-001 REQ-ACE-004 fixes THAT the
outcome record enumerates the debts; it does not settle whether the same
call's machine-local trail line carries them too
Why unresolved: the dispatch's fix direction names "the persisted record"
with the §G line grammar; persistOutcome writes two lines, and a trail line
that still drops the inventory preserves the same defect class one line below
the fix
Default: both lines the same persistOutcome call writes carry the inventory
(rule: preserves current behavior for every other outcome; undo is a single
revert of this SPEC's own commits; the trail is machine-local
`.moai/state/` state, so no additional user-visible surface beyond the
mandated §G fix)
Alternate: §G durable record only
Operator verdict: DEFAULT-APPLIED 2026-10-07T00:45:00Z manager-spec (lane-2, card t1560)

### Q2: What encoding does the `debts=` record field use for the per-debt ID / dispose_in / description triple?

Label: FOUNDER
Class: implementation-level (additive field content encoding; the
one-line + machine-parseable + escaped constraints are dispatch-settled)
Authority anchor: none — the dispatch delegates the exact record format
("You choose the exact record format — keep it one-line, machine-parseable,
and consistent with the existing line grammar")
Why unresolved: the existing grammar is `key=value` with Go-%q-quoted free
text (`reasons=%q`), but a multi-field, multi-entry inventory needs an entry
separator a free-text description may collide with under naive %q joining
Default: a JSON array value inside the `debts=` key
(`{"id":...,"dispose_in":...,"description":...}` per entry) — the stdlib
encoder guarantees a single line and complete escaping of quotes, semicolons,
and newlines with no custom scanner (rule: preserves current behavior — all
debt-free lines byte-identical, one additive field; single revert; smallest
user-visible delta)
Alternate: per-debt %q-quoted field triples in the reasons= style (requires a
quote-aware entry splitter once descriptions carry separators)
Operator verdict: DEFAULT-APPLIED 2026-10-07T00:45:00Z manager-spec (lane-2, card t1560)

### Q3: Where does the D1 dual-family resolution live — at the previousAuditedSHA site, or inside iterationOf (extended to dual-family parsing)?

Label: FOUNDER
Class: implementation-level (internal resolution site; behavior identical
either way at this call graph)
Authority anchor: none — no completed SPEC row or operator setting names the
resolution site
Why unresolved: `iterationOf`'s name and doc contract say "convention-family
file name"; its only two callers are inside `previousAuditedSHA`
(`audit_ceiling.go:274,286`), so extending it vs. resolving locally are
behaviorally equivalent today
Default: resolve the latest round number at the `previousAuditedSHA` site,
leaving `iterationOf`'s convention-only contract and signature unchanged
(rule: preserves current behavior — the helper's contract and its other call
site are untouched; smallest surface; avoids a helper-contract refactor this
card's scope discipline excludes)
Alternate: extend `iterationOf` to dual-family parsing and update its doc
contract
Operator verdict: DEFAULT-APPLIED 2026-10-07T00:45:00Z manager-spec (lane-2, card t1560)

### Q4: Does card t1560 absorb the third defect (§G append atomicity, D3) or does it split into a new card?

Label: FOUNDER
Class: product-level (changes the card's shipped deliverable surface — adds a
behavior fix to the SPEC's scope, not merely its implementation shape)
Authority anchor: none in the committed register — the settling act is the
leader ruling of 2026-10-07 (scope extension adopted; basis: 주제당 한 장 —
one card per topic with milestones; that rule lives in card-issuance memory,
not in a committed artifact, so no verifiable anchor exists)
Why unresolved at authoring: the dispatch originally defined a two-defect
surface (D1, D2); D3 surfaced post-authoring from the turn-end codex review
gate's overlay repro on `903ccd028`
Operator verdict: ADOPTED — leader ruling 2026-10-07: scope extension
adopted on the 주제당 한 장 basis; D3 joins this SPEC as REQ-ACR-008 /
AC-ACR-013 and folds into the persistence milestone M3 (recorded by
manager-spec per the leader's instruction)

### Q5: Does card t1560 absorb the fourth defect (CountAuditRounds overflow collapse at audit_counter.go:111) or split it to a new card?

Label: FOUNDER
Class: product-level (adds a round-counting semantics change — unparseable
iteration numbers count as their own round — to the card's shipped behavior)
Authority anchor: none in the committed register — the settling act is the
leader ruling #2 of 2026-10-07 (folded into the D1 requirement family;
basis: 주제당 한 장 — one card per topic; that rule lives in card-issuance
memory, not in a committed artifact, so no verifiable anchor exists)
Why unresolved at authoring: surfaced post-plan-audit-1 as the leader's
second scope ruling; the counter defect was not part of the original
dispatch surface
Operator verdict: ADOPTED — leader ruling #2 2026-10-07:
`audit_counter.go:111` folded into the D1 requirement family as REQ-ACR-009
/ AC-ACR-014, M2 extended, the leader-required one-line semantics note
carried in the AC wording (recorded by manager-spec per the leader's
instruction)

### Q6: Does the base report (plan-audit.md, convention match with no number) fold into the same D4 round-counting fix as the overflow suffix?

Label: FOUNDER
Class: product-level (same round-counting semantics axis as Q5 — the base
report's round identity changes)
Authority anchor: none in the committed register — the settling act is the
leader ruling #3 of 2026-10-07 (base-report fold approved; same 주제당 한 장
basis as Q4/Q5)
Why unresolved at authoring: the base-report collapse is the third mode of
the same n=1-init mechanism (`audit_counter.go:109` initializes `n = 1`
before the Atoi attempt) but was not named in ruling #2; the leader's gate
measured sources=3, count=1 on `903ccd028` (one normal + two overflow files)
Operator verdict: ADOPTED — leader ruling #3 2026-10-07: base-report
parity folded into REQ-ACR-009 / AC-ACR-014 (same IDs, cells extended — base
+ numbered → 2, one normal + N unparseable → 1+N, bare base alone
unchanged); no new IDs; the R1 stale-count sweep executed in the same pass
(recorded by manager-spec per the leader's instruction)

### Q7: Does the card proceed to a fifth audit round, and on what closure rule?

Label: FOUNDER
Class: implementation-level (process scope of the card's closure — no shipped
behavior change)
Authority anchor: none in the committed register — the settling act is the
operator's approval relayed by the leader, 2026-10-07 (per the Q4-Q6
convention)
Why unresolved at authoring: iteration 3/3 reached the ceiling (final hit);
whether a fifth round runs, and what happens if it fails again, is a
ceiling-policy disposition no committed rule settles for this card
Operator verdict: ADOPTED — operator approval 2026-10-07: a fifth delta
round is authorized; a fifth FAIL closes the card with a recorded debt
unconditionally (no further repair rounds after it; the debt inventory is
recorded per the ceiling policy's final-hit record)
