# `project-plan`

operation `project.plan` · 코드 `src/ops/project_plan.rs`

`project-plan PROJECT` 는 stdin 의 프로필(schema v2)로 프로젝트 프로필 미리보기를 낸다. 모르는 키나 v2 아닌 schema 는 값을 남기지 않고 거절한다(REQ-03A-08). `src/main.rs` 에서 그대로 옮김(#2834 R4a).
