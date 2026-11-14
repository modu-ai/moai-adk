# `worktree-gc-plan`

operation `worktree.gc.plan` · 코드 `src/ops/worktree_gc_plan.rs` · WT-4 #2380

`worktree-gc-plan PROJECT --integration-ref REF` 는 연결된 worktree 마다 `remove`·`trim`(cargo 빌드 출력만)·`skip` 과 이유, digest 를 보여 준다. 지우지 않고, fetch 하지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R2c).
