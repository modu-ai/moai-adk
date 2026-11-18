---
id: SPEC-20261119-hooks-followups
title: "The hooks step tells what the file and the record hold, and the go-hooks apply proves its put-back"
status: draft
revision: 2
created: 2026-11-19
---

# The hooks step tells what the file and the record hold, and the go-hooks apply proves its put-back

## Intent

`moai project update`의 `hooks` 단계와 `go-hooks-apply`는 앞선 감사(#3137)에서 PASS를 받았지만 일곱 가지 후속(N1~N7)이 남았어요. 공통점은 **단계가 읽는 것과 `hooks-plan`·`hooks-status`가 읽는 것이 달라서 "unchanged/current"라고 말하는데 실제로는 hooks가 없거나 다른 상태**인 자리, 그리고 **되돌리는 코드가 시험 없이 있는 자리**예요. 이 SPEC은 그 일곱 가지 중 `src/project_update.rs`와 `src/go_hooks.rs` 안에서 풀 수 있는 것을 고쳐요: 단계가 사실대로 말하고, 되돌림은 시험으로 고정해요.

## Non-goals

- Codex·OMP 파일이 다른 내용으로 바뀌었을 때의 `requires preservation review` 막힘에 길을 내는 일(N4 뒷부분): 그 글은 `src/hooks.rs`·`src/project_hosts.rs`에서 나와서 이 조각의 쓰기 범위 밖이에요. 리더에게 범위 확장 또는 후속 카드로 넘겨요.
- 설정 이동도 기록도 없는 프로젝트에 hooks 명령을 `ready`로 내미는 일: 앞선 사이클의 규칙(이동이 없으면 제안도 없음)을 지켜요. 이 단계는 그 자리를 말로만 알려요(D-2).
- `hooks-apply`가 사용자가 고친 `hooks`를 덮어쓰게 하는 일: 사용자의 수정을 지키는 현재 규칙은 그대로예요(D-1).
- `src/main.rs`, `src/claude.rs`, `src/hooks.rs`, `src/contract.rs`의 변경.

## Requirements

- REQ-1: The hooks step shall consult the hooks plan of the Claude host whenever the profile hosts include `claude`.
- REQ-2 (REQ-1): IF the file holds no MoAI hook set, no ownership record exists and no settings move is recorded, THEN the hooks step shall stay `unchanged` with no command and name the gap, the reading of the hooks plan and the commands to run by hand.
- REQ-3: The go-hooks apply shall run its install step through a closure argument, so that a test can supply an installer of its own.
- REQ-4 (REQ-3): IF the installer returns an error, or returns success while the settings file does not hold the previewed hooks, THEN the go-hooks apply shall put the settings file, the ownership record and the scripts back and mark its receipt `rolled_back`.
- REQ-5: The hooks step shall not call a `hooks` key that differs from the MoAI hook set "current".
- REQ-6 (REQ-5): IF the `hooks` key is `{}` or `null`, THEN the hooks step shall read `needs_decision` and name the way out.
- REQ-7: The claude-settings step shall say so when the status line key is gone from the file while MoAI's record still owns it.
- REQ-8: IF the ownership record cannot be read, THEN the hooks step shall name the record's path and the next step.
- REQ-9: The hooks step shall plan an older MoAI hook set that the file and the record both hold as `ready`.
- REQ-10 (REQ-9): IF the file holds an older MoAI hook set with a user's edit, THEN the hooks step shall read `conflict`.
- REQ-11: The undo order shall name `hooks-revert` for a hooks step that ran `hooks-apply --host claude`.
- REQ-12 (REQ-4): IF the read-back finds that the settings file does not hold the previewed hooks, THEN the go-hooks apply shall fail with the kind `conflict`.

## Acceptance

- AC-1 (REQ-1, REQ-2): 프로필에 `claude`가 있고 파일에 hooks가 없고 기록도 이동도 없으면 단계는 `unchanged`이고 명령이 없고, 사유가 "no MoAI hook set is installed"·`hooks-plan`이 읽는 값(`update`)·`moai hooks-apply … --host claude`를 말하고, `preview.configured`가 `hooks-status`와 같은 값(`drifted`)이에요; 사유가 말한 명령으로 설치하면 단계가 "the MoAI hook set is current"로 바뀌어요. 프로필에 `claude`가 없으면 `hooks-plan`을 읽지 않고 `unchanged`예요.
  verify: `cargo test --locked --offline --test project_update_hooks_cli a_project_with_nothing_yet_names_the_gap_and_prints_no_command`; `cargo test --locked --offline --test project_update_hooks_cli a_file_without_hooks_and_without_a_record_says_drifted_like_the_status`; `cargo test --locked --offline --test project_update_hooks_cli a_profile_without_claude_leaves_the_hooks_step_unchanged_without_a_command`
- AC-2 (REQ-3, REQ-4, REQ-12): 설치 함수가 성공을 말하고 아무것도 쓰지 않거나 다른 것을 쓰거나 오류를 내면 설정 파일·기록·스크립트가 처음 바이트로 돌아가고 영수증이 `rolled_back`이에요; 되읽기 실패는 `conflict` 종류이고 `internal`이 아니에요.
  verify: `cargo test --locked --offline --test go_hooks_migrate an_install_that_reports_success_without_writing_the_hooks_is_put_back`; `cargo test --locked --offline --test go_hooks_migrate an_installer_that_writes_something_else_is_undone_with_its_record`; `cargo test --locked --offline --test go_hooks_migrate an_installer_that_fails_is_put_back_the_same_way`; `cargo test --locked --offline --test go_hooks_migrate the_read_back_failure_is_a_conflict_not_an_internal_error`
- AC-3 (REQ-5, REQ-6): MoAI 세트가 아닌 `hooks`는 "current"로 읽히지 않아요: 사용자가 고친 세트는 `unchanged`·명령 없음·`drifted`이고, 기록 없이 사용자가 쓴 키는 `conflict`로 읽히는 `hooks-plan`과 같은 말을 하고, `{}`·`null`은 `needs_decision`이고 손으로 `hooks` 키를 지우고 다시 계획하라고 말해요.
  verify: `cargo test --locked --offline --test project_update_hooks_cli a_hooks_key_that_is_not_the_moai_set_is_not_called_current`; `cargo test --locked --offline --test project_update_hooks_cli an_empty_or_null_hooks_key_reads_needs_decision_and_names_the_way_out`; `cargo test --locked --offline --test project_update_hooks_cli a_hooks_key_of_the_users_own_with_no_record_says_conflict_and_prints_no_command`
- AC-4 (REQ-7): 기록은 상태줄을 가졌는데 파일에서 손으로 지워졌으면 claude-settings 단계가 `unchanged`이되 사유가 "removed by hand and is not put back"을 말하고 `preview.status_line_removed`가 참이에요.
  verify: `cargo test --locked --offline --test project_update_hooks_cli a_status_line_removed_by_hand_is_named_not_called_current`
- AC-5 (REQ-8): 읽을 수 없는 기록이면 hooks 단계는 `blocked`이고 명령이 없고, 사유가 `manifests/claude-settings.json`·"no moai command repairs it"·"plan again"을 말해요.
  verify: `cargo test --locked --offline --test project_update_hooks_cli an_unreadable_record_names_its_file_and_the_next_step`
- AC-6 (REQ-9, REQ-10): 파일과 기록이 같은 옛 세트를 가지면 `ready`이고 `hooks-plan`의 명령과 같은 명령이고 적용하면 지금 세트가 돼요; 옛 세트에 사용자의 수정이 얹히면 `conflict`이고 명령이 없어요.
  verify: `cargo test --locked --offline --test project_update_hooks_cli an_older_hook_set_held_by_the_file_and_the_record_reads_ready`; `cargo test --locked --offline --test project_update_hooks_cli an_older_hook_set_with_a_users_edit_reads_conflict`
- AC-7 (REQ-11): `hooks-apply --host claude`가 기존 파일 위에 돈 hooks 단계의 되돌리는 순서 글이 `hooks-revert`를 말하고, 첫 설치에는 되돌릴 것이 없다고 말해요.
  verify: `cargo test --locked --offline --test project_update_hooks_cli the_undo_order_names_hooks_revert_for_a_hooks_apply_update`
- AC-8 (REQ-1…REQ-12): 이웃 검사: 업데이트 계획을 읽는 시험 파일들(고정된 AC 13 시험 포함)이 그대로 통과해요.
  verify: `cargo test --locked --offline --test project_update_plan_cli`; `cargo test --locked --offline --test project_update_hooks_cli`; `cargo test --locked --offline --test go_hooks_migrate`; `cargo test --locked --offline --test home_isolation_scan`; `cargo clippy --offline --locked --all-targets -- -D warnings`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261119-hooks-followups/spec.md`, `src/project_update.rs`, `src/go_hooks.rs`, `tests/project_update_hooks_cli.rs`, `tests/go_hooks_migrate.rs`
- read-only: `src/hooks.rs`, `src/claude.rs`, `src/contract.rs`, `src/status_line.rs`, `src/project_hosts.rs`, `tests/project_update_plan_cli.rs`
- forbidden: 실제 HOME·실제 프로젝트의 `.claude/settings.json`(시험은 임시 프로젝트와 임시 HOME만), 네트워크, 새 의존성, `src/main.rs`

## Plan

1. 시험 먼저: `tests/go_hooks_migrate.rs`에 설치 함수를 바꿔 끼우는 시험 4개, `tests/project_update_hooks_cli.rs`에 단계가 말하는 것 시험을 쓰고 RED(실패 9개)를 읽어요.
2. `src/go_hooks.rs`에 `Installer` 인자와 되읽기 실패(`conflict`)를 넣고, `src/project_update.rs`의 hooks 단계가 프로필의 `claude`를 보고 `hooks-plan`을 읽게 해요(`unchanged_hooks`, `record_error`, 상태줄 문구, 되돌리는 순서 글).
3. 리더의 답(결정 D-2)에 따라 "기록도 이동도 없으면 제안 없음"을 지키도록 맞추고, 고정된 `project_update_plan_cli` 시험 59개가 그대로 통과하는지 읽어요.
4. 이웃 검사와 변이 시험을 돌려요.

## Risks

- 단계가 말하는 상태가 바뀌어 기존 시험이 깨질 수 있어요 → `project_update_plan_cli`를 필터 없이 돌려 고정된 AC 13 시험이 그대로 통과함을 확인해요; 바뀌는 문구는 이 SPEC의 시험만 가져요.
- 사유 글이 길어져 화면이 어수선해질 수 있어요 → 단계의 `preview`는 키만 더하고 `commands`는 그대로 둬요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 사용자가 고친 `hooks`는 덮어쓰지 않아요: `unchanged`로 두고 "current"라고 부르지 않아요 | 사용자의 수정을 지키는 현재 규칙; `hooks-status`도 `drifted`로 읽어요 | 없음 |
| D-2 | 설정 이동도 기록도 없는 프로젝트의 hooks 단계는 `unchanged`·명령 없음이고, 사유가 빈틈과 직접 실행할 명령을 말하고 `preview.configured`는 `hooks-status`처럼 읽어요 | 리더의 결정 B(카드 노트 2281): 앞선 사이클의 AC 13과 고정 시험 8개를 그대로 두고 범위를 넓히지 않아요; 옵션 A(`ready`로 제안)는 그 시험 8개를 다시 쓰는 일이라 받지 않았어요 | 다음 SPEC이 AC 13을 바꾸면 이 갈래의 `unchanged`를 `ready`로 바꿔요 |
| D-3 | `{}`·`null`인 `hooks`는 `needs_decision`이고 명령이 없어요 | 어떤 명령도 사용자가 고친 키를 바꾸지 않으므로 손으로 지우는 길만 있어요 | 없음 |
| D-4 | 되읽기 실패는 `conflict` 종류예요 | `contract::classify`가 "digest mismatch"를 `conflict`로 읽어요; 사용자 파일과 어긋난 상태이지 내부 오류가 아니에요 | 글을 바꾸면 종류가 바뀌어요 |
| D-5 | 옛 세트 시험은 같은 쓰기 함수를 이 프로세스에서 불러요 | `hook_legacy` 관문이 PATH 첫 `moai`가 Rust 빌드이기를 요구하는데 시험 환경에는 없어요 | 시험 환경이 PATH를 가지면 명령으로 바꿔요 |
| D-6 | Codex·OMP의 `requires preservation review` 길 내기는 하지 않고 리더에게 넘겨요 | 글이 `src/hooks.rs`·`src/project_hosts.rs`에서 나와 쓰기 범위 밖이에요 | 후속 카드 |

## Result

- AC-1: met — `cargo test --locked --offline --test project_update_hooks_cli`로 `a_project_with_nothing_yet_names_the_gap_and_prints_no_command`·`a_file_without_hooks_and_without_a_record_says_drifted_like_the_status`·`a_profile_without_claude_leaves_the_hooks_step_unchanged_without_a_command` ok (파일 21 passed). 시험 먼저: 구현 전 RED에서 9개가 실패(exit 101). 프로필의 `claude` 판단·사유 글·`preview.configured`를 바꾼 변이 P1·P8·P10이 모두 KILLED예요. 리더에게 물은 N1("기록도 이동도 없는 프로젝트에 `ready` 제안")에 **B로 답함**(카드 노트 2281): 그 제안은 하지 않고 고정된 AC 13 시험 8개를 그대로 두었고, `project_update_plan_cli` 59 passed로 확인했어요.
- AC-2: met — `cargo test --locked --offline --test go_hooks_migrate` 17 passed(새 4개: `an_install_that_reports_success_without_writing_the_hooks_is_put_back`, `an_installer_that_writes_something_else_is_undone_with_its_record`, `an_installer_that_fails_is_put_back_the_same_way`, `the_read_back_failure_is_a_conflict_not_an_internal_error`). 되읽기 비교를 지운 변이 G1은 3개 시험이, "digest mismatch" 글을 지운 변이 G2는 `conflict` 시험이 KILLED로 잡았어요.
- AC-3: met — `a_hooks_key_that_is_not_the_moai_set_is_not_called_current`, `an_empty_or_null_hooks_key_reads_needs_decision_and_names_the_way_out`, `a_hooks_key_of_the_users_own_with_no_record_says_conflict_and_prints_no_command` ok. "current" 판정을 항상 참으로 바꾼 P2, 빈 키 판정을 `null`만으로 줄인 P3, `drifted` 표시를 지운 P4, 충돌 갈래를 뒤집은 P9가 모두 KILLED예요.
- AC-4: met — `a_status_line_removed_by_hand_is_named_not_called_current` ok(상태줄이 파일에 있는 동안은 "removed"라고 말하지 않는 확인을 더함). 파일 쪽 조건을 지운 변이 P6은 처음에 SURVIVED였고 그 확인을 더한 뒤 KILLED예요. 변이 P7(`unchanged`일 때만 문구를 바꾸는 조건을 지움)은 SURVIVED로 남겼어요: 이 세대의 소유 기록에는 상태줄 키만 들어 있고 그 키가 지워지면 설정 단계는 언제나 `unchanged`라서 시험으로 닿는 상태가 없고(숨은 다른 쓰기가 생기는 날을 위한 방어선), 그래서 시험을 억지로 만들지 않았어요.
- AC-5: met — `an_unreadable_record_names_its_file_and_the_next_step` ok. 기록 오류 글을 감싸는 조건을 뒤집은 P5가 KILLED예요.
- AC-6: met — `an_older_hook_set_held_by_the_file_and_the_record_reads_ready`(같은 쓰기 함수를 이 프로세스에서 불러 적용: `hook_legacy` 관문이 PATH 첫 `moai`를 요구해요, D-5)와 `an_older_hook_set_with_a_users_edit_reads_conflict` ok; 변이 P9가 후자를 KILLED로 잡았어요.
- AC-7: met — `the_undo_order_names_hooks_revert_for_a_hooks_apply_update` ok(`hooks-revert`와 "a first install has nothing to revert"를 확인). 이 글에는 변이를 만들지 않았어요(글 한 줄).
- AC-8: met — 코드 head `9e7d9fb539`에서 `cargo test --locked --offline --lib` 1434 passed와 `--test`로 error_wording_cli·generation_apply_result_cli·go_hooks_migrate·go_project_real_run_cli·go_settings_cli·go_state_classes_cli·go_state_codex_cli·go_state_preview_cli·home_isolation_scan·project_update_hooks_cli·project_update_plan_cli·status_line_cli·update_cli·update_end_to_end_cli·update_route·update_tux_golden 모두 실패 0(필터 없이 돌렸어요; 그 뒤 커밋 `e0c4722d70`은 `project_update_hooks_cli`의 시험 확인 두 줄만 더해서 그 파일을 다시 돌려 21 passed); `cargo clippy --offline --locked --all-targets -- -D warnings`·`cargo fmt --check` exit 0; `spec check` errors 0·warnings 0. 변이는 12개 중 11 KILLED, 1 SURVIVED(P7, 위).
- 남은 것: Codex·OMP 파일이 다른 내용일 때의 `requires preservation review`에 길을 내는 일(N4 뒷부분)은 쓰기 범위 밖(`src/hooks.rs`·`src/project_hosts.rs`)이라 리더에게 넘겨요. SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.
