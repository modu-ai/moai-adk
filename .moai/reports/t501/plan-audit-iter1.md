# Plan-Phase Audit Report — SPEC-CODEX-TEST-GAPS-001 (card t501)

Iteration: 1/2 (Tier M ceiling per `harness.yaml` `plan_audit_tier_ceilings` — see D1: declared Tier S is refuted below; correct tier is M)
Verdict: **FAIL**
Overall Score: **0.875** (mean of 4 dimensions; clears both the Tier M 0.80 and Tier S 0.75 thresholds — the FAIL rests on the blocking-defect route, not threshold arithmetic; see Verdict Basis)

Auditor: plan-auditor (independent). Baseline attribution: all file reads, greps, and measurements below were taken in this run against worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t501`, branch `WT-codex-uncovered`, HEAD `30bbb1753` (plan-phase commit), parent `ace1c5440` (origin/develop base, matches plan.md §C.1).

Reasoning context: the dispatching card context (premise-refutation narrative) was treated as unverified claims per M1 Context Isolation; every assertion was re-verified from the artifacts, the committed raw evidence, and the production source.

Cross-model convergence (audit_multi, `project_root` = this worktree): claude anchor (this auditor) → initially PASS, revised to FAIL after independently verifying the codex findings; codex (required gate) → **FAIL** with 5 findings, all 5 adjudicated below (4 accepted, 1 accepted as advisory); glm → inconclusive (backend unavailable, fail-open, no effect on verdict). Note: the moai MCP server binary (e79c010b8) is an ancestor of HEAD — its catalogue may be stale; all load-bearing evidence was read directly via Read/Grep/Bash, not via the MCP catalogue.

---

## Must-Pass Results

- [PASS] **MP-1 REQ number consistency**: REQ-CTG-001..010 (spec.md:55-95) — sequential, zero-padded, no gaps, no duplicates. Verified by direct read.
- [PASS] **MP-2 GEARS format compliance (requirement layer)**: all 10 REQ entries match GEARS patterns — REQ-CTG-001 compound `[When][While]` + `[Where]` (spec.md:57-59), 002-007/010 `[When]`, 008 Ubiquitous ("The SPEC artifact itself … shall", spec.md:87), 009 `[While]` (spec.md:91). No informal language, no Given/When/Then content in the requirement layer. Judged against spec.md §C only.
- [PASS] **MP-3 YAML frontmatter validity**: all 12 canonical fields present with correct types (spec.md:2-15); `created:`/`updated:`/`tags:`/`id:` canonical (no snake_case aliases); `status: draft` valid enum; `phase: "v3.1.4 target"` is a release target (not a prohibited stage value); `era: V3R6` and `tier: S` are valid optional fields. `related_specs` is an extra field outside the optional catalogue — tolerated, noted only. Corroborated by codex: `moai spec lint --strict` → "No findings".
- [PASS/N/A] **MP-4 Section 22 language neutrality**: N/A — single-language SPEC (Go tests in `internal/cli`); no multi-language tooling surface.
- [PASS] **MP-5 D7 cross-SPEC reconciliation**: references extracted — SPEC-CODEX-WIRING-001 (×2), SPEC-CODEX-SESSION-MSG-001 (×2), self ×6. Both referenced SPECs exist with `status: completed` — no retired/superseded/archived reference, no reconciliation owed. No BLOCKING finding.
- [PASS] **MP-6 D8 cross-platform discipline**: literal `syscall` appears 0 times across spec.md/plan.md/acceptance.md — auto-PASS per D8-4. (The production comment at codex_job_control.go:92 mentioning `syscall.SIGTERM` is outside the SPEC body and is itself the documented rationale for choosing `os.Process.Kill` — no build-tag hazard is introduced by the SPEC.)
- [PASS] **MP-7 clarification gate**: `grep -n 'NEEDS CLARIFICATION'` across plan.md/spec.md/acceptance.md/progress.md → 0 matches (rc=1). No unresolved markers.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.0 | full | Every REQ single-interpretation with exact `file:line` anchors; all 8 distinct anchor sites verified to resolve (codex_job_control.go:94; mcp_codex.go:494, :602-603, :888, :912; codex_contract.go:186-194; codex_init.go:56-58, :136-138). AC Given/When/Then columns all populated and measurable. |
| Completeness | 1.0 | full | HISTORY (§A), context/WHY (§B with both measurement axes), REQUIREMENTS (§C, 10 REQs), AC layer (acceptance.md §D), skip record (§D table, 3 rows), constraints (§E), Out of Scope with four `### Out of Scope — <topic>` H3 headings each carrying `-` bullets (§F), cross-refs (§G). Frontmatter complete. Lettered-section layout is the repo convention; content-complete. |
| Testability | 0.75 | one-band-down | AC Then-columns are binary, but the verification verb of AC-CTG-001..007 (bare `go test -run <Selector> … rc=0`) cannot distinguish executed-pass from selector-zero-match pass — the repo's own named hazard (test-selector-zero-match lesson family; verification-completeness §1.1 "report-not-verdict"). Codex verified empirically on this tree: the selector run prints `[no tests to run]`, `PASS`, exit 0 today. 7 of 11 ACs share the defect → D2. |
| Traceability | 0.75 | one-band-down | REQ→AC complete (10/10 covered). AC→REQ complete for AC-CTG-001..010. AC-CTG-011 anchors to "spec.md §E (Constraints 2, 3)" — no REQ-XXX anchor exists for it (orphan AC) → D5. |

## Tier Judgment (dispatch question 3)

**The correct tier is M, not the declared S.** Verified mechanically:

- Authored governed plan artifacts: `spec.md` + `plan.md` + `acceptance.md` = 3 files — the Tier M artifact set (Tier S = 2 files with ACs inline in spec.md §3; acceptance.md would not exist).
- REQ count 10 > Tier S ceiling 8; AC count 11 > Tier S ceiling 8. Both fit Tier M ceilings (16/16). Per `spec-workflow.md` § SPEC Complexity Tier: "Exceeding either ceiling is a signal to tier up or to split the SPEC, not to relax the budget."
- Consequences if left unfixed: audit PASS threshold (0.75 vs 0.80), retry ceiling (S=1 vs M=2), delegation-template applicability, and the ownership-matrix commit-subject contract (`{tier}, {N} artifacts` — the landed commit `30bbb1753` carries neither).

This audit scores against **both** thresholds: aggregate 0.875 clears Tier M 0.80 and Tier S 0.75 either way. The tier reclassification is D1 (must-fix).

## Evidence Verification (dispatch question 1 — refuted-premise narrowing)

Independently re-verified against the committed raw evidence (`.moai/reports/t501/`):

- `coverage-run.log` — verbatim `ok github.com/modu-ai/moai-adk/internal/cli 515.612s coverage: 80.7% of statements` + `rc=0`. Matches spec.md §B.2 exactly.
- `coverage-perfunc.txt` — 118 function rows across exactly the 6 target files (codex_contract/init/job_control/launcher/doctor_codex/mcp_codex `.go`). Distribution: **exactly 4 at 0.0%** (`terminateCodexProcess` :94, `pid` :494, `Error` :602, `Unwrap` :603), exactly 1 non-zero below 60% (`codexIDMatches` 55.6), next band 60.0/66.7. Matches §B.2's measured residue.
- All 16 card-named function percentages in §B.2 match the extract value-for-value (80.0, 92.9, 75.0, 0.0, 85.7, 100.0×4, 100.0, 100.0, 91.7, 100.0, 77.3, 100.0, 100.0, 75.0).
- `namegrep-counts.txt` — 114 rows, 62 with count 0 → "62 of 114" confirmed **at name granularity** (see D7 for the receiver-collapse caveat codex raised; `pid` at count 45 aggregates mcp_codex.go:494 pid 0.0% with :701 pid 66.7% — both present in the extract).
- The seam-var shadowing claim (REQ-CTG-001's premise) verified: `codex_job_control_test.go:136-143` swaps `codexTerminateProcess`, so the real body at :94 never runs via those tests. Sound.
- The 7 scoped items map onto the measured gaps completely: the 0.0% set → REQ-001 (terminate), REQ-007 (Error+Unwrap), §D skip (pid); the <60% set → REQ-002 (codexIDMatches); card-cluster arm residue → REQ-003/004/005/006 (functions at 92.9/77.3/75.0/66.7%, each explicitly named in the REQs and therefore inside §F's exclusion carve-out). Everything else dropped is documented: §F clause 1 (all >60% unnamed functions — this covers `defaultLoginStatusRunner` 60.0% and the 66.7% band), §B.3 (behaviorally-covered idioms — `fakeCodexConn`, `codexTestExecImports` at codex_init_test.go:962, AC-CX2-017 arms ×6 in codex_protocol_liveness_test.go all verified present), §D (3 skip surfaces). **No measured-but-unscoped surface is dropped silently.** The narrowing is sound as to completeness.
- 6-file measurement closure: `internal/cli` contains 12 codex production files; the 6 unmeasured ones (codex_jobs/readiness/review_gate/task/hook_harness_codex/update_codex_wiring) were never in the card's cluster claim and the SPEC explicitly scopes to "the 6 target files" (§B.1) — explicit scope, not a silent drop. Not a defect.
- Progress §E.1 (dispatch question 5): present, parser-safe §E.1 heading, `status: draft` consistent with frontmatter, worktree/branch/base verified true (branch `WT-codex-uncovered`, base `ace1c5440` = HEAD~1), evidence paths all resolve, id-check regex `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` matches the 3-segment domain. `tier: S` at progress.md:9 mirrors the D1 defect. `premise_verification: "8/8 named functions located"` — all 8 distinct citation sites verified by this audit.

Dispatch question 4 (M1-M7 ordering and constraints): verified compliant. M1=[U3], M2-M3=[U2] — the U2·U3 clusters occupy M1-M3 per the lead's directive, with the least-reversible decision (cross-platform strategy: helper-process re-exec, rationale recorded at plan.md §F M1) locked first. Priority descends High→Medium→Low with the completion re-measurement (M7, High) correctly sequenced last. Tests-only (REQ-009 + plan §D/§G), cleanup-before-first-assertion (REQ-001 + plan M1 + §G), scoped ≥600s budget (spec §E.4 + plan §B), no local `go test ./...` (spec §E.4 + plan §D/§G) — all present. One arithmetic defect inside the recipe → D6.

## Defects Found

- **D1** — tier-contract misdeclaration — spec.md:14 (also plan.md:3, progress.md:9, commit 30bbb1753 subject) — `tier: S` contradicts the authored 3-artifact set and breaches both Tier S budgets (10 REQ > 8; 11 AC > 8). Correct tier: **M**. — Severity: **must-fix** — Class: **blocking** — Required fix: set `tier: M` in spec.md frontmatter and progress.md §E.1.
- **D2** — vacuous-green AC verification verb (AC-CTG-001..007) — acceptance.md:9-15 (inherited by plan.md:43 E1) — bare `go test -run <Selector> ./internal/cli/ rc=0` returns 0 with `[no tests to run]` on the current tree (codex-executed probe, observed on this baseline); rc=0 cannot distinguish "test exists and passed" from "selector matched nothing". — Severity: **must-fix** — Class: **blocking** — Required fix: verification verb must be `-count=1 -v` plus an assertion that the named test's `--- PASS: Test<Name>` line was observed (or a `go test -json` run-event check); record today's zero-match output as the RED-now cell so the gate has an observed failure baseline.
- **D3** — production-diff gate blind to committed/staged/untracked scope (AC-CTG-009, plan §E.4 E4) — acceptance.md:17, plan.md:46 — `git diff --name-only` reads unstaged tracked edits only; a production `.go` change committed before the completion check is invisible to the gate (codex measured 7 committed non-test paths on this branch — all documentation, but demonstrating the blindness). — Severity: **must-fix** — Class: **blocking** — Required fix: pin the run-phase base SHA and gate on the union of `base..HEAD` + staged + unstaged + untracked paths, filtered to `.go$`, asserting every path matches `internal/cli/**/*_test.go` (SPEC/progress/evidence allowlisted explicitly).
- **D4** — skip-record rationale factually refuted (`(realCodexConn).pid`) — spec.md:103 (asserted by AC-CTG-008 acceptance.md:16 and AC-CTG-010 acceptance.md:18) — §D claims the function is "not reachable from hermetic unit tests". Verified false by inspection: `pid()` (mcp_codex.go:494-500) is a 3-branch field read; `realCodexConn` is same-package constructible (struct at :455) and `os.Process.Pid` is an exported field — `&realCodexConn{}` alone reaches the first branch hermetically; codex additionally demonstrated a passing overlay test. REQ-CTG-008's deliverable (the record) therefore carries a false reason, and REQ-CTG-010 preserves the package's last 0.0% function on it. — Severity: **must-fix** — Class: **blocking** — Required fix: either (a) add the trivial 3-branch hermetic test as an 8th item and retarget REQ-CTG-010/AC-CTG-010 to "zero remaining 0.0% functions", or (b) rewrite the §D row as an explicit low-value decision (a field-read method deliberately skipped), never "unreachable".
- **D5** — orphan AC anchor (AC-CTG-011) — acceptance.md:19 — anchors to spec.md §E Constraints 2-3; no REQ-XXX exists for vet/lint/gofmt cleanliness. — Severity: **should-fix** — Class: **blocking** — Required fix: add REQ-CTG-011 (cleanliness; legal at Tier M budget) or demote AC-CTG-011 into plan §E.2's E2 verification item, leaving 10 ACs.
- **D6** — verification timeout arithmetic — plan.md:39-42 — go-test `-timeout 700s` exceeds the Bash tool's stated 600,000 ms ceiling in the same section, so the go-level hang-guard can never fire before the wrapper kills the command. Mitigated by the measured 515.6 s and the `-run` split path. — Severity: **advisory** — Class: optional — Required fix: `-timeout 550s`, or state explicitly that the Bash ceiling fires first.
- **D7** — evidence-labeling precision — spec.md:31 ("62 of 114 functions") and spec.md:46 ("Only 1 function is below 60%") — the census is name-granular (unique names; `Error`/`pid` each aggregate multiple receivers — :494 pid 0.0% vs :701 pid 66.7%), so "functions" overstates it; and `terminateCodexProcess` (0.0%) is also below 60%, scoped only by the preceding bullet. Scope decisions derive from the function-granular execution axis, so no decision changes. — Severity: **advisory** — Class: optional — Required fix: "62 of 114 unique function names" and "only 1 non-zero function below 60%".

## Regression Check (Iteration 2+ only)

N/A — iteration 1.

## Verdict Basis (explicit, because the score alone would read PASS)

All seven must-pass criteria PASS and the aggregate (0.875) clears both tier thresholds. The verdict is nevertheless **FAIL** because four blocking-class defects (D1-D4) stand inside the SPEC's verification contract itself: the acceptance layer — the artifact whose sole job is to deliver run-phase verdicts — cannot distinguish executed-pass from zero-match on 7 of 11 ACs (D2), its production-diff gate cannot detect the one violation REQ-CTG-009 exists to forbid (D3), and its sole skip-record deliverable carries a factually refuted rationale (D4) that AC-CTG-008/010 assert as content. Per verification-completeness §1.1 these are report-not-verdict instruments: output that looks like verification while carrying no decision. Blocking-class findings route a fix round before the verdict is revisited; the required codex gate independently reached FAIL with the same substance. Iteration 2 re-audits the D1-D7 delta only.

## Gaps (what this audit did NOT observe)

- The test suite was not executed and the coverage run was not re-performed — this is a plan-phase artifact audit; the committed evidence files were taken as the baseline (their internal consistency, the run-log verbatim line, and the per-function extract were cross-checked, and the code anchors were read from source).
- Codex's vacuous-green probe (F2) and overlay pid test (F4) were adopted after mechanism verification by inspection (go testing selector semantics; `os.Process.Pid` exported; same-package struct constructibility), not re-executed by this auditor.
- Full-suite CI status was not consulted (plan-phase commit, tests-only change pending; per lane protocol the develop-window push is the verdict surface).
- The original card text was not present in the checkout; scope judgment is anchored to the SPEC's own §B record plus the committed evidence, both of which checked out.

## Residual-risk

- The arm-level residue for REQ-CTG-003/004/005/006 rests on card cluster naming plus function-level coverage; if any of those arms turns out to be already exercised by an existing idiom, the item degrades to a cheap no-op — bounded, not hazardous.
- If the run-phase re-measures coverage after unrelated develop-window merges touch `internal/cli`, the "118 functions" denominator may shift; AC-CTG-010's completion count should be re-derived from the fresh extract at run time (plan M7 already requires the fresh extract).

## Recommendation (numbered fix route for manager-spec)

1. D1 — set `tier: M` in spec.md:14 and progress.md:9; note the tier correction in §A History.
2. D2 — amend the Verification column of AC-CTG-001..007 (acceptance.md:9-15) and plan.md §E E1 to `-count=1 -v` + required `--- PASS: Test<Name>` observation; add the zero-match RED-now baseline line.
3. D3 — replace AC-CTG-009's verb and plan §E E4 with the base-pinned union check (committed+staged+unstaged+untracked `.go` paths ⊆ `internal/cli/**/*_test.go`).
4. D4 — resolve the pid skip: add the trivial hermetic test and retarget REQ-CTG-010 to zero remaining 0.0%, or rewrite the §D row as a deliberate low-value skip (remove "not reachable").
5. D5 — give AC-CTG-011 a REQ anchor (new REQ-CTG-011) or fold it into plan §E.2 E2.
6. D6/D7 — one-line precision edits (timeout value; census labeling).
7. After fixes land, iteration 2 confirms the delta only; Implementation Kickoff Approval remains mandatory and is not affected by this audit's verdict mechanics.
