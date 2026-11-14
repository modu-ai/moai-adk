# `go-hooks-apply`

operation `go_hooks.apply` · 코드 `src/ops/go_hooks_apply.rs`, `src/go_hooks.rs` · Issue #2877

`go-hooks-apply --plan-digest D --permission-reference REF [PROJECT]` 는 확인한 `go-hooks-plan` 그대로 바꾼다. 순서: 영수증 폴더(`<상태>/receipts/go-hooks-<id>/`)에 settings·소유 영수증 바이트를 남기고, `.claude/hooks/moai` 를 그 폴더의 `archive/moai` 로 옮기고, settings 에서 `hooks` 키만 빼고(다른 키는 그대로), `hooks-apply` 와 같은 쓰기·소유 영수증으로 MoAI hook 묶음을 넣는다. 중간에 실패하면 모두 되돌리고 영수증은 `rolled_back`. digest 가 다르면 쓰지 않는다.
