---
title: moai goal 목표 루프
weight: 72
draft: false
---

`moai goal` 은 현재 세션에 대해 조건을 선언한 에이전틱 목표 루프를 arm/조회/해제합니다. MoAI goal 엔진은 선언한 조건이 충족되거나 턴 상한에 도달할 때까지 세션이 다음 턴으로 넘어가서도 작업을 이어가도록 합니다.

네이티브 `/goal`(사용자 전용 TUI 커맨드)을 오케스트레이터가 직접 다룰 수 있게 옮겨 온 것으로, 사용자가 `/goal` 을 직접 입력하지 않아도 목표를 등록하고 arm 할 수 있습니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai goal arm "<condition>"` | 활성 세션에 목표 등록 + arm (`moai goal "<condition>"` 도 arm 별칭) |
| `moai goal status` | 활성 세션의 목표 상태 출력 |
| `moai goal clear` | 활성 세션의 목표 해제 |

## 공통 플래그

| 플래그 | 설명 |
|--------|------|
| `--session <id>` | 세션 id 재정의 (기본: `moai session current` 로 해석) |
| `--json` | 기계 판독 JSON 출력 |
| `--all` | (`status` 전용) 활성 세션뿐 아니라 모든 세션의 목표 나열 |

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
