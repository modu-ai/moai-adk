---
id: SPEC-20261125-update-installed-from-link
title: "moai update shows the version the execution link points to and keeps it in the receipt of an apply that writes nothing"
status: draft
revision: 1
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
- REQ-6: When the apply records the version for a host that wrote no file, it shall copy the previous receipt's owned list unchanged. · failure: none — a copy of an existing list.
- REQ-7: When the apply records the version for a host that wrote no file, it shall write the receipt through the existing host receipt writer only. · failure: none — one writer, the one every apply uses.
- REQ-8 (REQ-7): IF the receipt write fails, THEN the apply shall report the error as every other receipt write does.
- REQ-9 (REQ-2): IF the preview carries no `installed_version` key, THEN the header shall show the receipt's source version, or none without a receipt.

## Acceptance

- AC-1 (REQ-1, REQ-2): The link reader returns the folder name for a symlink into the versions folder in the fixed order of candidates, and none for a missing, dangling, plain-file, outside-layout or non-version link; the preview carries the key only with a link.
  verify: `cargo test --offline --locked --lib -- version_folder_tests`
- AC-2 (REQ-1, REQ-2): In a temporary HOME the preview carries `installed_version` when the link names a version folder, and the key is absent without a link.
  verify: `cargo test --offline --locked --test update_user_layer_cli -- the_preview_names_the_version_the_execution_link_points_to`
- AC-3 (REQ-3, REQ-9): The header reads `installed 4.2648.26 → package` in English and `설치본 4.2648.26 → 패키지` in Korean from the key, and keeps the receipt version without it.
  verify: `cargo test --offline --locked --test update_installed_by_name -- the_header_prefers_the_link_version_in_english_and_korean`
- AC-4 (REQ-4, REQ-5, REQ-6, REQ-7): In a temporary HOME, an apply that writes no file leaves this build's version in the host receipt with the owned list and the skill bytes unchanged, a conflict-only host does the same, and a second such apply writes no receipt.
  verify: `cargo test --offline --locked --test update_installed_by_name -- a_no_write_apply a_conflict_only_no_write_apply a_second_no_write_apply`
- AC-5 (REQ-5, REQ-8): A blocked or pending host and a previous receipt that names this build write no receipt, and a failing save reports an error.
  verify: `cargo test --offline --locked --test update_installed_by_name -- a_blocked_or_pending_host_writes_no_version_only_receipt`

## Scope

- write: `src/update.rs`, `src/update_tux.rs`, `tests/update_installed_by_name.rs`, `tests/update_user_layer_cli.rs`, this SPEC folder
- read-only: `src/gate.rs`, `src/home_root.rs`, `src/headver.rs`, `tests/update_tux_golden.rs`
- forbidden: changing a field of the update plan or its digest, deleting or moving a file, touching a blocked or pending host, reading the process HOME or PATH in the link reader, writing under the real HOME in any test

## Plan

1. Link reader in `src/update.rs` with its unit tests (read-only).
2. Preview key `installed_version`, only with a link, and its end-to-end test.
3. Header prefers the key, English and Korean, and its test.
4. Version-only receipt in the no-write branch of the apply, guarded as REQ-4 to REQ-8 say, with the tests of AC-4 and AC-5 in a temporary HOME.
5. Neighbours: the tests that read the preview key set, the header and the receipt version keep passing.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The execution link is read in the order `path/moai`, `path/moai-cli`, `bin/moai` | The product layout links `path/`; the reported machine has only `bin/moai` | Drop a candidate |
| D2 | The value is a conditional preview key `installed_version`; the update plan and its digest stay as they are | The plan is digest-bound and deserialised with `deny_unknown_fields`; a new field would change three tests and a document | Move the value into the plan with a digest change |
| D3 | A failing version-only save is an error, like every receipt write | One rule for receipt writes | Report it and keep the outcome |
| D4 | A conflict-only host is refreshed too | It is the reported screen: 45 up to date and 4 conflict | Refresh only hosts with no conflict |
| D5 | `installed` is the link's version as the card asks; skill copies that lag show in the rows | The card's criterion | Show both versions |
| D6 | Tier 2 by the floor list (`src/update`) and by the HOME state the receipt write touches; the link reader is read-only | CLAUDE.md names HOME state; the leader asked for this SPEC (card note) | The leader sets the tier |

## Result

변경은 계획의 다섯 단계를 그대로 따랐어요. 링크 읽기(`execution_link_version`, `src/update.rs`), 미리보기 키 `installed_version`(`update_command_core`, `src/update.rs`)과 헤더가 그 키를 먼저 읽는 것(`src/update_tux.rs`), 쓰기 없는 적용의 버전만 남기는 영수증(`record_version_only`, `src/update.rs`)이에요. D2 대로 업데이트 계획과 `plan_digest` 는 그대로 두고 미리보기 키를 조건부로만 넣었어요. 영수증은 이전 영수증이 committed 또는 recovered 이고 다른 버전을 가리킬 때만 쓰고, owned 목록은 그대로 복사하고, `save_host_receipt` 로만 저장해요. blocked 와 pending 호스트는 이 분기에 닿기 전에 건너뛰어요(계획이 pending 영수증을 `preflight: blocked` 로 표시해요).

- AC-1: met — `cargo test --offline --locked --lib -- version_folder_tests` 4 passed, 0 failed (`src/update.rs` 의 `version_folder_tests`).
- AC-2: met — `the_preview_names_the_version_the_execution_link_points_to` (`tests/update_user_layer_cli.rs`) 통과, 같은 바이너리 24 passed. 링크가 없으면 키가 없고 `plan_digest` 는 그대로예요.
- AC-3: met — `the_header_prefers_the_link_version_in_english_and_korean` (`tests/update_installed_by_name.rs`) 통과, 같은 바이너리 9 passed. 키가 없으면 헤더 문구는 그대로예요.
- AC-4: met — `a_no_write_apply_leaves_this_builds_version_in_the_receipt`, `a_conflict_only_no_write_apply_also_refreshes_the_receipt_version`, `a_second_no_write_apply_writes_no_receipt` 통과.
- AC-5: met — `a_blocked_or_pending_host_writes_no_version_only_receipt` 통과. unix 전용 저장 실패 경우(읽기 전용 영수증 폴더에서 적용이 오류를 보고함)를 포함해요. root 프로세스는 그 부분을 건너뛰어요. 이전 영수증이 이 빌드의 버전을 가리키는 경우는 AC-4 의 두 번째 적용 테스트가 확인해요.

이웃 테스트(통과): `update_cli` 11, `update_headver` 6, `update_known_good` 4, `exe_link_cli` 4, `update_tux_golden` 17, `update_user_layer_cli` 24, `update_installed_by_name` 9. `cargo clippy --lib --tests -D warnings` 는 경고가 없어요.

가드 뮤턴트(한 줄씩 바꾸고 해당 테스트만 돌렸고, exit 101 은 테스트 실패로 잡힌 거예요): m1 본문이 `Ok(false)` 를 반환, m2 상태 match 에서 `recovered` 제거, m3 같은 버전 가드 제거, m4 상태 가드 제거, m5 저장 오류를 `let _ =` 로 버림. 다섯 개 모두 exit 101 이었어요.

위험한 경로: 적용 테스트 넷은 임시 HOME 픽스처에서 실제 적용을 돌렸고, 실제 HOME 상태는 건드리지 않았어요.

바뀐 파일: `src/update.rs`, `src/update_tux.rs`, `tests/update_installed_by_name.rs`, `tests/update_user_layer_cli.rs`, 이 SPEC 폴더.

Tier: tier 2 (D6: `src/update` 가 floor 목록에 있고 영수증 쓰기가 HOME 상태를 건드려요). 승인은 리더가 위임 승인으로 기록했고 `approval_digest` 는 `sha256:b42d5ee16fd382b59ac676f75a58b28669cab46a9d2820b1696e560d29f973ff` 예요. 사용자가 직접 고른 승인이 아니에요. 이 파일의 frontmatter 는 아직 `status: draft` 이고 `approval` 키가 없어요.

Not run:

- 전체 `cargo test` — 실행하지 않았어요.
- `moai check affected --card` 영수증 — READY 때 부모가 돌려요.
- 실제 `~/.moai` 와 `~/.claude` 를 대상으로 한 실행 — 하지 않았어요.
- Windows 와 Linux 결과 — Mac 픽스처만 썼어요.
- 가드 "이전 영수증 없음" — `update-apply` 를 통해서는 시험할 수 없어요. 그 경우는 adopt 경로로 가기 때문이에요. 그래서 `not_run` 이에요.
