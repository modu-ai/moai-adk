---
id: SPEC-20261124-home-guard-flake
title: "HOME 가드가 부하 아래에서 테스트가 읽지 않은 HOME 읽기를 테스트 탓으로 실패시키는 일을, 먼저 증거로 가려내요"
status: draft
revision: 2
created: 2026-11-24
---

# HOME 가드가 부하 아래에서 테스트가 읽지 않은 HOME 읽기를 테스트 탓으로 실패시키는 일을, 먼저 증거로 가려내요

## Interview

- none: 리더 배차(카드 본문)가 요구를 정했어요 — 가드를 느슨하게 하지 않고, 진짜 HOME 쓰기는 계속 실패시키며, 원인을 재현으로 찾아요. 사용자 인터뷰는 없었어요.

## Intent

`tools/test_home_guard.sh`(SPEC-20261123-test-home-isolation)가 통과한 테스트를 `HOME-GUARD: touched its HOME` 으로 실패시키는 일이 2026-11-24 에 세 레인에서 다시 나왔어요(카드 노트 8260, 8267, 8269). 같은 테스트는 혼자 돌리거나 같은 head 로 한 번 더 돌리면 통과해요. 레인마다 `check affected` 를 한 번 더 돌려 넘기고 있어서, 레인 수만큼 시간이 새요.

지금 아는 것과 모르는 것을 나눠 적어요. 읽은 명령과 숫자는 이 SPEC 의 Result 에 옮겨요(`target/` 의 로그는 카드가 열려 있는 동안만 남아요).

아는 것 (이 실행에서 직접 읽고 잰 것):

- 실제 실패 기록 하나가 남아 있어요: 다른 레인의 `check affected` 로그 `05-test.log` 의 두 줄이에요. 두 줄 모두 찍힌 항목이 정확히 decoy 폴더, `.moai`, `.moai/user-assets.json` 셋뿐이에요. `language.yaml`, `settings.json`, `.claude` 는 찍히지 않았어요.
  - `language_setting` 의 `the_rerun_keeps_the_stored_conversation_language` (0.448 s)
  - `review_section_cli` 의 `review_md_writes_default_sourced_decisions_but_not_hold`
- 두 테스트 모두 기준선(`baseline.tsv`)에 없고, 두 테스트 파일은 `moai` 자식마다 `.env("HOME", <자기 임시 폴더>)` 를 줘요(`tests/language_setting.rs` 의 `run`, `tests/review_section_cli.rs` 의 `cli`). 그러니 테스트와 그 자식이 가드의 decoy 를 읽을 길이 코드에 없어요. 읽은 쪽은 테스트 바깥이에요.
- `user-assets.json` 을 여는 코드는 `home_state::home_go_markers` 하나예요. `home_root::path` 가 HOME 의 상태 폴더를 정할 때마다 불러요. 읽힌 모양이 그 코드가 남기는 모양(폴더 둘 읽기 + 파일 하나 읽기)과 같아요.
- 바깥 읽기는 아주 드물어요. 한 번의 `check affected` 테스트 단계에서 4083개 중 2개예요.

모르는 것:

- 읽은 쪽이 누구인지. 원인은 uncertain 이에요.

재현 시도는 전부 0건이었어요. 이 숫자는 "없다"가 아니라 "이 방법으로는 못 잡았다"예요.

| 시도 | 횟수 | 부하 평균 | 거짓 실패 |
| --- | --- | --- | --- |
| 실패한 테스트 한 개를 가드(observe)로 반복, 병렬 8 | 40 | 44 | 0 |
| 아무것도 안 하는 프로세스(8 s)를 가드로 반복, 병렬 20 | 60 | 37~45 | 0 |
| 가드와 같은 모양의 decoy 800개를 20 s 두고 비교(`.metadata_never_index` 없음) | 800 | 42~105 | 0 |
| 같은 모양 800개, `.metadata_never_index` 있음 | 800 | 42~105 | 0 |
| 이전 레인: 아무것도 안 하는 프로세스를 가드로 반복 | 300 | - | 0 |
| 이전 레인: 실패 이력이 있는 바이너리 5개(59개 테스트)를 nextest 모양으로 | 295 | 65~95 | 0 |

배제한 후보 (각각 근거가 있어요):

- Spotlight 같은 시스템 색인기: 가만히 둔 decoy 1600개에 읽기 0건, 마커 유무와 무관해요.
- 테스트 자신이나 그 `moai` 자식: 위에 적은 대로 코드에 읽을 길이 없고, 실패한 테스트를 40번 단독으로 돌려도 읽지 않아요.
- `hygiene.rs`/`worktree_report.rs` 의 임시 폴더 훑기: 최상위만 보거나 크기만 재고, `user-assets.json` 을 열지 않아요.
- `ps` 로 프로세스 표를 읽는 `home_state::process_command_lines`: 명령줄 인자만 읽고 환경 변수 HOME 은 안 읽어요.

여기서 코드를 고치면 원인이 아닌 곳을 고치게 돼요. 거짓 실패가 한 번 더 나올 때 그 한 번이 원인을 알려주도록, 증거를 남기는 장치를 먼저 만들어요. 가드의 판정(통과·실패·종료 코드)은 바꾸지 않아요.

## Non-goals

- 가드를 느슨하게 하는 일. 진짜 HOME 읽기·쓰기·삭제는 계속 exit 1 이에요. 판정 규칙, 기준선, `enforce` 기본값은 이 SPEC 이 건드리지 않아요.
- 실패한 테스트를 가드가 다시 돌리는 일(다시 돌리면 부작용이 있는 테스트가 있어요).
- 원인을 모른 채 읽기를 눈감아 주는 규칙(예: 폴더만 찍히면 통과). 이 SPEC 의 증거가 원인을 가리킨 뒤 SPEC 개정(revision 2)이나 새 SPEC 으로 다뤄요.
- Windows 호스트. 가드가 `sh` 를 쓰므로 unix 만이에요.

## Requirements

<!-- This SPEC adds evidence only and no blocking check: the guard's verdict stays as it is, so there is no observe step to record. -->
- REQ-1: When the wrapper prints a `HOME-GUARD: touched its HOME` line, it shall append one `HOME-GUARD-DETAIL:` line per hit entry to the file named by `MOAI_HOME_GUARD_LOG`, with the entry's access time and modification time in seconds with a fraction.
- REQ-2 (REQ-1): The `HOME-GUARD-DETAIL:` lines shall carry the modification time of the reference file, which the wrapper makes just before it starts the test, and the time it compared the HOME, in seconds with a fraction, so a read can be placed before the test, during it or after it.
- REQ-3 (REQ-1): The `HOME-GUARD-DETAIL:` lines shall carry, for each other folder named `moai-test-home.*` under `TMPDIR`, the access time of its `.moai/user-assets.json`, read with `stat` only.
- REQ-4: The wrapper shall never list, open or read any entry inside another guard folder.
- REQ-5: The wrapper shall keep every exit code and every `HOME-GUARD:` line it printed before this change.
- REQ-6 (REQ-1): IF a `HOME-GUARD-DETAIL:` line cannot be written, THEN the wrapper shall ignore the failure and keep the test's result.
- REQ-7: Where `perl` is not on the host, the wrapper shall write the times in whole seconds.
- REQ-8: The wrapper shall write the paths of a `HOME-GUARD-DETAIL:` line relative to the test's HOME, and name a peer folder by its own file name, so the line holds neither the user's home path nor the `TMPDIR` path nor the test's arguments.
- REQ-9 (REQ-1): The `HOME-GUARD-DETAIL:` lines shall carry the process id and the file name of the command (neither its arguments nor its folder) of every running process whose command file name starts with `moai`.
- REQ-10 (REQ-9): IF the process table cannot be read, THEN the wrapper shall write one `procs=unavailable` detail line and keep the test's result.

## Acceptance

- AC-1 (REQ-1, REQ-2): a test that reads the decoy `user-assets.json` under a log file gives a `HOME-GUARD:` line and one `HOME-GUARD-DETAIL:` line for each hit entry, each with an access time and a modification time that has a fraction, and the test's start time is not later than the comparison time. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-2 (REQ-3, REQ-4): with another guard folder planted beside the test's, the detail line names it with its `user-assets.json` access time, and the planted folder's own access times are the same after the wrapper ran as before. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-3 (REQ-5): every test of `tests/test_home_guard.rs` that existed before this change passes unchanged: a decoy read, a write and a removal still exit 1 in the default mode, observe keeps the test's exit code, a listed test is tolerated as before. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-4 (REQ-6): a log path that cannot be written changes neither the exit code nor the printed `HOME-GUARD:` line. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-5 (REQ-7): with `perl` hidden from `PATH` the detail line holds whole seconds and the verdict is the same. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-6 (REQ-8): no detail line holds the `TMPDIR` path, the user's home path or an argument of the test, and the hit entries are written as `.`, `.moai` and `.moai/user-assets.json`. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-7 (REQ-9, REQ-10): with a process named `moai-probe` running beside the test, a detail line carries its id and that name and none of its arguments; with `ps` hidden from `PATH` one `procs=unavailable` line is written and the verdict is the same. · verify: `cargo nextest run --locked --offline --test test_home_guard`
- AC-8 (REQ-1): through the scratch crate and `cargo nextest run`, with `MOAI_HOME_GUARD_LOG` set outside it, the log file holds the `HOME-GUARD-DETAIL:` lines of the reading plant, so the variable reaches the wrapper through nextest. · verify: `cargo nextest run --locked --offline --test test_home_guard`

## Scope

- write: `tools/test_home_guard.sh`, `tests/test_home_guard.rs`, `.moai/specs/SPEC-20261124-home-guard-flake/spec.md`
- read-only: `src/`, `tests/fixtures/test_home_guard/baseline.tsv`, `.config/nextest.toml`
- forbidden: the baseline's lines (it never grows), the `enforce` default, the decoy files and their old access time, any rule that excuses a hit

## Plan

1. RED: the acceptance tests of AC-1, AC-2, AC-5, AC-6, AC-7 and AC-8 in `tests/test_home_guard.rs` fail on the current wrapper; AC-3 and AC-4 pass before and after.
2. `tools/test_home_guard.sh` reads the start time from the reference file and, only after a `touched` verdict that is not a baseline `tolerated` one, writes the detail lines; the verdict code is not touched and the path of a test that touched nothing gains no new command. The lines are:
   - `HOME-GUARD-DETAIL: kind=hit test=<binary>:<test> path=<relative path> atime=<t> mtime=<t> ref=<t> end=<t>`, one per hit entry (`.` is the HOME itself);
   - `HOME-GUARD-DETAIL: kind=peer test=<binary>:<test> folder=<moai-test-home.name> atime=<t>`, one per other guard folder (`atime=none` when it holds no `user-assets.json`);
   - `HOME-GUARD-DETAIL: kind=proc test=<binary>:<test> pid=<id> name=<file name>`, one per running `moai*` process, or one `kind=procs ... procs=unavailable` line.
3. Guard mutants, one line each in the READY note: removing the detail write fails AC-1; reading a planted peer with `ls` or `find` instead of `stat` fails AC-2; changing an exit code fails AC-3; writing the absolute path fails AC-6; writing the arguments of a process fails AC-7.
4. `moai check affected` once at the head; the receipt goes in the READY note, with a reading of whether `MOAI_HOME_GUARD_LOG` reaches nextest through `check affected` (`run_step` in `src/check_affected.rs` clears no variable) and through nextest (AC-8).
5. Later, with no new card: the next real `touched its HOME` that is not a real read leaves the detail lines in the log; the reader is read from them and revision 2 fixes the source.

## Risks

- A planted peer check that opens a file would change the peer's access time and make the guard cause the very flake it measures. REQ-4 and AC-2 pin that.
- The detail lines say who did not read, not who did. If the instants of several guard folders cluster at one moment the reader swept the temporary folder; if only one folder moved the reader knew its path. Neither names a program; it narrows the search.
- `perl` is not on every host: REQ-7 keeps the lines but with coarser times.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Evidence first: this revision changes no verdict. | proposed by the lane, awaiting the leader (LANE-BLOCKED kind=decision); default if the leader does not answer: this option | drop the detail lines |
| D-2 | A control folder (an unreachable twin of the decoy whose read marks a hit as an outsider's) is not part of this revision. | proposed: it would change what passes when a real read and an outside read coincide, a contract change that needs the cause first; the leader's decision note 8329 agrees (B only after the cause is known) | add it in a later revision after the evidence names the reader |
| D-3 | Option A is approved: the verdict, exit codes, baseline and `enforce` default stay as they are, and the detail lines are evidence only. | leader decision note 8329 of card C-451 (2026-11-24T07:39:54Z, under the user's delegation of 2026-11-23T13:57Z, not a direct user pick), which approved revision 1 at digest `sha256:81da21c5` | revert to the earlier wrapper |
| D-4 | The detail lines hold paths relative to the HOME and no user path (REQ-8). | leader note 8329, addition 1 | none needed: it only narrows what is written |
| D-5 | The detail lines also name the running `moai*` processes by id and command file name (REQ-9, REQ-10). | leader note 8329, addition 2, left to the lane's judgement: the only code that opens `user-assets.json` is `home_state::home_go_markers`, so a `moai` process alive at the hit that is not the test's child tells whether another test's `moai` ran with a wrong HOME; the cost is one `ps` on a `touched` verdict only, and no argument is kept | drop REQ-9 and REQ-10 |
| D-6 | Revision 2 holds the three additions of note 8329 and is a new text, so its digest is new: the lane records no approval of it; the leader reads the digest in the READY note. | CLAUDE.md rule that a lane never records an approval it did not get | none |

## Result

Measured 2026-11-24 (08:06Z) on macOS arm64 in this branch's worktree, nextest 0.9.152, in-tree `target/`, heavy slots `--max 3`, load about 25 to 40. Linux and Windows were not run. The cause of the false `touched its HOME` failures is still not known (uncertain); this revision records evidence and changes no verdict, so the card's criterion 1 stays `not_met`.

- AC-1 met: a reading test under a log file gives the `HOME-GUARD:` line and one `kind=hit` detail line per hit entry (the count equals the entries of the printed line), each with `atime`, `mtime`, `ref` and `end` that have a fraction, `ref` not later than `end`, and for `.moai/user-assets.json` `ref` < `atime` <= `end`; a test that touches nothing leaves no detail line (`a_touched_verdict_writes_one_detail_line_per_hit_entry`).
- AC-2 met: a planted `moai-test-home.PEER01` whose files carry an access time of 2000 is named with that time, and its file, its `.moai` folder and the folder itself keep that access time after the run (`another_guard_folder_is_named_by_its_access_time_and_never_opened`).
- AC-3 met: the 42 tests of `tests/test_home_guard.rs` that existed before this change pass unchanged (a decoy read, a write and a removal fail in the default mode, observe keeps the exit code, a listed test is tolerated); the whole file is 49 tests, 49 passed (`target/c451/green2.log`, 2.9 s, with this wrapper wrapping the test processes themselves).
- AC-4 met: a log path that cannot be written changes neither the exit code nor the printed line and leaks no shell error (`a_detail_that_cannot_be_written_changes_nothing`, which also passed before the change).
- AC-5 met: with a `perl` that fails, the times are whole seconds and the verdict is still 1 (`without_perl_the_times_are_whole_seconds_and_the_verdict_is_the_same`).
- AC-6 met: no detail line holds the `TMPDIR` path, the repository path, the HOME of the test process or a marker planted in the test's arguments; hit paths are `.`, `.moai` and `.moai/user-assets.json` and none starts with `/` (`the_detail_lines_hold_relative_paths_and_no_argument`).
- AC-7 met: a process named `moai-probe` (a link to `sleep`; a copied system binary is killed by macOS at start) is named by id and file name with its argument and folder absent, and with a `ps` that fails one `procs=unavailable` line is written and the verdict is still 1 (`a_moai_process_is_named_by_id_and_file_name_only`, `an_unreadable_process_table_leaves_one_line_and_the_same_verdict`).
- AC-8 met: with `MOAI_HOME_GUARD_LOG` set outside the scratch crate, the log holds the detail line of the reading plant after `cargo nextest run`, so the variable reaches the wrapper through nextest (`the_log_variable_reaches_the_wrapper_through_nextest`). That `moai check affected` passes it on to nextest is read from `run_step` in `src/check_affected.rs` (the child gets the caller's environment plus the step's own variables, no `env_clear`); the receipt of the READY note says whether the run used it.

RED (before the wrapper change, head `eead0d6fae`): 49 tests run, 42 passed, 7 failed, each because the log held no detail line (`target/c451/red-kept.log`).

Guard mutants (each applied to the committed head `51459aed6e`, the whole `test_home_guard` binary run under nextest, then restored; the tree was clean after):
- remove the detail write → killed by 7 tests (`a_touched_verdict_writes_one_detail_line_per_hit_entry` among them).
- read the other guard folder with `cat` and `ls` instead of `stat` → killed by `another_guard_folder_is_named_by_its_access_time_and_never_opened`.
- change the enforce exit code from 1 to 0 → killed by 15 tests (`enforce_fails_a_test_that_reads_a_decoy` among them).
- write the absolute path of a hit → killed by `a_touched_verdict_writes_one_detail_line_per_hit_entry`, `the_detail_lines_hold_relative_paths_and_no_argument`, `the_log_variable_reaches_the_wrapper_through_nextest`.
- write the whole command path of a process → killed by `a_moai_process_is_named_by_id_and_file_name_only`.
- keep the fraction switch on without `perl` → killed by `without_perl_the_times_are_whole_seconds_and_the_verdict_is_the_same`.
- change the `procs=unavailable` line (the first try replaced a comment and survived, an invalid mutant; the second hit the code line) → killed by `an_unreadable_process_table_leaves_one_line_and_the_same_verdict`.
- stop silencing the stderr of the detail write → killed by `a_detail_that_cannot_be_written_changes_nothing`.

Not measured: the cost per test. The path of a test that touches nothing gains the parsing of three function definitions and no command; a hit costs one `perl`, two `stat` per entry, one `stat` per peer and one `ps`.

Evidence of the false failure that this revision serves (kept here because `target/` logs live only while the card is open): two lines of one `check affected` test step (4083 tests), each naming only the decoy folder, `.moai` and `.moai/user-assets.json`, for tests that give every `moai` child its own HOME. Reproduction attempts under load found 0 of about 1,800 (the failing test alone 40 times; an idle process 60 times; guard-shaped decoys 1,600, with and without `.metadata_never_index`). Ruled out: Spotlight, the test and its children, the temporary-folder scans of `hygiene.rs` and `worktree_report.rs`, the `ps` read of `home_state`. Next: when a `touched` line appears for a test that cannot have read its HOME, the detail lines of the card's log show when the read happened (before, during or after the test), whether other guard folders were read in the same instant, and which `moai` processes were alive; revision 3 or a new SPEC fixes the source then.
