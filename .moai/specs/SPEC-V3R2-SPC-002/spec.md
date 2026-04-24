---
id: SPEC-V3R2-SPC-002
title: "@MX TAG v2 with hook JSON integration and sidecar index"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: Wave 4 SPEC Writer
priority: P1 High
phase: "v3.0.0 — Phase 7 — Extension"
module: "internal/mx/, .claude/rules/moai/workflow/mx-tag-protocol.md"
dependencies:
  - SPEC-V3R2-CON-001
  - SPEC-V3R2-RT-001
related_gap: []
related_problem: []
related_pattern:
  - T-1
  - T-5
  - X-1
related_principle:
  - P1
  - P8
related_theme: "Layer 2: SPEC & TAG"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "v3r2, mx-tag, hook, json-sidecar, anchor, post-tool-use"
---

# SPEC-V3R2-SPC-002: @MX TAG v2 with hook JSON integration and sidecar index

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-23 | Wave 4 SPEC Writer | Initial draft from master-v3 Layer 2, pattern-library T-1/T-5 |

---

## 1. Goal (목적)

Preserve the @MX TAG inline protocol (NOTE / WARN / ANCHOR / TODO / LEGACY) exactly as defined in `.claude/rules/moai/workflow/mx-tag-protocol.md` — a FROZEN invariant per master-v3 §1.3 — while layering two additive capabilities on top:

1. **Machine-readable JSON sidecar index** at `.moai/state/mx-index.json` that materializes every @MX TAG in the codebase as a structured record (Kind, File, Line, Body, Reason, AnchorID, CreatedBy, LastSeenAt). This enables tool consumption (SPEC-V3R2-SPC-004 resolver, evaluator-active, codemaps generators) without re-scanning source files on every query.

2. **Hook JSON integration** via the `PostToolUse` JSON protocol (SPEC-V3R2-RT-001) such that a post-tool-use handler can emit `{"additionalContext": "@MX:WARN at line 42: goroutine leak (no Done() signal)", "hookSpecificOutput": {"mxTags": [...]}}` to inject newly-detected @MX findings back into the model-turn context **and** append them to the sidecar index atomically.

The inline source-code convention stays FROZEN: `// @MX:NOTE some text` and its 5 sibling kinds remain the canonical on-disk form. The sidecar is a **view** built from scanning the canonical form; it is never the source of truth.

## 2. Scope (범위)

### 2.1 In Scope

- Go types in `internal/mx/tag.go`: `TagKind` enum (MXNote, MXWarn, MXAnchor, MXTodo, MXLegacy), `Tag` struct with Kind/File/Line/Body/Reason/AnchorID/CreatedBy/LastSeenAt.
- Cross-language `TagScanner` that walks source files across the 16 supported languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift) and extracts @MX TAGs from their line-comment syntaxes.
- Sidecar index file `.moai/state/mx-index.json` with atomic write semantics.
- `/moai mx` subcommand extension to produce/refresh the sidecar; `/moai mx --index-only` for fast index rebuild.
- PostToolUse hook handler integration: when a Write/Edit tool mutates a source file, the handler re-scans the file and emits a JSON `additionalContext` with new/changed tags plus a `hookSpecificOutput.mxTags` structured array.
- AnchorID resolver API at `internal/mx/resolver.go` — used by SPEC-V3R2-SPC-004.
- MX TAG sub-line parser for `@MX:WARN` `@MX:REASON` requirement per mx-tag-protocol.md.
- Retention: `LastSeenAt` per tag enables stale-tag detection (tag present in index but not found on last scan → candidate for archive after 7 days).

### 2.2 Out of Scope

- Inline @MX TAG syntax changes (FROZEN — mx-tag-protocol.md verbatim).
- Acceptance-criteria tree shape (→ SPEC-V3R2-SPC-001).
- MX anchor fan-in resolver query surface (→ SPEC-V3R2-SPC-004).
- Hook protocol spec itself (→ SPEC-V3R2-RT-001).
- MX TAG generation by agents (continues per mx-tag-protocol.md; the agent decides to add @MX, this SPEC persists the result).
- Evaluator scoring based on MX TAGs (→ SPEC-V3R2-HRN-003 or design-constitution).

## 3. Environment (환경)

Current moai state (v2.13.2):

- `.claude/rules/moai/workflow/mx-tag-protocol.md` defines the inline protocol: line-comment convention `// @MX:KIND text`, the 5 kinds, the sub-line requirement for `@MX:WARN` → `@MX:REASON`, autonomous agent add/update/remove without human approval.
- `/moai mx` command exists but is a thin wrapper that delegates to agent-side scanning.
- No sidecar index today; every consumer (codemaps, resolver, reports) re-scans the codebase.
- Hook protocol is exit-code-only today (SPEC-V3R2-RT-001 fixes this); PostToolUse handlers cannot currently inject `additionalContext` or structured outputs.
- 16-language neutrality: each language has its own comment syntax (`#` for Python/Ruby/Elixir, `//` for Go/Rust/Java/JS/TS/C#/C++/Kotlin/Scala/Swift/Dart, `--` for R-TeX, etc.). Cross-language scanner needs per-language comment-prefix detection.

References: master-v3 §4 Layer 2 @MX Go type sketch; design-principles.md §P8 Hook JSON Protocol; pattern-library.md §T-5 Hook JSON dual protocol.

## 4. Assumptions (가정)

- @MX inline syntax stays FROZEN per mx-tag-protocol.md; this SPEC is strictly additive (sidecar + hook integration).
- SPEC-V3R2-RT-001 ships before SPC-002 is activated; this SPEC depends on the JSON hook protocol to inject tags via PostToolUse.
- The 16-language comment-prefix mapping can be tabulated in `internal/mx/comment_prefixes.go` as a static lookup (line-comment symbols for each language).
- Sidecar index is rebuildable from scratch (`/moai mx --full`) at any time; partial updates are optimizations, not correctness requirements.
- Atomic JSON write uses temp-file + rename pattern (standard Go idiom).
- @MX:ANCHOR IDs are stable across edits (consumers may reference them); the scanner preserves AnchorID even across line-number shifts.

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-SPC-002-001: The system SHALL preserve the inline @MX TAG syntax defined in `.claude/rules/moai/workflow/mx-tag-protocol.md` verbatim; no changes to on-disk form.
- REQ-SPC-002-002: The system SHALL define Go type `internal/mx.Tag` with fields Kind (TagKind enum), File (string), Line (int), Body (string), Reason (string; MXWarn only), AnchorID (string), CreatedBy (string; agent name or "human"), LastSeenAt (time.Time).
- REQ-SPC-002-003: The system SHALL maintain a sidecar index at `.moai/state/mx-index.json` that lists every @MX TAG currently present in source files under the project root.
- REQ-SPC-002-004: The sidecar index SHALL be atomically written (temp-file + rename) to prevent partial reads.
- REQ-SPC-002-005: The TagScanner SHALL support all 16 supported languages via a static comment-prefix lookup table.
- REQ-SPC-002-006: Every `@MX:WARN` tag SHALL have a sibling `@MX:REASON` sub-line; absence SHALL be flagged as a scanner warning with the offending file:line.
- REQ-SPC-002-007: AnchorIDs SHALL be unique within the project; duplicate-anchor detection SHALL emit a scanner error.
- REQ-SPC-002-008: The sidecar index SHALL include a top-level field `schema_version: 2` to distinguish from any future v3 sidecar changes.

### 5.2 Event-driven

- REQ-SPC-002-010: WHEN a Write or Edit tool mutates a source file and the PostToolUse hook handler fires, the handler SHALL re-scan the mutated file and detect newly-added, changed, or removed @MX TAGs.
- REQ-SPC-002-011: WHEN the PostToolUse handler detects changes, it SHALL emit a HookResponse JSON with `additionalContext` (human-readable summary) and `hookSpecificOutput.mxTags` (structured array of Tag records).
- REQ-SPC-002-012: WHEN the PostToolUse handler returns tag updates, the sidecar index SHALL be atomically updated to reflect the new state of the mutated file (only that file's entries change; other files unchanged).
- REQ-SPC-002-013: WHEN `/moai mx --full` is invoked, the system SHALL rescan the entire project and rewrite the sidecar index from scratch.
- REQ-SPC-002-014: WHEN a tag is no longer found on its last seen file/line during a full scan, its `LastSeenAt` SHALL NOT be updated; stale tags are preserved for up to 7 days before archival.

### 5.3 State-driven

- REQ-SPC-002-020: WHILE a tag's `LastSeenAt` is older than 7 days AND the tag is not found in the current scan, the tag SHALL be removed from the sidecar and archived to `.moai/state/mx-archive.json`.
- REQ-SPC-002-021: WHILE two tags claim the same AnchorID, the scanner SHALL emit `DuplicateAnchorID` error naming both file:line pairs and refuse to write the index.
- REQ-SPC-002-022: WHILE the sidecar index file is corrupt or unparseable, `/moai mx` SHALL treat it as empty and emit a repair suggestion; the next full scan rebuilds it.

### 5.4 Optional

- REQ-SPC-002-030: WHERE a source file matches a `.gitignore`-style pattern in `.moai/config/sections/mx.yaml` `ignore:` list, the scanner SHALL skip it.
- REQ-SPC-002-031: WHERE `MOAI_MX_HOOK_SILENT=1` is set, the PostToolUse handler SHALL update the sidecar but NOT emit `additionalContext` into the model turn (useful for CI).
- REQ-SPC-002-032: WHERE `/moai mx --json` is invoked, the system SHALL print the current sidecar to stdout; this is the canonical read API for external tools.

### 5.5 Complex

- REQ-SPC-002-040: WHILE scanning a file AND a line contains both `@MX:WARN` and no sibling `@MX:REASON` within the next 3 lines, THEN the scanner SHALL emit `MissingReasonForWarn` warning (not error — reason is authored progressively).
- REQ-SPC-002-041: IF the PostToolUse handler's JSON output includes `mxTags` but `hookSpecificOutput.hookEventName` does not match "PostToolUse", THEN the hook protocol SHALL reject the output with `HookSpecificOutputMismatch` error (inherits from SPEC-V3R2-RT-001 contract).
- REQ-SPC-002-042: WHILE a Tag's Kind is MXAnchor AND the fan_in (callers referencing the AnchorID) is less than 3, THEN `/moai mx --anchor-audit` SHALL flag it as a low-value anchor candidate (not removed automatically; reviewer decides).

## 6. Acceptance Criteria

- AC-SPC-002-01: Given a Go source file with `// @MX:NOTE explains why handler forks`, When TagScanner scans it, Then a Tag{Kind: MXNote, File: "...", Line: N, Body: "explains why handler forks"} is produced. (maps REQ-SPC-002-001, REQ-SPC-002-002, REQ-SPC-002-005)
- AC-SPC-002-02: Given the project has 42 @MX tags across 15 files, When `/moai mx --full` runs, Then `.moai/state/mx-index.json` contains exactly 42 entries with `schema_version: 2`. (maps REQ-SPC-002-003, REQ-SPC-002-008, REQ-SPC-002-013)
- AC-SPC-002-03: Given the sidecar write is interrupted mid-way (simulated SIGKILL), When the process restarts and reads the sidecar, Then it is either empty or fully valid (never partial). (maps REQ-SPC-002-004)
- AC-SPC-002-04: Given a Python file adds `# @MX:WARN missing timeout on requests.get`, When the PostToolUse handler fires post-Edit, Then the HookResponse contains `additionalContext` referencing that WARN and `hookSpecificOutput.mxTags` has the new Tag entry. (maps REQ-SPC-002-010, REQ-SPC-002-011)
- AC-SPC-002-05: Given a file with `@MX:WARN` but no sibling `@MX:REASON` within 3 lines, When the scanner runs, Then it emits `MissingReasonForWarn` warning naming the file:line. (maps REQ-SPC-002-006, REQ-SPC-002-040)
- AC-SPC-002-06: Given two files each containing `@MX:ANCHOR auth-handler-v1`, When the scanner runs, Then it emits `DuplicateAnchorID` and refuses to write the index. (maps REQ-SPC-002-007, REQ-SPC-002-021)
- AC-SPC-002-07: Given a tag present in the previous index but missing from the current scan, When `LastSeenAt` is still within 7 days, Then the tag remains in the sidecar with its original LastSeenAt. (maps REQ-SPC-002-014)
- AC-SPC-002-08: Given a tag not seen for 8 days, When the next scan runs, Then it is removed from the sidecar and appended to `.moai/state/mx-archive.json`. (maps REQ-SPC-002-020)
- AC-SPC-002-09: Given the sidecar file is corrupt JSON, When `/moai mx` reads it, Then it logs a repair suggestion and proceeds with empty state; the next `--full` rebuilds the index. (maps REQ-SPC-002-022)
- AC-SPC-002-10: Given a mx.yaml `ignore: ["vendor/", "dist/"]`, When the scanner runs, Then no tags from `vendor/` or `dist/` appear in the index. (maps REQ-SPC-002-030)
- AC-SPC-002-11: Given `MOAI_MX_HOOK_SILENT=1` is set, When PostToolUse detects a new tag, Then sidecar is updated but `additionalContext` is empty. (maps REQ-SPC-002-031)
- AC-SPC-002-12: Given `/moai mx --json` is invoked, When it runs, Then it prints the current sidecar content as-is to stdout. (maps REQ-SPC-002-032)
- AC-SPC-002-13: Given a PostToolUse response with `hookSpecificOutput.hookEventName: "PreToolUse"` but `mxTags` present, When the hook protocol validator runs, Then `HookSpecificOutputMismatch` error is returned. (maps REQ-SPC-002-041)
- AC-SPC-002-14: Given an MXAnchor tag with fan_in 1, When `/moai mx --anchor-audit` runs, Then it is flagged as a low-value anchor candidate in the report. (maps REQ-SPC-002-042)
- AC-SPC-002-15: Given all 16 supported languages each have at least one @MX tag in a test fixture, When the scanner runs, Then tags from all 16 languages appear in the sidecar. (maps REQ-SPC-002-005)

## 7. Constraints (제약)

- Inline @MX TAG syntax is FROZEN per CON-001; no field additions to the inline form.
- Sidecar index is a view — never the source of truth; loss of sidecar is recoverable via `/moai mx --full`.
- 16-language neutrality: every language has an entry in the comment-prefix table; adding a 17th language requires a new language rule per SPEC-V3R2-WF-005.
- Performance: full-scan budget 2 seconds for a 10,000-file codebase; incremental PostToolUse update budget 100ms per file.
- Sidecar file size budget: <5MB for 10,000 tags (≈500 bytes per tag entry).
- Atomic writes: no partial reads even under concurrent access.
- No new third-party dependency; Go stdlib `encoding/json` sufficient.

## 8. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Sidecar drift from source truth | MEDIUM | MEDIUM | REQ-SPC-002-013 full-scan rebuilder; CI check in-repo runs `/moai mx --verify` |
| Cross-language comment parsing misses edge cases (nested comments, string literals) | MEDIUM | MEDIUM | Test fixture per language; escape-aware scanner; opt into language rule per SPEC-V3R2-WF-005 refinements |
| PostToolUse JSON injection balloon model-turn tokens | LOW | MEDIUM | `additionalContext` budget cap per mx.yaml `hook.max_additional_context_bytes` |
| Stale tags linger past 7-day TTL | LOW | LOW | Archive sweep is additive; reviewer can inspect mx-archive.json anytime |
| Duplicate AnchorID blocks index rebuild | MEDIUM | MEDIUM | Scanner names both sites; human fixes before next scan |
| Binary size grows from 16-language parsers | LOW | LOW | Prefix-table approach; no per-language AST parsing |

## 9. Dependencies

### 9.1 Blocked by

- SPEC-V3R2-CON-001 (ensures @MX TAG protocol is recorded as FROZEN in zone registry).
- SPEC-V3R2-RT-001 (JSON hook protocol must be available before PostToolUse can emit `hookSpecificOutput.mxTags`).

### 9.2 Blocks

- SPEC-V3R2-SPC-004 (@MX anchor resolver consumes the sidecar index).
- SPEC-V3R2-HRN-003 (evaluator-active may score against MX tag density per harness rubric).

### 9.3 Related

- `.claude/rules/moai/workflow/mx-tag-protocol.md` — inline syntax FROZEN source of truth.
- `/moai mx` subcommand (O-6 Agentless pipeline per master-v3 §7.4).
- design-principles.md §P8 (Hook JSON Protocol); pattern-library.md §T-5.

## 10. Traceability

- Theme: Layer 2 SPEC & TAG (master-v3 §4).
- Principles: P1 SPEC as Constitutional Contract (@MX is the inline analog of SPEC-level contract); P8 Hook Output = JSON Protocol (PostToolUse integration).
- Patterns: T-1 Agent-Computer Interface (sidecar is an ACI-shaped machine-readable response per T-1 priority 1); T-5 Hook JSON-OR-ExitCode Dual Protocol (PostToolUse mxTags emission); X-1 Markdown + YAML Frontmatter (sidecar JSON is a sibling convention).
- Wave 1 sources: R1 §11 SWE-agent ACI, R3 §2 Dec 5 Hook JSON.
- Wave 2 sources: design-principles.md §P8, pattern-library.md §T-5 priority 2.
