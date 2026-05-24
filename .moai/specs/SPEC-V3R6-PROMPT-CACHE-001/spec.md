---
id: SPEC-V3R6-PROMPT-CACHE-001
title: "Anthropic Prompt Caching 도입 (1h 세션 시작 + 5m SPEC body)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/runtime, internal/hook"
lifecycle: spec-anchored
tags: "prompt-cache, anthropic, cost-optimization, hit-rate, telemetry, v3.0"
depends_on:
  - SPEC-V3R6-RULES-PATH-SCOPE-001
  - SPEC-V3R6-RULES-COMPRESS-001
tier: M
---

# SPEC-V3R6-PROMPT-CACHE-001 — Anthropic Prompt Caching 도입 (1h 세션 시작 + 5m SPEC body)

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-05-23 | GOOS행님 | Initial plan-phase draft (v0.1.0) — Sprint 2 Wave 6 |

## 1. Background

`.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 5 "Prompt Caching 적극 활용" 결정 사항:

> 1h cache write at session start + 5min cache for SPEC body. Cache Read 90% off, Cache Write 5min +25%, Cache Write 1h +100%.
>
> 손익분기 (§ 2.3): 1h cache write 1회 + 14회 hit → $16.6 (no cache) → $7.06 (with cache), 일일 -57%.
>
> 1.4 가격 레버: cache_creation_input_tokens는 1회만 발생, cache_read_input_tokens는 매 후속 turn마다 90% 할인 적용된다.

§ 5 Wave 2 표에서 본 SPEC는 `SPEC-V3R6-PROMPT-CACHE-001 Tier M`로 명시되어 있다. 현재 baseline (`moai-adk-current-state-2026-05-22.md` § 2)은 cache 미사용으로 0% hit rate이며, settings.json 책임 분리 (§ 7.4)로 cache_control 주입 위치는 `internal/cli/cc.go` SDK wrapper 진입점이다.

Sprint 1 Lane A 4/4 SPEC (2026-05-23 commit `fa658d927`) 머지로 always-loaded rule baseline은 ~54K 감축되었고, 캐시 prefix exact-match 요구사항(R2)의 churn 위험이 사전 해소되어 본 SPEC의 진입 조건이 충족되었다.

## 2. Goal

단일 `/moai run` 턴 평균 비용 $1.10 → ≤ $0.45로 감축한다 (design doc § 6.1 KPI). cache hit rate ≥ 80% (7-day rolling) 달성을 통해 종량제 Agent SDK 풀에서도 지속 가능한 토큰 이코노미를 확보한다.

## 3. Scope

### In Scope

1. **1h cache write at session start**: CLAUDE.md + always-loaded rules (Sprint 1 후 ~15K) + output style + MCP initial context을 단일 cache breakpoint로 묶어 `ttl: "1h"` 지정. 세션 첫 turn에 cache_creation 발생, 이후 모든 turn에 cache_read (90% off) 적용.
2. **5min cache for SPEC body**: `/moai run SPEC-XXX` 진입 시 해당 SPEC의 `spec.md` + `acceptance.md` + `plan.md` 묶음 뒤에 `ttl: "5m"` breakpoint 주입. RUN phase 내 multi-turn 작업에서 SPEC body 재로드 비용 90% 절감.
3. **Implementation surface**:
   - `internal/cli/cc.go` (또는 신규 `internal/runtime/cache_control.go`)에서 Claude Code SDK wrapper outgoing Anthropic API request에 `cache_control: {type: "ephemeral", ttl: "1h"|"5m"}` 자동 주입.
   - 신규 `.moai/config/sections/cache.yaml` (또는 기존 `runtime.yaml` 확장) 도입: `cacheStrategy: {enabled: bool, session_ttl: "1h"|"5m"|"off", spec_ttl: "5m"|"off"}`.
4. **Telemetry hook**: PostToolUse 후크가 API response에서 `cache_creation_input_tokens` / `cache_read_input_tokens` 필드 추출 후 `.moai/state/cache-usage.jsonl`에 append. 일별 집계 → cache hit rate metric 산출.
5. **Smart 손익분기 guard**: 세션 길이 < 5분 (no second turn) 시 `cacheStrategy.session_ttl: "off"` bypass 권고 → 단일-turn 세션에서 +100% cache_write 페널티 회피.

### Out of Scope — Cache breakpoint scope limits

- Auto-placing additional cache breakpoints inside long messages (per-paragraph optimization) — deferred to SPEC-V3R6-CACHE-GRANULAR-001.
- GLM (Z.AI) cache compatibility — uses different control fields; deferred to Wave 3 SPEC-V3R6-BACKEND-ROUTING-001.
- Pre-emptive background cache warming (cron-style) — out of scope; cache activates only on first user-triggered API call.

Each of the three excluded categories is documented in the H4 sub-headings below with full rationale.

#### Out of Scope: Per-message cache breakpoint optimization

Auto-placing additional cache breakpoints inside long messages (예: per-paragraph)은 본 SPEC 범위 밖이다. 본 SPEC는 2개 고정 breakpoint (session start + SPEC body)만 설정한다. 향후 SPEC-V3R6-CACHE-GRANULAR-001로 분리.

#### Out of Scope: GLM cache compatibility

GLM (Z.AI) cache 메커니즘은 다른 control 필드를 사용한다. 본 SPEC는 Anthropic Claude SDK에만 적용된다. GLM cache 지원은 Wave 3 SPEC-V3R6-BACKEND-ROUTING-001에서 다룬다.

#### Out of Scope: Pre-emptive cache warming

사용자가 세션 진입 전 background process로 cache를 미리 워밍업하는 기능 (예: cron job)은 본 SPEC 범위 밖이다. 본 SPEC는 첫 user-triggered API call 시점에만 cache를 활성화한다.

## 4. EARS Requirements

Where notation guidance: GEARS notation 사용 (When/Where/While/If 합성 허용). Zero legacy IF/THEN.

### REQ-PC-001 (Ubiquitous)

The orchestrator shall inject `cache_control: {type: "ephemeral", ttl: "1h"}` on the LAST item of the system prompt portion containing always-loaded rules + CLAUDE.md + output style at session start.

### REQ-PC-002 (When)

When the user invokes `/moai run SPEC-XXX`, the orchestrator shall inject `cache_control: {type: "ephemeral", ttl: "5m"}` on a breakpoint that follows the SPEC `spec.md` + `acceptance.md` + `plan.md` bundled context.

### REQ-PC-003 (Where)

Where the active backend is GLM (`llm.mode == "glm"`), the system shall skip cache_control injection entirely (deferred to Wave 3 SPEC-V3R6-BACKEND-ROUTING-001).

### REQ-PC-004 (Ubiquitous)

The PostToolUse hook shall extract `cache_creation_input_tokens` and `cache_read_input_tokens` fields from each Anthropic API response and append a JSONL entry to `.moai/state/cache-usage.jsonl` containing timestamp, turn index, and both token counts.

### REQ-PC-005 (When)

When `cacheStrategy.session_ttl == "off"` in config, the orchestrator shall not inject any session-level cache breakpoint and shall log the bypass reason at session start.

### REQ-PC-006 (Ubiquitous)

The `moai doctor` command shall report current cache hit rate (last 7 days, calculated as `sum(cache_read_input_tokens) / (sum(cache_read_input_tokens) + sum(cache_creation_input_tokens))`) when `cacheStrategy.enabled == true`.

### REQ-PC-007 (While)

While a SPEC `/moai run` session is active AND total elapsed wall-time < 5 minutes AND only 1 turn has occurred, the system shall log a "single-turn cache write penalty risk" warning at session end with concrete recommendation to set `session_ttl: "off"` for the affected workflow.

## 5. Acceptance Criteria

See `acceptance.md` for 9 binary AC-PC-001..009.

## 6. Risks

| ID | Severity/Likelihood | Risk | Mitigation |
|----|---------------------|------|------------|
| R1 | Medium/High | 1h cache_write +100% 페널티가 단일-turn 세션 (예: 사용자 세션 즉시 종료)에서 발생 | REQ-PC-007 단일-turn 경고 + REQ-PC-005 opt-out flag (`session_ttl: "off"`) |
| R2 | Medium/Medium | Cache prefix exact-match 요구사항 — 턴 사이 ANY rule 변경 시 cache 무효화 | Sprint 1 Lane A (rules-compress/skill-compress) 머지로 always-loaded 안정화 → rule churn 가정 무효 |
| R3 | Low/Medium | Anthropic SDK 업그레이드 시 cache_control 필드 스키마 변경 가능 | SDK 버전 pin + cache_control 스키마 integration test (AC-PC-003) |
| R4 | Medium/Medium | **Cross-Sprint with SPEC-V3R6-AGENT-MODEL-ROUTING-001** — sonnet model은 1024-token minimum cacheable, haiku는 2048. Session start payload (Sprint 1 후 ~50K 추정)은 양쪽 임계값 위. | M1 구현 시 실측 + 임계값 미달 시 cache_control 미주입 fallback. KNOWN CONFLICT: AGENT-MODEL-ROUTING-001 머지 후 model-specific token count 변동 재검증 필요. |
| R5 | Low/Low | `.moai/state/cache-usage.jsonl` 무한 성장 | 월별 rotation 또는 truncate — 본 SPEC는 append-only, 회수는 후속 telemetry SPEC로 deferred |

## 7. Out of Scope Sections

본 SPEC에 명시된 3개 h3 Out of Scope 섹션 (§3 In Scope 뒤)을 참조한다.

## 8. References

- `.moai/research/v3.0-design-2026-05-22.md` § 1.4 (가격 레버), § 2.3 (손익분기 계산), § 4 Layer 5 (Prompt Caching 적극 활용), § 5 (Wave 2 표), § 6.1 (KPI), § 7.4 (settings.json 책임 분리)
- `.moai/research/moai-adk-current-state-2026-05-22.md` § 2 (현재 cache 0% baseline), § 7.4 (settings.json baseline)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical 12-field SSOT)
- `.claude/rules/moai/development/sprint-wave-naming.md` (Sprint = multi-SPEC, Wave = within-SPEC)
- Anthropic Prompt Caching official docs: cache_control mechanism (ephemeral with `ttl: "5m"` | `ttl: "1h"`); breakpoint position in system / messages array; minimum cacheable tokens (1024 Sonnet/Opus, 2048 Haiku); cache hit on EXACT prefix match
- GEARS-MIGRATION-001 (머지된 `134a43fac` 2026-05-22) — GEARS notation 사용 의무

## 9. Cross-References

- **Sprint 2 sibling SPECs**:
  - `SPEC-V3R6-AGENT-MODEL-ROUTING-001` (R4 — model-specific minimum cacheable token thresholds, sonnet 1024 / haiku 2048)
  - `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` (independent — REQ-PC-004 PostToolUse hook은 본 SPEC 단독 구현)
  - `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (R3 — SDK schema change cascade)
- **Sprint 1 (머지 완료)**: SPEC-V3R6-RULES-PATH-SCOPE-001 + SPEC-V3R6-RULES-COMPRESS-001 (depends_on — baseline 감축 선행 조건)

## 10. Section A-E Tier M MANDATORY References

상세는 `plan.md` 참조:

- **Section A** (Baseline): design doc § 4 Layer 5 + § 2.3 손익분기 verbatim 인용
- **Section B** (Goal/KPI): 평균 비용 $1.10 → ≤ $0.45, cache hit rate ≥ 80% (7-day)
- **Section C** (Requirements): REQ-PC-001..007 ↔ AC-PC-001..009 100% traceability
- **Section D** (Milestones): M1 cache_control inject (cc.go/runtime) → M2 cache.yaml config schema → M3 PostToolUse telemetry hook → M4 moai doctor metric + REQ-PC-007 warning → M5 docs-site 4-locale mirror
- **Section E** (Risks/Out of Scope): R1-R5 + 3 h3 Out of Scope (이 spec.md §3 In Scope 뒤)
