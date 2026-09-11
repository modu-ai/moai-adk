# F21 카드 판정

## Claim

Phase 7 감사와 문서 초안 fan-out 전에 immutable `sync_snapshot_id`를 봉인하고, D1-D5가 같은 snapshot을 읽도록 하여 Phase 11 결과를 선행 입력으로 요구하지 않는다.

## Evidence

- `bash .claude/hooks/tests/test-sync-snapshot-contract.sh`
- 관찰된 출력: `PASS: sync drafters consume one immutable snapshot before parallel launch`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f21`에서 수정된 `doc-execution.md`를 기준으로 실행하였다.

## Gaps

실제 Claude fan-out에서 snapshot ID가 전파되는 runtime trace와 draft 폐기율은 이 정적 계약 시험에서 관찰하지 않았다.

## Residual-risk

snapshot 생성·저장 primitive 자체는 아직 orchestrator runtime의 별도 구현 대상이며, 이 카드는 소비 순서와 입력 계약을 고정한다.
