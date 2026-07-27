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
내장 `Explore`)입니다. **No-Haiku 정책** 아래에서 Haiku는 어디에도 등장하지
않습니다. 멀티턴 에이전틱 행은 모두 Opus가 담당하고 Sonnet은 단발성·입력
지배적 행에만 한정되며, 정책 티어는 각 에이전트가 Opus effort 사다리의 어디에
놓이는지를 제어할 뿐 어떤 모델 클래스를 받는지를 제어하지 않습니다.

## 3단계 정책 개요

| 정책 (profile) | CLI 플래그 | Opus 셀 | Sonnet 셀 | 적합한 용도 |
|---------------|-----------|---------|-----------|-----------|
| **high** | `--model-policy high` | 11개 중 9개 | 11개 중 2개 | 최고 품질, 호출 빈도가 가장 낮은 두 행에 `max` effort |
| **medium** (기본) | `--model-policy medium` | 11개 중 9개 | 11개 중 2개 | 품질과 비용의 균형, 비용/점수 곡선의 무릎 |
| **low** | `--model-policy low` | 11개 중 7개 | 11개 중 4개 | 작업당 최저 비용, 에이전틱 행은 Opus `low`로 내려감 |

> **이름 축**: `llm.yaml`의 `profile` 필드, legacy `performance_tier` 별칭, CLI
> 플래그 `--model-policy`는 모두 동일하게 `high`/`medium`/`low` 세 값을
> 사용하며 1:1로 매핑됩니다 (별도 변환 없음). 기본값은 `medium`입니다. 과거
> 최상위 티어 이름 `max`는 기존 설정이 계속 해석되도록 `high`의 별칭으로 여전히
> **읽히지만**, 저장 시에는 항상 `high`로 기록됩니다 — 마이그레이션 단계는
> 필요하지 않습니다. `performance_tier`는 `profile`이 없을 때만 읽힙니다.
> 사용자 이름 등은 `user.yaml`에 따로 보관됩니다.

> **왜 중요한가요?** 정책을 낮추는 것은 더 이상 약한 모델 클래스로 바꾼다는
> 뜻이 아닙니다. 장기 지평 에이전틱 작업에서는 Opus의 `low` effort가 어떤
> effort의 Sonnet보다도 점수가 높고 **동시에** 작업당 비용도 낮습니다. 청구액을
> 결정하는 것은 토큰당 단가가 아니라 모델이 작업을 끝내기까지 소비한 스텝
> 수이기 때문입니다. 그래서 `low` 정책은 추론 깊이를 낮춰 Opus *안에서*
> 절약하고, 멀티스텝 완료 실패가 적용되지 않는 단발성 행에서만 Sonnet을
> 사용합니다.

## 에이전트별 모델 배정표

아래 33개 셀이 프로필 매트릭스(에이전트 11개 × 프로필 3개)입니다. 각 셀은
리졸버가 spawn 시점에 주입하는 `{model, effort}` 쌍입니다. (오케스트레이터 메인
세션은 spawn되는 에이전트가 아니므로 표에 포함하지 않습니다.)

### Manager Agents (5개)

| 에이전트 | high | medium | low |
|---------|------|--------|-----|
| manager-spec | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| manager-design | opus / high | opus / medium | opus / low |

### Evaluator · Advisor · Builder · Specialist Agents (5개)

| 에이전트 | high | medium | low |
|---------|------|--------|-----|
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |

### Built-in Agent (1개)

| 에이전트 | high | medium | low |
|---------|------|--------|-----|
| Explore | sonnet / low | sonnet / low | sonnet / low |

> `Explore`는 디스크에 에이전트 파일이 없어 frontmatter로 effort를 고정할 수
> 없습니다 — 매트릭스는 `sonnet / low`를 호출 시점 기본값으로 기록하며, 이는
> spawn 프롬프트에 명시됩니다. Agent Teams 정적 계층(정적 role profile)은
> v3.0에서 은퇴했으며, 병렬 작업은 sub-agent 병렬 실행과 동적 워크플로우가
> 대체합니다. `moai cg`의 teammate 런타임(tmux pane)은 그대로 유지됩니다.

> **Haiku 제거 (v3.0)**: 과거 Haiku 슬롯(문서화·MX 태깅·Git 절차)은 더 낮은
> 모델 클래스가 아니라 더 낮은 추론 깊이로 대체되었습니다 — 비용은 모델 교체가
> 아니라 effort 티어링으로 절감합니다.

## 배정 원칙

- **모든 에이전틱 행은 Opus**: `manager-spec`, `manager-develop`, `plan-auditor`, `sync-auditor`, `manager-design`, `builder-harness`, `manager-docs`, `e2e-tester` — 멀티턴 작업은 전부 Opus에 머무릅니다. Opus의 `low`가 어떤 effort의 Sonnet보다 점수가 높으면서 작업당 비용은 더 낮기 때문입니다
- **Sonnet은 단발성 행에만**: `manager-git`의 기계적 작업과 `Explore` 탐색은 입력 지배적인 단일 패스로 끝나므로 멀티스텝 완료 실패가 적용되지 않고, Sonnet의 낮은 입력 단가가 결정적 요인이 됩니다. 이 두 행은 세 프로필 전체에서 고정입니다
- **`max`는 두 셀에 한정**: `manager-develop`와 `super-advisor`, 그리고 `high` 프로필에서만 — 호출 빈도가 가장 낮으면서 한 번의 판단이 하류 비용을 불균형하게 좌우하는 행입니다
- **`xhigh`는 어디에도 쓰이지 않음**: Opus에서 점수는 `high`와 같은데 비용은 49% 더 높습니다
- **`low`는 모델 클래스가 아니라 effort를 낮춤**: 에이전틱 행은 Opus `low`로 이동하고, `manager-docs`와 `e2e-tester`만 추가로 Sonnet으로 폴백합니다

계획을 만든 에이전트가 감사하지 않도록 `plan-auditor`와 `sync-auditor`는
`manager-spec`과 독립된 배정을 유지합니다 — 편향 방지는 셀 값이 아니라 카탈로그
자체의 구조적 속성입니다.

## v3.0 확장: Tier×Phase 선언 축

v3.0에서는 에이전트 단위 배정 위에 **작업 단계(phase)와 SPEC 크기(Tier)**
축이 더해졌습니다. `internal/config/model_routing.go`가 Tier×Phase →
{model, effort} 매트릭스를 선언적으로 관리합니다:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (추론 깊이): low / medium / high / xhigh / max
- **tier** (SPEC 크기): S / M / L
- **phase** (작업 단계): plan / run / sync / mx

에이전트별 model+effort 배정은 단일 프로필 매트릭스가 담당합니다. 활성
프로필(`profile` — `high`/`medium`/`low`)이 매트릭스의 한 열을 선택하며,
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
moai init my-project --model-policy high    # 최고 품질 (2개 행에 max effort)
moai init my-project --model-policy medium  # 균형 (기본값)
moai init my-project --model-policy low     # 작업당 최저 비용
```

`--model-policy`는 `high`/`medium`/`low` 세 값을 받으며 `llm.yaml`의
`performance_tier` 필드에 저장됩니다. 과거 최상위 티어 이름 `max`는 입력값으로
여전히 허용되며 `high`의 별칭으로 처리됩니다.

> 기본 정책은 `medium`입니다 (llm.yaml `performance_tier: "medium"`, CLI `--model-policy medium`에 해당 — 값이 없으면 `medium`으로 해석). GLM 설정은 `settings.local.json`에 격리되어 Git에 커밋되지 않습니다.

## 다음 단계

- [CG 모드](/ko/multi-llm/cg-mode) — Claude + GLM 하이브리드로 비용 절감
- [에이전트 가이드](/ko/advanced/agent-guide) — 에이전트 커스터마이징
- [CLI 레퍼런스](/ko/getting-started/cli) — moai init, moai update 상세
