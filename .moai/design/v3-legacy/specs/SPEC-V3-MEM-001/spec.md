---
id: SPEC-V3-MEM-001
title: "Memory 2.0 Alignment — 4-type taxonomy, 200/25K cap, freshness, path validation"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 4 Memory 2.0"
module: "internal/core/memory/"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - gm#44
  - gm#45
  - gm#50
  - gm#53
related_theme: "Theme 5 — Memory 2.0 Alignment"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "memory, v3, taxonomy, truncation, freshness, path-validation"
---

# SPEC-V3-MEM-001: Memory 2.0 Alignment

## HISTORY

| Version | Date       | Author | Description                               |
|---------|------------|--------|-------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial v3 draft (Wave 4, Memory bundle)  |

---

## 1. Goal (목적)

Align moai-adk-go's MEMORY.md/memdir contract with Claude Code's authoritative 4-type taxonomy, 200-line / 25KB entrypoint truncation, freshness-warning wrapping, and path-security validation. Today moai's auto-memory uses a MEMORY.md index per project (`~/.claude/projects/{hash}/memory/MEMORY.md`) but enforces none of CC's safety rails: oversize indexes silently overflow context, stale memories propagate as fact without caveat, and path handling does not reject UNC/tilde/NFKC-attack inputs. findings-wave1-query-context.md §6 documents the CC contract in detail (memdir.ts:35-38, 57-103; memoryTypes.ts:14-19; memoryAge.ts:33-53; paths.ts:109-150). This SPEC ports the contract to Go with faithful constants, types, and rejection rules.

### 1.1 배경

findings-wave1-query-context.md §6.5: CC enforces `MAX_ENTRYPOINT_LINES = 200` (`memdir.ts:35`) and `MAX_ENTRYPOINT_BYTES = 25_000` (`memdir.ts:38`, rationale "p97 today; p100 observed 197KB under 200 lines"). Truncation appends a machine-readable warning: `"WARNING: MEMORY.md is {reason}. Only part of it was loaded. Keep index entries to one line under ~200 chars; move detail into topic files."` (memdir.ts:57-103).

findings-wave1-query-context.md §6.6: CC's `MEMORY_TYPES = ['user', 'feedback', 'project', 'reference']` is a fixed 4-tuple (`memoryTypes.ts:14-19`). Each memory file has YAML frontmatter with `name`, `description`, `type`. MEMORY.md itself is an index — one line per pointer entry.

findings-wave1-query-context.md §6.12: CC wraps memories older than 24h in `<system-reminder>…</system-reminder>` with explicit staleness caveat (`memoryAge.ts:33-53`). Rationale (L27-30): "user reports of stale code-state memories being asserted as fact — citation makes the stale claim sound more authoritative, not less."

findings-wave1-query-context.md §6.3: `validateMemoryPath` (`paths.ts:109-150`) rejects non-absolute paths, length < 3, Windows drive-root regex, UNC paths, null byte, tilde-only (`~/`, `~`, `~/..`), NFKC Unicode attacks. Always returns NFC-normalized with one trailing separator.

findings-wave1-moai-current.md §6 confirms moai's current state: `~/.claude/projects/{hash}/memory/MEMORY.md` exists in agent memory directory but has no truncation, no freshness, no type enforcement, no path validator.

### 1.2 Non-Goals

- LLM-based memory relevance top-k selection (SPEC-V3-MEM-002; opt-in, deferred default).
- KAIROS daily-log mode (`{autoMemPath}/logs/YYYY/MM/YYYY-MM-DD.md` + `/dream` distillation) — master-v3 §10 reject (out of v3.0 scope).
- Team memory directory (`teamMemPaths.ts` — CC gated on Statsig + moai does not surface team-shared memory in v3.0).
- CLAUDE.md 40K cap (already enforced per `.claude/rules/moai/development/coding-standards.md`; this SPEC addresses MEMORY.md only).
- Memory content migration — existing MEMORY.md files over 200 lines/25KB are snapshot-truncated on first load, not rewritten on disk in v3.0.
- Cross-project memory sync (remote memory via `CLAUDE_CODE_REMOTE_MEMORY_DIR`).
- Memory file format change — stays `.md` with YAML frontmatter; no JSON/binary alternatives.

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/core/memory/` new package with the following files:
  - `types.go` — `MemoryEntry`, `MemoryType` (enum-ish string), constants.
  - `truncate.go` — line + byte truncation with warning appendix.
  - `freshness.go` — age computation and `<system-reminder>` wrapping.
  - `validate.go` — path security rules port.
  - `taxonomy.go` — 4-type classification + frontmatter parsing.
- Integration point: memory loader invoked by moai agent runtime (SPEC-V3-AGT-001 consumes via `memory: user|project|local` field) and by `moai doctor memory` (sibling concern in SPEC-V3-CLI-001).
- Configuration: new YAML section `.moai/config/sections/memory.yaml` with fields:
  - `memory.truncation.max_lines` (default 200, validated ≥ 50)
  - `memory.truncation.max_bytes` (default 25000, validated ≥ 5000)
  - `memory.freshness.stale_threshold_hours` (default 24, validated ≥ 1)
  - `memory.llm_relevance.*` — schema stubs only (full impl in SPEC-V3-MEM-002).
- Schema registration: `QualityConfig` counterpart — `MemoryConfig` struct with validator/v10 tags (SPEC-V3-SCH-001).
- MEMORY.md frontmatter validation: accept absent frontmatter for backward compat (warning logged), enforce 4-type enum when present.
- Migration-free: existing MEMORY.md content remains on disk. Truncation is at load time; full content snapshot written to `.moai/reports/memory-truncation-{timestamp}.md` when truncation fires.
- Path validator rules ported 1:1 from CC `paths.ts:109-150`.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- LLM-based relevance selection (deferred to SPEC-V3-MEM-002).
- Rewriting existing MEMORY.md files larger than 200 lines / 25KB on disk.
- KAIROS daily-log pipeline or `/dream` distillation.
- Team memory shared directories.
- Memory content editing UI (moai has no REPL).
- Memory backup to cloud.
- Nontrivial memory deduplication beyond what `readFileState` already provides at the CC layer.
- Changing MEMORY.md entrypoint filename (stays `MEMORY.md` for CC compatibility).
- Changing memory base directory (stays `~/.claude/projects/{hash}/memory/` for CC interop).
- Encryption / at-rest protection (memory files remain plaintext Markdown).

---

## 3. Environment (환경)

- 런타임: Go 1.23+, moai-adk-go v3.0.0+.
- Claude Code v2.1.111+ (memdir consumer).
- Platforms: macOS / Linux / Windows (with Windows-specific path rejection rules active everywhere).
- Target directories: `internal/core/memory/` (new), `.moai/config/sections/memory.yaml` (new), `internal/config/schema/memory_schema.go` (new).
- Dependencies: `gopkg.in/yaml.v3` (frontmatter parse — already in go.mod), `github.com/go-playground/validator/v10` (via SPEC-V3-SCH-001).
- Stdlib only for Unicode NFC: `golang.org/x/text/unicode/norm` is acceptable as a new dep ONLY if existing indirect deps (`golang.org/x/text`) already pull it (W1.6 §14.1 — verified below in §4).
- Memory base dir resolution: `os.UserHomeDir()` + `~/.claude/projects/{project-hash}/memory/`. `CLAUDE_CODE_REMOTE_MEMORY_DIR` env overrides when set (honor CC compat; findings-wave1-query-context.md §6.1 L85-90).
- Project hash: moai reuses CC's sanitization scheme (`sanitizePath(getAutoMemBase())`; §6.2 L203-205, `findCanonicalGitRoot()` so worktrees share memory).

---

## 4. Assumptions (가정)

- A-MEM-001: `golang.org/x/text/unicode/norm` is already transitively available via existing dependencies (W1.6 §14.1 lists it as indirect). If not, it is added as a single new indirect dep — no direct top-level addition.
- A-MEM-002: Existing MEMORY.md files conform to one-line-per-entry format compatible with CC's MEMORY.md schema. Non-conforming files degrade gracefully (no crash; warning only).
- A-MEM-003: Users understand that truncation is display-only in v3.0; full content remains on disk and is preserved verbatim.
- A-MEM-004: `.moai/reports/memory-truncation-{timestamp}.md` snapshot files are tolerated by git (not in `.gitignore`); users may choose to gitignore them locally.
- A-MEM-005: All NFC normalization happens at path validation boundary; stored MEMORY.md content is not re-normalized on disk.
- A-MEM-006: Agent-scoped memory (SPEC-V3-AGT-001 `memory: user|project|local`) reuses this package's `validateMemoryPath` helper.
- A-MEM-007: Memory freshness is computed from file `mtime` (as CC does); no custom timestamp written into frontmatter.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MEM-001-001 (Ubiquitous) — constants**
The `internal/core/memory` package **shall** define constants `MaxEntrypointLines = 200`, `MaxEntrypointBytes = 25000`, `StaleThresholdHours = 24` (mirroring CC `memdir.ts:35-38` and `memoryAge.ts:33-53`).

**REQ-MEM-001-002 (Ubiquitous) — 4-type taxonomy**
The package **shall** define `MemoryType` as the fixed set `{user, feedback, project, reference}`. Unknown or missing type in frontmatter **shall** degrade to empty string `""` without erroring out (graceful degradation per CC `memoryTypes.ts parseMemoryType`).

**REQ-MEM-001-003 (Ubiquitous) — MemoryEntry struct**
The package **shall** expose `MemoryEntry` with fields `Path string`, `Type MemoryType`, `Name string`, `Description string`, `LastValidated time.Time`, `Content string`, `SizeBytes int`, `LineCount int`, `AgeHours float64`.

**REQ-MEM-001-004 (Ubiquitous) — config schema**
The file `internal/config/schema/memory_schema.go` **shall** declare `MemoryConfig` with validator/v10 tags enforcing:
- `truncation.max_lines` ≥ 50 (default 200)
- `truncation.max_bytes` ≥ 5000 (default 25000)
- `freshness.stale_threshold_hours` ≥ 1 (default 24)

### 5.2 Event-Driven Requirements

**REQ-MEM-001-010 (Event-Driven) — truncation**
**When** `Truncate(content string, maxLines int, maxBytes int)` is called with content exceeding either limit, the function **shall** return `(truncated string, mark string)` where `truncated` is cut at the last full line before the limit and `mark` contains a warning matching the format: `"WARNING: MEMORY.md is {reason}. Only part of it was loaded. Keep index entries to one line under ~200 chars; move detail into topic files."` (findings-wave1-query-context.md §6.5).

**REQ-MEM-001-011 (Event-Driven) — truncation order**
**When** both line and byte limits are exceeded, the function **shall** apply line truncation first (natural boundary at 200 lines) then byte truncation at last newline before maxBytes (CC `memdir.ts:57-103`).

**REQ-MEM-001-012 (Event-Driven) — freshness wrap**
**When** `FreshnessPreamble(entry MemoryEntry)` is called for an entry with `AgeHours > StaleThresholdHours`, the function **shall** return a string wrapped in `<system-reminder>…</system-reminder>` tags with text: `"Memory file {path} last modified {human-age}; content may be stale. Memories are point-in-time observations, not live state — claims about code behavior or file:line citations may be outdated. Verify against current code before asserting as fact."` (faithfully porting CC `memoryAge.ts:33-53`).

**REQ-MEM-001-013 (Event-Driven) — fresh case**
**When** `FreshnessPreamble` is called for an entry with `AgeHours ≤ StaleThresholdHours`, the function **shall** return an empty string.

**REQ-MEM-001-014 (Event-Driven) — snapshot on truncation**
**When** entrypoint truncation fires during memory load, the loader **shall** write the full pre-truncation content to `.moai/reports/memory-truncation-{ISO-8601-timestamp}.md` as a side effect and emit one `systemMessage` notifying the user of the snapshot path. Subsequent loads in the same session **shall NOT** re-emit the notification (dedup via in-memory set keyed by file path).

**REQ-MEM-001-015 (Event-Driven) — frontmatter parse**
**When** the memory loader reads a `.md` file under the memory directory, the loader **shall** attempt YAML frontmatter parse; on parse error it **shall** log a warning, treat the entire file as body (no frontmatter), and continue (never block the load).

### 5.3 State-Driven Requirements

**REQ-MEM-001-020 (State-Driven) — path validation**
**While** the function `ValidateMemoryPath(p string) (string, error)` is invoked, it **shall** reject inputs matching any of the following and return `ErrInvalidMemoryPath` with a discriminator tag:
- Non-absolute path (`non-absolute`)
- Length < 3 (`too-short`)
- Windows drive-root pattern `^[A-Za-z]:\\?$` (`windows-drive-root`)
- UNC path prefix `\\` on Windows (`unc-path`)
- Contains null byte `\0` (`null-byte`)
- Tilde-only (`~`, `~/`, `~/..`, `~/`) before expansion (`tilde-only`)
- NFKC-normalized form differs from NFC-normalized form in a way that indicates Unicode homoglyph attack (`nfkc-attack`)
On success it **shall** return the NFC-normalized path with exactly one trailing separator.

**REQ-MEM-001-021 (State-Driven) — config drives constants**
**While** `memory.yaml` is loaded and validated, runtime calls to `Truncate` and `FreshnessPreamble` **shall** use the config-supplied values rather than the compiled-in defaults.

### 5.4 Optional Features

**REQ-MEM-001-030 (Optional) — agent-scoped memory path**
**Where** an agent declares `memory: user|project|local` in its frontmatter (SPEC-V3-AGT-001), the resolver **shall** compose the agent-scoped directory using this package's `ValidateMemoryPath` at every join step.

**REQ-MEM-001-031 (Optional) — doctor output**
**Where** `moai doctor memory` is invoked (sibling SPEC-V3-CLI-001), this package **shall** expose a `Report(baseDir string) ([]MemoryEntry, []ValidationIssue, error)` helper that the CLI consumes.

### 5.5 Unwanted Behavior

**REQ-MEM-001-040 (Unwanted) — no silent truncation**
If MEMORY.md is truncated, then the loader **shall NOT** silently drop content. It **shall** (a) append the warning appendix, (b) snapshot full content to `.moai/reports/memory-truncation-{timestamp}.md`, and (c) emit a one-time-per-file notification.

**REQ-MEM-001-041 (Unwanted) — no path escape**
If `ValidateMemoryPath` is bypassed and a path escape is attempted (via manually-constructed join), the loader **shall** refuse to read files whose canonical path falls outside the configured memory base directory.

**REQ-MEM-001-042 (Unwanted) — no unknown-type coercion**
If a memory file's frontmatter declares `type: {unknown}`, then the loader **shall NOT** coerce the value to one of the 4 known types. It **shall** classify the entry as `type: ""` and log the unknown value once.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-MEM-001-01**: Loading a 500-line MEMORY.md returns truncated content ending at line 200 with the exact warning appendix; full content written to `.moai/reports/memory-truncation-{ts}.md`. (maps REQ-MEM-001-010, -014)
- **AC-MEM-001-02**: Loading a 50KB MEMORY.md (under 200 lines but over 25KB) truncates at last newline before byte 25000; warning appendix references "{reason}=byte-limit". (maps REQ-MEM-001-011)
- **AC-MEM-001-03**: A memory file with `mtime` 25 hours ago, when wrapped via `FreshnessPreamble`, is returned with `<system-reminder>…</system-reminder>` containing the literal staleness caveat string. A memory file modified 2 hours ago returns empty string. (maps REQ-MEM-001-012, -013)
- **AC-MEM-001-04**: `ValidateMemoryPath` rejects `../etc/passwd`, `C:\`, `\\server\share`, `"foo\0bar"`, `~`, `~/..`, and a path with NFKC-homoglyph 'a' (U+0430 CYRILLIC SMALL LETTER A) with the correct discriminator tag. (maps REQ-MEM-001-020)
- **AC-MEM-001-05**: `ValidateMemoryPath` accepts `/Users/x/.claude/projects/hash/memory/` on macOS/Linux and `C:\Users\x\.claude\projects\hash\memory\` on Windows, returning NFC-normalized with exactly one trailing separator. (maps REQ-MEM-001-020)
- **AC-MEM-001-06**: A memory file with `type: project` in frontmatter parses into `MemoryEntry.Type = MemoryTypeProject`. A file with `type: bogus` parses into `MemoryEntry.Type = ""` with one-shot warning log. (maps REQ-MEM-001-002, -042, -015)
- **AC-MEM-001-07**: `memory.yaml` with `truncation.max_lines: 30` fails validation (min 50); with `truncation.max_lines: 100` passes and runtime truncate uses 100 as the cap. (maps REQ-MEM-001-004, -021)
- **AC-MEM-001-08**: A second load of the same oversize MEMORY.md in the same session does NOT emit the truncation notification again (dedup). (maps REQ-MEM-001-014)
- **AC-MEM-001-09**: `go test ./internal/core/memory/...` passes with ≥ 90% coverage (memory is safety-critical).

---

## 7. Constraints (제약)

- **[HARD] CC-compat constants**: `MaxEntrypointLines=200`, `MaxEntrypointBytes=25000`, `StaleThresholdHours=24` are compile-time defaults; config may override within validated bounds but cannot disable enforcement.
- **[HARD] Path validator rules ported 1:1**: every rejection in CC `paths.ts:109-150` has a corresponding rejection here.
- **[HARD] No content rewrite on disk**: truncation is at load/view time; MEMORY.md file on disk is never modified by this SPEC.
- **[HARD] Snapshot retention**: `.moai/reports/memory-truncation-*.md` files are retained until the user deletes them; moai does not auto-clean.
- **[HARD] 4-type taxonomy fixed**: no extension, no custom types.
- **[HARD] UTF-8 only**: memory files must be UTF-8; non-UTF-8 is a warning with raw-byte fallback.
- **[HARD] 9-direct-dep budget preserved**: no new direct Go deps. `golang.org/x/text/unicode/norm` may enter as indirect dep only.
- **[HARD] Deterministic output**: the warning appendix text is byte-exact per REQ-MEM-001-010 for test reproducibility.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 확률 | 영향 | 완화 |
|--------|------|------|------|
| Users with 1000-line MEMORY.md surprised by truncation display | High | Low | Snapshot file preserves full content; one-time notification explains; migration guide calls out the 200-line best practice |
| NFKC attack detection false positive on legitimate Unicode input (e.g., CJK composition forms) | Medium | Medium | Only reject when NFKC differs from NFC AND the path contains suspicious characters (confusable script mixing); add allowlist escape via `MOAI_MEMORY_UNICODE_STRICT=0` |
| Freshness wrapper adds noise to stable/curated memory files | Medium | Low | 24h threshold matches CC default; users can extend via `memory.yaml freshness.stale_threshold_hours` up to compile-time MAX (e.g. 720h) |
| `.moai/reports/` directory balloons with truncation snapshots | Low | Low | Dedup logic per REQ-MEM-001-014 prevents repeated snapshots per file per session; future `moai doctor memory --clean` deferred |
| Windows path rejection breaks legitimate absolute paths (e.g., paths with `\\?\` long-path prefix) | Low | Medium | Explicit allowance for `\\?\` prefix (canonical Windows long path); regression test `TestWindowsLongPath` |
| Oversized memory files during a single session cause repeated `.moai/reports/` writes | Low | Low | REQ-MEM-001-014 dedup via in-memory set keyed by absolute path |
| Agent-scoped memory paths bypass validation via relative segments | Low | High | REQ-MEM-001-041 requires canonical-path containment check; enforced at every join step |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** — `MemoryConfig` schema registration and validator/v10 tags depend on the shared validator plumbing.

### 9.2 Blocks

- **SPEC-V3-MEM-002** — LLM-based relevance selection reuses `MemoryEntry` + `MemoryScan` helpers and must compose atop freshness/truncation.
- **SPEC-V3-AGT-001** — agent-scoped memory (`memory: user|project|local`) reuses `ValidateMemoryPath` for path joining.

### 9.3 Related

- **SPEC-V3-CLI-001** — `moai doctor memory` (if added) consumes `Report` helper.
- **SPEC-V3-MIG-001** — no migration required; existing files are untouched on disk.

---

## 10. Traceability (추적성)

- Theme: master-v3 §3.5 (Theme 5 — Memory 2.0 Alignment).
- Gap rows: gm#44 (MEMORY.md unbounded, Critical), gm#45 (no truncation), gm#50 (staleness missing, High), gm#53 (path validation).
- Wave 1 sources:
  - findings-wave1-query-context.md §6.1 (base dir resolution)
  - findings-wave1-query-context.md §6.3 (path validator rules `paths.ts:109-150`)
  - findings-wave1-query-context.md §6.5 (truncation `memdir.ts:57-103`)
  - findings-wave1-query-context.md §6.6 (4-type taxonomy `memoryTypes.ts:14-19`)
  - findings-wave1-query-context.md §6.12 (freshness `memoryAge.ts:33-53`)
  - findings-wave1-moai-current.md §6 (current moai memory state)
- BC-ID: none (additive — existing files load with warnings only).
- REQ 총 개수: 15 (Ubiquitous 4, Event-Driven 6, State-Driven 2, Optional 2, Unwanted 3 discrete).
- 예상 AC 개수: 9.
- 코드 구현 예상 경로:
  - `internal/core/memory/types.go`, `truncate.go`, `freshness.go`, `validate.go`, `taxonomy.go`
  - `internal/config/schema/memory_schema.go`
  - `internal/template/templates/.moai/config/sections/memory.yaml` (new)
  - Test files: `internal/core/memory/*_test.go` (≥ 90% coverage per constraint).

---

End of SPEC.
