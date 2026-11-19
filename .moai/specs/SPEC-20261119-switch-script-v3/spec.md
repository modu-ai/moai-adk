---
id: SPEC-20261119-switch-script-v3
title: "Name-switch script v3: bundle by action, every Go word checked against the Rust list (C-136)"
status: draft
revision: 1
created: 2026-11-19
---

# Name-switch script v3: bundle by action, every Go word checked against the Rust list (C-136)

## Intent

카드 C-136(p1). 사용자가 방금 이름 바꾸기 스크립트 v2 로의 교체를 골랐는데, v2 도 `bundle` 을 Go 전용 낱말 줄에 그대로 둬요. 그래서 `moai bundle build`·`verify`·`check`·`ready`·`body`(리더가 쓰는 M-Flow 묶음, Rust 에만 있음)가 Go 로 가서 Go 의 `bundle`(`add`·`remove`·`list`뿐)이 받아요. 실사용이 막혀요(설치 경로 `moai-cli` 로만 돼요).

v3 는 v2 와 낱말 하나만 달라요: `bundle` 을 `spec`·`session` 처럼 **동작 단위로 나눠요** — Go 에만 있는 `add`·`remove`·`list` 는 Go, 나머지(`ready` `check` `body` `build` `verify`, 낱말만 칠 때, 도움말)는 Rust. 그리고 Go 로 보내는 낱말 63개 전부를 지금 Rust 명령 목록과 대조해 표로 두고, Rust 가 나중에 같은 낱말을 가지면 시험이 실패하게 해요.

`moai update` 의 전환 부분(`src/update/switch.rs`)은 지금 "v1 → v2 하나"만 알아요. 이미 v2 를 적용한 컴퓨터를 v3 로 올릴 수 있도록, 알려진 옛 스크립트(v1, v2)를 내장 v3 로 바꾸는 미리보기·적용·되돌리기로 넓혀요.

## Non-goals

- 실제 HOME 의 `~/.moai-cli/path/moai`·`~/go/bin/moai` 교체 — 보호된 단계예요. `moai update` 미리보기의 switch 확인으로 사용자가 해요. 이 SPEC 은 코드와 시험만 내요.
- 표의 다른 낱말을 Rust 로 옮기는 일 — 대조해 보니 Rust 와 겹치는 Go 전용 낱말은 `bundle` 하나뿐이었어요(Result 참고).
- Go 쪽 `bundle add|remove|list` 의 이전, `hook`·`statusline`·`mcp-server`·Go 세대 프로젝트 경로의 변경 — 그대로예요.
- `src/help.rs` — `--confirm-switch` 도움말이 버전을 말하지 않아서 바꿀 곳이 없어요(그 파일은 다른 카드가 쥐고 있어요).

## Requirements

- REQ-1: When the v3 script is run with `bundle` alone, or with `bundle` and a second word that is not one of `add`, `remove` or `list`, the script shall hand the call, with every argument, to the Rust build without looking for the Go build. · failure: none
- REQ-2: When the v3 script is run with `bundle` and a second word that is one of `add`, `remove` or `list`, the script shall hand the call, with every argument, to the Go build.
- REQ-3 (REQ-2): IF the Go build is not found when `bundle add`, `bundle remove` or `bundle list` is run, THEN the script shall print one line that names `bundle` and the path searched and exit with status 1.
- REQ-4: The v3 script shall send to the Go build exactly the words that v2 sent there, less `bundle`, with every other word, launcher word, unknown word and empty command line routed as v2 routes it.
- REQ-5: The test suite shall hold a table of every word of the v1 script's Go list with its route in v3 (Go, Rust or split by action), failing when the table differs from the v1 list, from the v3 script's Go line or from its split branch.
- REQ-6: When a word that the table sends to the Go build whole is a command word of the built Rust program, the test suite shall fail and name that word. · failure: none
- REQ-7: When `moai update` previews a `path/moai` that is the known v1 or v2 script, the update shall offer replacing it with the v3 script, with a confirm token that names the script it replaces, keeping the old file as `moai.before-<UTC>` byte for byte.
- REQ-8 (REQ-7): IF the `path/moai` given to the apply is not a known older script any more, or its token is not the preview's, THEN the update shall write nothing.
- REQ-9: When `moai update --revert-switch <stamp>` is run, the update shall put back the saved copy only if it is byte-equal to the v1 or the v2 script, and only over a `path/moai` that is the v2 or the v3 script. · failure: none
- REQ-10: While `path/moai` is the v3 script, a link, absent or an unknown file, the update shall offer no switch.
- REQ-11: The work shall write no file in the real HOME: the tests run only in temporary folders with fake builds.

## Acceptance

- AC-1 (REQ-1, REQ-2): `ready`, `check`, `body`, `build`, `verify`, the bare word and the help forms reach the fake Rust build; `add`, `remove`, `list` reach the fake Go build; arguments arrive unchanged; only the first two words decide.
  verify: `cargo test --locked --offline --test switch_script_v3 v3_sends_the_rust_bundle_actions_and_its_help_to_rust`
  verify: `cargo test --locked --offline --test switch_script_v3 v3_keeps_the_go_bundle_actions_on_go`
  verify: `cargo test --locked --offline --test switch_script_v3 v3_routes_bundle_by_the_first_two_words_only`
- AC-2 (REQ-3): a missing Go build fails only the Go actions of `bundle`; the called build's exit status is the script's.
  verify: `cargo test --locked --offline --test switch_script_v3 v3_without_a_go_build_fails_only_for_the_go_bundle_actions`
- AC-3 (REQ-4): the Go list of v3 is the list of v2 less `bundle`; spec, session and bundle are the split words; every other call gives the same result in v2 and v3.
  verify: `cargo test --locked --offline --test switch_script_v3 v3_changes_exactly_one_word_of_v2`
- AC-4 (REQ-5, REQ-6): the table is the v1 list word for word (63 = 59 Go, 3 split, 1 moved) and agrees with the v3 script; the Rust command list is read from the built program; no word sent to Go whole is a Rust command.
  verify: `cargo test --locked --offline --test switch_script_v3 the_table_lists_every_word_of_the_v1_list_and_agrees_with_the_v3_script`
  verify: `cargo test --locked --offline --test switch_script_v3 the_rust_command_list_is_read_from_the_built_program`
  verify: `cargo test --locked --offline --test switch_script_v3 no_word_sent_to_go_whole_is_a_rust_command`
- AC-5 (REQ-7, REQ-8, REQ-10): a v1 and a v2 script are each offered v3 with a token of their own; a token made for v1 is refused for v2; the apply keeps the old script byte for byte, installs v3 at mode 755 and reads it back; v3 is not offered again.
  verify: `cargo test --locked --offline --test update_switch_layer plan_offers_v3_for_the_known_v1_script`
  verify: `cargo test --locked --offline --test update_switch_layer plan_offers_v3_for_the_known_v2_script_with_its_own_token`
  verify: `cargo test --locked --offline --test update_switch_layer apply_saves_v1_and_installs_v3`
  verify: `cargo test --locked --offline --test update_switch_layer apply_over_v2_keeps_v2_byte_for_byte_and_the_revert_puts_it_back`
  verify: `cargo test --locked --offline --test update_switch_cli a_v2_script_is_offered_v3_and_the_revert_puts_v2_back`
- AC-6 (REQ-9): the revert restores a saved v1 or v2 over v2 or v3, refuses a saved v3 or any other file, and refuses a `moai` that is neither.
  verify: `cargo test --locked --offline --test update_switch_layer revert_over_v3_restores_a_saved_v1_and_still_refuses_a_saved_v3`
  verify: `cargo test --locked --offline --test update_switch_layer a_v2_script_that_was_applied_by_an_earlier_build_still_reverts_to_v1`
  verify: `cargo test --locked --offline --test update_switch_layer revert_puts_v1_back_and_refuses_anything_else`
- AC-7 (REQ-11): neighbours (targeted; the full run is cited from the PR's CI run).
  verify: `cargo test --locked --offline --test update_switch_layer --test update_switch_cli --test switch_script_v3 --test switch_script_v2`
  verify: `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `.moai/specs/SPEC-20261119-switch-script-v3/**`, `install/moai-switch.sh`, `src/update/switch.rs`, `src/update.rs` (one doc comment), `tests/switch_script_v3.rs`, `tests/update_switch_layer.rs`, `tests/update_switch_cli.rs`
- read-only: `.moai/specs/SPEC-20261119-switch-script-v2/preview/moai.v1` and `moai.v2` (the two older scripts, byte copies), `src/main.rs` (the Rust command list)
- forbidden: any write in the real HOME (the name-switch script of the state folder, the Go build's folder, the MoAI state folders), starting the Go build outside an empty temporary HOME, deleting anything

## Plan

1. SPEC, then the script (`bundle` joins the split branch, leaves the Go line), `switch.rs` (known older scripts v1, v2; the built-in v3), the tests.
2. Checks, mutants, a read-only audit by another agent (tier 2: the install path; the leader audits).

## Risks

- A Go-generation project may call `moai bundle ready`-like words expecting Go: Go's `bundle` has only `add`, `remove`, `list` (read from `moai.old bundle --help` in an empty temporary HOME), so no Go caller of the moved actions can exist.
- The table is only as true as the Rust list: the test reads `moai help --all` of the built program, so a word Rust gains later fails the table test instead of staying silently on Go.
- The revert over a `moai` that is v2 is allowed for a saved v1: that is the undo an earlier build of this update offered, and it must keep working.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A known older script (v1, v2) is replaced by the built-in v3 in one step; v1 does not go through v2 | Fewer steps for the user; the preview names the script it replaces | Revert restores the saved copy |
| D-2 | The token covers the file, the old script's digest and the v3 digest | A preview made for v1 can never apply to v2 | none needed |
| D-3 | `src/help.rs` is not changed | The `--confirm-switch` help names no version; another card holds the file | none |

## Result

2026-11-19(UTC). 구현자 에이전트 agent-1, 감사는 tier 2(설치 경로)라 리더가 해요.

- 조사: 설치된 빌드 4.2647.9 의 `moai-cli help --all`(첫 열 낱말)과 v2 스크립트의 Go 전용 줄을 대조해 보니 Rust 에도 있는 낱말은 `bundle` 하나뿐이에요(`gate`·`verify`·`status`·`worktree`·`profile`·`plan`·`version`·`doctor` 같은 낱말은 Rust 에 그 이름이 없어요; Rust 에 `moai` 라는 명령은 없고 `help --all` 첫머리의 사용법 줄에만 글자가 나와요). Go 의 `bundle` 동작은 `moai.old bundle --help` 를 빈 임시 HOME 에서 읽어 `add`·`remove`·`list` 셋임을 확인했고, 그 HOME 에는 아무 것도 만들어지지 않았어요.
- 표(`tests/switch_script_v3.rs` `TABLE`): v1 목록 63 = Go 에 남는 59 + 동작 단위로 나누는 3(`spec` `session` `bundle`) + Rust 로 옮긴 1(`statusline`, v2 에서 이미). v3 는 v2 보다 `bundle` 하나만 Go 줄에서 빠져 분기로 가요(`v3_changes_exactly_one_word_of_v2`가 v2 와 v3 를 나란히 돌려 다른 낱말 59 + 런처·모르는 낱말·빈 줄·spec·session·statusline·mcp-server 호출이 같은 결과임을 확인).
- AC-1~AC-6 의 시험 실행(이 head, `cargo test --locked --offline`): `--test switch_script_v3` 8 통과, `--test update_switch_layer` 15 통과(기존 11 + 새 4), `--test update_switch_cli` 9 통과(기존 8 + 새 1). `switch_script_v2` 는 v2 미리보기 파일을 읽는 시험이라 그대로예요.
- 변이 10개(한 번에 하나, `git restore` 로 되돌림, 모두 시험이 잡음): 스크립트 S1 `bundle` 을 Go 줄에 되돌림, S2 `bundle:add` 삭제, S3 분기에서 `bundle` 삭제, S4 `bundle:list` 삭제; 표 T1 `spec` 을 Go 로; switch.rs R1 v2 를 모르는 파일로, R2 토큰이 옛 스크립트를 무시, R3 되돌리기가 v1 만, R4 되돌리기가 `moai` 를 가리지 않음, R5 보관본에 틀린 바이트.
- 되돌리기 규칙: 이전 빌드가 v1 → v2 를 적용해 남긴 `moai.before-<stamp>`(v1)는 `moai` 가 v2 여도 v3 여도 되돌려져요(`a_v2_script_that_was_applied_by_an_earlier_build_still_reverts_to_v1`).
- 못 한 것·남은 것: 실제 HOME 의 `path/moai`(지금 v2, sha256 36dd242bf133…)는 바꾸지 않았어요 — `moai update` 미리보기에 v2 → v3 가 뜨고 사용자가 `--confirm-switch` 로 확인해요(카드 기준 3의 실제 적용 부분).
