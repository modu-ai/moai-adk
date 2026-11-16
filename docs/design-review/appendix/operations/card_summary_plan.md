# `card-summary-plan`

operation `card.summary.plan` · 코드 `src/ops/card_summary_plan.rs`, `src/forge/card_export.rs`, `src/ledger/summary.rs` · Issue #3071 H4

`card-summary-plan PROJECT --card C-N --to github [--number N] [--home DIR]` 는 카드 한 장의 요약을 그 카드의 이슈에 올리면 무엇이 올라가는지 미리 보여 준다. 올라갈 글(`text`: 제목·종류·상태와 합격 기준별 상태 한 줄씩), 대상(`target`: 카드의 출처 이슈, `--number` 로 PR 지정 가능), 카드마다 하나인 표지(`<!-- moai-card-summary C-N -->`), `plan_digest` 를 낸다. 카드의 본문·노트는 요약에 넣지 않는다. 카드가 `--to` 가 말하는 호스트에서 온 것이 아니거나 출처가 없으면 오류다. 로컬 원장만 읽고 통신하지 않으며 아무것도 쓰지 않는다.
