# SPEC Review Report: SPEC-CODEX-GHOST-SKILLS-PRUNE-001

Iteration: **3** (delta re-audit — ceiling exception approved by the lead; Tier M
ceiling is 2, `.moai/config/sections/harness.yaml:77`)
Scope: **E1 and E2 only**, per the [HARD] delta instruction.
Verdict: **FAIL**
Overall Score: **0.81**
**STOP signal raised** — see § Score trajectory and STOP.

Prior verdicts intact: `plan-audit-verdict.md` (iter-1, FAIL 0.76),
`plan-audit-verdict-iter2.md` (iter-2, FAIL 0.83).

Reasoning context ignored per M1 Context Isolation. Every claim in the brief was
treated as a claim to check. **One of them is not accurate** — see F1.

Tree: `.claude/worktrees/t506`, HEAD `ace1c5440`, working tree carrying only the two
untracked card directories (`git status --short` → `?? .moai/reports/t506/`,
`?? .moai/specs/SPEC-CODEX-GHOST-SKILLS-PRUNE-001/`).

---

## Headline

**E1 is fully resolved.** I re-tested it against the exact case I constructed, plus
the three fresh-conflict directions the brief named, and found no residue.

**E2's headline harm is closed and now rests on executed evidence rather than my
hand-walk.** The criterion moved from line appearance to parser state, and
AC-CGP-003 rows 7-d / 7-d' genuinely discriminate the two readings — I verified that
by walking a text-rescanning implementation and a state-based implementation through
the fixture separately and getting opposite outcomes.

**What blocks is one layer down.** The *criterion* is now right; the *carrier* the
plan prescribes for it is narrower than the criterion it must serve, and as written
it makes two of the three original hazards undetectable (F2). And the fixture that
does the discriminating has no mutant behind it, in a SPEC whose own doctrine
demands one for exactly this shape (F3).

---

## E1 — REQ-CGP-005 demotion + REQ-CGP-022 precedence

**RESOLVED.** Checked four ways.

**Does REQ-CGP-022 resolve my exact case?** The case was: `enabled = true` + path
absent + an unrecognised line in the span. It is now resolved **twice over**, and the
first of the two is the load-bearing one:

1. `spec.md:73` REQ-CGP-005 is now a bare Unwanted clause — "시스템은 `enabled` 값을
   적격 판정의 게이트로 사용해서는 안 된다" — and the removal obligation is gone from
   the requirement entirely. The rationale paragraph at `:75` states it explicitly:
   "**이 조항은 제거를 지시하지 않는다.**" With no `shall remove` anywhere in §C.2
   except REQ-CGP-004's necessary-condition form, there is no longer a clause for
   REQ-CGP-018 to contradict. The contradiction is dissolved, not adjudicated.
2. `spec.md:81` REQ-CGP-022 then adjudicates any residual case: "§C.3 의 실격 조항
   (REQ-CGP-006..011, REQ-CGP-018) 중 하나라도 걸리는 항목에 대해 … 그 항목을
   **보존해야 한다**. 실격이 적격을 이긴다."

Note for the record that REQ-CGP-022 currently has nothing to override — after the
demotion, no §C.2 clause commands removal. That is not a defect: it guards the next
edit, which is exactly the edit that created E1.

**Did the demotion lose the "enabled is not a gate" meaning?** No. That meaning was
always carried by the *first* sentence, and that sentence is now the whole
requirement, unchanged in substance from v0.1.0's intent. What was demoted is the
sentence the MP-2 repair added, which the rationale paragraph preserves verbatim
alongside its reason (`spec.md:77`). AC-CGP-005 still tests the axis
(`acceptance.md:80-84`), and `§D.2` still maps REQ-CGP-005 → AC-CGP-005.

**Fresh conflicts?** Checked the two the brief named plus one more:
- **vs REQ-CGP-004's "…만 제거해야 한다"** — compatible. REQ-004 restricts *removal*
  (necessary conditions); REQ-022 mandates *preservation*. A clause that narrows
  removal and a clause that mandates preservation point the same direction; they
  cannot disagree about any entry.
- **vs REQ-CGP-017's fail-open** — compatible. REQ-017's posture is "change nothing",
  which satisfies a preservation obligation trivially. No state exists where one
  requires a change the other forbids.
- **vs §C.3's own framing** — `spec.md:89` still says "일곱 조항이 안전 경계다", and
  §C.3 does hold exactly seven (006-011, 018). REQ-022 lives in the new §C.2.1 and is
  not miscounted into that seven.

---

## E2 — parser-state recognition

### The executed counterexample

The author executed the fixture I had listed under Gaps as unexecuted, and the result
matches my hand-walk exactly: `ENTRIES=1`, `path="/gone"`, with the narrow
header-less variant giving the same. That converts my reading of
`skills.go:83-90` + `:106-143` into a measurement, and it is recorded in three places
with its date and tree — `spec.md:103` (`ace1c5440`, 2026-09-07), `plan.md §B.3:56`,
`acceptance.md` row 7-d. I did not re-execute it; my attribution for that figure is
the author's run, and I say so rather than adopting it as my own measurement. The
scratch file is gone: `ls internal/codexwiring/ | grep -i 'zz_\|t506'` → no match,
rc 1.

### Q2 — does row 7-d discriminate the two readings? **Yes. Verified by walking both.**

Fixture, `acceptance.md:83-90`, with the Then at `:94`: after the run, L3/L4/L5 must
remain byte-identical.

- **Text-rescanning implementation.** Span L0..L5 re-scanned by line appearance:
  header / `path` / comment / header / `path` / comment — all six are among the five
  recognised shapes, so nothing disqualifies, the span is deleted, and L3-L5
  disappear. **FAILS 7-d.**
- **Parser-state implementation.** `multilineOpener` (`skills.go:83-90`) counts `"""`
  across the whole line including comments, so L2 opens a literal; the `openDelim`
  branch (`:113-121`) consumes L3 and L4 without ever entering the switch. Those two
  lines were therefore never consumed as any of the five kinds → unrecognised →
  disqualified → nothing is deleted. **PASSES 7-d.**

The claim at `acceptance.md:96` ("텍스트 재훑기로 구현한 프루너는 통과하지 못한다")
is true. 7-d' (`:98`) additionally kills the cheap "one header per span" patch, since
its swallowed region carries no header at all — I checked that variant separately and
it discriminates the same way.

### Q4 — do hazards (b) and (c) survive the state-based rule?

**Under the requirement as stated in `spec.md:95`, no — both are still covered.**
A line the parser did not consume as one of the five kinds is unrecognised, and
neither an unknown key nor an array-opening assignment is consumed as any of them.
For (c) I re-confirmed the mechanism independently: an `["x", "y"]` continuation line
can only follow an opening assignment (`args = [`), that opener sits inside the span,
and it is not one of the five — so the entry is disqualified before `anyTableRe`'s
over-match (`configtoml.go:76`) can close the span early.

**Under the carrier the plan prescribes, yes — both fall open.** That is F2.

---

## Defects Found

**F1. No requirement obliges the parser to report anything; the reporting decision
lives only in `plan.md`** — `spec.md:93-97` vs `plan.md §B.3:60` — Severity: **major**
— Class: **optional**.
The brief states "The requirement says the parser must report it." I checked: it does
not. `spec.md:95` states the *criterion* ("인식 여부는 파서의 상태로 판정하며") and
`:97` adds the `[HARD]` swallow clause, but no REQ requires `ParseSkillEntries` or
`SkillEntry` to expose the fact, and no REQ forbids the pruner from re-deriving it.
The obligation exists only at `plan.md:60` — "파서가 … boolean 으로 보고한다. 프루너는
그 플래그를 읽을 뿐 범위의 텍스트를 다시 해석하지 않는다".
I classify this **optional**, not blocking, and I want to be explicit about why,
because the brief anticipated the opposite. A pruner that faithfully re-derives the
state (re-running the same `openDelim` bookkeeping over the span) reaches the *correct*
answer and **passes 7-d** — I walked it. So re-derivation is an architecture-hygiene
problem (the duplication `plan.md:50` rejects as B-2, and the shape REQ-CGP-001 bans
on the classifier axis), not a correctness hole, and the fixture already stops the
incorrect reading. Escalating it to blocking after the author fixed the real thing
would be manufacturing a defect.
Suggested fix (one clause, pairs with F2): add to §C.3 — "시스템은 항목별 인식 판정을
파서가 산출한 값으로 받아야 하며, 프루너가 범위의 텍스트를 다시 해석해 그 값을
재유도해서는 안 된다."

**F2. The carrier `plan.md §B.3` prescribes is strictly narrower than the criterion
REQ-CGP-018 states — and with §B.3's own no-text-rescan clause, hazards (b) and (c)
become undetectable** — `plan.md:60` vs `spec.md:95` — Severity: **major** — Class:
**blocking**.
§B.3's decision is a **single boolean per entry**: "이 범위 안에 `openDelim` 분기
(`skills.go:113-121`)가 소비한 줄이 있었는가". That answers hazards (a) and (d) — the
literal-swallow family — and nothing else. Walk the other two against it:
- **(b) unknown key.** `foo = "x"` inside an entry: no literal is open, so the boolean
  is `false`. Reading the boolean alone, the entry is "clean" → eligible → deleted →
  the assignment re-parents onto the preceding table. The exact loss `spec.md:99`
  clause (b) describes.
- **(c) array opener.** `args = [` inside the span, `["x", "y"]` closing the span
  early: again no literal, boolean `false` → eligible → deleted → the continuation
  line is orphaned.
The pruner cannot recover either case without re-scanning the span's text, which
`plan.md:60`'s second sentence forbids. **As specified, a compliant implementation is
impossible**: the boolean cannot decide what the requirement asks, and the only other
route is prohibited.

There is a second, sharper reason the boolean is the wrong shape, and it is the same
overstatement iter-1's D2 caught and iter-2's E2 caught, appearing a third time.
`plan.md:60` says "사실을 아는 유일한 주체가 파서이므로, 추론하게 두지 않고 보고하게
만든다" — but for (b) and (c) **the parser does not know the fact either.** Read
`skills.go:127-140`: a blank line, a full-line comment, and an unknown key assignment
all fall through the same `case inEntry:` arm, match neither `skillPathKeyRe` nor
`skillEnabledKeyRe`, and do nothing. The parser's present control flow **cannot
distinguish two of the five recognised kinds from the unrecognised kind** — they share
one no-op path. So "report what the parser knows" is not available; the parser has to
be taught the distinction first, exactly as `plan.md:56` already says about the extent
("기록만 하면 된다"가 아니라 없는 것을 새로 만드는 일).
Required fix: replace §B.3's boolean with a per-entry verdict covering **all five
kinds plus the swallow branch** — e.g. `AllLinesRecognized bool` (or the index of the
first unrecognised line, which is strictly more useful for the skip report
REQ-CGP-018 requires at `spec.md:113`) — computed inside the parser by classifying
every line in the extent, including those the `openDelim` branch consumes. Then
§B.3's no-text-rescan sentence becomes satisfiable rather than contradictory.

**F3. The fixture that does all the discriminating has no mutant behind it** —
`acceptance.md:19-100` vs `acceptance.md:64-79` — Severity: **major** — Class:
**blocking**.
`acceptance.md:96` *asserts* that a text-rescanning pruner cannot pass 7-d. I verified
the assertion is true, so this is not a correctness dispute — it is an evidence
dispute. No AC pins a mutant on the recognition predicate at AC-CGP-004's strength:
AC-CGP-004's Given/When (`:66-68`) name the indeterminate class and the eligibility
branch specifically, and `§D.3:216`'s mutant-evidence bullet cites only AC-CGP-004.
The mutant matters more here than it usually would, and for a reason the SPEC itself
articulates two lines away: **7-d's PASS state is "nothing was deleted"**, which is
also what a mis-built fixture produces. If the comment's `"""` count is even, or the
fixture is adjusted so `ParseSkillEntries` reports two entries, 7-d passes while
proving nothing — the empty-operand pass `acceptance.md:100` prohibits by name. The
positive firing evidence is already in hand (the executed `ENTRIES=1`), which makes
this cheap to pin.
Required fix: add AC-CGP-017 at AC-CGP-004's strength — **diff site**: the recognition
predicate, mutated from parser-state to line-shape; **discriminant**: the named test
covering 7-d and 7-d' appears in the FAIL list by name; **both observations**
recorded; and a **precondition** that the fixture makes `ParseSkillEntries` report
exactly one entry with `path="/gone"` (the fixture's own firing evidence). Then list
it in `§D.1` and `§D.2` under REQ-CGP-018.

**F4. `§D.2`'s completeness sentence still names the old counts** —
`acceptance.md:209` — Severity: **minor** — Class: **blocking**.
"이 표는 **양방향**으로 완결이다. REQ **21개**가 모두 왼쪽 열에 있고, AC **15개**가
모두 오른쪽 열 어딘가에 있다." The counts are now 22 and 16. **The table itself is
correct** — I verified both directions by enumeration (below) — so this is a false
sentence above a true table, which is the same defect class as E3 (a count sentence
not updated when rows were added), recurring one section away from where E3 was fixed.
Two digits. I mark it blocking only because it is an assertion of verification that is
now untrue, and this SPEC's own doctrine is that an unverified claim is a defect
whatever its size.

**F5. `spec.md §C`'s numbering note still describes a 21-REQ document** —
`spec.md:57` — Severity: **minor** — Class: **optional**.
"그래서 **001-021** 이 문서 순서대로 나오지 않는다(018 은 §C.3, 019 는 §C.5, 017 은
§C.7 에 있다)" — REQ-CGP-022 exists and sits in the new §C.2.1, which the parenthetical
does not mention. Same stale-count family as F4, without the verification claim.

**F6. AC-CGP-016's own subject traces to no requirement** — `acceptance.md:163-171`,
mapped at `:203` — Severity: **minor** — Class: **optional**.
The EOF AC is mapped under REQ-CGP-019 (line-ending state) — correct for its *two
variants*, but its actual subject is "EOF is a close", which is `plan.md §B.1` item 2
and has no REQ. So the AC is not an orphan (D3 has not recurred), but its primary
assertion is untraced. Outside the delta scope — the E4 repair introduced it, no E1/E2
edit touched it — reported for the record, not as a gate.

**F7. Row 7-d's fixture instruction is mildly self-contradictory** —
`acceptance.md:83` vs `:92` — Severity: **minor** — Class: **optional**.
"아래 7줄을 **그대로 쓴다**" against "`/gone` 은 존재하지 않고 `/exists` 는 **실재하도록**
픽스처를 준비한다" — a temp-dir path cannot be both verbatim and real. It does not
weaken the AC (the Then asserts byte survival of L3-L5, which holds whether or not
`/exists` resolves, since the parser never sees that entry at all), but the run phase
will have to pick one reading. Say which: substitute both paths, keep the shape.

---

## What I checked and cleared

- **Counts, both directions, by enumeration rather than by reading the claim.**
  `grep -o '\*\*REQ-CGP-[0-9]*\*\*' spec.md | sort | uniq -c` → 22 declarations, every
  count `1`, `001..022` contiguous. `grep -c '^### AC-CGP-' acceptance.md` → `16`;
  unique ids `AC-CGP-001..016`, contiguous. **Left column** of `§D.2` names all 22
  REQs including `REQ-CGP-022` (`:207`). **Right column**, collected across all rows:
  {001,002,003,004,005,006,007,008,009,010,011,012,013,014,015,016} — all 16,
  **AC-CGP-016 included** (via the REQ-CGP-019 row, `:203`). **No orphan. D3 has not
  recurred.** The lead's independent recount (22 / 16, contiguous) matches mine.
- **E3 did not regress under the E1/E2 edits.** `acceptance.md:21-25` now reads
  "**여덟 행**(부류로는 일곱)" in the Given, "그 **여덟 행의 모든 항목**" in the Then,
  and `:25` states which unit is being counted. Row 7 was extended from three variants
  to four (a/b/c/d) and the Given was updated in the same pass to say "네 변형" — the
  count and the table moved together this time, which is the specific failure E3 named.
- **E4/E5/E6 survived the E1/E2 edits.** AC-CGP-016 exists (`:163`) with two
  newline variants; AC-CGP-014 now carries the CRLF variant (`:159`) and explains the
  composition axis; `plan.md §F`'s risk table was not re-broken by the §B.3 insertion.
- **The three restatements of the executed evidence agree with each other** and with
  the mechanism: `spec.md:105`, `plan.md:56`, `acceptance.md:92` all report
  `ENTRIES=1` / `path="/gone"`, all name the same cause (`multilineOpener` counting
  `"""` without stripping comments, `skills.go:83-90`), and all record tree and date.
  No overclaim: none of the three says the pruner was tested, only the parser.
- **The struck-through false sentence is preserved, not deleted.** `plan.md:32` keeps
  the v0.2.0 claim under `~~strikethrough~~` with the correction beside it, and
  `spec.md:77` does the same for the demoted REQ-CGP-005 sentence. Both cite the
  finding that caught them. That is the right shape for a record a later reader will
  audit.
- **Mechanical gates re-run in this tree**: `moai spec lint .../spec.md` → rc 0,
  `✓ No findings` (frontmatter and REQ-id axes only — the modality axis is vacuous for
  a Korean SPEC, per `lint.go:790`, and the SPEC now says so in three places).
  `grep -rn 'NEEDS CLARIFICATION'` → rc 1. `grep -c syscall spec.md` → 0.
  Frontmatter unchanged and still 12/12.
- **Must-pass**: MP-1 PASS (22 contiguous, no dup), MP-2 PASS (I re-read REQ-CGP-022 —
  "시스템은 … 보존**해야 한다**", Ubiquitous, modal — and re-read the demoted
  REQ-CGP-005, which is now a clean Unwanted clause), MP-3 PASS, MP-4 N/A, MP-5 PASS
  (reference set unchanged), MP-6 PASS, MP-7 PASS.

---

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | E1's contradiction is gone and the precedence rule is now explicit rather than inferred (`spec.md:81`). Deducted for F2: `plan.md:60`'s two sentences cannot both be honoured. |
| Completeness | 0.80 | 0.75-1.0 | Precedence clause, EOF AC, CRLF variant, and the swallow rows all added. Deducted for the missing parser-reporting requirement (F1) and the missing mutant AC (F3). |
| Testability | 0.75 | 0.75-1.0 | 7-d / 7-d' are strong, executed-evidence-backed, and genuinely discriminating — I verified both readings against them. Deducted because the discriminator's PASS state is "nothing happened" with no mutant behind it (F3). |
| Traceability | 0.85 | 0.75-1.0 | Two-way map verified complete by enumeration; REQ-CGP-022 and AC-CGP-016 both absorbed. Deducted for F4 (false count sentence over a true table) and F6 (AC-016's subject untraced). |

Aggregate = **0.8125 → 0.81**, above the Tier M threshold of 0.80.

---

## Score trajectory and STOP

0.76 (iter-1) → 0.83 (iter-2) → **0.81 (iter-3)**. This is a **regression**, so per the
retry-loop contract's score-regression clause I raise **STOP** rather than recommend
another unconditional iteration.

The regression is real, not a scoring artifact, and it has a specific shape worth
naming: **each repair round closed its target and opened a smaller one adjacent to
it.** Traceability fell from 1.00 because the E3-class count defect reappeared in
`§D.2`'s summary sentence; Completeness fell because the E2 fix introduced a carrier
that is narrower than the criterion it serves. Both new findings are cheaper than the
ones they replaced — the amplitude is decaying — but three rounds of this pattern is
the signal the STOP clause exists to surface.

---

## Baseline-attribution

All figures produced in this run against this tree (`ace1c5440`,
`.claude/worktrees/t506`); source line numbers read from the working tree at that
HEAD; `moai spec lint` run with `/tmp/t506-audit-moai`, built from this tree in
iteration 1 (`go build -o /tmp/t506-audit-moai ./cmd/moai`, rc 0) — the tree has not
moved since (`git rev-parse --short HEAD` → `ace1c5440`). The `ENTRIES=1` measurement
is **the author's**, cited with its stated tree and date, not re-run by me. The
population figures in `spec.md §A.1` remain cited from
`.moai/reports/t506/baseline-measurement.md`.

## Gaps — what I did NOT observe

- I did not re-execute the counterexample. My walk of it is a reading of
  `skills.go:83-90` and `:106-143`; the author's execution is the measurement, and I
  did not independently reproduce it.
- I did not execute the two implementation walks in Q2 — they are readings of the
  parser's control flow against the fixture, not runs. No pruner exists to run.
- No Go tests run; no implementation exists.
- Per the [HARD] delta scope I did **not** re-audit E3/E4/E5/E6 on their merits — only
  far enough to confirm no E1/E2 edit broke them. F6 and F7 surfaced incidentally and
  are reported without a full re-audit of the sections they sit in.
- I did not re-verify the five referenced SPECs' `status:` (reference set unchanged
  since iteration 1) and did not re-measure this machine's `~/.codex/config.toml`.
- I did not search `internal/cli` for a pre-existing stat seam; `plan.md:140` still
  correctly records that as M2's first task.

## Residual risk

- **The pattern behind F2 is now three-for-three.** Iteration 1's D2, iteration 2's
  E2, and now F2 are the same mistake at successively smaller scales: a rule written
  about what the *text looks like*, or about what the parser is *assumed to already
  know*, when the deciding fact is something the parser must be taught to compute. The
  durable fix is not another patch at this layer — it is to make the parser emit a
  per-entry recognition verdict over every line in the extent, once, and let
  everything downstream read it.
- **F3's absence is why F2 could still ship silently.** With no mutant on the
  recognition predicate, an implementation that reads only the `openDelim` boolean
  passes 7-d (its span *does* contain swallowed lines) while failing hazards (b) and
  (c) — and rows 7-b and 7-c would catch that only if their fixtures are built
  precisely. A mutant makes the discriminator self-checking.
- The outer risk is unchanged and correctly bounded: byte-preservation over arbitrary
  hand-edited TOML. REQ-CGP-018's disqualification remains the right posture — a
  disqualified entry is a surviving entry — and the residue shrinks with each round.

---

## Recommendation

FAIL on two blocking findings (F2, F3) plus one two-digit correction (F4). E1 is
resolved; E2's harm is closed and evidenced.

Because the score regressed, I am **not** recommending a fourth auditor spawn. The
remaining work is three edits, each mechanically checkable by reading the diff, and a
fourth full audit would cost more than it can find:

1. **F2** — `plan.md §B.3:60`: replace the single `openDelim` boolean with a per-entry
   verdict over **all five kinds plus the swallow branch** (`AllLinesRecognized`, or
   the first unrecognised line index — the latter also feeds REQ-CGP-018's required
   skip report). Acceptance for the lead's diff read: the sentence must name a value
   that can decide an unknown key and an array opener, not only a swallowed line.
2. **F3** — add **AC-CGP-017**, a mutant AC at AC-CGP-004's strength: diff site = the
   recognition predicate mutated to line-shape; discriminant = the 7-d/7-d' test named
   in the FAIL list; both observations recorded; precondition = the fixture makes
   `ParseSkillEntries` report exactly one entry. List it in `§D.1` and `§D.2` under
   REQ-CGP-018.
3. **F4** — `acceptance.md:209`: `21` → `22`, `15` → `16`.

F1, F5, F6, F7 are optional; surface them, do not gate. F1's one-line REQ is worth
taking in the same pass since it pairs with F2's fix.

**Disposition I recommend to the operator**: apply these three edits and have the lead
verify the diff against the three acceptance conditions written above — **no fourth
auditor iteration**. If the operator prefers to close now instead, F4 alone is
acceptable as documented debt; F2 and F3 are not, because F2 leaves a specification
that cannot be satisfied as written and F3 leaves the repair's only discriminator
unverified.
