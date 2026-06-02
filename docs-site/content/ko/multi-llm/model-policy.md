---
title: 모델 정책
weight: 30
draft: false
---

## 모델 정책이란?

MoAI-ADK는 8개의 유지된 에이전트(7개 MoAI 커스텀 + Anthropic 내장 Explore) 각각에 최적의 AI 모델을 배정합니다. Claude Code 구독 플랜에 맞춰 품질을 극대화하면서 요율 제한 에러를 방지합니다.

## 3단계 정책 개요

| 정책 | 플랜 | 🟣 Opus | 🔵 Sonnet | 🟡 Haiku | 적합한 용도 |
|------|------|---------|-----------|----------|-----------|
| **High** | Max $200/월 | 5 | 1 | 1 | 최고 품질, 최대 처리량 |
| **Medium** | Max $100/월 | 2 | 3 | 2 | 품질과 비용의 균형 |
| **Low** | Plus $20/월 | 0 | 4 | 3 | 저예산, Opus 미포함 |

> **왜 중요한가요?** Plus $20 플랜은 Opus에 접근할 수 없습니다. `Low` 정책을 설정하면 모든 에이전트가 Sonnet과 Haiku만 사용하여 요율 제한 에러를 방지합니다. 상위 플랜은 핵심 에이전트(보안, 전략, 아키텍처)에 Opus를 배정하고 일상 작업에는 Sonnet/Haiku를 사용합니다.

## 에이전트별 모델 배정표

### Manager Agents (4개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| manager-spec | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-develop | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-docs | 🔵 sonnet | 🟡 haiku | 🟡 haiku |
| manager-git | 🟡 haiku | 🟡 haiku | 🟡 haiku |

### Evaluator & Builder Agents (3개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| plan-auditor | 🟣 opus | 🟣 opus | 🔵 sonnet |
| sync-auditor | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| builder-harness | 🟣 opus | 🔵 sonnet | 🟡 haiku |

> Team(팀) 모드의 역할(researcher, analyst, architect, implementer, tester, designer, reviewer)은 정적 에이전트가 아니라 `workflow.yaml`의 role profile로 `Agent(general-purpose)`를 통해 동적 생성됩니다.

## 배정 원칙

- **항상 Opus**: 계획 감사(plan-auditor), SPEC 작성(manager-spec) — 높은 추론 능력 필요
- **항상 Haiku**: 문서화(manager-docs), Git(manager-git) — 가볍고 빠른 작업
- **플랜에 따라 변동**: 구현(manager-develop, cycle_type=tdd/ddd) — 플랜이 높을수록 Opus

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
