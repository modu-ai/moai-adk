# `reports-move-plan`

operation `reports_move.plan` · 코드 `src/ops/reports_move_plan.rs`, `src/reports_move.rs` · 카드 C-126 (SPEC-20261119-update-moves-reports)

`reports-move-plan [PROJECT]` 는 이전 세대가 저장소 안에 남긴 보고서 파일 중 git 이 추적하지 않는 것을 사용자 HOME 의 보고서 폴더(`<HOME>/.moai/reports/<프로젝트>`)로 옮기면 어떻게 되는지 보여 준다. 읽기만 하고 아무것도 만들거나 옮기지 않는다.

- 대상 폴더: 프로젝트의 `.moai-cli/reports` 와 `.moai/reports`. 두 폴더 아래 일반 파일 중 git 이 추적하지 않는 것(`.gitignore` 로 무시되는 파일 포함)이 옮길 파일이고, 목적지는 폴더 구조를 그대로 둔 같은 상대 경로다.
- 추적 파일: git 이 추적하는 파일은 제자리에 둔다. 수만 `tracked` 로 센다.
- 이미 있는 이름: 목적지에 같은 상대 경로가 있으면 덮어쓰지 않고 `skipped_existing` 에 이름을 남긴다.
- 링크와 폴더가 아닌 것: 따라가지도 옮기지도 않고 `unmovable` 로 센다. 파일은 최대 20,000 개까지 본다.
- git 저장소가 아니거나 git 이 추적 여부를 말하지 못하면 `state` 는 `blocked` 이고 `moves` 는 비어 있다.

필드: `protocol_version`, `project`, `destination`, `state` (`ready` | `unchanged` | `blocked`), `reason`, `sources`, `moves` (`{from, to, bytes}`), `skipped_existing`, `tracked`, `unmovable`, `plan_digest`. 계획 전체가 `reports-move-apply` 의 표준입력으로 들어간다.

종료 코드는 다른 연산과 같다(성공 0). 쓰지 않으므로 되돌릴 것이 없다.
