# `goal-show`

operation `goal.show` · 코드 `src/ops/goal_show.rs`, `src/goal.rs` · Issue #2679

`goal-show PROJECT` 는 이 clone 의 지속 목표를 읽는다. 저장 위치는 claim journal 과 같은 git 공통 폴더 아래 `.moai-cli/goal/goal.json` 이라, 주 worktree 와 모든 linked worktree 가 같은 파일을 본다. 결과: `state`(`active`·`cleared`·`none`), `store`, `goal`(`text`, `end_condition`, `exclusions`, `source`, `started_at`, `updated_at`, `revision`, `active`), `sha256`. 목표가 없으면 `state: none` 과 `detail` 을 낸다. 읽기만 한다.
