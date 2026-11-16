# `card-summary-apply`

operation `card.summary.apply` · 코드 `src/ops/card_summary_apply.rs`, `src/forge/card_export.rs` · Issue #3071 H4 (AC7)

`card-summary-apply PROJECT --card C-N --to github [--number N] [--home DIR] --plan-digest D --permission-reference REF` 는 카드 요약을 호스트에 올리는 유일한 명령이다. 같은 입력으로 다시 만든 계획의 digest 가 같고 권한 참조가 있을 때만 움직인다(아니면 `stale or altered plan` 또는 `permission reference required`, 아무것도 올리지 않음). 올리기 전에 대상의 댓글을 끝까지 읽어, 구성원의 댓글에 이 카드의 표지가 이미 있으면 아무것도 올리지 않고 `already_posted` 와 그 댓글 번호를 낸다. 올렸으면 `posted`. 글은 연결부가 비밀 모양을 가린 뒤 올린다. 자동으로는 올리지 않는다.
