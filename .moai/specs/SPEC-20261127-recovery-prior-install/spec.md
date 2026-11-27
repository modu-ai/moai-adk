---
id: SPEC-20261127-recovery-prior-install
title: "recovery: 버전 업그레이드 뒤 새 설치에서도 직전 설치의 백업 복원 제안을 이어서 보여요"
status: draft
revision: 1
created: 2026-11-27
---

# recovery: 버전 업그레이드 뒤 새 설치에서도 직전 설치의 백업 복원 제안을 이어서 보여요

## Intent

backup-then-replace(`update-apply --confirm-backup-replace`)는 원래 폴더를 백업으로 통째로 옮기고, 그 설치의 영수증 `backup` 으로만 복원을 제안해요. 배포 레이아웃(`versions/<버전>/bin`)에서는 새 버전이 새 설치 id 라서, 업그레이드하면 `recovery-plan` 이 새 id 의 영수증만 읽어 그 제안이 사라져요. 원본 폴더는 백업에 그대로 있는데 되돌리는 길이 보이지 않는 거예요. 이 SPEC 은 업그레이드 뒤에도 직전 설치의 복원 제안을 이어서 보이고, 따라 했을 때 지우는 파일이 없게 해요. 앞의 SPEC(`SPEC-20261126-keep-restore-backup`)이 D-4 로 따로 열기로 한 후속이에요.

**이 SPEC 이 닿는 범위(승인 전에 확인해 주세요).** 아래는 코드를 읽은 결과이고 실행해 보지는 않았어요.

- 업그레이드가 만드는 모양은 둘이에요. (a) 새 빌드 B 가 설치만 되고 아직 적용하지 않았어요(B 의 영수증 없음). (b) B 가 보통 적용을 했어요(B 의 영수증은 `committed`, `backup` 없음, 쓴 파일만 entries). `moai update` 는 설치 뒤 B 로 적용하는 길이라 실제 업그레이드는 대부분 (b) 예요. (a) 만 고치면 업그레이드 직후의 짧은 구간에만 제안이 보여요.
- (b) 를 덮으려면 앞의 SPEC 이 못박은 것 하나를 뒤집어야 해요. 그 SPEC 의 AC-6(`keep_restore_is_not_taken_from_another_installation`)의 글, 곧 B 의 새 영수증에 `backup` 이 없고 A 의 영수증 바이트가 그대로라는 것은 바뀌지 않아요. 바뀌는 것은 그 테스트의 도우미 `assert_writes_only_and_no_offer` 가 B 의 `recovery-plan` 에 `restore_backup` 행이 없다고 단언하는 부분이에요(`tests/update_backup_replace.rs`).
- 복원은 지금 skill 폴더를 `<backup>/replaced` 로 옮기고 원본을 되돌려요. 지우는 파일은 없어요. B 가 쓴 파일은 `<backup>/replaced` 에 남아요.
- 승인 전에 정할 것은 아래 Decisions 의 D-1(어느 모양까지), D-2(어느 영수증을 닫을지), D-3(제안이 막힐 때 행을 보일지)예요.

## Interview

- Q1: 업그레이드의 어느 모양까지 덮나요? → A: 기본값은 (a) 와 (b) 둘 다예요. 실제 업그레이드가 (b) 라서 (a) 만으로는 쓸모가 작아요(D-1).
- Q2: 복원이 끝난 뒤 어느 영수증을 닫나요? → A: 기본값은 백업을 가진 직전 설치의 영수증과, `committed` 인 지금 설치의 영수증 둘 다 `recovered` 로 닫아요. 둘 다 이제 없는 파일을 소유한다고 적고 있기 때문이에요(D-2).
- Q3: 복원이 막힐 때(사용자 편집, 백업이 사라짐) 다른 설치의 행을 보이나요? → A: 기본값은 행을 보이지 않아요. 다른 설치의 막힌 백업이 이 설치의 `recovery-apply` 를 `partial` 로 만들면 안 돼요(D-3).
- Q4: 앞의 SPEC 의 AC-6 은 바뀌나요? → A: AC-6 의 글은 그대로예요. 그 테스트가 쓰는 도우미의 "복원 행 없음" 단언만 이 SPEC 의 AC-1 로 옮겨요.
- Q5: 승인 노트의 선택은 누가 정하나요? → A: 사용자의 직접 선택이에요. 이 SPEC 을 쓴 레인이나 리더가 정한 것으로 읽지 않아요.

## Non-goals

- `src/update.rs`, `src/home_layout.rs`, 영수증의 필드와 스키마를 바꾸는 것. 업데이트가 쓰는 영수증은 그대로예요.
- 직전 설치가 아닌 더 오래된 설치의 백업을 여럿 보이는 것. 가장 새 하나만 다뤄요.
- 실패하거나 중단된 적용(`pending`)의 복구. 파일별 복구가 그대로 먼저예요.
- 백업 보존과 정리.

## Requirements

- REQ-1: When `recovery-plan` reads a host whose receipt under this installation's id is absent, or is `committed` or `recovered` without a `backup`, and exactly one earlier installation in the same state folder holds the newest `committed` receipt of that host with a `backup` whose folder has no `RESTORED.json`, the plan shall classify that backup as it classifies an own one and, when the action is `restore_backup`, list it as the host's `restore_backup` with the earlier installation's id in a field `from_install`.
- REQ-2 (REQ-1): While this installation holds a receipt of that host, the plan shall list a restore from an earlier installation only when every file under the projection root matches, by path and bytes, an entry of the earlier receipt or an entry of this installation's receipt.
- REQ-3 (REQ-1): IF this installation's receipt of the host is `pending` or unreadable, THEN the plan shall list no restore from an earlier installation.
- REQ-4 (REQ-1): IF the newest candidates tie in finish time, or a candidate's receipt is unreadable, a symlink, or names another installation or host, THEN the plan shall list no restore from it and plan the other hosts as before.
- REQ-5 (REQ-1): When `recovery-apply` runs a restore with `from_install`, it shall read that receipt again, classify again right before any move, move the installed folder into `<backup>/replaced` and the original back, delete no file, and save the earlier receipt and this installation's `committed` receipt as `recovered` with no owned files.
- REQ-6 (REQ-5): IF the classification right before the move differs from the plan, THEN the recovery shall move nothing and write nothing.
- REQ-7: The plan shall carry `from_install` only for a restore taken from an earlier installation, so that the plans of today stay byte for byte the same.
- REQ-8: The recovery shall change no code outside `src/recovery.rs` and no receipt field.

## Acceptance

- AC-1 (REQ-1, REQ-2): 교체한 빌드 A 의 설치 뒤 B 가 내용을 바꾸고 더한 단계 패키지로 보통 적용을 하면, B 의 `recovery-plan` 의 `restore_backup.action` 이 `restore_backup`, `original_at` 이 `backup`, `from_install` 이 A 의 id 예요. B 의 영수증에는 `backup` 이 없고 A 의 영수증 바이트는 그대로예요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_offers_the_restore_after_an_ordinary_apply`
- AC-2 (REQ-1): B 를 설치만 하고 적용하기 전에도 B 의 `recovery-plan` 이 같은 제안을 내요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_offers_the_restore_before_the_new_build_applies`
- AC-3 (REQ-5, REQ-6): 제안을 따라 `recovery-apply` 하면 임시 HOME 의 어떤 파일도 사라지지 않아요. 원본은 바이트와 모드까지 돌아오고, B 의 패키지는 `<backup>/replaced` 에, `RESTORED.json` 은 백업 폴더에 있고, 두 영수증은 `recovered` 예요. 위험한 길이라 READY 에 이 실행의 결과를 적어요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_restore_moves_and_deletes_nothing`
- AC-4 (REQ-2): 설치 폴더에 두 영수증이 모르는 파일이 있으면 제안이 없고, `recovery-apply` 는 아무것도 옮기지 않아요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_offer_is_blocked_by_a_user_edit`
- AC-5 (REQ-3): B 의 영수증이 `pending` 이면 A 의 제안이 없고 파일별 복구가 그대로예요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_offer_waits_for_own_pending_recovery`
- AC-6 (REQ-4): 백업을 가진 영수증 둘이 같은 시각에 끝났으면 제안이 없어요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_offer_is_not_guessed_between_two_backups`
- AC-7 (REQ-1, REQ-5): 복원에 쓴 백업은 B 나 새 빌드 C 의 `recovery-plan` 에 다시 나오지 않아요 · verify: `cargo test --locked --offline --test update_backup_replace prior_install_used_backup_is_not_offered_again`
- AC-8 (REQ-7, REQ-8): 지금 설치가 자기 `backup` 을 가진 경우의 제안은 전과 같고(`keep_restore_*` 테스트), 이웃 테스트와 HOME 이름 표 검사가 통과하고, `git diff <base>..<head> --stat -- src` 가 `src/recovery.rs` 하나예요 · verify: `cargo test --locked --offline --test update_backup_replace --test recovery_cli --test update_conflict_baseline_cli --test update_adopt_cli`; `cargo test --offline --locked --lib every_rust_written_home_name_is_in_the_table`
- AC-9 (REQ-1): `docs/rust-product-path.md` 의 복원 안내가 업그레이드 뒤 직전 설치의 제안을 적어요 · verify: `grep -n 'earlier installation' docs/rust-product-path.md`

## Scope

- write: `src/recovery.rs`, `tests/update_backup_replace.rs`, `.moai/specs/SPEC-20261127-recovery-prior-install/**`
- write (카드 범위 밖, `card line-lease` 로 잠깐 빌려 씀): `docs/rust-product-path.md`
- read-only: `src/update.rs`, `src/apply.rs`, `src/home_layout.rs`
- forbidden: `src/update.rs`, `src/home_layout.rs`, `tools/repo_scans.json`; 실제 사용자 HOME 의 설치 파일은 건드리지 않아요

## Plan

1. 빨강: `tests/update_backup_replace.rs` 에 AC-1~7 의 테스트 일곱 개를 더하고 `assert_writes_only_and_no_offer` 의 "복원 행 없음" 단언을 AC-1 로 옮겨요. 상태 폴더 이름은 `state_dir` 와 `put_build` 도우미로만 다뤄요. AC-1, AC-2, AC-3, AC-7 은 지금 코드에서 실패해야 하고 AC-4~6 은 지켜야 할 쪽을 고정해서 실제 결과를 그대로 적어요.
2. 수정(`src/recovery.rs`): 지금 설치에 자기 `backup` 이 없을 때만 `installations/` 의 다른 설치 영수증을 읽어 후보를 고르고, `RestoreBackup` 에 `from_install` 을 더하고, `root_holds_only_ours` 가 두 영수증의 entries 를 함께 보게 하고, `restore_backup_host` 가 `from_install` 의 영수증을 읽고 닫아요. 기존 선택 로직 `resolve_host_receipt` 는 가장 새 영수증 하나를 돌려줘서 (b) 의 A 를 못 고르므로 쓰지 않고, 같은 점검(심볼릭 링크, id 와 호스트 이름 일치)을 따라 해요.
3. 초록과 이웃: AC-8 의 타깃을 heavy slot 에서 돌려요.
4. `moai bundle check` 로 floor 를 확인해요(`src/recovery.rs` 는 바닥 목록 밖이라 1 로 읽히지만 복원 경로라서 감사에 이 점을 적어요).
5. 가드 뮤턴트(하나씩, 되돌린 뒤 트리 확인): 후보를 고르는 조건, 두 영수증의 entries 합치기, 재분류 비교, 두 번째 영수증 닫기.
6. `cargo fmt --all`, `git diff --check`, 커밋, bare `moai check affected --card <base>..<head>`, READY. AC-3 의 테스트를 위험 경로 실행 근거로 적어요.

## Risks

- 업그레이드 뒤 오래 지난 설치에서도 `recovery-apply` 가 원본 폴더를 되돌리고 지금 skill 폴더를 `<backup>/replaced` 로 옮겨요. 지우는 것은 없지만 되돌리면 새 버전의 skill 이 사라지고 `moai update` 가 폴더를 다시 관리되지 않는 것으로 봐요. 제안 행이 이 사실을 알려야 해요.
- 처음으로 `recovery-apply` 가 다른 설치 id 폴더의 영수증을 써요(D-2). 그 쓰기가 실패하면 폴더는 옮겨졌는데 영수증이 `committed` 로 남아 다음 `recovery-plan` 이 `release` 를 보일 수 있어요(원본이 이미 돌아와 있어 아무것도 옮기지 않고 영수증만 닫아요).
- 후보가 하나뿐이라고 가정해요. 같은 호스트에 백업을 가진 영수증이 둘 있으면 가장 새 하나만 보고, 같은 시각이면 아무것도 보이지 않아요(AC-6).
- `src/recovery.rs` 는 `src/gate.rs` 바닥 목록에 없어서 floor 가 1 이지만 HOME 안의 폴더를 옮기는 코드예요. 이 SPEC 이 승인될 때까지 구현하지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | (a) 와 (b) 둘 다 덮어요. | 실제 업그레이드는 설치 뒤 B 가 적용하는 (b) 예요. (a) 만 고치면 적용 전 짧은 구간에만 보여요. | (a) 만 남기고 AC-1 과 REQ-2 를 빼요. |
| D-2 | 복원 뒤 직전 설치의 영수증과 `committed` 인 지금 설치의 영수증을 둘 다 `recovered` 로 닫아요. | 둘 다 이제 없는 파일을 소유한다고 적고 있어요. 하나만 닫으면 다음 `update-plan` 의 기준 영수증 선택이 가장 새 완료 시각을 따라가서 레이아웃 밖 설치에서는 열린 쪽이 기준이 돼요. | 지금 설치의 영수증은 닫지 않고 직전 설치 것만 닫아요. |
| D-3 | 다른 설치의 제안이 막히면(사용자 편집, 백업 사라짐, 매니페스트 불일치) 행을 보이지 않아요. | 다른 설치의 막힌 백업이 이 설치의 `recovery-apply` 를 `partial` 로 만들면 안 돼요. 사용자가 오래된 백업을 지운 HOME 에서 모든 업그레이드가 `none` 행을 만들게 돼요. | `none` 행을 보이도록 REQ-1 의 "when the action is `restore_backup`" 를 빼요. |
| D-4 | 새 필드는 `RestoreBackup.from_install` 하나이고 `schema_version` 은 1 그대로예요. | 값이 있을 때만 쓰는 덧붙임이고 읽는 곳은 `recovery.rs` 뿐이에요(`deny_unknown_fields` 라서 옛 바이너리가 만든 계획을 새 쪽이 읽는 데는 문제가 없어요). | `schema_version` 을 2 로 올려요. |
| D-5 | 후보는 `installations/` 의 폴더를 읽어 고르고 `update.rs` 의 레이아웃 선택(`resolve_host_receipt`)은 쓰지 않아요. | 그 함수는 가장 새 영수증 하나만 돌려줘서 (b) 에서 B 자신이 이겨요. 이름을 바꾸거나 pub(crate) 를 더하려면 `src/update.rs`(tier 2 바닥)를 건드려야 해요. | `update.rs` 에 후보 목록을 돌려주는 함수를 더하고 tier 2 감사를 받아요. |
