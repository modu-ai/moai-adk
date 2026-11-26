---
id: SPEC-20261126-tracked-state-move
title: "무시될 곳으로 가는 추적 상태 파일은 묻고, flaky 원장은 링크 폴더에 쓰지 않아요"
status: draft
revision: 1
created: 2026-11-26
---

## Interview

- none: 두 문제는 감사 노트 9952 와 조정 세션의 노트 10043·10088 에서 왔고, 사용자에게 새로 물은 것은 없어요. 남은 선택은 Decisions 에 기본값과 함께 적었어요.

# 무시될 곳으로 가는 추적 상태 파일은 묻고, flaky 원장은 링크 폴더에 쓰지 않아요

## Intent

옛 상태 폴더(`.moai-cli/`)만 있거나, `.moai/` 에 SPEC 폴더와 local 폴더만 있는 프로젝트에서는 상태 이동 계획이 `.gitignore` 를 확인하지 않아요. 그래서 git 이 추적하던 파일이 무시되는 곳으로 옮겨지면 사용자에게 한마디 없이 버전 관리에서 빠져요. 또 flaky 원장 기록은 `.moai` 나 `.moai/local` 이 링크이면 따라가서 프로젝트 밖에 써요. 이 SPEC 은 두 곳을 더 엄격한 방향으로만 고쳐요: 이동은 파일 이름을 대고 사용자의 결정을 기다리고, 기록은 링크를 거절해요.

## Non-goals

- Go 세대 폴더 이동(`folder_moves`)의 무시 규칙 확인, 보존 이름 확인이 항목 단위라 안쪽의 무시된 파일을 놓치는 것, 프로젝트가 저장소의 하위 폴더일 때 확인을 건너뛰는 것: 같은 부류지만 이 변경 밖이에요(D-4).
- `check flaky` 의 읽기가 링크를 따라가는 것.
- 새 선택 플래그, 스킬·문서 변경.
- 이미 옮겨진 프로젝트를 되돌리거나 고치는 것.

## Requirements

- REQ-1: When the state move plans the earlier state into a `.moai/` that is absent or holds nothing beyond the SPECs folder, the local folder and `.DS_Store`, the plan shall name each tracked file whose destination git ignores, with the rule that hides it, at most 20 names and the total.
- REQ-2 (REQ-1): IF such a destination is ignored, or the marker is ignored while a moving source holds a tracked file, THEN the update preview shall read the go-state step as `needs_decision` with no apply command and a reason that names them.
- REQ-3 (REQ-1): IF the folder is in a git work tree and git cannot list the tracked files or judge a destination, THEN the plan shall mark the check `unknown` and stop the move as for an ignored destination.
- REQ-4: When `go-state-apply` receives a move of that kind, the apply shall judge the destinations again before its first write.
- REQ-5 (REQ-4): IF that judgement is ignored or unknown, THEN the apply shall refuse and leave every file and folder of the project unchanged.
- REQ-6: Where every destination is allowed, no moving source holds a tracked file, or the folder is in no git work tree, the plan shall offer the move as it did before this change.
- REQ-7: When a flaky record or a SPEC write adds the local folder or the SPECs folder to the `.moai/`, the plan shall answer the same go-state state and name the same files as before that write.
- REQ-8 (REQ-7): IF the `.moai/` holds another kept name or a `backups/` folder, THEN an ignored target shall keep blocking the move as it did before this change.
- REQ-9: When a flaky name is recorded, the record shall append to the ledger only through real folders and a real file below the project root.
- REQ-10 (REQ-9): IF `.moai`, `.moai/local` or the ledger file is a link, THEN the record shall return a refusal that names the path, write nothing through the link and leave the check result unchanged.
- REQ-11: When a flaky record adds the local folder to a `.moai/` that holds `backups/`, no kept name and no go-state receipt, the state classification shall answer what it answered before the record.
- REQ-12 (REQ-11): IF that `backups/` holds a reverted go-state receipt, THEN the state classification shall answer `none` as it does after a revert today.

## Acceptance

- AC-1 (REQ-1, REQ-2): `.moai/` 가 없고 추적 중인 옛 상태 파일의 목적지가 무시되면, 계획이 파일과 규칙을 대고 업데이트 단계는 `needs_decision` 이며 적용 명령이 없어요 · verify: `cargo test --locked --offline --test state_tracked_move_cli an_absent_state_folder_names_the_hidden_tracked_files_and_offers_no_command`
- AC-2 (REQ-1, REQ-2): local 폴더만 있는 `.moai/` 에서도 같아요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_local_only_state_folder_names_the_hidden_tracked_files_and_offers_no_command`
- AC-3 (REQ-7): flaky 기록과 SPEC 쓰기 앞뒤로 가려진 배치와 허용된 배치의 상태와 파일 답이 같아요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_flaky_record_and_a_spec_write_leave_the_move_answer_as_it_was`
- AC-4 (REQ-1, REQ-2): 표지 파일이나 안쪽 폴더의 파일만 무시돼도 이동이 멈춰요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_hidden_marker_or_nested_file_stops_the_move`
- AC-5 (REQ-4, REQ-5): 무시되는 곳으로의 적용은 거절되고 프로젝트의 파일과 폴더가 그대로예요 · verify: `cargo test --locked --offline --test state_tracked_move_cli an_apply_into_an_ignored_place_is_refused_and_writes_nothing`
- AC-6 (REQ-3, REQ-5): git 이 답하지 못하면 적용 명령이 없고 적용은 거절돼요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_check_git_cannot_answer_offers_no_command`
- AC-7 (REQ-6): 허용된 목적지, 추적 파일 없음, git 작업 트리가 아닌 폴더에서는 이동이 제안되고 적용돼요 · verify: `cargo test --locked --offline --test state_tracked_move_cli an_allowed_or_untracked_move_is_offered_and_applies`
- AC-8 (REQ-8): 다른 보존 이름이나 `backups/` 가 있으면 지금처럼 막혀요 · verify: `cargo test --locked --offline --test state_tracked_move_cli other_kept_names_or_backups_keep_the_move_blocked`
- AC-9 (REQ-9, REQ-10): 링크인 `.moai` 는 기록이 거절되고 링크 대상이 그대로예요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_linked_state_folder_refuses_the_flaky_record`
- AC-10 (REQ-10): 링크인 local 폴더도 같아요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_linked_local_folder_refuses_the_flaky_record`
- AC-11 (REQ-10): 링크이거나 끊어진 링크인 원장 파일도 같아요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_linked_ledger_file_refuses_the_flaky_record`
- AC-12 (REQ-9): 진짜 폴더에는 원장 줄이 적혀요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_real_local_folder_takes_the_flaky_record`
- AC-13 (REQ-10): 원장 오류는 검사 결과를 바꾸지 않아요 · verify: `cargo test --locked --offline --test check_affected_rerun ac5_an_unwritable_ledger_does_not_change_the_result`
- AC-14 (REQ-7): flaky 기록이 상태 판정과 업데이트 미리보기를 바꾸지 않는 기존 기준이 유지돼요 · verify: `cargo test --locked --offline --test check_flaky_verdict a_flaky_record_leaves_the_state_verdict_and_the_update_preview_as_they_were`
- AC-15 (REQ-12): 이동과 되돌리기 뒤에 local 폴더가 남는 기존 기준이 유지돼요 · verify: `cargo test --locked --offline --test check_flaky_verdict a_state_move_and_its_revert_leave_a_local_folder_in_place`
- AC-16 (REQ-8): 기존의 막는 확인이 그대로예요 · verify: `cargo test --locked --offline --test go_state_cli a_target_that_git_ignores_blocks_the_plan_until_the_rule_allows_it`
- AC-17 (REQ-11): `backups/` 만 있는 `.moai/` 의 판정을 flaky 기록이 바꾸지 않아요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_flaky_record_beside_backups_alone_keeps_the_verdict`
- AC-18 (REQ-12): 되돌린 영수증과 local 폴더가 함께 있으면 여전히 `none` 이에요 · verify: `cargo test --locked --offline --test state_tracked_move_cli a_reverted_receipt_beside_a_local_folder_still_reads_none`

## Scope

- write: `.moai/specs/SPEC-20261126-tracked-state-move/spec.md`
- write: `src/go_state.rs`
- write: `src/check_flaky.rs`
- write: `src/project_update.rs`
- write: `src/state_root.rs`
- write: `tests/state_tracked_move_cli.rs`
- read-only: `src/check_affected.rs`, `tests/check_flaky_verdict.rs`, `tests/check_affected_rerun.rs`, `tests/go_state_cli.rs`, `tests/areas.toml`
- forbidden: 기존 거절을 약하게 하는 것, 테스트에서 실제 HOME 을 읽고 쓰는 것(임시 HOME 만), 공용 파일을 범위에 넣는 것

## Plan

0. 앞 조건: `src/go_state.rs` 와 `src/check_flaky.rs` 를 함께 바꾸는 head `b86eb558d8` 이 통합 브랜치에 들어온 뒤 그 tip 을 합치고 구현해요. `src/project_update.rs` 를 쥔 작업의 담당에게 먼저 알려요. 겹치는 곳은 늦게 착지하는 쪽이 합쳐요.
1. 실패하는 테스트를 새 파일 `tests/state_tracked_move_cli.rs` 에 먼저 써요. 실제 바이너리를 임시 HOME 과 임시 저장소에 대고 돌리고, 폴더 이름은 상수로 적어요. 두 배치(`.moai/` 없음, local 만)와 `backups/` 만 있는 배치를 먼저 재현해요.
2. `check_flaky::record`: `create_dir_all` 대신 `go_state::make_dirs` 를 쓰고, 열기 전에 `go_state::linked_component` 로 원장 경로의 링크를 거절해요. 거절은 `Err` 라서 호출한 쪽은 지금처럼 표준 오류에만 적어요.
3. `go_state`: 무시 확인을 하지 않는 배치에서 `rev-parse --is-inside-work-tree`, `ls-files -z`, 목적지마다 `check-ignore --no-index -v -n`(기존 묶음 크기)으로 판정해 계획에 선택 필드 `ignored_tracked`(상태, 파일과 규칙 최대 20개, 전체 수, 표지, 이유)를 더해요.
4. 적용: 첫 폴더나 영수증을 만들기 전에 같은 판정으로 거절해요.
5. 업데이트 단계: `ignored_tracked` 가 있으면 `needs_decision`, 적용 명령 없음, 계획의 이유를 그대로 보여 줘요.
6. `state_root`: 보존 이름 없이 `local/` 과 `backups/` 가 함께 있으면 `backups/` 아래 go-state 영수증이 있을 때만 지금처럼 읽어요.
7. `cargo fmt --all`, 표적 테스트, 새 테스트마다 가드 뮤턴트, `check affected --card` 한 번. 위험 경로 확인은 AC-5 와 AC-9~AC-11 의 임시 HOME Rust 테스트예요.

## Risks

- 의도한 새 멈춤: 옛 상태 폴더의 파일을 추적하면서 새 위치를 무시하는 프로젝트는 적용 명령을 받지 못하고, 그 단계 뒤에서 기다리는 설치·합치기 단계도 사용자가 고칠 때까지 기다려요.
- 나중에 사용자가 `.moai/` 에 보존 이름이나 `backups/` 를 만들면 같은 프로젝트의 답이 `needs_decision` 에서 `blocked` 로 바뀌어요. flaky 기록은 둘 다 만들지 않아요.
- 비어 있거나 `.DS_Store` 만 있는 `.moai/` 는 코드를 읽으면 기록 뒤 `none` 이 되어 제안되지 않던 이동이 열릴 수 있어요. 파일이 가려지면 새 확인이 멈추지만 나머지는 "local 만이면 `none`" 이라는 기존 기준과 부딪혀 두어요.
- 일부러 링크한 `.moai/local` 은 flaky 횟수를 더 모으지 않아요. 표준 오류 한 줄이 알려요.
- 두 기록이 동시에 폴더를 만들면 한 줄이 빠질 수 있어요(표준 오류에 남아요).
- `src/go_state.rs` 와 `src/project_update.rs` 는 다른 작업도 바꾸고 있어 합칠 때 충돌할 수 있어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "묻기"는 `needs_decision` + 적용 명령 없음 + 적용 거절. 사용자는 무시 규칙을 고치거나 파일 추적을 끝낸 뒤 다시 미리보기 해요 | 기본값(승인 대기). 대안: `blocked` / 새 선택 플래그 / 이름만 알리고 명령은 유지(확인 한 번에 파일이 옮겨져요) | 단계의 분기 하나를 되돌려요 |
| D-2 | SPEC 폴더만, SPEC+local 배치도 같은 확인 | 기본값. 빼면 `spec new` 가 멈춘 이동을 확인 없는 이동으로 바꿔요 | 조건을 두 배치로 좁혀요 |
| D-3 | `backups/` 만 있는 배치(노트 10043)도 이 변경에서 고쳐요(REQ-11, REQ-12) | 기본값. 앞 SPEC 의 "local 만이면 `none`" 읽기를 `backups/` 가 있는 경우에 좁혀요 | REQ-11·12 와 AC-17·18 을 빼고 후속 일로 적어요 |
| D-4 | 같은 부류 세 가지(Go 폴더 이동 무확인, 항목 단위 확인, 하위 폴더 프로젝트)는 여기서 안 고쳐요 | 기본값. 카드 기준 밖이고 재현하지 않았어요(코드 읽기만) | 후속 일로 적어요 |
| D-5 | 파일 이름은 20개까지와 전체 수 | 기본값. 대안: 전부 또는 3개 | 상수 하나 |
| D-6 | 스킬·운영 문서는 안 바꿔요(이유 문장이 결정을 알려요) | 기본값. 바꾸면 공용 원장 파일과 다른 작업이 쥔 파일이 필요해요 | 문서 조각을 더해요 |
| D-7 | 관찰 기간 없이 바로 거절해요. 오늘도 이 단계는 `needs_decision` 으로 기다리고, 가려진 경우에만 명령을 빼요 | 기본값. 카드가 멈춤을 요구해요. 대안: 경고만(파일이 조용히 빠지는 길이 남아요) | D-1 의 대안으로 바꿔요 |

## Result

확인한 커밋은 `3980416fbb73457c5ae7d3e727bccb3bc6efcd89`(기준 `cb3b046473`), 실패를 먼저 본 커밋은 `41e0495a24` 예요. 아래 명령은 heavy slot 안에서 돌렸고, 뮤턴트는 커밋한 코드에 하나씩 넣고 되돌렸어요(`git diff --quiet HEAD -- src` 로 확인). `cargo clippy --locked --offline --all-targets -- -D warnings` 는 exit 0 이에요.

- AC-1: met — `an_absent_state_folder_names_the_hidden_tracked_files_and_offers_no_command` exit 0(1 통과). 구현 전 실패. 뮤턴트 `go_state.rs:3126`(판정 호출 제거), `project_update.rs:466`(단계 분기 무효화) 모두 exit 101.
- AC-2: met — `a_local_only_state_folder_names_the_hidden_tracked_files_and_offers_no_command` exit 0. 구현 전 실패. 뮤턴트 `go_state.rs:3123`(`.moai/` 가 없을 때만 판정) exit 101.
- AC-3: met — `a_flaky_record_and_a_spec_write_leave_the_move_answer_as_it_was` exit 0. 구현 전 실패. 뮤턴트 `go_state.rs:3123`(같은 좁힘) exit 101.
- AC-4: met — `a_hidden_marker_or_nested_file_stops_the_move` exit 0. 구현 전 실패. 뮤턴트 `go_state.rs:4142`(표지 규칙 버림) exit 101.
- AC-5: met — `an_apply_into_an_ignored_place_is_refused_and_writes_nothing` exit 0. 구현 전 실패. 뮤턴트 `go_state.rs:4764`(적용 거절 무효화) exit 101.
- AC-6: met — `a_check_git_cannot_answer_offers_no_command` exit 0. 구현 전 실패. 뮤턴트 `go_state.rs:4043`(`unknown` 을 허용으로 바꿈) exit 101.
- AC-7: met — `an_allowed_or_untracked_move_is_offered_and_applies` exit 0. 구현 전에도 통과(바뀌면 안 되는 동작). 뮤턴트 `go_state.rs:4109`(추적 파일이 없어도 표지를 물음) exit 101.
- AC-8: met — `other_kept_names_or_backups_keep_the_move_blocked` exit 0. 구현 전에도 통과. 뮤턴트 `go_state.rs:3117`(기존 막는 확인을 건너뜀) exit 101.
- AC-9: met — `a_linked_state_folder_refuses_the_flaky_record` exit 0. 구현 전 실패. 뮤턴트 `check_flaky.rs:477`(`make_dirs` 를 `create_dir_all` 로) exit 101.
- AC-10: met — `a_linked_local_folder_refuses_the_flaky_record` exit 0. 구현 전 실패. 뮤턴트 `check_flaky.rs:477`(같은 교체) exit 101.
- AC-11: met — `a_linked_ledger_file_refuses_the_flaky_record` exit 0. 구현 전 실패. 뮤턴트 `check_flaky.rs:479`(링크 거절 제거) exit 101(일반 파일 확인이 남아 거절 문구에서 실패), `:479`와 `:484` 를 함께 제거해도 exit 101.
- AC-12: met — `a_real_local_folder_takes_the_flaky_record` exit 0. 구현 전에도 통과. 뮤턴트 `check_flaky.rs:477`(기록이 늘 거절) exit 101. 추가 테스트 `a_ledger_that_is_no_regular_file_refuses_the_flaky_record` 는 전체 대상 실행에서 통과, 뮤턴트 `check_flaky.rs:484` exit 101(FIFO 에서 30초 안에 답이 없음).
- AC-13: met — `check_affected_rerun ac5_an_unwritable_ledger_does_not_change_the_result` exit 0, 대상 전체 12 통과.
- AC-14: met — `check_flaky_verdict a_flaky_record_leaves_the_state_verdict_and_the_update_preview_as_they_were` exit 0, 대상 전체 3 통과.
- AC-15: met — `check_flaky_verdict a_state_move_and_its_revert_leave_a_local_folder_in_place` exit 0.
- AC-16: met — `go_state_cli a_target_that_git_ignores_blocks_the_plan_until_the_rule_allows_it` exit 0, 대상 전체 77 통과.
- AC-17: met — `a_flaky_record_beside_backups_alone_keeps_the_verdict` exit 0. 구현 전 실패. 뮤턴트 `state_root.rs:523`(`local_counts = local`) exit 101.
- AC-18: met — `a_reverted_receipt_beside_a_local_folder_still_reads_none` exit 0. 구현 전에도 통과. 뮤턴트 `state_root.rs:523`(`local && !backups`) exit 101.

그 밖의 통과: `state_tracked_move_cli` 15, `go_state_preview_cli` 44, `project_update_plan_cli` 73, `update_end_to_end_cli` 8, `state_root_cli` 6, 라이브러리 `go_state::` 16 · `check_flaky::` 5 · `state_root::` 29 · `project_update::` 9.

기존 테스트 코드는 한 줄만 바꿨어요. `src/go_state.rs` 단위 테스트의 `sample()` 에 `ignored_tracked: None,` 을 더했어요(새 필드라 컴파일에 필요, 단언은 그대로).

남은 한계(코드 읽기 기준, 재현하지 않음): 계획이 이미 막힌 경우에는 `ignored_tracked` 를 만들지 않아요(아무것도 옮기지 않아요). 이동하는 폴더 안의 `.gitignore` 가 새 위치에서 가리는 경우는 판정하지 않아요. 인덱스와 디스크의 대소문자가 다른 항목은 판정에서 빠질 수 있어요. git 이 "저장소가 아님" 이라고 답하는 깨진 `.git` 은 작업 트리가 아닌 것으로 봐요.

실행하지 않은 것(not_run): `moai check affected`(상위 세션이 실행), 전체 `cargo test`, 위에 없는 테스트 대상, Windows·Linux 확인, Plan 0 의 `src/project_update.rs` 담당 알림.
