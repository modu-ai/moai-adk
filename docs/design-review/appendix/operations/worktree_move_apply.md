# `worktree-move-apply`

operation `worktree.move.apply` · 코드 `src/ops/worktree_move_apply.rs`, `src/worktree_move.rs` · Issue #2723

`worktree-move-apply PROJECT --root DIR [--observations FILE] --plan-digest D --permission-reference REF` 는 같은 입력으로 다시 계산한 미리보기의 digest 가 같을 때만 `move` 항목을 옮긴다(다르면 아무것도 쓰지 않고 `stale or altered plan`). 사용자가 이동을 승인한 뒤에만 부른다. 항목마다 옮기기 직전에 프로세스·열린 파일·추적 변경을 다시 보고(`refused`), `<root>/<프로젝트 키>` 폴더만 만들고, `git worktree move` 로 옮긴다(`--force` 없음). 그다음 git 이 새 경로에 같은 HEAD·가지로 등록했고 그 안에서 git 이 도는지 확인하고, 아니면 `git worktree repair` 를 한 번 돌린 뒤 다시 확인한다. 옮긴 항목마다 되돌리는 `git worktree move` 명령(`revert`)을 준다. 다시 미리보기하면 옮긴 worktree 는 `keep` 이다.
