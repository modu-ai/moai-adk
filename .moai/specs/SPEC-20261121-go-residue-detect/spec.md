---
id: SPEC-20261121-go-residue-detect
title: "A read-only go-residue step finds the files the Go generation left behind, proven against the manifests it can still find"
status: draft
revision: 1
created: 2026-11-21
---

# A read-only go-residue step finds the files the Go generation left behind, proven against the manifests it can still find

## Intent

The Go-generation update survey (2026-11-21, mo.ai.kr measured) found the root gap G1/G2: the update recognises the Go generation only by three marker files and `.moai/manifest.json`, so once go-state has moved the state (or the markers were removed by hand) the Go files left behind — on mo.ai.kr 222 tracked files byte-identical to the backup manifest — are invisible, and the plan answers "nothing to move". The survey's slice S1 adds the detection half: a read-only `go-residue` step of `moai project-update-plan` that finds those files, proves them against a Go manifest it can still find, and classifies them. Applying and reverting are the next slice (S2), on its own card.

## Non-goals

- Applying, moving or reverting anything: the step writes nothing; the apply and revert commands are S2.
- The three Go markers and `state_root` classification: unchanged; a `mixed` verdict still blocks go-state as today (S3 owns the resurrected-marker case).
- The `claude-md` step, the leftovers bookkeeping, `all_unchanged` semantics and the managed ignore block: untouched (S4, S5 own them).
- Codex: withdrawn by decision, so no Codex step or text.
- The `settings_referenced`, `tracked_reports` and `unmanaged_host` classes: the slices that own them (the claude-md step, the leftovers and reports bookkeeping, the unmanaged-host step) extend this survey; this slice adds none of them.

## Requirements

- REQ-1: The project-update plan shall include a read-only `go-residue` step that surveys the project for files the Go generation left behind and writes nothing.
- REQ-2: The go-residue step shall prove a Go-made file against the first manifest it finds in the order `.moai/manifest.json`, then a `manifest.json` under any `.moai-backup*` folder, then a `manifest.json` under `.moai/backups/go-markers-*`, then the backup manifest that the last committed go-state receipt names.
- REQ-3: The go-residue step shall sort every file it surveys into `go_same` (byte-identical to a `deployed_hash` of the found manifest), `go_changed` (named by the manifest with different bytes), `go_marker` (a marker file back after a committed move), `user_data` (in no manifest) and `leftover_folder` (`.moai-cli/`, a `.moai.rust-<UTC>` folder, `.moai-backups/`).
- REQ-4: Each residue item shall carry its path, its sha256, whether git tracks it, whether a committed go-state receipt says it moved and it is back, and the suggested action.
- REQ-5: The go-residue step shall count secret-shaped names without opening their files and never follow a link.
- REQ-6: When no manifest proves any file, the go-residue step shall report no proof with zero candidates and the plan still succeeds. · failure: none — the no-manifest case is this requirement's own subject and leaves nothing to retry
- REQ-7: A second run of the plan shall read the go-residue step unchanged with the same items and the same digests.
- REQ-8: When every surveyed file is user data that no manifest names, the go-residue step shall read unchanged and still list every surveyed file. · failure: none — the user-data-only case is this requirement's own subject and leaves nothing to retry

## Acceptance

- AC-1 (REQ-1, REQ-7): on the survey's temp frame the plan's `go-residue` step reports `go_same` for the byte-identical files, `go_changed` for the changed one, `user_data` for the unmanifested one, counts the one secret-shaped name without opening it, marks the tracked files `tracked: true`, and a second plan run reads the same · verify: `cargo test --test project_update_go_residue_cli`
- AC-2 (REQ-2): with `.moai/manifest.json` absent the step still proves the same files from the backup manifest and the go-markers manifest · verify: `cargo test --test project_update_go_residue_cli`
- AC-3 (REQ-6): with no manifest anywhere the step reports no proof and zero candidates while the plan exits 0 · verify: `cargo test --test project_update_go_residue_cli`
- AC-4 (REQ-3, REQ-4, REQ-5): every item carries path, sha256, tracked, resurrected and a suggested action, a symlink is never followed, and the snapshots around the plan are identical · verify: `cargo test --test project_update_go_residue_cli`
- AC-5 (REQ-8): a project whose only surveyed file is an unmanifested committed file reads the step as unchanged, lists the file as user_data with tracked true and action keep, and does not turn the update's all_unchanged false · verify: `cargo test --test project_update_go_residue_cli`

## Scope

- write: `src/go_residue.rs` (new), `src/project_update.rs` (the step), `tests/project_update_go_residue_cli.rs` (new), the step-list pins `tests/go_settings_cli.rs`, `tests/project_update_plan_cli.rs` and `tests/reports_move_cli.rs` (each gains the read-only step name), one module line in `src/lib.rs` (a shared file, written by line lease), one row of `tests/fixtures/state_folder_literals/allow.tsv` for the new test file, `.moai/specs/SPEC-20261121-go-residue-detect/**`
- read-only: `src/go_state.rs`, `src/state_root.rs`, the survey report sections 4-8
- forbidden: any card number in shipped text, real HOME or real project writes, apply or revert behaviour

## Plan

The temp frame follows the survey's section 8: a git repository with a valid Rust marker and a profile, a backup manifest with `deployed_hash` values, a resurrected policies file and a workflow file byte-identical (`go_same`), a changed config section (`go_changed`), an unmanifested design file (`user_data`), one secret-shaped name, an output style a settings file references, and a `.moai-backup` manifest chain with the live `.moai/manifest.json` removed for AC-2. RED: `tests/project_update_go_residue_cli.rs` fails on a tip without the step. GREEN: the detection module plus the plan step.

## Risks

- A `deployed_hash` collision across manifests: the first-found manifest wins and the item names which manifest proved it.
- Large projects: the survey walks the state folders and the manifest-named paths only, never the whole tree, and stops at the same depth limits go-state uses.
- The step must stay read-only even on error: the test asserts identical snapshots after a failing run.

## Result

`status: draft`예요. 승인은 이 파일에 기록하지 않았어요. 구현 head는 `ba3659ba637915955d044ef0de59ba357ee44f2c`이고, base는 NeXT의 `64b948a105932a0168d4b835b7bffb47b9a90b91`이에요. 이 작업은 같은 SPEC의 이전 브랜치를 이어받아요. 그 브랜치의 커밋 네 개(SPEC, 조사 모듈과 step, house-rules 수정, step 목록 고정 테스트)를 충돌 없이 cherry-pick으로 현재 NeXT 끝에 옮겼고, 그 위에 수리 커밋 둘을 더했어요. 첫 수리는 위 `user_data` 회귀이고, 둘째 수리(열째 묶음 넓은 검사가 찾은 것)는 `src/go_residue.rs` 가 옛 state 폴더 이름을 상수로 직접 쓰던 것을 `state_root::EARLIER_PROJECT_DIR` 로 바꾸고, 그 때문에 필요 없어진 `src/lib.rs` 의 state 폴더 allow-list 한 줄을 뺀 거예요. 수리 전 `path_root_direct_sites::the_repository_spells_the_earlier_folder_only_where_the_baseline_lists` 가 left `[("go_residue.rs", 1)]` 로 실패했고(RED, 이 세션에서 관찰), 수리 뒤 그 시험과 `state_dir_tests::project_state_paths_are_resolved_per_project` 가 통과했어요(`path_root_direct_sites`·`project_update_go_residue_cli`·`project_update_plan_cli` 82개 통과 exit 0). 처음 영수증이 이 시험을 고르지 못한 건 `check affected --card` 가 모듈 기준으로 시험을 골랐기 때문이에요. 체크 영수증(`check affected --card`)은 READY 노트에 있어요.

수리가 고친 것: 수리 전에는 평범하게 업데이트된 프로젝트의 plan이 `.moai/workflows/project/` 아래 현재 세대의 자체 파일을 `user_data`로 나열했고, step이 `needs_decision`으로 읽혔어요. 그래서 기존 테스트 `declined_skip_reads_skip_and_names_the_way_back`이 `all_unchanged`가 false라며 실패했어요. 새 테스트 `user_data_alone_leaves_the_go_residue_step_unchanged`의 RED는 `left: "needs_decision"`, `right: "unchanged"`였고(이 세션에서 관찰), 변경 뒤에는 GREEN이에요.

Verification (이 세션에서 관찰한 실행):

- 옮겨 온 커밋에서 `cargo nextest run --locked --test project_update_go_residue_cli` → 3 passed (exit 0).
- 수리 뒤 `cargo nextest run --locked --no-fail-fast --test project_update_go_residue_cli --test project_update_plan_cli --test go_settings_cli --test reports_move_cli` → 126 tests run, 126 passed, exit 0 (새 테스트 `user_data_alone_leaves_the_go_residue_step_unchanged` 포함).

Acceptance:

- AC-1 met: `project_update_go_residue_cli`의 원래 테스트 중 분류 프레임 테스트가 통과했어요(위 3 passed와 126 passed 실행).
- AC-2 met: 같은 파일의 백업 manifest 증명 테스트(`.moai/manifest.json` 없음)가 통과했어요.
- AC-3 met: 같은 파일의 manifest 없음 테스트가 통과했어요.
- AC-4 met: 위 원래 세 테스트가 덮는 범위로만 말해요. 위 두 실행에서 `project_update_go_residue_cli`가 모두 통과했어요.
- AC-5 met: 새 테스트 `user_data_alone_leaves_the_go_residue_step_unchanged`가 통과했어요(126 passed 실행).

Not run: clippy는 `not_run`이에요. guard-mutant 줄도 `not_run`이에요(tier 1이라 요구되지 않아요). 문서 단계는 `moai-docs`가 했고, `@cc:` 주석은 추가하지 않았고 필요하지도 않았어요.
