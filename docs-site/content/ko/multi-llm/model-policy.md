---
title: 모델 정책
weight: 30
draft: false
---
# 모델 정책

## 모델 정책이란?

MoAI-ADK는 24개 이상의 에이전트 각각에 최적의 AI 모델을 배정합니다. Claude Code 구독 플랜에 맞춰 품질을 극대화하면서 요율 제한 에러를 방지합니다.

## 3단계 정책 개요

| 정책 | 플랜 | 🟣 Opus | 🔵 Sonnet | 🟡 Haiku | 적합한 용도 |
|------|------|---------|-----------|----------|-----------|
| **High** | Max $200/월 | 16 | 5 | 3 | 최고 품질, 최대 처리량 |
| **Medium** | Max $100/월 | 3 | 17 | 4 | 품질과 비용의 균형 |
| **Low** | Plus $20/월 | 0 | 13 | 11 | 저예산, Opus 미포함 |

> **왜 중요한가요?** Plus $20 플랜은 Opus에 접근할 수 없습니다. `Low` 정책을 설정하면 모든 에이전트가 Sonnet과 Haiku만 사용하여 요율 제한 에러를 방지합니다. 상위 플랜은 핵심 에이전트(보안, 전략, 아키텍처)에 Opus를 배정하고 일상 작업에는 Sonnet/Haiku를 사용합니다.

## 에이전트별 모델 배정표

### Manager 에이전트 (8개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| manager-spec | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-strategy | 🟣 opus | 🟣 opus | 🔵 sonnet |
| manager-ddd | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-tdd | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| manager-project | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| manager-docs | 🔵 sonnet | 🟡 haiku | 🟡 haiku |
| manager-quality | 🟡 haiku | 🟡 haiku | 🟡 haiku |
| manager-git | 🟡 haiku | 🟡 haiku | 🟡 haiku |

### Expert 에이전트 (8개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| expert-backend | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-frontend | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-security | 🟣 opus | 🟣 opus | 🔵 sonnet |
| expert-debug | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-refactoring | 🟣 opus | 🔵 sonnet | 🔵 sonnet |
| expert-devops | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| expert-performance | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| expert-testing | 🟣 opus | 🔵 sonnet | 🟡 haiku |

### Builder 에이전트 (3개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| builder-agent | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| builder-skill | 🟣 opus | 🔵 sonnet | 🟡 haiku |
| builder-plugin | 🟣 opus | 🔵 sonnet | 🟡 haiku |

### Team 에이전트 (5개)

| 에이전트 | High | Medium | Low |
|---------|------|--------|-----|
| team-reader | 🔵 sonnet | 🔵 sonnet | 🔵 sonnet |
| team-coder | 🔵 sonnet | 🔵 sonnet | 🔵 sonnet |
| team-tester | 🔵 sonnet | 🔵 sonnet | 🔵 sonnet |
| team-designer | 🔵 sonnet | 🔵 sonnet | 🔵 sonnet |
| team-validator | 🟡 haiku | 🟡 haiku | 🟡 haiku |

## 배정 원칙

- **항상 Opus**: 보안(expert-security), 전략(manager-strategy) — 높은 추론 능력 필요
- **항상 Haiku**: 품질 검증(manager-quality), Git(manager-git) — 가볍고 빠른 작업
- **플랜에 따라 변동**: 구현 에이전트(backend, frontend, ddd, tdd) — 플랜이 높을수록 Opus

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
