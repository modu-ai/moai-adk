# Research — SPEC-TODO-RUNTIME-STORE-001

## Claim

기존 Factory 실행 기록 공개 API와 LoadPure를 사용하는 최소 persistence child다. 별도 Factory 기록의 lifecycle 권위를 새로 만드는 것이 아니다.

## Evidence

작성 전 git branch --show-current / git rev-parse HEAD 출력:

```text
WT-todo-unified
a315dad9af0d3a0e04862e6106b3993d9a3812f7
```

ID Bash regex 검사 출력: PASS. source에서 factory_runtime.go의 RecordFactoryRunStart→homestate.OpenFactory/RecordRun, RecordFactoryCardAssignment→RecordFactoryCardState→RecordCard를 읽었다. backlog_store.go:204의 기존 BacklogRecord에는 Version/LastSeq/Items/Findings/Archived가 있고 runtime 필드는 없다. 이 소스 관측은 신규기능 runtime 측정이 아니다.

## Baseline-attribution

HEAD a315dad9a, t648 기존 작업 공간. 기존 독립 보고서 .moai/reports/t648/red-baseline.md의 1개 RED는 테스트 담당의 실제 관측이며 본 계획 담당은 재실행하지 않았다. 감사2 D1을 PASS로 전환하지 않는다.

## Gaps

현재4 AC의 full RED ledger, 구현/race/rollback/운영이전 미실행. 이후3–4 AC는 선행 확장미존재로 setup이 안되면 미채택을 유지한다.

## Residual-risk

core/확장 version 혼용, MaxOpenConns(1) 내부 재진입, whole-record writer의 오래된 runtime overwrite를 실행 검사해야 한다. 이러한 위험을 관측된 결함으로 단정하지 않는다.
