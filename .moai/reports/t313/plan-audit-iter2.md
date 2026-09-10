# SPEC Review Report: SPEC-WORKTREE-BASEREF-001

Card: t313 · Iteration: **2/2 (Tier M ceiling — final)** · Auditor: plan-auditor
Tree measured: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t313`, branch `WT-worktree-baseref`, HEAD `48eb945df` (re-read by this auditor with `git rev-parse --show-toplevel` / `--short HEAD` / `git branch --show-current`, not inherited from the dispatch).
Artifacts read: `spec.md` (0.3.0), `plan.md`, `acceptance.md`, `progress.md` — the full Tier M input set. Prior verdict `plan-audit-iter1.md` (189 lines) and reviser summary `spec-iter2-changes.md` (127 lines) both read in full before auditing.

Reasoning context ignored per M1 Context Isolation. The reviser's change summary was treated as **the claim under test**, never as evidence: every disposition below was re-derived from the artifacts and, where mechanical, re-measured in this tree.

**Verdict: PASS-WITH-DEBT**
**Aggregate score: 0.92 (harmonic mean) / 0.925 (arithmetic mean) — Tier M PASS threshold 0.80**

Eleven of the twelve iteration-1 findings are genuinely repaired at the layer the verdict named. One (D9) is partially repaired: the judgement steps are gone, but the diff-scoped pipeline that replaced check (1) **fails a correct implementation of this SPEC's own plan §D write list** — measured, not inferred. That, plus one ambiguity introduced by the D1 repair, is the debt. Neither is a must-pass failure and neither justifies burning the final iteration; both are small, precisely-bounded edits enumerated in § Debt Carried Into Run-Phase.

Score movement: **+0.14 harmonic** (0.78 → 0.92). Traceability and Testability — the two dimensions that carried the iteration-1 FAIL — both improved materially. Clarity held. Completeness improved to the rubric ceiling.

---

## Must-Pass Results (7/7 clear — re-verified, not carried over)

A repair can break a must-pass criterion, so all seven were re-run against the 0.3.0 artifacts rather than inherited from iteration 1.

- **[PASS] MP-1 REQ number consistency** — `grep -oh 'REQ-WBR-[0-9]\{3\}' *.md | sort -u` returns exactly `REQ-WBR-001 … REQ-WBR-016`, count **16**, sequential, no gaps, no duplicates, uniform 3-digit padding. `grep -c '^\*\*REQ-WBR-' spec.md` returns **16**, so every id has a definition and none is reference-only. Matches the orchestrator's stated count; **no contradiction**.
- **[PASS] MP-2 GEARS format compliance** — all 16 `REQ-WBR-*` entries in `spec.md` §B carry an explicit pattern label and match it. Enumerated from source: Ubiquitous 001/002/014/015; Unwanted 003/008; Event-driven 004 (compound)/007/009/010/012/016; State-driven 005/006/011; Where 013. The three requirements iteration 2 rewrote were re-checked against their labels — REQ-WBR-004 (`spec.md:82`) is `When … while …, shall …; and while …, shall …`, GEARS-compound and PASS-equivalent; REQ-WBR-011 (`:102`) retains its `While … shall …` head clause with the parity binding appended as elaboration; REQ-WBR-013 (`:108`) retains its `Where … shall …` head with the preservation clause as a second `shall` conjunct under the same gate. Judgement made against the **requirement layer only**; the Given/When/Then entries in `acceptance.md` are verification-layer `AC-XXX` and were graded under Group 4, never here. One minor note recorded below (N3), not a failure.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types, read at `spec.md:2-13`: `id`, `title`, `version: "0.3.0"` (quoted semver, bumped from 0.2.0), `status: draft`, `created: 2026-08-27`, `updated: 2026-08-27`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated string). No rejected snake_case alias (`created_at` / `updated_at` / `labels` / `spec_id`). `tier: M` and `related_specs` are additive, not schema violations.
- **[PASS] MP-4 language neutrality** — unchanged in substance by iteration 2. The SPEC targets this repository's own Go codebase plus one shipped config key; it names no programming-language-specific tooling, and the one shipped artifact it touches is a language-agnostic config template. D5's grep change was checked for regression against this criterion: see § B3, where the narrowed pipeline was re-run against five fixtures — it still binds a shipped **value**, and REQ-WBR-003 + AC-WBR-002 remain the two-layer assertion. **Neutrality survives D5's change.**
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — two external ids referenced. Measured: `SPEC-WORKTREE-BRANCH-GUARD-001` reports `status: completed`; `SPEC-SYNC-STRATEGY-KEY-001` reports `status: in-progress`. Neither is `retired` / `superseded` / `archived`; both directories exist. No BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c 'syscall'` returns `0` in each of the four artifacts. Auto-PASS; no build-tag concern.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' .moai/specs/SPEC-WORKTREE-BASEREF-001/` exits rc=1 with no match. `progress.md:13` independently records `blocker: none` with the D2 widget ruling CLOSED by the operator. No open marker.

---

## Category Scores (rubric-anchored, with iteration-1 delta)

| Dimension | iter1 | iter2 | Δ | Direction | Why |
|-----------|-------|-------|---|-----------|-----|
| Clarity | 0.90 | 0.90 | 0.00 | **held** | iter1's named deduction (AC-WBR-012's `Then` narrower than the REQ it mapped to) is closed. It is replaced at comparable weight by a new one: AC-WBR-016 half 1's "read seam" admits two denotations that disagree on the empty-value branch its own `Given` admits (N2). REQ-WBR-004's compound form is long, but each clause has exactly one reading and each is separately criterion-backed — the Clarity cost the reviser predicted is real and small, not the driver of this hold. |
| Completeness | 0.95 | 1.00 | +0.05 | **improved** | iter1's sole deduction — `plan.md` §D omitting any file for the round-trip test — is closed at `plan.md:137`. Rubric 1.0 conditions all hold: every required section present; frontmatter 12/12; six `### Out of Scope — <topic>` H3 sub-headings each carrying a specific `-` bullet (`spec.md:120-142`); `plan.md` §B still enumerates G1-G7 verbatim (none deleted) and §E carries five risks. One cosmetic non-scoring note: HISTORY rows run 0.1.0 → 0.3.0 → 0.2.0 (`spec.md:26-28`), out of order. |
| Testability | 0.70 | 0.85 | +0.15 | **improved** | The dominant iter1 deduction — eight MUST criteria dischargeable while executing zero tests — is closed **mechanically**, not by prose: `acceptance.md:330-338` requires `-v` on every `-run`, `grep -c '^=== RUN' >= 1`, and records both the RUN count and the exit code as PASS conditions. D8's unpinned check-name string is fixed at the requirement layer. Held back from higher by D9's partial repair — AC-WBR-013 check (1) is now binary but **wrong**, emitting a false failure for a correct implementation (N1, measured) — and by N2. |
| Traceability | 0.65 | 0.95 | +0.30 | **improved** | Both iter1 exceptions closed and re-measured here. The `§D` matrix REQ column reads `001,002 / 003 / 005 / 006 / 007 / 008 / 010 / 011 / 012 / 013 / 014 / 015 / 016 / 013 / 009 / 004` — **16 of 16 REQs covered, 0 of 16 ACs orphaned**. Every `AC-WBR-*` id appearing anywhere in the four artifacts is defined in `acceptance.md` (16 distinct across all files equals 16 distinct in `acceptance.md`), so there is no dangling reference either. Held below 1.0 for the folding thinness: REQ-WBR-004's and REQ-WBR-013's second obligations are each carried by one AC *half* or by a shared AC rather than a distinct id, so an id-level coverage check would still read "covered" if AC-WBR-016 half 2 were quietly dropped in run-phase. |

Harmonic mean = 4 / (1/0.90 + 1/1.00 + 1/0.85 + 1/0.95) = **0.9216 → 0.92**. This clears the Tier M threshold of 0.80 on the **binding** (harmonic) aggregate, not merely on the arithmetic one — the reverse of iteration 1's situation, where only the arithmetic mean reached the line and only by rounding.

---

## Per-Defect Repair Table

Every disposition re-derived from the artifacts. "REPAIRED" means the repair is present **and** closes the defect — checked against the *effect* iteration 1 named, not against the mere presence of an edit at the cited line.

| # | iteration-1 finding | Disposition | Evidence |
|---|---------------------|-------------|----------|
| **D1** | REQ-WBR-004 has no acceptance criterion | **REPAIRED** | `acceptance.md:29` carries matrix row `AC-WBR-016 \| MUST \| 004`; the scenario is at `:287-315`. Half 1 asserts the **read** seam — explicitly distinguished from the `git remote set-head` write seam AC-WBR-003/-005 assert on — is invoked exactly 1 time per `Handle` invocation, with "Zero invocations FAIL (the task was never registered)" stated at `:294`. It closes the defect the verdict named: the `moai doctor`-only implementation that satisfied all four outcome criteria now fails an explicit criterion. The verdict's *preferred* fix (a new AC rather than extending AC-WBR-003) is what landed. Carries N2. |
| **D2** | eight MUST criteria pass vacuously | **REPAIRED** | `acceptance.md:330-338`, first bullet of §D.3. Binds by name — 003, 004, 005, 006, 010, 011, 012, 015, 016 (the original eight plus the new one). Not prose-only: `-v` is mandated, `grep -c '^=== RUN' out.txt` must be `>= 1`, a `[no tests to run]` line is declared VACUOUS, and **both** the RUN count and the exit code are recorded under `.moai/state/verify/t313/` with PASS requiring both. The iteration-1 measurement is quoted inline as the reason, so a later reader cannot mistake the bullet for ceremony. One edit covering all affected criteria, as the verdict directed. |
| **D3** | AC-WBR-012 verified one third of REQ-WBR-015 | **REPAIRED — beyond the literal ask** | `acceptance.md:195-214`. The `Then` is now the explicit three-part conjunction (present in `settings.AllFields()` / present in rendered HTML / reaches a consumer), stated as "a conjunction, not a disjunction". The command at `:206` targets `./internal/web/... ./internal/hook/... ./internal/cli/...` — `./internal/web/...`, where the guard lives, added. Two additions close the deeper hole: `:211` asserts the **guard's existence** (a named test beside `internal/web/dead_config_guard_test.go`) and states that AC-WBR-005/-007's consumer tests do NOT discharge it; `:212` asserts its **failure mode** by mutation (delete the `FieldDef` from `gitStrategyFields()`, re-run, observe non-zero, restore). |
| **D4(a)** | R1's "one repository, one key" premise false | **REPAIRED** | `plan.md:149-165`. The sentence is quoted, named **false**, and named **load-bearing**. The tracked-file measurement is inline at `:152-155`; I re-ran it — `git ls-files --error-unmatch .moai/config/sections/git-strategy.yaml` returns rc **0**, confirming the count inversion. `:157` states the continuous re-creation mechanism (lane A's write creates exactly the difference lane B's next session start reverses) and the external-actor case that makes it repeated even in a single-lane repository. `:159` states the real steady-state invariant *and* that it is "not a property this SPEC can assume; it is one it must arrange". `:165` states residual risk rather than dismissing it. No trace of the false premise survives anywhere in the artifacts. |
| **D4(b)** | consumer-1 concurrency narrowing | **REPAIRED** | Six coordinated landings, all verified: `spec.md:82-84` (REQ-WBR-004 compound, primary-checkout narrowing plus the linked-worktree total-no-op clause, with the discriminant cited); `plan.md:50-52` (decision record D3.1, so the narrowing is filed as a design decision rather than buried in a risk); `plan.md:91` (M2 gates on the discriminant **before** reading the configured value, reusing the existing shape rather than inventing a second test); `acceptance.md:287-315` half 2 (the verifying criterion); `acceptance.md:35` (§D.1 preface making the primary checkout the implicit `Given` of consumer-1 scenarios and stating consumer-2 scenarios carry no such precondition); `acceptance.md:326` (edge case replacing the mis-analysed "misconfiguration" reading). Citation re-measured at source: `inGitWorktreeReal` is at `internal/cli/session_worktree.go:234-241` and does compare `git rev-parse --git-dir` against `--git-common-dir` — **exactly as cited**. |
| **A5** | consumer parity asserted in the plan, unbound by the requirements | **REPAIRED — at the requirement layer, which is where iteration 1 said it escaped** | See § B5 for the full ruling. `spec.md:102` (REQ-WBR-011) now reads: *"'Unresolvable' shall be decided by the predicate REQ-WBR-009 specifies, which shall be implemented **once** as a single shared helper and shall be the **sole** resolvability authority for both consumers"*, and declares each of the three divergent implementations iteration 1 named a violation **"even when their runtime behaviour agrees"** — the clause that actually closes the hole. `acceptance.md:134-138` gives AC-WBR-008 a third, **structural** assertion and states that a behavioural-equivalence check does NOT satisfy it. `plan.md:98` (M3) restates it as calling M2's exported helper through a seam. |
| **D5** | AC-WBR-002's grep bound the whole line | **REPAIRED — measurement independently replicated** | `acceptance.md:54-62` takes the value-side narrowing; `plan.md:83` (M1) is aligned to it so the two no longer contradict. I re-ran the pipeline against the reviser's two fixtures plus three of my own — see § B3. It discriminates correctly; there is no quoting or field-splitting inversion. |
| **D6** | AC-WBR-014 traced to `R4 / G3` (orphaned) | **REPAIRED** | `spec.md:108` — REQ-WBR-013 absorbs the preservation clause naming all three keys with their file locations and their absence from `ModeProfile`. `acceptance.md:27` — REQ column now `013`, severity MUST. `acceptance.md:340` — the §D.3 "single SHOULD criterion" bullet is replaced; the set is all-MUST and an AC-WBR-014 failure escalates as a blocker rather than being absorbed. `plan.md:137` — the write list gains `internal/settings/sectionapply_test.go` (or a `gitstrategy_roundtrip_test.go` sibling), which also closes iteration 1's Completeness deduction. `plan.md:177` — R4 correctly *retained* (the underlying `ModeProfile` gap is still unrepaired) while recording that preservation is now required and verified. The verdict's preferred disposition (promote to the requirement layer) was taken. |
| **D7** | citation drift at REQ-WBR-008 | **REPAIRED — verified at source** | `spec.md:92` now cites `internal/hook/session_start.go:176-181`. Measured in this tree: the contract sentence *"Best-effort contract preserved: Handle never returns a non-nil error from these steps"* begins at **`:176`** and its comment block runs to `:179`, with the surrounding branch closing at `:181`. `plan.md` §D3 already cited `:176-181`; spec and plan now agree about the same anchor. |
| **D8** | AC-WBR-009 pinned a check name no requirement fixed | **REPAIRED — verified at source** | `spec.md:106` — REQ-WBR-012 fixes `DiagnosticCheck.Name` to exactly the string `Worktree Base Branch` and states the name is part of the contract, not an implementation detail. The justification is measured: `internal/cli/doctor.go:232` reads `if filterCheck != "" && c.name != filterCheck`, exact-name equality, exactly as cited. AC-WBR-009's manual step (`acceptance.md:148`) now names a string the requirement layer fixes. |
| **D9** | AC-WBR-013 judgement step + uninterpretable loop | **PARTIALLY REPAIRED** | The judgement steps are genuinely gone: `acceptance.md:216-249` replaces both with diff-scoped pipelines against `BASE=$(git merge-base HEAD origin/develop)`, and the vacuous-by-construction reading of check (2) is now stated explicitly at `:249` rather than left uninterpretable — that part is a clean repair. `git merge-base HEAD origin/develop` resolves in this tree to `48eb945df606eea7d6d3d1b9a1020adbfe79b2e6`, so `BASE` is real and the commands run. **What remains**: check (1) emits a false `NO-TEMPLATE-COUNTERPART` for this SPEC's own `.moai/config/sections/git-strategy.yaml` (N1, measured), and check (2) carries a latent path-mixing bug (N4). Both in § B4. |
| **D10** | `grep -n -A0` no-op flag | **REPAIRED** | `acceptance.md:55` — `-A0` is gone; the line is now `grep -n 'worktree_base_branch' internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`. |

**Tally: 11 REPAIRED (two of them stronger than the literal ask), 1 PARTIAL, 0 NOT REPAIRED.** No repair moved its defect rather than closing it.

---

## B1-B5 Rulings

### B1 — The ceiling-driven folding: **accepted; both obligations are present and enforceable**

The orchestrator instructed two *new* requirements. The reviser instead folded them into REQ-WBR-004 and REQ-WBR-013 to stay at 16/16, and said so openly. Three questions decide this.

**Is the ceiling reasoning founded?** Yes, and I re-read the rule rather than taking it on trust. `spec-workflow.md:146-150` gives the table (Tier M: 16 requirements, 16 acceptance criteria) and `:152` states the ceilings *"apply **independently** to the requirement count and to the acceptance-criterion count — never to their sum"*, adding that exceeding either *"is a signal to tier up or to split the SPEC, not to relax the budget"*. The reviser's alternative (18 REQ / 17 AC) would have breached both axes. So the constraint is real, and repairing one audit finding by manufacturing another would not have been a repair. (Minor note: the reviser cites `:152`, the orchestrator cites `:148`; both point into the same block — the values are at `:149`, the independence sentence at `:152`. No contradiction of substance.)

**Is the first obligation — the primary-checkout narrowing — enforceable in its compound home?** Yes. This is the test that matters, because a compound requirement whose second clause has no criterion is a regression rather than a repair. REQ-WBR-004's second clause (`spec.md:82`) is not decorative prose: it states a **total** no-op ("no read of either, no write, and shall emit no output"), it names the discriminant with a source citation I verified, and it is backed by **AC-WBR-016 half 2** (`acceptance.md:296-299`), which asserts four conjuncts — read seam 0, write seam 0, stderr empty, nil error. Better still, `acceptance.md:308` forecloses the obvious cheat: *"A half-2 test that only exercises the empty-value path does not discharge this criterion, because the empty value already no-ops under REQ-WBR-005."* Half 2 is pinned to the exact configuration AC-WBR-005 requires a **write** for, which is what makes it a real negative rather than a restatement of the empty-value path. That is a criterion a divergent implementation actually fails.

**Is the second obligation — round-trip preservation — enforceable in REQ-WBR-013?** Yes. `spec.md:108`'s second conjunct names all three keys, their file locations, and their absence from `ModeProfile`, and states the *reason* it belongs here ("a correctness property of the write path this SPEC introduces"). It is backed by AC-WBR-014, now MUST and now mapped to `013` (`acceptance.md:27`), with the test file named in `plan.md:137`. Nothing was dropped in the fold.

**The Clarity cost the reviser predicted.** It is real but smaller than the reviser feared. REQ-WBR-004 is now a four-sentence compound with an explanatory block-quote, which is more to hold in view than two short requirements would be. Against that: the block-quote at `spec.md:84` carries the *reason* for the narrowing next to the narrowing itself, and explicitly bounds it ("binds **consumer 1 only**"), which two separate ids would have separated. I scored Clarity as held, not reduced, and the deduction I did apply is N2, which has nothing to do with the folding. **The folding was the right call.**

One structural consequence I do record, and it is the reason Traceability is 0.95 rather than 1.00: id-level coverage no longer distinguishes "REQ-WBR-004's firing point is verified" from "REQ-WBR-004's *narrowing* is verified", because both live under one id and one AC. If AC-WBR-016 half 2 were quietly dropped in run-phase, a matrix check would still read `004 → covered`. That is a real, small loss of resolution, correctly traded for staying inside the ceiling.

### B2 — AC-WBR-016's read-seam presumption: **an acceptable plan-phase constraint, not an unverifiable criterion**

The reviser's stated residual risk is that AC-WBR-016 half 1 presumes the read path is implemented as a seam, that no existing seam carries it, and that an implementer using a direct `exec.Command` "would find the criterion unsatisfiable and would have to refactor rather than merely add a test."

**The premise is right but its consequence is over-stated, and the difference is decisive.** I measured both halves of it:

- No function-variable test seam of the swap-and-restore kind exists in `internal/hook` today. `grep` over `internal/hook/*.go` (excluding tests) for function-variable assignment returns one hit, `branch_guard.go:113`, which is a computed package-level var rather than an injection seam. So "no existing seam carries it" is accurate.
- **But the read path is new code, not existing code.** `plan.md:127` names `internal/hook/worktree_base_branch.go` as a **new** file, and M2 (`plan.md:89-93`) writes the whole helper. There is nothing to refactor: the criterion constrains the *shape of code this SPEC's own milestone is about to write*.
- The pattern is already in-repo and one package away. `internal/cli/session_worktree.go:48-53` carries exactly this idiom, with a comment reading *"Function-variable seams for test injection. Each has a Real counterpart below; tests swap these via `swapSessionWorktreeSeams` and restore on cleanup."* AC-WBR-008's own third assertion (§ B5) reuses that same mechanism, so the SPEC is consistent about it.
- The plan states the obligation where the implementer will read it: `plan.md:93` — *"Both assert on a seam call count, so the read path is a seam, not a direct `exec.Command`."*

**Ruling: this is the SPEC legitimately dictating a testable shape, and it will not stall run-phase.** A criterion is unverifiable when satisfying it requires something outside the change's scope or unavailable to the implementer; here it requires one `var readAlignmentState = readAlignmentStateReal` line in a file the same milestone creates, following a precedent the repository already ships. The cost is a few lines; the thing bought is the only assertion in the set that catches an implementation that is behaviourally perfect and never wired into the errgroup — which is precisely the failure iteration 1's D1 was raised over. Removing it to avoid the constraint would re-open D1.

**One qualification, and it is N2, not B2.** The criterion is safe in its *mechanism* and ambiguous in its *denotation* — see N2 below. That is a wording fix, not a reason to doubt the seam constraint.

### B3 — D5's `sed`/`cut` pipeline: **re-run independently; it discriminates correctly, with one benign over-match**

The reviser claims a measurement. I did not accept it; I re-ran the exact pipeline from `acceptance.md:57-58` against the reviser's two fixtures plus three of my own. Verbatim:

```
$ printf 'worktree_base_branch: ""  # e.g. main, develop; empty = no action\n' > /tmp/wbr1.yaml
$ grep 'worktree_base_branch' /tmp/wbr1.yaml | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main; echo "wbr1 rc=$?"
wbr1 rc=1                                   # comment naming both branches → PASSES the AC ✓

$ printf 'worktree_base_branch: "develop"\n' > /tmp/wbr2.yaml
$ grep 'worktree_base_branch' /tmp/wbr2.yaml | sed 's/#.*//' | cut -d: -f2- | grep -e develop -e main; echo "wbr2 rc=$?"
 "develop"
wbr2 rc=0                                   # value naming a branch → FAILS the AC ✓

$ printf 'worktree_base_branch: "{{.MainBranch}}"  # main\n' > /tmp/wbr4.yaml
$ ... ; echo "wbr4 rc=$?"
wbr4 rc=1                                   # placeholder value + comment naming main → passes

$ printf 'worktree_base_branch: "feature/main-thing"\n' > /tmp/wbr5.yaml
$ ... ; echo "wbr5 rc=$?"
 "feature/main-thing"
wbr5 rc=0                                   # substring over-match → fails
```

**The reviser's measurement replicates exactly.** There is no quoting or field-splitting error, and the criterion is not inverted: comment-side mentions pass, value-side mentions fail. `cut -d: -f2-` correctly keeps everything after the first colon (so a value containing a colon survives intact), and `sed 's/#.*//'` runs first, so a `#` inside the value would truncate it — irrelevant here, since REQ-WBR-003 requires the shipped value to be empty.

Two observations, neither a defect:

- **wbr5 (over-match).** A value like `feature/main-thing` fails on the substring `main`. This is harmless in context: REQ-WBR-003 requires the shipped template value to be **empty**, so *any* non-empty value already violates the requirement. The grep is a belt over a brace.
- **wbr4 (the one real gap, and it is a carry-over, not new).** AC-WBR-002's *first* command (`acceptance.md:55`) has no mechanical pass condition — `# expect a line whose value is ""` is read by a human. So a Go-template placeholder value that renders to a branch name would pass both commands. This weakness is identical in iteration 1 (whose whole-line grep would equally have missed `{{.MainBranch}}`, since `grep -e main` is case-sensitive), so **D5's narrowing introduced no regression**; I record it as N5 debt for completeness, not as a repair failure.

**Template neutrality survives the change** (MP-4 re-checked above): the prohibition still binds the shipped value at both layers.

### B4 — D9's diff-scoped pipelines: **runnable, decidable when empty — but check (1) is wrong, and it is wrong for this SPEC specifically**

Three questions were asked. The first two pass; the third exposes the one finding that keeps this verdict from being a clean PASS.

**Are they runnable as written?** Yes. `BASE` resolves — I ran `git merge-base HEAD origin/develop` in this worktree and it returned `48eb945df606eea7d6d3d1b9a1020adbfe79b2e6`. The pipelines use no undefined variable, no unavailable tool, and no shell construct that would fail under `sh`.

**Is the pass condition decidable when the diff is empty?** Yes, and I measured it rather than reasoning about it. At present `HEAD` *is* the merge base, so the diff is empty:

```
$ git diff --name-only 48eb945df606eea7d6d3d1b9a1020adbfe79b2e6..HEAD -- .claude .moai
$ echo rc=$?
rc=0
```

Both `while read` loops therefore enumerate nothing, emit nothing, and the criterion's stated pass condition ("expect no `NO-TEMPLATE-COUNTERPART` lines" / "expect no `DRIFT` lines") is satisfied. An empty diff passes cleanly rather than hanging or erroring — and `acceptance.md:249` now states that reading for check (2) explicitly, which is exactly the repair iteration 1 asked for.

**Is check (1) correct?** **No — and this is N1.** Check (1) takes each changed `.claude/` or `.moai/` path `$f` and requires that the same diff also changed `internal/template/templates/$f`. That construction assumes the template counterpart is a **same-named plain file**. For this SPEC's own primary template surface it is not. Measured:

```
$ ls internal/template/templates/.moai/config/sections/
… git-strategy.yaml.tmpl …          # only the .tmpl form exists; no plain git-strategy.yaml
```

`plan.md:125` lists `.moai/config/sections/git-strategy.yaml` in the §D write list (M1, "Local mirror"). A correct implementation therefore changes that path, check (1) looks for `internal/template/templates/.moai/config/sections/git-strategy.yaml`, finds nothing in the diff because only the **`.tmpl`** counterpart was changed, and emits `NO-TEMPLATE-COUNTERPART .moai/config/sections/git-strategy.yaml`. The criterion says "expect no `NO-TEMPLATE-COUNTERPART` lines`", so **AC-WBR-013 FAILS for a correctly-executed implementation of this SPEC's own plan.**

The SPEC already knows the underlying fact — `plan.md:70` G5 states *"No template mirror of `.moai/config/sections/git-strategy.yaml` exists as a plain file … The new key must be added there, not to a non-existent plain mirror."* The D9 repair simply did not account for its own §B G5. By contrast the other templated path in the write list works: `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` **does** exist as a plain file (measured), so `plan.md:138-139`'s pair passes check (1) as written. The failure is specific and predictable, not general.

**Ruling on B4:** runnable — yes; decidable on an empty diff — yes; *correct* — no, for one enumerable path. Fix named in § Debt.

### B5 — A5's repair: **verified at the requirement layer, which is where iteration 1 said it escaped**

Iteration 1's A5 was the finding "most likely to actually bite in run-phase" because the plan's prose read as though the obligation were discharged while the requirement layer carried nothing. So the only ruling that matters is whether the binding now exists in `spec.md`, not whether `plan.md` says it more emphatically.

**Requirement layer — present and strong.** `spec.md:102` (REQ-WBR-011), verbatim: *"'Unresolvable' shall be decided by the predicate REQ-WBR-009 specifies, which shall be implemented **once** as a single shared helper and shall be the **sole** resolvability authority for both consumers; consumer 2 shall not carry a second resolvability rule of its own (`git rev-parse --verify`, a `git branch --list` scrape, or a local-branch check are each a violation of this requirement even when their runtime behaviour agrees)."*

Three properties make this a real closure rather than a restatement:

1. It names REQ-WBR-009 as the authority, so the cross-reference iteration 1 found missing now exists at the requirement layer.
2. It requires the predicate be implemented **once**, which is the structural obligation — not merely that the two consumers agree.
3. **"even when their runtime behaviour agrees"** is the load-bearing clause. Without it, a divergent second rule that happens to produce the same answers would satisfy a naive reading. This sentence makes divergence itself the violation, which is what the plan's D4 note always intended and never bound.

**Criterion layer — structural, as requested.** `acceptance.md:134-138` adds AC-WBR-008's third assertion: swap the shared resolvability helper through the same seam mechanism the package already uses for `sessionWorktreeGitWorktreeAdd` (I verified that mechanism exists at `internal/cli/session_worktree.go:48-53`, with its swap-and-restore comment), and assert the fake was invoked exactly once with the configured value during materialization. The text states outright that *"A behavioural-equivalence check (both paths reject `no-such-branch`) does NOT satisfy this — it passes for a divergent second rule, which is the defect the assertion exists to catch."* That is precisely the distinction iteration 1 asked for, stated in the criterion rather than left to the reader.

**Plan layer — consistent.** `plan.md:98` (M3) now says resolvability is decided *"by **calling M2's exported helper**, never by a second rule — REQ-WBR-011 now carries this at the requirement layer, so a divergent … implementation is a requirement violation, not merely a style deviation"*, and requires the helper be exposed through a seam so the assertion is testable. `plan.md:90` (M2) keeps the "single exported helper" instruction.

**Ruling: A5 is fully repaired, and repaired at the layer that was gapped.** An implementation with a second, divergent resolvability rule now violates a requirement *and* fails a criterion, whichever way its runtime behaviour falls. This was iteration 1's highest-risk finding and it is the cleanest of the eleven repairs.

---

## New Defects Introduced by Iteration 2

At the ceiling these matter more than surviving findings, because there is no further revision cycle to catch them.

**N1 — AC-WBR-013 check (1) fails a correct implementation of this SPEC's own plan §D.** `acceptance.md:229-236` — Severity: **major** — Class: **blocking (carried as debt; see the verdict rationale)**.
Measured and demonstrated in § B4. The check requires a same-named plain-file template counterpart; `.moai/config/sections/git-strategy.yaml`'s counterpart is `git-strategy.yaml.tmpl`, a fact `plan.md:70` G5 already records. A correct M1 landing emits one false `NO-TEMPLATE-COUNTERPART` line and the criterion FAILS.
**If it ships as written:** run-phase hits a red MUST criterion on a correct change and must either debug the criterion mid-run or — the worse outcome — conclude the criterion is noise and start disregarding `NO-TEMPLATE-COUNTERPART` output generally, which destroys the check's value for the paths where it does work.
**Required fix (two lines):** make the counterpart probe accept the `.tmpl` form, e.g. replace the inner probe with one that tests both `internal/template/templates/$f` and `internal/template/templates/$f.tmpl` and reports only when *neither* was changed in this diff.

**N2 — AC-WBR-016 half 1's "read seam" has two denotations that disagree on the empty-value branch its own `Given` admits.** `acceptance.md:292-294` vs `plan.md:89` — Severity: **minor-to-major** — Class: **blocking (carried as debt)**.
Half 1's `Given` admits *"`worktree_base_branch` carries any value (**set or empty**)"*, and its `Then` requires the read seam — defined as *"the one that reads the configured value **and** `refs/remotes/origin/HEAD`"* — to be invoked exactly 1 time. But `plan.md:89` prescribes the helper's ordering as *"read config → no-op silently on empty → read `refs/remotes/origin/HEAD`"*, and REQ-WBR-005 requires the empty path to perform **no** git-metadata read. Under the natural reading — the seam is the `origin/HEAD` read — a compliant implementation short-circuits before reaching it and half 1 records **0** invocations on the empty branch, which half 1 declares a FAIL. Under the other reading — the seam is one combined entry-point helper invoked once regardless of value — it passes. Both readings are available from the text.
**If it ships as written:** run-phase either writes a test that cannot pass against a compliant implementation, or silently picks the permissive reading and the criterion loses the precision D1's repair was for.
**Required fix (one clause):** either narrow half 1's `Given` to a **non-empty** configured value, or restate the seam as the single alignment-entry helper (the config read alone), so the assertion is "the alignment helper is entered once per `Handle`" rather than "the `origin/HEAD` read happens once".

**N3 — REQ-WBR-012 now mixes two GEARS patterns inside one requirement.** `spec.md:106` — Severity: **minor** — Class: **optional**.
The requirement is labelled Event-driven and opens correctly (`When 'moai doctor' runs, it shall report…`), but the D8 repair appended a Ubiquitous-form obligation (`The item's 'DiagnosticCheck.Name' shall be exactly…`) inside the same entry. This is formal-plus-formal, not formal-plus-informal, so it does not fail MP-2 — but it does make one id carry two modalities. Optional: split it, or leave it. It is enforceable either way (AC-WBR-009's manual step names the string).

**N4 — AC-WBR-013 check (2) carries a latent path-mixing bug.** `acceptance.md:240-245` — Severity: **minor** — Class: **optional (dormant for this SPEC)**.
The `[ -f … ]` guard reconstructs the *template-tree* path (`internal/template/templates/${b#internal/template/templates/}.tmpl`), but the subsequent `diff -q "$b" "$b.tmpl"` compares `$b` against a sibling in `$b`'s own tree. For a local-side wrapper (`.claude/hooks/moai/x.sh`) that sibling does not exist — I measured that `.claude/hooks/moai/` contains no `.tmpl` files at all, while `internal/template/templates/.claude/hooks/moai/` contains the `.sh`/`.sh.tmpl` pairs — so `diff -q` would error and the check would emit a false `DRIFT`. **Dormant here**: `plan.md` §D lists no hook wrapper, so check (2) enumerates nothing (verified in § B4). Fix if a later card touches a wrapper; not worth a cycle now.

**N5 — AC-WBR-002's first command still has a prose-only pass condition.** `acceptance.md:55-56` — Severity: **minor** — Class: **optional (carry-over, not introduced by iteration 2)**.
`# expect a line whose value is ""` is read by a human. Recorded for completeness because § B3's wbr4 fixture demonstrates the gap concretely; explicitly **not** a D5 regression, since iteration 1's whole-line form had the identical hole.

**Cosmetic, non-scoring:** HISTORY rows at `spec.md:26-28` run 0.1.0 → 0.3.0 → 0.2.0, out of order.

---

## Re-Checks Mandated by the Dispatch

- **All 7 must-pass criteria** — re-run against the 0.3.0 artifacts, not carried over. 7/7 clear; full evidence in § Must-Pass Results. Specifically checked for the two hazards named: no compound REQ lost its GEARS label (all 16 labels enumerated and matched), and no renumbering orphaned a reference (16 distinct AC ids across *all four* artifacts equals 16 distinct in `acceptance.md`, so no artifact cites an id that does not exist).
- **REQ→AC coverage, both directions** — measured. Forward: all 16 REQ ids appear in the `§D` matrix REQ column. Backward: all 16 AC ids trace to a REQ id (`AC-WBR-014` now to `013`, `AC-WBR-016` to `004`). Iteration 1's one uncovered REQ and one orphaned AC are both closed.
- **No operator decision re-opened** — checked all four individually. (a) **Both-consumers scope**: intact and now *strengthened* — REQ-WBR-011 binds both consumers to one predicate, and the primary-checkout narrowing explicitly states at `spec.md:84` and `acceptance.md:309` that it binds consumer 1 only and consumer 2 is unaffected, so the scope is narrowed in *firing location* without dropping either consumer. (b) **SessionStart + doctor surfacing**: intact — REQ-WBR-004 still fires from SessionStart (`plan.md:92`, fifth errgroup task), REQ-WBR-012 still carries the doctor item and now fixes its name. (c) **`TypeText` widget**: intact and untouched — `spec.md:110` (REQ-WBR-014) still mandates `TypeText`, still rejects `TypeSelect`/`TypeRadio`, `acceptance.md:168` still records the ruling as CLOSED, and `acceptance.md:344` re-states that M5 does not re-open it. (d) **t316 boundary**: intact — `spec.md:120-122` Out of Scope unchanged, `acceptance.md:345` retains the no-modification condition, and I confirmed no `tab_schema.json` appears in `git status`.
- **Scope discipline** — `git status --porcelain` in this worktree returns exactly two entries, both untracked directories: `?? .moai/reports/t313/` and `?? .moai/specs/SPEC-WORKTREE-BASEREF-001/`. No `tab_schema.json`, no source file, no config file, no template file. Clean.
- **Template neutrality survives D5's grep change** — § B3 and MP-4. It does.

---

## Debt Carried Into Run-Phase

Precise enough to carry knowingly. Ordered by when run-phase will hit it.

1. **N1 — fix AC-WBR-013 check (1) before running it (M1/M6).** The check will emit a false `NO-TEMPLATE-COUNTERPART` for `.moai/config/sections/git-strategy.yaml`. Do **not** interpret that line as a parity failure. Fix: extend the inner probe to accept the `.tmpl` counterpart and report only when neither the plain nor the `.tmpl` form was changed in the diff. Two lines. This is the one debt item that will actively mislead if left.
2. **N2 — pin AC-WBR-016 half 1's seam denotation before writing the test (M2).** Choose one: narrow half 1's `Given` to a non-empty configured value, **or** define the seam as the alignment-entry helper (config read alone) so "invoked once per `Handle`" holds for every value. Record which, in the criterion, so the run-phase test and the criterion agree. Do not resolve it by silently writing the permissive test.
3. **The seam obligation is not optional (M2/M3).** `plan.md:93` requires the read path be a function-variable seam and `acceptance.md:134-138` requires the resolvability helper be swappable through the seam mechanism at `internal/cli/session_worktree.go:48-53`. `internal/hook` carries no such seam today, so M2 must introduce one in its own new file. Budget for it; it is what makes AC-WBR-016 and AC-WBR-008's third assertion testable at all (§ B2).
4. **N4 — AC-WBR-013 check (2) is dormant, not correct.** It enumerates nothing for this SPEC. If scope ever grows to touch a hook wrapper, fix the path mixing first (`diff` compares against a sibling that only exists in the template tree).
5. **N5 — AC-WBR-002's first command is human-read.** If run-phase wants it mechanical, add an explicit value-emptiness assertion.
6. **N3 — REQ-WBR-012 carries two modalities.** Cosmetic; no run-phase action needed.
7. **The folding's traceability thinness (§ B1).** REQ-WBR-004's narrowing and REQ-WBR-013's preservation clause are each carried by one AC half or one shared AC. An id-level matrix check will not notice if AC-WBR-016 half 2 is dropped. Run-phase should treat both halves of AC-WBR-016 as separately mandatory — `acceptance.md:289` says so ("either alone leaves REQ-WBR-004 half-verified"); honour it.
8. **Inherited, unchanged from iteration 1** — G2 (`EnterWorktree`'s read of `origin/HEAD` is inferred, not read from source; the doctor item is the stated fallback), G4 (`moai doctor` never executed), G6 (rendered attribute order unmeasured; AC-WBR-011's two-branch condition must be collapsed in run-phase). All three are honestly disclosed in `plan.md` §B and none was regressed.

---

## Coverage Statement

What this audit actually checked, so an empty finding is distinguishable from an unlooked-at surface.

**Measured in this tree (commands run, output observed):**
- Toplevel / HEAD / branch re-read: worktree confirmed, `48eb945df`, `WT-worktree-baseref`.
- Distinct id counts: REQ 16 (across all four artifacts), AC 16 (in `acceptance.md`; also 16 across all four, so no dangling reference). REQ definition count in `spec.md`: 16.
- All 16 GEARS pattern labels enumerated from `spec.md` and matched against their clause forms.
- `§D` matrix REQ column extracted row by row → forward and backward coverage both complete.
- `grep -c 'syscall'` → 0 in each artifact; `grep -rn '\[NEEDS CLARIFICATION'` → rc=1.
- Frontmatter fields read at `spec.md:2-13` → 12/12 canonical, no rejected alias.
- Referenced-SPEC statuses: `completed` (BRANCH-GUARD-001), `in-progress` (SYNC-STRATEGY-KEY-001).
- `git ls-files --error-unmatch .moai/config/sections/git-strategy.yaml` → rc 0 (D4(a)'s premise, independently reconfirmed).
- `git merge-base HEAD origin/develop` → `48eb945df606eea7d6d3d1b9a1020adbfe79b2e6` (D9 runnability).
- `git diff --name-only <BASE>..HEAD -- .claude .moai` → empty (D9 empty-diff decidability).
- AC-WBR-002's `sed`/`cut` pipeline re-run against **five** fixtures (§ B3), full output quoted.
- `internal/template/templates/.moai/config/sections/` listed → only `git-strategy.yaml.tmpl` exists (N1).
- `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` → exists as a plain file (the counterpart that *does* work).
- `.claude/hooks/moai/` listed → no `.tmpl` files present locally (N4).
- `internal/hook/*.go` scanned for function-variable seams → one computed var, no injection seam (B2).
- `git status --porcelain` → two untracked directories only.

**Citations verified against source (read, line numbers confirmed):** `internal/cli/session_worktree.go:234-241` (`inGitWorktreeReal`, `--git-dir` vs `--git-common-dir` — as cited by REQ-WBR-004 and D3.1); `internal/cli/session_worktree.go:48-53` (the seam block and its swap-and-restore comment — as cited by AC-WBR-008's third assertion); `internal/hook/session_start.go:176-181` (the best-effort contract sentence begins at `:176` — D7's corrected citation is right); `internal/cli/doctor.go:232` (exact-name filter — D8's justification is right); `internal/settings/schema_sections.go:160-177` (`gitStrategyFields()` opens at `:160` and returns at `:177` — as cited by REQ-WBR-013 and AC-WBR-012's mutation step); `spec-workflow.md:146-152` (the Tier ceiling table and the independence sentence — B1's foundation).

Every citation checked resolved to what the artifact claimed. Iteration 1's single citation-drift finding (D7) is closed and no new drift was found.

**NOT checked (gaps in this audit):**
- No Go test, build, `vet`, or lint was run. Plan-phase audit; no implementation exists. In particular I did **not** re-run iteration 1's vacuous-`-run` probe, because the §D.3 repair is a documentary obligation whose correctness I judged by reading, not by re-measuring the failure it forbids.
- `moai doctor` was not executed (the SPEC's own G4 records the same gap); the `--check` flag's filter was read, not run.
- AC-WBR-013's two pipelines were **not** executed end to end against a populated diff — no implementation diff exists. N1 was derived by tracing the pipeline against measured filesystem facts (the `.tmpl`-only counterpart) plus `plan.md`'s own write list, not by observing the false line. I regard that as sufficient to state N1 as a defect, and I flag the distinction rather than hide it.
- `EnterWorktree`'s actual read of `origin/HEAD` was not verified from source (G2 — Claude Code is not in this repository). The SPEC's premise that consumer 1 has any effect still rests on inference, honestly disclosed.
- The rendered attribute order of a `TypeText` control was not measured (G6, unchanged).
- I did not re-audit surfaces iteration 1 cleared and iteration 2 did not touch, except where a repair could plausibly have regressed them (MP-4 and the four operator decisions, both explicitly re-checked).

**Residual risk in this verdict:** the design survived a second pass unchanged, and I again found no reason to doubt it. The two debt items are both *criterion* defects — a check that misfires and a criterion that is ambiguous — not defects of requirement or of design. The specific risk a reader should hold: N1 will surface as a red MUST criterion during M1, and the wrong response to it (treating `NO-TEMPLATE-COUNTERPART` output as noise) is more damaging than the false line itself. That is why it is item 1 of the debt list.

---

## Recommendation

**PASS-WITH-DEBT.** The SPEC proceeds to Implementation Kickoff Approval. Iteration 2 repaired eleven of twelve findings at the layer each was raised against, cleared the must-pass firewall on re-verification, and moved the binding harmonic aggregate from 0.78 to 0.92 — comfortably past the Tier M threshold rather than scraping it. Two of the repairs (D3's guard-existence-plus-mutation, A5's "even when their runtime behaviour agrees") are stronger than what the verdict asked for.

**Why not a clean PASS.** N1 is a criterion that fails a correct implementation of the SPEC's own plan, and N2 is an ambiguity inside the repair that closed iteration 1's central traceability gap. Both are blocking-class findings under the finding-consumption discipline, and neither should reach run-phase unrecorded.

**Why not a FAIL.** Neither is a must-pass failure; the aggregate clears on the binding mean by a wide margin; and both are bounded, named, two-line edits rather than defects of design or of requirement. Failing the final iteration over them would escalate to the operator a decision the operator has no better information to make than run-phase does — the fixes are already written out above. Under the ceiling rule a FAIL here escalates rather than loops, and escalation is the disproportionate response to a check that needs one `.tmpl` case added.

**Would a narrow out-of-cycle fix close the gap to a clean PASS?** Yes, and it is worth stating in case the orchestrator prefers that route: a single out-of-cycle edit to `acceptance.md` — the N1 probe change at `:229-236` and the N2 `Given` narrowing at `:292` — would close both. That edit touches one artifact, re-opens no operator decision, adds no id, and stays inside the 16/16 ceiling. If the orchestrator elects it, no re-audit is needed for the fix itself: both changes are mechanically checkable against the two statements in § B4 and § N2, and the verdict above is unchanged by them except that the debt list loses its first two items and Testability would sit at the rubric ceiling.

**Stagnation assessment:** not applicable in the adverse sense. No defect appeared unchanged across both iterations; every iteration-1 finding was addressed and eleven were closed. The one PARTIAL (D9) was addressed in substance — the judgement steps that were the finding are gone — and failed only on an implementation detail of its replacement. This is progress, not a reviser that did not understand the finding.

**Tier M ceiling reached.** No further revision cycle remains within this audit contract.

