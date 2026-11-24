---
id: SPEC-20261124-home-writer-links
title: "홈 상태 아래에 쓰는 나머지 세 곳도 심볼릭 링크 폴더를 거절해요: 아웃박스, 영수증, layout-apply"
status: approved
revision: 3
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-24
  basis: "리더 결정 노트 8214 (2026-11-24T04:53:42Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단이고 사용자 직접 선택은 아님, 노트 8213 에 답함: 개정 3 승인, 노트 8207 선택 1 그대로, write_area 는 바꾸지 않고 호출부 거절을 자식 프로세스 시험으로 못 박아요; 승인 대상 커밋 98464bf278245d793d1c3ae695c54e59a3c51805); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:6dd18987f64d8f4ffae8f64cd0087a6c0a52f3d928c6d86b7daa669091e7c320"
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

개정 2(2026-11-24)는 같은 카드의 기준 3 을 더해요(개정 1 감사 노트 8157 의 N-3 을 리더가 새 기준으로 접었어요, 노트 8185 의 P-1). 홈 폴더 길을 만들어 주는 세 함수가 링크 낀 폴더를 그대로 돌려줘요. 읽은 결과(NeXT `04f3cea651`):

- `ProjectHome::write_area`(`src/project_home.rs`): `<홈 상태 뿌리>/projects/<키>/<영역>` 을 링크 검사 없이 돌려줘요. 부르는 곳은 `hooks::home_area_dir` 하나이고, 쓰는 호출부는 둘이에요: 훅 영수증 쓰기(`hooks::append_receipt_line_in`)와 lane Stop 훅의 상태 저장(`lane_loop::save_state_home`). 둘 다 쓰기 전에 `hooks::make_private_dirs`(`src/hooks.rs`)로 홈 상태 뿌리부터 폴더까지 걸어서 폴더가 아닌 구성요소(링크 포함)에서 `Err` 를 내요: 훅은 `recorded: false` 로 답하고 프로젝트 쪽에도 쓰지 않으며, Stop 훅은 세지 못한 block 을 보내지 않아요. 개정 2 는 이 호출부가 링크를 따라간다고 적었는데 틀렸어요(개정 3 정정, 카드 노트 8206·8207: `write_area` 에 거절을 넣자 기존 테스트 `hooks::receipt_home_tests::a_link_or_a_file_in_the_way_on_the_home_side_is_refused` 가 실패해서 확인했어요). 읽는 호출부(`optimize` 의 훅 로그 읽기, `receipt_lines_in`)가 링크를 따라가는지는 읽지 않았어요(unverified, 후속 후보).
- `ProjectHome::backup_area`: `<dir>/backups/<stamp>/<종류>` 를 검사 없이 돌려줘요. 테스트 밖에서 부르는 곳이 없어요(`grep -rn "backup_area(" src` 가 정의와 `project_home.rs` 의 테스트만 답해요). 백업을 쓰는 Go 상태 코드는 자기 길(`go_state.rs`)을 써요. 부르는 곳이 생기면 이 도우미가 링크 낀 길을 내주지 않아야 해요.
- `audit_clone::evidence_dir`: 감사 증거 폴더 `<홈 상태 뿌리>/projects/<id>/audits/<pr>-<head8>` 의 이름만 만들고 `plan` 이 그 문자열을 감사자에게 알려 줘요(`evidence_dir` 필드). moai 가 쓰지는 않지만 감사자가 로그를 거기에 복사해요. `identity` 가 홈 상태 뿌리 자신이 링크인 것은 이미 거절해요(`home state root is a symlink; refused`). 그 아래 `projects/`, `projects/<id>/`, `audits/` 가 링크인 것은 검사가 없어요.

개정 3(2026-11-24): `write_area` 는 바꾸지 않아요. 호출부가 이미 거절하므로 카드 기준 3 의 `write_area` 몫은 "호출부가 거절한다"로 충족하고 자식 프로세스 테스트로 못 박아요. `backup_area` 와 감사 계획은 같은 `layout::refuse_linked_home_path` 를 같은 문장으로 불러요. 새로 만드는 것은 `ProjectHome` 의 한 걸음 도우미 하나뿐이에요.

## Non-goals

- `src/claim.rs`·`src/worktree_move.rs`(C-427 이 이미 닫았어요)와 `refuse_linked_home_path` 의 동작과 문장은 바꾸지 않아요.
- `copy_verified` 의 원본(프로젝트 쪽) 걷기가 원본 트리 안의 링크를 따라가는 것은 이 카드 밖이에요. 이 카드는 홈 상태 쪽 목적지만 봐요. unverified: 원본 쪽 링크 처리는 이 SPEC 에서 시험하지 않았어요.
- `src/archive.rs` 의 `copy_verified`(이름이 같은 다른 함수)는 홈 상태 뿌리 아래에 쓰는지 읽지 않았어요. 카드 본문이 가리킨 것은 `layout-apply` 의 것이에요(unverified, 후속 후보).
- Windows junction 은 `symlink_metadata().is_symlink()` 가 아니라서 이 검사가 못 잡아요(C-427 과 같은 한계). Windows 결과는 not_run 이에요.
- 검사와 쓰기 사이에 링크를 바꿔 끼우는 경쟁은 막지 못해요(C-427 과 같은 잔여 위험).
- 개정 2·3: `ProjectHome::area`(읽는 길 도우미)와 `write_area` 는 바꾸지 않아요. `area` 는 부르는 곳이 테스트뿐이고 `write_area` 의 쓰는 호출부는 이미 링크를 거절해요(D-7). `hooks.rs`, `optimize.rs`, `lane_loop.rs` 는 읽기만 하고 고치지 않아요. 홈 상태 뿌리 자신이 링크이면 `hooks::home_area_dir` 가 홈 쪽 자리 없음(`None`)으로 보고 프로젝트 쪽 옛 자리를 써요(기존 동작이고 링크 너머로 쓰지 않아요).
- 개정 2: 감사자가 증거 폴더에 로그를 복사하는 쉘 명령은 moai 밖이에요. `plan` 이 링크 낀 폴더를 이름으로 알려 주지 않을 뿐이에요(D-8).

## Requirements

- REQ-1: The `switched_dir_checked` function shall refuse a switched store folder through `refuse_linked_home_path` before it returns the folder.
- REQ-2: IF no layout pointer switches the store, THEN `switched_dir_checked` shall answer no folder and refuse no link of the home state.
- REQ-3: The forge outbox shall refuse a switched outbox folder that is, or sits below, a symbolic link, for the read and for the write.
- REQ-4: The apply receipt store shall refuse a switched receipt folder that is, or sits below, a symbolic link, for the receipt, the before bytes and the recovery read.
- REQ-5: The `layout::apply_at` function shall refuse a home project folder that is, or sits below, a symbolic link before it copies or writes any file.
- REQ-6: IF the home state root of `layout::apply_at` does not exist yet, THEN the function shall copy the stores as before.
- REQ-7: The hook receipt writer shall record nothing, behind the link or on the project side, when a symbolic link lies between the home state root and the receipts folder.
- REQ-8: The lane Stop hook shall send no block when a symbolic link lies between the home state root and the state folder.
- REQ-9: The `ProjectHome::backup_area` function shall refuse a backup folder that is, or sits below, a symbolic link.
- REQ-10: IF a symbolic link lies between the home state root and the evidence folder, THEN the audit clone plan shall answer no evidence folder and give the refusal sentence as the reason.
- REQ-11: The comments and test names that this SPEC adds shall name the SPEC and no audit or card number.

## Acceptance

- AC-1 (REQ-1, REQ-2): `switched_dir_checked` 가 전환 없음이면 `Ok(None)`, 링크 없는 전환이면 `Ok(Some(홈 쪽 폴더))`, `projects/` 가 링크인 전환이면 `home state folder is a symlink; refused` 로 거절해요. 환경 HOME 은 자식 프로세스에만 둬요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(switched_dir_checked)'`
- AC-2 (REQ-3): 환경 HOME 만 바꾼 자식에서 `forge::outbox::save` 와 `load` 가 `projects/` 링크, `projects/<id>/` 링크, `forge/outbox` 링크 셋 모두 `home state folder is a symlink; refused` 로 거절하고 링크 너머에 아무것도 생기지 않아요. 링크 없는 전환 트리는 쓰고 읽어요(대조). · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(outbox_refuses)'`
- AC-3 (REQ-4): 환경 HOME 만 바꾼 자식에서 `apply::Receipt::save`, `apply::save_before_bytes` 가 `projects/` 링크, `projects/<id>/` 링크, `receipts` 링크 셋 모두 같은 문장으로 거절하고 링크 너머에 아무것도 생기지 않아요. 링크 없는 전환 트리는 쓰고(대조) 복구(`apply::recover`)도 같은 거절을 받아요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(receipt_refuses)'`
- AC-4 (REQ-5): `layout::apply_at` 가 `projects/` 링크, `projects/<id>/` 링크일 때 거절하고(홈 상태 뿌리 자신이 링크이면 `home state root is a symlink; refused`), 링크 너머에 폴더도 파일도 `layout.json` 도 생기지 않으며 원본(프로젝트 쪽)이 그대로예요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(layout_apply_refuses)'`
- AC-5 (REQ-5, REQ-6): 링크 없는 트리(홈 뿌리와 `projects/` 가 아직 없는 첫 적용)는 이전처럼 복사하고 표식을 써요. · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(layout_apply_keeps)'`
- AC-6 (REQ-1..REQ-11): 이 SPEC 의 변경 경로가 Scope write 목록 안이에요. · verify: `git diff --stat <base>..<head>`
- AC-7 (REQ-1..REQ-10): 이웃 테스트가 그대로 통과해요: `layout`·`apply`·`forge::outbox`·`project_home`·`audit_clone`·`hooks`·`lane_loop`·`optimize` 모듈의 기존 테스트와 C-427 의 `home_state_link_refusal`. · verify: `cargo nextest run --offline --locked --lib -E 'test(layout::) | test(apply::) | test(forge::outbox) | test(project_home::) | test(audit_clone::) | test(hooks::) | test(lane_loop::) | test(optimize::)'` 와 `cargo nextest run --offline --locked --test home_state_link_refusal`
- AC-8 (REQ-7, REQ-8): 환경 HOME 만 바꾼 자식에서 `hooks::receipt` 가 `projects/` 링크, `projects/<id>/` 링크, `receipts` 링크일 때 `recorded: false` 로 답하고 링크 너머와 프로젝트 쪽 `hooks.jsonl` 에 아무것도 생기지 않아요. `lane_loop::stop_hook`(읽기 `Unknown`)이 `projects/`, `projects/<id>/`, `state` 링크일 때 빈 문자열을 답하고 링크 너머가 그대로예요. 링크 없는 트리는 영수증을 홈 쪽에 쓰고 block 을 보내 상태를 저장해요(대조). `write_area` 의 코드는 바뀌지 않아요(주석만). · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(hook_receipt_) | test(stop_hook_)'`
- AC-9 (REQ-9): `backup_area` 가 `projects/` 링크, `projects/<id>/` 링크, `backups` 링크, 뿌리 링크에서 같은 문장으로 거절하고, 링크 없는 트리는 `<dir>/backups/<stamp>/<종류>` 를 돌려줘요(대조). · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(backup_area_)'`
- AC-10 (REQ-10): 환경 HOME 만 바꾼 자식에서 `audit_clone::plan` 이 `projects/` 링크, `projects/<id>/` 링크, `audits` 링크일 때 `evidence_dir` 를 `null` 로, `evidence_reason` 을 `home state folder is a symlink; refused` 로 답하고(`plan` 자체는 성공해요) 링크 너머가 그대로예요. 링크 없는 트리는 `evidence_dir` 를 이름으로 답해요(대조). · verify: `cargo nextest run --offline --locked --test home_writer_links -E 'test(audit_plan_)'`
- AC-11 (REQ-11): 이 SPEC 이 더한 주석과 테스트 이름에 `AUDIT #`, 카드 번호(`C-` 와 숫자)가 없어요. · verify: `git diff <base>..<head> -- src tests | grep -E '^\+.*(AUDIT #|C-[0-9])'` 가 비어 있어요(이 개정이 지운 줄은 `-` 쪽이라 세지 않아요)

## Scope

- write: `src/layout.rs`, `src/forge/outbox.rs`, `src/apply.rs`, `src/project_home.rs`, `src/audit_clone.rs`, `tests/home_writer_links.rs`, `.moai/specs/SPEC-20261124-home-writer-links/spec.md`
- read-only: `src/claim.rs`, `src/worktree_move.rs`, `src/identity.rs`, `src/home_root.rs`, `src/forge/mod.rs`, `src/hooks.rs`, `src/optimize.rs`, `src/lane_loop.rs`, `src/go_state.rs`, `tests/home_state_link_refusal.rs`, `.moai/specs/SPEC-20261124-home-state-link-refusal/spec.md`
- forbidden: `CLAUDE.md`

`src/layout.rs` 는 여러 카드가 만진 공유 파일이에요. 이 카드는 함수 하나(`switched_dir_checked`, `switched_dir_in` 바로 아래)와 `check_destinations` 의 호출 한 줄만 더하고 기존 줄을 바꾸지 않아요. 이 카드의 예약 범위에 이 파일이 들어 있어서 `moai card line-lease` 는 쓰지 않아요(개정 2 는 `layout.rs` 의 주석 한 줄만 고쳐요). `src/project_home.rs` 와 `src/audit_clone.rs` 는 개정 2 의 예약 범위에 넣어 쓰고, 호출부(`hooks.rs`, `optimize.rs`, `lane_loop.rs`)는 읽기만 해요. `tests/areas.toml` 은 바꾸지 않아요(새 파일의 첫 낱말 `home` 이 이미 `home` 영역에 있어요). 실제 HOME 은 읽기 전후 비교로도 건드리지 않아요: 테스트는 임시 HOME 만 쓰고 환경 HOME 은 자식 프로세스에만 둬요. 네트워크도 쓰지 않아요.

## Plan

1. 승인을 리더 결정 노트로 받아요(승인은 쓰지 않아요).
2. RED: `tests/home_writer_links.rs` 에 AC-1..AC-5 테스트를 쓰고 `switched_dir_checked` 가 없어 컴파일이 실패하는 것을 봐요. 이어서 도우미만 거절하지 않는 뼈대로 더해 컴파일하고 거절 테스트가 실패하는 것(exit 100)을 봐요.
3. GREEN `src/layout.rs`: `pub fn switched_dir_checked(project, subdir) -> Result<Option<PathBuf>>` 를 더해요. HOME 을 한 번(`environment_home`) 읽고 `switched_dir_in` 의 답이 있으면 `default_home_root_in` 의 뿌리로 `refuse_linked_home_path` 를 불러요. `check_destinations` 는 `home_project` 에 같은 검사를 `layout.json` 읽기보다 먼저 불러요.
4. GREEN `src/forge/outbox.rs`: 전환된 갈래가 `switched_dir_checked` 를 불러요; 거절은 `ForgeError::invalid` 에 같은 문장을 그대로 실어요.
5. GREEN `src/apply.rs`: `receipts_dir` 가 `switched_dir_checked` 를 불러요. `receipt_path`·`save_before_bytes`·`before_bytes_path` 가 모두 `receipts_dir` 를 거치므로 한 곳이에요.
6. guard-mutant: 새 테스트마다 거절 호출 하나를 지우고 그 필터가 exit 100 으로 실패하는 것을 봐요(복구는 `git checkout`).
7. 위험 경로: 임시 HOME 아래 링크 폴더(자식 프로세스에만 HOME)로 세 곳을 실제로 돌려 링크 너머가 비어 있음을 READY 에 적어요.
8. `moai check affected <base>..<head>` 를 맨몸으로 한 번 돌려요.
9. 개정 2 승인을 리더 결정 노트로 받아요(승인은 쓰지 않아요).
10. RED(개정 2·3): `tests/home_writer_links.rs` 에 AC-9, AC-10 테스트를 더하고 거절하지 않아서 실패하는 것(exit 100)을 봐요. AC-8 은 호출부가 이미 거절하므로 처음부터 통과하는 특성화 테스트예요; 호출부의 거절 줄을 지운 뮤턴트가 실패시키는지로 가드를 확인해요.
11. GREEN `src/project_home.rs`: `ProjectHome` 에 비공개 도우미 하나를 더하고(`dir` 의 두 단계 위가 홈 상태 뿌리이므로 `refuse_linked_home_path(뿌리, 길)` 를 불러요, D-6) `backup_area` 만 그것을 거쳐요(D-9). `write_area` 는 바꾸지 않아요(D-7).
12. GREEN `src/audit_clone.rs`: `plan` 이 `evidence_dir` 로 이름을 만든 뒤 `refuse_linked_home_path` 를 불러서 거절이면 `evidence_dir` 를 `null`, `evidence_reason` 을 그 문장으로 답해요(D-8).
13. N-7: 이 SPEC 이 더한 주석과 테스트 설명의 `AUDIT #434` 를 SPEC 이름으로 바꿔요(AC-11).
14. guard-mutant 와 위험 경로를 다시 돌려요(`audit-clone-plan` 은 링크 낀 임시 HOME 으로 실제 실행). `moai check affected <base>..<head>` 를 맨몸으로 한 번 돌려요.

## Risks

- 동작이 바뀌는 곳: `layout.json` 이 아웃박스·영수증 폴더를 홈 상태로 돌려 놓았고 `projects/`·`projects/<id>/`·그 폴더 자신이 링크인 사용자는 아웃박스·영수증이 거절돼요(C-427 이 저널에 한 것과 같아요). 홈 상태 뿌리 자신이 링크인 사용자(예 `~/.moai-cli` 를 dotfiles 링크로 둔 경우)의 `layout-apply` 도 거절돼요; 저널은 이미 그랬어요.
- 읽기도 거절해요: `outbox::load` 와 영수증 복구 읽기가 링크 낀 전환 폴더를 읽던 사용자는 거절돼요. 쓰기와 읽기가 같은 폴더를 가리키는 한 갈래라서 한 곳에서 거절하는 편이 단순해요(D-2).
- 한 번 더 부르는 비용: 경로 구성요소마다 `symlink_metadata` 한 번이에요.
- 개정 3, 동작이 바뀌는 곳: `backup_area`(테스트 밖 호출부 없음)와 감사 clone 계획(링크 낀 증거 폴더를 이름으로 알리지 않고 이유를 말해요)뿐이에요. 훅 영수증과 lane Stop 훅은 지금처럼 링크에서 이미 거절하고 이 SPEC 은 바꾸지 않아요.
- 환경 HOME 읽기: `switched_dir_checked` 는 `environment_home` 을 한 번 읽고 같은 값을 `switched_dir_in` 과 뿌리 계산에 써요(HOME 을 두 번 읽다 서로 다른 값을 받는 일이 없어요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 아웃박스와 영수증은 호출부마다 검사를 붙이지 않고 `layout::switched_dir_checked` 하나를 불러요 | 두 호출부가 같은 3줄(홈 뿌리 읽기, 검사, 오류 변환)을 되풀이하면 HOME 을 두 번 읽는 틈이 생겨요. 도우미가 HOME 을 한 번 읽어요. `claim.rs` 의 인라인 형태는 `_in` 변형을 쓰므로 그대로 둬요 | 도우미를 지우고 호출부를 이전 줄로 되돌리기 |
| D-2 | 영수증과 아웃박스는 읽기 길도 거절해요 | 영수증은 `receipts_dir` 한 곳이 저장·백업·복구를 모두 거쳐요. 쓰기만 가르려면 호출부마다 갈라야 해요. 거절 문장은 C-427 과 같아요(`home state folder is a symlink; refused`) | 호출 줄을 지우기 |
| D-3 | `layout-apply` 의 검사는 `check_destinations` 에서 뿌리부터 `home_project` 까지 한 번, 그리고 `KINDS` 의 집 폴더마다 뿌리부터 그 폴더까지 한 번 걸어요(개정 2 가 개정 1 감사 B-1 의 정정을 본문에 옮겼어요) | `check_destinations` 는 `apply_at` 의 첫 검사이고 `apply_locked` 가 잠금 뒤에 다시 불러서 두 번 보아요. 아웃박스의 집 폴더는 두 단계(`forge/outbox`)라서 `home_project` 까지만 걸으면 `forge` 가 링크일 때 마지막 이름 검사가 NotFound 로 통과하고 복사가 링크를 따라가요. 표식(`layout.json`)은 `home_project` 안이에요 | 호출 줄을 지우기 |
| D-4 | `store_dir`(`layout.rs:447`)는 부르는 곳이 없어서 고치지 않아요 | `grep -rn "store_dir(" src` 가 정의 한 줄만 답했어요. 부르는 곳이 생기면 그 카드가 같은 검사를 불러요 | 해당 없음 |
| D-5 | 홈 뿌리가 아직 없는 첫 적용은 거절하지 않아요 | `refuse_linked_home_path` 는 없는 폴더를 링크 없음으로 봐요(C-427 판 1). 첫 적용이 막히면 옵트인이 시작도 못 해요 | 해당 없음 |
| D-6 | `backup_area` 는 `ProjectHome` 의 비공개 도우미 하나를 거쳐요; 홈 상태 뿌리는 새 필드가 아니라 `dir` 의 두 단계 위에서 읽어요 | `dir` 는 `resolve` 만 만들고 늘 `<뿌리>/projects/<키>` 예요(src 안에 직접 만드는 곳 없음; `grep -rn "ProjectHome {" src` 가 정의와 `resolve` 하나). 필드를 더하면 `PartialEq` 와 모든 생성처가 바뀌어요 | 도우미와 호출을 지우기 |
| D-7 | `write_area` 는 바꾸지 않아요(개정 2 의 거절을 내렸어요) | 쓰는 호출부 둘(훅 영수증, lane Stop 훅)이 `hooks::make_private_dirs` 로 이미 링크를 거절하고 기존 테스트가 그것을 못 박아요. `write_area` 가 `Err` 를 내면 `home_area_dir` 의 `.ok().flatten()` 이 `None` 으로 바꿔서 그 거절이 프로젝트 쪽 옛 자리에 조용히 쓰는 것으로 약해져요(`hooks::receipt_home_tests::a_link_or_a_file_in_the_way_on_the_home_side_is_refused` 가 실패해서 드러났어요). 그래서 호출부 거절을 새 자식 프로세스 테스트로 못 박는 것이 카드 기준 3 의 `write_area` 몫이에요 | 해당 없음 |
| D-8 | `audit_clone::plan` 은 링크 낀 증거 폴더를 `Err` 로 막지 않고 `evidence_dir: null` 과 이유로 답해요 | 계획의 나머지(clone 명령)는 링크와 상관없이 쓸 수 있어요. 홈 상태 뿌리 자신이 링크이면 `identity` 가 이미 `plan` 전체를 거절해요(기존 동작). 프로필 id 가 없을 때의 `evidence_reason` 갈래와 같은 모양이에요 | `plan` 의 새 `match` 갈래를 지우기 |
| D-9 | `backup_area` 는 부르는 곳이 없어도 같은 도우미를 거쳐요 | 부르는 곳이 생기면 링크 낀 길을 받으면 안 돼요. 함수를 지우는 것은 이 카드 밖이에요 | 호출을 지우기 |

## Interview

- Q1: 영수증 `receipt_path` 의 기존 마지막 이름 검사는 어떻게 되나요? → A: 그대로 남아요. 이 SPEC 은 그 앞에 뿌리부터 마지막 폴더까지의 검사를 더할 뿐이에요(D-1, D-2).
- Q2: 아웃박스 프로젝트 쪽 갈래(`outbox_dir` 의 기존 링크 거절)는 바꾸나요? → A: 아니에요. 홈 쪽 갈래만 새 도우미를 거쳐요.
- Q3: `claim.rs` 도 새 도우미로 바꾸나요? → A: 아니에요. C-427 이 닫았고 `journal_root_in` 이 `_in` 변형(주어진 HOME)을 써서 도우미의 환경 HOME 읽기와 맞지 않아요. 후속 정리 후보로만 적어요.
- Q4: 감사 증거 폴더(`evidence_dir`)의 이름 만드는 함수 자체를 `Result` 로 바꾸나요? → A: 아니에요. 이름 함수는 그대로(시그니처와 테스트 유지)이고 `plan` 이 그 이름에 검사를 붙여요(D-8).
- Q5: `write_area` 가 링크를 거절하게 하지 않나요? → A: 아니에요(개정 3). 쓰는 호출부가 이미 거절하고, `write_area` 가 `Err` 를 내면 그 거절이 조용한 프로젝트 쪽 대체로 약해져요(D-7).

## Result

판 1(승인: 리더 결정 노트 8120, 리더 판단이고 사용자 직접 선택은 아니에요). 브랜치 `spec/home-writer-links-20261124`, 기준 NeXT `e15220f90f` 를 쫓지 않았어요. 구현 커밋은 `3a5b85a1f6`(RED 표와 뼈대), `cbb477badf`(GREEN)이에요. 로그는 `target/c434/` 에 있고 카드가 열려 있는 동안만 둬요.

- RED: 테스트 파일만 있고 도우미가 없어서 컴파일이 exit 101 로 실패했어요(`E0425 cannot find function switched_dir_checked`, 네 곳). 거절하지 않는 뼈대를 더하고 `cargo nextest run --offline --locked --test home_writer_links --no-fail-fast` → exit 100, 16개 중 5 통과(대조) · 11 실패(거절 테스트 전부). 실패 이유가 모두 "거절하지 않아서"예요: 아웃박스는 `unwrap_err()` 가 `Ok(())` 를 만났고, 영수증은 `Ok(())`(링크 너머에 씀) 또는 옛 문장 `receipt directory is a symlink; refused`, `layout-apply` 는 링크 너머에 복사하고 표식을 쓰고 `switched: true` 로 성공했어요(`marker` 가 링크 너머 경로). 첫 RED 는 내 픽스처 탓에 `stale or altered plan` 으로 실패해서(계획 뒤에 클레임 저널 폴더가 새로 생김) 계획 전에 저널을 열어 두도록 고친 뒤 위 결과를 얻었어요.
- GREEN: `--test home_writer_links --test home_state_link_refusal` → exit 0, 30/30(새 16 + 이웃 저널 14).
- AC-1: 충족. `switched_dir_checked_answers_none_without_a_pointer`, `..._answers_the_home_folder_on_a_plain_tree`, `..._refuses_a_linked_projects_folder`, `..._refuses_a_linked_project_folder`(환경 HOME 은 자식 프로세스에만).
- AC-2: 충족. `outbox_refuses_a_linked_projects_folder`, `..._project_folder`, `..._outbox_folder`(`save` 와 `load` 모두 `home state folder is a symlink; refused`, 링크 너머 파일 바이트까지 그대로), `outbox_refuses_no_plain_tree`(대조).
- AC-3: 충족. `receipt_refuses_a_linked_projects_folder`, `..._project_folder`, `..._receipts_folder`(`Receipt::save`, `save_before_bytes`, `recover` 모두 같은 문장, 링크 너머가 그대로), `receipt_refuses_no_plain_tree`(대조).
- AC-4: 충족. `layout_apply_refuses_a_linked_projects_folder`, `..._project_folder`(폴더 문장), `..._home_root`(뿌리 문장); 링크 너머와 프로젝트 쪽 원본이 그대로예요.
- AC-5: 충족. `layout_apply_keeps_a_first_apply_on_a_plain_tree`(홈 뿌리와 `projects/` 가 없는 첫 적용이 복사하고 표식을 써요).
- AC-6: 충족. `git diff --stat e15220f90f..HEAD` 가 `src/apply.rs`, `src/forge/outbox.rs`, `src/layout.rs`, `tests/home_writer_links.rs` 와 이 SPEC 다섯 개만 보여요.
- AC-7: 충족. `cargo nextest run --offline --locked --lib -E 'test(layout::) | test(apply::) | test(forge::outbox)'` → exit 0, 38 통과. C-427 의 `home_state_link_refusal` 14개도 위 30/30 안에서 통과해요.
- 뮤턴트(적용을 `git diff --stat` 으로 확인하고 `git checkout -- <파일>` 로 되돌린 뒤 `git status --short` 가 비었음을 확인; 모두 `--test home_writer_links`, exit 100): ① `switched_dir_checked` 안의 `refuse_linked_home_path` 호출 삭제 → 8개 실패(`switched_dir_checked_refuses` 2, `outbox_refuses` 3, `receipt_refuses` 3) ② `check_destinations` 의 호출 삭제 → 3개 실패(`layout_apply_refuses` 전부) ③ 아웃박스 호출부를 검사 없는 `switched_dir` 로 되돌림 → 3개 실패(`outbox_refuses`) ④ 영수증 호출부를 `switched_dir` 로 되돌림 → 3개 실패(`receipt_refuses`).
- 위험 경로: 깨끗한 head `cbb477badf` 에서 만든 `target/debug/moai` 로 임시 폴더(HOME 은 자식 명령에만 줘요)에서 `layout-plan` 과 `layout-apply` 를 돌렸어요(`target/c434/danger.log`, `danger-root.log`). `projects/` 링크와 `projects/<id>/` 링크는 exit 1, `home state folder is a symlink; refused`, 링크 너머가 `keep.txt` 하나 그대로(바이트 `bytes behind the link`), 원본 `old.json` 그대로. 상태 루트 링크(`~/.moai-cli` 가 빈 폴더로의 링크)는 exit 1, `home state root is a symlink; refused`, 링크 너머가 비어 있어요. 링크 없는 대조는 exit 0 으로 복사하고 표식을 써요. 첫 프로브는 내 픽스처 탓에 `stale` 이 나와서(링크 대상에 파일이 있어 상태 루트 이름이 계획과 적용 사이에 바뀜) 링크를 계획 전에 만들도록 고쳐 다시 돌렸어요. `~/.moai` 를 링크로 두면 이름 해석이 `.moai-cli` 로 넘어가는 것은 이 카드 밖의 기존 동작이에요.
- 동작이 바뀌는 곳: 위 Risks 와 같아요. 영수증과 아웃박스는 읽기 길도 거절하고, 상태 루트가 링크인 사용자의 `layout-apply` 도 거절돼요.
- 하지 않은 것: `claim.rs`·`worktree_move.rs`(D-1 대로 그대로), 원본(프로젝트 쪽) 트리 안의 링크 처리, `archive.rs` 의 `copy_verified`(unverified, 후속 후보).
- not_run: Windows·Linux(junction 은 이 검사가 못 잡아요), 실제 HOME 읽기 전후 비교(테스트와 위험 경로는 임시 HOME 만 써요), 전체 `cargo test --locked`(리더 몫).

### 수리: AUDIT #434 B-1 (head `4f2d949919` 의 FAIL 뒤)

수리 커밋은 `50b6c258bc`(코드와 테스트)와 `33bec29c34`(테스트 주석)예요. 첫 `check affected`(head `d7930da508`)는 `state_folder_literal_scan` 한 개가 FAIL 이었어요(내 주석이 상태 폴더 이름을 적어서, 나머지 2923 통과): 주석을 이름 없이 고쳐 새 head 로 다시 돌렸어요. 판은 올리지 않았어요(승인 digest 그대로): 본문의 D-3 은 고치지 않고 아래에 정정으로 적어요(리더 결정, 노트 8157).

- D-3 정정: "`home_project` 까지의 구성요소만 보면 충분해요" 는 틀렸어요. 아웃박스의 집 폴더는 두 단계(`forge/outbox`)라서 `forge` 가 링크이고 그 뒤에 `outbox` 가 없으면 마지막 이름 검사(`symlink_metadata`)가 NotFound 로 통과하고, 복사(`create_dir_all`)가 링크를 따라가요. 맞는 문장은 이거예요: `check_destinations` 는 `KINDS` 의 집 폴더마다 `refuse_linked_home_path(home_root, home_project/<subdir>)` 로 뿌리부터 그 폴더까지 모든 구성요소를 걸러요. 마지막 이름이 이미 링크이면 충돌 문장 대신 같은 링크 거절 문장이 먼저 나와요(둘 다 거절이고 아무것도 쓰지 않아요).
- 수정: `src/layout.rs` `check_destinations` 의 `KINDS` 루프에 `refuse_linked_home_path` 한 줄.
- RED(수리 전 코드): `layout_apply_refuses_a_linked_forge_folder` 가 `unwrap_err()` 가 `Ok` 를 만나 실패했어요. 값이 `switched: true`, `marker` 가 링크 너머 경로예요(B-1 재현). 홈 상태 뿌리가 링크인 아웃박스·영수증 테스트(N-2)는 수리 전에도 통과했어요(`switched_dir_checked` 가 이미 뿌리를 걸러요): 테스트가 모자랐을 뿐 코드 결함이 아니었어요.
- 새 테스트 3개: `layout_apply_refuses_a_linked_forge_folder`(링크 너머 바이트와 목록 그대로, 프로젝트 쪽 원본 그대로, `layout.json` 안 써요), `outbox_refuses_a_linked_home_root`, `receipt_refuses_a_linked_home_root`(둘 다 `home state root is a symlink; refused`, `save`·`load`·`before bytes`·`recover` 전부, 링크 너머 그대로). 뿌리 링크는 `.moai` 이름으로 두면 표식을 잃어 `.moai-cli` 로 판정이 바뀌므로 `.moai-cli` 이름의 링크로 두고 `home_root::path` 가 그 링크로 풀리는지 픽스처가 확인해요.
- GREEN: `cargo nextest run --offline --locked --test home_writer_links --test home_state_link_refusal --no-fail-fast` → 33/33. `--lib -E 'test(layout::) | test(apply::) | test(forge::outbox)'` → 38/38.
- 뮤턴트(적용을 `git diff --stat` 으로 확인, `git checkout -- src/layout.rs` 로 되돌린 뒤 `git status --short` 비었음; 모두 `--test home_writer_links`, exit 100): ⑤ `check_destinations` 루프의 새 호출 삭제 → `layout_apply_refuses_a_linked_forge_folder` 한 개만 실패 ⑥ `refuse_linked_home_path` 안의 뿌리 검사(`home state root`) 삭제 → `layout_apply_refuses_a_linked_home_root`, `outbox_refuses_a_linked_home_root`, `receipt_refuses_a_linked_home_root` 셋 실패.
- 위험 경로(수리 head `50b6c258bc` 의 `target/debug/moai`, 임시 폴더, HOME 은 자식 명령에만; `target/c434/danger-repair.log`): `projects/<id>/forge` 가 `outbox` 없는 폴더로의 링크 → exit 1, `home state folder is a symlink; refused`, 링크 너머가 `keep.txt` 하나 그대로, `layout.json` 안 써요, 원본 그대로. `~/.moai-cli` 가 링크 → exit 1, `home state root is a symlink; refused`, 링크 너머 그대로. 링크 없는 대조 → exit 0, 복사하고 표식을 써요. 첫 계획은 클레임 저널 폴더가 아직 없어 stale 일 수 있어서 한 번 다시 계획한 뒤 적용했어요(대조만 두 번째 시도에서 성공, 나머지 둘은 첫 시도에서 거절).
- 카드 기준 3(N-3: `project_home::write_area`·`backup_area`, `audit_clone::evidence_dir`)은 이 수리에 넣지 않았어요(리더 결정: 새 기준으로 접음, 열린 채).

### 개정 2·3: 카드 기준 3(N-3)과 N-7

브랜치 `spec/home-writer-links-20261124-r2`, 기준 NeXT `04f3cea651` 를 쫓지 않았어요. 승인: 개정 2 는 리더 결정 노트 8199, 개정 3 은 노트 8214(둘 다 리더 판단이고 사용자 직접 선택은 아니에요). 커밋은 `e5e1067982`(개정 2 초안), `b36cb87152`(개정 2 승인), `14a182c8f8`(N-7 주석), `f54469a4d0`(RED), `f6a6b4be75`(GREEN), `c56fa85595`(`write_area` 거절을 내리고 호출부 시험), `98464bf278`(개정 3 초안), `163c028259`(개정 3 승인)이에요. 로그는 `target/c434/` 에 카드가 열려 있는 동안만 둬요.

- 정정(개정 3, 노트 8206·8207): 개정 2 의 D-7 전제는 틀렸어요. 개정 2 대로 `write_area` 에 거절을 넣자 기존 테스트 `hooks::receipt_home_tests::a_link_or_a_file_in_the_way_on_the_home_side_is_refused` 한 개가 실패했어요(lib 100개 중 99 통과, `target/c434/lib.log`). 원인: `write_area` 의 쓰는 호출부 둘(훅 영수증, lane Stop 훅)이 `hooks::make_private_dirs` 로 이미 링크를 거절하는데, `write_area` 가 `Err` 를 내면 `home_area_dir` 의 `.ok().flatten()` 이 `None` 으로 바꿔 그 거절이 프로젝트 쪽에 조용히 쓰는 것으로 약해졌어요. 그래서 `write_area` 변경은 내렸고(코드는 원래대로, 주석만 남아요) 호출부 거절을 새 시험으로 못 박았어요.
- RED(개정 2): 새 시험 35개 중 거절 시험 11개가 거절하지 않아서 실패하고(`unwrap_err` 가 `Ok(Some(경로))` 를 만나요; 감사 계획은 `evidence_dir` 에 경로가 있어요) 24개가 통과했어요 → exit 100(`target/c434/red.log`).
- GREEN: `cargo nextest run --offline --locked --test home_writer_links --test home_state_link_refusal --no-fail-fast` → exit 0, 51/51(새 시험 포함 37 + 이웃 저널 14, `target/c434/green2.log`).
- AC-8: 충족. `hook_receipt_refuses_a_linked_projects_folder`, `..._project_folder`, `..._receipts_folder`(`hooks::receipt` 가 `recorded: false`, 링크 너머와 프로젝트 쪽 `hooks.jsonl` 에 아무것도 없어요), `stop_hook_refuses_a_linked_projects_folder`, `..._project_folder`, `..._state_folder`(`lane_loop::stop_hook` 이 빈 문자열, 링크 너머 그대로), 대조 `hook_receipt_keeps_a_plain_tree`, `stop_hook_keeps_a_plain_tree`. 이 시험은 호출부가 이미 거절하므로 처음부터 통과하는 특성화 시험이고, 거절 줄을 지운 뮤턴트 ③이 6개를 실패시켜요. `write_area` 코드는 바뀌지 않았어요(`git diff 04f3cea651 HEAD -- src/project_home.rs` 에서 `write_area` 본문 줄 없음, 주석만).
- AC-9: 충족. `backup_area_refuses_a_linked_projects_folder`, `..._project_folder`, `..._backups_folder`, `..._home_root`, 대조 `backup_area_keeps_a_plain_tree`. `backup_area` 는 테스트 밖 호출부가 없어요(`grep -rn "backup_area(" src` 가 정의와 `project_home.rs` 테스트만 답해요).
- AC-10: 충족. `audit_plan_refuses_a_linked_projects_folder`, `..._project_folder`, `..._audits_folder`(`plan` 은 성공하고 `evidence_dir: null`, `evidence_reason` 이 `home state folder is a symlink; refused`), 대조 `audit_plan_keeps_a_plain_tree`, 그리고 기존 동작 `audit_plan_still_refuses_a_linked_home_root`(`identity` 가 뿌리 링크로 계획 전체를 거절).
- AC-11: 충족. 이 개정이 더한 줄에서 `AUDIT #` 와 카드 번호가 0개예요(`target/c434/added.diff` 를 단어 경계를 둔 식 `AUDIT #|(?<![A-Za-z])C-\d` 로 세었어요). 주의: AC-11 의 verify 문장에 적은 `grep -E '…C-[0-9]'` 는 `AC-8`, `SPEC-…` 같은 이름까지 잡아서 25줄이 나와요(거짓 양성; 판을 올리지 않으려고 문장은 그대로 두고 여기에 정정해요).
- AC-6: 충족. `git diff --stat 04f3cea651..HEAD` 가 이 SPEC, `src/audit_clone.rs`, `src/layout.rs`(주석 한 줄), `src/project_home.rs`, `tests/home_writer_links.rs` 다섯 개만 보여요.
- AC-7: 충족. `cargo nextest run --offline --locked --lib -E 'test(layout::) | test(apply::) | test(forge::outbox) | test(project_home::) | test(audit_clone::) | test(hooks::) | test(lane_loop::) | test(optimize::)'` → exit 0, 100/100(`target/c434/lib2.log`). C-427 의 `home_state_link_refusal` 14개는 위 51/51 안에서 통과해요.
- 뮤턴트(적용을 `git diff --stat` 으로 확인하고 `git checkout -- <파일>` 로 되돌린 뒤 `git status --short` 가 비었음을 확인; ③은 작업 트리 밖 임시 복사본에서 했어요): ① `backup_area` 안의 `self.refuse_linked(&path)?` 삭제 → `backup_area_refuses_*` 4개 실패(대조 1개 통과) ② `audit_clone::plan` 의 `refuse_linked_home_path` 호출을 `Ok(())` 로 바꿈 → `audit_plan_refuses_*` 3개 실패(대조·뿌리 2개 통과) ③ `hooks::make_private_dirs` 의 `Ok(_) => return Err(…)` 를 `Ok(_) => continue` 로 바꿈(`src/hooks.rs` 는 이 카드의 쓰기 범위 밖이라 `git archive` 복사본에서만) → `hook_receipt_refuses_*` 3개와 `stop_hook_refuses_*` 3개 실패(대조 2개 통과). 모두 exit 100.
- 위험 경로(`target/debug/moai-cli` 는 head `163c028259` 의 깨끗한 트리에서 빌드, 임시 폴더, HOME 은 자식 명령에만, `target/c434/danger-r2.log`): `audit-clone-plan` 은 `projects/`, `projects/widgets/`, `audits` 가 `keep.txt` 한 개 든 폴더로의 링크일 때 모두 `ok: true`, `evidence_dir: null`, 이유 `home state folder is a symlink; refused` 이고 링크 너머가 바이트 그대로예요. `hook-receipt` 는 `projects/`, `projects/widgets/`, `receipts` 링크에서 링크 너머와 프로젝트 쪽 `hooks.jsonl` 에 아무것도 쓰지 않아요. 링크 없는 대조는 `evidence_dir` 가 이름으로 나오고 영수증이 홈 쪽에 써져요. `backup_area` 는 CLI 에서 닿는 길이 없어서 시험으로만 확인했어요.
- 동작이 바뀌는 곳: 감사 clone 계획이 링크 낀 증거 폴더를 이름으로 알리지 않고 이유를 말해요. `backup_area` 는 링크 낀 길을 돌려주지 않아요(부르는 곳 없음). 훅 영수증과 lane Stop 훅은 그대로예요.
- 하지 않은 것: `write_area`·`hooks.rs`·`optimize.rs`·`lane_loop.rs` 변경, 읽는 호출부(`receipt_lines_in`, `optimize` 훅 로그 읽기)가 링크를 따라가는지 읽기(unverified, 후속 후보), 홈 상태 뿌리 자신이 링크인 경우의 훅 영수증(`home_area_dir` 가 `None` 으로 보고 프로젝트 쪽에 써요; 기존 동작).
- not_run: Windows·Linux(junction 은 이 검사가 못 잡아요), 실제 HOME 읽기 전후 비교(시험과 위험 경로는 임시 HOME 만 써요), 전체 `cargo test --locked`(리더 몫).
