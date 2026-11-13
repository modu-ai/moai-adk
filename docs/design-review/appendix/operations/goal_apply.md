# `goal-apply`

operation `goal.apply` · 코드 `src/ops/goal_apply.rs`, `src/goal.rs` · Issue #2679

`goal-apply PROJECT --request FILE --plan-digest D [--now YYYY-MM-DDTHH:MM:SSZ]` 는 같은 요청과 같은 파일로 다시 계산한 미리보기의 digest 가 같을 때만 쓴다(다르면 `stale or altered plan`). `unchanged`·`none` 은 쓰지 않는다. 처음 쓰기는 배타 생성이고, 그다음은 원자 교체다(쓰기 직전 바이트가 미리보기 때와 같아야 함). `revision` 은 1부터 하나씩 늘고, `started_at` 은 문구(`text`)가 그대로면 유지된다. `clear` 는 파일을 지우지 않고 `active: false` 로 남긴다.
