# Research — SPEC-V3R6-GRAPH-FRESHNESS-001

Two halves: (1) the operator-mandated evidence base, referenced not re-derived; (2) anchor re-verification measured in THIS tree at authoring time, because the evidence base measured at `294b4b6ab` and the base had moved.

## §1 Evidence Base (operator-mandated, read-first)

`.moai/reports/graft/graft-analysis-20260824.md` — Graft (NanoNets, MIT) adoption analysis. Path note: the report file is untracked and exists only in the primary checkout's `.moai/reports/graft/`; it is absent from worktrees (verified when the worktree read failed and the primary read succeeded). Decision: adopt 4 design principles (freshness+drift gate, content-addressed citations, AST symbol layer, MCP code queries), do NOT install the tool (TS/npm runtime vs Go single binary + 16-language neutrality), do NOT adopt the LLM concept-node layer (moai-adk already has a human semantic layer; two semantic layers without a disagreement detector). The report's structure of Graft (content-hash cache, query-time structural refresh, Crux-stores-code-text-not-line-numbers, 6 MCP tools, `graft check` exit 1) is the design source this SPEC transfers.

Foreign figures that must NOT propagate as promises: `~3ms` doc-based refresh (124-file foreign repo), −42% tokens / −46% tools / −60% time controlled sweep, SWE-bench 54%→66% — all self-reported by Graft, unverified by third parties.

## §2 Anchor Re-Verification (measured in WT-graph-freshness @ baa100ce5, 2026-08-25)

Every anchor the lead flagged was re-verified by command in this tree. Where a fact moved since the report, both values are recorded.

| Anchor | Report (at `294b4b6ab`) | This tree (`baa100ce5`) | Command evidence |
|---|---|---|---|
| codemaps last regeneration | 2026-08-12 (mtime) | commit `6da952899` 2026-08-12 (git date; mtime unusable — see below) | `git log -1 --date=short -- .moai/project/codemaps/` |
| Drift since codemaps date | 713 commits | **740 commits** on origin/main since 2026-08-12 (738 measured at review time against a different ref — the report's figure was origin/main at report time, the review's was 294b4b6ab; see spec.md §A provenance note) | `git rev-list --count --since=2026-08-12 origin/main` |
| mx-index provenance | (not examined) | **absent** — top-level keys are `schema_version`, `tags`, `scanned_at` only; no commit SHA | python3 read of primary checkout's `.moai/state/mx-index.json` |
| mx-index in a fresh worktree | (not examined) | **does not exist** (untracked runtime artifact) | `ls .moai/state/mx-index.json` → No such file |
| edges.jsonl in a fresh worktree | (not examined) | **does not exist**; `git ls-files .moai/project/graph/` → empty (untracked derived artifact) | `ls`, `git ls-files` |
| codemaps mtimes in worktree | — | all reset to checkout timestamp (Aug 25 00:08) — mtime-based freshness would misread a fresh checkout as freshly regenerated | `ls -la .moai/project/codemaps/` |
| edges aggregation site | `internal/cli/graph.go` ~233-240 | confirmed: `moai graph build` aggregates 5 doc-derived layers (import / mx-spec / spec-depends / report-milestone / milestone-card), default out `.moai/project/graph/edges.jsonl`, fails open per layer | Read of graph.go 200-280 |
| astx queries | 16 `.scm` | 16 `.scm` confirmed; **all capture declarations only** (`@symbol.function/method/type` in go.scm) — no call/import captures exist | `ls`, `cat queries/go.scm` |
| tree-sitter dep | present | `github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82` at go.mod:23 | `grep tree-sitter go.mod` |
| `moai mx query` | — | exists (`internal/cli/mx_query.go`) — M2 has a live query surface to refresh | grep of cli |
| gate machinery | — | `quality.NewQualityGate` at `internal/hook/quality/gate.go:305`; CLI gate shares config loader with PreToolUse hook (SSOT seam, ast-grep precedent for warn-only default) | grep + Read of `internal/cli/gate.go` |
| MCP registration | 21 tools | `add()` pattern + named tool constants in `internal/cli/mcp_server.go`; catalog fn `mcpcat.MoaiMCPToolNames()` | grep |
| CI landing surface | — | 17 tracked workflow files; `spec-lint.yml` is the small-focused-workflow precedent for a standalone `graph-freshness.yml` | `ls .github/workflows/` (17 files counted) |
| astx importers | navigator-only | confirmed: importers are navigator internals + `internal/cli/navigator_enrich.go` only | grep `navigator/astx` |
| astx provenance precedent | — | `Provenance` struct + `CurrentProvenance(projectRoot)` already exist in `internal/navigator/astx/enrich.go` — §2 of design.md reuses the shape convention | grep of exported surface |

Two derived findings that changed the SPEC beyond the report:

1. **Absence semantics are a first-class state.** A fresh worktree has no mx-index and no edges.jsonl at all. Any freshness design that only grades "old vs new" misclassifies absence; REQ-GF-002/004 make `absent` an explicit, failing verdict, and AC-GF-005 tests it.
2. **The tracking-status split forces the per-layer metric split.** Tracked (codemaps) can use endpoint git diffs; untracked (mx-index, edges) have no history to consult and mtime lies — content fingerprinting is the only sound signal. This is why REQ-GF-002 is a three-row table with per-row rationale rather than one formula.

## §3 Deduplication Check

No existing SPEC covers graph freshness, staleness gating, code-derived edges, or MCP code queries: `ls .moai/specs/ | grep -i -E "graph|fresh|graft|edge|codemap"` → `SPEC-DWF-CODEMAPS-PILOT-001`, `SPEC-V3R6-DOCS-CODEMAPS-V3-001` only; both are codemaps content/SSOT work, neither touches freshness or the symbol layer. `SPEC-V3R6-GRAPH-FRESHNESS-001` is unique (regex-validated `PASS` before write).

## §4 Open Items Settled by Design (no user clarification pending)

- Gate step default warn-only (ast-grep precedent) vs CI blocking — settled in design.md §7 with rationale.
- M2 refresh scope = mechanical layers only — settled in spec.md §B.2 (auto-rewriting curated docs without an LLM would produce garbage).
- edges provenance carrier = sidecar `.meta.json` (keeps JSONL line-per-edge purity) — design.md §2.
- Citation hash = region-content hash, not whole-file — design.md §4.

No `[NEEDS CLARIFICATION]` markers remain: thresholds and budgets carry documented defaults with a calibration obligation inside M1/M2 (acceptance.md §D.7), which is a run-phase measurement, not an open design question.
