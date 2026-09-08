# Plan-Phase Audit Report — SPEC-CODEX-TEST-GAPS-001 (card t501)

Iteration: 2/2 (Tier M ceiling per `harness.yaml` `plan_audit_tier_ceilings`; final)
Verdict: **PASS**
Overall Score: **1.00** (Clarity 1.0 · Completeness 1.0 · Testability 1.0 · Traceability 1.0; Tier M threshold 0.80)

Delta-only confirmation per the Retry Loop Contract — scope is the enumerated D1-D7 defect delta from iteration 1 plus a regression check, not a from-scratch re-audit.

Baseline attribution: all reads and measurements in this run against worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t501`, branch `WT-codex-uncovered`, HEAD `24df2ae45` (fix round, on top of plan-phase `30bbb1753`, base `ace1c5440`). Working tree clean; no foreign writes (the iter-1 codex overlay test was confirmed removed in iter 1).

## Must-Pass Results (re-verified on 24df2ae45)

- [PASS] MP-1 REQ numbering: REQ-CTG-001..012 sequential, zero-padded, no gaps/duplicates (spec.md §C).
- [PASS] MP-2 GEARS (requirement layer): 001-010 unchanged and compliant; new REQ-CTG-011 `[While]` (spec.md:102) and REQ-CTG-012 `[When]` (spec.md:106) both match patterns. No informal language.
- [PASS] MP-3 frontmatter: 12 canonical fields valid; `tier: M` is a valid optional enum value; `version: "0.2.0"` quoted semver; no snake_case aliases (spec.md:2-16).
- [PASS/N/A] MP-4 language neutrality: N/A single-language SPEC.
- [PASS] MP-5 D7 reconciliation: SPEC refs unchanged (SPEC-CODEX-WIRING-001, SPEC-CODEX-SESSION-MSG-001 — both `status: completed`); no new SPEC references introduced by the fix round. No BLOCKING.
- [PASS] MP-6 D8: `syscall` count 0 across spec.md/plan.md/acceptance.md — auto-PASS.
- [PASS] MP-7 clarification gate: 0 `[NEEDS CLARIFICATION]` matches across all four artifacts (rc=1).

## Regression Check (prior-iteration defects)

- D1 (tier misdeclaration) — **RESOLVED**: `tier: M` at spec.md:14 and progress.md:9; `version: 0.2.0` + History row (spec.md:24); plan.md:3 header "Tier M … 12 REQ/12 AC … above the Tier S ceiling". Budget respected: 12 REQ ≤ 16, 12 AC ≤ 16.
- D2 (vacuous-green AC verification verbs) — **RESOLVED**: AC-CTG-001..007 and AC-CTG-012 verification cells now require `go test -run <Selector> -v` rc=0 AND at least one observed `--- PASS: <TestName>` line, with `[no tests to run]` / zero PASS lines named as explicit FAIL (acceptance.md:9-15, :20); DoD item 1 carries the same requirement (acceptance.md:38). The instrument now has an observed-failure condition (the zero-match state codex demonstrated on 30bbb1753 is the named red).
- D3 (production-diff gate blindness) — **RESOLVED**: AC-CTG-009 (acceptance.md:17) and plan §E E4 (plan.md:46) rewritten as the base-pinned UNION check — `git diff --name-only <FIX-ROUND-BASE-SHA>..HEAD` UNION `git status --short`, filtered to non-test `.go`, must be empty; base SHA cited verbatim in run-phase §E. Coverage: committed (diff range) + staged/unstaged/untracked (`status --short`). The `<FIX-ROUND-BASE-SHA>` placeholder retained in the artifact is the sanctioned self-reference pattern (same physics as the `sync_commit_sha` `pending-backfill` D3 backfill exemption — a commit cannot contain its own hash; the debt is named and owned by the run-phase citation duty). Accepted, not a defect.
- D4 (skip-record false rationale) — **RESOLVED**: `(realCodexConn).pid` removed from §D (now exactly 2 rows) with the v0.2.0 removal note (spec.md:117); REQ-CTG-012/AC-CTG-012 added — 3-branch same-package construction (`&realCodexConn{}`, `&realCodexConn{cmd: &exec.Cmd{}}`, `os.FindProcess(os.Getpid())`), no subprocess spawn, matching my iter-1 source inspection (mcp_codex.go:494-499; `os.Process.Pid` exported); REQ-CTG-010 retargeted to ZERO remaining 0.0% among the 118 rows (spec.md:98, acceptance.md:18); milestones renumbered M1-M8 with pid as M6 (plan.md:79-82). Cross-layer sweep verified: AC-CTG-008 now asserts exactly the two remaining rows + the removal note (acceptance.md:16), plan M8 says "exactly the two remaining surfaces" (plan.md:93) — no stale pid requirement anywhere.
- D5 (orphan AC) — **RESOLVED**: REQ-CTG-011 (quality gates, `[While]` pattern) added; AC-CTG-011 anchored to it. Traceability now bidirectional 12/12.
- D6 (timeout arithmetic) — **RESOLVED** via the ceiling-first route: plan §E background-run note (plan.md:42) acknowledges the 600s foreground ceiling against the 515.6s measurement with little headroom, routes the full scoped run through a background task with output persisted to `.moai/state/verify/t501/cover-after.log` (evidence-persistence-conformant — not `/tmp`), and keeps `-run` scoping for iteration loops. With the run detached from the foreground tool ceiling, the retained `-timeout 700s` go-guard is now reachable; the iter-1 arithmetic conflict is dissolved.
- D7 (evidence-labeling precision) — **RESOLVED (denominators)**: §B.2 census paragraph (spec.md:38) defines 114 = unique function names, 118 = coverprofile function rows, difference 4 = duplicated names across receivers. **Numerically verified this run**: duplicated names are exactly `pid`, `close`, `UnmarshalJSON`, `Error` (each ×2) → 118 − 4 = 114. The sub-bullet "Only 1 function is below 60%" (spec.md:49) remains slightly imprecise in isolation (the 0.0% set is also below 60%) but is scoped by the preceding bullet — residual is advisory, optional class, non-gating per M6.

Declared deviation accepted: final counts REQ 12 / AC 12 (not the 11/11 my D5 arithmetic projected) — D4's 8th test item required its own REQ+AC pair to preserve 100% AC→REQ coverage. Within Tier M ceilings; 12/12 is the intended state.

## Category Scores

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | 1.0 | All 12 REQs single-interpretation with exact anchors; census denominators now defined once and used consistently; plan §A "8 test items" reconciles with REQ-001..007 + REQ-012. |
| Completeness | 1.0 | All sections present; skip record intact with retraction provenance; History carries the fix-round row. |
| Testability | 1.0 | All 12 ACs binary-testable with observed-failure conditions: 8 test ACs require an observed `--- PASS` line (empty sweep = explicit FAIL); AC-CTG-008 exact-row count; AC-CTG-009 union-empty gate; AC-CTG-010 count==0; AC-CTG-011 verbatim tool outputs. |
| Traceability | 1.0 | 12/12 bidirectional REQ↔AC; no orphans; AC-CTG-008/plan-M8 swept to the retracted skip record. |

## Gaps (what this delta audit did NOT observe)

- No test execution, coverage re-measurement, or lint run — plan-phase artifacts only; run-phase evidence duties (E1-E4) remain to be discharged at run time.
- The `<FIX-ROUND-BASE-SHA>` citation in §E is a run-phase obligation, not yet discharged (by design).
- AC-CTG-010's "among the 118" denominator is descriptive of the plan-phase measurement; if develop-window merges add `internal/cli` functions before the run-phase re-measurement, the fresh extract's row count will differ — the count==0 check itself is denominator-independent (residual-risk note, not a defect).

## Residual-risk

- The "Only 1 function is below 60%" narrative imprecision (optional, D7 residue) could re-seed a minor miscount if quoted without the preceding bullet.
- Background execution of the full scoped suite is bounded by go's own `-timeout 700s` (framework-level outside bound, cleanup-guaranteed); a machine slow enough to approach 700s wall time remains possible — the `-run` iteration path is the documented fallback.

## Recommendation

Verdict **PASS** at 1.00 ≥ 0.80 (Tier M). Proceed to Implementation Kickoff Approval (mandatory, score-independent — this PASS never auto-bypasses it). On run entry, the lane must: cite the literal fix-round base SHA (`24df2ae45`) verbatim in §E evidence per AC-CTG-009, discharge E1-E4 per plan §E, and persist evidence under `.moai/state/verify/t501/` or the SPEC directory — never `/tmp` alone.
