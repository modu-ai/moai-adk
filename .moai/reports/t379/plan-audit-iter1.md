# SPEC Review Report: SPEC-BOARDLOCK-ERRNO-001

Card: t379
Iteration: 1/1 (Tier S ceiling — `harness.plan_audit_tier_ceilings` S=1)
Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t379`, branch `WT-boardlock-errno`, HEAD `3f03d9c36`
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85** (harmonic mean of the four dimensions below; Tier S PASS threshold 0.75)

Reasoning context ignored per M1 Context Isolation. The dispatching session's pre-plan narrative
was read as a set of claims to be re-verified, not as findings to be adopted; every citation below
was re-resolved against this tree.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-BLE-001` … `REQ-BLE-005` at `spec.md:94,98,102,106,110`.
  Sequential, no gaps, no duplicates, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-XXX` in
  `spec.md`), never against the ACs. All five match a GEARS pattern:
  001/002 event-driven (`When [trigger], the <subject> shall …`); 003/005 ubiquitous
  (`The <subject> shall …`); 004 state-driven (`While [condition], … shall …`). The six
  Given-When-Then entries in `acceptance.md` are verification-layer artifacts and are graded under
  Group 4, per M3 § Scope.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with canonical names
  (`spec.md:2-13`): `id`, `title`, `version: "0.1.0"` (quoted semver), `status: draft`,
  `created`/`updated` ISO (`2026-08-31`), `author`, `priority: P3`, `phase: "v3.1.4 target"`
  (a release target, not a prohibited lifecycle stage), `module`, `lifecycle: spec-anchored`,
  `tags`. No rejected snake_case alias present. `acceptance.md`/`plan.md`/`progress.md` carry no
  `status:` field — compliant with § Artifact Statelessness.
- **[N/A] MP-4 Section 22 language neutrality** — single-language SPEC (Go, `internal/kanban`);
  no multi-language tooling surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — extracted refs (`grep -Eo` over the SPEC dir):
  `SPEC-KANBAN-BOARD-001` → `status: completed` (`spec.md:5`), `SPEC-STRESS-INVARIANT-VERDICT-001`
  → `status: implemented` (`spec.md:5`). Neither is `retired`/`superseded`/`archived`; both exist
  on disk. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `syscall` appears once per document
  (`plan.md:76` prose "syscall 경로"; `spec.md §1.3-6` inside the filename
  `zsyscall_darwin_arm64.go`). Explicit build-tag discipline is present and load-bearing:
  `plan.md:40` (`//go:build !windows`, `GOOS=windows` build mandatory), `plan.md:97` (new test file
  carries the tag), `spec.md:174` (`GOOS=windows go build ./...` must keep passing). Proximity note:
  the tag clauses are not in the same section as `plan.md:76`, but the cross-platform obligation is
  unambiguously stated and mechanically checkable, so this is not escalated to BLOCKING.
- **[PASS] MP-7 clarification gate** — `grep -rn "NEEDS CLARIFICATION" .moai/specs/SPEC-BOARDLOCK-ERRNO-001/`
  → no output, rc=1. Tier S carries no `research.md`; `plan.md` scanned and clean.

---

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | Every REQ has one reading; the single open design decision is fenced and named (`plan.md:74-79`, option A recommended with its own risk stated). Deduction: `spec.md §1.3-4` and `§1.3-5` contradict each other (D1). |
| Completeness | 0.88 | 0.75-1.0 | 6 `### Out of Scope — <topic>` H3 sub-headings each with `-` bullets (`spec.md:149-165`); HISTORY/WHY/WHAT/REQ/AC-map/constraints all present; unmeasured items recorded as unmeasured in three places (`spec.md §1.3-5`, `§4`, `acceptance.md §D.3`). Deduction: EINTR absent from the unmeasured list (D1). |
| Testability | 0.72 | 0.50-0.75 | AC-BLE-005's mutation pair is a genuine bidirectional non-vacuity instrument and is mandatory, not optional. Deductions: no RED-now cell on any of six landing-blocking ACs (D3); AC-BLE-003's "equivalent" alternative names no mechanism and REQ-BLE-004 has no mutant guarding it (D4). |
| Traceability | 1.00 | 1.0 | `spec.md §3` maps REQ-BLE-001→AC-BLE-001a, 002→001b, 003→002, 004→003, 005→004, plus AC-BLE-005 as a joint non-vacuity criterion over 001+002. Verified against `acceptance.md §D`: 0 uncovered REQ, 0 orphan AC. AC bodies live in exactly one file. |

---

## Answers to the eight audit questions

1. **Measured zero stated plainly and early — YES, strongly.** `spec.md §1.1` is titled
   "먼저, 숫자부터" and its first body sentence is *"이 호출 지점에서 실측된 오분류는 0건이다"*
   (`spec.md:33`), before any discussion of the defect shape. `§1.2` then separates the shape claim
   from the misclassification claim explicitly. The card's "over-broad" wording cannot be read as
   live breakage from this document. This is the SPEC's strongest property.
2. **Bidirectional ACs — YES, and neither passes trivially.** `acceptance.md:22` states the trap
   directly: 001a alone admits an always-true predicate (the current defect), 001b alone admits an
   always-false one (rule disablement). Checked mechanically against the mutant table: an
   always-false predicate fails AC-BLE-001a (real path, contention must still report held);
   an unchanged predicate fails AC-BLE-001b (`ENOLCK`/`EBADF` must report not-held). AC-BLE-001b's
   `err != nil` assertion additionally blocks the swallow-to-nil escape (`acceptance.md:42`).
3. **Vacuity — mostly guarded, with one misattribution and one hole.** The mutation pair is
   coherent as a *pair*: M-broad REDs 001b only, M-narrow REDs 001a only, so no single AC responds
   to both. But `plan.md:79` names the wrong mutant as the guard against its own stated wiring risk
   (D2), and the fd-hygiene branch (REQ-BLE-004 / AC-BLE-003) has no mutant at all (D4).
   Zero-execution is blocked for AC-BLE-001a by the selector-match-count clause (`acceptance.md:33`)
   and globally by `§D.2` ("각 판정에 명령 + 출력 전문 + 선택자 매치 수").
4. **Unmeasured recorded as unmeasured — YES for the named items, NO for EINTR.**
   `ENOLCK`/`EOPNOTSUPP`: "0 이 아니라 미측정이다" (`spec.md:75`), repeated at `§4` and
   `acceptance.md:107`. Linux: "CI 는 ubuntu 이므로, 아래는 리눅스 커널의 관측이 아니다"
   (`spec.md:57`) plus `acceptance.md:108`. EINTR is named at `spec.md §1.3-6` but never enters
   the unmeasured list — see D1.
5. **Consumer behaviour change stated — YES, both sentences.** `spec.md §2.1` carries the table of
   three consumers, then: *"이것이 의도다 … 그리고 오늘 관측 가능한 차이는 0 이다 … 두 문장이 함께
   있어야 한다"* (`spec.md:124`). REQ-BLE-005 pins the second sentence as non-negotiable
   (`spec.md:110`, `spec.md:175`), with the mis-attribution rationale stated at `spec.md:112`.
6. **Tier S derivation — sound.** Re-derived independently: 1 package, 2 expected files, no external
   contract change (`IsBoardLockHeld` signature and sentinel unchanged — confirmed at
   `board_lock.go:25,28`), no migration, one reversible design decision. Tier S budgets are
   REQ ≤ 8 (SPEC has 5) and AC ≤ 8 (SPEC has 6) — both clear. The 4-artifact set exceeds the Tier S
   2-artifact default; `plan.md:28` records the deviation and its reason (dispatch required
   `acceptance.md`; AC bodies kept in one place to prevent divergence). Documented deviation, not
   drift. Minor gap: the derivation table omits the canonical primary axis, LOC (D9).
7. **Scope discipline — clean.** Windows is stated as the reference surface, not a target
   (`spec.md:88,150`); `plan.md:33` PRESERVE-lists `board_lock_windows.go`, `board_lock.go`,
   the three consumers, `board_recover.go`, `internal/spec/**`, `.github/workflows/**`. No
   requirement reaches outside `internal/kanban`. `internal/spec/lock_unix.go` is explicitly
   excluded (`spec.md:161-162`).
8. **Citation accuracy — 10 of 12 exact, 2 drifted.** Verified in this tree:
   `board_lock.go:28` ✓ · `board_lock_windows.go:69` ✓ · `board_store.go:173` ✓ ·
   `integration_lock_mutation.go:103` ✓ · `backlog_store.go:736` ✓ · `board_store.go:289` ✓ ·
   `backlog_store.go:677` ✓ · `board_store.go:237` ✓ · `board_recover.go:77` ✓ ·
   `backlog_store.go:623` ✓ · `backlog_store.go:644` ✓ · `board_lock_unix.go:43` (sentinel return) ✓ ·
   `board_lock_unix.go:42` (close) ✓. Two drifted — see D5.

---

## Defects Found

**D1 — `spec.md:74` vs `spec.md:75` — internal contradiction in the reachability claim, and an
unenumerated errno stated as exhaustive — Severity: major — Class: blocking.**
`§1.3-4` concludes *"이 호출 지점에서 실제로 도달 가능한 errno 는 EWOULDBLOCK/EAGAIN 하나뿐"* —
a universal claim over an errno set the probe never enumerated (three cases were induced: A, B, C).
The very next bullet, `§1.3-5`, says `ENOLCK`/`EOPNOTSUPP` *"다른 파일시스템·커널에서 여전히
그럴듯하다"*. Both cannot hold: -4 says one errno is reachable, -5 says two others remain plausible.
Separately, `§1.3-6` records that `unix.Flock` is a raw wrapper with no EINTR retry — which makes
EINTR a concrete reachable-in-principle candidate at this call site, and one whose reclassification
is not behaviour-neutral: today it maps to the contention sentinel and is absorbed by the retry
budget at all three consumers; after narrowing it becomes an immediate hard error. That is a
behaviour change on a *plausibly reachable* input, which REQ-BLE-005 ("오늘 도달 가능한 입력에서
행동 불변") declares non-negotiable. The SPEC never connects -6 to -4.
**Required fix:** rewrite `§1.3-4` to state only what was measured — "of the errnos induced here,
only EWOULDBLOCK/EAGAIN is producible through the real call site; the set was not enumerated
exhaustively" — and add EINTR to `§1.3-5` and `acceptance.md §D.3` as unmeasured, with the
retry→hard-error consequence stated explicitly. If EINTR is to be treated as contention-equivalent
(retryable), say so in REQ-BLE-002; if not, record it as an accepted behaviour change.

**D2 — `plan.md:79` — the vacuity guard is attributed to the wrong mutant — Severity: major —
Class: blocking.**
`plan.md:79` states that under design option A the risk is "분류 함수는 옳은데 호출 지점이 그것을
쓰지 않아도 AC-BLE-001a 는 통과한다", and that "M3 의 M-broad 뮤턴트가 정확히 이 구멍을 겨냥한다".
Traced against the AC bodies, that is backwards. AC-BLE-001b feeds synthetic errnos to the
classifier directly (`acceptance.md:38-39`, "substrate 가 그 실패를 분류하면"), so it never
traverses `acquireBoardLockImpl`; M-broad mutates the classifier and therefore REDs 001b whether or
not the call site is wired — it detects nothing about wiring. M-narrow is the wiring detector:
it makes the classifier return non-contention for every errno, and AC-BLE-001a (which *does* go
through `AcquireBoardLock` twice, `acceptance.md:29`) goes RED **only if** the call site actually
consumes the classifier. An unwired implementation leaves 001a green under M-narrow, which is
exactly the AC-BLE-005 failure the pair is supposed to surface.
**Required fix:** in `plan.md:79`, replace M-broad with M-narrow and state the reason — M-narrow's
RED on AC-BLE-001a is the evidence that `acquireBoardLockImpl` returns the classifier's result
rather than a hard-coded sentinel. Mirror the note into `acceptance.md` AC-BLE-005's M-narrow row.

**D3 — `acceptance.md:26-85` — six landing-blocking ACs adopted with no RED-now cell —
Severity: major — Class: blocking.**
`§D.1` classifies all six ACs as blocking ("하나라도 실패하면 착지 불가"), but not one carries the
two-cell adoption pair required by `.claude/rules/moai/development/verification-completeness.md`
§2 [HARD], nor the four elements §2.1 requires of a blocking criterion (single-invocation command /
verbatim stdout / exit code as its own field / pinned tree SHA). AC-BLE-001a cites the plan-phase
probe (`acceptance.md:34`) — but that is a *green* prior observation, not a RED-now cell; it records
that the behaviour already holds, which is the vacuity direction §2 exists to catch. `§D.2` partly
compensates by demanding a selector-match count on every verdict, and AC-BLE-005 is a stronger
non-vacuity instrument than RED-now for this SPEC's shape — but neither substitution is declared,
and neither is a starting observation pinned to `3f03d9c36`.
**Required fix:** add a RED-now cell per AC pinned to `3f03d9c36`, each carrying the command, its
verbatim stdout, and its exit code. Where the RED is "the contract test does not exist yet", record
the runner's own empty-sweep token (`no tests to run` / `[no test files]`) rather than a bare
non-zero exit — §1.1 names a zero-match selector as the exact false-RED hazard. Where a RED-now cell
is genuinely unobtainable, apply §2.1's undecidable disposition (reclassify as regression-guard) and
say so.

**D4 — `acceptance.md:59-60` — AC-BLE-003's "equivalent" alternative names no mechanism, and
REQ-BLE-004 has no mutant — Severity: major — Class: blocking.**
The primary judgment (fd count does not increase monotonically over N≈200 induced failures) is
binary-testable, though the counting mechanism is unnamed. The alternative offered at
`acceptance.md:60` — "코드 경로 단언 테스트로 두 분기 모두 `unix.Close` 를 지나감을 잠근다" — names
no mechanically checkable observation at all; "asserting a code path" is not something a Go test
does without instrumentation the SPEC does not specify, yet it is declared "동등 인정". This fails
verification-completeness §1.2(b): a check specification must name the input that turns it red.
Compounding it, neither AC-BLE-005 mutant touches the fd-close branch, so REQ-BLE-004 — the property
`plan.md`/`spec.md:108` themselves identify as "분기가 둘로 갈리는 변경에서 가장 흔하게 깨지는
성질" — is the one requirement with no non-vacuity guard.
**Required fix:** strike the alternative, or name its mechanism concretely (e.g. count entries under
`/dev/fd` before and after, or assert on a `unix.Close` counter injected for the test). Add a third
mutant M-leak (remove `unix.Close` from the non-contention branch) whose designated RED is
AC-BLE-003, so REQ-BLE-004 is guarded like the other two directions.

**D5 — `spec.md:72` and `plan.md:81` — two file:line citations drifted — Severity: minor —
Class: optional.**
`spec.md §1.3-2` cites `board_lock_unix.go:38-43` for the "Flock immediately follows a successful
`unix.Open`" adjacency; in this tree the `unix.Open` is at `:37`, its error return at `:38-40`, and
the `unix.Flock` at `:41`. `plan.md:81` cites `board_lock_unix.go:36` as the home of the
`open board lock %s: %w` idiom it copies; `:36` is the function signature — the string is at `:39`.
Both cite a real fact at a wrong line. Every other cited line resolved exactly.
**Required fix:** change to `board_lock_unix.go:37-41` and `board_lock_unix.go:39`.

**D6 — `spec.md:116` — the consumer count is production-only and the test consumers are unnamed —
Severity: minor — Class: optional.**
"`IsBoardLockHeld` 소비자는 셋" is accurate for non-test code and I confirmed it
(`board_store.go:173`, `integration_lock_mutation.go:103`, `backlog_store.go:736` — the only three
non-test call sites). The narrowing also reaches five test consumers that constrain the baseline
AC-BLE-004 will be diffed against: `board_lock_wait_test.go:189`, `backlog_store_test.go:198`,
`backlog_concurrency_test.go:59` and `:275`, and the negative guard
`integration_lock_cross_test.go:362` (which asserts the board sentinel does *not* leak out of the
integration scope). AC-BLE-004 covers them generically as "기존 경합 테스트들".
**Required fix:** name the five test call sites in `spec.md §2.1` or in AC-BLE-004, so a baseline
diff is attributable per-site rather than as an aggregate pass/fail.

**D7 — `plan.md:85` — the proposed error text says "board lock" on a path that may be the
integration mutation lock — Severity: minor — Class: optional.**
`integration_lock_mutation.go:99` calls `acquireBoardLockImpl(path)` directly, with the *integration
mutation lock* path — not the board lock path. The proposed non-contention wrapper
`fmt.Errorf("lock board lock %s: %w", lockPath, err)` would therefore emit "lock board lock
<integration-lock-path>" for that caller. AC-BLE-002 asserts on this message
(`acceptance.md:49`), so it is an operator-facing string the SPEC has committed to.
**Required fix:** use a scope-neutral phrasing (e.g. `"lock %s: %w"`), or have the caller supply the
scope label the way `joinBoardReleaseErr` already takes an `op` string (`board_store.go:289`).

**D8 — `spec.md:16` — `related_specs:` is an undocumented frontmatter field — Severity: minor —
Class: optional.**
Not in `spec-frontmatter-schema.md` § Optional Fields (which lists `issue_number`, `depends_on`,
`lint.skip`, `bc_id`, `amendment_of`, `tier`). It is not rejected either — `FrontmatterSchemaRule`
checks only the 12 required fields — so no lint error results. `era: V3R6` is documented, in
`.claude/rules/local/lifecycle-sync-gate.md`.
**Required fix:** none required; either drop it or use `depends_on:`/prose cross-references, which
are the documented carriers.

**D9 — `plan.md:17-24` — the Tier derivation table omits the canonical primary axis —
Severity: minor — Class: optional.**
Six axes are listed (package / files / behaviour delta / design decisions / external contract /
migration), but not the LOC estimate, which is the first column of the canonical Tier table in
`spec-workflow.md` § SPEC Complexity Tier. The conclusion is right either way.
**Required fix:** add an estimated-LOC row.

**D10 — `plan.md:89` — the EWOULDBLOCK/EAGAIN portability note is slightly wrong —
Severity: minor — Class: optional.**
"두 상수는 리눅스에서 같은 값이고 darwin 에서도 프로브가 둘 다 `true` 로 관측했으므로" implies the
identity is a Linux property. Both constants are identical on darwin as well, which is precisely why
the probe reported `EWOULDBLOCK=true EAGAIN=true` for a single errno. The chosen `errors.Is(…) ||
errors.Is(…)` form is correct on both platforms regardless (and is the right choice over a `switch`,
which would not compile with two identical cases).
**Required fix:** restate as "the two constants are the same value on both Linux and darwin; writing
both is portability notation, and the `||` form is used rather than a `switch` because duplicate
case values would not compile".

---

## Verification actually run (Claim / Evidence / Baseline-attribution)

- **Tree identity.** `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t379`;
  `git branch --show-current` → `WT-boardlock-errno`; `git rev-parse --short HEAD` → `3f03d9c36`.
- **Tree cleanliness / probe reversion.** `git status --short` → exactly two untracked entries
  (`.moai/reports/t379/`, `.moai/specs/SPEC-BOARDLOCK-ERRNO-001/`). No probe file remains in
  `internal/kanban`, confirming `spec.md:52`'s "패키지는 지금 깨끗하다" claim at the tracked-file level.
- **Package compiles.** `go vet ./internal/kanban/` → no output, exit 0. This grounds "clean" at the
  compile/vet level only; it is not a test-suite claim.
- **Citations.** `cat -n internal/kanban/board_lock_unix.go`, `cat -n internal/kanban/board_lock.go`,
  `grep -n "IsBoardLockHeld" internal/kanban/*.go`,
  `grep -n "joinBoardReleaseErr\|joinBacklogReleaseErr" internal/kanban/*.go`,
  `grep -n "ErrBoardLockHeld" internal/kanban/board_lock_windows.go`. Results in question 8 above.
- **Probe evidence.** `cat .moai/reports/t379/errno-probe-output.txt` — the six logged lines match
  the block quoted at `spec.md:59-67` in content (whitespace reflowed). Both cited evidence files
  exist and are non-empty.
- **Dependency version.** `grep -n "golang.org/x/sys" go.mod` → `golang.org/x/sys v0.47.0`, matching
  `spec.md §1.3-6`.
- **D7/D8/MP-7 scans.** Commands and outputs recorded in the Must-Pass section above.

---

## Gaps — what I did NOT verify

1. **`go test ./internal/kanban/...` was not run.** This is the lock axis and the brief forbids
   background load; the pre-repair baseline is designated run-phase pre-flight work (`plan.md:49`).
   I therefore have **no measurement** of the current pass/fail set, no confirmation that the
   `TestConcurrencyStress` flake lineage (`plan.md:38`) is real in this tree, and no evidence for or
   against AC-BLE-004's feasibility beyond compilation.
2. **`moai spec lint` was not re-run.** The dispatch reports `0 error(s), 1096 warning(s)`, rc 0 from
   a tree-built binary. I did not build a binary from this tree and did not reproduce it. That figure
   is the dispatcher's measurement, not mine, and is not carried as evidence here.
3. **`GOOS=windows GOARCH=amd64 go build ./...` was not run.** The Windows-build constraint
   (`spec.md:174`) is unverified against this tree.
4. **EINTR reachability was not measured.** D1 asserts that EINTR is reachable *in principle* at a
   raw non-retrying `flock` wrapper and that the SPEC's `§1.3-4`/`§1.3-5` contradict each other. The
   contradiction is read directly off the text and is verified. The EINTR reachability claim is an
   inference from `spec.md §1.3-6` plus flock(2) semantics — I induced no EINTR and make no claim
   about its frequency.
5. **`ENOLCK` / `EOPNOTSUPP` were not induced.** Consistent with the SPEC's own exclusion; I add no
   measurement here and confirm none exists.
6. **Linux behaviour unobserved.** All of my work is darwin/arm64. Nothing here is evidence about
   the ubuntu CI runner.
7. **The three consumers' runtime behaviour under the narrowing was not simulated.** The consumer
   analysis in question 5 is a static read of `board_store.go:165-181`,
   `integration_lock_mutation.go:95-129`, and `backlog_store.go:736`; no behavioural difference was
   executed.
8. **`.moai/reports/t372/verdict.md` §9.3 candidate B was not read.** It is not present on this
   branch (the SPEC notes it is reachable only via `git show origin/develop:`). The SPEC's
   characterization of its own origin is therefore unverified; `spec.md §1.4`'s independent
   re-verification of the `errors.Join` axis *was* verified against this tree (all four fold call
   sites install after successful acquisition — confirmed at `board_store.go:237`,
   `board_recover.go:77`, `backlog_store.go:623`, `backlog_store.go:644`).
9. **No cross-model backend audit was run.** This is a Claude-only verdict; `audit_multi` /
   `codex_audit` / `glm_audit` were not invoked.

---

## Residual risk

The SPEC's single greatest strength — leading with the measured zero — is also where its one
substantive defect lives: having correctly refused to overstate the defect, `§1.3` then overstates
the *safety*, concluding an exhaustive reachability that its own next bullet contradicts and that
its own EINTR note undercuts. If run-phase proceeds on `§1.3-4` as written, an EINTR (or an
`ENOLCK` on a future NFS/overlay deployment) becomes a hard failure at all three consumers with no
record that the change was anticipated — the exact mis-attribution REQ-BLE-005 was written to
prevent, arriving through the door the SPEC left open rather than the one it closed.

Beyond that, the AC set is genuinely bidirectional and the mutation requirement is the right
instrument for a branch that cannot be induced today. The debt is concentrated in D1-D4 and all four
are cheap to close in the plan artifacts before run-phase entry.

---

## Recommendation

**PASS-WITH-DEBT.** All seven must-pass criteria clear and the 0.85 aggregate exceeds the Tier S
threshold of 0.75, so the SPEC is not blocked. Four blocking-class defects (D1-D4) should be closed
in the plan artifacts before run-phase M1 begins, in this order:

1. **D1** — repair the `§1.3-4`/`§1.3-5` contradiction; add EINTR to the unmeasured list with its
   retry→hard-error consequence, and decide explicitly whether REQ-BLE-002 treats it as contention.
2. **D2** — swap the mutant attribution at `plan.md:79` to M-narrow and state why.
3. **D4** — strike or mechanize AC-BLE-003's alternative judgment; add an M-leak mutant so
   REQ-BLE-004 is guarded.
4. **D3** — add RED-now cells pinned to `3f03d9c36`, using runner empty-sweep tokens rather than a
   bare non-zero exit.

D5-D10 are optional and left to the orchestrator's discretion; D5 (two drifted line numbers) is a
one-line fix worth taking while the file is open.

This is iteration 1 of 1 — the Tier S ceiling permits no iteration 2. The ceiling bounds iteration
count, not verdict authority: this verdict is the auditor's, and the four blocking findings above
are the fix route the orchestrator routes from directly rather than re-auditing from scratch.
