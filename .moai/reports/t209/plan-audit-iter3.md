# SPEC Review Report: SPEC-WORKTREE-REAPER-001

Iteration: 3/3 (final)
Verdict: **PASS**
Overall Score: **0.875** (Tier L threshold 0.85)

Reasoning context ignored per M1 Context Isolation. The dispatch's
characterisations — "the EC-9 result is corrected", "the fork is settled by
measurement, not preference", "the t208 path pre-exists M1" — were treated as
hypotheses to test, not premises. Every judgment below is anchored to an artifact
line or a command run in
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t209` (branch
`WT-worktree-reaper`, HEAD `6a9b7c66a`) on 2026-08-24.

**No fixture was built.** The measurement-hygiene instruction is honoured
literally: this session sits inside a live worktree, the worktree-isolation guard
refuses `git -C` into sibling trees, and iteration 2's own EC-9 error was produced
by a fixture built in exactly this position. Where a claim could only be settled
by a fixture, it is recorded as a **gap**, not measured in a contaminated tree.
Two such gaps are named in § Gaps.

**Score trajectory: 0.55 → 0.84 → 0.875.** No regression; no STOP signal.

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `grep -o 'REQ-WR-[0-9]*' spec.md | sort -u` → `REQ-WR-001 … REQ-WR-024`, 24 unique, zero-padded, no gaps, no duplicates; `grep -c '^- \*\*REQ-WR-' spec.md` → `24`. Declared count matches (`spec.md` §D "24 requirements"; `progress.md` §E.1 "**24**").
- **[PASS] MP-2 GEARS format compliance — judged against the `REQ-WR-XXX` requirement layer in `spec.md` §D only.** The Given/When/Then entries in `acceptance.md` are the verification layer and are graded under Group 4, not here. `grep -o '(Ubiquitous)\|(Event-driven)\|(State-driven)\|(Unwanted)\|(Where — capability gate)' spec.md | sort | uniq -c` → 9 Ubiquitous + 8 Event-driven + 2 State-driven + 4 Unwanted + 1 Where = **24 labels for 24 requirements**. `grep -c 'Event-detected' spec.md` → `0` (iteration-1 D17 stays fixed). The four v0.3.0 additions/rewrites (REQ-WR-021, 023, 024, and the amended 019) each carry a valid pattern and a `shall`. Mechanical confirmation: `moai spec lint .moai/specs/SPEC-WORKTREE-REAPER-001/spec.md` → exit 0, `✓ No findings — all SPEC documents are valid`.
- **[PASS] MP-3 YAML frontmatter validity** — spec.md:L1-15 carries all 12 canonical fields (`id`, `title`, `version: "0.3.0"`, `status: draft`, `created`/`updated` ISO, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags`) plus optional `tier: L`. No rejected snake_case alias. `moai spec lint` exit 0. Version bump to `0.3.0` closes iteration-2 N6.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `module: internal/cli`); no template-bound or 16-language content.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — `grep -Eoh 'SPEC-([A-Z][A-Z0-9]+-)+[0-9]+' *.md | sort -u` → `SPEC-SESSION-WORKTREE-001`, `SPEC-WORKTREE-REAPER-001`. `grep '^status:' .moai/specs/SPEC-SESSION-WORKTREE-001/spec.md` → `completed` — not retired/superseded/archived. No BLOCKING.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -rn 'syscall' *.md` → 2 matches (`design.md:311`, `plan.md:147`), both carrying an explicit cross-platform clause in the same sentence: `plan.md:147` "wrapping the existing syscall, **with the per-platform mapping in `design.md` §B.5**"; `design.md` §B.5 carries the five-row unix/windows probe table and the "windows can never assert dead, so never widens removal" clause; `spec.md` §E states the Windows constraint and AC-WR-022 runs `GOOS=windows go vet`. Explicit cross-platform exemption present ⇒ PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn 'NEEDS CLARIFICATION' <spec dir>` → exactly one line, `progress.md:38`, a **negation** ("no `[NEEDS CLARIFICATION]` markers remain"), not a marker. `plan.md` and `research.md` both exist and carry zero markers. **Judged, not merely counted:** `design.md` §A.7's open fork is the substantive question this criterion exists to catch, and it is examined on its merits in § 3 below. It does not fail MP-7 — the fork is named, gated by a [HARD] M1 precondition, and cannot silently ship — but its decision rule is defective, which is finding F1.

No must-pass criterion fails.

---

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | 0.75-1.0 | Every iteration-2 clarity defect is repaired and I verified each: `spec.md` §E's inverted constraint now reads "git does NOT backstop the dirty guard for ignored content" with the measured commands (L253-262); REQ-WR-021 (L188-211) no longer asserts a condition that does not exist and states its pre-detection set as explicitly **non-exhaustive**; `spec.md` §D's preamble contradiction is gone ("018-024 are amendment additions", L129-131). Deductions: `design.md` §A.7's decision rule maps half its outcome space to two policies while the same section claims "a procedure, not a judgement call" (F1); four textual residues of the last edit (F4-F7); the §A HISTORY entry asserts the `--ignored` probe "survives" where §A.7 declines to assert it (F8). |
| Completeness | 0.85 | 0.75-1.0 | All required sections present; `grep -c '^### Out of Scope — ' spec.md` → **8** H3 sub-headings, each with specific bullets (submodules added at v0.3.0 with a measured reason, closing N2). `research.md` corrects itself rather than being patched: §D.4a rewritten to "git does NOT check ignored content at removal time", §F **strikes through** its own wrong `--merged-only` exclusion and states why it was wrong. The third blind consumer is named in all four artifacts. Deductions: F1 (the fork's second branch has no selecting rule, and the second datum AC-WR-025 collects feeds nothing); F2 (REQ-WR-021's defining limb has no criterion); F4-F7. |
| Testability | 0.88 | 0.75-1.0 | **The falsifiability property survives the growth, and I re-verified it end to end: 29 distinct test names across all 26 criteria, 0 discrepancies** (table in §1). The two new-name criteria were run in their exact criterion form and returned `0` as recorded. AC-WR-024's assertion is rebuilt from string inequality to a positive token match on `locked` (N4 closed). AC-WR-026 uses the anchored `^…$` form (no D12 regression). Deductions: REQ-WR-021's fail-open limb is untested (F2); AC-WR-025 is a procedure-completion criterion with no discriminator — honestly declared, but it is the sole criterion for REQ-WR-024, so one requirement of 24 carries no behavioural test (an unavoidable consequence of the open fork, not a concealment). |
| Traceability | 0.92 | 0.75-1.0 | `grep -o 'REQ-WR-[0-9]*' acceptance.md | sort -u` → **all 24**. Every criterion carries a `Covers:` line; AC-WR-020 still honestly declares it covers a craft constraint rather than a REQ. The DoD label is corrected to `REQ-WR-001…024` (N5(a) closed) and AC-WR-024 has been moved out of §C into §B beside the other `prMergeCleanup` criteria (N5(c) closed). Deductions: REQ-WR-023 enumerates five cause tokens but is cross-cited at one criterion asserting one token (F3); REQ-WR-021's two covering criteria both exercise pre-detection only (F2). |

Aggregate: (0.85 + 0.85 + 0.88 + 0.92) / 4 = **0.875** ≥ 0.85.

---

## 1. Criterion re-verification (D1 non-regression) — 29 of 29 names, 0 discrepancies

The task asks how many criteria I ran. **All of them, by the method appropriate to each baseline.** 26 criteria cite 29 distinct test names. Two are new at v0.3.0 (`TestPRMergeCleanup_RefusalClassNamesCause`, replacing v0.2.0's `…_RefusalClassPreDetected`; `TestCleanMergedOnly_LockAnchoredWorktreeKept`); the rest are unchanged.

**Eleven `1`-baseline names executed as real test runs** (three package invocations; the alternation `-run` produces `--- PASS:` lines byte-identical to the per-name criterion form):

```
go test ./internal/cli/ -run '^(TestPRMergeCleanup_GhPresentMergedRemoves|…|TestParseWorktreeList_BranchExtraction)$' -v -count=1 -timeout 600s
--- PASS: TestPRMergeCleanup_AnchoredSessionSkipsRemoval        --- PASS: TestPRMergeCleanup_ToggleOffNoOp
--- PASS: TestPRMergeCleanup_NilCfgNoOp                          --- PASS: TestPRMergeCleanup_GhPresentMergedRemoves
--- PASS: TestPRMergeCleanup_GhPresentSeesSquashMerge            --- PASS: TestPRMergeCleanup_GhAbsentBranchMergedRemoves
--- PASS: TestPRMergeCleanup_GhAbsentEmitsBlindnessNoticeOnce    --- PASS: TestPRMergeCleanup_WorktreeListErrorFailOpen
--- PASS: TestParseWorktreeList_BranchExtraction                 ok  internal/cli  1.367s

go test ./internal/cli/worktree/ -run '^(TestCleanStale_KeepsAnchoredWorktree|TestCleanStale_PreviewsByDefault|TestCleanStale_RemovesWithYes)$' -v -count=1
--- PASS: TestCleanStale_KeepsAnchoredWorktree   --- PASS: TestCleanStale_PreviewsByDefault
--- PASS: TestCleanStale_RemovesWithYes          ok  internal/cli/worktree  22.172s
```

All eleven recorded `1` → observed `1`. `TestParseWorktreeList_BranchExtraction` is `acceptance.md` §0's positive control and it passes, so the criterion form itself still discriminates.

**Eighteen `0`-baseline names proved absent mechanically** — a name with no `func` definition cannot produce a `--- PASS:` line, so source absence is a sound and complete proof of `0`:

```
grep -rn "func TestPRMergeCleanup_GhNoAnswerConsultsGitFallback(\|…\|func TestLockAnchor_\|func TestAnchorDecision_\|…\|func TestCleanMergedOnly_LockAnchoredWorktreeKept(" internal/
grep-exit=1          # zero matches for all 18
```

**The two new/changed criteria additionally run in their exact criterion form:**

```
go test ./internal/cli/ -run '^TestPRMergeCleanup_RefusalClassNamesCause$' -v -count=1 -timeout 600s 2>&1 \
  | grep -c '^--- PASS: TestPRMergeCleanup_RefusalClassNamesCause'          → 0   (AC-WR-024, recorded 0) ✓
go test ./internal/cli/worktree/ -run '^TestCleanMergedOnly_LockAnchoredWorktreeKept$' -v -count=1 2>&1 \
  | grep -c '^--- PASS: TestCleanMergedOnly_LockAnchoredWorktreeKept'       → 0   (AC-WR-026, recorded 0) ✓
```

**Grep-token baselines reproduce**: `mergeStateUndetermined` → **0** (admissible); `lockReason` → **22** (correctly declared inadmissible); `func Test.*NoAskUserQuestion` in `internal/cli/` → **31**, matching AC-WR-020's figure exactly.

**Mandatory sample re-verified against the lock path, not the registry path.** `git worktree list --porcelain` for `.claude/worktrees/t207`:

```
branch refs/heads/WT-web-live-todo
locked claude session t207 (pid 36912 start Sun Aug 23 07:26:09 2026)
git branch --no-merged origin/main | grep -c 'WT-web-live-todo'   → 1
```

AC-WR-014 exercises this fixture **with the session registry empty**, so the lock source alone must answer — which is the only construction that is not 4-of-5 blind. AC-WR-023(a) exercises the real tree non-destructively. The sample is judged NOT-disposable and is bound to the lock path. Correct.

---

## 2. Regression Check

### 2a. Iteration-2 findings N1-N5 (plus N6, N7)

| # | Claim | Verdict | Evidence |
|---|---|---|---|
| **N1** EC-9 does not reproduce | **RESOLVED — fully, across all six downstream sites** | `ec9-measurement.md` is now **v2** and states in its own header that §Q1 *replaces* a wrong v1 result and that v1 text is deliberately not preserved. Every site is corrected, and I checked each rather than accepting the disposition: `spec.md` §E now reads "git does NOT backstop the dirty guard for ignored content" with the measured commands; `design.md` §A.6 is retitled "git does NOT backstop…" and states "There is no third backstop layer"; §A.4 reason 3 now ends "**For gitignored content it does not**"; §B.6a's table records ignored-only as "**removes, exit 0**"; `research.md:152` §D.4a rewritten; EC-9 now reads "**Removed today, and the ignored content is destroyed**"; EC-11 re-derived. The inversion of the probe's *rationale* is carried through to REQ-WR-024. Note the correction goes further than the finding required: the SPEC does not merely restate the measurement, it withdraws the scope the bad measurement had motivated. |
| **N2** refusal class incomplete (submodule) | **RESOLVED** | `spec.md` §G now carries `### Out of Scope — submodule-bearing worktrees` with the measured reason (`.gitmodules` absent — I confirmed: `ls .gitmodules` → No such file or directory), and EC-12 records the case. `design.md` §B.6a re-derives the class from what actually refuses rather than patching the enumeration. The disposition offered by iteration 2 ("either cover it or record it in §G with the measured reason") was taken in its second form, correctly. |
| **N3** third consumer unnamed | **RESOLVED** | Verified against source, not against the claim: `grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` → `95:` and `163:`; `sed -n '88,100p'` shows the verbatim in-code comment "`--merged-only` has no dirty guard of its own, so this is the only protection between the sweep and a live lane's tree" immediately above line 95. REQ-WR-019 now names all three sites, AC-WR-026 exercises the third, `research.md` §F strikes through its own wrong exclusion. |
| **N4** AC-WR-024 assertion weaker than its requirement | **RESOLVED** | AC-WR-024's Then clause now reads "the preserve notice **contains the cause-specific token `locked`** — not merely a string that differs from some other notice", and the closing paragraph states both reasons for the rebuild. The non-existent second fixture (ignored-only as a refusal) is removed. |
| **N5** count and placement residues | **RESOLVED (all three limbs)** | (a) DoD now reads `REQ-WR-001…024` — and coverage is in fact complete, 24/24. (b) `spec.md` §D preamble no longer carries the contradictory parenthetical. (c) AC-WR-024 has moved into §B (M2) beside the other `prMergeCleanup` criteria; `grep -n '^### AC-WR-'` shows the order 016 → 024 → 025 → 026 → §C 017. |
| **N6** version not bumped (optional) | **RESOLVED** | `version: "0.3.0"` with its own HISTORY bullet enumerating the five folded-in consequences. |
| **N7** EC-8 "nothing is lost" | **RESOLVED** | EC-8 now reads "nothing **committed, tracked, or untracked** is lost … **Ignored content is NOT protected** — see EC-9". |

**7 of 7 dispositioned; 7 fully resolved; zero regressed.**

### 2b. Iteration-1 findings D1-D19 — no regression across the growth

Spot-checked every finding whose subject matter the v0.3.0 edit touched, and mechanically re-verified the four that a growth edit is most likely to break:

- **D1 (falsifiability)** — re-verified in full: 29/29 names, 0 discrepancies (§1). The two new criteria adopt the falsifiable form correctly. **No regression.**
- **D8/D9 (counts and coverage)** — `grep -c '^### AC-WR-'` → **26**; header sub-counts 7 + 12 + 4 + 3 = 26 and each range verified against the actual IDs; 24/24 requirements covered. **No regression.**
- **D10 (test names asserted as "existing" that do not exist)** — the two new names are recorded `0`, not `1`; the one existing name reused by AC-WR-015 (`TestCleanStale_KeepsAnchoredWorktree`) runs and passes. **No regression.**
- **D12 (unanchored `-run`)** — AC-WR-026's pattern is `'^TestCleanMergedOnly_LockAnchoredWorktreeKept$'`, anchored. **No regression.**
- **D2, D3, D4, D5, D6, D7, D11, D13, D14, D15, D16, D17, D18, D19** — each re-checked at its artifact site and each still closed; D3 is now closed *completely* rather than partially (N3). `grep -c 'Event-detected'` → 0 (D17); `research.md` present and materially improved (D16); `design.md` §B.3's stored-reason table intact (D19); `design.md` §D's honest "correctness and legibility gain, not a rescue" intact (D6).

**19 of 19 remain closed. Zero regressions across two growth rounds.**

---

## 3. The central open question (design.md §A.7) — judged

### 3a. Is deferring to a measurement legitimate here, or an unresolved decision wearing a procedure's clothes?

**Partly legitimate, and the author's argument is sound for exactly half the outcome space.**

The argument under test is: a `[NEEDS CLARIFICATION]` marker would be wrong because the fork is settled by measurement, not preference. Tested against the rule as written:

- **P1 vs {P2, P3} is genuinely measurement-settled.** The measurement (how many M1-unblocked trees carry ignored content) determines whether "preserve on any ignored content" cancels M1. That is a fact about the repository, not a preference, and no amount of plan-phase deliberation substitutes for it. Deferring it is correct, and marking it `[NEEDS CLARIFICATION]` — a marker whose convention routes to `AskUserQuestion`, i.e. to a *user preference* — would genuinely be the wrong instrument. **The author is right on this half.**
- **P2 vs P3 is not measurement-settled, and the SPEC provides no rule for it.** The table's second row reads: "P1 preserves **more than half** → P1 is too blunt … Choose between **P2** and **P3** below." That is a judgement call, and `design.md`'s own descriptions of the two make it one: P2's cost is "the allowlist is a new thing to maintain and is wrong by default for any project MoAI has not seen"; P3's is "honest, and matches git's own default, but a local `.env` is genuinely unrecoverable". Those are preference statements weighed against each other, not outcomes a `git status --porcelain --ignored` run selects between.

So §A.7's closing claim — "**A run-phase implementer therefore has a procedure, not a judgement call**" — is falsified by §A.7's own table, for the branch the SPEC's own hypothesis says is the likely one. That is finding **F1**.

### 3b. Is the decision rule decisive?

**No.** Every possible measurement outcome does *not* map to exactly one policy. Outcome space:

| Measured | Policy selected |
|---|---|
| ignored content in ≤ half the M1-unblocked set | P1 — **decisive** |
| ignored content in > half | P2 **or** P3 — **not decisive** |

There is a second symptom of the same defect, and it points at the fix: AC-WR-025 instructs the measurer to record two data — how many unblocked trees carry ignored content, **and** "how many carry ignored content **outside** regenerable paths" — but the rule consumes only the first. The second datum is exactly the input that discriminates P2 from P3 (if nothing irreplaceable is out there, P3's cost argument collapses; if something is, P2's does). The rule is one table row short of being decisive, and the missing row's input is already being collected.

### 3c. Does the gate bind?

**Yes, as far as a plan-phase artifact can make it bind — and more than most.** It is stated in four places, three of them [HARD]: `plan.md` §C.1 (a top-level pre-flight section titled "[HARD] C.1 — M1 precondition gate"), `acceptance.md` AC-WR-025 ("[HARD] **This is a gate on M1, not a report**"), the DoD checkbox ("AC-WR-025 run from outside every worktree, and `design.md` §A.7's fork closed, BEFORE M1 lands"), and `plan.md` §G's anti-pattern list. `progress.md` carries it as "[HARD] OPEN DECISION — gated on a measurement, not deferred by omission".

One weakness, recorded rather than inflated: **M1's own step list (`plan.md` §F M1, steps 1-5) does not reference the gate.** A run-phase implementer working from the milestone section alone would not encounter it; they would have to have read §C.1. It is a documented gate, not a mechanical one — which is the ceiling for any plan-phase artifact — but the cheapest possible strengthening is one cross-reference line at the head of §F M1. Recorded as F9 (optional).

---

## 4. The REQ-WR-021 re-derivation — audited

**Is defining a requirement over a deliberately incomplete enumeration sound?**

**Yes, and it is the better of the two available forms.** The requirement is stated over the *observable* — "when a worktree is in a state that non-forced `git worktree remove` refuses, the sweep shall preserve the worktree and emit a cause-naming notice" — with the enumeration demoted to an optimisation ("where that state is already observable from data the sweep has in hand, pre-detect it") and explicitly flagged non-exhaustive. That inverts the failure mode iteration 2 exposed: under the v0.2.0 form, an unlisted member (the submodule) fell *silently outside* a requirement claiming completeness; under this form it falls *inside* the requirement and lands on the fail-open path. `design.md` §B.6a states this reasoning explicitly and records that reverting to the v0.1.0 lock-only scope was considered and rejected for that reason.

The requirement is **not** unfalsifiable. Its predicate — "non-forced `git worktree remove` refuses" — is decidable by running the command, and the obligation it triggers (preserve + name the cause) is observable in the sweep's output. A requirement quantified over an open class is still falsifiable by any single member.

**Does the fail-open path have a criterion? No — and that is finding F2.**

`REQ-WR-021` is cited by exactly two criteria (`grep -n '^\*\*Covers:'`): AC-WR-012 (confirmed-dead lock produces no removal attempt) and AC-WR-024 (refusal-class tree pre-detected, notice contains `locked`). **Both exercise the pre-detection limb, on the same condition — the lock.** No criterion exercises the limb the whole re-derivation was built on: a refusal condition *outside* the pre-detection set, where the sweep attempts removal, git refuses, and the sweep must preserve and name the cause without aborting its caller. EC-12 asserts that behaviour in prose ("Falls through to fail-open: git refuses, a cause-naming notice is emitted, nothing is lost") with nothing behind it.

This matters more than an ordinary coverage gap because it is the *defining* limb: the SPEC deliberately moved the requirement's weight from the enumeration onto the observable, and the observable is the untested half. The fix is small — the removal seam is already injectable, so a test in which it returns exit 128 asserting (a) the tree survives, (b) the notice names a cause, and (c) the caller is not aborted, is one criterion of the same shape as AC-WR-016.

---

## 5. The two scope corrections — weighed

**(a) The `--ignored` probe's necessity is contingent on the fork.** **Correct, and correctly placed** — in `design.md` §A.7 under "Consequence for the probe", which states plainly that the probe is P1's implementation and P2's classifier and "under **P3 it is not needed at all**", and that "this SPEC does not assert the probe is required". REQ-WR-024 matches: it commits only to "the sweep's handling shall be governed by the ignored-content policy selected per `design.md` §A.7", not to a probe. This is the right correction and it is placed at the layer that owns the decision. The SPEC also flags, rather than silently adopting, `ec9-measurement.md` v2's assertion that the probe is "still the right mechanism" — which is a genuinely good instance of not inheriting a conclusion from a source it otherwise trusts.

One residue: `spec.md` §A's v0.3.0 HISTORY bullet, point (1), says "the `--ignored` probe **survives** but its rationale inverts — … it **is** the only thing between the sweep and destruction of ignored content". That asserts the probe exists in the design, which §A.7 declines to assert. Finding **F8**, minor.

**(b) M1 does not introduce the ignored-content data-loss path.** **Correct, and stated with appropriate confidence — including on the t208 point, which I checked specifically because it is an inference about a tree that no longer exists.**

The claim as made in `design.md` §A.6 is narrow: "`prMergeCleanup` already removes merged, porcelain-clean trees today — that is how worktree t208 was removed during the investigation. What M1 changes is the *rate* … That distinction sets the priority …, not the blame." That claim is supported without the risky step: t208's removal is evidenced by a removal notice in the investigation transcript (`spec.md` §B.2 attributes it there), and the sweep only removes trees it judges merged and porcelain-clean, so the removal itself establishes the class. **The SPEC does not claim t208 carried ignored content.** The dispatch's stronger inference — that as a session-hosting tree it would have carried `.moai/state/` — is *absent from the artifacts*, which is the right call, because the tree is gone and cannot be inspected. The author took the supportable half and left the unsupportable half out. No finding; recorded as a positive.

---

## 6. Growth check (17 → 23 → 24 REQ, 18 → 24 → 26 AC)

- **Restatement**: none that rises to a defect. REQ-WR-021 and REQ-WR-023 overlap on the notice obligation, but 021 is scoped to the refusal class and 023 generalises to every preserve path — 023 subsumes 021's clause rather than repeating it, and `design.md` §B.6a states that widening explicitly ("widened from 'distinguish these two causes' to 'every preserve notice names its cause'"). REQ-WR-024 occupies ground neither covers (the ignored-content policy, which is not a refusal).
- **Scope creep past "repair the shipped reaper"**: the v0.3.0 net movement is **inward, not outward**. The v0.2.0 generalisation that iteration 2 identified as the one unearned growth was re-derived and shrunk (pre-detection set back to `{lock line}`), the submodule case was pushed to Out of Scope rather than absorbed, and the one addition — REQ-WR-024 — is a *deferral marker with a gate*, not new build scope. This is the first revision in the series that removed as much as it added.
- **Counts**: REQ 24 ✓ (declared and measured), AC 26 ✓, header sub-counts 7+12+4+3 = 26 ✓, `progress.md` §E.1 states 24/26 ✓, DoD says `001…024` ✓.
- **AC↔REQ mapping**: complete in both directions.
- **Consistency breaks introduced by the last edit**: four, all minor — F4 (the `clean.go:162` mis-citation), F5 (`progress.md`'s stale "both consumers" bullet), F6 (`research.md` §F's claim about a table it did not change), F7 (`plan.md` §F M2's duplicated step number). None affects a decision; all are edit residue of the kind this SPEC has been good at sweeping and did not sweep this time.

---

## Defects Found (structured defect-list)

**F1.** IGNORED-CONTENT-DECISION-RULE-NOT-DECISIVE — `design.md` §A.7 (decision-rule table + the "What is NOT deferred" paragraph) and `acceptance.md` AC-WR-025 (the `Decision:` line) — The rule maps one of its two branches to two policies: "P1 preserves **more than half** → … Choose between **P2** and **P3** below", and AC-WR-025 repeats it ("otherwise choose P2 or P3"). The same section then claims "A run-phase implementer therefore has a procedure, not a judgement call", which its own table falsifies — and falsifies precisely on the branch the SPEC's stated hypothesis predicts (`.moai/state/` gitignored, written into every session-hosting tree). Compounding: AC-WR-025 collects a **second** datum ("how many carry ignored content **outside** regenerable paths") that no rule consumes, and that datum is exactly the P2/P3 discriminator. — **Evidence**: `design.md` §A.7 table, two rows, second row terminal in "Choose between P2 and P3"; `acceptance.md` AC-WR-025 `Decision:` line; the measurement block two lines above it recording both data. — Severity: **major** — Class: **blocking** — **Required fix**: add the third row keyed to the datum already being collected, e.g. "> half **and** every `!!` entry in the unblocked set lies under a declared regenerable path → **P3**; > half **and** any entry lies outside → **P2**"; mirror it in AC-WR-025's `Decision:` line; and either keep the "procedure, not a judgement call" claim (now true) or delete it. This lands inside the existing [HARD] M1 gate — it does not need a new one.

**F2.** REQ-WR-021'S DEFINING LIMB HAS NO CRITERION — `spec.md` REQ-WR-021 (L188-211) / `acceptance.md` AC-WR-012, AC-WR-024 / EC-12 — REQ-WR-021 is deliberately stated over the observable ("when a worktree is in a state that non-forced `git worktree remove` refuses, the sweep shall preserve … and emit a cause-naming notice"), with pre-detection demoted to an optimisation over a non-exhaustive set. Both criteria citing it exercise **only** the pre-detection limb, and both on the same condition (the lock). Nothing exercises the fail-open path: removal attempted, git refuses, tree preserved, cause named, caller not aborted — the behaviour EC-12 relies on in prose for the submodule case and that REQ-WR-016 requires stay non-blocking. The re-derivation moved the requirement's weight onto the observable and left the observable untested. — **Evidence**: `grep -n '^\*\*Covers:\*\*' acceptance.md` → REQ-WR-021 appears at L261 (AC-WR-012) and L358 (AC-WR-024) only; both criteria's Then clauses read "no removal is attempted". EC-12's expected column asserts the fall-through behaviour with no criterion cited. — Severity: **major** — Class: **blocking** — **Required fix**: add one criterion of AC-WR-016's shape — removal seam returns exit 128 with a non-lock message; assert the tree survives, the notice carries a cause token, and the caller is not aborted; `Covers: REQ-WR-021, REQ-WR-016, REQ-WR-023`.

**F3.** REQ-WR-023'S FIVE CAUSE TOKENS CROSS-CITED AT ONE CRITERION — `spec.md` REQ-WR-023 (L213-217) / `acceptance.md` — REQ-WR-023 requires *every* preserve notice to carry a cause-specific token and enumerates five (refusal-class, dirty, anchored-by-lock, anchored-by-registry, undetermined-merge). Only AC-WR-024 cites it, asserting one token (`locked`). Four of the five causes *are* in fact asserted somewhere — AC-WR-004 (undetermined merge), AC-WR-013 (registry source), AC-WR-014 (lock source) — but none cites REQ-WR-023, and the **dirty** cause has no token assertion anywhere. — **Evidence**: `grep -o 'REQ-WR-023' acceptance.md` → 1 occurrence, at AC-WR-024's `Covers:` line. — Severity: minor — Class: **optional** — **Required fix**: add `REQ-WR-023` to the `Covers:` lines of AC-WR-004/013/014, and add a dirty-cause token assertion to AC-WR-007(b), which already fixtures a dirty tree.

**F4.** `clean.go:162` MIS-CITATION IN FOUR PLACES, CONTRADICTING `:163` IN TWO — `spec.md` §B.3 (L100), `acceptance.md` AC-WR-015 (L324), `design.md` §C.4 (L525), `plan.md` §B B5 (L30) cite `cleanStaleWorktrees`'s `LiveAnchoredSessions` call at `clean.go:162`; `spec.md` REQ-WR-019 (L184) and `design.md` §B.9 (L436) cite `:163`. Measured: the call is at **163**; line 162 is a bare `continue` inside the preceding `case`. The underlying defect is real and independently verified — this is a citation-hygiene failure in a SPEC that leans on cited line numbers as evidence. — **Evidence**: `grep -n 'LiveAnchoredSessions' internal/cli/worktree/clean.go` → `95:`, `163:`; `sed -n '155,168p' internal/cli/worktree/clean.go` shows `162: continue` / `163: case len(session.LiveAnchoredSessions(...)) > 0:`. — Severity: minor — Class: **blocking** (internal inconsistency, mechanical) — **Required fix**: `162` → `163` at the four sites.

**F5.** `progress.md` CONTRADICTS ITSELF ON THE CONSUMER COUNT AND ON "RESOLVED" — `progress.md` L38-52 vs L74-89 — Under the heading "**Design decisions resolved at plan-phase** (no `[NEEDS CLARIFICATION]` markers remain)", bullet D3 still reads "the anchor decision … reaches **both** consumers, `prMergeCleanup` and `cleanStaleWorktrees`" — the v0.2.0 text, contradicted 40 lines below by "**Third blind consumer named.** REQ-WR-019 now covers all three". The same heading's "resolved at plan-phase" claim sits above "**[HARD] OPEN DECISION**". — **Evidence**: `progress.md:44-45` verbatim vs `progress.md:84-89`. — Severity: minor — Class: **optional** — **Required fix**: update D3 to three consumers; qualify the heading to "resolved at plan-phase **except the §A.7 fork, below**".

**F6.** `research.md` §F DESCRIBES A CHANGE IT DID NOT MAKE — `research.md` §F (L206-209) vs §A (L19) — §F's withdrawal note closes "See §A, whose two-sweep table is correspondingly a three-sweep table." §A's table is still two-column (`| | prMergeCleanup | cleanStaleWorktrees |`); the correction is carried by a blockquote above it, not by the table. — **Evidence**: `sed -n '10,20p' research.md`. — Severity: minor — Class: **optional** — **Required fix**: add the third column, or reword to "see the correction note above §A's table".

**F7.** `plan.md` §F M2 STEP NUMBERING DUPLICATED AND OUT OF ORDER — `plan.md` L163-176 — The step sequence is 1, 2, 3, 4, 5, 6, **5**, 7: the v0.3.0 edit inserted a new step 5 (`--merged-only`) and step 6 (ignored-content deferral) ahead of the original step 5 (`cleanStaleWorktrees`), leaving two 5s and placing a wiring step *after* the "NOT implemented in M2" note. Markdown renders the list sequentially, so nothing is lost to a reader — but the source is inconsistent and the ordering now reads oddly. — **Evidence**: `sed -n '150,178p' plan.md`. — Severity: minor — Class: **optional** — **Required fix**: renumber; move the `cleanStaleWorktrees` step adjacent to the other two wiring steps.

**F8.** HISTORY ASSERTS THE PROBE SURVIVES; §A.7 DECLINES TO — `spec.md` §A, v0.3.0 bullet, point (1) — "the `--ignored` probe **survives** but its rationale inverts — … it **is** the only thing between the sweep and destruction of ignored content". `design.md` §A.7 states the opposite posture: "under **P3 it is not needed at all**. So this SPEC does not assert the probe is required." REQ-WR-024 follows §A.7, so the HISTORY entry is the outlier. — **Evidence**: `spec.md:21` vs `design.md` §A.7 "Consequence for the probe". — Severity: minor — Class: **optional** — **Required fix**: "the probe survives **under P1 and P2**; under P3 it is not needed — its necessity is contingent on the open fork".

**F9.** THE M1 GATE IS ABSENT FROM M1'S OWN STEP LIST — `plan.md` §F M1 (steps 1-5) — The AC-WR-025 gate is stated in `plan.md` §C.1 [HARD], AC-WR-025 [HARD], the DoD, `plan.md` §G, and `progress.md`; it is **not** referenced in the M1 milestone section an implementer works from. — **Evidence**: `grep -n 'AC-WR-025\|C.1' plan.md` → L48, L167 (M2 step 6), L231 — no hit inside §F M1. — Severity: minor — Class: **optional** — **Required fix**: one line at the head of §F M1: "**Blocked on `plan.md` §C.1 (AC-WR-025) — do not start until the §A.7 fork is closed.**"

### Positive findings (recorded, not defects)

- **The correction went further than the finding required.** N1 asked for a measurement to be corrected. The revision corrected it, withdrew the scope the bad measurement had motivated (the two-member refusal class), re-derived the surviving requirement over a form that cannot repeat the failure, and pushed the newly-found member (submodules) to Out of Scope with a measured reason. Shrinking scope in response to an audit is rarer and harder than adding to it.
- **`ec9-measurement.md` v2 destroys v1 rather than appending to it**, and says why: "leaving a superseded version in place invites one of those to be re-derived from it." That is the correct handling of a withdrawn measurement and it is the direct reason the six downstream sites are now consistent.
- **`research.md` §F strikes through its own reasoning and states why it was wrong** — "which is an argument for *including* it" — rather than quietly deleting the sentence. Same for `design.md` §B.9.
- **The evidence record across three iterations is the strongest part of this SPEC.** 29/29 criterion baselines reproduce at iteration 3, as 23/23 did at iteration 2. Across two growth rounds not one recorded observation has failed to reproduce except the single one the SPEC itself has now withdrawn.
- **The t208 inference was left out.** The supportable claim (the removal path pre-exists M1) is stated and evidenced; the unsupportable one (that t208 specifically held ignored content) is absent. Declining to make an unfalsifiable claim that would have strengthened the argument is exactly the discipline `verification-claim-integrity.md` §1 asks for.

---

## Gaps (what this audit did NOT observe)

Stated explicitly, because an empty Gaps section would itself be a claim.

1. **I built no git fixture.** Every EC-9-class behavioural claim (`git worktree remove` on an ignored-only tree, on a submodule tree, on a missing directory) is carried forward from iteration 2's measurements and `ec9-measurement.md` v2, not re-measured here. Both records now agree, and iteration 2's were taken outside this tree — but I did not independently reproduce them, because this session sits inside a live worktree and that position is what corrupted the v1 measurement. Re-measurement belongs to the AC-WR-025 run, from the primary checkout.
2. **AC-WR-025 remains unrun, by construction.** The worktree-isolation guard refuses `git -C` into sibling trees from here, so the ignored-content prevalence across the 154 trees is unmeasured — which is the SPEC's own recorded position, not a discrepancy.
3. **AC-WR-022 and AC-WR-023 were not executed.** Both are completion gates rather than discriminators; AC-WR-023(b) is destructive and operator-gated. `go build ./...` was not run as part of this audit; the two package test runs above compiled `internal/cli` and `internal/cli/worktree` successfully, which is weaker evidence than AC-WR-022 requires and is not offered as satisfying it.

## Residual risk

The SPEC's correctness now rests on the AC-WR-025 measurement being run **and on F1 being fixed before it is interpreted**. If the measurement is taken while the rule still terminates in "choose between P2 and P3", the implementer will make a preference decision at run-phase that plan-phase declared it had eliminated — and it will be made with the gate showing green. That is the one place where this SPEC could still ship a judgement disguised as a procedure.

---

## Recommendation

**PASS at 0.875 against a Tier L threshold of 0.85.**

Defending the score rather than asserting it: all seven must-pass criteria pass and were checked mechanically, not inferred. All seven iteration-2 findings are closed, and I verified each at its artifact site rather than accepting the disposition — including the one that required checking six separate downstream sites. All nineteen iteration-1 findings remain closed, and D3 is now closed completely rather than partially. The falsifiability property that drove iteration 1's FAIL survives two growth rounds intact: 29 distinct test names, 0 discrepancies, with the two new criteria run in their exact form. The mandatory t207 sample is judged not-disposable and is bound to the lock path with the registry emptied, which is the only construction that is not 4-of-5 blind.

Iteration 2's blocker was a false claim about the world that had propagated into six artifacts and grown a requirement, a criterion and two edge cases on top of itself. Nothing of that kind survives. What remains is one genuinely defective decision rule, one uncovered requirement limb, and seven pieces of edit residue — and none of them is a claim that is untrue about the code or about git.

**Fix before M1 lands** (both are inside gates the SPEC already declares, so neither needs a new one):

1. **F1** — add the third row to `design.md` §A.7's table, keyed to the regenerable-path datum AC-WR-025 already collects, and mirror it in AC-WR-025's `Decision:` line. This must land **before** the §A.7 fork is closed, not merely before M1 — otherwise the gate goes green on a judgement call.
2. **F2** — add one fail-open criterion for REQ-WR-021's defining limb (removal seam returns 128; assert preserve + cause token + caller not aborted).

**Sweep at leisure**: F4 (`162` → `163`, four sites), then F3, F5, F6, F7, F8, F9 at the orchestrator's discretion. F3 and F5-F9 are optional-class and do not gate anything.

No STOP signal: the score rose. No escalation required despite this being iteration 3 — the SPEC clears the bar.
