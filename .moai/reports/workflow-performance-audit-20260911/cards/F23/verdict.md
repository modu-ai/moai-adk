# F23 카드 판정

## Claim

run 4차원 결과는 tree·AC·rubric에 묶인 증거 묶음으로만 재사용하고, `INCOMPLETE`·`CONTESTED`·식별자 불일치에서는 sync 판정 소유자에게 재검토를 넘긴다.

## Evidence

- `bash .claude/hooks/tests/test-evidence-decision-boundary.sh`
- 관찰된 출력: `PASS: run evidence reuse is separated from sync verdict ownership`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f23`의 run/sync 문서 계약을 기준으로 실행하였다.

## Gaps

실제 run→sync 호출 trace에서 재실행 횟수 감소와 cache hit 비율은 관찰하지 않았다.

## Residual-risk

계약은 재사용 조건과 판정 소유자를 고정하지만, runtime ledger가 각 reuse 사유를 실제로 기록하는지는 후속 계측으로 확인해야 한다.
