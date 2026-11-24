---
id: SPEC-20261124-home-writer-links
title: "홈 상태 아래에 쓰는 나머지 세 곳도 심볼릭 링크 폴더를 거절해요: 아웃박스, 영수증, layout-apply"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8120 (2026-11-24T02:29:47Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 8118 에 답함: 판 1 승인 선택지 1, 읽기 길도 같은 문장으로 거절 D-2, claim.rs 는 건드리지 않아요; 승인 대상 커밋 b39b63b037); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:ee8d1424dd64e546ccdd539790f907cdf463dfa94d95dc445bfb14dafb685573"
---

# 홈 상태 아래에 쓰는 나머지 세 곳도 심볼릭 링크 폴더를 거절해요: 아웃박스, 영수증, layout-apply

## Intent

카드 C-434 는 SPEC-20261124-home-state-link-refusal(C-427)이 남긴 열린 질문을 닫아요. C-427 은 `layout::refuse_linked_home_path` 를 두고 클레임 저널 뿌리와 worktree 이동 뿌리만 그것을 불러요. 같은 홈 상태 아래에 쓰는 다른 곳이 뿌리와 쓰는 폴더 사이의 링크를 따라가는지는 읽지 않았어요(unverified). 이 SPEC 은 그 세 곳을 읽은 결과와 고치는 방법이에요.

읽은 결과(이 작업트리, NeXT `e15220f90f`):

- `crate::layout::switched_dir` 를 부르는 곳은 셋이에요: `src/claim.rs`(C-427 이 이미 검사), `src/forge/outbox.rs:202`, `src/apply.rs:368`. `layout::store_dir` 는 정의만 있고 부르는 곳이 없어요(`grep -rn "store_dir(" src` 가 `src/layout.rs:447` 하나만 답해요).
- 아웃박스(`outbox_dir`): 전환된 폴더(`<홈 상태>/projects/<id>/forge/outbox`)가 있으면 그 경로로 `create_dir_all` 하고 그대로 돌려줘요. 프로젝트 쪽 갈래에 있는 링크 거절이 이 갈래에는 없어요. `save` 가 그 폴더 아래에 임시 파일과 이름 바꾸기를 해요. 쓰기 틈이에요.
- 영수증(`receipts_dir`): 전환된 폴더를 돌려줘요. 그 폴더를 쓰는 `receipt_path`(저장), `save_before_bytes`, `before_bytes_path` 는 마지막 이름(`receipts` 또는 `<id>.before`)이 링크인지만 봐요. `projects/` 나 `projects/<id>/` 가 링크이면 거절 없이 링크 너머에 써요. 쓰기 틈이에요.
- `layout-apply`(`apply_at` → `check_destinations` → `copy_verified`, 표식 쓰기): 목적지 검사는 `projects/<id>/<store>` 가 이미 있는지만 봐요(`symlink_metadata`, 마지막 이름). 그 앞의 `projects/` 나 `projects/<id>/` 가 링크이면 `create_dir_all` 이 링크를 따라 링크 너머에 폴더를 만들고 파일을 복사하고 `layout.json` 을 써요. 쓰기 틈이에요. 이 쓰기는 사용자가 확정한 계획을 실행하는 것이라 링크 하나가 홈 상태 밖으로 새 데이터를 보내요.

세 곳은 같은 틈이고 회귀가 아니에요. 같은 검사를 같은 문장으로 불러요. 새로 만드는 것은 한 단계 도우미 하나뿐이에요.

## Non-goals

- `src/claim.rs`·`src/worktree_move.rs`(C-427 이 이미 닫았어요)와 `refuse_linked_home_path` 의 동작과 문장은 바꾸지 않아요.
- `copy_verified` 의 원본(프로젝트 쪽) 걷기가 원본 트리 안의 링크를 따라가는 것은 이 카드 밖이에요. 이 카드는 홈 상태 쪽 목적지만 봐요. unverified: 원본 쪽 링크 처리는 이 SPEC 에서 시험하지 않았어요.
- `src/archive.rs` 의 `copy_verified`(이름이 같은 다른 함수)는 홈 상태 뿌리 아래에 쓰는지 읽지 않았어요. 카드 본문이 가리킨 것은 `layout-apply` 의 것이에요(unverified, 후속 후보).
- Windows junction 은 `symlink_metadata().is_symlink()` 가 아니라서 이 검사가 못 잡아요(C-427 과 같은 한계). Windows 결과는 not_run 이에요.
- 검사와 쓰기 사이에 링크를 바꿔 끼우는 경쟁은 막지 못해요(C-427 과 같은 잔여 위험).

## Requirements

- REQ-1: The `switched_dir_checked` function shall refuse a switched store folder through `refuse_linked_home_path` before it returns the folder.
- REQ-2: IF no layout pointer switches the store, THEN `switched_dir_checked` shall answer no folder and refuse no link of the home state.
- REQ-3: The forge outbox shall refuse a switched outbox folder that is, or sits below, a symbolic link, for the read and for the write.
- REQ-4: The apply receipt store shall refuse a switched receipt folder that is, or sits below, a symbolic link, for the receipt, the before bytes and the recovery read.
- REQ-5: The `layout::apply_at` function shall refuse a home project folder that is, or sits below, a symbolic link before it copies or writes any file.
- REQ-6: IF the home state root of `layout::apply_at` does not exist yet, THEN the function shall copy the stores as before.

## Acceptance

- AC-1 (REQ-1, REQ-2): `switched_dir_checked` 가 전환 없음이면 `Ok(None)`, 링크 없는 전환이면 `Ok(Some(홈 쪽 폴더))`, `projects/` 가 링크인 전환이면 `home state folder is a symlink; refused` 로 거절해요. 환경 HOME 은 자식 프로세스에만 둬요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(switched_dir_checked)'`
- AC-2 (REQ-3): 환경 HOME 만 바꾼 자식에서 `forge::outbox::save` 와 `load` 가 `projects/` 링크, `projects/<id>/` 링크, `forge/outbox` 링크 셋 모두 `home state folder is a symlink; refused` 로 거절하고 링크 너머에 아무것도 생기지 않아요. 링크 없는 전환 트리는 쓰고 읽어요(대조). · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(outbox_refuses)'`
- AC-3 (REQ-4): 환경 HOME 만 바꾼 자식에서 `apply::Receipt::save`, `apply::save_before_bytes` 가 `projects/` 링크, `projects/<id>/` 링크, `receipts` 링크 셋 모두 같은 문장으로 거절하고 링크 너머에 아무것도 생기지 않아요. 링크 없는 전환 트리는 쓰고(대조) 복구(`apply::recover`)도 같은 거절을 받아요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(receipt_refuses)'`
- AC-4 (REQ-5): `layout::apply_at` 가 `projects/` 링크, `projects/<id>/` 링크일 때 거절하고(홈 상태 뿌리 자신이 링크이면 `home state root is a symlink; refused`), 링크 너머에 폴더도 파일도 `layout.json` 도 생기지 않으며 원본(프로젝트 쪽)이 그대로예요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(layout_apply_refuses)'`
- AC-5 (REQ-5, REQ-6): 링크 없는 트리(홈 뿌리와 `projects/` 가 아직 없는 첫 적용)는 이전처럼 복사하고 표식을 써요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(layout_apply_keeps)'`
- AC-6 (REQ-1..REQ-6): 이 SPEC 의 변경 경로가 Scope write 목록 안이에요. · verify: `git diff --stat <base>..<head>`
- AC-7 (REQ-1..REQ-6): 이웃 테스트가 그대로 통과해요: `layout`·`apply`·`forge::outbox` 모듈의 기존 테스트와 C-427 의 `home_state_link_refusal`. · verify: `cargo nextest run --offline --locked --lib -E 'test(layout::) | test(apply::) | test(forge::outbox)'` 와 `cargo nextest run --offline --locked --test home_state_link_refusal`

## Scope

- write: `src/layout.rs`, `src/forge/outbox.rs`, `src/apply.rs`, `tests/home_writer_links.rs`, `.moai/specs/SPEC-20261124-home-writer-links/spec.md`
- read-only: `src/claim.rs`, `src/worktree_move.rs`, `src/identity.rs`, `src/home_root.rs`, `src/forge/mod.rs`, `tests/home_state_link_refusal.rs`, `.moai/specs/SPEC-20261124-home-state-link-refusal/spec.md`
- forbidden: `CLAUDE.md`

`src/layout.rs` 는 여러 카드가 만진 공유 파일이에요. 이 카드는 함수 하나(`switched_dir_checked`, `switched_dir_in` 바로 아래)와 `check_destinations` 의 호출 한 줄만 더하고 기존 줄을 바꾸지 않아요. 한 줄은 `moai card line-lease` 로 짧게 쥐고 써요. `tests/areas.toml` 은 바꾸지 않아요(새 파일의 첫 낱말 `home` 이 이미 `home` 영역에 있어요). 실제 HOME 은 읽기 전후 비교로도 건드리지 않아요: 테스트는 임시 HOME 만 쓰고 환경 HOME 은 자식 프로세스에만 둬요. 네트워크도 쓰지 않아요.

## Plan

1. 승인을 리더 결정 노트로 받아요(승인은 쓰지 않아요).
2. RED: `tests/home_writer_links.rs` 에 AC-1..AC-5 테스트를 쓰고 `switched_dir_checked` 가 없어 컴파일이 실패하는 것을 봐요. 이어서 도우미만 거절하지 않는 뼈대로 더해 컴파일하고 거절 테스트가 실패하는 것(exit 100)을 봐요.
3. GREEN `src/layout.rs`: `pub fn switched_dir_checked(project, subdir) -> Result<Option<PathBuf>>` 를 더해요. HOME 을 한 번(`environment_home`) 읽고 `switched_dir_in` 의 답이 있으면 `default_home_root_in` 의 뿌리로 `refuse_linked_home_path` 를 불러요. `check_destinations` 는 `home_project` 에 같은 검사를 `layout.json` 읽기보다 먼저 불러요.
4. GREEN `src/forge/outbox.rs`: 전환된 갈래가 `switched_dir_checked` 를 불러요; 거절은 `ForgeError::invalid` 에 같은 문장을 그대로 실어요.
5. GREEN `src/apply.rs`: `receipts_dir` 가 `switched_dir_checked` 를 불러요. `receipt_path`·`save_before_bytes`·`before_bytes_path` 가 모두 `receipts_dir` 를 거치므로 한 곳이에요.
6. guard-mutant: 새 테스트마다 거절 호출 하나를 지우고 그 필터가 exit 100 으로 실패하는 것을 봐요(복구는 `git checkout`).
7. 위험 경로: 임시 HOME 아래 링크 폴더(자식 프로세스에만 HOME)로 세 곳을 실제로 돌려 링크 너머가 비어 있음을 READY 에 적어요.
8. `moai check affected <base>..<head>` 를 맨몸으로 한 번 돌려요.

## Risks

- 동작이 바뀌는 곳: `layout.json` 이 아웃박스·영수증 폴더를 홈 상태로 돌려 놓았고 `projects/`·`projects/<id>/`·그 폴더 자신이 링크인 사용자는 아웃박스·영수증이 거절돼요(C-427 이 저널에 한 것과 같아요). 홈 상태 뿌리 자신이 링크인 사용자(예 `~/.moai-cli` 를 dotfiles 링크로 둔 경우)의 `layout-apply` 도 거절돼요; 저널은 이미 그랬어요.
- 읽기도 거절해요: `outbox::load` 와 영수증 복구 읽기가 링크 낀 전환 폴더를 읽던 사용자는 거절돼요. 쓰기와 읽기가 같은 폴더를 가리키는 한 갈래라서 한 곳에서 거절하는 편이 단순해요(D-2).
- 한 번 더 부르는 비용: 경로 구성요소마다 `symlink_metadata` 한 번이에요.
- 환경 HOME 읽기: `switched_dir_checked` 는 `environment_home` 을 한 번 읽고 같은 값을 `switched_dir_in` 과 뿌리 계산에 써요(HOME 을 두 번 읽다 서로 다른 값을 받는 일이 없어요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 아웃박스와 영수증은 호출부마다 검사를 붙이지 않고 `layout::switched_dir_checked` 하나를 불러요 | 두 호출부가 같은 3줄(홈 뿌리 읽기, 검사, 오류 변환)을 되풀이하면 HOME 을 두 번 읽는 틈이 생겨요. 도우미가 HOME 을 한 번 읽어요. `claim.rs` 의 인라인 형태는 `_in` 변형을 쓰므로 그대로 둬요 | 도우미를 지우고 호출부를 이전 줄로 되돌리기 |
| D-2 | 영수증과 아웃박스는 읽기 길도 거절해요 | 영수증은 `receipts_dir` 한 곳이 저장·백업·복구를 모두 거쳐요. 쓰기만 가르려면 호출부마다 갈라야 해요. 거절 문장은 C-427 과 같아요(`home state folder is a symlink; refused`) | 호출 줄을 지우기 |
| D-3 | `layout-apply` 의 검사는 `check_destinations` 에 `home_project` 한 경로로 둬요 | `check_destinations` 는 `apply_at` 의 첫 검사이고 `apply_locked` 가 잠금 뒤에 다시 불러서 두 번 보아요. 대상 폴더(`<id>/receipts` 등)는 이미 있으면 충돌로 거절하므로 `home_project` 까지의 구성요소만 보면 충분해요. 표식(`layout.json`)도 `home_project` 안이에요 | 호출 줄을 지우기 |
| D-4 | `store_dir`(`layout.rs:447`)는 부르는 곳이 없어서 고치지 않아요 | `grep -rn "store_dir(" src` 가 정의 한 줄만 답했어요. 부르는 곳이 생기면 그 카드가 같은 검사를 불러요 | 해당 없음 |
| D-5 | 홈 뿌리가 아직 없는 첫 적용은 거절하지 않아요 | `refuse_linked_home_path` 는 없는 폴더를 링크 없음으로 봐요(C-427 판 1). 첫 적용이 막히면 옵트인이 시작도 못 해요 | 해당 없음 |

## Interview

- Q1: 영수증 `receipt_path` 의 기존 마지막 이름 검사는 어떻게 되나요? → A: 그대로 남아요. 이 SPEC 은 그 앞에 뿌리부터 마지막 폴더까지의 검사를 더할 뿐이에요(D-1, D-2).
- Q2: 아웃박스 프로젝트 쪽 갈래(`outbox_dir` 의 기존 링크 거절)는 바꾸나요? → A: 아니에요. 홈 쪽 갈래만 새 도우미를 거쳐요.
- Q3: `claim.rs` 도 새 도우미로 바꾸나요? → A: 아니에요. C-427 이 닫았고 `journal_root_in` 이 `_in` 변형(주어진 HOME)을 써서 도우미의 환경 HOME 읽기와 맞지 않아요. 후속 정리 후보로만 적어요.

## Result

판 1(승인: 리더 결정 노트 8120, 리더 판단이고 사용자 직접 선택은 아니에요). 구현 전이에요.
