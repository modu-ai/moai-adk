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
내장 `Explore`)입니다. **No-Haiku 정책** 아래에서 모든 워커 에이전트는 티어와
무관하게 **Sonnet 5로 고정**되며, 정책 티어가 제어하는 축은 두 가지뿐입니다 —
(a) Opus를 어디에 배치할지, (b) Sonnet의 추론 깊이(effort)를 얼마나 낮출지.

## 3단계 정책 개요

| 정책 (performance_tier) | CLI 플래그 | 플랜 | Opus 배치 | 워커 | 적합한 용도 |
|------------------------|-----------|------|-----------|------|-----------|
| **max** | `--model-policy max` | Max $200/월 | 5개 지점 | Sonnet 고정 | 최고 품질, 최대 처리량 |
| **medium** (기본) | `--model-policy medium` | Max $100/월 | 2개 지점 (on-demand) | Sonnet 고정 | 품질과 비용의 균형 |
| **low** | `--model-policy low` | Plus $20/월 | 없음 (0) | Sonnet 고정 | 저예산, Opus 미포함 |

> **이름 축**: `llm.yaml`의 `performance_tier` 필드와 CLI 플래그 `--model-policy`는
> 동일하게 `max`/`medium`/`low` 세 값을 사용하며 1:1로 매핑됩니다 (별도 변환
> 없음). 기본값은 `medium`입니다. `--high` 플래그는 `--model-policy max`의 더
> 이상 사용하지 않는 별칭입니다 (한 사이클 하위 호환, `--low`도 마찬가지).
> `performance_tier`는 `profile`(프로필 매트릭스 열)의 legacy 별칭 필드로,
> `profile`이 없을 때만 읽히며 `high`→`max`로 정규화됩니다. 두 필드는 같은
> `max`/`medium`/`low` 축입니다. 사용자 이름 등은 `user.yaml`에 따로 보관됩니다.

> **왜 중요한가요?** Plus $20 플랜은 Opus에 접근할 수 없습니다. `low` 정책을 설정하면 모든 에이전트가 Sonnet만 사용하여 요율 제한 에러를 방지합니다. 상위 플랜은 핵심 지점(계획 작성, 감사, 자문)에만 Opus를 배정하고 나머지는 Sonnet을 사용합니다.

## 에이전트별 모델 배정표

모든 워커 에이전트는 Sonnet 5로 고정되며, Opus는 아래 특정 지점에만
배치됩니다. (오케스트레이터 메인 세션도 `max`에서 Opus로 동작하지만, 이는
spawn되는 에이전트가 아니므로 표에는 포함하지 않습니다.)

### Manager Agents (5개)

| 에이전트 | max | medium | low |
|---------|-----|--------|-----|
| manager-spec (plan) | opus | opus (Tier L만) | sonnet |
| manager-develop | sonnet | sonnet | sonnet |
| manager-docs | sonnet | sonnet | sonnet |
| manager-git | sonnet | sonnet | sonnet |
| manager-design | sonnet | sonnet | sonnet |

### Evaluator · Advisor · Builder Agents (4개)

| 에이전트 | max | medium | low |
|---------|-----|--------|-----|
| plan-auditor | opus | sonnet | sonnet |
| sync-auditor | opus | sonnet | sonnet |
| super-advisor | opus | opus | sonnet |
| builder-harness | sonnet | sonnet | sonnet |

> Anthropic 내장 `Explore`는 읽기 전용 탐색 에이전트로 별도 배정 없이
> 동작합니다. Agent Teams 정적 계층(정적 role profile)은 v3.0에서
> 은퇴했으며, 병렬 작업은 sub-agent 병렬 실행과 동적 워크플로우가
> 대체합니다. `moai cg`의 teammate 런타임(tmux pane)은 그대로 유지됩니다.

> **Haiku 제거 (v3.0)**: 과거 Haiku 슬롯(문서화·MX 태깅·Git 절차)은
> `sonnet`/`effort:low`로 대체되었습니다. 모델은 Sonnet이지만 추론 깊이를
> 낮춰 비용을 절감하는 방식이며, 모델 클래스를 낮춘 것이 아닙니다.

## 배정 원칙

- **모든 워커는 Sonnet 고정**: manager-develop, manager-docs, manager-git, manager-design, builder-harness — 티어는 Opus 배치 위치와 Sonnet effort 조정 폭만 제어
- **max에서 Opus 배치 (5개 지점)**: 오케스트레이터, super-advisor, manager-spec(plan), plan-auditor, sync-auditor — 높은 추론 능력이 필요한 곳
- **medium은 Opus 최소화 (2개 지점, on-demand)**: super-advisor와 Tier L 계획(manager-spec)에만 Opus, 나머지는 Sonnet
- **low는 Opus 0**: 자문(super-advisor)까지 포함해 전부 Sonnet, effort 티어링으로만 조절

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

에이전트별 model+effort 배정은 단일 프로필 매트릭스가 담당합니다. 활성
프로필(`profile` — `max`/`medium`/`low`)이 매트릭스의 한 열을 선택하며,
`profile`이 없으면 legacy `performance_tier`가 별칭으로 읽히고, 그마저 없으면
`medium`으로 해석됩니다. 상세한 에이전트별 매핑은
[프로필 매트릭스](/ko/advanced/profile-matrix/) 페이지를 참조하세요.

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
