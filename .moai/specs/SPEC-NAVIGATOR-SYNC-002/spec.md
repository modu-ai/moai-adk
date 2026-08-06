---
id: SPEC-NAVIGATOR-SYNC-002
title: "Navigator Sync (BAS M1) — Falconer Detect: PostToolUse changed-path → affected-graph-rows mapping"
version: "0.1.0"
status: implemented
created: 2026-08-06
updated: 2026-08-06
author: manager-spec
priority: P1
phase: "v3.3 target"
module: navigator-sync
lifecycle: spec-anchored
tier: M
era: V3R6
tags: "navigator, sync, falconer, detect, posttooluse, hook, bas-epic"
related_specs: [SPEC-NAVIGATOR-SYNC-001, SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002, SPEC-PROJECT-NAVIGATOR-003]
---

# SPEC-NAVIGATOR-SYNC-002 — Navigator Sync (BAS Epic M1) — Falconer Detect

## HISTORY

- 2026-08-06 (initial draft) — BAS (Blueprint-Anchored Synchronization) Epic M1, plan-phase. M0 (SPEC-NAVIGATOR-SYNC-001, PR #1375 squash `f4bd58acbd`, MERGED 2026-08-05, `status: completed`) delivered the SSOT binding-token trio (`@NAV:DEC-<id>` / `@NAV:SYM:<symbol>` / `@MX:SPEC:<id>` bridged not absorbed per REQ-NS-005) and the graph-join schema layer at `internal/navigator/sync/` (9 files) producing `.moai/project/navigator/nav-graph.json` (nodes: decision/spec/symbol; edges: dec-edge/spec-edge/sym-edge, each carrying `source_path` + `line_number`). M1 is the **Detect** element of the Falconer 3-element loop (detect → route → fix) from the design report `.moai/reports/navigator-redesign-bas-20260805.{md,html}` §3.3(C), §4, §6. M1 builds ON TOP of `nav-graph.json`: it consumes the graph, it does not extend the token grammar or the join layer. M2 (Route — audit advisory → work item) and M3 (Fix — AI-drafted incremental regen) are separate SPECs, out of scope here. Design-report asset-reuse directive (`navigator-audit.sh` token-normalization matching → Detect's path→doc mapping engine) is honored as REQ-NS2-010. Bridge-not-absorb principle (M0 REQ-NS-005, code-verified at `internal/mx/spec_association.go:21-32` — the `type SpecAssociator` struct at L21 + `NewSpecAssociator` constructor at L32 — and `internal/navigator/sync/mx_bridge.go:28-64` — the `BridgeMxAssociations` consumer) carries forward as REQ-NS2-005: M1 consumes M0 outputs and the existing navigator-audit matching logic, never mutates producers.

## §A. User Story

**As a** developer or agent editing a source file, design doc, or SPEC,
**I want** the Navigator graph to surface — in real time, at the moment of my edit — which decision records, SPEC anchors, and symbol bindings my changed file touches,
**so that** I am warned before the graph drifts and before downstream consumers (audit, capability-map, blueprint readers) act on a stale picture.

Today the only change signal is a SessionStart staleness hint (a count, not a trigger); there is no per-edit impact surfacing. M1 closes that gap with a PostToolUse Detect hook that maps a changed file path to the affected navigator-graph rows via reverse edge traversal.

**Falconer framing**: "갱신 인프라 없는 문서는 살아있는 게 아니라 출처가 더 좋은 스냅샷이다." M1 is the **detect** element of the 3-element living-document loop (detect → route → fix). M1 alone does not fix drift; it makes drift visible at the cheapest possible moment (the edit itself).

## §B. Scope

### In scope

- A PostToolUse detect layer that reads `.moai/project/navigator/nav-graph.json` and maps a changed file path to the affected graph rows (nodes + edges) via reverse traversal.
- Trigger surface scoped to Write/Edit/NotebookEdit (structured `file_path` in tool input).
- Read-only advisory output: a `systemMessage` naming affected rows + a machine-readable impact record under `.moai/state/navigator-detect/` for M2 Route to consume.
- Fail-open semantics: detection never blocks a tool call.
- Concurrency-safe atomic read of nav-graph.json during regeneration by any Navigator chain.
- Reuse of `navigator-audit.sh`'s token-normalization matching as the path→doc mapping foundation.
- Template-first distribution: any new distributed config keys ship in `internal/template/templates/` first.
- ≥80% Detect coverage success metric, mechanically measurable.

### Excluded from M1 (full list in §F)

M2 Route (advisory → work item promotion), M3 Fix (AI-drafted incremental regen), M4 4-tier map, M5 brownfield reverse-extraction. Bash file-mutation triggering. Modifying `internal/navigator/sync/` or `internal/mx/`. Modifying the three predecessor Navigator chains. The canonical exclusions live in §F with `### Out of Scope — <topic>` sub-headings.

## §C. Functional Requirements (GEARS)

#### REQ-NS2-001 (Event-driven — trigger surface)

**When** a PostToolUse event for `Write`, `Edit`, or `NotebookEdit` arrives at the post-tool handler, the Detect layer SHALL map the changed file path (extracted from the tool input's `file_path` field) to the affected navigator-graph rows. **When** the tool is `Bash` or any tool without a structured `file_path`, the Detect layer SHALL NOT attempt path→row mapping (Bash file mutations have no structured path and are out of scope for M1).

#### REQ-NS2-002 (Ubiquitous — reverse edge traversal)

The Detect layer SHALL traverse `nav-graph.json` in reverse from the changed file path: it SHALL collect every edge whose `source_path` matches the changed path (both normalized to absolute, project-root-anchored form) and, for each matching edge, the target node referenced by `source_node` / `target_node`. The output SHALL be the affected-row set: a deduplicated list of affected nodes (with `entity_type` ∈ {decision, spec, symbol} and `identifier`) and the originating edges (with `edge_type`, `source_path`, `line_number`).

#### REQ-NS2-003 (Ubiquitous — read-only advisory output)

The Detect layer SHALL emit its result as (a) a read-only advisory `systemMessage` naming the affected graph rows in human-readable form, AND (b) an append-only machine-readable impact record at `.moai/state/navigator-detect/<session-id>.jsonl` with one JSON line per detection `{ "changed_path": "...", "affected_nodes": [...], "affected_edges": [...] }`. The Detect layer SHALL NOT promote findings to an actionable work item, SHALL NOT block the tool call, and SHALL NOT mutate any source file, any SPEC, or any M0 producer artifact — Route (M2) owns work-item promotion.

#### REQ-NS2-004 (Event-detected — fail-open on every error mode)

**When** the Detect layer encounters any failure — `nav-graph.json` absent (not yet generated by M0), stale, unparseable as JSON, graph-schema invalid, traversal error, or per-detection timeout exceeded — the Detect layer SHALL degrade silently to "no impact surfaced": return an empty affected-row set, emit exit-0 `Decision: "allow"`, surface NO user-facing error, and append a single diagnostic line to `.moai/logs/navigator-sync.log`. The Detect layer SHALL NEVER abort the PostToolUse handler or cascade a failure into sibling PostToolUse branches (LSP, AST, MX validation).

#### REQ-NS2-005 (Ubiquitous — consumer-only on M0 + mx layer; bridge-not-absorb carries forward)

The Detect layer SHALL treat `internal/navigator/sync/` outputs (nav-graph.json) and `internal/mx/` outputs as READ-ONLY inputs. The Detect layer SHALL NOT modify `internal/navigator/sync/`, `internal/mx/`, or any M0 producer surface. This carries forward the M0 bridge-not-absorb principle (REQ-NS-005, REQ-NS-012) to the M1 read surface: M1 consumes the graph + the existing navigator-audit matching logic, never mutates producers.

#### REQ-NS2-006 (State-driven — concurrency-safe atomic read)

**While** a Navigator chain (001 regen / 003 enrich / M0 sync join) is regenerating `nav-graph.json` via the atomic-rename pattern (`internal/navigator/sync/write.go` atomicWrite — `.tmp` then `os.Rename`), the Detect layer SHALL read the graph in a single `os.ReadFile` call and tolerate a stale snapshot without error. A read that lands during regeneration SHALL observe the prior committed graph atomically (never a partial file), and — if that prior graph yields no affected rows for the changed path — SHALL fail-open per REQ-NS2-004 rather than block or retry.

#### REQ-NS2-007 (Capability gate — ≥80% Detect coverage, mechanically measured)

**Where** the ≥80% Detect coverage success metric is evaluated (per the design report §성공 지표), the coverage percentage SHALL be produced by the mechanical measurement procedure named in acceptance.md §D.AC-NS2-007 (a fixture corpus of changed files + the reverse-traversal command emitting the ratio of in-scope changed files that yield ≥1 affected row). The coverage percentage SHALL NOT be a narrative claim, a sampled estimate, or a carried-over number from a prior measurement — every coverage assertion MUST be attributable to a run command + observed output (per `verification-claim-integrity.md` §1.1 surface 3 + §2 attribution).

#### REQ-NS2-008 (Ubiquitous — non-overlap with the 3 predecessor Navigator chains)

The Detect layer SHALL NOT write to `.moai/project/navigator/capability-map.md`, `audit-report.{md,json}`, `capability-symbols.{md,json}`, `nav-graph.json`, or any `.moai/specs/` SPEC surface. The Detect layer SHALL NOT modify the three predecessor Navigator chains' regen / enrich / audit surfaces — carries forward REQ-PN-016 (001), REQ-NA-011 (002), REQ-NT-018/019 (003), and the M0 non-overlap guards REQ-NS-013 / REQ-NS-015 / REQ-NS-016.

#### REQ-NS2-009 (Capability gate — integration as a new branch, NOT a forked hook chain)

**Where** the Detect layer plugs into the PostToolUse event, it SHALL register as a new conditional branch inside the existing `postToolHandler.Handle` dispatcher (`internal/hook/post_tool.go:145`) — mirroring the pattern of `runAstScan`, `runMxValidation`, `runMemoryAudit`, and `logEvidence`. The Detect layer SHALL NOT fork the PostToolUse hook chain, SHALL NOT duplicate the dispatch, SHALL NOT create a parallel `moai hook navigator-detect` subcommand that bypasses the existing `handle-post-tool.sh` wrapper, and SHALL NOT introduce a new wrapper script.

#### REQ-NS2-010 (Ubiquitous — asset reuse: navigator-audit.sh as INSPIRATION, not the mapping engine)

The Detect layer's primary path→graph-edge mapping is absolute-path string equality (per plan.md §C.2: both changed path and `Edge.SourcePath` are absolute, normalized via `filepath.Abs`). The `navigator-audit.sh` `heuristic_match()` last-segment resolution (`.claude/skills/moai-workflow-project/scripts/navigator-audit.sh:406-422`) is cited as **inspiration for the directory-prefix fallback case** ONLY (when a changed path is a directory prefix, not a file path), NOT as the path→graph-edge mapping engine. The Detect layer SHALL cite this inspiration by file path + function name in plan.md §C (asset-reuse map) AND SHALL NOT claim the matching engine itself is reused — the engine is absolute-path string equality, not the shell script's commit-title/path → design-doc name matching.

#### REQ-NS2-011 (Capability gate — template-first distribution)

**Where** the Detect layer ships any new distributed surface — a new `settings.json` PostToolUse config key enabling detection, a new hook block, or any artifact under `.claude/` — the template source at `internal/template/templates/.claude/` SHALL carry the change first, `make build` SHALL regenerate the embedded catalog (`catalog.yaml`), and the local copy SHALL mirror — per CLAUDE.local.md §2 [HARD] Template-First Rule. The run-phase touch list (template paths + mirror paths) is named in plan.md §D so compliance is planned, not discovered late.

#### REQ-NS2-012 (Event-detected — PostToolUse never blocks)

**When** the Detect layer completes — success (affected rows surfaced) or fail-open (no rows surfaced) — the post-tool handler SHALL return `Decision: "allow"` and the process SHALL exit 0. Detection is advisory-only; a detection finding SHALL NEVER cause `Decision: "block"`, SHALL NEVER emit exit 2, and SHALL NEVER prevent the triggering Write/Edit/NotebookEdit from completing.

## §D. Constraints (Non-Functional)

- **Performance budget**: the Detect layer runs inside the MoAI 5-second PostToolUse policy timeout (CLAUDE.local.md §7). The reverse traversal of nav-graph.json SHALL complete in < 200ms p99 for a graph of up to 10,000 edges (the M0 graph for this repo is O(hundreds) of edges; the budget headroom is for fan-out). The Detect branch SHALL honor the existing PostToolUse context-cancellation discipline (no hardcoded sleeps; `context.WithTimeout`).
- **Read-only invariant**: the Detect layer's only write surfaces are `.moai/state/navigator-detect/<session-id>.jsonl` (append-only impact record) and `.moai/logs/navigator-sync.log` (advisory diagnostic). No other file mutation.
- **Idempotence / determinism**: for the same `nav-graph.json` baseline + the same changed path, the Detect layer SHALL produce the same affected-row set (sorted, deduplicated). The impact-record JSON line is order-stable.
- **Provenance alignment** (`verification-claim-integrity.md`): every coverage / mapping-rate claim in an AC is attributable to a mechanical command + observed output. The impact record is the Evidence; the affected-row claim is the Claim.
- **Concurrency**: atomic-read of nav-graph.json (REQ-NS2-006) is the load-bearing guarantee; no mutex / file-lock is added on top of M0's atomic-rename.
- **Language neutrality**: the Detect layer's path-matching is language-neutral (operates on graph edges, not source ASTs); it inherits the language coverage of the M0 graph without re-detecting language markers.

## §E. Verification Surface

- **Coverage AC**: acceptance.md §D.AC-NS2-007 names the exact fixture-corpus directory, the reverse-traversal CLI smoke command, and the ratio formula. The percentage is the observed output of that command, not a narrative.
- **Fail-open ACs**: one AC per error mode (graph absent / unparseable / schema-invalid / traversal error / timeout) — each Given-When-Then, each asserting exit 0 + empty affected-row set + a log line.
- **Non-overlap AC**: a grep-based test asserting the Detect source files do not name forbidden write surfaces (carries forward M0's `nonoverlap_test.go` pattern).
- **Concurrency AC**: a test using M0's `NAVIGATOR_PRE_RENAME_BARRIER` test hook to hold the graph write mid-rename while the Detect reader runs, asserting the reader observes the prior graph without error.

## §F. Out of Scope

### Out of Scope — M2 Route (advisory → work item promotion)

- Promoting a Detect advisory into a tracked work item (GitHub issue, Linear ticket, `.moai/specs/` draft, TODO file). Owned by M2 Route (separate SPEC).
- Surfacing an owner (code path / agent / team) for each affected row. M2 owns owner-assignment.

### Out of Scope — M3 Fix (AI-drafted incremental regen)

- Auto-drafting an incremental regeneration of `capability-map.md` / `audit-report.{md,json}` / `nav-graph.json` to absorb the change. Owned by M3 Fix.
- Invoking any LLM to draft a fix, diff, or patch for the affected rows.

### Out of Scope — Bash file-mutation triggering

- Detecting changes from `Bash` tool calls (`git checkout`, `sed -i`, `mv`, `rm`). Bash has no structured `file_path` in tool input; path-extraction heuristics are unreliable and out of scope for M1. Write/Edit/NotebookEdit only (REQ-NS2-001).

### Out of Scope — M0 / mx producer modification

- Modifying `internal/navigator/sync/` (M0 join layer), `internal/mx/` (scanner + SpecAssociator), the token grammar, or the graph schema. REQ-NS2-005 / REQ-NS2-008 forbid this.

### Out of Scope — the three predecessor Navigator chains

- Modifying `internal/navigator/astx/` (001/003), `navigator-regen.sh` (001), `navigator-enrich.sh` (003), or `navigator-audit.sh` (002). The Detect layer reuses their outputs / matching logic read-only.

### Out of Scope — M4 4-tier map + M5 brownfield reverse-extraction

- Tier 0 Contract / Tier 1 Blueprint / Tier 2 ADR / Tier 3 SCIP symbol layers (M4). Reverse-extracting docs from existing code via the tessl `document --code` pattern (M5).

## §G. Related SPECs (non-blocking)

- **SPEC-NAVIGATOR-SYNC-001** (M0, `status: completed`, PR #1375 `f4bd58acbd`) — produces `nav-graph.json`, the input M1 consumes. Referenced via `related_specs` (non-blocking), not `depends_on`: M0 is merged and drift-stable.
- **SPEC-PROJECT-NAVIGATOR-001** (regen, completed) — produces `capability-map.md`, a graph input.
- **SPEC-PROJECT-NAVIGATOR-002** (audit, completed) — produces `audit-report.json`, a graph input; its `navigator-audit.sh` matching logic is reused (REQ-NS2-010).
- **SPEC-PROJECT-NAVIGATOR-003** (enrich, completed) — produces `capability-symbols.json`, a graph input.

## §H. Cross-References

- `.moai/reports/navigator-redesign-bas-20260805.{md,html}` — design report. §3.3(C) Falconer 3-element loop, §4 asset reuse (navigator-audit.sh), §6 M1 milestone card, §성공 지표 (≥80% Detect coverage).
- `.claude/rules/moai/workflow/nav-tokens.md` — the binding-token trio author surface (M0).
- `internal/navigator/sync/{schema,scan,join,write,mx_bridge}.go` — M0 graph-join layer (consumed).
- `internal/hook/post_tool.go:145` (`postToolHandler.Handle`) — the dispatcher M1 extends (REQ-NS2-009).
- `.claude/skills/moai-workflow-project/scripts/navigator-audit.sh` — token-normalization matching (reused, REQ-NS2-010).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3 + §2 — coverage-claim attribution.
- CLAUDE.local.md §2 [HARD] Template-First Rule, §7 Hook Development Guidelines.
