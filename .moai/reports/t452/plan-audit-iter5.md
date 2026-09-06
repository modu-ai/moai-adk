# SPEC-CODEX-SKILL-LOADER-001 — final confirm (iter-5)

Two items only. No re-audit, no rescoring.

**Item 1 — Y1 is CLOSED**, confirmed against the literal row. One new item planted (Y2), one clause wide.
**Item 2 — the author's argument FAILS as stated (outcome iii)**, but their *disposition* is right on a
different ground, so the required edit is (ii)'s scoping clause plus a corrected reason.
**Close — YES, with a named pre-M0 debt.** The iter-2 PASS-WITH-DEBT 0.86 **can be recorded
debt-closed**; Y2 and the AC-CSL-009 clause are new, separate, one-line debt that rides into
run-phase under the condition stated at the end.

## Hashes read

| artifact | sha256 |
|---|---|
| acceptance.md | `20dd94bf79d94f3e0201e0e2de7c69cd94c4a2663dcf3e2b29f2c5cd849f360c` |
| spec.md | `b19d920e0c3bc50fa1ffbd7b55027767a3e8ac060c53aac131c50239ca2d9e8f` |
| plan.md | `562c0872d2045ee57847b9f170821075ff55797939144da53d566242dcac5599` |
| progress.md | `f8b7844e27cb8c0a07c8ff557586dd826dbe2512986f94df18cf53541abb0ef9` |

## Item 1 — Y1 closed; one new item planted

**CONFIRMED.** The `inconclusive` row now reads AC-CSL-002 = `inconclusive`, AC-CSL-001 = PASS,
**AC-CSL-010 = PASS** on the branch-B row's stated grounds, remaining 10 (AC-CSL-003~009 · 011 · 012 ·
013) = 해당 없음. Arithmetic: 003-009 is 7, plus 011/012/013 is 10; 3 + 10 = 13. The [HARD]
anti-refold clause is present and grounded in AC-CSL-010's `Given`/`When` holding on that path.
Stale count references (`나머지 11`, `003~013`) → 0, with `나머지` → 1 as a live control.

Two things in this edit are better than what I asked for and worth recording: the row explains **why
010 and 011 diverge** (one judges an *execution*, the other a *change set*), and it states that
AC-CSL-001 and AC-CSL-010 are PASS **by their own criteria** rather than by the row's permission —
which is the correct relationship between a disposition table and a criterion body.

**Y2 — the row's justifying sub-bullet asserts a property of the ten that is false for one of them —
`acceptance.md` 완료 정의, `inconclusive` row, the "이 줄이 없으면…" sub-bullet — Severity: minor. NEW.**
It now reads "…`inconclusive` 회차에서 **자기 `Given`/`When` 이 성립하지 않는** 10 개 판정에 아무 처분도
허가되지 않는다". That qualifier is false for AC-CSL-009, whose `Given` holds on this path (item 2
below establishes it), and AC-CSL-009 is one of the ten. The disposition is unaffected; the stated
reason for it is wrong for one member. The qualifier was added in this pass, so it is planted, not
inherited.
**Required fix:** drop the qualifier, or carve out AC-CSL-009 alongside the 001/010 carve-out already
in the parenthetical.

## Item 2 — adjudication: **(iii), the argument fails; remedy is (ii)'s clause**

I read AC-CSL-009 at `acceptance.md:73-76` rather than accepting the summary, and the argument does
not survive the reading.

**The `Given` holds — the author is right about that.** `:73` is "프로브가 사용한 codex-cli 버전이 기록되어
있을 때", and AC-CSL-001's five recorded elements and AC-CSL-003(e) both guarantee it on branch B and
on `inconclusive`.

**The `When` also holds — this is where the argument fails.** `:74` is not "the manifest is touched";
it is two greps: `grep -A4 'class: skill-loader' …` and `grep 'codex_measured_version' …`. The
manifest is a checked-in file present on every path, so both greps run and produce output whether or
not this SPEC edits it. "Those two paths never touch the manifest" is true and irrelevant — nothing
in the `When` requires touching it.

**So the criterion is fully evaluable on both paths, and its `Then` is false.** `:75` requires the
`skill-loader` row's rationale to contain this round's `codex --version`. On branch B, REQ-CSL-003
forbids editing that file, so the row still carries its `deferred-m1` rationale with no version
string. Read alone, **AC-CSL-009 FAILs on branch B and on `inconclusive`** while the disposition table
says 해당 없음 — and unlike the 012·013 case, no canonical-on-divergence clause covers 009 (that clause
is scoped to 012·013 in its own sub-bullet). This is not latent: it is live on the leading branch
today.

**The disposition is nevertheless correct, on a ground the author did not give.** REQ-CSL-003 forbids
the manifest edit on branch B; REQ-CSL-009 (`spec.md:159`) asks the version to be recorded "이번에 잰
대상 옆에 — `skill-loader` 행의 처분 rationale 에", which that prohibition makes unreachable; and
AC-CSL-003(e) (`acceptance.md:31`) already relocates the version anchor to the blocker report, saying
so explicitly. So 해당 없음 is right because **the requirement's own obligation is discharged elsewhere
on that path**, not because the `When` is unsatisfiable.

**Verdict: (iii) for the argument, (ii) for the remedy.** AC-CSL-009 is the one criterion in this
situation that never received the treatment its siblings got — AC-CSL-003 gained an 적용 조건 naming
branch A and `inconclusive`; AC-CSL-012 and AC-CSL-013 gained [HARD] empty-sweep clauses; AC-CSL-009
gained nothing.

**Required fix, one clause on AC-CSL-009, mirroring AC-CSL-003's:**
> **적용 조건**: 분기 B 와 `inconclusive` 회차에서는 해당 없음으로 기록한다 — REQ-CSL-003 이 매니페스트
> 편집을 금지하므로 이 판정이 요구하는 rationale 갱신 자체가 도달 불가능하고, 그 경로의 버전 고정점은
> AC-CSL-003(e) 의 blocker report 다. `Given`/`When` 이 성립하는데도 해당 없음인 이유가 이것이며,
> 처분표만으로 이 자리를 덮지 않는다.

## Close

**Record the iter-2 PASS-WITH-DEBT 0.86 as debt-closed.** N1, N2 and N3 are closed; X1-X5 are closed;
Y1 is closed. Nothing from those rounds remains open.

**Plan-phase may close, carrying Y2 + the AC-CSL-009 clause as one named pre-M0 debt** — on the
lead's stated preference, and I agree with it here, with one condition attached because the AC-CSL-009
item is not a wording nit:

- The debt entry MUST carry the **exact clause text above** and instruct the run-phase to apply both
  edits **before M0 begins**. Neither touches a requirement, a criterion's identity, or the branch
  structure, so applying them is a mechanical pre-flight step, not a re-plan.
- The reason for the condition: unapplied, a run that reaches branch B — the leading outcome — reads
  AC-CSL-009's body, finds `Given` and `When` satisfied and `Then` false, and records **FAIL** on a
  path where nothing failed, with no canonical-on-divergence clause to override it. That is a live
  wrong-verdict risk, not a cosmetic one. With the clause text carried verbatim in the debt entry, the
  risk is a two-line edit ahead of M0.

A sixth iteration is not warranted: both items are single clauses with prescribed text, and each is
verifiable by reading the line it lands on.

**On the base rate.** Five repair rounds, five plants — N1, then X2 and X3, then Y1, now Y2 — every
one the same shape, a new rule leaving a neighbouring sentence stale, and each smaller than the last
(Y2 is a false qualifier in a justification, affecting no disposition). The author's refusal to read
their own clean sweep as evidence of no plant is the right posture and is what surfaced item 2; note
that item 2 was found by them and that my role here was to reject their explanation, not their
observation.
