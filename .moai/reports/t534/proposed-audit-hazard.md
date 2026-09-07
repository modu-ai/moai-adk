# Proposed plan-audit hazard — cross-artifact ordering conflict

Drafted by lane-3 from card t534, at the lead's request. Wording for a card, not yet doctrine.

## The gap this closes

Plan-audit already checks, per artifact, that each document is internally sound. It does not check
that the documents can be **followed together**. A SPEC can pass every existing hazard while
containing two obligations that cannot both be satisfied — and the implementer discovers it only
by hitting the wall.

An earlier hazard (t534's D3) covers the neighbouring case: an obligation stated only in `plan.md`
is unenforced, because a plan is guidance and an acceptance criterion is the gate. The case proposed
here is the harder sibling: the obligation IS in both, and they **disagree**.

## The proposed hazard, as it would appear in an audit dispatch

> **Cross-artifact ordering conflict.** `plan.md`'s milestone sequence and `acceptance.md`'s
> ordering obligations must be jointly satisfiable. For every clause in §D.4 Definition of Done and
> in any AC's Then that constrains ORDER — a "before", "after", "first", "prior to", "captured
> against the pre-change tree" — locate the milestone that clause binds and check the plan schedules
> it on the permitted side. Report any pair an implementer could not satisfy simultaneously. A
> conflict here is blocking: the implementer must pick one document over the other, and whichever
> they pick, the SPEC is violated.

## Discriminant — how an auditor actually checks it

Mechanical enough to run, narrow enough to finish:

1. Extract the ordering clauses. In `acceptance.md`, grep §D.4 and every AC's **Then** for
   `before|after|first|prior to|pre-change|already|not yet`. Each hit is a candidate.
2. For each hit, name the milestone it binds (the AC's exit milestone, or the DoD line's owning
   milestone).
3. Read `plan.md` §F's milestone order and ask: **does the plan schedule that milestone on the side
   the clause requires?**
4. Where it does not, that pair is a conflict. Quote both texts side by side — the conflict is only
   visible when the two are read together, which is exactly why per-artifact review misses it.

**Empty is a real result** — most SPECs carry no ordering clause at all, and an auditor reporting
"no ordering clauses found" should show the grep so the zero is an observed absence rather than an
unrun check.

## The worked example — t534, where it actually bit

`plan.md` §F ordered the milestones **M1 → M2 → M3**, M1 being the render-and-bucket change and M2
the guard test.

`acceptance.md` §D.4 Definition of Done required:

> AC-SSF-001's RED output recorded verbatim in `progress.md` §E.2, captured BEFORE the M1 render
> change, with the command and its exit code.

AC-SSF-001 is M2's exit criterion. So the DoD requires M2's RED before M1, and the plan schedules M1
before M2. **Both are binding and they cannot both be followed.**

The implementer took `acceptance.md` — correctly, since a RED reconstructed after the change is not
an observation — committed the RED as its own commit before touching production, and **reported the
tension rather than silently reordering**. That report is the only reason the conflict is on record.

The cost had the other branch been taken is concrete: following the plan's order means the guard
test is authored after the fix, its RED is never observed, and AC-SSF-001 passes as a test-after
guard — which is precisely the hole the DoD clause was added to close. The SPEC would have shipped
with its central criterion unproven while every check reported green.

## Provenance, including what I got wrong

The DoD clause that created the conflict was added in response to plan-audit finding **D3** —
"the M2 RED obligation is unbound by any criterion". So the conflict was introduced by a correction,
which is worth stating plainly: **a fix applied to one artifact can contradict another, and nothing
in the current audit re-reads the pair afterwards.** That argues for running this hazard on
correction rounds too, not only on first audits.

The nine hazards lane-3 dispatched for t534's audit — tier honesty, supersession precision, the
conditional-vs-unconditional claim, the assertion inventory, vacuity and selector zero-match,
controls, the regression-guard classification, reversal-rationale discipline, scope discipline —
were **each scoped inside a single artifact**. None asked whether the artifacts agreed with one
another. That omission is lane-3's, and it is why the conflict reached run-phase.

## Scope note

Stated for ORDER because that is the instance measured. The wider family — any obligation pair
where satisfying one violates the other — is real but harder to make checkable, and a rule that
cannot be run is worse than none. Recommend adopting the order-scoped version, and widening only if
a second instance appears with a different shape.
