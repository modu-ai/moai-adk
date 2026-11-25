---
id: SPEC-20261125-lane-orca-reason-meaning-first
title: "lane-screen unreachable reasons keep their meaning when the orca path is long"
status: draft
revision: 1
created: 2026-11-25
---

# lane-screen unreachable reasons keep their meaning when the orca path is long

## Intent

`moai lane-screen --scan` 는 닿지 않는 터미널의 이유를 `the terminal could not be read: ` 뒤에 `ProcessOrca::call` 의 오류 글을 160자로 잘라 붙여요. 그 글이 프로그램의 전체 경로로 시작해서, 경로가 길면(`moai check affected` 가 TMPDIR 아래에 폴더를 하나 더 만들어요) 잘림이 뜻(시간 제한 문구)을 지워요. 리더 실측(2026-11-25): 같은 tip 에서 TMPDIR 이 한 단계 길면 `tests/lane_screen_cli.rs` 의 `lane_screen_scan_an_orca_child_holding_the_output_does_not_hold_the_scan` 이 실패하고 짧으면 통과했어요. 이 SPEC 은 뜻을 앞에 두고 경로를 뒤에 둬서, 경로가 길어도 뜻이 남게 해요.

## Requirements

- REQ-1: When a `ProcessOrca` call fails, the error text shall state what happened first (`did not answer within <N> seconds`, `did not finish its output within <N> seconds`, `cannot run the orca program: <kind>` or `cannot wait for the orca program: <kind>`) and name the program path last, as ``: `<program>` ``.
- REQ-2 (REQ-1): IF the program path is longer than the 160 characters that `lane-screen --scan` keeps of an error text, THEN the scan shall still show the meaning of the error in the reason of the `unreachable` row and cut only the path.
- REQ-3: When `lane-send` meets a `did not answer` or `did not finish its output` error of the `orca terminal send` call, it shall add `; the text may have reached the lane` to the error.
- REQ-4 (REQ-3): IF the `orca terminal send` call exits non-zero, THEN `lane-send` shall report its exit code without the may-have-reached hint.

## Acceptance

- AC-1 (REQ-1, REQ-2): a held output at a path over 200 characters gives a row whose reason holds `within 1 seconds` · verify: `cargo test --locked --test lane_screen_cli -- lane_screen_scan_a_held_output_under_a_path_over_200_characters_keeps_its_time_limit`
- AC-2 (REQ-1, REQ-2): a read that never answers at a path over 200 characters gives a row whose reason holds `did not answer within 1 seconds` · verify: `cargo test --locked --test lane_screen_cli -- lane_screen_scan_a_hanging_read_under_a_path_over_200_characters_keeps_its_time_limit`
- AC-3 (REQ-1, REQ-2): an absent program at a path over 200 characters gives a failure that holds `cannot run` and the error kind · verify: `cargo test --locked --test lane_screen_cli -- lane_screen_scan_an_absent_orca_under_a_path_over_200_characters_says_why_it_cannot_run`
- AC-4 (REQ-1): a send whose output a child holds names the program after the meaning · verify: `cargo test --locked --test lane_send_cli -- lane_send_a_send_whose_output_a_child_holds_names_the_program_after_the_meaning`
- AC-5 (REQ-1, REQ-2): the earlier scan tests still pass unchanged · verify: `cargo test --locked --test lane_screen_cli -- lane_screen_scan_an_orca_child_holding_the_output_does_not_hold_the_scan lane_screen_scan_fails_when_the_list_cannot_be_read_and_sends_and_writes_nothing`
- AC-6 (REQ-3, REQ-4): the may-have-reached hint is still added to a send that did not finish and not to a send that exited non-zero · verify: `cargo test --locked --test lane_send_cli -- lane_send_a_send_whose_output_a_child_holds_says_the_text_may_have_reached_the_lane lane_send_sends_failure_is_never_reported_as_sent`
- AC-7 (REQ-1): no sentence of `ProcessOrca::call` starts with the program path · verify: `rg -n '"`\{shown\}`|cannot (run|wait for) `\{shown\}`' src/lane_send.rs` prints nothing and exits 1
- AC-8 (REQ-1, REQ-2): the receipt run on the range holds `lane_screen_cli` and passes · verify: `moai check affected --card <base>..<head>`

## Scope

- write: `src/lane_send.rs`
- write: `tests/lane_screen_cli.rs`
- write: `tests/lane_send_cli.rs`
- write: `.moai/specs/SPEC-20261125-lane-orca-reason-meaning-first/spec.md`

## Plan

1. `tests/lane_screen_cli.rs`: 길이 200자가 넘는 중첩 폴더 도우미와 AC-1~AC-3 테스트를 먼저 쓰고, 고치기 전 head 에서 실패하는 것을 봐요.
2. `src/lane_send.rs`: `ProcessOrca::call` 의 네 문장을 뜻 먼저, 경로 마지막으로 바꿔요.
3. `tests/lane_send_cli.rs`: AC-4 테스트를 더하고, `lane_send_sends_failure_is_never_reported_as_sent` 에 힌트가 없음을 확인하는 한 줄을 더해요(AC-6 의 REQ-4 증거). AC-6 의 테스트 둘이 그대로 통과하는지 봐요.
4. AC-5~AC-7 을 돌리고, `cargo fmt --all` 뒤에 `moai check affected --card <base>..<head>` 영수증을 받아요.
