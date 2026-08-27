---
id: SPEC-MOVING-REF-GUARD-001
title: "Moving-ref invariant guard: warn on unpinned invariant claims, with the anchor-or-subject exemption predicate shipped alongside"
version: "0.1.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/spec, .claude/rules/moai/core"
lifecycle: spec-anchored
tags: "lint, verification, evidence, moving-ref, invariant, preserve, baseline-attribution"
era: V3R6
tier: M
related_specs: [SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-VERIFICATION-COMPLETENESS-001]
---

# SPEC-MOVING-REF-GUARD-001 — Moving-ref invariant guard

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.2.0 | 2026-08-28 | Lead addendum absorbed. **(1)** Grounded instance 3 re-attributed: the stale-base occurrence is the **lead's own dispatch block**, named as such rather than as an anonymous occurrence, and §B.4 now states that it sits on the same axis as the card's side-discipline — relaying a once-measured value without re-measuring is the form this card prohibits. The attribution was re-verified in this turn before being written (`git rev-parse ec15ec2cd^` → `44095ddc2…`; `git reflog show WT-moving-ref-guard` → created from `origin/develop` at `ec15ec2cd`), because citing the dispatched figure unverified would have made this SPEC an instance of its own defect. **(2)** Fifth grounded instance added (§B.5, §D.3): the lead's change to the dispatch base line, from `base: origin/develop <sha>` to `base: measure at entry … (dispatch-time reference value: <sha>)`. **(3)** The predicate is **kept two-way** at the top level but gains **Test 4**, which routes within the SUBJECT class, and the remedy space expands from three branches to **four** with the addition of R4 (state the measuring command as the criterion, demote the value to a dated reference). §D.1 now states explicitly that classification and remedy selection are two steps, because conflating them is a second, subtler route to the card's dominant failure mode. Why the new case does not become a third top-level branch, and why it does not fold into ANCHOR, is argued in §D.1.1. **(4)** New REQ-MRG-010 / AC-MRG-013: the R4 form must not be flagged. **(5)** New detection limit L6 — the detector cannot tell a live reference value from a rotted one — with AC-MRG-011's grep count raised from 5 to 6. | manager-spec |
| 0.1.0 | 2026-08-28 | Initial plan-phase authoring for card t342, in worktree `.claude/worktrees/t342` at HEAD `ec15ec2cd`. Every figure in §B was measured in this tree; the corpus survey (§B.3) and the two-dot/three-dot discriminator experiment (§B.2) were run here rather than carried over. The card's own dispatch supplied a third live instance of its subject (§B.4), verified against `git reflog show WT-moving-ref-guard`. The exemption predicate (§D.1) is specified as the primary deliverable per the card's [HARD] clause, and milestone ordering (plan.md §F) puts it first for that reason. | manager-spec |

## §A. Problem Statement

An invariant claim — "byte-unchanged", "PRESERVE", "this path is absent from the diff" — is
asserted by naming a comparison anchor and reporting what the comparison returned. When that anchor
is a **moving ref** (`origin/main`, `origin/develop`, `origin/HEAD`), the anchor's value changes
every time upstream commits. The claim's *text* does not change. Its *truth* does.

The failure is silent in the direction that matters. Nothing errors; the claim simply stops meaning
what it meant when it was written, and the next reader who re-runs it gets a different answer with
no signal that the anchor moved underneath them.

The card's measured instance: a `plan.md` recorded its PRESERVE proof as
`git diff --stat origin/main -- <path>` and the command returned `18 files changed / 2825 deletions`
**before any work had been done in that tree** — `origin/main` had advanced ten commits in the
interval between authoring and reading. The evidence offered for "every existing file is
byte-unchanged" was false on arrival, and false for a reason that had nothing to do with the work
it was describing.

This is a `verification-claim-integrity.md` §2 defect specifically: the claim is not attributed to
an actually-measured baseline, because the thing it names as its baseline is not a fixed value.

### What this SPEC does, and what it deliberately does not

It adds a **warning** that surfaces the shape — an invariant claim decided by a moving ref with no
pin — and it ships, in the same delivery, the **predicate that decides whether the warning should
be acted on by pinning at all**.

The predicate is not an accompaniment. It is the primary deliverable, and §D.1 exists because a
warning shipped without it has a well-understood failure mode, stated by the card: the next person
mechanically pins every occurrence. Pinning is right for most of them and **wrong** for the class
where the moving ref is the claim's *subject* rather than its anchor — a provenance narrative about
what mainline currently carries is weakened, not strengthened, by being frozen to one commit.

## §B. Measured Baseline

Every figure below was measured in `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342`, branch
`WT-moving-ref-guard`, HEAD `ec15ec2cd`, on 2026-08-28. Each row names the command that produced it.

### B.1 — The tree this was measured in

```
git rev-parse --show-toplevel  → /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t342
git branch --show-current      → WT-moving-ref-guard
git rev-parse --short HEAD     → ec15ec2cd
git status --short             → (empty)
```

No work had been performed in this tree at measurement time. That matters for §B.4: every non-empty
diff reported there is upstream drift and nothing else.

### B.2 — Three-dot is **not** a safe form (a tempting mitigation, disproved)

A plausible first reading is that `git diff A...B` (merge-base relative) is immune to upstream
advance, and that the guard could therefore exempt the three-dot form. Measured here, it is not.

Same claim ("`internal/hook` is unchanged in my work"), same tree, two anchor values — `44095ddc2`
(one integration stale) and `ec15ec2cd` (current, `== HEAD`):

| Command | Output | Claim reads |
|---|---|---|
| `git diff --stat ec15ec2cd -- internal/hook` | *(empty)* | true |
| `git diff --stat 44095ddc2 -- internal/hook` | `3 files changed, 288 insertions(+)` | **false** |
| `git diff --stat 44095ddc2...HEAD -- internal/hook` | `3 files changed, 288 insertions(+)` | **false** |

The three-dot form gave the identical wrong answer. The reason is that merge-base is stable only
while the upstream advance **diverges** from the branch; once the branch absorbs that advance — by
merge, rebase, or by being created from the newer tip — `merge-base(stale, HEAD)` moves to the stale
ref itself and three-dot degenerates to two-dot. Confirmed:
`git merge-base ec15ec2cd 44095ddc2 → 44095ddc2cc1c9fed2b3bd5ac946f48017988aba`.

Consequence for the detector: `...` MUST NOT be treated as an exemption (AC-MRG-004). Recording the
disproof here is deliberate — it is the mitigation a later reader is most likely to propose.

### B.3 — Corpus survey: how much of this shape exists today

Over `.moai/specs/**/*.md` in this tree:

| # | Filter | Count | Command |
|---|---|---|---|
| 1 | Lines naming a moving ref at all | **1477** | `grep -rnE 'origin/(main\|develop\|HEAD)' .moai/specs --include='*.md' \| wc -l` |
| 2 | (1) inside a `git <verb>` command context | **527** | (1) piped through `grep -E 'git [a-z-]+[^\`]*origin/…'` |
| 3 | (2) also carrying an invariant-claim marker | **53** | (2) piped through the claim-marker alternation |
| 4 | (3) with no 7-40 hex SHA on the line | **42** | (3) piped through `grep -v -E '\b[0-9a-f]{7,40}\b'` |

**42 is the raw candidate set, not 42 defects.** The set was enumerated and read; it contains all
three dispositions the predicate produces, which is the useful result:

- **Anchor, unpinned — genuine finding.** `AC-TST-010` (`SPEC-V3R6-TEST-REFACTOR-001/acceptance.md:35`,
  `git diff origin/main -- .moai/specs/…` deciding "no predecessor SPEC body modified"); `AC-AFS-012`
  (`SPEC-V3R6-AGENT-FOLDER-SPLIT-001/progress.md:126`, `git diff origin/main -- <9 files> | wc -l`
  deciding "0 (byte-identical)"); `AC-MSD-010`; `AC-WC-009`; `AC-CFS-010`; `AC-NS2-005`.
- **Subject, correctly unpinned — must not be flagged.** `AC-COORD-016`
  (`SPEC-V3R6-MULTI-SESSION-COORD-001/progress.md:296`) asserts that the *literal string*
  `git rev-list --count --left-right origin/main...HEAD` is preserved verbatim inside a document.
  The ref token is quoted subject matter, not a measurement anchor. `REQ-LB-006`
  (`SPEC-V3R5-LATE-BRANCH-001/spec.md:168`) documents `git reset --hard origin/main` as a
  doctrine step — again the command is the subject.
- **Already remediated by freezing, not pinning.** `SPEC-DESIGN-MOAIWEBV2-002/plan.md:36` records
  `BASELINE_SHA=$(git rev-parse origin/main)` captured *before* the first run-phase commit, with
  `acceptance.md:14` deciding its criteria against `$BASELINE_SHA`. This is a third remediation
  branch that pre-dates this SPEC and is promoted to the recommended default in §D.2 (R2).

### B.4 — The lead's dispatch reproduced this card's own subject (live, this tree)

**Attributed to the lead**, who owns the instance and asked for it to be recorded as theirs rather
than as an anonymous occurrence. The dispatch block named the base as `origin/develop 44095ddc2`.
The tree was created from a later tip:

```
git reflog show WT-moving-ref-guard
  ec15ec2cd WT-moving-ref-guard@{0}: Branch: renamed refs/heads/worktree-t342 to refs/heads/WT-moving-ref-guard
  ec15ec2cd WT-moving-ref-guard@{1}: branch: Created from origin/develop
```

`origin/develop` was `ec15ec2cd` at creation; `44095ddc2` is its parent
(`git log --oneline -3 origin/develop`). The dispatch's ref value went stale between being written
and being acted on — one integration's worth — and the §B.2 table is the consequence: a
`44095ddc2`-anchored PRESERVE check in this tree reports `3 files changed, 288 insertions(+)`
against work that does not exist.

This is the card's subject occurring in the act of dispatching the card, and it is used as
grounding instance 3 in §D.3.

**It is the same axis as the card's side-discipline, not a neighbouring one.** The side-discipline
says a once-measured divergence value must not keep being cited without re-measurement; the stale
base line is that rule violated on a base SHA instead of a divergence count. Both are a measurement
whose validity expired being re-served as current. §D.6 therefore treats them as one defect on two
carriers rather than as two rules.

The attribution above was **re-verified in this turn before being written** rather than copied from
the dispatch — `git rev-parse ec15ec2cd^` → `44095ddc2cc1c9fed2b3bd5ac946f48017988aba`, and the
reflog line quoted above. Citing the dispatched figure on the dispatcher's word would have been the
prohibited form, in the document that prohibits it.

### B.5 — The lead's dispatch-format change (attributed decision, 2026-08-28)

Recorded on the lead's authority as a design decision, **not** as an observed practice — see the
Gap note in `progress.md` §E.1. The base line changes from:

```
base: origin/develop <sha>
```

to:

```
base: measure at entry with git fetch origin develop (dispatch-time reference value: <sha>)
```

The lead's stated reasoning, quoted because it names the mechanism more precisely than a paraphrase
would: *writing the value first makes it read as the criterion, and demotes re-measurement to a
confirmation step. Inverting the order is the fix.*

This is the observation that produces remedy **R4** (§D.2) and Test 4 (§D.1). Its significance is
that the correct remedy here is neither pinning nor keeping a bare moving ref: it is **stating the
measuring command as the criterion and demoting the value to a dated reference**. A reader who
knows only "pin it" has no way to reach this form.

## §C. Scope

In scope: SPEC artifacts under `.moai/specs/**` (`spec.md`, `plan.md`, `acceptance.md`,
`progress.md`), the `internal/spec` lint engine, one always-loaded doctrine section, and the
template mirror of that doctrine.

## §D. Judgments

### D.1 — The exemption predicate: **is the ref the ANCHOR or the SUBJECT?**

This is the deliverable the card marks [HARD]. It is a three-test decision procedure, applied in
order, and every test is answerable by reading the sentence the ref appears in.

**Test 1 — Substitution.** Replace the ref token with the SHA it resolves to *right now*. Re-read
the sentence.

- It still says what it meant → the ref is an **ANCHOR** (an address at which a measurement was
  taken). Anchors are pinned.
- It now says something different, narrower, or weaker → the ref is the **SUBJECT** (the claim is
  *about* mainline as a living thing). Subjects keep the moving ref.

**Test 2 — Falsification source.** Applies when the claim currently reads false. Attribute the flip:
were the commits that caused it authored by this SPEC's work, or not?

- Authored by this work → **true signal**. The claim is genuinely broken. Fix the work; do not
  touch the ref. Pinning here would hide a real defect, and that is the worse error of the two.
- Not authored by this work → **spurious red** from upstream drift. Remediate per §D.2.

**Test 3 — Re-measurement expectation.** Re-run this claim next week with no work done in between.
Is the same answer expected?

- Yes → anchor. Pin.
- No, *and that variance is the point of the claim* → subject. Keep the moving ref.

Tests 1 and 3 will normally agree; where they disagree, Test 1 governs, because it reads the
sentence's meaning rather than predicting a future run.

**Test 4 — Read-time action.** Applies only once Tests 1-3 have returned SUBJECT. Ask: must a later
reader *act* on this claim by measuring something?

- **No — the claim is narrative.** It describes what mainline carries, quotes a command as text, or
  records a coordinate as the subject of a correction. Nothing is measured at read time. → **S1**,
  remedied by R3.
- **Yes — the claim asserts the current state of a moving thing, and a reader will act on it.** →
  **S2**, remedied by **R4**: state the measuring command as the criterion, demote any value to a
  dated reference.

Test 4 exists because the two SUBJECT sub-shapes take structurally different remedies, and because
S2 written in anchor form is exactly how the lead's dispatch failed (§B.4, §B.5).

**Classification and remedy are two steps, not one.** The tests return a class; the class does not
name the remedy by itself — ANCHOR selects between R1 and R2, SUBJECT between R3 and R4. This is
stated because collapsing the two steps is a second, subtler route to the card's dominant failure
mode: a reader who believes the class *is* the remedy has only as many remedies as there are
classes, and reaches for the first one that fits.

#### D.1.1 — Why the predicate stays two-way at the top level

The lead asked whether S2 should be a third top-level branch. It should not, and the reason is that
the top-level tests answer one question — *is this ref an address, or the thing being talked
about?* — and S1 and S2 give the same answer to it. Separating them requires a different kind of
question (what will a reader **do** with this claim?), which is naturally a follow-on rather than a
peer; promoting it would mean the top-level branch is decided by two unrelated criteria at once,
which is how decision procedures become unusable.

It also must not fold into ANCHOR, and the reason is sharper: an anchor's entire purpose is that its
value is **fixed**, while an S2 claim's entire purpose is that its value **must not be**. They are
opposite requirements on the same property. Routing S2 to ANCHOR is precisely the error that
produced the stale base line — treating "the tip you will start from" as though it were an address.

What the addendum genuinely changes is therefore the **remedy space**, which goes from three
branches to four (§D.2), plus the routing test above. Recording this as "two classes, four remedies"
rather than "three classes" keeps the classification honest and puts the new case where a reader
will actually find it.

#### D.1.2 — Generality

The predicate is not specific to git refs. It governs any *moving coordinate* — a source line
number is the same hazard on a moving tree, and §D.3 instance 2 is that case. What is detected here
is the git-ref form only; what is *published* is the predicate.

### D.2 — Four remediation branches

The failure mode the card names is a reader who learns one branch. The finding message therefore
names all four (AC-MRG-008), and the count matters more than any single entry: with one remedy on
offer the reader pins; with four, choosing requires reading the predicate.

| | Branch | Class | When | Form |
|---|---|---|---|---|
| **R1** | Pin the literal SHA | ANCHOR | the anchor value is already known at authoring time | replace `origin/main` with the resolved 40-hex SHA, recorded with the tree and date it was resolved in |
| **R2** | Freeze at pre-flight *(recommended default for anchors)* | ANCHOR | the value is not knowable when the criterion is written — the usual case for a run-phase PRESERVE criterion | `BASELINE_SHA=$(git rev-parse origin/main)` captured before the first run-phase commit; criteria decided against `$BASELINE_SHA`, resolved value recorded in `progress.md` |
| **R3** | Keep the moving ref, declare the exemption | SUBJECT / S1 | narrative — nothing is measured at read time | leave the ref; add the inline marker with a stated reason |
| **R4** | State the measuring command; demote the value to a dated reference | SUBJECT / S2 | the claim asserts the current state of a moving thing and a reader will act on it | lead with the command that must be run at read time; any value follows it, parenthesized and dated, explicitly labelled a reference |

R2 is recommended over R1 for run-phase criteria because it removes the authoring-time knowledge
requirement that makes R1 awkward, while giving the same fixed-value guarantee. It is not novel —
§B.3 records it already in use.

**R4's ordering is load-bearing, not stylistic.** Per §B.5: a value written first reads as the
criterion and demotes re-measurement to a confirmation step. The remedy is the inversion — command
first, value second and marked as a reference — so a reader who skims the line still sees an
instruction to measure rather than a number to trust.

### D.3 — The three grounded instances, mapped onto the predicate

**Instance 1 — t292 (the card's own citation): provenance narrative.** The claim's subject is what
`origin/main` *currently carries*. Test 1: substituting a SHA converts "what mainline carries" into
"what this one commit carried", which is a different and weaker statement. → **SUBJECT → R3.**

**Instance 2 — t322, SPEC-GRAPH-FRESHNESS-CADENCE-001 v0.2.2 (landed `44095ddc2`).** Its citation
refresh deliberately left `internal/mx/provenance.go:196` and the `:208` / `:219` values
unrefreshed. The SPEC's own HISTORY row states the reason: those three coordinates are "the
*subject* of audit finding N2 rather than addresses into the tree, and refreshing them would erase
the correction N2 records." Test 1, applied to a line number rather than a ref: substituting the
current coordinate destroys the record of the miscitation. → **SUBJECT → R3.** This instance is
what establishes the predicate's generality (§D.1).

**Instance 3 — the lead's dispatch base line (§B.4).** Attributed to the lead at their request; the
attribution was verified in this turn rather than taken on the dispatcher's word. The line named
`44095ddc2` meaning "the tip of develop you will start from". Test 1: substituting the *current*
value corrects it, so the ref is not narrative — but the substitution decays immediately, because
the sentence means "whatever the tip is when you enter". Test 3: re-measuring next week gives a
different tip **and that variance is the point** — the reader is supposed to get today's tip. →
**SUBJECT / S2 → R4.**

This instance is the reason Test 4 exists. Under the v0.1.0 predicate it was classified ANCHOR → R2,
and that classification was **wrong in a way worth recording**: R2 would have had the *dispatcher*
freeze a value, which is exactly the shape that failed. The defect is not that the wrong SHA was
chosen; it is that a value was stated at all where a command belonged. The measured cost of the
v0.1.0 reading is in §B.2 — a `44095ddc2`-anchored PRESERVE check reports
`3 files changed, 288 insertions(+)` against work that does not exist.

**Instance 4 (found by the §B.3 survey) — AC-COORD-016.** Asserts a literal command string is
preserved verbatim in a document. The ref token sits inside quoted subject matter, and no reader
measures anything on the strength of it. Test 4: no read-time action. → **SUBJECT / S1 → R3.**
Retained because it is the class the detector will most often flag wrongly, and it is the fixture
for AC-MRG-003.

**Instance 5 — the lead's dispatch-format change (§B.5).** The remedy of instance 3, adopted as a
standing format: `base: measure at entry with git fetch origin develop (dispatch-time reference
value: <sha>)`. It is instance 3 correctly classified — S2, remedied by R4 — and it is the only one
of the five that shows the remedy rather than the defect. It supplies the detector's R4-recognition
fixture (AC-MRG-013), so the guard does not flag the very form it recommends.

**Classification tally, and a skew worth naming.** All five grounded instances are SUBJECT-class —
S1 ×3 (instances 1, 2, 4), S2 ×2 (instances 3 and 5, the same case as defect and as remedy). **No
grounded instance is ANCHOR.** The ANCHOR class is evidenced instead by the seven corpus lines named
in §B.3 (`AC-TST-010`, `AC-AFS-012`, `AC-MSD-010`, `AC-WC-009`, `AC-CFS-010`, `AC-NS2-005`, and the
already-frozen `SPEC-DESIGN-MOAIWEBV2-002` pair).

The skew is not an accident of collection and is the strongest available argument for shipping the
predicate: the instances that got *noticed* and escalated into cards are overwhelmingly the ones
where pinning would have been wrong. Anchor-class defects are quietly correct to pin and generate no
incident; subject-class ones destroy information when pinned, which is what makes them memorable.
A guard tuned only on the noticed cases would therefore be tuned entirely on the exemption class.

Both SUBJECT sub-shapes and both of their remedies are represented, which is the minimum needed to
validate Test 4 at all. Five instances remains a small validation set (`progress.md` §E.1 residual
risk).

### D.4 — The exemption is author-declared, and it costs a reason

The predicate's tests are judgments about meaning. No regex decides them, and pretending otherwise
would produce a detector that is confidently wrong on the subject class.

The exemption is therefore an inline marker the author writes **after** applying the predicate:

```
<!-- moving-ref-ok: <reason> -->
```

on the flagged line or the line immediately above it. An HTML comment is chosen so it is invisible
in rendered markdown.

**The reason is mandatory and non-empty (AC-MRG-003).** A bare marker would make "silence the
warning" cheaper than "pin the SHA", which inverts the incentive this SPEC exists to set. With a
reason required, declaring and pinning cost about the same, and the author picks on the merits.

The existing per-SPEC `lint.skip` mechanism (`applylintSkip`, `internal/spec/lint.go`) remains
available but is deliberately **not** the exemption path: it silences the code for a whole document,
which is the wrong granularity for a per-claim judgment.

### D.5 — Severity is `warning`, never `error`

42 candidate lines exist today (§B.3). Shipping this at `SeverityError` would red the corpus on the
first run and the rational response would be a bulk suppression — the outcome this SPEC is designed
to prevent. `SeverityWarning` is non-blocking by default in the existing engine (`internal/spec/lint.go`,
the `Strict` branch of the exit-code computation) and escalates only under `--strict`.

### D.6 — Disposition of the side-discipline (divergence figures)

The card asks whether "a divergence value measured once (`0 0`) must not keep being cited without
re-measurement" is a requirement, guidance, or out of scope.

**It is the same defect as §A, on a different carrier — not a neighbouring rule.** §B.4 establishes
this: the lead's stale base line is the side-discipline violated on a base SHA rather than on a
divergence count, and both are a measurement whose validity expired being re-served as current. The
remedy is correspondingly the same: **R4**, not R1. One does not pin `0 0`; one writes the
re-measuring command as the criterion and demotes `0 0` to a dated reference — which is exactly the
transformation the lead applied to the dispatch base line (§B.5).

**Split by carrier, and recorded as such:**

- **In scope as a requirement (REQ-MRG-006) for the SPEC-document carrier.** A `progress.md` or
  `plan.md` line citing a `rev-list --count --left-right` figure with no accompanying SHA or
  timestamp is the same defect on a different carrier — a measurement whose validity has expired
  being re-served as current — and it is detectable in exactly the same place.
- **Guidance only for the dispatch carrier.** A dispatch message is not a file in the tree; the
  detector cannot see it, and claiming otherwise would be an acceptance criterion that cannot fail.
  It is stated as doctrine (REQ-MRG-005) and nothing more. §B.4 is that doctrine's worked example.

The split is recorded rather than resolved silently because the two carriers look identical in
prose and are not identical to any mechanism.

## §E. Requirements

- **REQ-MRG-001**: The SPEC lint engine shall emit a `MovingRefUnpinned` finding of severity
  `warning` for each line in a SPEC artifact that names a moving ref (`origin/main`,
  `origin/develop`, `origin/HEAD`) inside a git-command context, carries an invariant-claim marker,
  and carries neither a 7-40 character hexadecimal SHA nor a frozen-baseline variable reference.
- **REQ-MRG-002**: Where a flagged line, or the line immediately preceding it, carries a
  `<!-- moving-ref-ok: <reason> -->` marker with a non-empty reason, the linter shall suppress the
  finding for that line.
- **REQ-MRG-003**: Where such a marker carries an empty or whitespace-only reason, the linter shall
  emit a finding reporting the marker as incomplete rather than suppressing.
- **REQ-MRG-004**: The `MovingRefUnpinned` finding message shall name all four remediation branches
  of §D.2 — pin, freeze, declare, and state-the-command — and shall not present pinning as the sole
  remedy.
- **REQ-MRG-005**: The doctrine section shall state the anchor-or-subject predicate and its four
  tests (§D.1), the five grounded instances (§D.3), the four remediation branches (§D.2), and the
  six detection limits (§F), and shall be mirrored into the distributed template.
- **REQ-MRG-006**: The linter shall emit a `MovingRefUnpinned` finding for a line citing a
  branch-divergence count (`rev-list --count --left-right` against a moving ref) that carries
  neither a SHA nor a date.
- **REQ-MRG-007**: While the three-dot (`A...B`) form is present, the linter shall treat it as
  equivalent to the two-dot form and shall not suppress the finding.
- **REQ-MRG-008**: The linter shall not emit `MovingRefUnpinned` for a line whose moving ref carries
  a resolved SHA pin or a frozen-baseline variable reference (`$BASELINE_SHA` or equivalent).
- **REQ-MRG-009**: Where only `MovingRefUnpinned` findings are present, `moai spec lint` shall exit
  0; under `--strict` it shall exit non-zero.
- **REQ-MRG-010**: Where a line is written in the R4 form — a measuring command stated as the
  criterion, with any value following it parenthesized and labelled a reference — the linter shall
  not emit `MovingRefUnpinned` for that line.

## §F. Detection limits (stated, not discovered later)

Required by the card. Each limit names what the mechanism does not see, and no acceptance criterion
in this SPEC implies coverage of any of them.

- **L1 — Refs expressed without an `origin/` token are invisible.** `git diff --stat main`,
  `git diff @{u}`, `git diff HEAD~10`, and the prose form "compared against mainline" all carry the
  same hazard and none of them match. This is the lane-14 recorded Gap, carried forward unresolved.
- **L2 — The detector reads shape, never subject.** It cannot apply the §D.1 predicate. Every
  finding is a question put to a human, never a verdict, and the message says so.
- **L3 — Detection is line-scoped.** A claim whose command and whose invariant word sit on different
  lines — a wrapped table row, a fenced block with its assertion in prose above — is missed.
- **L4 — A documented command is indistinguishable from an asserted claim.** `REQ-LB-006` (§B.3)
  documents `git reset --hard origin/main` as doctrine and will be flagged. This is why the
  exemption is a marker rather than a cleverer regex.
- **L5 — No carrier outside the SPEC tree is covered.** Dispatch messages, commit messages, PR
  bodies, and reports under `.moai/reports/` carry the same defect and are not scanned (§D.6). The
  lead's dispatch base line — grounded instance 3, the occurrence that motivated R4 — sits on
  exactly such a carrier and would not have been caught by this guard.
- **L6 — A rotted reference value is indistinguishable from a live one.** REQ-MRG-010 exempts the
  R4 form, and the detector cannot tell whether the parenthesized reference value is current or
  years stale; it reads the *shape* of the demotion, not the freshness of the number. This limit is
  **created by** the R4 exemption rather than pre-existing it, and it is accepted deliberately: the
  alternative is flagging the recommended form, which would teach readers to avoid it. The residual
  is bounded by R4's own design — the command is stated first, so a reader who follows the line
  re-measures regardless of what the stale value says.

## §G. Out of Scope

### Out of Scope — corpus remediation

- Bulk-pinning or bulk-marking the 42 candidate lines of §B.3. The run-phase deliverable is a
  triage record that classifies them; editing them is separate work under their owning SPECs.
- Retrofitting closed or grandfather-era SPECs.

### Out of Scope — mechanism reach

- Detecting the L1 unpinned forms that carry no `origin/` token.
- Scanning any carrier outside `.moai/specs/**` — reports, dispatches, commit messages, PR bodies.
- Automatic remediation. The guard never rewrites a claim; a fix-it mode would apply R1 blindly and
  is precisely the failure mode §D.1 exists to prevent.

### Out of Scope — adjacent concerns

- The `graph-freshness` stamp cadence (SPEC-GRAPH-FRESHNESS-CADENCE-001) and stamp reachability
  (SPEC-STAMP-REACHABILITY-001). Both involve moving anchors; neither is this defect.
- Any change to `lint.skip`, to the era-demotion table, or to existing rule severities.

## §H. Open questions for the operator

- **Q0 — R4 form recognition (new, and the least settled).** REQ-MRG-010 exempts the R4 form, but
  the form is prose and its recognizable signature is not yet fixed — the lead's dispatch line is
  one instantiation, not a grammar. If the exclusion is written too loosely it becomes a bypass
  (any line mentioning a command escapes the guard); too tightly and it misses legitimate
  variations. Recorded as the run-phase decision most likely to need operator input, and deliberately
  not guessed here.
- **Q1 — Marker syntax.** `<!-- moving-ref-ok: … -->` is proposed on the grounds that it renders
  invisibly. An alternative is a visible inline tag readable in the rendered document. Left open
  because it is a taste call about whether the exemption should be visible to a reader of the
  rendered artifact; a reviewer may reasonably prefer visibility.
- **Q2 — REQ-MRG-006's scope.** The divergence-figure rule is the least certain requirement here:
  it may over-fire on progress records that legitimately narrate a sequence of measurements. It is
  specified narrowly (SHOULD-tier criterion) and can be withdrawn at run-phase without disturbing
  the rest; recorded rather than guessed.
- **Q3 — L1.** Closing L1 requires either a parser over fenced command blocks or a ref-resolution
  step, both materially larger than this card. Left explicitly open rather than implied away.
