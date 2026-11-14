# `worktree-gc-apply`

operation `worktree.gc.apply` · 코드 `src/ops/worktree_gc_apply.rs` · WT-9 #2498

`worktree-gc-apply PROJECT --integration-ref REF --plan-digest D --permission-reference R` 는 같은 입력으로 계획을 다시 계산해 digest 가 같을 때만 나열된 cargo 출력을 줄이고 worktree 를 지운다(`--force` 없음, 가지는 남김). `src/main.rs` 에서 그대로 옮김(#2834 R2c).
