---
id: SPEC-20261125-go-residue-apply
title: "Go-residue apply moves the byte-proven Go files into a receipted backup and revert puts them back, refusing every path outside the project"
status: draft
revision: 1
created: 2026-11-25
---

# Go-residue apply moves the byte-proven Go files into a receipted backup and revert puts them back, refusing every path outside the project

## Intent

슬라이스 S1(`SPEC-20261121-go-residue-detect`)은 Go 세대가 남긴 파일을 읽기만 하는 `go-residue` 단계로 찾아서 분류해요. 이 SPEC은 그 다음 절반(S2)이에요. 사용자가 고른 묶음만 `.moai/backups/<UTC>/go-residue/` 아래로 옮기고(지우지 않아요), 영수증을 남기고, 영수증으로 바이트까지 똑같이 되돌려요. 조사 보고서 4-8절의 규칙을 따라요: 미리보기 → 고르기 → 적용 → 영수증 → 되돌리기, 멱등, 바이트로 증명된 파일만 후보.

이 SPEC은 삭제와 되돌리기를 다루는 위험 경로(tier 2)예요. S1 감사가 남긴 비차단 지적을 입력으로 받아요: 매니페스트가 이름 붙인 경로를 `project.join(rel)`로 그대로 읽으면, 절대 경로(`join`이 기준 경로를 통째로 바꿔요), `..`, 중간 링크가 프로젝트 밖을 가리켜요. 쓰기와 지우기 전에 그런 경로는 모두 거절해야 해요.

## Non-goals

- `go_marker`(되살아난 표식)는 옮기지 않아요: 이 슬라이스가 아니라 S3의 일이에요(조사 보고서 8절 조각 표).
- `leftover_folder`(`.moai-cli/`, `.moai.rust-<UTC>/`, `.moai-backups/`)는 옮기지 않아요: S5의 일이에요. `.moai-cli/`를 옮기면 그 안의 영수증을 읽는 go-state 되돌리기가 깨져요.
- `user_data`(어떤 매니페스트도 이름 붙이지 않은 파일)와 비밀 모양 이름은 절대 옮기지 않아요. 보여 주기만 해요.
- 마무리 커밋 제안과 `update.md` 글은 S6의 일이에요. 이 SPEC은 git 색인을 건드리지 않아요: 옮긴 추적 파일은 `git status`에 삭제로 보이고, 결과가 "커밋하지 않으면 checkout 때 되살아나요"라고 말해요.
- `claude-md` 단계(S4), 안 쓰는 호스트(S9), 추적 중인 보고서는 건드리지 않아요.
- go-state의 동작과 영수증 형식은 바꾸지 않아요. 새 영수증은 `go-residue-receipt.json`으로 따로 두어서 `state_root` 분류기와 go-state 되돌리기가 읽는 `go-state-receipt.json`에 영향이 없어요.

## Requirements

- REQ-1: When the user applies a previewed go-residue plan, `go-residue-apply` shall move only the files of the chosen bundles, the proven `go_same` bundle and each explicitly picked `go_changed` path, into `.moai/backups/<UTC>/go-residue/<same path>` and delete nothing.
- REQ-2 (REQ-1): IF the plan digest differs from a fresh plan's digest or the permission reference is rejected, THEN `go-residue-apply` shall refuse with a stale-plan or permission answer and write nothing.
- REQ-3: The one path guard shall refuse a path that is absolute, holds a `..` component or passes through a link before the survey, the apply or the revert reads, writes or deletes anything at it.
- REQ-4: When the apply starts, `go-residue-apply` shall save a pending receipt `.moai/backups/<UTC>/go-residue-receipt.json` under the one size limit its reader uses and read it back before it moves any file.
- REQ-5 (REQ-4): IF the pending receipt cannot be saved, read back or kept under that limit, THEN `go-residue-apply` shall move no file and report the failure.
- REQ-6: When the last file has moved, `go-residue-apply` shall read every moved file back from the backup and mark the receipt `committed` only when each sha256 equals its receipt entry.
- REQ-7 (REQ-6): IF a move or a read-back fails, THEN `go-residue-apply` shall stop at that file, keep the receipt that lists every planned move and leave the files already moved revertible.
- REQ-8: When the user reverts a receipt, `go-residue-revert` shall put a backup file back at its original path only when that place is empty and the backup file's sha256 equals the receipt entry's.
- REQ-9 (REQ-8): IF the original place holds a file or the backup file differs from the receipt, THEN `go-residue-revert` shall leave both untouched and report that item as `conflict`.
- REQ-10: The go-residue step shall read `needs_decision` only while a `go_same` or `go_changed` file is left, whatever other classes it still lists.
- REQ-11: `go-residue-apply` shall move no file of the classes `user_data`, `go_marker` or `leftover_folder` and no secret-shaped name.
- REQ-12: When a moved file is tracked by git, `go-residue-apply` shall report how many tracked files moved and that a checkout brings them back unless the deletions are committed. · failure: none — the report is text on an apply that already finished and leaves nothing to retry

## Acceptance

- AC-1 (REQ-1, REQ-10, REQ-11): on the survey's temp frame, applying the proven bundle moves `a.md`, `gtd.md` and `moai-easy.md` into the backup folder and leaves `user.yaml`, `x.md` and every secret-shaped file byte for byte where they were; a second preview lists only the changed file and asks about it; applying it with its explicit pick makes the next preview read `unchanged` · verify: `cargo test --test project_update_go_residue_cli`
- AC-2 (REQ-2): a plan digest from before a file changed, a made-up digest and a malformed permission reference each leave the whole tree identical before and after the apply · verify: `cargo test --test project_update_go_residue_cli`
- AC-3 (REQ-3): a manifest that names `../outside/x.md`, an absolute path to a file outside the project, a path through a linked folder and a linked file never makes the survey or the apply read, move or write that outside file (its sha256 is the same before and after), and a forged receipt that names a path outside the project is refused by the revert before it moves anything · verify: `cargo test --test project_update_go_residue_cli`
- AC-4 (REQ-4, REQ-5): a project with enough proven files for a receipt above 256 KiB applies to `committed` and reverts, and with the backup area made unwritable the apply moves no file (the tree is identical before and after) · verify: `cargo test --test project_update_go_residue_cli`
- AC-5 (REQ-6, REQ-7): a destination that already exists makes the apply stop at that file with the receipt still listing every planned move, and the revert then restores the files that did move · verify: `cargo test --lib go_residue_apply`
- AC-6 (REQ-8, REQ-9): after a committed apply the revert leaves every file byte for byte as it was before the apply; a new file at an original path, or a changed backup copy, makes only that item `conflict` and touches neither side · verify: `cargo test --test project_update_go_residue_cli`
- AC-7 (REQ-12): on a git project where the moved files are tracked, the apply result names the tracked count and the checkout warning · verify: `cargo test --test project_update_go_residue_cli`

## Scope

- write: `.moai/specs/SPEC-20261125-go-residue-apply/**`, `src/go_residue.rs`, `src/go_residue_apply.rs` (new), `src/go_state.rs` (visibility of the helpers the apply reuses, no behaviour change), `src/project_update.rs` (the step's decision rule and apply command), `src/main.rs` (the two command variants and their dispatch lines), `src/lib.rs` (the module line, by line lease), `src/help.rs` and `tests/fixtures/help/*.txt` (the command rows, by line lease), `tests/project_update_go_residue_cli.rs`, the step-list pins of `tests/project_update_plan_cli.rs`, `tests/go_settings_cli.rs` and `tests/reports_move_cli.rs` only if the new commands change one
- read-only: `src/state_root.rs`, `src/apply.rs`, the survey report sections 4-8
- forbidden: any card number in shipped text, any write outside the project folder, real HOME or real project writes in tests, deleting any file (apply and revert only move), changing go-state behaviour

## Plan

1. RED: 같은 테스트 파일에 AC마다 실패하는 테스트를 먼저 써요. 프레임은 S1 테스트 파일의 임시 프로젝트(`TempHome` + git init)를 그대로 써요.
2. 하나의 포함 검사 함수를 `src/go_residue.rs`에 두고 세 곳(조사, 적용, 되돌리기)이 같이 써요. go-state의 `plain_relative`와 `linked_component`를 재사용해요(지금 비공개라서 `pub(crate)`로만 바꿔요, 동작은 그대로).
3. `src/go_residue_apply.rs`: 계획(묶음 선택과 다이제스트), 적용(영수증 pending 저장 → 읽어 확인 → 파일마다 한 번 더 해시 확인 후 이동 → 읽어 확인 → committed), 되돌리기(영수증을 믿지 않고 경로·해시를 다시 확인). 이동은 go-state의 `move_exclusive`(하드 링크 후 제거, 덮어쓰지 않음)와 `still_as_previewed`를 써요.
4. 영수증 크기 제한은 쓰는 쪽과 읽는 쪽이 같아요(`state_root::GO_STATE_RECEIPT_MAX`). 영수증은 맨 처음에 계획된 이동을 모두 `moved: false`로 적어 두고, 되돌리기는 `moved` 값이 아니라 디스크의 실제 상태(원래 자리가 비었는지, 백업 해시가 같은지)로 판단해요.
5. `src/project_update.rs`의 단계가 `go_same`/`go_changed`가 남았을 때만 `needs_decision`이 되고, 그때 적용 명령을 보여 줘요.
6. 명령 이름과 도움말을 더해요(공유 파일은 줄 임대로).
7. 위험 경로 실행: 빌드한 바이너리를 격리 폴더의 임시 프로젝트와 임시 HOME에서 돌려, 심은 링크와 프로젝트 밖 경로를 실제로 거절하는 영수증을 READY에 넣어요. 가드 테스트마다 뮤턴트 한 줄(최대 5개)을 만들어요.

## Risks

- 계획과 적용 사이에 파일이 바뀌면 영수증이 백업에 없는 바이트를 가리킬 수 있어요: 파일마다 이동 직전에 해시를 다시 확인하고, 다르면 거기서 멈춰요(REQ-7).
- 영수증이 수백 KB를 넘으면 256 KiB 한도를 쓰는 다른 읽기가 실패할 수 있어요: 쓰는 쪽과 읽는 쪽이 같은 한도를 쓰고(REQ-4), AC-4가 256 KiB 넘는 프레임으로 확인해요.
- 되돌리기가 영수증을 믿으면 위조된 영수증이 프로젝트 밖 파일을 움직일 수 있어요: 영수증의 모든 경로를 같은 포함 검사에 다시 통과시켜요(REQ-3, AC-3).
- 옮긴 추적 파일이 checkout으로 되살아나요: 이 SPEC은 말해 주기만 하고(REQ-12), 마무리 커밋은 S6이에요.
- 프로젝트와 백업이 다른 파일 시스템이면 하드 링크 이동이 거절돼요: 기존 go-state와 같은 거절이고, 파일을 하나도 지우지 않은 채 멈춰요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | go_changed는 사용자가 경로를 골랐을 때만 옮겨요 | 조사 보고서 7.1: 바이트로 증명된 것만 자동 후보, 다르면 파일마다 물어요 | 고르지 않으면 그대로예요 |
| D-2 | 별도 plan 명령 없이 project-update-plan의 go-residue 단계가 계획이에요 | 리더 디스패치와 계획 노트의 기본값, 명령 수를 줄여요 | 나중에 plan 명령을 더해도 단계는 그대로예요 |
| D-3 | 되돌리기는 `--receipt <UTC>`를 받고 영수증 상태가 아니라 디스크의 실제 상태로 판단해요 | 조사 보고서 7.2의 6단계, 중간에 멈춘 적용도 되돌려야 해요 | 영수증 파일은 그대로 남아요 |
| D-4 | 새 영수증은 `go-residue-receipt.json`이에요 | `state_root`가 `go-state-receipt.json`만 읽어서 이 영수증이 분류기와 go-state 되돌리기에 닿지 않아요 | 이름만 바꾸면 돼요 |

## Result

`status: draft`예요. 승인은 이 파일에 기록하지 않았어요(리더의 위임 승인은 카드 노트에 있어요). 구현 head는 `12f50d2d981e79bb556f88331513be66844729b0`이고, 그 뒤의 문서 커밋 하나가 이 Result예요. base는 NeXT의 `e946ae8102`에 S1 수리 head `1aaccb74e1`을 한 번 합친 `70aa4a907d`예요.

한 일: `src/go_residue_apply.rs`(새 파일)에 후보 선별, 계획 digest, 적용, 되돌리기, 영수증을 넣었어요. 하나뿐인 경로 검사 `go_residue::contained`와 `secret_path`를 조사, 적용, 되돌리기가 같이 써요. `go_state.rs`는 helper 가시성만 `pub(crate)`로 바꿨어요(동작 그대로). `go-residue` 단계는 `go_same`/`go_changed`가 남았을 때만 `needs_decision`이고 `plan_digest`와 `--bundle proven` 적용 명령을 내요. 앞선 백업 단계(go-settings, go-state, reports)에 적용이 남아 있으면 명령을 내지 않아요(그 이동이 이 digest를 낡게 만들어요). `src/main.rs`, `src/lib.rs`, `src/help.rs`와 도움말 골든 8개에는 명령 행만 더했어요.

보안 검토가 짚은 것: 매니페스트가 이름 붙인 전체 경로를 파일 이름용 비밀 검사에 넘겨서 `.moai/.env`, `.npmrc`, `credentials*`가 걸러지지 않았어요(S1부터 있던 결함이고 REQ-11을 깼어요). 경로의 모든 성분을 검사하도록 고쳤고, 조사, 후보, 적용 가드, 되돌리기가 모두 같은 `secret_path`를 써요. 테스트 `a_manifest_that_names_a_secret_shaped_path_never_moves_it`가 덮어요.

Verification (이 세션에서 관찰한 실행):

- RED: `cargo nextest run --locked --no-fail-fast --test project_update_go_residue_cli` → 13 tests, 4 passed, 9 failed (새 테스트 9개가 명령이 없어서 실패).
- 구현 뒤 같은 명령 → 13 passed. 비밀 경로 수정 뒤 → 14 passed, 단위 테스트 `go_residue_apply::tests` 2 passed.
- `help_layout` 11 passed(골든은 `MOAI_HELP_GOLDEN_WRITE`로 다시 만들고 다시 비교했어요).
- 첫 `check affected --card` 시도는 `ledger` 단계에서 FAIL: 상태 폴더 리터럴 스캔이 테스트 파일의 `.moai` 철자 58곳(기준 18)을 짚었어요. 허용표를 올리지 않고 매크로 한 곳으로 모아 고쳤고(커밋 `12f50d2d98`), 같은 스캔 2개와 CLI 테스트 14개가 통과했어요.
- 위험 경로 실제 실행(`target/keziah/danger.sh`, 격리 폴더, 진짜 빌드): 매니페스트가 `../outside/x.md`, 절대 경로, 링크 폴더, 링크 파일, `.moai/.env`를 이름 붙인 프로젝트에서 계획이 `refused_paths: 4`를 내고 바깥 파일과 `.env`를 목록에 올리지 않았어요. 인쇄된 적용 명령이 `a.md` 하나만 옮겼고(`committed`), 바깥 파일 3개와 `.env`의 sha256은 적용 전후 같았어요. 위조 영수증(`../outside/x.md`, 절대 경로, 링크 경로, `.moai/.env`)은 각각 `invalid_input`으로 거절됐고 아무것도 쓰지 않았어요. 진짜 되돌리기는 `a.md`를 바이트 그대로 되살렸어요(`reverted`).

Guard mutants (한 슬롯에서 정확히 한 번씩 편집, 지정 테스트, git으로 복원, 끝난 뒤 트리 깨끗함):

- `guard-mutant: paths_that_leave_the_project_are_never_read_or_moved | go_residue::survey의 contained 검사 | exit 100`
- `guard-mutant: the_apply_refuses_a_candidate_that_leaves_the_project | apply_chosen의 경로 가드 | exit 100`
- `guard-mutant: a_forged_receipt_naming_a_path_outside_the_project_is_refused | revert의 receipt 경로 가드 | exit 100`
- `guard-mutant: a_stale_digest_or_a_rejected_permission_reference_changes_nothing | apply의 plan digest 비교 | exit 100`
- `guard-mutant: a_manifest_that_names_a_secret_shaped_path_never_moves_it | survey의 secret_path 검사 | exit 100`

추가로 되돌리기의 충돌 판정(`original_there ||` 줄)을 지운 변이도 `a_new_file_or_a_changed_backup_leaves_only_that_item_as_a_conflict`가 죽였어요(exit 100).

Acceptance:

- AC-1 met: `applying_the_proven_bundle_moves_only_the_go_same_files`와 `a_manifest_that_names_a_secret_shaped_path_never_moves_it`가 통과했어요.
- AC-2 met: `a_stale_digest_or_a_rejected_permission_reference_changes_nothing`와 `only_a_changed_file_can_be_picked_and_one_choice_is_needed`가 통과했어요.
- AC-3 met: `paths_that_leave_the_project_are_never_read_or_moved`, `a_forged_receipt_naming_a_path_outside_the_project_is_refused`, 단위 테스트 `the_apply_refuses_a_candidate_that_leaves_the_project`가 통과했고 위험 경로 실행이 같은 결과를 보였어요.
- AC-4 met: `a_receipt_above_256_kib_applies_and_reverts`(영수증 256 KiB 초과, `committed`와 `reverted`)와 `an_unwritable_backup_area_moves_no_file`이 통과했어요.
- AC-5 met: 단위 테스트 `a_taken_destination_stops_the_apply_and_the_revert_restores_what_moved`가 통과했어요(모듈 안 테스트라서 이 AC의 `verify:` 줄을 `cargo test --lib go_residue_apply`로 고쳤어요; 승인 digest는 `verify:`를 읽지 않아요).
- AC-6 met: `the_revert_restores_every_byte`와 `a_new_file_or_a_changed_backup_leaves_only_that_item_as_a_conflict`가 통과했어요.
- AC-7 met: `applying_the_proven_bundle_moves_only_the_go_same_files`가 `tracked_moved: 3`과 checkout 경고 문구를 확인해요.

Not run / 남은 것: clippy와 넓은 테스트 집합은 이 Result의 범위 밖이고 READY 영수증(`check affected --card`)이 맡아요. 되돌리기는 `leftover_folder`와 `go_marker`를 옮기지 않아요(S3, S5의 일). 마무리 커밋 제안은 S6의 일이에요.
