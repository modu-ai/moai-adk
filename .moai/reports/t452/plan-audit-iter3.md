# SPEC-CODEX-SKILL-LOADER-001 — scoped delta-audit (iter-3)

Scope: ONE axis — the semantics of the new fourth terminal state `inconclusive`. Not a full
iteration. The mechanical closure of N1/N2/N3 was verified by the lead and is NOT re-derived here.

Verdict on this axis: **DEFECTS FOUND — 2 blocking, 1 major, 2 minor.**
Debt-closed recording: **NOT YET.** One of the two blocking findings (X2) reinstates the exact
failure N2 was raised to prevent, so the N2 closure claim is not currently supportable.
Plan-phase close: **hold until X1 and X2 land.** Both are single-clause edits, not another iteration.

## Hashes read

Verified before judging; all four match the frozen set:

| artifact | sha256 |
|---|---|
| spec.md | `ffbeb656e6a77ee3c9e07622ec06c8836fa3a2973493367b6a08262b6b2bdfbc` |
| plan.md | `8fc138ec51e7099f782a5ac2991672f058956e04dd6f19d769fe18bd72598752` |
| acceptance.md | `7aeca741c34602ef40751dfcccd8aece6a0a8d15f28955344f9496a477020107` |
| progress.md | `f8b7844e27cb8c0a07c8ff557586dd826dbe2512986f94df18cf53541abb0ef9` |

## X1 — the AC-006/007 mutual-exclusivity clause is unscoped, and now fires on two terminal paths — `acceptance.md:62` — Severity: major — Class: blocking

> "**AC-CSL-006 과 AC-CSL-007 은 상호 배타적이다** — 정확히 하나가 PASS 로 판정되고, 다른 하나는
> 해당 없음으로 기록된다. 둘 다 해당 없음이면 REQ-CSL-006·007 위반이므로 **FAIL**."

The clause is appended to AC-CSL-007, whose own `Given` is branch-A-scoped (`:59`), but the clause
itself carries no scope. Read as written it binds every run:

- **On branch B** the completion definition (`:106`) records AC-CSL-004~009 = 해당 없음. That makes
  006 and 007 *both* 해당 없음 — and this clause then declares FAIL. The branch the SPEC calls its
  leading outcome terminates in a mandated FAIL on a pair of criteria that had nothing to judge.
- **On `inconclusive`** the same collision occurs and is worse, because `:108` states neither branch
  path applies, so nothing licenses 해당 없음 for the pair in the first place.

**I did not catch this in iter-1 or iter-2.** The branch-B half is pre-existing, not planted by this
pass; `inconclusive` only added a second path onto an already-broken clause. Recording that plainly
because the iter-2 verdict rested in part on the branch-B row being coherent.

**Required fix:** scope the clause — "분기 A 에서 정확히 하나가 PASS…" — and state that on branch B
and on `inconclusive` both are 해당 없음 without violating REQ-CSL-006·007, since neither requirement's
`Where`/`When` guard is satisfied on those paths.

## X2 — the branch-recording sentence still routes a no-fire run to branch B, three lines below the rule forbidding it — `plan.md:102` vs `plan.md:97` — Severity: major — Class: blocking. NEW, planted by the N2 repair

`plan.md:97` (Rule 4, positive control):

> [HARD] **어느 뿌리에서도 발화하지 않은 회차는 `inconclusive` 이지 분기 B 가 아니다.**

`plan.md:102`, five lines later, unchanged from before the repair:

> **분기 기록**: R1(`.agents/skills`)이 적재되면 분기 A, **아니면 분기 B** 로 `progress.md` §E.2 에 적는다.

"아니면" is unconditional. A run where the signal never fires satisfies "R1 이 적재되지 않음", so this
sentence routes it to **branch B** — which is precisely the false confirmation on a broken instrument
that N2 was raised to prevent. The repair added the rule and left the operative recording instruction
untouched, so the file now contains its own contradiction, and the stale sentence is the one an
implementer reads at the moment of recording.

**Required fix:** `R1 이 적재되면 분기 A; R1 이 적재되지 않았고 **어느 뿌리에서든 신호가 발화한** 회차는
분기 B; 어느 뿌리에서도 발화하지 않은 회차는 inconclusive`.

## X3 — R1 fires but R2 does not: no disposition anywhere, and a directly-evidenced branch A would be denied — `acceptance.md:22` — Severity: major — Class: optional. NEW

AC-CSL-002's `Given` was tightened to "AC-CSL-001 의 관측이 끝났고 **양성 대조(R2)가 발화했을 때**".
Consider a run where R1 fires and R2 does not — entirely possible, since the five roots are
independent and R2's standing is a documented prior, not a guarantee:

- The `Given` is unsatisfied, so the main `Then` cannot be evaluated.
- The `[HARD]` fallback at `:25` triggers only on "**어느 뿌리에서도** 신호가 발화하지 않은 회차" — R1
  fired, so it does not apply either.
- `plan.md:96` ("이 뿌리에서조차 적재가 관측되지 않으면…") likewise addresses only the all-silent case.

No clause in either file dispositions this run. Worse, the plain reading denies a branch A that was
directly evidenced: R1 firing *is* proof the instrument works, which is all the positive control was
ever for. R2's role is to distinguish "nothing loads" from "nothing is being measured" — a question
that only arises when no root fires.

**Required fix:** restate the positive control as satisfied when **any** root fires, with R2 as the
fallback prior for the all-silent case: "양성 대조는 어느 뿌리에서든 신호가 발화하면 충족된다. 아무 뿌리도
발화하지 않은 회차에 한해 R2 를 기준으로 계측기를 의심하고 규칙 4 의 재선택으로 간다."

## X4 — the state-set enumeration licenses `inconclusive` everywhere; the next line restricts it — `acceptance.md:104` vs `:108` — Severity: minor — Class: optional

`:104` — "위 13 개 판정이 각각 PASS / FAIL / 해당 없음 / `inconclusive` 중 하나로 기록되어 있다" — grants
all four states to all thirteen. `:108` then restricts the fourth to AC-CSL-002 alone. The two bullets
disagree on scope, and the permissive one comes first, so a run could record e.g. AC-CSL-005 as
`inconclusive` and cite `:104` for it. This is the leak the axis was opened to look for; it is narrow,
but it is real and it is one clause wide.

**Required fix:** fold the restriction into the enumeration — "각각 PASS / FAIL / 해당 없음 중 하나로
(AC-CSL-002 에 한해 `inconclusive` 추가)".

## X5 — an `inconclusive` run leaves AC-CSL-003 and 004~009 with no licensed disposition — `acceptance.md:108`, `:32`, `:106` — Severity: minor — Class: optional

`:108` states that on `inconclusive` neither branch path applies. But the per-AC 적용 조건 clauses
license 해당 없음 only within a branch: AC-CSL-003's (`:32`) covers "분기 A 이면", and 004~009's coverage
comes from the branch-B row (`:106`). With no branch recorded, neither source licenses anything —
while `:104`'s 미기록 = FAIL rule still demands a value. The run is pushed toward recording either an
unlicensed 해당 없음 or a FAIL that nothing failed. X1 is the acute instance of this general shape;
this entry is the shape itself.

**Required fix:** one line in the completion definition — "`inconclusive` 회차에서는 AC-CSL-001 을 제외한
나머지가 모두 해당 없음이며, 그 사실과 blocker report 경로를 §E.2 에 기록한다."

## Question 3 — is the blocker exit reachable and terminal?

**Yes, on both counts, and it does not loop.** `plan.md:98` bounds reselection at 3 and then stops with
a blocker report whose contents it enumerates (the 3 signals tried with their observations, the R2
non-fire, and the conclusion that no session-output method of observing a load is confirmed at this
codex version). `acceptance.md:25` mirrors the bound and correctly records the AC as `inconclusive`
rather than FAIL. The stated rationale — that unbounded reselection becomes "고를 때까지 고른다", fitting
the signal to the answer after the fact — is the right reason for the bound.

One gap, minor and worth folding into X5's edit: **no AC judges that blocker report's contents.**
AC-CSL-003's (a)-(e) list governs the *branch-B* blocker report only; AC-CSL-002's `Then` requires a
branch in `progress.md` §E.2, which by definition does not exist on this path. The `inconclusive`
terminal state is the one outcome whose deliverable no criterion checks.

## Question 4 — did this repair pass plant anything new?

**Yes: X2 and X3, both from the N2 repair.** The base rate the lead cited is borne out — N1 was planted
by the iter-2 repair, and this pass planted two more. Both share one shape: the repair added a new rule
and did not sweep the sentences the rule made stale (`plan.md:102`'s recording instruction) or the
sentences the rule newly interacts with (AC-CSL-002's `Given`). That is the cross-layer sweep obligation
in `verification-completeness.md` §3 — a revision does not end in the clause it started in.

X1 is *not* newly planted; it is pre-existing on branch B and was merely given a second path by
`inconclusive`. X4 is planted, but it is a scope-wording slip rather than a mechanism defect.

I found nothing else on this axis: `inconclusive` does not appear in any requirement or criterion
beyond the three places the lead named, and the four-state set does not otherwise interact with the
branch rows.

## Recommendation

Do **not** record the iter-2 PASS-WITH-DEBT as debt-closed yet. N1 and N3 are closed. **N2 is not** —
`plan.md:102` still performs the routing N2 exists to forbid, so a no-fire run reaches branch B by the
recording instruction even though the rule above it says otherwise.

Five edits close this axis, none of which touches a requirement, a criterion's identity, or the branch
structure:

1. **X2** — rewrite `plan.md:102`'s branch-recording sentence to three cases. (Closes N2 for real.)
2. **X1** — scope the mutual-exclusivity clause at `acceptance.md:62` to branch A.
3. **X3** — restate the positive control at `acceptance.md:22` as any-root-fires, R2 as the all-silent fallback.
4. **X5** — one completion-definition line dispositioning an `inconclusive` run's other twelve criteria, and naming the blocker report as its deliverable.
5. **X4** — fold the `inconclusive` scope restriction into the state enumeration at `acceptance.md:104`.

After those land, this axis is clear and the iter-2 0.86 can be recorded debt-closed. A further
full re-audit is not warranted — the Tier M iteration ceiling is already spent, and every item above is
a wording fix whose landing is verifiable by reading the five cited lines.
