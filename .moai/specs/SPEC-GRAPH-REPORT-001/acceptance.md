---
id: SPEC-GRAPH-REPORT-001
title: "Acceptance criteria — graph report toolchain"
version: "0.2.0"
created: 2026-09-02
author: manager-spec
---

## §A. Verification Conventions

- Every AC is binary-testable: command → expected observable outcome. RED-first per TDD; the verbatim RED output is §E.2 evidence (plan.md §E, item E8).
- Fixtures live under `t.TempDir()`; graph fixtures are hand-written `edges.jsonl` files (no extraction dependency unless the AC says CGO-gated).
- Determinism ACs compare two consecutive executions with `cmp` (exit 0), same tree, same fixture.

## §D. AC Matrix

### M1 — graph_shortest_path

- **AC-GR-001** — Given the MCP server registers its tools, When the registry test enumerates the code-graph tools, Then `graph_shortest_path` is present alongside `graph_file_api` / `graph_find_code` / `graph_trace_calls`, annotated read-only, with `from`/`to` required parameters and `project_root` optional. Verify: targeted `go test ./internal/cli/... -run TestMCPServer` exits 0 with the tool named in output.
- **AC-GR-002** — Given a fixture chain of exactly 8 code-call edges a1→a2→…→a9 (8 hops), When `graph_shortest_path` is queried from a1 to a9, Then the path is returned with its hop count asserted ≤ 8. And given a fixture chain of 9 hops (10 edges), When queried end to end, Then the response is the structured not-found naming both endpoints and the cap (a truncated non-reaching path is never returned). Verify: targeted tests asserting hop count ≤ 8 on the first fixture and the not-found shape on the second.
- **AC-GR-003** — Given a fixture with no edge path between two nodes, When the query runs, Then the response is a structured not-found result naming both endpoints, the cap, and provenance — tool error count 0. And given a bare symbol name resolving to two nodes in different files, When the query runs with that bare name as an endpoint, Then the response is a structured candidates list (name → matching node ids) with no path. Verify: tests assert the not-found shape, no `IsError` result; and the candidates shape on the ambiguity fixture.
- **AC-GR-004** — Given the same fixture tree, When the shortest-path query executes twice, Then the two responses are byte-identical. Verify: `cmp` of the two serialized responses, exit 0.
- **AC-GR-005** — Given a request carrying an invalid `project_root`, When the tool runs, Then the root is rejected with an explicit error (no fallback to another tree). Verify: test asserts the rejection error, mirroring the existing tools' `resolveToolProjectRoot` contract.

### M2 — moai graph report

- **AC-GR-006** — Given a fixture edges.jsonl, When `moai graph report --root <fixture>` runs, Then the report file exists at the fixture's fixed rotating path `.moai/reports/graph-report.md` containing the three section headings (god nodes, surprising connections, import cycles) and exit code 0. Verify: command + file-content assertions in a targeted test.
- **AC-GR-007** — Given a fixture where package X has 3 distinct importers and package Y has 3 distinct importers, When god nodes are rendered, Then both rank at the same tier and ties resolve by node id ascending. Verify: test asserts the deterministic ordering on the fixture.
- **AC-GR-008** — Given an INFERRED code-call edge whose endpoints' source-file directories differ (cross-package via the directory proxy) and an INFERRED intra-package edge of equal confidence, When surprising connections are rendered, Then the boundary-crossing edge ranks first, and an ambiguous bare-callee edge (callee name matching 2+ nodes) is excluded from the section. Verify: fixture test asserting the ordering and the exclusion.
- **AC-GR-009** — Given import edges forming cycle a→b→c→a, When the report renders, Then one SCC is reported with its representative cycle's members in the canonical rotation (smallest node id first), and the SCC count (not a simple-cycle enumeration) is what the section states. Verify: fixture test.
- **AC-GR-010** — Given a tree whose code layer is absent (nocgo fixture, no code edges), When `moai graph report` runs, Then the report is still emitted, the code-dependent sections are present-but-empty with the stated reason, and exit code is 0. Verify: command + assertions.
- **AC-GR-011** — Given the same fixture tree, When `moai graph report` runs twice, Then the two report files are byte-identical. Verify: `cmp`, exit 0.

### M3 — shrink guard

- **AC-GR-012** — Given an existing edges.jsonl and a rebuild whose scanned source set excludes one source file that still exists on disk and contributed edges to the existing artifact, When the refresh path reaches the write step, Then the overwrite is refused BEFORE any write: the existing edges.jsonl and meta sidecar are byte-identical before/after (SHA-compared), and the refusal names the removed edges and the unscanned source. And given the same shrink condition, When `moai graph build` runs explicitly, Then it exits non-zero naming the removed edges with the prior artifact SHA-identical. Verify: tests with injected partial rebuild; `sha256` equality asserted on both the refresh and build paths.
- **AC-GR-013** — Given a legitimate rebuild where a source file was genuinely deleted from disk and its edges dropped, When the rebuild runs, Then the overwrite proceeds normally — the deleted source is absent from disk, therefore outside the scanned set by the existence discriminator, and is NOT reported as a shrink defect. Verify: test asserting the write succeeds and the artifact shrinks by exactly the departed source's edges.
- **AC-GR-014** — Given the query-time refresh refused by AC-GR-012's condition, When the user runs a graph query, Then the answer is produced from the EXISTING artifact with a stated shrink-refusal warning on stderr, and exit code is 0. Verify: CLI-level test asserting answer + warning.

### M4 — deferred refresh

- **AC-GR-015** — Given a session start over a tree whose edges layer is stale, When the SessionStart handler runs in synchronous-deferred-scans mode with the `DeferredEdgesRefresh` DI seam wired to a wrapper around `refreshEdgesArtifact`, Then after the advisory payload is delivered the edges artifact is refreshed (staleness predicate now false), the duration was measured through the seam, and no entry is staged in git (`git status --porcelain` shows no staged lines). Given a fresh edges layer, Then no refresh write occurs (artifact mtime/SHA unchanged). Verify: hook-package test using `WithSynchronousDeferredScans` and the injected seam (a nil seam skips the step — backward-compat assertion); goleak-clean; separate assertion for the fresh case; budget-overrun warning observed via an injected over-budget duration through the `edgesRefreshClock` seam.
- **AC-GR-016** — Given a report fixture with fan-in over mixed edge kinds, When the report's god-nodes section renders, Then it names the edge kinds it counted (e.g. "fan-in over: code-call, import"). And across the whole change set, `internal/hook/mx/validator.go`'s `fanInIndex` computation is byte-unchanged. Verify: fixture test asserting the kinds line; diff-level assertion that validator.go's fan-in computation is untouched by this SPEC's commits.

## §D.1 Severity

- MUST (release-blocking): AC-GR-001..016.
- SHOULD: none reserved.
- MAY: none.

## §D.2 Traceability

| AC | REQ | Milestone |
|---|---|---|
| AC-GR-001 | REQ-GR-001 | M1 |
| AC-GR-002 | REQ-GR-002/003 | M1 |
| AC-GR-003 | REQ-GR-003/001 | M1 |
| AC-GR-004 | REQ-GR-004 | M1 |
| AC-GR-005 | REQ-GR-001 | M1 |
| AC-GR-006 | REQ-GR-005 | M2 |
| AC-GR-007 | REQ-GR-005 | M2 |
| AC-GR-008 | REQ-GR-005 | M2 |
| AC-GR-009 | REQ-GR-005 | M2 |
| AC-GR-010 | REQ-GR-006 | M2 |
| AC-GR-011 | REQ-GR-005 | M2 |
| AC-GR-012 | REQ-GR-008 | M3 |
| AC-GR-013 | REQ-GR-008 | M3 |
| AC-GR-014 | REQ-GR-009 | M3 |
| AC-GR-015 | REQ-GR-010/011/012 | M4 |
| AC-GR-016 | REQ-GR-007 | M2 |

## §D.3 Edge cases covered

- Exactly-8-hop chain (path found) vs >8-hop chain (structured not-found) and disconnected components (AC-GR-002/003); ambiguous bare-name endpoints (candidates list, no path — AC-GR-003); nocgo absence (AC-GR-010); genuine-source-deletion shrink vs existing-but-unscanned-source refusal (AC-GR-012/013) plus the explicit build-path refusal (AC-GR-012 second clause); ambiguous bare-callee edges excluded from surprising connections (AC-GR-008); fresh-layer no-op (AC-GR-015 second Given); tie-breaking determinism (AC-GR-007).

## §D.4 Closure gates

1. All 16 ACs PASS with per-AC attribution triples (plan.md §E).
2. Determinism double-run evidence (`cmp` exit 0) attached for AC-GR-004 and AC-GR-011.
3. Absorbed-tree re-measure recorded for M1-M3 (plan.md §F.0) before their first commit.
4. Mutant check on AC-GR-012: a rebuild that silently shrinks MUST fail this AC — verified by observing the AC fail against a guard-less build (RED evidence).
