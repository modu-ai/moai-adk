---
id: SPEC-20261125-update-installed-from-link
title: "moai update shows the version the execution link points to and keeps it in the receipt of an apply that writes nothing"
status: draft
revision: 2
created: 2026-11-25
---

# moai update shows the version the execution link points to and keeps it in the receipt of an apply that writes nothing

## Intent

`moai update` 화면의 `installed` 줄은 지금 호스트 영수증의 `package_version` 을 보여 줘요. 쓰기 없는 적용(바꿀 파일이 없는 호스트, 충돌만 있는 호스트)은 영수증을 남기지 않아서, 실행 링크가 이미 새 버전(예: 26)을 가리켜도 화면은 옛 번호(24)로 남아요. 이 SPEC 은 두 가지를 해요: 미리보기가 실행 링크의 대상 버전을 읽어 `installed` 줄에 보이고(읽기 전용), 쓰기 없는 적용도 이전 영수증이 있을 때 호스트 영수증에 이 빌드의 버전을 남겨요(HOME 상태 쓰기, 안전장치 넷).

## Requirements

- REQ-1: When `moai update` previews for a HOME, the preview shall carry `installed_version`, the version of the folder that the execution link under that HOME points to. · failure: none — a read-only key, and REQ-2 covers a link that names no version.
- REQ-2 (REQ-1): IF the execution link is missing, dangling, not a symlink, points outside the versions folder, or its folder name is not version-shaped, THEN the preview shall carry no `installed_version` key.
- REQ-3: When the preview screen shows the installed version, it shall prefer `installed_version` over the receipt's source version in English and in Korean. · failure: none — without the key the earlier header stays unchanged.
- REQ-4: When an apply writes no file for a host, the apply shall record this build's version in that host's receipt only when the previous receipt is committed or recovered and names another version. · failure: none — REQ-5 lists the cases that write nothing.
- REQ-5 (REQ-4): IF the previous receipt is absent, pending or names this build's version, or the host is blocked, THEN the apply shall write no receipt for that host.
- REQ-6: When the apply records the version for a host that wrote no file and this installation has its own host receipt, it shall change only that receipt's `package_version`, `finished_at` and owned list, taking the owned list from the previous receipt. · failure: none — every recovery field of the own receipt, the backup and the entries included, stays as it was.
- REQ-7: When the apply records the version for a host that wrote no file, it shall write the receipt through the existing host receipt writer only. · failure: none — one writer, the one every apply uses.
- REQ-8 (REQ-7): IF the receipt write fails, THEN the apply shall report the error as every other receipt write does.
- REQ-9 (REQ-2): IF the preview carries no `installed_version` key, THEN the header shall show the receipt's source version, or none without a receipt.
- REQ-10 (REQ-4): IF this installation's own host receipt exists and is neither committed nor recovered, THEN the apply shall write no receipt for that host.
- REQ-11: When the apply records the version for a host that wrote no file and this installation has no own host receipt, it shall write a new receipt with the previous receipt's owned list, no entries and no backup. · failure: none — no receipt of this installation exists to lose.
- REQ-12: When the apply records the version for a host that wrote no file, the host outcome shall report the saved receipt's status as `receipt_status`. · failure: none — the value of the file just written.

## Acceptance

- AC-1 (REQ-1, REQ-2): The link reader returns the folder name for a symlink into the versions folder in the fixed order of candidates, and none for a missing, dangling, plain-file, outside-layout or non-version link; the preview carries the key only with a link.
  verify: `cargo test --offline --locked --lib -- version_folder_tests`
- AC-2 (REQ-1, REQ-2): In a temporary HOME the preview carries `installed_version` when the link names a version folder, and the key is absent without a link.
  verify: `cargo test --offline --locked --test update_user_layer_cli -- the_preview_names_the_version_the_execution_link_points_to`
- AC-3 (REQ-3, REQ-9): The header reads `installed 4.2648.26 → package` in English and `설치본 4.2648.26 → 패키지` in Korean from the key, and keeps the receipt version without it.
  verify: `cargo test --offline --locked --test update_installed_by_name -- the_header_prefers_the_link_version_in_english_and_korean`
- AC-4 (REQ-4, REQ-5, REQ-6, REQ-7, REQ-12): In a temporary HOME, an apply that writes no file leaves this build's version in the host receipt, keeps the entries, operation, previous version and status of an own receipt, leaves the skill bytes unchanged, does the same for a conflict-only host, the outcome reports the saved receipt's status, and a second such apply writes no receipt.
  verify: `cargo test --offline --locked --test update_installed_by_name -- a_no_write_apply a_conflict_only_no_write_apply a_second_no_write_apply`
- AC-5 (REQ-5, REQ-8): A blocked or pending host and a previous receipt that names this build write no receipt, and a failing save reports an error.
  verify: `cargo test --offline --locked --test update_installed_by_name -- a_blocked_or_pending_host_writes_no_version_only_receipt`
- AC-6 (REQ-6, REQ-10): In a temporary HOME, after a no-write apply over a backup-then-replace receipt `recovery-plan` still offers `restore_backup` and `recovery-apply` restores the original bytes and modes, both for a version change under the same installation and for going back with `moai update --allow-older` in the version layout; an own receipt that is neither committed nor recovered stays byte-identical.
  verify: `cargo test --offline --locked --test update_backup_replace -- a_no_write_apply`
- AC-7 (REQ-11): In a temporary HOME, an installation with no own receipt writes a new receipt under its own id and leaves the earlier receipt as it was.
  verify: `cargo test --offline --locked --test update_installed_by_name -- a_no_write_apply_over_an_earlier_folders_receipt_writes_this_installations_receipt`

## Scope

- write: `src/update.rs`, `src/update_tux.rs`, `tests/update_installed_by_name.rs`, `tests/update_user_layer_cli.rs`, `tests/update_backup_replace.rs`, this SPEC folder
- read-only: `src/recovery.rs`, `src/gate.rs`, `src/home_root.rs`, `src/headver.rs`, `tests/update_tux_golden.rs`
- forbidden: changing a field of the update plan or its digest, deleting or moving a file, touching a blocked or pending host, reading the process HOME or PATH in the link reader, writing under the real HOME in any test

## Plan

1. Link reader in `src/update.rs` with its unit tests (read-only).
2. Preview key `installed_version`, only with a link, and its end-to-end test.
3. Header prefers the key, English and Korean, and its test.
4. Version-only receipt in the no-write branch of the apply, guarded as REQ-4 to REQ-8 say, with the tests of AC-4 and AC-5 in a temporary HOME.
5. Neighbours: the tests that read the preview key set, the header and the receipt version keep passing.
6. Repair after the first audit (revision 2): update the own receipt in place as REQ-6, REQ-10, REQ-11 and REQ-12 say, with the tests of AC-4, AC-6 and AC-7.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The execution link is read in the order `path/moai`, `path/moai-cli`, `bin/moai` | The product layout links `path/`; the reported machine has only `bin/moai` | Drop a candidate |
| D2 | The value is a conditional preview key `installed_version`; the update plan and its digest stay as they are | The plan is digest-bound and deserialised with `deny_unknown_fields`; a new field would change three tests and a document | Move the value into the plan with a digest change |
| D3 | A failing version-only save is an error, like every receipt write | One rule for receipt writes | Report it and keep the outcome |
| D4 | A conflict-only host is refreshed too | It is the reported screen: 45 up to date and 4 conflict | Refresh only hosts with no conflict |
| D5 | `installed` is the link's version as the card asks; skill copies that lag show in the rows | The card's criterion | Show both versions |
| D6 | Tier 2 by the floor list (`src/update`) and by the HOME state the receipt write touches; the link reader is read-only | CLAUDE.md names HOME state; the leader asked for this SPEC (card note) | The leader sets the tier |
| D7 | An existing own receipt is updated in place; only version, finish time and owned list change | The audit of the first head failed on data loss: a replaced receipt dropped the backup, so `recovery-plan` no longer offered the user's original folder | Write no version when the own receipt holds a backup |
| D8 | The owned list comes from the receipt the apply was planned against; the refreshed finish time makes this receipt the next baseline | The next preview must compare against the same list; recovery reads the entries, not the owned list | Keep the own owned list |

## Result

변경은 계획의 여섯 단계를 따랐어요. 링크 읽기(`execution_link_version`, `src/update.rs`), 미리보기 키 `installed_version`(`update_command_core`, `src/update.rs`)과 헤더가 그 키를 먼저 읽는 것(`src/update_tux.rs`), 쓰기 없는 적용의 버전만 남기는 영수증(`record_version_only`, `src/update.rs`)이에요. D2 대로 업데이트 계획과 `plan_digest` 는 그대로 두고 미리보기 키를 조건부로만 넣었어요.

첫 head `bcd2eeb26b6d6937c0375513fc9e61958085ec78` 는 독립 감사에서 데이터 손실로 FAIL 이었어요. 버전만 남기는 쓰기가 이 설치의 자기 영수증 위에 새 영수증을 만들어서 `backup` 과 `entries` 가 사라졌고, `recovery-plan` 이 사용자의 원래 폴더에 대한 `restore_backup` 을 더 이상 내놓지 않았어요. 개정 2 와 커밋 `164ab233eb13af8a0eb96e4770e188873f1d1778` 이 이를 고쳤어요(계획 6단계).

`record_version_only` 의 지금 동작:

- 기록 여부는 계획이 기준으로 삼은 영수증으로 정해요. 그 영수증이 committed 또는 recovered 이고 다른 버전을 가리킬 때만 이어져요.
- 이 설치의 자기 영수증이 있으면 그 자리에서 고쳐요. `package_version`, `finished_at`, owned 목록(기준 영수증의 것)만 바꿔요.
- 자기 영수증이 committed 도 recovered 도 아니면 아무것도 쓰지 않아요.
- 자기 영수증이 없을 때만 새로 써요. owned 목록은 기준 영수증의 것이고 `entries` 는 비어 있으며 `backup` 은 없어요.
- 저장은 `save_host_receipt` 로만 하고, 결과의 `receipt_status` 는 저장한 영수증의 상태예요(쓰지 않았으면 값이 없어요).
- blocked 와 pending 호스트는 이 분기에 닿기 전에 건너뛰어요(계획이 pending 영수증을 `preflight: blocked` 로 표시해요).

SPEC 의 verify 명령은 head `ee2b3c8612` 에서 적힌 그대로 돌렸고 모두 exit 0 이에요.

- AC-1: met — `cargo test --offline --locked --lib -- version_folder_tests` 4 passed, 0 failed (`src/update.rs` 의 `version_folder_tests`).
- AC-2: met — `the_preview_names_the_version_the_execution_link_points_to` (`tests/update_user_layer_cli.rs`) 1 passed. 링크가 없으면 키가 없고 `plan_digest` 는 그대로예요.
- AC-3: met — `the_header_prefers_the_link_version_in_english_and_korean` (`tests/update_installed_by_name.rs`) 1 passed. 키가 없으면 헤더 문구는 그대로예요.
- AC-4: met — verify 명령 4 passed: `a_no_write_apply_leaves_this_builds_version_in_the_receipt`, `a_conflict_only_no_write_apply_also_refreshes_the_receipt_version`, `a_second_no_write_apply_writes_no_receipt`, 그리고 이름 필터에 함께 걸리는 AC-7 의 테스트.
- AC-5: met — `a_blocked_or_pending_host_writes_no_version_only_receipt` 1 passed. unix 전용 저장 실패 경우(읽기 전용 영수증 폴더에서 적용이 오류를 보고함)를 포함하고, root 프로세스는 그 부분을 건너뛰어요. 이전 영수증이 이 빌드의 버전을 가리키는 경우는 AC-4 의 두 번째 적용 테스트가 확인해요.
- AC-6: met — `cargo test --offline --locked --test update_backup_replace -- a_no_write_apply` 3 passed: `a_no_write_apply_keeps_the_own_backup_that_recovery_restores`, `a_no_write_apply_going_back_keeps_the_own_backup_and_refreshes_the_receipt`, `a_no_write_apply_never_rewrites_an_own_failed_receipt`.
- AC-7: met — `a_no_write_apply_over_an_earlier_folders_receipt_writes_this_installations_receipt` (`tests/update_installed_by_name.rs`) 1 passed.

빨강 확인: head `f202e43af4`(`src` 는 감사받은 `bcd2eeb26b` 와 같아요)의 `src` 에 새 테스트를 돌렸어요. `update_backup_replace` 의 새 테스트 3개가 실패했어요(둘은 `restore_backup.action` 이 기대한 `restore_backup` 대신 Null, 셋째는 자기 failed 영수증이 다시 쓰임). `update_installed_by_name` 의 다시 쓴 테스트 2개도 실패했어요. `tests/update_backup_replace.rs` 는 그 뒤에 두 번 더 고쳐서(owned 구분 항목, 이어서 픽스처 패키지), head `ee2b3c8612` 에서 `src/update.rs` 만 `bcd2eeb26b` 의 것으로 바꿔 최종판을 다시 돌렸어요. 같은 5개가 의도한 단언에서 실패했고(`update_backup_replace` 0 passed 3 failed, `update_installed_by_name` 2 passed 2 failed, 각각 exit 101), 끝난 뒤 `src` 는 git 으로 되돌렸어요.

영수증 1차: head `f7c1fa3641` 의 `moai check affected --card` 는 ledger 단계에서 실패했어요. 되돌아가기 테스트가 저장소의 `skills/moai` 와 `agents` 를 복사해 읽었는데 그 파일을 이름 대는 스캔 항목이 없었어요. 커밋 `ee2b3c8612` 는 테스트가 버전 폴더에 픽스처 패키지와 빈 `agents` 폴더를 직접 만들게 해서 고쳤고, `repo_scans_table` 33 passed 와 lib 의 스캔 단계 테스트 2 passed 를 확인했어요. 스캔 표와 가드는 바꾸지 않았어요.

수리 코드의 초록 실행(커밋 `164ab233eb` 직전의 작업 트리): `update_backup_replace` 23, `update_installed_by_name` 10, `update_user_layer_cli` 24, `update_headver` 6, `update_known_good` 4, `state_folder_literal_scan` 10 모두 passed, lib 의 `version_folder_tests` 4 passed. `cargo clippy --offline --locked --lib --tests -- -D warnings` 는 exit 0 이에요.

수리 가드 뮤턴트(head `ee2b3c8612` 에서 `src/update.rs` 를 한 곳씩 바꾸고 해당 테스트만 돌렸어요. 여덟 개 모두 테스트 실패로 exit 101):

- (a) 자기 영수증 대신 기준 영수증의 복사본을 고침 — `a_no_write_apply_going_back_keeps_the_own_backup_and_refreshes_the_receipt`
- (b) `own.backup = None` 추가 — `a_no_write_apply_keeps_the_own_backup_that_recovery_restores`
- (c) `own.entries.clear()` 추가 — `a_no_write_apply_keeps_the_own_backup_that_recovery_restores`
- (d) `own.finished_at = Some(now)` 제거 — `a_no_write_apply_leaves_this_builds_version_in_the_receipt`
- (e) 자기 영수증 상태 가드(REQ-10) 제거 — `a_no_write_apply_never_rewrites_an_own_failed_receipt`
- (f) `own.owned = owned` 제거 — `a_no_write_apply_going_back_keeps_the_own_backup_and_refreshes_the_receipt`
- (g) `receipt_status` 를 committed 로 고정 — `a_no_write_apply_leaves_this_builds_version_in_the_receipt`
- (h) 새 영수증(REQ-11)을 저장하지 않고 `None` 반환 — `a_no_write_apply_over_an_earlier_folders_receipt_writes_this_installations_receipt`

위험한 경로: `ee2b3c8612` 빌드를 `env -i` 아래 임시 HOME 에서 돌렸고 결과는 PASS 예요. 두 길(같은 설치 id 에서 버전만 다른 경우, 버전 폴더 배치에서 `update --allow-older` 로 되돌아가는 경우) 모두 backup-then-replace 뒤 쓰기 없는 적용이 `unchanged` 였고, 자기 영수증의 `backup` 이 남았고, `recovery-plan` 이 `restore_backup` 을 내놓았고, `recovery-apply` 가 원래 바이트와 모드를 되돌렸어요. 버전 차이는 픽스처 영수증의 `package_version` 을 고쳐 만든 테스트 데이터예요.

알려진 한계:

- backup-then-replace 뒤에 파일을 쓰거나 baseline 을 옮기는 적용은 `backup` 이 없는 새 영수증을 만들어서, 그때는 `restore_backup` 을 더 이상 내놓지 않아요. 코드를 읽고 확인했고 테스트하지 않았어요. 이 SPEC 전부터 있던 동작이고 바꾸지 않았어요.
- 되돌아가기 길에서 새 빌드가 처음 쓰기 없이 남기는 자기 영수증에는 `backup` 이 없어서, 그 빌드에서 돌린 복구는 복원을 내놓지 않아요. 이 SPEC 전에는 그 빌드에 영수증이 아예 없었어요. 코드를 읽고 확인한 것이에요.

바뀐 파일: `src/update.rs`, `src/update_tux.rs`, `tests/update_installed_by_name.rs`, `tests/update_user_layer_cli.rs`, `tests/update_backup_replace.rs`, 이 SPEC 폴더.

Tier: tier 2 (D6: `src/update` 가 floor 목록에 있고 영수증 쓰기가 HOME 상태를 건드려요). 승인은 리더가 기록한 위임 승인이에요(개정 1, 이어서 개정 2). 개정 2 의 `approval_digest` 는 `sha256:df9f4e2ae7f30b36720b5ec3376d8d3dbd46175e917c53e6896afbf711d2e81f` 예요. 사용자가 직접 고른 승인이 아니에요. 이 파일의 frontmatter 는 아직 `status: draft` 이고 `approval` 키가 없어요.

Not run:

- 전체 `cargo test` — 실행하지 않았어요.
- `moai check affected --card` 영수증 — READY 때 부모가 돌려요.
- 이웃 테스트 `update_cli`, `exe_link_cli`, `update_tux_golden` — 수리 전에는 통과했고(11, 4, 17), 위의 수리 실행에서는 다시 돌리지 않았어요.
- 실제 `~/.moai` 와 `~/.claude` 를 대상으로 한 실행 — 하지 않았어요.
- Windows 와 Linux 결과 — Mac 픽스처만 썼어요.
- 가드 "이전 영수증 없음" — `update-apply` 를 통해서는 닿지 않아요. 그 경우는 adopt 경로로 가기 때문이에요. 그래서 `not_run` 이에요.
