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

MoAI-ADK v3.0의 에이전트 카탈로그는 **11개** (MoAI 커스텀 10개 + Anthropic
내장 `Explore`)이며, 아래 배정표는 그중 모델 정책이 직접 배정하는 핵심 7개
에이전트를 다룹니다.

## 3단계 정책 개요

| 정책 (performance_tier) | CLI 플래그 | 플랜 | Opus | Sonnet | 적합한 용도 |
|------------------------|-----------|------|------|--------|-----------|
| **max** | `--model-policy max` | Max $200/월 | 5 | 2 | 최고 품질, 최대 처리량 |
| **medium** (기본) | `--model-policy medium` | Max $100/월 | 2 | 5 | 품질과 비용의 균형 |
| **low** | `--model-policy low` | Plus $20/월 | 0 | 7 | 저예산, Opus 미포함 |

> **이름 축**: `llm.yaml`의 `performance_tier` 필드와 CLI 플래그 `--model-policy`는
> 동일하게 `max`/`medium`/`low` 세 값을 사용하며 1:1로 매핑됩니다 (별도 변환
> 없음). 기본값은 `medium`입니다. `--high` 플래그는 `--model-policy max`의 더
> 이상 사용하지 않는 별칭입니다 (한 사이클 하위 호환, `--low`도 마찬가지).
> `performance_tier`는 서브에이전트 모델 배정만 제어하며, 요금제 종류(api /
> subscription)를 결정하는 `plan_type` 필드와는 별개 축입니다. 사용자 이름 등은
> `user.yaml`에 따로 보관됩니다.

> **왜 중요한가요?** Plus $20 플랜은 Opus에 접근할 수 없습니다. `low` 정책을 설정하면 모든 에이전트가 Sonnet만 사용하여 요율 제한 에러를 방지합니다. 상위 플랜은 핵심 에이전트(계획, 감사)에 Opus를 배정하고 일상 작업에는 Sonnet을 사용합니다.

## 에이전트별 모델 배정표

### Manager Agents (4개)

| 에이전트 | max | medium | low |
|---------|-----|--------|-----|
| manager-spec | opus | opus | sonnet |
| manager-develop | opus | sonnet | sonnet |
| manager-docs | sonnet | sonnet | sonnet |
| manager-git | sonnet | sonnet | sonnet |

### Evaluator & Builder Agents (3개)

| 에이전트 | max | medium | low |
|---------|-----|--------|-----|
| plan-auditor | opus | opus | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| builder-harness | opus | sonnet | sonnet |

> Anthropic 내장 `Explore`는 읽기 전용 탐색 에이전트로 별도 배정 없이
> 동작합니다. Agent Teams 정적 계층(정적 role profile)은 v3.0에서
> 은퇴했으며, 병렬 작업은 sub-agent 병렬 실행과 동적 워크플로우가
> 대체합니다. `moai cg`의 teammate 런타임(tmux pane)은 그대로 유지됩니다.

> **Haiku 제거 (v3.0)**: 과거 Haiku 슬롯은 `sonnet`/`effort:low`로
> 대체되었습니다. `manager-git`과 `manager-docs`의 가벼운 작업이 이에
> 해당하며, 모델은 Sonnet이지만 추론 깊이를 낮춰 비용을 절감합니다.

## 배정 원칙

- **항상 Opus**: 계획 감사(plan-auditor), SPEC 작성(manager-spec) — 높은 추론 능력 필요
- **항상 Sonnet/effort:low**: Git(manager-git) — 가볍고 빠른 작업
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
요금제 종류(`plan_type` — `api` 또는 `subscription`) 프로파일이 요금제별
매트릭스를 분리 적용합니다. `plan_type`은 `performance_tier`와 독립된 축으로,
값이 없으면 `subscription`으로 해석됩니다.

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

### CLI 플래그로 직접 설정

```bash
moai init my-project --model-policy max     # 최고 품질 (Opus 중심)
moai init my-project --model-policy medium  # 균형 (기본값)
moai init my-project --model-policy low     # Sonnet만, Opus 미사용
```

`--model-policy`는 `max`/`medium`/`low` 세 값을 받으며 `llm.yaml`의
`performance_tier` 필드에 그대로 저장됩니다. 더 이상 사용하지 않는 `--high`
플래그는 `--model-policy max`의 별칭입니다.

> 기본 정책은 `medium`입니다 (llm.yaml `performance_tier: "medium"`, CLI `--model-policy medium`에 해당 — 값이 없으면 `medium`으로 해석). GLM 설정은 `settings.local.json`에 격리되어 Git에 커밋되지 않습니다.

## 다음 단계

- [CG 모드](/ko/multi-llm/cg-mode) — Claude + GLM 하이브리드로 비용 절감
- [에이전트 가이드](/ko/advanced/agent-guide) — 에이전트 커스터마이징
- [CLI 레퍼런스](/ko/getting-started/cli) — moai init, moai update 상세
