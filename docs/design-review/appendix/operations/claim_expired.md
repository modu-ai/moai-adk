# `claim-expired`

operation `claim.expired` · 코드 `src/ops/claim_expired.rs`, `src/claim_expired.rs` · Issue #3015 C7a (계획 6894357631, 조사 N19)

`claim-expired [PROJECT] [--repo OWNER/REPO] [--now T]` 는 열린 Issue·PR 중 `moai-claim:` 줄이 있는 것을 한 번에 찾아, **임대가 끝난 보유자**를 `result.expired` 로 돌려준다. 리더가 Issue 를 하나씩 이름 붙이지 않고도 조용해진 레인을 찾기 위한 읽기 전용 목록이다. GET 만 쓰고 아무것도 쓰지 않으며 인수도 하지 않는다.

- 찾기: `repo:OWNER/REPO is:open in:comments "moai-claim:"` 검색 한 쪽(100개). 총수가 보인 수보다 크거나, 꽉 찬 쪽인데 총수가 없거나, GitHub 가 불완전하다고 알리면 `partial` 에 이유를 적는다. 한 번에 댓글을 읽는 스레드는 60개까지이고, 넘으면 `partial` 에 적는다.
- 판정: 스레드마다 댓글을 끝까지 읽고 `claim-verify` 와 같은 규칙(`claim_verify::evaluate`)으로 보유자를 정한다. 보유자의 `lease_until <= now` 이면(끝나는 그 초도 만료) 목록에 넣는다. 해제(`moai-claim-release:`)·인계(`moai-claim-handoff:`)로 닫힌 claim 은 넣지 않고, 같은 세션의 더 높은 epoch 가 살아 있으면 그 임대로 판단한다. 댓글 작성자는 따지지 않는다(작성자 제한은 C7c).
- 결과: `now`, `threads_read`, `expired`(`issue`·`claim_comment`·`session`·`epoch`·`lease_until`·`expired_secs`, 오래 만료된 순), `unread`(댓글을 끝까지 읽지 못한 스레드 번호 — “만료 없음”으로 읽지 않는다), `partial`, `takeover`(인수는 여기서 하지 않는다는 안내).
- 인수는 그대로 `claim-timeout-check` 의 세 조건과 handoff 줄 규칙을 따른다.
