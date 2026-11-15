# `audit-claim-stale`

operation `audit.claim_stale` · 코드 `src/ops/audit_claim_stale.rs`, `src/audit_stale.rs` · Issue #3015 C7b (계획 6895140034, 조사 N2)

`audit-claim-stale [PROJECT] [--repo OWNER/REPO] [--lane NAME] [--since-hours 48] [--now T]` 는 **판정 없이 닫힌·병합된 PR 에 남은 감사 예약**을 `result.stale` 로 돌려준다(예: 게이트 밖 병합). GET 만 쓰고 아무것도 쓰지 않는다.

- 대상: `pulls?state=closed&sort=updated&direction=desc&per_page=30` 한 쪽 중 `closed_at` 이 `--since-hours`(1~720, 기본 48) 안인 PR. 경계 시각 자체는 안에 든다. 한 쪽이 꽉 찼고 마지막 PR 도 창 안이면 더 오래된 닫힌 PR 을 못 읽었을 수 있어 `partial` 에 적는다.
- 판정: 각 PR 의 댓글을 끝까지 읽어, 닫힌 시점의 머리(`head.sha`)에 `moai-audit-claim: lane=… head=…` 줄이 있고 그 머리를 든 `AUDIT … verdict: PASS|FAIL` 줄이 없으면 나열한다(`lane_status::verdict_on` 과 같은 규칙). 다른 머리의 판정은 이 머리의 예약에 답하지 않는다. `holder` 는 `audit-claim` 과 같은 규칙(가장 낮은 예약 번호, 리더의 `by=leader` 인계 줄이 옮김)이고 `reservations` 는 그 머리의 모든 예약이다.
- `--lane NAME`: 그 레인이 예약했거나 `holder` 인 것만. 레인 이름은 `audit-claim` 의 `same_lane` 규칙으로 비교한다(`agent-1-h2` 는 `agent-1`).
- 결과: `now`, `since`, `pull_requests_read`, `stale`(`pr`·`merged`·`closed_at`·`head`·`holder`·`reservations`·`verdict_seen: false`·`guidance`, PR 번호가 큰 순), `unread`(댓글을 끝까지 읽지 못한 PR 번호 — “남은 예약 없음”으로 읽지 않는다), `partial`.
- 해제 줄은 새로 만들지 않는다. 닫힌 PR 의 예약은 `audit-claim` 규칙상 효력이 없고(예약은 열린 PR 의 현재 머리에만 쓰인다) 산문 해제는 읽히지 않는다. `guidance` 는 그 감사가 더는 필요 없고 레인의 감사 worktree 는 `worktree-gc-plan` 대상이라고만 알린다. 댓글 삭제·수정·인계는 하지 않는다.
