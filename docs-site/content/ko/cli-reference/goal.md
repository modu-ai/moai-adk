---
title: moai goal 목표 루프
weight: 72
draft: false
---

`moai goal` 은 조건을 선언해 두는 에이전틱 목표 루프를 현재 세션에 arm 하고, 상태를 조회하고, 해제합니다. 선언한 조건이 충족되거나 턴 상한에 닿을 때까지, goal 엔진이 턴을 넘겨 가며 작업을 이어 갑니다.

네이티브 `/goal`(사용자만 입력할 수 있는 TUI 커맨드)을 오케스트레이터가 직접 다룰 수 있게 옮겨 온 커맨드입니다. 사용자가 `/goal` 을 일일이 입력하지 않아도 목표를 등록하고 arm 할 수 있습니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai goal arm "<condition>"` | 활성 세션에 목표 등록 + arm (`moai goal "<condition>"` 도 arm 별칭). arm-only — 그 자체로는 작업을 시작하지 않는다 |
| `moai goal status` | 활성 세션의 목표 상태 출력 (`--all` 로 전 세션 나열) |
| `moai goal clear` | 활성 세션의 목표 해제 |
| `moai goal render` | 활성 세션의 목표 대시보드를 self-contained HTML 파일로 렌더 (`.moai/state/goal/` 옆에 저장). arm 된 goal이 없으면 0이 아닌 종료 코드로 끝남. v3.1(PR #1388)부터 판정 섹션(천장 exit 시 사이드카에서 로드)과 재무장 조건부 보기가 대시보드에 표시됩니다 — 자세한 내용은 [/moai goal 대시보드 섹션](/utility-commands/moai-goal#목표-대시보드)을 참고하세요 |

## 공통 플래그

| 플래그 | 설명 |
|--------|------|
| `--session <id>` | 세션 id 재정의 (기본: `moai session current` 로 해석) |
| `--json` | 기계 판독 JSON 출력 |
| `--all` | (`status` 전용) 활성 세션뿐 아니라 모든 세션의 목표 나열 |

## arm 플래그

| 플래그 | 설명 |
|--------|------|
| `--max-turns <N>` | 턴 상한. `0` = 무한 (SPEC-INFINITE-GOAL-001); 생략 시 기본 `30` (완전 호환). **`0`(무한)은 `--max-duration <sec>` 을 필수로 요구** (arm 시점 fail-closed). |
| `--max-duration <sec>` | 실행 시간 상한 (arm 시점 이후 초 단위). **무한 goal(`--max-turns 0`)의 실제 벽시계 상한** — 무한 goal은 이 플래그 없이 arm 할 수 없다. |
| `--cost-cap <value>` | 호출 수 상한으로 **기록 전용(recorded-only)** — 현재 enforce 로직이 없어 실제 bound가 아니다. `--max-turns 0` 에 대한 실제 bound 요구를 충족하지 못해 cost-cap 단독으로는 거부된다. |

## 상태와 평가

목표 상태는 `.moai/state/goal/<session-id>.json` (세션당 파일 1개)에 저장됩니다. Stop 훅 `moai hook stop-goal` 이 매 턴 종료마다 목표를 평가합니다.

**조건 파싱**:

- 실행 가능한 셸 커맨드(선택적으로 `exits <N>` 접미)는 **기계적(mechanical) 조건**이 됩니다.
- 대화 transcript를 참조하는 주장은 오케스트레이터가 평가하는 **모델(model) 조건**이 됩니다.

## 예시

```bash
# 테스트 스위트가 통과할 때까지 계속 작업
moai goal arm "go test ./... exits 0"

# 현재 목표 상태 확인
moai goal status

# 목표 해제
moai goal clear
```

## 관련 문서

- [자율 연속 루프](/ko/advanced/autonomous-loops) — `/goal` 과 `/moai loop` 비교
- [moai loop](/ko/cli-reference/loop)
- [CLI 개요](/ko/getting-started/cli)
