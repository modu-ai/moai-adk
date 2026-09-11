# F24 카드 판정

## Claim

4차원 judge가 `xhigh`를 고정하지 않고 orchestrator가 넘긴 `judge_effort`를 `low|medium|high` 정책으로 해석하며, 누락·지원하지 않는 값은 `high`로 안전하게 대체한다.

## Evidence

- `bash .claude/hooks/tests/test-judge-effort-contract.sh`
- 관찰된 출력: `PASS: 4-dimension judges use the resolved supported effort profile`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f24`의 JS 실행 입력 계약과 rule을 기준으로 실행하였다.

## Gaps

profile별 실제 품질·토큰·호출시간 비교와 orchestrator가 `judge_effort`를 주입하는 runtime trace는 관찰하지 않았다.

## Residual-risk

허용값 clamp는 추가되었으나 실제 모델 이름 해석·비용 계측은 현재 카드 범위에 포함하지 않았다.
