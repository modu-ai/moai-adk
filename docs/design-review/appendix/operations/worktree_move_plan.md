# `worktree-move-plan`

operation `worktree.move.plan` · 코드 `src/ops/worktree_move_plan.rs`, `src/worktree_move.rs` · Issue #2723, #2835

`worktree-move-plan PROJECT [--root DIR] [--observations FILE]` 는 이 clone 의 연결 worktree 목록과, 각각을 공통 루트 바로 아래 `<root>/<프로젝트>-<이름>` 으로 옮기는 미리보기를 낸다(#2835 D3 — `worktree-plan` 이 새 worktree 를 만드는 자리와 같은 모양, 이름은 `home_workdir::worktree_name`). 쓰지 않는다. `--root` 는 절대 경로이고 `.`·`..` 가 없으며 어느 worktree 안에도 있지 않아야 한다. `--root` 가 없으면 `<HOME>/.moai/worktrees` 다(#2835 D3; `$HOME`, Windows 는 `%USERPROFILE%`). HOME 이 없거나 절대 경로가 아니면 거부한다. 이름이 같은 두 프로젝트는 같은 대상 이름이 되므로 둘째는 `destination_exists` 로 막힌다(합치지 않음). 프로젝트 키 `<저장소 폴더 이름>-<clone_id 앞 12자>` 는 clone 을 가리키는 값으로 계속 보고한다.

- 항목마다 `location`(`primary`, `home` — 이미 루트 안, `old_root` — 예전 공통 루트 `<HOME>/.moai-cli/worktrees/<키>/<이름>`, `go` — Go 세대의 `<HOME>/.moai/worktrees/projects/…`, `project`, `claude`, `other`)과 `move`, `keep`(`home`·`old_root`·`go` 는 옮기지 않음), `blocked` 와 이유. 막는 이유: `main_worktree`, `bare`, `missing`(prunable), `locked`, `has_submodules`(git 이 옮기지 못함), `uncommitted`, `untracked`, `unpushed`(어느 원격 추적 ref 에도 없는 커밋 — fetch 하지 않음), `stash`(그 가지의 stash), `in_use`(그 안을 작업 폴더로 둔 프로세스), `current_directory`, `hold`·`handoff_pending`(관측 파일), `hold_check_not_run`(관측 파일에 `holds` 가 없음), `destination_exists`, `destination_collision`. 읽지 못한 사실은 `<사실>_unknown` 으로 막는다.
- 관측 파일: `schema_version` 1, `holds` {`source`, `branches` [{`branch`, `reason`: `hold`|`handoff_pending`}]}. 빈 목록은 "막힌 가지 없음"을 관측한 것이다.
- 결과: `project_key`, `project_root`, `state`(`move` 가 없으면 `unchanged`), `totals`, `items`, `plan_digest`.
