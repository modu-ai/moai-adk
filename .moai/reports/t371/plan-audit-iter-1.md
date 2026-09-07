# SPEC Review Report: SPEC-SPECLINT-GITBLIND-001

Iteration: 1/1 (Tier S ceiling)
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.75** (Tier S PASS threshold 0.75 — at threshold, not above it)

Audit tree: `.claude/worktrees/t371` @ `WT-lint-shallow-clone`, HEAD `1e5199b88`.
Artifacts read (Tier S contract): `spec.md`, `plan.md`, `progress.md`.
Reasoning context ignored per M1 Context Isolation — the operator's dispatch supplied a
ground-truth A/B table; that table was audited **against the code**, not accepted.

---

## Ground-truth verification (operator's table, re-derived from source)

Every premise the SPEC rests on was checked against the tree. All hold:

| Claim | Status | Evidence |
|---|---|---|
| `cachedMainBranch` checks local `main` only, falls back to literal `"master"` | CONFIRMED | `internal/spec/gitquery_cache.go:101-104`, `:112-115` — `git rev-parse --verify main`; on error `branch = "master"`, no remote candidate |
| Second consumer at `drift.go:68` (`DetectDrift`) | CONFIRMED | `internal/spec/drift.go:68` `branch: cachedMainBranch` inside `realDriftDeps`; `grep -rn cachedMainBranch internal --include "*.go"` returns exactly 2 call sites (`:68`, `:303`) |
| `StatusGitConsistency` emitted with `Advisory: true` | CONFIRMED | `internal/spec/lint.go:1316` — citation is line-exact |
| `OwnershipTransitionInvalid` emitted with **no** `Advisory` | CONFIRMED | `internal/spec/lint_ownership.go:429` is literally `Severity: SeverityWarning,` and the Finding literal (`:425-440`) carries no `Advisory` field. Citation is line-exact |
| A non-advisory warning reddens `--strict` | CONFIRMED | `internal/spec/lint.go:61` `if r.Strict && f.Severity == SeverityWarning && !f.Advisory` |
| Info never reddens `--strict` | CONFIRMED | same function, `:56-64`; Info hits neither branch. Regression fixture `lint_test.go:574-579` ("only info", `want: false`) exists as cited |
| `applyEraDemotion` marks grandfathered/terminal warnings advisory | CONFIRMED | `internal/spec/lint.go:284-300`. Note: the switch handles Error and Warning only — Info passes through untouched, which is what REQ-SLGB-005 needs |
| `spec-lint.yml` checkout has no `with:` block | CONFIRMED | `.github/workflows/spec-lint.yml:31` — bare `- uses: actions/checkout@v7` |
| `ci.yml` already carries `fetch-depth: 0` at 6 sites `:129 :264 :382 :431 :486 :543` | CONFIRMED | `grep -n fetch-depth .github/workflows/ci.yml` returns exactly those 6 lines |
| `main = 48239c7dc` | CONFIRMED | `git show-ref --verify refs/heads/main` → `48239c7dc7428c8751a04f6321887c2d36123884` |
| `StatusGitUnreachable` does not exist today | CONFIRMED | `grep -rn StatusGitUnreachable internal .github` → rc=1, zero matches |
| §B test-seam precedents exist | CONFIRMED | `withFakeOwnershipLookup` at `lint_ownership_test.go:81,183,225`; `drift_chore_skip_test.go`, `archive_git_test.go`, `closer_test.go` all present |

**One premise is wrong.** See D1.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-SLGB-001`..`008`, sequential, no gaps, no duplicates, consistent 3-digit zero-padding. `grep -o "REQ-SLGB-[0-9]*" spec.md | sort -u` returns exactly 8 entries.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`REQ-SLGB-*` in `spec.md` §2), not the AC layer. 001/005/006 Ubiquitous (`<subject> …한다`), 002 event-driven (`When … 실패하면`), 003/007 state-driven (`While …이면`), 004 unwanted (`… 내지 않는다` = canonical `shall not`), 008 Ubiquitous. All eight match a GEARS pattern. Nit recorded as D5b: REQ-SLGB-001 is *labelled* `(Ubiquitous)` but carries the embedded condition "관측하지 못한 경우", making it structurally state-driven/compound. Pattern-valid, label-wrong.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types (`id`, `title`, `version` quoted `"0.1.0"`, `status: draft`, `created`/`updated` ISO `2026-08-31`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` comma-separated string) plus `tier: S` and `era: V3R6`. No rejected snake_case alias (`created_at`/`updated_at`/`labels`/`spec_id`) present.
- **[N/A] MP-4 language neutrality** — single-programming-language SPEC (Go internal tooling + one GitHub Actions workflow). No multi-language tooling surface. Auto-passes per the MP-4 N/A clause.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — five SPEC-IDs referenced in the body; all five exist under `.moai/specs/`. Statuses: `SPEC-KANBAN-TODO-CLI-001` in-progress, `SPEC-UPDATE-DOC-DRIFT-001` draft, `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001` implemented, `SPEC-V3R6-SKILL-GEARS-ALIGN-001` implemented, **`SPEC-LSPMCP-001` superseded**. The terminal-status reference is reconciled in-text: §1.3 names it precisely *because* it is terminal ("SPEC-LSPMCP-001 은 terminal status 라 `applyEraDemotion` 이 advisory 로 낮춘다") — it is measurement provenance, not a live dependency, and its terminal state is the load-bearing fact of the sentence. No BLOCKING finding.
- **[N/A] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md plan.md` → `0`, `0`. Auto-PASS per D8-4.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-SPECLINT-GITBLIND-001/` → rc=1, zero matches. `research.md` absent (Tier S, expected).

---

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | 0.75 — minor ambiguity a reasonable engineer resolves consistently | Problem statement is measurement-attributed and unusually precise; most line citations are exact. Deductions: §2.1's "두 모양" enumeration is wrong (D1); §1.2 cites `lint.go:1310-1313` for a block actually at `:1305-1308` (D5a); REQ-001 label mismatch (D5b) |
| Completeness | 0.75 | 0.75 — one non-critical area sparse; frontmatter complete | All sections present. Out of Scope has three `### Out of Scope — <topic>` H3 sub-headings each with specific `-` bullets — well above the lint convention's floor. Deductions: §4 omits two material risks (D2 output volume, D6 DetectDrift critical-path change, the latter recorded in `plan.md` §B but never promoted to the SPEC's own residual-risk list) |
| Testability | 0.75 | 0.75 — measurable, but not every AC is a binary discriminator | No weasel words anywhere; every AC names a command and an observable. Deductions: AC-SLGB-003 and AC-SLGB-006 cannot fail on today's tree (D8), and AC-003's fixture has an unstated status trap that makes it vacuous for a second, independent reason (D7). AC-002's fixture shape is under-determined (D1) |
| Traceability | 0.75 | 0.75 — one REQ effectively uncovered | REQ→AC mapping: 001→AC-001, 003→AC-002, 004→AC-003, 005→AC-004, 006→AC-005, 007→AC-006, 008→AC-007/008. **REQ-SLGB-002's discriminating observable — the unresolved ref name in the message — is asserted by no AC** (D4). No orphaned ACs; every AC traces to a real REQ |

Aggregate: (0.75 + 0.75 + 0.75 + 0.75) / 4 = **0.75**. Tier S PASS threshold is 0.75 — the SPEC clears it by exactly zero margin.

---

## Defects Found

**D1.** `spec.md`:§2.1 (L120-126) and `plan.md`:§F M1.1 (L60-64) — **The claimed error-shape split is wrong: `getGitImpliedStatus` returns THREE error shapes, not two.** Verified at `internal/spec/drift.go`:
  1. `:311` `git log failed: %s` — ref unresolvable / git error
  2. `:316` `no git history found for %s` — git ran, zero output
  3. `:365` `no classifiable commit within window of %d for %s` — **git ran, matched commits existed, but every one was a skip-pattern / unclassifiable / word-boundary-rejected candidate**

  Shape 3 is named in neither document. Two consequences. (a) The planned branch condition — "ref resolution failure OR (zero-match AND shallow)" — leaves shape 3 **silent even in a shallow repo**, which is precisely the blindness this SPEC exists to end: a truncated window that retains only a cosmetic commit for a SPEC returns shape 3, and M1 as specified says nothing. REQ-SLGB-003's own prose ("해당 SPEC 의 이력을 찾지 못함") reads as covering shape 3, so the requirement and its stated implementation basis disagree. (b) AC-SLGB-002's fixture is under-determined: "얕은 창 안에 매칭 커밋이 없는 SPEC" does not pin whether the window yields shape 2 (`git log --grep` empty) or shape 3 (matches present, all skipped), and the AC passes under one and fails under the other. Note this is not hypothetical — `classification-18.md` class C-2 ("cosmetic docs/chore 커밋이 최신 슬롯을 차지", 6 cases) is the same walker regime that produces shape 3. — Severity: **critical** — Class: **blocking** — Required fix: enumerate all three shapes in `spec.md` §2.1 and `plan.md` §F M1.1; state explicitly whether shape 3 under `is-shallow-repository == true` emits `StatusGitUnreachable` (it should, by REQ-SLGB-003's own wording); pin AC-SLGB-002's fixture to one named shape and add a sibling AC for the other.

**D2.** `spec.md`:§2 REQ-SLGB-001, §4; `plan.md`:§G — **Output-volume amplification is unexamined and contradicts the plan's own stated anti-pattern.** The condition M1 detects (`main` unresolvable; repository shallow) is **repository-level**, but `StatusGitConsistencyRule.Check` runs **per SPEC** (`internal/spec/lint.go:1287`). In the exact CI state this SPEC was written to fix, every non-terminal SPEC in the corpus takes the same branch, so a single repo-level fact renders as one Info finding per SPEC — on this corpus, hundreds of identical lines. `plan.md` §G names "완전한 저장소에서 Info 를 남발하는 것" as an anti-pattern with the rationale "출력이 소음이 되고, 진짜 눈멂 신호가 그 안에 묻힌다" — that rationale applies with full force here, yet the only guard (AC-SLGB-003) covers the *non-shallow, main-resolved* direction and bounds nothing in the blind case. No AC bounds finding count. — Severity: **major** — Class: **blocking** — Required fix: make the decision explicit — either (a) accept per-SPEC emission and record the volume consequence in §4 as an accepted cost, or (b) emit the repo-level Info once (a directory-level finding, for which `Lint()` already has the `dirFindings` channel at `lint.go:177`). Add an AC that bounds the count either way.

**D3.** `plan.md`:§F opening (L52-54) — **The order claim is only half load-bearing; M1→M3 is ceremony, and the SPEC's own measurement refutes it.** M1→M2 IS a genuine dependency: §F M2.2 requires M2's "해소 불가" to flow into the Info path M1 creates, so M2 alone would substitute one silent path for another. But M1→M3 is not. `spec.md` §1.1 run (3) — `--unshallow` + `git fetch origin main:main`, executed against the **unmodified binary** — produced 18 `StatusGitConsistency` findings. M3 in CI reproduces exactly that state, so **M3 alone closes AC-SLGB-008**, with no M1 and no M2. The stated rationale ("M2/M3 를 먼저 하면 … 다시 침묵 속에서 판단해야 한다") is false for M3: counting `StatusGitConsistency` findings in the run log is the non-silent judgment, and it needs nothing from M1. The chosen order is harmless; the assertion that it is fixed *by dependency* is not accurate. — Severity: **minor** — Class: **optional** — Required fix: restate §F as "M1 → M2 is a dependency; M3 is independent and ordered last by preference" — or move M3 first, since it alone restores CI sight at the cost of one workflow edit.

**D4.** `spec.md`:§2 REQ-SLGB-002, §3 — **REQ-SLGB-002 is untraced: no AC observes its distinguishing content.** The requirement's whole payload is that the finding carries "해소 실패한 ref 이름". AC-SLGB-001 asserts only that a `StatusGitUnreachable` finding naming the SPEC appears; AC-SLGB-004 asserts exit code and severity. `plan.md` §F M1.4 implements it, but the implementation can be omitted entirely and all eight ACs still pass. — Severity: **major** — Class: **blocking** — Required fix: extend AC-SLGB-001 (or add an AC) asserting the finding's message contains the unresolved ref name; for the shallow branch, assert it states the shallow fact.

**D5a.** `spec.md`:§1.2 (L64-67) — **Stale line citation.** The text cites `internal/spec/lint.go:1310-1313` for the `err != nil → return nil` silent skip. The actual block is `:1305-1308` (`gitStatus, err := getGitImpliedStatus(fm.ID)` is at `:1304`; the comment "If git history is unavailable, skip this check" at `:1306`; `return nil` at `:1307`). Every other citation in both documents was checked and is exact — `lint.go:1287`, `lint.go:1316`, `lint_ownership.go:429`, `drift.go:68`, `gitquery_cache.go:89-117`, `lint.go:285-300`, `lint_test.go:574-580`, the six `ci.yml` line numbers — so this is an isolated slip, not a pattern. — Severity: **minor** — Class: **optional** — Required fix: correct to `:1305-1308`. (`spec.md` §1.1 and `plan.md` §H cite `drift.go:299` / `:299-330` for `getGitImpliedStatus`, which declares at `:300` and runs to `:366` — off by one at the start and truncated at the end; correct to `:300-366`.)

**D5b.** `spec.md`:§2 REQ-SLGB-001 — **GEARS label mismatch.** Labelled `(Ubiquitous)` but the sentence carries the condition "관측하지 못한 경우", making it state-driven. Pattern-valid either way (MP-2 passes), label wrong. — Severity: **minor** — Class: **optional** — Required fix: relabel `(state-driven)`, or lift the condition out of the sentence.

**D6.** `spec.md`:§4 — **M2's effect on the SessionStart critical path is absent from the SPEC's residual risks.** `plan.md` §B states the two-consumer fact plainly and does not bury it (challenge #3: **the plan states it correctly**). But the consequence never reaches `spec.md` §4, and no AC covers the `DetectDrift` surface. The change is real: on a checkout with `origin/main` but no local `main`, `DetectDrift` today walks the non-existent `"master"` (log fails → `commits` nil → every SPEC skipped → **zero drift records**, silently, per `drift.go:206-210`); after M2 it walks `origin/main` and drift records appear. `DetectDrift` runs synchronously inside the SessionStart hook (`drift.go` `@MX:REASON`), so this is a behaviour change on the session-launch path introduced as a side effect of a lint fix. Mitigating: `internal/spec/drift_entrypoints_test.go` exercises the production `cachedMainBranch` against real fixtures built with `git init -b main` (`drift_chore_skip_test.go:42`, `closer_test.go:876`, `drift_era_terminal_test.go:72`), so the chain's first candidate resolves and `go test ./internal/spec/...` covers it incidentally. — Severity: **minor** — Class: **optional** — Required fix: add a §4 residual-risk bullet naming the `DetectDrift` behaviour change and the SessionStart path; consider an AC asserting `DetectDrift` yields the same result under a local-`main` and an `origin/main`-only fixture.

**D7.** `spec.md`:§3 AC-SLGB-001/002/003 — **Unstated fixture trap makes AC-SLGB-003 vacuous for a second, independent reason.** `StatusGitConsistencyRule.Check` returns `nil` at `internal/spec/lint.go:1300` for any SPEC whose status is in `terminalStatusEnum` = `{superseded, archived, rejected, completed}` (`:1271-1276`) — **before** `getGitImpliedStatus` is ever called. A fixture SPEC authored with `status: completed` therefore produces zero findings regardless of git state. AC-SLGB-003 asserts exactly "zero findings", so a terminal-status fixture satisfies it without the rule running at all. No AC states the fixture's status constraint. — Severity: **major** — Class: **blocking** — Required fix: state in AC-001/002/003 that the fixture SPEC carries a non-terminal status (`draft` / `in-progress` / `implemented`); for AC-003, additionally require a positive control proving the rule executed on that fixture.

**D8.** `spec.md`:§3 AC-SLGB-003, AC-SLGB-006; `plan.md`:§E — **Two ACs cannot show RED, and the plan's own vacuous-green guard does not reach them.** `grep -rn StatusGitUnreachable internal .github` returns zero matches, so on today's tree AC-003 ("`StatusGitUnreachable` findings = 0") passes trivially, and AC-006 (cache preservation) passes because `cachedMainBranch` already memoizes via `mainBranchSet` (`gitquery_cache.go:96-107`). Both are legitimate guards — one against over-emission, one against regression — but `plan.md` §E commits to "M1 의 새 테스트는 먼저 RED 를 보여야 한다" and AC-003 is one of M1's four closing ACs. A test that cannot fail before the change is the exact failure class §E cites the repository's own history for. Answering the dispatch's challenge #1 directly: **AC-SLGB-008 is indeed the only CI-only AC** (correctly flagged in `spec.md` §3 and `progress.md` §E.1), but it is **not** the only AC that fails to discriminate — AC-003 and AC-006 pass whether or not the change lands. — Severity: **major** — Class: **blocking** — Required fix: exempt AC-003/006 from the RED-first clause explicitly and pair each with a mutation check instead (for AC-003: after M1 lands, force the ref-unresolvable state and confirm the *same* fixture flips to ≥1 finding — proving the assertion is live; for AC-006: temporarily disable `mainBranchSet` and confirm the test fails).

**D9.** `.github/workflows/spec-lint.yml`:4-14; `spec.md`:§3 AC-SLGB-008 — **The job's path filter is `.moai/specs/**` on both triggers, so M1/M2 code changes never re-run it.** After this SPEC lands, a future regression in `internal/spec/lint.go` or `gitquery_cache.go` will not re-trigger `SPEC Lint`. AC-SLGB-008's evidence is obtainable only because this SPEC's own artifacts live under `.moai/specs/` and therefore match the filter — which holds only if the SPEC documents land in the same push as the workflow fix. Since `plan.md` §F orders M3 last, an implementer splitting the milestones across pushes could land M3 in a commit touching no SPEC file, and the job would not fire. — Severity: **minor** — Class: **optional** — Required fix: note in `plan.md` §F M3 that M1-M3 plus the SPEC artifacts must land in one push for AC-008's evidence to exist; separately consider whether `internal/spec/**` and `.github/workflows/spec-lint.yml` belong in the path filter (out of this card's scope, but a follow-up candidate).

---

## Answers to the five challenges

1. **AC verifiability.** AC-SLGB-008 is the only CI-log-only AC and is correctly flagged as such. But "the rest genuinely fail when the behaviour is absent" is **false**: AC-SLGB-003 and AC-SLGB-006 both pass on today's unmodified tree (D8), and AC-003 has a second, independent vacuity path through the terminal-status early return (D7). AC-001/002/004/005/007 do genuinely fail today — `StatusGitUnreachable` does not exist (grep rc=1), and `cachedMainBranch` returns literal `"master"` for AC-005's fixtures 2 and 4.

2. **The M1 discriminant.** The claimed error-shape split **does not exist as described** — there are three shapes, not two (D1). The shallow predicate's placement on the per-run cache is, however, **correct**: `is-shallow-repository` is a repository-level constant, and the cache's lifetime is exactly one `Lint()` run (created `lint.go:174`, discarded by `defer stopGitQueryCache()` at `:175`), so the value cannot go stale within its scope. One caveat the plan does not state: the cache is a **package-level global** (`gitquery_cache.go:20-23`), so concurrent `Lint()` calls clobber it — pre-existing, not introduced by M1, but any new test touching `startGitQueryCache` must not run under `t.Parallel()`.

3. **M2 blast radius.** The second consumer is real and exactly where the author says (`drift.go:68`, `realDriftDeps`), and `plan.md` §B **states the consequence plainly rather than burying it** ("M2 는 회귀 표면이 lint 하나가 아니다"). The gap is downstream: the consequence never reaches `spec.md` §4, and no AC covers the `DetectDrift` surface (D6).

4. **The order claim.** **Half load-bearing, half ceremony.** M1→M2 is a genuine dependency (M2's unresolvable state must land in M1's Info path or M2 creates a new silent failure). M1→M3 is ceremony, and the SPEC's own run (3) refutes the stated rationale: M3 alone restores the 18 findings and closes AC-SLGB-008 without M1 or M2 (D3).

5. **Residual risks.** Both recorded risks are **confirmed against the code**. (a) A shallow repo whose truncated window still yields a classifiable commit returns `(status, nil)` from `drift.go:358` and surfaces as an ordinary `StatusGitConsistency` warning — never as Unreachable. (b) `OwnershipTransitionInvalid` is emitted at `lint_ownership.go:425-440` with `Severity: SeverityWarning` and **no `Advisory` field**, and `lint.go:61` escalates exactly that shape under `--strict`; today only `applyEraDemotion` (via `SPEC-LSPMCP-001`'s terminal status) holds it back. Material risks the plan **omits**: the output-volume amplification (D2), the `DetectDrift`/SessionStart behaviour change (D6), the shape-3 silent path that survives M1 (D1), and the path-filter durability gap (D9).

---

## Recommendation

The measurement foundation is the strongest part of this SPEC and survived adversarial re-derivation intact — every cited premise except one line range was verified line-exact against the tree, and the A/B attribution of 1 finding to history depth and 18 to the `main` ref is sound. The verdict is **PASS-WITH-DEBT** rather than PASS because five blocking-class defects would let the SPEC ship with a requirement unimplemented (D4), two acceptance criteria that cannot fail (D7, D8), a fourth silent path left open by the very milestone that exists to close silent paths (D1), and an unbounded output-volume consequence the plan's own anti-pattern section argues against (D2). All five are sentence-level edits to `spec.md` / `plan.md`; none require structural rework.

Resolve before run-phase entry, in this order:

1. **D1** — enumerate all three `getGitImpliedStatus` error shapes; decide shape 3's behaviour under `is-shallow-repository == true`; pin AC-SLGB-002's fixture to a named shape.
2. **D2** — decide per-SPEC vs repo-level Info emission explicitly; add a count-bounding AC.
3. **D4** — add the ref-name assertion so REQ-SLGB-002 is observable.
4. **D7** — state the non-terminal-status fixture constraint in AC-001/002/003.
5. **D8** — exempt AC-003/006 from the RED-first clause and pair each with a mutation check.

Then optionally: D3 (restate the order claim), D5a/D5b (citation and label corrections), D6 (§4 residual-risk bullet), D9 (single-push note in §F M3).

---

## Gaps (not observed in this audit)

- No test was executed. `go build ./internal/spec/...` and `go test ./internal/spec/...` were **not** run — the plan's §C.3/§C.4 baselines remain unmeasured, and this audit asserts nothing about them.
- GitHub Actions runtime behaviour (whether `actions/checkout` at `fetch-depth: 0` populates `refs/remotes/origin/*`) was not verified — matching `spec.md` §2.2's own honest statement that it is unverifiable locally. The SPEC's decision to add an explicit fetch rather than rely on it is sound under that uncertainty.
- The 18 classified findings in `classification-18.md` were not re-derived; the file's existence and its A/B/C summary table were read, its per-row classification was not re-checked.

## Residual risk

The score sits exactly at the Tier S threshold with no margin. If any one of the four dimensions is judged half a band lower than recorded here — and Testability is the most exposed, with two of eight ACs unable to discriminate — the aggregate falls below 0.75 and the verdict becomes FAIL. Resolving D7 and D8 is what moves this SPEC off the threshold rather than merely past it.

---

# Addendum — two lead-supplied points, verified

Added after the iter-1 verdict on two points raised by the lead. Both were re-derived
from source rather than accepted. **Verdict unchanged: PASS-WITH-DEBT, 0.75.** One escalation
was considered and rejected on evidence (see D10).

## Point 1 — `cachedMainBranch` directory binding and AC-SLGB-005 satisfiability

**The mechanism is confirmed; the conclusion is (a), not (b).**

Confirmed: `internal/spec/gitquery_cache.go:102` and `:113` call
`exec.Command("git", "rev-parse", "--verify", "main").Output()` with no `.Dir`, so resolution
follows the **test process working directory**. `grep -rn cachedMainBranch internal/spec` over
`*_test.go` returns one line — `drift_entrypoints_test.go:13`, a comment — so no test calls it
directly, as the lead read.

But the package already solves this, and has for some time. `setupDriftCorpusFixture`
(`internal/spec/drift_characterization_test.go:98-106`) ends with `chdirForTest(t, root)`, and
`chdirForTest` (`:55-64`) saves the original cwd, `os.Chdir`s into the fixture, and restores via
`t.Cleanup`. Every existing fixture-based drift test resolves `cachedMainBranch` against its
fixture precisely this way. AC-SLGB-005's four fixtures are therefore satisfiable with an
**existing helper and an established precedent** — no `os.Chdir` invention, and no directory
parameter threaded into the helper.

So: **(a)**. `plan.md` §B's claim that ACs verify against real `t.TempDir()` + `git init` fixtures
holds, and **M2's declared diff is not understated on this axis** — the signature change the lead
posited as the alternative is not required.

The parallel-safety concern is also already addressed, though only by convention — see D11.

**D10.** `spec.md`:§3 AC-SLGB-005, AC-SLGB-006 — **Neither AC declares its per-run-cache
precondition, and the two want opposite states.** AC-005 needs `gitQueryCacheV == nil` (the default
in a unit test) so each of its four fixtures takes the direct path at `gitquery_cache.go:112-116`;
AC-006 explicitly needs the cache **active and preserved**. A test for AC-006 that calls
`startGitQueryCache()` without a paired `stopGitQueryCache()` leaks the memoized branch through the
package-level global (`gitquery_cache.go:20-23`) into whatever runs next, including AC-005's
fixtures. — Severity: **minor** — Class: **optional** — Required fix: state each AC's cache
precondition; require `defer stopGitQueryCache()` in AC-006's test.

*Escalation considered and rejected.* I initially scored this blocking on the theory that fixtures
2-4 would silently read fixture 1's memoized `"main"`. That is wrong: those fixtures assert
`origin/main`, `master`, and unresolvable respectively, so a leaked cache makes them **fail loudly**,
not pass vacuously. It is a cross-test pollution hazard producing a red test, not a third vacuity
vector. Testability therefore stays at 0.75 and the aggregate stays at 0.75 — had this been a
genuine vacuity vector, Testability would have dropped a band and the verdict would have become FAIL.

**D11.** `internal/spec/drift_characterization_test.go:53-64` — **The no-parallel rule protecting
the chdir fixtures is a comment, not a mechanism.** `chdirForTest` uses `os.Chdir` +
`t.Cleanup`, and the rule lives in the comment at `:53` ("Tests using this helper MUST NOT call
`t.Parallel()`: os.Chdir mutates …"). The convention holds today — the three `t.Parallel()` calls in
`internal/spec` (`archive_git_test.go:178`, `:259`, `:284`) are all non-chdir cases
(`NonGitDirIsEmpty`, `FallsBackOutsideGit`, `RefusesToClobber`). But `go.mod` declares **go 1.26.4**,
so `t.Chdir` is available and **panics** when the calling test is parallel — the same rule, enforced
by the runtime instead of by a reader. AC-SLGB-005 adds four more chdir-dependent fixtures, which
multiplies reliance on the comment. — Severity: **minor** — Class: **optional** — Required fix:
switch `chdirForTest` to `t.Chdir`, or state in `plan.md` §E that AC-005's tests must not
parallelize. Pre-existing, not introduced by this SPEC.

## Point 2 — Info findings render in the table but not in the summary

**Confirmed exactly as read, and AC-SLGB-001 survives.**

`printTable` (`internal/cli/spec_lint.go:124-132`) iterates `report.Findings` with **no severity
filter**, printing `strings.ToUpper(string(f.Severity))` — so a `StatusGitUnreachable` row appears as
`INFO`. The summary immediately below (`:136-145`) counts only `SeverityError` and
`SeverityWarning` and prints `%d error(s), %d warning(s)`, so an Info finding is invisible there.

AC-SLGB-001 asserts the finding appears "출력에" (in the output) — a table row satisfies that, so the
AC survives. AC-SLGB-004's "severity 필드는 info" is observable on both surfaces: the table renders
`INFO`, and `--json` emits `"severity":"info"` (`internal/spec/lint.go:34` `json:"severity"`).
The AC does not name which surface; harmless, but see the fix below.

**On whether M1's stated purpose survives — it does, by a mechanism neither document names.**
`printTable` short-circuits at `:114-117`: when `len(report.Findings) == 0` it prints
`✓ No findings — all SPEC documents are valid`. That line is today's blind-state output — the
false all-clear this SPEC exists to remove. After M1, the blind state yields N Info findings, the
short-circuit no longer fires, and **the false all-clear disappears**. That is the real visibility
win, and it is stronger than a changed count would be.

The residual is narrower than the lead's framing but real: a reader scanning only the count line
sees `0 error(s), 0 warning(s)`, unchanged. Combined with D2, the blind-state output becomes
hundreds of Info rows sitting under a zero summary — which is precisely the shape D2 asks the
authors to decide on deliberately.

**D12.** `spec.md`:§2 REQ-SLGB-001/005, §3 AC-SLGB-001/004 — **The Info finding is invisible in the
summary line, and no AC observes the surface that actually changes.** The load-bearing behavioural
change in the blind state is the *disappearance* of `✓ No findings — all SPEC documents are valid`
(`internal/cli/spec_lint.go:114-117`), not the appearance of a count. No AC asserts it, so an
implementation that emitted the Info only on the `--json` path — leaving the table's false all-clear
intact — would pass every AC as written. — Severity: **minor** — Class: **optional** —
Required fix: name the output surface in AC-SLGB-001 (table row) and AC-SLGB-004 (`--json`
`severity` field); add an assertion that the `✓ No findings` line is absent in the blind-state
fixture. Note the `--json` and `--sarif` paths were **not** examined beyond the `Finding` struct's
JSON tags — see Gaps.

## Revised defect ledger

Blocking (unchanged, 5): D1, D2, D4, D7, D8.
Optional (was 4, now 7): D3, D5a, D5b, D6, D9, **D10**, **D11**, **D12**.

## Gaps added by this addendum

- The `--json` and `--sarif` output paths were not exercised. The `Finding` struct's JSON tags were
  read (`lint.go:33-45`); no serializer output was produced, and `sarif.go` was not located by grep.
  Claims above about `--json` rest on the struct tags alone.
- No test was run to confirm the cache-pollution hazard in D10; it is derived from reading
  `gitquery_cache.go:20-23` and `:39-52`, not observed.
