---
id: SPEC-V3-MEM-002
title: "Memory Relevance Selection — LLM-based top-k (opt-in)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P3 Low
phase: "v3.0.0 — Phase 6a Tier 2 Strategic Differentiators"
module: "internal/core/memory/relevance/"
dependencies:
  - SPEC-V3-MEM-001
related_gap:
  - gm#48
  - gm#49
  - gm#52
related_theme: "Theme 5 — Memory 2.0 Alignment (opt-in extension)"
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "memory, relevance, llm, haiku, sonnet, opt-in, v3"
---

# SPEC-V3-MEM-002: Memory Relevance Selection

## HISTORY

| Version | Date       | Author | Description                                        |
|---------|------------|--------|----------------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial v3 draft (Wave 4, Memory relevance bundle) |

---

## 1. Goal (목적)

Add opt-in LLM-based relevance selection for memory retrieval that matches Claude Code's `findRelevantMemories` flow (findings-wave1-query-context.md §6.10, `memdir/findRelevantMemories.ts`). When enabled, the system runs a Haiku-class side query per conversation turn to select the top-k most relevant memory entries from the scanned index, reducing context noise at the cost of a small per-turn API call. v3.0 ships this feature **default off** (master-v3 §9 question #3 recommended default); telemetry-driven evaluation in v3.2 may re-baseline the default.

### 1.1 배경

findings-wave1-query-context.md §6.10 documents CC's `findRelevantMemories(query, memoryDir, signal, recentTools, alreadySurfaced)`:
1. `scanMemoryFiles(memoryDir, signal)` returns all `.md` headers except MEMORY.md.
2. Filters out `alreadySurfaced` paths.
3. Calls `selectRelevantMemories()` — a **Sonnet sideQuery** with system prompt: "You are selecting memories that will be useful to Claude Code as it processes a user's query..." (`findRelevantMemories.ts:18-24`). Returns up to 5 filenames via JSON schema `{selected_memories: string[]}`. Max 256 tokens.
4. Filters `recentTools` — excludes tools Claude is already exercising.
5. Returns `RelevantMemory[] = {path, mtimeMs}[]`.

findings-wave1-query-context.md §6.11 `memdir/memoryScan.ts`: `MAX_MEMORY_FILES = 200`, `FRONTMATTER_MAX_LINES = 30`. Single-pass read with mtime (halves syscalls vs stat-sort-read).

findings-wave1-query-context.md §6.14: CC kicks off prefetch concurrently with model streaming (`query.ts:301-304`); the ~1s Haiku/Sonnet selector call is hidden behind the 5-30s model streaming window.

master-v3 §3.5 design approach #5: "LLM-based relevance selection (T2-MEM-02, opt-in): `memory.yaml.llm_relevance.enabled: false` default. When enabled, runs a Haiku-class side query with memory candidates and picks top-5 relevant (configurable via `top_k`). Cost: ~\$0.01-0.03 per turn; telemetry-driven evaluation for default-on in v3.2."

Default model per master-v3: Haiku (cost-conscious). CC uses Sonnet; moai chooses Haiku for v3.0 to keep cost predictable (gap-matrix #52 cost budget concern).

### 1.2 Non-Goals

- Default-on behavior (master-v3 §9 question #3 — opt-in in v3.0).
- Per-memory fine-grained scoring (this SPEC returns top-k paths; no per-entry score exposed).
- Query embedding / vector similarity (not matching CC's approach; deferred).
- Multi-turn relevance learning (no feedback loop in v3.0).
- Team-memory relevance (teamMemPaths.ts gated on Statsig; out of v3.0 scope per SPEC-V3-MEM-001 non-goals).
- `KAIROS` daily-log distillation integration.
- Relevance caching across conversation turns beyond the current turn (per-turn cache only).
- Cost budget enforcement beyond config `max_usd_per_session` advisory.

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/core/memory/relevance/` new sub-package with:
  - `selector.go` — LLM relevance selector with retryable API invocation.
  - `scanner.go` — lightweight `MemoryScan` reading header portion only (frontmatter + first 30 lines per CC `FRONTMATTER_MAX_LINES = 30`).
  - `cache.go` — per-conversation-turn decision cache (memoized by turn ID + candidate-set hash).
  - `types.go` — `RelevantMemory`, `SelectorConfig`, `SelectorResult`.
- Configuration extension in `memory.yaml` (schema defined by SPEC-V3-MEM-001 but `llm_relevance` block finalized here):
  - `memory.llm_relevance.enabled: bool` (default `false`)
  - `memory.llm_relevance.model: string` (default `"haiku"`, enum `haiku|sonnet|inherit`)
  - `memory.llm_relevance.top_k: int` (default `5`, validated 1–20)
  - `memory.llm_relevance.timeout_seconds: int` (default `5`, validated 1–30)
  - `memory.llm_relevance.max_candidates: int` (default `200`, validated 10–500, matches CC `MAX_MEMORY_FILES`)
  - `memory.llm_relevance.max_output_tokens: int` (default `256`, matches CC JSON schema output cap)
  - `memory.llm_relevance.max_usd_per_session: float` (default `0.50`, advisory-only; no enforcement in v3.0)
  - `memory.llm_relevance.cache.scope: string` (default `"turn"`, enum `"turn"|"off"`)
- System prompt and JSON schema ported from CC `findRelevantMemories.ts` verbatim (`selected_memories: string[]` with max-items = `top_k`).
- Integration point: invoked by memory loader when `enabled: true`; results injected as `RelevantMemory[]` before freshness/truncation pipeline (SPEC-V3-MEM-001).
- Fallback: on API error, timeout, or `enabled: false`, return all scanned candidates (up to `max_candidates`) unfiltered (matches CC graceful degradation).
- Observability: every selector call emits a trace log entry with `{turn_id, candidate_count, selected_count, latency_ms, est_cost_usd, model}`.
- Cost estimation heuristic: `input_tokens ≈ sum(header_bytes)/4`, `output_tokens ≈ 256`, `cost_usd = Haiku table price × (in+out)`.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- Default-on behavior.
- Per-entry relevance scoring exposed to callers (return paths only).
- Vector embedding / local similarity index.
- Multi-turn learning / feedback loop.
- Team-memory relevance.
- Hard cost-budget enforcement (advisory config only; enforcement deferred).
- Cross-session caching of relevance decisions (turn-scoped only).
- Custom prompt override (CC-compat prompt is fixed; custom prompts deferred).
- Non-JSON output parsing (strict JSON schema only; malformed responses treated as fallback).
- Language-specific prompt variants (English prompt only in v3.0; user's `conversation_language` does not apply here — relevance selector is an internal side query).
- Multiple simultaneous selectors (one in-flight per turn).

---

## 3. Environment (환경)

- 런타임: Go 1.23+, moai-adk-go v3.0.0+.
- Claude Code v2.1.111+ (Haiku/Sonnet API available via CC's model router).
- API invocation path: moai does not directly call Anthropic API; instead it uses CC's slash-command or tool mechanism to request a side query OR uses a thin HTTP client with user's existing `CLAUDE_CODE_*` or `GLM_*` credentials. For v3.0, the approach is **stdlib `net/http` with the existing CC-compatible auth env vars** (no new SDK dependency).
- Platforms: macOS / Linux / Windows (network-stack agnostic).
- Target directories: `internal/core/memory/relevance/` (new).
- Dependencies: `net/http` (stdlib), `encoding/json` (stdlib). No new direct Go deps.
- Reuses: `internal/core/memory` types (SPEC-V3-MEM-001), `internal/config/schema/memory_schema.go` (SPEC-V3-MEM-001).

---

## 4. Assumptions (가정)

- A-MEM-002-001: User running with `memory.llm_relevance.enabled: true` has valid `ANTHROPIC_API_KEY` / `CLAUDE_CODE_*` credentials or `ZAI_API_KEY` (GLM) configured; moai does not authenticate.
- A-MEM-002-002: Haiku default model is sufficient for relevance judgment; CC uses Sonnet but moai's cost-consciousness justifies Haiku for v3.0. Users may opt into Sonnet via config.
- A-MEM-002-003: Memory directory contains < 200 files on average; scan cost is bounded.
- A-MEM-002-004: JSON schema response parsing is reliable with Haiku at 256 tokens for the prompt variant used.
- A-MEM-002-005: Per-turn cache sufficient; no cross-session persistence of relevance decisions needed.
- A-MEM-002-006: Graceful fallback on API failure is acceptable (return all candidates); relevance feature is opt-in, so failure mode = v3.0-default behavior.
- A-MEM-002-007: Prompt language (English) is compatible with all supported project languages since the selector operates on file headers in English-structured YAML frontmatter.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous Requirements

**REQ-MEM-002-001 (Ubiquitous) — default disabled**
The `memory.llm_relevance.enabled` config key **shall** default to `false` in the template `memory.yaml` shipped by this SPEC.

**REQ-MEM-002-002 (Ubiquitous) — config schema**
The `MemoryConfig.LLMRelevance` struct **shall** include fields `Enabled bool`, `Model string`, `TopK int`, `TimeoutSeconds int`, `MaxCandidates int`, `MaxOutputTokens int`, `MaxUSDPerSession float64`, `Cache CacheConfig` with validator/v10 tags per §2.1.

**REQ-MEM-002-003 (Ubiquitous) — types**
The package **shall** expose `RelevantMemory struct { Path string; MtimeMs int64 }` and `SelectorResult struct { Selected []RelevantMemory; Fallback bool; ErrorReason string; LatencyMs int; EstCostUSD float64 }`.

**REQ-MEM-002-004 (Ubiquitous) — prompt fidelity**
The relevance selector's system prompt **shall** contain (byte-for-byte prefix match): `"You are selecting memories that will be useful to Claude Code as it processes a user's query"` — ported from CC `findRelevantMemories.ts:18-24`.

**REQ-MEM-002-005 (Ubiquitous) — JSON schema output**
The selector **shall** request output with JSON schema `{"type":"object","properties":{"selected_memories":{"type":"array","items":{"type":"string"},"maxItems":<top_k>}},"required":["selected_memories"]}`.

### 5.2 Event-Driven Requirements

**REQ-MEM-002-010 (Event-Driven) — invocation**
**When** memory load is invoked with `enabled: true`, the loader **shall** call `relevance.Select(ctx, query, candidates, cfg)` and use the returned `Selected` list as the filtered candidate set for subsequent freshness/truncation processing.

**REQ-MEM-002-011 (Event-Driven) — cache hit**
**When** `relevance.Select` is called with a `(turnID, candidate-set-hash)` key that is already cached in the current turn, the selector **shall** return the cached `SelectorResult` without issuing an API call.

**REQ-MEM-002-012 (Event-Driven) — timeout**
**When** the API call exceeds `timeout_seconds`, the selector **shall** cancel the request, set `Fallback = true`, `ErrorReason = "timeout"`, return all candidates up to `MaxCandidates`, and emit a `memory.relevance.timeout` trace event.

**REQ-MEM-002-013 (Event-Driven) — API error**
**When** the API call returns a non-2xx status or unparseable JSON, the selector **shall** set `Fallback = true`, `ErrorReason = "api-error:{status}"` or `"parse-error"`, return all candidates up to `MaxCandidates`, and emit a `memory.relevance.error` trace event.

**REQ-MEM-002-014 (Event-Driven) — recentTools filter**
**When** the selector is invoked with a non-empty `recentTools` set, it **shall** exclude memory paths whose frontmatter `tags` field overlaps with tool names by keyword (port of CC `findRelevantMemories.ts` dedup against recentTools).

**REQ-MEM-002-015 (Event-Driven) — already-surfaced filter**
**When** the selector is invoked with a non-empty `alreadySurfaced` set of paths, it **shall** filter those paths out of the candidate list before the LLM call.

### 5.3 State-Driven Requirements

**REQ-MEM-002-020 (State-Driven) — disabled path**
**While** `enabled: false`, the selector wrapper **shall** return all candidates unfiltered (`Fallback = true`, `ErrorReason = "disabled"`) without issuing any API call.

**REQ-MEM-002-021 (State-Driven) — cache scope**
**While** `cache.scope: "turn"`, the cache **shall** be cleared at the start of each conversation turn; while `cache.scope: "off"`, no caching occurs.

**REQ-MEM-002-022 (State-Driven) — session cost advisory**
**While** cumulative `EstCostUSD` for the session exceeds `max_usd_per_session`, the selector **shall** emit a `memory.relevance.budget-warning` trace event. Selection **shall** continue (advisory only; no hard enforcement in v3.0).

### 5.4 Optional Features

**REQ-MEM-002-030 (Optional) — model selection**
**Where** `memory.llm_relevance.model: sonnet` is configured, the selector **shall** use Sonnet instead of Haiku. `model: inherit` **shall** use the same model as the parent conversation.

**REQ-MEM-002-031 (Optional) — max_candidates cap**
**Where** the scanned memory file count exceeds `max_candidates`, the scanner **shall** sort by mtime (newest first) and truncate to `max_candidates` before invoking the selector.

---

## 6. Acceptance Criteria (수용 기준 요약)

- **AC-MEM-002-01**: Default `memory.yaml` ships with `llm_relevance.enabled: false`; a fresh `moai init` project has relevance selection disabled out of the box. (maps REQ-MEM-002-001)
- **AC-MEM-002-02**: Invoking memory load with `enabled: true`, 10 candidate files, Haiku model: selector returns up to 5 paths (top_k default), emits trace event with model=haiku, latency, est cost. (maps REQ-MEM-002-005, REQ-MEM-002-010)
- **AC-MEM-002-03**: Second invocation within the same turn with identical candidate set returns cached result without API call (verified by zero HTTP requests on second call). (maps REQ-MEM-002-011)
- **AC-MEM-002-04**: Simulated 10-second API delay with `timeout_seconds: 5` triggers timeout; selector returns all candidates with `Fallback=true, ErrorReason="timeout"`. (maps REQ-MEM-002-012)
- **AC-MEM-002-05**: Simulated HTTP 500 response sets `Fallback=true, ErrorReason="api-error:500"`; emits `memory.relevance.error` trace. (maps REQ-MEM-002-013)
- **AC-MEM-002-06**: `memory.yaml` with `llm_relevance.top_k: 25` fails validation (max 20); with `top_k: 5` passes and selector returns max 5 entries. (maps REQ-MEM-002-002)
- **AC-MEM-002-07**: With `enabled: false`, calling `Select()` returns all candidates up to `MaxCandidates` without any network I/O. (maps REQ-MEM-002-020)
- **AC-MEM-002-08**: System prompt contains the exact porting prefix from CC (byte-exact match via test fixture). (maps REQ-MEM-002-004)
- **AC-MEM-002-09**: `go test ./internal/core/memory/relevance/...` passes with ≥ 85% coverage; network calls mocked via httptest.Server.

---

## 7. Constraints (제약)

- **[HARD] Opt-in default**: `enabled: false` is the only shipped default; no environment flag may override to true silently.
- **[HARD] No new direct Go deps**: reuse stdlib `net/http` and existing `encoding/json`; 9-direct-dep budget preserved.
- **[HARD] Graceful degradation**: every failure mode returns the unfiltered candidate list (never blocks memory load).
- **[HARD] Turn-scoped cache**: no cross-session or cross-turn persistence of relevance decisions.
- **[HARD] Prompt fidelity**: the system prompt prefix matches CC source byte-for-byte for downstream comparability.
- **[HARD] No hard cost enforcement**: `max_usd_per_session` is advisory only in v3.0 per master-v3 §9 question #3 (telemetry-driven re-evaluation in v3.2).
- **[HARD] Single in-flight request per turn**: concurrent calls in same turn coalesce via cache singleflight pattern.
- Binary size delta ≤ 150 KB.
- Latency budget: selector call p99 ≤ 3 seconds (Haiku); prefetch pattern hides this behind model streaming.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크 | 확률 | 영향 | 완화 |
|--------|------|------|------|
| User bill explodes on sessions with hundreds of turns | Medium | High | Default off in v3.0; `max_usd_per_session` advisory warning at 1× threshold; docs-site migration guide sections explicitly flag cost |
| Haiku returns lower-quality selections than Sonnet for complex queries | Medium | Medium | `model: sonnet` opt-in; `model: inherit` uses parent model for transparency |
| JSON parse errors on malformed Haiku output | Low | Low | REQ-MEM-002-013 falls back to all candidates; logged for telemetry |
| Candidate set hashing produces false cache hits | Very Low | Medium | Use SHA-256 of sorted paths + mtimes; collision risk negligible |
| LLM selector fails for non-English memory filenames | Low | Medium | Scanner operates on frontmatter `description` field (English-normalized) + filename hints; test corpus includes non-ASCII filenames |
| Prefetch pattern (starting selector before query loaded) not integrated in v3.0 | Medium | Low | v3.0 synchronous call is acceptable; hidden behind user opt-in; prefetch optimization deferred to v3.2 |
| Interaction with Freshness wrapper surfaces stale memories as "relevant" | Medium | Medium | Selector runs BEFORE freshness; freshness wrapper still applied to selected results. Order documented. |
| User with GLM backend expects GLM-compatible API call | Medium | Medium | Auth detection uses `ZAI_API_KEY` first, then `ANTHROPIC_API_KEY`; GLM endpoint honored when configured |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-MEM-001** — provides `MemoryEntry`, `MemoryType`, freshness wrapper, truncation, path validator, and `MemoryConfig` schema base.

### 9.2 Blocks

- None in v3.0. v3.2 features (relevance prefetch integration, default-on telemetry trigger) build atop this SPEC.

### 9.3 Related

- **SPEC-V3-SCH-001** — `MemoryConfig.LLMRelevance` inherits validator/v10 tagging pattern.
- **SPEC-V3-CLI-001** — `moai doctor memory --relevance` could surface selector config/status (optional future extension; not in scope here).

---

## 10. Traceability (추적성)

- Theme: master-v3 §3.5 (Theme 5 — Memory 2.0 Alignment, design approach #5 LLM-based relevance).
- Gap rows: gm#48 (LLM relevance missing), gm#49 (recentTools filter), gm#52 (cost budget advisory).
- Wave 1 sources:
  - findings-wave1-query-context.md §6.10 (`findRelevantMemories.ts` full flow)
  - findings-wave1-query-context.md §6.11 (`memoryScan.ts` MAX_MEMORY_FILES=200, FRONTMATTER_MAX_LINES=30)
  - findings-wave1-query-context.md §6.14 (prefetch pattern — future optimization)
  - findings-wave1-moai-current.md §6 (current moai memory state — no relevance selection exists)
- BC-ID: none (additive, opt-in).
- REQ 총 개수: 15 (Ubiquitous 5, Event-Driven 6, State-Driven 3, Optional 2 discrete — sum 16, collapsed Optional overlap to 15 unique IDs).
- 예상 AC 개수: 9.
- 코드 구현 예상 경로:
  - `internal/core/memory/relevance/types.go`, `selector.go`, `scanner.go`, `cache.go`
  - `internal/core/memory/relevance/selector_test.go` (httptest.Server mocking)
  - Template `.moai/config/sections/memory.yaml` updated with `llm_relevance:` block.

---

End of SPEC.
