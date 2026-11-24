---
id: SPEC-20261123-hook-entry-skeleton
title: "moai hook entry point skeleton and hook latency tool"
status: draft
revision: 1
created: 2026-11-23
---

# moai hook entry point skeleton and hook latency tool

## Intent

새 MoAI 훅 설계의 첫 조각이에요 (최종 계획서 `~/.moai/reports/moai-adk/hooks-rules-final-plan-2026-11-23.md` §4 공통 사항과 §7 단계 1-a). Claude Code 설정이 부를 Rust 진입점 `moai hook <name>` 의 뼈대와, 훅 지연을 재는 Rust 측정 도구 `moai hook latency` 만 더해요 (같은 `hook` 낱말 아래의 예약 이름이라 `src/help.rs`, `src/lib.rs` 와 help 골든을 건드리지 않아요).

뼈대는 다음만 해요: 표준 입력 JSON 을 한 번 읽고, MoAI 프로젝트가 아니면 출력 없이 exit 0, 모르는 이름도 exit 0, 엔진 실패는 영수증 한 줄(상한 넘으면 경고), 네트워크와 Orca 환경은 읽지 않아요. 이름 세 개(`session-context`, `bash-boundary`, `write-boundary`)는 H1, H2, H3 의 자리만 잡고, 이 조각의 핸들러는 모두 "결정 없음"이에요. 판정 코드는 1-b 이후 조각이 더해요. 철회된 옛 설계와 옛 핸들러(`src/hook_engine.rs`, `src/hook_guards.rs`, `src/claim_verify.rs`)와 Go 자료는 재사용하지 않아요.

설치는 하지 않아요: settings 를 쓰지 않고, 이 저장소의 운영 훅 4개(`hook-dispatch`, `claim-verify`, `hook-receipt`)는 그대로예요.

### 이미 있는 것

- `moai hook <event>` 는 지금 Go 세대 훅 호출을 조용히 받는 호환 경로예요 (`src/main.rs` `go_hook_main`: stdout 비움, stderr 한 줄, exit 0; `tests/diff_go.rs` 가 `hook pre-tool` 로 고정). 새 진입점은 등록된 세 이름만 맡고, 그 밖의 이름은 이 호환 경로를 그대로 지나가요.
- 프로젝트 상태 폴더 `<HOME>/.moai/projects/<project_id>/` 는 `register` 만 만들어요 (`src/ledger/location.rs`). 훅은 이 폴더를 만들지 않아요.
- `tools/hook_latency.py` 는 옛 `hook-dispatch` 용 Python 측정기예요. 이 SPEC 은 새 진입점용 Rust 측정기를 더하고 그 파일은 건드리지 않아요.
- `src/help.rs` 와 `src/lib.rs` 는 다른 카드가 예약 중이라 (2026-11-23 `card take` 충돌), 새 모듈 두 개는 `src/main.rs` 안의 `mod` 로 두는 binary 쪽 모듈이에요. 이후 카드가 lib 로 옮기거나 `hook-latency` 를 help 에 올릴 수 있어요.

## Non-goals

- 훅 설치와 settings 쓰기 (단계 1-e). 이 조각은 `.claude/settings.json` 을 읽지도 쓰지도 않아요.
- H1 맥락 출력, H2 Bash 경계, H3 쓰기 경계의 판정 (단계 1-b, 1-c, 1-d).
- 옛 훅 핸들러의 은퇴와 `claim-verify`·`hook-receipt` 제거 (단계 3).
- 측정 숫자에 합격/불합격을 거는 일. 목표(p50 ≤ 50 ms, p90 ≤ 100 ms)는 1-b 이후 조각의 합격 기준이에요.
- Windows 에서의 실행 확인. 이 조각의 검증은 macOS 에서만 해요.

## Requirements

<!-- This SPEC adds no check that blocks a tool call: every handler answers no decision. -->
- REQ-1: When `moai hook <name>` starts, the hook entry shall read standard input once, bounded to 4 MiB, before it decides anything · failure: none — input that cannot be read or is too large is the failure REQ-5 handles.
- REQ-2: When the name is not `session-context`, `bash-boundary` or `write-boundary`, the hook entry shall leave stdout empty and exit 0 · failure: none — an unknown name changes nothing, and a Go-generation call such as `hook pre-tool` keeps its one stderr line.
- REQ-3: When the project folder (`CLAUDE_PROJECT_DIR`, else the working folder) has no regular file `.moai/project.yaml`, the hook entry shall print nothing on stdout and stderr, write nothing and exit 0 · failure: none — an unreadable or non-file `project.yaml` reads as no project.
- REQ-4: When the name is registered and the project folder is a MoAI project, the hook entry shall hand the parsed input to that name's handler, which in this slice answers no decision by printing nothing and exiting 0 · failure: none — a handler that fails is the engine failure REQ-5 handles.
- REQ-5 (REQ-4): IF the input of a registered name in a MoAI project is larger than 4 MiB, is not UTF-8, is not JSON or is not a JSON object, THEN the hook entry shall leave stdout empty, exit 0 and record one engine-failure receipt line.
- REQ-6 (REQ-5): IF the engine-failure receipt of REQ-5 is recorded, THEN the hook entry shall append one line with the name, a reason word and a time, never any input bytes, to `<HOME>/.moai/projects/<project_id>/hook-entry.jsonl` only when that project folder already exists, and write one stderr line instead when it does not.
- REQ-7 (REQ-6): IF `hook-entry.jsonl` is already 256 KiB or larger, THEN the hook entry shall append nothing, write one stderr warning line naming the limit and exit 0.
- REQ-8: The hook entry shall make no network call, start no process, read no environment variable other than `CLAUDE_PROJECT_DIR` and `HOME`, and write no settings file.
- REQ-9: When `moai hook latency --bin <path>` runs, the hook latency tool shall start `<path> hook <name>` `--samples` times for each of two cases, a folder that is not a MoAI project and a MoAI project with a temporary HOME, and print the p50, p90 and maximum in milliseconds per case with the sample count and a build profile label · failure: none — a sample whose process fails to start stops the run under REQ-10.
- REQ-10 (REQ-9): IF the binary cannot be started, `--samples` is outside 1 to 1000, or `--name` is not a registered name, THEN the hook latency tool shall refuse with a usage message on stderr and exit 2 before any sample is taken.
- REQ-11: The hook latency tool shall compute each percentile as the nearest rank (the smallest value with at least that share of the values at or below it), and a measured run exits 0 whatever the milliseconds are · failure: none — the tool judges no number.
- REQ-12: When `moai hook latency --values <list>` runs, the hook latency tool shall print the same summary for the given millisecond values without starting any process · failure: none — an empty or non-numeric list is refused under REQ-10.

## Acceptance

- AC-1 (REQ-1, REQ-4): in a temporary MoAI project, each registered name with a valid JSON object on stdin prints nothing on stdout and stderr and exits 0, and the project folder and HOME hold the same files before and after · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac1_/)'`
- AC-2 (REQ-2): an unknown name (`pre-tool`, `nonsense`, an empty-looking name) exits 0 with stdout empty, with and without a MoAI project, and `pre-tool` keeps the stderr line `ignored a Go-generation hook call` · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac2_/)'`
- AC-3 (REQ-3): a registered name run where `.moai/project.yaml` is missing, is a folder, or is a link to a missing file prints nothing on either stream, exits 0 and writes no file in HOME or the folder, and `CLAUDE_PROJECT_DIR` wins over the working folder when both are set · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac3_/)'`
- AC-4 (REQ-5, REQ-6): in a MoAI project whose HOME holds the project folder, input that is empty, `not json`, a JSON array, invalid UTF-8 or larger than 4 MiB (4194305 bytes) exits 0 with stdout empty and adds exactly one JSON line to `hook-entry.jsonl` with the name, a reason word and a time and none of the input bytes, while a JSON object of exactly 4 MiB (4194304 bytes) is accepted with no receipt line · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac4_/)'`
- AC-5 (REQ-6): the same bad input where the HOME has no project folder exits 0, creates no folder, and writes one stderr line, and a receipt path that is a link is not followed and a `project_id` such as `../../outside` is refused, each with one stderr line and nothing written · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac5_/)'`
- AC-6 (REQ-7): with `hook-entry.jsonl` at 262144 bytes a bad input appends nothing, writes one stderr warning naming the limit and exits 0, and at 262143 bytes it still appends · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac6_/)'`
- AC-7 (REQ-8): with `ORCA_*` variables set the answers and files equal those without them, `src/hook_entry.rs` names no process-start, network or settings-path word, and no `settings.json` or `settings.local.json` exists afterwards in the project or HOME · verify: `cargo nextest run --test hook_entry_cli -E 'test(/^ac7_/)'`
- AC-8 (REQ-9, REQ-10): `hook latency` against the built binary with `--samples 3` answers both cases with `samples` 3 and `p50_ms`, `p90_ms`, `max_ms` ordered p50 ≤ p90 ≤ max, and refuses a missing binary, `--samples 0`, `--samples 1001`, `--name pre-tool` and a bad `--values` list with a usage message on stderr and exit 2, accepts `--samples 1` and `--samples 1000` (the range check comes before any sample, so 1000 against a failing stand-in binary starts exactly one process), and removes only the measuring folder it made under the temporary folder · verify: `cargo nextest run --test hook_latency_cli -E 'test(/^ac8_/)'`
- AC-9 (REQ-11, REQ-12): `hook latency --values` on literal lists gives the nearest rank at the boundaries (10 values p50 → 5th, p90 → 9th; 20 values p90 → 18th; one value → itself), starts no process, and no field of the answer is a verdict · verify: `cargo nextest run --test hook_latency_cli -E 'test(/^ac9_/)'`
- AC-10 (REQ-2): the existing hook commands (`hook-dispatch`, `hook-receipt`, `claim-verify`, `lane-stop-hook`, the Go-generation `hook pre-tool` call) keep their behavior, shown by their test targets still passing · verify: `cargo nextest run --test hook_dispatch_cli --test hook_mode_cli --test hook_legacy_cli --test lane_stop_hook_cli --test diff_go`
- AC-11 (REQ-4, REQ-12): `moai help` and its goldens are unchanged by this slice, shown by the help layout target passing with no golden file touched · verify: `cargo nextest run --test help_layout`
- AC-12 (REQ-1…REQ-12): the SPEC itself passes the spec check · verify: `moai spec check SPEC-20261123-hook-entry-skeleton`

## Scope

- write: `.moai/specs/SPEC-20261123-hook-entry-skeleton/**`, `src/hook_entry.rs` (new), `src/hook_latency.rs` (new), `src/main.rs` (two `mod` lines and the `hook` word reaching the new entry before the Go compatibility path), `tests/hook_entry_cli.rs` (new), `tests/hook_latency_cli.rs` (new).
- read-only: `src/hook_engine.rs`, `src/hook_guards.rs`, `src/claim_verify.rs`, `src/hooks.rs`, `src/ledger/location.rs`, `src/home_workdir.rs`, `tools/hook_latency.py`.
- forbidden: `src/help.rs`, `src/lib.rs` and the help goldens (another card holds them), any settings file (the `.claude/**` folder of the home and of the project), the installed hooks of this clone, the real HOME and its `.moai/projects` folder, and any network call. Tests use temporary HOMEs and folders only.

## Plan

1. (this slice) SPEC, status draft; the user's approval is not written by a lane.
2. Tests first (RED): `tests/hook_entry_cli.rs` and `tests/hook_latency_cli.rs` against stubs that fail.
3. `src/hook_entry.rs` (the entry: bounded stdin, project check, registered names, handler table, receipt with cap) and `src/main.rs` (the `hook` word).
4. `src/hook_latency.rs` (nearest-rank percentile, sample runner, `--values`) reached from the `hook` word in `src/main.rs` (so `src/hook_entry.rs` itself starts no process).
5. Mutants: remove each guard in turn and show the test that fails.
6. Dangerous-path run in a throwaway: the real binary with a hostile project folder (link, folder named `project.yaml`), oversized and non-UTF-8 input, a full receipt file.
7. Measurement run: `hook latency` on the built binary, p50/p90 recorded as numbers with the build label.
8. Targeted tests, one `check affected`, READY with head, base and the receipt.

## Risks

- R-1: The new `hook` word sits in front of `go_hook_main`. A wrong order would change what Go-generation projects see. AC-2 and AC-10 pin the `pre-tool` call and the other hook commands.
- R-2: A registered name collides with a future Claude event spelling. The three names are not Go event names (`session-start`, `pre-tool`, `post-tool`, …) and are not Claude event names, so no installed Go hook script reaches a handler.
- R-3: `hook-entry.jsonl` is written only into an existing project folder, so on a machine where the project was never registered a failure leaves one stderr line and no record. That is the cost of "the folder `register` creates and nothing else does"; a later slice may register the folder first.
- R-4: Wall-clock numbers depend on machine load. The tool labels the build profile and the sample count and judges nothing (REQ-11); it does not read the machine load, so whoever reads the numbers notes the load beside them.
- R-5: The exec form (`args`) is documented, but the minimum Claude Code version is unverified (see Sources). This slice installs nothing, so the risk belongs to 1-e; a quoted shell-form `command` is the fallback.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The registered names are `session-context` (H1), `bash-boundary` (H2) and `write-boundary` (H3). | The plan names H1 to H3 by role and leaves the spelling to this slice; the names avoid the Go event words. | Rename in the handler table and the tests. |
| D-2 | Every handler of this slice answers no decision. | The plan puts the judgments in 1-b to 1-d; a skeleton that decides anything now would be an unreviewed guard. | Not applicable: nothing decides. |
| D-3 | An unknown name keeps the Go compatibility answer (stderr line, exit 0), before the project check. | `tests/diff_go.rs` pins `hook pre-tool`; the plan asks that an unknown event exit 0. | Drop the stderr line when the Go test is retired. |
| D-4 | The receipt file is `hook-entry.jsonl`, capped at 256 KiB, one JSON line per failure, no input bytes. | Plan §4: one receipt line per engine failure, a warning at the limit; input may hold secrets (principle: secrets never enter reports). | Change the constant. |
| D-5 | The measuring tool is `moai hook latency`, a reserved word of the `hook` entry, with a plain JSON answer and exit 2 for a usage error. | `src/help.rs` and `src/lib.rs` are held by another card; `latency` is not a registered hook name; no host installs it. | Promote it to a `src/ops/` word (`hook-latency`) in a later slice. |
| D-6 | The input limit is 4 MiB, as the older hook engine's limit. | A Write tool input carries file content; 4 MiB is the size the older entry already reads. | Change the constant. |

### Sources

Official hooks documentation, read for the plan §8 open item "exec form". Retrieved by a research subagent with WebFetch (a small model summarised the page, so the quotes are the tool's rendering, not raw bytes); the clock read before the first fetch was 2026-11-23T05:54:47Z, and this session's own clock read was 2026-11-23T05:57:10Z with local Claude Code 2.1.330. The page had 293,973 characters; the first 200,000 were read.

- https://code.claude.com/docs/en/hooks, section "Exec form and shell form": "A command hook runs as exec form when `args` is set, and shell form when `args` is omitted." and "Claude Code resolves `command` as an executable on `PATH` and spawns it directly with `args` as the argument vector. There is no shell, so each `args` element is one argument exactly as written". The page's example: `{ "type": "command", "command": "node", "args": ["${CLAUDE_PROJECT_DIR}/.claude/hooks/check-command.js"], "onFailure": "block" }`. In exec form `command` is the executable name or path only; an absolute path with spaces is one valid executable.
- Shell form needs quoting for paths with spaces: `"command": "node \"${CLAUDE_PLUGIN_ROOT}\"/scripts/format.js --fix"`, and the page says to prefer exec form for any hook that references a path placeholder.
- Exit codes: 0 success, 2 blocking error, any other code or a failure to start is non-blocking unless `onFailure` is `"block"` (documented as requiring v2.1.295 or later). Default `timeout` for command hooks is 600 seconds.
- `${CLAUDE_PROJECT_DIR}` is "the project root where the session started" and does not follow a worktree switch; the hook input's `cwd` does.
- Unverified: the minimum Claude Code version for `args` (the page states none; `onFailure` v2.1.295 and the page's own `onFailure` example using `args` suggest 2.1.330 supports it, but no source says so); whether `${CLAUDE_PROJECT_DIR}` expands in a user-level settings file (the settings page was not fetched); the last 93,973 characters of the hooks page. Consequence for 1-e: the installed form is chosen by a real load check on 2.1.330, with a quoted shell-form `command` as the fallback; this slice installs nothing.

## Result

Status: implemented on the code slice. The SPEC stays `draft`: approval is the user's pick, not a lane's. Nothing is installed, no settings file is read or written, and the installed hooks of this clone are unchanged.

- AC-1: met — `hook_entry_cli` test `ac1_a_registered_name_with_good_input_prints_nothing_and_writes_nothing`
- AC-2: met — `ac2_an_unknown_name_exits_zero_with_nothing_on_stdout` (with and without a project; `pre-tool` keeps its stderr line). This test passed before the code too: it pins the compatibility path that already existed.
- AC-3: met — `ac3_outside_a_moai_project_a_registered_name_is_silent`, `ac3_claude_project_dir_wins_over_the_working_folder`
- AC-4: met — `ac4_bad_input_adds_one_receipt_line_without_the_input` (empty, text, array, scalar, non-UTF-8, 4 MiB + 1 byte; the marker word never reaches the file), `ac4_the_stdin_limit_takes_exactly_the_limit_and_refuses_one_more` (an object of exactly 4194304 bytes: no receipt; 4194305 bytes: one `stdin_too_large` line)
- AC-5: met — `ac5_without_a_project_folder_the_failure_is_one_stderr_line`, `ac5_a_receipt_link_is_not_followed_and_a_hostile_project_id_is_refused` (the second was added after the dangerous-path run showed both guards untested)
- AC-6: met — `ac6_a_full_receipt_file_warns_and_appends_nothing` (262144 bytes: no append, one warning; 262143: appends)
- AC-7: met — `ac7_orca_variables_change_nothing_and_no_settings_file_appears` (answers and receipt equal with and without `ORCA_*` and `CLAUDE_CONFIG_DIR`; the source scan finds no process-start, network or settings word, and exactly the variables `CLAUDE_PROJECT_DIR` and `HOME`)
- AC-8: met — `hook_latency_cli` tests `ac8_both_cases_are_measured_with_ordered_percentiles`, `ac8_bad_arguments_are_refused_with_exit_2_before_any_sample`, `ac8_the_samples_range_takes_both_ends_and_refuses_one_beyond`, `ac8_only_the_folder_the_run_made_is_removed`
- AC-9: met — `ac9_values_give_the_nearest_rank_and_judge_nothing`
- AC-10: met — one `cargo nextest run` over `hook_dispatch_cli`, `hook_mode_cli`, `hook_legacy_cli`, `lane_stop_hook_cli`, `diff_go` and `help_layout`: 85 passed. In `diff_go`, the Go-hook case returns early when no Go binary is found, so AC-2 is the evidence for the `pre-tool` line.
- AC-11: met — `help_layout` is in the same 85-test run; no golden file and neither `src/help.rs` nor `src/lib.rs` is in the diff.
- AC-12: met — `moai spec check SPEC-20261123-hook-entry-skeleton` answers 0 errors and 0 warnings.

Against the code that existed before, 9 of the 11 new tests failed (commit `1bcc31beb2`); the two that passed were AC-2, the compatibility test, and the first AC-5 test, which passed by accident through the old Go stderr line and was then tightened to name the hook and the reason. After the code: 12 of 12 pass; `cargo fmt --check` and `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0.

Mutants (each edits one guard of the committed code and runs `hook_entry_cli` and `hook_latency_cli`): project check removed, stdin limit, receipt limit off by one, project folder created, `CLAUDE_PROJECT_DIR` ignored, registered name sent to the Go path, unknown name loses its Go line, percentile rounds down, sample range removed, NaN accepted, receipt link followed, project id unchecked — 12 of 12 killed, none by a compile error. Not covered by a mutant: that the receipt never holds input bytes (the `SECRETMARKER` check in AC-4 covers it by test only).

Dangerous path, run with the real debug binary in a throwaway HOME and project: a 5 MiB input and a non-UTF-8 input each left one receipt line and exit 0; a full receipt log gave one stderr warning; a receipt path linked to a file outside left that file as it was; the project id `../../outside` was refused and nothing was written outside; `project.yaml` as a folder and as a link to nothing stayed silent; the names `$(touch PWNED)`, `; touch PWNED2`, `../../x` and an empty name exited 0 and created no file; closed stdin exited 0.

Measurement (`moai hook latency --bin <debug binary> --samples 50`, debug build, machine load 80 to 90, 2026-11-23): `session-context` p50 7.0 ms and p90 11.7 ms outside a project, p50 7.5 ms and p90 13.4 ms in a project; `bash-boundary` p50 9.1 and 7.8 ms, p90 16.7 and 14.9 ms; `write-boundary` p50 8.2 and 7.9 ms, p90 13.3 and 11.5 ms. These are the cost of the skeleton alone (no handler decides anything), under load and in a debug build. The plan's targets (H1 p50 ≤ 50 ms, p90 ≤ 100 ms; H2 and H3 p50 ≤ 20 ms, p90 ≤ 40 ms) are judged by the slices that add the handlers, on a release build.

Not done here: a real Claude Code session loading the entry (nothing is installed), a Windows run, and the exec form version check (see Sources). The `check affected` receipt of the final head is on the card's READY note.
