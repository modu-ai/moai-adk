# `ledger-recover-plan`

operation `ledger.recover.plan` · 코드 `src/ops/ledger_recover_plan.rs`, `src/ledger_recover.rs`, `src/ledger/recover.rs` · Issue #3068 P6

`ledger-recover-plan PROJECT --source FILE [--home DIR] [--journal DIR] [--observations FILE] [--now UTC]` 는 망가진 원장을 읽을 수 있는 사본에서 다시 만들면 무엇이 생기는지 미리 보인다. 쓰지 않는다(사본에도). `--source` 는 원장 파일 자체, `.bak`, 또는 원장 모양의 파일이다. 링크·폴더·WAL 파일·다른 프로젝트나 다른 `application_id` 의 파일·옆에 남은 `-journal` 이 있는 파일은 거절한다. 결과: `tables`(표마다 `state` ok|partial|unreadable, `rows_read`, `last_good_id`, `lost_estimate`, `copied`, `skipped_duplicate`, `rejected`), `unreadable_tables`, `decisions`(활성 예약마다 `restore_held`|`held_for_review` 와 `reasons`), `rejected`(스키마가 거절한 행, 최대 200개만 이름으로), `stale_temp`, `target_state`, `plan_digest`, `writes: false`. 쓰는 규칙·JSON 모양은 `docs/design-review/appendix/ledger_recover.md`.
