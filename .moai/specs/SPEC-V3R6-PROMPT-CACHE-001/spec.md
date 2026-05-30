---
id: SPEC-V3R6-PROMPT-CACHE-001
title: "Anthropic Prompt Caching 도입 (1h 세션 시작 + 5m SPEC body)"
version: "0.1.1"
status: draft
created: 2026-05-23
updated: 2026-05-30
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/cli, internal/runtime, internal/hook"
lifecycle: spec-anchored
tags: "prompt-cache, anthropic, cost-optimization, hit-rate, telemetry, v3.0"
depends_on:
  - SPEC-V3R6-RULES-PATH-SCOPE-001
  - SPEC-V3R6-RULES-COMPRESS-001
related_specs:
  - SPEC-V3R6-AGENT-MODEL-ROUTING-001
tier: M
---

# SPEC-V3R6-PROMPT-CACHE-001 — Anthropic Prompt Caching 도입 (1h 세션 시작 + 5m SPEC body)

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-05-23 | GOOS행님 | Initial plan-phase draft (v0.1.0) — Sprint 2 Round 6 |
| 2026-05-30 | manager-spec | Plan patch (v0.1.1) — D1 economic premise re-grounded as estimated/illustrative (cited research baselines absent on disk); D2 R4 decoupled from obsolete SPEC-V3R6-AGENT-MODEL-ROUTING-001 (now `related_specs`, not a hard dependency) via self-contained `min_cacheable_tokens` config default; D3 AC-PC-010 added for REQ-PC-007; D5 `sprint-wave-naming.md` → `sprint-round-naming.md`, Wave → Round terminology |

## 1. Background

본 SPEC는 Anthropic Prompt Caching (ephemeral `cache_control` with `ttl: "5m"` | `ttl: "1h"`)을 MoAI-ADK 워크플로우에 도입하여, 종량제 Agent SDK 풀에서 always-loaded 컨텍스트(CLAUDE.md + always-loaded rules + output style)와 SPEC body의 재로드 비용을 절감하는 것을 목표로 한다.

### 비용 모델 (estimated / illustrative — 검증된 출처 없음)

> **[추정 근거 — verbatim 인용 아님]** 아래 손익분기 수치는 **공개된 Anthropic Prompt Caching 가격 정책 구조**(Cache Read 입력 단가 대비 약 90% off, Cache Write 5min +25%, Cache Write 1h +100%)에 기반한 **illustrative 산술 예시**이며, 특정 프로젝트 측정값이나 commit된 research 문서의 verbatim 인용이 아니다. 초안(v0.1.0)이 출처로 명시했던 `.moai/research/v3.0-design-2026-05-22.md` / `.moai/research/moai-adk-current-state-2026-05-22.md` 는 현재 디스크에 존재하지 않으므로(§ 8 출처 정정 참조) 해당 수치는 검증 가능한 baseline이 아니다. 실제 손익분기는 워크플로우 패턴(턴 수, 컨텍스트 크기, 모델 단가)에 따라 달라지며, 머지 후 K1-K5 KPI로 실측 검증한다.

- **가격 레버 (Anthropic 공개 정책 구조)**: `cache_creation_input_tokens`는 세션 첫 turn 1회만 발생하고, `cache_read_input_tokens`는 매 후속 turn마다 입력 단가 대비 약 90% 할인이 적용된다.
- **illustrative 손익분기 (Sonnet 단가 가정, 일일 ~15 turn 가정)**: cache 미사용 시 약 `15 turn × turn당 단가`, 1h cache 사용 시 `1회 write (+100%) + 14회 read (−90%)` 구조로, 일일 비용이 약 50% 안팎 절감될 수 있다는 추정. 정확한 수치는 머지 후 실측한다(§ 2 Goal + plan.md § B KPI 참조).

이 estimated 손익분기는 본 SPEC의 **동기(motivation)**이며, **acceptance 기준은 아니다**. 구현 정합성은 § 4 REQ-PC-001..007 + acceptance.md의 binary AC로만 판정한다.

### 진입 조건

현재 baseline은 cache 미사용(0% hit rate)이다. cache_control 주입 위치는 `internal/cli/cc.go` SDK wrapper 진입점(또는 신규 `internal/runtime/cache_control.go`)이다.

Sprint 1 Lane A 4/4 SPEC (2026-05-23 commit `fa658d927`) 머지로 always-loaded rule baseline이 감축되었고, 캐시 prefix exact-match 요구사항(R2)의 churn 위험이 사전 해소되어 본 SPEC의 진입 조건이 충족되었다.

## 2. Goal

cache hit rate ≥ 80% (7-day rolling) 달성을 통해 종량제 Agent SDK 풀에서도 지속 가능한 토큰 이코노미를 확보한다.

평균 turn 비용 절감(예: turn당 비용 감축)은 illustrative 손익분기(§ 1) 기반의 **방향성 목표(estimated)**이며 hard acceptance 기준이 아니다. 절감 금액의 구체 수치는 검증된 baseline 출처가 없으므로(§ 1, § 8 참조) 머지 후 K1-K5 KPI(plan.md § B)로 실측한다. cache hit rate ≥ 80%만이 측정 가능한 1차 성공 지표이고, 비용 절감액은 측정값에서 파생된다.

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
- GLM (Z.AI) cache compatibility — uses different control fields; deferred to Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001.
- Pre-emptive background cache warming (cron-style) — out of scope; cache activates only on first user-triggered API call.

Each of the three excluded categories is documented in the H4 sub-headings below with full rationale.

#### Out of Scope: Per-message cache breakpoint optimization

Auto-placing additional cache breakpoints inside long messages (예: per-paragraph)은 본 SPEC 범위 밖이다. 본 SPEC는 2개 고정 breakpoint (session start + SPEC body)만 설정한다. 향후 SPEC-V3R6-CACHE-GRANULAR-001로 분리.

#### Out of Scope: GLM cache compatibility

GLM (Z.AI) cache 메커니즘은 다른 control 필드를 사용한다. 본 SPEC는 Anthropic Claude SDK에만 적용된다. GLM cache 지원은 Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001에서 다룬다.

#### Out of Scope: Pre-emptive cache warming

사용자가 세션 진입 전 background process로 cache를 미리 워밍업하는 기능 (예: cron job)은 본 SPEC 범위 밖이다. 본 SPEC는 첫 user-triggered API call 시점에만 cache를 활성화한다.

## 4. EARS Requirements

Where notation guidance: GEARS notation 사용 (When/Where/While/If 합성 허용). Zero legacy IF/THEN.

### REQ-PC-001 (Ubiquitous)

The orchestrator shall inject `cache_control: {type: "ephemeral", ttl: "1h"}` on the LAST item of the system prompt portion containing always-loaded rules + CLAUDE.md + output style at session start.

### REQ-PC-002 (When)

When the user invokes `/moai run SPEC-XXX`, the orchestrator shall inject `cache_control: {type: "ephemeral", ttl: "5m"}` on a breakpoint that follows the SPEC `spec.md` + `acceptance.md` + `plan.md` bundled context.

### REQ-PC-003 (Where)

Where the active backend is GLM (`llm.mode == "glm"`), the system shall skip cache_control injection entirely (deferred to Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001).

### REQ-PC-004 (Ubiquitous)

The PostToolUse hook shall extract `cache_creation_input_tokens` and `cache_read_input_tokens` fields from each Anthropic API response and append a JSONL entry to `.moai/state/cache-usage.jsonl` containing timestamp, turn index, and both token counts.

### REQ-PC-005 (When)

When `cacheStrategy.session_ttl == "off"` in config, the orchestrator shall not inject any session-level cache breakpoint and shall log the bypass reason at session start.

### REQ-PC-006 (Ubiquitous)

The `moai doctor` command shall report current cache hit rate (last 7 days, calculated as `sum(cache_read_input_tokens) / (sum(cache_read_input_tokens) + sum(cache_creation_input_tokens))`) when `cacheStrategy.enabled == true`.

### REQ-PC-007 (While)

While a SPEC `/moai run` session is active AND total elapsed wall-time < 5 minutes AND only 1 turn has occurred, the system shall log a "single-turn cache write penalty risk" warning at session end with concrete recommendation to set `session_ttl: "off"` for the affected workflow.

## 5. Acceptance Criteria

See `acceptance.md` for 10 binary AC-PC-001..010 (REQ-PC-007 ↔ AC-PC-010 추가로 모든 normative `shall`이 매핑됨).

## 6. Risks

| ID | Severity/Likelihood | Risk | Mitigation |
|----|---------------------|------|------------|
| R1 | Medium/High | 1h cache_write +100% 페널티가 단일-turn 세션 (예: 사용자 세션 즉시 종료)에서 발생 | REQ-PC-007 단일-turn 경고 + REQ-PC-005 opt-out flag (`session_ttl: "off"`) |
| R2 | Medium/Medium | Cache prefix exact-match 요구사항 — 턴 사이 ANY rule 변경 시 cache 무효화 | Sprint 1 Lane A (rules-compress/skill-compress) 머지로 always-loaded 안정화 → rule churn 가정 무효 |
| R3 | Low/Medium | Anthropic SDK 업그레이드 시 cache_control 필드 스키마 변경 가능 | SDK 버전 pin + cache_control 스키마 integration test (AC-PC-003) |
| R4 | Low/Medium | **Model-specific minimum cacheable tokens** — Anthropic 공개 정책상 모델별 최소 캐시 가능 토큰이 다르다(예: sonnet/opus 1024, haiku 2048). Session start payload가 임계값 미달 시 cache_control 주입이 무효가 될 수 있다. | **자체 완결적 처리(self-contained)**: M2 `cache.yaml`에 `min_cacheable_tokens` 설정 키(보수 default 2048 — haiku 임계값 기준 안전 상한)를 도입하고, M1은 payload 크기가 이 임계값 미만일 때 cache_control을 주입하지 않는 **threshold-agnostic fallback**으로 구현한다. SPEC-V3R6-AGENT-MODEL-ROUTING-001의 per-agent model 배정에 **의존하지 않는다** (decoupled — 아래 R4 NOTE 참조). |
| R5 | Low/Low | `.moai/state/cache-usage.jsonl` 무한 성장 | 월별 rotation 또는 truncate — 본 SPEC는 append-only, 회수는 후속 telemetry SPEC로 deferred |

> **R4 NOTE — SPEC-V3R6-AGENT-MODEL-ROUTING-001 의존성 제거 (decoupled)**: 본 SPEC는 SPEC-V3R6-AGENT-MODEL-ROUTING-001을 **hard dependency로 두지 않는다** (frontmatter `depends_on`에 포함하지 않으며, `related_specs`로만 참조한다). 이유: (1) 해당 SPEC는 현재 `status: draft` 상태로 미머지/미구현이며, plan-audit에서 **신뢰 불가(obsolete 전제 기반 — 폐기된 23-agent 카탈로그 가정)**로 판정되어 본 SPEC의 진입 게이트가 될 수 없다. 실제 retained MoAI 에이전트 카탈로그는 7개(flat)이다. (2) 모델별 최소 캐시 가능 토큰 임계값은 Anthropic 공개 정책에서 직접 유래하는 상수이므로, per-agent model 배정 결정 없이도 `cache.yaml`의 `min_cacheable_tokens` 설정 키(보수 default 2048)로 자체 완결 처리가 가능하다. 따라서 M1 cache_control fallback은 **threshold-agnostic**이며, AGENT-MODEL-ROUTING-001이 머지되든 폐기되든 본 SPEC run-phase 정합성에 영향이 없다. AGENT-MODEL-ROUTING-001이 향후 안정화되면 `min_cacheable_tokens`를 per-model로 세분화하는 것은 별도 후속 작업이다(본 SPEC 범위 외).

## 7. Out of Scope Sections

본 SPEC에 명시된 3개 h3 Out of Scope 섹션 (§3 In Scope 뒤)을 참조한다.

## 8. References

> **출처 정정 (2026-05-30, v0.1.1)**: 초안(v0.1.0)이 verbatim 인용 출처로 명시했던 `.moai/research/v3.0-design-2026-05-22.md` 및 `.moai/research/moai-adk-current-state-2026-05-22.md` 두 파일은 **현재 디스크에 존재하지 않는다**(committed 흔적 없음 — 미커밋이거나 rename된 것으로 추정). 따라서 § 1 손익분기 비용 모델은 verbatim 인용에서 **estimated / illustrative** 프레이밍으로 정정되었다. 본 SPEC의 구현 정합성은 § 4 REQ + acceptance.md AC로만 판정되므로 출처 부재가 요구사항 자체를 무효화하지 않는다.

- **Anthropic Prompt Caching official docs** (1차 출처 — run-phase에서 WebFetch verify 대상): `cache_control` mechanism (ephemeral with `ttl: "5m"` | `ttl: "1h"`); breakpoint position in system / messages array; model-specific minimum cacheable tokens (공개 정책상 모델별 상이); cache hit on EXACT prefix match
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical 12-field SSOT)
- `.claude/rules/moai/development/sprint-round-naming.md` (Sprint = multi-SPEC, Round = within-SPEC; "Wave" terminology는 AP-SRN-004로 retired)
- GEARS-MIGRATION-001 (머지된 `134a43fac` 2026-05-22) — GEARS notation 사용 의무

## 9. Cross-References

- **Sprint 2 sibling SPECs**:
  - `SPEC-V3R6-AGENT-MODEL-ROUTING-001` (**related_specs only — NOT a hard dependency**; § 6 R4 NOTE 참조. 현재 `status: draft` + plan-audit 신뢰 불가 판정으로 본 SPEC가 의존하지 않음. model-specific minimum cacheable token은 `cache.yaml` `min_cacheable_tokens` 보수 default로 자체 처리)
  - `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` (independent — REQ-PC-004 PostToolUse hook은 본 SPEC 단독 구현)
  - `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (R3 — SDK schema change cascade)
- **Sprint 1 (머지 완료)**: SPEC-V3R6-RULES-PATH-SCOPE-001 + SPEC-V3R6-RULES-COMPRESS-001 (depends_on — baseline 감축 선행 조건)

## 10. Section A-E Tier M MANDATORY References

상세는 `plan.md` 참조:

- **Section A** (Baseline): Anthropic Prompt Caching 공개 정책 구조 기반 estimated/illustrative 손익분기 (검증된 verbatim 출처 없음 — § 1, § 8 참조)
- **Section B** (Goal/KPI): cache hit rate ≥ 80% (7-day) 1차 지표; 비용 절감액은 estimated 방향성 목표(머지 후 K1-K5 실측)
- **Section C** (Requirements): REQ-PC-001..007 ↔ AC-PC-001..010 100% traceability
- **Section D** (Milestones): M1 cache_control inject (cc.go/runtime, threshold-agnostic fallback) → M2 cache.yaml config schema (`min_cacheable_tokens` 포함) → M3 PostToolUse telemetry hook → M4 moai doctor metric + REQ-PC-007 warning → M5 docs-site 4-locale mirror
- **Section E** (Risks/Out of Scope): R1-R5 + 3 h3 Out of Scope (이 spec.md §3 In Scope 뒤)
