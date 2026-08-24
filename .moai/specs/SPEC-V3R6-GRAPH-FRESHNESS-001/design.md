# Design — SPEC-V3R6-GRAPH-FRESHNESS-001

Design decisions below the WHAT-level of spec.md, above the HOW-level of run-phase code. Each section names the decision, the alternative rejected, and why.

## §1 Layer Model (drives M1 metrics, M2 refresh scope)

| Property | codemaps | mx-index | edges.jsonl |
|---|---|---|---|
| Path | `.moai/project/codemaps/*.md` | `.moai/state/mx-index.json` | `.moai/project/graph/edges.jsonl` |
| Git-tracked | yes | no (runtime state) | no (derived) |
| Producer | curated + generation pipeline | `moai mx scan` (mechanical) | `moai graph build` (deterministic aggregation) |
| Sources | described source trees (`internal/`, `cmd/`, `pkg/`) | Go/16-lang sources carrying @MX | codemaps + @MX + SPEC dir + reports dir |
| Refresh owner | human/agent at sync phase (M1 gate only) | M2 query-time refresh | M2 query-time refresh (rebuild) |
| Staleness metric | endpoint content-diff vs stamped gen commit | content-diff of scan inventory | source-fingerprint mismatch |
| Absent verdict | n/a (tracked — present in every checkout) | `absent` (fresh worktree) | `absent` (fresh worktree / never built) |

Why the metrics differ (the load-bearing rationale, from measured facts in research.md):

- **Tracked layer**: git gives exact endpoint comparison. `git diff --name-only <gen-SHA> -- <described trees>` counts files whose endpoint content differs — a revert-to-identical burst counts zero, which commit-count would misreport. The generation commit does not exist in codemaps today → REQ-GF-003 stamps it.
- **Untracked layers**: no git history exists, so history-based metrics are impossible, and mtime lies (worktree checkout resets it — measured). The only sound signals are content-derived: the scanner's inventory (what did the index read, and does that content still match) for mx-index; source fingerprints for the derived artifact.
- **Derived layer**: a derived artifact is stale iff its sources moved — measured by recomputing the four source-set fingerprints (codemaps dir, mx-index file, SPEC dir, reports dir) and comparing to those stamped at build time. Cheap (hashing a few paths), exact, no rebuild needed to judge.

**Alternative rejected**: one unified metric (age-since-mtime) — always-green in fresh worktrees (all mtimes new) and always-red in long-lived primary checkouts; both uninformative (card directive 1's exact failure mode).

**Alternative rejected**: drift-commit-count for codemaps — measured against this repo: 740 commits since the codemaps regeneration, but the file-level endpoint diff is the number that maps to "how wrong are the docs", and reverts/churn inflate commit-count without changing endpoints.

## §2 Provenance Block (M1)

Single schema stamped into all three artifacts (mx-index gains a top-level field; edges.jsonl gains a header line or sidecar `.meta.json` — sidecar preferred: keeps JSONL line-per-edge purity; codemaps gains a frontmatter/footer block):

```json
{
  "provenance": {
    "schema_version": 1,
    "tree_root": "/abs/path/of/generating/tree",
    "commit_sha": "<40-hex or null>",
    "dirty": false,
    "content_fingerprint": "<sha256 of described/inventoried content, present when dirty>",
    "source_fingerprints": { "<layer-specific sources>": "<sha256>" },
    "generated_by": "mx-scan|graph-build|codemaps-gen",
    "generated_at": "<RFC3339, display-only — never a freshness signal>"
  }
}
```

`generated_at` is display-only: encoding it as a freshness input would reintroduce the mtime defect through the back door.

Precedent: nav-graph `extract_commit_sha` (nav-tokens.md) and astx `Provenance`/`CurrentProvenance` — reuse the existing shape conventions where they fit rather than inventing a third.

## §3 Refresh Cache (M2)

- **Key**: `<abs-tree-root>` + per-file `<sha256(content)>`. Two worktrees of one repository never share an entry (AC-GF-010). Never keyed by repo identity, remote URL, or module path alone.
- **Incremental unit**: the file. mx rescan touches only files whose hash differs from the inventory; edges rebuild recomputes only fingerprint-mismatching source sets.
- **Answer provenance**: every answer embeds tree root + commit (or `dirty` + fingerprint) — the same block §2 defines, so query consumers inherit the freshness contract without a second mechanism.
- **Cost budget**: measured per refresh (duration + files-re-read), compared against a config budget; overrun warns and still answers. Warn-not-fail because a stale-but-labeled answer beats no answer.

**Alternative rejected**: cache under `.moai/state/graph-cache/` keyed by HEAD SHA only — breaks on dirty trees (uncommitted edits are the common case M2 exists for) and collides across worktrees at the same SHA with different dirty states.

## §4 Citation Anchor (M3)

Form: `excerpt + sha256(region-content) [+ file path] [+ convenience line] [+ measured-tree SHA]`.

- The hash covers the **cited region's content**, not the whole file: whole-file hashing breaks on any edit anywhere in the file; region hashing survives unrelated edits and line drift above the region (AC-GF-013/014).
- Resolution: match excerpt (normalized whitespace) within the named file; hash confirms; on mismatch (the region itself was edited), report the mismatch — honest staleness beats force-resolution.
- Line numbers remain renderable for humans but no resolver may depend on them.

**Alternative rejected**: symbol-name anchoring alone — names collide (method sets across types) and rename invisibly; excerpt+hash carries its own verification data.

## §5 Code-Derived Edge Layers (M4)

Edge schema (additive to the 5 doc-derived types; fields beyond the doc layers' base set):

```
{ "type": "code-call" | "code-import", "source": <symbol>, "target": <symbol>,
  "grade": "full" | "name-based", "origin": "astx", ...base fields }
```

- **Additivity invariant**: enabling code layers must leave the doc-derived edge set byte-identical (AC-GF-018 asserts E_doc ⊆ E_out unchanged). The build keeps its fail-open-per-layer posture for doc layers; the ONE strict path is the matrix defect check.
- **Disagreement**: same `(source, target, relationship-kind)` claimed by doc and code layers with contradictory polarity/shape ⇒ both edges emitted + `disagrees_with` cross-reference marker. Never a pick.
- **astx seam**: the extractor gains call/import capture (new `.scm` captures — today's 16 files capture declarations only, measured) and a consumer-facing API; navigator keeps its current enrichment path unchanged. Dependency direction: `internal/graph` (or a new `internal/graph/symbol` seam) → astx, with no navigator-tier imports (AC-GF-016 via `go list -deps`).
- **Resolution grading**: per language, `full` (scope-aware resolution of call targets) | `name-based` (name match without scope) | `none` (no capture available). The matrix publishes all 16 cells; empty cell ⇒ defect verdict (AC-GF-019). `IsScaffolded`-style placeholder queries must grade honestly (`none` or `name-based`), not pass silently.

## §6 MCP Tool Shapes (M5)

All three: signature-level output, provenance block in every response, read-only.

- `graph_file_api(file) -> { file, symbols: [{name, kind, signature, exported}], provenance }`
- `graph_find_code(query) -> { matches: [{symbol, file, signature, grade}], provenance }`
- `graph_trace_calls(symbol, direction=callers|callees, depth) -> { edges: [{from, to, grade}], provenance }`

Registration follows the existing `add()` pattern + tool catalog (catalog constant list — the count moves 21 → 24; `.claude/rules/moai/core/moai-mcp-tools.md` + template mirrors sync in M5 per Template-First).

**Alternative rejected**: full-source bodies in responses — the graft analysis's ~10% token figure is foreign, but the direction (signatures carry the query value; bodies are a Read away) holds regardless of the exact number, and signatures-only keeps responses bounded.

## §7 Gate & CI Wiring (M1)

- `moai gate`: new step in the quality-gate step set, default **warn-only** (ast-grep precedent — pre-commit is the wrong place to force codemaps regeneration), enabled/blocking via gate config (SSOT loader path already shared between CLI and PreToolUse hook).
- CI: standalone `.github/workflows/graph-freshness.yml` (spec-lint.yml precedent — small focused workflow, keeps ci.yml stable). The job BOOTSTRAPS the mechanical layers before checking, in sequence: build the moai binary from the PR head → `moai mx scan` (mx-index refreshed to head) → `moai graph build` (edges refreshed to head) → `moai graph check`. The bootstrap scopes the CI signal to codemaps drift — the tracked, curated layer: the untracked mechanical layers are refreshed to head by the job itself, so an `absent` verdict for them cannot fire in CI (without the bootstrap the job would be structurally always-red on every fresh checkout — the D1 audit chain). A code PR that skips codemaps regeneration still shows the accumulated codemaps drift — the intended signal.
- Exit contract: 0 fresh · 1 stale/absent · 2 system error (internal/cli conventions).

## §8 What This Design Does NOT Do

- No Graft installation, no TS/npm runtime, no Graft hook set (excluded).
- No LLM anywhere in M1-M5 (excluded; the semantic layer stays human-authored).
- No auto-rewrite of codemaps from a query path (§1 refresh-owner row).
- No LSP resolution (excluded; grades come from AST extraction only).
- No foreign benchmark numbers as defaults or promises (constraint; §3's budget is measured locally).
