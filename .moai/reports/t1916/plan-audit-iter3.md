auditor-model: glm-5.3-flash

# SPEC Review Report: SPEC-FACTORY-LEASE-BUSY-001
Iteration: 3/3 (delta re-audit — final iteration before the ceiling policy)
Verdict: FAIL
Overall Score: 0.97
Plan Artifact Hash: 35f4d5a21d78051d6b4619b4ca6d0e013587ed6a7cdaaf713f6ea96d61c78f29 (SHA-256 over spec.md + plan.md + acceptance.md concatenated in that order, Tier M set, at HEAD ae3bfb729f)
Auditor Version: plan-auditor/v3.2 (served by glm-5.3-flash; independent of manager-spec)

verdict: FAIL
audited_sha: ae3bfb729fb7e4ba332db88e82fd238bd9057fca
overall_score: 0.97
must_pass_failed: 0
blocking_count: 2
plan_artifact_hash: 35f4d5a21d78051d6b4619b4ca6d0e013587ed6a7cdaaf713f6ea96d61c78f29
scope: delta
fix_scope: plan.md#§F-M4
defect_class: design (D16); design (D17)
reread_hunks: plan.md#§F-M4
convergence_overall: fail
required_backend: codex fail

Delta basis: the iter2 report's fix_scope and riders, repaired in commit c2740ece2b ("apply
plan-audit iter2 repairs D11-D14") plus the iter2 verdict-file commit; audited at HEAD
ae3bfb729fb7e4ba332db88e82fd238bd9057fca (pinned by `git rev-parse HEAD` at audit start, clean
tree). The full repair diff ad8603214c..c2740ece2b was read (acceptance.md 19 lines,
decision-index 2, plan 22, progress 7, spec 11 changed). Reasoning context from the SPEC author
was not present; M1 Context Isolation holds. Tier M input contract: spec.md + plan.md +
acceptance.md read in full.

## Must-Pass Results (re-measured this run)
- [PASS] MP-1 REQ number consistency: `grep -oE 'REQ-FLB-[0-9]+' spec.md | sort -u | wc -l` → 12
  (REQ-FLB-001..012, sequential, zero-padding consistent); repairs renumbered nothing.
- [PASS] MP-2 EARS/GEARS format compliance (requirement layer; ACs graded under Group 4): the one
  repaired requirement — REQ-FLB-012 (spec.md:136-141) — remains Event-driven ("When the
  post-change measurement is taken, the `factory status` read path shall satisfy the outcome the
  baseline selects: …") with the ceiling now carried in the requirement layer; all other REQs
  unchanged from iter2's PASS assessment (REQ-FLB-005 legacy negative in-window; REQ-FLB-009
  canonical Where). No IF/THEN; no informal normative language.
- [PASS] MP-3 YAML frontmatter validity: untouched by the repair diff (spec.md hunks touch only
  HISTORY line 36 and REQ-FLB-012); all 12 canonical fields stand as measured in iter1/iter2.
- [N/A] MP-4 language neutrality: single-language (Go) SPEC — auto-passes per LN-3.
- [PASS] MP-5 D7 cross-SPEC reconciliation: the repair diff added no new SPEC references; the
  three references and their live statuses (in-progress/completed/completed) stand as measured in
  iter2; zero REVIEW/SHOULD conditions.
- [PASS] MP-6 D8 cross-platform discipline: `grep -c syscall` → 0/0/0 (spec/plan/acceptance),
  re-measured fresh. Auto-pass per D8-4.
- [PASS] MP-7 clarification gate: `grep -n 'NEEDS CLARIFICATION' plan.md` → exit 1 on the
  repaired plan.md; research.md correctly absent (Tier M).
- [N/A] MP-8 RED-now cell completeness: `grep -c 'release-blocking' acceptance.md` → 0; no
  criterion is release-blocking, the four-element obligation does not bind (reason stated), and
  no RED-now command was executed at audit time.
- [PASS] MP-9 cross-artifact ordering consistency: the CN-4 awk verb remains refused-class (iter1
  guard refusal; not retried per the standing instruction); the plain-grep substitution re-run
  fresh → 12 candidates (L5, 36, 45, 46, 50, 70, 75, 92, 93, 97, 126, 131), each hand-read
  against plan §F order M0→M1→M2→M3→M4→M5→M6 (no `Exit:` bindings in the plan). L70 (AC-FLB-008
  baseline precedes the first read-path commit) = M4-before-M5 ✓; L126 (RED before GREEN) =
  M1→M2 ✓; L131 (M0 gate before any implementation commit) ✓; L75 ("selected by the baseline
  first") is outcome-selection precedence inside AC-FLB-009, not milestone ordering; the rest are
  intra-invocation arms. Zero conflicts; no new ordering obligation introduced by the repairs.

## Category Scores (0.0-1.0, rubric-anchored)
| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 1.00 | single-interpretation requirements | REQ-FLB-012's repaired form (spec.md:136-141) selects the outcome by the baseline with the ceiling in-line — one interpretation, same conditional shape as AC-FLB-009; all other requirements stand as assessed in iter2 |
| Completeness | 0.90 | one narrow residue | D11 (stale counts, 4 locations) closed; D13's binary-path collision closed (plan.md:135-137 "exactly one token changed: the binary path"); D15 closed (one root serves both runs, read-only justification). Residue: the replacement seed procedure (clause (3)) is not executable as written — D16/D17 below — so M4 still cannot produce its artifact end-to-end |
| Testability | 1.00 | every AC binary-testable | AC-FLB-009 (acceptance.md:73-77) is now baseline-first precedence — disjoint by construction (baseline < 200 vs ≥ 200 partition the space); mutants re-probed: post=500 at baseline=100 fails the ceiling (500 > 200) ✓, post=40 passes with the named outcome ✓. AC-FLB-007 (acceptance.md:61-66) carries the causal quantifier ("neither site's own outcome is attributed to a race loss") plus the explicit downstream-race exclusion; the mutation that attributes the peek failure to a race still fails it ✓. §C's "Latency floor escape" (acceptance.md:119-120) now agrees with AC-FLB-009 by construction — both record "defect not reproduced" when baseline p95 < 200 ms, the iter2 §A-vs-§C contradiction is gone |
| Traceability | 1.00 | full coverage | §B untouched by the repair diff (REQ-FLB-008 → AC-FLB-006 + AC-FLB-011 stands); AC ids 001..011 and REQ ids 001..012 re-measured; DoD #1 "All eleven" consistent with plan §E's repaired "all eleven AC-FLB ids" (plan.md:82) |

Aggregate 0.97 ≥ Tier M threshold 0.80, but 2 blocking defects (D16, D17) force FAIL. Score
monotonicity holds: 0.69 → 0.90 → 0.97 (no STOP signal).

## Defects Found (structured defect-list)

Repairs verified CLOSED this iteration (full adjudication in the Regression Check):

- D11 — [RESOLVED]: plan.md:82 "all eleven AC-FLB ids"; plan.md:156 "AC-FLB-001..011";
  spec.md:36 "criterion count 11"; decision-index.md:44 Q4 heading now names
  `TestFactoryLeaseQueueLockStallBounded` and its subtests instead of 443/464. All four token
  edits landed.
- D12 — [RESOLVED]: acceptance.md:73-77 replaces "exactly one … holds" with baseline-first
  selection ("when baseline p95 < 200 ms the outcome is 'defect not reproduced' … otherwise
  (baseline p95 ≥ 200 ms) the outcome is 'defect fixed' when post-change p95 ≤ 50% of baseline
  p95") — mutually exclusive by construction; the iter2 computed failure case (100/40) now
  passes with the named outcome while the 10x-regression mutant (500 > 200) still fails;
  REQ-FLB-012 aligned to the same shape with the ceiling in the requirement layer; §C agrees by
  construction.
- D13 — [RESOLVED]: plan.md:135-137 — "reissued verbatim for the post-change measurement with
  exactly one token changed: the binary path, the sole revision-varying measurement parameter —
  every other argument, environment variable, and fixture input is byte-identical across the two
  runs." Clauses (1) and (4) are now jointly satisfiable.
- D14 — [RESOLVED]: acceptance.md:61-66 — causal quantifier ("neither site's own outcome is
  attributed to a race loss; no rendered line names this serial-slot outcome … as the cause")
  plus the explicit exclusion ("A genuine race lost downstream of the opened peek … is
  REQ-FLB-002's own mandated `raced` outcome and does not violate this criterion"). The
  over-breadth is gone; the gating mutation still fails.
- D15 — [RESOLVED]: plan.md:126-128 — "one fresh `mktemp -d` project directory serves both runs
  … (`factory status` is read-only, so one root keeps the fixture and the measurement
  comparable)."

New blocking defects (surfaced by the codex convergence leg; each corroborated independently
from this tree's source before adoption):

D16. FLB-D16 — plan.md:128-130 (M4 clause (3), seed line) — the seed procedure fails at its
  second verb as written: the plan seeds with plain `moai todo add <fixture-card>` then
  `moai factory assign <fixture-card> --run <fixture-run>`, but `factory assign` refuses
  non-picked queue items — the guard is `internal/cli/factory_card.go:3033`:
  `return fmt.Errorf("queue item %s is %s, not picked", cardID, state)` — and plain `todo add`
  creates a queued (not picked) item, so the seed's factory record is never created. The
  corrective flag exists in the tree: `todo add --pick` (internal/cli/todo.go:924; the t71
  pick-marking fold "add --pick folds the add and the [pick]", todo.go:14). Codex reported the
  measured failure (`queue item t1 is queued, not picked`, exit 1, freshly built binary in an
  isolated root); my corroboration is the source guard plus the flag's existence — the recipe's
  seed step cannot execute as written. — Severity: major — Class: blocking — Required fix: seed
  with the pick-marked issuance — `moai todo add --pick <fixture-card>` then `moai factory
  assign <issued-id> --run <fixture-run>` — and record the issued id in the baseline artifact.
D17. FLB-D17 — plan.md:130-132 (M4 clause (3), reissue parenthetical) — "the exact seed command
  lines … are recorded verbatim in the baseline artifact and reissued unchanged for the
  post-change run" cannot be executed against the same root clause (2) now mandates: the
  reissued seed's first verb (`todo add`) refuses a duplicate with "already holds this card"
  (internal/cli/todo_analysis.go:42 — observed by codex as the post-change re-seed exit 1), and
  `--force` would add a second card, breaking clause (4)'s own "every … fixture input is
  byte-identical across the two runs". The re-seed reading also contradicts clause (3)'s "seeded
  once into that root". — Severity: major — Class: blocking (same one-clause rewrite as D16) —
  Required fix: strike the reissue directive — the seed runs exactly once before the first run;
  the post-change measurement reuses the already-seeded root and changes only the binary path;
  the seed lines stay recorded verbatim in the baseline artifact for reproducibility, not
  re-execution.

One repair commit closes both: rewrite M4 clause (3)'s seed line with `--pick` and the
issued-id form, and replace the "reissued unchanged for the post-change run" parenthetical with
"recorded verbatim for reproducibility; executed once; the post-change run reuses the seeded
root unchanged."

## Regression Check (Iteration 3)
- iter2 D11-D15: all five RESOLVED (adjudications above, each with the artifact line read).
- iter1 D1-D10: remain RESOLVED — the iter3 repair diff touched none of their hunks (the diff
  paths are M4 clause (3), §E/§H tokens, spec HISTORY count, REQ-FLB-012, AC-FLB-007/009,
  decision-index Q4; none of the iter1 repair sites).
- New this iteration: D16, D17 — both confined to plan.md §F M4 clause (3); no requirement-layer
  or acceptance-layer defect is open.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

| # | Claim | Evidence (command → verbatim output) | Baseline-attribution |
|---|-------|--------------------------------------|----------------------|
| R1 | Tree identity | `git rev-parse HEAD` → `ae3bfb729fb7e4ba332db88e82fd238bd9057fca`; `git log --oneline -4` → ae3bfb729f / c2740ece2b / ad8603214c / 013ace676f; `git status --porcelain` → empty | This run, this worktree |
| R2 | Repair subject set | `git diff ad8603214c..c2740ece2b --stat` → 5 files; full diff read (acceptance 19, decision-index 2, plan 22, progress 7, spec 11 lines) | This run |
| R3 | D16 corroboration | `grep -n "picked" internal/cli/factory_card.go` (3040-3240 band) → :3033 `return fmt.Errorf("queue item %s is %s, not picked", cardID, state)`; :3181 "only a picked card can move to assigned"; `grep -n '"pick"' internal/cli/todo.go` → :924 `cmd.Flags().BoolVar(new(bool), "pick", false, …)` | This run, tree ae3bfb729f |
| R4 | D17 corroboration | `grep -rn "already holds this card" internal/cli/` (non-test) → todo_analysis.go:42 `"todo add: %s already holds this card (%q) — pass --force to add it anyway"`, todo.go:1111 | This run |
| R5 | Seed verbs exist | `grep -rn 'Use:.*"assign' internal/cli/` → factory_card.go:3042 `Use: "assign <card>"`; `--run` flag usage in factory_bundle_test.go:207-253 | This run |
| R6 | MP measurements | REQ ids → 12; AC ids → 001..011; `release-blocking` → 0; `syscall` → 0/0/0; `NEEDS CLARIFICATION` → exit 1 | This run |
| R7 | CN-4 substitution | ordering-word grep → 12 candidates, hand-read, zero conflicts | This run |
| R8 | Convergence | `mcp__moai__audit_multi` (target ad8603214c..ae3bfb729f) → codex verdict fail, 2 findings (seed precondition; seed reissue), receipt rcpt-c010efb4f9d4c78c7b1e6ce8; digest written and `--result-file` re-check → `convergence_check: {ok: true, unmet: []}` — the required backend answered; its findings were adopted only after the R3/R4 source corroborations above | This run, tree ae3bfb729f |
| R9 | Plan-artifact hash | `cat spec.md plan.md acceptance.md \| shasum -a 256` → 35f4d5a21d78051d6b4619b4ca6d0e013587ed6a7cdaaf713f6ea96d61c78f29 | This run |

**Gaps**: (1) the CN-4 awk verb and the traceability awk verb remain in the iter1 refusal class
("awk … construct too complex to verify") and were not retried; the plain-grep substitutions and
hand-reads are named per VCI §3.1. (2) codex's transient verifications (its isolated-root seed
reproductions with a freshly built binary) were not re-executed by me; each finding was adopted
on my own source corroboration (R3-R5), which is stronger evidence than the runs alone. (3)
build_lag notice: the installed moai binary (8249bff8c) predates this tree's HEAD — no moai CLI
measurement is cited in this verdict, so no claim rests on the stale build. (4) No RED-now cell
was re-executed (none is release-blocking).

**Residual-risk**: This is iteration 3 of 3 — the tier ceiling. The verdict carries `fix_scope`
with a single scoped anchor (plan.md#§F-M4), and the eligible next delta (clause (3)'s seed
rewrite only, plus progress.md/.moai/reports) stays inside the anchor range with the REQ/AC id
sets unchanged — the ceiling policy's `auto_delta_rounds` delta audit is therefore available
without a question. If that delta leaves D16/D17 unresolved, the final-hit admission ladder
applies; the split arm stays reachable because the blocking findings carry scoped fix anchors.
Everything outside plan.md clause (3) is admission-clean: the requirement layer, the acceptance
layer, traceability, decision routing, and the M0 sequencing hold are all sound, and both
blocking defects are one clause away from closed.

## decision-index routing assessment
Unchanged from iter2 and untouched by the repair diff except the Q4 heading anchor (now
name-level): Q1's published default (exit 4 for every class) remains encoded in
REQ-FLB-002/003/004/007; Q2 remains baseline-relative (now with the aligned REQ-FLB-012/AC-FLB-009
conditional shape); Q3 POLICY-COVERED (TDD); Q4 DEFAULT-APPLIED. No new decision surfaced.

## Recommendation (numbered fix for manager-spec, the delta round)
1. D16+D17 (one clause, plan.md §F M4 (3)): seed with `moai todo add --pick <fixture-card>` →
   `moai factory assign <issued-id> --run <fixture-run>` (record the issued id verbatim in the
   baseline artifact); replace the "reissued unchanged for the post-change run" parenthetical
   with "recorded verbatim for reproducibility; the seed executes exactly once and the
   post-change measurement reuses the seeded root with only the binary path changed."

Fit to enter run phase after M0: NOT in current form — D16/D17 must land first; after that
one-clause repair, every prior defect family (iter1 D1-D10, iter2 D11-D15) is closed and the
contract is fit for the kickoff review.
