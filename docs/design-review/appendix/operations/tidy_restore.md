# `tidy-restore`

operation `tidy.restore` · 코드 `src/ops/tidy_restore.rs`, `src/tidy/restore.rs`, `src/tidy/archive.rs` · SPEC-20261126-home-tidy-apply (부모 SPEC-20261122-home-tidy)

`tidy-restore HOME --archive NAME --permission-reference REF [--pick PATH…]` 는 `HOME/.moai/archive/<NAME>` 의 항목을 되돌려요. `--pick` 이 없으면 전부예요.

- 쓰기 전에 거절해요: 쓸 수 없는 권한 참조, 보관 이름 모양이 아님, `.moai`·보관 폴더가 없거나 링크, 보관본에 없는 선택.
- 항목마다 그 항목의 파일을 하나도 쓰기 전에 전부 확인해요. 목록이 없거나 링크, 절대 경로·`..`·항목 밖·비밀 이름 경로, 목록에 없는 보관 항목, 해시·크기 차이가 있으면 그 항목은 `failed` 이고 아무것도 쓰지 않아요.
- 아무것도 없는 자리에만 만들고, 만든 것에는 목록의 mode 와 수정 시각을 줘요. 이미 있는 경로는 `existing` 에 이름이 나오고 그대로예요. 링크는 따라가지 않아요.
- 답: `archive`, `state`(`restored`·`partial`), `items`.
- 보관본은 남아요. 명령을 시작하지 않고 `HOME/.moai` 밖에는 쓰지 않아요.
