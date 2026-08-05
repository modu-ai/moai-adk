---
id: SPEC-PROJECT-NAVIGATOR-003
title: "Project Navigator — tree-sitter auto-derivation into /moai codemaps (16-language AST-based capability rows)"
version: "0.1.0"
status: completed
created: 2026-08-05
updated: 2026-08-05
author: manager-spec
priority: P3
phase: "v3.3 target"
module: project-navigator
lifecycle: spec-anchored
tier: L
era: V3R6
tags: "navigator, tree-sitter, codemaps, ast, auto-derivation, 16-language"
related_specs: [SPEC-PROJECT-NAVIGATOR-001, SPEC-PROJECT-NAVIGATOR-002]
---

# SPEC-PROJECT-NAVIGATOR-003 — tree-sitter auto-derivation (Project Navigator Epic, P3)

## HISTORY

- 2026-08-05 (stub) — Reserved SPEC ID + recorded boundary decision. Stub deferred full authoring until SPEC-PROJECT-NAVIGATOR-001 (capability-map format) and SPEC-PROJECT-NAVIGATOR-002 (audit algorithm) landed.
- 2026-08-05 (plan-phase expansion, iter-1) — Full Tier L expansion. 001 MERGED (PR #1354 squash `2c87d195f`, status: completed) and 002 MERGED (PR #1361 squash `927e268c9` + backfill #1362 `da4cfd258`, status: completed). This SPEC mechanically enriches 001's capability-map rows with AST-extracted file/symbol/status columns via tree-sitter, integrated into `/moai codemaps` so a codemaps regeneration produces enriched rows as a sibling output. Tier L (16-language grammar coverage is a genuinely large verification surface; cgo + binary dependency + parsing-compatibility surface; /moai codemaps integration touches multiple files across Go source + template + script). Research grounding: Aider repo map (tree-sitter + PageRank) and Codebase-Memory (tree-sitter + SQLite), carried forward from 001's `research.md`; this SPEC's `research.md` adds tree-sitter-specific grounding (grammar availability across the 16 languages, the parsing/symbol model, CI matrix implications).

## §A. User Story

**As a** MoAI-ADK user maintaining a polyglot project (any of the 16 supported languages, not only Go) who relies on the Project Navigator to reorient returning sessions to the current frontier,
**I want** the capability map to carry MECHANICALLY EXTRACTED file/symbol/status columns derived from the actual AST — not just hand-derived text from the SPEC registry — so that a returning session can see, at a glance, which symbols (functions, types, classes, methods) implement each capability and whether the implementation path the capability-map names actually exists on disk,
**so that** a session resuming after a `/clear` or a multi-day gap reads ONE enriched capability row and knows the capability's owning SPEC, its declared implementation path, the symbols that constitute that implementation, and a机械 extraction of "is the code actually there" — instead of trusting a hand-derived path that may have drifted from the source.

**Outcome hypotheses:**
- `/moai codemaps` regeneration produces an enriched output file (`.moai/project/codemaps/capability-symbols.md`) whose rows join 001's capability-map text columns (spec-id, title, implementation-path, commit-sha, captured-at — read header-driven per 002's REQ-NA-007 lesson) with AST-derived columns (primary-files, primary-symbols, symbol-count, on-disk-verified).
- 14 of 16 supported languages ship working tree-sitter grammars; the remaining 2 (`r`, `flutter`/dart) degrade fail-open with a `Supported: false` marker per row, mirroring the existing `internal/hook/mx/complexity` package's scaffolded-language contract.
- The 16-language neutrality invariant (CLAUDE.local.md §15) is preserved: the generator's per-language query files ship as data, not code, and adding a 17th language is a query-file addition, not a Go source edit.

## §B. Context and Background

### §B.1 The gap (research-backed — see research.md)

001's capability-map is hand-derived: rows are produced by `navigator-regen.sh` reading SPEC frontmatter + git log. The implementation-path column is whatever text the SPEC's `module:` field names — it is not verified against the disk, and it carries no symbol-level breakdown. This is the right contract for 001 (hand-derived is enough at the substrate layer), but it leaves two gaps a returning session feels:

| Gap | What 001 produces | What a returning session also wants |
|-----|-------------------|-------------------------------------|
| Path verification | `implementation-path: internal/spec` (text from SPEC `module:`) | "Does `internal/spec/` actually exist on disk, and what symbols does it contain?" |
| Symbol rollup | (none) | "Which functions/types/classes implement this capability?" |
| Status derivation | `status: implemented` (SPEC frontmatter) | "Is there actual code on disk matching the declared path?" |

This SPEC closes the three gaps by adding a MECHANICAL layer on top of 001's hand-derived layer. The mechanical layer does not replace 001; it enriches it.

### §B.2 What 003 does NOT touch (boundary, non-negotiable)

- **001's artifact set** (`.moai/project/navigator/{navigator,capability-map,progress-map}.md`) — REQ-PN-001 mandates "exactly three living documents ... and no other top-level Navigator files." 003 SHALL NOT add a fourth file there, SHALL NOT modify `navigator-regen.sh`, and SHALL NOT modify the three files' contents. 001 is `status: completed` and its contract is frozen.
- **002's audit algorithm** — 002 parses 001's capability-map header-driven (REQ-NA-007). 003's enriched rows live in a SIBLING file under `.moai/project/codemaps/`. 002's heuristic continues to operate on 001's text columns; it does NOT depend on 003's AST columns. (Forward-compat: a future SPEC MAY teach 002's heuristic to also consult 003's on-disk-verified column, but that is not in scope here.)
- **LSEL surfaces** — `.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*` skills. 003 is project-scoped; LSEL is harness-scoped (per 001's REQ-PN-016 non-overlap, carried forward here as REQ-NT-013).

### §B.3 Output surface (decided — see plan.md §C for full rationale)

- **Output location**: `.moai/project/codemaps/capability-symbols.md` — a SIBLING to the existing codemaps output files (`overview.md`, `modules.md`, `dependencies.md`, `entry-points.md`, `data-flow.md`). This is the natural home per the Epic boundary statement ("integrated into `/moai codemaps`") and avoids any collision with 001's three-file contract.
- **Output shape**: one row per capability (joined to 001's capability-map by spec-id, header-driven), with columns `spec-id | title | implementation-path | primary-files | primary-symbols | symbol-count | on-disk-verified | extract-language | extract-commit-sha | captured-at`. The first three columns are echoed from 001's capability-map (header-driven per the 002 lesson); the rest are AST-derived.
- **Optional machine output**: `.moai/project/codemaps/capability-symbols.json` (stable schema, one entry per spec-id, carries the full symbol list per row for downstream tooling) — emitted alongside the markdown, mirroring 002's dual-output contract (REQ-NA-005).

### §B.4 Mechanism (decided — see plan.md §C for full rationale)

- **Pipeline classification** (preserved): `/moai codemaps` is Agentless fixed-pipeline (localize → repair → validate). 003 adds an AST-enrichment step INSIDE the existing Phase 3 (map generation), NOT a new phase. The pipeline contract is unchanged.
- **Grammar library**: reuse `github.com/smacker/go-tree-sitter` (already in `go.mod`, already used by `internal/hook/mx/complexity`). NO new top-level dependency. The library requires cgo; 003 follows the existing `internal/hook/mx/complexity/measure_{cgo,nocgo}.go` build-tag split so the project still builds without cgo.
- **Per-language query files**: tree-sitter query files (`.scm`) define what counts as a "symbol" per language. They ship as `//go:embed` data, mirroring the complexity package. 14 query files for 14 grammars; 2 scaffolded stubs (`r.scm`, `dart.scm`) that return `Supported: false` because `smacker/go-tree-sitter` ships no grammar for those languages today.
- **New Go package**: `internal/navigator/astx/` — self-contained symbol-extraction library. Exports `Extract(ctx, filePath, language) (SymbolSet, error)` and `SupportedLanguages() []string`. No dependency on the `moai` binary; no AskUserQuestion surface; no hook registration.
- **New script**: `scripts/navigator-enrich.sh` (sibling to `scripts/navigator-regen.sh` and `scripts/navigator-audit.sh`) — self-contained bash that reads 001's `capability-map.md` header-driven, resolves each row's implementation-path, invokes a small `moai navigator-astx` Go subcommand (or an embedded extractor binary) per file, and writes the enriched output atomically. Inherits 001's portability + atomic-write + idempotence contracts.
- **No SessionStart hook**: AST extraction is NOT latency-sensitive (it walks the whole codebase). Per Advisory-Check Discipline, extraction runs ONLY on `/moai codemaps` invocation, NOT on every session start.

### §B.5 Inputs (read-only)

The extractor consumes; it does not own:

- `.moai/project/navigator/capability-map.md` (001's output — read header-driven, joining by spec-id)
- The source files under each row's `implementation-path` (the codebase itself)
- `.moai/specs/SPEC-*/spec.md` frontmatter (for fallback language detection per the `module:` path)
- `git log` (commit-sha + committer date for provenance — identical source to 001 + 002)
- The 14 tree-sitter grammars embedded at build time (compile-time constant)

### §B.6 Non-inputs (boundaries)

- Does NOT read 001's `progress-map.md` (per-SPEC progress is irrelevant to AST extraction).
- Does NOT read 002's `audit-report.{md,json}` (audit is downstream of 001, not upstream of 003).
- Does NOT read `@MX` tag corpus (symbol-level annotations are a separate concern; a future SPEC may cross-reference `@MX:ANCHOR` rows with 003's symbol set, but that is not in scope).
- Does NOT read or write any LSEL surface.

## §C. Requirements (GEARS)

### §C.1 Codemaps integration (P3 core)

#### REQ-NT-001 (Capability-gate — codemaps emits enriched file)
**Where** `/moai codemaps` is invoked AND 001's `.moai/project/navigator/capability-map.md` exists on disk, the codemaps workflow SHALL run the AST-enrichment step during its Phase 3 (map generation) and emit `.moai/project/codemaps/capability-symbols.md` alongside the existing codemaps output files, so a single `/moai codemaps` invocation leaves the enriched surface current.

#### REQ-NT-002 (Event-driven — 001 absence is graceful)
**When** `/moai codemaps` is invoked AND 001's `capability-map.md` does NOT exist (001 not yet run, or zero-SPEC project), the AST-enrichment step SHALL be skipped with an info log naming the missing input, and codemaps SHALL complete its remaining phases normally, so 003 never blocks codemaps on 001's absence.

#### REQ-NT-003 (Ubiquitous — non-modification of 001 + 002)
The AST-enrichment step SHALL NOT modify `.moai/project/navigator/*` (001's surface) or `.moai/project/navigator/audit-report.{md,json}` (002's surface), so 001 and 002 remain `status: completed` with frozen contracts.

### §C.2 16-language grammar coverage

#### REQ-NT-004 (Ubiquitous — 14 grammars via smacker/go-tree-sitter)
The extractor SHALL ship working tree-sitter grammars for exactly 14 of the 16 supported languages — go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift — using the `github.com/smacker/go-tree-sitter` library already in `go.mod`, so the most common polyglot projects get mechanical symbol extraction without a new top-level dependency.

#### REQ-NT-005 (Capability-gate — 2 scaffolded languages fail-open)
**Where** the detected language is `r` OR `flutter` (dart) — the 2 of 16 supported languages for which `smacker/go-tree-sitter` ships no grammar — the extractor SHALL return a `Supported: false` marker for that row, emit an info log naming the language, and continue processing the remaining rows, so r/flutter source trees do not block extraction and the scaffolded-language contract mirrors `internal/hook/mx/complexity`'s existing pattern.

#### REQ-NT-006 (Ubiquitous — 16-language neutrality)
The extractor's per-language configuration SHALL ship as data (`.scm` query files + a language-registration table), NOT as Go source per language, so adding a 17th language or upgrading an existing grammar is a data-file change, not a code change — preserving the 16-language neutrality invariant (CLAUDE.local.md §15).

### §C.3 Symbol extraction

#### REQ-NT-007 (Ubiquitous — symbol definition per language)
The extractor SHALL use per-language tree-sitter query files (`.scm`) to define what node types count as "symbols" for each language — minimally: function definitions, method definitions, type/class/interface definitions — so the symbol set is language-appropriate and not biased toward Go's syntax.

#### REQ-NT-008 (Event-driven — malformed source tolerance)
**When** the extractor encounters a source file that fails to parse (syntax error, encoding issue, binary file, file removed mid-walk), the extractor SHALL skip that file, append a warning line to `.moai/logs/navigator-astx.log`, and continue with the remaining files, so a single broken file does not abort extraction for the whole capability row.

#### REQ-NT-009 (Ubiquitous — provenance per row)
Every enriched row SHALL carry an `extract-commit-sha` (the HEAD commit SHA at extraction time, sourced from `git rev-parse HEAD`) plus an ISO-8601 `captured-at` timestamp (sourced from `git log` for the commit's committer date, NOT wall-clock), so any enriched claim is attributable to a measured baseline per `.claude/rules/moai/core/verification-claim-integrity.md` §2.

### §C.4 Output format

#### REQ-NT-010 (Ubiquitous — dual output)
The AST-enrichment step SHALL emit exactly two output files — `.moai/project/codemaps/capability-symbols.md` (human-readable, one row per capability, grouped by detected language) and `.moai/project/codemaps/capability-symbols.json` (machine-readable, stable schema with `extracted_at`, `extract_commit`, `rows[]` fields, where each row carries the full symbol list) — and no other top-level enrichment file, mirroring 002's dual-output contract (REQ-NA-005).

#### REQ-NT-011 (Capability-gate — header-driven join to 001)
The enriched output SHALL join to 001's capability-map **by header name** (parsed from the capability-map's header row, per 002's REQ-NA-007 lesson), NOT by fixed column position, so 003 remains robust to 001's unfrozen column order and does not redefine 001's schema.

### §C.5 Determinism + concurrency

#### REQ-NT-012 (Ubiquitous — idempotence)
The AST-enrichment step SHALL be idempotent: two extractions over the same HEAD commit + same 001 capability-map + same source tree SHALL produce byte-identical `capability-symbols.{md,json}` output, so a no-op re-extraction is a safe operation and the report carries no wall-clock timestamps (the only time field is `extract_commit`'s committer date from `git log`, identical to 001 + 002's provenance contract).

#### REQ-NT-013 (Event-driven — atomic writes)
**When** two sessions run extraction concurrently, the writes SHALL land via an atomic-rename strategy (write to `<file>.tmp` then `mv` into place), so the later write wins and no reader observes a partially-written file — inheriting 001's REQ-PN-008 contract.

#### REQ-NT-014 (Capability-gate — bounded extraction cost)
**Where** a capability row's `implementation-path` resolves to a directory tree exceeding a configurable file-count ceiling (default `navigator.astx.max_files_per_path: 2000`, overridable via config), the extractor SHALL cap the walk at the ceiling, emit a `truncated: true` marker on the row, and continue, so a monorepo-scale path cannot make `/moai codemaps` unbounded.

### §C.6 cgo + build-tag split

#### REQ-NT-015 (Ubiquitous — cgo / nocgo build-tag split)
The extractor SHALL ship under build tags `//go:build cgo` (real tree-sitter path) and `//go:build !cgo` (stub path returning `Supported: false` for every language with an explanatory comment), mirroring `internal/hook/mx/complexity/measure_{cgo,nocgo}.go`, so the project continues to build in cgo-disabled environments and the nocgo path degrades cleanly rather than failing to compile.

### §C.7 Template / distribution

#### REQ-NT-016 (Ubiquitous — template-neutrality, constraint inherited from 001/002)
The template-distributed 003 surfaces (the new `references/navigator-astx.md` Level-3 reference, the new `scripts/navigator-enrich.sh`, and the extended `codemaps.md` skill body) SHALL be template-neutral (no internal SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs — the C2/C3/C7 forbidden classes per CLAUDE.local.md §25.1) and SHALL carry no Go-bias in any per-language example, so the feature ships equally to all 16 supported languages.

#### REQ-NT-017 (Ubiquitous — Template-First)
The new Go package source, the new script, the new query files, and the extended skill body SHALL ship via `internal/template/templates/` + `make build`, and the local `.claude/` + `internal/` copies SHALL be regenerated from the template source, so a downstream `moai update` preserves the feature.

### §C.8 Boundary + non-overlap

#### REQ-NT-018 (Ubiquitous — non-overlap with LSEL)
The extractor SHALL operate strictly on `.moai/project/navigator/capability-map.md` (read-only), `.moai/specs/SPEC-*/spec.md` frontmatter (read-only), the project source tree (read-only), and `git log` (read-only) — and SHALL NOT read or write any LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`), carrying forward 001's REQ-PN-016 non-overlap to the 003 boundary.

#### REQ-NT-019 (Ubiquitous — non-overlap with 002 audit)
The extractor SHALL NOT read or write `.moai/project/navigator/audit-report.{md,json}` (002's surface) and SHALL NOT invoke the audit algorithm, so 002's audit remains a downstream consumer of 001's text columns and is not coupled to 003's AST columns (forward-compat: a future SPEC MAY teach 002 to consult `on-disk-verified`, but that is out of scope here).

### §C.9 Pipeline-contract preservation

#### REQ-NT-020 (Ubiquitous — Agentless pipeline preserved)
The AST-enrichment step SHALL execute as executor delegation INSIDE `/moai codemaps` Phase 3 (map generation), NOT as a new workflow phase and NOT as an LLM-driven control-flow branch, so `/moai codemaps` retains its Agentless fixed-pipeline classification (localize → repair → validate) per `.claude/rules/moai/workflow/spec-workflow.md` § Subcommand Classification.

## §D. Acceptance Criteria Matrix (summary)

Full Given-When-Then scenarios live in `acceptance.md`. Summary:

| AC | REQ | Check |
|----|-----|-------|
| AC-NT-001 | REQ-NT-001 | `/moai codemaps` produces `capability-symbols.{md,json}` when 001's capability-map exists |
| AC-NT-002 | REQ-NT-002 | 001-absent case → skip + info log + codemaps completes |
| AC-NT-003 | REQ-NT-003 | `navigator/` + `audit-report.*` unchanged across codemaps invocation |
| AC-NT-004 | REQ-NT-004 | 14 grammars extract symbols on a polyglot fixture |
| AC-NT-005 | REQ-NT-005 | r/flutter rows carry `Supported: false`, others extract normally |
| AC-NT-006 | REQ-NT-006 | Adding a `.scm` query file (no Go edit) extends the language set |
| AC-NT-007 | REQ-NT-007 | Per-language `.scm` defines symbols (function/method/type for ≥8 languages) |
| AC-NT-008 | REQ-NT-008 | Malformed source file → skip + warning logged, others extract |
| AC-NT-009 | REQ-NT-009 | Every enriched row carries `extract-commit-sha` + ISO-8601 `captured-at` |
| AC-NT-010 | REQ-NT-010 | Dual output (md + json), stable schema, no extras |
| AC-NT-011 | REQ-NT-011 | Header-driven join to 001 verified against a fixture with reordered columns |
| AC-NT-012 | REQ-NT-012 | Two extractions on same HEAD → byte-identical output |
| AC-NT-013 | REQ-NT-013 | Atomic-rename write strategy verified |
| AC-NT-014 | REQ-NT-014 | Path ceiling truncates extraction + marks `truncated: true` |
| AC-NT-015 | REQ-NT-015 | `GO_CGO_ENABLED=0 go build ./...` succeeds; nocgo path returns `Supported: false` |
| AC-NT-016 | REQ-NT-016 | Template neutrality grep clean (C2/C3/C7 classes absent) |
| AC-NT-017 | REQ-NT-017 | Template-First: `.claude/` + `internal/` regenerated from template source via `make build` |
| AC-NT-018 | REQ-NT-018 | Extractor grep touches no LSEL surface |
| AC-NT-019 | REQ-NT-019 | Extractor grep touches no 002 audit surface |
| AC-NT-020 | REQ-NT-020 | `/moai codemaps` remains Agentless pipeline (no LLM-driven phase selection introduced) |

## §E. Out of Scope

### Out of Scope — Navigator artifact generation (001's surface)

- Generating or regenerating `.moai/project/navigator/{navigator,capability-map,progress-map}.md` is owned by **SPEC-PROJECT-NAVIGATOR-001** (status: completed). 003 reads `capability-map.md` header-driven; it never writes to 001's directory and never invokes `navigator-regen.sh`.

### Out of Scope — drift / completeness audit (002's surface)

- The `--audit` mode that diffs design intent against the implemented capability-map is owned by **SPEC-PROJECT-NAVIGATOR-002** (status: completed). 003's enriched rows live under `.moai/project/codemaps/`, a SIBLING surface; 002's audit continues to operate on 001's text columns. Forward-compat: a future SPEC MAY teach 002 to consult 003's `on-disk-verified` column, but that coupling is not in scope here.

### Out of Scope — semantic / RAG / PageRank symbol ranking

- Aider-style PageRank symbol ranking (research.md §Aider) and embeddings-based semantic similarity over the source tree are out of scope. 003 emits a flat per-capability symbol list with no importance ranking. Symbol ranking is a possible future SPEC; it is NOT required for the enriched capability row to be useful (the symbol COUNT + on-disk verification alone close the three gaps named in §B.1).

### Out of Scope — r-language and dart/flutter tree-sitter grammars

- Writing or vendoring a tree-sitter grammar for `r` or for `dart`/`flutter` is out of scope. 003 fails-open with `Supported: false` for those two languages (REQ-NT-005). When `smacker/go-tree-sitter` (or a successor package) adds those grammars, 003's data-file extension (REQ-NT-006) makes adoption a query-file + registration-table change, not a source edit.

### Out of Scope — LSEL-owned harness self-evolution

- Closing the PROPOSE→APPLY seam for harness self-edit is owned by **SPEC-LSEL-LOCAL-EVOLUTION-001**. 003 consumes the SPEC registry + the source tree; it does not consume the lessons-inbox. REQ-NT-018 codifies the non-overlap.

### Out of Scope — MX tag cross-reference

- Cross-referencing the extracted symbol set with the `@MX:ANCHOR` / `@MX:SPEC` tag corpus (e.g. "which extracted symbols carry an `@MX:ANCHOR` tag and which are untagged") is out of scope. A future SPEC may address it once 003's enriched output has stabilized.

### Out of Scope — SessionStart ambient auto-brief of enriched rows

- 001's SessionStart hook (`handle-session-start-navigator.sh`) emits a ≤500-token brief drawn from `navigator.md`. 003's enriched rows are NOT surfaced ambiently; they live in `.moai/project/codemaps/capability-symbols.md` and are read on demand by `/moai codemaps` consumers. Wiring 003 into the SessionStart brief would violate Advisory-Check Discipline (extraction is not latency-sensitive).

## §F. Constraints (Non-Functional)

- **Provenance**: every enriched row carries `extract-commit-sha` + ISO-8601 `captured-at` (REQ-NT-009), aligns with verification-claim-integrity.md §2 (no unobserved claims).
- **Idempotent extraction**: same HEAD + same 001 capability-map + same source tree → byte-identical output (REQ-NT-012). Timestamps sourced from `git log`, not wall-clock, mirroring 001 + 002.
- **Atomic writes**: `capability-symbols.{md,json}.tmp` → `mv` (REQ-NT-013, inherited from 001's REQ-PN-008).
- **Bounded extraction**: per-path file-count ceiling (default 2000, overridable), `truncated: true` marker when hit (REQ-NT-014) — so `/moai codemaps` never becomes unbounded.
- **Fail-open**: missing 001, missing grammar, malformed source, missing design docs — every failure mode degrades to a log line + continue, never aborts codemaps (REQ-NT-002, REQ-NT-005, REQ-NT-008).
- **cgo / nocgo parity**: project builds with cgo disabled; nocgo path is a stub (REQ-NT-015).
- **16-language neutrality**: no Go-bias (constraint inherited from CLAUDE.local.md §15, carried from 001/002).
- **Template-First**: new files ship via `internal/template/templates/` + `make build` + §25 neutrality (REQ-NT-016, REQ-NT-017).
- **PR-mandatory**: `enforce_admins:true` (CLAUDE.local.md §23). Plan-phase artifacts committed via the standard plan→PR flow; run-phase via Route B PR.
- **No new Go CLI subcommand for end users**: the extraction is invoked BY `/moai codemaps` internally. A `moai navigator-astx` Go subcommand MAY exist as an internal extractor entry point (sibling to 002's pattern of self-contained bash), but the user-facing surface remains `/moai codemaps` (constraint inherited from 001 §B.1).
- **Pipeline-class preservation**: `/moai codemaps` stays Agentless fixed-pipeline (REQ-NT-020).

## §G. Dependencies / Related SPECs

- **SPEC-PROJECT-NAVIGATOR-001** (`status: completed`, PR #1354 squash `2c87d195f`, Tier M) — the Navigator substrate. 003 reads 001's `capability-map.md` header-driven (REQ-NT-011); 003 does NOT write to 001's surface (REQ-NT-003). 001's `capability-map` column order is NOT a stable contract (001's spec.md declares column 1 as `capability`; 001's acceptance.md AC-PN-013 enumerates spec-id-first) — 003 therefore joins by header name, the same immunization 002 adopted (REQ-NA-007).
- **SPEC-PROJECT-NAVIGATOR-002** (`status: completed`, PR #1361 squash `927e268c9` + backfill #1362 `da4cfd258`, Tier M) — the audit layer. 003's enriched rows are SIBLING to 002's audit output; non-overlap codified in REQ-NT-019.
- **SPEC-LSEL-LOCAL-EVOLUTION-001** (`status: completed`) — non-overlap boundary (REQ-NT-018).
- `internal/hook/mx/complexity/` (existing Go package) — the prior-art tree-sitter user in this repo. 003's `internal/navigator/astx/` package follows the same build-tag split and the same `//go:embed` query-file pattern; the two packages remain independent (no import edge) because they serve different consumers (complexity is an `@MX`-tag scoring input; astx is a Navigator/codemaps enrichment input).
- `/moai codemaps` workflow (`.claude/skills/moai/workflows/codemaps.md`) — EXTENDED with an AST-enrichment step inside Phase 3 (REQ-NT-001). The pipeline classification is preserved (REQ-NT-020).
- `github.com/smacker/go-tree-sitter` v0.0.0-20240827094217 — already in `go.mod`; 003 adds NO new top-level dependency.

## §H. Resolved decisions (recorded for traceability)

1. **Output location** — CHOSEN: `.moai/project/codemaps/capability-symbols.{md,json}` (codemaps surface, sibling to existing codemaps output files). Rejected alternatives: (a) `.moai/project/navigator/capability-map-enriched.md` — violates 001's REQ-PN-001 "exactly three ... and no other top-level Navigator files"; (b) a new top-level directory — fragments the project-context surface unnecessarily. The codemaps surface is the natural home because 003's enrichment IS code-structural (per the Epic boundary statement "integrated into `/moai codemaps`").
2. **Grammar library** — CHOSEN: reuse `github.com/smacker/go-tree-sitter` (already in `go.mod`, already used by `internal/hook/mx/complexity`). Rejected: a new tree-sitter Go binding, or shell-out to a separate tree-sitter CLI binary. Reusing the existing binding keeps the dependency surface at 1 (no new top-level dep), follows the prior-art pattern, and avoids a subprocess-per-file cost model.
3. **cgo / nocgo split** — CHOSEN: mirror `internal/hook/mx/complexity/measure_{cgo,nocgo}.go`. The nocgo path returns `Supported: false` for every language with an explanatory comment; cgo-disabled environments (some CI matrices, some downstream users) continue to build.
4. **Symbol definition per language** — CHOSEN: per-language `.scm` query files, embedded via `//go:embed`, mirroring `internal/hook/mx/complexity/queries/*.scm`. Rejected: a single universal query (tree-sitter node types differ too much across languages; a universal query would either miss symbols in some languages or produce false positives in others); in-source reflection (no portable reflection API across the 14 grammars).
5. **r / dart / flutter coverage** — CHOSEN: fail-open with `Supported: false`. Rejected: vendoring a third-party grammar (would add a new top-level dependency and a maintenance burden for 2 of 16 languages); dropping those languages from the supported set (would violate the 16-language neutrality invariant). When `smacker/go-tree-sitter` adds those grammars, adoption is a data-file change (REQ-NT-006).
6. **Era frontmatter** — `era: V3R6` set explicitly to avoid H-2 misclassification during the early window when progress.md is sparse (per `.claude/rules/moai/workflow/lifecycle-sync-gate.md` § When to set era: explicitly, item 3 — newly created SPECs before progress.md is populated), mirroring 002's decision.
