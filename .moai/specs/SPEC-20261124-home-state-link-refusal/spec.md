---
id: SPEC-20261124-home-state-link-refusal
title: "홈 상태 뿌리와 쓰는 폴더 사이의 심볼릭 링크를 거절해요: 클레임 저널 뿌리와 worktree 이동 뿌리"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7908 (2026-11-23T21:22:01Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 7906 에 답함: 판 1 승인 — D-1 검사 함수 하나, D-2 문장, D-3 이동 뿌리 범위, D-5 N1·R1 비범위, D-6 쓰기 범위; 승인 커밋 ca269cda30); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:dd49e7494cfb5feabc9274e6e958437945f80135c849e13a74ee8f79ee3dfea4"
---

# 홈 상태 뿌리와 쓰는 폴더 사이의 심볼릭 링크를 거절해요: 클레임 저널 뿌리와 worktree 이동 뿌리

## Intent

카드 C-427 은 감사 #409(조각 2a)가 남긴 비차단 N1·N2 를 닫아요. 두 곳이 홈 상태 아래에 쓰는데, 뿌리와 쓰는 폴더 사이의 링크를 보지 않아요.

- N1(클레임 저널): `claim::journal_root_in` 은 `layout.json` 이 저널을 홈 상태 아래로 돌려 놓았을 때 홈 상태 뿌리(`<HOME>/.moai` 나 `<HOME>/.moai-cli`) 하나만 링크인지 봐요(`src/claim.rs:104-107`). 뿌리는 진짜 폴더이고 그 아래 `projects/` 나 `projects/<id>/` 가 링크이면 거절하지 않고, 이어서 `Journal::open`(`src/claim.rs:825`)이 마지막 이름과 `.lock` 만 보고 링크 너머에 폴더와 잠금 파일을 만들어요(감사 #409 프로브 S3). 링크 너머에 `layout.json` 이 있으면 `active_home` 이 그 링크를 따라 읽어서 이 길이 열려요.
- N2(worktree 이동): `worktree_move::plan` 의 이동 뿌리 `<HOME>/.moai/worktrees`(`default_root`, 또는 그 아래를 가리키는 `--root`)에 링크가 있는지 아무도 보지 않아요. `apply` 는 `plan` 을 다시 돌려 그 뿌리 아래에 `create_dir_all` 과 `git worktree move` 를 해요. 지금 막히는 것은 우연한 보호예요(감사 #409 N2).

둘 다 예전부터 있던 틈이고 회귀가 아니에요. 이 SPEC 은 "홈 상태 뿌리와 쓰는 폴더 사이의 모든 링크 구성요소를 같은 문장 모양으로 거절"하는 한 가지 검사를 `src/layout.rs` 에 두고 두 곳이 그것을 불러요. 새 검사는 읽기만 하고 아무것도 쓰지 않아요.

## Non-goals

- `src/forge/outbox.rs`·`src/apply.rs` 의 `switched_dir` 호출과 `layout-apply` 의 복사(`copy_verified`)도 홈 상태 아래에 써요. 이 카드의 쓰기 범위 밖이라 같은 검사를 부르는 일은 따로 카드로 열어요(Risks).
- `Journal::open` 의 검사를 바꾸지 않아요. 그 함수는 홈 상태 뿌리를 몰라서 받은 경로의 마지막 이름만 볼 수 있어요. 홈 상태 아래 경로는 `journal_root_in` 이 길 전체를 본 뒤에 돌려줘요.
- N1(감사 #410, `journal_root_in` 의 `identity_in(project, None)` 이 환경 HOME 의 git 설정을 읽음)과 R1(home-env-readers SPEC 판 4 Risks 의 끝 슬래시 문장)은 이 카드에서 하지 않아요(D-5).
- 홈 상태 밖의 사용자가 고른 `--root`(예 `/tmp/x`)는 거절하지 않아요. macOS 의 `/var`·`/tmp` 는 링크라서 그 규칙은 정상 경로를 막아요(D-3).

## Requirements

- REQ-1: The `refuse_linked_home_path` function shall refuse a path when the home state root it is given, or a folder between that root and the path, is a symbolic link.
- REQ-2: IF the path lies outside the home state root, THEN `refuse_linked_home_path` shall accept it without reading any link.
- REQ-3: The `refuse_linked_home_path` function shall answer `home state root is a symlink; refused` for the root and `home state folder is a symlink; refused` for a folder below it.
- REQ-4: The `journal_root_in` function shall refuse a switched claim journal folder through `refuse_linked_home_path` before it returns the folder.
- REQ-5: IF no layout pointer switches the journal, THEN `journal_root_in` shall return the clone side folder and refuse no link of the home state.
- REQ-6: The `worktree_move::plan` function shall refuse a move root below `<HOME>/.moai` when `.moai` or a folder between it and the root is a symbolic link.
- REQ-7: IF the move root lies outside `<HOME>/.moai`, THEN `worktree_move::plan` shall not refuse it for a link.
- REQ-8: The `worktree_move::apply` function shall refuse such a move root and move no worktree.

## Acceptance

- AC-1 (REQ-1, REQ-3): 홈 뿌리 자신, 뿌리 바로 아래 폴더, 두 단계 아래 폴더, 쓰는 폴더 자신이 각각 링크일 때 `refuse_linked_home_path` 가 거절하고, 뿌리는 `home state root is a symlink; refused` 로, 아래 폴더는 `home state folder is a symlink; refused` 로 답해요. 링크가 없는 같은 모양의 폴더는 통과해요. · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(refuse_linked_home_path)'`
- AC-2 (REQ-2): 홈 상태 뿌리 밖의 경로는 그 경로의 앞쪽 구성요소가 링크여도 통과해요. 끝 슬래시를 더하거나 `.` 를 낀 같은 경로도 같은 답이에요(경로 구성요소 비교). · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(refuse_linked_home_path_ignores)'`
- AC-3 (REQ-4): 홈 상태 뿌리는 진짜 폴더이고 `projects/` 가, `projects/<id>/` 가, `claims` 가 링크인 세 경우 모두 `journal_root_in` 이 전환된 폴더를 돌려주기 전에 거절하고, 링크 너머에 아무것도 생기지 않아요. · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(journal_root_in_refuses)'`
- AC-4 (REQ-4): 환경 HOME 만 바꾼 자식 프로세스에서 `claim::journal_root` 와 쓰기 진입점 `worktree::open_locked_journal` 이 같은 거절을 하고, 링크 너머 폴더에 `.lock` 도 폴더도 안 생겨요. · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(journal_root_refuses)'`
- AC-5 (REQ-5): `projects/` 가 링크여도 그 너머와 홈 상태 아래 어디에도 `layout.json` 이 없으면 `journal_root_in` 은 클론 쪽 폴더를 돌려주고 거절하지 않아요(감사 #409 판 4 의 경계). 링크 없는 전환 트리는 홈 쪽 폴더를 돌려줘요(대조). · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(journal_root_in_keeps)'`
- AC-6 (REQ-6): `<HOME>/.moai` 가 링크인 경우, `<HOME>/.moai/worktrees` 가 링크인 경우, `--root <HOME>/.moai/worktrees/sub/x` 의 `sub` 가 링크인 경우에 `worktree-move-plan` 이 거절 문장으로 실패하고 링크 너머에 아무것도 생기지 않아요. · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(worktree_move_plan_refuses)'`
- AC-7 (REQ-7): `--root` 가 `<HOME>/.moai` 밖이고 그 앞쪽 구성요소가 링크이면 `worktree-move-plan` 이 이전처럼 계획을 돌려줘요. · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(worktree_move_plan_accepts)'`
- AC-8 (REQ-8): 링크가 낀 이동 뿌리로 `worktree-move-apply` 를 부르면 거절하고 어떤 worktree 도 움직이지 않으며 링크 너머에 폴더가 안 생겨요. · verify: `cargo nextest run --offline --locked --test home_state_link_refusal -E 'test(worktree_move_apply_refuses)'`
- AC-9 (REQ-1..REQ-8): 이 SPEC 의 변경 경로가 Scope write 목록 안이에요. · verify: `git diff --stat <base>..<head>`
- AC-10 (REQ-1..REQ-8): 이웃 테스트가 그대로 통과해요: `layout`·`claim`·`worktree` 모듈의 기존 테스트와 `worktree_move` 통합 테스트. · verify: `cargo nextest run --offline --locked --lib -E 'test(journal_root) | test(layout::) | test(worktree_move)'` 와 `cargo nextest run --offline --locked --test worktree_move`

## Scope

- write: `src/layout.rs`, `src/claim.rs`, `src/worktree_move.rs`, `tests/home_state_link_refusal.rs`, `.moai/specs/SPEC-20261124-home-state-link-refusal/spec.md`
- read-only: `src/identity.rs`, `src/worktree.rs`, `src/home_root.rs`, `src/home_workdir.rs`, `src/contract.rs`, `src/forge/outbox.rs`, `src/apply.rs`, `.moai/specs/SPEC-20261124-home-env-readers/spec.md`, `.moai/specs/SPEC-20261124-identity-home-arg/spec.md`
- forbidden: `CLAUDE.md`

`src/layout.rs` 는 카드 C-410 이 쥐었다가 판 4 가 착지한 파일이에요(NeXT `fb97c6af9a` 에 합쳐진 것을 읽었어요). 이 카드는 그 파일에 함수 하나와 테스트를 더할 뿐 기존 줄은 바꾸지 않아요. `tests/areas.toml` 은 바꾸지 않아요(새 파일의 첫 낱말 `home` 이 이미 `home` 영역에 있어요). 실제 HOME 은 읽기 전후 비교로도 건드리지 않아요: 테스트는 임시 HOME 만 써요. 네트워크도 쓰지 않아요.

## Plan

1. 승인을 리더 노트로 받아요(승인은 쓰지 않아요).
2. RED: `tests/home_state_link_refusal.rs` 에 AC-1..AC-8 테스트를 쓰고 `refuse_linked_home_path` 가 없어 컴파일이 실패하는 것을 봐요. 함수만 뼈대로 더해 컴파일하고(`Ok(())`) 거절 테스트가 실패하는 것(exit 100)을 봐요.
3. GREEN: `src/layout.rs` 에 `refuse_linked_home_path(home_root, path)` 를 구현해요. 뿌리가 `path` 의 앞쪽이 아니면 `Ok`, 맞으면 `identity::refuse_symlink(root, "home state root")` 뒤 뿌리 아래 구성요소를 하나씩 `"home state folder"` 로 봐요.
4. `src/claim.rs` `journal_root_in`: 전환된 폴더가 있으면 기존 `identity::refuse_symlink(&root, "home state root")` 호출을 `refuse_linked_home_path(&root, &switched)` 로 바꿔요.
5. `src/worktree_move.rs` `plan`: `check_root` 바로 뒤에, 홈이 있고 이동 뿌리가 `home_workdir::root(home)` 아래이면 `refuse_linked_home_path` 를 불러요. `apply` 는 `plan` 을 다시 돌리니 같은 거절을 받아요.
6. guard-mutant: 테스트마다 거절 호출 하나를 지우고 그 필터가 exit 100 으로 실패하는 것을 봐요(복구는 `git checkout` 이나 `cp`, `cp -p` 는 쓰지 않아요).
7. `moai check affected <base>..<head>` 를 맨몸으로 한 번 돌려요.

## Risks

- 검사와 쓰기 사이에 링크를 바꿔 끼우는 경쟁은 막지 못해요. `Journal::open` 의 문서가 이미 이를 알려진 잔여 위험으로 적었고(`src/claim.rs:823`), 이 검사도 같아요.
- Windows 의 junction 은 `symlink_metadata().is_symlink()` 가 아니라서 이 검사가 못 잡아요. Windows 결과는 not_run 이에요.
- 같은 틈이 `src/forge/outbox.rs`·`src/apply.rs`(`switched_dir`)와 `layout-apply` 의 복사에도 있어요. 이 카드는 닿지 않아요. `refuse_linked_home_path` 가 `pub` 이라 그 카드가 같은 함수를 불러요. 이 SPEC 이 승인되면 리더가 후속 카드를 열지 정해요(unverified: 두 모듈이 홈 상태 아래에 쓰는 줄은 이 SPEC 에서 읽지 않았어요).
- R2(감사 #410): 같은 HOME 의 다른 철자(`..`, 링크, 대소문자)를 `Path` 구성요소 비교는 같은 HOME 으로 보지 않아요. 이 검사는 `strip_prefix` 로 구성요소를 비교하므로 끝 슬래시와 `.` 는 같은 HOME 으로 보고, 다른 철자는 "뿌리 밖"으로 보아 검사를 건너뛰어요. 받은 HOME 을 밖에서 얻는 호출자가 생기면 정규화 비교가 필요해요(후속, 이 카드는 하지 않아요).
- 동작이 바뀌는 곳은 하나예요: `projects/` 나 `projects/<id>/` 가 링크이고 `layout.json` 이 저널을 홈 상태로 돌려 놓은 사용자는, 전에 열리던 저널이 거절돼요. 뿌리 링크는 이미 거절이었고, 이 SPEC 은 같은 거절을 뿌리 아래로 넓혀요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 검사는 `src/layout.rs` 의 `pub fn refuse_linked_home_path(home_root: &Path, path: &Path) -> Result<()>` 하나예요. 두 호출부가 같은 함수를 불러요 | 홈 상태 뿌리를 아는 모듈이 `layout`(`default_home_root_in`)이고, `claim` 과 `worktree_move` 가 이미 `layout` 이나 `home_workdir` 를 불러요. `identity::refuse_symlink` 가 한 이름을 보는 기존 도우미라 그것을 부려요 | 함수를 지우고 두 호출부를 이전 줄로 되돌리기 |
| D-2 | 문장 모양은 `<무엇> is a symlink; refused` 예요: 뿌리는 기존 `home state root`, 뿌리 아래는 새 `home state folder` | `contract::classify` 가 `refused` 가 든 문장을 InvalidInput 으로 분류해서 분류표를 바꾸지 않아요. 기존 문장 둘(`home state root`, `project state root`)이 같은 모양이에요 | 낱말 하나 바꾸기 |
| D-3 | 이동 뿌리는 `home_workdir::root(home)`(`<HOME>/.moai`) 아래일 때만 검사해요; `.moai` 자신부터 봐요. 그 밖의 `--root` 는 검사하지 않아요 | 이동 뿌리 기본값이 `<HOME>/.moai/worktrees`(`worktree_move::default_root`)이고, 홈 밖 경로는 macOS `/var`·`/tmp` 가 링크라서 거절하면 정상 경로가 막혀요. HOME 자신은 보지 않아요(임시 HOME 이 `/var/folders` 아래라서 테스트가 깨져요) | 호출 줄을 지우기 |
| D-4 | worktree 쪽 검사는 `plan` 안에 한 번 둬요 | `apply` 가 먼저 `plan` 을 다시 돌려서 한 곳이 둘 다 막고, `plan` 은 아무것도 쓰지 않아서 거절이 쓰기 전에 와요 | 호출 줄을 지우기 |
| D-5 | N1(`journal_root_in` 의 git 설정 읽기)과 R1(home-env-readers SPEC 문장)은 이 카드 범위에 넣지 않아요 | 리더 배차가 R1 은 home-env-readers SPEC 을 범위에 넣을 때만 하라고 했고 넣지 않았어요. N1 은 `identity_in` 의 HOME 인자를 바꾸는 일이라 링크 거절(D-7 의 경계)과 얽혀 따로 보는 편이 안전해요 | 해당 없음 |
| D-6 | 쓰기 범위에 `src/worktree_move.rs` 를 더해요. `src/worktree.rs`·`src/identity.rs` 는 예약 범위에 있지만 바꾸지 않아요 | 배차가 적은 `src/worktree.rs` 에는 이동 뿌리 검사가 없고(`refuse_in_build_folder` 만 있어요) 이동 뿌리 코드는 `src/worktree_move.rs`(`default_root`, `plan`, `apply`)에 있어요. `identity::refuse_symlink` 는 이미 `pub(crate)` 라 `identity.rs` 도 그대로예요 | 해당 없음 |

## Interview

- Q1: `Journal::open` 에서 막으면 안 되나요? → A: 안 돼요. 그 함수는 받은 경로의 홈 상태 뿌리를 몰라서(클론 쪽 저널에도 쓰여요) "뿌리 아래"를 가를 수 없어요. 홈 상태 아래 경로를 만드는 `journal_root_in` 이 가려요(D-1).
- Q2: 링크 대상 아래에 이미 있는 저널은 어떻게 되나요? → A: 거절돼요. 거절 문장이 바뀐 사용자는 링크를 풀거나(`layout.json` 을 지워 클론 쪽으로 돌리기) 링크 안의 내용을 진짜 폴더로 옮겨요.
- Q3: `claims` 자신이 링크인 경우도 같은 검사를 거치나요? → A: 네. 길의 마지막 구성요소도 봐요(AC-3). `Journal::open` 의 기존 마지막 이름 검사는 그대로 남아요.

## Result

(구현 뒤에 적어요.)
