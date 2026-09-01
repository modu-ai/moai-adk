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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

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
