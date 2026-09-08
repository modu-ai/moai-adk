# SPEC-CODEX-SKILL-LOADER-001 — confirm-only pass (iter-4)

Scope: confirm-only. No re-audit, no rescoring, nothing re-derived. One question answered — are
X1-X5 closed, and did closing them plant anything new.

**Result: X1-X5 all closed. Both self-swept edits confirmed. Both residue claims verified with live
control groups. ONE new defect planted (Y1), one line wide.**

Debt-closed recording: **withhold until Y1 lands.** Plan-phase close: **not structurally blocked** —
Y1 is a single-line correction inside the row that introduced it, verifiable by reading that one
line, and it needs no further audit round.

## Hashes read

Verified before judging; all four match the frozen set:

| artifact | sha256 |
|---|---|
| spec.md | `b19d920e0c3bc50fa1ffbd7b55027767a3e8ac060c53aac131c50239ca2d9e8f` |
| plan.md | `562c0872d2045ee57847b9f170821075ff55797939144da53d566242dcac5599` |
| acceptance.md | `151eed57482892eb98563bfa390d8f4570aa417a18783fb4125799f677b4c41e` |
| progress.md | `f8b7844e27cb8c0a07c8ff557586dd826dbe2512986f94df18cf53541abb0ef9` |

## (a) X1-X5

**X2 — CLOSED, and N2 is now closed for real.** `plan.md:105-107` carries the three arms verbatim
(R1 loads → A; R1 does not load AND some root fired → B; no root fired → `inconclusive`, no branch),
and `plan.md:109` is the [HARD] anti-collapse clause with the reason stated — the sentence read at
the moment of recording is the instruction, so a stale instruction beats the rule above it.
I swept for any surviving sentence that routes a no-fire run to branch B and found **none**:
`grep -n '아니면'` across all three artifacts returns three hits, and all three are non-routing —
`spec.md:23` (the HISTORY entry describing the defect), `plan.md:109` (the prohibition itself, which
must contain the phrase to forbid it), and `plan.md:127` (unrelated prose about 오값 거동).
`spec.md:149` (REQ-CSL-002) also gained the `inconclusive` arm, so the requirement layer and the
instruction layer now agree.

**X1 — CLOSED.** `acceptance.md:62` is scoped ("분기 A 안에서 상호 배타적", "분기 A 회차에서"), and `:63`
adds the [HARD] statement that on branch B and `inconclusive` both are 해당 없음 and not a violation,
grounded exactly where I asked — in REQ-CSL-006's `Where 분기 A 가 확정되고` and REQ-CSL-007's
`분기 A 이지만…` guards being unmet.

**X3 — CLOSED.** `acceptance.md:22` restates the positive control as any-root ("다섯 뿌리 중 어느
하나에서든 신호가 발화했을 때"), explicitly including R1 and saying why R1 firing is itself proof the
instrument works. `plan.md:96-97` matches: any-root satisfies the control, R2 demoted to the
all-silent fallback prior, with the observation that the R2 question "어느 뿌리도 발화하지 않았을 때만"
arises.

**X5 — CLOSED (with Y1 inside it).** The completion definition gains a dedicated `inconclusive` path
row: AC-CSL-002 = `inconclusive`, AC-CSL-001 = PASS, the remaining 11 = 해당 없음, recorded in
`progress.md` §E.2. Arithmetic checks (13 − 2 = 11, AC-CSL-003~013). It also names the deliverable
and — the part I did not ask for and that is right — states explicitly that the M0 blocker report and
AC-CSL-003's (a)-(e) branch-B report are **different documents**, so the (a)-(e) list is not silently
applied to a path it was not written for.

**X4 — CLOSED.** The restriction is folded **into** the enumerating bullet ("…중 하나로 기록되어 있다
(**AC-CSL-002 에 한해 `inconclusive` 를 추가로 쓸 수 있다**)"), with a sub-bullet stating the reason the
ordering matters: opening the state set wide and narrowing afterwards lets the permissive sentence win.
That is the correct fix rather than the cheap one.

## (b) The two self-swept edits

**Sweep edit 1 — CONFIRMED in both places, and it was a real second X3.** The exhaustion condition now
reads any-root in `plan.md:99` ("3 회 뒤에도 **어느 뿌리에서도 발화하지 않으면**(R2 만이 아니다 — 양성
대조는 어느 뿌리로도 충족되므로, 소진 조건도 같은 기준이어야 한다)") and in `acceptance.md:25` with the
same parenthetical. The blocker-report contents at `plan.md:99` were updated to match ("다섯 뿌리 전부의
미발화 사실(R2 포함)"). Left unfixed this would have let a run where R1 fired exhaust reselection on R2's
silence alone — the author's own reading of it is correct.

**Sweep edit 2 — CONFIRMED.** `acceptance.md:32` names `inconclusive` in AC-CSL-003's own 적용 조건
("분기 A 이면, **그리고 `inconclusive` 회차에서도**…"), with the recording surface split by case (§E.3 for
branch A, §E.2 for `inconclusive`) and the reason given — the AC's `Given` does not hold in either case.
The reasoning for not leaning on a single completion-definition line is right: that is precisely X5's
shape reproduced one layer up.

**Both residue claims verified independently, each with a live control** rather than a bare zero:

- `R2 가 발화하지 않으면` → **0** in all three files. Control: `발화` → 8 / 4 / 5 occurrences
  (plan / acceptance / spec), so the selector reaches live text and the zero is a real absence.
- base-less `git diff` → **0 selectors**. My inverted filter returns two hits, and both are
  *prohibition prose*, not selectors (`acceptance.md:45`, `plan.md:34`) — the same shape the lead
  reported. Controls: total `git diff` occurrences 8, `BASELINE_SHA` occurrences 7. All four actual
  selectors carry `"$BASELINE_SHA"`: `acceptance.md:31`, `:44`, `:86`, `:93`.
  Noted in passing: `:86` (AC-CSL-011) now carries an explicit
  `git diff --name-only "$BASELINE_SHA" -- '*.go'` where iter-2 found only prose. That goes beyond
  what X1-X5 required and closes the weakest of N1's three sites.

## (c) NEW — planted by this pass

**Y1 — the `inconclusive` row drops AC-CSL-010, contradicting the branch-B row's own stated principle — `acceptance.md` 완료 정의, `inconclusive` 경로 row — Severity: major.**

The new row assigns 해당 없음 to all eleven of AC-CSL-003~013. For AC-CSL-010 that is wrong on the
document's own reasoning:

- AC-CSL-010's `Given` (`:80`) is "M0 진입 전에 해시를 떠 두었을 때" and its `When` (`:81`) is "모든
  프로브가 끝난 뒤 같은 방식으로 다시 뜨면". On an `inconclusive` run **both hold** — the probes ran; the
  row itself records AC-CSL-001 = PASS for exactly that reason.
- The branch-B row justifies AC-CSL-010·011 = **PASS** with "프로브 실행만으로도 주어가 있다". That
  justification applies verbatim to an `inconclusive` run, so the two rows now state opposite
  dispositions for the same criterion under the same condition.
- The cost is not bookkeeping. AC-CSL-010 is the judging criterion for HARD constraint (a) — the dev
  repo's `~/.codex` is never written. The `inconclusive` path is the one with the **most** probe
  executions (up to four full five-root sweeps plus negative controls, per the 3-reselection bound),
  and this row is the only place that would have confirmed the isolation held across all of them.

AC-CSL-011 = 해당 없음 on this path is **correct** and should not be changed — its subject is changed
Go source, and an `inconclusive` run changes none. Only 010 is misplaced.

**Required fix, one line:** `AC-CSL-010 = PASS`(프로브 실행만으로도 주어가 있다 — 분기 B 행과 같은 근거),
`AC-CSL-011 = 해당 없음`, 나머지 10 개(AC-CSL-003~009, 012·013) = 해당 없음. Adjust the count wording
from 11 accordingly.

Nothing else was planted. One cosmetic drift, **not** a defect and needing no edit: the
`inconclusive` deliverable sub-bullet summarises `plan.md:99`'s blocker contents as "R2 미발화" where
`plan.md:99` now says "다섯 뿌리 전부의 미발화 사실(R2 포함)". The sub-bullet names plan.md as the
authority for that list, so the summary cannot diverge in effect.

## (d) Debt-closed and plan-phase close

**Do not record debt-closed yet — but the remaining gap is one line, not a round.** N1, N2 and N3 are
all genuinely closed as of these bytes (N2 by X2, verified by the `아니면` sweep above). X1-X5 are
closed. The only open item is Y1, which sits inside the row X5 added.

**Plan-phase may close once Y1 lands.** It does not warrant another audit iteration: the fix is a
single disposition in one row, its correctness is decided by reading AC-CSL-010's own `Given`/`When`
against the branch-B row's justification, and no requirement, criterion identity, or branch structure
moves.

On the base rate the lead flagged: it held again — iter-2's repair planted N1, iter-3's planted X2 and
X3, and this pass planted Y1. Each was the same shape (a new rule leaving a neighbouring sentence
stale), and each was smaller than the last. The author's own post-repair sweep caught two of the three
this time before I saw them, which is the first pass where that happened.
