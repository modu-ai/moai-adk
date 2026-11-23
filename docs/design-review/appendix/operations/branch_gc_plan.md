# `branch-gc-plan`

operation `branch.gc.plan` · 코드 `src/ops/branch_gc_plan.rs`, `src/branch_gc.rs` · SPEC-20261123-branch-gc

`branch-gc-plan [PROJECT] --integration-ref REF` 는 로컬 브랜치마다 `delete` 또는 `keep` 과 이유, digest 를 보여 준다. `delete` 는 통합 ref 의 조상이고, 어떤 worktree 에도 체크아웃되지 않았고, 열린 카드 예약(`held`, `expired_candidate`, `held_for_review`)이 이름 짓지 않았고, 보존 ref(`main`, `master`, `v2026.11.0`, `2026.11.0-rc.1`, `archive/*`, 통합 ref 가 가리키는 로컬 브랜치)가 아닌 것뿐이다. `keep` 의 이유는 `preserved`, `checked_out`, `card_reserved`, `not_merged`, `symbolic_ref`, `reservation_check_unavailable`. 프로필이나 원장을 읽을 수 없으면 지울 후보는 모두 `reservation_check_unavailable` 로 남는다. 아무것도 쓰지 않고 fetch 하지 않는다. digest 는 통합 tip 과 `delete` 브랜치의 이름·tip 에만 달려 있다.
