# `card-times`

operation `card.times` · 코드 `src/ops/card_times.rs`, `src/card_times.rs` · Issue #3067 (S6)

`card-times` 는 stdin 의 `{schema_version:1, owner, repo, issues:[1~50개, 겹치지 않는 양수]}` 로 카드마다 **서버가 적은 단계 시각**을 읽기만 해서 돌려준다. 카드마다 `card`, `state`(`open`·`closed`·`not_found`·`unknown`), `held`·`held_by`(`needs:decision`·`hold`·`<축>:hold` 라벨), `dispatch`(`sent_at`·`received_at`·`decided_at`·`started_at` 과 사이 구간 `spans`, `ack_alert`), `claim_at`(가장 이른 구성원 `moai-claim:` 댓글), `pr_wait_state`·`pr_wait`(`pr-wait:` 댓글의 서버 시각, 본문에 적힌 시각은 `stated_at` 로 따로 두고 계산에 쓰지 않음), `prs`(카드를 `Moai-Claim: <카드>#…` 본문이나 `issue/<카드>-…` 가지로 가리키는 PR 마다 `opened`·`ready`·`ci_queued`·`ci_running`·`audit_reserved`·`verdict`·`repairs`·`merged` 와 `spans`)가 나온다.

각 단계 값은 `{state, at, at_epoch, source, comment_id?, reason?}` 이고 `state` 는 셋 중 하나다. `observed` 는 서버 시각(댓글 `createdAt`, PR `createdAt`·`mergedAt`, 타임라인 사건, 검사 묶음)을 읽은 것이다. `absent` 는 읽기를 끝까지 했는데 기록이 없는 것이다(미병합, 미착수, 보류). `unknown` 은 읽기가 잘렸거나(댓글·커밋·목록이 한 번에 읽는 개수를 넘음), 요청 예산(30)이 바닥났거나, 시각이 지금보다 미래이거나 `YYYY-MM-DDTHH:MM:SSZ` 꼴이 아닌 것이다. 구간은 양쪽이 `observed` 이고 끝이 처음보다 이르지 않을 때만 `secs` 가 있고, 끝이 이르면 `unknown`, 한쪽이 `unknown` 이면 `unknown`, 아니면 `absent` 다. 수리(`repairs`)는 `REPAIR` 줄이 없어서 `AUDIT … FAIL` → 실패한 head 의 다음 커밋(`source: "commit_date"`, 커밋 날짜) → 다른 head 의 다음 판정으로 다시 짠다.

배차 단계는 `dispatch` 의 읽기 점검(`dispatch-check`)을 이미 읽은 댓글에 그대로 돌려 얻는다(두 번째 읽기 없음). 구성원이 아닌 사람의 `pr-wait:`·`AUDIT`·`moai-claim:` 줄은 무시하고 `ignored_control_lines` 에 센다. GraphQL `query` 만 보내고(요청 2개 카드마다 1개, 최대 `api_budget` 30), 쓰기·디스크 기록이 없으며 `writes:false`, `claim_authorized:false`, `ownership_transferred:false`, `audit_gate_integrated:false` 다. 이름과 `observed|absent|unknown` 세 낱말은 리더 결정(#3068 과 조율)을 기다리는 기본값이고 `src/card_times.rs` 상수 한 곳에 있다.
