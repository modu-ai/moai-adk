---
title: 프로필 매트릭스
weight: 4
draft: false
---

MoAI-ADK는 유지되는 에이전트 11개를 하나의 **프로필 매트릭스**를 통해 각각의 `{model, effort}` 쌍에 매핑합니다. 활성 **프로필**(`high` / `medium` / `low`)이 매트릭스의 한 열(column)을 선택하고, 그 열의 값이 모든 서브에이전트 spawn에 적용됩니다. 매트릭스는 에이전트 이름 단위의 **33셀**(에이전트 11개 × 프로필 3개)이며, 이전의 그룹 추상화와 `plan_type × tier` 축을 모두 대체합니다.

## 프로필 축

프로필은 세 가지 값을 가집니다:

- `high` — 품질 우선 열. Opus 5가 모든 멀티턴 에이전틱 행을 담당하며, `max`는 호출 빈도가 가장 낮은 두 행(`manager-develop`, `super-advisor`)에만 배정됩니다. `xhigh`는 어떤 셀에도 없습니다 — Opus 5에서는 `high`와 같은 점수를 내면서 비용만 뚜렷하게 더 듭니다.
- `medium` (기본값) — 균형 열이자 나머지 매트릭스가 파생되는 기준점. `manager-develop`이 Opus 5 `medium`에 놓이며, 이 지점이 비용/점수 곡선의 무릎(knee)입니다. 값이 없거나 비어 있으면 `medium`으로 해석됩니다.
- `low` — 경제 열. Opus 5의 `low`는 어떤 effort의 Sonnet 5보다도 점수가 높으면서 **동시에** 과제당 비용이 낮으므로, 모든 에이전틱 행에 Opus를 유지합니다. Sonnet은 단발(single-shot)·입력 지배 행에만 등장합니다.

`max`는 `high`의 **읽기 전용 별칭**입니다. 기존 설정의 `profile: max`는 그대로 `high`로 해석되며, 저장 시에는 항상 정규 이름 `high`로 기록됩니다. 마이그레이션 작업은 필요하지 않습니다.

프로필은 `performance_tier`와 별개 필드가 아니라 동일한 축입니다 — `llm.profile`이 우선이고, 없으면 legacy `performance_tier`가 별칭으로 읽힙니다. 두 필드 모두 `high`/`medium`/`low` 어휘를 공유합니다. 리졸버는 이 유효 프로필을 읽어 각 에이전트의 셀을 결정합니다.

## 프로필 설정

```bash
moai init . --profile high             # 초기화 시 설정
moai update --profile low              # 사후 전환
```

허용 값은 `high` / `medium` / `low`이며, legacy `max`도 입력으로 받아 `high`로 정규화합니다. 현재 값은 `.moai/config/sections/llm.yaml`의 `llm.profile` 필드에서 확인할 수 있습니다.

## 프로필 매트릭스

유지되는 에이전트 11개가 아래 매트릭스에서 각자의 `{model, effort}`를 직접 받습니다. 사용자가 추가한 에이전트만 `inherit`(부모 세션 모델 상속)로 해석되어 model 주입 대상에서 제외됩니다. 매트릭스 어디에도 Haiku는 없습니다.

| 에이전트 | high | medium (기본) | low |
|---|---|---|---|
| manager-spec | opus / high | opus / medium | opus / low |
| plan-auditor | opus / high | opus / medium | opus / low |
| sync-auditor | opus / high | opus / medium | opus / low |
| manager-develop | opus / max | opus / medium | opus / low |
| super-advisor | opus / max | opus / high | opus / medium |
| manager-design | opus / high | opus / medium | opus / low |
| builder-harness | opus / high | opus / medium | opus / low |
| e2e-tester | opus / medium | opus / low | sonnet / low |
| manager-docs | opus / medium | opus / low | sonnet / low |
| manager-git | sonnet / low | sonnet / low | sonnet / low |
| Explore | sonnet / low | sonnet / low | sonnet / low |

33개 셀의 모델 분포는 Opus 25 / Sonnet 8입니다. Fable은 어떤 셀에도 없으며, `xhigh`를 쓰는 셀도 없습니다.

`manager-git`과 `Explore` 행은 프로필과 무관하게 `sonnet / low`로 고정됩니다 — 기계적 작업과 읽기 전용 탐색은 프로필이 올라가도 모델 클래스를 올리지 않습니다.

각 행은 단조(monotone)입니다: `high` ≥ `medium` ≥ `low`. 프로필을 낮추면 어떤 에이전트도 이전보다 강한 조합을 받지 않습니다.

### 왜 이 셀 배치인가

이 셀들은 토큰 단가가 아니라, 점수·과제당 비용·출력 토큰·에이전트 스텝을 **effort 단계마다** 보고하는 장기 호라이즌 코딩 에이전트 벤치마크에서 도출했습니다. 세 가지 실측이 배치를 이끕니다:

- **Opus는 모든 effort에서 Sonnet을 지배합니다.** Opus 5 `low`(58%, $1.66/과제, 36스텝)는 어떤 단계의 Sonnet 5보다도 점수가 높고 과제당 비용이 낮으며, 여기에는 Sonnet 5 `max`(54%, $26.40/과제, 268스텝)도 포함됩니다. 과제당 비용을 결정하는 것은 완주 효율 — 과제를 끝내는 데 쓴 스텝과 출력 토큰 — 이지 토큰당 단가가 아닙니다. 따라서 Sonnet은 멀티스텝 완주가 적용되지 않는 곳, 즉 단발·입력 지배 행(`Explore` 검색, `manager-git` 기계 작업)에만 남습니다. 그곳에서는 낮은 입력 단가가 실질 변수이기 때문입니다.
- **`xhigh`는 Opus에서 완전히 열등합니다.** `high`는 $6.08에 73%를, `xhigh`는 같은 73%를 $9.07에 냅니다 — 이득 없이 비용 +49%, 스텝 +22%. 매트릭스에서 퇴출했습니다(6셀 → 0). `max`는 호출 빈도가 가장 낮은 두 셀에만 살아남습니다.
- **`medium`이 곡선의 무릎입니다.** 그 위로는 점수 1점당 한계 비용이 몇 배로 뜁니다: `low`→`medium`은 점당 $0.15, `medium`→`high`는 점당 $0.70(4.7배). `manager-develop`을 `medium`에 두어 기본 열의 기준점으로 삼은 이유가 이것입니다.

{{< icon warning warn >}} **근거의 적용 범위**: 이 벤치마크가 측정하는 것은 *코딩* 에이전트입니다. 문서 저작, 감사 판단, SPEC 저작 품질은 **직접 측정되지 않았으며**, 해당 행 배치는 멀티턴 에이전틱 작업과의 유사성 추론에 기댑니다. 어떤 행이든 `llm.agent_overrides`로 에이전트별 되돌리기가 가능합니다.

Anthropic 내장 `Explore`는 더 이상 `inherit`이 아니라 자기 셀(`sonnet / low`)로 해석됩니다. `inherit` 센티널은 이제 사용자가 추가한 에이전트에만 남습니다.

## 하네스 스페셜리스트 model + effort

`/moai:harness`가 생성하는 스페셜리스트는 **모델이 `opus`로 통일**되고 **effort로만 차별화**됩니다. 하네스 에이전트는 사용자 소유의 지속적 스페셜리스트이며, 이들을 가르는 축은 모델 티어가 아니라 추론 깊이이기 때문입니다. 모든 비-Haiku 모델이 1M 컨텍스트를 갖게 되어 모델을 고정해도 컨텍스트 손실이 없습니다.

effort는 각 목적 클래스가 대응하는 유지 에이전트 행에서 빌려옵니다:

| 목적 클래스 | effort 출처 행 | high | medium | low |
|---|---|---|---|---|
| `read-only-extract` | Explore | opus / low | opus / low | opus / low |
| `mechanical-transform` | manager-git | opus / low | opus / low | opus / low |
| `synthesize` | manager-docs | opus / medium | opus / low | opus / low |
| `research` | plan-auditor | opus / high | opus / medium | opus / low |
| `verify-judge` | sync-auditor | opus / high | opus / medium | opus / low |
| `implement` | manager-develop | opus / max | opus / medium | opus / low |
| `design-architecture` | manager-design | opus / high | opus / medium | opus / low |

`llm.harness_agents[프로필][클래스].effort`로 클래스별 effort를 덮어쓸 수 있습니다. 모델은 어떤 경로로도 바뀌지 않습니다. 인식되지 않는 클래스는 `implement`로 폴백합니다.

## 리졸버 우선순위

각 에이전트의 유효 `{model, effort}`는 다음 순서로 결정됩니다:

1. `llm.agent_overrides[agent]`가 있으면 그것이 이깁니다.
2. 없으면 활성 프로필의 에이전트 셀(config `llm.profiles`)을 사용합니다.
3. config에 셀이 없으면 Go 기본 매트릭스(`template.DefaultProfileMatrix`)의 에이전트 셀을 사용합니다.
4. 매트릭스에 없는 에이전트(사용자 추가)는 `inherit`(주입 안 함)입니다.

`agent_overrides`는 정규 에이전트 이름을 키로 하며 카탈로그 + enum에 대해 검증됩니다:

```yaml
llm:
  agent_overrides:
    manager-develop: { model: opus, effort: high }
```

enum은 여전히 모델로 `fable`을, effort로 `xhigh`를 받습니다 — 기본 매트릭스에서 빠졌을 뿐 어휘에서 제거된 것은 아니므로, override로는 지금도 둘 중 어느 쪽이든 선택할 수 있습니다.

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

- 모델 슬롯 매핑: `fable` → `glm-5.2` (Fable 슬롯, `ANTHROPIC_DEFAULT_FABLE_MODEL`). 이 슬롯은 프로필 매트릭스와 무관한 GLM 환경 바인딩이며, 매트릭스의 어떤 셀도 Fable을 선택하지 않지만 배선은 그대로 유지됩니다.
- Claude의 5단 effort를 z.ai가 도달 가능한 3-state로 collapse:
  - `low` → **thinking-off**
  - `medium` / `high` → **reasoning-high**
  - `xhigh` / `max`(legacy effort 값) → **reasoning-max**
  - (인식 불가 값 → reasoning-max, 과소 추론 방지)
- coding-max override: `manager-develop`은 collapse 결과와 무관하게 **reasoning-max** 강제
- `manager-git`은 low effort → **thinking-off**

z.ai가 Anthropic-compat shim으로 `ANTHROPIC_REASONING_EFFORT` 값을 실제로 소비하는지는 라이브 GLM 세션 아웃바운드 관측이 필요한 실증 과제입니다. 런타임 SSOT는 `internal/template/glm_effort_overlay.go`입니다.

## 다음 단계

- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — DeepSWE 리더보드 근거와 3-티어 정의
- [토크노믹스 개요](/ko/advanced/tokenomics-overview/) — 4-층 토크노믹스 구조의 B층 라우팅
- [모델 정책](/ko/multi-llm/model-policy/) — performance_tier 별칭과 GLM 백엔드 상세
