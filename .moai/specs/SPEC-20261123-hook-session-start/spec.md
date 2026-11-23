---
id: SPEC-20261123-hook-session-start
title: "moai hook session-context: the SessionStart handler (H1)"
status: draft
revision: 1
created: 2026-11-23
---

# moai hook session-context: the SessionStart handler (H1)

## Intent

새 MoAI 훅 설계의 두 번째 조각이에요 (최종 계획서 `~/.moai/reports/moai-adk/hooks-rules-final-plan-2026-11-23.md` §4 H1 과 §7 단계 1-b). 조각 1-a (`SPEC-20261123-hook-entry-skeleton`)가 만든 진입점 `moai hook session-context` 의 빈 처리기를 채워요: 이 체크아웃의 브랜치와 예약 카드를 한 줄로, 압축(`compact`) 뒤에는 다시 읽을 것(카드의 최신 RESUME 노트 위치)을 맥락(`additionalContext`)으로 내요.

처리기는 git 하위 프로세스 없이 `HEAD` 파일을 읽고 원장을 읽기 전용으로 한 번만 열어요. 네트워크는 없고, 어떤 실패에서도 막지 않아요(exit 0, 맥락만 줄어요). 철회된 옛 설계와 옛 핸들러(`src/hook_engine.rs`, `src/hook_guards.rs`, `src/claim_verify.rs`)와 Go 자료는 재사용하지 않아요.

1-a 감사(AUDIT #310)가 미룬 항목도 이 조각이 H1~H3 공통 계약으로 닫아요: N1 (`HOME/.moai` 계열 링크를 따라감), N2 (stderr 가 끊긴 파이프면 exit 101, 4 MiB 를 넘는 stdin 을 읽지 않고 끝냄), D1 (`src/hook_latency.rs` 문서 주석이 문장 중간에서 끊김).

설치는 하지 않아요: settings 와 하네스를 쓰지 않아요. 이 저장소의 운영 훅 4개는 그대로예요. 로컬 빌드로 하네스에 적용하는 일은 사용자 선택 뒤 리더가 해요.

### 이미 있는 것

- `src/hook_entry.rs` (1-a): 입력 한도 4 MiB, 비MoAI 는 침묵, 실패 영수증 `hook-entry.jsonl`(256 KiB 상한), 등록 이름 셋. `session-context` 처리기는 지금 "결정 없음"이에요.
- `moai_cli::ledger::open_read_only_within` 은 `.moai`, `projects`, `<id>` 세 곳의 링크를 거절하고, 읽기 전용으로 열며, 기다림 시간을 인자로 받아요. 이 조각은 새 원장 읽기 코드를 쓰지 않고 이 함수를 써요.
- 예약은 `reservation` 표의 `worktree`, `branch` 열에 있고, RESUME 노트는 `card_note.body` 가 `RESUME ` 로 시작하는 노트예요 (C-310 의 노트가 그 형식).
- 계획서 §4 H1 의 현재 측정값은 중앙값 205.6 ms, p90 263 ms (옛 설계, 노트 4107), 목표는 p50 ≤ 50 ms, p90 ≤ 100 ms 예요.

## Non-goals

- 훅 설치와 settings·하네스 쓰기 (단계 1-e), 이 저장소 하네스에 적용 (사용자 선택 뒤 리더).
- H2 Bash 경계와 H3 쓰기 경계의 판정 (단계 1-c, 1-d). 두 처리기는 "결정 없음"인 채로 남아요.
- 옛 4개 훅의 은퇴 (단계 3).
- `fork` 시작의 처리 (계획서가 보류). 처리기는 `source` 값에 상관없이 같은 첫 줄을 내고, 등록할 때 matcher 가 `fork` 를 뺄 뿐이에요.
- 패닉을 잡는 일. 처리기에 패닉이 없도록 쓰고, 잡는 일은 필요가 보이면 후속이에요.
- Windows 확인. macOS 에서만 검증해요.

## Requirements

<!-- This SPEC adds no check that blocks a tool call: the SessionStart handler only adds context. -->
- REQ-1: When `moai hook session-context` runs in a MoAI project with a JSON object on stdin, the handler shall print one JSON object `{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":<text>}}` on stdout and exit 0 · failure: none — a failure that stops the text from being built is handled under REQ-8.
- REQ-2: The context text shall have at most 600 characters and no control character other than the line feed between its lines · failure: none — the title and the branch are cut to their limits under REQ-3 and REQ-4 before the whole is cut.
- REQ-3: The first line of the text shall be `MoAI: branch <name>` with the name read from the `HEAD` file of the checkout holding the input's `cwd` (else the working folder) and cut to 100 characters, the trailing `…` included, `MoAI: detached <12 hex digits>` for a detached head, or `MoAI: branch unknown` when `HEAD` cannot be read · failure: none — an unreadable `HEAD` reads as unknown.
- REQ-4: When an active reservation of the ledger names the checkout's root as its `worktree` or the branch as its `branch`, the first line shall continue with ` · card C-<id> "<title>" (<state>)`, the title cut to 80 characters (the trailing `…` included) with its whitespace collapsed and the reservation that names the root preferred · failure: none — a ledger that cannot be read is handled under REQ-8.
- REQ-5 (REQ-4): IF no active reservation names the checkout, THEN the first line shall continue with ` · no card reserved here`.
- REQ-6: When the input's `source` is `compact` and a card is reserved here, the text shall have a second line `The newest RESUME note of C-<id> is #<note_id> (moai card show C-<id>)` naming the card's newest note whose body starts with `RESUME `, or `C-<id> has no RESUME note yet (moai card show C-<id>)` when the card has none, and for any other `source` or without a reserved card no second line · failure: none — a ledger that cannot be read is handled under REQ-8.
- REQ-7: The handler shall read the ledger once through one read-only connection, read `HEAD` and `.git` as plain files, start no process (git included), make no network call, read no environment variable other than `CLAUDE_PROJECT_DIR` and `HOME`, and leave every file of the project and of HOME unchanged except the receipt of REQ-8 · failure: none.
- REQ-8 (REQ-7): IF the ledger is absent, damaged, busy for longer than 200 ms or behind a link, THEN the handler shall still print the first line without the card part, exit 0 and record one receipt line (REQ-9) with the reason word `ledger_absent`, `ledger_unreadable`, `ledger_busy` or `ledger_link`.
- REQ-9: IF the hook entry records a receipt and `<HOME>/.moai`, `<HOME>/.moai/projects` or `<HOME>/.moai/projects/<project_id>` is a link, or the receipt file is a link or is not a regular file, THEN the hook entry shall write nothing there and write one stderr line instead, the receipt file being opened without following a link · failure: none (the SPEC 1-a rule that no receipt is written without an existing project folder stays).
- REQ-10: The hook entry shall exit 0 when its standard output or standard error cannot be written (a closed pipe) · failure: none.
- REQ-11: IF the standard input of a registered name is larger than 4 MiB, THEN the hook entry shall read it to its end before it exits, so the writer is not cut off, and handle it as the SPEC 1-a "stdin_too_large" failure · failure: none.
- REQ-12: When `moai hook latency` measures a name, the hook latency tool shall register a ledger in its temporary HOME and reserve one card for the measured project folder, so the `moai_project` case runs the card path of REQ-4 · failure: none — the tool still judges no number.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): in a temporary MoAI project that is a git checkout with a reservation on its worktree, `session-context` with `source` `startup`, `resume` and `clear` prints exactly one JSON object of the REQ-1 shape whose text is one line `MoAI: branch <b> · card C-<id> "<title>" (active)`, exits 0 with stderr empty, and `bash-boundary` and `write-boundary` stay silent · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac1_/)'`
- AC-2 (REQ-4, REQ-5): a reservation naming the root, one naming the root only through a link or a `..` spelling, an input `cwd` that is not absolute (ignored, the working folder decides), one naming only the branch, a released one, one of another worktree and branch, and none each give the expected card part (the root-naming reservation wins over a branch-only one) · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac2_/)'`
- AC-3 (REQ-6): with `source` `compact` the text gets the second line with the highest note id of the notes starting `RESUME ` (a later note of another kind and an earlier RESUME note are ignored), or `no RESUME note yet`, and `startup`, `resume`, `clear` and a compact start without a reserved card give one line and no second line · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac3_/)'`
- AC-4 (REQ-2, REQ-3, REQ-4): a title of 80 characters is shown whole and one of 81 is cut with `…`, a branch name of 100 characters is whole and one of 101 is cut, multi-byte characters are counted as characters and never split, line feeds, tabs and escape characters in the title or the branch never reach the text, and with every field at its limit the text has at most 600 characters · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac4_/)'`
- AC-5 (REQ-3, REQ-7): the branch is read from a plain checkout, a linked worktree (a `.git` file with `gitdir:`), the input's `cwd` below the root and the working folder when the input has none, a detached head gives 12 hex digits, a missing `HEAD` and a `HEAD` that is a FIFO give `branch unknown` without hanging, and a `HEAD` of 1024 bytes is read while 1025 bytes reads as unknown · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac5_/)'`
- AC-6 (REQ-7): with a stand-in `git` that leaves a marker first on `PATH` no marker appears, the files of the project and of HOME (the ledger bytes included, and no journal file) are the same before and after, `src/hook_session_start.rs` names no process-start, network or settings word and no environment variable, and `src/hook_entry.rs` still reads exactly `CLAUDE_PROJECT_DIR` and `HOME` · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac6_/)'`
- AC-7 (REQ-8): an absent ledger, a damaged ledger (bytes left as they were), a ledger held under an exclusive lock and a missing project folder each print the branch part, exit 0 and add one receipt line with the reason word (the missing project folder writes one stderr line instead), and the locked case answers within 3 seconds · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac7_/)'`
- AC-8 (REQ-8, REQ-9): with `.moai`, `projects` or the project folder under HOME a link to a folder that holds a valid ledger and a receipt file, `session-context` shows no card and writes nothing in the target, a bad input of `bash-boundary` writes nothing in the target, each with one stderr line, and a receipt file that is a link or a FIFO is refused without hanging and its target stays as it was · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac8_/)'`
- AC-9 (REQ-10): with stderr closed (a pipe whose reader is gone) a bad input without a project folder exits 0, and with stdout closed a good input exits 0 · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac9_/)'`
- AC-10 (REQ-11): a JSON object of 16 MiB written to a registered name is written completely (the writer sees no broken pipe), exits 0 and adds exactly one `stdin_too_large` receipt line · verify: `cargo nextest run --test hook_session_start_cli -E 'test(/^ac10_/)'`
- AC-11 (REQ-12): `hook latency --samples 3` against a stand-in that runs the built binary and keeps its stdout shows the card line in every `moai_project` sample and none in the `not_moai_project` samples, and the tool still removes only the folder it made · verify: `cargo nextest run --test hook_latency_cli -E 'test(/^ac11_/)'`
- AC-12 (REQ-1): the entry's earlier answers are unchanged for the other names: `bash-boundary` and `write-boundary` print nothing for good input, unknown names exit 0, and the 1-a tests still pass · verify: `cargo nextest run --test hook_entry_cli --test hook_latency_cli`
- AC-13 (REQ-1…REQ-12): the existing hook commands and the help layout keep their behavior · verify: `cargo nextest run --test hook_dispatch_cli --test hook_mode_cli --test hook_legacy_cli --test lane_stop_hook_cli --test diff_go --test help_layout`
- AC-14 (REQ-1…REQ-12): the SPEC itself passes the spec check · verify: `moai spec check SPEC-20261123-hook-session-start`

## Scope

- write: `.moai/specs/SPEC-20261123-hook-session-start/**`, `src/hook_session_start.rs` (new), `src/hook_entry.rs`, `src/hook_latency.rs`, `tests/hook_session_start_cli.rs` (new), `tests/hook_entry_cli.rs` (the `session-context` expectation of AC-1 of 1-a only), `tests/hook_latency_cli.rs`, `tests/areas.toml` only if a new first word needs a line (`hook` is listed).
- read-only: `src/ledger/**` and `src/profile.rs` (called, not changed), `src/hook_engine.rs`, `src/hook_guards.rs`, `src/claim_verify.rs`, `src/hooks.rs`, `tools/hook_latency.py`.
- forbidden: `src/main.rs` (card C-318 holds it, leader decision of 2026-11-23: the handler is attached through `src/hook_*` files only), `src/lib.rs`, `src/help.rs`, any settings file (`.claude/**` of the home and of the project), the installed hooks of this clone, the real HOME and its `.moai/projects` folder, any network call. Tests use temporary HOMEs and folders only.

## Plan

1. (this slice) SPEC, status draft; the user's approval is not written by a lane.
2. Tests first (RED): `tests/hook_session_start_cli.rs` and the changes of the two 1-a test files, run against the 1-a binary; each must fail for the stated reason (an empty handler prints nothing, a link is followed, a closed pipe exits 101, a short read cuts the writer off).
3. `src/hook_session_start.rs`: `HEAD` reader, reservation and RESUME query, text builder with the limits. `src/hook_entry.rs`: declares the new module with `#[path]`, calls the handler with the ledger error mapped to a receipt reason, opens the receipt without following a link, drains the rest of stdin in `read_input`, and writes its own answer ignoring write errors (so `src/main.rs`, which prints what `run` returns, stays untouched).
4. `src/hook_latency.rs`: register a ledger and reserve one card in the temporary HOME; mend the doc comment that stops mid-sentence (D1).
5. Mutants: remove each new guard in turn and show the test that fails; one `guard-mutant:` line per new guard test goes into the READY note.
6. Dangerous-path run in a throwaway: the real debug binary with a `.git` that is a FIFO, a `HEAD` of 5 MiB, a ledger under a lock, `.moai` as a link to a folder with a ledger, a closed stderr.
7. Measurement run: `hook latency --name session-context` on a release build, p50 and p90 recorded with the build label and the machine load.
8. Targeted tests, one `check affected`, READY with head, base and the receipt.

## Risks

- R-1: The handler runs at every session start, so a slow ledger open delays every session. The wait is cut to 200 ms (D-2) and the result is measured (AC-11, plan step 7); the plan's target is judged on a release build.
- R-2: The text goes into the model's context. A card title or a branch name is text someone typed; it is shortened and stripped of control characters, but it is still data the model reads, so nothing of it is ever run and the note's body is never copied (D-5).
- R-3: A reservation matches by the `worktree` text or the branch name. A lane that works in another folder than it reserved shows `no card reserved here`; the ledger, not this line, decides ownership.
- R-4: The exec form of the hook settings (`args`) has no stated minimum Claude Code version (see 1-a Sources). This slice installs nothing, so the check belongs to 1-e.
- R-5: Wall-clock numbers depend on machine load; the tool prints none of it, so the load is noted beside the numbers.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The context text is English. | `CLAUDE.md` writes instruction text in English; the text is read by the model, not the user. | Change the strings. |
| D-2 | The ledger wait is 200 ms. | A session start must not hang on a busy ledger; the status line's reader uses a short wait for the same reason. | Change the constant. |
| D-3 | An active reservation is one whose status is `held`, `held_for_review` or `expired_candidate` (the statuses of the one-active index). | The ledger schema's `reservation_one_active` index. | Narrow the list. |
| D-4 | The limits are: text 600 characters (the plan), title 80, branch 100, `HEAD` file 1024 bytes. | Plan §4 H1 (600); the rest keep one card line short. | Change the constants. |
| D-5 | The `compact` line names the note and the command; it never copies the note. | A note holds free text and can be long; the model reads it with `moai card show`. | Add a short excerpt later. |
| D-6 | The H1 code is a binary-side module `src/hook_session_start.rs`, declared inside `src/hook_entry.rs` with `#[path = "hook_session_start.rs"]`; `src/main.rs` and `src/lib.rs` are not changed. | Leader decision 2026-11-23: C-318 owns `src/main.rs`; the 1-a scan of `src/hook_entry.rs` stays valid. | Move it to the library in a later slice. |
| D-9 | `hook_entry::run` writes the answer to stdout and stderr itself, ignoring write errors, and returns an empty answer. | `src/main.rs` prints a returned answer with `println!`, which panics on a closed pipe, and cannot be edited here. | Return the answer again when `src/main.rs` prints it safely. |
| D-7 | The ledger is read through `ledger::open_read_only_within`, the receipt through `ledger::folders::refuse_links` plus an open with `O_NOFOLLOW`. | Rule 3: the cheapest source first; both exist and are tested. | Not applicable. |
| D-8 | The old `session-start` handler of `hook_engine` is not read or ported. | User decision of 2026-11-23 (plan §0-1). | Not applicable. |
| D-10 | The second line of a `compact` start is a statement of fact (`The newest RESUME note of C-<id> is #<note_id> …`), not an order. | The audit of this slice (N2, leader decision 2026-11-23) read the official hooks page: context added at a session start is better read as information than as an instruction. | Change the two strings. |
| D-11 | The stdin drain of REQ-11 keeps no upper bound in this slice. | A bound would cut the writer off again above it, which REQ-11 forbids; the bounded read (a time limit) belongs to the install slice 1-e, where the host's hook timeout is the limit (audit N4, optional). | Add a drain limit with its own REQ. |

### Sources

- Plan: `~/.moai/reports/moai-adk/hooks-rules-final-plan-2026-11-23.md` §4 H1, §7 step 1-b, read whole in this session.
- Official hooks documentation, https://code.claude.com/docs/en/hooks, fetched by a research subagent with WebFetch on 2026-11-23 (clock read 2026-11-23T08:16:20Z by this session). The page has 293,973 characters; the SessionStart section lay in a range the tool returned only as a model summary, so the points below are the tool's rendering, not raw bytes. Verbatim rows: the SessionStart matcher values "`startup`, `resume`, `clear`, `compact`, `fork`"; "Claude Code adds plain-text stdout as context that Claude can see and act on."; "Claude never sees stderr from a hook that exits 0."; default command timeout 600 seconds; "SessionStart runs on every session, so keep these hooks fast." Summary only: the common input fields (`session_id`, `cwd`, `hook_event_name` …) and the SessionStart field `source`; the JSON output `{"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "..."}}`; `additionalContext` and plain stdout capped at 10,000 characters (more is saved to a file and only the first 2,000 characters are shown); exit 2 does not block SessionStart and shows stderr to the user only. Unverified: the exact wording of those summarised lines and the parts of the page past character 179,900. The 600-character text stays far below the cap.

## Result

Status: implemented on the code slice. The SPEC stays `draft`: approval is the user's pick, not a lane's. Nothing is installed, no settings file is read or written, the installed hooks of this clone are unchanged, and `src/main.rs` and `src/lib.rs` are not in the diff.

- AC-1: met — `hook_session_start_cli` test `ac1_a_start_names_the_branch_and_the_reserved_card_in_one_json_line`
- AC-2: met — `ac2_the_card_part_follows_the_reservation_of_this_checkout` (eight reservation shapes, among them a link and a `..` spelling of the root with another branch, the root-over-branch order, and an input `cwd` that is not absolute)
- AC-3: met — `ac3_a_compact_start_names_the_newest_resume_note` (the second line is the fact form of D-10)
- AC-4: met — `ac4_limits_hold_at_the_edge_and_control_characters_never_reach_the_text` (title 80/81, branch 100/101, single- and multi-byte, controls, the 600-character bound with every field at its limit)
- AC-5: met — `ac5_the_branch_is_read_from_head_files_only` (plain checkout, linked worktree with absolute and relative `gitdir`, cwd below the root and no cwd, detached head, `HEAD` missing, a folder, a FIFO, `.git` a FIFO, `gitdir` naming nothing, `HEAD` of 1024, 1025 and 5 MiB bytes)
- AC-6: met — `ac6_no_process_no_write_and_two_variables` (a stand-in `git` first on `PATH` never starts; project and HOME unchanged; the source scan and the two-variable scan)
- AC-7: met — `ac7_an_unreadable_ledger_gives_the_branch_and_one_receipt_line` (absent, damaged, exclusive lock, no project folder; the locked case answered within 3 seconds)
- AC-8: met — `ac8_links_under_home_are_not_followed` (a link at `.moai`, `projects` and the project folder; a receipt file that is a link and a FIFO); the 1-a test `ac5_a_receipt_link_is_not_followed_and_a_hostile_project_id_is_refused` still passes
- AC-9: met — `ac9_a_closed_pipe_still_exits_zero` (stderr closed, stdout closed, both closed)
- AC-10: met — `ac10_an_input_over_the_limit_is_read_to_its_end` (16 MiB written completely)
- AC-11: met — `hook_latency_cli` test `ac11_the_project_case_runs_the_card_path`
- AC-12: met — `hook_entry_cli` and `hook_latency_cli` pass; the 1-a test `ac1_…` no longer lists `session-context` (it now answers context) and `ac4_the_stdin_limit_…` takes the context answer at the limit for that name
- AC-13: met — one `cargo nextest run` over `hook_dispatch_cli`, `hook_mode_cli`, `hook_legacy_cli`, `lane_stop_hook_cli`, `diff_go`, `help_layout` and the three new-or-changed hook targets: 111 passed, 0 failed (exit 0)
- AC-14: met — `moai spec check SPEC-20261123-hook-session-start` answers 0 errors and 0 warnings

Against the 1-a binary, all 11 new tests failed (commit `859ebca1fe`): the empty handler printed nothing, a link under HOME was followed, a closed pipe gave exit 101, an input over the limit cut the writer off, and the latency tool reserved no card. After the code (commit `708fae71b3`): the three hook targets 26 of 26 pass, `cargo clippy --offline --locked --all-targets -- -D warnings` exits 0.

Mutants (each edits one guard of the committed code, then runs the test of that guard; "killed" means exit 100): wrong `hookEventName`, root match off, RESUME prefix loosened to "contains", title limit 81, `HEAD` limit 1025, a `git` process started in `read_head`, ledger wait 5 s, receipt link refusal removed, panicking output, stdin drain removed, no card reserved by the latency tool, receipt opened without `O_NOFOLLOW`, `HEAD` opened without `O_NONBLOCK` (the FIFO case then hangs until the test's 20-second kill), second line for every `source` — 14 of 14 killed. The first run of the `git` mutant was killed only by the source scan because the stand-in's `touch` was not found on its `PATH`; the stand-in now calls `/usr/bin/touch` and the marker kills it.

Repair after the audit (verdict FAIL on head `211071dfd3`, blockers B1 and B2, notes N2 and N4): B1 and B2 are test gaps only. `ac2_` now has a reservation whose `worktree` is a link to the checkout and one spelled with `..`, each with another branch, and an input with a relative `cwd` next to a reservation that names only the folder. Two mutants, each run against `ac2_` only, were both killed (exit 100): `|| false && fs::canonicalize(stored)` in `same_folder` (the link case fails) and `.filter(|_c| true)` on the input `cwd` (the relative-`cwd` case fails). N2 changed the strings (D-10; RED first: `ac3_` failed on the old strings, then passed). N4 is not done by decision D-11. The code change of the repair is the two strings of the second line.

Dangerous path: every hostile shape is a test of the real debug binary in a temporary checkout and HOME (FIFO and folder as `HEAD`, `.git` as a FIFO, a 5 MiB `HEAD`, a locked ledger, `.moai`, `projects` and the project folder as links to a folder holding a valid ledger, a closed stderr and stdout, a 16 MiB stdin); a separate shell run of the same shapes was refused by the session's worktree guard and was not made.

Measurement (`moai hook latency --name session-context`, 2026-11-23, load average 130 at the time): release build, 200 samples, `not_moai_project` p50 6.563 ms, p90 8.877 ms, max 41.096 ms; `moai_project` (git checkout, registered ledger, one reserved card) p50 8.187 ms, p90 9.570 ms, max 13.113 ms. Debug build, 100 samples, load 47: p50 3.635 ms, p90 3.826 ms. The plan's targets (p50 ≤ 50 ms, p90 ≤ 100 ms) are met in the project case. The tool judges no number; the comparison is the reader's.

Not done here: a real Claude Code session loading the entry (nothing is installed), a Windows run, the exec-form version check, and a run against the real HOME ledger (read-only, left for the leader's application after the user's pick). Deferred by decision: a drain limit for a stdin over 4 MiB (D-11, audit N4); catching a panic of the handler; N3, N5 and N7 of the 1-a audit (optional there). The `check affected` receipt of the final head is on the card's READY note.
