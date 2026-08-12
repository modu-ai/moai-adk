---
title: moai loop 피드백 루프
weight: 76
draft: false
---

`moai loop` 은 SPEC 라이프사이클 Ralph 피드백 루프 컨트롤러를 관리합니다. SPEC 하나를 대상으로, 도구 진단이 찾아낸 작업을 반복 처리하는 상태 기계를 제어합니다.

진단 도구가 한 번에 모든 결함을 잡아 내지 못하기 때문에, 루프 컨트롤러는 진단→처리→재진단을 턴 단위로 되풀이하며 SPEC 의 품질 게이트가 녹색(GREEN) 으로 떨어질 때까지 작업을 이어 갑니다. 따라서 하네스(harness)가 룰을 고쳤을 때 그것이 코드베이스 전체로 퍼지는지를 기계적으로 검증하는 안전망 역할을 합니다.

> 이 CLI 커맨드는 Claude Code 대화창의 `/moai loop` 스킬과 별개입니다. CLI는 루프 컨트롤러의 상태를 조작할 뿐이고, 실제 반복 수정은 `/moai loop` 스킬을 받은 오케스트레이터가 수행합니다.

## 하위 명령어

| 명령어 | 설명 |
|--------|------|
| `moai loop start <SPEC-ID>` | SPEC에 대한 피드백 루프 시작 |
| `moai loop status` | 현재 루프 상태 표시 |
| `moai loop pause` | 실행 중인 루프 일시정지 |
| `moai loop resume <SPEC-ID>` | 일시정지된 루프 재개 |
| `moai loop cancel` | 실행 중인 루프 취소 |

## 예시

```bash
# SPEC에 대한 루프 시작
moai loop start SPEC-AUTH-001

# 현재 상태 확인
moai loop status

# 일시정지 후 나중에 재개
moai loop pause
moai loop resume SPEC-AUTH-001

# 루프 취소
moai loop cancel
```

## 관련 문서

- [자율 연속 루프](/ko/advanced/autonomous-loops) — Ralph 엔진과 목표 기반 루프
- [moai goal](/ko/cli-reference/goal)
- [CLI 개요](/ko/getting-started/cli)
