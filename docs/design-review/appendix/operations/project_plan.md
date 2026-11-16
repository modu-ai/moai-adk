# `project-plan`

operation `project.plan` · 코드 `src/ops/project_plan.rs`

`project-plan PROJECT` 는 stdin 의 프로필(schema v2)로 프로젝트 프로필 미리보기를 낸다. 모르는 키나 v2 아닌 schema 는 값을 남기지 않고 거절한다(REQ-03A-08). `src/main.rs` 에서 그대로 옮김(#2834 R4a).

버저닝(card C-69): 프로필의 선택 키 `versioning` 은 `scheme`(`semver`·`headver`·`calver`·`custom`)과, 고른 경우의 `format`(CalVer 형식: 점으로 이은 토큰 2~4개, 같은 토큰은 한 번만, 연도 토큰은 맨 앞에 하나만. `custom` 이면 사용자의 말 한 줄, 128글자 이하이고 제어 문자·`<`·`>`·백틱·줄 머리의 `#`·보이지 않는 문자(U+200B–200F, U+2028, U+2029, U+202E)·`/` 가 없어야 함)·`head`(HeadVer 의 head, 0 이상 9999 이하의 정수)를 담는다. 모르는 방식, SemVer·HeadVer 의 `format`, HeadVer 가 아닌 방식의 `head`, 모르는 필드, 객체가 아닌 값(배열·문자열)은 값을 남기지 않고 거절한다. 키가 없으면 선택이 기록되지 않은 것이고, 키가 생기기 전에 쓴 프로필은 그대로 읽힌다.
