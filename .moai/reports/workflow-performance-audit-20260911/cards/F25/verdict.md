# F25 카드 판정

## Claim

run이 승인된 `plan_artifact_hash`와 `task_graph_id`를 검증하여 불변 계획이면 manager-spec 재분석을 생략하고, hash·scope·의존성·위험 분류 변화가 있을 때만 영향 범위를 재계획한다.

## Evidence

- `bash .claude/hooks/tests/test-plan-handoff-contract.sh`
- 관찰된 출력: `PASS: run reuses an approved plan identity and replans only on drift`

## Baseline-attribution

위 출력은 `WT-workflow-audit-f25`의 run handoff 계약을 기준으로 실행하였다.

## Gaps

실제 plan→run 세션에서 manager-spec 호출 횟수와 중복 질문 수는 계측하지 않았다.

## Residual-risk

문서 계약은 plan identity와 재계획 조건을 고정하지만 실제 orchestrator handoff 저장소·hash 전달 wiring은 후속 구현 검증이 필요하다.
