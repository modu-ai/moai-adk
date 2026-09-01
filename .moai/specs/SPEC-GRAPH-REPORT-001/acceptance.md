---
id: SPEC-GRAPH-REPORT-001
title: "Acceptance criteria — graph report toolchain"
version: "0.1.0"
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
- **AC-GR-002** — Given a fixture chain of 10 code-call edges a1→a2→…→a10, When `graph_shortest_path` is queried from a1 to a10, Then the returned path is capped at 8 hops with the cap named in the response. Verify: targeted test asserting hop count ≤ 8.
- **AC-GR-003** — Given a fixture with no edge path between two nodes, When the query runs, Then the response is a structured not-found result naming both endpoints, the cap, and provenance — tool error count 0. Verify: test asserts the not-found shape, no `IsError` result.
- **AC-GR-004** — Given the same fixture tree, When the shortest-path query executes twice, Then the two responses are byte-identical. Verify: `cmp` of the two serialized responses, exit 0.
- **AC-GR-005** — Given a request carrying an invalid `project_root`, When the tool runs, Then the root is rejected with an explicit error (no fallback to another tree). Verify: test asserts the rejection error, mirroring the existing tools' `resolveToolProjectRoot` contract.

### M2 — moai graph report

- **AC-GR-006** — Given a fixture edges.jsonl, When `moai graph report --root <fixture>` runs, Then a report file exists under the fixture's `.moai/reports/` containing the three section headings (god nodes, surprising connections, import cycles) and exit code 0. Verify: command + file-content assertions in a targeted test.
- **AC-GR-007** — Given a fixture where package X has 3 distinct importers and package Y has 3 distinct importers, When god nodes are rendered, Then both rank at the same tier and ties resolve by node id ascending. Verify: test asserts the deterministic ordering on the fixture.
- **AC-GR-008** — Given an INFERRED code-call edge crossing a package boundary and an INFERRED intra-package edge of equal confidence, When surprising connections are rendered, Then the boundary-crossing edge ranks first. Verify: fixture test asserting the ordering.
- **AC-GR-009** — Given import edges forming cycle a→b→c→a, When the report renders, Then the cycle is listed with members in the canonical rotation (smallest node id first). Verify: fixture test.
- **AC-GR-010** — Given a tree whose code layer is absent (nocgo fixture, no code edges), When `moai graph report` runs, Then the report is still emitted, the code-dependent sections are present-but-empty with the stated reason, and exit code is 0. Verify: command + assertions.
- **AC-GR-011** — Given the same fixture tree, When `moai graph report` runs twice, Then the two report files are byte-identical. Verify: `cmp`, exit 0.

### M3 — shrink guard

- **AC-GR-012** — Given an existing edges.jsonl and a rebuild whose scanned source set excludes one source file that contributed edges to the existing artifact, When the refresh path reaches the write step, Then the overwrite is refused: the existing edges.jsonl and meta sidecar are byte-identical before/after (SHA-compared), and the refusal names the removed edges and the unscanned source. Verify: test with injected partial rebuild; `sha256` equality asserted.
- **AC-GR-013** — Given a legitimate rebuild where a source file was genuinely deleted and its edges dropped, When the rebuild runs, Then the overwrite proceeds normally (the removed edges' source is inside the rebuild's scanned set). Verify: test asserting the write succeeds and the artifact shrinks by exactly the departed source's edges.
- **AC-GR-014** — Given the query-time refresh refused by AC-GR-012's condition, When the user runs a graph query, Then the answer is produced from the EXISTING artifact with a stated shrink-refusal warning on stderr, and exit code is 0. Verify: CLI-level test asserting answer + warning.

### M4 — deferred refresh

- **AC-GR-015** — Given a session start over a tree whose edges layer is stale, When the SessionStart handler runs in synchronous-deferred-scans mode, Then after the advisory payload is delivered the edges artifact is refreshed (staleness predicate now false), the duration was measured through the seam, and no entry is staged in git (`git status --porcelain` shows no staged lines). Given a fresh edges layer, Then no refresh write occurs (artifact mtime/SHA unchanged). Verify: hook-package test using `WithSynchronousDeferredScans`; goleak-clean; separate assertion for the fresh case; budget-overrun warning observed via an injected over-budget duration through the `edgesRefreshClock` seam.

## §D.1 Severity

- MUST (release-blocking): AC-GR-001..015.
- SHOULD: none reserved.
- MAY: none.

## §D.2 Traceability

| AC | REQ | Milestone |
|---|---|---|
| AC-GR-001 | REQ-GR-001 | M1 |
| AC-GR-002 | REQ-GR-002 | M1 |
| AC-GR-003 | REQ-GR-003 | M1 |
| AC-GR-004 | REQ-GR-004 | M1 |
| AC-GR-005 | REQ-GR-001 | M1 |
| AC-GR-006 | REQ-GR-005 | M2 |
| AC-GR-007 | REQ-GR-005 | M2 |
| AC-GR-008 | REQ-GR-005 | M2 |
| AC-GR-009 | REQ-GR-005 | M2 |
| AC-GR-010 | REQ-GR-006 | M2 |
| AC-GR-011 | REQ-GR-004 | M2 |
| AC-GR-012 | REQ-GR-008 | M3 |
| AC-GR-013 | REQ-GR-008 | M3 |
| AC-GR-014 | REQ-GR-009 | M3 |
| AC-GR-015 | REQ-GR-010/011/012 | M4 |

## §D.3 Edge cases covered

- >8-hop chains and disconnected components (AC-GR-002/003); nocgo absence (AC-GR-010); genuine-source-deletion shrink vs partial-failure shrink (AC-GR-012/013); fresh-layer no-op (AC-GR-015 second Given); tie-breaking determinism (AC-GR-007).

## §D.4 Closure gates

1. All 15 ACs PASS with per-AC attribution triples (plan.md §E).
2. Determinism double-run evidence (`cmp` exit 0) attached for AC-GR-004 and AC-GR-011.
3. Absorbed-tree re-measure recorded for M1-M3 (plan.md §F.0) before their first commit.
4. Mutant check on AC-GR-012: a rebuild that silently shrinks MUST fail this AC — verified by observing the AC fail against a guard-less build (RED evidence).
