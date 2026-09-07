# SPEC Review Report: SPEC-CODEX-GHOST-SKILLS-PRUNE-001

Iteration: **2/2** (Tier M ceiling — `.moai/config/sections/harness.yaml:77` `M: 2`)
Verdict: **FAIL**
Overall Score: **0.83** — above the Tier M PASS threshold (0.80, `spec-workflow.md:141`).
**The FAIL does not rest on the score.** It rests on two blocking findings, one of
which is a demonstrated contradiction between two requirements and the other a
worked counter-example in which the newly added REQ-CGP-018 permits the deletion
of a healthy registration. Both are small, local edits to the SPEC text.

Iteration 1 verdict left intact at `.moai/reports/t506/plan-audit-verdict.md`
(FAIL, 0.76). Score movement 0.76 → 0.83 is an improvement, so no STOP-on-regression
signal is raised.

Reasoning context ignored per M1 Context Isolation. The author's repair summary
was treated as a set of claims to check, not as evidence; every row below was
re-read from the artifacts and, where it names source, from the source.

Tree: `.claude/worktrees/t506`, branch `WT-codex-ghost-skills`, HEAD `ace1c5440`.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — 21 ids, `REQ-CGP-001..021`, each declared
  **exactly once**. Evidence: `grep -o '\*\*REQ-CGP-[0-9]*\*\*' spec.md | sort | uniq -c`
  → 21 rows, every count `1`; `grep -o 'REQ-CGP-[0-9][0-9][0-9]' spec.md | sort -u | wc -l`
  → `21`. No gap, no duplicate, uniform padding. The out-of-document-order numbering
  (018 in §C.3, 019 in §C.5, 017 in §C.7) is declared and explained at `spec.md:57`;
  it is not a gap symptom, and I confirmed the set rather than the ordering.

- **[PASS] MP-2 GEARS format compliance** — **the iteration-1 failure is resolved.**
  I read all 21 REQ entries individually, not a sample:
  - `spec.md:69` REQ-CGP-004 — "시스템은 … 항목**만 제거해야 한다**" → Ubiquitous,
    modal, actor named. The "만" (only) is load-bearing and correct: it states three
    *necessary* conditions, which is what lets REQ-CGP-018 add a fourth without
    contradiction.
  - `spec.md:73` REQ-CGP-005 — "시스템은 `enabled` 값을 … 사용**해서는 안 된다**" →
    Unwanted, modal. (Its *second* sentence is the subject of defect E1 below; the
    modality itself is correct.)
  - New entries: `:83` REQ-CGP-018 "제거해서는 안 된다" (Unwanted), `:102`
    REQ-CGP-019 "보존해야 한다" (Ubiquitous), `:106` REQ-CGP-020 "실행해서는 안 된다"
    (Unwanted), `:107` REQ-CGP-021 "유지해야 한다" (Ubiquitous). All four modal.
  - The 15 carried-over entries are unchanged and were already compliant.
  Judgment layer: the `REQ-XXX` requirement layer in `spec.md`. The Given/When/Then
  entries in `acceptance.md` are the verification layer and were graded under
  Testability, not here (M3 § Scope).

  **The lint-vacuity side-finding was carried into three surfaces, and it is
  accurate on all three.** `spec.md:59`, `progress.md:34-40` (marked `[HARD]`), and
  `acceptance.md:186` each state that a green `moai spec lint` says nothing about
  GEARS modality for a Korean SPEC. I re-read `internal/spec/lint.go:790`
  `isModalityMalformed`: it tests `strings.HasPrefix(upper, "WHEN "/"WHILE "/
  "WHERE "/"IF "/"THE ")` and returns false unconditionally otherwise. The three
  restatements are correct, and `acceptance.md:186`'s prohibition ("린트 초록을
  모달리티 근거로 인용하지 않는다") is the right shape — it binds a future run-phase
  reader, which is where the misreading would otherwise land.

- **[PASS] MP-3 YAML frontmatter validity** — frontmatter unchanged from iteration 1
  (`spec.md:1-16`), all 12 canonical fields present and typed, no rejected alias.
  `moai spec lint .../spec.md` → rc 0, `✓ No findings` (binary built from this tree
  in iteration 1: `go build -o /tmp/t506-audit-moai ./cmd/moai`, rc 0). The green is
  cited for the frontmatter and REQ-id axes only, per MP-2 above.

- **[N/A] MP-4 language neutrality** — unchanged: scoped to this repository's own Go
  source, not template-bound or multi-language tooling.

- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `related_specs` unchanged
  (`spec.md:15`); the three named SPECs plus the two cited in `plan.md` were verified
  `status: completed` in iteration 1 and the reference set did not change. No
  BLOCKING finding.

- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md` → `0`.
  Auto-PASS.

- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION'` over the SPEC
  directory → rc 1, no match. `research.md` absent (Tier M).

---

## Regression Check — iteration-1 defects

| # | iter-1 finding | Disposition | Evidence |
|---|---|---|---|
| MP-2 | REQ-CGP-004/005 non-modal | **RESOLVED** | `spec.md:69`, `spec.md:73` — read above |
| D1 | AC-CGP-013 vacuous; `splitLines` line-ending loss undeclared | **RESOLVED** | `acceptance.md:115-129` is now a unit criterion on the reassembly function with three fixtures (newline present / absent / CRLF) and explicitly forbids the end-to-end form, naming the reason (`:127`). `acceptance.md:131-137` AC-CGP-014 forces ≥1 actual removal. `plan.md:84-102` records the `splitLines` loss as a **known fact** with the source quoted, and `spec.md:102` REQ-CGP-019 requires the line-ending state to be carried separately. Residual: E6 below |
| D2 | `plan.md` false extent claim; three span hazards | **PARTIALLY RESOLVED** | The false sentence is corrected at `plan.md:52-62` with the original quoted rather than silently dropped, and the three real obligations (all-lines span incl. the `openDelim` branch, EOF-as-close, trailing-blank exclusion) are stated. The three hazards are addressed by inversion (REQ-CGP-018) — which closes them only under a reading its own text does not require. See **E2**, and **E4** for the untested EOF obligation |
| D3 | AC-CGP-012 orphan | **RESOLVED** | `spec.md:106` REQ-CGP-020 added; `acceptance.md:171` maps it. Two-way completeness verified independently below |
| D4 | mutant RED not bound to the right test | **RESOLVED** | `acceptance.md:42-56` pins all three: diff site = the pruner's own eligibility branch and explicitly **not** `doctor_codex.go:442-450` (`:50`), RED discriminant = the named test rather than the suite (`:51`), both observations recorded (`:52`) |
| D5 | `skip` offered as an option for the must-pass fixture | **RESOLVED** | The skip row is gone from the `plan.md:135` risk table, replaced by "**stat 씨앗 주입이 유일한 경로다** … 플랫폼 조건부 skip 은 선택지가 아니다"; `plan.md:144` declares skip an AC-CGP-004 FAIL; `acceptance.md:54` says the same on the AC itself. `plan.md:138-142` §F.1 states the seam is **new work** and cites `doctor_codex.go:426` correctly (I re-read it: `_, serr := os.Stat(statPath)`, no seam). M2 promoted ahead of the eligibility judge (`plan.md:109`, rationale at `:115`) |
| D6 (optional) | `path = ""` uncovered | **TAKEN** | `acceptance.md:32` row 5b, with the mechanism named (`skillPathKeyRe`, `skills.go:66`) |
| D7 (optional) | `clean.go:34` help text | **TAKEN** | `spec.md:107` REQ-CGP-021, `acceptance.md:139-143` AC-CGP-015, `plan.md:113`/`:119` M6 |
| D8 (optional) | sha256 boundary undefined | **TAKEN** | `acceptance.md:184` fixes it at "run 단계의 첫 명령 직전 / 마지막 명령 직후" |

No iteration-1 defect is unresolved-and-unchanged, so no stagnation flag.

---

## Verification of the three questions asked

### Q1 — Does REQ-CGP-018 contradict, or make unreachable, any existing requirement?

**It makes nothing unreachable**, and against REQ-CGP-004 it is compatible by
construction: `spec.md:69` states three *necessary* conditions ("만 제거해야 한다"),
so a fourth necessary condition composes cleanly. REQ-CGP-006..011 are all
"제거해서는 안 된다" and compose with another prohibition trivially. A plain entry
(header / `path` / `enabled`) remains eligible, so REQ-CGP-004 and AC-CGP-001 are
still reachable.

**It does contradict one — see E1.** `spec.md:73` REQ-CGP-005's second sentence is
an *unconditional obligation to remove*. That is new in this iteration: the modal
rewrite that fixed MP-2 converted a definitional statement into a `shall remove`.

### Q2 — Does the AC-CGP-013 / AC-CGP-014 split leave a path where neither fires?

I enumerated the axes each covers. AC-CGP-013 (`acceptance.md:115-125`) exercises
parse→reassemble **with no deletion**, over three line-ending states. AC-CGP-014
(`:131-137`) exercises **deletion** end-to-end over two line-ending states. The
composition *deletion on a CRLF file* is covered by neither — see **E6** (minor;
AC-CGP-013c constrains the join rule enough that the risk is thin, but the
composition is untested and the fix is one fixture).

A second uncovered path is more consequential: **no AC pins an entry terminated by
EOF**, although `plan.md:61` makes EOF-as-close an explicit obligation — see **E4**.
`acceptance.md:137`'s "마지막 개행을 갖지 않는 경우" variant constrains the file's
last byte, not the removed entry's position, so an implementation that mishandles
an EOF-terminated span passes both.

### Q3 — Did the AC-CGP-007 narrowing drop coverage?

**No.** I diffed the three axes the old AC-CGP-007 carried against where each now
lives:

| Axis | iter-1 home | iter-2 home |
|---|---|---|
| A literal **outside** an entry must not manufacture a phantom entry | AC-CGP-007 (fixture clause) | AC-CGP-007, now its whole subject (`acceptance.md:70-76`) |
| Non-removed bytes identical after a real removal | AC-CGP-007 (Then clause) | AC-CGP-014 (`:131-137`), whose fixture explicitly carries 다른 테이블·주석·여러 줄 리터럴 |
| A literal **inside** an entry | uncovered in iter-1 | AC-CGP-003 row 7-a (`:34`) |

`acceptance.md:78` states this partition itself, and `§D.2:166` maps REQ-CGP-015 to
**both** AC-CGP-007 and AC-CGP-014, so the byte-preservation requirement did not
lose its second leg. Nothing was dropped; the narrowing sharpened three axes that
were previously fused into one fixture.

### Counts and two-way map — verified independently, not accepted

- REQ: `grep -o '\*\*REQ-CGP-[0-9]*\*\*' spec.md | sort | uniq -c` → 21 declarations,
  each exactly once. Claim "17 → 21" holds.
- AC: `grep -c '^### AC-CGP-' acceptance.md` → `15`; unique ids `AC-CGP-001..015`,
  contiguous. Claim "13 → 15" holds.
- **Left column complete**: the `§D.2` table (`acceptance.md:157-172`) names
  001,002,003,004,005,006..011,012,013,014,015,016,017,018,019,020,021 — all 21.
- **Right column complete**: collecting every AC id appearing on the right yields
  {001,002,003,004,005,006,007,008,009,010,011,012,013,014,015} — all 15, **no
  orphan**. The `§D.2:174` claim of two-way completeness is true. D3 did not
  regenerate.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.70 | 0.50-0.75 | The E1 contradiction is not an ambiguity a reasonable engineer resolves consistently — two requirements give opposite instructions for the same entry, and which one an implementer follows is unpredictable. Everything else reads unambiguously; the numbering note (`spec.md:57`) and the AC-CGP-007/014/003 partition note (`acceptance.md:78`) actively remove ambiguity that iteration 1 left. |
| Completeness | 0.85 | 0.75-1.0 | All sections present; five `### Out of Scope — …` H3 sub-headings with specific bullets, unchanged. The command surface (REQ-020/021), the line-ending state (REQ-019), and the span-content disposition (REQ-018) all now have requirements where iteration 1 had none. Deducted for E4 (a named plan obligation with no AC). |
| Testability | 0.75 | 0.75-1.0 | AC-CGP-004 is now genuinely non-vacuous (three pinned clauses + skip-is-FAIL); AC-013/014 split is sound and each names why the other cannot substitute. Deducted for E3 (the binding Given/Then of the feeder AC still says 여섯) and because the fixture set cannot detect E2's case. |
| Traceability | 1.00 | 1.0 | Verified two-way by enumeration above, not by reading the claim: 21/21 REQs on the left, 15/15 ACs on the right, zero orphans. Row `:162` even names which AC-003 table row serves which REQ. |

Aggregate (mean) = **0.8375 → 0.83**.

---

## Defects Found

**E1. REQ-CGP-005's second sentence is an unconditional removal obligation that
contradicts REQ-CGP-018 (and every other never-prune clause)** — `spec.md:73` vs
`spec.md:83` — Severity: **critical** — Class: **blocking**.
The MP-2 repair rewrote REQ-CGP-005 as two sentences. The first is correct Unwanted
form. The second is:

> `enabled = true` 이면서 경로가 부재한 항목도 유령이며, **시스템은 그 항목을 제거해야 한다**.

That is a `shall remove`, with no qualifier. Take an entry that is `enabled = true`,
declares an absolute path that does not exist, **and** carries an unrecognized line
inside its span. REQ-CGP-005 says the system shall remove it; REQ-CGP-018 says the
system shall not. Both are Ubiquitous/Unwanted with no precedence rule, and
`spec.md:89`'s framing ("§C.3 의 일곱 조항이 이 SPEC 의 안전 경계다") is prose, not a
conflict-resolution clause. The same collision arises against REQ-CGP-011 if a
future reading of "경로가 부재한" and "해석에 성공한" ever overlap, though that pair
is disjoint today.
This is a **repair-planted** defect: iteration 1's REQ-CGP-005 was definitional
("게이트가 아니다") and carried no removal obligation, so the collision did not exist
before the modal rewrite.
Required fix (one clause): keep the first sentence as the requirement and demote the
second to a rationale, or qualify it — e.g. "…유령이며, 다른 never-prune 조항에
걸리지 않는 한 시스템은 그 항목을 제거해야 한다". Note the demotion loses nothing:
the eligibility obligation already lives in REQ-CGP-004, whose condition set does
not mention `enabled`, and AC-CGP-005 tests it.

**E2. REQ-CGP-018's recognition test is textual, and a line swallowed by a
multi-line literal can be textually "recognized" — the rule then permits deleting a
healthy registration** — `spec.md:83`, `plan.md:70-74` — Severity: **critical** —
Class: **blocking**.
The inversion is the right move and it does close all three named hazards **under a
parser-state reading** — that is, if "인식되지 않는 줄" means *a line the parser did
not consume as one of the five kinds*. But neither `spec.md:83` nor `plan.md:68`
says that: both enumerate five **line shapes** (`[[skills.config]]` 헤더 / `path` 대입
/ `enabled` 대입 / 빈 줄 / 온전한 주석 줄), which is a test a reader will naturally
implement as a re-scan of the span's text. `plan.md:72` states the false premise
explicitly: "여러 줄 리터럴의 어느 줄도 다섯 종류에 들지 않으므로". A literal's lines
can be exactly those shapes.

Worked counter-example, walked against `internal/codexwiring/skills.go:106-143`:

```
[[skills.config]]        L0
path = "/gone"           L1
# uses """ in prose      L2
[[skills.config]]        L3
path = "/exists"         L4
# and """ again          L5
[other]                  L6
```

`multilineOpener` (`skills.go:83-90`) counts `"""` on L2 → 1, odd → `openDelim` set,
so the parser's `openDelim` branch (`:113-121`) **skips L3, L4** and closes on L5.
`[other]` at L6 toggles `inEntry` false. The parser therefore reports **one** entry,
`Path == "/gone"`, missing → eligible, with a span of L0..L5. Re-scanning that span
textually finds: header, `path`, comment, header, `path`, comment — every line one of
the five recognized shapes → **not disqualified** → the pruner deletes L0-L5,
destroying the healthy `/exists` registration. That is the exact harm `spec.md:89`
declares the safety boundary exists to prevent, and no AC in the current set builds
this fixture.
A narrower variant with no second header (the swallowed region carrying only a
second `path =` line and comments) defeats a "at most one header per span" patch,
so that patch is not sufficient.
Required fix: make the recognition **parser-state-based** and say so in
REQ-CGP-018 — "범위 안의 어떤 줄이라도 파서의 `openDelim` 분기(`skills.go:113-121`)에
의해 소비되었다면, 그 줄의 모양과 무관하게 인식되지 않은 줄로 본다" — and correct
`plan.md:72`'s premise. Then add the fixture above as AC-CGP-003 row 7-d, asserting
both entries survive.

**E3. AC-CGP-003's binding Given/Then still says "여섯 항목" while its own table
carries eight rows** — `acceptance.md:21`, `:23` vs `:25-34`, `:36` — Severity:
**major** — Class: **blocking**.
The Given reads "아래 **여섯** 항목을 각각 하나 이상 담은 픽스처가" and the Then
"**여섯** 항목이 모두 파일에 남는다", but the table now has rows 1,2,3,4,5a,5b,6,7 —
eight rows over seven classes — and the prose two lines below says "**여덟** 행(5a/5b
포함)". A tester implementing the Given/Then literally builds a six-class fixture and
never constructs row 5b (`path = ""`) or row 7 (a/b/c), which is the entire D2
repair. This AC is also the feeder for the must-pass AC-CGP-004, so a short fixture
here weakens the one criterion the SPEC declares cannot fail.
Repair-planted: the rows were added without updating the sentence that counts them.
Required fix: change both occurrences to match the table (여덟 행 / 일곱 부류), and
state which count the sentence means — rows or classes.

**E4. `plan.md:61` makes EOF-as-close an explicit obligation; no AC verifies it** —
`plan.md:61` vs `acceptance.md` (all 15) — Severity: **major** — Class: **blocking**.
§B.1 item 2 states "**EOF 는 닫힘이다** … 범위 계산이 EOF 를 명시적으로 닫힘으로
다루어야 그렇다". I read all 15 ACs for a fixture whose eligible entry is the last
block in the file: AC-CGP-001's fixture is unspecified as to position; AC-CGP-014's
two variants pin the file's final byte, not the entry's position. An implementation
that computes the span only on a close *event* — and therefore drops or mis-bounds
an entry the loop exits on — passes every current AC. This is the same class as the
span hazards: an off-by-one at the end of file deletes or preserves the wrong lines.
Cost of the fix is one fixture.
Required fix: add to AC-CGP-014 (or AC-CGP-001) a variant in which the eligible entry
is the **last** block, tested in both trailing-newline states.

**E5. `plan.md:133`'s mitigation cites a milestone that no longer holds that
fixture** — `plan.md:133` vs `plan.md:108-113` — Severity: **minor** — Class:
**optional**.
The §F risk table's first row still mitigates with "M1 의 무손실 왕복 + **M2 의**
'제거 후 나머지 바이트 동일' 픽스처". After the renumbering that promoted the stat
seam, M2 is the stat seam (`:109`) and the byte-identity fixture is AC-CGP-014,
landing under M3/M5. A reader following the mitigation to M2 finds nothing.

**E6. AC-CGP-014 has no CRLF variant — removal on a CRLF file is untested** —
`acceptance.md:137` — Severity: **minor** — Class: **optional**.
AC-CGP-013 covers CRLF at the unit level and AC-CGP-014 covers removal at `\n`. The
composition is uncovered. The risk is thin — AC-CGP-013c forces the join rule to be
self-consistent, since `splitLines` leaves each line's trailing `\r` in place and
only the final terminator is at stake — but it costs one variant to close.

---

## What I checked and found nothing wrong with

- **Every claim the repair summary made was verified against the artifact, and all
  nine held.** The two I most expected to be overstated were not: `plan.md §B.1`
  really does preserve the false sentence as a quotation before correcting it
  (`:54`), and the skip option really is *removed* from the `§F` risk table rather
  than merely annotated (`:135` now names seam injection as the only route).
- **The two-way `§D.2` map.** Enumerated both columns by hand rather than reading
  `:174`'s assertion. Complete in both directions; the iteration-1 orphan is gone and
  no new one appeared — which was the specific regression risk named in the brief.
- **REQ-CGP-004's "만" (only).** This single word is what makes the whole inversion
  legal. Had the repair written "…세 조건을 만족하는 항목을 제거해야 한다" (sufficient
  rather than necessary), REQ-CGP-018 would contradict it exactly as REQ-CGP-005's
  second sentence now does. It is correct as written.
- **The lint-vacuity carry-through.** Re-read `internal/spec/lint.go:790` and
  compared it against all three restatements (`spec.md:59`, `progress.md:36`,
  `acceptance.md:186`). The mechanism is described accurately in each, with no
  overclaim about what the green *does* cover (both `spec.md:59` and `progress.md:40`
  correctly scope it to frontmatter and REQ ids).
- **Hazard (c) closure.** Checked independently of the author's reasoning: an array
  continuation line such as `["x", "y"]` can only exist beneath an opening key
  assignment (`args = [`), and that opener is inside the span and is not one of the
  five shapes — so the entry is disqualified before `anyTableRe`'s over-match
  (`configtoml.go:76`) can matter. `plan.md:74` reaches the same conclusion by the
  same route. Closed under **either** reading of REQ-CGP-018, unlike hazards (a) and
  (b).
- **Scope discipline.** Re-checked after four new REQs landed: none of REQ-018/019/
  020/021 reaches axis 2. `grep 'codex_measured_version\|agents-codex' spec.md` still
  returns only the exclusion line at `:124`. REQ-CGP-021 widens the card to a help-text
  edit in `clean.go`, which is a consequence of this card's own change and is declared
  in-scope at `:107` — not a leak.
- **Non-goal integrity.** `spec.md:117-120` unchanged. I re-read all 15 ACs (13 carried
  + 2 new) for one that would execute against the live config: AC-CGP-014 is fixture-
  based, AC-CGP-015 reads `--help` output only. `plan.md:127` still confines every
  test to `t.TempDir()` and the seams. `acceptance.md:182-184` strengthens the proof
  obligation with a defined boundary.
- **Structural gates.** REQ set 21 contiguous; AC set 15 contiguous; `moai spec lint`
  rc 0; no `NEEDS CLARIFICATION`; `syscall` count 0.

---

## Baseline-attribution

All figures produced in this run against this tree (`ace1c5440`,
`.claude/worktrees/t506`), reading the four artifacts as they stand now. Source line
numbers read from the working tree at that HEAD. The linter is `/tmp/t506-audit-moai`,
built from this tree in iteration 1 (`go build -o /tmp/t506-audit-moai ./cmd/moai`,
rc 0) — the tree has not moved since (`git rev-parse --short HEAD` → `ace1c5440`,
artifacts untracked), so the tool-provenance coordinate still holds. The population
figures in `spec.md §A.1` are cited from `.moai/reports/t506/baseline-measurement.md`,
not re-measured.

## Gaps — what I did NOT observe

- I did not execute E2's counter-example. It is a hand-walk of
  `skills.go:83-90` and `:106-143` against a fixture I wrote in this report — a
  reading of the code, stated as such. It is falsifiable by running
  `ParseSkillEntries` on those seven lines, which I did not do because no pruner
  exists to run it through.
- No Go tests run; no implementation exists.
- I did not search `internal/cli` for a pre-existing stat seam under another name.
  `plan.md:140` correctly records that this check is M2's first task and that plan
  phase did not do it — the gap is declared, not hidden.
- I did not re-verify the five referenced SPECs' `status:` in this iteration; the
  reference set is unchanged from iteration 1, where I read all five.
- I did not re-measure this machine's `~/.codex/config.toml`.

## Residual risk

- **E2 is the shape this SPEC keeps producing: a rule that is correct in intent and
  defeated by the parser's actual behaviour.** Iteration 1 found it in the span
  computation; the repair moved it into the span's *content test*. The common cause
  is that both rules are written about **text** while the thing that decides is
  **parser state**. Fixing E2 by naming parser state is the durable form; another
  textual patch will move the hole again.
- **The safety boundary still cannot be observed on this machine** (four never-classes
  at 0, `spec.md:36-44`), so everything rests on fixtures that do not exist. AC-CGP-004
  is now tightly specified, which removes the vacuity route — but E3's stale count
  can still produce a short fixture set, and a short set is how a tight AC ends up
  proving less than it says.
- **Even with E1-E4 fixed, byte-preservation over arbitrary hand-edited TOML remains
  the design's outer risk.** REQ-CGP-018's disqualification is the right posture and
  materially reduces it — a disqualified entry is a survived entry — but the residue
  is not zero while recognition is textual.

---

## Recommendation

FAIL at iteration 2. This is the **Tier M ceiling** (`harness.yaml:77` `M: 2`), so
the retry loop cannot continue on the auditor's own authority — this goes to the
operator with the standard three options. My recommendation on that choice, stated
plainly:

**The repair was good.** Six of six iteration-1 findings were addressed, none was
disputed, every claim the author made about the repair is true against the artifacts,
and the score moved 0.76 → 0.83, above the Tier M threshold. Two of the three
questions the brief asked came back clean (the AC-CGP-007 narrowing dropped nothing;
the map is genuinely two-way complete). What blocks is small and local:

1. **E1** — qualify or demote REQ-CGP-005's second sentence (`spec.md:73`). One clause.
2. **E2** — make REQ-CGP-018's recognition parser-state-based (`spec.md:83`), correct
   the false premise at `plan.md:72`, add the swallow fixture as AC-CGP-003 row 7-d.
   Two sentences and one table row.
3. **E3** — reconcile 여섯 / 여덟 in AC-CGP-003's Given/Then (`acceptance.md:21,23`).
4. **E4** — one fixture variant pinning an EOF-terminated eligible entry.

E5 and E6 are optional; surface them, do not gate.

Given that, **PASS-with-debt is the wrong disposition for E1 and E2** — one is a
contradiction that makes the implementation's behaviour unpredictable and the other
has a worked path to destroying a healthy user registration, which is this card's
cardinal harm. The disposition I would recommend to the operator is a **bounded
iteration 3 scoped to exactly these four edits**, confirmed by a delta re-audit rather
than a full re-audit. E3 and E4 alone would be acceptable as documented debt; E1 and
E2 would not.
