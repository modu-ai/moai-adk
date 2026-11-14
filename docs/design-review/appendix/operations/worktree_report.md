# `worktree-report`

operation `worktree.report` · 코드 `src/ops/worktree_report.rs` · WT-1 #2380

`worktree-report PROJECT [--tmp-root DIR] [--budget-ms N]` 는 저장소의 worktree 와 빌드 폴더, 임시 `moai-*` 폴더가 쓰는 공간을 보인다. 사실만, 지울 후보는 판단하지 않고 쓰지 않는다. `src/main.rs` 에서 그대로 옮김(#2834 R3).
