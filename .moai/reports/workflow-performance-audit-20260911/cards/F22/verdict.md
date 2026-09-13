# F22 카드 판정

## Claim

docs drafters, MX shards, judge, verification 작업이 각자 별도 동시성 예약을 하지 않고 하나의 `max_concurrency` bounded queue를 공유한다.

## Evidence

- `bash .claude/hooks/tests/test-resource-budget-contract.sh`
- 관찰된 출력: `PASS: sync fanout uses one bounded resource budget`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f22`의 resource rule과 sync 지침을 기준으로 실행하였다.

## Gaps

실제 런타임에서 peak concurrency와 queue wait 수치를 수집하는 실행 trace는 관찰하지 않았다.

## Residual-risk

이 카드는 admission·기록 계약을 추가했으며, 실제 scheduler 구현과 지표 저장 wiring은 별도 runtime 작업으로 남아 있다.
