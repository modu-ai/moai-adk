# `project-plan`

operation `project.plan` · 코드 `src/ops/project_plan.rs`

`project-plan PROJECT` 는 stdin 의 프로필(schema v2)로 프로젝트 프로필 미리보기를 낸다. 모르는 키나 v2 아닌 schema 는 값을 남기지 않고 거절한다(REQ-03A-08). `src/main.rs` 에서 그대로 옮김(#2834 R4a).

버저닝(card C-69): 프로필의 선택 키 `versioning` 은 `scheme`(`semver`·`headver`·`calver`·`custom`)과, 고른 경우의 `format`(CalVer 형식: 점으로 이은 토큰 2~4개에 연도 토큰 포함. `custom` 이면 사용자의 말 한 줄, 128자 이하)·`head`(HeadVer 의 head, 0 이상의 정수)를 담는다. 모르는 방식, SemVer·HeadVer 의 `format`, HeadVer 가 아닌 방식의 `head`, 모르는 필드는 값을 남기지 않고 거절한다. 키가 없으면 선택이 기록되지 않은 것이고, 키가 생기기 전에 쓴 프로필은 그대로 읽힌다.
