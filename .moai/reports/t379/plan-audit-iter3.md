# SPEC Review Report: SPEC-BOARDLOCK-ERRNO-001 — iteration 3 (terminal)

Card: t379
Iteration: 3/3 — **final iteration permitted by the Retry Loop Contract**
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t379`, branch `WT-boardlock-errno`, HEAD `3f03d9c36`
(`git rev-parse --show-toplevel` → the worktree above; `git branch --show-current` → `WT-boardlock-errno`;
`git rev-parse --short HEAD` → `3f03d9c36`; `git status --porcelain` → exactly two untracked entries,
`.moai/reports/t379/` and `.moai/specs/SPEC-BOARDLOCK-ERRNO-001/`)

Verdict: **PASS-WITH-DEBT**
Overall Score: **0.83** (harmonic mean; Tier S PASS threshold 0.75)
Delta vs iter-2 (0.83): **0.00 — flat**
Delta vs iter-1 (0.85): **−0.02 — the regression is NOT recovered**

Reasoning context ignored per M1 Context Isolation. The repair pass's own account of its twelve edits
was read as a set of claims to re-verify against the files, never as findings to adopt.

Scope: delta-scoped re-audit over the enumerated iter-2 defects N1-N9 plus the edited regions the lead
named, per the Retry Loop Contract. Must-pass criteria were re-run because their inputs sit inside the
edited regions.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-BLE-001`…`005` at `spec.md:114, 118, 124, 128, 132`.
  Sequential, no gaps, no duplicates, uniform zero-padding. iter-3 edited REQ-001/002/005 bodies, not
  their numbers.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`) only; the six Given-When-Then entries in `acceptance.md` are verification-layer artifacts
  and are graded under Group 4 per M3 § Scope. 001 / 002 event-driven (`When …, the <subject> shall …`);
  003 / 005 ubiquitous; 004 state-driven (`While …, … shall …`). `spec.md:132`: *"The observable
  behaviour of `AcquireBoardLock` and of every `IsBoardLockHeld` consumer **shall** be unchanged on
  every measured-reachable input…"* — clean ubiquitous form.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical names at
  `spec.md:2-13`; `version: "0.3.0"` (quoted semver), `created`/`updated: 2026-08-31` (ISO), `status:
  draft`, `priority: P3`, `lifecycle: spec-anchored`, `tags` a comma-separated string. No rejected
  snake_case alias (`created_at`/`updated_at`/`labels`/`spec_id`) anywhere. `plan.md` /
  `acceptance.md` / `progress.md` carry no `status:` field (lead-established count: 1/0/0/0).
- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC (Go, `internal/kanban`).
  Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — re-run this iteration.
  Extracted set: `SPEC-KANBAN-BOARD-001`, `SPEC-STRESS-INVARIANT-VERDICT-001`. Both exist on disk;
  `grep -m1 '^status:'` → `completed` and `implemented` respectively. Neither is
  `retired`/`superseded`/`archived`. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears at `spec.md:82` (prose *"생 syscall
  래퍼"* + the filename `zsyscall_darwin_arm64.go`), `plan.md:83`, `acceptance.md:147`. The build-tag
  obligation is present and load-bearing in four places: `plan.md:40` (`//go:build !windows`,
  `GOOS=windows` build mandatory), `plan.md:110` (new test file carries the tag), `spec.md:200`,
  `acceptance.md:102`/`:140`. No BLOCKING finding.
  *Observation, not a finding*: `spec.md` itself carries no `//go:build` literal, so the D8 verb read
  strictly file-by-file over `spec.md` alone would fire. The discipline is substantively present across
  the artifact set, which is what lesson #21 exists to secure; I record the technicality rather than
  manufacturing a BLOCKING from it.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BOARDLOCK-ERRNO-001/`
  → no output, rc=1. Tier S carries no `research.md`; `plan.md` scanned and clean.

**No must-pass failure. The verdict is not forced by the M5 firewall.**

---

## Category Scores

| Dimension | Score | Δ vs iter-2 | Δ vs iter-1 | Rubric band | Evidence |
|---|---|---|---|---|---|
| Clarity | 0.80 | −0.05 | −0.05 | 0.75-1.0 | Real gain: the retracted universal claim is gone from `plan.md:8`, `:21`, `:69`, and EINTR joins the unmeasured list at `plan.md:41` (N1 closed, verified by reading each line). The inference/measurement boundary is now labelled consistently. Offsetting, and it costs more than N1 did: **four live internal contradictions**, three of them created or left by this pass — `plan.md:111` orders the very measurement `acceptance.md:26` retracts as impossible (P1); `plan.md:73`/`:118` and `acceptance.md:104`/`:107`/`:116` demand three mutants after the tables withdrew one (P2); `spec.md:31` states the audit found 003 rather than 004 unguarded, contradicting `acceptance.md:36`/`:43` and the iter-2 record (P5); `spec.md:165` reads "004 미피복" two lines above "미피복 REQ 0건" (P7). |
| Completeness | 0.93 | +0.01 | +0.05 | 0.75-1.0 | `plan.md §B.1` is a genuine debt record and says so in its own first line (*"이 절은 부채 기록이지 가드가 아니다"*). Both deferred defects are mirrored where a reader meets them (`acceptance.md:35`, `:42`, `:91`, `:92`, `:118`, `:129`). Six `### Out of Scope — <topic>` H3s at `spec.md:175-191`, each with `-` bullets. HISTORY 0.3.0 row added. Deductions: HISTORY 0.2.0 (`spec.md:26`) still omits the `§D.0` work (N9 open); `§B.1` records AC-BLE-003 as *guard-less* but not as **failing `verification-completeness.md` §2's mutant probe**, which is the stronger and doctrine-cited consequence (P8). |
| Testability | 0.72 | +0.04 | 0.00 | 0.50-0.75 | The four false affirmatives iter-2 penalised are genuinely withdrawn: the `3f03d9c36` RED-now pin (`acceptance.md:26`), the blanket mutant claim (`:39`, `:43`), the M-leak→003 designation (`:114`, `:92`), the RED-now-6 demand (`plan.md:73`). Debt is placed where it will be hit, not in an appendix. But the instrument set is still not executable as written: a milestone instruction points at a tree the same SPEC says cannot yield the measurement (P1); `§E` E8 demands mutant evidence the SPEC says cannot be produced (P2); two ACs (003, 004) remain unguarded, one of them by a mechanism that returns `ok` on the CI runner (N6 debt); `§2.1` is cited while its operative consequence is overridden (P6). Honesty up, executability flat. |
| Traceability | 0.90 | 0.00 | −0.10 | 0.75-1.0 | Mapping is still complete and single-homed: `spec.md §3` maps 001→001a, 002→001b, 003→002, 004→003, 005→004, with AC-BLE-005 as joint non-vacuity; `acceptance.md §D` agrees; 0 uncovered REQ, 0 orphan AC, AC bodies in exactly one file. Gain: N5's unsupported coverage relation is withdrawn. Offsetting: the withdrawal's phrasing introduces P7 — in a column headed 피복 REQ, "004 는 … 현재 미피복" reads as REQ-BLE-004 being uncovered, two lines above `spec.md:167`'s "미피복 REQ 0건". |

Harmonic mean: `4 / (1/0.80 + 1/0.93 + 1/0.72 + 1/0.90)` = `4 / 4.8253` = **0.829 → 0.83**

---

## Answers to the lead's eight questions

### 1. Is iter-3 actually subtraction-only?

**Substantially yes — I found no new guarantee, guard, or mechanism smuggled in — with three
exceptions, all small, and one of them consequential to the pass's own justification.**

I checked every edited region named in the brief for sentences that assert rather than retract. What
the pass added is overwhelmingly of four permitted kinds: retraction (`acceptance.md:26`, `:39`,
`:92`, `:114`), scope-down (`plan.md:73`, `spec.md:165`), debt record (`plan.md §B.1`, `acceptance.md:91`),
and history (`spec.md:27`, `:31`). No new AC, no new REQ, no new mutant, no new mechanism, no new
threshold, no new coverage claim. The `§B.1` header even carries its own anti-guard disclaimer.

The three exceptions:

- **`acceptance.md:36` — an affirmative universal.** The rewritten AC-BLE-004 row ends *"대신 수리
  이전에 실측한 기준선이 없으면 대조할 것 자체가 없어 이 AC 는 공허해진다 — **그것이 이 AC 가 걸려 있는
  유일한 조건이다**"*. "The only condition" is a new, unmeasured universal claim, and it is not
  obviously true: a differential over failure sets is equally vacuous when the compared runs sweep no
  test exercising the three consumers, and equally unreadable when known flake noise moves the set.
  This is the precise defect class — an affirmative claim that does not hold — whose introduction
  produced the iter-2 regression, appearing in the pass whose whole justification is that it adds none.
  (P3.)
- **`acceptance.md:28` — a mechanism claim.** *"RED-now 를 잴 수 있는 트리는 `3f03d9c36` + 새 계약
  테스트(분류 수리 이전 상태)이며"* asserts what the measurable tree **is**. Read literally, base + a
  test file is exactly the configuration `:26` had just condemned two lines earlier: under option A the
  test names `classifyBoardFlockErr`, which does not exist at `3f03d9c36`, so the tree yields a build
  failure, not a RED. The parenthetical *"분류 수리 이전 상태"* may be intended to carry an M1 stub, but
  it does not say so. (P4.)
- **`spec.md:31` — a history record that misstates the record.** *"실제로 가드가 비어 있는 것은
  AC-BLE-004 가 아니라 **AC-BLE-003** 이었다."* iter-2 did not retract its AC-BLE-004 finding; N5
  sustained it (*"Nothing designates 004"*), and N2 raised 003 independently. Both are unguarded, which
  `acceptance.md:36` and `:43` say in this very artifact set. The section claims to assert nothing about
  the code, and that is true — but it asserts something about the audit record, and that is false. (P5.)

**The larger problem with this pass is not addition — it is incomplete subtraction.** P1 and P2 below
are both withdrawals that stopped short of every surface carrying the withdrawn thing.

### 2. N1 — is the retracted universal claim gone from every surface?

**The N1 claim itself: yes, closed. But the same file-boundary failure mode has relapsed on a different
claim, and it is now worse than the version it replaced.**

N1 closed, verified line by line:

- `plan.md:8` — rewritten as a measurement statement mirroring `spec.md §1.3-4`, with the explicit
  disclaimer *"이것은 잰 것에 대한 진술이지 이 호출 지점이 낼 수 있는 errno 에 대한 전수 주장이 아니다"*.
- `plan.md:21` — the Tier table's behaviour-delta cell now reads *"0 — 측정된 도달 가능
  errno(`EWOULDBLOCK`/`EAGAIN`) 한정"* and points at `spec.md §1.3.1` for EINTR.
- `plan.md:69` (was `:62`) — carries the measured-reachable qualifier and closes with *"'오늘 행동
  불변'이라는 넓은 읽기는 철회됐다"*.
- `plan.md:41` — EINTR added to B-미측정.
- Residue grep over all four artifacts for `하나뿐` / `도달 불가 errno 만` / `오늘 행동 불변` /
  `도달 가능한 errno` returned only `spec.md:116` (*"측정된 유일한 errno"* — correctly scoped) and
  `plan.md:69` (the retraction itself).

**The relapse (P1).** `acceptance.md:26` retracts the `3f03d9c36` RED-now pin, giving the reason: at
that tree the observation is a build failure, so *"그 pin 은 잴 수 없는 측정을 가리키고 있었다"*.
`plan.md:111` still instructs, verbatim: *"M2 착수 전에 그 표의 RED-now 셀을 **`3f03d9c36` 에서 실제로
재고 기록한다**"*. The HISTORY row's own wording gives the pattern away — it records *"`acceptance.md` 의
`3f03d9c36` RED-now pin 철회"*, scoped to one file. That is the same sentence shape iter-2's N1 was
written about, one file over.

This relapse is more damaging than the original N1 was, because `plan.md:111` is not a stale label — it
is a **live instruction to a milestone**. An implementer following M2 in order will attempt a
measurement that `acceptance.md` says cannot be taken, and will then either report a build failure as a
RED cell or improvise a tree the SPEC does not describe.

### 3. The RED-now pin — is the wording honest and actionable, or deferred into prose?

**Honest and largely actionable, with one under-specified sentence, and undermined from outside its own
file.**

The choice not to invent a SHA is correct on the merits, and I sustain it. Under
`verification-completeness.md` §2.1 the RED-now cell's fourth element is *"the tree SHA the measurement
was taken on"* — a SHA written before the measurement is not that element, it is its counterfeit, and
§2.1's own evidence row records exactly this failure ("the cells pinned a tree and carried no output,
so the pin named a measurement that had never been taken"). Writing a SHA now would have been the new
assertion the constraint forbade.

Actionability is real, not prose-deep, on three counts: `acceptance.md:47` says the cells are unmeasured
and forbids pretending otherwise; `§D.2:135` makes it a checklist line requiring *"명령 + 출력 전문 +
종료 코드 + 실측 시점의 트리 SHA"* with *"잴 수 없는 트리 pin 금지"*; and the ownership is assigned
(*"그 측정 트리를 만드는 것이 M1 의 일이다"*).

Two deductions. The tree description at `:28` is under-specified in the way P4 records — as written it
names the configuration `:26` just rejected. And `plan.md:111` (P1) actively contradicts the whole
paragraph, which is the difference between a deferral and a hole: a deferral tells the implementer
*when* to measure, and here two artifacts tell them different things about *where*.

### 4. N2 and N6 as debt — are the records adequate?

**Both adequate on the question asked. Neither reads as a guard. One gap remains on N2.**

**AC-BLE-003 reads honestly as having no non-vacuity guard.** It is stated in five places, including
inside the AC itself where an implementer meets it:

- `acceptance.md:35` (§D.0 row) — *"비공허성 공급원으로 적혀 있던 M-leak 은 현 정의로는 발화하지
  못한다 … 이 AC 는 **비공허성 가드 없음**으로 읽는다"*
- `acceptance.md:42` — same, in the narrowed alternative-means declaration
- `acceptance.md:92` (inside AC-BLE-003) — *"**따라서 이 AC 는 현재 비공허성 가드가 없다.**"*
- `acceptance.md:118` (inside AC-BLE-005) — *"REQ-BLE-004 는 지금 비공허성 가드가 없는 요구다"*
- `plan.md:118` (M3) — *"AC-BLE-003 이 M-leak 으로 보증된다고 적지 않는다"*

The mutant table row is struck through with its designation column reading *"(AC-BLE-003 지정이
성립하지 않는다)"*. No surface reads as a guard. This is the correct shape.

**The gap (P8).** The record stops at "no guard". `verification-completeness.md` §2's mutant probe is
stronger than that: *"Before adopting a criterion, try to write a mutant that satisfies the criterion
while violating its requirement. If such a mutant is writable, the criterion is **too shallow to
adopt**."* For AC-BLE-003 such a mutant is precisely what iter-2 identified — remove the close on the
non-contention return path; the contention-only induction at `acceptance.md:89` still passes, and
REQ-BLE-004 is violated. So AC-BLE-003 does not merely lack a guard; by the cited doctrine it is
**unadoptable as written**. Recording that is a subtraction (a stricter debt label), so it was available
under this pass's constraint and was not taken.

**The `/dev/fd` silent-`ok` failure mode is stated where an implementer will hit it.** `acceptance.md:91`
sits directly under the mechanism it qualifies, opens with *"이 기제는 M1 로 넘기는 부채를 안고 있다 —
지금 상태로는 판정을 이루지 못한다"*, and enumerates all three holes iter-2 raised: (a) the unpinned
slack, (b) the platform sentence relabelled as *"어느 쪽에서도 측정되지 않은 추론"* with the ubuntu-CI
mismatch named, (c) *"조용히 `ok` 로 통과하고 이 AC 는 아무것도 단언하지 않는다"*. Mirrored at
`plan.md:48`. Adequate.

### 5. The declined required fix (N4's mutant-probe half) — sustained or withdrawn?

**The refusal is SUSTAINED on the evidence. My iter-2 required fix was defective.**

iter-2 instructed the pass to *"cite §2's mutant probe as the adoption basis for the three
regression-guards (**which they pass**)"*. I read §2 in this tree
(`verification-completeness.md:161-168`). The parenthetical is false, and demonstrably so from iter-2's
own findings issued in the same report: AC-BLE-003 admits a satisfying mutant (N2), so by §2 it **fails**
the probe; AC-BLE-004 has no designated mutant at all (N5), so nothing was measured either way. Writing
that citation would have asserted coverage that does not hold — the exact defect that produced the
iter-2 regression, committed at the audit's own instruction. The pass was right to refuse, and the
refusal is a correct application of the constraint it was given.

Sustaining this is not consistency with my prior position; it is the evidence: `verification-
completeness.md:161-168` is the mutant probe, and `acceptance.md:92` records that 003's designated
mutant cannot fire.

**But N4's other half is open, and the pass did not address it (P6).** N4 was two findings. The second:
`acceptance.md:32` invokes *"§2.1 undecidable disposition 적용"* by name and then declines its operative
consequence. §2.1 (`verification-completeness.md:201-208`) reads: *"the criterion **loses
release-blocking eligibility**, is classified as a **regression-guard**, and is **not recorded as a
pass**. This is the disposition, not one option among several."* The SPEC takes the middle clause and
overrides the outer two — `acceptance.md:129` retains *"착지 차단 성질은 유지된다"*, and `§D.2:134`
requires *"AC-BLE-001a ~ 005 **전부 PASS**"*, which for the three regression-guards is exactly the
recording §2.1 forbids. Separately, §2.1's trigger is a *cited RED that cannot be re-executed*
(historical event / already-merged state / external CI result); a criterion that is green-now is none
of those, so the citation is a misapplication as well as an override.

The deviation is conservative in effect — stricter than doctrine, and `:47` avoids §2.1's actual harm by
refusing to record unmeasured cells as passes. That is why this is a documentation defect rather than a
firewall failure. But a [HARD] clause is named and silently overridden, and any reader routing from this
SPEC inherits the misreading. Declaring the deviation is a subtraction and was available.

### 6. New-defect sweep and citation check on iter-3's edited regions

Regions swept: `spec.md` §1.3 / §1.3.1 / §2.1 / §3 / §4 / §5, REQ-BLE-001 / 002 / 005, HISTORY;
`plan.md` §A / §A.1 / §B / §B.1 / §D / §E / §F; `acceptance.md` §D.0 / §D.1 / §D.2 / §D.3,
AC-BLE-003 / 004 / 005.

New defects: **P1** (retracted pin survives in `plan.md:111`), **P2** (three-mutant demand survives the
withdrawal), **P3** (`acceptance.md:36` universal), **P4** (`acceptance.md:28` tree description),
**P5** (`spec.md:31` misstates the audit record), **P7** (`spec.md:165` coverage ambiguity), **P8**
(§2's mutant-probe consequence unrecorded). **P6** is N4's surviving half, carried forward rather than
new. Full entries below.

Citation check — every code citation in the edited regions re-read against the tree at `3f03d9c36`:

| Citation | Verified |
|---|---|
| `board_lock_unix.go:37` `unix.Open` / `:39` `open board lock %s: %w` / `:41` `unix.Flock` / `:42` `_ = unix.Close(fd)` / `:43` `return nil, ErrBoardLockHeld` | correct (read directly) |
| `board_lock.go:28` `IsBoardLockHeld` | correct — `func IsBoardLockHeld(err error) bool` |
| `board_store.go:173`, `integration_lock_mutation.go:103`, `backlog_store.go:736` (§2.1 consumer table) | all three are `if !IsBoardLockHeld(err) {` |
| `board_store.go:289` `joinBoardReleaseErr`, `backlog_store.go:677` `joinBacklogReleaseErr` (§1.4) | both correct |
| `board_lock_windows.go:69` | `return nil, ErrBoardLockHeld` — consistent with §1.5's characterisation of the narrow substrate |

One imprecision persists (N8, optional): `spec.md:45` labels the fenced block `board_lock_unix.go:43`
while the fence spans `:41-44`. The claim made about `:43` is true at `:43`.

### 7. Score and monotonicity

| Iteration | Score | Δ |
|---|---|---|
| iter-1 | 0.85 | — |
| iter-2 | 0.83 | −0.02 (STOP signal raised) |
| **iter-3** | **0.83** | **0.00 vs iter-2; −0.02 vs iter-1** |

**The regression was not recovered.** Stated plainly: this pass did what it was told and removed four
false affirmatives, which is real repair and shows in Testability (+0.04). It paid that back in Clarity
(−0.05), because two of its withdrawals stopped one surface short and thereby created live
contradictions between artifacts — and a contradictory live instruction (`plan.md:111`) costs an
implementer more than the stale label it replaced.

The LEAN STOP-escalation clause fires on `iter(N+1) < iter(N)`; 0.83 = 0.83, so **no new STOP signal is
raised**. The iter-2 STOP's underlying condition — the score is below iter-1 and has not returned —
persists and is reported here rather than treated as resolved.

### 8. Terminal recommendation

**Accept-with-debt and proceed to the operator Implementation Kickoff Approval gate.**

Not reduce-scope: the SPEC's substance is a two-file defensive narrowing with a bidirectional contract
test and one open design decision (M1 A/B) that is fenced, named, and reviewed first. Nothing in the
remaining defects is a scope problem; every one of them is a sentence.

Not reject: all seven must-pass criteria pass or are N/A, the score clears the Tier S threshold by
0.08, the requirement layer is clean GEARS, traceability is complete, and the two deferred defects are
recorded honestly and in the places an implementer meets them. M6 is explicit that a list of findings
does not by itself manufacture a FAIL, and I will not manufacture one here.

**Two retractions should land before kickoff, and neither needs a fourth audit.** Both are deletions of
the same kind this pass already performed twelve times, and both are verifiable by grep rather than by
judgment — which is the property that makes them safe to apply without re-auditing:

1. **P1** — delete `3f03d9c36` from `plan.md:111`, leaving the instruction pointing at
   `acceptance.md §D.0`'s measure-at-measurement-time rule. Verification: `grep -n '3f03d9c36' plan.md`
   returns only `:5` (the base-ref fact).
2. **P2** — replace the three-mutant demands with two, at `plan.md:73` (E8), `plan.md:118`,
   `acceptance.md:104`, `:107`, `:116`. Verification: `grep -n '세 뮤턴트\|뮤턴트 3종\|3방향'` returns
   nothing outside the withdrawal note itself.

Leaving P1 in place sends M2 to take a measurement the SPEC says cannot be taken; leaving P2 in place
requires the §E report to either fabricate M-leak evidence or under-deliver against its own checklist.
Both are cheap and both are the kind of hole that surfaces as a fabricated verdict later.

**Debt travelling into run phase** — this is what the operator is accepting at the kickoff gate:

1. **REQ-BLE-004 enters run with no working non-vacuity guard.** AC-BLE-003's designated mutant cannot
   be written against the prescribed shared-close design, and by `verification-completeness.md` §2's
   mutant probe the criterion is unadoptable as written, not merely unguarded (P8). M1 must re-lay both
   together with its A/B decision. `spec.md:130` itself calls this the property *"분기가 둘로 갈리는
   변경에서 가장 흔하게 깨지는 성질"* — the guard is missing on exactly the property the SPEC names as
   most fragile.
2. **AC-BLE-003's `/dev/fd` mechanism is not runnable as specified** — unpinned slack, an unlabelled
   platform assumption measured on neither platform, and no empty-sweep clause, so an ubuntu runner that
   cannot read `/dev/fd` returns `ok` and the criterion asserts nothing.
3. **AC-BLE-004 has no mutant guard**, and its sole stated vacuity condition (P3) is under-enumerated.
4. **The RED-now cells for 001b / 002 / 005 are unmeasured and their measurement tree does not exist.**
   M1 constructs it; until then three of six ACs have no starting observation.
5. **`verification-completeness.md` §2.1 is cited while its operative consequence is overridden**
   (P6) — release-blocking status retained at `acceptance.md:129` and `§D.2:134`'s "전부 PASS" against
   §2.1's "not recorded as a pass".
6. **EINTR's reclassification is an accepted, unmeasured behaviour change** — recorded by design at
   `spec.md §1.3.1`, outside REQ-BLE-005's scope. This is debt the SPEC chose deliberately and
   documents well; it travels as a known, not as a defect.

---

## Defects Found (structured defect-list)

**P1 — `plan.md:111` vs `acceptance.md:26` — the retracted `3f03d9c36` RED-now pin survives as a live
M2 instruction; the withdrawal stopped at the `acceptance.md` file boundary — Severity: major —
Class: blocking.**
`acceptance.md:26` retracts the pin and gives the reason (*"그 pin 은 잴 수 없는 측정을 가리키고
있었다"*). `plan.md:111` still reads *"M2 착수 전에 그 표의 RED-now 셀을 `3f03d9c36` 에서 실제로 재고
기록한다"*. `spec.md:27`'s HISTORY records the retraction as scoped to `acceptance.md`, which is where
the incompleteness is visible. This is the N1 pattern relapsing on N3's subject, and it is worse than a
stale label because it is an executable instruction.
**Required fix:** delete `3f03d9c36` from `plan.md:111` and point the sentence at `acceptance.md §D.0`'s
measure-at-measurement-time rule.

**P2 — `plan.md:73` (E8), `plan.md:118`, `acceptance.md:104`, `:107`, `:116` — the three-mutant demand
survives M-leak's withdrawal — Severity: major — Class: blocking.**
The mutant table (`acceptance.md:114`) strikes M-leak and `plan.md:118` says it must not be claimed as
003's guarantee. Yet `§E` E8 demands *"뮤턴트 3종(M-broad·M-narrow·M-leak)의 RED 전문과 되돌린 뒤
GREEN"* — in the same sentence that retracts the RED-now-6 over-demand for exactly this reason —
AC-BLE-005's header reads *"(뮤테이션 3방향)"*, its Given requires *"아래 **세** 뮤턴트"*, and both
`:116` and `plan.md:118` require the three to RED distinct ACs. The §E report must fabricate M-leak
evidence or under-deliver against its own checklist.
**Required fix:** reduce every one of the five sites to two mutants, with a one-line note that the third
is withdrawn pending M1 (`plan.md §B.1` 부채-1).

**P6 — `acceptance.md:32`, `:129`, `:134` — `verification-completeness.md` §2.1 is cited by name and its
operative consequence declined, undeclared (N4 residue, carried from iter-2) — Severity: major —
Class: blocking.**
§2.1 (`verification-completeness.md:201-208`): the criterion *"loses release-blocking eligibility … and
is **not recorded as a pass**. This is the disposition, not one option among several."* `acceptance.md:129`
retains blocking status explicitly; `§D.2:134` requires all six ACs recorded PASS. Separately, §2.1's
trigger is a cited RED that cannot be re-executed (historical / already-merged / external CI); a
green-now criterion is none of those, so the citation is a misapplication independent of the override.
The deviation is conservative and `:47` avoids §2.1's actual harm, which is why this is documentation
debt rather than a firewall failure.
**Required fix:** state the retained blocking status as a declared, stricter-than-doctrine deviation,
naming `:47` as the clause honouring §2.1's purpose; and either drop the §2.1 citation at `:32` or
qualify it as an analogy, since the trigger condition is not met.

**P8 — `plan.md:47` (§B.1 부채-1), `acceptance.md:92` — AC-BLE-003 is recorded as "no guard" where §2's
mutant probe makes it "too shallow to adopt" — Severity: major — Class: blocking.**
`verification-completeness.md:161-168`: a criterion admitting a mutant that satisfies it while violating
its requirement *"is too shallow to adopt"*. Removing the close from the non-contention return path is
such a mutant for AC-BLE-003 — the contention-only induction at `acceptance.md:89` still passes. The
debt record understates its own severity, and stating it correctly is a subtraction, available under this
pass's constraint.
**Required fix:** in `plan.md §B.1` 부채-1, record that AC-BLE-003 currently fails §2's mutant probe and
is therefore unadopted, not merely unguarded, until M1 re-lays it.

**P3 — `acceptance.md:36` — a new affirmative universal in a subtraction-only pass — Severity: minor —
Class: blocking.**
*"그것이 이 AC 가 걸려 있는 유일한 조건이다"*. Unmeasured, and plausibly false: the differential is also
vacuous when the compared runs sweep no test through the three consumers, and unreadable under known
flake noise (which `:101` itself anticipates). This is the one place the pass's subtraction-only
justification is falsified by its own text.
**Required fix:** replace *"유일한 조건"* with *"주된 조건"*, or enumerate the empty-sweep condition
alongside it.

**P4 — `acceptance.md:28` — the measurable-tree description reproduces the configuration `:26` rejects —
Severity: minor — Class: blocking.**
*"RED-now 를 잴 수 있는 트리는 `3f03d9c36` + 새 계약 테스트(분류 수리 이전 상태)이며"*. Under option A
a test naming `classifyBoardFlockErr` does not compile at base, which is the build-failure `:26` had
just condemned. The parenthetical may intend an M1 stub carrying a broad seam; it does not say so.
**Required fix:** say *"`3f03d9c36` + M1 이 놓은 분류 seam 스텁(넓은 매핑 상태) + 새 계약 테스트"*, or
delete the tree description and leave only "M1 constructs it".

**P5 — `spec.md:31` — the self-withdrawal record misstates the iter-2 record and contradicts this
artifact set — Severity: minor — Class: blocking.**
*"실제로 가드가 비어 있는 것은 AC-BLE-004 가 아니라 AC-BLE-003 이었다."* iter-2 sustained its
AC-BLE-004 finding (N5, *"Nothing designates 004"*) and raised 003 separately (N2); both are unguarded,
which `acceptance.md:36` (*"뮤턴트 가드 없음"*) and `:43` (*"지정된 뮤턴트가 없다"*) state in this same
SPEC. The section's disclaimer covers claims about the code, not claims about the audit.
**Required fix:** rewrite as *"감사는 003 과 004 를 각각 다른 이유로 무가드로 지목했다 — 003 은 지정
뮤턴트가 발화하지 못해서, 004 는 지정 뮤턴트가 없어서"*. The *"세 번째 사례"* batch claim is unverified
by me and should be dropped or attributed.

**P7 — `spec.md:165` vs `:163`, `:167` — "004 미피복" reads as REQ-BLE-004 being uncovered — Severity:
minor — Class: blocking.**
In a table whose column is 피복 AC, and mirrored in `acceptance.md:20` under a 피복 REQ column, the
parenthetical *"004 는 M-leak 정의 철회로 현재 미피복"* reads as REQ-BLE-004. `:163` maps REQ-BLE-004 →
AC-BLE-003 and `:167` states *"미피복 REQ 0건"*. The intended meaning is that REQ-BLE-004's
**non-vacuity** is uncovered; nothing in the sentence says so.
**Required fix:** *"REQ-BLE-004 의 **비공허성**은 M-leak 정의 철회로 현재 미피복 — 피복 자체는
AC-BLE-003 이 유지한다"*.

### Carried over, unchanged, all optional

- **N8** (`spec.md:45` labels a `:41-44` fence as `:43`) — **open**, optional. The claim is true at the
  cited line; a one-character improvement.
- **N9** (`spec.md:26` HISTORY 0.2.0 omits the `acceptance.md §D.0` work) — **open**, optional. The
  0.3.0 row is thorough; 0.2.0 still records D1/D4/D5 and not the D3 response, which was that pass's
  largest edit.
- **D6** (consumer count is production-only; six test consumers unnamed) — open, optional.
- **D7** (`"lock board lock %s"` mislabels the integration-mutation path, `plan.md:99`) — open, optional.
- **D8** (`related_specs:` undocumented frontmatter field, `spec.md:16`) — open, harmless.
- **D9** (Tier table omits the LOC axis, `plan.md:17-24`) — open, optional.
- **D10** (EWOULDBLOCK/EAGAIN portability note, `plan.md:104`) — open, optional.

---

## Regression Check — iter-2 defects N1-N9

| # | Status | Evidence |
|---|---|---|
| **N1** — universal claim survives in `plan.md` | **CLOSED** | `plan.md:8` rewritten as a measurement statement with an explicit non-exhaustiveness disclaimer; `:21` scoped to *"측정된 도달 가능 errno … 한정"* + `spec.md §1.3.1` pointer; `:69` carries the measured-reachable qualifier and states the broad reading is withdrawn; `:41` adds EINTR. Residue grep across all four artifacts returns only correctly-scoped uses. |
| **N2** — M-leak cannot fire; AC-BLE-003 unguarded | **DELIBERATELY DEFERRED AS DEBT** — record adequate, one gap | Recorded at `plan.md:47` (§B.1 부채-1) and mirrored at `acceptance.md:35`, `:42`, `:92`, `:118`, `plan.md:118`. The mutant table row is struck. No surface reads as a guard. Gap: the §2 mutant-probe consequence ("too shallow to adopt") is not recorded — P8. |
| **N3** — RED-now cells pinned to an unmeasurable tree | **PARTIALLY CLOSED** | Retracted in `acceptance.md:26-28`, `:33`, `:34`, `:37`, `§D.2:135`; no per-row SHA is pre-written. But `plan.md:111` still orders the measurement at `3f03d9c36` — P1. And `:28`'s replacement tree description is under-specified — P4. |
| **N4** — §2.1 cited then overridden; mutant probe as adoption basis | **PARTIALLY CLOSED** | The mutant-probe half: the pass's refusal is **sustained on the evidence** — my required fix asserted a coverage claim false for AC-BLE-003 (see answer 5). The §2.1 half: **OPEN**, unaddressed — `acceptance.md:32` still cites the disposition, `:129` and `§D.2:134` still override its consequence — P6. |
| **N5** — AC-BLE-004 swept into a blanket non-vacuity claim | **CLOSED** | `acceptance.md:39` retracts the blanket claim by name and re-states it per-AC at `:41-43`; `:36`'s grade cell reads *"회귀-가드 — 뮤턴트 가드 없음"*; `§D.1:129` narrows to *"대체 수단도 셋에 고르게 있지 않다"*. Introduced a new universal in the same cell — P3 — and the coverage-column ambiguity — P7. |
| **N6** — `/dev/fd` threshold, platform label, empty sweep | **DELIBERATELY DEFERRED AS DEBT** — record adequate | `acceptance.md:91`, directly under the mechanism, opens *"지금 상태로는 판정을 이루지 못한다"* and enumerates all three holes including the silent-`ok` empty sweep and the ubuntu-CI mismatch; the platform sentence is relabelled *"어느 쪽에서도 측정되지 않은 추론"*. Mirrored at `plan.md:48`. |
| **N7** — RED-now cell count 3 vs 6 | **CLOSED** for its stated subject | `plan.md:73` now demands *"RED-now 셀 3개(001b·002·005) … + 회귀-가드 3개(001a·003·004)의 등급 보고"* and states the 6-cell demand is withdrawn, matching `§D.2:135-136`. The same line retains the parallel three-mutant over-demand — P2. |
| **N8** — `board_lock_unix.go:43` labels a `:41-44` fence | **OPEN** (optional) | `spec.md:45` unchanged; verified against the tree (`:41` Flock, `:42` close, `:43` sentinel return). |
| **N9** — HISTORY 0.2.0 omits the `§D.0` work | **OPEN** (optional) | `spec.md:26` still records only D1 / D4 / D5. |

**Stagnation check.** No defect appears unchanged across all three iterations. N8 and N9 are minor and
optional across two. The pass engaged every blocking finding it was given — accepting six, deferring two
as declared debt with reasons, and refusing one with an argument that holds. That is not stagnation; it
is a pass whose withdrawals were correct in kind and incomplete in extent.

---

## Gaps — what I did NOT verify

- **I could not diff iter-2 → iter-3.** The SPEC directory is untracked
  (`git status --porcelain` → `?? .moai/specs/SPEC-BOARDLOCK-ERRNO-001/`), so no git baseline of the
  pre-iter-3 text exists in this tree. My subtraction-only judgment (question 1) is therefore made by
  reading the **current** text against the excerpts iter-2 quoted verbatim. For edited regions iter-2
  did not quote, I cannot mechanically exclude an added assertion — only report that I found none by
  reading. This is the single largest gap in this report.
- **I ran no Go test and no build.** No `go test ./internal/kanban/...`, no `go vet`, no
  `GOOS=windows go build`. I did not measure any RED-now cell, did not run any mutant, and did not
  establish the pre-repair baseline AC-BLE-004 depends on.
- **I did not verify that contention tests through the three `IsBoardLockHeld` consumers exist**, which
  is the fact that would settle whether AC-BLE-004's differential can be vacuous for a second reason
  (P3). I read the three call sites, not their test coverage.
- **I did not verify `/dev/fd` behaviour on darwin or linux.** N6's platform claim remains unmeasured by
  the SPEC and by me; I judged its labelling, not its truth.
- **I did not re-run `./bin/moai spec lint`.** I consumed the lead's established measurement
  (`0 error(s), 1096 warning(s)`, rc 0, tree-built binary, no pipe) rather than re-executing, and it is
  attributed to the lead, not to this run.
- **I did not verify `spec.md:31`'s batch-level claim** (*"이 배치에서 … 세 번째 사례"*). It concerns
  cards outside this SPEC; I make no claim either way.
- **I did not re-derive D6-D10** (iter-1 optional carry-overs) beyond confirming their cited lines still
  read as iter-2 described. iter-2's own correction to D6's count (six test consumers, not five) is
  carried forward unverified by me.
- **I did not verify `board_lock_windows.go`'s `os.IsExist` characterisation** beyond reading `:69`
  itself.
- **No cross-model audit was run.** This is a Claude-only verdict; `mcp__moai__audit_multi` /
  `codex_audit` / `glm_audit` were not invoked, so no independent backend corroborates it.
- **I did not check the template mirror** or any surface outside the four SPEC artifacts, the four cited
  `internal/kanban` files, and `verification-completeness.md`.

## Residual risk

- **The score is flat, not recovered, and this is the terminal iteration.** A fourth pass would very
  likely close P1-P8 — they are all sentences — but the contract does not permit one without an explicit
  operator override. The risk the operator accepts at the kickoff gate is that these eight land as
  drive-by fixes during M1 rather than under audit, and that the two majors (P1, P2) are the kind that
  produce a fabricated verdict downstream if they do not.
- **My subtraction-only finding rests on reading, not on a diff.** If iter-3 added an assertion in a
  region iter-2 never quoted, I would not have caught it structurally — only by noticing it read as an
  assertion. Two of the three exceptions I did find (P3, P4) were caught that way, which is weak
  evidence the method works and no evidence that it is complete.
- **AC-BLE-003 and AC-BLE-004 enter run phase unguarded, and the SPEC's central requirement
  (REQ-BLE-004, fd hygiene on a branching change) is the one with no working instrument.** The SPEC says
  so honestly in five places — which is why this is PASS-WITH-DEBT and not FAIL — but honest labelling
  does not close the hole. If M1 does not re-lay the guard, the landing evidence will show a green whose
  meaning has never been established.
- **The `verification-completeness.md` §2.1 override (P6) is conservative in this SPEC**, but it is a
  named [HARD] clause overridden without declaration, and SPECs are read as precedent. The risk is
  propagation, not local harm.
