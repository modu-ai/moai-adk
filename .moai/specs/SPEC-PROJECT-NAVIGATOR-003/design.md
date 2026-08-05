# Design — SPEC-PROJECT-NAVIGATOR-003

> System design for the tree-sitter integration. The artifact set is plan-phase only; this document records architecture decisions for plan-auditor review and for run-phase M1–M6 implementation. Cross-references `research.md` for grounding and `plan.md` for milestone sequencing.

## §1. System architecture overview

```
                            /moai codemaps  (Agentless fixed-pipeline)
                                            │
                                            ▼
                        ┌─────────────────────────────────────┐
                        │  Phase 1: Explore (Explore subagent) │  (existing, unchanged)
                        └─────────────┬───────────────────────┘
                                      │
                                      ▼
                        ┌─────────────────────────────────────┐
                        │  Phase 2: Architecture Analysis      │  (existing, unchanged)
                        └─────────────┬───────────────────────┘
                                      │
                                      ▼
                        ┌─────────────────────────────────────┐
                        │  Phase 3: Map Generation             │
                        │   ├── existing maps (overview, ...)  │  (existing, unchanged)
                        │   └── NEW: AST-enrichment step ──────┼──┐
                        └─────────────────────────────────────┘  │
                                                                 │
                          ┌──────────────────────────────────────┘
                          ▼
                ┌─────────────────────────────────────────────┐
                │  scripts/navigator-enrich.sh                │  (new, self-contained bash)
                │   1. read 001's capability-map.md           │
                │      (header-driven join, per REQ-NT-011)   │
                │   2. for each row:                          │
                │      resolve implementation-path            │
                │      detect language (path/extension)       │
                │      invoke extractor (Go entry point)      │
                │   3. emit capability-symbols.{md,json}      │
                │      (atomic-rename, idempotent)            │
                └────────────┬────────────────────────────────┘
                             │
                             ▼
                ┌─────────────────────────────────────────────┐
                │  internal/navigator/astx/ (new Go package)  │
                │   ├── SupportedLanguages() []string         │
                │   ├── Extract(ctx, path, lang) SymbolSet    │
                │   ├── queries/*.scm  (//go:embed)           │
                │   │   14 working + 2 scaffolded stubs       │
                │   ├── measure_cgo.go    (//go:build cgo)    │
                │   └── measure_nocgo.go  (//go:build !cgo)   │
                └────────────┬────────────────────────────────┘
                             │
                             ▼
                ┌─────────────────────────────────────────────┐
                │  github.com/smacker/go-tree-sitter          │  (already in go.mod)
                │   14 grammar sub-packages                   │
                └─────────────────────────────────────────────┘
                             │
                             ▼
                .moai/project/codemaps/capability-symbols.md
                .moai/project/codemaps/capability-symbols.json
```

The enrichment step is **strictly additive** to `/moai codemaps` Phase 3. The existing five codemaps output files are unchanged; 003 emits two new sibling files. The pipeline classification (Agentless) is preserved because the enrichment step is deterministic Go + bash, not an LLM-driven control-flow branch.

## §2. Grammar-per-language model

### §2.1 Registration table

The package's `supportedLanguages` map is the single point of registration for the 14 working grammars + 2 scaffolded stubs. Each entry carries:

| Field | Type | Example |
|-------|------|---------|
| `language` | string | `"go"` |
| `grammar` | `*sitter.Language` | `golang.GetLanguage()` |
| `queryFile` | embedded `[]byte` | `//go:embed queries/go.scm` |
| `extensions` | `[]string` | `[".go"]` |
| `supported` | bool | `true` for 14, `false` for r/dart |
| `symbolNodeTypes` | `[]string` | `["function_declaration", "method_declaration", "type_declaration"]` |

Adding a 17th language (or upgrading r/dart from scaffolded to working) is a registration-table entry + a `.scm` query file. NO Go source logic change (REQ-NT-006).

### §2.2 Language detection

The extractor detects the language of a source file by extension, using the `extensions` field of the registration table. Files whose extension matches no registered language are skipped with an info log (not a warning — unknown extensions are expected in any polyglot repo's `vendor/`, `.md`, `.yaml`, etc.).

### §2.3 Query design — per-language `.scm`

Each `.scm` query file captures the symbol kinds named in research §3.1 (function/method/type). Example for Go:

```scheme
(function_declaration name: (identifier) @symbol.function)
(method_declaration name: (field_identifier) @symbol.method)
(type_declaration (type_spec name: (type_identifier) @symbol.type))
```

The captures use a naming convention `@symbol.<kind>` so the extractor can group symbols by kind without language-specific logic. The 14 working query files are authored in M2; the 2 scaffolded stubs (`r.scm`, `dart.scm`) are empty files whose presence in the embed set documents that the language is known-but-unsupported.

## §3. Symbol-extraction approach

### §3.1 The extraction algorithm

```
Extract(ctx, sourcePath, language) -> SymbolSet
  1. Look up language in registration table.
     - If supported=false or absent → return SymbolSet{Supported: false}.
  2. Read sourcePath. On read error → return SymbolSet{Supported: ..., Error: err}.
  3. Parse with tree-sitter grammar. On parse error:
     - If error-recovery produced a partial tree → use the partial tree.
     - If hard parse failure → return SymbolSet{..., Error: err}.
  4. Run the language's .scm query against the tree.
  5. Collect captures, group by @symbol.<kind>, deduplicate by name.
  6. Return SymbolSet{Supported: true, Symbols: grouped, SourceBytes: n}.
```

### §3.2 The walk algorithm (per capability row)

```
EnrichRow(row) -> EnrichedRow
  1. Resolve row.implementation-path against project root.
     - If path does not exist → EnrichedRow{on_disk_verified: false, ...}.
  2. Walk the path recursively, collecting files whose extension matches a registered language.
     - Apply the per-path file-count ceiling (REQ-NT-014). If exceeded, mark truncated: true.
  3. For each file, call Extract(ctx, file, detectedLanguage).
     - Skip unsupported languages (Supported: false).
     - Skip files that error-open (REQ-NT-008) — log warning, continue.
  4. Aggregate symbols across files:
     - primary_files: top-N files by symbol count (default N=5)
     - primary_symbols: top-N symbols across the path (default N=10, configurable)
     - symbol_count: total (dedup) symbol count across the path
     - on_disk_verified: true if path exists AND ≥1 file parsed successfully
     - extract_language: the dominant language by file count
  5. Stamp provenance: extract_commit_sha = `git rev-parse HEAD`, captured_at = committer date.
```

### §3.3 The top-N heuristic

`primary_symbols` is a top-N (default ≤10) by a frequency heuristic (occurrence count in the path), NOT a PageRank ranking (research §1 Aider — PageRank is out of scope per spec.md §E). The top-N keeps the enriched markdown small (research §1 Codebase-Memory — map reduces context, not inflates it) while surfacing the most representative symbols. The full symbol list per row is available in the JSON output for downstream tooling (REQ-NT-010 dual output).

## §4. `/moai codemaps` integration points

### §4.1 Where 003 hooks in

| Codemaps phase | Existing behavior | 003 change |
|----------------|-------------------|------------|
| Phase 1 (Explore) | Explore subagent scans codebase | UNCHANGED |
| Phase 2 (Architecture Analysis) | orchestrator-direct analysis | UNCHANGED |
| Phase 3 (Map Generation) | orchestrator-direct generation of 5 map files | EXTENDED: after the 5 existing files, invoke `scripts/navigator-enrich.sh` to emit the 2 new sibling files |
| Phase 4 (Verification) | verify references, dependencies, entry points | EXTENDED: verify the enriched output references real files (the on-disk-verified column) |
| Phase 5 (Report) | report to user | UNCHANGED |

### §4.2 Capability gate (REQ-NT-001 vs REQ-NT-002)

```
Phase 3 extension:
  IF .moai/project/navigator/capability-map.md EXISTS:
    run scripts/navigator-enrich.sh
    emit capability-symbols.{md,json}
  ELSE:
    log "navigator: capability-map.md absent, skipping AST enrichment"
    (continue to Phase 4 — codemaps completes normally)
```

This is a pure capability gate, not an LLM dispatch. The Agentless pipeline classification is preserved (REQ-NT-020).

### §4.3 Repeatability

Each `/moai codemaps` invocation runs the enrichment once (idempotent — REQ-NT-012). Re-running `/moai codemaps` on the same HEAD produces byte-identical output. The script reads 001's existing `capability-map.md` (which itself is regenerated on `/moai sync` per 001's REQ-PN-003); 003 does NOT trigger 001's regeneration.

## §5. Output schema

### §5.1 `capability-symbols.md` (human-readable)

```markdown
# Capability Symbols (AST-derived)

Extracted at: <ISO-8601 commit date>
Extract commit: <git rev-parse HEAD>
Source capability-map: .moai/project/navigator/capability-map.md

## go

| spec-id | title | implementation-path | on-disk | files | symbols | primary-symbols |
|---------|-------|---------------------|---------|-------|---------|-----------------|
| SPEC-AUTH-001 | OAuth2 | internal/auth | ✓ | 12 | 47 | AuthHandler, Login, RefreshToken, ... |
| ... | | | | | | |

## python

| ... |

## (unsupported: r)

- (no rows — language scaffolded, Supported: false)

## (unsupported: flutter)

- (no rows — language scaffolded, Supported: false)
```

### §5.2 `capability-symbols.json` (machine-readable, stable schema)

```json
{
  "extracted_at": "<ISO-8601 commit date>",
  "extract_commit": "<sha>",
  "source_capability_map": ".moai/project/navigator/capability-map.md",
  "rows": [
    {
      "spec_id": "SPEC-AUTH-001",
      "title": "OAuth2 Authentication",
      "implementation_path": "internal/auth",
      "on_disk_verified": true,
      "extract_language": "go",
      "primary_files": ["internal/auth/handler.go", "internal/auth/token.go", ...],
      "primary_symbols": [
        {"name": "AuthHandler", "kind": "type", "file": "internal/auth/handler.go"},
        {"name": "Login", "kind": "function", "file": "internal/auth/handler.go"},
        ...
      ],
      "symbol_count": 47,
      "truncated": false,
      "supported": true
    },
    ...
  ]
}
```

Schema stability follows 002's contract: forward-compatible (additive field additions only; no field removals; no semantic redefinitions of existing fields).

## §6. Failure modes + recovery

| Failure | Detection | Recovery |
|---------|-----------|----------|
| 001 capability-map absent | Phase 3 capability gate | Skip enrichment, log, continue (REQ-NT-002) |
| implementation-path not on disk | Path resolution | Mark `on_disk_verified: false`, continue |
| Source file parse failure | tree-sitter error / file read error | Skip file, log warning, continue (REQ-NT-008) |
| Language unsupported (r/flutter) | Registration table `supported: false` | Emit row with `Supported: false`, continue (REQ-NT-005) |
| Path exceeds file-count ceiling | Walk counter | Mark `truncated: true`, stop walk at ceiling (REQ-NT-014) |
| cgo disabled at build time | Build tag | nocgo stub returns `Supported: false` for every language (REQ-NT-015) |
| Two sessions extract concurrently | Atomic-rename | Later write wins (REQ-NT-013) |
| `git rev-parse HEAD` fails | Shell exit code | Use `<unknown>` placeholder + log; do not abort |

Every failure mode degrades to a log line + continue, never aborts `/moai codemaps`. This invariant aligns with Advisory-Check Discipline (extraction is NOT latency-sensitive; the user invoked `/moai codemaps` explicitly) and with the verification-claim-integrity §1.1 surface 3 (no unobserved defect claims — the `on_disk_verified: false` marker IS the observed defect claim, attributable to the path-resolution check).

## §7. Boundary integrity (cross-SPEC)

| Surface | 003 reads? | 003 writes? | Notes |
|---------|-----------|-------------|-------|
| `.moai/project/navigator/navigator.md` (001) | NO | NO | REQ-NT-003 |
| `.moai/project/navigator/capability-map.md` (001) | YES (header-driven) | NO | REQ-NT-011 |
| `.moai/project/navigator/progress-map.md` (001) | NO | NO | REQ-NT-003 |
| `.moai/project/navigator/audit-report.{md,json}` (002) | NO | NO | REQ-NT-019 |
| `.moai/project/codemaps/capability-symbols.{md,json}` (003 NEW) | YES (own output) | YES (own output) | REQ-NT-010 |
| `.moai/project/codemaps/{overview,modules,...}.md` (codemaps) | NO | NO | 003 adds sibling files only |
| `.moai/specs/SPEC-*/spec.md` (SPEC registry) | YES (frontmatter only) | NO | Language fallback detection |
| Project source tree | YES (read-only walk) | NO | REQ-NT-018 |
| `.moai/lessons-inbox.jsonl` (LSEL) | NO | NO | REQ-NT-018 non-overlap |
| `.moai/state/lsel/` (LSEL) | NO | NO | REQ-NT-018 non-overlap |
| `memory/feedback_*.md` (LSEL) | NO | NO | REQ-NT-018 non-overlap |
| `hns-lsel-*` skills (LSEL) | NO | NO | REQ-NT-018 non-overlap |

The read/write sets are fully disjoint from 001 (except the single read-only header-driven join on `capability-map.md`), from 002, and from LSEL.

## §8. Performance model

| Workload | Cost model | Ceiling |
|----------|------------|---------|
| Files parsed per capability row | O(file count under implementation-path) | bounded by REQ-NT-014 ceiling (default 2000) |
| Bytes parsed per file | O(source-file size) | typical source file <100KB |
| Parse time per file | O(file size) — tree-sitter is linear in practice | dominated by file I/O, not parsing |
| Symbols emitted per row | O(symbols in path) | top-N capture (default 10) keeps markdown bounded; JSON carries full set |
| Total enrichment time | O(sum of file counts across all capability rows) | bounded by total source-tree size under all implementation-paths; monorepo-scale paths truncated per REQ-NT-014 |

The cost is paid once per `/moai codemaps` invocation. There is no per-session cost (REQ — no SessionStart hook) and no per-turn cost (no Stop hook). The user pays only when they explicitly invoke `/moai codemaps`.

## §9. Open design questions (forward-looking, NOT blockers)

These questions are NOT blocking plan-phase sign-off; they are recorded for run-phase M1–M6 to resolve at implementation time, with the listed default.

| Question | Default | Forward-compat |
|----------|---------|----------------|
| Top-N value for `primary_symbols` | 10 | Configurable via `navigator.astx.primary_symbols_n` |
| Top-N value for `primary_files` | 5 | Configurable via `navigator.astx.primary_files_n` |
| Per-path file-count ceiling | 2000 | Configurable via `navigator.astx.max_files_per_path` |
| Whether to surface `@MX:ANCHOR` symbols specially | No (out of scope, spec.md §E) | Future SPEC may cross-reference |
| Whether to cache parse results across invocations | No (idempotence suffices) | Future SPEC may add content-hash cache |

These defaults are recorded in plan.md §D Constraints and are NOT in-spec variables (the SPEC text hard-codes the defaults; the config keys override at runtime, mirroring 001's `navigator.staleness_cycles` pattern).
