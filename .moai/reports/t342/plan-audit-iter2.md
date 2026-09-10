# SPEC Review Report: SPEC-MOVING-REF-GUARD-001 (card t342) — iteration 2

Iteration: 2/2 (Tier M ceiling — this verdict closes plan-phase)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.86** (Tier M PASS threshold 0.80)
Delta against iter-1: **0.80 → 0.86, +0.06 — monotonic upward.** No STOP escalation.

Reasoning context ignored per M1 Context Isolation. The audit reads only the four artifacts
plus the tree they describe. The remediation summary supplied in the dispatch was treated as a
set of claims to verify, not as evidence.

> **Report location note.** Written inside the worktree, as in iter-1: the isolation guard
> refuses a write from this session to the primary checkout. `.moai/reports/t342/` is not
> gitignored; the lead commits it into the branch.

## Audit attribution

| Item | Value |
|---|---|
| Tree | `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342` (`git rev-parse --show-toplevel`) |
| Branch | `WT-moving-ref-guard` |
| Commit judged | **`3dc45b361`** (`git log --oneline -1` → `docs(t342): commit the plan-audit verdict as the decision record`) |
| Tree state | `git status --short` → empty at entry **and** at verdict |
| spec.md sha256 | `8dc3b3fbcfc7…` (identical at entry and at verdict) |
| acceptance.md sha256 | `b0b924dc8ee1…` (identical at entry and at verdict) |
| Artifact version | v0.3.0 |

**The tree did not move under this audit.** HEAD, `git status`, and both artifact hashes were
read at entry and again immediately before this verdict and are identical. The iter-1 P1
condition (a writer active inside the audit window) did not recur. Every finding below was
measured against `3dc45b361`.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -c '^- \*\*REQ-MRG-'` → 11; enumerated IDs
  `REQ-MRG-001`..`REQ-MRG-011`, sequential, no gaps, no duplicates, uniform 3-digit padding.
  `REQ-MRG-011` is the D9 addition and extends the range without breaking it.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer**
  (`REQ-XXX` in spec.md), not the verification layer. Opening modality per requirement, read
  mechanically (`grep -oE '^- \*\*REQ-MRG-[0-9]+\*\*: [A-Za-z]+'`): `The` ×5 (001, 004, 005,
  006, 011 — Ubiquitous), `When` ×2 (002, 003 — Event-driven), `While` ×4 (007, 008, 009, 010
  — State-driven). **Zero `Where` conditionals**, confirming the optional D8 relabel was
  applied in full. All 11 carry subject + `shall` + response.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present and well-typed:
  `version: "0.3.0"` (quoted semver, now consistent with the HISTORY head row — D6 closed),
  `updated: 2026-08-28`, `status: draft`, ISO dates, `priority: P2`, `lifecycle: spec-anchored`,
  `tags` comma-string. No rejected snake_case alias. Optional `era: V3R6` / `tier: M` /
  `related_specs` valid.
- **[N/A] MP-4 Section 22 language neutrality** — single-domain SPEC (this repo's Go lint engine
  plus one doctrine section). No multi-language tooling claim. Auto-passes per the MP-4 N/A rule.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — 10 referenced SPEC IDs extracted across all
  four artifacts; all 10 resolve on disk. Statuses read individually:
  `SPEC-AGENT-PARALLEL-OPT-001` completed · `SPEC-DESIGN-MOAIWEBV2-002` completed ·
  `SPEC-GRAPH-FRESHNESS-CADENCE-001` in-progress · `SPEC-STAMP-REACHABILITY-001` completed ·
  `SPEC-V3R5-LATE-BRANCH-001` completed · `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` **superseded** ·
  `SPEC-V3R6-MULTI-SESSION-COORD-001` completed · `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001`
  implemented · `SPEC-V3R6-TEST-REFACTOR-001` completed · `SPEC-VERIFICATION-COMPLETENESS-001`
  completed. The three new references since iter-1 (`SPEC-AGENT-PARALLEL-OPT-001`,
  `SPEC-STAMP-REACHABILITY-001`, `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001`) are all live and
  cited evidentially. The single superseded reference is unchanged from iter-1 and reconciled by
  §G "Retrofitting closed or grandfather-era SPECs" being out of scope. No BLOCKING finding.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall` → 0 across all four artifacts.
  Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -c 'NEEDS CLARIFICATION'` → 0 in `plan.md` and
  `progress.md`. No `research.md` at Tier M. No open marker.

## Category Scores

| Dimension | iter-1 | iter-2 | Rubric Band | Evidence |
|---|---|---|---|---|
| Clarity | 0.78 | **0.88** | 0.75–1.0 boundary | D1's fix is real and reaches R4 on its own (verified by independent walk, below). Test 1 now states its evaluation time; the tie-break routes to Test 4; §D.3 carries a step-by-step trace *and* names the ANCHOR-branch validation gap as a limitation. Q4 opens the residual honestly rather than closing it. Deducted for two minor wording residues (D15, D16). |
| Completeness | 0.82 | **0.84** | 0.75–1.0 boundary | D5 (DoD now Q0-Q4), D6 (version), D7 (plan.md sweep), D9 (REQ-MRG-011), D10 (three alternations recorded verbatim) all closed and independently verified. Deducted for D14 — the SPEC states an obligation on M1's doctrine that no criterion decides — and for the uncharacterized R4 residual scope (part of D13). |
| Testability | 0.72 | **0.78** | 0.75 band | Four criteria repaired and two added, all verified genuinely falsifiable: AC-MRG-014's mutant is real, CM-2 catches the fetch-verb key *and* generalizes beyond it, AC-MRG-005's era precondition is now stated, AC-MRG-010 decides on both surfaces. Held down by D13 — AC-MRG-013 rests on a fixture no correct rule would flag, so the R4 exemption, the [HARD] surface, has no criterion that exercises it. |
| Traceability | 0.88 | **0.92** | 1.0 band, one gap | All 11 REQs covered; all 14 ACs trace to a REQ; all 14 ACs assigned to a milestone (M1: 007/011 · M2: 003 · M3: 001/002/004/005/006/008/009/013/014 · M4: 010 · M5: 012 — union is exactly 001-014, no orphan in either direction). D9's orphan closed by `AC-MRG-010 → **Traces to:** REQ-MRG-011`. One residual gap of the same kind: D14. |

Aggregate (arithmetic mean, the method used at iter-1): (0.88 + 0.84 + 0.78 + 0.92) / 4 = **0.855 → 0.86**.

## Regression check — every iter-1 defect, verified individually

| # | iter-1 severity | Status | Evidence |
|---|---|---|---|
| D1 | critical / blocking | **RESOLVED** | Test 1 carries the read-time clause verbatim; the tie-break now reads "run Test 4 — do not resolve to ANCHOR"; Test 4's gate extended with "and also whenever Tests 1 and 3 disagree". Walked independently — see below. |
| D2 | critical / blocking | **RESOLVED** | AC-MRG-014 present. Verified falsifiable: its fixture carries `origin/main`, git context, no SHA, and no claim marker, so the claim-marker-dropped mutant fires on it and turns the criterion red. Over-fire figure re-measured here: **495**, reproduces exactly. |
| D3 | major / blocking | **RESOLVED** | `acceptance.md` §A carries a `[HARD] Fixture era precondition` naming both satisfying forms and citing `isGrandfatheredSpecDir` → `applyEraDemotion` → `Report.HasErrors` as the reason. |
| D4 | major / blocking | **RESOLVED** | AC-MRG-010 now runs both `git diff --name-only "$BASELINE_SHA"..HEAD -- .moai/specs \| grep -v …` and `git status --short -- .moai/specs`, with the mutation restated as a *committed* pin and the `SPEC-DESIGN-MOAIWEBV2-002` AC-MWA-007a precedent cited. |
| D5 | major / optional | **RESOLVED** | DoD reads "Open questions **Q0-Q4**", with Q4 newly opened by D1's tie-break change. |
| D6 | major / blocking | **RESOLVED** | `version: "0.3.0"`, `updated: 2026-08-28`, HISTORY head row 0.3.0. The miss is recorded in the HISTORY row rather than silently corrected. |
| D7 | minor / blocking | **RESOLVED** | `plan.md:24` now "both predicate classes and all four remedies"; `plan.md` §G now "the message names four branches". |
| D8 | minor / optional | **RESOLVED** | Zero `Where` conditionals (measured, MP-2 above). |
| D9 | minor / optional | **RESOLVED** | REQ-MRG-011 added; AC-MRG-010 carries `**Traces to:** REQ-MRG-011`. |
| D10 | minor / optional | **RESOLVED** | Three alternations recorded verbatim in `progress.md` §E.1. All three reproduce exactly in this tree — see the reconciliation table below. |
| D11 | critical / blocking | **RESOLVED** | CM-2 added; Q0 settles the imperative-structure key. Verified to close the named bypass and to generalize — see below. |
| D12 | major / blocking | **RESOLVED** | §D.4 restated for four remedies with a pricing table; R4's cost stated and its enforcement layer named honestly in L6's second residual. See below. |
| P1 | major / process | **NOT RECURRED** | Tree quiescent throughout this audit; hashes identical at entry and verdict. |

**No defect appeared unchanged across both iterations. No stagnation.**

## The measurement reconciliation (D10's real test)

All four load-bearing figure sets re-run in this tree at `3dc45b361`, using the commands now
recorded verbatim in `progress.md` §E.1:

| Figure | SPEC states | I measured | Command |
|---|---|---|---|
| Corpus filter 2 | 527 | **527** | `grep -rnE 'git [a-z-]+[^\`]*origin/(main\|develop\|HEAD)' .moai/specs --include='*.md' \| grep -vc SPEC-MOVING-REF-GUARD-001` |
| Corpus filter 3 | 53 | **53** | filter 2 → `grep -ciE 'byte-unchanged\|byte unchanged\|unchanged\|preserv\|보존\|no diff\|empty\|0 files\|부재\|absent\|변경 ?없\|그대로'` |
| Corpus filter 4 | 42 | **42** | filter 3 → `grep -cvE '\b[0-9a-f]{7,40}\b'` |
| §B.6 unpinned divergence lines | 117 | **117** | `grep -rn 'rev-list --count --left-right' .moai/specs --include='*.md' \| grep -v SPEC-MOVING-REF-GUARD-001 \| grep -cvE '\b[0-9a-f]{7,40}\b'` |
| §B.6 of those carrying the fetch verb | 76 | **76** | same, then `grep -c 'git fetch'` |
| AC-MRG-014 all-matching-mutant over-fire | 495 | **495** | `grep -rnE 'origin/(main\|develop\|HEAD)' .moai/specs --include='*.md' \| grep -v SPEC-MOVING-REF-GUARD-001 \| grep -E 'git [a-z-]+' \| grep -cvE '\b[0-9a-f]{7,40}\b'` |

**Six of six reproduce exactly**, including the two figures (527, 53) my own iter-1 reconstruction
missed. D10 was the right diagnosis: the divergence was never about the tree, it was about an
alternation that existed only as a description. Recording it verbatim made the numbers
reconcilable, and they reconcile. The 117/76 re-targeting is also the better measurement — the
divergence class is where this bypass actually lands, and 65% is worse than my 36% estimate.

## D1, walked independently

Not accepted from §D.3's trace. I applied the v0.3.0 procedure to instance 3 myself, and then
tried to break it.

**The founding case reaches R4 on its own.** The line `base: origin/develop 44095ddc2` meaning
"the tip of develop you will start from":

- **Test 1** — substitute today's tip, then read *as a later reader will act on it*. A reader
  entering next week gets that frozen SHA rather than the then-current tip, so the meaning
  decays. Test 1's SUBJECT branch now names this case explicitly ("a meaning that is correct at
  the instant of substitution and decays afterwards"). → **SUBJECT**, without importing any
  consideration the test's own text lacks. This is the exact defect D1 named, and it is fixed at
  the text.
- **Test 3** — re-measured next week gives a different tip and that variance is the point →
  **SUBJECT**.
- **Tie-break** — not needed, the tests agree. Had Test 1 been read at substitution time it
  would have returned ANCHOR, and the tie-break would then route to Test 4 rather than to
  ANCHOR. **Both roads terminate at R4**, which is what "the procedure now reaches R4" has to
  mean.
- **Test 4** — a later reader must measure to obtain their base → **S2 → R4**.

**Attempts to break it, reported as negative results.**

1. *Does the read-time Test 1 now over-return SUBJECT for genuine anchors?* Applied to
   `AC-TST-010` (`git diff origin/main -- .moai/specs/… → empty`, deciding no predecessor SPEC
   body was modified): substituting a fixed SHA makes the claim stable and re-checkable, and it
   stays so when acted on later → ANCHOR. Test 3 agrees. No disagreement, no misroute. The
   near-miss reading — "but the original measurement was taken against an *earlier* tip, so
   today's value is the wrong pin" — is about which value to substitute, not about whether the
   meaning decays after substitution; Test 1 asks the latter, and the stale-anchor case is
   Test 2's, which routes it to §D.2 as a spurious red. The fix does not over-fire here.
2. *Can Tests 1 and 3 disagree in the direction (Test1 = ANCHOR, Test3 = SUBJECT)?* That
   requires "pinning preserves the meaning" **and** "the variance is the point" to hold together,
   which are close to contradictory — and the D1 fix is precisely what routes that combination
   to SUBJECT. I could not construct one.
3. *The direction (Test1 = SUBJECT, Test3 = ANCHOR)?* Test 3's SUBJECT branch requires both a
   different answer and that the variance be the point, so Test 3 can only return SUBJECT for
   subject claims. I could not construct a case where a stable-answer claim has a decaying
   meaning.

**The residual is real and the SPEC names it.** Test 4 has **no exit back to ANCHOR** — once the
tie-break routes a disagreement there, the outcome is S1 or S2, never ANCHOR, so a false
disagreement is unrecoverable. That is exactly what Q4 says ("a deferral rather than a decision
procedure"), and Q4 is now in the DoD's disposition list. Shipping that at plan-phase is
acceptable: no such case has been found by the author or by me, the deferral is written down
rather than assumed, and the DoD forces a recorded disposition before close. I would not accept
it as an unstated assumption; as a named open question with a disposition obligation, it is
honest scoping.

**D1 is closed.**

## D11, attacked again

**The named bypass is closed, and CM-2 does more work than the fix claims for it.** Under a
fetch-verb-keyed exclusion: CM-2's line (`git fetch origin main && git rev-list --count
--left-right origin/main...HEAD` → `0 0`) is exempted and goes red, while AC-MRG-001, AC-MRG-006
and CM-1 stay green — exactly the green-to-red pattern the criterion demands. Verified against
the fixture texts.

**Beyond the fetch verb, I tried the next exclusion shapes.** CM-1 and CM-2 together cover more
than "fetch":

- *"a git verb before the ref and a parenthesized value after it"* — exempts CM-1
  (`… internal/ (unchanged)`) → CM-1 red → caught by CM-1, as before.
- *"a command verb plus a parenthesized value"* — exempts CM-2 (`… → 0 0 (no divergence)`) →
  CM-2 red → **caught by CM-2**. This shape is the natural over-broad reading of "imperative
  structure with a demoted value", and CM-1 alone does not catch it. So CM-2 generalizes past
  the token key it was written for.

**Is imperative structure less forgeable than a token?** Partly, and the SPEC does not overclaim.
Forging it requires writing an instruction to measure, which is R4's actual content and carries
§D.4's price — but the price is enforced by review, not by the detector, and the SPEC says so in
L6's second residual ("An author determined to silence a line can still do it"). A shape key
remains forgeable in principle; what changed is that forging it now costs a plausible deciding
command. That is a bounded, stated residual rather than a closed hole presented as closed.

**What I did find on this surface is D13, below** — and it lands harder than any exclusion shape,
because the criterion that is supposed to prove the R4 exclusion exists cannot fail.

## D12, checked rather than accepted

The four-row pricing table is present, and the "count does the work" claim is defensible but
narrower than it reads: the count argument defeats *ignorance* (a reader who knows one remedy
pins everything), not *deliberate silencing* (an author takes the cheapest door regardless of how
many there are). The SPEC covers the second population separately, in L6's second residual, which
states plainly that R4's price is review-enforced and that a determined author can still silence
a line.

One thing worth the lead's eye, stated as an observation rather than a defect: R1 and R2's costs
are **detector-checkable** (a SHA or a baseline variable must be present on the line), R3's is
partly so (non-empty reason), and **R4's is not at all** — the detector reads shape, and shape is
what is being forged. So R4 is strictly the cheapest remedy at the layer where an author trying to
silence a warning actually meets resistance. §D.4 does not say this in those terms; L6 says it in
substance. Since both halves are written down and the residual is accepted with a reason, I score
this closed rather than reopening it.

**D12 is closed.**

## Every acceptance criterion, falsifiability named

| AC | Input that makes it fail | Falsifiable? |
|---|---|---|
| 001 | rule unregistered in `l.rules`, or the two-dot pattern missed; mutation `origin/main` → `origin/mainx` drops the count to 0 | **Yes** |
| 002 | delete the hex-SHA exclusion → the pinned fixture is flagged | **Yes** |
| 003 | remove the non-empty-reason check → fixture (c) reports zero | **Yes** |
| 004 | add a `strings.Contains(line, "...")` early return → the three-dot fixture stops firing | **Yes** |
| 005 | emit at `SeverityError` → the non-strict exit becomes non-zero. Depends on the §A era precondition, which is now stated (D3) | **Yes**, precondition stated |
| 006 | append a resolved SHA to the divergence fixture → the finding must disappear | **Yes** (SHOULD-tier) |
| 007 | delete a test name, or delete Test 4, from the doctrine → the grep count drops | **Yes** (document-grep, decidable) |
| 008 | shorten the message to name only pinning; or drop only R4 → the string assertion fails | **Yes** |
| 009 | restrict the rule to `doc.Body` → fails while 001 still passes | **Yes** |
| 010 | pin one occurrence in another SPEC directory **and commit it** → the diff surface is non-empty | **Yes** (D4 fixed) |
| 011 | delete the L1 or L6 paragraph → `grep -c 'L[1-6] —'` drops to 5 | **Yes** |
| 012 | copy the doctrine verbatim into the template → the neutrality guard reports the SPEC-ID and SHA tokens | **Yes** |
| **013** | **stated mutation is "remove the R4-form exclusion" — which does not turn it red** | **NO — see D13** |
| 013 CM-1 | positional exclusion → CM-1 goes dark | **Yes** |
| 013 CM-2 | any command-token-keyed exclusion → CM-2 goes dark while 001/006/CM-1 stay green | **Yes** |
| 014 | delete the claim-marker conjunct → the fixture is flagged | **Yes**, and the over-fire is measured (495) |

**On "SPECIFIED falsifiable, not OBSERVED falsifiable" (§E.1 and the dispatch's question 3).**
That scoping is honest, and it is the correct scoping at plan-phase: no Go code exists, so no
mutation can have been planted, and `progress.md` §E.1 lists this explicitly under Gaps rather
than implying otherwise. The DoD carries the obligation forward in binding form — "Every
criterion's stated mutation was actually planted, observed to turn the criterion red, and
reverted — recorded in `progress.md` §E.2 with the verbatim failing output" — plus two named
non-acceptance clauses for CM-2 and AC-MRG-014. A criterion whose falsifying input is *named and
mechanically checkable* is a criterion that can be checked; the checking is run-phase work. That
is not a criterion that cannot be checked.

The exception is D13, where the named input is checkable **and I checked it, and it does not
falsify**. That is a different thing from "not yet observed", and it is why D13 is blocking.

## Defects Found

**D13 — AC-MRG-013's fixture is unflaggable for three independent reasons, so the R4 exemption — the [HARD] surface — has no criterion that exercises it — `acceptance.md`: AC-MRG-013 — Severity: critical — Class: blocking**

The fixture line is:

```
base: measure at entry with git fetch origin develop (dispatch-time reference value: 44095ddc2)
```

The stated mutation is *"remove the R4-form exclusion from the rule; the finding reappears."*
It does not reappear. Tested mechanically against each of REQ-MRG-001's conjuncts:

```
grep -cE 'origin/(main|develop|HEAD)'  → 0     # no moving-ref token at all
grep -oE '\b[0-9a-f]{7,40}\b'          → 44095ddc2   # carries a SHA (9 hex chars, in range)
grep -ciE '<the recorded claim-marker alternation>' → 0   # no invariant-claim marker
```

- The line says `git fetch origin develop` — **two arguments**, not the token `origin/develop`.
  REQ-MRG-001 enumerates the slash forms only, and **§F L1 states outright** that refs without an
  `origin/` token are invisible. The fixture is an L1 blind-spot line.
- `44095ddc2` is a 7-40 hex run, so REQ-MRG-008's SHA-pin exclusion already exempts it.
- It carries no claim marker, so REQ-MRG-001's third conjunct is unsatisfied.

Any one of the three makes the line unflaggable. **AC-MRG-013 therefore passes whether or not the
R4 exclusion exists**, and its PASS is delivered by a stated detection limit rather than by the
mechanism it names. That is the §A [HARD] contract violated on the SPEC's own most delicate
surface: "A guard whose acceptance criterion cannot fail is indistinguishable from a guard that
is switched off."

This does not weaken D11's finding or CM-1/CM-2 — those are separately sound and I verified them
above. What it means is narrower and worse: the suite now protects well against an R4 exclusion
that is *too wide*, and has nothing at all that proves an R4 exclusion *exists and works*.

**A second, related gap the fixture choice reveals: REQ-MRG-010's residual scope was never
characterized.** Once the other exclusions are applied, most well-formed R4 lines are already
exempt without it:

- On the REQ-MRG-006 (divergence) class, R4's form demotes the value to a **dated** reference
  (§D.2 R4: "parenthesized and dated"), and REQ-MRG-006 fires only on a line carrying "neither a
  SHA nor a date". Every well-formed R4 divergence line is therefore **already exempt by the date
  conjunct** — the R4 exclusion has nothing to do there.
- On the REQ-MRG-001 class the residual is non-empty: a line with an `origin/` token, a claim
  marker, and a dated-but-not-SHA reference value would still fire and does need REQ-MRG-010.
  Example shape: `run git diff --name-only origin/main -- internal/ at entry (reference: empty
  as of 2026-08-28)`.

So the class REQ-MRG-010 must actually cover is narrow and specific, and Q0 — the least-settled
question, driving M3's most delicate work — is being asked without that class having been named.
Naming it is cheap now and expensive after M3.

**Required fix (two lines of work, both plan-phase):**
1. Replace AC-MRG-013's fixture with a line from the REQ-MRG-001 residual class — carrying an
   `origin/<branch>` slash token, a claim marker, **no** 7-40 hex run — so that removing the R4
   exclusion genuinely turns the criterion red. Re-check CM-1 and CM-2 against the new fixture
   (both remain valid as written; they do not depend on it).
2. State REQ-MRG-010's residual scope in §D.2 or §H Q0: which lines the R4 exclusion must exempt
   *that the SHA-pin, baseline-variable, date and claim-marker conjuncts do not already exempt*.

**Mitigating fact, stated so this is not read as worse than it is:** the DoD already contains the
mechanism that would catch this at run-phase — "every criterion's stated mutation was actually
planted, observed to turn the criterion red". Planting "remove the R4-form exclusion" and
observing AC-MRG-013 stay green is precisely that obligation discharging. D13 is therefore a
defect the SPEC's own process would surface later; the finding is that it can be fixed now for
free instead.

**D14 — the SPEC states an obligation on M1's doctrine that no criterion decides — `spec.md`:§D.3 vs `acceptance.md`:§C — Severity: major — Class: blocking**

§D.3 carries the heading **"The skew is also a limitation, and the doctrine must publish it as
one"**, and the paragraph is the direct remediation of iter-1's judgment call 3. But nothing
decides whether M1's doctrine actually publishes it:

```
grep -n 'ANCHOR' acceptance.md   → (no matches)
```

- AC-MRG-007's decider greps for the four test names, the four branch labels `R1`-`R4`, and the
  five instance identifiers. The validation-gap statement is none of those.
- AC-MRG-011's decider is `grep -c 'L[1-6] —' → 6`. The gap is stated in §D.3, not as a limit in
  §F, so it is outside that count too.
- REQ-MRG-005 enumerates what the doctrine shall state — predicate, instances, remedies, limits —
  and does not include it.

So the one limitation this audit round surfaced, and which the SPEC itself says the doctrine must
carry, is the one thing M1 could omit while passing every criterion. That is the same shape as
D13 arriving from the other direction: there, a criterion with no failing input; here, an
obligation with no criterion.

**Required fix:** either add the ANCHOR-branch validation gap to §F as **L7** (which pulls it into
AC-MRG-011 automatically, with the grep count raised 6 → 7 and the second-mutation clause
extended), or extend REQ-MRG-005 and AC-MRG-007's decider to name it explicitly. L7 is the
cheaper of the two and matches how L6 was handled at v0.2.0.

**D15 — Test 4's gate is literally unsatisfiable for a claim that reads true — `spec.md`:§D.1 Test 4 — Severity: minor — Class: optional**

Test 4 "Runs when Tests 1-3 return SUBJECT". Test 2 is explicitly conditional — "Applies when the
claim currently reads false" — and its outcomes are "true signal" / "spurious red", never SUBJECT.
So for a claim that reads true, Test 2 returns nothing and the gate as written is never satisfied.
§D.3's trace resolves this by skipping Test 2 entirely, which is the natural reading, but the gate
text does not authorize the skip. Given how much weight §D.1 now carries as published doctrine, a
literal-minded reader is a plausible reader.

**Required fix (optional):** "Runs when the applicable tests among 1-3 return SUBJECT", or simply
"Runs when Tests 1 and 3 return SUBJECT".

**D16 — §B.3 attributes the three-way triage taxonomy to the predicate, which produces two classes — `spec.md`:114 — Severity: minor — Class: optional**

§B.3 reads "it contains all **three dispositions the predicate produces**". The predicate produces
two classes and four remedies; the three-way split (anchor / subject / already-frozen) is the
**triage** taxonomy AC-MRG-010 requires, not the predicate's output. `plan.md:24` was corrected for
exactly this at D7 ("both predicate classes and all four remedies"); this line is the same residue
one file over. It matters slightly more than a typo because §D.1 makes a point of the class-vs-
remedy distinction and names conflating them as a route to the dominant failure mode.

**Required fix (optional):** "all three triage dispositions", or "both predicate classes and the
already-frozen case".

## What I verified and found sound

Stated so the findings above are not read as the whole picture.

- **Counts are exactly as claimed.** 11 REQ · 14 AC · 4 tests (`grep -c '^\*\*Test [0-9]'`) ·
  4 remedies (`grep -c '^| \*\*R[0-9]\*\*'`) · 6 limits (`grep -c '^- \*\*L[0-9]'`) ·
  5 instances (`grep -c '^\*\*Instance '`) · Q0-Q4 · zero `Where`. Every figure the lead listed
  reproduces.
- **No orphans in either direction.** All 11 REQs have a covering AC; all 14 ACs trace to a REQ;
  all 14 ACs are assigned to exactly one milestone and the union is 001-014.
- **AC-MRG-014 and CM-2 are a genuinely well-made pair.** They differ by exactly one thing — CM-2
  cites a result (`→ 0 0 (no divergence)`), AC-MRG-014's fixture cites none — which is precisely
  the claim-marker conjunct under test. Isolating a conjunct by a one-token fixture delta is the
  right construction, and it is what makes both criteria decisive rather than suggestive.
- **The re-measurement is better evidence than the audit figure it replaced.** Re-targeting from
  the diff-shaped class (52/6) to the divergence class (76/117) is the correct move — the bypass
  lands on the class REQ-MRG-006 covers — and it makes the author's own finding worse than the
  auditor's estimate. Recording that rather than taking the softer number is the behaviour this
  lane's doctrine asks for.
- **D6's handling is the right form.** The version miss is recorded in the HISTORY row as a miss
  ("bumped here to 0.3.0 with the miss recorded rather than silently corrected") rather than
  quietly fixed. In a SPEC about documented-vs-actual divergence, that is the consistent choice.
- **P1 did not recur.** One writer, quiescent tree, identical hashes at entry and verdict.
- **The Gaps section is honest and load-bearing.** `progress.md` §E.1's v0.3.0 Gaps name the three
  things that were not observed — no planted mutations, the bypass reasoned rather than executed,
  and imperative-structure recognition untested — and none of them is contradicted by a criterion
  claiming coverage.
- **Tier M remains correct.** One rule file, one test file, fixtures, one doctrine section, one
  template mirror. 11 REQ / 14 AC sit inside the Tier M 16/16 budget.

## Gaps — what this audit did not observe

- **No Go code was built or run.** No mutation was planted; no rule output was produced. Every
  statement above about what a rule "would" flag is derived from REQ-MRG-001/006/008/010's stated
  conjuncts applied to fixture text, not from an executed detector. D13 is established at that
  level — the fixture fails three conjuncts as written — and would need the run-phase mutation to
  be *observed* falsifying rather than *shown* non-falsifying.
- **No `go test`, no `go vet`, nothing under `e2e/`** — per the dispatch constraints.
- **The doctrine file does not exist yet**, so AC-MRG-007's and AC-MRG-011's greps were reasoned
  against §D.1/§D.2/§D.3/§F rather than run against M1's output. D14 is established from
  `acceptance.md`'s own text (zero `ANCHOR` matches), which does not depend on the doctrine
  existing.
- **The corpus figures are grep prevalence, not rule output** — the SPEC states this itself, and
  no criterion here assumes the two agree.
- **Q4's undecidable region was probed, not exhausted.** I could not construct an ANCHOR case that
  trips the tie-break; that is a failed search, not a proof that none exists.

## Recommendation

**Eight of eight iter-1 blocking defects are genuinely closed, and I verified each one
individually rather than accepting the summary.** D1's fix reaches R4 on its own, by both routes,
and survived three attempts to break it. D11's fix closes the fetch-verb bypass and CM-2
generalizes past the shape it was written for. D2's negative control is real and its mutant is
measured. D12's pricing is stated with its enforcement layer named honestly. Every figure in the
SPEC reproduces exactly in this tree, including the two that did not reconcile at iter-1 — D10 was
the right diagnosis and its fix worked. The score movement is monotonic, 0.80 → 0.86, and no
must-pass criterion failed.

**Two new blocking defects exist, and I would be misreporting if I said otherwise.** Both were
found by attacking the surfaces the dispatch pointed at, and neither is a restatement of anything
from iter-1:

1. **D13 (critical) before M3.** AC-MRG-013's fixture is unflaggable three ways over, so the
   criterion cannot fail and the R4 exemption has nothing proving it exists. This is the §A [HARD]
   contract broken on the card's own [HARD] deliverable, and the fixture choice also reveals that
   REQ-MRG-010's residual scope — after the SHA, date, and marker conjuncts do their work — has
   never been characterized. Both halves are plan-phase edits: swap the fixture, name the class.
2. **D14 (major) before M1.** The ANCHOR-branch validation gap is the one thing this audit round
   added to the doctrine's obligations, and it is the one thing M1 can omit while passing every
   criterion. Adding it as L7 pulls it into AC-MRG-011 for free.

**D15 and D16 are optional** — a gate-wording literalism and a residual count attribution.
Surfaced for the orchestrator's discretion; I am not routing them into a revision, and neither
affects the verdict.

**Verdict rationale.** No must-pass criterion failed (MP-1/2/3/5/7 PASS, MP-4/6 N/A). The
aggregate of 0.86 clears the Tier M threshold of 0.80 with margin, unlike iter-1's zero-margin
0.80. The two new defects are both one-edit repairs on surfaces whose milestones have not started,
and D13 in particular would be caught by the SPEC's own DoD at run-phase — the finding is that it
is free to fix now and expensive to discover after M3 has written the exclusion.

**This is the closing iteration, so the debt cannot be discharged by another audit round.** My
recommendation to the orchestrator: apply D13 and D14 as a targeted pre-run edit — a fixture
swap, a scope sentence, and an L7 paragraph, no plan-phase re-entry — and enter run-phase with
them closed. If the orchestrator prefers to carry them instead, they must be written into the DoD
as named pre-M3 and pre-M1 obligations rather than left in this report, because a report is not a
gate.

The SPEC is in materially better shape than at iter-1 and is well above the standard this lane
usually reaches. It is a PASS-WITH-DEBT, and the debt is small, named, and cheap.
