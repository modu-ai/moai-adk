# t648 — 단계별 검증 순서 예외 승인

## 승인 원문과 귀속

- 일자: 2026-09-12.
- root가 제시한 결정: “내부 저장 기반 먼저 → 해당 baseline의 실패 검증 → 통과 후 외부 연결” 순서의 이번 작업 한정 예외.
- 사용자 응답 원문: **“승인!!!”**
- 응답은 위 직전 질문에 대한 명시 승인으로 root가 이 문서 작성 담당에게 전달했다. 작성 담당의 독자적인 예외 승인이 아니다.
- 대상: 단일 카드 t648, SPEC-TODO-UNIFIED-001 및 최초 내부 저장 기반 SPEC-TODO-RUNTIME-STORE-001.

## 허용 범위 — verification-order only

1. 현재 도달 가능한 공개 API의 RED를 유지한 채 내부 Todo 저장 기반을 먼저 구현한다.
2. 그 실제 baseline에서 assignment 실패/rollback, 활성 WAL snapshot, 동시 writer 및 extension 설치 rollback의 실패 주입·RED→GREEN을 수행한다. 같은 기능의 negative control/mutant 검증이 필요하면 그 근거도 보존한다.
3. 위 필수 안전 검증을 모두 통과하기 전 CLI/hooks 자동완료 연결 및 운영 전환을 금지한다.
4. 이 처리는 해당 검증 **순서** 게이트만 narrowly **BYPASSED** 한 것이다. 기존 plan-audit.md와 plan-audit-2.md의 FAIL을 수정하거나 PASS로 재명명하지 않는다.

## 면제되지 않는 의무

- umbrella 전체23AC, 최초 child4AC 및 기존85% 품질 기준.
- 권한·소유권·identity·migration 안전, 원본 보존, 오류 전파, 실제 검증 출력과 기준 귀속.
- 독립 감사·최종 누적 회귀와 완료 readback.
- 운영 DB 변경, 설치 바이너리 교체, 카드 완료 조작, push/PR/병합/배포에 대한 별도 범위 확인. 이 예외는 이 작업들을 승인하지 않는다.
- 회귀 조건 삭제·약화, 미검증 PASS, 다른 작업/후속 기능에 대한 일반 정책 예외.

## 만료와 재검토

해당 내부 저장 기반의 필수 검증이 완료되는 시점에 이 순서 예외는 만료한다. 실패 또는 새 위험이 드러나면 외부 연결을 진행하지 않고 실패 근거와 수정 결과를 다시 검증한다. 이후 owner/완료/Graph/UI 단계는 해당 baseline과 정상 게이트를 따른다. 범위 확장은 새 명시 결정 없이는 허용되지 않는다.

## Evidence · Baseline-attribution

승인 기록 직전 실제 확인:

```text
$ git branch --show-current
WT-todo-unified
$ git rev-parse HEAD
a315dad9af0d3a0e04862e6106b3993d9a3812f7
$ git rev-list --count --left-right origin/main...HEAD
0	3005
```

fetch origin main 후 비교했다. 이 HEAD는 기존 uncommitted Factory guard 및 테스트/설계 파일 내용을 포함하지 않으며 각 실행 증거에 파일 hash/변경 상태를 별도로 남겨야 한다. 기존 다른 담당의 변경은 보존했다.

## Gaps · Residual-risk

이 문서는 승인 기록이지 구현·검증 성공 증거가 아니다. 아직 구현되지 않은 transaction 경계의 실패 검증은 수행 후 실행 담당이 progress.md §E.2/§E.3에 기록한다. 승인 자체로 데이터 안전성이 증명되지는 않는다.
