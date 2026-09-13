# SPEC Review Report: SPEC-AUT-PERMMODES-001 (card t584) — Iteration 2

Iteration: 2/2 (Tier M ceiling = 2)
Verdict: PASS
Overall Score: 1.0 (Tier M PASS threshold: 0.80)

Auditor: plan-auditor (independent). Re-audit scope per the Retry Loop Contract: delta over iter-1 D1 + regression over D2–D5 + protected-set drift check. Delta surface: `git diff 0ddd1a282..96354b216` — exactly 4 files (spec.md, acceptance.md, progress.md, the auditor's own iter-1 report). HEAD at audit: `96354b216`; working tree clean; revision commit `docs(SPEC-AUT-PERMMODES-001): plan revision iter2 — AC coverage for REQ-009/010 (audit D1)` is the sole commit in range.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

**Claim.** The D1 blocking defect is fully resolved via the AC route (AC-011 for REQ-009, AC-012 for REQ-010, with §D.2 rewritten to a full REQ→AC map that honestly demotes plan §E.4 to supplementary drift detection); all four optional defects (D2–D5) are also resolved; the protected set (requirement bodies, M1 blocker gate in plan.md, test-preservation list, Out of Scope) shows zero drift. All four dimensions re-score to 1.0 → aggregate 1.0 → PASS.

**Evidence.** All from `git diff 0ddd1a282..96354b216` (verbatim, this run, this tree):

- **D1 RESOLVED** — acceptance.md gains two rows: `AC-011 | REQ-009` (Given `moai init --help` output, When the `--autonomy-tier` flag description is read, Then it names the three permission-mode choices, states the acceptEdits default, and the closed-set tokens are unchanged) and `AC-012 | REQ-010` (Given the godoc of both functions, When each block is read, Then the new three-way mapping and the re-scoped REQ-004 bounded-delta invariant are stated). Both are binary-testable presence assertions with no weasel words. §D.2 is rewritten: "REQ-009→AC-011 · REQ-010→AC-012. Every REQ carries at least one AC; plan.md §E.4's grep guards (E4) are supplementary drift detection only and are NOT the verification route for any REQ" — the false iter-1 claim is retracted, not papered over. Consistency ripple complete: §D.1 → "Should-pass: AC-009..AC-012", §D.7 → "AC-001..AC-012", progress.md §E.1 → `ac_count: 12`. Budget: 10 REQ / 12 AC ≤ Tier M ceiling 16/16.
- **D2 RESOLVED** — spec.md gains a `## HISTORY` section (2-row version table with provenance, citing iter-1 audit path).
- **D3 RESOLVED** — `version: 0.1.0` → `version: "0.1.1"` (quoted), with a matching HISTORY row.
- **D4 RESOLVED** — §H relabeled to "AC-AUTONOMY-TIERS-006 at :167, AC-AUTONOMY-TIERS-007 at :168 — the AC summary rows carrying the REQ-006/REQ-007 invariants", matching what iter-1 measured at those lines.
- **D5 RESOLVED** — REQ-003 subject de-named ("The tier-to-mode mapping SHALL map…"), REQ-005 tail de-named ("The gating logic's behavior is otherwise unchanged"). Both retain the GEARS SHALL pattern (MP-2 unaffected). REQ-010's retention of the two function names is correct — those godocs ARE its subject.
- **Protected-set regression: ZERO DRIFT** — the diff contains NO plan.md hunks (M1 blocker gate untouched at plan.md §F exit), no hunks in REQ-001/002/004/006/007/008/009 bodies, and no hunks in §D Out of Scope. The AC-008 test-preservation list is byte-identical. `syscall` and `[NEEDS CLARIFICATION` appear nowhere in the delta (MP-6/MP-7 carry forward on iter-1 measurement + clean delta).

**Baseline-attribution.** HEAD `96354b216`, worktree `.claude/worktrees/t584`, 2026-09-13, this run.

**Gaps.** Same as iter-1 and unchanged: no WebFetch on this surface, so the six-value-enum claim remains accepted-on-citation (2026-09-13, progress.md §E.1) with plan M1's blocker gate as the residual control. No tests executed (plan-phase scope).

**Residual-risk.** None new. The auditor's own iter-1 report was committed by the author into the same commit — accepted: it is the card evidence path convention, the file is unmodified (87 insertions = the file as written), and committing it makes the cited evidence path resolve post-merge.

## Must-Pass Re-verification (delta-scoped)

- MP-1 PASS (REQ-001..010 unchanged, sequential — no hunk touches §B numbering)
- MP-2 PASS (requirement layer; REQ-003/REQ-005 rewordings keep SHALL patterns)
- MP-3 PASS (frontmatter: only the `version` line changed, now quoted; all 12 fields + `tier: M` intact)
- MP-4 N/A (unchanged, single-language)
- MP-5 PASS (no change to SPEC references; iter-1 measurement carries)
- MP-6 PASS (`syscall` absent from spec.md and from the delta)
- MP-7 PASS (no markers; artifact set unchanged at 3 + progress)

## Category Scores (iteration 2)

| Dimension | iter1 | iter2 | Evidence |
|-----------|-------|-------|----------|
| Clarity | 0.75 | 1.0 | D5 fixed; REQ-010's function names are its legitimate subject; no ambiguity remains. |
| Completeness | 0.75 | 1.0 | D2 fixed (HISTORY present); all sections + 12/12 frontmatter. |
| Testability | 1.0 | 1.0 | AC-011/AC-012 binary-testable; still zero weasel words across AC-001..012. |
| Traceability | 0.50 | 1.0 | D1 fixed: every REQ has ≥1 AC in both directions; E4 demoted honestly. |

**Overall Score: (1.0 + 1.0 + 1.0 + 1.0) / 4 = 1.0 → PASS** (score trajectory 0.75 → 1.0: no regression, no STOP trigger).

## Regression Check (Iteration 2)

- D1 (blocking, TRACE-001): **RESOLVED** — AC-011/AC-012 added; §D.2 corrected. Evidence above.
- D2 (STRUCT-001): **RESOLVED** — HISTORY section added.
- D3 (FM-001): **RESOLVED** — version quoted, bumped, HISTORY-reconciled.
- D4 (XREF-001): **RESOLVED** — line labels now name the AC rows measured at :167/:168.
- D5 (REQ-STYLE-001): **RESOLVED** — REQ-003/REQ-005 subjects de-named; GEARS patterns preserved.
- New defects introduced by the revision: none found.

## Recommendation

PASS. The SPEC is plan-audit-clear at Tier M threshold with margin. For the lane: proceed per the normal gate sequence (Implementation Kickoff Approval remains mandatory and is not bypassed by this verdict). During run-phase, the two items this audit could not observe remain owned where the SPEC already placed them: (1) M1's version-floor research must land before M2 (its own exit gate makes a contradiction a blocker, not a workaround); (2) AC-011/AC-012 are presence assertions on `--help` output and godoc — they must be verified against real `go run`/`go doc` output in progress.md §E, not against the source text alone.
