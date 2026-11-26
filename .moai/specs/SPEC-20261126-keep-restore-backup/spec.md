---
id: SPEC-20261126-keep-restore-backup
title: "update: 같은 설치의 보통 적용이 백업 복원 제안을 잃지 않게"
status: draft
revision: 1
created: 2026-11-26
---

# update: 같은 설치의 보통 적용이 백업 복원 제안을 잃지 않게

## Intent

backup-then-replace(`update-apply --confirm-backup-replace`)는 원래 폴더를 백업으로 통째로 옮기고, 영수증의 `backup` 으로만 복원을 제안해요. 그 뒤 같은 설치가 파일을 쓰거나 소유 기록을 옮기는 보통 적용을 하면, 쓰기 전에 저장하는 pending 영수증이 `backup` 없이 자기 영수증을 덮어써요. 원본 폴더는 백업에 남지만 `recovery-plan` 은 더 이상 복원을 제안하지 않아요. 이 SPEC 은 그 보통 적용이 성공했을 때 복원 제안이 남고, 되돌리기가 원본을 바이트와 모드까지 그대로 돌려놓게 해요.

**이 수정이 닿는 범위(승인 전에 확인해 주세요).** 아래는 코드를 읽은 결과이고 실행해 보지는 않았어요.

- 같은 설치 id(같은 실행 파일 경로)의 적용에만 효과가 있어요. 배포 레이아웃(`versions/<버전>/bin`)에서는 새 버전이 새 id 이고 `recovery-plan` 은 지금 id 의 영수증만 읽어서, 업그레이드하면 보통 적용을 하기 전부터 제안이 보이지 않아요. 이 카드 전과 같아요.
- 교체한 뒤 같은 단계 패키지로 다시 적용하면 파일을 쓰는 길이 없어요(지운 파일도 `preserved`). 그래서 이 수정이 지키는 길은 같은 설치가 내용이 다른 단계 패키지를 적용하는 경우예요.
- 실제 설치에서 그 길에 닿는 모양은 두 가지로 읽혀요. 배포 레이아웃에서는 교체한 버전 A 에서 B 로 올렸다가 A 로 되돌아가 다시 적용하는 경우(A→B→A)예요. 그 밖에는 같은 실행 파일 경로의 패키지가 바뀐 경우(같은 버전 폴더 다시 설치, 레이아웃 밖 설치, `update-apply` 에 다른 단계 폴더)예요. 새 버전으로 올리는 보통 업그레이드에는 닿지 않아요.

## Interview

- Q1: 어떤 적용에서 `backup` 을 지키나요? → A: 같은 설치의 보통 적용이 `committed` 로 끝날 때예요(수락 기준 1과 2, D-1).
- Q2: 중간에 실패한 적용은요? → A: 지금과 같게 둬요. 파일별 복구는 그대로이고 복원 제안은 사라져요(D-3, 승인 노트의 선택).
- Q3: 버전을 올린 새 설치에도 보이게 하나요? → A: 이 SPEC 에서는 하지 않아요. 리더가 후속 작업을 따로 열었어요(D-4, 승인 노트의 선택).
- Q4: 복원 제안이 계속 남아도 되나요? → A: 남겨요. 카드 기준 1이 요구하는 동작이에요(D-5, 승인 노트의 선택).
- Q5: 동작을 적은 문서는요? → A: `docs/rust-product-path.md` 는 카드 범위에 넣지 않고 `card line-lease` 로 잠깐 빌려 고친 뒤 커밋하고 바로 돌려줘요(D-6, 승인 노트의 선택).
- Q6: 승인 노트의 선택은 누가 정했나요? → A: 원장 노트 10154(2026-11-26 07:53Z)에서 리더가 사용자 목표에 따라 판단했어요. 사용자의 직접 선택은 아니에요.

## Non-goals

- 다른 설치 id 의 영수증에서 `backup` 을 가져오는 것. 옮겨진 상태 폴더의 옛 id 도 같아요.
- `src/recovery.rs`, 영수증의 필드와 스키마를 바꾸는 것.
- 실패하거나 중단된 적용 뒤에도 복원 제안을 지키는 것(D-3).
- 이미 사라진 제안을 되살리는 것. 백업 폴더는 남아 있어 문서의 손 복원 절차를 그대로 써요.
- 백업 보존과 정리, `extra_skill` 영수증.

## Requirements

- REQ-1: When an ordinary update apply of a host commits and the receipt of this installation held a `committed` status and a `backup` before that apply, the update shall save the new receipt with that same `backup`.
- REQ-2 (REQ-1): While the new receipt carries a `backup`, the update shall list in its entries the writes of this apply first and in write order, followed by one entry for every other owned path with action `unchanged`, the owned digest as both `before_sha256` and `after_sha256`, `published` false and no error.
- REQ-3 (REQ-1): The recovery of the offer of such a receipt shall restore the original folder byte for byte and mode for mode, keeping the installed package in `<backup>/replaced` undeleted.
- REQ-4 (REQ-2): IF a file under the projection root differs from its entry or has no entry, THEN the recovery shall offer no restore (`none`, `user_changed`) and write nothing under the root and the backup.
- REQ-5 (REQ-1): IF the receipt of this installation before the apply is absent or is not `committed` or has no `backup`, THEN the update shall save the receipt as it does today, with no `backup` and entries for the writes only.
- REQ-6 (REQ-1): The update shall take `backup` only from the receipt of this installation, read before the pending receipt is saved, and never from the baseline receipt of another installation.
- REQ-7 (REQ-1): IF the apply fails before the commit save, including a verification error, THEN the saved receipt shall be the pending receipt of today, with no `backup` and entries for the writes only.
- REQ-8: The update shall add no receipt field, keep the operation `update.apply` and change no recovery code.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 교체한 뒤 같은 설치가 내용을 바꾼 단계 패키지로 보통 적용을 두 번 해요. 두 번 모두 `recovery-plan` 의 `restore_backup.action` 이 `restore_backup`, `original_at` 이 `backup` 이에요. `backup` 은 교체 영수증과 같고, entries 의 경로와 digest 는 설치 폴더의 파일과 같아요. `recovery-apply` 뒤 설치 폴더는 원본과 바이트·모드가 같고, 적용했던 패키지는 `<backup>/replaced` 에 있으며 `RESTORED.json` 이 생기고 영수증은 `recovered` 예요. 지금 코드에서는 첫 적용 뒤 단언에서 실패해야 해요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_an_ordinary_apply_keeps_the_backup`
- AC-2 (REQ-5): 이전 백업이 없는 설치의 보통 적용 두 번은 `backup` 이 없고, entries 는 쓴 파일뿐이며, `restore_backup` 행이 없어요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_adds_nothing_without_an_earlier_backup`
- AC-3 (REQ-1, REQ-2): 파일을 쓰지 않고 소유 기록만 옮기는 적용 뒤에도 제안이 있고 복원이 성공해요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_covers_a_baseline_move`
- AC-4 (REQ-4): 남긴 사용자 편집이나 충돌이 있으면 `none`/`user_changed` 이고, `recovery-apply` 는 `partial` 이며 설치 폴더와 `<backup>/moai` 는 그대로이고 `RESTORED.json` 이 없어요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_is_blocked_by_a_kept_user_edit`
- AC-5 (REQ-7): 중간에 실패한 적용은 pending 영수증에 `backup` 이 없고, 파일별 복구가 전처럼 동작하며, 복원 행이 없어요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_leaves_a_failed_apply_to_per_file_recovery`
- AC-6 (REQ-6): 다른 설치가 적용하면 그 설치의 새 영수증에는 `backup` 이 없고, 교체한 설치의 영수증 바이트는 그대로예요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_is_not_taken_from_another_installation`
- AC-7 (REQ-5): 복원에 이미 쓴 백업(`recovered` 영수증)은 다음 적용에 옮기지 않아요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_never_carries_a_used_backup`
- AC-8 (REQ-8): 이웃 테스트와 HOME 이름 표 검사가 통과하고, `src/recovery.rs` 와 영수증 구조체의 필드는 바뀌지 않아요 · verify: `cargo test --locked --offline --test update_backup_replace --test update_conflict_baseline_cli --test update_adopt_cli --test update_cli --test recovery_cli`; `cargo test --offline --locked --lib every_rust_written_home_name_is_in_the_table`
- AC-9 (REQ-1, REQ-7): 문서 범위가 승인되면 `docs/rust-product-path.md` 의 두 문단이 새 동작(성공한 적용은 제안을 지키고, 실패한 적용과 다른 버전 폴더는 지키지 않음)을 적어요 · verify: `grep -n 'keeps the restore offer' docs/rust-product-path.md`
- AC-10 (REQ-1, REQ-2, REQ-3): 버전 폴더 배치에서 A 가 교체하고 B 가 파일을 바꾸고 더한 뒤, A 가 다시 보통 적용을 해요. A 의 `recovery-plan` 이 `restore_backup` 을 내놓고 `recovery-apply` 가 원본 바이트와 모드를 되돌려요. 지금 코드에서는 실패해야 해요 · verify: `cargo test --locked --offline --test update_backup_replace keep_restore_covers_what_another_build_wrote`

## Scope

- write: `src/update.rs`, `tests/update_backup_replace.rs`, `.moai/specs/SPEC-20261126-keep-restore-backup/**`
- write (카드 범위 밖, `card line-lease` 로 잠깐 빌려 씀): `docs/rust-product-path.md`
- read-only: `src/recovery.rs`, `src/apply.rs`, `src/home_layout.rs`
- forbidden: `src/recovery.rs`, `src/home_layout.rs`, `tools/repo_scans.json`; 실제 사용자 HOME 의 설치 파일은 건드리지 않아요

## Plan

1. 선행: `SPEC-20261125-update-installed-from-link` 의 수리본이 감사를 통과해 착지할 때까지 코드를 쓰지 않아요. 착지한 tip 을 fast-forward 로 받고(이 가지에는 아직 커밋이 없어요), 테스트 도우미, `record_version_only`, 줄 번호를 다시 읽어요.
2. SPEC 커밋.
3. 빨강: `tests/update_backup_replace.rs` 에 AC-1~7 과 AC-10 의 테스트 8개를 더해요. AC-1, AC-3, AC-10 은 지금 코드에서 실패해야 하고, AC-2, AC-4~7 은 지켜야 할 쪽을 고정하는 테스트라 실제 결과를 그대로 적어요. 재현은 임시 HOME 에서 실제 바이너리를 돌리는 Rust 통합 테스트로만 해요. `src/` 에 새 단위 테스트나 `home_root::path` 호출을 더하지 않아요.
4. 수정(`src/update.rs`): pending 영수증을 저장하기 전에 이 설치의 영수증을 읽어 `committed` 일 때의 `backup` 을 잡아 두고, 세 재검증이 끝난 뒤 commit 저장 직전에만 entries 를 채우고 `backup` 을 넣어요. `HostReceipt.entries` 와 `backup` 의 문서 주석을 영어로 고쳐요.
5. 초록: AC-8 의 타깃을 heavy slot 에서 돌려요.
6. 가드 뮤턴트(하나씩, 되돌린 뒤 트리 확인):

| 테스트 | 빼거나 바꾸는 코드 |
| --- | --- |
| AC-1 | `backup` 대입 |
| AC-2 | 백업이 있을 때만 entries 를 채우는 조건 |
| AC-3 | entries 를 `owned_now` 대신 `pre_owned` 로 채움 |
| AC-4 | entries digest 를 `owned_now` 대신 디스크에서 읽음 |
| AC-5 | `backup` 옮기기를 pending 저장 앞으로 옮김 |
| AC-6 | 원본을 자기 영수증 대신 `previous` 로 |
| AC-7 | `committed` 조건 |
| AC-10 | entries 를 `owned_now` 대신 자기 영수증의 이전 `owned` 로 채움 |

   AC-6 과 AC-10 의 두 빌드는 모두 버전 폴더 배치에 두어요. 배치 밖 빌드는 다른 영수증을 빌려 읽지 않아서 AC-6 의 뮤턴트가 죽지 않아요.

7. `cargo fmt --all`, `git diff --check`, 커밋, bare `moai check affected --card <base>..<head>`, READY. `git diff --stat <base>..<head> -- src/recovery.rs src/home_layout.rs` 가 비어 있는지와 `HostReceipt`, `HostReceiptEntry` 의 필드가 그대로인지 diff 로 읽어요. AC-1 의 테스트를 위험 경로 실행 근거로 적어요.

## Risks

- 복원 제안이 남아 있는 동안 `recovery-apply` 를 실행하면, 업데이트를 여러 번 한 뒤에도 지금 skill 폴더가 `<backup>/replaced` 로 가고 교체 전의 옛 폴더가 돌아와요. 지우는 것은 없어요. 지금은 교체 직후(파일을 쓰는 적용 전)에만 가능한 일이에요. `moai update` 결과의 `revert.skills` 안내 문구는 pending 호스트만 말해요.
- 남긴 편집이 있거나 백업을 손으로 지운 `committed` 호스트에서 `recovery-plan` 이 `none` 행을 보이고 `recovery-apply` 가 `partial` 이 돼요. 지금은 교체 직후에만 그래요.
- entries 의 digest 는 `owned` 에서 오므로, 다른 빌드의 기준 영수증에서 이어받은 파일도 들어가요(`owned` 가 지금도 그래요). 되돌리기는 그 파일을 지우지 않고 `<backup>/replaced` 로 옮겨요.
- 이 수정이 없는 옛 바이너리가 같은 영수증 자리에서 적용하면 `backup` 이 다시 사라져요.
- `src/recovery.rs` 의 "this receipt created" 주석은 조금 낡아져요(범위 밖).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `backup` 은 이 설치의 영수증이 `committed` 일 때만 옮기고, operation 은 보지 않아요. 그래서 다음 적용에도 이어져요. | `recovered` 영수증은 `backup` 을 그대로 두지만 그 백업은 이미 쓰였어요. | 옮기는 코드를 지워요. |
| D-2 | 덮는 entries 는 commit 직전의 `owned_now` 로 만들어요. 이전 entries 를 합치거나 디스크를 훑지 않아요. | 복원 판정은 설치 폴더의 모든 파일이 entry 와 같아야 해요. 디스크를 훑으면 사용자 편집이 통과해 옮겨져요. | 이전 영수증의 entries 를 합치는 방식으로 바꿔요. |
| D-3 (노트 10154) | `backup` 은 commit 저장에만 넣어요. 실패한 적용은 지금처럼 파일별 복구를 하고 제안은 잃어요. | pending 영수증에 넣으면 `src/recovery.rs` 가 파일별 복구를 건너뛰어 옛 폴더 전체 복원만 남고, 경로마다 entry 가 하나면 반만 적용된 폴더가 어느 쪽에도 맞지 않아 호스트가 pending 에 갇혀요. 둘 다 지키려면 `src/recovery.rs` 를 바꿔야 해요(후속 카드). | 후속 카드에서 복구 코드를 바꿔요. |
| D-4 (노트 10154) | 버전 업그레이드 뒤의 새 설치 id 에서는 제안을 이어 주지 않아요. 후속 p1 작업은 리더가 따로 열었어요. | `recovery-plan` 은 지금 id 만 읽어요. 이어 주려면 복구 설계가 따로 필요해요. | 이 카드의 범위를 넓혀요. |
| D-5 (노트 10154) | 같은 설치의 성공한 적용이 이어지는 동안 복원 제안을 계속 남겨요. | 카드 기준 1. 원본을 되돌릴 길이 조용히 사라지지 않아요. | 제안 기간을 따로 정해요. |
| D-6 (노트 10154) | `docs/rust-product-path.md` 를 고치고 AC-9 를 유지해요. 카드 범위에는 넣지 않고 `card line-lease` 로 잠깐 빌려 쓰고, 커밋 뒤 바로 돌려줘요. 지금 문서는 제안이 사라지는 것을 동작으로 적어요. 그 파일은 진행 중인 다른 작업 둘도 예약하고 있어서(2026-11-26 읽음), 착지한 tip 을 받은 뒤의 내용에 맞춰 고쳐요. | 결정 기록 22번 행은 "막고, 따로 확인한 뒤 백업하고 교체"만 정해요. 제안이 사라지는 것은 #2505 가 한계로 적었어요. | 문서 변경을 되돌려요. |

## Result

2026-11-26, 기준 `75b69200669812ef3ff83aa976b4b649fc6d0180`. 실행 환경은 Mac 한 대이고 Windows 와 Linux 에서는 돌리지 않았어요.

- 커밋: SPEC `265ff52572`, 빨강 `a91e03fb7b`, 수정 `ffbb84769f`, 문서 `ba21bbd8a2`.
- 빨강(`a91e03fb7b`): 새 테스트 8개 중 AC-1, AC-3, AC-4, AC-10 의 4개가 단언에서 실패했고 AC-2, AC-5, AC-6, AC-7 의 4개는 통과했어요. Plan 3 은 AC-4 를 통과 쪽으로 적었지만 이 head 에서는 `restore_backup` 행이 없어서 실패해요.
- 초록: `moai tdd green` 이 빨강 때의 확인 명령을 다시 돌려 exit 0 이었어요.
- `ba21bbd8a2` 에서 AC-1~AC-10 의 verify 명령을 적힌 그대로 돌려 모두 exit 0 이었어요. AC 별 필터는 각각 테스트 1개, AC-8 의 타깃 다섯 개는 84개 통과, HOME 이름 표 검사는 1개 통과, AC-9 의 grep 은 2줄이에요. `cargo nextest run --test update_backup_replace` 는 31개 통과예요.
- 가드 뮤턴트(`ba21bbd8a2`, Plan 6 의 표): 8개 모두 지목한 테스트가 exit 100 으로 실패했고, 하나씩 되돌린 뒤 `git status --short` 가 비었어요.
- `src/recovery.rs`, `src/apply.rs`, `src/home_layout.rs`, `tools/repo_scans.json` 은 바꾸지 않았고, `HostReceipt` 와 `HostReceiptEntry` 의 필드도 그대로예요(diff 는 문서 주석과 `apply_update_in` 두 곳).
- 위험 경로 실행 근거: AC-1 의 테스트가 임시 HOME 에서 실제 바이너리로 교체, 보통 적용 두 번, `recovery-apply` 를 돌리고 원본의 바이트와 모드, `<backup>/replaced` 를 확인해요.

남은 것(코드 읽기만, 실행 안 함):

- 다른 빌드가 파일을 바꾸거나 더한 뒤, 원래 설치가 그 내용과 같은 패키지를 적용하면 쓰기가 없어서 이 설치의 entries 가 갱신되지 않아요. 그때 복원 제안은 `user_changed` 로 막혀요. 커밋하는 적용이 아니라서 REQ-1 밖이에요.
- `docs/rust-product-path.md` 의 손 복원 문단 마지막의 "reports the folder as unmanaged again" 은 이 변경 전부터 있던 문장이고, 영수증이 그 경로를 소유하면 맞지 않아 보여요.
- `src/recovery.rs` 의 `root_holds_only_ours` 주석("one this receipt created")은 범위 밖이라 낡은 채예요.
