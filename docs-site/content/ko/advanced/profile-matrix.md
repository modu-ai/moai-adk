---
title: 프로필 매트릭스
weight: 4
draft: false
---

MoAI-ADK는 유지되는 각 에이전트를 하나의 **프로필 매트릭스**를 통해 `{model, effort}` 쌍에 매핑합니다. 활성 **프로필**(`max` / `medium` / `low`)이 매트릭스의 한 열(column)을 선택하고, 그 열의 값이 모든 서브에이전트 spawn에 적용됩니다. 이 단일 3-열 프로필 축은 이전의 `plan_type × tier` 60-셀 매트릭스를 대체합니다(SPEC-MODEL-PROFILE-MATRIX-001).

## 프로필 축

프로필은 세 가지 값을 가집니다:

- `max` — 최고 품질 열. 추론 지점에 Fable을, 설계·하네스·E2E에 Opus를 배치합니다.
- `medium` (기본값) — 균형 열. 추론과 실행에 Opus/high를 배치합니다. 값이 없거나 비어 있으면 `medium`으로 해석됩니다.
- `low` — 경제 열. Opus를 낮은 effort로 배치하고 기계적 작업을 Sonnet으로 돌립니다.

프로필은 `performance_tier`와 별개 필드가 아니라 동일한 축입니다 — `llm.profile`이 우선이고, 없으면 legacy `performance_tier`가 별칭으로 읽힙니다(`high` → `max` 정규화, `max`/`medium`/`low`는 그대로). 리졸버는 이 유효 프로필을 읽어 각 에이전트의 셀을 결정합니다.

## 프로필 설정

```bash
moai init . --profile max              # 초기화 시 설정
moai update --profile low              # 사후 전환
```

현재 값은 `.moai/config/sections/llm.yaml`의 `llm.profile` 필드에서 확인할 수 있습니다. `moai init` 대화형 마법사에서 `high` 답변은 `max`로 정규화됩니다.

## 프로필 매트릭스

10개의 그룹화된 에이전트가 아래 매트릭스에서 `{model, effort}`를 받습니다. `Explore`와 사용자 정의 에이전트는 그룹이 없으므로 `inherit`(부모 세션 모델 상속)로 해석되며 model 주입 대상이 아닙니다. 매트릭스 어디에도 Haiku는 없습니다.

| 에이전트 (그룹) | max | medium (기본) | low |
|---|---|---|---|
| manager-spec (spec_auditors) | fable / medium | opus / high | opus / low |
| plan-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| sync-auditor (spec_auditors) | fable / medium | opus / high | opus / low |
| manager-develop (develop) | fable / low | opus / high | opus / medium |
| super-advisor (advisor) | fable / medium | fable / low | opus / high |
| manager-design (design_harness_e2e) | opus / high | opus / medium | opus / low |
| builder-harness (design_harness_e2e) | opus / high | opus / medium | opus / low |
| e2e-tester (design_harness_e2e) | opus / high | opus / medium | opus / low |
| manager-docs (docs) | sonnet / medium | sonnet / medium | sonnet / medium |
| manager-git (git) | sonnet / low | sonnet / low | sonnet / low |
| Explore (—) | inherit | inherit | inherit |

`docs`와 `git` 행은 프로필과 무관하게 고정됩니다(각각 sonnet/medium, sonnet/low) — 기계적 작업은 프로필이 바뀌어도 모델 클래스를 올리지 않습니다.

## 에이전트 그룹

매트릭스는 에이전트 이름이 아니라 6개 **그룹** 단위로 정의됩니다. 그룹 → 에이전트 멤버십은 다음과 같습니다:

| 그룹 | 에이전트 |
|---|---|
| `spec_auditors` | manager-spec, plan-auditor, sync-auditor |
| `develop` | manager-develop |
| `advisor` | super-advisor |
| `design_harness_e2e` | manager-design, builder-harness, e2e-tester |
| `docs` | manager-docs |
| `git` | manager-git |

`Explore`와 사용자가 추가한 에이전트는 멤버십이 없어 `inherit`로 해석됩니다.

## 리졸버 우선순위

각 에이전트의 유효 `{model, effort}`는 다음 순서로 결정됩니다:

1. `llm.agent_overrides[agent]`가 있으면 그것이 이깁니다.
2. 없으면 활성 프로필의 그룹 셀(config `llm.profiles`)을 사용합니다.
3. config에 셀이 없으면 Go 기본 매트릭스(`template.DefaultProfileMatrix`)의 그룹 셀을 사용합니다.
4. 그룹 멤버십이 없으면 `inherit`(주입 안 함)입니다.

`agent_overrides`는 정규 에이전트 이름을 키로 하며 카탈로그 + enum에 대해 검증됩니다:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: xhigh }
```

**model**과 **effort**의 소비 경로는 다릅니다. 리졸브된 **model**은 오케스트레이터가 spawn 시점에 `Agent(model: <alias>)` 런타임 인자로 주입하는 값입니다(`[1m]`-safe, frontmatter `model:` 필드와 별개). 에이전트 `.md` frontmatter는 `model: inherit`로 유지되며 init/update/web 저장이 이를 변경하지 않습니다. 리졸브된 **effort**는 NAMED 서브에이전트에 대한 *문서화된 의도*입니다 — Agent/Task 도구는 named 서브에이전트에 per-spawn effort 인자를 받지 않으므로, effort는 (a) 에이전트 frontmatter effort 기본값, (b) GLM effort 오버레이, (c) Workflow / `Agent(general-purpose)` 프롬프트 수준 steering을 통해서만 소비됩니다.

## moai model profile

활성 프로필로 리졸브된 에이전트별 model+effort는 읽기 전용 접근자로 확인합니다:

```bash
moai model profile          # 사람용 표
moai model profile --json   # 기계 판독용
```

이 명령은 아무것도 변경하지 않습니다 — 오케스트레이터가 spawn 시 주입할 값을 그대로 노출합니다.

## GLM 백엔드 effort 오버레이

{{< icon warning warn >}} **정직성 고지**: GLM 백엔드 effort 오버레이는 **구현 + 배선 완료** 상태이나, wire 유효성(라이브 유효성)은 실증 예정입니다 — "동작 보장"으로 서술하지 않습니다.

GLM 백엔드(`moai glm` / `moai cg` GLM 패널)에서는 프로필 매트릭스 위에 오버레이가 적용됩니다:

- 모델 슬롯 매핑: `fable` → `glm-5.2` (Fable 슬롯, `ANTHROPIC_DEFAULT_FABLE_MODEL`)
- Claude의 5단 effort를 z.ai가 도달 가능한 3-state로 collapse:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max` → **reasoning-max**
  - (인식 불가 값 → reasoning-max, 과소 추론 방지)
- coding-max override: `manager-develop`은 collapse 결과와 무관하게 **reasoning-max** 강제
- `manager-git`은 low effort → **thinking-off**

z.ai가 Anthropic-compat shim으로 `ANTHROPIC_REASONING_EFFORT` 값을 실제로 소비하는지는 라이브 GLM 세션 아웃바운드 관측이 필요한 실증 과제입니다. 런타임 SSOT는 `internal/template/glm_effort_overlay.go`입니다.

## 다음 단계

- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — DeepSWE 리더보드 근거와 3-티어 정의
- [토크노믹스 개요](/ko/advanced/tokenomics-overview/) — 4-층 토크노믹스 구조의 B층 라우팅
- [모델 정책](/ko/multi-llm/model-policy/) — performance_tier 별칭과 GLM 백엔드 상세
