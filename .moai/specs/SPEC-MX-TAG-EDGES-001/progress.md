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

## §F Phase 4 Mode Selection

- Inputs: tier M · scope internal/mx + internal/graph + internal/hook/mx +
  internal/hook (session_end) + internal/cli (mx_query/graph query) · domains
  1-2 (Go backend + hook wiring) · language mix Go 100% · concurrency benefit
  LOW (coding-heavy) · agent-team prereqs: not requested.
- Mode evaluation: direct — not selected (semantic multi-milestone change);
  fanout — not selected (coding-heavy, Anthropic caveat); sweep — not selected
  (not mechanical-uniform, <30 files); agent-team — not selected (no operator
  request).
- Decision: serial
- Justification: coding-heavy Tier M implementation across 5 milestones;
  Anthropic's coding-task parallelism caveat favors one sequential
  manager-develop delegation. Progression axis: AUTONOMOUS — Implementation
  Kickoff Approval PASSED 2026-09-02 (operator selected TDD start + autonomous
  goal arming, continuing the t411 pattern); `/moai goal` armed post-gate
  (mechanical condition: scoped graph/hook-mx/hook/cli tag-edge + fan-in tests
  exit 0; ceiling 30 turns). Semantic-failure escalation remains
  operator-owned.
- Baseline: origin/develop (5043acbd3) absorbed pre-spawn (CHANGELOG resolved
  --theirs: develop's entry set was a strict superset) → HEAD 63435427c;
  build OK; internal/graph+symbol+hook/mx tests green (this run, this tree).

## §E.2 Run-phase Evidence

Run phase executed 2026-09-02 in worktree `.claude/worktrees/t412`
(branch `WT-mx-tag-edges`, TDD RED-GREEN-REFACTOR, serial, 5 milestones,
one commit each — base `63435427c`, nothing pushed).

### Milestones (commit SHA → landed content)

| M | Commit | Content |
|---|---|---|
| M1 | `7443a3523` | Six `mx-*` kind constants (plan B.1, uniform over all six tag kinds); `(File, Kind, Line)`-only edge mapping (B.3); endpoint join = innermost retained `astx.FuncRange` containing the tag line, else self-edge (B.2); ONE `ScanDir` per build via the `scanDirFn` seam (REQ-MTE-006); `BuildWithCodeLayersMode` extracts FIRST so the doc layer joins against retained ranges (extraction failure fails open to the self-edge form); golden extended + regenerated, base `63435427c` named (B.5). `FileDecls` gains retained `Ranges` (retention, no second parse). |
| M2 | `303b53524` | Determinism double-build lock incl. mx-* lines (AC-MTE-001); mx-less legacy artifact load (AC-MTE-005); DEBT-pair identical key sets + metadata-absence scan (AC-MTE-007); reverse-only traversal additivity (AC-MTE-008). Mutant probe: stamping `CreatedBy`+`RotRisk` into mx-debt targets turned `TestDocEdgesByteIdentical` RED (observed, then restored GREEN). |
| M3 | `70025f7d1` | `SymbolFanIn` pure query (REQ-MTE-009: distinct caller files, evidence = extracted/intra-package, inferred itemized separately, declaring file excluded — B.7); `EdgeFanInSource` in `internal/graph/fanin_edge.go` (structurally compatible primitive-typed method — NO hook-layer import; hub exclusion via the shared `mx.IsTestFileWithPatterns` predicate, exported so ONE definition governs both counters — B.8); `FanInEvidenceSource` interface at the consumer + `NewValidatorWithSource` batch constructor; default constructor byte-identical (REQ-MTE-010); LAZY textual index (a source that answers never pays the project walk); REQ-MTE-013 reason shape `fan_in(graph)=N evidence-backed (+M inferred-only)`; parity fixture (graph 3 / inferred 1 / textual larger — AC-MTE-009) + hub exclusion fixture (AC-MTE-012); layering lock `go list -deps` test (acceptance §D.3). |
| M4 | `9ffa524cc` | `session_end.go` constructs the validator with the edge-backed source via the `newFanInEvidenceSourceFn` seam — the SOLE batch injection site (REQ-MTE-010); static guard pins `post_tool.go` to `mx.NewValidator` only; race-free one-shot artifact load (parallel ValidateFiles workers share one source); label contract pinned: fresh=`edges`, unjudgeable sidecar=`edges(stale)` + stderr note, absent/unusable→textual labeled `textual-fallback` (REQ-MTE-011). |
| M5 | `26ca94ad4` | `DebtFanIn` ranking + `moai graph query --debt-fanin` (REQ-MTE-014: desc, ties by target; file-scope DEBT at fan-in 0 LISTED with `(self)` marker — D7); `--fanin` behavior retained, help rewritten; stand-in strings retired to zero (grep + pinned static guard); AC-MTE-010 negative pinned as a test (`mx_scan.go` constructs no validator); CGO gate: `CGO_ENABLED=0` build, `!cgo` self-edge + zero-code-call + source-degrade test, untagged labeled-fallback validator test (REQ-MTE-015). |

### AC matrix (14/14 PASS)

| AC | Status | Verification command (verbatim) | Observed output (this run, HEAD `26ca94ad4`) |
|---|---|---|---|
| AC-MTE-001 | PASS | `go test ./internal/graph/ -run TestEdgesJSONLDeterministicWithTags -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.326s` |
| AC-MTE-002 | PASS | `go test ./internal/graph/ -run TestTagEdgeKindDomain -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.348s` |
| AC-MTE-003 | PASS | `go test ./internal/graph/ -run TestTagEdgeEndpoints -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.348s` |
| AC-MTE-004 | PASS | `go test ./internal/graph/ -run TestDocEdgesByteIdentical -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.348s` |
| AC-MTE-005 | PASS | `go test ./internal/graph/ -run TestLegacyArtifactLoad -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.371s` |
| AC-MTE-006 | PASS | `go test ./internal/graph/ -run TestSingleScanPerBuild -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.347s` |
| AC-MTE-007 | PASS | `go test ./internal/graph/ -run TestTagEdgesCarryNoMetadata -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.346s` |
| AC-MTE-008 | PASS | `go test ./internal/graph/ -run TestTraversalAdditivityWithTags -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.350s` |
| AC-MTE-009 | PASS | `go test ./internal/graph/ -run TestGraphFanInParityFixture -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.352s` |
| AC-MTE-010 | PASS | `go test ./internal/hook/ -run TestPostToolUseKeepsTextualFanIn -count=1` + `go test ./internal/hook/mx/ -run TestConstructorDefaultsTextualSource -count=1` + `go test ./internal/hook/ -run TestSessionEndSelectsEdgeSource -count=1` + negative `go test ./internal/cli/ -run TestMxScanConstructsNoValidator -count=1` | all `ok` (0.426s / 0.197s / 0.428s / cli ok) |
| AC-MTE-011 | PASS | `go test ./internal/hook/mx/ -run TestFanInFallbackLabeled -count=1` | `ok github.com/modu-ai/moai-adk/internal/hook/mx 0.196s` |
| AC-MTE-012 | PASS | `go test ./internal/graph/ -run TestHubExclusionTestCallers -count=1` | `ok github.com/modu-ai/moai-adk/internal/graph 0.353s` |
| AC-MTE-013 | PASS | `go test ./internal/cli/ -run TestGraphQueryDebtFanIn -count=1`; retirement `grep -rnE "stands in for\|stand-in\|no tag-kind edges yet" internal/graph/query.go internal/cli/graph.go` → exit 1 (0 matches, observed) | `ok github.com/modu-ai/moai-adk/internal/cli 0.869s` |
| AC-MTE-014 | PASS | `CGO_ENABLED=0 go build ./...` + `CGO_ENABLED=0 go test ./internal/graph/ -run TestNoCGOTagEdgesSelfEdge -count=1` + `CGO_ENABLED=0 go test ./internal/hook/mx/ -run TestNoCGOFanInTextualFallback -count=1` | build exit 0; both `ok` (0.209s / 0.175s) |

AC-MTE-004 artifact spot-check: `grep -E '"kind":"mx-' internal/graph/testdata/edges-doc-golden.jsonl | grep -cE 'resolution|confidence|rot_risk'` → `0`.

### E-items

- **E2 builds** (HEAD `26ca94ad4`): `go build ./...` → 0; `GOOS=windows
  GOARCH=amd64 go build ./...` → 0; `CGO_ENABLED=0 go build ./...` → 0.
- **E3 coverage**: `go test -cover ./internal/graph/... ./internal/mx/...
  ./internal/hook/mx/ -count=1` → graph 85.7%, graph/symbol 87.8%,
  mx 88.9%, hook/mx 90.0% — all ≥ 85%.
- **E4 boundary grep**: `grep -rn 'AskUserQuestion' internal/mx
  internal/graph internal/hook/mx | grep -v _test.go | grep -v "// "` →
  empty (exit 1). Layering lock additionally pinned by
  `TestHookMxLayeringLock` (`go list -deps`: neither internal/graph nor
  internal/cli in hook/mx's dep set).
- **E5 lint**: `golangci-lint run` over internal/mx, internal/graph,
  internal/graph/symbol, internal/hook/mx, internal/hook, internal/cli →
  `0 issues.` Baseline was `0 issues.` → NEW issues: 0.
- **E6 push state**: branch `WT-mx-tag-edges` HEAD `26ca94ad4`;
  `git ls-remote --heads origin WT-mx-tag-edges` → empty. NOTHING PUSHED
  (integration via the lead's window, per repo git-flow lane protocol).
- **E8 RED evidence** (verbatim, captured before GREEN):
  - M1: `undefined: KindMXNote ... undefined: KindMXDebt ... FAIL
    github.com/modu-ai/moai-adk/internal/graph [build failed]`
  - M3 (graph): `undefined: EdgeFanInSource / undefined:
    NewEdgeFanInSource ... [build failed]`; M3 (hook/mx, observed against
    the reverted pre-seam tree): `v.fanInSource undefined ... undefined:
    NewValidatorWithSource ... viol.Source undefined ...`
  - M4: `undefined: newFanInEvidenceSourceFn ... [build failed]`
  - M5: `unknown flag: --debt-fanin` (observed against the reverted
    pre-M5 tree).
  - M2: RED is the golden divergence itself (observed during M1 as the
    pre-extension `TestDocEdgesByteIdentical` byte-compare failure); the
    M2 locks GREEN on arrival after M1's golden extension, with the
    mutant probe re-demonstrating RED sensitivity (above).
  - AC-MTE-014 nocgo paths: RED structurally unreachable (t411 precedent
    — tests observed RUNNING and sweeping non-empty under
    `CGO_ENABLED=0 -v`; gap recorded).

### Behavior notes for the sync phase

- PostToolUse cost profile unchanged: the default constructor never
  touches the artifact; the textual index now builds lazily only when a
  P1 candidate actually needs it (identical results, no new cost).
- Documented verdict flip (plan B.7 / acceptance §D.2): on same-file-only
  callers the PostToolUse textual source can flag P1 while the SessionEnd
  graph-backed authority does not — deliberate sharpening, batch verdict
  is the enforced one.
- Artifact growth: mx-* lines are the full tag population (accepted with
  observation, plan B.1); this repo's next `moai graph build` will show
  six new kind-count lines' worth of growth in `edges.jsonl`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-02
run_commit_sha: "26ca94ad4"
run_status: "complete"
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "not-run (worktree-isolated lane; pre-spawn absorption verified by the orchestrator)"
l44_post_push_fetch: "not-run (nothing pushed — integration rides the lead's window)"
new_warnings_or_lints_introduced: 0
cross_platform_build.darwin: "pass"
cross_platform_build.windows: "pass (GOOS=windows GOARCH=amd64 go build ./...)"
cross_platform_build.cgo_off: "pass (CGO_ENABLED=0 go build ./...)"
total_run_phase_files: 24
m1_to_m5_commit_strategy: "per-milestone commits on WT-mx-tag-edges (5 commits, 7443a3523..26ca94ad4), no push, no --no-verify, no --amend"
golden_regenerated_base_sha: "63435427c"
gaps:
  - "full-suite verdict NOT claimed locally — origin/develop CI owns it (repo-local verification-load discipline)"
  - "golden pins the CGO-available output only (per plan B.5); CGO-off output covered by table assertions"
  - "stale-artifact stderr note is per-EvidenceBacked-call slog.Warn (injected-sink tests assert the label; production note volume unmeasured on large batches)"
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
