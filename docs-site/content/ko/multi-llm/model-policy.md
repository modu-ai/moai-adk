---
title: 모델 정책
weight: 30
draft: false
---

## 모델 정책이란?

모델 정책은 MoAI-ADK 토크노믹스의 뼈대입니다. "모든 작업에 최고 모델"이
아니라, 에이전트마다 — 계획·감사처럼 추론이 무거운 일과 문서화·Git처럼
가벼운 일마다 — 알맞은 모델을 선언적으로 배정합니다. Claude Code 구독 플랜에
맞춰 품질을 극대화하면서 요율 제한 에러를 방지합니다.

MoAI-ADK v3.0의 에이전트 카탈로그는 **10개** (MoAI 커스텀 9개 + Anthropic
내장 `Explore`)이며, 아래 배정표는 그중 모델 정책이 직접 배정하는 핵심 7개
에이전트를 다룹니다.

## 3단계 정책 개요

| 정책 | 플랜 | Opus | Sonnet | Haiku | 적합한 용도 |
|------|------|---------|-----------|----------|-----------|
| **High** | Max $200/월 | 5 | 1 | 1 | 최고 품질, 최대 처리량 |
| **Medium** | Max $100/월 | 2 | 3 | 2 | 품질과 비용의 균형 |
| **Low** | Plus $20/월 | 0 | 4 | 3 | 저예산, Opus 미포함 |

> **왜 중요한가요?** Plus $20 플랜은 Opus에 접근할 수 없습니다. `Low` 정책을 설정하면 모든 에이전트가 Sonnet과 Haiku만 사용하여 요율 제한 에러를 방지합니다. 상위 플랜은 핵심 에이전트(계획, 감사)에 Opus를 배정하고 일상 작업에는 Sonnet/Haiku를 사용합니다.

## 에이전트별 모델 배정표

### Manager Agents (4개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | haiku | haiku |
| manager-git | haiku | haiku | haiku |

### Evaluator & Builder Agents (3개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | haiku |

> Anthropic 내장 `Explore`는 읽기 전용 탐색 에이전트로 별도 배정 없이
> 동작합니다. Agent Teams 정적 계층(정적 role profile)은 v3.0에서
> 은퇴했으며, 병렬 작업은 sub-agent 병렬 실행과 동적 워크플로우가
> 대체합니다. `moai cg`의 teammate 런타임(tmux pane)은 그대로 유지됩니다.

## 배정 원칙

- **항상 Opus**: 계획 감사(plan-auditor), SPEC 작성(manager-spec) — 높은 추론 능력 필요
- **항상 Haiku**: Git(manager-git) — 가볍고 빠른 작업
- **플랜에 따라 변동**: 구현(manager-develop, cycle_type=tdd/ddd) — 플랜이 높을수록 Opus

계획을 만든 에이전트가 감사하지 않도록 plan-auditor와 sync-auditor는 독립
배정을 유지합니다 — 비용 축과 품질 축(편향 방지)이 함께 설계된 표입니다.

## v3.0 확장: Tier×Phase 선언 축

v3.0에서는 에이전트 단위 배정 위에 **작업 단계(phase)와 SPEC 크기(Tier)**
축이 더해졌습니다. `internal/config/model_routing.go`가 Tier×Phase →
{model, effort} 매트릭스를 선언적으로 관리합니다:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (추론 깊이): low / medium / high / xhigh / max
- **tier** (SPEC 크기): S / M / L
- **phase** (작업 단계): plan / run / sync / mx

같은 워크플로우라도 API 종량제와 구독 요금제는 최적 배분이 다르기 때문에,
요금제 인지 (plan_type) 프로파일이 요금제별 매트릭스를 분리 적용합니다.

## 설정 방법

### 프로젝트 초기화 시

```bash
moai init my-project
# 대화형 위자드에서 모델 정책 선택 포함
```

### 기존 프로젝트 재설정

```bash
moai update
# 대화형 프롬프트:
# - Reset model policy? (y/n) — 모델 정책 재설정
# - Update GLM settings? (y/n) — GLM 환경변수 설정
```

> 기본 정책은 `High`입니다. GLM 설정은 `settings.local.json`에 격리되어 Git에 커밋되지 않습니다.

## 다음 단계

- [CG 모드](/ko/multi-llm/cg-mode) — Claude + GLM 하이브리드로 비용 절감
- [에이전트 가이드](/ko/advanced/agent-guide) — 에이전트 커스터마이징
- [CLI 레퍼런스](/ko/getting-started/cli) — moai init, moai update 상세
