# SPEC-MX-TAG-EDGES-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-01 by manager-spec in worktree
  `.claude/worktrees/t412` (branch `WT-mx-tag-edges`, base = origin/develop
  `07ee6e74a` + absorbed SPEC-GRAPH-EDGE-CONFIDENCE-001; HEAD `57d2f3ae3`).
- SPEC ID pre-write check: `SPEC-MX-TAG-EDGES-001` regex PASS (verbatim
  `PASS` output cited in-session); ID unique in `.moai/specs/` — 0 prior
  directory matches (only cross-reference mentions inside
  SPEC-GRAPH-EDGE-CONFIDENCE-001's artifacts, which name this card as
  sibling).
- Current-state facts re-derived on THIS tree (not carried from the
  main-based analysis report): `Edge` struct incl. `Resolution`/`Confidence`
  (t411 already merged), 7 existing edge kinds, `mxSpecEdges` single-scan at
  graph.go:252-277, `Tag` 6-kind domain + wall-clock `LastSeenAt`,
  `fanInIndex` + threshold 3 at validator.go (P1 check lines 270-283),
  PostToolUse 500ms / SessionEnd 4000ms budgets, `--fanin` stand-in text at
  query.go:15 + graph.go:96-97, golden scope (doc kinds + `code-import`),
  `FileDecls` drops ranges today, `internal/hook` cannot import
  `internal/cli`. Drift vs the report's main-based citations recorded in
  plan.md §J.
- Artifacts: spec.md · plan.md · acceptance.md · progress.md (Tier M set +
  this file). Tier argued M: two coupled deliverables across internal/mx +
  internal/graph + internal/hook/mx + internal/cli; 15 REQs / 14 ACs, both
  under the Tier M ceiling of 16.
- Key decisions recorded (plan.md §B): B.1 one edge kind per tag kind
  (uniform over all six, not just the four named); B.2 file → enclosing
  symbol with self-edge fallback; B.3 rot/ceiling metadata stays
  scanner-side; B.4 no new Edge fields — new kinds are new lines;
  B.5 golden extends (same file, regeneration procedure inherited from
  t411); B.6 two-tier wiring (hook keeps textual under the 500ms budget;
  SessionEnd — the sole batch validator site after repair-round D1 —
  graph-backed authority; SessionStart deferred refresh rejected as sibling
  t413 scope); B.7 blocking count =
  evidence-backed (extracted/intra-package) distinct EXTERNAL caller files
  (declaring file excluded), inferred itemized in the reason; B.8 hub
  exclusion IN via the REQ-SPC-004-040 test patterns; B.9 `--debt-fanin`
  query mode + stand-in retirement, `--fanin` behavior retained, self-edges
  rank at fan-in 0.
- plan_status: audit-ready
- plan_complete_at: 2026-09-02
- Repair round 1 (2026-09-02, post plan-audit iter-1 FAIL 0.75 / Testability
  0.50; auditor-verified-sound items untouched — seam cycle-freeness, budget
  constants, freshness probe existence, retention-not-reparse, wall-clock
  hazard aim, golden extension point, ceilings, quality-gate omission):
  - D1 (BLOCKING) `moai mx scan` was a PHANTOM validator surface — verified
    in tree (mx_scan.go constructs scanner+manager only): REQ-MTE-010 and
    plan §B.6/§C/§E-M4 re-scoped to SessionEnd as the SOLE batch injection
    site; mx scan re-characterized as sidecar producer feeding the freshness
    probes; AC-MTE-010 corrected (+ negative assertion).
  - D2 (BLOCKING) declaring-file ruling stated in the requirement layer:
    REQ-MTE-009 now excludes the declaring file (ANCHOR = external
    dependents); §D.2's flip case rewritten honestly (textual does NOT
    exclude same-file callers — deliberate sharpening, hook-vs-batch verdict
    flip accepted); direction argument re-verified — all three graph-side
    filters only remove textual candidates, so graph ≤ textual holds by
    construction (plan §B.7, §D.2).
  - D3 AC-MTE-013 retirement grep widened to the hyphenated form:
    `grep -rnE "stands in for|stand-in|no tag-kind edges yet" ...` → 0.
  - D4 parity-fixture test home named (plan §C): `internal/graph/
    fanin_edge_test.go` with an IN-TEST textual-matched reference counter —
    the only package where both sources are observable without violating the
    hook/mx layering lock; wiring observations stay in `internal/hook` +
    `internal/hook/mx` test files.
  - D5 `depends_on: [SPEC-GRAPH-EDGE-CONFIDENCE-001]` added to spec.md
    frontmatter (run-gate pre-flight reads it).
  - D6 mx.yaml `test_paths` divergence from the graph hub-exclusion recorded
    as ACCEPTED with candidate future work (spec.md §E, plan §B.8) —
    minimal-change leaning.
  - D7 self-edge ranking ruled: file-scope DEBT targets rank at fan-in 0 and
    are listed explicitly with a `(self)` marker (REQ-MTE-014, plan §B.9,
    AC-MTE-013).
  - D8 citation fixed: the test-path predicate is the package-level
    unexported `isTestFileWithPatterns` at `internal/mx/fanin.go:47`, not a
    method (plan §B.8, AC-MTE-012).
  - Version 0.1.0 → 0.2.0; both auditor code measurements independently
    re-verified in this tree before applying (mx_scan.go:64-114;
    fanin.go:47-69; validator.go:132-138).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
