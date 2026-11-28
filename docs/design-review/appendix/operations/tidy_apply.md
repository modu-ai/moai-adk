# `tidy-apply`

operation `tidy.apply` · 코드 `src/ops/tidy_apply.rs`, `src/tidy/apply.rs`, `src/tidy/archive.rs` · SPEC-20261126-home-tidy-apply (부모 SPEC-20261122-home-tidy)

`tidy-apply HOME --pick PATH… --confirm DIGEST --permission-reference REF` 는 새로 계산한 `tidy-plan` 에서 고른 항목만 실행해요. `--threshold`·`--entry-limit`·`--time-limit-secs` 는 계획을 읽을 때와 같은 값을 줘요.

- 쓰기 전에 거절해요: 선택이 없음, 계획에 없거나 `keep` 인 선택, `--confirm` 이 지금 계획의 `digest` 와 다름, 쓸 수 없는 권한 참조, `.moai` 가 링크, `.moai/archive` 가 링크이거나 폴더가 아님.
- `archive` 항목: `.moai/archive/<UTC>/item-NNN/` 에 `data.tar.gz` 와 `list.json`(경로·종류·크기·sha256·mode·수정 시각)을 새로 만들어요(폴더 0700, 파일 0600). 보관본과 목록을 다시 읽고 원본도 다시 확인한 뒤에만, 목록의 파일을 하나씩 지우고 폴더는 비었을 때만 지워요. 재귀 삭제는 없어요.
- 확인에 실패하면 그 항목의 원본은 하나도 지우지 않고, 이번 실행이 만든 그 항목의 파일만 지워요(`failed`).
- 목록 뒤에 생긴 파일은 남고 `left` 에 이름이 나와요. 지우기 직전에 달라진 파일이나 링크로 바뀐 폴더는 남아요(`partly_removed`). 링크는 따라가지 않아요.
- `delete` 항목은 빈 폴더만 지워요. 그 사이에 파일이 생겼으면 남아요(`kept`, `not_empty`).
- 비밀 모양 이름, 링크, 특수 파일, 읽을 수 없는 부분, 한도를 넘는 항목은 통째로 건너뛰어요(`skipped`). 비밀 이름 파일은 열지 않아요.
- 옛 도구(`retired_tool`) 항목은 Go 세션이 살아 있을 수 있으면 `in_use`, 알 수 없으면 `in_use_unknown` 으로 건너뛰어요. 지우기 직전에 한 번 더 봐요.
- 영수증 `receipt.json` 은 첫 삭제 전에 `pending` 이고, 항목마다 상태를 남기며 `committed` 또는 `partial` 로 끝나요.
- 답: `archive`(보관 이름), `archive_root`, `state`(`committed`·`partial`), `items`, `free_bytes`(앞·뒤, 읽을 수 없으면 `not_checked`).
- 명령을 시작하지 않고 `HOME/.moai` 밖은 바꾸지 않아요. 보관본은 이 연산이 지우지 않아요.
