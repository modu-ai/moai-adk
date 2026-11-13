# `goal-plan`

operation `goal.plan` · 코드 `src/ops/goal_plan.rs`, `src/goal.rs` · Issue #2679

`goal-plan PROJECT --request FILE` 는 지속 목표의 변경을 미리 보여 준다. 요청은 `schema_version`, `action`(`set`·`clear`), `text`, `end_condition`, `exclusions`(`number`, `reason`), `source` 다. `set` 은 `text`·`end_condition`·`source` 가 필요하고, 각 2,000자 이하이며, 자격증명 모양 문자열은 거부한다. 결과: `change`(`create`·`update`·`clear`·`unchanged`·`none`), `changed_fields`, `current`, `requested`, `current_sha256`, `plan_digest`(지금 파일의 sha256 과 요청을 묶음), `asks_person`(바뀌는 것이 있을 때만 true), `writes: false`. 아무것도 쓰지 않는다.
