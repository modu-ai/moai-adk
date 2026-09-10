# Implementation Plan — SPEC-TODO-LANDING-ATTRIBUTION-001

Card: t472. Tree: `.claude/worktrees/t472`, branch `WT-landed-drift-detect`, measured at HEAD
`4bcac7079`.

Ordered by decision-reversibility: the decisions most likely to change are stated first, mechanical
work last.

---

## §A Decisions taken, and the alternatives rejected

### A.1 The ref chain — option A3 selected

Four options were measured in `.moai/reports/t472/axis-bf-measurement.md`. The selection and the
grounds:

| | Option | Verdict | Ground |
|---|---|---|---|
| A1 | Wait for `develop` to reach `main` | **Rejected** | Lead verdict. Its timing is bound to the t204 deploy gate, and the closed-but-picked backlog recurs every release cycle regardless — a remedy that expires is not a remedy. |
| A2 | Write `develop` into the primary checkout's config | **Rejected** | Two independent blockers, both measured. The primary checkout is parked on `main`, so the edit is not committable there; and `moai update` redeploys `.moai/config` wholesale from the template, so any hand edit is reverted on the next update (`CLAUDE.local.md` §2.3). A remedy that a routine maintenance command silently undoes is not durable. |
| **A3** | **Three-level chain: config → `refs/remotes/origin/HEAD` → `DefaultLandedRef`** | **SELECTED** | The repository already records the answer: `git symbolic-ref refs/remotes/origin/HEAD` returns `refs/remotes/origin/develop` in this tree. Today's resolver ignores that record and reaches for a compiled-in constant. Preferring the repository's own recorded default over a hardcoded branch name is the same direction as `CLAUDE.local.md` §14 (no hardcoding), and it needs neither a config edit nor a release to take effect. A project that configures the key is unaffected — level 1 still answers first (C-4). |
| A4 | Read the landed ref from the *worktree's* config instead of the primary checkout's | **Rejected** | It inverts the stated ground at `internal/cli/todo.go:81-90` (doc comment `:81-87`, func `:88-90` — the same range `spec.md` §A.1 now cites; version 0.1.0's two artifacts disagreed by one line, plan-audit D6) — "the queue and the integration branch are properties of one repository, not of whichever worktree the command happens to run in". That ground is not refuted by anything measured here, and overturning it would make the landed verdict differ between two worktrees of one repository. Rejecting A4 is what keeps C-5 intact. |

**Residual risk carried by A3 — re-measured, and smaller than version 0.1.0 recorded (plan-audit D2).**
Version 0.1.0 stated that level 2 reads a ref `internal/hook/worktree_base_branch.go:156` **writes**,
and carried that as an unbounded coupling. Both halves were wrong. Measured at HEAD `e227871b4`:

- **The citation.** `:155` is `worktreeBaseBranchReadConfigReal`, a config **read**. The write is
  `WorktreeBaseBranchSetHead(configured)` at **`:125`**, backed by `worktreeBaseBranchSetHeadReal`
  (`git remote set-head`) at **`:170`**.
- **The coupling is disjoint, not unbounded.** `RunWorktreeBaseAlignment` gates on the primary
  checkout at `:92` and returns at `:97-100` when the configured key is empty (REQ-WBR-005 — "the
  neutral value performs no git-metadata read at all"). Chain level 2 fires only when level 1 is
  **empty**; the writer fires only when level 1 is **non-empty**, from the same primary-checkout root.
  **Mutually exclusive by construction — no cycle exists.**

What actually remains is the converse, and it is not a risk this SPEC carries: a project that
*configures* the key never reaches level 2, so the two surfaces never interact on this path. Recorded
as an exclusion in `spec.md` §D rather than as an accepted risk, because an unverified premise dressed
as an accepted risk is itself the defect (VCI §1).

### A.2 The predicate's discriminator — attribution position, not occurrence

The naive repair ("match the subject only") was measured insufficient: it leaves 2 of the 7
`develop` false positives standing, because those are other cards' subjects mentioning the queried
card (`spec.md` §A.3). The discriminator is the **position at which the card is attributed**; the
positional shapes (forms 1, 2, 2b, 3a, 3b, 3c) and the non-attribution rule are enumerated in `spec.md`
§A.4. Form 3 was repaired from an occurrence test to two positional shapes at version 0.2.0
(plan-audit iter-1 D1) — the occurrence wording admitted 146 merge subjects against 5 attributing
ones and is now the named mutant `MUT-MERGE-ANY-TOKEN`.

Version 0.3.0 made three further corrections, each on a measurement rather than a preference
(plan-audit iter-2 D1/D2/D3/D4):

- **Form 3c added** — the `merge: <card>` local-merge spelling, 31 subjects with the delivering card
  as the first token in 31 of 31. It is the largest family the four-form enumeration missed, and
  admitting it takes the measured under-count from 19 cards to 7 (`spec.md` §A.4.2).
- **Form 3b narrowed and its target made derived** — the group must carry exactly one card token,
  and the target is read from the resolved landed ref rather than spelled `develop`
  (`spec.md` §A.4.1, REQ-TLA-013). Both were unfalsified before; AC-TLA-003 clauses 4-6 now falsify
  them.
- **The non-attribution rule contraposed** — target mismatch rather than a `WT-` prefix, because
  nine merge targets in this repository's own history are card or agent worktrees carrying no `WT-`
  prefix, three of them no prefix at all.

Every entry in the enumeration now has a falsifier, and the per-form "ids this form alone
attributes" column in `acceptance.md` AC-TLA-005 is the check a further form must pass before it is
added: a form measuring 0 there cannot be falsified from the corpus.

**Open shape decision, deferred to run-phase but named here so review sees it:** whether the
shapes are expressed as a widened `git log --grep` pattern, or as a subject-level filter applied
in Go over `git log --format=%s`. Both satisfy REQ-TLA-001..004. The trade-off is that a widened
`--grep` keeps the one-subprocess budget of `GitLandedQuerier` intact but concentrates the
correctness into a regex whose failure mode is silent (an unmatched pattern is byte-identical to
"not landed" — the hazard `prlink_landed.go:4-9` was written about); a Go-side filter is inspectable
and unit-testable per form but changes what the querier consumes. **Whichever is chosen, REQ-TLA-005
binds: one exported builder, asserted against by the tripwire.**

### A.3 Milestone order — F before A, and why it is not a preference

**[HARD] Milestone 1 (axis F) lands before Milestone 2 (axis A+B).**

The ground is measured, not stylistic. Landing the ref correction first does not fix a symptom: it
**grows the false-positive population from 2 to 9**, of which 7 are wrong (`spec.md` §A.2). Nine
cards would read `landed` against a ref the operator has just been told to trust. Repairing the
predicate first means the ref correction lands on a predicate that can bear it.

A run-phase that reverses this order, or that lands M2 alongside M1 in one commit, has not delivered
this SPEC.

**The ordering was challenged at plan-audit iteration 2 and survives on a measurement.** The audit
argued that the M1-only window makes a correct form-3b implementation indistinguishable from a
`develop`-hardcoding one, and treated that as a reason the target derivation could not be left to
run-phase. Re-measured on `origin/main` `7ad9f8534` (`spec.md` §A.7): **0** of its 101 merge subjects
target `develop` with a card-bearing trailing group, and 0 target `main`, so form 3b contributes
nothing either way in the window and the two implementations are behaviourally identical there. The
window is real and harmless; the derivation defect is real and **permanent downstream**, which is
why it is closed by REQ-TLA-013 + AC-TLA-003 clause 6 rather than by disturbing this ordering. The
ordering's own ground — the measured 2→9 false-positive growth — is untouched.

---

## §B Milestones

### M1 — the attribution predicate (REQ-TLA-001..006, REQ-TLA-013)

Priority: High. Blocks M2.

1. Establish RED against the three axis-F mutants before writing the repair. All three and their
   falsifying inputs are fixed in `acceptance.md` §A:
   - **MUT-WHOLE-MESSAGE** — today's whole-message `--grep`. Falsified by the t237 fixture.
   - **MUT-SUBJECT-ONLY** — the naive subject match. Falsified by the t216 and t443 fixtures.
   - **MUT-MERGE-ANY-TOKEN** — form 3 read as an occurrence test. Falsified by `9a3837b5c`
     (t386/t387) and `c4ae1ecbd` (t284), via AC-TLA-003's third clause.
   - **MUT-NO-FORM-3B / MUT-NO-FORM-3A / MUT-NO-FORM-3C** — an enumeration entry silently omitted.
     Falsified by `t412` (AC-TLA-003 clause 4), `t244` (AC-TLA-003b clause 1), and `t79`
     (AC-TLA-003b clause 3) respectively — the one id each form alone attributes.
   - **MUT-GROUP-ANY-TOKEN / MUT-NO-TARGET-TEST** — form 3b's group read as an occurrence set, or
     its target test dropped. Falsified by `t412`'s three absorb siblings (AC-TLA-003 clause 5).
   - **MUT-HARDCODED-DEVELOP** — form 3b's target spelled rather than derived. Falsified only by
     AC-TLA-003 clause 6, which varies the resolved ref; no corpus fixture distinguishes it.
   - **MUT-FIRST-TOKEN-OF-GROUP** — the declined widening. Falsified by `t80` (AC-TLA-003b clause 4).
2. Express the §A.4 shapes **and the non-attribution rule** in one named place (REQ-TLA-002), with
   form 3b's target and the rule's target comparison derived from the resolved landed ref rather
   than spelled (REQ-TLA-013), and build the query argv in the single exported builder
   (REQ-TLA-005).
3. Assert both directions on every criterion: a title-attributed commit reads `landed`, and a
   body-mention-only commit reads `not-landed`. A one-directional suite lets "everything is landed"
   pass and is not acceptable evidence.
4. Preserve the three-valued answer on the unanswerable path (REQ-TLA-006).

### M2 — the ref chain and its disclosure (REQ-TLA-007..012)

Priority: High. Depends on M1.

1. Insert level 2 between the existing config read and `DefaultLandedRef`, with the level-3
   fall-through on any failure (REQ-TLA-008, REQ-TLA-009).
2. Carry which level answered out of the resolver, so the caller can disclose it without
   re-deriving it — the same shape `LandedRef()` already uses to let a refusal name the ref without
   re-resolving.
3. Append the answering ref to the `todo done` verdict line, preserving the `done <id> ` prefix
   (REQ-TLA-010); disclose the sub-config level on stderr (REQ-TLA-011).
4. Assert the read-only property (REQ-TLA-012) by a subprocess census over the resolution path — the
   package already has that seam (`CommandRunner`).

### M3 — the tripwire and the regression surface

Priority: Medium. Mechanical; ordered last deliberately.

1. Keep the existing `-E` engine-flag tripwire green, or replace it with the equivalent assertion
   for whatever query shape §A.2 settles on. The tripwire exists because the failure mode is silent;
   removing it without a replacement is a regression regardless of the suite's colour.
2. Extend the `todo pr` outcome documentation to state the predicate's limit in the terms the
   repaired predicate actually holds. **Non-gating** (plan-audit D7): this item maps to no REQ-TLA and
   to no acceptance criterion, and `acceptance.md` §F records it explicitly as documentation work that
   does not gate close. It is carried on M3 for sequencing, not as a deliverable the SPEC is judged on.
   M3 item 1, by contrast, IS gated — via AC-TLA-006.

---

## §C Technical approach

- Both changes are confined to `internal/kanban/prlink_landed.go` and its two callers
  (`internal/cli/todo.go`, `internal/cli/todo_pr.go`). No schema change, no new subcommand, no new
  config key.
- The verdict line change is additive at the tail; the `done <id> ` prefix that existing readers key
  off is unchanged.
- The chain's level 2 is a local ref read. No network I/O is added (C-3).

---

## §D Risks

| Risk | Direction | Mitigation |
|---|---|---|
| A widened regex silently matches nothing | A predicate that answers `not-landed` for every card is byte-identical to a working one that found nothing | Positive control in every criterion (REQ-TLA-001 green direction), plus the retained tripwire |
| The six enumerated forms miss a further convention in use | Under-counts true positives — cards read `not-landed` and stay open | Failure is loud rather than silent; the enumeration lives in one place so a sixth form is a one-place diff (REQ-TLA-002). **A sixth form is admitted only on the AC-TLA-005 falsifier check**: measure the ids that form alone would attribute, and reject it if that count is 0 — a form with no unique contribution cannot be falsified from the corpus, which is the vacuity iter-2 D1 found in form 3b |
| Form 3b's target is spelled rather than derived | Silent: invents attributions in a repository that never merged `into develop`, and misses every merge into the branch it does integrate on | REQ-TLA-013 states the derivation as [HARD]; AC-TLA-003 clause 6 falsifies a spelled target by resolving the ref to something other than `develop`. Note the M1-only window does **not** falsify it — measured on `origin/main` `7ad9f8534`, 0 of 101 merge subjects target `develop` or `main` with a card-bearing group, so the two implementations are behaviourally identical there (`spec.md` §A.7) |
| Level 2 disagrees with the operator's intent because another surface wrote `origin/HEAD` | The landed verdict follows a value MoAI itself maintains | Measured disjoint (§A.1 above): the writer fires only on a non-empty key, level 2 only on an empty one. The residual is bounded to a future change that breaks that disjointness; REQ-TLA-011's disclosure names the answering level either way, so the operator sees the source rather than inferring it |
| Narrowing §A.4 under-counts a genuine landing | A card whose landing subject uses a shape the enumeration does not carry reads `not-landed` | **The figure has been wrong three times and the "archived-era" ground is WITHDRAWN as false (plan-audit iter-3 D3-1).** History: 1 (v0.2.0) → 19, reduced to 7 by adopting form 3c (v0.3.0) → ≥10 at v0.4.0, after adopting form 2b removed 39 ids the `)$` anchor had been silently dropping → **28 measured** by card t482's exhaustive classification (`.moai/reports/t482/verdict.md` §4) (`spec.md` §A.4.2). Version 0.3.0 grounded its tolerance on "all seven are archived-era ids so today's operational impact is nil". That ground was **false when written** — the 39 it did not know about are this repository's most recent merges, PR `#1600`-`#1670` — and it does not hold for the residual either: at least four residual ids (`t216`, `t409`, `t443`, `t460`) appear in subjects dated `2026-09-03`, the corpus's newest day (`git log 7835148d3 --perl-regexp --grep='\b(t443\|t460\|t409\|t216)\b' --format='%ad %s' --date=short`). **No dating or operational-impact claim replaces it**: the residual now carries an exhaustive classification — **28 measured** clear omissions of 38 = 5 correct exclusions + 5 judgment-deferred (`.moai/reports/t482/verdict.md` §4) — and 28 is a floor, not a total: the 5 judgment-deferred items were never opened to diff level and can only raise it, to at most 33 (`spec.md` §A.4.2 [HARD]); the live-queue status this row once carried as unmeasured is measured by the same audit (34 of 38 absent from the queue store, 4 archived, 0 live — verdict.md §2.5), a fact about today's queue that closes nothing. The tolerance now rests on the failure **direction** alone, which is measured rather than dated: the under-count is **loud** (an operator who knows the card landed sees `not-landed`), and the one widening that would recover most of the residual is measurably **silent** in the wrong direction — it reads the branch name in `(branch WT-t80)` as an attribution. Trading a loud under-count for a silent over-count is the wrong direction for this SPEC; a reader who finds that insufficient should classify the 38 rather than re-argue the dating |
| M2 landed without M1 | 7 wrong closures become reachable | §A.3, stated as a [HARD] ordering decision with its measured ground |

---

## §E Anti-patterns

- **Landing M2 first, or landing both in one commit.** Measured to grow the defect, not shrink it.
- **A one-directional acceptance criterion.** Asserting only that a landed card reads `landed` lets
  a predicate that answers `landed` for everything pass.
- **Adding a new query subcommand.** Declared a non-goal in `spec.md` §D — the detection surface
  already ships as `moai todo pr`.
- **Writing `develop` into the primary checkout's config as the "quick" fix.** Not committable there
  and reverted by the next `moai update`.

---

## §F Cross-references

- `spec.md` §A.7 — the ordering decision this plan's milestone order implements.
- `acceptance.md` §A — the named mutants and their falsifying inputs; §D — the edge cases, including the two-attribution tiebreak.
- `.moai/reports/t472/axis-bf-measurement.md` — the measurements every figure here is carried from.
