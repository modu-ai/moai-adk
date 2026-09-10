# SPEC Review Report: SPEC-BOARDLOCK-ERRNO-001 — iteration 2

Card: t379
Iteration: 2/3 (Tier S ceiling is 1; this iteration was explicitly commissioned by the lead to judge the repair pass)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t379`, branch `WT-boardlock-errno`, HEAD `3f03d9c36`
(`git rev-parse --short HEAD` → `3f03d9c36`; `git branch --show-current` → `WT-boardlock-errno`;
`git status --short` → two untracked entries only: `.moai/reports/t379/`, `.moai/specs/SPEC-BOARDLOCK-ERRNO-001/`)

Verdict: **PASS-WITH-DEBT** — **with a STOP signal** (score regression, see § Score monotonicity)
Overall Score: **0.83** (harmonic mean; Tier S PASS threshold 0.75)
Delta vs iter-1: **−0.02 (0.85 → 0.83) — DOWN**

Reasoning context ignored per M1 Context Isolation. The repair pass's own account of what it did was
read as a set of claims to be re-verified against the files, not as findings to adopt.

Scope: per the Retry Loop Contract, this is a delta-scoped re-audit over the enumerated iter-1 defects
D1-D5 plus the edited regions the lead named, not a from-scratch full re-audit. Must-pass criteria were
re-run because their inputs (frontmatter, REQ set, cross-refs) sit inside the edited regions.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-BLE-001`…`005` at `spec.md:109,113,119,123,127`.
  Sequential, no gaps, no duplicates, uniform padding. The narrowing edit to 005 changed its body, not
  its number.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`), never against the ACs. 001/002 event-driven (`When …, the <subject> shall …`);
  003/005 ubiquitous (`The <subject> shall …`); 004 state-driven (`While …, … shall …`). The rewritten
  REQ-BLE-005 (`spec.md:127`) remains a clean ubiquitous form: *"The observable behaviour of
  `AcquireBoardLock` and of every `IsBoardLockHeld` consumer shall be unchanged on every
  **measured-reachable** input…"*. The six Given-When-Then entries in `acceptance.md` are
  verification-layer artifacts, graded under Group 4 per M3 § Scope.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical names
  (`spec.md:2-13`); `version` bumped to `"0.2.0"` (quoted semver), `updated: 2026-08-31`, `status: draft`.
  No rejected snake_case alias. `plan.md` / `acceptance.md` carry no frontmatter block at all and
  therefore no `status:` field — compliant with § Artifact Statelessness.
- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC (Go, `internal/kanban`). Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — re-run this iteration:
  `grep -n '^status:' .moai/specs/SPEC-STRESS-INVARIANT-VERDICT-001/spec.md .moai/specs/SPEC-KANBAN-BOARD-001/spec.md`
  → `implemented` and `completed` respectively. Neither is `retired`/`superseded`/`archived`; both exist
  on disk. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears in `spec.md §1.3-6` (prose "생 syscall
  래퍼" + the filename `zsyscall_darwin_arm64.go`) and `plan.md:76`… the build-tag obligation is present
  and load-bearing in four places: `plan.md:40` (`//go:build !windows`, `GOOS=windows` build mandatory),
  `plan.md:103` (new test file carries the tag), `spec.md:195`, `acceptance.md:93`. No BLOCKING finding.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-BOARDLOCK-ERRNO-001/`
  → no output, rc=1. Tier S carries no `research.md`; `plan.md` scanned and clean.

---

## Category Scores

| Dimension | Score | Δ vs iter-1 | Rubric band | Evidence |
|---|---|---|---|---|
| Clarity | 0.85 | 0.00 | 0.75-1.0 | The `spec.md §1.3-4` ↔ `§1.3-5` contradiction is genuinely gone (`spec.md:75` is now a measurement statement, explicitly disclaiming exhaustiveness). `§1.3.1` (`spec.md:79-91`) is the strongest new prose in the document. Offsetting: the retracted universal claim **survives verbatim in `plan.md:8`, `:21`, `:62`** — the same defect relocated one layer up (N1). |
| Completeness | 0.92 | +0.04 | 0.75-1.0 | EINTR now in the unmeasured list in three places (`spec.md:77`, `spec.md:176`, `acceptance.md:138`) with its retry→hard-error consequence stated; `§1.3.1` records the decision as an accepted change; a sixth Out-of-Scope H3 added (`spec.md:185`). 6 `### Out of Scope — <topic>` H3s, each with `-` bullets. Deductions: `spec.md:26` HISTORY omits the `§D.0` work; `plan.md:41` B-미측정 omits EINTR. |
| Testability | 0.68 | −0.04 | 0.50-0.75 | Real machinery added (`§D.0` matrix, M-leak, the `/dev/fd` mechanism, the deleted escape hatch). But iter-2 now makes **affirmative claims that do not hold**: M-leak mutates a branch the prescribed design does not have and its designated RED is unreachable by the specified induction (N2); the RED-now cells are pinned to a tree on which they cannot be measured (N3); `/dev/fd`'s pass threshold is an unpinned constant with no empty-sweep clause (N6). A false affirmative in a verification artifact is worse than the omission it replaced. |
| Traceability | 0.90 | −0.10 | 0.75-1.0 | `spec.md §3` correctly adds REQ-BLE-004 to AC-BLE-005's joint coverage (`spec.md:160`), matching `acceptance.md:20`. 0 uncovered REQ, 0 orphan AC, AC bodies still in exactly one file. Deduction: `acceptance.md:37` asserts a coverage relation — that AC-BLE-005's three mutants supply non-vacuity for 001a·003·**004** — that the mutant table (`acceptance.md:101-105`) does not carry: no mutant designates 004 (N5). |

Harmonic mean: `4 / (1/0.85 + 1/0.92 + 1/0.68 + 1/0.90) = 4 / 4.845 = 0.826` → **0.83**.

---

## Answers to the lead's six questions

### 1. D1 — is the contradiction gone, or merely relocated?

**The contradiction is genuinely resolved in `spec.md`, and the narrowing is NOT "true and useless".
But the same defect now exists one layer up, in `plan.md`, untouched.**

On the narrowing itself, the repair is right and the "true but useless" worry does not land. Three
reasons, in order of weight:

1. **REQ-BLE-005 retains a falsifying input.** A narrowed implementation that classifies
   `EWOULDBLOCK`/`EAGAIN` as non-contention breaks it — and that is not a hypothetical, it is exactly
   the failure M-narrow simulates and AC-BLE-001a catches. A requirement with a real, named, plausible
   violator is not scoped to inputs that cannot break it.
2. **The discarded scope was re-homed, not dropped.** What the old wording implicitly promised — that
   the change is invisible in production — is now carried explicitly by `§1.3.1` as an *accepted
   behaviour change* with a stated decision, three reasons, and an explicit "this is unmeasured"
   label. Relocation-without-disclosure would be the defect; relocation-with-disclosure is how a
   scope reduction is supposed to be recorded.
3. **The reduction propagated consistently across `spec.md` and `acceptance.md`.** Verified at every
   surface that carried the old scope: `spec.md:127` (REQ body), `:131` (its note), `:145` (§2.1's new
   third sentence), `:176` and `:185-186` (§4), `:196` (§5 constraint), `acceptance.md:19` (§D matrix),
   `:85` (AC-BLE-004 title), `:138` (§D.3). That is the cross-layer revision sweep
   `verification-completeness.md` §3 asks for — done properly, within those two files.

The sweep stopped at the file boundary. `plan.md` still asserts the retracted claim:

- `plan.md:8` — *"도달 가능한 errno 가 `EWOULDBLOCK`/`EAGAIN` 하나뿐이고"* — the exhaustive reachability
  claim `§1.3-4` was rewritten to abandon. This is the plan's **founding premise sentence**.
- `plan.md:21` — Tier table: *"오늘의 행동 변화 | 0 (도달 불가 errno 만 재분류)"* — "only unreachable
  errnos are reclassified", now directly contradicted by `§1.3.1`, which reclassifies EINTR and
  establishes its unreachability only as *implausible*, not impossible.
- `plan.md:62` — *"REQ-BLE-005(오늘 행동 불변)는 협상 대상이 아니다"* — restates the requirement at its
  **pre-narrowing** scope. `spec.md:196` was updated to say "measured-reachable"; this line was not.

`plan.md` §D is what the implementer reads as the constraint list. A run-phase agent working from
`plan.md:62` would hold REQ-BLE-005 at the wider scope `spec.md` has explicitly retracted. This is
`verification-completeness.md` §3's named shape — *"the layer a rule constrains is its blind spot"* —
and it is why D1 is **partially closed**, not closed.

### 2. D3 — judging the disagreement

**On the direction, the repair pass is right and my iter-1 requirement was wrong. On the execution, it
is not adequate — and the AC it leaves unguarded is 003, not 004.**

**Where the repair is right.** AC-BLE-001a and AC-BLE-003 assert properties the pre-repair tree already
holds; they are regression guards by construction, green-before and green-after. Requiring a RED-now
cell for them would demand either a fabricated observation or the "impossible" direction §2 names. My
iter-1 remedy — record the runner's empty-sweep token (`no tests to run` / `[no test files]`) as the
RED — is worse than it looked: §2.1's own evidence row condemns precisely this, *"nine release-blocking
criteria… resting on a single premise — that an absent test turns its suite red — where a runner given
a selector matching zero tests exits 0 and prints `ok`."* An absent test is not a red criterion. The
doctrine's alternative adoption test is §2's **mutant probe**, and 001a passes it decisively (M-narrow
satisfies the requirement's negation and 001a catches it). The repair's substitution is therefore
**adequate in kind**, and `acceptance.md:39`'s refusal to pre-fill the cells — *"계획 단계에서 채워 넣은
척하지 않는다"* — is exactly right.

**Where the execution fails.** Four counts, all citable:

- **The wrong disposition is cited, and its operative consequence is overridden silently.**
  `acceptance.md:30` grades AC-BLE-001a as 회귀-가드 "(§2.1 undecidable disposition 적용)". §2.1's
  disposition is scoped to *"a cited RED [that] cannot be re-executed on the current tree — a
  historical event, an already-merged state, or an externally observed CI result"*. None of the three
  applies: these criteria are green-now, not red-but-unreproducible. Worse, §2.1's stated effect is
  that the criterion *"loses release-blocking eligibility… and is not recorded as a pass. This is the
  disposition, not one option among several: a criterion whose starting observation cannot be
  reproduced does not become a release gate."* `acceptance.md:120` keeps them as release gates:
  *"착지 차단 성질은 유지된다."* The deviation runs in the conservative direction and §2's actual harm
  (recording an unobserved RED as a pass) is avoided by line 39 — but a [HARD] clause is invoked by
  name and then declined, undeclared. (N4)
- **The RED-now pins name a tree that cannot carry the measurement.** (N3, below.)
- **AC-BLE-004's substitution source does not exist.** (N5, below.)
- **AC-BLE-003's substitution source is not writable.** (N2, below — the headline finding.)

**On the AC-BLE-004 row specifically, which the lead asked about.** The row (`acceptance.md:34`) claims
no substitution — "회귀-가드 (정의상)" — while line 37 sweeps it into a blanket claim that all three
regression-guards draw non-vacuity from AC-BLE-005's three mutants. The mutant table maps
M-broad→001b, M-narrow→001a, M-leak→003. **Nothing targets 004.** So line 37 over-claims.

Is 004 therefore an unguarded criterion? Probably not, but the SPEC does not say why. AC-BLE-004 is a
differential (pre-tree failure set == post-tree failure set), so running it against a mutated tree is
itself a comparison — and M-narrow, which would break `IsBoardLockHeld` for real contention, plausibly
flips the existing contention tests (`board_lock_wait_test.go:189`, `backlog_store_test.go:198`,
`backlog_concurrency_test.go:59` / `:275`) and thus REDs 004 incidentally. **I did not run this and do
not carry it as verified** — it is a candidate the SPEC should either name (and then reconcile with
`acceptance.md:107`'s "each mutant must RED a *different* AC" clause) or explicitly reject with a
reason. As written, the row is an **undeclared gap dressed as a definition**.

The genuinely unguarded criterion right now is **AC-BLE-003**, whose only declared guard cannot be
planted — see N2.

### 3. D4 — is the escape hatch closed, and does M-leak fire?

**The escape hatch is closed. The replacement mechanism is under-specified. And M-leak, as written,
cannot fire.**

Closed, verified: `acceptance.md:80` records the deletion in-line — *"종전의 '대안 판정(동등 인정)'은
삭제됐다. '코드 경로를 지나감을 단언'은 어떤 입력이 그것을 RED 로 만드는지 말하지 않으므로 검사 명세가
아니다"* — and no "동등 인정" alternative remains anywhere in the file. That is a clean fix and it cites
the right reason (`verification-completeness.md` §1.2(b)).

**M-leak does not compose with the design the plan recommends.** Two independent reasons, both read
directly off the SPEC's own text:

1. **The prescribed design has no per-branch `unix.Close`.** `plan.md:92` specifies the error shape as:

   ```
   경합      → ErrBoardLockHeld            (변경 없음)
   비경합    → fmt.Errorf("lock board lock %s: %w", lockPath, err)
   양쪽      → 반환 전에 unix.Close(fd)     (기존 성질 보존)
   ```

   `양쪽` — *both* — one shared close before the classification split, preserving the current shape
   (`board_lock_unix.go:42`, verified in this tree: a single `_ = unix.Close(fd)` at `:42` inside the
   one `Flock`-failure block). M-leak is defined as *"비경합 분기에서 `unix.Close(fd)` 제거"*
   (`acceptance.md:105`, `plan.md:110`) — **removing a close from a branch `plan.md:92` does not
   create.** The mutant is not writable against the design it is meant to guard.

2. **The induction cannot reach the non-contention branch anyway.** `acceptance.md:81` induces failures
   by *"같은 root 에 대해 반복 획득 실패를 N회(N=200) 유도"* — repeated acquisition failure on an
   already-held root. At the real call site that is **contention only**; the SPEC's own `§1.3` is the
   finding that no non-contention errno is inducible there. Under design option A (the recommended
   one — `plan.md:76`), `classifyBoardFlockErr` is a pure function that never owns the descriptor, so
   the non-contention *return path* is unreachable from a test without the option-B seam the plan
   argues against. So even a per-branch M-leak would sit on a branch the AC never executes.

   This is a design-coupled finding: it interacts with M1's one open decision. Choosing option B would
   dissolve reason 2 but not reason 1.

**The `/dev/fd` mechanism is named but not runnable as specified.** Three gaps:

- **The pass threshold is an unpinned constant.** `acceptance.md:81` requires
  `post ≤ pre + 고정 여유(테스트 자신이 여는 파일 수)` — the slack is never given a value. The implementer
  must invent it, which turns a binary judgment into a calibration.
- **The platform claim is stated as fact and was measured on neither platform.** *"darwin·리눅스 모두 이
  경로가 프로세스 자신의 열린 디스크립터를 노출한다"* is asserted flat. The rest of this SPEC is
  scrupulous about the inference/measurement boundary — `§1.3-6` labels its own reachability argument
  *"syscall 계약을 읽은 추론이지 측정이 아니다"* — and this sentence does not get the same treatment.
  On Linux `/dev/fd` is a symlink into `/proc/self/fd` and depends on `/proc` being mounted; the CI
  runner is ubuntu and the plan-phase measurement was darwin. I did not verify it either, and I make no
  claim that it is false — only that it is unlabelled.
- **A `/dev/fd`-unavailable runner produces a silent green.** A test that skips when the path is
  unreadable makes `go test` print `ok`, and AC-BLE-003 carries **no** selector-match-count or
  empty-sweep clause — unlike AC-BLE-001a, which has one at `acceptance.md:54` (*"0 매치는 초록이 아니라
  미실행이다"*). This is `verification-completeness.md` §1.1's empty-sweep hazard sitting in the one AC
  whose mechanism can legitimately be unavailable.

### 4. New-defect sweep on the edited regions

Findings N1-N7 and N9 below. Regions checked: `spec.md` §1.3/§1.3.1/§2.1/§3/§4/§5 + REQ-BLE-001/002/005;
`plan.md` §F M1/M2/M3 + §E E8; `acceptance.md` §D.0/§D.1/§D.2/§D.3 + AC-BLE-003/004/005.

Clean in the edited regions: the `§3` matrix (all five REQs covered, AC-BLE-005's joint coverage now
matches `acceptance.md:20`, 0 orphans); the Out-of-Scope H3 set (6 headings, each with `-` bullets);
`spec.md §2.1`'s new third sentence, which is correctly conditional rather than contradicting the
"차이는 0" sentence above it; `plan.md:104`'s single-holder discipline for `§D.0` (correct
anti-divergence move); `acceptance.md:108`'s survival-as-signal note, which is a faithful mirror of
`plan.md:85`.

One cosmetic note, not carried as a defect: `spec.md:79` `### 1.3.1` is an H3 sibling of `### 1.3`
(`spec.md:51`), so the numbering implies a nesting the heading level does not. Harmless.

### 5. Citation sweep on edited regions

All newly written file:line citations resolve in this tree. Verified this iteration:

| Citation | Where | Resolves to | Status |
|---|---|---|---|
| `board_lock_unix.go:37-41` | `spec.md:73` (D5 fix) | `unix.Open` at `:37`, `unix.Flock` at `:41` | ✓ exact |
| `board_lock_unix.go:39` | `plan.md:87` (D5 fix) | `fmt.Errorf("open board lock %s: %w", …)` | ✓ exact |
| `board_lock_unix.go:42` | `spec.md:125`, `acceptance.md:33` | `_ = unix.Close(fd)` | ✓ exact |
| `board_store.go:165-181` | `spec.md:83` (new, §1.3.1) | `acquireBoardLockSerialized` opens at `:165` | ✓ |
| `zsyscall_darwin_arm64.go:1337` | `spec.md:77` (new, §1.3-6) | `func Flock(fd int, how int)` at `:1337` in `x/sys@v0.47.0` (`go.mod:30` confirms the version) | ✓ exact |
| `board_store.go:173` · `integration_lock_mutation.go:103` · `backlog_store.go:736` | `acceptance.md:89`, `spec.md:139-141` | the only three non-test `IsBoardLockHeld` call sites | ✓ |

**On the one the repair pass deliberately left — `spec.md §1.2`'s `board_lock_unix.go:43` label over a
fence spanning `:41-44`: acceptable, but change it while the file is open.** The reasoning: `:43` is
`return nil, ErrBoardLockHeld`, which is the precise line where `err` is discarded — and discarding
`err` is exactly the claim `§1.2` makes (*"`err` 가 버려진다"*). The citation is therefore **true**, not
drifted, which is why iter-1 resolved it as accurate. The cost is that a reader resolving the label
lands mid-block rather than at the fence's first line. That is imprecision, not falsity; it does not
meet the bar for a defect, and `:41-44` is the one-character fix. Same disposition for the identical
label at `plan.md:7` and `plan.md:134`.

### 6. Score monotonicity

**0.85 → 0.83. Down by 0.02.**

Per `plan-auditor.md` § LEAN Workflow Additions, a lower iter(N+1) aggregate emits a **STOP** signal.
It is emitted here, and it should be read for what it is rather than as an alarm: the regression is
narrow and traceable to one dimension. Completeness rose materially (+0.04). Clarity held. The loss
sits in Testability (−0.04) and Traceability (−0.10), and it has one cause: **iter-2 replaced omissions
with affirmative claims, and three of those claims do not hold** — M-leak REDs AC-BLE-003 (it cannot be
planted), the RED-now cells are measurable at `3f03d9c36` (they are not), and the three mutants supply
non-vacuity for 001a·003·004 (nothing targets 004). Iter-1's failures were gaps; iter-2's are
assertions. Under `verification-claim-integrity.md` §1 that is the more expensive direction, because a
gap announces itself and a false affirmative does not.

The STOP clause requires a scope-reduction proposal. Mine, and the reason I do **not** recommend a
plan-audit iteration 3:

- **N2 is design-coupled and belongs inside M1, not inside another plan iteration.** The M-leak/branch
  mismatch cannot be settled without settling `plan.md §F M1`'s open A-vs-B decision, which the plan
  already schedules as its first, review-gated step. Routing it into iter-3 would decide a design
  question in an audit loop.
- **N1, N3, N4, N5, N7, N9 are all documentation edits** confined to `plan.md §A`/`§D` and
  `acceptance.md §D.0`/`§D.2`. None requires re-deriving anything.
- **N6 splits**: the threshold constant and the empty-sweep clause are documentation; the `/dev/fd`
  platform claim wants one measurement on each platform, which is run-phase pre-flight work.

So: **option 2 — accept iter-2 with the debt enumerated below, close N1/N3/N4/N5/N7 as plan-artifact
edits before M1 begins, and fold N2 and N6 into M1's design review as inputs it must resolve.** Option 1
(scope reduction / SPEC split) is not warranted — the SPEC is not structurally over-scoped; Tier S
remains correctly derived and the artifact set is small. Option 3 (iterate again) buys the same edits at
a higher price.

---

## Defects Found (structured defect-list)

**N1 — `plan.md:8`, `:21`, `:62` (and `:41`) — the retracted universal reachability claim survives in
`plan.md`; D1 was repaired in `spec.md` only — Severity: major — Class: blocking.**
`plan.md:8` states *"도달 가능한 errno 가 `EWOULDBLOCK`/`EAGAIN` 하나뿐이고"*, the exhaustive claim
`spec.md §1.3-4` was rewritten to abandon. `plan.md:21` states the behaviour delta as *"0 (도달 불가
errno 만 재분류)"*, contradicted by `§1.3.1`, which reclassifies EINTR and establishes only that its
arrival is *implausible*. `plan.md:62` restates REQ-BLE-005 at its pre-narrowing scope (*"오늘 행동
불변"*) while `spec.md:196` now reads *"측정된 도달 가능 입력"*. `plan.md:41` (B-미측정) lists
`ENOLCK`/`EOPNOTSUPP` and omits EINTR, which `spec.md §4` and `§1.3-6` both add.
**Required fix:** rewrite `plan.md:8` as a measurement statement mirroring `spec.md:75`; change
`plan.md:21` to "0 on measured-reachable errnos; EINTR's reclassification is an accepted change
(`spec.md §1.3.1`)"; append the "measured-reachable" qualifier to `plan.md:62` and point it at
`spec.md §1.3.1`; add EINTR to `plan.md:41`.

**N2 — `plan.md:92` vs `plan.md:110` / `acceptance.md:83`, `:105` — M-leak mutates a branch the
prescribed design does not have, and its designated RED is unreachable by the specified induction;
AC-BLE-003 is consequently the one AC with no functioning non-vacuity guard — Severity: major —
Class: blocking.**
`plan.md:92` prescribes **one shared** `unix.Close(fd)` before the classification split
(`양쪽 → 반환 전에 unix.Close(fd)`), preserving the current single close at `board_lock_unix.go:42`.
M-leak is defined as removing `unix.Close(fd)` *from the non-contention branch* — a branch that shape
does not create, so the mutant is not writable. Independently, `acceptance.md:81` induces failures by
repeated acquisition on a held root, which at the real call site yields **contention only** (the SPEC's
own `§1.3` finding), and under design option A `classifyBoardFlockErr` never owns the descriptor — so
the non-contention return path is unreachable from a test without the option-B seam `plan.md:77` argues
against. Since `acceptance.md:33` and `:37` designate M-leak as AC-BLE-003's *sole* non-vacuity source,
AC-BLE-003 is currently unguarded.
**Required fix:** re-specify M-leak as removing the **shared** `unix.Close(fd)` that `plan.md:92`
prescribes — which the contention induction does reach — and record that the guard then covers
REQ-BLE-004's contention half only, with the non-contention half stated as unguarded and why. Resolve
this inside M1's A-vs-B decision, not after it: option B dissolves the reachability half of this finding
but not the branch-shape half.

**N3 — `acceptance.md:26`, `:31`, `:32`, `:126` — the RED-now cells for AC-BLE-001b and AC-BLE-002 are
pinned to `3f03d9c36`, a tree on which they cannot be measured — Severity: major — Class: blocking.**
The document-level pin (`:26`) and `§D.2`'s checklist (`:126`, *"`3f03d9c36` 에서 실측"*) both bind these
cells to the pre-repair tree. But both ACs exercise the classification seam — `acceptance.md:60`
*"substrate 가 그 실패를 분류하면"* — and under design option A no such entry point exists at
`3f03d9c36`; a test written against it does not compile. What one would actually observe there is a
build failure, not a criterion RED. Row `:31` half-concedes this by equating the pre-repair mapping with
"M-broad", but M-broad is a mutant of the *repaired* code, so the honest RED is a post-M1-stub
observation. Under `verification-completeness.md` §4, a cell must pin the SHA where the evidence was
actually collected; as written the pin names a measurement that cannot be taken — the exact shape §2.1's
evidence row records.
**Required fix:** drop the `3f03d9c36` pin for the 001b/002/005 rows and replace it with "pinned at
measurement time to the M1-stub commit"; keep the document-level `3f03d9c36` pin only for the
regression-guard rows, whose green-now observation genuinely is a `3f03d9c36` measurement.

**N4 — `acceptance.md:30`, `:37`, `:119-120` — `verification-completeness.md` §2.1's undecidable
disposition is cited as the authority and then its operative consequence is declined, undeclared —
Severity: major — Class: blocking.**
§2.1's disposition is scoped to a RED that *cannot be re-executed* (historical event / already-merged
state / external CI result); none applies to a criterion that is green-now. And its stated effect is
that the criterion *"loses release-blocking eligibility… and is not recorded as a pass"*.
`acceptance.md:120` retains release-blocking status explicitly: *"착지 차단 성질은 유지된다."* The
deviation is conservative and §2.1's actual harm is avoided by `:39`'s refusal to record unmeasured
cells as passes — but a [HARD] clause is invoked by name and overridden silently, which would mislead
any reader routing from this SPEC.
**Required fix:** cite §2's **mutant probe** as the adoption basis for the three regression-guards
(which they pass), not §2.1; and state the retention of blocking status as a declared, stricter-than-
doctrine deviation, naming `:39` as the clause that honors §2.1's purpose.

**N5 — `acceptance.md:34`, `:37` — AC-BLE-004 is swept into a blanket non-vacuity claim that no mutant
supports — Severity: major — Class: blocking.**
`:37` asserts that the three mutants supply non-vacuity for 001a·003·004; the mutant table (`:101-105`)
maps M-broad→001b, M-narrow→001a, M-leak→003. Nothing designates 004. Row `:34` names no source at all,
recording it as "회귀-가드 (정의상)" — a definition standing in for a guard. A candidate exists (M-narrow
would plausibly flip the existing contention tests and thus the differential), but the SPEC does not
name it, and I did not run it.
**Required fix:** either name M-narrow as 004's incidental guard and reconcile with `:107`'s
each-mutant-REDs-a-different-AC clause, or state plainly that AC-BLE-004 is a comparative baseline gate
with no mutant guard, and give the reason that is acceptable. Then narrow `:37` to the two ACs it
actually covers.

**N6 — `acceptance.md:81` — the `/dev/fd` mechanism has an unpinned pass threshold, an unlabelled
platform claim, and no empty-sweep clause — Severity: major — Class: blocking.**
(a) `post ≤ pre + 고정 여유(테스트 자신이 여는 파일 수)` never gives the slack a value, so the implementer
calibrates the threshold rather than reading it. (b) *"darwin·리눅스 모두 이 경로가 프로세스 자신의 열린
디스크립터를 노출한다"* is stated as fact and was measured on neither platform — a lone unlabelled
inference in a document that elsewhere labels them scrupulously (`spec.md:77`). On Linux `/dev/fd`
resolves through `/proc`, whose availability is environment-dependent, and the CI runner is ubuntu while
the plan-phase measurement was darwin. (c) AC-BLE-003 carries no selector-match-count or skip-surfacing
clause, unlike AC-BLE-001a (`:54`), so a runner where `/dev/fd` is unreadable yields `ok` and the
criterion asserts nothing — `verification-completeness.md` §1.1's empty-sweep hazard.
**Required fix:** pin the slack to a literal (0, or a named constant with its derivation); relabel the
platform sentence as an assumption to be confirmed in M2 pre-flight on both platforms; and add to
AC-BLE-003 the same empty-sweep clause AC-BLE-001a carries, stating explicitly that a skip is a gap and
never a pass.

**N7 — `acceptance.md:126` ("RED-now 셀 3개") vs `plan.md:66` E8 ("RED-now 셀 6개") — Severity: minor —
Class: blocking.**
`§D.0` has six rows but only three carry a RED starting colour. `§D.2` counts three; `plan.md` §E E8
demands six verbatim measurements. The §E report will either over-claim three cells that do not exist or
under-deliver against E8.
**Required fix:** change `plan.md:66` E8 to "RED-now 셀 3개(001b·002·005) + 회귀-가드 3개(001a·003·004)의
등급 보고", matching `§D.2`'s two checklist lines (`:126`, `:127`).

**N8 — `spec.md:40` (mirrored at `plan.md:7`, `:134`) — `board_lock_unix.go:43` labels a fence spanning
`:41-44` — Severity: minor — Class: optional.**
The cited fact is true at the cited line (`:43` is `return nil, ErrBoardLockHeld`, where `err` is
discarded — precisely §1.2's claim), so this is imprecision rather than drift. Judgment: **acceptable**;
not a defect that blocks.
**Required fix:** none required; `:41-44` is a one-character improvement worth taking while the file is
open.

**N9 — `spec.md:26` — the HISTORY 0.2.0 row omits the `acceptance.md §D.0` work — Severity: minor —
Class: optional.**
The row records the D1, D4 (§3 matrix) and D5 changes but not the D3 response, which is the single
largest edit of the pass. `acceptance.md` carries no HISTORY of its own, so `spec.md`'s is the only
version record for the SPEC directory.
**Required fix:** add "`acceptance.md §D.0` 신설 — AC 별 시작 색 + 대체 수단 선언(D3)" to the 0.2.0 row.

### Carried over from iter-1, unchanged, all optional

- **D6** (consumer count is production-only; test consumers unnamed) — open. Correction to my iter-1
  count: there are **six** test consumers, not five — `backlog_store_errors_test.go:76` also calls
  `IsBoardLockHeld` and was missing from my list.
- **D7** (`"lock board lock %s"` on the integration-mutation path, `plan.md:91`) — open, and slightly
  reduced: AC-BLE-002 (`acceptance.md:70`) now asserts only that the message *contains the lock path*,
  no longer the exact string, so the SPEC is less committed to the mislabel than iter-1 judged. The
  operator-facing wording is still wrong for `integration_lock_mutation.go:99`.
- **D8** (`related_specs:` undocumented frontmatter field, `spec.md:16`) — open, harmless.
- **D9** (Tier table omits the LOC axis, `plan.md:17-24`) — open.
- **D10** (EWOULDBLOCK/EAGAIN portability note, `plan.md:95`) — open, unchanged text.

---

## Regression Check — iter-1 defects D1-D5

| # | Status | Evidence read |
|---|---|---|
| **D1** | **partially closed** | `spec.md:75` is now a measurement statement explicitly disclaiming exhaustiveness; `:77` adds EINTR to the unmeasured list with a labelled inference; `:79-91` (`§1.3.1`) records the reclassification as an accepted, non-neutral behaviour change with a stated decision; `:127`/`:131` narrow REQ-BLE-005 with the reason; propagated to `:145`, `:176`, `:185-186`, `:196`, `acceptance.md:19`, `:85`, `:138`. **Open:** the retracted claim survives at `plan.md:8`, `:21`, `:62` (N1). |
| **D2** | **closed** | `plan.md:81-85` names M-narrow as the wiring detector, with the correct mechanism: 001b feeds the classifier directly and never traverses `acquireBoardLockImpl`, so M-broad says nothing about wiring; 001a goes through `AcquireBoardLock` twice, so its RED under M-narrow is the evidence the call site returns the classifier's result. `plan.md:85` and `acceptance.md:108` both state the survival-as-signal corollary. Mirrored into AC-BLE-005's note as required. |
| **D3** | **partially closed** | `acceptance.md §D.0` (`:22-39`) added with a per-AC starting-colour matrix, an explicit substitution declaration (`:37`), and an honest refusal to pre-fill the cells (`:39`). The direction of the disagreement is decided **in the repair's favour** (see Q2). **Open:** N3 (unmeasurable pins), N4 (§2.1 mis-cited and overridden), N5 (004 unsourced), N7 (count mismatch). |
| **D4** | **partially closed** | The "or equivalent" alternative is struck and the deletion is recorded in-line with the correct doctrinal reason (`acceptance.md:80`); the `/dev/fd` mechanism is named (`:81-82`) with `t.Cleanup` and no background load; M-leak added to the mutant table (`:105`) and to `plan.md:110`. **Open:** N2 (M-leak not writable / RED unreachable), N6 (mechanism under-specified). |
| **D5** | **closed** | `spec.md:73` now cites `board_lock_unix.go:37-41` — verified: `unix.Open` at `:37`, `unix.Flock` at `:41`. `plan.md:87` now cites `board_lock_unix.go:39` — verified: `fmt.Errorf("open board lock %s: %w", lockPath, err)`. Both exact. |

No defect appears unchanged across both iterations in a way that indicates no progress; the stagnation
clause does not fire. Every iter-1 blocking defect moved.

---

## Gaps — what I did NOT verify

1. **No Go test was executed.** `go test ./internal/kanban/...` was not run — this is the lock axis and
   the brief forbids background load. I have no measurement of the current pass/fail set, no evidence
   about the `TestConcurrencyStress` flake lineage, and no evidence for or against AC-BLE-004's
   feasibility.
2. **No mutant was planted.** M-broad, M-narrow, and M-leak were all left unwritten. **N2 is a reading
   of the SPEC's own two statements** — `plan.md:92` prescribes one shared `unix.Close`, while
   `acceptance.md:105` mutates a per-branch one; plus `spec.md §1.3`'s own finding that only contention
   is inducible at the real call site. It is a documentary contradiction verified as text. I ran nothing
   to confirm the runtime consequence.
3. **My M-narrow → AC-BLE-004 inference is unverified.** In Q2 I named M-narrow as a plausible
   incidental guard for AC-BLE-004. I did not run it and do not carry it as a finding — only as a
   candidate the SPEC should name or reject.
4. **`/dev/fd` behaviour was measured on neither platform.** N6 asserts that the SPEC's platform claim is
   *unlabelled and unmeasured*, not that it is false. I made no observation of `/dev/fd` on darwin or
   Linux.
5. **`moai spec lint` was not run by me.** The `0 error(s), 1096 warning(s)`, rc 0 figure from the
   tree-built `./bin/moai` is the lead's measurement, attributed to the lead. I did not build a binary
   and did not reproduce it; it is not carried as evidence in this report.
6. **`GOOS=windows GOARCH=amd64 go build ./...` was not run.** The Windows-build constraint
   (`spec.md:195`, `acceptance.md:93`) is unverified against this tree. `go vet` was not re-run either
   (iter-1's rc 0 stands as iter-1's measurement, not this one's).
7. **`progress.md` was not read.** It was not in the edited-region list; its `§E`/`§F` state is
   unobserved and its `status:`-freedom rests on the lead's stated count, not mine.
8. **EINTR, ENOLCK, and EOPNOTSUPP were not induced.** Consistent with the SPEC's own exclusions. I add
   no measurement and confirm none exists.
9. **Linux unobserved.** All reading was on darwin/arm64. Nothing here is evidence about the ubuntu CI
   runner.
10. **`.moai/reports/t372/verdict.md` §9.3 was not read** — not present on this branch.
11. **No cross-model backend audit.** Claude-only verdict; `audit_multi` / `codex_audit` / `glm_audit`
    were not invoked.
12. **Delta-scoped citation sweep.** Per the Retry Loop Contract I re-verified citations in the edited
    regions plus the consumer and `zsyscall` references those edits touch. I did not re-verify iter-1's
    full 13-citation set.

---

## Residual risk

The repair pass did the harder half of its job well: it accepted a finding it could have argued away
(D1) and repaired it with more honesty than the finding demanded, and it pushed back on a finding I got
wrong (D3) with a correct reading of the doctrine. Both are the right instincts.

The risk it introduces is the mirror of the risk it removed. Iter-1's SPEC had gaps; iter-2's SPEC has
scaffolding, and three struts of that scaffolding do not bear load — M-leak cannot be planted against
the design the plan recommends, the RED-now pins name a tree that cannot carry the measurement, and the
non-vacuity claim reaches an AC no mutant targets. A gap is visible to whoever next reads the document.
A stated guard that cannot fire is not: it will be discovered at M3, by a manager-develop agent that
tries to plant M-leak and finds nothing to remove — and the cheapest failure mode there is a blocker
report, while the expensive one is a plausible-looking substitute mutant invented on the spot to make
the step complete.

The narrow reading of this audit is therefore: the SPEC is sound, its two open items are cheap, and one
of them (N2) is genuinely a design question that M1 was already going to answer. The wider reading is
that a verification artifact gets more dangerous as it gets more elaborate, and that this SPEC has now
crossed the point where its verification machinery needs to be checked as carefully as the code it
guards.

---

## Recommendation

**PASS-WITH-DEBT at 0.83 (Tier S threshold 0.75), with a STOP signal on the −0.02 regression.** The
SPEC is not blocked. Route the fixes in this order:

1. **N2** — resolve M-leak's branch and induction inside `plan.md §F M1`'s A-vs-B design review, since
   the answer depends on that decision. Do not settle it in another audit iteration.
2. **N6** — pin the `/dev/fd` slack constant and add the empty-sweep clause now; schedule the two
   platform confirmations as M2 pre-flight.
3. **N1** — sweep `plan.md §A`/`§D` to the narrowed scope. One-file edit, four lines.
4. **N3, N4, N5, N7** — `acceptance.md §D.0`/`§D.2` and `plan.md:66`. Documentation only.
5. **N8, N9** and the carried-over D6-D10 — optional, orchestrator's discretion; N8 and N9 are one-line
   fixes worth taking while the files are open.

Per the STOP clause, this verdict is presented to the operator with three options: **(recommended)**
accept iter-2 with the debt above and proceed to M1 with N2/N6 as named M1 inputs; **or** reduce scope
(not warranted — Tier S is correctly derived and the artifact set is small); **or** explicitly override
and run iteration 3 (buys the same documentation edits at a higher price, and cannot settle N2 without
M1's design decision).

Verdict authority is this agent's. The defect list above is the machine-consumable fix route; a
confirming re-audit, if the operator orders one, should be scoped to that delta rather than run from
scratch.
