---
id: SPEC-20261119-update-moves-reports
title: "The project update moves the earlier untracked report files to the user-level reports folder"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-reports-outside-repo
---

# The project update moves the earlier untracked report files to the user-level reports folder

## Intent

사용자 지시(2026-11-19, 이 일의 카드 결정 노트 3273): `/moai update` 가 프로젝트의 옛 보고서 폴더(`.moai-cli/reports`, 그리고 조사에서 나오는 `.moai/reports`)에 남은 파일을 새 자리 `~/.moai/reports/<project>/` 로 옮겨요. 조건: update 의 기존 미리보기(plan) → 적용(apply) 흐름 안에서 하고 미리보기에 옮길 파일 수와 목적지를 보여 줘요; 추적되지 않는 파일만 옮기고 git 이 추적하는 파일은 "추적 중, 건드리지 않음" 으로 수만 보여 줘요; 목적지에 같은 이름이 있으면 덮어쓰지 않고 건너뛰며 수를 알려요; 두 번 돌려도 같은 결과(멱등)예요; 폴더 구조를 유지하고 목적지 권한은 본인만이에요; 시험은 임시 HOME + 임시 프로젝트에서 옮김·추적 파일 유지·이름 충돌·재실행 네 경우예요.

이 조각은 실제 HOME 에 쓰는 길이라 tier 2 감사예요. 목적지는 부모 SPEC 의 `home_workdir::reports_dir` 가 줘요(새 해석기를 만들지 않아요).

## Non-goals

- 추적 파일 옮기기·지우기: 삭제는 사용자 선택이에요(부모 SPEC 의 결정 노트).
- `.gitignore` 고치기, 옛 폴더 지우기: 옮긴 뒤 비는 폴더도 그대로 둬요.
- Go 세대 `~/.moai/reports` 의 옛 파일 옮기기.

## Requirements

- REQ-1: The command `reports-move-plan` shall list the files that git does not track under `.moai-cli/reports` and `.moai/reports` of the project with their destinations under the HOME reports folder.
- REQ-2: The command `reports-move-plan` shall count the files that git tracks and name them as tracked and left alone.
- REQ-3: The command `reports-move-apply` shall move each planned file to the same relative path under the HOME reports folder and keep the folder structure.
- REQ-4: The command `reports-move-apply` shall skip a file whose destination name already exists and count it as skipped without overwriting it.
- REQ-5: The command `reports-move-apply` shall refuse a plan made before the files moved and leave every file as it is.
- REQ-6: The command `reports-move-apply` shall make the folders it creates readable by the owner only and clear the group and other permission bits of the files it moves.
- REQ-7: The command `reports-move-apply` shall refuse a stale plan digest or a missing permission reference and write nothing.
- REQ-8: The command `project-update-plan` shall show one step `reports` whose preview carries the counts to move and to skip and the tracked count and the destination.
- REQ-9: The step `reports` shall offer its apply command only with the plan digest of the same preview.
- REQ-10: The command `reports-move-plan` shall report the state `blocked` and plan no move when the project is not a git repository or git cannot tell what is tracked.
- REQ-11: The command `reports-move-plan` shall report the state `blocked` and plan no move when a folder on the way to an earlier report folder is a link, when that folder belongs to another repository, when HOME is not an absolute path, when the reports folder would lie inside the project, when more than 20,000 files lie under the earlier folders or when a folder of the destination is not a plain folder.
- REQ-12: The command `reports-move-plan` shall name the destination folder after the primary checkout when the project is a linked worktree.

## Acceptance

- AC-1 (REQ-1, REQ-2): 임시 HOME·임시 프로젝트에서 `reports-move-plan` 이 추적되지 않은 파일 수와 목적지와 "추적 중, 건드리지 않음" 수를 보여 줘요.
  verify: `cargo test --locked --offline --test reports_move_cli no_earlier_folder_is_unchanged_and_preview_counts_are_reported`
- AC-2 (REQ-3): 적용하면 추적되지 않은 파일이 폴더 구조 그대로 새 자리로 옮겨지고 추적 파일은 제자리에 남아요.
  verify: `cargo test --locked --offline --test reports_move_cli untracked_report_files_move_with_their_folders_and_tracked_ones_stay`
- AC-3 (REQ-4): 목적지에 같은 이름이 있으면 덮어쓰지 않고 건너뛰며 그 수를 알려요(목적지 바이트가 그대로예요).
  verify: `cargo test --locked --offline --test reports_move_cli a_name_the_destination_holds_is_kept_and_counted_never_overwritten`
- AC-4 (REQ-5): 옮긴 뒤 새 미리보기는 `unchanged`·옮길 것 0 이고, 옛 계획을 다시 적용하면 낡은 계획으로 거절되어 어떤 파일도 바꾸지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli a_second_run_moves_nothing_and_changes_nothing`
- AC-5 (REQ-6): 새로 만든 폴더는 0700 이고 옮긴 파일에서 그룹·다른 사람 권한 비트가 지워져요(Unix).
  verify: `cargo test --locked --offline --test reports_move_cli moved_files_and_folders_are_owner_only`
- AC-6 (REQ-7): 어긋난 digest 와 빠진 permission reference 는 거절되고 아무 파일도 옮겨지지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli a_plan_made_stale_by_a_new_file_or_a_wrong_digest_or_reference_is_refused`
- AC-7 (REQ-8, REQ-9): `project-update-plan` 의 `reports` 단계가 수와 목적지를 미리 보여 주고 apply 줄은 같은 미리보기의 digest 를 써요.
  verify: `cargo test --locked --offline --test reports_move_cli update_plan_shows_the_reports_step`
- AC-8 (REQ-10): git 저장소가 아닌 프로젝트에서는 `blocked` 이고 아무것도 옮기지 않아요.
  verify: `cargo test --locked --offline --test reports_move_cli a_folder_git_cannot_read_blocks_the_move_and_nothing_is_created`
- AC-9 (REQ-11): `.moai-cli`(또는 `.moai`)가 다른 저장소로 가는 링크이거나 중첩 저장소이면, HOME 이 상대 경로이거나 프로젝트 안이면, 파일이 20,000개를 넘으면, 목적지 폴더가 링크·파일이면 `blocked` 이고 아무것도 옮겨지거나 지워지지 않아요(링크 너머 저장소의 추적 파일이 그대로예요); 20,000개는 계획되고 20,001개는 막혀요.
  verify: `cargo test --locked --offline --test reports_move_cli a_linked_state_folder_is_never_followed_into_another_repository`; `cargo test --locked --offline --test reports_move_cli a_report_folder_inside_a_nested_repository_is_never_moved`; `cargo test --locked --offline --test reports_move_cli a_home_that_is_relative_or_inside_the_project_blocks_the_move`; `cargo test --locked --offline --test reports_move_cli the_file_limit_holds_at_n_and_blocks_at_n_plus_one`; `cargo test --locked --offline --test reports_move_cli a_destination_that_is_a_link_or_a_file_blocks_the_move`
- AC-10 (REQ-12): 연결된 작업 트리에서 돌리면 목적지가 그 작업 트리 폴더 이름이 아니라 primary 체크아웃 이름의 폴더예요.
  verify: `cargo test --locked --offline --test reports_move_cli a_linked_worktree_moves_to_the_primary_checkouts_folder`

## Scope

- write: `.moai/specs/SPEC-20261119-update-moves-reports/spec.md`, `src/reports_move.rs`(새), `src/lib.rs`(모듈 한 줄), `src/ops/reports_move_plan.rs`·`src/ops/reports_move_apply.rs`(새), `docs/design-review/appendix/operations/reports_move_plan.md`·`reports_move_apply.md`(새), `src/help.rs`, `tests/fixtures/help/*.txt`, `src/project_update.rs`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`, `tests/reports_move_cli.rs`(새), `tests/project_update_plan_cli.rs`(reports 단계가 go-state 명령을 붙잡는 시험 하나)
- read-only: `src/home_workdir.rs`(부모 SPEC 이 더한 `reports_dir`), `src/state_root.rs`, `src/plan_input` 관련
- forbidden: 실제 HOME 에 쓰는 시험, 네트워크, 추적 파일 이동·삭제

## Plan

1. 부모 SPEC 의 `reports_dir` 가 있는 가지에서 시작해요(없으면 먼저 부모 조각을 받아요).
2. RED: `tests/reports_move_cli.rs` 의 네 경우(옮김·추적 유지·이름 충돌·재실행)와 권한·digest·단계·git 아님 시험을 먼저 써서 실패를 눈으로 봐요. 새 시험이 git 을 띄우면 `tests/common/git.rs` 도우미로만 해요(환경을 씻어요, 리더 공지).
3. GREEN: `src/reports_move.rs`(plan/apply), 두 연산 파일과 문서 조각, 도움말 표·예시·골든.
4. `project_update.rs` 의 `reports` 단계와 `update.md` 지침 한 줄, 매니페스트 해시.
5. 시험을 돌려요: 위 verify 명령들, fmt, ops rustfmt, clippy `--all-targets`, `project_update` 를 읽는 기존 시험, `--bins`.

## Risks

- 옮김은 실제 HOME 에 쓰는 길이에요: 목적지 이름이 있으면 건너뛰고(덮어쓰기 없음) 추적 파일은 건드리지 않으며 시험은 전부 임시 HOME 에서만 돌려요.
- `project-update-plan` 에 단계를 더하면 단계 목록·골든을 읽는 시험이 달라질 수 있어요: 먼저 그 시험들을 찾아 같이 고쳐요.
- 새 연산 두 개는 도움말 표·예시·문서 조각·골든 네 곳을 함께 건드려요(`operations_doc_matches_code`). 기준 가지에서 이미 실패하는 `--bins` 시험 5개와 섞이지 않게 전후 실패 목록을 따로 비교해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 옮김 대상은 프로젝트의 `.moai-cli/reports` 와 `.moai/reports` 안의 추적되지 않는 파일(무시 규칙이 걸린 파일 포함)이에요. `plan-audit/.gitkeep` 처럼 추적되는 파일은 자동으로 제자리에 남아요 | 사용자 지시와 `.gitignore` 의 `.moai/reports/` 블록 | 대상 폴더 줄이기 |
| D-2 | git 저장소가 아니거나 추적 여부를 알 수 없으면 `blocked` 로 두고 아무것도 옮기지 않는다 | 추적 파일은 그대로라는 지시를 어길 위험이 있는 쪽은 옮기지 않는 쪽이 안전해요 | 저장소 아님을 "모두 추적 안 됨"으로 |
| D-3 | `reports` 단계는 backup 단계(phase)의 마지막에 두고, 옮길 파일이 있으면 같은 단계의 `go-state` 명령이 `re-run after the reports apply` 로 기다리게 한다. install·merge 단계는 기다리게 하지 않는다 | 옮김은 `.moai/reports` 를 비워서 `go-state` 미리보기가 읽는 보고서 목록이 달라지므로(그 digest 가 낡아져요) `go-state` 는 새 미리보기를 기다려야 하고, install·merge 가 읽는 파일은 바뀌지 않아요 | install 단계로 옮기기 |
| D-4 | 같은 파일 시스템이면 하드 링크를 만든 뒤 원본 이름을 지워서 옮기고(목적지 이름이 이미 있으면 링크가 실패해서 덮어쓰지 않아요), 아니면 `create_new` 로 복사해 바이트가 같을 때만 원본을 지운다 | 하드 링크는 이름이 있으면 실패해서 덮어쓰기가 없고, 복사는 원본을 확인 뒤에만 지워요 | 항상 복사 |
| D-6 | 링크는 따라가지 않아요: 이른 보고서 폴더로 가는 길의 어느 한 칸이 링크이고 그 너머에 보고서 폴더가 있으면 `blocked`, 중첩 저장소 안이면 `blocked`; 보고서 폴더가 없는 링크는 무시해요 | 링크 너머는 다른 저장소일 수 있고 이 저장소의 git 은 그곳의 추적 파일을 몰라서 지우게 돼요 | 링크 허용 규칙 더하기 |
| D-7 | 연결된 작업 트리의 목적지는 `git worktree list` 의 첫 항목(primary)의 이름으로 정한다 | `optimize-plan` 이 같은 선택을 해서 두 연산이 같은 폴더를 가리켜요 | 작업 트리 이름 쓰기 |
| D-5 | 심볼릭 링크와 폴더가 아닌 것은 옮기지 않고 수만 알린다 | 링크를 따라가면 저장소 밖 파일을 옮길 수 있어요 | 링크 처리 규칙 더하기 |

## Result

구현됨(2026-11-20, 가지 `spec/20261119-reports-outside-repo`).

- AC-1: met — `no_earlier_folder_is_unchanged_and_preview_counts_are_reported`(`tests/reports_move_cli.rs`) 통과
- AC-2: met — `untracked_report_files_move_with_their_folders_and_tracked_ones_stay` 통과
- AC-3: met — `a_name_the_destination_holds_is_kept_and_counted_never_overwritten` 통과
- AC-4: met — `a_second_run_moves_nothing_and_changes_nothing` 통과
- AC-5: met — `moved_files_and_folders_are_owner_only` 통과(Unix)
- AC-6: met — `a_plan_made_stale_by_a_new_file_or_a_wrong_digest_or_reference_is_refused` 통과
- AC-7: met — `update_plan_shows_the_reports_step` 통과
- AC-8: met — `a_folder_git_cannot_read_blocks_the_move_and_nothing_is_created` 통과
- AC-9: met — 여섯 시험 통과: `a_linked_state_folder_is_never_followed_into_another_repository`, `a_linked_state_folder_inside_the_same_repository_blocks_the_move`, `a_report_folder_inside_a_nested_repository_is_never_moved`, `a_home_that_is_relative_or_inside_the_project_blocks_the_move`, `a_destination_that_is_a_link_or_a_file_blocks_the_move`, 그리고 20,000개 계획·20,001개 막힘 경계 `the_file_limit_holds_at_n_and_blocks_at_n_plus_one`
- AC-10: met — `a_linked_worktree_moves_to_the_primary_checkouts_folder` 통과
- 위 AC-1..AC-8 시험은 head `0a6b335558` 의 영역 접두 대상 실행(`reports_*` 포함 63개, `cargo test` 종료 코드 0)에서 돌았어요. 감사가 찾은 결함을 고친 뒤(AC-9·AC-10 과 `project_update_plan_cli` 의 `a_ready_reports_step_holds_the_go_state_command_until_it_is_applied`)의 확인은 `moai check affected` 요약 줄이에요(노트에 있어요).

- 두 연산 `reports-move-plan`(읽기 전용 미리보기)과 `reports-move-apply`(digest·permission reference 를 확인하고 옮김)는 `src/reports_move.rs` 의 한 곳을 써요. 옛 폴더 이름(`.moai-cli/reports`, `.moai/reports`)은 `state_root` 의 이름 상수로 적어요(`project_state_dir_is_spelled_once` 와 허용 목록 시험 둘 다 통과).
- `project-update-plan` 에 `reports` 단계가 수와 목적지를 보여 주고 apply 줄은 같은 미리보기의 digest 를 써요. 도움말 표·예시와 골든 네 개에 두 연산 줄이 들어갔고(한 줄 설명은 56자), 연산 문서 조각 두 개를 더했어요.
- AC-1~AC-8 은 `tests/reports_move_cli.rs`·`tests/project_update_plan_cli.rs` 가 확인해요.
- 증거(head `0a6b335558`, 모두 직접 돌린 종료 코드): diff-check 0, `cargo fmt --check` 0, clippy `--all-targets -D warnings` 0, `cargo test --lib --bins` 와 지정 대상 0, `product_cli skill_package` 0, 파이썬 5개 모듈 0, trailer-check 0, spec check 둘 0. 영역 접두 대상 63개(`go_*`·`help_*`·`home_*`·`optimize_*`·`project_*`·`reports_*`·`update_*` 와 macOS 전용 시험) `cargo test` 0, `tools/repo_scans.json` 의 검사 39개 0, 영역별 파이썬 9개 모듈 0.
- 좁힌 시험이 놓친 것을 넓은 시험이 찾았고 고쳤어요: 철자 시험, 한 줄 설명 61자, `AGENTS.md` 크기 한도, `home_retire_cli` 의 `reports` 기대값.
- 돌리지 않은 것: 병합된 트리에서의 골든 실행, CI(이 저장소의 PR CI 는 리눅스 x86_64 만 돌아요).
