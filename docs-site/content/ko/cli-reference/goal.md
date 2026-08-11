---
title: moai goal 목표 루프
weight: 72
draft: false
---

`moai goal` 은 조건을 선언해 두는 에이전틱 목표 루프를 현재 세션에 arm 하고, 상태를 조회하고, 해제합니다. 선언한 조건이 충족되거나 턴 상한에 닿을 때까지, goal 엔진이 턴을 넘겨 가며 작업을 이어 갑니다.

네이티브 `/goal`(사용자만 입력할 수 있는 TUI 커맨드)을 오케스트레이터가 직접 다룰 수 있게 옮겨 온 커맨드입니다. 사용자가 `/goal` 을 일일이 입력하지 않아도 목표를 등록하고 arm 할 수 있기 때문에, 긴 SPEC 실행이나 자율 루프에서 턴 단위로 사용자가 개입하는 부담을 크게 줄여 줍니다. 따라서 이 커맨드는 관리자 에이전트가 반자율 모드로 장시간 작업할 때 핵심 진입점 역할을 합니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai goal arm "<condition>"` | 활성 세션에 목표 등록 + arm (`moai goal "<condition>"` 도 arm 별칭) |
| `moai goal status` | 활성 세션의 목표 상태 출력 |
| `moai goal clear` | 활성 세션의 목표 해제 |
| `moai goal render` | 활성 세션의 목표 대시보드를 단일 HTML 파일로 렌더 (외부 의존성 없음) |

## 공통 플래그

| 플래그 | 설명 |
|--------|------|
| `--session <id>` | 세션 id 재정의 (기본: `moai session current` 로 해석) |
| `--json` | 기계 판독 JSON 출력 |
| `--all` | (`status` 전용) 활성 세션뿐 아니라 모든 세션의 목표 나열 |

## arm 플래그

| 플래그 | 설명 |
|--------|------|
| `--max-turns <N>` | 턴 상한. `0` = 무한 (오토컴팩트 기반); 생략 시 기본 `30` (완전 호환). `0`은 `--max-duration` 또는 `--cost-cap` 중 하나를 필수로 요구 (arm 시점 fail-closed). |
| `--max-duration <sec>` | 실행 시간 상한 (arm 시점 이후 초 단위). 무한 goal의 1차 실행 상한. |
| `--cost-cap <value>` | `Ceiling`에 기록되는 비용 상한. 실제 적용은 후속 작업 (현재 호출/토큰 계산이 없음); 무한 goal은 `--max-duration`과 정체 가드로 여전히 묶인다. |

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
