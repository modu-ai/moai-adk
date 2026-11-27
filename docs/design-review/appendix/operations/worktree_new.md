# `worktree-new`

operation `worktree.new` · 코드 `src/ops/worktree_new.rs`, `src/worktree_new.rs` · SPEC-20261124-lane-worktree-command

레인 카드의 작업트리를 명령 하나로 만든다: 본 체크아웃의 `.claude/worktrees/<이름>` 에, 통합 브랜치(또는 `--base`)에서 딴 새 브랜치로, 절대 경로로 만들고 `git worktree list` 로 다시 읽어 확인한 뒤 그 경로를 출력한다.

- `worktree-new NAME --branch BRANCH [--base REF] [--project DIR]`: `NAME` 은 폴더 이름 하나다(소문자·숫자·`-`, 64자까지, `-` 로 시작하지 않음). 경로를 받지 않으므로 상대 경로, `.`/`..`, `/`, 중첩 자리를 쓸 수 없다. 본 체크아웃은 git 공통 디렉터리에서 읽으므로 연결된 작업트리 안에서 불러도 새 폴더는 본 체크아웃 아래에 생긴다.
- 기본 base 는 통합 브랜치의 지금 커밋이다(프로필의 `work_flow` 나 가드 기록). 통합 브랜치를 알 수 없고 `--base` 도 없으면 거절하고 `--base` 를 알려 준다. `--base` 는 커밋을 가리키는 이름이면 된다.
- 만든 뒤 `git worktree list --porcelain` 을 다시 읽어 그 경로가 새 브랜치로, base 커밋에서, 폴더가 있는 채로 나올 때만 답을 낸다. 답: `path`(절대 경로, `EnterWorktree` 에 그대로 넘기는 값), `name`, `branch`, `base`, `head`, `from`, `primary`. 맞지 않으면 명령이 실패하고 만든 경로를 알린다: 폴더는 그대로 두고 아무것도 지우지 않는다.
- 거부(아무것도 만들지 않음): 이름이 위 규칙을 어길 때, 폴더가 이미 있을 때(무엇이든), 브랜치 이름이 올바르지 않거나 이미 있을 때, base 가 커밋이 아닐 때, `.claude`·`.claude/worktrees` 가 연결(링크)일 때(`home_workdir::place_verdict`), 지역 제외 파일을 맞출 수 없을 때.
- 지우기, 강제, 가지치기, 옮기기 옵션은 없다. 이미 만든 작업트리를 정리하는 것은 `worktree-gc-plan`·`worktree-gc-apply` 의 몫이다.
- 호스트는 `EnterWorktree` 로 들어온 세션에서 본 체크아웃 쪽 `git worktree add` 를 직접 막는다. work.md 의 순서는 그대로다: 먼저 `ExitWorktree(keep)`, 그다음 이 명령, 마지막으로 출력된 `path` 로 `EnterWorktree`.
