# `project-plan`

operation `project.plan` · 코드 `src/ops/project_plan.rs`

`project-plan PROJECT` 는 stdin 의 프로필(schema v2)로 프로젝트 프로필 미리보기를 낸다. 모르는 키나 v2 아닌 schema 는 값을 남기지 않고 거절한다(REQ-03A-08). `src/main.rs` 에서 그대로 옮김(#2834 R4a).

버저닝(card C-69): 프로필의 선택 키 `versioning` 은 `scheme`(`semver`·`headver`·`calver`·`custom`)과, 고른 경우의 `format`(CalVer 형식: 점으로 이은 토큰 2~4개, 단위마다 하나씩 연도 → 월 또는 주 → 일 → `MINOR` → `MICRO` → `MODIFIER` 순서이고 연도가 맨 앞. 일은 월 뒤에만 오므로 주와 함께 쓸 수 없고, `MODIFIER` 는 연도 바로 뒤에 올 수 없으며, `MAJOR` 는 거절한다(calver.org 에서 MAJOR 는 첫 숫자인데 여기서는 연도가 첫 숫자). `custom` 이면 사용자의 말 한 줄, 128글자 이하이고 제어 문자·`<`·`>`·백틱·`[`·`]`·`*`·`_`·`/`·`\`·줄 머리의 `#`·줄과 문단 구분 문자(U+2028, U+2029)·Unicode 가 기본 무시(Default_Ignorable_Code_Point, Unicode 18.0.0)로 정한 보이지 않는 문자(U+00AD, U+034F, U+061C, U+115F–U+1160, U+17B4–U+17B5, U+180B–U+180F, U+200B–U+200F, U+202A–U+202E, U+2060–U+206F, U+3164, U+FE00–U+FE0F, U+FEFF, U+FFA0, U+FFF0–U+FFF8, U+1BCA0–U+1BCA3, U+1D173–U+1D17A, U+E0000–U+E0FFF)가 없어야 함)·`head`(HeadVer 의 head, 0 이상 9999 이하의 정수)를 담는다. 모르는 방식, SemVer·HeadVer 의 `format`, HeadVer 가 아닌 방식의 `head`, 모르는 필드, 객체가 아닌 값(배열·문자열)은 값을 남기지 않고 거절한다. 키가 없으면 선택이 기록되지 않은 것이고, 키가 생기기 전에 쓴 프로필은 그대로 읽힌다.
