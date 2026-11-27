---
id: SPEC-20261127-update-leftovers-action
title: "Update gives each leftover of an earlier generation an archive action or a recorded keep, with a receipt and a revert, and the managed ignore block names the Rust set-aside folders"
status: draft
revision: 1
created: 2026-11-27
---

# Update gives each leftover of an earlier generation an archive action or a recorded keep, with a receipt and a revert, and the managed ignore block names the Rust set-aside folders

## Intent

`project-update-plan`은 Rust 세대로 옮긴 프로젝트에 이전 세대가 남긴 것(`.moai-cli/` 폴더, 죽은 Go 업데이트 lock)을 `leftovers` 배열에 사실로만 알려 줘요. 조치도, "남기기"를 고르는 길도 없어서 `/moai update`는 이것들을 끝까지 정리하지 못하고, `.moai.rust-<UTC>/`(go-state 되돌리기가 Rust 상태를 치워 둔 폴더)와 `.moai-backups/`는 배열에 아예 없어요. 사용자 지시(2026-11-21): Go 세대에서 완벽하게 마이그레이션되어야 해요. 조사 보고서 4-8절의 규칙을 따라요: 지우지 않고 옮겨요, 미리보기 → 고르기 → 적용 → 영수증 → 되돌리기, 멱등.

이 SPEC은 조각 S5예요. 다섯 가지(`.moai-cli/`, `.moai.rust-<UTC>/`, `.moai-backups/`, `.moai/.update.lock`, `.moai-backup/.update.lock`) 각각에 보관 명령 또는 "보관할 수 없는 이유"를 붙이고, 어느 쪽이든 "남기기"를 기록할 수 있게 해요. 기록한 남기기는 `all_unchanged`에 반영돼요. 보관은 새 명령 둘(apply, revert)과 영수증이 맡고, 관리 무시 블록은 `.moai.rust-*/` 한 줄을 더 가져요. 폴더를 옮기므로 tier 2예요.

## Non-goals

- `state_root::KEPT`(관리 블록의 `!` 허용 목록)를 Rust 소유 하위 경로로 좁히지 않아요. `KEPT`는 `classify`와 go-state의 `kept_name`도 읽고, 좁히는 일은 승인된 `SPEC-20261120-moai-layout`의 규칙을 바꿔요. 후속 SPEC에서 사용자 승인 뒤에 해요(D-2). 이 SPEC은 "이미 git이 추적하는 파일은 그대로 추적된다"만 테스트로 고정해요(AC-14).
- 단계(step)를 새로 더하지 않아요. `ORDER` 핀을 쥔 `update_end_to_end_cli`·`project_update_plan_cli`의 단계 목록은 그대로예요.
- `CLAUDE.md` leftover는 사실로만 남아요. 다른 카드(claude-md 단계)의 일이에요.
- 이름 붙은 `.moai-backup-<UTC>/` 안의 lock은 후보가 아니에요(D-9). 정확히 `.moai-backup/.update.lock` 한 경로만 봐요.
- `src/go_state.rs`, `src/go_residue.rs`, `src/go_residue_apply.rs`의 동작은 바꾸지 않아요. go-residue 조사가 `.moai-cli`와 `.moai-backups`를 `leftover_folder`로 이미 내지만 그 apply가 옮기지 않는 중복 신호는 `docs/later.md` 한 줄로만 남겨요.
- `skills/**`는 바꾸지 않아요. `update.md:44`가 이미 "leftover마다 조치 또는 사용자가 남김을 보이고, 답 없는 leftover는 남은 것으로 말하라"고 해요.
- git 색인은 건드리지 않아요. 옮긴 추적 파일은 `git status`에 삭제로 보이고, 결과가 "커밋하지 않으면 checkout 때 되살아나요"라고 말해요. 마무리 커밋 제안은 이 SPEC의 일이 아니에요.
- Windows·Linux 결과를 주장하지 않아요. 다른 파일 시스템 이동은 Mac 픽스처로 실제 재현할 수 없어서 시험용 고장 지점으로만 증명해요(결과에 "simulated"라고 적어요).

## Requirements

- REQ-1: When `project-update-plan` runs on a project whose state generation is exactly `rust`, each listed leftover of the five kinds shall carry an `action` (the argv of `project-update-leftovers-apply` for it) or a `keep_only` reason, and a `keep` argv (that of `project-update-decline`).
- REQ-2 (REQ-1): IF another backup-phase step still has a command to run, THEN every leftover shall print its `keep` argv and no `action`, because that move would make the other step's plan digest stale.
- REQ-3: A folder leftover shall be archivable only when its path passes the one path guard (`go_residue::contained`), the state generation is `rust`, `go_state::nested_checkouts_under` finds no nested checkout in it and it is on the same filesystem as `.moai/backups`; every other folder leftover is `keep_only` with the reason.
- REQ-4: The `.moai-cli/` leftover shall be archivable only when no go-state receipt inside it is pending, unreadable, over the receipt limit or in an unknown state and no merge step still reads a backed-up copy (`backed_up_root_file`, `backed_up_settings`) from it.
- REQ-5: A `.moai.rust-<UTC>/` leftover shall be archivable only when its name passes `state_root::utc_stamp` after the prefix and every go-state receipt inside it is `reverted`, because a `committed` one means a revert stopped.
- REQ-6: The `.moai-backups/` leftover shall be archivable as an exact-name real folder with no receipt rule.
- REQ-7: A lock at `.moai/.update.lock` or `.moai-backup/.update.lock` shall be archivable only as a regular file of at most 4096 bytes whose JSON `pid` names a process that `home_state::pid_alive` reports not running.
- REQ-8: `all_unchanged` shall be false while a leftover of the five kinds is neither archived nor kept for its current source sha256, and `CLAUDE.md` is not one of the five kinds.
- REQ-9: When `project-update-decline` receives `--step leftover:<path>` with the current source sha256 of one of the five kinds, it shall record the keep in `manifests/update-declined.json`.
- REQ-10 (REQ-9): IF the sha256 differs from the current source or the path is not one of the five kinds, THEN `project-update-decline` shall refuse and record nothing.
- REQ-11: The source sha256 of a folder leftover shall cover its sorted top-level entry names, kinds and file sizes, that of a lock the sha256 of its bytes, so a changed source reads as open again and `--ask-again` clears the record.
- REQ-12: When the user applies a previewed plan, `project-update-leftovers-apply <project> --plan-digest D --permission-reference R --item P…` shall move each chosen item by one rename (folder) or one link-then-unlink (file) to `.moai/backups/<UTC>/leftovers/<same path>`, never overwriting and deleting nothing.
- REQ-13 (REQ-12): IF the plan digest is stale, the permission reference is rejected, an item is unknown, or a path is absolute, holds `..` or passes through a link, THEN `project-update-leftovers-apply` shall refuse and write nothing.
- REQ-14 (REQ-12): IF the preflight (same filesystem, fresh eligibility, the lock's process still not running) fails, THEN `project-update-leftovers-apply` shall stop before it writes the pending receipt.
- REQ-15: When the apply starts, `project-update-leftovers-apply` shall save a pending receipt `.moai/backups/<UTC>/leftovers-receipt.json` under the one size limit its reader uses, read it back before the first move and mark it `committed` only after every move is read back.
- REQ-16 (REQ-15): IF the pending receipt cannot be saved, read back or kept under that limit, THEN `project-update-leftovers-apply` shall move nothing and report the failure.
- REQ-17: The receipt shall hold per item only the kind, the entry count, the size and the top-level listing digest, never a file's contents or child names.
- REQ-18: `project-update-leftovers-apply` shall count secret-shaped names inside a moved folder and never open them.
- REQ-19: When the user reverts a receipt, `project-update-leftovers-revert <project> --receipt <UTC>` shall put an item back only when its original place is empty and the archived listing digest equals the receipt's, judging by the disk and not by the receipt's `moved` flags.
- REQ-20 (REQ-19): IF the original place is taken, the archived digest differs or an item path fails the path guard, THEN `project-update-leftovers-revert` shall report that item as `conflict` and leave both sides untouched.
- REQ-21: The apply result shall report how many moved files git tracks and that a checkout brings them back unless the deletions are committed.
- REQ-22: The printed `undo_order` shall begin with the leftovers revert only while a leftovers receipt that is not `reverted` exists, so an archived `.moai-backup/.update.lock` goes back before `go-state-revert` reads the folder digest.
- REQ-23: The managed ignore block shall end with the line `.moai.rust-*/` written verbatim (a glob, not through `escape_ignore_path`), and `state_root::KEPT` stays as it is.
- REQ-24: The finish section of `project-update-guide` shall list the open leftovers in `finish.open` and not claim that every step is settled while one is open.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-8): in a temp `rust` project with `.moai-cli/` holding a committed receipt, `.moai.rust-20261101T000000Z/` holding a reverted one, `.moai-backups/` and a dead lock at each path, the plan lists the five items with an `action` and a `keep`, prints no `action` while an earlier backup step still has a command, and never counts `CLAUDE.md` · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_plan_gives_each_item_an_action_or_a_keep_only_reason`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-6): a `.moai-cli/` with a pending receipt, a `.moai.rust-*` with a committed receipt, a folder with a nested `.git`, and a stamp-less `.moai.rust-notes/` are each `keep_only` with their reason · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_folder_with_an_unfinished_receipt_is_keep_only`
- AC-3 (REQ-12, REQ-15, REQ-17, REQ-19): applying the folders and the lock moves them, a second plan shows them as gone, and reverting restores every byte and every listing digest · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_moves_and_revert_puts_back_every_byte`
- AC-4 (REQ-13): a plan digest from before a change, a made-up digest, a malformed permission reference and an unknown item each leave the whole tree identical before and after · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_refuses_stale_digest_bad_reference_and_unknown_item`
- AC-5 (REQ-13, REQ-20): an item that is a link, an absolute path or holds `..` is never moved, and a forged receipt that names a path outside the project makes the revert report that item as `conflict` before it moves anything (the outside file's sha256 is the same before and after) · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_never_moves_a_link_or_an_escaping_path`
- AC-6 (REQ-7, REQ-14): a lock whose pid is alive at apply time, or whose bytes changed after the plan, is not moved · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_lock_is_archived_only_while_its_process_is_dead`
- AC-7 (REQ-8, REQ-9, REQ-10, REQ-11): recording a keep makes the item `kept` and `all_unchanged` true; a keep with a wrong sha256 or for another path is refused; changing a top-level name or size of the folder asks again; `--ask-again` clears the record · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_keep_is_recorded_counted_and_asked_again_when_the_folder_changes`
- AC-8 (REQ-14): with the cross-device fault point set, the apply stops before any move and before the pending receipt, and says the check was simulated · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_stops_before_any_move_on_another_filesystem`
- AC-9 (REQ-21): on a git project where the moved files are tracked, the apply result names the tracked count and the checkout warning · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_says_how_many_tracked_files_moved`
- AC-10 (REQ-18): a secret-shaped file with mode 000 inside an archived folder moves without being read and is counted, not named · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_never_opens_a_secret_shaped_name`
- AC-11 (REQ-22): the printed undo order has no leftovers row before an apply and starts with it after one · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_undo_order_names_the_revert_only_after_an_apply`
- AC-12 (REQ-23): git ignores `.moai.rust-20261101T000000Z/x` under the managed block and the block's last line is `.moai.rust-*/` · verify: `cargo test --locked --offline --test gitignore_block_cli the_block_ignores_the_rust_set_aside_folders`
- AC-13 (REQ-23): the two protocol goldens carry the new block text and hash, and nothing else in them changed · verify: `cargo test --locked --offline --test product_cli protocol_fixtures_record_current_contract`
- AC-14 (REQ-23): a file git already tracks under the state folder stays tracked after the block applies · verify: `cargo test --locked --offline --test gitignore_block_cli a_file_git_already_tracks_stays_tracked`
- AC-15 (REQ-8): the end-to-end walk of shape G answers each open leftover with its printed keep command and then reads `all_unchanged` true · verify: `cargo test --locked --offline --test update_end_to_end_cli shape_g_reaches_all_unchanged`
- AC-16 (REQ-7, REQ-22): archiving `.moai-backup/.update.lock` makes `go-state-revert` refuse with nothing written, and after the leftovers revert the go-state revert works · verify: `cargo test --locked --offline --test update_end_to_end_cli shape_e_backup_lock_needs_the_leftovers_revert_first`
- AC-17 (REQ-24): the guide lists an open leftover in `finish.open` · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_open_item_is_listed_by_the_guide`
- AC-18 (REQ-15, REQ-16): with the backup area made unwritable the apply moves no item and reports the failure, and the tree is identical before and after · verify: `cargo test --locked --offline --test project_update_leftovers_cli leftover_apply_moves_nothing_when_the_receipt_cannot_be_saved`

## Scope

- write: `.moai/specs/SPEC-20261127-update-leftovers-action/**`, `src/project_update/leftovers.rs` (new), `src/project_update/leftovers_archive.rs` (new), `src/ops/project_update_leftovers_apply.rs` (new), `src/ops/project_update_leftovers_revert.rs` (new), `src/project_update.rs` (only `leftovers()` and the lock reader with its constants, the `all_unchanged` line, the `"leftovers"` line of `plan()`, a conditional first row of `undo_order`, the two `pub mod` lines), `src/project_update/declined.rs` (the `leftover:<path>` key only), `src/project_update/guide.rs` (`finish.open`), `src/agents.rs` (`render_ignore` and its doc comment only), `src/help.rs` and `tests/fixtures/help/all.txt`, `all-60.txt`, `all-80.txt`, `all-dumb.txt` (the two command rows, by line lease), `tests/project_update_leftovers_cli.rs` (new), `tests/gitignore_block_cli.rs`, `tests/update_end_to_end_cli.rs`, `tests/fixtures/protocol/v1/current/16-agents-plan-create.json`, `tests/fixtures/protocol/v1/current/17-agents-apply-ok.json`, `docs/design-review/appendix/operations/project_update_leftovers_apply.md` (new), `docs/design-review/appendix/operations/project_update_leftovers_revert.md` (new), `docs/design-review/appendix/operations/project_update_plan.md`, `docs/design-review/appendix/operations/project_update_decline.md`, `docs/design-review/appendix/state-layout.md`, `docs/rust-product-path.md`; only if a red run shows a pin flipping: `tests/project_update_plan_cli.rs`, `tests/project_update_go_residue_cli.rs`, `tests/update_moved_receipts.rs`, `tests/fixtures/state_folder_literals/allow.tsv`, `src/ops/project_update_decline.rs` (doc comment only)
- read-only: `src/go_state.rs`, `src/go_residue.rs`, `src/go_residue_apply.rs`, `src/state_root.rs`, `src/baseline.rs`, `src/home_state.rs`, the survey report sections 4-8
- forbidden: any card number in shipped text, any write outside the project folder, real HOME or real project writes in tests, deleting any file (apply and revert only move), changing go-state or go-residue behaviour, `src/main.rs` and `docs/design-review/appendix/interfaces.md` (a one-file operation is collected by `build.rs` and listed there it fails `operations_doc_matches_code`), `skills/**`

## Plan

1. SPEC 승인 뒤 RED: 무시 블록 줄부터 시작해요. `tests/gitignore_block_cli.rs`의 줄 수 단언(`managed block lines: 50 added` → 51, 두 곳)과 새 테스트 `the_block_ignores_the_rust_set_aside_folders`를 먼저 쓰고, 두 골든은 아직 다시 만들지 않아서 `product_cli`가 실패해요. `38 added` 핀은 그대로예요(`with_earlier_block`이 같은 줄을 양쪽에 넣어요).
2. GREEN(무시 블록): `render_ignore`가 `STATE_SECRETS` 반복 뒤에 `.moai.rust-*/`를 덧붙여요(글롭이라 `escape_ignore_path`를 거치지 않아요). 골든 둘은 `MOAI_UPDATE_FIXTURES=1`로 다시 만들고 `detail`과 `rendered_sha256`만 달라졌는지 읽어요. `state-layout.md`에 한 문장. 이 단계만 먼저 따로 착지할 수 있어요(D-1).
3. RED(단위): `leftovers.rs`·`declined.rs`의 `#[cfg(test)]`: 목록 digest가 안정적이고 이름·크기가 바뀌면 달라져요, `leftover:` 키가 `parse`·`encode`를 왕복해요, 알 수 없는 키는 거절돼요, `.moai.rust-<이름>`은 `utc_stamp` 접미사일 때만 유효해요, 영수증 상태 분류.
4. 계획 쪽: `leftovers.rs`의 `view(root, generation, others_pending, record)`가 기존 `leftovers()` 사실에 `action`/`keep_only`/`keep`/`kept`를 붙여요. 공개 규칙은 REQ-3~7과 같아요. `declined.rs`는 `leftover:<알려진 경로>` 키를 받아요(`STEPS`는 그대로). `project_update.rs`는 `all_unchanged`에 `&& open == 0`, `"leftovers"` 줄은 `view`, 그리고 `guide.rs`는 `finish.open`에 열린 항목을 넣어요. 기존 `fact` 문구는 그대로 둬서 기존 단언이 유지돼요. 이어서 `update_end_to_end_cli`의 `assert_walks_to_the_end`가 열린 leftover를 인쇄된 남기기 명령으로 답하게 바꿔요.
5. 적용·되돌리기: `leftovers_archive.rs`에 사전 점검 → pending 영수증 저장 → 읽어 확인 → 이동 → 읽어 확인 → committed. 이동은 `go_state`의 비공개 `move_folder`를 건드리지 않으려고 이 파일에 15줄짜리 `rename_folder`를 두고, 시각 이름은 `go_residue_apply`의 `free_stamp` 방식을 따라 써요(그 파일은 수정하지 않아요). 되돌리기는 영수증이 아니라 디스크 상태로 판단해요. 시험용 고장 지점 `leftovers_cross_device`는 `update::test_fault`에 걸어요.
6. `undo_order` 첫 행(조건부)과 AC-16 테스트.
7. 공유 파일: `help.rs`의 그룹 행·예제 행과 도움말 골든 4개(`MOAI_HELP_GOLDEN_WRITE`), 운영 문서 조각 둘과 `project_update_plan.md`·`project_update_decline.md`·`rust-product-path.md` 수정. 새 명령 둘은 `build.rs`가 `src/ops/*.rs`에서 모으므로 `main.rs`와 `interfaces.md`는 건드리지 않아요.
8. 검증: `cargo fmt --all`, `moai check affected --card` 맨몸으로, 진짜 바이너리로 임시 프로젝트·임시 HOME에서 위험 경로 실행(Rust 통합 테스트), 가드 테스트마다 뮤턴트 한 줄(apply의 digest 비교, 항목의 `contained` 호출, 영수증 상태 검사, apply 때 pid 재확인, 되돌리기의 "원래 자리가 비었는지" 검사, 사전 점검의 장치 비교, 새 무시 줄).

## Risks

- 새 무시 줄이 들어가면 이미 마이그레이션된 프로젝트의 `.gitignore` 관리 블록 해시가 첫 적용 때 달라져요. 계획은 `merge`로 보여 주고 적용 때만 써요(`tests/gitignore_block_cli.rs:555-582`의 선례). 사용자가 고친 블록은 그대로 두어요. 두 버전의 바이너리가 번갈아 돌면 블록이 두 모양 사이를 오가요.
- 골든 둘이 각각 두 줄(`detail`, `rendered_sha256`) 바뀌어요: `16-agents-plan-create.json:47-48`, `17-agents-apply-ok.json:40-41`.
- 단계를 더하지 않아도 열린 leftover가 `all_unchanged`를 거짓으로 만들어 `update_end_to_end_cli`의 walk(`assert_walks_to_the_end`)가 깨져요. AC-15가 그 walk를 남기기 답변으로 맞추고, `undo_order` 핀(`ran == 7`)은 행을 조건부로 두어서 지켜요.
- `.moai-backup/.update.lock`을 옮기면 `go-state-revert`가 그 폴더의 digest 불일치로 `.moai is missing`이라며 거절해요(`src/go_state.rs:5529-5533`, `5769-5778`). 보관 되돌리기를 먼저 하도록 `undo_order`와 항목 효과 문구에 적고(REQ-22), AC-16이 순서를 증명해요.
- `.moai-cli/`를 옮기면 그 안의 백업 복사본을 읽는 병합 단계(`newest_backed_up`, `backed_up_root_file`)가 원본을 잃어요. REQ-4의 병합 원본 가드가 그런 폴더는 `keep_only`로 둬요.
- 옮긴 추적 파일은 `git status`에 삭제로 보이고 checkout으로 되살아나요(REQ-21이 말해 주기만 해요).
- 다른 파일 시스템이면 `rename`과 `hard_link`가 errno 18로 실패해요. 사전 점검이 pending 영수증보다 먼저 막아요(REQ-14).
- `.moai.rust-*/` 글롭은 사용자의 `.moai.rust-notes/` 같은 폴더도 가려요. 글롭이 가리는 것과 보관 후보는 달라요(후보는 `utc_stamp` 접미사만).
- `go_residue`의 `find_manifest`가 `.moai-backup`으로 시작하는 폴더를 모두 읽어서 `.moai-backups/` 안의 `manifest.json`은 보관 뒤에 못 찾게 돼요. 실제 프로젝트에 그런 파일이 있는지는 확인하지 못했어요.
- 같은 파일을 다른 카드도 고쳐요(`src/project_update.rs`, `src/agents.rs`, `src/project_update/declined.rs`, `src/project_update/guide.rs`, `tests/update_end_to_end_cli.rs`, `docs/rust-product-path.md`). 이 SPEC의 조각은 Plan에 적은 자리로 좁게 두고, 나중에 착지하는 쪽이 팁에 한 번 합쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 하나의 SPEC으로 가되, 무시 블록 줄(Plan 1-2)을 첫 커밋으로 두어 필요하면 먼저 따로 착지해요 | 리더가 다른 카드(`agents.rs` 겹침)와의 순서를 고를 수 있어요. 레인 기본값, 리더 확인 필요 | 커밋을 나눠도 SPEC은 그대로예요 |
| D-2 | `STATE_KEPT` 좁히기는 이 SPEC에서 빼요(Non-goals) | 승인된 `SPEC-20261120-moai-layout`의 규칙을 바꾸는 일이고, `KEPT`를 `classify`와 go-state도 읽어요. 사용자 승인 필요: 리더 결정 항목. 카드 기준의 마지막 문장은 AC-14와 `KEPT` 불변으로만 충족돼요 | 후속 SPEC에서 허용 목록과 `!` 줄을 더해요 |
| D-3 | `.moai-cli/`는 마친(committed·reverted) 영수증만 든 폴더를 보관 후보로 두고, 병합 원본 가드를 더해요 | 카드의 "committed 영수증만"을 읽은 것이고, 더 엄격한 안(모두 reverted)은 정상 마이그레이션 뒤에 보관을 거의 안 내요. 레인 기본값 | 규칙을 엄격하게 바꿔도 명령은 그대로예요 |
| D-4 | `.moai-backup/.update.lock`은 보관하되, 보관 되돌리기를 go-state 되돌리기보다 먼저 하도록 `undo_order`에 적어요 | 카드가 두 경로 모두를 요구해요. 대안: 보관 불가(keep만) 또는 go-state 영수증이 남은 동안만 막기. 레인 기본값, 리더 확인 필요 | 항목을 `keep_only`로 바꿔요 |
| D-5 | 계획은 배열에 항목별 조치를 붙이는 모양이에요(새 단계 없음) | 단계 목록·`ORDER` 핀을 다른 카드와 부딪치지 않아요 | 나중에 단계로 올려도 항목은 그대로예요 |
| D-6 | 남기기는 `update-declined.json`의 `leftover:<경로>` 키와 기존 `project-update-decline`을 써요 | 새 명령·새 파일 없이 같은 출처 sha 규칙을 재사용해요 | 키만 지우면 돼요 |
| D-7 | `all_unchanged`에 반영하는 것은 다섯 가지뿐이고 `CLAUDE.md`는 세지 않아요 | 카드 기준과 claude-md 단계의 분리 | 한 줄 조건이에요 |
| D-8 | apply는 반복 가능한 `--item <경로>`를 한 plan digest 아래 받아요 | 항목별 digest보다 명령 수가 적어요 | 항목별 digest로 바꿔도 영수증은 그대로예요 |
| D-9 | 이름 붙은 `.moai-backup-<UTC>/`의 lock은 후속으로 미뤄요 | 카드가 말한 "두 경로"는 `.moai/`와 `.moai-backup/`이에요 | 후보 목록에 이름 규칙을 더해요 |
| D-10 | 영수증은 `.moai/backups/<UTC>/leftovers-receipt.json`, 옮긴 것은 `.moai/backups/<UTC>/leftovers/<같은 경로>`예요 | `go-state-receipt.json`이 아니라서 분류기와 go-state 되돌리기가 읽지 않아요(go-residue와 같은 이유) | 이름만 바꾸면 돼요 |

## Result

`status: draft`예요. 승인은 이 파일에 기록하지 않았어요(리더가 카드 노트에 인용한 사용자 승인이 레저에 있어요). 구현 head 는 이 Result 를 담은 문서 커밋 바로 앞이에요. 작업 범위는 `fa2b49dc44..HEAD`예요.

한 일: 무시 블록 끝에 `.moai.rust-*/` 한 줄(`render_ignore`)과 골든 16·17, 계획 쪽(`src/project_update/leftovers.rs` 새 파일: 다섯 종류의 보관 가능 판정, 출처 sha256, 보관 명령과 남기기 명령, `declined.rs`의 `leftover:<경로>` 키, `all_unchanged`, `guide.rs`의 `finish.open`), 적용·되돌리기(`leftovers_archive.rs` 새 파일, `project-update-leftovers-apply`·`-revert` 새 명령 둘, 영수증 `.moai/backups/<UTC>/leftovers-receipt.json`, 보관 `leftovers/<같은 경로>`), `undo_order` 의 조건부 첫 행, 도움말 행과 골든 넷, 운영 문서와 `docs/later.md` 두 줄이에요. 이동은 한 번의 rename(폴더) 또는 link 뒤 unlink(파일)이고 아무것도 지우지 않아요. 되돌리기는 영수증이 아니라 디스크로 판단해요.

결정과 해석:

- D-2: `STATE_KEPT` 좁히기는 하지 않았고 후속 카드는 리더가 열어요. 카드 기준의 "좁히기" 문장은 met 으로 표시하지 않아요(AC-14 와 `KEPT` 불변으로만 충족돼요). 랜딩은 `--keep-open`이에요.
- D-4: `.moai-backup/.update.lock` 은 보관하고, 보관 되돌리기를 `go-state-revert` 보다 먼저 하도록 `undo_order` 첫 행과 AC-16 테스트로 고정했어요.
- 되돌리기 명령은 `--receipt` 에 더해 `--permission-reference` 도 받아요(`go-residue-revert` 와 같아요). 본문 REQ-19 는 최소 형태만 적었어요.
- REQ-1 의 키는 `action`·`keep_only`·`keep`·`kept`·`ask_again` 이고, 다른 백업 단계가 명령을 가진 동안 보관 가능한 항목은 `action` 도 `keep_only` 도 없이 `action_waits` 문장만 가져요(REQ-2).
- 폴더 이름은 프로젝트의 글이라서 `.moai.rust-` 뒤가 ASCII 영숫자와 `-_.`(64자 이하)인 폴더만 목록에 올려요(보안 검토 지적을 반영했어요).

다른 카드 테스트를 바꾼 곳(리뷰어가 약화 여부를 볼 곳): `tests/update_end_to_end_cli.rs`의 `shape_e_reaches_all_unchanged` 는 이제 `.moai-backup/.update.lock` 도 목록에 있어서 기대 목록이 둘이에요. `the_printed_undo_order_takes_the_chain_back_to_the_start_bytes` 의 `may_stay_behind` 는 정확히 `.moai-cli/manifests/update-declined.json` 한 파일만 더 허용해요(walk 가 열린 `.moai-cli/` 를 남기기 명령으로 답해서 생겨요). `ran == 7` 핀과 다른 단언은 그대로예요. `tests/fixtures/state_folder_literals/allow.tsv` 는 `tests/gitignore_block_cli.rs` 의 기준선을 74 에서 81 로 올렸어요(새 테스트가 `.moai.rust-…` 이름을 써요). `tests/help_text_names_moai.rs` 의 `the_help_goldens_keep_the_old_name_only_for_folders_and_the_alias_file` 는 규칙은 그대로고(옛 이름은 점 바로 뒤와 별칭 문장 `Preview the moai link beside moai-cli` 의 끝에서만 허용), 그 문장이 줄바꿈되는 자리만 일반화했어요: 새 명령 이름이 길어져서 60열 도움말의 열 폭이 바뀌고 줄바꿈이 `beside` 앞으로 옮겨지자 옛 코드는 `…link beside` 뒤에서만 끊긴 줄을 알아봐서 실패했어요. 새 코드는 마지막 네 줄을 이어 붙인 글이 `Preview the moai link beside moai-cli` 로 끝날 때만 허용해요(다른 자리의 `moai-cli` 는 여전히 걸려요).

Verification (이 세션과 에이전트 실행에서 본 결과; 같은 헤드의 체크 결과는 READY 노트가 인용해요):

- 무시 블록: RED → `gitignore_block_cli` 48개 중 3개 실패(단언), GREEN 뒤 48개 통과, `product_cli protocol_fixtures_record_current_contract` 1개 통과(골든은 `MOAI_UPDATE_FIXTURES=1` 로 다시 만들었고 바뀐 필드는 `detail`·`rendered_sha256`·`manifest_sha256` 뿐이에요).
- 계획 쪽: RED 새 테스트 4개와 끝에서 끝 2개가 단언으로 실패, GREEN 뒤 통과.
- apply/revert: RED 9개와 끝에서 끝 1개가 단언으로 실패, GREEN 뒤 `project_update_leftovers_cli` 15개, `update_end_to_end_cli` 10개, `project_update_plan_cli` 74개, `project_update_go_residue_cli` 17개, `update_moved_receipts` 8개, `--lib project_update::` 통과.
- 도움말: `help_layout` 11개 통과. 네 골든은 새 행 두 줄 단어만 늘고 나머지는 열 폭과 줄바꿈이에요(60·80 열은 단어 단위 비교).
- clippy: 라이브러리와 바이너리, 그리고 세 테스트 대상 `-D warnings` 모두 exit 0.
- 위험 경로(이동·되돌리기·위조 영수증·권한 0 파일)는 임시 프로젝트와 임시 HOME 의 진짜 바이너리로 도는 Rust 통합 테스트로만 실행했어요(AC-3, 4, 5, 6, 8, 10, 18 과 되돌리기 충돌 테스트 둘).

Guard mutants(각 뮤턴트는 슬롯에서 정확히 한 번씩 Edit 으로 넣고 지정 테스트를 돌린 뒤 되돌렸어요. 처음 실행에서 3개가 `exit 0` 으로 살아남아서 테스트를 보강한 뒤 같은 뮤턴트가 `exit 100` 이 됐어요):

- `guard-mutant: leftover_apply_refuses_stale_digest_bad_reference_and_unknown_item | apply 의 plan digest 비교 | exit 100`
- `guard-mutant: an_item_behind_a_link_is_refused_by_the_look_before_a_move | still_as_planned 의 contained 검사 | exit 100`
- `guard-mutant: apply_names_a_pick_behind_a_link_as_such | apply 의 pick contained 검사 | exit 100`
- `guard-mutant: an_archive_place_behind_a_link_is_refused_before_the_receipt_is_saved | preflight 의 보관 자리 contained 검사 | exit 100`
- `guard-mutant: leftover_apply_never_moves_a_link_or_an_escaping_path | revert 의 영수증 경로 가드 | exit 100`
- `guard-mutant: leftover_folder_with_an_unfinished_receipt_is_keep_only | 계획의 pending 영수증 검사 | exit 100`
- `guard-mutant: a_lock_whose_process_is_alive_when_the_apply_starts_is_refused | apply 때 pid 재확인 | exit 100`
- `guard-mutant: leftover_revert_leaves_a_taken_place_and_the_archive_as_they_are | revert 의 원래 자리 검사 | exit 100`
- `guard-mutant: leftover_revert_refuses_an_archive_that_differs_from_the_receipt | revert 의 보관본 listing digest 비교 | exit 100`
- `guard-mutant: leftover_apply_stops_before_any_move_on_another_filesystem | preflight 의 leftovers_cross_device 고장 지점 | exit 100`
- `guard-mutant: same_filesystem_compares_the_device_of_the_item_and_of_the_backup_area | 장치 id 비교 | exit 100`
- `guard-mutant: the_block_ignores_the_rust_set_aside_folders | render_ignore 의 새 마지막 줄 | exit 100`
- `guard-mutant: a_set_aside_folder_is_listed_only_under_a_plain_name | plain_suffix 필터 | exit 100`

Acceptance:

- AC-1: met · `leftover_plan_gives_each_item_an_action_or_a_keep_only_reason` 통과
- AC-2: met · `leftover_folder_with_an_unfinished_receipt_is_keep_only` 통과
- AC-3: met · `leftover_apply_moves_and_revert_puts_back_every_byte` 통과
- AC-4: met · `leftover_apply_refuses_stale_digest_bad_reference_and_unknown_item` 통과
- AC-5: met · `leftover_apply_never_moves_a_link_or_an_escaping_path` 통과
- AC-6: met · `leftover_lock_is_archived_only_while_its_process_is_dead` 통과(apply 때 pid 재확인은 단위 테스트가 따로 덮어요)
- AC-7: met · `leftover_keep_is_recorded_counted_and_asked_again_when_the_folder_changes` 통과
- AC-8: met · `leftover_apply_stops_before_any_move_on_another_filesystem` 통과(다른 파일 시스템은 시험용 고장 지점으로만 증명했어요: simulated)
- AC-9: met · `leftover_apply_says_how_many_tracked_files_moved` 통과
- AC-10: met · `leftover_apply_never_opens_a_secret_shaped_name` 통과
- AC-11: met · `leftover_undo_order_names_the_revert_only_after_an_apply` 통과
- AC-12: met · `the_block_ignores_the_rust_set_aside_folders` 통과
- AC-13: met · `protocol_fixtures_record_current_contract` 통과(1 passed)
- AC-14: met · `a_file_git_already_tracks_stays_tracked` 가 `gitignore_block_cli` 48개 안에서 통과
- AC-15: met · `shape_g_reaches_all_unchanged` 통과
- AC-16: met · `shape_e_backup_lock_needs_the_leftovers_revert_first` 통과
- AC-17: met · `leftover_open_item_is_listed_by_the_guide` 통과
- AC-18: met · `leftover_apply_moves_nothing_when_the_receipt_cannot_be_saved` 통과

Not run / 남은 것: Windows·Linux 결과는 주장하지 않아요(Mac 픽스처만). 실제 다른 파일 시스템 이동은 재현하지 못했고 시험용 고장 지점과 `same_filesystem` 단위 테스트로만 덮었어요. 계획 문장의 `.moai-cli/` fact 는 맨 위 항목 이름을 비밀 이름 거르기 없이 최대 8개 보여 줘요(`docs/later.md` 한 줄). go-residue 조사가 같은 폴더를 `leftover_folder` 로 한 번 더 알리는 중복도 `docs/later.md` 에 한 줄로 남겼어요. 이 Result 커밋 뒤 헤드의 `check affected --card` 결과는 READY 노트가 인용해요(여기에 적으면 헤드가 바뀌어요).
