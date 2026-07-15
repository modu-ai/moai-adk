---
title: 토크노믹스 개요
weight: 1
draft: false
---

토크노믹스(Token Economics)는 MoAI-ADK v3.0의 첫 번째 기둥입니다. 토큰 단가가 낮아져도 에이전틱 개발은 토큰을 대량으로 소비하므로, 비용을 결정하는 것은 모델 가격이 아니라 토큰 운용 방식입니다. 이 페이지는 토크노믹스의 전체 구조를 개관하고, 각 하위 주제의 심화 페이지로 연결합니다.

## 왜 토크노믹스인가

에이전트가 여러 개 돌고, 컨텍스트가 길어지고, 추론이 깊어질수록 단일 세션의 토큰 소비는 급격히 증가합니다. 토큰 단가 하락이 토큰 사용량 증가를 따라가지 못하는 상황에서, 하네스가 토큰을 어떻게 측정하고 라우팅하고 다이어트하고 방어하는지가 비용 경쟁력의 핵심 축이 됩니다.

MoAI-ADK의 답은 세 가지입니다.

1. **작업마다 맞는 모델과 추론 깊이를 배정한다** — 계획은 깊게, 구현은 싸게, 검증은 독립적으로.
2. **컨텍스트를 다이어트한다** — 항시 로드되는 지침을 최소화하고, 프롬프트 캐시 적중률을 측정한다.
3. **예산을 시스템이 지킨다** — 토큰 사용을 추적하고, 임계 초과 전에 안전하게 멈춘다.

## 세 기둥 서사

v3.0의 제품 차별화는 세 기둥으로 구성됩니다. 토크노믹스는 그 첫 번째 기둥이며, 나머지 두 기둥과 긴밀하게 연결됩니다.

{{< icon target >}} **토크노믹스** (이 페이지) — 측정하고, 라우팅하고, 다이어트하고, 방어한다.

{{< icon rotate >}} **자율 연속 루프** — 언제 멈추고 언제 계속할 것인가. [자율 연속 루프](/ko/advanced/autonomous-loops/) 페이지에서 다룹니다.

{{< icon database >}} **에이전틱 하네스** — 어떤 에이전트가, 어떤 티어로, 어떻게 진화하는가. [3-티어 아키텍처](/ko/advanced/no-haiku-3tier/), [plan_type 티어 프로필](/ko/advanced/plan-type-profiles/), [하네스 자가 진화](/ko/advanced/self-evolving/) 페이지에서 다룹니다.

## 4-층 토크노믹스 구조

토크노믹스는 네 개의 층으로 구성됩니다. 각 층은 독립적으로 작동하면서도 서로 보완합니다.

```mermaid
flowchart TD
    A["A층 — 계측 Metering<br/>per-SPEC 토큰 회계"]
    B["B층 — 라우팅 Routing<br/>Tier × Phase 선언적 모델/effort"]
    C["C층 — 검증 다이어트 Verify-diet<br/>verbatim 증거는 파일, 컨텍스트엔 요약"]
    D["D층 — 예산 방어 Budget defense<br/>90% hard-limit graceful stop"]

    A --> B
    B --> C
    C --> D
```

### A층 — 계측 (Metering)

{{< icon database >}} 모든 에이전트 호출의 토큰 사용량을 per-SPEC 단위로 회계합니다. `moai spec audit` 출력의 토큰 컬럼과 progress.md의 토큰 회계 섹션이 이 층의 산출물입니다. 무엇이 토큰을 소비했는지 모르면 최적화할 수 없습니다.

### B층 — 라우팅 (Routing)

{{< icon package >}} 작업 단계(phase: plan / run / sync)와 SPEC 크기(Tier S / M / L)에 따라 모델과 추론 깊이(effort)를 선언적으로 배정합니다. 깊은 추론이 필요한 계획 단계에는 고추론 모델을, 기계적 반복이 많은 구현 단계에는 가벼운 모델을 배정하여 비용 대비 품질을 극대화합니다. 상세한 60-셀 프로필 매트릭스는 [plan_type 티어 프로필](/ko/advanced/plan-type-profiles/) 페이지를 참조하세요.

### C층 — 검증 다이어트 (Verify-diet)

{{< icon wrench >}} 검증 명령의 장문 출력을 디스크 파일로 리다이렉트하고, 컨텍스트에는 exit code와 bounded tail(최대 50줄)만 남깁니다. 이 파일-리다이렉트 계약(file-redirect contract)은 검증 증거의 무결성을 유지하면서 컨텍스트 소비를 줄입니다. 상세한 메커니즘은 [토큰 예산 관리와 우아한 중단](/ko/advanced/token-budget/) 페이지를 참조하세요.

### D층 — 예산 방어 (Budget defense)

{{< icon warning >}} 에이전트별 토큰 사용량이 hard-limit(기본 90%)에 도달하면 우아한 중단(graceful abort)을 수행합니다. 진행 상태를 progress.md에 저장하고, 붙여넣기 가능한 핸드오프 메시지(paste-ready resume)를 발행하며, 자동 `/clear`는 절대 하지 않습니다. 상세한 절차는 [토큰 예산 관리와 우아한 중단](/ko/advanced/token-budget/) 페이지를 참조하세요.

## 모델 티어 라우팅

B층의 라우팅을 구체화하는 것이 모델 티어 정책입니다. MoAI-ADK v3.0은 Haiku를 라우팅 모델 세트에서 배제하고, 3-티어 구조(Sonnet / Opus / Fable)로 작업을 분산합니다. 이 설계의 근거와 ApplyTierProfile 구현은 다음 두 페이지에서 다룹니다.

- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — 왜 Haiku를 배제했는가, DeepSWE 리더보드 근거
- [plan_type 티어 프로필](/ko/advanced/plan-type-profiles/) — api vs 구독 요금제별 60-셀 프로필 매트릭스

## CG 모드 (비용 최적화)

`moai cg`는 Claude 리더와 GLM 워커를 결합한 하이브리드 모드입니다. 전략, 계획, 감사는 Claude가 담당하고, 대량 구현 작업은 GLM이 담당합니다. 구현 중심 작업에서 60-70% 비용 절감 효과가 있습니다.

GLM-5.2는 1M 컨텍스트 단일 모델로 입력 $2 / 출력 $8 (1M 토큰당) 단가이며, z.ai 암시적 프롬프트 캐싱이 자동 적용됩니다. CG 모드와 GLM 단독 세션(`moai glm`)에 대한 상세는 다중 LLM 섹션을 참조하세요.

## 검증된 사실과 로드맵

이 페이지에 기재된 내용 중 구현 상태를 명확히 구분합니다.

{{< icon check ok >}} **구현 완료 (배포 중)** — 4-층 구조(A/B/C/D) 전 층, 3-티어 모델 정책(ApplyTierProfile), CG 모드, 검증 다이어트 파일-리다이렉트 계약, 우아한 중단 메커니즘.

{{< icon clock >}} **설계 단계 (로드맵)** — GLM 백엔드 effort 오버레이의 wire 유효성은 라이브 GLM 세션 아웃바운드 관측이 필요한 실증 과제입니다. plan_type 티어 프로필 페이지에서 이 구분을 명시합니다.

## 다음 단계

- [토큰 예산 관리와 우아한 중단](/ko/advanced/token-budget/) — D층 심화 (모델별 임계치, paste-ready resume 구조)
- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — 하네스 아키텍처 기초
- [plan_type 티어 프로필](/ko/advanced/plan-type-profiles/) — 60-셀 프로필 매트릭스
