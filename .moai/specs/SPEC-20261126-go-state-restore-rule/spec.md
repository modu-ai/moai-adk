---
id: SPEC-20261126-go-state-restore-rule
title: "go-state apply restores the files the Go manifest does not record and leaves the byte-proven Go templates in the backup"
status: approved
revision: 3
created: 2026-11-26
approval:
  by: leader
  at: 2026-11-26
  basis: "LEADER-APPROVAL 노트 10466 (개정 1 승인, 리더 판단이고 사용자 직접 선택은 아님, 사용자 목표 2026-11-26 에 따름, D-1·D-2·D-4 기본값) 과 노트 10519 (AC-4 문구 정정, 결정 10516 A, 개정 2) 와 노트 10615 (업데이트 락 제외, 개정 3, 리더 판단이고 사용자 직접 선택은 아님); 리더가 노트에 적었고 레인은 그 노트를 인용만 해요"
  digest: "sha256:02a7948b9b8a8c3a2f32cf5fe36dbb1695f5c6cdca96411a5477fa036268f646"
---

# go-state apply restores the files the Go manifest does not record and leaves the byte-proven Go templates in the backup

## Intent

이 SPEC 의 카드(조각 S7)는 조사 보고서 `go-generation-update-2026-11-21.md` 4절 첫 줄과 5절 G5 가 짚은 빈틈이에요. `moai go-state-apply` 는 Go 세대의 `.moai/` 를 통째로 `.moai-backup` 으로 옮기고, 사용자 자료를 담는 열 개 폴더(`specs`, `project`, `plans`, `reports`, `docs`, `research`, `memory`, `proposals`, `todo`, `release-notes`)만 새 `.moai/` 로 되돌려 복사해요(`src/go_state.rs` 의 `USER_DATA_FROM` 과 `plan_restores`). 그 밖의 폴더에 있던 사용자 파일(mo.ai.kr 의 `.moai/design/` 478개, `decisions/` 사용자 파일, `school/`, `brain/` 등)은 백업에만 남아요. 그런데 git 이 추적하던 파일은 되살아나서, 옮겼다고 믿은 파일이 다시 보이고 나머지는 사라진 것처럼 보여요.

사용자 지시(2026-11-21, 카드 본문)와 보고서의 규칙은 이래요: 옮기기만 하고 지우지 않아요, 바이트로 증명된 Go 파일만 후보예요, 비밀 이름은 열지 않고 개수로만 보여 줘요. 이 SPEC 은 그 규칙을 복원에 적용해요. 열 개 폴더 밖에서는 Go 매니페스트가 기록하지 않은 파일이 사용자 자료라서 복원하고, 매니페스트가 기록한 파일은 Go 도구가 만든 것이라서 백업에 남겨요.

## Interview

- Q1: 사용자 자료는 무엇으로 알아보나요? → A: Go 매니페스트(`.moai/manifest.json` 의 `files`)가 `.moai/<경로>` 키로 기록하지 않은 파일이에요(카드 본문, 보고서 5절 G5: "Go 매니페스트에 없는 파일 = 사용자 데이터 → 전부 복원 (비밀 이름·링크 제외)").
- Q2: 매니페스트가 기록한 파일은 바이트가 같을 때만 남기나요, 고친 파일도 남기나요? → A: 둘 다 백업에 남겨요. 지금 동작과 같아요. 고친 Go 파일을 하나씩 묻는 일은 `go-residue` 단계(S1·S2)의 몫이에요(보고서 7.2).
- Q3: 열 개 폴더 안의 파일도 매니페스트를 보나요? → A: 아니요. 열 개 폴더는 지금처럼 전부 복원해요. 이 SPEC 은 줄이지 않고 넓히기만 해요(D-1).
- Q4: 되살아난 Go 표식이 생기지 않나요? → A: 표식 세 개(`.moai/manifest.json`, `.moai/config/config.yaml`, `.moai/config/sections/system.yaml`)는 복원하지 않아요. 복원하면 새 `.moai/` 가 다시 Go 로 분류돼요(`src/state_root.rs`).

## Non-goals

- 이미 `.moai-backup` 으로 옮긴 프로젝트(표식을 지운 뒤 남은 Go 파일)의 정리: `go-residue` 단계의 일이에요.
- 고친 Go 템플릿 파일을 되살리거나 묻는 일, 되살아난 표식의 보관: S1·S2·S3 의 일이에요.
- 복원 파일의 git 마무리 커밋 제안: S6 과 PROJECT 8단계의 일이에요.
- 영수증 크기 한도와 영수증 저장 순서 변경: `SPEC-20261120-go-state-large-receipt` 가 정한 그대로예요.
- 열 개 폴더의 복원 규칙 변경.

## Requirements

- REQ-1: WHEN `go-state-plan` plans the move of the Go `.moai/` folder, it shall list as a restore row every regular file under `.moai/` that lies outside the ten user-data folders and whose `.moai/<path>` key the Go manifest does not record. · failure: none — a preview only
- REQ-2: The restore list shall hold no file the Go manifest records outside the ten user-data folders, whether its bytes equal the recorded hash or not, so the byte-proven Go templates stay in the Go backup. · failure: none — nothing is deleted, the file stays in `.moai-backup`
- REQ-3: The restore list shall hold none of the three Go markers (`.moai/manifest.json`, `.moai/config/config.yaml`, `.moai/config/sections/system.yaml`), not the Go tool's update lock (`.moai/.update.lock`), and nothing under `.moai/backups/`.
- REQ-4: A restore row of REQ-1 shall be judged like a row of the ten user-data folders: a secret name is never opened and gets `excluded_secret`, a link or special file gets `leave_link`, a file over the file limit gets `excluded_unclassifiable`, a config-shaped file is read by the credential filter fail-closed, a file whose target exists is `conflict` and is never overwritten.
- REQ-5: The files of the ten user-data folders shall be listed and judged exactly as before this SPEC.
- REQ-6: The apply shall copy each restore row of REQ-1 whose action is `copy` from the Go backup into the new `.moai/` and record it in the receipt.
- REQ-7: The receipt check shall accept a restored path under any folder of `.moai/` except `.moai/backups/` and the three markers, by one rule on the write side and on the read side.
- REQ-8: WHEN the apply has finished, a new `go-state-plan` shall list no restore row that still has to be copied. · failure: none — a preview only
- REQ-9: The revert shall put the Go tree back byte for byte, with the restored copies going aside with the new `.moai/`.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary project shaped like mo.ai.kr, with a Go manifest that records `.moai/policies/core/a.md` with its real hash, `.moai/design/x.md` unrecorded, and an unrecorded `.moai/decisions/user.md`, the preview lists `.moai/design/x.md` and `.moai/decisions/user.md` as `copy` and no row for `.moai/policies/core/a.md`; a recorded file whose bytes differ from the recorded hash gets no row either. · verify: `cargo nextest run --offline --locked --test go_state_cli -E 'test(restore_rule_)'`
- AC-2 (REQ-3): an unrecorded `.moai/config/config.yaml`, an unrecorded `.moai/.update.lock` and anything under `.moai/backups/` get no restore row. · verify: `cargo nextest run --offline --locked --test go_state_cli -E 'test(restore_rule_)'`
- AC-3 (REQ-4): an unrecorded secret-named file is listed as `excluded_secret` with an empty hash and was not opened, a link gets `leave_link`, and an existing target gets `conflict`. · verify: `cargo nextest run --offline --locked --test go_state_cli -E 'test(restore_rule_)'`
- AC-4 (REQ-5): the existing restore tests of the ten folders pass; only the three assertions this SPEC changes are edited: a Go file the manifest does not record is restored in `decisions/` and `state/` (`go_state_classes_cli`), and a new such file after the preview makes the digest stale (`go_state_digest_cli`). · verify: `cargo nextest run --offline --locked --test go_state_cli --test go_state_classes_cli --test go_state_digest_cli --test go_state_preview_cli --test go_state_blocked_folders_cli`
- AC-5 (REQ-6, REQ-8, REQ-9): the apply copies `.moai/design/x.md` into the new `.moai/design/x.md` with the same bytes, the receipt lists it, a new preview lists no row to copy, the revert gives the Go tree back with the same bytes, and a receipt that names a restored path under `.moai/backups/` or a marker is refused. · verify: `cargo nextest run --offline --locked --test go_state_cli -E 'test(restore_rule_)'`
- AC-6 (REQ-7): the receipt check unit tests accept a restored path under a folder outside the ten and refuse one under `.moai/backups/` or a marker. · verify: `cargo nextest run --offline --locked --lib -E 'test(restore_rule_)'`
- AC-7 (REQ-3): the end-to-end update test that expects the dead Go update lock to stay in the Go backup passes once the lock is never restored. · verify: `cargo nextest run --offline --locked --test update_end_to_end_cli -E 'test(shape_e_reaches_all_unchanged)'`

## Scope

- write: `.moai/specs/SPEC-20261126-go-state-restore-rule/**`, `src/go_state.rs` (the function `plan_restores`, its helpers and the `allowed` check of the receipt validation), `tests/go_state_cli.rs`, `docs/rust-product-path.md` (the `restores` words of the `go_state.plan` row)
- read-only: `src/state_root.rs`, `src/migrate.rs`, `src/apply.rs`, `tests/go_state_classes_cli.rs`, `tests/go_state_digest_cli.rs`
- forbidden: `src/project_update.rs`, `src/go_residue*.rs` (other cards hold them), the real HOME and any real project (tests use temporary projects and fake data), deleting a file, a change to the receipt format or its size limit, `CLAUDE.md`

## Plan

1. RED: the `restore_rule_` tests of AC-1~AC-3 and AC-5 in `tests/go_state_cli.rs`, the receipt check tests of AC-6 in the lib; record the failing run.
2. GREEN: `plan_restores` walks the rest of the Go `.moai/` after the ten folders, skips the markers, `backups/` and the recorded keys, and judges each file with the code the ten folders use; the receipt check allows the wider folder set by one rule used on both sides.
3. Dangerous path: the apply and the revert on a temporary project (real binary, temporary HOME, fake files), including a conflict at a target.
4. Docs: the `restores` words of the `go_state.plan` row.
5. READY: `moai check affected --card`, a guard mutant for each new guard test.

## Risks

- A Go `.moai/` that holds many unrecorded files (logs, caches) restores them all; the walk is already bounded by `unrecorded_limit()` through the nested-checkout walk, a folder that passes it stays where it is, and each file is bounded by the file limit. The receipt size is the large-receipt SPEC's.
- A restored file that lands where the earlier-root product state also lands is a `conflict`, and a conflict stops the whole apply with nothing written, as it does for `specs/` today. The user resolves it; nothing is overwritten.
- A file the apply itself writes under the new `.moai/` (the backup area under `backups/`) cannot collide: `backups/` is excluded.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The ten user-data folders keep restoring every file, the manifest rule applies only outside them. | Widening only: no existing restore shrinks, so no earlier apply or test changes meaning. | a later SPEC |
| D-2 | A recorded file stays in the backup whether its bytes match the recorded hash or not. | Q2: a changed Go file is asked one by one by the residue step, never decided here; nothing is deleted. | a later SPEC |
| D-3 | The three Go markers, the Go update lock `.update.lock` and `backups/` are never restored. | A restored marker makes `state_root` classify the new `.moai/` as `mixed`; `backups/` holds this apply's own area; the update lock is the Go tool's process lock that the update reports as a leftover fact and expects to stay in the Go backup (found by the batch check: a restored copy stayed as a leftover). | none needed |
| D-4 | A conflict at a restore target stops the apply with nothing written. | The existing rule for the ten folders. | a later SPEC |

## Result

- AC-1: met — `cargo nextest run --offline --locked --test go_state_cli -E 'test(restore_rule_)'` passed (restore_rule_lists_unrecorded_files_and_leaves_the_recorded_templates); a copy row for `design/x.md`, `decisions/user.md` and a SPEC, no row for the recorded `policies/core/a.md` or the recorded edited `config/sections/b.yaml`.
- AC-2: met — same command, restore_rule_skips_the_markers_and_the_backups_folder: no row for the three markers or anything under `backups/`.
- AC-3: met — same command, restore_rule_judges_each_file_like_the_ten_folders: secret name → `excluded_secret` with an empty hash, credential-looking `cfg.yaml` → `excluded_secret`, link → `leave_link`, existing target → `conflict`.
- AC-4: met — `cargo nextest run --offline --locked --no-fail-fast --test go_state_cli --test go_state_classes_cli --test go_state_digest_cli --test go_state_preview_cli --test go_state_blocked_folders_cli`: 225 run, 225 passed, exit 0 (log target/lamech/ac4-b.log). Only the three assertions this SPEC changes were edited (leader decision A, note 10516); one product-path phrase test needed its sentence kept in the `go_state.plan` row.
- AC-5: met — go_state_cli restore_rule_apply_copies_the_file_and_the_revert_gives_the_go_tree_back (apply copies `design/x.md`, the receipt lists it, a new preview has no restore row, the revert returns the Go tree) and restore_rule_a_forged_restored_path_is_refused (receipt with `.moai/backups/…` or a marker path is refused).
- AC-6: met — `cargo nextest run --offline --locked --lib -E 'test(restore_rule_)'` passed (restore_rule_allows_any_folder_but_not_backups_or_a_marker, restore_rule_keeps_the_earlier_root_to_its_specs).
