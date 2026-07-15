---
title: /moai goal
weight: 25
draft: false
---

완료 조건을 선언하면 세션이 그 조건을 충족할 때까지 스스로 일하는 **조건 선언형 자율 루프** 명령어입니다. `/moai goal "<조건>"`으로 완료 조건을 arm하면, 매 턴 종료 시 `stop-goal` Stop 훅이 조건 충족 여부를 평가하여 충족될 때까지 다음 턴을 자동으로 시작합니다.

{{< callout type="info" >}}
**한 줄 요약**: `/moai goal`은 "끝 상태를 선언하는 범용 루프" 입니다. `/moai loop`가 "진단 도구가 찾은 문제를 다 없앨 때까지"라는 조건이 미리 정해진 프리셋이라면, `/moai goal`은 완료 조건을 **직접 선언**하는 범용 엔진입니다.
{{< /callout >}}

{{< callout type="info" >}}
**프로그래매틱 명령어**: 네이티브 Claude Code `/goal`은 사용자만 입력할 수 있는 (HUMAN-ONLY) TUI 커맨드입니다. `/moai goal`은 동일한 의미를 **파이프라인에서 프로그래매틱하게** 구현한 MoAI 소유 명령어로, `moai` 스킬 라우팅과 `moai goal` CLI를 통해 진입합니다.
{{< /callout >}}

## 개요

에이전트에게 "이 조건이 만족될 때까지 알아서 계속 일해줘"라고 시키고 싶을 때 사용합니다. 조건은 두 종류를 섞어 쓸 수 있습니다.

- **기계적 조건 (mechanical)**: 셸 명령어로 검증되는 조건. 예: `go test ./... exits 0`. 명령을 실행하고 종료 코드를 관찰합니다.
- **모델 평가 조건 (model-evaluated)**: 트랜스크립트에 대한 판단으로 검증되는 조건. 예: `모든 AC 행이 PASS로 기록됨`. 세션이 지금까지 남긴 내용을 근거로 평가합니다.

이 루프가 v3의 두 번째 기둥, **에이전틱 루프 엔지니어링**의 범용 엔진입니다. goal 상태는 `.moai/state/goal/<session-id>.json`에 세션별로 저장되며 (공유 파일이 아님), **턴 상한 (기본 30)** 이 루프를 유계로 만듭니다. 상한에 도달하면 평가기는 5-섹션 판정 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk) 을 내고 블로킹을 멈춥니다.

## 동사 (verbs)

### `/moai goal "<조건>"` — 등록 + arm

조건 텍스트를 등록하고 활성 세션에 goal을 arm합니다. 조건은 `conditions[]` 배열로 파싱됩니다 — 순수 셸 명령 문자열은 기계적 조건, 트랜스크립트를 참조하는 주장은 모델 조건입니다. arm하면 `.moai/state/goal/<session-id>.json`이 원자적으로 (temp+rename) 기록되고, `stop-goal` Stop 훅이 다음 턴 종료 시 이를 집어 평가를 시작합니다.

```bash
> /moai goal "go test ./... exits 0; 모든 AC가 PASS로 기록, 또는 30턴 후 중단"
```

### `/moai goal status [--all]`

활성 세션의 goal (또는 `--all`로 모든 세션의 goal) 을 출력합니다 — 조건 텍스트, conditions 배열, 사용한 턴 수 vs 상한, 진행 로그, 라이프사이클 상태 (`armed` / `satisfied` / `ceiling-exit` / `cleared`).

### `/moai goal clear`

활성 세션의 goal을 해제합니다 (상태 파일 삭제). Stop 훅은 arm된 goal이 없음을 보고 블로킹을 멈춥니다. 오케스트레이터가 모델 조건을 충족했다고 판정해 루프를 끝낼 때 씁니다.

{{< callout type="info" >}}
**`resume` 동사는 제공되지 않습니다.** 예전에 논의되던 `resume` (해제된 goal을 아카이브에서 복원) 동사는 현재 CLI에 없습니다 — `moai goal --help`는 `arm` / `status` / `clear`만 나열합니다. `clear`가 상태 파일을 **삭제**하기 때문에 (아카이브로 tombstone하지 않음) 복원할 원본이 남지 않습니다.
{{< /callout >}}

## 진행 모드 (자율 / 반자율)

오케스트레이터가 구현 착수 승인 (plan→run 경계의 `AskUserQuestion`) 을 실행할 때, 승인/거절 결정과 **구분되는 별도 축**으로 **자율 vs 반자율** 진행 모드를 선택하게 합니다. 선택한 모드는 goal 상태의 `progression_mode` 필드에 저장됩니다 (사용자가 고르지 않으면 기본 `autonomous`).

| 모드 | 동작 |
|------|------|
| **자율 (autonomous, 기본)** | 평가기가 조건 충족 또는 상한 도달까지 매 턴 블로킹하며, 턴마다 사용자에게 묻지 않습니다. 기존 Stop 훅 동작 그대로입니다. |
| **반자율 (semi-autonomous)** | `stop-goal` 훅이 매 턴 경계에서 **체크포인트 신호** 블록 JSON을 내보내고, 오케스트레이터가 이를 읽어 `AskUserQuestion` 확인 라운드 (계속 / goal 해제 / 자율로 전환) 를 돌립니다. 훅 자체는 절대 `AskUserQuestion`을 호출하지 않습니다 (훅·서브에이전트 경계 — 구조화 JSON만 방출). |

{{< callout type="warning" >}}
**승인은 두 모드 모두에서 필수입니다.** 진행 모드 축은 게이트가 통과된 **이후** 무엇을 할지 선택할 뿐, 게이트를 우회하지 않고 구현 착수 승인을 완화하지도 않습니다. arm된 goal은 어떤 모드에서도 run-phase 진입을 승인하거나, PR을 만들거나, 파괴적 작업을 수행하지 않습니다.
{{< /callout >}}

## 안전 불변식

1. **구현 착수 승인은 두 모드 모두 필수** — 진행 모드는 승인 이후의 진행 선택이지 게이트 완화가 아니며, 점수와 무관하게 유지됩니다.
2. **arm된 goal은 게이트를 우회하지 않음** — PR을 자동 생성하지 않고, 파괴적 작업을 수행하지 않습니다. 평가기는 턴을 계속할지 여부만 결정하며, 되돌릴 수 없는 작업을 사전 승인하지 않습니다.
3. **`stop-goal` 훅은 `AskUserQuestion`을 호출하지 않음** — 구조화 JSON만 방출합니다 (훅·서브에이전트 경계).
4. **정체 가드 (stagnation guard)** — N회 연속 무진전 반복이 감지되면 루프를 멈추고 E1/E3 에스컬레이션 노트를 담은 5-섹션 판정을 냅니다.

## goal 조건은 빨라야 합니다

평가기는 매 턴 끝날 때마다 실행됩니다. 전체 스위트보다 `go test -run <pattern>`을, 오래 걸리는 명령보다 결정론적 명령을 쓰세요 — `stop-goal`의 Stop 훅 타임아웃은 120초지만, 빠른 명령이 턴 루프를 촘촘하게 유지합니다.

## /moai loop과의 관계

`/moai loop`는 **goal 엔진 위의 프리셋**입니다. `/moai goal`이 사용자가 완료 조건을 직접 선언하는 범용 루프라면, `/moai loop`는 "진단 도구가 찾은 이슈 큐를 다 비울 때까지"라는 조건을 미리 채워 넣은 프리셋입니다.

| 엔진 | 목표 | 완료 조건 |
|------|------|----------|
| `/moai goal` | 조건 선언형 범용 루프 | 사용자 정의 조건식 만족 |
| `/moai loop` | 진단 수정 루프 (프리셋) | 이슈 큐 비움 + 진단 클린 (0 에러 / 테스트 통과 / 커버리지) |

끝 상태를 조건식으로 표현할 수 있다면 `/moai goal`, "도구가 찾는 문제를 전부 없애줘"라면 `/moai loop`가 맞습니다.

## 관련 문서

- [/moai loop - 반복 수정 루프](/utility-commands/moai-loop)
- [/moai fix - 일회성 자동 수정](/utility-commands/moai-fix)
- [/moai - 완전 자율 자동화](/utility-commands/moai)
