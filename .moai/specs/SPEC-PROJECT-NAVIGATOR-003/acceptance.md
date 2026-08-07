# Acceptance Criteria — SPEC-PROJECT-NAVIGATOR-003

> Given-When-Then scenarios. The `REQ-NT-*` cross-references point at `spec.md` §C requirements. This file is the SSOT for the AC count (Tier L ceiling: ≤25 AC; this SPEC carries 20 AC).

## §D. AC Matrix

### AC-NT-001 — codemaps emits enriched file when 001 capability-map exists

**Given** a project where `.moai/project/navigator/capability-map.md` exists (001 has been run) and at least one SPEC carries a non-empty `module:` field
**When** `/moai codemaps` is invoked
**Then** `.moai/project/codemaps/capability-symbols.md` and `.moai/project/codemaps/capability-symbols.json` both exist after the invocation completes, with non-empty `rows[]` content.

*REQ-NT-001*

### AC-NT-002 — 001 absence is graceful

**Given** a project where `.moai/project/navigator/capability-map.md` does NOT exist (001 not yet run, or zero-SPEC project)
**When** `/moai codemaps` is invoked
**Then** codemaps completes its existing 5 output files normally, an info log naming the missing input is emitted, and NO `capability-symbols.*` file is written (or, if written, carries `rows: []`).

*REQ-NT-002*

### AC-NT-003 — non-modification of 001 + 002 surfaces

**Given** a snapshot of `.moai/project/navigator/navigator.md`, `.moai/project/navigator/capability-map.md`, `.moai/project/navigator/progress-map.md`, `.moai/project/navigator/audit-report.md`, and `.moai/project/navigator/audit-report.json` taken before `/moai codemaps`
**When** `/moai codemaps` is invoked
**Then** the after-snapshot of all five files is byte-identical to the before-snapshot (verified by `sha256sum` diff).

*REQ-NT-003 + REQ-NT-019*

### AC-NT-004 — 14 grammars extract symbols on a polyglot fixture

**Given** a fixture project with one source file per language for each of the 14 working languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, swift), each declaring known symbols (function + type)
**When** the extractor is invoked over the fixture
**Then** for each of the 14 languages, the returned `SymbolSet.Supported == true` AND the returned symbol list contains the expected function + type names.

*REQ-NT-004*

### AC-NT-005 — r/flutter fail-open with Supported: false

**Given** a fixture project containing `analysis.r` and `main.dart` files declaring known symbols
**When** the extractor is invoked with `language: "r"` or `language: "flutter"`
**Then** the returned `SymbolSet.Supported == false` AND an info log naming the language is emitted AND the other rows (for working languages) extract normally.

*REQ-NT-005*

### AC-NT-006 — adding a query file extends the language set (no Go edit)

**Given** the extractor with a NEW `queries/yaml.scm` file added under `internal/navigator/astx/queries/` AND a registration-table entry for `"yaml"` added in a SINGLE data/config location (no Go logic change)
**When** the extractor is rebuilt (`make build`) and invoked with `language: "yaml"`
**Then** the extractor returns `SymbolSet.Supported == true` for yaml (demonstrating REQ-NT-006's data-file extension property).

*REQ-NT-006*

### AC-NT-007 — per-language .scm defines symbols

**Given** the 14 working `.scm` query files shipped under `internal/navigator/astx/queries/`
**When** each query file is inspected
**Then** each query file captures at minimum a function-definition node type and a type/class/interface-definition node type for its language (verified by grepping the query file for `@symbol.function` and `@symbol.type` or equivalent).

*REQ-NT-007*

### AC-NT-008 — malformed source tolerance

**Given** a fixture where one source file under the implementation-path is malformed (binary content, encoding issue, OR syntactically broken), AND the remaining files are well-formed
**When** the extractor is invoked over the implementation-path
**Then** the malformed file is skipped, a warning line is appended to `.moai/logs/navigator-astx.log`, the remaining files' symbols are returned, and the extractor does NOT abort.

*REQ-NT-008*

### AC-NT-009 — provenance per row

**Given** an enriched `capability-symbols.json` produced by a run at HEAD `<sha>`
**When** each row in `rows[]` is inspected
**Then** the row carries `extract_commit_sha` == `<sha>` AND the top-level `extracted_at` matches the committer date of `<sha>` from `git log` (NOT wall-clock — verified by comparing to `git log -1 --format=%cI <sha>`).

*REQ-NT-009*

### AC-NT-010 — dual output with stable schema

**Given** an enriched run
**When** the output is inspected
**Then** exactly two files exist under `.moai/project/codemaps/`: `capability-symbols.md` and `capability-symbols.json` (no other top-level enrichment files), AND the JSON's top-level fields are `{extracted_at, extract_commit, source_capability_map, rows[]}`, AND each row carries the fields listed in design.md §5.2.

*REQ-NT-010*

### AC-NT-011 — header-driven join to 001 capability-map

**Given** a synthetic 001 capability-map fixture whose column ORDER has been permuted (e.g. spec-id moved from column 1 to column 3) but whose header names are preserved
**When** the extractor joins enriched rows to the capability-map
**Then** the join resolves columns by header name and produces correctly-paired enriched rows (no misalignment), demonstrating REQ-NT-011's robustness against 001's unfrozen column order.

*REQ-NT-011*

### AC-NT-012 — idempotence

**Given** two consecutive `/moai codemaps` invocations on the same HEAD commit with no intervening change to 001's capability-map or the source tree
**When** the two `capability-symbols.{md,json}` outputs are compared
**Then** both files are byte-identical between the two runs (verified by `sha256sum` diff; the only time field is `extract_commit`'s committer date, which is identical).

*REQ-NT-012*

### AC-NT-013 — atomic writes

**Given** a test harness that performs synchronized reads of `capability-symbols.md` at sub-write-interval cadence during an extraction
**When** the extractor writes the file via `<file>.tmp` → `mv`
**Then** no read ever observes a partial file (the file content is either the pre-write version or the complete post-write version, never truncated mid-record). Verified via the `NAVIGATOR_PRE_RENAME_BARRIER` test hook (mirroring 001's atomic-rename fixture pattern).

*REQ-NT-013*

### AC-NT-014 — path ceiling truncates with marker

**Given** an implementation-path fixture containing 2500 source files AND the config `navigator.astx.max_files_per_path: 2000`
**When** the extractor walks the path
**Then** extraction stops at 2000 files, the resulting row carries `truncated: true`, and the `symbol_count` reflects only the first 2000 files (not all 2500).

*REQ-NT-014*

### AC-NT-015 — cgo disabled: project builds, nocgo path returns Supported: false

**Given** a build environment with `CGO_ENABLED=0`
**When** `go build ./...` is run
**Then** the build succeeds (the nocgo stub path compiles), AND `SupportedLanguages()` returns an empty list (or every entry returns `Supported: false`), AND invoking `Extract(ctx, anyPath, anyLang)` returns `SymbolSet{Supported: false}` without panicking.

*REQ-NT-015*

### AC-NT-016 — template neutrality grep clean

**Given** the template-distributed 003 surfaces: `references/navigator-astx.md`, `scripts/navigator-enrich.sh`, and the extended `codemaps.md` skill body, all under `internal/template/templates/`
**When** the template neutrality grep is run (searching for internal SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs — the C2/C3/C7 forbidden classes per CLAUDE.local.md §25.1)
**Then** zero matches are found (CI guard: `internal/template/internal_content_leak_test.go` + `.github/workflows/template-neutrality-check.yaml`).

*REQ-NT-016*

### AC-NT-017 — Template-First mirror in place

**Given** the 003 plan-phase artifacts committed
**When** `make build` is run
**Then** the local `.claude/skills/moai-workflow-project/scripts/navigator-enrich.sh`, the local `.claude/skills/moai/workflows/codemaps.md`, and the local `internal/navigator/astx/` package all reflect the template source under `internal/template/templates/` (verified by `sha256sum` comparison).

*REQ-NT-017*

### AC-NT-018 — extractor grep touches no LSEL surface

**Given** the extractor source files (the new Go package + the new bash script) and the extended codemaps skill body
**When** grepped for any path or symbol referencing `.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_`, or `hns-lsel-`
**Then** zero matches are found.

*REQ-NT-018*

### AC-NT-019 — extractor grep touches no 002 audit surface

**Given** the extractor source files
**When** grepped for any path or symbol referencing `.moai/project/navigator/audit-report.md`, `.moai/project/navigator/audit-report.json`, or `audit-known-matches.yaml`
**Then** zero matches are found (demonstrating 003 + 002 non-overlap — REQ-NT-019).

*REQ-NT-019*

### AC-NT-020 — codemaps pipeline classification preserved

**Given** the extended `.claude/skills/moai/workflows/codemaps.md` skill body
**When** inspected for LLM-driven control flow (Agent() invocations that select the next phase, conditional phase branching driven by an LLM)
**Then** zero such patterns are found — the AST-enrichment step is executor delegation inside Phase 3, NOT a phase-selection mechanism (the Agentless pipeline contract is preserved — REQ-NT-020). Verified by the existing CI guard `internal/template/agentless_audit_test.go`.

*REQ-NT-020*

## §D.1 Severity classification

| AC | Severity | Rationale |
|----|----------|-----------|
| AC-NT-001, AC-NT-004, AC-NT-007, AC-NT-010 | MUST-PASS | core feature contract |
| AC-NT-002, AC-NT-005, AC-NT-008, AC-NT-014, AC-NT-015 | MUST-PASS | fail-open / graceful-degradation contracts |
| AC-NT-003, AC-NT-011, AC-NT-018, AC-NT-019 | MUST-PASS | boundary integrity (001/002/LSEL non-overlap) |
| AC-NT-009, AC-NT-012, AC-NT-013 | MUST-PASS | provenance + idempotence + atomicity (verification-claim-integrity alignment) |
| AC-NT-006, AC-NT-017, AC-NT-020 | MUST-PASS | extensibility + template-First + pipeline-class invariants |
| AC-NT-016 | MUST-PASS | template neutrality (CLAUDE.local.md §25 CI gate) |

All 20 AC are MUST-PASS (none deferred; none downgraded to SHOULD). The SPEC carries no PASS-WITH-DEBT items at plan-phase.

## §D.2 Traceability

Every REQ-NT-* requirement in `spec.md` §C is covered by at least one AC here:

| REQ | Covering AC |
|-----|-------------|
| REQ-NT-001 | AC-NT-001 |
| REQ-NT-002 | AC-NT-002 |
| REQ-NT-003 | AC-NT-003 |
| REQ-NT-004 | AC-NT-004 |
| REQ-NT-005 | AC-NT-005 |
| REQ-NT-006 | AC-NT-006 |
| REQ-NT-007 | AC-NT-007 |
| REQ-NT-008 | AC-NT-008 |
| REQ-NT-009 | AC-NT-009 |
| REQ-NT-010 | AC-NT-010 |
| REQ-NT-011 | AC-NT-011 |
| REQ-NT-012 | AC-NT-012 |
| REQ-NT-013 | AC-NT-013 |
| REQ-NT-014 | AC-NT-014 |
| REQ-NT-015 | AC-NT-015 |
| REQ-NT-016 | AC-NT-016 |
| REQ-NT-017 | AC-NT-017 |
| REQ-NT-018 | AC-NT-018 |
| REQ-NT-019 | AC-NT-019 |
| REQ-NT-020 | AC-NT-020 |

Coverage: 20/20 REQ covered, 20 AC total. Within the Tier L ≤25 ceiling for each, counted independently.

## §D.3 Indirect verification (gaps + indirect-AC mapping)

The following SPEC invariants are verified indirectly via the AC matrix above:

| Invariant | Indirect verification |
|-----------|-----------------------|
| 16-language neutrality (no Go-bias) | AC-NT-004 (14 grammars) + AC-NT-005 (r/flutter fail-open) + AC-NT-006 (data-file extension) + AC-NT-016 (template neutrality grep) |
| Provenance per row (verification-claim-integrity §2) | AC-NT-009 (extract-commit-sha + committer date) |
| Non-overlap with LSEL (carry-forward from 001 REQ-PN-016) | AC-NT-018 |
| Non-overlap with 002 audit | AC-NT-019 |
| Pipeline classification preserved | AC-NT-020 |
| 001 contract preserved | AC-NT-003 (byte-identical before/after) + AC-NT-011 (header-driven join) |

## §D.4 Closure gates (Definition of Done)

This SPEC is `completed` when ALL of the following hold:

1. All 20 AC in §D show PASS (no FAIL, no DEFERRED).
2. `make build` succeeds and regenerates template-mirrored artifacts (AC-NT-017).
3. The polyglot fixture (AC-NT-004) is checked in under `internal/navigator/astx/testdata/` and runs in CI.
4. The template-neutrality CI guard passes (AC-NT-016).
5. The codemaps Agentless-pipeline CI guard passes (AC-NT-020).
6. Run-phase + sync-phase land via Route B PR (per CLAUDE.local.md §23 — PR-mandatory, `enforce_admins:true`).
7. No `[NEEDS CLARIFICATION]` markers remain in `plan.md` or `research.md`.

## §D.5 Forward-looking checks (sanity, not blocking)

These checks run at sync-phase as defense-in-depth; they are NOT blocking for `completed` status:

- `moai spec audit --json` classifies this SPEC as `era: V3R6` (H-4 detection via progress.md §E.2-§E.4 markers).
- The progress.md `sync_commit_sha` field is populated in the sync commit (backfilled per the SHA-placeholder backfill exemption D3 if necessary).
- The closed-SPEC lint ownership-transition check passes (the `draft → in-progress` transition is performed by `manager-develop`'s first run-phase commit; the `in-progress → implemented → completed` transition rides the sync commit from `manager-docs`).
