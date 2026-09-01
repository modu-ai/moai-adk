---
id: SPEC-GRAPH-REPORT-001
title: "Implementation plan — graph report toolchain (P2-4..P2-7 adoption)"
version: "0.1.0"
created: 2026-09-02
author: manager-spec
---

## §A. Context

- Work tree: card t413 worktree (`.claude/worktrees/t413`), branched from origin/develop tip `9145806d8`. SPEC artifacts: `.moai/specs/SPEC-GRAPH-REPORT-001/{spec,plan,acceptance,progress}.md`.
- Method: TDD (`constitution.development_mode: tdd`) — every REQ lands RED-first, one behavior per test.
- Existing infrastructure to EXTEND: 3-tool MCP registration pattern (`internal/cli/mcp_server.go:486-509` + handlers in `mcp_code_tools.go`), query layer (`internal/graph/codequery.go`), refresh path (`internal/cli/graph_refresh_cli.go`), deferred advisory scans (`internal/hook/session_start.go:680-702`), duration seam (`edgesRefreshClock`), staleness predicate (`edgesRefreshNeeded`).
- Existing infrastructure to PRESERVE: `WriteJSONL` byte-stable write, `edges.meta.json` sidecar contract, `EdgesSourcesMovedFor`/`MXIndexNeedsRefresh` predicates, `WithSynchronousDeferredScans` + `deferredScansAsyncEnabled` test seams, the grep-based MX validator fan-in.
- Dependency: SPEC-MX-TAG-EDGES-001 (t412) lands mx-tag edge layer on develop before M1-M3. M4 is gate-free.

## §B. Known Issues (domain-relevant subset)

- **B1 cross-platform**: new code is pure stdlib file/string work; verify `GOOS=windows GOARCH=amd64 go build ./...`; no syscall.
- **B2 cross-SPEC policy conflict**: t412 is mid-flight in the same packages. The absorb gate (§F) exists precisely because `mcp_code_tools.go`, `cli/graph.go`, `graph/query.go`, `graph/graph.go` will have moved on develop. Re-read the shared files after absorbing; never plan against this worktree's pre-absorb copies.
- **B3/B11 subagent boundary**: no AskUserQuestion anywhere in `internal/cli` / `internal/graph` / `internal/hook`; error paths return structured results (per `cli/CLAUDE.md` C-HRA-008).
- **B5 CI tiers**: `CGO_ENABLED=0` legs will exercise "graph layer absent" paths — tests must skip-or-assert deliberately, not fail (predecessor REQ-GFR-001/010 pattern).
- **B8 working-tree hygiene**: `.moai/reports/` and `.moai/project/graph/` are untracked derived surfaces; tests write only under `t.TempDir()`.
- **B12**: sync-phase CHANGELOG entry counts MCP tools accurately (28 → 29 with the 4th graph tool — re-count at sync, never carry this plan's number).

## §C. Pre-flight

```bash
git branch --show-current && git rev-parse --short HEAD
git merge-base --is-ancestor <t412-merge-sha> HEAD && echo "t412 absorbed" || echo "GATE: absorb required"
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m ./internal/graph/... ./internal/cli/... ./internal/hook/... 2>&1 | tail -5
```

GATE line not green ⇒ M1-M3 do not start; M4 may proceed (no shared files).

## §D. Constraints (DO NOT VIOLATE)

- Absorb gate: M1-M3 AFTER `git merge origin/develop` post-t412 + re-measure (§F.0).
- No new Go dependencies; path search and cycle detection in stdlib.
- Determinism: stable sorts, total order (node id, line); no wall-clock in output bodies.
- Targeted tests only (`go test ./internal/<pkg>/...`); never the local full suite; CI judges.
- `t.TempDir()` isolation; no goroutine leaks (join via the `completed` seam in SessionStart tests).
- Do not modify the MX validator, its fan-in, or any `.claude/` surface.
- Conventional Commits (`feat(SPEC-GRAPH-REPORT-001): M{N} …`), per-M commits; `🗿 MoAI` trailer.

## §E. Self-Verification (delegation contract)

Each §E item reported with the attribution triple (command / verbatim output / tree SHA):

- **E1** AC binary PASS/FAIL matrix against `acceptance.md` (AC-GR-001..015).
- **E2** `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` — both exit 0.
- **E3** `go test -cover ./internal/graph/... ./internal/cli/... ./internal/hook/...` — new code ≥85%.
- **E4** Subagent-boundary grep 0 matches on touched packages.
- **E5** Lint: NEW issues vs recorded baseline, separately named.
- **E6** Commit SHAs + push state.
- **E8** Verbatim RED outputs per milestone (TDD): captured before GREEN, one per REQ.
- Determinism proof: report generated twice on the same fixture, `cmp` exit 0 (AC-GR-005 evidence).

## §F. Milestones

Ordered by decision-reversibility: M1's tool contract (new user-facing MCP surface) is the most likely to change and is reviewed first; M4 is the most mechanical and lands last. M4 carries no absorb gate and MAY execute while t412 is still in flight; M1-M3 execute in order after the gate.

### §F.0 Absorb + re-measure (gate step for M1-M3, not a milestone)

1. Confirm t412 merged to develop (`git fetch origin && git merge-base --is-ancestor` per §C).
2. `git merge origin/develop` in the card worktree.
3. Re-measure on the absorbed tree: re-read the shared files (`mcp_code_tools.go`, `cli/graph.go`, `graph/query.go`, `graph/graph.go`, `graph/reader.go`); re-run `go build ./...`, targeted `go test` on the three packages, and re-baseline lint. Record absorbed HEAD SHA + measured baselines in `progress.md §E.2` before M1's first commit. Any conflict between this plan and the absorbed tree ⇒ blocker report, not a silent adaptation.

### M1 — `graph_shortest_path` MCP tool (REQ-GR-001..004) [gated]

Scope: graph query function + MCP registration + tests.

Key files:
- `internal/graph/codequery.go` (or a sibling new file) — `ShortestPath(projectRoot, from, to string) (PathResult, error)`: BFS over `KindCodeCall` edges indexed by caller function, neighbor iteration in total order (node id, then line), depth bounded by the shared `maxTraceDepth` const; returns the hop list, a not-found result shape, and provenance via `AnswerProvenance`. Edge confidence rides along per edge (consumes the landed EDGE-CONFIDENCE fields).
- `internal/cli/mcp_code_tools.go` — `handleGraphShortestPath`: `from`/`to` required strings, optional `project_root` via `resolveToolProjectRoot`, `toolJSON`/`toolErr` shaping.
- `internal/cli/mcp_server.go` — `add("graph_shortest_path", …)` with description restating the 8-hop cap, read-only hint annotation.

Tests: registration/discovery; happy path on a chain fixture; >8-hop chain → structured no-path; disconnected → no-path; determinism (two runs byte-identical); bad `project_root` rejected; absent artifact → actionable "run 'moai graph build'" error; `CGO_ENABLED=0` skip/absent semantics.

### M2 — `moai graph report` (REQ-GR-005..007) [gated]

Scope: report computation + CLI subcommand + output artifact.

Key files:
- `internal/graph/report.go` (new) — pure functions: `GodNodes(edges, limit)` (distinct-source fan-in per target over code-call + import layers, labeled by kinds counted), `SurprisingConnections(edges)` (INFERRED code-call edges whose `splitCodeNode(source)` file's package differs from the target's package, ranked by confidence then total order), `ImportCycles(edges)` (import-edge cycle enumeration with canonical rotation — smallest node id first), all returning stable-sorted results.
- `internal/cli/graph.go` — `newGraphReportCmd()` added to `newGraphCmd()`; `--root`, `--out` (default `.moai/reports/graph-report.md` under the resolved root), `--limit`; renders the three sections deterministically; empty sections emitted with stated reason (REQ-GR-006).

[NEEDS CLARIFICATION: report artifact naming — fixed rotating `.moai/reports/graph-report.md` (proposed; deterministic name, gitignored like all reports) vs a dated directory per the `.moai/reports/{TYPE}-{DATE}/` convention. Resolve via orchestrator AskUserQuestion before run-phase entry.]

Tests: fixture with known fan-in ranking (ties broken by id); boundary-INFERRED edge ranked above same-confidence intra-package edge; cycle fixture (a→b→c→a) rendered with canonical rotation; empty code layer (nocgo fixture) still emits with reason; `cmp` of two consecutive runs byte-identical.

### M3 — edges shrink guard (REQ-GR-008..009) [gated — enforcement sites are unshared, but verification depends on the absorbed edge inventory; re-measure applies]

Scope: guard function + enforcement at the automatic write paths.

Key files:
- `internal/graph/shrink.go` (new) — `DetectUnexplainedShrink(existing, rebuilt []Edge, scannedSources map[string]bool) ShrinkReport`: names removed edges whose `splitCodeNode(Source)`/edge source file lies outside `scannedSources`; empty report = overwrite permitted.
- `internal/cli/graph_refresh_cli.go` — `refreshEdgesArtifact` computes the guard between `BuildWithCodeLayers` and `WriteJSONL`; on refusal: skip both writes, return a typed refusal carrying the report; the query path warns and answers from the existing artifact (REQ-GR-009 fail-safe shape). The build path (`internal/cli/graph.go` build command, shared file) applies the same guard and exits non-zero naming the removed edges.
- `scannedSources` derives from the rebuild's own source inventory (the fingerprint source set `SourceFingerprintsForEdges` already enumerates).

Tests: injected partial failure (rebuild missing one source's edges) → refusal, prior artifact SHA-identical, message names the edges; legitimate shrink (source genuinely deleted → its edges removed, source inside the rebuild set) → overwrite proceeds; both automatic paths covered; no-path "prior artifact intact" assertion via byte comparison.

### M4 — deferred SessionStart edges refresh (REQ-GR-010..012) [ungated]

Scope: extend the deferred scan pattern to the edges layer.

Key files:
- `internal/hook/session_start.go` — in `spawnDeferredAdvisoryScans`'s goroutine, after the advisory send and alongside the `mxScanNeeded` durable side effect: when `edgesRefreshNeeded(...)` (reusing the existing predicate + `DefaultThresholds()` red line), call a `runDeferredEdgesRefresh(projectDir)` that wraps `refreshEdgesArtifact` with the `edgesRefreshClock` duration seam and the budget-overrun warning; failure logs and exits the step (fail-open). Computed synchronously like `mxScanNeeded` (cheap fingerprint probe) so the goroutine reads a snapshot, per the existing comment contract.
- Tests: `WithSynchronousDeferredScans` path asserts edges refreshed when stale, untouched when fresh; goleak-clean join; `git status --porcelain` shows no staged entries after refresh (fixture repo); injected over-budget duration produces the warning (reuses the duration seam — no real-timing dependence).

## §G. Anti-Patterns (blocked)

- Planning M1-M3 against pre-absorb file contents; "the shared files probably didn't move" is the exact race the gate exists for.
- A second, parallel rebuild path in the hook layer instead of wrapping `refreshEdgesArtifact`.
- An override flag that forces a known-shrunk write past the guard.
- Wall-clock timestamps in report/tool bodies (breaks determinism ACs).
- Switching the MX validator's fan-in source "while we're in there" (t412's gate, not this SPEC's).
- Real-timing assertions on the deferred refresh (inject through the seam).

## §H. Cross-References

- SPEC: `.moai/specs/SPEC-GRAPH-REPORT-001/spec.md` (REQ-GR-001..012) · ACs: `acceptance.md` (AC-GR-001..015)
- Source analysis: `.moai/reports/graphify-codegraph-analysis-20260901.html` (P2-4..P2-7)
- Dependency: SPEC-MX-TAG-EDGES-001 (t412, branch `WT-mx-tag-edges`)
- Anchors: `codequery.go:21` (maxTraceDepth) · `graph_refresh_cli.go:53` (refreshEdgesArtifact) · `session_start.go:680` (spawnDeferredAdvisoryScans) · `mcp_server.go:486-509` (registration pattern)
