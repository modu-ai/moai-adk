---
id: SPEC-TODO-LANDING-WORDING-001
title: "Five measured wording amendments to SPEC-TODO-LANDING-ATTRIBUTION-001: the 28-floor, the S1 name, non-closure as a property, the t359 termination path, and form 3b casing"
version: "0.1.0"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec (card t486)
priority: P2
phase: "v3.2.0 target"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "kanban, landed-verdict, attribution, wording-amendment, documentation-debt"
tier: S
related_specs:
  - SPEC-TODO-LANDING-ATTRIBUTION-001
---

# SPEC: Five measured wording amendments to SPEC-TODO-LANDING-ATTRIBUTION-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-07 | Initial plan-phase authoring (card t486). Wording-only card: every figure below is carried from card t482's committed measurements (`.moai/reports/t482/verdict.md` — measured in tree `.claude/worktrees/t482`, branch `WT-landing-detect-redesign`, HEAD `25a3212a9`, pinned corpus commit `7835148d3`, 2026-09-04) and is cited by evidence path, never re-measured or re-derived here. Target of the amendments: `SPEC-TODO-LANDING-ATTRIBUTION-001` v0.4.0 (`status: completed`), files `spec.md` and `plan.md` only. |

---

## §A Context

### A.1 Why this SPEC exists

`SPEC-TODO-LANDING-ATTRIBUTION-001` (card t472) closed at v0.4.0 with its residual population of 38
subject-present-but-unattributed ids stated as "**at least 10** — a floor, not a total", on the
honest ground that nobody had classified the 38 exhaustively. Card t482 then performed that
exhaustive classification (`.moai/reports/t482/verdict.md` §4): **28 clear omissions (M) + 5 correct
exclusions (C) + 5 judgment-deferred (A) = 38**. The stated floor understates the measured clear
omissions by 2.8×. The same audit produced four further findings — the residual's largest single
shape went unnamed through three author versions and three plan-audit iterations; the enumeration's
non-closure was demonstrated structurally rather than numerically; the correct long-term solution
(persisting landing evidence at landing time) is already routed to card t359 but not named as the
termination path; and form 3b's wording leaves the merge verb's casing unspecified, with a measured
76-vs-77 count divergence.

This SPEC's single deliverable is applying **five wording amendments** to the completed SPEC's
`spec.md` and `plan.md` so that the next reader inherits t482's measurements instead of
rediscovering them. It is a documentation-debt card: **zero Go code, zero templates, zero new
attribution forms.**

### A.2 Measurement provenance (VCI §2)

Every figure this SPEC carries is attributed to t482's own recorded baseline, cited by file:

| Figure | Value | Evidence path (t482) |
|---|---|---|
| Residual classification | 38 = C 5 + M 28 + A 5 | `.moai/reports/t482/verdict.md` §4 (verdict line "합계 38 = C 5 + M 28 + A 5"), §4 table; raw 38-id dump in `.moai/reports/t482/residual-evidence.txt` |
| S1 shape size / residual contribution | 18 subjects / 18 distinct ids (10 `WT-` / 8 `worktree-`); residual contribution 14 (4 of the 18 attributed by other forms) | `.moai/reports/t482/s1-reconcile.txt`; `.moai/reports/t482/verdict.md` §4.1 |
| Form 3b casing divergence | `^[Mm]erge` counts 77, `^Merge` counts 76; delta subject `merge: WT-ci-test-observability into develop (t358)` | `.moai/reports/t482/form3b-delta.txt` |
| Open-population evidence | `t409` (2026-09-01), `t460` (2026-09-03) — new shapes entering up to two days before the corpus pin | `.moai/reports/t482/verdict.md` §2.7 |
| Ancestor-propagation refutation | `git log d2ad26c90^2 --not d2ad26c90^1` yields 1 commit carrying no card token — no inheritable attribution | `.moai/reports/t482/verdict.md` §2.6 |

t482's baseline, restated from its verdict: measured 2026-09-04 in tree `.claude/worktrees/t482`
at HEAD `25a3212a9` against the pinned corpus commit `7835148d3` (5,837 subjects, 414 merges). The
corpus pin is an SHA, not a branch name, so no figure here rides a moving ref. Nothing in this SPEC
re-measures or re-derives these numbers; a reader who needs them fresh re-runs t482's committed
scripts (`forms.py`, `s1_reconcile.py`, `form3b_delta.py` — all in `.moai/reports/t482/`).

### A.3 What the amendments do NOT do

- They do **not** reopen the completed SPEC: its frontmatter `status:` stays `completed`, no
  `amendment_of:` field is added to it, and no `completed → in-progress` transition occurs. The
  amendments land as t486's own commits; this SPEC documents them (operator decision — no precedent
  for reopening closed cards).
- They do **not** adopt S1 (or any shape) as an attribution form. Naming only, with the operator's
  rejection recorded.
- They do **not** state the residual count as a reduction target, and do not invite a future round
  to shrink it.

---

## §B Requirements (GEARS)

### REQ-TLW-001 — replace the floor figure with t482's measured 28, still a floor

The amended `spec.md` §A.4.2 shall state the residual's clear-omission floor as **28 measured**
(carrying the classification 38 = C 5 + M 28 + A 5 with its evidence path,
`.moai/reports/t482/verdict.md` §4 + `residual-evidence.txt`), and shall state that 28 remains a
FLOOR: the 5 judgment-deferred items were never opened to diff level, and opening them can only
raise M — to at most 33 — never lower it. The history line "1 → 19 → 7 → ≥10" shall gain
"→ 28 measured", and `plan.md` §D's residual risk row shall replace both occurrences of the
"at least 10" figure accordingly.

**[HARD]** The amended text shall not read as "the residual is now closed or fully counted" — it
updates a floor to a better-measured floor, and must say so.

### REQ-TLW-002 — name the residual's largest single shape as S1, recording the rejection

The amended `spec.md` §A.4.2 shall name the residual's largest single shape **S1 (the
release-integrate shape)**: subject begins `merge(WT-…)` or `merge(worktree-…)`, integrates into a
release branch (e.g. `: integrate into release/v3.1.1`), and the card token appears ONLY inside the
merge scope (the branch name) — never as attribution (example: `merge(worktree-t132): integrate
into release/v3.1.1`). The text shall state BOTH measured figures as distinct: **shape size 18
subjects / 18 distinct ids** (spellings: 10 `WT-` / 8 `worktree-`) and **residual contribution 14**
(4 of the 18 are attributed by other forms) — evidence: `.moai/reports/t482/s1-reconcile.txt` and
verdict.md §4.1. The text shall record that S1 passed 3 author versions and 3 plan-audit iterations
unnamed, so the name exists to prevent rediscovery.

**[HARD]** NAMING ONLY. The amendment shall record the operator's explicit REJECTION of adopting S1
as a rule/form: a rule that reads `merge(WT-t131):` as attributing t131 attributes a branch name —
a positional exception the non-attribution rule must not pay for (verdict.md §5.1 cause 2). No
attribution form is added.

### REQ-TLW-003 — state "the enumeration does not close" as a property, not a number

The amended `spec.md` §A.4.2 shall state the enumeration's non-closure as a **structural
property**, grounded in three measured facts (verdict.md §5.1): (a) the residual population is
OPEN — `t409` entered 2026-09-01 and `t460` 2026-09-03, so the enumeration can close only for the
past, never for the future; (b) expanding the largest residual shape into a form collides with
§A.4's existing [HARD] branch-name non-attribution rule — the same silent over-count direction the
SPEC already rejects at `(branch WT-t80)`; and (c) the structural bypass — attributing a card from
inside a merge scope via ancestor propagation — is refuted by the corpus: `git log d2ad26c90^2
--not d2ad26c90^1` yields no inheritable attribution (verdict.md §2.6). The text shall carry t482's
ground sentence: shape size was 18, of which 4 accidentally matched other forms and were excluded
from the residual — therefore searching for a shape by reading only the residual list
systematically underestimates that shape's true size; three audit rounds re-counted the residual
each time (1 → 19 → 7 → ≥10), exactly that under-counting procedure.

**[HARD]** The residual count shall not be written as a number a future round could be asked to
shrink. Future rounds that re-count the residual repeat the exact defect this card fixes.

### REQ-TLW-004 — state t359 as the sole termination path

The amended `spec.md` §D "Out of Scope — axis C, landing evidence storage" section shall state
explicitly: this enumeration is a **transitional instrument** until card t359 lands;
landing-time recording of landing evidence is the **termination path** for the enumeration debt;
and further enumeration rounds are not the plan (verdict.md §5.2(1) and §6 recommendation 4).

### REQ-TLW-005 — state form 3b's merge-verb casing latitude and correct the count 76 → 77

The amended `spec.md` §A.4 form 3b row (and its surrounding section where the count appears) shall
correct the observed count **76 → 77** and shall state the merge verb's **casing latitude**: form
3b's definition (named integration target + exactly one card token in the trailing group) does not
prescribe the merge verb's casing, so implementations shall not add a case restriction the corpus
refutes — `^[Mm]erge …` counts 77, `^Merge …` counts 76, and the one diverging subject is `merge:
WT-ci-test-observability into develop (t358)` (evidence: `.moai/reports/t482/form3b-delta.txt`).

**[HARD]** The residual figure shall remain INVARIANT under this amendment: t358's attribution is
unchanged (carried by its trailing `(t358)` group), so the 347 / 309 / 38 figures and the 28-floor
shall not drift as a consequence of the 76 → 77 correction.

### REQ-TLW-006 — target immutability guard

While the amendments are applied, the target `SPEC-TODO-LANDING-ATTRIBUTION-001` frontmatter
`status:` shall remain `completed`; the amendments shall touch only the target's `spec.md` and
`plan.md` — never its `acceptance.md`, `progress.md`, or any other file.

---

## §C Acceptance Criteria (inline, Tier S)

### AC-TLW-001 — floor figure replaced, floor semantics preserved

**Given** `SPEC-TODO-LANDING-ATTRIBUTION-001/spec.md` §A.4.2 after amendment,
**When** a reader greps the residual discussion for the floor figure,
**Then** the text states 28 as the measured clear-omission count (with the 38 = C 5 + M 28 + A 5
classification and its evidence path `.moai/reports/t482/verdict.md` §4 +
`.moai/reports/t482/residual-evidence.txt`), states that 28 is still a floor (the 5 deferred items
can only raise M, to at most 33), the history line reads "1 → 19 → 7 → ≥10 → 28 measured", and no
sentence in the section reads as "the residual is now closed or fully counted".
**And** `plan.md` §D's residual risk row carries the 28 figure in place of both "at least 10"
occurrences.

### AC-TLW-002 — S1 named with both figures and the rejection recorded

**Given** the amended §A.4.2,
**When** a reader searches for the residual's largest single shape,
**Then** the text names S1 (release-integrate shape) with its subject definition and example,
states BOTH 18 (shape size: 18 subjects / 18 distinct ids, 10 `WT-` / 8 `worktree-`) and 14
(residual contribution) as distinct measures citing `.moai/reports/t482/s1-reconcile.txt`, records
that S1 was unnamed across 3 author versions and 3 plan-audit iterations, and records the
operator's explicit rejection of adopting S1 as a form with its ground (attributing a branch name;
positional exception).

### AC-TLW-003 — non-closure stated as a property with three grounds

**Given** the amended §A.4.2,
**When** a future round proposes adding another form,
**Then** the text answers from property, not count: it states the open population (t409 2026-09-01,
t460 2026-09-03), the collision with the [HARD] branch-name non-attribution rule, and the refuted
ancestor-propagation bypass (`git log d2ad26c90^2 -- not d2ad26c90^1` — no inheritable
attribution), and carries the under-estimation ground sentence (residual-list reading
systematically underestimates a shape's true size; 3 rounds re-counted, 1 → 19 → 7 → ≥10).
**And** no sentence anywhere in the amended files states the residual count as a reduction target.

### AC-TLW-004 — t359 stated as the sole termination path

**Given** the amended §D "Out of Scope — axis C, landing evidence storage",
**When** a reader asks how the enumeration debt ends,
**Then** the section states that the enumeration is a transitional instrument until t359 lands,
that landing-time recording is the termination path, and that further enumeration rounds are not
the plan.

### AC-TLW-005 — form 3b count corrected with casing latitude; residual invariant

**Given** the amended §A.4 form table row for 3b,
**When** a reader or implementer checks the observed count and the verb casing,
**Then** the row (or its section) states 77 and states the casing latitude (`^[Mm]erge` 77 vs
`^Merge` 76, delta subject `merge: WT-ci-test-observability into develop (t358)`, evidence
`.moai/reports/t482/form3b-delta.txt`),
**And** the 347 / 309 / 38 figures and the 28-floor elsewhere in the same file are unchanged by
this correction.

### AC-TLW-006 — target untouched outside the two files; status stays completed

**Given** the amendment commits on t486's branch,
**When** `git diff --name-only` is taken over the commits that apply the amendments,
**Then** the changed paths are exactly `…/SPEC-TODO-LANDING-ATTRIBUTION-001/spec.md` and
`…/SPEC-TODO-LANDING-ATTRIBUTION-001/plan.md` (plus t486's own SPEC artifacts), and the target's
frontmatter still reads `status: completed` with no `amendment_of:` field added.

---

## §D Constraints

- **C-1** — Wording only: no Go code, no templates, no new attribution forms, no queue-store
  changes.
- **C-2** — All figures carried from t482's committed evidence with path citations; none
  re-measured, re-derived, or hand-enumerated by this card.
- **C-3** — Target files: `SPEC-TODO-LANDING-ATTRIBUTION-001/spec.md` and `plan.md` only.
- **C-4** — Target frontmatter `status: completed` is immutable for this card; no reopening, no
  `amendment_of:`, no version bump of the target beyond what a wording amendment note requires.
- **C-5** — No residual count is stated as a reduction target anywhere in the amended text.

---

## §E Exclusions

This section states what is **out of scope** for this SPEC and why, so no successor re-derives the
boundary.

### Out of Scope — attribution form additions

- Adding S1 (or any residual shape) as attribution form 7 or otherwise. The operator explicitly
  rejected S1 adoption; this card names it and records the rejection (REQ-TLW-002).
- Any widening or narrowing of forms 1/2/2b/3a/3b/3c beyond the form 3b casing-latitude wording
  (REQ-TLW-005), which changes no subject's attribution.

### Out of Scope — target SPEC lifecycle changes

- Reopening `SPEC-TODO-LANDING-ATTRIBUTION-001`: its `status:` stays `completed`; no
  `completed → in-progress` transition, no `amendment_of:` field, no new plan-audit round on the
  target (operator decision — no precedent for reopening closed cards).

### Out of Scope — implementation and measurement work

- Go code, templates, hooks, or queue-store changes (the card is documentation-only).
- Re-measuring or re-deriving the 28 / 18 / 14 / 77 figures; t482's committed evidence is the
  cited source of record.
- Opening the 5 judgment-deferred (A) items to diff level — that would be a new classification
  round, which REQ-TLW-003's property statement exists to retire.
- Reading or amending card t359's content; REQ-TLW-004 cites the target's own §D routing, which is
  the recorded ground.

### Out of Scope — target files beyond spec.md and plan.md

- `SPEC-TODO-LANDING-ATTRIBUTION-001/acceptance.md` and `progress.md` are not amended (the five
  amendments have no acceptance-matrix or evidence-log consequence there).
