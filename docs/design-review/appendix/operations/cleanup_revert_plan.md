# `cleanup-revert-plan`

operation `cleanup.revert_plan` · 코드 `src/ops/cleanup_revert_plan.rs`, `src/update_cleanup.rs` · Issue #2872 (1단계)

`cleanup-revert-plan HOME --stamp STAMP` 는 `<상태 폴더>/cleanup/<STAMP>/receipt.json` 을 읽어(링크는 거절, 16 MiB 상한) 항목마다 상태를 보인다. 쓰지 않는다.

- `restore`: 보관본이 있고 digest 가 같으며 원래 경로가 비어 있다.
- `done`: 원래 경로에 같은 digest 가 있고 보관본이 없다(이미 되돌림).
- `conflict`: 원래 경로에 다른 것이 있다.
- `missing`: 보관본이 없거나 digest 가 다르다.
- `action`: `conflict`·`missing` 이 하나라도 있으면 `blocked`, `restore` 가 있으면 `restore`, 아니면 `none`.
