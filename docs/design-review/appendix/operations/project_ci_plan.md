# `project-ci-plan`

operation `project.ci_plan` · 코드 `src/ops/project_ci_plan.rs`, `src/project_ci_plan.rs` · SPEC-20261122-project-ci-design (REQ-6)

`project-ci-plan PROJECT` 는 프로젝트의 CI 워크플로 파일을 미리 보여 준다. stdin 의 요청(`protocol_version`, 선택 `runs_per_month`, `language`, `remove`)과 프로필에 저장한 CI 답·선언한 검사 목록으로 파일 내용을 만들고, 디스크의 파일·기록된 기준(`manifests/ci.json`)과 3-way 로 비교한다. 아무것도 쓰지 않는다.

- 파일마다 `action`: `create`·`merge`·`unchanged`·`conflict`(디스크에 있고 기준이 없거나 사용자가 고친 파일은 그대로 두고 충돌로 보인다)·`stale`·`remove`. 그 밖의 `.github/` 파일은 `preserved` 로 나열만 하고 건드리지 않는다.
- `stale`: 기록된 파일인데 지금 답이 더 이상 만들지 않는 것(m-flow 를 벗어난 뒤의 야간 실행, 배포를 끈 뒤의 배포 파일). 디스크에 그대로 남고 매 계획에 다시 보인다. 사용자가 `remove` 에 그 경로를 적고 파일이 기록된 기준과 같을 때만 `remove` 가 된다. 고친 파일, 아직 계획에 있는 파일, 사용자 자신의 워크플로, 그 밖의 경로는 계획이 거절한다.
- `estimate`: 달마다 드는 비용 추정(공개 저장소는 무료, 비공개는 달러와 가격표 날짜), 예산을 넘으면 더 싼 선택지를 말한다. 도구는 비교만 하고 예산을 설정하지 않는다.
- `plan_digest`: 이 계획의 지문. `project-ci-apply` 가 같은 값과 권한 참조를 받아야 쓴다.
