# `card-summary-apply`

operation `card.summary.apply` · 코드 `src/ops/card_summary_apply.rs`, `src/forge/card_export.rs` · Issue #3071 H4 (AC7)

`card-summary-apply PROJECT --card C-N --to github [--number N] [--home DIR] --plan-digest D --permission-reference REF` 는 카드 요약을 호스트에 올리는 유일한 명령이다. 같은 입력으로 다시 만든 계획의 digest 가 같고 권한 참조가 있을 때만 움직인다(아니면 `stale or altered plan` 또는 `permission reference required`, 아무것도 올리지 않음). 올리기 전에 대상의 댓글을 끝까지 읽어, 구성원의 댓글, 또는 토큰 계정 자신이 쓴 댓글에 이 카드의 표지가 이미 있으면 아무것도 올리지 않고 `already_posted` 와 그 댓글 번호를 낸다. 올렸으면 `posted`. 낯선 사람이 표지만 먼저 써 두어도 막히지 않는다(구성원도 토큰 계정도 아니면 무시). 올리는 글은 계획이 보인 `text` 와 바이트까지 같다. 올리는 요청(POST)은 한 번만 보내고 되풀이하지 않는다 — 속도 제한이나 시간 초과가 나면 호스트를 읽어 본 뒤 같은 명령을 다시 실행한다. 자동으로는 올리지 않는다.
