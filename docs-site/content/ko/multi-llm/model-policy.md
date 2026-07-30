---
title: 모델 정책
weight: 30
draft: false
---

## 모델 정책이란?

모델 정책은 MoAI-ADK 토크노믹스의 기본 틀입니다. "모든 작업에 최고 모델"을
쓰는 대신, 계획·감사처럼 추론이 무거운 일과 문서화·Git처럼 가벼운 일을 갈라
에이전트마다 알맞은 모델을 선언적으로 배정합니다. 덕분에 Claude Code 구독 플랜
안에서 품질을 최대한 끌어올리면서도 요율 제한 에러를 피할 수 있습니다.

MoAI-ADK v3.0의 에이전트 카탈로그는 **11개**입니다(MoAI 커스텀 10개 + Anthropic
내장 `Explore`). **No-Haiku 정책**에 따라 Haiku는 어디에도 쓰지 않습니다.
멀티턴 에이전틱 행은 모두 Opus가 맡고, Sonnet은 단발성·입력 지배적 행에만
씁니다. 정책 티어가 정하는 것은 각 에이전트가 Opus effort 사다리의 어디에
놓이느냐이지, 어떤 모델 클래스를 받느냐가 아닙니다.

## 3단계 정책 개요

| 정책 (profile) | CLI 플래그 | Opus 셀 | Sonnet 셀 | 적합한 용도 |
|---------------|-----------|---------|-----------|-----------|
| **high** | `--model-policy high` | 11개 중 9개 | 11개 중 2개 | 최고 품질, 호출 빈도가 가장 낮은 두 행에 `max` effort |
| **medium** (기본) | `--model-policy medium` | 11개 중 9개 | 11개 중 2개 | 품질과 비용의 균형, 비용/점수 곡선의 무릎 |
| **low** | `--model-policy low` | 11개 중 7개 | 11개 중 4개 | 작업당 최저 비용, 에이전틱 행은 Opus `low`로 내려감 |

> **이름 정리**: `llm.yaml`의 `profile` 필드, legacy `performance_tier` 별칭, CLI
> 플래그 `--model-policy`는 모두 `high`/`medium`/`low` 세 값을 그대로 쓰며 1:1로
> 대응합니다(별도 변환 없음). 기본값은 `medium`입니다. 예전 최상위 티어 이름인
> `max`는 기존 설정이 계속 해석되도록 지금도 `high`의 별칭으로 **읽히지만**,
> 저장할 때는 항상 `high`로 기록됩니다. 따로 마이그레이션할 일은 없습니다.
> `performance_tier`는 `profile`이 없을 때만 읽습니다. 사용자 이름 같은 값은
> `user.yaml`에 따로 둡니다.

> **왜 중요한가요?** 정책을 낮춘다는 것이 더 약한 모델 클래스로 갈아탄다는
> 뜻은 아닙니다. 호흡이 긴 에이전틱 작업에서는 Opus의 `low` effort가 어떤
> effort의 Sonnet보다도 점수가 높고, **동시에** 작업당 비용도 쌉니다. 청구액을
> 가르는 것은 토큰당 단가가 아니라 모델이 작업을 끝낼 때까지 밟은 스텝 수이기
> 때문입니다. 그래서 `low` 정책은 추론 깊이를 낮춰 Opus *안에서* 아끼고,
> 멀티스텝 완료 실패가 문제되지 않는 단발성 행에서만 Sonnet을 씁니다.

## 에이전트별 모델 배정표

아래 33개 셀이 프로필 매트릭스(에이전트 11개 × 프로필 3개)입니다. 각 셀에는
리졸버가 spawn 시점에 주입하는 `{model, effort}` 쌍이 들어 있습니다.
오케스트레이터 메인 세션은 spawn되는 에이전트가 아니라서 표에서 뺐습니다.

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
> 없습니다. 대신 매트릭스가 `sonnet / low`를 호출 시점 기본값으로 기록하고, 이
> 값은 spawn 프롬프트에 그대로 적힙니다. Agent Teams 정적 계층(정적 role
> profile)은 v3.0에서 물러났고, 그 자리는 sub-agent 병렬 실행과 동적 워크플로우가
> 채웁니다. `moai cg`의 teammate 런타임(tmux pane)은 그대로 남아 있습니다.

> **Haiku 제거 (v3.0)**: 예전 Haiku 슬롯(문서화·MX 태깅·Git 절차)은 더 낮은
> 모델 클래스가 아니라 더 낮은 추론 깊이로 바뀌었습니다. 비용은 모델을 갈아
> 끼워서가 아니라 effort를 층층이 나눠서 줄입니다.

## 배정 원칙

- **모든 에이전틱 행은 Opus**: `manager-spec`, `manager-develop`, `plan-auditor`, `sync-auditor`, `manager-design`, `builder-harness`, `manager-docs`, `e2e-tester` 등 멀티턴 작업은 전부 Opus에 남깁니다. Opus의 `low`가 어떤 effort의 Sonnet보다 점수는 높고 작업당 비용은 싸기 때문입니다
- **Sonnet은 단발성 행에만**: `manager-git`의 기계적 작업과 `Explore` 탐색은 입력이 대부분인 단일 패스로 끝나 멀티스텝 완료 실패를 걱정할 일이 없고, 그러면 Sonnet의 싼 입력 단가가 결정적입니다. 이 두 행은 세 프로필 모두에서 고정입니다
- **`max`는 두 셀에만**: `high` 프로필의 `manager-develop`와 `super-advisor`뿐입니다. 호출 빈도는 가장 낮으면서 한 번의 판단이 이후 비용을 크게 좌우하는 행입니다
- **`xhigh`는 어디에도 쓰지 않음**: Opus에서는 점수가 `high`와 같은데 비용만 49% 더 듭니다
- **`low`는 모델 클래스가 아니라 effort를 낮춤**: 에이전틱 행은 Opus `low`로 내려가고, `manager-docs`와 `e2e-tester`만 Sonnet까지 물러납니다

계획을 세운 에이전트가 자기 계획을 감사하지 않도록 `plan-auditor`와
`sync-auditor`는 `manager-spec`과 따로 배정합니다. 편향을 막는 힘은 셀 값이
아니라 카탈로그 구조 자체에서 나옵니다.

## v3.0 확장: Tier×Phase 선언 매트릭스

v3.0에서는 에이전트 단위 배정 위에 **작업 단계(phase)와 SPEC 크기(Tier)** 두
차원을 더 얹었습니다. Tier×Phase → {model, effort} 매트릭스는
`internal/config/model_routing.go`가 선언적으로 관리합니다:

- **model**: inherit / sonnet / opus / glm / fable
- **effort** (추론 깊이): low / medium / high / xhigh / max
- **tier** (SPEC 크기): S / M / L
- **phase** (작업 단계): plan / run / sync / mx

에이전트별 model+effort 배정은 프로필 매트릭스 하나가 맡습니다. 활성
프로필(`profile`, 값은 `high`/`medium`/`low`)이 매트릭스에서 열 하나를
고릅니다. `profile`이 없으면 legacy `performance_tier`를 별칭으로 읽고, 그마저
없으면 `medium`으로 봅니다. 에이전트별 매핑은
[프로필 매트릭스](/ko/advanced/profile-matrix/) 페이지에 자세히 정리했습니다.

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

`--model-policy`는 `high`/`medium`/`low` 세 값을 받고, 결과는 `llm.yaml`의
`performance_tier` 필드에 저장됩니다. 예전 최상위 티어 이름 `max`도 입력값으로는
여전히 받아 주며, `high`의 별칭으로 처리합니다.

> 기본 정책은 `medium`입니다(llm.yaml `performance_tier: "medium"`, CLI `--model-policy medium`에 해당하며 값이 없으면 `medium`으로 봅니다). GLM 설정은 `settings.local.json`에 따로 두므로 Git에 커밋되지 않습니다.

## 다음 단계

- [CG 모드](/ko/multi-llm/cg-mode) — Claude + GLM 하이브리드로 비용 절감
- [에이전트 가이드](/ko/advanced/agent-guide) — 에이전트 커스터마이징
- [CLI 레퍼런스](/ko/getting-started/cli) — moai init, moai update 상세
