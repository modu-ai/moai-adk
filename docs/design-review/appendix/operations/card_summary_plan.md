# `card-summary-plan`

operation `card.summary.plan` · 코드 `src/ops/card_summary_plan.rs`, `src/forge/card_export.rs`, `src/ledger/summary.rs` · Issue #3071 H4

`card-summary-plan PROJECT --card C-N --to github [--number N] [--home DIR]` 는 카드 한 장의 요약을 그 카드의 이슈에 올리면 무엇이 올라가는지 미리 보여 준다. 올라갈 글(`text`: 제목·종류·상태와 합격 기준별 상태 한 줄씩에 맨 끝 표지 줄까지, 올라가는 바이트 그대로), 대상(`target`: 카드의 출처 이슈, `--number` 로 PR 지정 가능), 카드마다 하나인 표지(`<!-- moai-card-summary C-N -->`), `plan_digest` 를 낸다. 카드의 본문·노트는 요약에 넣지 않는다. 글은 공개 댓글이 되므로 먼저 씻는다: 제어 문자는 공백으로, 제어 줄(`moai-claim`·`repair:`·`AUDIT #` 등으로 시작하는 줄)은 앞에 `· ` 를 붙여, `@이름` 은 `@` 뒤에 폭 없는 공백을 넣어 알림이 가지 않게, `<!--` 는 `&lt;!--` 로, 비밀 모양의 낱말은 `[redacted]` 로 바꾼다. 바꾼 개수는 `sanitized`, 사람이 읽을 경고는 `warnings`, digest 는 이 바이트 위에서 계산한다. 카드의 출처는 `source`(`호스트 owner/repo#번호`)에 있고, 원장에 다른 저장소에서 온 카드가 있으면 `ledger_sources` 와 한 줄 `ledger_notice` 가 그렇게 알린다. 카드가 `--to` 가 말하는 호스트에서 온 것이 아니거나 출처가 없으면 오류다. 로컬 원장만 읽고 통신하지 않으며 아무것도 쓰지 않는다.
