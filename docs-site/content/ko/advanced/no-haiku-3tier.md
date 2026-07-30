---
title: "3-티어 에이전트 아키텍처 (No-Haiku)"
weight: 3
draft: false
---

MoAI-ADK v3.0은 Haiku를 라우팅 모델 세트에서 배제하고, 작업 성격에 맞춘 3-티어 구조로 작업을 분산합니다. 이 설계는 DeepSWE 리더보드의 실측 데이터에 근거합니다. 이 페이지는 왜 Haiku를 배제했는지, 3-티어가 어떻게 구성되는지, 그리고 설계 의도와 구현된 동작을 구분해서 설명합니다.

## 왜 Haiku를 배제했는가

DeepSWE 리더보드의 핵심 발견은 "**약한 모델 + 높은 effort = 가용성의 적**"이라는 점입니다. 약한 모델은 긴 호흡의 과제를 더 싸게 끝내지 못합니다 — 수렴에 실패하면서 스텝과 출력 토큰만 더 태웁니다. `max` effort에서 Sonnet 5는 268스텝에 214k 출력 토큰을 쓰는데, 같은 과제 세트를 Opus 5는 99스텝에 끝냅니다.

아래 수치는 리더보드의 **"All effort levels"** 뷰(113 tasks / 91 repos / 5 languages, mini-swe-agent 하네스)에서 가져왔습니다. effort를 단계별로 보고하므로 운용점 하나가 아니라 각 모델의 비용/점수 곡선 모양에서 티어를 도출할 수 있습니다.

| 모델 | effort | 점수 | $/과제 | 출력 토큰 | 스텝 |
|---|---|---|---|---|---|
| Opus 5 | low | 58% | $1.66 | 20k | 36 |
| Opus 5 | medium | 69% | $3.29 | 37k | 52 |
| Opus 5 | high | 73% | $6.08 | 64k | 73 |
| Opus 5 | xhigh | 73% | $9.07 | 92k | 89 |
| Opus 5 | max | 74% | $11.84 | 118k | 99 |
| Sonnet 5 | low | 31% | $2.19 | 36k | 77 |
| Sonnet 5 | medium | 40% | $4.08 | 57k | 108 |
| Sonnet 5 | high | 48% | $7.43 | 87k | 147 |
| Sonnet 5 | xhigh | 50% | $11.89 | 121k | 186 |
| Sonnet 5 | max | 54% | $26.40 | 214k | 268 |
| Fable 5 | high | 69% | $9.18 | 57k | 59 |
| Fable 5 | max | 70% | $21.63 | 119k | 88 |

MTok당 정가(입력/출력): Opus 5 $5/$25 · Sonnet 5 $2/$10 (도입 가격, 2026-08-31까지, 이후 $3/$15) · Fable 5 $10/$50.

{{< icon warning warn >}} **단가 역전**: Sonnet의 토큰당 단가는 Opus보다 *낮지만*, 비교 가능한 모든 지점에서 과제당 비용은 오히려 더 듭니다 — Opus 5 `low`는 $1.66에 58%, Sonnet 5 `max`는 $26.40에 54%입니다. "싼 모델로 돌리면 쿼터가 절약된다"는 통념은 긴 호흡의 에이전틱 작업에서는 성립하지 않습니다. 청구서를 정하는 것이 단가가 아니라 완주 효율이기 때문입니다.

이 데이터를 놓고 보면, Haiku를 라우팅에 넣어봐야 역량은 늘지 않고 스텝 낭비만 늘어납니다. 대신 Sonnet은 멀티스텝 완주 실패가 걸리지 않는 단발·입력 지배 작업으로 한정합니다.

## 3-티어 정의

작업 성격에 따라 3개 티어로 모델과 effort를 배정합니다.

```mermaid
flowchart TD
    T1["Tier 1 — 단발 Single-shot<br/>Sonnet low<br/>git mechanics · read-only search"]
    T2["Tier 2 — 에이전틱 Agentic<br/>Opus low/medium/high<br/>spec · develop · audit · design · harness"]
    T3["Tier 3 — 피크 Peak<br/>Opus max<br/>develop · advisor (high profile only)"]

    T1 --> T2 --> T3
```

### Tier 1 — 단발 (Single-shot)

{{< icon database >}} 한 번에 끝나고, 반복보다 입력이 비용을 좌우하는 작업입니다. 약한 모델을 비싸게 만드는 원인, 즉 멀티스텝 완주 실패가 여기서는 나타나지 않으므로 Sonnet의 낮은 입력 단가가 실질적인 변수가 됩니다. Sonnet `low` effort로 스텝 수를 최소로 줄입니다. 담당 에이전트는 `manager-git`, `Explore`이며, 이 두 행은 세 프로필 모두에서 고정입니다.

### Tier 2 — 에이전틱 (Agentic)

{{< icon flash >}} 계획, 구현, 감사, 설계, 하네스 생성, 문서화, E2E — 멀티턴 행 전부입니다. Opus `low`가 이미 어떤 effort의 Sonnet보다 점수가 높으면서 과제당 비용은 낮으므로 이 행 전부를 Opus가 맡습니다. 프로필은 각 행을 Opus effort 단계 가운데 어디에 앉힐지 고릅니다. 경제 열은 `low`, 기본 열은 `medium`, 품질 열은 `high`입니다. 담당 에이전트: `manager-spec`, `manager-develop`, `plan-auditor`, `sync-auditor`, `manager-design`, `builder-harness`, `manager-docs`, `e2e-tester`.

### Tier 3 — 피크 (Peak)

{{< icon sparkles >}} `max` effort는 `high` 프로필에서 호출 빈도가 가장 낮은 두 행, 즉 `manager-develop`과 `super-advisor`에만 씁니다. `medium` 위로는 점수 1점당 한계 비용이 가파르게 오르기 때문입니다(`low`→`medium`은 점당 $0.15, `medium`→`high`는 점당 $0.70). 그래서 한 번의 결정이 이후 비용을 크게 좌우하는 자리에만 피크 effort를 배정합니다. `xhigh`는 어디에도 쓰지 않습니다 — Opus에서 `high`와 점수가 같으면서 비용만 49% 더 듭니다.

## DeepSWE 리더보드 근거

effort 단계별 실측에서 나온 결론은 네 가지입니다.

1. **Opus 5는 모든 effort에서 Sonnet 5를 파레토 지배합니다.** Opus `low`(58%, $1.66)가 Sonnet의 다섯 지점을 두 축 모두에서 앞서며, Sonnet `max`(54%, $26.40)도 예외가 아닙니다. "바쁜 에이전트는 싼 모델로 보낸다"는 라우팅 명제는 긴 호흡의 에이전틱 작업에서는 반증됩니다.
2. **원인은 단가가 아니라 완주 효율입니다.** Sonnet은 같은 과제 세트를 끝내는 데 스텝을 약 2.7배 씁니다. 과제를 비싸게 만드는 것은 토큰당 요율이 아니라 그렇게 더 든 스텝과 출력 토큰입니다.
3. **`xhigh`는 Opus에서 순손실입니다.** `high`와 `xhigh` 모두 73%인데, `xhigh`는 비용이 49% 더 들고 스텝도 22% 더 씁니다. Fable에서도 같은 자리에서 천장이 평평해집니다. 무릎을 넘긴 effort는 점수가 아니라 토큰만 더 씁니다.
4. **`medium`이 무릎입니다.** 점수 1점당 한계 비용은 `low`→`medium` $0.15, `medium`→`high` $0.70(4.7배), `xhigh`→`max` $2.77(18.6배)입니다. 기본 프로필이 `manager-develop`을 `medium`에 고정하는 이유가 여기 있습니다.

{{< icon info >}} **한계 고지**: 이 벤치마크가 측정하는 대상은 **코딩** 에이전트입니다. 문서 저작, 감사 판단, SPEC 저작 품질은 직접 측정하지 않았으므로, 해당 행 배치는 관측이 아니라 멀티턴 에이전틱 작업과 비슷하리라는 추론에 기댑니다. 신뢰구간도 함께 봐야 합니다. `medium`(69%±1)과 `high`(73%±2)는 겹치지 않지만 `max`(74%±4)는 `high`와 겹칩니다 — `max`를 거의 호출되지 않는 두 셀로 묶어둔 이유입니다. 모든 기본값은 `llm.agent_overrides`로 에이전트마다 되돌릴 수 있습니다.

{{< icon info >}} **Fable 5에 대하여**: Fable은 코딩 작업에서 모든 effort에 걸쳐 밀립니다. Fable `high`(69%, $9.18)는 Opus `medium`(69%, $3.29)과 같은 점수를 거의 3배 비용에 냅니다. 그래서 어떤 매트릭스 셀에도 넣지 않았습니다. 모델 enum에서는 여전히 유효한 값이고 GLM 백엔드의 Fable 슬롯 배선도 그대로 살아 있습니다. 바뀐 것은 기본값뿐입니다.

## 설계 보고서 vs 구현

{{< icon warning warn >}} **REQ-DA-061 정직성 구분**: 이 페이지에서는 설계 단계와 실제 구현된 동작을 분명히 갈라 둡니다.

**설계 단계** (`.moai/reports/agent-architecture-redesign-v2-20260709.html`) — v2 아키텍처의 설계 의도입니다. 3-티어 모델 정책의 원칙과 DeepSWE 근거를 제시합니다.

**구현된 동작** — 실제 라우팅은 단일 프로필 매트릭스가 수행합니다. 활성 프로필(`high`/`medium`/`low`)이 매트릭스의 한 열을 고르면 리졸버가 각 에이전트의 `{model, effort}`를 정하고, spawn 시점에 model을 런타임 인자로 넣습니다. 자세한 매트릭스는 [프로필 매트릭스](/ko/advanced/profile-matrix/) 페이지를 보세요.

읽는 쪽에서도 설계 의도(이 페이지의 DeepSWE 근거)와 구현된 동작(단일 프로필 매트릭스)을 구분해서 봐야 합니다.

## 하네스 자가 진화와의 연결

3-티어 아키텍처는 하네스 자가 진화의 바탕입니다. 진화 루프(관찰 → 반추 → 승격)가 제 몫을 하려면 관찰 단계의 라우팅 결정부터 알맞은 모델에 알맞은 effort로 이뤄져야 합니다. 자세한 내용은 [하네스 자가 진화](/ko/advanced/self-evolving/) 페이지를 보세요.

## 다음 단계

- [프로필 매트릭스](/ko/advanced/profile-matrix/) — 단일 3-열 per-agent 프로필 매트릭스 (11 에이전트 × 3 프로필 = 33 셀)
- [토크노믹스 개요](/ko/advanced/tokenomics-overview/) — 4-층 토크노믹스 구조의 B층 라우팅
