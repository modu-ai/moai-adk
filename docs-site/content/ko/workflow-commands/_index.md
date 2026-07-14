---
title: 워크플로우 명령어
weight: 30
draft: false
---

SPEC 기반 3-Phase 라이프사이클 (plan → run → sync)을 실행하는 명령어 모음입니다.

{{< mascot coding >}}

## 에이전틱 하네스의 중심 — 3-Phase 라이프사이클

MoAI-ADK v3의 핵심 가치 중 하나는 **에이전틱 하네스** (Agentic Harness)입니다. 코드를 직접 쓰는 대신, 에이전트가 잘 일할 환경 — SPEC 문서, 품질 게이트, 피드백 루프 — 을 설계한다는 뜻입니다. 워크플로우 명령어는 이 하네스의 중심축인 **plan → run → sync** 파이프라인을 실행합니다.

각 단계는 전문화된 에이전트가 담당하고, 만든 사람이 검사하지 않도록 **계획과 감사가 분리**되어 있습니다. plan 단계의 산출물은 plan-auditor가 독립 감사하고, sync 단계의 결과물은 sync-auditor가 4차원 (Functionality·Security·Craft·Consistency) 으로 평가합니다. run 단계 진입 직전에는 **구현 착수 승인** (휴먼 게이트)이 항상 사용자에게 돌아옵니다.

```mermaid
flowchart TD
    A["/moai project<br>프로젝트 문서 생성"] --> B["/moai plan<br>SPEC 문서 생성"]
    B --> D["/moai run<br>DDD/TDD 구현"]
    D --> E["/moai sync<br>문서 동기화 및 PR"]
    E -.-> B
    D -.-> B
    F["/moai harness<br>하네스 학습 시스템"] -.-> D
```

## 명령어 요약

| 명령어 | 단계 | 담당 에이전트 | 토큰 예산 | 목적 |
|--------|------|---------------|-----------|------|
| [`/moai project`](./moai-project) | Phase 0 | manager-docs | - | 프로젝트 문서 자동 생성 |
| [`/moai plan`](./moai-plan) | Phase 1 | manager-spec | 30K | SPEC 문서 생성 |
| [`/moai run`](./moai-run) | Phase 2 | manager-develop | 180K | DDD/TDD 방식 구현 |
| [`/moai sync`](./moai-sync) | Phase 3 | manager-docs | 40K | 문서 동기화 및 PR 생성 |
| [`/moai harness`](./moai-harness) | 보조 | builder-harness | - | 하네스 생성 및 학습 라이프사이클 관리 |

단계별 토큰 예산이 다른 것도 v3의 **토크노믹스** (Token Economics) 설계의 일부입니다. 계획은 깊은 추론이 필요하지만 산출물이 작고 (30K), 구현은 코드량이 많아 넉넉한 예산이 필요하며 (180K), 문서 동기화는 그 중간 (40K)입니다. 단계 사이에 `/clear`로 컨텍스트를 비우는 관행도 같은 이유에서 나옵니다 — 이전 단계의 대화를 다음 단계로 끌고 가지 않아야 각 단계가 예산을 온전히 씁니다.

{{< callout type="info" >}}
처음 사용하신다면 `/moai project`부터 시작하세요. 프로젝트 문서가 있어야 이후 단계에서 AI가 프로젝트를 정확히 이해하고 작업할 수 있습니다.

`/moai harness`는 하네스 학습 서브시스템 관리용 보조 명령어입니다 — CLAUDE.md 변경을 모니터링하고 티어 기반 자동 업데이트를 제안합니다.
{{< /callout >}}

## 전체 서브커맨드 (15개)

`/moai` 오케스트레이터는 15개 서브커맨드를 라우팅합니다. 이 섹션(워크플로우)은 SPEC 3-Phase 라이프사이클 명령어를, [유틸리티 명령어](/utility-commands/) 섹션은 자동화·수정 루프·코드 관리·피드백 명령어를 다룹니다.

**워크플로우 명령어 (이 섹션):**

| 서브커맨드 | 목적 |
|-----------|------|
| [`/moai plan`](./moai-plan) | SPEC 문서 생성 |
| [`/moai run`](./moai-run) | DDD/TDD 구현 |
| [`/moai sync`](./moai-sync) | 문서 동기화 및 PR |
| [`/moai project`](./moai-project) | 프로젝트 문서 생성 |
| [`/moai design`](./moai-design) | 디자인 단계 협업 (manager-design D1-D5) |
| [`/moai harness`](./moai-harness) | 하네스 생성 및 학습 라이프사이클 |

**유틸리티 명령어 ([유틸리티 섹션](/utility-commands/)):**

| 서브커맨드 | 목적 |
|-----------|------|
| [`/moai fix`](/utility-commands/moai-fix) | 일회성 자동 수정 |
| [`/moai loop`](/utility-commands/moai-loop) | 반복 수정 루프 |
| [`/moai mx`](/utility-commands/moai-mx) | @MX 코드 주석 |
| [`/moai feedback`](/utility-commands/moai-feedback) | GitHub 이슈 피드백 |
| [`/moai review`](/utility-commands/moai-review) | 다관점 코드 리뷰 (보안·@MX) |
| [`/moai clean`](/utility-commands/moai-clean) | 데드 코드 제거 |
| [`/moai codemaps`](/utility-commands/moai-codemaps) | 아키텍처 코드맵 생성 |
| [`/moai gate`](/utility-commands/moai-gate) | 커밋 전 품질 게이트 |
| [`/moai e2e`](/utility-commands/moai-e2e) | 멀티 플랫폼 E2E 테스트 |
| [`/moai goal`](/utility-commands/moai-goal) | 조건 선언형 자율 루프 |

## 빠른 시작

```bash
# Phase 0: 프로젝트 문서 생성 (최초 1회)
> /moai project

# Phase 1: SPEC 생성
> /moai plan "사용자 인증 기능 구현"
> /clear

# Phase 2: DDD 구현
> /moai run SPEC-AUTH-001
> /clear

# Phase 3: 문서 동기화 및 PR
> /moai sync SPEC-AUTH-001

# 보조: 하네스 학습 관리 (선택)
> /moai harness status
> /moai harness apply
```

자연어로 바로 요청해도 됩니다. `/moai "로그인 버그 고쳐줘"`처럼 서브커맨드 없이 입력하면 **Analyze-First 라우팅**이 의도를 분석해 알맞은 워크플로우로 자동 연결합니다.

## 관련 문서

- [SPEC 기반 개발](/core-concepts/spec-based-dev) - SPEC과 EARS/GEARS 형식 상세 설명
- [DDD 방법론](/core-concepts/ddd) - ANALYZE-PRESERVE-IMPROVE 사이클 상세 설명
- [TRUST 5 품질 시스템](/core-concepts/trust-5) - 품질 게이트 상세 설명
- [하네스 엔지니어링](/core-concepts/harness-engineering) - 하네스 학습 서브시스템 개요
- [빠른 시작](/getting-started/quickstart) - 처음부터 따라하는 튜토리얼
