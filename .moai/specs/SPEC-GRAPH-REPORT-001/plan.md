---
id: SPEC-GRAPH-REPORT-001
title: "Implementation plan — graph report toolchain (P2-4..P2-7 adoption)"
version: "0.2.2"
created: 2026-09-02
author: manager-spec
---

## §A. Context

- Work tree: card t413 worktree (`.claude/worktrees/t413`), branched from origin/develop tip `9145806d8`. SPEC artifacts: `.moai/specs/SPEC-GRAPH-REPORT-001/{spec,plan,acceptance,progress}.md`.
- Method: TDD (`constitution.development_mode: tdd`) — every REQ lands RED-first, one behavior per test.
- Existing infrastructure to EXTEND: 3-tool MCP registration pattern (`internal/cli/mcp_server.go:486-509` + handlers in `mcp_code_tools.go`), query layer (`internal/graph/codequery.go`), refresh path (`internal/cli/graph_refresh_cli.go`), deferred advisory scans (`internal/hook/session_start.go:680-702`), duration seam (`edgesRefreshClock`), staleness predicates (`EdgesSourcesMovedFor` / `MXIndexNeedsRefresh` — the exported constituents the cli-private refresh wrapper composes), hook handler option pattern (`Option` / `WithSynchronousDeferredScans` — the DI seam for `DeferredEdgesRefresh`).
- Existing infrastructure to PRESERVE: `WriteJSONL` byte-stable write, `edges.meta.json` sidecar contract, `EdgesSourcesMovedFor`/`MXIndexNeedsRefresh` predicates, `WithSynchronousDeferredScans` + `deferredScansAsyncEnabled` test seams, the grep-based MX validator fan-in (`fanInIndex`, `internal/hook/mx/validator.go`), `internal/graph/report.go` (the existing t107 report cross-check edge layer — M2's new file is named `architecture_report.go` to avoid the collision).
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
- Do not modify the MX validator, its `fanInIndex`, or any `.claude/` surface.
- Conventional Commits (`feat(SPEC-GRAPH-REPORT-001): M{N} …`), per-M commits; `🗿 MoAI` trailer.

## §E. Self-Verification (delegation contract)

Each §E item reported with the attribution triple (command / verbatim output / tree SHA):

- **E1** AC binary PASS/FAIL matrix against `acceptance.md` (AC-GR-001..016).
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
3. Re-measure on the absorbed tree: re-read the shared files (`mcp_code_tools.go`, `cli/graph.go`, `graph/query.go`, `graph/graph.go`, `graph/reader.go`); re-run `go build ./...`, targeted `go test` on the three packages, and re-baseline lint. This re-measure applies to M4 as well: if M4 executed pre-absorb (permitted by §B.2 of spec.md), re-run M4's tests on the absorbed tree after the merge — the DI seam M4 wraps is cli-wired, and an absorbed tree may have moved that wiring. Record absorbed HEAD SHA + measured baselines in `progress.md §E.2` before M1's first commit. Any conflict between this plan and the absorbed tree ⇒ blocker report, not a silent adaptation.

### M1 — `graph_shortest_path` MCP tool (REQ-GR-001..004) [gated]

Scope: graph query function + MCP registration + tests.

Key files:
- `internal/graph/codequery.go` (or a sibling new file) — `ShortestPath(projectRoot, from, to string) (PathResult, error)`: BFS over `KindCodeCall` edges indexed by caller function, neighbor iteration in total order (node id, then line), depth bounded by the shared `maxTraceDepth` const; returns the hop list, a not-found result shape, and provenance via `AnswerProvenance`. Edge confidence rides along per edge (consumes the landed EDGE-CONFIDENCE fields). Intermediate-hop expansion follows the REQ-GR-003 rule deterministically: an ambiguous intermediate name on a candidate path (a bare callee matching 2+ nodes) is never joined through the bare-name caller index — it is treated as no continuation from that node — and a duplicate-name intermediate fixture covers this at test time.
- `internal/cli/mcp_code_tools.go` — `handleGraphShortestPath`: `from`/`to` required strings, optional `project_root` via `resolveToolProjectRoot`, `toolJSON`/`toolErr` shaping.
- `internal/cli/mcp_server.go` — `add("graph_shortest_path", …)` with description restating the 8-hop cap, read-only hint annotation.

Tests: registration/discovery; happy path on an exactly-8-hop chain fixture; >8-hop chain → structured not-found naming endpoints + cap; disconnected → no-path; ambiguous bare-name endpoint (same function name in two files) → structured candidates list, no path; determinism (two runs byte-identical); bad `project_root` rejected; absent artifact → actionable "run 'moai graph build'" error; `CGO_ENABLED=0` skip/absent semantics.

### M2 — `moai graph report` (REQ-GR-005..007) [gated]

Scope: report computation + CLI subcommand + output artifact.

Key files:
- `internal/graph/architecture_report.go` (new — named to avoid the existing t107 `internal/graph/report.go` collision) — pure functions: `GodNodes(edges, limit)` (distinct-source fan-in per target over code-call + import layers, labeled by kinds counted), `SurprisingConnections(edges)` (INFERRED code-call edges whose endpoints' package directories differ, ranked by confidence then total order), `ImportCycles(edges)` (import edges grouped into strongly connected components; one rendered entry per SCC — the canonical member list, smallest node id first; a branched SCC can contain no simple cycle through all members, so membership, not a cycle, is rendered), all returning stable-sorted results.
- `internal/cli/graph.go` — `newGraphReportCmd()` added to `newGraphCmd()`; flags: `--root`, `--limit` only — no output-path flag exists (the output path is the fixed `.moai/reports/graph-report.md` under the resolved root, pinned by REQ-GR-005; a user-selectable output path would let the report land on a committed location, defeating the derived-artifacts-never-committed contract); renders the three sections deterministically; empty sections emitted with stated reason (REQ-GR-006).
- `internal/template/templates/.gitignore` — add the anchored `.moai/reports/graph-report.md` rule alongside the existing plan-audit rule, with template parity (Template-First, CLAUDE.local.md §2 — the distributed template surface is what generated user projects receive, so the durable fix lives there). The repo-root `.gitignore` needs no edit: its existing `.moai/reports/*.md` rule already covers this repo.

D1 resolution (disclosed): report artifact naming is RESOLVED as the fixed rotating `.moai/reports/graph-report.md` — a REGENERATING derived artifact, never committed; the dated `.moai/reports/{TYPE}-{DATE}/` directory convention is for one-off analyses and archival naming would contradict the derived-artifacts-never-committed principle. The operator retains veto at the Implementation Kickoff Approval gate.

Tests: fixture with known fan-in ranking (ties broken by id); boundary-INFERRED edge ranked above same-confidence intra-package edge; cross-package attribution derived from endpoint file directories (`splitCodeNode`), ambiguous bare-callee edges excluded; SCC fixture (a→b→c→a) rendered as one SCC with its member list in canonical rotation; a branched-SCC fixture (edges A↔C, B↔C) renders the member list, never a fabricated cycle; empty code layer (nocgo fixture) still emits with reason; report written to the fixed path (tests assert `.moai/reports/graph-report.md` under the fixture root — no output-path flag exists to override it); `cmp` of two consecutive runs byte-identical.

### M3 — edges shrink guard (REQ-GR-008..009) [gated — enforcement sites are unshared, but verification depends on the absorbed edge inventory; re-measure applies]

Scope: guard function + enforcement at the automatic write paths.

Key files:
- `internal/graph/shrink.go` (new) — `DetectUnexplainedShrink(existing, rebuilt []Edge, scannedSources map[string]bool, projectRoot string) ShrinkReport`: names removed edges whose `Source` is a project-relative file path (kind-scoped per REQ-GR-008: the code-call kind and any other file-sourced kind) where that file — the file part of the compound `Source` payload, extracted before the existence test (code-call `file:function` via `splitCodeNode`; a literal stat of the undecoded `file:function` string is never performed) and validated as project-relative (no `..`, absolute, or symlink escape) before the stat — resolved under `projectRoot` (which the disk-existence test needs to turn the relative path into a stat-able location) — BOTH still exists on disk AND lies outside `scannedSources`; empty report = overwrite permitted. Removed edges of non-file-sourced kinds (doc-import `Source` = package/directory name, spec-depends `Source` = SPEC ID) are skipped by the guard, never evaluated against a file test — their shrinkage is explained by their own rebuild inputs (the parsed doc-dependency markdown and SPEC dependency set). A removed file-sourced source absent from disk is a genuine deletion — never reported as a shrink defect.
- `internal/cli/graph_refresh_cli.go` — `refreshEdgesArtifact` computes the guard between `BuildWithCodeLayers` and `WriteJSONL` (pre-write: refusal skips BOTH writes, so the prior artifact is byte-identical by construction); on refusal: return a typed refusal carrying the report; the query path warns and answers from the existing artifact (REQ-GR-009 fail-safe shape). The build path (`internal/cli/graph.go` build command, shared file) applies the same guard and exits non-zero naming the removed edges.
- `scannedSources` is the file list the rebuild's own extraction loop actually processed (captured inside `BuildWithCodeLayers` at build time). The extraction-seam change that yields it: `BuildWithCodeLayers` currently returns edges + fingerprints only (symbol.go:151-153, consumed at `graph_refresh_cli.go:59`), so it also returns the processed-file list alongside them — the capture rides the existing extraction loop's own file iteration, no second scan. NOT the `SourceFingerprintsForEdges` aggregate — that function returns four doc-side source-SET hashes (`internal/graph/meta.go:37`), not a per-file inventory, and does not cover the Go source tree.
- `internal/graph/meta.go` — the fixed `.tmp` temp path of the meta write (`metaPath + ".tmp"`, meta.go:90) hardened to a per-refresh unique suffix (stdlib `os.CreateTemp`-style name, or an atomic counter / crypto/rand value — NOT pid-based: two racing refreshes can be same-process, e.g. the SessionStart goroutine and a query-time refresh inside one binary, so a pid cannot separate them), closing the temp-file collision for any two refreshes; an interleaved edges/meta publication stays inside the SPEC's disclosed self-heal limitation (REQ-GR-011). Part of this milestone.

Tests: injected partial failure (rebuild missing one existing source's edges; the parse-skipped file variant — file on disk, outside the scanned set — refuses the same way) → refusal, prior artifact SHA-identical, message names the edges; genuine deletion (source file removed from disk → its edges dropped) → overwrite proceeds; existing-but-unscanned source → refuse; build path refuses with non-zero exit naming the removed edges; kind-scoped fixtures: a file-sourced (code-call) shrink refuses while a doc-import-edge shrink (package-name `Source`) and a spec-depends-edge shrink (SPEC-ID `Source`) are skipped by the guard, never misreported; the query-time refresh path and the build path covered (the deferred path exercises the same guarded wrapper once M4 lands); refusal leaves zero writes (byte comparison). Fixture discipline: the partial-failure fixture uses real extraction output (real `file:function` `Source` shapes exercising the `splitCodeNode` decode), never hand-written bare-path code-call Sources; an equal-cardinality remove+add mutant (one unscanned edge lost, one unrelated edge gained — same total count) refuses via the set-difference trigger.

### M4 — deferred SessionStart edges refresh (REQ-GR-010..012) [ungated]

Scope: extend the deferred scan pattern to the edges layer.

Key files:
- `internal/hook/session_start.go` — dependency-injection seam: a new handler option (option-pattern shape, e.g. `WithDeferredEdgesRefresh(fn func(projectDir string) error)`) stores `DeferredEdgesRefresh` on the handler; in `spawnDeferredAdvisoryScans`'s goroutine, after the advisory send and alongside the `mxScanNeeded` durable side effect: the hook-side staleness probe composes the EXPORTED predicates directly — `graph.EdgesSourcesMovedFor(root, edgesFile) || graph.MXIndexNeedsRefresh(root, graph.DefaultThresholds().MXIndexChangedFiles)` (the same composition the cli-private refresh wrapper makes, cited here as its exported constituents rather than by exporting the wrapper itself) — and invokes the injected seam for the project dir only when stale. This declares a new `internal/hook → internal/graph` import, cycle-free: `internal/graph` imports `internal/mx`, never the `internal/hook` package. Disclosure: the deferred goroutine is fire-and-forget in production — it may be killed when the hook process exits (the code's own comment, `session_start.go:1515`) — so the deferred refresh is best-effort at the process-lifetime axis, idempotent and self-healing on the next staleness check, with the query-time refresh as the liveness guarantee; a durable-worker/subprocess redesign is rejected as over-engineering (it would fork the established house pattern this SPEC adopts). The hook layer NEVER imports `internal/cli` (cli imports hook — compile-time cycle). Computed synchronously like `mxScanNeeded` (cheap fingerprint probe) so the goroutine reads a snapshot, per the existing comment contract.
- `internal/cli` (the handler construction/wiring site, shared file with M3) — wires the seam with a thin wrapper around `refreshEdgesArtifact` that applies the `edgesRefreshClock` duration seam and the budget-overrun warning; failure logs and exits the step (fail-open). M4 wraps, never forks, the rebuild path.
- Tests: `WithSynchronousDeferredScans` path asserts edges refreshed when stale, untouched when fresh (the injected seam records invocations — a nil/unset seam skips the step, preserving backward compatibility); goleak-clean join; `git status --porcelain` shows no staged entries after refresh (fixture repo); injected over-budget duration produces the warning (reuses the duration seam — no real-timing dependence).

## §G. Anti-Patterns (blocked)

- Planning M1-M3 against pre-absorb file contents; "the shared files probably didn't move" is the exact race the gate exists for.
- `internal/hook` importing `internal/cli` to reach `refreshEdgesArtifact` — compile-time cycle; the DI seam is the only crossing.
- A second, parallel rebuild path in the hook layer instead of the cli-injected wrapper around `refreshEdgesArtifact`.
- An override flag that forces a known-shrunk write past the guard.
- Wall-clock timestamps in report/tool bodies (breaks determinism ACs).
- Switching the MX validator's fan-in source "while we're in there" (t412's gate, not this SPEC's).
- Real-timing assertions on the deferred refresh (inject through the seam).

## §H. Cross-References

- SPEC: `.moai/specs/SPEC-GRAPH-REPORT-001/spec.md` (REQ-GR-001..012) · ACs: `acceptance.md` (AC-GR-001..016)
- Source analysis: `.moai/reports/graphify-codegraph-analysis-20260901.html` (P2-4..P2-7)
- Dependency: SPEC-MX-TAG-EDGES-001 (t412, branch `WT-mx-tag-edges`)
- Anchors: `codequery.go:21` (maxTraceDepth) · `graph_refresh_cli.go:53` (refreshEdgesArtifact) · `session_start.go:680` (spawnDeferredAdvisoryScans) · `mcp_server.go:486-509` (registration pattern)
