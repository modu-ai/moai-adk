---
id: SPEC-20261124-lane-worktree-command
title: "One command that makes a lane's worktree in the primary checkout and reads it back"
status: draft
revision: 1
created: 2026-11-26
branch: spec/lane-worktree-command-20261124
---

# One command that makes a lane's worktree in the primary checkout and reads it back

## Interview

- Q1: 무엇을 만드나요? → A: 리더 배정 글과 카드 본문: moai 명령 하나가 레인 워크트리를 본 체크아웃의 `.claude/worktrees/<이름>` 에 절대 경로로 만들고, `git worktree list` 로 다시 읽어 확인한 뒤 그 경로를 출력해요. 상대 경로나 중첩은 거절해요.
- Q2: 왜 필요한가요? → A: 2026-11-24 에 레인 두 곳이 이전 워크트리 안에서 상대 경로로 `git worktree add` 를 해서 워크트리가 중첩으로 만들어졌고, 없는 경로로 `EnterWorktree` 를 불러 권한 질문에 멈췄어요. 글(work.md)만으로는 잘못 읽으면 다시 생기므로, 같은 일을 명령 하나로 만들어 재발을 막아요(재발 방지 규칙).
- Q3: 어디까지 하나요? → A: 만들기와 다시 읽기만 해요. 지우기, 강제 옵션, 가지치기, 옮기기는 넣지 않아요(리더 배정 글).

## Intent

`moai worktree-new <이름> --branch <브랜치>` 는 통합 브랜치(또는 `--base` 로 준 커밋)에서 새 브랜치를 따서 `<본 체크아웃>/.claude/worktrees/<이름>` 에 워크트리를 만들어요. 본 체크아웃은 `git` 의 공통 디렉터리에서 읽기 때문에 이 명령을 어느 워크트리 안에서 불러도 새 폴더는 항상 본 체크아웃 아래에 생겨요. 만든 뒤 `git worktree list --porcelain` 을 다시 읽어 그 경로와 브랜치와 커밋이 맞고 폴더가 있을 때만 절대 경로를 출력해요. 출력된 경로를 `EnterWorktree` 에 그대로 넘기면 호스트가 권한 질문을 하지 않아요(없는 경로로 들어갈 때만 질문이 나와요).

이미 있는 명령은 이 일을 덮지 않아요: `worktree-plan`, `worktree-apply` 는 클레임(이슈)용이고, `worktree-slot ensure` 는 고정 자리를 분리된 HEAD 로 만들어요. 통합 브랜치에서 새 브랜치로 만드는 레인 카드용 명령은 없어요.

읽는 사람: 이 변경을 판단하는 리더, 구현하는 레인, 독립 감사자, 그리고 work.md 를 따르는 레인.

## Non-goals

- 워크트리를 지우거나, 가지치기하거나, 옮기거나, 강제로 만드는 일은 하지 않아요. 그런 옵션은 없어요.
- 브랜치 이름을 추측하지 않아요. `--branch` 는 필수예요(날짜 SPEC 브랜치는 SPEC 날짜와 맞아야 해서 명령이 고르지 않아요).
- `EnterWorktree` 를 대신 부르지 않아요. 명령은 호스트 도구를 만지지 않고, 레인이 출력된 경로로 직접 들어가요.
- 고정 자리(`worktree-slot`)와 정리(`worktree-gc-*`)의 규칙은 바꾸지 않아요.

## Requirements

- REQ-1: When `moai worktree-new <name> --branch <branch>` is run in a repository, the command shall create the worktree at `<primary checkout>/.claude/worktrees/<name>` on the new branch cut from the integration branch, and print its absolute path.
- REQ-2: When the add has finished, the command shall read `git worktree list --porcelain` again and print the path only when that list holds the path on the new branch at the base commit and the folder exists.
- REQ-3 (REQ-1): IF the name is empty, longer than 64 characters, holds a character other than a lower-case letter, a digit or `-`, starts with `-`, or holds a separator or a `.` or `..` part, THEN the command shall refuse and create nothing.
- REQ-4 (REQ-1): IF the target path already exists, the branch already exists, the base names no commit, the place runs through a link, or no base can be known (no `--base` and no integration branch), THEN the command shall refuse and create nothing.
- REQ-5: When the command is run from inside a linked worktree, the new worktree shall still be at `<primary checkout>/.claude/worktrees/<name>`, never below the folder the command was run in. · failure: none — the folder comes from the git common directory, so the folder the command is run in cannot move it.
- REQ-6: The command shall have no option that forces, deletes, prunes or moves a worktree.
- REQ-7: The `work.md` step that makes a card's worktree shall name `moai worktree-new` as the way to make it.
- REQ-8 (REQ-2): IF the list read after the add does not hold the path on the new branch at the base commit, or the folder is missing, THEN the command shall fail, name the path it left in place and delete nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary repository with an integration branch in its profile, `worktree-new lane-a --branch spec/x-20261124` makes `.claude/worktrees/lane-a` on that branch at the integration tip, prints the absolute path, and `git worktree list` holds it · verify: `cargo test --offline --locked --test worktree_new_cli ac1_`
- AC-2 (REQ-3): names with a separator, `..`, `.`, upper case, a leading `-`, a space, an empty name or 65 characters are refused and no folder or branch appears · verify: `cargo test --offline --locked --test worktree_new_cli ac2_`
- AC-3 (REQ-4): an existing folder, an existing branch, an unknown `--base` and a profile with no integration branch (without `--base`) are each refused and change nothing; a link at `.claude/worktrees` is refused · verify: `cargo test --offline --locked --test worktree_new_cli ac3_`
- AC-4 (REQ-5): run from inside a linked worktree, the new worktree is under the primary checkout and the answer's path is absolute · verify: `cargo test --offline --locked --test worktree_new_cli ac4_`
- AC-5 (REQ-2, REQ-8): a worktree that git lists at another path or branch than the one added is reported as a failure that names the path and leaves it (the read-back check, run on a throwaway repository with a `git` shim) · verify: `cargo test --offline --locked --test worktree_new_cli ac5_`
- AC-6 (REQ-6): the help text names no `--force`, `--remove`, `--delete`, `--prune` or `--move` option · verify: `cargo test --offline --locked --test worktree_new_cli ac6_`
- AC-7 (REQ-7): `skills/moai/workflows/work.md` names `moai worktree-new` in its worktree step and in the paragraph that makes a card's worktree · verify: `cargo test --offline --locked --test worktree_new_cli ac7_`

## Scope

- write: `.moai/specs/SPEC-20261124-lane-worktree-command/**`, `src/worktree_new.rs`, `src/ops/worktree_new.rs`, `docs/design-review/appendix/operations/worktree_new.md`, `tests/worktree_new_cli.rs`
- write (한 줄, line-lease): `src/lib.rs`, `src/help.rs`, `tests/fixtures/help/*.txt`, `tools/repo_scans.json`, `skills/moai/references/source-manifest.json`, `skills/moai/workflows/work.md`
- read-only: `src/worktree_slot.rs`, `src/home_workdir.rs`, `src/card/landed.rs`
- forbidden: 워크트리 삭제·강제·가지치기 경로, 실제 HOME, 네트워크, Python

## Decisions

- D-1: 이름은 `moai worktree-new <이름> --branch <브랜치> [--base <ref>] [--project <폴더>]` 예요. 이름은 폴더 이름 하나만 받아서 상대 경로와 중첩이 문장으로 만들어질 수 없어요(경로를 받지 않으므로 거절이 아니라 구조로 막아요; 경로처럼 생긴 이름은 거절해요).
- D-2: 기본 base 는 통합 브랜치의 지금 커밋이에요(`card::landed::integration_branch` 와 `tip_of`). 통합 브랜치를 알 수 없고 `--base` 도 없으면 거절해요. 브랜치 이름은 `git check-ref-format --branch` 로 확인해요.
- D-3: 새 폴더 자리는 `worktree_slot::primary_of(common_dir)` 로 읽은 본 체크아웃 아래 `.claude/worktrees/<이름>` 이에요. 자리에 연결(링크)이 있으면 `home_workdir::place_verdict` 로 거절하고, 지역 제외 파일(`ensure_place_ignored`)은 이미 있는 부품으로 맞춰요.
- D-4: `git worktree add -b <브랜치> <절대 경로> <커밋>` 한 번만 부르고, 실패하면 아무것도 지우지 않아요. 다시 읽기가 맞지 않으면 만든 폴더를 그대로 두고 경로를 알려요(지우기가 없기 때문이에요).
- D-5: 이 명령은 `EnterWorktree` 안에서 불러도 같은 자리에 만들어요. 다만 호스트는 `EnterWorktree` 로 들어온 세션에서 본 체크아웃 쪽 `git worktree add` 를 직접 막으므로, work.md 의 순서(먼저 `ExitWorktree(keep)`)는 그대로 두고 명령 이름만 더해요.

## Plan

1. RED: `tests/worktree_new_cli.rs` 에 ac1_ … ac7_ 를 쓰고 실패를 기록해요.
2. GREEN: `src/worktree_new.rs`, `src/ops/worktree_new.rs`, 도움말 조각, `src/lib.rs` 한 줄, `src/help.rs` 두 줄과 도움말 골든 파일을 더해요.
3. work.md 의 두 곳이 명령을 가리키게 하고 매니페스트 해시를 맞춰요.
4. 영수증: `cargo fmt --all`, `check affected --card`, READY.

## Risks

- R-1: 스캔 표와 도움말 골든 파일은 공유 파일이라 줄 임대로만 써요. 한 줄씩만 건드리고 바로 돌려줘요.
- R-2: `git worktree add` 가 만든 폴더를 다시 읽기가 못 찾는 드문 경우에는 폴더가 남아요. 지우기는 사람이나 `worktree-gc-*` 의 몫이고, 이 명령은 경로를 알려요.
- R-3: work.md 는 다른 카드가 쓰기 범위로 쥐고 있을 수 있어요. 줄 임대로 두 곳만 고치고 바로 돌려줘요.

## Result

구현 커밋 `ba32f145d3`, 거절 분류를 맞춘 커밋 `e7c2453261`, RED 커밋 `7585e0c41f` 위, NeXT `398b66bda0` 에서 갈라졌어요. 아래는 이 실행에서 본 값이에요.

- RED: 구현 전 `cargo test --offline --locked --test worktree_new_cli` 는 7개 모두 실패했어요(명령이 없었어요, `target/wt/red.log`). 첫 GREEN 에서 5개가 실패했어요: 다시 읽기가 `refs/heads/` 접두사가 붙은 브랜치 이름과 비교했는데 파서는 접두사를 떼어 줘요. 고친 뒤 7개 통과, 0개 실패(`target/wt/green4.log`). 이름 단위 시험 1개 통과(`target/wt/green3.log`).
- AC-1: met — `ac1_the_worktree_is_made_on_a_new_branch_at_the_integration_tip`
- AC-2: met — `ac2_a_name_that_is_not_one_plain_folder_name_is_refused`
- AC-3: met — `ac3_a_taken_name_an_unknown_base_and_a_linked_place_are_refused` (거절은 `invalid_input`)
- AC-4: met — `ac4_run_from_a_linked_worktree_the_new_one_is_in_the_primary_place`
- AC-5: met — `ac5_a_read_back_that_does_not_match_fails_and_names_the_path`
- AC-6: met — `ac6_the_help_has_no_option_that_forces_deletes_prunes_or_moves`
- AC-7: met — `ac7_work_md_names_the_command_in_both_worktree_places`
- 위험 경로 실행(임시 저장소, 만든 바이너리, `target/wt/throwaway2.log`): 정상 생성, `../escape`·`a/b`·절대 경로·이미 있는 폴더는 거절(exit 1, 아무것도 안 만듦), 워크트리 안에서 불러도 본 체크아웃의 `.claude/worktrees/lane-b` 에 생성, `.claude/worktrees` 가 링크인 저장소는 거절이고 링크 뒤에는 항목 0개·새 브랜치 없음.
- 같은 날 세션에서 본 증거: 워크트리에 격리된 세션에서 본 체크아웃 아래로 `git worktree add` 를 부르면 호스트가 "gives worktree add a path in the shared checkout … Refusing to run it" 로 거절했고, `ExitWorktree(keep)` 뒤에는 통과했어요. 그래서 work.md 의 순서(먼저 나오기)는 그대로 두었어요.
- 지우기·강제·가지치기·옮기기 옵션과 코드는 없어요. work.md 는 고정된 다섯 문장을 그대로 두고 명령을 가리키는 문장만 더했고, 도움말 표·골든 파일·스캔 표·매니페스트 해시를 줄 임대로 맞췄어요.
- 남은 것: 영수증은 READY 노트에 있어요. 독립 감사와 넓은 `check affected` 는 아직이에요. Windows 와 Linux 는 Mac 에서만 돌렸어요.
