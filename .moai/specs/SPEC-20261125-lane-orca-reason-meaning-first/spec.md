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

## Result

구현 head 077cc1558278bc292d1241ef7c7feeb4ea71960e(짧게 077cc15582, base 59ac3344cd). 아래는 이 실행에서 본 값이에요.

- 고치기 전: `target/lena/red_screen.log` 에서 새 `lane_screen_cli` 시험 3개가 모두 실패했어요(`test result: FAILED. 0 passed; 3 failed`). `target/lena/red_send.log` 에서 `lane_send_a_send_whose_output_a_child_holds_names_the_program_after_the_meaning` 이 실패했어요(1 passed, 1 failed).
- AC-1 met: `lane_screen_scan_a_held_output_under_a_path_over_200_characters_keeps_its_time_limit` 통과. `target/lena/green_screen.log` 의 `cargo test --locked --test lane_screen_cli -- <시험 5개>` 가 `test result: ok. 5 passed; 0 failed`.
- AC-2 met: `lane_screen_scan_a_hanging_read_under_a_path_over_200_characters_keeps_its_time_limit` 통과(같은 `green_screen.log` 실행).
- AC-3 met: `lane_screen_scan_an_absent_orca_under_a_path_over_200_characters_says_why_it_cannot_run` 통과(같은 `green_screen.log` 실행).
- AC-4 met: `lane_send_a_send_whose_output_a_child_holds_names_the_program_after_the_meaning` 통과. `target/lena/green_send.log` 의 `cargo test --locked --test lane_send_cli -- <시험 5개>` 가 `test result: ok. 5 passed; 0 failed`.
- AC-5 met: `lane_screen_scan_an_orca_child_holding_the_output_does_not_hold_the_scan` 과 `lane_screen_scan_fails_when_the_list_cannot_be_read_and_sends_and_writes_nothing` 이 고치지 않은 채로 통과(`green_screen.log` 의 같은 실행).
- AC-6 met: `lane_send_a_send_whose_output_a_child_holds_says_the_text_may_have_reached_the_lane` 과 `lane_send_sends_failure_is_never_reported_as_sent` 통과(`green_send.log` 의 같은 실행; 같은 실행의 `lane_send_unsafe_a_hanging_orca_is_cut_off_and_nothing_is_sent` 와 `lane_send_unsafe_an_absent_orca_is_an_unreachable_lane` 도 통과). `lane_send_sends_failure_is_never_reported_as_sent` 에 더한 `!message.contains("may have reached the lane")` 줄은 고치기 전에도 통과했어요(REQ-4 를 지키는 줄이고 뮤턴트는 주장하지 않아요).
- AC-7 met: `rg -n '"`\{shown\}`|cannot (run|wait for) `\{shown\}`' src/lane_send.rs` 가 아무것도 찍지 않고 exit 1(구현 에이전트의 실행). `cannot wait for` 문장은 `try_wait` 가 OS 오류로만 실패해서 시험으로 일으킬 수 없고, AC-7 과 코드 읽기로만 덮여요.
- AC-8: not_run in this file: the receipt names the head it ran on and this file is part of that head, so the receipt (`moai check affected --card <base>..<head>`, summary line, step exit codes) is quoted in the card's READY note.
- 문서: `docs/design-review/appendix/operations/lane_send.md` 13행과 `lane_screen.md` 24행은 `did not finish its output` / `did not finish its output within N seconds` 문구만 인용하고 둘 다 그대로 맞아서 문서는 고치지 않았어요.
- not_run: 전체 `cargo test`, 뮤턴트, Windows 확인.
- 이 SPEC 은 `status: draft` 로 두어요(tier 1, 기록된 승인 없음).
