# SPEC Review Report: SPEC-CLI-WORKTREE-FLAG-RACE-001
Iteration: 1/2 (Tier M ceiling = 2, per `.moai/config/sections/harness.yaml` `plan_audit_tier_ceilings.M`)
Auditor tree: `.claude/worktrees/t464` @ `d592b0551`
Verdict: **PASS-WITH-DEBT**
Overall Score: **0.85**
Tier-M PASS threshold applied: **0.80** (SSOT: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier M `0.80`)

Reasoning context ignored per M1 Context Isolation. The lane's ground-truth items supplied in the
dispatch were treated as claims to be re-verified, not as author reasoning; every one of them was
re-measured in this tree before use, and one was found over-broad (see D1).

---

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency** — `REQ-WFR-001` … `REQ-WFR-006` at spec.md:103,107,111,115,119,125. Sequential, no gaps, no duplicates, uniform 3-digit padding.
- **[PASS] MP-2 GEARS format compliance** — judged against the **requirement layer** (`spec.md` §B) only, per M3 § Scope. 001 Ubiquitous ("The `internal/cli` test package shall contain no data race…"), 002 Event-driven ("**When** `go test …` is executed …, the run shall exit `0`…"), 003/004 Unwanted-negative ("The repair shall not change…" / "shall not reduce…"), 005 Where-gate ("**Where** the run-phase implementer selects the injection option, the production-side change shall be confined to…"), 006 Event-driven ("**When** the repair lands, a `-race` repetition run shall exist…"). 6/6 match a GEARS pattern. The Given-When-Then entries in `acceptance.md` are the **verification layer** and were graded under Group 4, never here.
- **[PASS] MP-3 YAML frontmatter validity** — all 12 canonical fields present with correct types at spec.md:2-13 (`id`, `title`, `version: "0.1.0"` quoted semver, `status: draft`, `created`/`updated` ISO dates, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` comma-separated string) plus `tier: M`. Zero rejected snake_case aliases (`created_at`/`updated_at`/`labels`/`spec_id`) present.
- **[N/A] MP-4 language neutrality** — single-language SPEC (Go, `internal/cli`). No multi-language tooling surface. Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation** — one external reference: `SPEC-TEMPDIR-CLEANUP-RACE-001`. `.moai/specs/SPEC-TEMPDIR-CLEANUP-RACE-001/spec.md` exists; `status: completed` ∉ {retired, superseded, archived}. No reconciliation obligation, no BLOCKING finding.
- **[PASS] MP-6 D8 cross-platform discipline** — `grep -c syscall spec.md plan.md acceptance.md` → `0/0/0`. D8-4 auto-PASS.
- **[PASS] MP-7 clarification gate** — `grep -rn '\[NEEDS CLARIFICATION' plan.md` → no matches. `research.md` absent (correct for Tier M).

No must-pass failure. The M5 firewall does not force FAIL.

---

## Category Scores (rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|---|---|---|---|
| Clarity | 0.85 | requirement layer 1.0 / context layer 0.75 | Requirement layer is unambiguous and single-interpretation (spec.md:103-127); two internal contradictions live in the context/risk layer — spec.md:51 vs spec.md:77-79 (D1) and spec.md:139-142 + :245 vs spec.md:86-89 (D2). A reasonable engineer resolves both consistently in favour of the correct reading, so the deduction is one band on the context layer only. |
| Completeness | 0.90 | 1.0 band, one deduction | HISTORY (:21), Context/WHY (§A), WHAT (§B), HOW (`plan.md` §F), REQUIREMENTS (§B), ACCEPTANCE CRITERIA (`acceptance.md` §D.1), Out of Scope (five `### Out of Scope — <topic>` H3s at spec.md:136,148,153,159,165, each with `-` bullets). Frontmatter 12/12. Tier M artifact set complete (spec+plan+acceptance, + progress). Deduction: the tier claim carries no LOC/file measurement, which is the SSOT table's stated determinant (D5). |
| Testability | 0.85 | 0.75-1.0 | Zero weasel words (`grep -nEi 'appropriate\|adequate\|reasonable\|proper\|as needed\|if necessary'` over all three artifacts → no matches). AC-WFR-001b (acceptance.md:56-70) is exemplary: dual independent signal, plus the `grep -c` inverted-exit-code caveat spelled out. AC-WFR-003 carries a mandatory positive control (:110-112). Deductions: no vacuous-green control on AC-WFR-001b (D4); AC-WFR-006's "the entry was written before the first code edit" (:153) states no verification verb and is not mechanically checkable after the fact (D7). |
| Traceability | 0.80 | 0.75 band ("one mapping is indirect") | 6/6 REQs covered, 7/7 ACs traced to existing REQs, no orphans (acceptance.md:164-171 verified row by row against spec.md §B). One false mapping: REQ-WFR-003 ← AC-WFR-005 (D3). |

Arithmetic mean = (0.85 + 0.90 + 0.85 + 0.80) / 4 = **0.85** ≥ 0.80. Score gate passes.

---

## Verified-correct findings (recorded so a re-audit does not re-open them)

These were pressed adversarially and held:

- **(a) The [HARD] "full suite green is NOT repair evidence" constraint is BINDING, not decorative.** It appears three times, and critically **inside the AC body**, not only in the constraints section: acceptance.md:72-74 — "A green run of the full suite, or of this package at `-count=1`, does **NOT** satisfy this AC and MUST NOT be substituted for it." AC-WFR-004, the only AC a package-wide run could discharge, disclaims repair-evidence status in its own body (acceptance.md:125-127: "This is a **non-regression** check, not repair evidence… The repair evidence is AC-WFR-001b"). I searched for an AC whose text a suite run could satisfy; there is none.
- **(b) The GREEN AC verifies BOTH signals independently.** acceptance.md:56-70 requires `exit 0` **and** a printed `grep -c 'WARNING: DATA RACE'` of exactly `0`, with the two-command recipe supplied and the `grep -c` exit-code inversion explicitly called out ("read the printed count, not the exit status of the grep"). Neither an exit-code-only nor a grep-only reading satisfies it as written.
- **(c) Neither repair option is pre-picked; no default is smuggled in.** spec.md:188-190 — "declares no winner. A run-phase implementer who finds this section already decided should treat that as drift and stop." plan.md:25-26 and plan.md:36 close the two obvious smuggling routes: the plan restates no option, and "The default is not 'Option A because it is smaller'." spec.md:204 declares the criteria table deliberately unweighted. AC-WFR-002 (:89-91) and AC-WFR-003 (:102-106) are both written option-conditionally rather than assuming one. One residual tilt is recorded as D6 (MINOR).
- **(f) Traceability 6/6 is real, not asserted.** Verified per-row, both directions. Forward: 001→AC-001b; 002→AC-001a,001b; 003→AC-004,005; 004→AC-002; 005→AC-003; 006→AC-001a,006 — every REQ has ≥1 AC. Reverse: all seven ACs name a REQ that exists in spec.md §B; zero orphans. The quality of one row is a separate finding (D3).
- **Numeric claims re-measured and correct.** `grep -rl 'findProjectRootFn = \|launcherWorktreeMaterialize = ' internal/cli/*_test.go | wc -l` → **23** (spec.md:138 ✓). Per-file: `cc_test.go` 18, `mx_query_test.go` 24, `coverage_improvement_test.go` 14 (spec.md:139-140 ✓ exact). Production references → **26**, satisfying "24+" (spec.md:140 ✓). Four siblings at lines **65 / 95 / 125 / 160** each with `t.Parallel()` on the following line and `t.Cleanup` restores (spec.md:61-64 ✓ exact — the card's "three" is correctly superseded, and `_NoFlagIsNoop`'s frames 67/68/72 do appear in the RED output). Production seam `worktree_branch_flag.go:50` (`var launcherWorktreeMaterialize`), `:70` (`findProjectRootFn()`), `:74` (`launcherWorktreeMaterialize(...)`) — all three line numbers verified by reading the file (spec.md:77-79 ✓). RED file: `wc -l` → **1369**, `grep -c 'WARNING: DATA RACE'` → **23** (spec.md:45-47 ✓). `git merge-base --is-ancestor a05b9c4d8 origin/develop` → exit `0` (spec.md:34-35 ✓).
- **(e) Tier M is defensible, and is not tier-inflation.** Tier M raises the PASS bar from 0.75 to 0.80; the author chose the stricter gate, so the classification does not buy a lower bar. See D5 for what is missing from the justification.

---

## Defects Found

**D1 — `spec.md:51` — "No frame lands outside that file" is false, and contradicts `spec.md:77-79` three paragraphs later. — Severity: SHOULD-FIX — Class: blocking (internal consistency).**
Measured in this tree: `grep -o 'internal/cli/[a-z_]*\.go' .moai/reports/t464/red-race-d592b0551.txt | sort | uniq -c` → `52 worktree_branch_flag_test.go`, **`5 worktree_branch_flag.go`**, `46 main_test.go`. The production-file frames are race participants, not stack decoration — report line 289-291 reads `Previous read at 0x…b0c8 by goroutine 28: resolveWorktreeExistingBranch() … worktree_branch_flag.go:70`, and line 918-920 reads `Read at 0x…b508 by goroutine 46: resolveWorktreeExistingBranch() … worktree_branch_flag.go:74`. `main_test.go:230` appears in the `Goroutine N created at:` stacks (TestMain). The claim is a false verification claim under `AGENTS.md` §1, on a SPEC whose own centerpiece is evidence discipline. It is also self-contradicting: §A.4 (:77-79) correctly states the production function reads both globals at exactly those two lines. Separately, the line enumeration at :50-51 omits two test-file lines that do appear inside `WARNING: DATA RACE` blocks — `:69` and `:175` (the latter 9 times, e.g. report lines 519 and 587).
*Required fix*: replace the absolute sentence with the measured one — "every **test-side** race frame lands in `worktree_branch_flag_test.go` (lines 67-69, 72, 98-105, 128-134, 162-175); the **read** side lands in `worktree_branch_flag.go:70` and `:74`, which is the production function under test; `main_test.go:230` appears only in goroutine-creation stacks."

**D2 — `spec.md:139-142` and `spec.md:245` (F1) — the residual-risk claim overstates what was left unobserved, and contradicts `spec.md:86-89`. — Severity: SHOULD-FIX — Class: blocking (internal consistency).**
§C asserts "Whether any of those other files also combine assignment with `t.Parallel()` is **not established by this card's measurement**", and F1 asserts "Their status is unknown, not clean." But §A.5 rejects `coverage_test.go:621` and `launcher_test.go:793` as *false positives* — candidates that can only have been produced by a package-wide scan. The SPEC consumes that scan's output while denying the scan exists.
I re-ran the scan independently, function-body-bounded over **all** `internal/cli/*_test.go` (`awk` resetting at each top-level `}`, so no doc-comment bleed into the following function — the known weakness of the lane's version):
```
internal/cli/worktree_branch_flag_test.go:65   _NoFlagIsNoop
internal/cli/worktree_branch_flag_test.go:95   _WiresAndStrips
internal/cli/worktree_branch_flag_test.go:125  _RejectsBadUsage
internal/cli/worktree_branch_flag_test.go:160  _MaterializeErrorPropagates
```
Exactly the four in-file siblings, **zero** out-of-file hits, and zero false positives. Under-count risk was checked and closed: every non-`= ` mention of either global in the package (97 lines) is an `orig := <global>` save or a comment — there is no helper-mediated assignment form the pattern would miss.
This does **not** make a package sweep in scope, and the exclusion itself is correct. What is wrong is the *epistemic framing*: a measurement exists and found nothing, which is materially different from "unknown."
*Required fix*: keep both Out-of-Scope entries; restate F1 as "a function-body scan over all `internal/cli/*_test.go` returned only the four in-file siblings; the residual risk is that the scan matches a literal `<global> = ` assignment and would miss a helper-mediated one" — and cite the scan in §A.5 as the source of the two rejected candidates.

**D3 — `acceptance.md:20` and `acceptance.md:168` — REQ-WFR-003 (behaviour preservation) is traced to AC-WFR-005 (cross-compile `go vet`), which by the AC's own admission establishes nothing about behaviour. — Severity: SHOULD-FIX — Class: blocking (a criterion this document states).**
acceptance.md:144-146 says of AC-WFR-005: "`go vet` establishes that the package and its tests compile on those targets. It does **not** run them." A compile check cannot verify that `--branch` token stripping, error propagation, and the no-flag no-op path are unchanged. The remaining mapped AC (AC-WFR-004) asserts only `exit 0` for the whole package — a coarse proxy. The ACs that actually exercise REQ-WFR-003's three named behaviours are AC-WFR-001b (runs the four scenarios ×20) and AC-WFR-002 (their assertions survive), and neither is cited in REQ-003's row. This matters most under Option B, where a production signature change makes behaviour preservation the principal risk.
*Required fix*: in acceptance.md §D.3, map REQ-WFR-003 → AC-WFR-001b, AC-WFR-002, AC-WFR-004; drop AC-WFR-005 from that row (leave AC-005 as a SHOULD-severity portability check tracing to nothing, or introduce a portability REQ for it).

**D4 — `acceptance.md:47-74` (AC-WFR-001b) — no positive control against a vacuous green. — Severity: MINOR — Class: optional.**
`exit 0` + `0` race warnings is also the exact signature of a run in which the `-run` selector matched nothing. This repository's own lesson base names that failure mode. It is *partially* mitigated by construction — AC-001a proves the selector matches on the identical command string (bolded "**same**" at :50), and AC-WFR-002 forbids renaming or removing any of the four functions — so it is not an open hole, only an unclosed one.
*Required fix*: add one line to the GREEN recipe — run with `-v` and assert `grep -c '^--- PASS: TestResolveWorktreeExistingBranch'` returns the expected non-zero count (4 tests × 20 iterations, plus subtests), so a zero-match selector cannot present as a repair.

**D5 — `plan.md:50-59`, `spec.md:225` — the Tier M claim never states the SSOT's own determinant. — Severity: MINOR — Class: optional.**
`spec-workflow.md` § SPEC Complexity Tier classifies by LOC and files affected (S: `<300 LOC`, `<5 files`). This change is 4 deleted lines across 1 file (Option A) or a signature change plus four test rewrites across 2 files (Option B) — Tier S by that table. plan.md:52 is honest that its reasons are "not about line count", but does not acknowledge that line count *is* the table's criterion, so the departure reads as an oversight rather than a decision. The classification is defensible (Tier M raises the bar to 0.80 and buys a second audit iteration; it does not lower any gate), and I applied the 0.80 threshold accordingly.
*Required fix*: add the measurement and the explicit departure — "measured scope is Tier S by LOC/files (1-2 files, ≤4 LOC under Option A); Tier M is claimed deliberately because …", so a reader sees a decision rather than a miscount.

**D6 — `spec.md:207-213` (§D.3 criteria table) — the unweighted table reads with a soft tilt toward Option A. — Severity: MINOR — Class: optional.**
Three of five "Favours A" cells are unhedged and concrete ("Nothing is lost that the suite depends on", "trivially reviewable", "stays inside that idiom"), while the matching B cells hedge ("either a bridgehead for a later sweep or an inconsistency, depending on whether that sweep is ever funded"). The prohibition on declaring a winner (:188-190) is respected in letter; the rhetorical balance is not. This does not rise to smuggling a default — plan.md:36 explicitly forbids the "A because it is smaller" default — but a run-phase implementer skimming the table absorbs a lean.
*Required fix*: match the hedging level across each row's two cells, or state the tilt openly ("A reads stronger on three rows; that is not a weighting").

**D7 — `acceptance.md:153` (AC-WFR-006) — "the entry was written before the first code edit" is asserted with no verification verb. — Severity: MINOR — Class: optional.**
Every other AC supplies the command a tester runs. This one requires establishing a temporal ordering after the fact, with no mechanism named.
*Required fix*: supply the verb — e.g. require the `progress.md` §E.2 entry to land as its own commit preceding the first code commit, verifiable with `git log --oneline -- .moai/specs/…/progress.md internal/cli/`.

**D8 — `spec.md:44-47`, `spec.md:247` (F3) — the RED run's early termination is not recorded. — Severity: MINOR — Class: optional.**
The captured RED does not complete 20 iterations: it ends in `panic: Log in goroutine after TestResolveWorktreeExistingBranch_NoFlagIsNoop has completed` followed by `FAIL github.com/modu-ai/moai-adk/internal/cli 1.543s` (report tail). So "23 warnings at `-count=20`" is a lower bound from a truncated run, and the GREEN run — which will complete all 20 — executes strictly more work than the RED did. That asymmetry is conservative and harmless, but it is unrecorded, and the second observable failure mode (the post-completion-log panic, itself a symptom of the same stale-stub mechanism) is never named.
*Required fix*: record in §A.2 that the RED aborted on a panic before exhausting `-count=20`, name the panic as a second manifestation of the same cause, and note in F3 that the RED warning count is therefore a floor.

---

## Recommendation

**Proceed to run-phase with debt.** All seven must-pass criteria pass with cited evidence, the
aggregate 0.85 clears the Tier M threshold of 0.80, and the three items the dispatch asked me to
press hardest — the binding force of the "full suite is not evidence" constraint (a), the dual-signal
GREEN (b), and the absence of a pre-picked repair option (c) — all held under adversarial reading.
The SPEC's measured claims are, with one exception, accurate to the byte.

The debt is three SHOULD-FIX items, all text-only edits to `spec.md` / `acceptance.md`, none of which
requires re-deriving evidence:

1. **D1** — correct `spec.md:51` to match the measured frame locality (the production file *is* in the race, at `:70` and `:74`).
2. **D2** — restate `spec.md` §C / F1 as "scanned, nothing found, here is the scan's weakness" instead of "unknown". As written, F1 is the mirror-image of the failure this repository's doctrine forbids: an unattributed claim, made in the pessimistic direction.
3. **D3** — repair the REQ-WFR-003 row in `acceptance.md` §D.3 so behaviour preservation traces to the ACs that actually exercise behaviour.

D4-D8 are optional; route them at the orchestrator's discretion. D4 is the one I would take anyway —
it is a single added line and it closes the vacuous-green hole outright rather than leaving it
mitigated-by-construction.

Nothing here blocks the option decision, which correctly remains run-phase's to make.
