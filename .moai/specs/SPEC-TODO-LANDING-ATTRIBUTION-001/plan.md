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
| A4 | Read the landed ref from the *worktree's* config instead of the primary checkout's | **Rejected** | It inverts the stated ground at `internal/cli/todo.go:81-90` — "the queue and the integration branch are properties of one repository, not of whichever worktree the command happens to run in". That ground is not refuted by anything measured here, and overturning it would make the landed verdict differ between two worktrees of one repository. Rejecting A4 is what keeps C-5 intact. |

**Residual risk carried by A3.** Level 2 reads a ref that `internal/hook/worktree_base_branch.go:156`
**writes**. The chain therefore reads a value another MoAI surface maintains, which is a coupling
rather than an independence. It is accepted on the ground that both surfaces want the same
(integration) meaning — measured across all four consumers, `spec.md` §A.6 — but a future divergence
between them would surface here first.

### A.2 The predicate's discriminator — attribution position, not occurrence

The naive repair ("match the subject only") was measured insufficient: it leaves 2 of the 7
`develop` false positives standing, because those are other cards' subjects mentioning the queried
card (`spec.md` §A.3). The discriminator is the **position at which the card is attributed**, and
the three attributing forms are enumerated in `spec.md` §A.4.

**Open shape decision, deferred to run-phase but named here so review sees it:** whether the
three forms are expressed as a widened `git log --grep` pattern, or as a subject-level filter applied
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

---

## §B Milestones

### M1 — the attribution predicate (REQ-TLA-001..006)

Priority: High. Blocks M2.

1. Establish RED against the two named mutants before writing the repair. Both mutants and their
   falsifying inputs are fixed in `acceptance.md` §D:
   - **MUT-WHOLE-MESSAGE** — today's whole-message `--grep`. Falsified by the t237 fixture.
   - **MUT-SUBJECT-ONLY** — the naive subject match. Falsified by the t216 and t443 fixtures.
2. Express the three attributing forms in one named place (REQ-TLA-002), and build the query argv in
   the single exported builder (REQ-TLA-005).
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
   repaired predicate actually holds.

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
| The three enumerated forms miss a fourth convention in use | Under-counts true positives — cards read `not-landed` and stay open | Failure is loud (a card the operator knows landed reads not-landed) rather than silent; the enumeration lives in one place so a fourth form is a one-line diff (REQ-TLA-002) |
| Level 2 disagrees with the operator's intent because another surface wrote `origin/HEAD` | The landed verdict follows a value MoAI itself maintains | REQ-TLA-011's disclosure names the level that answered, so the operator sees the coupling rather than inferring it |
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
- `acceptance.md` §D — the named mutants and their falsifying inputs.
- `.moai/reports/t472/axis-bf-measurement.md` — the measurements every figure here is carried from.
