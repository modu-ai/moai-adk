# SPEC-GRAPH-EDGE-CONFIDENCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-01 by manager-spec in worktree
  `.claude/worktrees/t411` (branch `WT-edge-confidence`, origin/develop base).
- SPEC ID pre-write check: `SPEC-GRAPH-EDGE-CONFIDENCE-001` regex PASS; ID unique
  in `.moai/specs/` (0 prior matches).
- Current-state facts re-derived on THIS tree (not carried from the main-based
  analysis report): Edge struct, 7 edge kinds, astx grade constants + 6
  call-grade languages all `name-based`, `CodeMatch`/`CallTraceEdge` shapes, MCP
  handlers, freshness `checkEdges` fingerprint-over-source, consumers
  (`ImportFanIn`/`UnreferencedSpecs`/`BlastRadius`/`FindCallers`).
- Artifacts: spec.md · plan.md · acceptance.md · progress.md (Tier M set + this
  file).
- Edge-model decision: additive `Resolution` + `Confidence` fields (plan.md §B).
- Repair round 1 (2026-09-01, post plan-audit FAIL 0.81 → re-audit expected at
  Tier M threshold 0.80): D1 T2 rescoped to Go-module imports (non-Go →
  T3/T4; specifier resolution out of scope as candidate future work);
  D2 AC-GEC-009 command fixed to `-run TestNoCGO` (real test:
  `TestNoCGO_FallbackStubsAreUnsupported`, verified in tree);
  D3 same-package tier added — `intra-package`/0.95, REQ-GEC-012 + AC-GEC-012;
  D4 re-tiered S→M, spec.md §D now REQ→AC mapping only, acceptance.md is the
  AC SSOT; D5 AMBIGUOUS-drop rationale recorded (spec.md §C, acceptance.md
  §D.2); D6 `related_specs` dropped from frontmatter (§G carries the link);
  D7 AC-GEC-006 pinned to committed golden
  `internal/graph/testdata/edges-doc-golden.jsonl` with regeneration procedure
  (plan.md §G). Version bumped 1.0.0 → 1.1.0; still `status: draft`.
- Plan-audit outcome (2026-09-01): iter-1 FAIL 0.81 (D1 BLOCKING) → repair
  round (above) → iter-2 delta re-audit **PASS-WITH-DEBT 0.9125**, Clarity gate
  0.90 ≥ M 0.80; all 7 iter-1 defects verified RESOLVED against code. Post-audit
  mechanical corrections applied by the orchestrator per the auditor's iter-2
  recommendation: D-NEW-1 `CGO_ENABLED=0` prefix on AC-GEC-009's second command
  (acceptance.md); D-NEW-2 stale sibling strings (plan.md tier header,
  acceptance.md REQ range, this file's artifact-set note). The plan-artifact
  hash therefore differs from the iter-2-audited bytes; the corrections are
  exactly the auditor-prescribed mechanical edits, recorded here as audit-chain
  closure.

## §F Phase 4 Mode Selection

- Inputs: tier M · scope ~4 packages (internal/graph, internal/graph/symbol,
  internal/navigator/astx, internal/cli handlers+tests) · domains 1 (Go backend)
  · language mix Go 100% · concurrency benefit LOW (coding-heavy) · agent-team
  prereqs: not requested.
- Mode evaluation: direct — not selected (semantic multi-milestone change);
  fanout — not selected (coding-heavy, Anthropic caveat); sweep — not selected
  (not mechanical-uniform, <30 files); agent-team — not selected (no operator
  request).
- Decision: serial
- Justification: coding-heavy Tier M implementation across 4 milestones;
  Anthropic's coding-task parallelism caveat favors one sequential
  manager-develop delegation. Progression axis: AUTONOMOUS — Implementation
  Kickoff Approval PASSED 2026-09-01 (operator selected TDD start + autonomous
  goal arming); `/moai goal` armed post-gate (mechanical condition: scoped
  graph/astx/cli confidence tests exit 0; ceiling 30 turns). Semantic-failure
  escalation remains operator-owned.
- Baseline: origin/develop absorbed pre-spawn (13 commits) → HEAD 5593e8cff;
  build OK; internal/graph+symbol+astx tests green (this run, this tree).

## §E.2 Run-phase Evidence

Run phase executed 2026-09-01 by manager-develop (cycle_type=tdd) in worktree
`.claude/worktrees/t411`, branch `WT-edge-confidence`, serial, 4 milestones,
4 commits (M1 68cf2f724 · M2 937c9f511 · M3 dbdcd94b7 · M4 629bbc12a; base
5593e8cff). All measurements below are (this run, this tree @ 629bbc12a).

### AC matrix (E1)

| AC | Status | Verification command (acceptance.md) | Observed output |
|----|--------|--------------------------------------|-----------------|
| AC-GEC-001 | PASS | `go test ./internal/graph/ -run TestEdgesJSONLDeterministic -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.403s` |
| AC-GEC-002 | PASS | `go test ./internal/graph/ -run TestCodeCallConfidenceDomain -count=1` | `ok … 0.369s` (fixture exercises all three tiers; domain closed) |
| AC-GEC-003 | PASS | `go test ./internal/graph/ -run TestSameFilePromotion -count=1` | `ok … 0.366s` |
| AC-GEC-004 | PASS | `go test ./internal/graph/ -run TestImportEvidencePromotion -count=1` | `ok … 0.363s` (Shared/Dup/Do incl. T2-before-T3 tie) |
| AC-GEC-005 | PASS | `go test ./internal/graph/ -run TestNameOnlyFallback -count=1` | `ok … 0.361s` (Go non-imported + undeclared + non-Go python case) |
| AC-GEC-006 | PASS | `go test ./internal/graph/ -run TestDocEdgesByteIdentical -count=1` | `ok … 0.350s` (golden `testdata/edges-doc-golden.jsonl`, generated on base 5593e8cff) |
| AC-GEC-007 | PASS | `go test ./internal/cli/ -run TestGraphMCPConfidenceExposed -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.802s` (per-edge values 1.0 and 0.85 asserted) |
| AC-GEC-008 | PASS | `go test ./internal/graph/ -run TestLegacyArtifactLoad -count=1` | `ok … 0.374s` |
| AC-GEC-009 | PASS | `CGO_ENABLED=0 go build ./...` + `CGO_ENABLED=0 go test ./internal/navigator/astx/ -run TestNoCGO -count=1` + `CGO_ENABLED=0 go test ./internal/graph/ -run TestNoCGOConfidenceInert -count=1` | exit 0; `ok … 0.173s`; `ok … 0.215s` (test observed running via -v) |
| AC-GEC-010 | PASS | `go test ./internal/graph/... ./internal/navigator/astx/ -count=1` | `ok` × 3 packages (graph 14.074s / symbol 0.883s / astx 1.961s) — grade matrix + traversal suites unchanged |
| AC-GEC-011 | PASS | `go test ./internal/graph/ -run TestProvenanceShapeUnchanged -count=1` | `ok … 0.397s` |
| AC-GEC-012 | PASS | `go test ./internal/graph/ -run TestSamePackagePromotion -count=1` | `ok … 0.397s` (both directions: import-holding caller and zero-import caller) |

### RED evidence (E8, verbatim, captured before GREEN)

- M1 (pre-implementation, @ 5593e8cff + tests only): `undefined:
  ResolutionExtracted / ResolutionIntraPackage / ResolutionInferred /
  ConfidenceFor … e.Resolution undefined (type Edge has no field or method
  Resolution) … too many errors — FAIL github.com/modu-ai/moai-adk/internal/graph
  [build failed]`.
- M2 regression locks proven by mutant probes (observed RED before accepted):
  A — confidence stamped on code-import edges → `doc+code-import lines diverged
  from the committed golden (691 vs 817 bytes)`; B — ConfidenceFor(inferred)→0.9
  → `ConfidenceFor("inferred") = 0.9, want 0.85`; C3 — first-iterated declaring
  dir decides T2 → `edges.jsonl not byte-identical: build 0 vs build 2` (the
  fixture's three-directory callee is the value-flip hazard). Emission-order
  mutants (sort removal) were normalized by the canonical sort — recorded as a
  non-finding, not a pass.
- M3: `find match Finish confidence = 0, want 1.0 … trace edge missing
  confidence key` (both handler responses pre-wiring).
- M4: red structurally unreachable under !cgo (failure mode = a CGO-off build
  producing edges); the test is observed RUNNING and sweeping (non-empty) via
  -v. Recorded as a Gap, not a pass of the red-observation obligation.

### E2 cross-platform (per B1) — `go build ./...` exit 0; `GOOS=windows
GOARCH=amd64 go build ./...` exit 0 (this run, this tree).

### E3 coverage — `go test -cover ./internal/graph/... ./internal/navigator/astx/`:
internal/graph 88.5% · internal/graph/symbol 87.1% · internal/navigator/astx
88.5% (all ≥ 85% target; ≥ pre-change level per the same-command baseline).

### E4 boundary grep — `grep -rn 'AskUserQuestion' internal/graph
internal/navigator/astx internal/cli/mcp_code_tools.go internal/cli/mcp_server.go
| grep -v _test.go | grep -v "// "` → empty (no output).

### E5 lint — `golangci-lint run ./internal/graph/... ./internal/navigator/astx/
./internal/cli/` → `0 issues.` Baseline was `0 issues.` (pre-flight); zero NEW.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-01"
run_commit_sha: "629bbc12a"
run_status: "complete"
ac_pass_count: 12
ac_fail_count: 0
ac_pass_with_debt: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "not-run (repo-local B9: no push; integration via lead window)"
l44_post_push_fetch: "not-run (nothing pushed)"
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: "pass (exit 0)"
cross_platform_build.windows: "pass (GOOS=windows GOARCH=amd64, exit 0)"
cross_platform_build.cgo_disabled: "pass (CGO_ENABLED=0 build + scoped tests, exit 0)"
coverage_internal_graph: "88.5%"
coverage_internal_graph_symbol: "87.1%"
coverage_internal_navigator_astx: "88.5%"
total_run_phase_files: 12
m1_to_m4_commit_strategy: "per-milestone commits on WT-edge-confidence; nothing pushed (B9)"
full_suite_verdict: "deferred to CI per repo-local verification-load discipline"
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
