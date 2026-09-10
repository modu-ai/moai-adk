# SPEC Review Report: SPEC-STATUS-TRANSITION-VALIDITY-001

Card: **t376** · Iteration: 2/2 (Tier M ceiling) · Tree of record: `3f03d9c36` (`WT-status-transition-gap`)
Auditor: plan-auditor (independent). Reasoning context ignored per M1 Context Isolation.

**Scope**: confirming iteration, scoped to the enumerated defect delta from iter-1 plus a bounded
regression sweep. Not a from-scratch re-audit. iter-1's report is a prior claim set, not authority —
every item below was re-measured in this tree.

Verdict: **PASS**
Overall Score: **0.96** (Tier M PASS threshold 0.80; iter-1 was 0.90 — monotonic increase, so the
LEAN STOP-on-regression clause does not fire)

---

## Must-Pass Results (re-confirmed mechanically, not carried over)

- **[PASS] MP-1 REQ number consistency** — `grep -o '^### REQ-STV-[0-9]*' spec.md` → 001-015, fifteen
  ids, no gaps, no duplicates, uniform padding. `REQ-STV-012` still sits between 009 and 010
  (spec.md:315) — reading order only, iter-1's optional D6, unchanged.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-STV-*` in
  spec.md), never against the Given-When-Then ACs, which are the correct verification-layer form.
  `grep -c '^### REQ-STV-[0-9]* ('` → 15: every REQ carries an explicit pattern label. The REQ set is
  unchanged from iter-1 in the clauses bearing on this criterion.
- **[PASS] MP-3 YAML frontmatter validity** — 12 canonical fields present with correct types
  (spec.md:1-15), plus `tier: M`. Measured, not eyeballed:
  `go run ./cmd/moai spec lint --json .moai/specs/SPEC-STATUS-TRANSITION-VALIDITY-001/spec.md` → `[]`,
  rc=0. Zero findings, so no `FrontmatterInvalid`.
- **[N/A] MP-4 language neutrality** — subject is this repository's own Go lint engine
  (`internal/spec`), single-language. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+'` across
  all three artifacts returns exactly one id: `SPEC-STATUS-TRANSITION-VALIDITY-001` (itself). No
  external reference, so nothing owed. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` → 0 in all three artifacts.
  Auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' <spec dir>/` → rc=1, no match.
  `progress.md` `open_clarifications: 0` agrees. No `research.md` (Tier M does not require one).

---

## The five commissioned defects

### D4 (High) — `AC-STV-016` had no failing condition — **CLOSED**

Current text, acceptance.md:191-206 (quoted, not summarized):

> ### AC-STV-016 — non-overlap with `StatusValueEnumRule`, measured
> **Given** the full corpus, **When** `moai spec lint --json` runs after the change and the documents
> reported under `StatusTokenUnrecognized` are intersected with those reported under
> `StatusValueInvalid`, **Then** the intersection is **empty** — zero documents appear in both sets.
>
> A non-empty intersection **fails this AC**. It may not be discharged by listing the overlap: an
> overlapping document is the duplication REQ-STV-015 prohibits until someone shows otherwise, and
> the card does not close on it.

**Can it now fail? Yes, determinately.** The Then is an outcome (`intersection is empty`), not an
activity. A run producing one overlapping document fails it with no interpretive latitude, and
acceptance.md:200 forecloses by name the exact discharge iter-1 identified. The hedge that made the
old version vacuous is now explicitly subordinated (acceptance.md:209-211: "an AC that could be
satisfied by reporting the overlap would measure nothing at all"). This is a structural repair — the
Then changed form — not a firmer re-wording of the same shape.

**Is the escape hatch bounded? Yes, with a residual risk I am recording rather than absorbing.**
Verbatim, acceptance.md:202-206:

> Two remedies are available, both of which must land before the AC passes: narrow the new rule so
> the overlap disappears, or — where a specific overlap is genuinely correct — enumerate the exact
> documents allowed to appear in both, record them in `progress.md` §E.2 with both messages and the
> reason each is not duplication, and amend this AC to name that bounded set, so that any document
> outside it still fails.

Three properties make it bounded: the allowed set must be **enumerated** (finite, named documents,
not a predicate), each entry must carry **both messages and a per-document reason**, and the amended
AC still **fails on anything outside the set**. The post-amendment AC therefore retains a failing
condition; it does not degenerate into "whatever we measured".

The residual risk, stated plainly: the amendment is performed by the same actor that measured the
overlap, and "genuinely correct" is a judgment, not a mechanical predicate. Nothing here prevents an
overlap from being justified into the allowed set. What it does prevent is that happening
**silently** — the widening is a SPEC diff plus per-document written reasons in `progress.md` §E.2,
both visible to the run-phase auditor. That is a materially different failure mode from the iter-1
version, which could be discharged with no artifact at all. I judge D4 closed, and hand the
amendment path to the run-phase auditor as something to read rather than assume.

### D2 (High) — no AC bound the gating-population measurement — **CLOSED**

`AC-STV-019` is new (acceptance.md:240-265), covers `REQ-STV-009`, and reads:

> **Given** the post-change corpus lint output, **When** the findings are split by their `advisory`
> field (`jq '[.[] | select(.advisory != true)] | length'` and the same filtered to the two new
> codes), **Then** the non-advisory count is reported — overall and for each of the two new codes —
> and, **when that count is non-zero**, an explicit decision is recorded in `progress.md` §E.2 before
> the card may close: either accept that `spec-lint --strict` now reddens on the integration branch,
> or hold the finding advisory and record why.

**Does it bind? Yes, and doubly.** Run phase cannot satisfy it without producing the number: the Then
names the number as the deliverable, supplies the exact command that produces it, and requires the
per-code split as well as the overall figure. The DoD repeats the binding independently
(acceptance.md:278-279: "the advisory split and gating decision of AC-STV-019 … recorded in
`progress.md` §E.2"). Both halves are binary-checkable by reading `progress.md` §E.2 after the run.

**It is not a vacuity in the D4 shape.** iter-1's objection to `AC-STV-016` was that an activity had
been substituted for the outcome its REQ prohibits. That does not apply here: `REQ-STV-009`'s
*behavioural* outcome is already bound by `AC-STV-018` (advisory `false` in the JSON, `HasErrors()`
true), and `AC-STV-019` adds the process gate on top of it. A process AC layered over a behavioural
AC is a different construction from a process AC replacing one.

**Not satisfiable in advance — confirmed intended, and correctly stated** (acceptance.md:264-265:
"this AC is deliberately not satisfiable in advance. The decision is recorded **after** M2 produces
the count, never predicted before it"). Correct; not counted against testability.

The AC's supporting measurements, re-run by me in this tree rather than read off the SPEC:

| Claim in acceptance.md:255-259 | My measurement | Result |
|---|---|---|
| 1096 baseline findings | `jq 'length' lint-baseline.json` → 1096 | matches |
| `advisory != true` count is 0 | `jq '[.[] \| select(.advisory != true)] \| length'` → 0 | matches |
| `spec-lint.yml:40` runs `--strict` | `sed -n '40p'` → `run: go run ./cmd/moai spec lint --strict` | **exact** |
| "green because nothing gates" | all 1096 findings are `severity: warning` (`group_by(.severity)` → one row), and `lint.go:56-66` escalates only on `Strict && warning && !Advisory` | **established, not merely asserted** |

### D1 (Medium) — §A.5 D2's rationale contradicted its own census — **CLOSED**

The decision line now stands on independent ground (spec.md:175-178): *"an emission-site `Advisory`
would make the rule unable to gate anything, anywhere, which would close the gap in appearance only."*
That is a property of the mechanism, not of the corpus — the census cannot contradict it, which is
what makes it a real replacement rather than a re-phrasing.

The withdrawn premise is retracted in place rather than quietly deleted (spec.md:180-190):

> D2 was first justified by the claim that "the overwhelming majority of what this rule would flag
> ends in `completed`" … The census says otherwise: of the ~98 projected findings, **50** end in
> `completed` and **48** end in `implemented`, which is not in `terminalStatusEnum` … So it is
> **roughly half**, not an overwhelming majority.

**Consistency with §A.6's own numbers**: §A.6 (spec.md:223-224) projects 98 = 50 `draft → completed` +
48 `completed → implemented`. The corrected block uses exactly those figures, and 50/98 = 51% —
"roughly half" is right. It also labels them as *projected*, keeping the correction on the right side
of the SPEC's own projection/observation boundary.

**The cited coordinate is exact.** `internal/spec/lint.go:1290-1295` is `terminalStatusEnum =
map[string]bool{superseded, archived, rejected, completed}`. I read the lines. `implemented` is
absent, so the 48 genuinely sit outside the layer-4 shelter, as the corrected text says.

**And the correction terminates properly** rather than leaving an unmeasured gap behind: it names the
unknown ("the gating population is therefore unknown until M2 measures it") and points at the AC that
closes it (`AC-STV-019`). D1's fix and D2's fix are wired to each other.

### D3 (Medium) — `~6` should be `~7` — **CLOSED**

`grep -rn '~6' <spec dir>/` → rc=1, **zero matches** (an unanchored search, so stricter than the
`~6\b` the lead ran). The four sites now read `~7` / `7`:

| Site | Text |
|---|---|
| spec.md:224 | `and **7** \`StatusTokenUnrecognized\`` |
| spec.md:396 | `Repairing the ~7 SPECs` (§C non-goal — the occurrence iter-1 did not list) |
| plan.md:77 | `the §A.6 hand projections (~98 / ~7)` |
| acceptance.md:162 | `§A.6 projects roughly 98 and 7 respectively` |

**Arithmetic verified against the census, not against the author's summary.** spec.md:224-226
enumerates `Completed` 3 + `synced` + `approved` + `Superseded` + `cancelled`, written "3+1+1+1+1".
That sums to 7, and the enumeration matches the §A.4 census rows exactly (`Completed → completed` 3;
`synced → completed` 1; `approved → completed` 1; `Superseded → superseded` 1; `cancelled → rejected`
1 — spec.md:146-151). The quote-wrapped `"in-progress"` is correctly excluded with a stated reason
(it sits behind a `(none)` skip and never reaches the token test).

The residual `6`s in the SPEC directory are all unrelated and correct: `InvalidREQID | 6` (a real
baseline count, spec.md:68, which I confirmed against `lint-baseline.json`), `matrix row 6`, and the
section number `§A.6`.

### D7 (Medium) — `AC-STV-003` unsatisfiable as worded — **CLOSED (with a smaller residual)**

Now reads (acceptance.md:39-40): *"both emit `StatusTransitionInvalid` for that SPEC, and the two
findings differ in no field other than the commit SHA and the file path."*

The blocking half is fixed. iter-1's objection was that the Given constructs two repositories, so
`file` necessarily differs and a **correct** implementation would fail the AC as written. `and the
file path` removes that.

**The residual, recorded as a new minor defect (D10 below):** the `Finding` shape has no SHA field. I
measured its keys: `jq '.[0] | keys_unsorted'` → `file, line, severity, code, message, advisory`. The
SHA lives *inside* `message` (`AC-STV-001`, acceptance.md:18-19, requires the message to name the
performing SHA). So the fields that actually differ between the two repositories are `file` and
`message`, and the AC excuses "the commit SHA", which is a value rather than a field. An implementer
comparing field-by-field still finds `message` differing. This is a smaller version of the same
value-vs-field imprecision, not a return of the unsatisfiability.

---

## Regression sweep (bounded, as commissioned)

**Traceability — still 1.00, re-derived.** All 20 `*Covers …*` lines name REQ ids that exist
(001-015). Every REQ has ≥1 AC: 001→001/002/014, 002→003, 003→001, 004→002, 005→004/005/006/007/007a,
006→008, 007→009, 008→011, 009→**018 + 019**, 010→013, 011→010, 012→012, 013→015, 014→017, 015→016.
`REQ-STV-009`'s coverage changed as expected — it gained `AC-STV-019` and kept `AC-STV-018`. No
orphaned AC, no uncovered REQ.

**AC set shape — confirmed.** `grep -c '^### AC-STV-'` → **20**. Ids in file order: 001, 002, 003,
004, 005, 006, 007, **007a**, 008, 009, 010, 011, 012, 013, 014, 015, 016, 017, 018, 019.
`sort | uniq -d` → empty (no duplicates). Numeric ids 001-019 contiguous, plus `007a`.
`progress.md` `ac_count: 20` agrees, and its inline comment ("19 numeric ids AC-STV-001..019 +
AC-STV-007a") is accurate. The DoD's arithmetic is internally consistent too (acceptance.md:269:
"AC-STV-001..013 incl. 007a, plus AC-STV-014..019" = 14 + 6 = 20).

**No new vacuous AC.** I walked all 20 and asked of each: what run outcome fails this? Every one has
a determinate answer. The two edited/new ACs are the ones that matter and both hold — `AC-STV-016`
fails on a non-empty un-enumerated intersection; `AC-STV-019` fails on an unreported count, or on a
non-zero count with no recorded decision. The silence ACs (004-009, 017) remain individually
satisfiable by a rule that was never wired in, which is exactly why `AC-STV-010` exists as the live
control in the same execution — that structure is intact (acceptance.md:5-7, 112-123) and the DoD
re-binds it (acceptance.md:271-272).

**One coupling worth naming; not a defect.** `AC-STV-019` passes on a zero non-advisory count, and a
zero count is produced both by "the rule fires and everything demotes" and by "the rule never fires
at all". `AC-STV-019` does not itself separate those. `AC-STV-013` does — it requires the new count
for **every** code side by side with baseline and states that "a projection that misses by a wide
margin is a finding to explain before the card closes" (acceptance.md:158-166), so
`StatusTransitionInvalid = 0` against a projection of ~98 cannot pass quietly. The AC set covers it;
I record the dependency so the run-phase auditor reads the two together rather than either alone.

**Optional-class defects from iter-1 — none got worse:**

| iter-1 | State now |
|---|---|
| D5 (REQ-STV-015 mixes a verification activity into the requirement) | Unchanged — spec.md:348 still reads "shall be shown, by measurement rather than by assumption". Not worse; note the activity it names is now genuinely discharged by `AC-STV-016`'s outcome form |
| D6 (`REQ-STV-012` out of sequence) | Unchanged — still at spec.md:315, between 009 and 010 |
| D8 (two loose line ranges) | Unchanged. spec.md:101 still cites `lint_ownership.go:78-94` for a function spanning 64-95; plan.md:15 still cites `:412-415` for the `rec.AuthoredByAgent == ""` guard I read at **414-416**. Both substantive claims remain true |
| D9 (`AC-STV-007a` suffix form) | Unchanged |

**Evidence spot-checks (the lead's figures, re-measured rather than accepted):**

| Lead's figure | My command | Result |
|---|---|---|
| `lint-after-fixes.json` 1096 findings | `jq 'length'` → 1096 | matches |
| no per-code delta vs baseline | `cmp lint-baseline.json lint-after-fixes.json` → **byte-identical** (both 373,334 bytes) | matches — but see the note below |
| 0 findings on this SPEC | `jq '[.[] \| select(.file \| test("STATUS-TRANSITION-VALIDITY"))] \| length'` → 0 | matches, **and independently corroborated** by my own single-file lint returning `[]` |
| 0 non-advisory in the baseline | `jq '[.[] \| select(.advisory != true)] \| length'` → 0 | matches |
| strict invocation at line 40, not 36 | `sed -n '36,40p' .github/workflows/spec-lint.yml` | **line 40 is right.** Line 36 is `- name: Run SPEC lint` (the step header); line 40 is `run: go run ./cmd/moai spec lint --strict`. The SPEC's own citation (acceptance.md:258) says 40 and quotes the command, so the SPEC is correct and iter-1's `:36` was the imprecise one. A stale coordinate is this card's own subject matter, so I read the file rather than adjudicating between two reports |
| `grep -rn '~6\b'` → 0 | `grep -rn '~6'` (unanchored, wider) → rc=1 | matches, on a stricter test |

Evidence-hygiene note, carried forward and non-blocking: `lint-after-fixes.json` being byte-identical
to `lint-baseline.json` is consistent with "the SPEC edits contribute zero findings", but a
byte-identical artifact is also exactly what a stale copy looks like, so it carries no evidentiary
weight standing alone. What establishes the claim is my independent
`go run ./cmd/moai spec lint --json <spec.md>` returning `[]` in this tree — that is the attributed
measurement; the JSON file is corroboration.

---

## Category Scores (rubric-anchored)

| Dimension | iter-1 | iter-2 | Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.90 | **0.95** | 1.0-band, small deduction | D1 and D3 both closed — the rationale now matches its own census (spec.md:180-190), and the arithmetic matches its own enumeration (spec.md:224-226, zero `~6` remaining). Deducted for the optional-class residue: D5's verification clause inside `REQ-STV-015`, D6's out-of-sequence REQ, D8's two loose ranges |
| Completeness | 0.90 | **0.95** | 1.0-band, small deduction | D2 closed: the gating-population risk now has a gate (`AC-STV-019`) plus a DoD line. All required sections present; `§C` carries seven `### Out of Scope — …` headings, each with bullets. Deducted for the one plan deliverable still carrying no AC (plan.md M1's check-order test — iter-1 recorded this and declined to block on it; I concur) |
| Testability | 0.80 | **0.95** | 1.0-band, small deduction | Both testability defects closed. `AC-STV-016` now has a determinate failing condition; `AC-STV-003` is satisfiable. Deducted for D10 (the residual field-vs-value wording in `AC-STV-003`) and for the `AC-STV-016` amendment path resting on a written judgment rather than a mechanical predicate |
| Traceability | 1.00 | **1.00** | 1.0 | Re-derived above. 15/15 REQs covered, 20/20 ACs name an existing REQ, no orphans |

Aggregate: (0.95 + 0.95 + 0.95 + 1.00) / 4 = **0.9625 → 0.96**.

---

## Defects Found

**D10 — `AC-STV-003` excuses a value where it names fields.** Severity: **minor**.
Class: **optional**.
acceptance.md:39-40 — *"the two findings differ in no field other than the commit SHA and the file
path."* The `Finding` type has no SHA field (measured: `file, line, severity, code, message,
advisory`); the SHA is a substring of `message`, which `AC-STV-001` requires it to be. So the
differing **fields** are `file` and `message`. An implementer comparing field-by-field still sees a
`message` mismatch and must decide on their own authority whether it counts.
*Required fix (one clause)*: "…differ in no field other than the file path and, within `message`, the
commit SHA."
*Why optional rather than blocking*: the property being pinned is unambiguous from the surrounding
text and from the probe row the AC cites, and the failure mode is an implementer pausing, not an
implementer shipping something wrong. It does not reproduce iter-1's D7, where a **correct**
implementation would have failed the AC as written.

**Carried forward, unchanged, optional-class — not routed into a revision** (per iter-1's own
recommendation and M6 finding-consumption discipline): D5, D6, D8, D9. None got worse; state
tabulated above.

No blocking-class defect remains.

---

## Regression Check (iteration 2)

Defects from iteration 1:

- **D1** (§A.5 D2 rationale contradicts census) — **RESOLVED**. spec.md:180-190 retracts the premise
  by name, restates as "roughly half" against the correct 50/48 split, cites `lint.go:1290-1295`
  (verified exact), and routes the unmeasured remainder to `AC-STV-019`.
- **D2** (no AC binds the gating population) — **RESOLVED**. `AC-STV-019` added
  (acceptance.md:240-265), covers `REQ-STV-009`, bound again by the DoD at acceptance.md:278-279.
- **D3** (`~6` → `~7`) — **RESOLVED**. `grep -rn '~6'` → rc=1 across the SPEC directory; four sites
  corrected; the enumeration sums to 7 and matches the census rows.
- **D4** (`AC-STV-016` cannot fail) — **RESOLVED**. Outcome-form Then; a non-empty intersection
  fails; the listing discharge is explicitly foreclosed; the amendment path stays bounded and leaves
  the amended AC still able to fail.
- **D5** (REQ-STV-015 mixes verification into the requirement) — **UNRESOLVED, deliberately**.
  Optional-class, left open on iter-1's recommendation. Not worse.
- **D6** (REQ-STV-012 out of sequence) — **UNRESOLVED, deliberately**. Optional-class. Not worse.
- **D7** (`AC-STV-003` unsatisfiable) — **RESOLVED** as to satisfiability; a smaller value-vs-field
  imprecision remains and is recorded as D10 (optional).
- **D8** (two loose line ranges) — **UNRESOLVED, deliberately**. Optional-class. Both ranges
  re-checked against the source; both substantive claims still true. Not worse.
- **D9** (`AC-STV-007a` suffix form) — **UNRESOLVED, deliberately**. Optional-class. Not worse.

No stagnation: every blocking-class defect from iter-1 closed in one revision, and no defect appears
unchanged across iterations except the four deliberately deferred as optional.

---

## Recommendation

**PASS.** All seven must-pass criteria hold, all five commissioned defects are genuinely closed
against the current text (not against the author's summary of it), and 0.96 clears the Tier M
threshold of 0.80 with margin.

The repair quality is worth stating precisely, because "the author says it is fixed" is the claim
this iteration exists to test. `AC-STV-016` did not merely acquire firmer language — it changed from
an activity-form Then to an outcome-form Then, which is the structural change that makes an AC able
to fail. `AC-STV-019` is a new gate that cannot be satisfied without producing a number that does not
yet exist. `§A.5 D2` now rests on a mechanism argument the census cannot contradict, with the
withdrawn premise retracted in place rather than deleted. Those are the three that mattered.

Two items to hand forward rather than block on:

1. **D10** — Priority Low. A one-clause wording fix to `AC-STV-003`; safe to fold into the run-phase
   change, or to leave.
2. **The `AC-STV-016` amendment path** — Priority Medium *for the run-phase auditor, not for plan
   phase*. If run phase enumerates an allowed-overlap set and amends the AC, that amendment is a
   judgment made by the party that measured the result. It is visible by construction (a SPEC diff
   plus per-document reasons in `progress.md` §E.2), which is what makes it acceptable here — but it
   must be read, not assumed, before the card closes.

`progress.md` `plan_audit_blocking_closed: [D1, D2, D3, D4, D7]` and
`plan_audit_optional_open: [D5, D6, D8, D9]` both agree with my independent finding.

Iteration ceiling reached (Tier M = 2). No further plan-phase iteration is warranted. This verdict
does not open, and does not substitute for, the Implementation Kickoff Approval human gate.
