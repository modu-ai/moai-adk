# SPEC Review Report: SPEC-IGNORED-EVIDENCE-CITATION-001

Iteration: 4 — second operator-override delta iteration
Granted scope: **P1-P5**. P6 optional/untouched; P7 excluded by operator decision, carried as
recorded residual risk.
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.89** (harmonic mean) — Tier M PASS threshold **0.80**
Score movement: **0.74 → 0.78 → 0.83 → 0.89**. Monotonic across four iterations. **No dimension
regressed in this round**, which was the specific risk of a subtraction-only pass.

Reasoning context ignored per M1 Context Isolation. Every claim below was re-executed in this tree.

Audit tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t381`, HEAD `3f03d9c36`;
`origin/develop` `9328a5242`, `11 0` ahead.

---

## Must-Pass Results — the firewall is cleared

- **[PASS] MP-1 REQ number consistency** — `grep -n '^### REQ-IEC-'` returns **001, 002, 003, 004,
  005, 006, 007, 008, 009, 010** at spec.md:165, 173, 181, 187, 194, 202, 211, 235, 240, 258.
  Contiguous, no gap, no duplicate, uniform padding. **The iter3 firewall failure is closed.**

  *On the surviving `REQ-IEC-011` string* — the coordinator asked me to rule on this explicitly. It
  occurs once, at spec.md:25, and reads: "numbered REQ-IEC-010 and REQ-IEC-011 **at the time**,
  renumbered to REQ-IEC-009 and REQ-IEC-010 in iter4." **Definitions is the correct reading**, and
  not merely because it is convenient here: MP-1 constrains the numbers *assigned to requirements*,
  and a HISTORY row narrating a renumber assigns nothing. Reading it as any string occurrence would
  make it impossible to ever document a renumber without failing the firewall — it would penalise
  exactly the transparency that closes the defect. The occurrence is also explicitly time-marked,
  which is the strongest form of the distinction.
- **[PASS] MP-2 GEARS format compliance** — renumbering changed ids, not forms. All ten requirements
  remain GEARS-conformant.
- **[PASS] MP-3 YAML frontmatter validity** — 12 canonical fields, `version: "0.4.0"` (spec.md:4),
  no rejected alias.
- **[N/A] MP-4 Section 22 language neutrality** — single-language scope.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `SPEC-EVIDENCE-CITATION-CANON-001` → `completed`;
  `SPEC-HIERARCHICAL-TEAM-001` → `completed`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall` = 0.
- **[PASS] MP-7 clarification gate** — rc=1, no markers.

**All seven pass.** Verdict is therefore driven by the rubric, which clears at 0.89.

---

## The granted delta

### P1 — closed. And the author's correction of my edit count is right.

Definitions are contiguous (above). On the count: I estimated "six edits"; the author counted 18
occurrence-lines across four artifacts and said my figure omitted spec.md §C/§D/HISTORY and both
plan.md and progress.md. **The author is correct and I was wrong.** My iter3 figure came from the
narrow question "what does it take to renumber §B and the acceptance references", not from the actual
occurrence set — and understating the cost of a fix is how a fix gets applied partially. That the
author refused the estimate and measured is the right instinct.

**On leaving the three audit reports unrewritten — correct, and it is the only defensible call.**
An audit report is a record of what was true when the audit ran. Rewriting `REQ-IEC-009` to
`REQ-IEC-008` inside iter1-iter3 would make those reports assert findings against a numbering that
did not exist at the time, which falsifies the audit trail to preserve a cosmetic consistency. The
reports are dated and iteration-labelled, so a reader has the frame they need. **The general rule:
renumber the live artifact, never the record of a past judgment.**

### P2 — closed, and the rebinding claim is verified

The mechanism is real, and I confirmed it end to end:

- At iter3 the two dangling references (then spec.md:417, :446) named `REQ-IEC-008` meaning the
  demoted **probe-boundary** requirement.
- After the iter4 renumber, `REQ-IEC-008` at spec.md:235 is **Collision avoidance** — a live
  requirement, covered by AC-IEC-007.
- So a renumber-only edit would have left both references pointing at a real requirement with the
  wrong meaning. **A dangling reference announces itself; a rebound one does not.** The claim holds.

The fix is right: both sites now reference the constraint **by name**, immune to future renumbering —
spec.md:426 `### C.7 The census probe and its blind spots (§D probe-boundary constraint)` and
spec.md:455 "That record is why the §D probe-boundary constraint exists."

**I swept for other reassignments taking the same path, as asked.** Three ids moved
(009→008, 010→009, 011→010), so old text saying `009` or `010` is at risk of silent rebinding. In the
*specification layer* every surviving reference is correct — I checked all of them:

| Reference | Names | Correct? |
|---|---|---|
| acceptance.md:91, :251 — AC-IEC-007 → REQ-IEC-008 | Collision avoidance | ✓ |
| acceptance.md:92, :278 — AC-IEC-010 → REQ-IEC-007 | Self-consistency | ✓ |
| acceptance.md:93, :324 — AC-IEC-011 → REQ-IEC-010 | Behavior preservation | ✓ |
| acceptance.md:94, :343 — AC-IEC-012 → REQ-IEC-009 | Stale coordinates | ✓ |
| spec.md:493 — carve-outs / t375 files "excluded by REQ-IEC-006 and REQ-IEC-008" | 006 Carve-out, 008 Collision | ✓ |
| spec.md:499 — "REQ-IEC-010, which AC-IEC-011 gates" | Behavior preservation | ✓ |
| spec.md:501, :559 | marked historical | ✓ |

All ten requirements map to exactly one MUST criterion; no criterion names a non-existent
requirement. **The specification layer survived the renumber intact.** The exceptions are in
`progress.md` — F1 below.

### P3 — closed, and I ran the full seven-path set you did not

```
$ git diff --exit-code --stat origin/develop...HEAD -- <all 7 AC-IEC-007 paths>   → exit=0
$ git diff --exit-code --stat origin/develop...HEAD -- <all 8 AC-IEC-006 paths>   → exit=0
```

Both hold on the complete sets the criteria actually name, not just the four you sampled. The
three-dot form diffs `merge-base(origin/develop, HEAD)...HEAD`, so after the branch absorbs
`origin/develop` the merge base advances to `9328a5242` and the comparison isolates this card's own
edits. That is exactly the property iter3's frozen `3f03d9c36` lacked, and it removes a scheduled
integration failure from two MUST criteria.

### P4 — closed; the remaining "will add" really is historical

spec.md:319 and :343 now read "the landed rule-body sentence at `manager-lead.md:150` (quoted in
§A.4)". `grep -n 'will add'` returns exactly three lines, and I read all three: **:27** (HISTORY row
recording the fix), **:130** (§A.4 quoting the retired framing — `iter1/iter2 described this as a
sentence t375 "will add"`), **:557** (§E.2's table row recording the occurrence). All three are
quotation or record. **This is not a fourth recurrence** — your reading is correct.

### P5 — closed

plan.md §H: the truncated bullet is restored ("the canon, the treatments, the do-not-touch list."),
the contradictory `draft, unmerged` duplicate is gone, one t375 entry remains reading `status:
completed`, landed, read at `9328a5242`, and "12-AC" is now "ten MUST criteria plus two §D structural
checks". HISTORY ordering is also correct now (0.1.0 → 0.2.0 → 0.3.0 → 0.4.0) — a defect the author
found and fixed unprompted.

### §E.1 — actionable, not decoration

You asked whether a follow-up card could act on it without re-deriving. It can. It carries the
requirement text location, the criterion location, the guard constraint that blocks the obvious fix,
the D2 history explaining how the gap was created, an explicit statement of what is *not* closed, a
bounded live-risk assessment, and one genuinely load-bearing mechanical insight: **`grep -c` counts
lines, so a same-line citation-plus-marker test is expressible as a single pattern while a ±N-line
window is not.** That last sentence is the difference between a follow-up card starting from the
analysis and starting from scratch.

§E.1 also states the honest reason it was not closed here — both resolutions touch requirement
substance, which the delta grant excludes. That is the correct handling of an operator exclusion.

---

## Defects

**F1 — the silent-rebind fix was applied to `spec.md` and NOT to `progress.md`; two references
re-bound there.** This is the fourth instance of the half-repair shape, in the round that named it,
inside the file that records the naming.

- **`progress.md`:100-101** — "neither creates coverage for **REQ-IEC-008**, which appears only in
  `spec.md`." In this iter3 narrative `REQ-IEC-008` meant the probe-boundary requirement. It now
  names **Collision avoidance**, which *is* covered (AC-IEC-007) and does *not* appear only in
  spec.md (acceptance.md:91, :251). **Both clauses are now false** — the exact shape §E.2 defines,
  undetectable by grep because the id resolves.
- **`progress.md`:61-62** — records the iter2 additions as "**REQ-IEC-009** (stale evidence
  coordinates) and **REQ-IEC-010** (behavior preservation)". At iter2 those were 010 and 011. The
  names happen to match the current ids so the sentence is not false today, but a past event is
  silently restated in present numbering — inconsistent with spec.md:25-26, which flags the
  reassignment explicitly. Same class, lower severity.

Severity: **major**. Class: **blocking** (for a follow-up round, not for run-phase — see below).
*Required fix*: at progress.md:101 replace the id with "the probe-boundary requirement (then
numbered `REQ-IEC-008`)"; at progress.md:61-62 add the same time-marking spec.md:25 uses. Then re-run
`grep -rn 'REQ-IEC-00[89]\|REQ-IEC-01[01]' progress.md` and read each hit for meaning, not existence.

**F2 — `progress.md`:103 still asserts the iter3 justification that iter4 overturned.** It reads
"REQ ids 009/010/011 keep their numbers so every AC cross-reference stays valid." P1 removed those
numbers, and spec.md:230-231 records that this ground "did not survive measurement". `REQ-IEC-011`
no longer exists. Severity: **minor**. Class: **blocking**. *Fix*: one sentence, marking it as the
iter3 position superseded by the iter4 renumber.

**F3 — `progress.md`:17 carries a stale requirement range.** "Requirements: REQ-IEC-001..009" — true
when written at iter1, now 001..010, and unmarked as historical. Severity: **minor**. Class:
**optional**. *Fix*: fold into F1's sweep.

**P6 (carried, unchanged, optional)** — spec.md:389/391/395 still frame t375 as a live parallel lane
("files owned by t375 (lane-8)", "lane-8 **edits** … as a pair", "t375 **creates** it"). t375 has
landed; REQ-IEC-008's stated rationale — a write race between two lanes — no longer describes
reality, though the exclusion remains correct as scope discipline. Explicitly untouched by operator
scope. Severity: **minor**. Class: **optional**.

**P7 (carried by operator decision)** — the adjacency gap, now properly recorded at §E.1 rather than
silently carried. Severity: **minor**. Class: **optional** — converted from a defect to documented
residual risk, which is the correct disposition for an excluded item.

**Note on F1-F3's blocking classification.** All three sit in `progress.md`'s narrative record. No
milestone, criterion, or requirement depends on them, so none blocks run-phase entry. They are marked
blocking because a card whose subject is *tracked statements that are false about their own
provenance* should not ship two of them in its own progress record — and because F1 is the precise
failure §E.2 was written to prevent.

---

## Category Scores

| Dimension | i1 | i2 | i3 | **i4** | Band | Movement |
|---|---|---|---|---|---|---|
| Clarity | 0.85 | 0.85 | 0.75 | **0.85** | 0.75 | Recovered. P4 (§C.2/§C.3), P5 (§H contradiction + truncation), P2 (dangling refs) all closed; §E adds a coherent residual-risk layer. Held off 1.0 by F1/F2 in progress.md and P6's stale framing. |
| Completeness | 0.95 | 0.90 | 0.90 | **0.95** | 1.0 | §H restored and corrected; §E.1/E.2/E.3 added; HISTORY ordering repaired; five `### Out of Scope` headings intact. |
| Testability | 0.55 | 0.65 | 0.90 | **0.92** | 0.75-1.0 | P3 converted two MUST criteria from "scheduled to fail at integration" to correct, verified on the full 7- and 8-path sets. No criterion weakened. |
| Traceability | 0.75 | 0.75 | 0.80 | **0.85** | 0.75-1.0 | Numbering contiguous; 10 requirements ↔ 10 criteria verified one-by-one after the renumber; dangling references retargeted by name rather than by id. Held off higher by F1's rebinds and the carried P7 under-coverage. |

Harmonic mean of {0.85, 0.95, 0.92, 0.85} = **0.8904 → 0.89**. Threshold 0.80.
**No dimension regressed** — the specific hazard of a round whose purpose was subtraction.

---

## Judgment on the repeated metadata deviation

**The second instance still qualifies — and your instinct that repetition means something is right,
but the conclusion points at the grant, not at the author.**

The rule I supplied at iter3 was: *would deferring create a fresh inaccuracy?* Applied here, iter4
rewrote §B ids across four artifacts, retargeted references, replaced two baselines, and added §E. A
frontmatter reading `0.3.0` after that would be false. So the rule fires, and the edit is bounded
exactly as the narrow exception requires: frontmatter and HISTORY rows only, describing edits already
granted, licensing no new requirement, criterion, or scope. I verified that — the requirement count
change is explained entirely by P1's renumber, and no criterion changed outside P3.

But a deviation declared every round is evidence of a **mis-shaped grant**, not of drift. Each
declaration costs an operator decision and re-litigates a boundary already settled, and a boundary
re-argued repeatedly is a boundary that will eventually be argued wider. The fix belongs upstream:

> **Fold "version bump plus a HISTORY row covering the edits made in this iteration" into the
> standing delta-grant template**, so it needs no declaration and no per-round judgment.

That converts a recurring exception into a stated term, which is where it should have been after
iter3. Making the author declare it a third time would be process theatre.

Two things to credit: the deviation was disclosed before judgment both times, and the author found
and fixed an out-of-order HISTORY sequence while there — a defect no auditor had raised. Self-caught
defects are the signal that the discipline is internalised rather than performed for the audit.

---

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria pass, the score clears the Tier M threshold by 0.09,
and the entire gating layer — requirements, criteria, milestones, baselines — is now verified clean.
Every defect I found is confined to `progress.md`'s narrative or to items the operator explicitly
excluded.

Recorded debt, in priority order:

1. **F1** — the two `progress.md` rebinds. Highest priority not because of consequence but because of
   *shape*: it is the failure §E.2 names, surviving in the round that named it. Fix before run-phase
   so the card's own record does not contradict its analysis.
2. **F2** — one superseded sentence at progress.md:103.
3. **F3** — one stale range at progress.md:17.
4. **P6** — restate §C.5's rationale as scope discipline rather than lane-race avoidance.
5. **P7** — the adjacency gap, already properly recorded at §E.1; needs a follow-up card, and §E.1
   carries enough for one to start without re-deriving.

**Proceed to Implementation Kickoff Approval.** F1-F3 are single-sentence edits to a progress record
and do not gate run-phase entry; they should be swept in the same turn that opens M1, using the rule
the card itself wrote: *grep for every occurrence of the old fact, and read each hit for meaning
rather than existence.*

Do **not** open a fifth audit iteration for F1-F3. The verdict is PASS; re-auditing three narrative
sentences would spend an operator decision on less than the decision costs.

---

## Claims that did not hold when I ran them

**From the author** — one, and it is the finding of the round:

1. **"P2: retargeted the two dangling probe-boundary references"** (HISTORY 0.4.0) and §E.2's
   "This card produced the same shape three times". Both are true of `spec.md` and **false of
   `progress.md`**, where two references re-bound and were not swept (F1). The count is three
   occurrences *found*; the fourth is live. The rule §E.2 states — grep before reassigning — was
   formulated correctly and executed on one artifact out of four, and progress.md was edited in this
   round, so it was in hand at the time.

**From the coordinator** — everything held:

- **P1 closed, definitions contiguous.** Confirmed, and your reading of MP-1 as keying on definitions
  is the correct one; my ruling and reasoning are above.
- **P3's three-dot form.** Confirmed, and extended to the full seven-path AC-IEC-007 set and the
  eight-path AC-IEC-006 set you had not run — `exit=0` on both.
- **P4's remaining "will add" is historical.** Confirmed; all three surviving occurrences are
  quotation or record.
- **The rebinding claim.** Confirmed as a mechanism, and it is the most consequential finding of the
  round exactly as you judged — though its own fix is the thing that turned out to be incomplete.

**From my own iter3 report** — my "six edits" estimate for P1 was wrong; the measured figure is 18
occurrence-lines across four artifacts. Understating a fix's extent is how a fix gets applied
partially, which is, with some irony, precisely what F1 records happening.
