# plan.md — SPEC-MX-SCANNER-DOCS-001

## §A. Context

P2-3 in the MX activation queue. Documentation-only: no scanner source changes (REQ-MSD-009). The deliverable is (a) a new docs-site page covering 4 advanced scanner behaviors across 4 locales, (b) cross-reference links from 2 existing pages, (c) README FAQ entries across 4 files.

The run/sync phase will produce the actual prose. This plan.md carries the **research findings** that are the factual basis for that prose — each finding cites the source file and line so the docs author can verify any claim against the code rather than against another doc.

## §B. Known Issues

- None blocking. The 4 features are confirmed to exist in code (see §C). All four behaviors are unambiguous in source; no user decision is required before run-phase entry.

## §C. Pre-flight — Research Findings (factual basis for documentation)

### C.1 rotRisk scoring

**What it is**: a string field on `@MX:DEBT` tags that flags whether the simplification has an exit condition.

**Source of truth**:
- `internal/mx/tag.go:62-65` — `RotRisk` field declared on `Tag`, json tag `rotRisk,omitempty`, comment "flags a DEBT tag whose @MX:UPGRADE trigger is absent".
- `internal/mx/resolver_query.go:127-130` — same field projected into `TagResult` (the query/sidecar shape).
- `internal/mx/scanner.go:214-219` — when a `@MX:DEBT` tag is appended, `RotRisk` is set to the pessimistic default `"no-trigger"`. `pendingDebtIdx` records the tag's position so a later sub-line can resolve it.
- `internal/mx/scanner.go:110-116` — when an `@MX:UPGRADE` sub-line follows the pending DEBT, `RotRisk` is cleared to `""` (i.e. omitted from JSON) and `pendingDebtIdx` resets.
- `internal/cli/mx_scan.go:66-72, 136-138` — `moai mx scan` counts DEBT tags where `RotRisk == "no-trigger"` and prints a `DEBT rotRisk (missing @MX:UPGRADE): N` summary line.

**Documentation gist**: `rotRisk` is a DEBT-only field. The scanner pessimistically assumes a DEBT rots (`"no-trigger"`) unless a subsequent `@MX:UPGRADE` sub-line clears it. The absence of `@MX:CEILING` is a quality note, NOT the rot gate — only `@MX:UPGRADE` presence clears rotRisk. The `moai mx scan` summary surfaces the rot count. (Aligns verbatim with `.claude/rules/moai/workflow/mx-tag-protocol.md` "When to Add Tags → @MX:DEBT" section.)

### C.2 LSP fan-in condition

**What it is**: the caller count for `@MX:ANCHOR` tags, computed LSP-primary with a textual fallback. The threshold itself (`fan_in >= 3`) is the ANCHOR suggestion gate and lives in the mx.yaml `thresholds.fan_in_anchor` config.

**Source of truth**:
- `internal/mx/fanin.go:11-23` — `FanInCounter` interface with two implementations: `TextualFanInCounter` (grep fallback) and `LSPFanInCounter`.
- `internal/mx/fanin.go:116-167` — textual `Count`: walks the project root, skips `vendor/`, `node_modules/`, `.git/`, optionally excludes test files (built-in `_test.go`/`testdata` plus user `test_paths` globs from mx.yaml), counts lines containing the AnchorID substring.
- `internal/mx/fanin_lsp.go:27-45` — `LSPFanInCounter` struct holds an `LSPReferencesClient` (the `FindReferences` + `IsAvailable` subset of the LSP client).
- `internal/mx/fanin_lsp.go:54-89` — LSP `Count` flow: (1) read `MOAI_MX_QUERY_STRICT` env; (2) if LSP unavailable → strict returns `LSPRequiredError`, non-strict falls back; (3) call `FindReferences` at the tag's position (LSP 0-based, tag line 1-based — `tag.Line - 1`); (4) on LSP error → strict returns `LSPRequiredError`, non-strict falls back; (5) filter `excludeTests` (drops `_test.go`/`testdata` paths); (6) return count with `fan_in_method = "lsp"`.
- `internal/mx/fanin_lsp.go:111-119` — `textualFallback` delegates to `TextualFanInCounter` with `fan_in_method = "textual"`.
- `internal/mx/resolver_query.go:118-122` — sidecar fields `FanIn int` and `FanInMethod string` (`"lsp"` or `"textual"`) projected into query output.

**Documentation gist**: ANCHOR fan-in is LSP-primary. `fan_in_method` in the query result tells you which engine produced the count. Non-strict mode (default) silently falls back to textual grep when LSP is unavailable or errors; strict mode (`MOAI_MX_QUERY_STRICT=1`) raises `LSPRequiredError` instead. Test files are excludable via `excludeTests` + the mx.yaml `test_paths` globs.

### C.3 CGO complexity path

**What it is**: cyclomatic-complexity measurement for the WARN/ANCHOR suggestion engine, implemented via tree-sitter AST queries. CGO and non-CGO builds get DIFFERENT `measure` symbols via Go build constraints.

**Source of truth**:
- `internal/hook/mx/complexity/complexity.go:1-44` — package doc + the `Measure` entry point. `Result{Cyclomatic, IfBranches, Supported}`. `maxFileSizeBytes = 1 << 20` (1 MiB) caps input. `Measure` delegates to `measure(...)`.
- `internal/hook/mx/complexity/measure_cgo.go:1` — `//go:build cgo`. The full tree-sitter implementation.
- `internal/hook/mx/complexity/measure_cgo.go:87-166` — CGO `measure`: file-size guard → scaffolded-language stub → seeded-language lookup → parse → find function node by name (and `startLine` hint) → compile decision query → `SetPointRange` constrained to the function's byte range → count `@decision` and `@if_branch` captures → return `Cyclomatic = decisionCount + 1` (McCabe).
- `internal/hook/mx/complexity/measure_nocgo.go:1-9` — `//go:build !cgo`. A 9-line stub: returns `Result{Supported: false}` for ALL inputs, because tree-sitter parsing requires CGO.
- `internal/hook/mx/complexity/complexity.go:33-39` — even on CGO, `Supported: false` is returned for: unsupported/scaffolded languages, content > 1 MiB, any parse error, any query compile error, function-not-found. Errors are swallowed (logged at `slog.Debug`), never propagated — the validation pipeline never blocks on a complexity measurement failure.

**WHY CGO needs a separate path**: tree-sitter's Go bindings (`github.com/smacker/go-tree-sitter`) are CGO-bound. A non-CGO build (e.g. `CGO_ENABLED=0` for static distro binaries) cannot load the native parser libraries, so the package ships a build-tag stub that returns `Supported: false` for everything rather than failing to compile. The validation pipeline treats `Supported: false` as "skip complexity-based suggestions for this function" — never as an error.

**Documentation gist**: complexity measurement is CGO-gated. Distro builds with `CGO_ENABLED=0` get a stub that reports `Supported: false` for every language; the suggestion engine silently skips complexity-gated WARN recommendations on such builds. Even on CGO builds, scaffolded (unseeded) languages, files >1 MiB, and parse/query errors all yield `Supported: false`.

### C.4 Scan automation timing

**What it is**: the lifecycle points that build or consult the `@MX` sidecar index (`.moai/state/mx-index.json`), beyond the explicit `moai mx scan` CLI.

**Source of truth**:
- `internal/cli/mx_scan.go:24-125` — the explicit `moai mx scan` CLI. Default scan root = project root; `--path` narrows; `--dry` previews without writing; writes `mx.Sidecar{SchemaVersion, Tags, ScannedAt}` to `.moai/state/mx-index.json`. Advisory-only: scanner warnings/errors go to stderr, never block.
- `internal/hook/session_start.go:267-271, 448-470` — SessionStart runs an advisory drift scan first, THEN triggers the deferred MX cold-start full scan so `moai mx query` returns fresh results on a new session.
- `internal/hook/session_start.go:1327-1346` — the cold-start scan is time-boxed by `mxIndexScanTimeout`, whose default is `mxIndexScanTimeoutDefault` = **2s** (`session_start.go:1340`). This is the cold-start SCAN ceiling on huge repos. On timeout the scan is abandoned for this session (fail-open, non-blocking).
- `internal/hook/session_start.go:1216-1232` — DISTINCT constant: `DefaultSessionStartDriftTimeout` = **2s** (`session_start.go:1223`) is the DRIFT-scan ceiling, NOT the cold-start scan ceiling. The two 2s constants are independent — they bound different scans (drift vs cold-start full scan) and happen to share the same 2s value by coincidence, not by sharing a single ceiling. The docs prose MUST name them separately so readers do not conflate them.
- `internal/hook/session_start.go:1376-1433` — `runMXColdStartScan`: builds a scanner with `mx.DefaultScanIgnore`, calls `ScanDir`, writes the sidecar. All failure paths log and continue (fail-open); the 5s hook budget is never blocked by the scan.
- `internal/hook/mx/config.go:10-32, 75-98` — the validation config: `PostToolUse` (default enabled, 500ms timeout) and `SessionEnd` (default enabled, 4000ms timeout) hooks run MX validation. The `Sync` section (default `strict`, skip flag `--skip-mx`) gates `/moai sync`. Enforcement levels: P1-anchor and P2-warn are `blocking`; P3-note and P4-todo are `advisory`.
- `internal/hook/mx/types.go:13-46` — `Priority` enum: P1 (exported func with fan_in ≥ 3 missing ANCHOR, blocking), P2 (goroutine missing WARN, blocking), P3 (exported func ≥ 100 lines missing NOTE, advisory), P4 (untested public func missing TODO, advisory). `IsBlocking()` true for P1/P2.
- `internal/hook/mx/validator.go` (header) — the validator is explicitly READ-ONLY: it never modifies source files or tags.

**Documentation gist**: the sidecar is rebuilt (a) explicitly via `moai mx scan`, (b) automatically via the SessionStart deferred cold-start scan (time-boxed ~2s, fail-open). PostToolUse and SessionEnd hooks validate tags against the index but do NOT rebuild it; the sync workflow enforces P1/P2 blocking gates with a `--skip-mx` escape hatch.

## §D. Constraints

- 4-locale same-PR parity for docs-site (REQ-MSD-007) and README (REQ-MSD-008).
- Mermaid TD-only (REQ-MSD-005); diagram source identical across locales.
- Icon shortcodes over emoji (REQ-MSD-006); typography arrows and branding-emoji-in-code-blocks permitted.
- Light single-theme; no dark-theme branches.
- URL whitelist: `adk.mo.ai.kr` only.
- No scanner source modification (REQ-MSD-009).

## §E. Self-Verification (plan-phase)

- [x] SPEC ID regex check executed: `SPEC-MX-SCANNER-DOCS-001` → PASS.
- [x] SPEC ID uniqueness verified against `.moai/specs/` — no prior occurrence.
- [x] All 12 canonical frontmatter fields present in spec.md.
- [x] GEARS notation used for all REQ-* (Ubiquitous + Unwanted patterns).
- [x] Out of Scope section carries ≥1 `### Out of Scope — <topic>` H3 with bullets.
- [x] All 4 research findings cite source files (§C) — the docs author can verify any claim against code.
- [x] Extend-vs-new-page decision documented (§F) with rationale.

## §F. Extend-vs-New-Page Decision

**Decision: NEW PAGE** at `docs-site/content/{ko,en,ja,zh}/advanced/mx-scanner-internals.md`, plus cross-reference links from the two existing pages.

**Rationale**:
- `advanced/mx-tags.md` (104 lines, all locales) is a "how to write tags" user reference — Tag Syntax, Types, Sublines, When to Add, Lifecycle. Adding 4 internal-behavior sections (rotRisk scoring semantics, LSP-vs-textual fan-in engine, CGO build-tag stub, scan lifecycle hooks) would shift its purpose and push it past its comfortable length.
- `utility-commands/moai-mx.md` (114 lines, all locales) is the `moai mx` command reference. Scan automation timing fits its scope partially, but rotRisk/fan-in/CGO are scanner internals, not command-shape docs.
- The 4 features share a common audience (the user trying to interpret query output or reason about index freshness) and a common voice (here is what the scanner does that is not obvious from the tag syntax). A single dedicated page serves that audience without fragmenting across two existing pages whose scope would be diluted.
- **Cross-reference links** (in scope, REQ-MSD-007): `mx-tags.md` "DEBT" section → links to the new page's rotRisk subsection; `moai-mx.md` "Execution flow" section → links to the new page's scan-automation subsection. This keeps the new page discoverable from both existing entry points.

**Menu wiring** (in scope per §G of spec.md): add the new page to `data/menu/main.yaml` and per-locale `content/<locale>/_meta.yaml` so it appears in the sidebar. No new icon values (reuse an existing one, e.g. `wrench` or `search`).

**Diagram plan** (Mermaid TD-only): one diagram on the new page — a `flowchart TD` of the scan lifecycle (SessionStart → drift check → cold-start scan → sidecar; PostToolUse/SessionEnd → validate-against-sidecar; sync → enforce). The rotRisk, fan-in, and CGO behaviors are described in prose + inline code snippets, not diagrams (their flow is too small to merit separate diagrams).

## §G. Anti-Patterns to Avoid

- **AP-MSD-001**: authoring canonical content in a non-canonical locale first (docs-site canonical = ko; README canonical = en). Author in the canonical locale, derive the rest.
- **AP-MSD-002**: translating diagram source between locales (the Mermaid block must be byte-identical across the 4 locales; only the surrounding prose translates).
- **AP-MSD-003**: describing a scanner behavior the code does not have (e.g. claiming non-CGO builds measure complexity via a "simpler heuristic" — they do not; they stub). Every behavior claim must trace to a §C citation.
- **AP-MSD-004**: using body-text emoji for callouts instead of the `{{</* icon */>}}` shortcode.
- **AP-MSD-005**: editing scanner source "while documenting it" — explicitly forbidden by REQ-MSD-009.

## §H. Cross-References

- `spec.md` — GEARS requirements (REQ-MSD-001..009).
- `acceptance.md` — Given-When-Then matrix.
- `.moai/docs/docs-site-i18n-rules.md` + CLAUDE.local.md §17.1 — design and i18n regime.
- Skill `hns-oss-docs-i18n-rules` — the loaded i18n rules digest.
- Skill `hns-oss-docs-verify` — the runnable verify recipe (run/sync phase).

## §I. Milestones (priority-ordered, no time estimates)

1. **M1 — New docs-site page (canonical locale)**: author `docs-site/content/ko/advanced/mx-scanner-internals.md` covering all 4 features using §C findings; add Mermaid TD scan-lifecycle diagram; add to `data/menu/main.yaml` + `content/ko/_meta.yaml`.
2. **M2 — New docs-site page (derived locales)**: translate to en/ja/zh preserving section structure and diagram source verbatim; add to each locale's `_meta.yaml`.
3. **M3 — Cross-reference links**: add links from `mx-tags.md` (DEBT section → rotRisk subsection) and `moai-mx.md` (Execution flow → scan automation subsection) in all 4 locales.
4. **M4 — README FAQ entries**: author English entries in `README.md`; derive into `README.ko.md`, `README.ja.md`, `README.zh.md` with heading/section parity.
5. **M5 — Verify**: run the `hns-oss-docs-verify` recipe (warning-free Hugo build, sitemap, URL blacklist, Mermaid TD-only, 4-locale file/section parity, README heading parity, body-emoji scan); confirm zero scanner source files modified.
