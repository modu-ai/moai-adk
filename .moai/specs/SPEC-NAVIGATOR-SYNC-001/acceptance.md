# SPEC-NAVIGATOR-SYNC-001 — Acceptance Criteria

> Companion to `spec.md` §C (REQ-NS-001..018) and `plan.md` §F (milestones).
> Format: each AC is `AC-XXX` labeled `Given ... When ... Then ...`, binary-testable.
> Every Evidence entry names the command + the observable output, per `verification-claim-integrity.md` §2.

## §A. Methodology

### §A.1 Non-overlap ACs use the grep-pair protocol

Per the repo lesson `feedback_deadfield_grep_pair_protocol.md`, non-overlap ACs (AC-013/014/015/016) use TWO grep lenses:

1. **Write-surface grep** — `internal/navigator/sync/` source code must reference the allowed write path `nav-graph.json` and must NOT reference the forbidden write paths.
2. **Runtime write grep** — the running binary's actual writes must land ONLY on `.moai/project/navigator/nav-graph.json` (+ `.tmp` transient), verified by a temp-dir fixture that captures all filesystem writes.

A single lens (source-only OR runtime-only) is insufficient — the source may reference a path without writing it, and the runtime may write a path the source constructs dynamically. Both lenses required.

### §A.2 Byte-identical re-run ACs use the characterization pattern

AC-009 (REQ-NS-009) uses: run the join twice on the same HEAD → `cmp` the two outputs → exit 0 on identical. The fixture pins HEAD via a temp git repo with a known commit SHA, so the test is deterministic.

### §A.3 Grep language: Go source + Go test fixtures

All grep patterns in ACs target Go source files (`*.go`) and Go test fixtures (`*_test.go`), which are English-only per CLAUDE.md §9. No CJK text appears in the patterns; the prohibition in `feedback_ac_command_vacuous_green_traps.md` (pattern/file language mismatch) does not bind.

## §B. AC Matrix (Given-When-Then)

### AC-001 (REQ-NS-001 — token trio recognition)

**Given** a temp project with a populated `capability-map.md` (001 format) and a single `.go` fixture carrying `@NAV:DEC-AUTH` and `@NAV:SYM:ParseHeader` tokens,
**When** the integration layer runs `moai navigator-sync --project-root <tmp>`,
**Then** the emitted `nav-graph.json` `edges` array contains at least one edge of each `edge_type` ∈ `{dec-edge, spec-edge, sym-edge}`.

**Evidence**: `jq '.edges | map(.edge_type) | unique' nav-graph.json` → `["dec-edge","spec-edge","sym-edge"]` (set equality).

### AC-002 (REQ-NS-002 — binding record 5 fields)

**Given** any binding record emitted by the scanner,
**When** the record is inspected,
**Then** it carries exactly the 5 fields `token_family`, `identifier`, `source_path`, `line_number`, `commit_sha`, with non-empty values.

**Evidence**: a Go unit test on the scanner output struct, asserting each field is non-zero on a populated fixture.

### AC-003 (REQ-NS-003 — `@NAV:DEC` scanner over design docs)

**Given** `.moai/project/tech.md` containing the line `Decision @NAV:DEC-AUTH-STRATEGY: adopt OAuth2`,
**When** the scanner runs,
**Then** it emits a binding record with `token_family="NAV:DEC"`, `identifier="AUTH-STRATEGY"`, `source_path` ending in `tech.md`, `line_number` matching the line, and `commit_sha` = HEAD.

**Evidence**: Go test fixture, asserted by struct equality.

### AC-004 (REQ-NS-004 — `@NAV:SYM` scanner over code + design)

**Given** a `.go` fixture with `// @NAV:SYM:pkg.ParseHeader` on line 5 and `.moai/project/structure.md` with `@NAV:SYM:pkg.WriteAtomic` on line 12,
**When** the scanner runs,
**Then** it emits two binding records, one for each occurrence, with the code-source record's `source_path` ending `.go` and the design-source record's `source_path` ending `.md`.

**Evidence**: Go test fixture.

### AC-005 (REQ-NS-005 — `@MX:SPEC` via mx-scanner, no `internal/mx/` modification)

**Given** the integration layer's binary,
**When** a source-code audit runs,
**Then** `internal/mx/` source files are NOT modified by this SPEC's commits, and `internal/navigator/sync/` CONSUMES `spec_association.NewSpecAssociator(...)` via import.

**Evidence (two-lens)**:
- Lens 1 — `git diff <pre-SPEC-commit>..<post-SPEC-commit> -- internal/mx/` → empty.
- Lens 2 — `grep -rn 'modu-ai/moai-adk/internal/mx' internal/navigator/sync/` → at least one import line.

### AC-006 (REQ-NS-006 — single graph artifact shape)

**Given** a successful join run,
**When** `nav-graph.json` is parsed,
**Then** its top-level keys are exactly `{provenance, nodes, edges}` (no extras, no missing).

**Evidence**: `jq 'keys' nav-graph.json` → `["edges","nodes","provenance"]` (set equality, order-insensitive).

### AC-007 (REQ-NS-007 — node set 3 entity types)

**Given** the emitted graph,
**When** the node `entity_type` values are collected,
**Then** the set of entity_types ⊆ `{decision, spec, symbol}` and ≥ 2 of the 3 are present in a populated fixture.

**Evidence**: `jq '.nodes | map(.entity_type) | unique' nav-graph.json` → subset check.

### AC-008 (REQ-NS-008 — edge set 3 edge types)

**Given** the emitted graph,
**When** the edge `edge_type` values are collected,
**Then** the set of edge_types ⊆ `{dec-edge, spec-edge, sym-edge}` and all 3 are present in a populated fixture (carries AC-001's evidence).

**Evidence**: `jq '.edges | map(.edge_type) | unique' nav-graph.json`.

### AC-009 (REQ-NS-009 — Provenance, byte-identical re-run)

**Given** a temp git repo at a pinned HEAD with populated inputs,
**When** the join runs twice in sequence,
**Then** `cmp run1/nav-graph.json run2/nav-graph.json` exits 0, and the `provenance.captured_at` equals `git log -1 --format=%cI` (not a wall-clock).

**Evidence**: characterization test; `jq '.provenance.captured_at'` == `git -C <tmp> log -1 --format=%cI`.

### AC-010 (REQ-NS-010 — atomic write + barrier test hook)

**Given** the `NAVIGATOR_PRE_RENAME_BARRIER=/tmp/barrier` env var set,
**When** the join runs,
**Then** the binary creates the barrier file, blocks until the test removes it, then renames `nav-graph.json.tmp` → `nav-graph.json`. The final `nav-graph.json` exists and `nav-graph.json.tmp` does not.

**Evidence**: Go concurrency test mirroring `internal/cli/navigator_enrich_test.go`'s barrier fixture (the test removes the barrier after detecting it and asserts the rename completed).

### AC-011 (REQ-NS-011 — fail-open when capability-map absent)

**Given** a temp project with NO `capability-map.md`,
**When** the join runs,
**Then** the exit code is 0, NO `nav-graph.json` is written, and a log line appears in `.moai/logs/navigator-sync.log` containing the substring `capability-map absent`.

**Evidence**: exit-code check + `test ! -f nav-graph.json` + log-substring grep.

### AC-012 (REQ-NS-012 — consumer-only on the 3 chains)

**Given** this SPEC's commit range,
**When** a diff audit runs,
**Then** `internal/navigator/astx/`, `internal/cli/navigator_enrich.go`, `scripts/navigator-audit.sh`, `scripts/navigator-regen.sh` are NOT modified.

**Evidence (two-lens)**:
- Lens 1 — `git diff <range> -- internal/navigator/astx/ internal/cli/navigator_enrich.go scripts/navigator-audit.sh scripts/navigator-regen.sh` → empty.
- Lens 2 — runtime: the join step is invoked AFTER 001/003 in `/moai project`, and each chain still produces its own output as before (characterization test on the chained output).

### AC-013 (REQ-NS-013 — LSEL non-overlap)

**Given** the integration layer's source and its runtime writes,
**When** the grep-pair protocol runs,
**Then** NO LSEL surface (`.moai/lessons-inbox.jsonl`, `.moai/state/lsel/`, `memory/feedback_*.md`, `hns-lsel-*`, `.moai/state/lsel/proposals/`) appears as a read OR write path.

**Evidence (two-lens per §A.1)**:
- Lens 1 — `grep -rn 'lessons-inbox\|state/lsel\|memory/feedback_\|hns-lsel' internal/navigator/sync/` → no matches.
- Lens 2 — runtime temp-dir fixture: the set of paths the binary touched (collected via a filesystem-audit helper) ∩ the LSEL surface set = ∅.

### AC-014 (REQ-NS-014 — write-surface isolation)

**Given** a successful join run in a temp dir,
**When** the set of paths created/modified by the binary is collected,
**Then** the set equals `{<tmp>/.moai/project/navigator/nav-graph.json}` (the `.tmp` transient is renamed away, not left behind).

**Evidence**: filesystem-audit helper in the test, comparing pre-run and post-run path snapshots.

### AC-015 (REQ-NS-015 — 002 audit write surface untouched)

**Given** a temp project with a pre-existing `.moai/project/navigator/audit-report.json`,
**When** the join runs,
**Then** the audit-report.json file content is byte-identical before and after (the join only READs it).

**Evidence**: `cmp` before/after.

### AC-016 (REQ-NS-016 — 003 enrich write surface untouched)

**Given** a temp project with a pre-existing `.moai/project/codemaps/capability-symbols.json`,
**When** the join runs,
**Then** the capability-symbols.json content is byte-identical before and after.

**Evidence**: `cmp` before/after.

### AC-017 (REQ-NS-017 — malformed-token diagnostic)

**Given** a fixture with `@NAV:DEC-` (empty id) and `@NAV:SYM:` (empty symbol) on separate lines,
**When** the scanner runs,
**Then** it emits two diagnostic warnings to `.moai/logs/navigator-sync.log` (one per malformed token), skips both records, and the join completes with exit 0 and a `nav-graph.json` that contains no edge sourced from either malformed line.

**Evidence**: log-substring grep + `jq '.edges | map(select(.source_path == "<fixture>" and .line_number == <bad-line>)) | length'` → 0.

### AC-018 (REQ-NS-018 — template-first documentation)

**Given** the documentation edit for the two new tokens,
**When** `make build` runs and the `internal/template/internal_content_leak_test.go` neutrality guard runs,
**Then** the local `.claude/rules/moai/workflow/nav-tokens.md` exists, matches the template source byte-for-byte, and the neutrality guard is green (no SPEC-ID / REQ-token / internal-date leakage).

**Evidence**:
- `test -f internal/template/templates/.claude/rules/moai/workflow/nav-tokens.md` → 0.
- `diff internal/template/templates/.claude/rules/moai/workflow/nav-tokens.md .claude/rules/moai/workflow/nav-tokens.md` → empty.
- `go test ./internal/template/ -run TestInternalContentLeak` → PASS.

## §C. Quality Gate Criteria (Definition of Done)

- All 18 ACs PASS with observed evidence (no unobserved claims per `verification-claim-integrity.md` §1.1).
- `go test ./internal/navigator/sync/...` green.
- `go test ./internal/cli/...` green (including the new navigator-sync subcommand test).
- `golangci-lint run ./internal/navigator/sync/...` clean.
- `go vet ./internal/navigator/sync/...` clean.
- `make build` green; `internal/template/internal_content_leak_test.go` green.
- `moai spec lint --strict .moai/specs/SPEC-NAVIGATOR-SYNC-001/` → 0 errors.
- Commit subject prefix `feat(SPEC-NAVIGATOR-SYNC-001): ...` per the repo's plan-phase commit-subject convention.

## §D. Edge Cases

- **Empty project (no design docs, no symbols, no capability-map)**: REQ-NS-011 fail-open triggers; no output; exit 0.
- **Capability-map present but capability-symbols.json absent**: spec nodes still surface (from capability-map header parse); symbol nodes degrade to `@NAV:SYM`-only. Graph emits with whatever nodes are available.
- **Audit-report.json malformed**: advisory-only; the join logs a warning and proceeds without audit-derived edges. (Audit edges are not in M0's required edge set — `dec/spec/sym` only.)
- **Massive monorepo (>10k files)**: scanner performance — M0 caps scan at the design-doc root + `.moai/project/` tree (NOT the whole repo). Code-side scan for `@NAV:SYM` is bounded by reading only files that contain the literal `@NAV:SYM:` substring (a pre-filter via `grep -lE '@NAV:SYM:'`).
- **Concurrent runs (two `navigator-sync` invocations)**: atomic-rename + barrier means one wins; the loser's `.tmp` is left behind (not a hazard — `.tmp` is not a recognized artifact).

## §E. Forward-looking checks (not M0 blockers)

- The `nav-graph.json` schema (REQ-NS-006/007/008) MUST be forward-compatible (additive only) because M1 (Detect hook) will consume it.
- The `binding_record` 5-field shape (REQ-NS-002) MUST be forward-compatible because M3 (Fix) will author new bindings.
- The Provenance contract (REQ-NS-009) MUST remain wall-clock-free because all later milestones inherit the byte-identical re-run property.
