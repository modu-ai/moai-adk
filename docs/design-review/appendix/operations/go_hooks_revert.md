# `go-hooks-revert`

operation `go_hooks.revert` · 코드 `src/ops/go_hooks_revert.rs`, `src/go_hooks.rs` · Issue #2877

`go-hooks-revert --receipt ID --permission-reference REF [PROJECT]` 는 `go-hooks-apply` 하나를 되돌린다. settings 와 소유 영수증이 apply 가 쓴 바이트 그대로이고 `.claude/hooks/moai` 가 비어 있을 때만, 두 파일을 이전 바이트로 돌리고 스크립트를 제자리로 옮긴다. 하나라도 바뀌었으면 아무것도 쓰지 않는다. 지우지 않는다.
