# SPEC-GRAPH-REPORT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-02

Plan-phase artifacts authored 2026-09-02 by manager-spec (spec.md, plan.md, acceptance.md; Tier M). Plan-auditor verdict pending — this section is updated to `audit-ready` only after an observed PASS verdict.

Revision 2026-09-02: plan-audit iter-1 FAIL 0.625 → artifacts revised to 0.2.0 resolving D1-D12, D14, D16 (D1 resolved by the lane as fixed rotating `graph-report.md` with operator veto open at the run gate; D5 resolved as a cli-injected `DeferredEdgesRefresh` DI seam). Awaiting iteration-2 re-audit (delta-scoped).

Revision 2026-09-02: plan-audit iter-2 PASS-WITH-DEBT 0.9375 → artifacts revised to 0.2.1 resolving the residual debt D17 (node-id shape `file:function`), D18 (shrink guard kind-scoped to file-sourced kinds + `projectRoot` param), D19 (hook-side staleness probe cites exported predicates), D20 (per-refresh unique temp suffix), D21 (output-path flag removed — fixed path only), N1 (path-coverage wording). The edit invalidates the plan-artifact hash, so the run-gate re-executes on the next `/moai run` per the audit debt terms.

Revision 2026-09-02: plan-audit review-3 (`.moai/reports/plan-audit/SPEC-GRAPH-REPORT-001-review-3.md`) — verdict PASS-WITH-DEBT 0.875, MP 7/7; final fold-in D22-D29 applied to 0.2.2; verdict accepted with recorded debt, no further audit iteration.

## §E.2 Run-phase Evidence

### §F.0 baselines (absorbed tree, re-measured 2026-09-02 before M4's first commit)

Absorbed HEAD: `bf05efb8b` (WT-graph-report; plan artifacts v0.2.2 + absorbed `origin/develop` `58fbc3b5e`; t412 NOT landed — `git merge-base --is-ancestor 6916ee5a7 origin/develop` fails, so M1-M3 stay gated).

| Command | Result |
|---|---|
| `go build ./...` | exit 0 (`BUILD_OK`) |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (`WINBUILD_OK`) |
| `go test -count=1 ./internal/graph/` | `ok github.com/modu-ai/moai-adk/internal/graph 16.413s` |
| `go test -count=1 ./internal/hook/` | `ok github.com/modu-ai/moai-adk/internal/hook 35.714s` (+ `internal/hook/{quality,security,testutil,trace}` all ok) |
| `golangci-lint run --timeout=2m ./internal/graph/... ./internal/hook/... ./internal/cli/...` | `0 issues.` |

Note: internal/cli full-package baseline delegated to CI (per lane-local verification discipline); M4's own targeted cli runs cover the touched surface.

### M4 — deferred SessionStart edges refresh (REQ-GR-010/011/012)

Implemented 2026-09-02 (TDD RED→GREEN; tree = `bf05efb8b` + M4 working-tree changes, pre-commit attribution — commit SHA rides the M4 commit itself).

Files: `internal/hook/session_start.go` (DI seam `DeferredEdgesRefresh` + `WithDeferredEdgesRefresh` option, synchronous staleness snapshot via exported predicates, deferred invocation after advisory send alongside `mxScanNeeded`, fail-open helper) · `internal/cli/graph_refresh_cli.go` (`deferredEdgesRefresh` thin wrapper around `refreshEdgesArtifact` + budget-overrun warning) · `internal/cli/deps.go` (seam wiring at handler construction) · tests: `internal/hook/session_start_deferred_edges_test.go` (5), `internal/cli/graph_deferred_refresh_test.go` (3).

| E-item | Command | Observed output |
|---|---|---|
| E1 AC-GR-015 | `go test -count=1 -v ./internal/hook/ -run TestSessionStartDeferredEdgesRefresh` | `--- PASS:` ×4 (+1 added later: `NilGuardIsNoOp`) → final full-run below |
| E1 AC-GR-015 | `go test -count=1 -v ./internal/cli/ -run TestDeferredEdgesRefresh` | `--- PASS: TestDeferredEdgesRefresh_StaleRefreshesAndStagesNothing (0.35s)` · `--- PASS: TestDeferredEdgesRefresh_FreshNoRewrite (0.21s)` · `--- PASS: TestDeferredEdgesRefresh_BudgetOverrunWarns (0.08s)` · `ok github.com/modu-ai/moai-adk/internal/cli 1.563s` |
| E2 | `go build ./...` | exit 0 (`BUILD_OK`) |
| E2 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (`WINBUILD_OK`) |
| E3 | `go test -count=1 -cover ./internal/hook/` | `ok github.com/modu-ai/moai-adk/internal/hook 35.824s coverage: 85.1% of statements`; new-code per-func: `runDeferredEdgesRefresh 100.0%`, `spawnDeferredAdvisoryScans 90.9%`, `Handle 87.0%` |
| E3 | `go test -count=1 ./internal/cli/ -run TestDeferredEdgesRefresh -coverprofile` | `deferredEdgesRefresh 85.7%`, `graphRefreshOverrun 100.0%` (per-func) |
| E4 | `grep -n 'AskUserQuestion' internal/hook/session_start.go internal/cli/graph_refresh_cli.go internal/cli/deps.go` | 0 matches (exit 1). Package-wide matches are pre-existing comment/doc/testdata lines in untouched files (baseline) |
| E5 | `golangci-lint run --timeout=2m ./internal/hook/... ./internal/cli/...` | first run: 1 NEW errcheck (test `w.Close()` unchecked) → fixed → `0 issues.` (matches `0 issues.` baseline) |
| E8 RED | `go test ./internal/hook/ -run TestSessionStartDeferredEdgesRefresh` (pre-implementation) | `undefined: WithDeferredEdgesRefresh` ×3 · `FAIL github.com/modu-ai/moai-adk/internal/hook [build failed]` |
| E8 RED | `go vet ./internal/cli/` (pre-wrapper) | `undefined: deferredEdgesRefresh` ×3 · exit 1 |

Additional AC-GR-015 assertions covered by name: stale predicate flips false (`edgesRefreshNeeded(...)` false after Handle, in `StaleRefreshesAndStagesNothing`); fresh artifact SHA+mtime unchanged (`FreshNoRewrite`); no staged git entries via `git status --porcelain` in a fixture repo (`StaleRefreshesAndStagesNothing`); over-budget duration injected through `edgesRefreshClock` seam produces the warning (`BudgetOverrunWarns`); nil seam skips step (`NilSeamSkipsStep` + `NilGuardIsNoOp`); fail-open on seam error (`FailOpenOnError`); goleak-clean via the package TestMain (sync-mode tests).

Deviations from plan (justified, not silent):
1. The hook-side probe uses `graph.EdgesSourcesMoved(projectDir)` — the exported DEFAULT-artifact form of `EdgesSourcesMovedFor` — because `internal/graph` exports no default-edges-path constant and hand-building the path in the hook would duplicate it. Same predicate, same artifact the deferred path refreshes; `internal/graph` production code untouched (M1-M3 gate).
2. The AC's full-stack budget-overrun assertion exercises the wrapper directly rather than through Handle: swapping `os.Stderr` across the whole sync Handle would also capture unrelated handler output into the pipe. The wrapper-through-Handle path is otherwise fully exercised by the stale/fresh tests.

Full-suite regression: `go test -count=1 ./internal/hook/` → `ok github.com/modu-ai/moai-adk/internal/hook 35.533s` (post-change). `go test -count=1 ./internal/cli/` (full package, deps.go is shared) → `ok github.com/modu-ai/moai-adk/internal/cli 375.583s`, exit 0. `go vet ./internal/hook/ ./internal/cli/` → exit 0.

### M1 — `graph_shortest_path` MCP tool (REQ-GR-001..004)

Implemented 2026-09-02 (TDD RED→GREEN; tree = `7cad11efd` + M1 working-tree changes, pre-commit attribution — commit SHA rides the M1 commit itself). Absorb gate OPEN: t412 landed and was absorbed (`b6231290d`); plan anchors re-verified post-absorb (`maxTraceDepth` codequery.go:21, `splitCodeNode`/`loadCodeEdges`/`AnswerProvenance` intact; t412's fan-in migration touched no M1-named surface — traversal consumes `KindCodeCall` only, no conflict).

Files: `internal/graph/shortestpath.go` (new — `ShortestPath` + `PathResult`/`PathHop`/`PathCandidate`; deterministic BFS over `KindCodeCall` edges indexed by caller node id, neighbor iteration in total order (node id, then line), depth bounded by the SHARED `maxTraceDepth`; ambiguous intermediate = no continuation) · `internal/cli/mcp_code_tools.go` (`handleGraphShortestPath`: from/to required, `resolveToolProjectRoot`, `toolJSON`/`toolErr`) · `internal/cli/mcp_server.go` (`add("graph_shortest_path", …)` with the cap restated in the description, read-only hint) · `internal/mcp/catalog.go` + `catalog_test.go` (29th entry, `wantCatalogSize` 28→29 — forced by the registration/catalog equality guard) · `.claude/rules/moai/core/moai-mcp-tools.md` + template mirror (Ten-tool `project_root` sentence — forced by `TestProjectRootDocMatchesServer`) · tests: `internal/graph/shortestpath_test.go` (9), `internal/graph/shortestpath_nocgo_test.go` (1, `!cgo`), `internal/cli/mcp_shortest_path_test.go` (7).

| E-item | Command | Observed output |
|---|---|---|
| E1 AC-GR-001 | `go test -count=1 -v ./internal/cli/ -run TestMCPServer_GraphShortestPathRegistered` | `--- PASS: TestMCPServer_GraphShortestPathRegistered (0.00s)` — tool named in tools/list alongside the three existing graph tools, effective read-only true, `from`/`to` in `required`, `project_root` declared but not required, description contains "capped at 8" |
| E1 AC-GR-002 | `go test -count=1 -v ./internal/graph/ -run 'TestShortestPath_EightHopChainFound|TestShortestPath_NineHopChainNotFound'` | both `--- PASS` — 8-hop chain found (HopCount=8 ≤ 8, Cap=8); 9-hop chain structured not-found naming both endpoints + "8", no hops |
| E1 AC-GR-003 | `go test -count=1 -v ./internal/graph/ -run 'TestShortestPath_DisconnectedNotFound|TestShortestPath_AmbiguousEndpointCandidates|TestShortestPath_AmbiguousIntermediateNoContinuation|TestShortestPath_TargetOnlyNameNotFound'` | all `--- PASS` — not-found names endpoints+cap+provenance; candidates = name→sorted node ids, no path; ambiguous intermediate never joined through; target-only name → plain not-found |
| E1 AC-GR-004 | `go test -count=1 -v ./internal/graph/ -run 'TestShortestPath_Deterministic|TestShortestPath_TieBreakTotalOrder'` | both `--- PASS` — 10× byte-identical serializations + `cmp` exit 0; tie breaks via node-id total order (b.go:B, not c.go:C) |
| E1 AC-GR-005 | `go test -count=1 -v ./internal/cli/ -run TestHandleGraphShortestPath_BadRootRejected` | `--- PASS` — bad `project_root` → IsError result naming the rejected root, no fallback |
| E2 | `go build ./...` | exit 0 (`BUILD_OK`) |
| E2 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 (`BUILD_OK`) |
| E3 | `go test -count=1 -coverprofile ./internal/graph/ -run TestShortestPath` + `go tool cover -func` | `ShortestPath 92.7%`, `resolveName 100.0%`, `sortedNeighbors 88.9%`, `reconstruct 100.0%` (per-func, all ≥ 85%) |
| E3 | same for `./internal/cli/ -run 'TestMCPServer_GraphShortestPathRegistered|TestHandleGraphShortestPath'` | `handleGraphShortestPath 92.3%` (per-func) |
| E4 | `grep -rn 'AskUserQuestion\|mcp__askuser' <touched non-test files>` | 0 matches (exit 1) |
| E5 | `golangci-lint run --timeout=2m ./internal/graph/... ./internal/cli/... ./internal/mcp/... ./internal/settings/... ./internal/template/...` | `0 issues.` (baseline `0 issues.` → 0 NEW) |
| B5 nocgo | `CGO_ENABLED=0 go test -count=1 ./internal/graph/ -run 'TestNoCGOShortestPath|TestShortestPath'` | `ok github.com/modu-ai/moai-adk/internal/graph 0.767s` |
| E8 RED | pre-GREEN `go test ./internal/graph/ -run TestShortestPath` | `undefined: ShortestPath` … `FAIL [build failed]`; pre-GREEN cli run: `undefined: handleGraphShortestPath` … `FAIL [build failed]` |

Full-suite regression (post-change): `go test -count=1 ./internal/graph/ ./internal/mcp/` → both `ok`; `go test -count=1 ./internal/cli/ ./internal/settings/` → both `ok` (cli 330.004s — includes the doc-parity + catalog guard tests this milestone's cascades feed); `go test -count=1 ./internal/template/` → `ok` (template mirror parity). `go vet` on graph/cli/mcp → exit 0.

Deviations from plan (justified, not silent):
1. `internal/mcp/catalog.go` + `catalog_test.go` (`wantCatalogSize` 28→29) are not named in plan §F M1's key files, but the registration/catalog equality guard (`TestMoaiMCPServer_RegistrationMatchesCatalog`, plus `TestMoaiMCPTools_Count*`) mechanically fails without the 29th entry — the catalog_test comment itself instructs updating it together with registration. Mechanical cascade of REQ-GR-001.
2. `.claude/rules/moai/core/moai-mcp-tools.md` + its template mirror (Nine→Ten-tool `project_root` sentence, 28→29 header, Code-queries family row) — forced by `TestProjectRootDocMatchesServer`, which checks BOTH copies against the live server schema. Edit kept identical in both copies (verified `diff` → identical); template content stays neutral (tool names only, no SPEC-IDs).
3. Pre-existing drift observed, NOT fixed (out of M1 scope): `moai-mcp-tools-catalogue.md` header still says "each of the 21 tools" — stale from before the graph family landed (28→29). No test pins it; left for a docs sweep.

### M2 — `moai graph report` (REQ-GR-005/006/007)

Implemented 2026-09-02 (TDD RED→GREEN; tree = `31566c117` + M2 working-tree changes, pre-commit attribution — commit SHA rides the M2 commit itself). Absorb gate already open at M1; M2 re-read the shared `internal/cli/graph.go` post-absorb (subcommand wiring pattern intact) and re-measured the §F.0 baselines before the first edit (builds OK, graph/hook/cli targeted tests ok, lint `0 issues.`).

Files: `internal/graph/architecture_report.go` (new — `GodNodes`/`SurprisingConnections`/`ImportCycles`/`RenderArchitectureReport` + `GraphReportRelPath` const; directory proxy via `codeNodeIndex` + `resolveName`, Tarjan SCC over import edges, canonical rotation for simple cycles / member list for branched SCCs) · `internal/cli/graph.go` (`newGraphReportCmd()` under `newGraphCmd()`; flags `--root` + `--limit` ONLY — no `--out`, fixed rotating path per D1 ADOPTED) · `internal/template/templates/.gitignore` (anchored `.moai/reports/graph-report.md` rule alongside the plan-audit rule, Template-First; repo-root `.gitignore` already carries `.moai/reports/*.md`, no edit needed) · tests: `internal/graph/architecture_report_test.go` (9), `internal/graph/architecture_report_nocgo_test.go` (1, `!cgo`), `internal/cli/graph_report_cmd_test.go` (7).

| E-item | Command | Observed output |
|---|---|---|
| E1 AC-GR-006 | `go test -count=1 -v ./internal/cli/ -run TestGraphReportCmd_WritesFixedPathWithSections` | `--- PASS` — report exists at `<fixture>/.moai/reports/graph-report.md`, carries `## God Nodes` / `## Surprising Connections` / `## Import Cycles`, exit 0 |
| E1 AC-GR-007 | `go test -count=1 -v ./internal/graph/ -run TestGodNodes_TieBrokenByNodeID` | `--- PASS` — two packages at fan-in 3 rank at the same tier, order = node id ascending (internal/aaa before internal/yyy) |
| E1 AC-GR-008 | `go test -count=1 -v ./internal/graph/ -run TestSurprisingConnections_BoundaryRankedFirstAmbiguousExcluded` | `--- PASS` — boundary INFERRED edge (internal/cli→internal/hook via directory proxy, conf 0.85) ranks first above the same-confidence intra-package edge (kept out by the cross-package selector); ambiguous bare callee ("Dup", 2 nodes) excluded; non-INFERRED (extracted) edge excluded |
| E1 AC-GR-009 | `go test -count=1 -v ./internal/graph/ -run TestImportCycles_SCCShapeAndCanonicalRotation` | `--- PASS` — 3 SCCs (simple cycle + branched + self-loop); simple cycle renders canonical rotation `pkg/a,pkg/c,pkg/b` (edge-following from smallest, NOT the sorted list); branched SCC renders sorted member list, no fabricated cycle; SCC count is the primary datum; acyclic packages never reported |
| E1 AC-GR-010 | `go test -count=1 -v ./internal/cli/ -run TestGraphReportCmd_EmptyCodeLayerStillEmits` + `CGO_ENABLED=0 go test -count=1 -v ./internal/graph/ -run TestNoCGOArchitectureReportEmptySections` | both `--- PASS` — doc-only artifact: report emitted at the fixed path, sections present-but-empty with `code layer absent: CGO disabled or no extraction`, exit 0 |
| E1 AC-GR-011 | `go test -count=1 -v ./internal/cli/ -run TestGraphReportCmd_DoubleRunByteIdentical` + shell double-run `cmp` | test `--- PASS` (byte-equal reads); shell: two `moai graph report` runs on the same fixture (binary built from this tree, fixture under `.moai/state/verify/m2-report/fixture`), `cmp run1.md graph-report.md` → `cmp_exit=0` |
| E1 AC-GR-016 | `go test -count=1 -v ./internal/graph/ -run 'TestGodNodes_MixedKindAggregationAndKindsLabel|TestRenderArchitectureReport_SectionsAndKindsLine'` + `git diff 31566c117 -- internal/hook/mx/validator.go \| wc -l` | tests `--- PASS` — section renders `fan-in over: code-call, import` naming the counted kinds; validator diff = **0 lines** (`fanInIndex` byte-unchanged; whole `internal/hook/` tree untouched by M2 — `git diff --stat 31566c117 -- internal/hook/` empty) |
| E2 | `go build ./...` | `BUILD_DARWIN_OK` (exit 0) |
| E2 | `GOOS=windows GOARCH=amd64 go build ./...` | `BUILD_WINDOWS_OK` (exit 0) |
| E3 | `go test -count=1 -coverprofile ./internal/graph/ ./internal/cli/` + `go tool cover -func` | graph pkg `coverage: 88.3% of statements`; new-code per-func: `GodNodes 100.0%`, `ImportCycles 100.0%`, `codeNodeIndex/packageDir/sortedKindSet/SimpleCycle/formatConfidence 100.0%`, `SurprisingConnections 95.7%`, `RenderArchitectureReport 95.5%`; `newGraphReportCmd 91.3%` (all ≥ 85%) |
| E4 | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/graph/architecture_report*.go internal/cli/graph_report_cmd_test.go`; `grep -c AskUserQuestion internal/cli/graph.go` | 0 matches / count `0` |
| E5 | `golangci-lint run --timeout=2m ./internal/graph/... ./internal/cli/... ./internal/template/...` | `0 issues.` (baseline `0 issues.` → 0 NEW) |
| E8 RED | pre-GREEN `go test ./internal/graph/ -run 'TestGodNodes|TestSurprisingConnections|TestImportCycles|TestRenderArchitectureReport'` | `undefined: GodNodes` ×5 · `undefined: SurprisingConnections` · `undefined: ImportCycles` · `undefined: ImportSCC` · `undefined: RenderArchitectureReport` · `FAIL github.com/modu-ai/moai-adk/internal/graph [build failed]` |
| E8 RED | pre-GREEN `go test ./internal/cli/ -run TestGraphReportCmd` | `undefined: newGraphReportCmd` · `FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]` |

Full-suite regression (post-change, env-scrubbed): `go test -count=1 ./internal/graph/ ./internal/cli/` → `ok …internal/graph 16.268s` · `ok …internal/cli 337.594s`; nocgo leg `CGO_ENABLED=0 go test ./internal/graph/ -run TestNoCGO` → `ok`. `go build` both targets exit 0. Preserve surfaces verified untouched: `git diff 31566c117 -- internal/graph/report.go internal/graph/shortestpath*.go internal/hook/` → empty.

Deviations from plan (justified, not silent):
1. REQ-GR-005's surprising-connections ranking clause ("ranked above same-confidence intra-package edges") is realized through the cross-package SELECTOR, per plan §F M2's implementation contract ("INFERRED code-call edges whose endpoints' package directories differ, ranked by confidence then total order"): a boundary edge outranks every same-confidence intra-package edge by keeping intra-package edges out of the section entirely — the section never carries non-surprising entries on a real tree (they would be noise). The AC-GR-008 assertion (boundary ranks first) holds under this reading and is tested.
2. `GodNodesResult` carries a section-level kinds union rather than per-node kinds — AC-GR-016's example ("fan-in over: code-call, import") names the section's counted kinds; per-node provenance would render the same information N times.

### M3 — edges shrink guard (REQ-GR-008/009)

Implemented 2026-09-02 (TDD RED→GREEN; tree = `2cea86ec2` + M3 working-tree changes, pre-commit attribution — commit SHA rides the M3 commit itself). Absorb gate open since M1; M3 re-measured the baselines before the first edit (`go build` both targets exit 0; `golangci-lint run ./internal/graph/... ./internal/cli/...` → `0 issues.`; `go test ./internal/graph/...` both ok; `go test ./internal/cli/` → `ok … 301.302s`).

Files: `internal/graph/shrink.go` (new — `DetectUnexplainedShrink` + `ShrinkReport`/`ShrinkDefect`/`ShrinkRefusalError`; set-difference trigger on (Kind,Source,Target) identity; kind scope = code-call + code-import; `splitCodeNode` decode; project-relative validation incl. symlink-escape; `*graph.ShrinkRefusalError` typed refusal) · `internal/graph/symbol/symbol.go` (`Extract` also returns the scanned-file list, captured inside the walk's own iteration — no second scan) · `internal/graph/symbol.go` (`BuildWithCodeLayers`/`BuildWithCodeLayersMode` return `scanned` alongside edges+matrix; fail-open extraction path returns an EMPTY scanned set, deliberate) · `internal/graph/meta.go` (`writeMetaFile` temp hardened to per-refresh unique suffix via `os.CreateTemp(dir, base+".*.tmp")` + 0644 chmod — NOT pid-based, closes the same-process double-refresh collision, D20) · `internal/cli/graph_refresh_cli.go` (guard between `BuildWithCodeLayers` and `WriteJSONL` — pre-write, refusal = zero writes; unreadable prior artifact skips the guard) · `internal/cli/graph.go` (query path: `errors.As(*ShrinkRefusalError)` → stated shrink-refusal warning + answer from existing artifact, exit 0; build path: same guard pre-write → non-zero exit naming removed edges) · tests: `internal/graph/shrink_test.go` (10), `internal/graph/shrink_extraction_test.go` (1, cgo, real-extraction shapes), `internal/graph/shrink_nocgo_test.go` (2, `!cgo`), `internal/graph/meta_tmp_test.go` (4), `internal/cli/graph_shrink_test.go` (5, cgo) + mechanical arity updates in 10 pre-existing test files (`BuildWithCodeLayers` now returns 4 values).

| E-item | Command | Observed output |
|---|---|---|
| E1 AC-GR-012 (1st clause) | `go test -count=1 -v ./internal/cli/ -run TestShrinkGuard_QueryPathRefusedAnswersFromExisting` | `--- PASS` — stderr carries `unexplained shrink` + `rootlevel.go:Fn`; answer from EXISTING artifact (`callers of InjectedTarget: 1` incl. the injected edge); edges.jsonl + meta SHA-identical; `cmd.Execute()` returns nil (exit 0) |
| E1 AC-GR-012 (2nd clause) | `go test -count=1 -v ./internal/cli/ -run TestShrinkGuard_BuildPathRefuses` | `--- PASS` — build exits non-zero naming `rootlevel.go:Fn` / `InjectedTarget` / `rootlevel.go`; edges.jsonl + meta SHA-identical (zero writes) |
| E1 AC-GR-013 (deletion) | `go test -count=1 -v ./internal/cli/ -run TestShrinkGuard_GenuineDeletionProceeds` | `--- PASS` — overwrite proceeds; removed edges = exactly the deleted `internal/demo/calls.go`'s edges (code-call + code-import), added 0 |
| E1 AC-GR-013 (kind scope) | `go test -count=1 -v ./internal/graph/ -run TestDetectUnexplainedShrink_KindScopeSkipsNonFileKinds` | `--- PASS` — removed doc-import (Source `internal/demo`, a REAL on-disk directory) + spec-depends (Source SPEC ID) edges produce an EMPTY report — never stat'ed |
| E1 D25 mutant | `go test -count=1 -v ./internal/graph/ -run TestDetectUnexplainedShrink_EqualCardinalityMutant` | `--- PASS` — equal-cardinality remove+add (same total count) refuses via the set-difference trigger |
| E1 D23 decode/real shapes | `go test -count=1 -v ./internal/graph/ -run 'TestDetectUnexplainedShrink_DecodeStep|TestDetectUnexplainedShrink_RealExtractionPartialFailure'` | both `--- PASS` — a file literally named `calls.go:Calls` never matches (the DECODED path is stat'ed, not the compound string); real `BuildWithCodeLayers` output partial-failure fixture refuses naming the real compound Source |
| E1 parse-skipped variant | `go test -count=1 -v ./internal/cli/ -run TestShrinkGuard_ScannedSetIsExtractionWalk` + `TestShrinkGuard_QueryPathRefusedAnswersFromExisting` fixture | both `--- PASS` — a Go file on disk outside the described roots (the extraction-skipped shape: file exists, outside the scanned set) refuses the same way on both write paths |
| E1 AC-GR-014 | `go test -count=1 -v ./internal/cli/ -run TestShrinkGuard_QueryPathRefusedAnswersFromExisting` | `--- PASS` (see AC-GR-012 1st clause row — same test: warning + existing-artifact answer + exit 0) |
| E1 REQ-GR-009 deferred inheritance | `go test -count=1 -v ./internal/cli/ -run TestShrinkGuard_DeferredPathInheritsRefusal` | `--- PASS` — `deferredEdgesRefresh` surfaces `*graph.ShrinkRefusalError` via `errors.As`, artifact byte-identical; guard inherited through the M4-wrapped `refreshEdgesArtifact`, NOT re-wired |
| E1 B5 nocgo deliberate | `CGO_ENABLED=0 go test -count=1 -v ./internal/graph/ -run TestDetectUnexplainedShrink_NoCGO` | `--- PASS: TestDetectUnexplainedShrink_NoCGOEmptyScannedSetRefuses` (empty scanned set + existing cgo artifact → REFUSE, pinned as deliberate) `+ --- PASS: …DocEdgesUnaffected` (doc-only shrink never refuses) |
| E1 D20 tmp hardening | `go test -count=1 -race ./internal/graph/ -run TestWriteEdgesMeta_Concurrent` (×3) + `-count=5` | `ok` — 8 concurrent same-process meta writes all succeed, parseable sidecar, no residual `.tmp`, `-race` clean |
| E2 | `go build ./...` | `BUILD_DARWIN_OK` (exit 0) |
| E2 | `GOOS=windows GOARCH=amd64 go build ./...` | `BUILD_WINDOWS_OK` (exit 0); `GOOS=windows GOARCH=amd64 go vet ./internal/graph/... ./internal/cli/` exit 0 (tests cross-compile) |
| E3 | `go test -count=1 -coverprofile ./internal/graph/` + `go tool cover -func` | new-code per-func: `DetectUnexplainedShrink 94.7%`, `Empty/Error/shrinkGuardedKinds/edgeID/decodeSourceFile 100.0%`, `Describe 100.0%`, `isSafeProjectRelative 87.5%`, `fileExistsUnderRoot 91.7%`; `writeMetaFile 58.3%` (see deviations §3) |
| E3 | `go test -count=1 -coverprofile ./internal/cli/` (full suite, exit 0) | guard-insertion sites: `deferredEdgesRefresh 100.0%`, `newGraphQueryCmd 92.1%`, `newGraphBuildCmd 84.8%`, `refreshEdgesArtifact 76.5%` (all NEW branches — refusal + pass — exercised; uncovered lines are pre-existing fault-injection-only error returns) |
| E4 | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/graph/shrink.go internal/graph/meta.go internal/graph/symbol.go internal/graph/symbol/symbol.go internal/cli/graph_refresh_cli.go internal/cli/graph.go` | 0 matches; `TestGraphCmd_NoAskUserQuestion` static guard passed in the full cli run |
| E5 | `golangci-lint run ./internal/graph/... ./internal/cli/...` | `0 issues.` (baseline `0 issues.` → 0 NEW) |
| E8 RED | pre-GREEN `go test ./internal/graph/ -run TestDetectUnexplainedShrink` | `undefined: DetectUnexplainedShrink` ×10 + `undefined: ShrinkRefusalError` · `FAIL github.com/modu-ai/moai-adk/internal/graph [build failed]` |
| E8 RED | pre-wiring `go test ./internal/cli/ -run TestShrinkGuard` | `build over an unexplained shrink must exit non-zero` (build path overwrote) · query: stderr without refusal warning, `callers of InjectedTarget: 0`, edges.jsonl + meta SHA CHANGED (silent shrunk write) · deferred: `got: <nil>` — 3 FAIL, exit 1 |
| E8 RED | pre-hardening `go test ./internal/graph/ -run TestWriteEdgesMeta_Concurrent` | `concurrent meta write must not collide … rename …/edges.meta.json.tmp …: no such file or directory` — the fixed-`.tmp` collision, FAIL |

Full-suite regression (post-change): `go test -count=1 ./internal/graph/...` → both `ok` (graph 16.602s, symbol 0.620s); `go test -count=1 ./internal/cli/` → `ok … 325.662s` (cgo leg). `go test -race` on the shrink/meta test set → `ok`. Preserve surfaces verified untouched: `git status` shows no `internal/hook`, no M1/M2 files (shortestpath*, architecture_report*, report cmd) beyond the named wiring points in `internal/cli/graph.go`.

Deviations from plan (justified, not silent):
1. Kind scope realized as `{code-call, code-import}` — the kinds whose Source is a file path drawn from the SAME extraction walk that produces the scanned set. `mx-spec` edges carry file-path Sources but their rebuild input is the mx-index scan universe (a different input); testing them against the extraction's scanned set would misreport ordinary refreshes, so they are out of scope on the same reasoning the REQ gives for doc-import/spec-depends. Stated in the `shrinkGuardedKinds` code contract.
2. Edge identity for the set difference is (Kind, Source, Target) — positional (Line) and annotation fields are not identity, so a line shift inside a SCANNED file never manufactures a removal (an unscanned file's edges are flagged wholesale, which is the defect being caught).
3. `writeMetaFile` per-func coverage 58.3%: the three uncovered branches (write/close/chmod failure AFTER a successful `CreateTemp`) are reachable only via filesystem fault injection (disk-full class); the happy path and two failure branches (mkdir-blocked parent, read-only dir) are covered, and the concurrent-collision branch — the milestone's actual change — is covered deterministically.
4. Pre-existing reds observed, NOT fixed (out of M3 scope, verified at HEAD `2cea86ec2` in a detached temp worktree): `CGO_ENABLED=0 go test ./internal/graph/symbol/` (3 tests) and `CGO_ENABLED=0 ./internal/cli/` (`TestGraphTools_HonorProjectRoot`, `TestHandleGraphFileAPI`, `TestHandleGraphFindAndTrace`) — nocgo legs that expect cgo extraction output; M3's own nocgo-scoped tests pass.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-09-02
```

Run phase complete: all four milestones (M4 `258013d0a` → M1 `31566c117` → M2 `2cea86ec2` → M3 `cfe86675c`, branch `WT-graph-report`, unpushed) carry their E1-E8 attribution matrices in §E.2 — every AC row PASS with command + verbatim output, both cross-platform builds exit 0, lint clean against the `0 issues.` baseline, and RED evidence captured pre-GREEN for each TDD cycle.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: complete
sync_complete_at: 2026-09-02
sync_commit_sha: "pending-backfill"
b12_self_test_a: "grep -c 'SPEC-GRAPH-REPORT-001' CHANGELOG.md → 0 before emission (rc=1), duplicate-entry gate passed"
b12_self_test_b: "AC count in acceptance.md = 16 (grep -oE 'AC-[A-Z0-9-]+' | sort -u | wc -l = 16); CHANGELOG entry cites no per-AC count, matches SSOT"
b12_self_test_c: "claimed file paths verified by ls (internal/graph/, internal/cli/, internal/hook/session_start.go present)"
canary_compliance_check:
  changelog_entry_position: "top of [Unreleased] → ### Added, matching existing entry style"
  frontmatter_status_transitions: "in-progress → implemented → completed merged into this single sync commit"
  body_content_modified: "none (spec/plan/acceptance bodies untouched; frontmatter status: + updated: only)"
```

### §E.2-MX — MX tag change summary (sync sub-step)

New-code `@MX` annotations introduced by this branch's milestone commits (measured: `git diff b6231290d..HEAD -- internal/graph internal/cli internal/hook | grep '^+' | grep '@MX:'` → 6 added lines):

- **@MX:NOTE × 3** — `internal/cli/graph.go` (fixed rotating report path — D1: regenerates in place, never committed; SPEC-GRAPH-REPORT-001), `internal/graph/architecture_report.go` (determinism contract — stable sorts + total order, no wall-clock; REQ-GR-005), `internal/hook/session_start.go:769` (deferred edges refresh — fail-open, fire-and-forget at process lifetime; SPEC-GRAPH-REPORT-001).
- **@MX:ANCHOR × 1** — `internal/graph/shrink.go:101` — single choke point for every automatic write path's shrink evaluation (refresh / build / deferred all consume one verdict), with `@MX:REASON` (forking the discriminator would let one path accept a shrink another refuses — the graphify #1116 incident shape) and `@MX:SPEC:SPEC-GRAPH-REPORT-001` sub-lines.
- **@MX:WARN × 0 new** — no new dangerous-pattern sites; `internal/hook/session_start.go`'s existing `@MX:WARN` lines predate this SPEC.
- **M1 (`graph_shortest_path`, `internal/graph/shortestpath.go`) carries no @MX tag** — read-only query following the existing `graph_file_api`/`graph_find_code`/`graph_trace_calls` registration pattern, which likewise carries none; no exported-symbol fan-in ≥ 3 site, so the P1 ANCHOR rule does not trigger.

Sync-phase deliverables: CHANGELOG `[Unreleased]` entry (4 deliverables, MCP tool count 28 measured via `grep -c 'add("' internal/cli/mcp_server.go` at `cfe86675c` — NOT the 29 claimed in the dispatch, which was not re-counted), spec.md frontmatter `completed` transition, codemaps skipped (no `codemaps/` directory exists in this tree). §E.4 `sync_commit_sha` carries the `pending-backfill` placeholder per the D3 backfill exemption — a commit cannot know its own SHA; the orchestrator backfills in a follow-up commit.

## §F Phase 4 Mode Selection

Input parameters: tier M · scope ≈ 4 files (session_start.go, hook options, cli refresh wiring, tests) · domains 2 (internal/hook, internal/cli) · language mix 100% Go · concurrency benefit LOW (coding-heavy, async lifecycle) · agent-team prereqs not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | semantic multi-file change, not a one-liner |
| serial | **YES** | coding-heavy per-milestone delegation (Anthropic coding-task caveat) |
| fanout | no | <3 domains, no independent research lenses |
| sweep | no | semantic new code, ~4 files — far under the mechanical ≥~30-file bar |

Decision: serial

Justification: M4 is a single cohesive async-lifecycle change (DI seam + deferred goroutine step + seam-injected tests) across two packages with tight coupling between the hook handler and its option pattern — one manager-develop spawn per milestone keeps the TDD loop coherent and the gate boundary (M1-M3 gated on t412 landing, M4 ungated) explicit. Boundary case: M4 executes before M1-M3 (ungated per spec §B.2) — milestone execution order M4 → (post-landing) M1 → M2 → M3 deviates from plan §F numbering by design, recorded here per §D.3.
