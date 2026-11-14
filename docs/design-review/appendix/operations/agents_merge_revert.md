# `agents-merge-revert`

operation `agents_merge.revert` · 코드 `src/ops/agents_merge_revert.rs`, `src/agents_local_merge.rs` · Issue #2936 (R10)

`agents-merge-revert --receipt ID --permission-reference REF [PROJECT]` 는 `agents-merge-apply` 하나를 되돌린다. `AGENTS.md` 가 apply 가 쓴 바이트 그대로일 때만(영수증의 `agents_after` 와 같을 때만) 영수증의 이전 `AGENTS.md` 로 되돌린다. 그 뒤 손으로 고쳤으면 아무것도 쓰지 않는다. 출처 파일과 백업은 건드리지 않고, 이미 되돌린 영수증은 `state: unchanged` 다. 쓰다가 끊긴 `pending` 영수증은 `AGENTS.md` 가 쓰려던 바이트면 `applied` 처럼 되돌리고, 이전 바이트 그대로면(쓰지 못함) `failed` 로 바꾸고 `state: unchanged` 를 내며, 그 밖이면 아무것도 쓰지 않는다. `failed` 영수증은 되돌릴 것이 없다.
