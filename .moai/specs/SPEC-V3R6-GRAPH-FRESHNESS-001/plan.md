# Implementation Plan — SPEC-V3R6-GRAPH-FRESHNESS-001

## §A. Context

Card t250 (operator-picked, Tier L). Adopt 4 Graft design principles into moai-adk without adopting the tool: freshness gate (M1), query-time refresh (M2), content-addressed citations (M3), code symbol layer (M4) + MCP code queries (M5). Evidence base: `.moai/reports/graft/graft-analysis-20260824.md` (measured at `294b4b6ab`), re-verified in this tree at `baa100ce5` (research.md §Anchor Re-Verification). Anchor re-verification moved two facts: drift grew 713 → 740 commits, and the fresh-worktree absence of untracked artifacts (mx-index, edges.jsonl) was directly observed — absence semantics (AC-GF-005) went from assumption to measurement.

Substrate already present: tree-sitter dep, 16 `.scm` queries (declarations only — call/import queries are new work), `moai graph build`/`query` with 5 doc-derived edge layers, `moai mx query`, MCP `add()` registration, quality-gate step machinery.

## §B. Known Issues (pre-flight risks)

1. **mx-index has no provenance today** (measured: top-level keys `schema_version`/`tags`/`scanned_at` only). REQ-GF-003 stamping must land before REQ-GF-002's mx-index metric can compute anything. Ordering inside M1: stamping first.
2. **Untracked artifacts have no git history** — any git-based metric is impossible for mx-index/edges. Content fingerprinting is the only option (design.md §Layer Model). An implementation that reaches for `git log` on these layers will compute garbage.
3. **mtime untrustworthy** (measured: worktree checkout resets all mtimes). All metrics content/git-based; `os.Stat` mtime on gated artifacts is a defect, verified per §D.3.
4. **Wrong-tree cache is the observed defect family** (CR #8/t246 lineage). Per-tree keying is a MUST-severity AC (AC-GF-010), not a nicety.
5. **Update-cost figures are foreign** (`~3ms` = 124-file foreign repo). Budget defaults are placeholders until measured here; M1's calibration step covers thresholds, M2's covers cost.
6. **Scope-aware call resolution is genuinely hard** in some languages. The grade matrix exists so the difficulty is published, not hidden; `name-based` and `none` are honest grades, and a scaffolded-but-graded cell is acceptable while an ungraded cell is a defect (AC-GF-019).
7. **`moai graph build` currently fails open per layer** (absent layer contributes zero edges). Additive code layers must preserve this posture for doc layers while the matrix defect check introduces the one place the build is allowed to be strict.

## §C. Pre-flight (before M1)

- [ ] Confirm no parallel SPEC touches `internal/graph`, `internal/cli/graph*.go`, `internal/cli/mx*.go`, `internal/navigator/astx/`, `internal/cli/mcp_server.go`, `internal/hook/quality/` (worktree isolation; lead's pre-spawn sync check).
- [ ] Baseline for M5 (REQ-GF-020): fixed task set defined + Grep/Read counts measured and recorded BEFORE any implementation commit — this is a pre-flight obligation precisely because it cannot be reconstructed later.
- [ ] Threshold calibration data: measure changed-described-source-files count across this repo's recent history to sanity-check the codemaps ≥40 default.
- [ ] Fixture strategy decided: `t.TempDir()` git fixture repositories for check/gate/cache tests (no reliance on the real repo's state).

## §D. Constraints (carried from spec.md §E)

No new deps · no LLM/network in freshness paths · foreign benchmarks never quoted as promises · hardcoding prevention (thresholds in config/defaults, env names in envkeys.go) · template neutrality for any mirrored file · exit codes 0/1/2 · 16-language equality · t.TempDir() isolation.

## §E. Self-Verification (E1..E7 plan-phase deliverable map)

| E# | Deliverable | Where |
|---|---|---|
| E1 | AC PASS/FAIL matrix with verbatim command output | progress.md §E.2 (run-phase) |
| E2 | Cross-platform build (`GOOS=windows`/`linux` vet or build) | run-phase |
| E3 | Affected-package coverage (graph, cli, astx ≥85%) | run-phase |
| E4 | Subagent-boundary grep (no AskUserQuestion in CLI paths) | run-phase |
| E5 | `golangci-lint run` clean on touched packages | run-phase |
| E6 | Branch/push state | run-phase |
| E7 | Mutant-kill executions recorded (AC-GF-004, AC-GF-008 red runs verbatim) | progress.md §E.2 — both mutants are closure-gate items (§D.4.2) |

## §F. Milestones

Execution order M1 → M5 (dependency-forced: M5 depends on M4; M2's cache substrate feeds M4/M5 query surfaces). Decision-review order — present the highest-change-likelihood decisions first for human review: (1) the per-layer metric table + provenance schema (M1, data-model for everything downstream), (2) edge schema for code-derived layers + disagreement representation (M4, new type interface), (3) citation format (M3, user-facing convention), (4) refresh cache keying (M2), (5) MCP output shapes (M5), (6) CI/gate wiring files (mechanical, last).

### M1 — drift gate (priority: High)

Decisions most likely to change: per-layer metric definitions (spec.md REQ-GF-002 table), threshold defaults, provenance block schema (design.md §Provenance Block).

1. Provenance block type + stamping for the three artifacts (mx-index first — it has nothing today).
2. Per-layer metric implementations (content/endpoint based; no mtime).
3. `moai graph check` subcommand: per-layer report + exit code (AC-GF-001/002/004/005).
4. Thresholds in config (defaults in `internal/config`; gate.yaml override) + calibration measurement recorded.
5. `moai gate` step (warn-only default, ast-grep precedent) (AC-GF-006).
6. CI workflow job (standalone `graph-freshness.yml` following the spec-lint.yml precedent): build binary from PR head → bootstrap the mechanical layers (`moai mx scan`, `moai graph build`) → `moai graph check` (AC-GF-007, red demonstrated on a non-main branch; healthy-head-green clause included). Day-one posture (explicit, not implicit): at landing time this repo carries 740+ commits of codemaps drift, so the job as specified lands red on day one. Chosen posture: **regenerate codemaps in the landing PR before the job is enabled** — the threshold keeps its reasoned default and is calibrated from real history in M1's calibration step afterward. Threshold-raising-to-pass-day-one is rejected as calibration-to-pass (a self-defeating gate).

### M2 — query-time refresh (priority: High)

Decisions: cache keying (tree root + content hash), budget mechanics (warn-only).

1. Content-hash inventory for mx-index scan (changed-files-only rescan) (AC-GF-008/009).
2. Source-fingerprint recompute for edges build on `moai graph query`.
3. Per-tree cache + answer provenance naming (AC-GF-010/015).
4. Update-cost budget: measurement, config default, overrun warning (AC-GF-011). Budget default calibrated by local measurement (recorded §E.2).

Scope note (spec.md §B.2): refresh applies to mechanical layers only; codemaps is never auto-rewritten by a query path.

### M3 — citation convention switch (priority: High)

Decisions: citation format (excerpt + region content hash + convenience line + co-stamped tree SHA).

1. Citation anchor type (excerpt + hash) and renderer for codemaps/report surfaces (AC-GF-012).
2. mx-index anchoring switch to file+hash (AC-GF-013).
3. Two-tree resolver + guarantee test (AC-GF-014); mismatch (cited region itself edited) reports honestly rather than force-resolving.
4. Migration posture: new/refreshed citations adopt the canon; existing line-only citations are not mass-rewritten in this SPEC (regeneration surfaces convert incrementally).

### M4 — symbol layer (priority: Medium)

Decisions: edge schema (`code-call`/`code-import` + grade field + disagreement marker), astx seam.

1. astx seam for non-navigator consumers; verify dependency isolation (AC-GF-016).
2. New `.scm` call/import captures for the supported languages (16 files carry declaration captures today; call/import captures are new per language).
3. Resolution grading per language (`full`/`name-based`/`none`) + matrix publication + defect check on empty cells (AC-GF-019).
4. Additive `code-call`/`code-import` layers in `moai graph build` (AC-GF-017/018); doc layers byte-preserved; disagreement marker for contradictions.
5. Fixture: undocumented A→B call appears in blast radius (AC-GF-017).

### M5 — MCP code queries (priority: Medium)

Decisions: tool output shapes (signature-level).

1. `graph_file_api` (first) (AC-GF-020).
2. `graph_find_code` + `graph_trace_calls` (AC-GF-021).
3. Provenance in every response (tree + commit).
4. Post-implementation task-set run vs pre-flight baseline; reduction recorded (AC-GF-022).
5. Docs touch: `.claude/rules/moai/core/moai-mcp-tools.md` tool count/catalog sync (21 → 24) — template-mirrored files follow Template-First + neutrality checklist.

## §G. Anti-Patterns to Avoid

- **Blanket threshold** — one staleness number across layers with different tracking status and update cycles produces always-red or always-green, both uninformative (card directive; the reason REQ-GF-002 is a per-layer table).
- **Silent layer pick** — doc vs code edge disagreement resolved by preference instead of exposure (REQ-GF-015).
- **Foreign benchmark importation** — `~3ms`/−42% figures quoted as moai-adk promises.
- **mtime trust** — any staleness signal from file modification time.
- **Repo-identity cache key** — cache shared across worktrees of one repository (wrong-tree answer family).
- **Reporting without exit code / stamping without refresh** — the two card-named mutants (AC-GF-004/008); every new check's completion condition includes making a failing input and observing the actual failure.
- **Auto-rewriting curated layers** — query paths regenerating codemaps without an LLM would produce mechanical garbage; the gate surfaces the drift instead.

## §H. Cross-References

- spec.md §D REQ table (per-layer metrics) · acceptance.md §D.5 mutant table
- design.md — layer model, provenance schema, cache design, edge schema, matrix format, MCP shapes
- research.md — anchor re-verification measurements + graft analysis provenance
- Related: SPEC-V3R6-DOCS-CODEMAPS-V3-001 (codemaps SSOT), SPEC-DWF-CODEMAPS-PILOT-001
- House rules: internal/cli CLAUDE.md (exit codes, subcommand registration), CLAUDE.local.md §2 Template-First, §6 test isolation
