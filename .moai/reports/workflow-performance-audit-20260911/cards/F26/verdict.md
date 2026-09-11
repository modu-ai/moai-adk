# F26 카드 판정

## Claim

신규 project의 verification·UI·external systems·team sharing 네 축을 네 번의 순차 질문이 아니라 한 번의 batched AskUserQuestion으로 수집한다.

## Evidence

- `bash .claude/hooks/tests/test-project-interview-batch-contract.sh`
- 관찰된 출력: `PASS: new-project extended axes are collected in one question batch`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f26`의 project interview 지침을 기준으로 실행하였다.

## Gaps

실제 신규 프로젝트 인터뷰에서 사용자 응답 품질과 재질문 수는 관찰하지 않았다.

## Residual-risk

AskUserQuestion의 구조화된 응답을 `harness-spec.yaml`로 변환하는 runtime wiring은 별도 계측이 필요하다.
