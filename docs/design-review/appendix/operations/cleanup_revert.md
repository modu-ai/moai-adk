# `cleanup-revert`

operation `cleanup.revert` · 코드 `src/ops/cleanup_revert.rs`, `src/update_cleanup.rs` · Issue #2872 (1단계)

`cleanup-revert HOME --stamp STAMP --permission-reference REF` 는 `cleanup-revert-plan` 의 `restore` 항목을 원래 경로로 `rename` 해 되돌리고(필요한 부모 폴더는 만든다) receipt 상태를 `reverted` 로 바꾼다. `conflict`·`missing` 이 하나라도 있으면 통째로 거절하고 아무것도 옮기지 않는다. 이미 되돌린 항목은 `done` 이라 다시 실행해도 안전하다. 지우지 않는다.
