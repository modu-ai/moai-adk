# `branch-gc-apply`

operation `branch.gc.apply` · 코드 `src/ops/branch_gc_apply.rs`, `src/branch_gc.rs` · SPEC-20261123-branch-gc

`branch-gc-apply [PROJECT] --integration-ref REF --plan-digest D --permission-reference R` 는 같은 입력으로 계획을 다시 계산해 digest 가 같을 때만 `delete` 브랜치를 `git branch -d`(소문자, `-D` 와 `--force` 는 쓰지 않음)로 지운다. 지우기 직전에 브랜치 tip 과 worktree 목록을 다시 읽고, tip 이 계획과 다르거나 어딘가에 체크아웃됐으면 `refused` 로 건너뛴다. git 이 거절한 브랜치는 git 메시지와 함께 `refused`, 지운 뒤 ref 를 다시 읽어 사라졌을 때만 `deleted` 다. 원격 브랜치와 원격 추적 ref 는 건드리지 않는다. 지우기는 사용자가 고른 뒤에만 돌린다(HOLD: deletion).
