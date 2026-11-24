---
id: SPEC-20261125-slow-test-top
title: "Slow test top: rank by run time times runs a day, then fix the top few with assertions kept"
status: draft
revision: 1
created: 2026-11-25
---

# Slow test top: rank by run time times runs a day, then fix the top few with assertions kept

## Intent

느린 테스트 상위를 같은 단언으로 빠르게 해요. 먼저 "실행 시간 × 하루 실행 횟수" 순위 표를 만들고, 그중 세 표본(`home_state_selective` 의 두 테스트와 부하 아래 플레이키인 `release_pr_three_os` 의 한 테스트)을 세팅, 외부 대기, 고정 sleep 으로 나눠 원인을 찾아 고쳐요. 테스트를 덜 하는 것이 아니라 같은 검증을 덜 기다리게 하는 카드라서 단언은 하나도 지우지 않아요.

## Requirements

- REQ-1: While the Result holds a rank table, each row shall give the test's median nextest time, its runs in the last day and the product of the two.
- REQ-2: While a `home_state_selective` test does not judge whether a program uses a folder, its `moai` command shall read the empty process table.
- REQ-3: While a `home_state_selective` test holds a program it started, the test shall stop that program's own children together with it.
- REQ-4: While the release workflow's decide step runs in a test, the stand-in programs it calls shall be complete files that no process holds open for writing.
- REQ-5: While a test file in the write scope is changed, the file shall keep every assertion it had.
- REQ-6: While the named-file scan lists `tests/home_state_selective.rs`, each of its tests that starts a holder program shall opt back into the real process table.

## Acceptance

- AC-1 (REQ-1): the Result gives the rank table and the data it came from · verify: `moai spec check SPEC-20261125-slow-test-top --project .`
- AC-2 (REQ-2, REQ-3, REQ-5): `home_state_selective` passes with the same test names, and the two named tests run in seconds with no leaked process · verify: `cargo test --locked --offline --test home_state_selective`
- AC-3 (REQ-4, REQ-5): `release_pr_three_os` passes with the same test names · verify: `cargo test --locked --offline --test release_pr_three_os`
- AC-4 (REQ-5): the guard of each fix is killed by a mutant that removes the fixed code, and no assertion line was removed · verify: `git diff NeXT...HEAD --stat`
- AC-5 (REQ-6): the named-file scan passes with the file listed, and fails when a holder test of the file loses its opt-in · verify: `cargo test --locked --offline --test process_table_default`

## Scope

- write: `.moai/specs/SPEC-20261125-slow-test-top/**`
- write: `tests/home_state_selective.rs`, `tests/release_pr_three_os.rs`
- write (line lease): `tests/common/**`
- write (line lease, one line, approved by the leader after a scope request): `tests/process_table_default.rs`, the `NAMED` list
- carried by one merge on the leader's decision: `.moai/specs/SPEC-20261124-test-cost-profile-git-path/**`
- read-only: everything else, `src/` included; a change in `src/` waits for a SCOPE-REQUEST note on the card
- not in this slice: the ranked tests outside these files (`tests/go_state_cli.rs`, `tests/update_hooks_cli.rs` and others); their rank and cause stay in the Result for the leader to scope

## Plan

1. 순위 표를 만든다(24시간 안의 `check-affected` `05-test.log` 71개).
2. 세 표본의 시간을 세팅, 외부 대기, 고정 sleep 으로 나눠 원인을 확인한다.
3. 같은 단언을 둔 채 고치고, 가드는 뮤턴트로 확인한다.
4. 번갈아 5회 잰 중앙값으로 전후를 비교한다.

## Result

머리: 이 가지의 코드 커밋은 테스트 파일 세 개만 바꿔요(`tests/release_pr_three_os.rs` 의 `894a88c709`, `tests/home_state_selective.rs` 의 `684157345c`, 둘 다 건드리는 `2c572ce434`, 거기에 `tests/process_table_default.rs` 한 줄). `src/` 와 `tests/common/` 은 바꾸지 않았어요. 이 SPEC 과 `SPEC-20261124-test-cost-profile-git-path`(머지 커밋 `609160e645` 로 실었어요)가 문서 변경이에요. 측정은 2026-11-24 에 macOS arm64 디버그 빌드에서 했고, 부하(load1)는 9~25(전후 표)와 25~50(순위 표의 원자료)이었어요. 도구와 원자료는 레인 워크트리의 `target/c473/` 에 있고 카드가 닫히면 사라져요(`rank.py`, `ab.py`, `summ.py`, `stress.sh`, `mutants.py`, `ab-hss.jsonl`, `rank-24h-all.tsv`). 카드 노트 8658, 8666 에 같은 내용이 있어요.

### 순위 표 (기준 1)

방법: 24시간 안(`find -mtime -1`)에 만들어진 `.claude/worktrees/*/target/check-affected-*/05-test.log` 71개를 읽어, 테스트마다 nextest 결과 줄(`PASS`·`FAIL`·`LEAK`)의 시간 중앙값과 그 테스트가 나온 로그 수(하루 실행 횟수)를 곱했어요(`target/c473/rank.py`). 테스트 7838개, 중앙값×횟수의 합계는 190,497초예요. 리더의 전체 테스트 로그는 읽지 못했어요(not_run: 이 워크트리에서 찾지 못했어요).

| 순위 | 중앙값(초) | 횟수 | 중앙값×횟수 | 테스트 | 이 슬라이스 |
|---|---|---|---|---|---|
| 1 | 81.76 | 71 | 5805 | `go_state_cli::a_receipt_above_256_kib_reverts_completely` | 범위 밖(노트 8666) |
| 2 | 38.87 | 71 | 2760 | `worktree_gc::tests::the_recheck_before_a_removal_refuses_files_and_an_install_folder` (src 단위 테스트) | 범위 밖 |
| 3 | 34.60 | 72 | 2492 | `update_hooks_cli::the_hook_receipt_keeps_a_protected_settings_file_privately_under_any_umask` | 범위 밖 |
| 4 | 29.56 | 71 | 2099 | `update_user_layer_cli::update_previews_and_installs_the_whole_user_layer_then_has_nothing_to_apply` | 범위 밖 |
| 5 | 24.78 | 71 | 1759 | `update_home::tests::apply_takes_only_the_previews_own_token` (src 단위 테스트) | 범위 밖 |
| 482 | 15.44 | 5 | 77 | `home_state_selective::more_than_five_hundred_entries_block_the_plan` | 고침 |
| 1046 | 6.45 | 5 | 32 | `home_state_selective::worktrees_and_entries_in_use_stay_in_place` (LEAK 5/5) | 고침 |
| 1316 | 4.70 | 5 | 24 | `home_state_selective::an_open_file_keeps_its_entry_in_place` (LEAK 5/5) | 고침 |
| 6589 | 0.92 | 1 | 1 | `release_pr_three_os::an_ordinary_pull_request_keeps_the_classifier_rule` | 고침(플레이키) |

- 시간은 부하 25~50 에서 병렬로 돈 값이라 부풀려져 있어요. 순서가 의미 있고 절대값은 아니에요.
- 리더가 넘긴 세 표본은 이 표에서 한참 아래예요. `home_state_selective` 는 그 파일이 영향받을 때만 돌아서 횟수가 5 예요. 상위 5개는 승인된 범위 밖 파일이라 이 슬라이스에서 고치지 않았고, 후보 원인만 노트 8666 에 남겼어요(미확인): 1위는 `many_specs_fixture(1000)` 의 픽스처 크기, 2·5위는 src 단위 테스트가 기계 전체 프로세스 표를 읽는 것(`SPEC-20261122-test-process-table-default` 의 F3 (b)), 3·4위는 확인 안 했어요.
- 기준 3 의 "상위 3~5개"는 표의 상위 5개가 아니라 리더가 넘긴 세 표본을 고쳤어요(이 카드의 지시와 리더의 표본 기록). 상위 5개는 not_run 이에요. 리더 결정: 상위 5개는 이 카드에서 하지 않고 이 카드의 READY 뒤에 파일 묶음마다 새 카드로 열고 측정을 먼저 해요(1위 `go_state_cli`, 2위 `src/worktree_gc.rs` 부터).

### 원인 (세팅, 외부 대기, 고정 sleep 으로 가른 결과)

- `an_open_file_keeps_its_entry_in_place` 의 60초: 외부 대기예요. 테스트가 `sh -c "sleep 60; :"` 를 띄우고 `Kill` 이 쉘만 죽여서, 손자 `sleep` 이 남아 하네스의 출력 파이프를 60초 동안 붙잡아요. 래퍼 아래에서 파이프를 EOF 까지 읽는 방식으로 재면 60.03초(`worktrees_and_entries_in_use_stay_in_place` 는 61.39초)이고, 프로세스 종료는 1.45초(2.80초)예요. nextest 에서는 LEAK 로 보여요(24시간 로그에서 두 테스트 모두 5/5). 같은 홀더가 `sleep 60` 이라는 고정 시간을 쓰지만 줄일 필요는 없고, 홀더가 남지 않게 하면 돼요.
- `more_than_five_hundred_entries_block_the_plan` 과 파일의 나머지: 세팅이에요. 이 파일만 공유 스위치(`common::process_table()`)를 안 써서, 명령마다 기계 전체 `lsof` 를 읽었어요.
- `an_ordinary_pull_request_keeps_the_classifier_rule` 의 플레이키: 스텁 쪽 경합이에요(타임아웃 아님: `decide` 스크립트에는 대기도 타임아웃도 없어요). 스크립트가 `printf ... | python3 -c ...` 로 분류기 답을 파이프로 넘기는데, 스텁의 `-c` 갈래는 입력을 읽지 않고 바로 끝나요. 쓰는 쪽이 늦으면 SIGPIPE 를 받고, `pipefail` 아래에서 `macos=...` 줄은 조건이 거짓이 돼 전체 행렬로 대체되며(`elif` 조건 안이라 `-e` 가 안 걸려요) `printf ... | python3 -c ...[macos_reasons]` 줄은 `-e` 로 단계가 조용히 실패해요(`d.ok` 가 stderr 없이 거짓). 진짜 Python 은 stdin 을 읽어요.

### 고친 것

- `tests/release_pr_three_os.rs`: 스텁이 `-c` 호출에서 `cat > /dev/null` 로 입력을 읽어요. 새 테스트 `the_json_reader_stand_in_reads_all_of_its_input` 이 1 MiB 를 쓰는 쪽을 앞에 세워 이를 고정해요. 테스트 8개가 9개가 됐어요.
- `tests/home_state_selective.rs`: `plan_env` 가 빈 프로세스 표를 실어요. 프로세스가 폴더를 쓰는지 판단하는 두 테스트(`worktrees_and_entries_in_use_stay_in_place`, `an_open_file_keeps_its_entry_in_place`)는 `use_real_process_table()` 로 실제 표를 읽어요. 홀더는 `spawn_group` 으로 띄우고 `Kill` 이 `kill_group` 으로 그룹째 멈춰요(`tests/common/mod.rs` 의 기존 도우미).
- `tests/process_table_default.rs`: `NAMED` 목록에 `tests/home_state_selective.rs` 를 더했어요(10→11, 리더가 승인하고 `card line-lease` 로 한 줄만 썼어요). 이 목록에 있으면 홀더 프로그램을 띄우는 테스트가 `use_real_process_table()` 를 안 부를 때 스캔이 실패해요. 스캔은 홀더를 `Command::new("/bin/sh")`·`"sleep"` 같은 표기로만 알아보는데 열린 파일 테스트는 `Command::new("sh")` 를 써서 첫 뮤턴트가 살아남았어요. 같은 프로그램이라 `"/bin/sh"` 로 적어서 스캔이 알아보게 했어요(주석으로 이유를 남겼어요).
- 지운 단언은 0개예요: 세 테스트 파일의 지운 줄 26개 중 `assert` 가 든 줄은 없어요(`git diff -U0 NeXT...HEAD -- tests/home_state_selective.rs tests/release_pr_three_os.rs tests/process_table_default.rs` 를 파일로 저장해 `^-` 줄을 센 값이에요).

### 전후 표: `home_state_selective` (기준 1·3)

방법(`SPEC-20261124-test-cost-profile-git-path` 의 번갈아 재기와 같아요): `tools/test_home_guard.sh <바이너리> --exact <이름> --nocapture` 로 테스트 하나씩, 기준선(`894a88c709` 앞의 NeXT 바이너리)과 수정본을 번갈아 5회씩(총 100회, 첫 실행 한 번은 재지 않아요) 돌려 중앙값을 냈어요(`target/c473/ab.py`). `t_exit` 는 프로세스가 끝난 시각, `t_eof` 는 출력 파이프가 EOF 에 닿은 시각이에요. 모든 실행의 종료 코드가 0 이었어요. load1 은 최소 9.0, 중앙값 13.4, 최대 25.1 이에요.

| 테스트 | 전 t_exit | 후 t_exit | 전 t_eof | 후 t_eof |
|---|---|---|---|---|
| `an_open_file_keeps_its_entry_in_place` | 1.45 | 1.50 | 60.03 | 1.50 |
| `more_than_five_hundred_entries_block_the_plan` | 7.84 | 2.75 | 7.84 | 2.75 |
| `worktrees_and_entries_in_use_stay_in_place` | 2.80 | 3.02 | 61.39 | 3.02 |
| `each_entry_gets_a_class_and_a_reason` | 1.69 | 0.48 | 1.69 | 0.48 |
| `secrets_are_listed_and_never_opened` | 3.40 | 0.37 | 3.40 | 0.37 |
| `taken_backup_names_move_to_the_stamp_then_block` | 2.77 | 0.29 | 2.77 | 0.29 |
| `a_live_go_session_blocks_the_selective_plan` | 1.49 | 0.21 | 1.49 | 0.21 |
| `the_whole_folder_mode_is_unchanged` | 1.74 | 0.19 | 1.74 | 0.19 |
| `an_unreadable_process_table_keeps_run_data_in_place` | 0.12 | 0.12 | 0.12 | 0.12 |
| `nothing_to_move_plans_none` | 1.35 | 0.14 | 1.35 | 0.14 |
| 중앙값의 합 | 24.66 | 9.07 | 141.84 | 9.07 |

- 두 홀더 테스트의 `t_exit` 는 그대로예요(1.45→1.50, 2.80→3.02). 실제 `lsof` 를 읽어야 하는 테스트라서 줄지 않는 게 맞고, 줄어든 것은 홀더가 남아 출력 파이프를 붙잡던 대기(60초)예요.
- 500개 테스트는 7.84초에서 2.75초가 됐어요(`SPEC-20261124-test-cost-profile-git-path` 이 잰 15초와 다른 것은 잰 때의 부하가 달라서로 보이지만 확인하지 않았어요). 남은 2.75초는 `moai` 본체와 파일 만들기이고 더 쪼개지 못했어요(not_run).
- 파일 전체를 `cargo test --locked --test home_state_selective` 로 한 번 돌리면 10개 통과, 3.56초였어요(수정 뒤 한 번만 쟀고 기준선과 짝지은 값은 아니에요).

### 플레이키 재현과 A/B (`release_pr_three_os`)

방법: 테스트 바이너리 전체(`--test-threads=16`, 한 프로세스의 스레드로 돌아 `cargo test` 와 같아요)를 `target/c473/stress.sh` 로 부하(자기 종료 반복문) 아래에서 되풀이했어요. 같은 시간대에 기준선 6줄과 수정본 6줄을 동시에 돌렸어요(각 20회, 합 120회).

| 바이너리 | 실패 횟수(줄마다) | 합 |
|---|---|---|
| 기준선 | 10, 10, 10, 11, 12, 7 | 60 / 120 |
| 수정본 | 0, 0, 0, 0, 0, 0 | 0 / 120 |

- 한 줄로(부하 반복문 14개, 20회) 돌렸을 때는 재현되지 않았어요(0/20). 8줄을 동시에 돌린 부하에서 재현됐고, 실패 위치는 줄 347·348·369·386·388 로 흩어졌어요(단계가 조용히 실패하거나 행렬이 전체로 바뀐 두 모양).
- 처음 세운 가설(스텁을 `fs::write` 로 만든 직후 실행하는 경합, `common::write_executable` 로 바꾸기)은 같은 방식 A/B 에서 기준선 27/120, 바꾼 쪽 19/120 이라 효과가 없어서 기각하고 되돌렸어요.
- 판정 스크립트가 분류기 실패를 다루는 방식은 이미 `STUB_SCOPE_FAIL`·`STUB_API_FAIL` 줄(`tests/release_pr_three_os.rs` 의 `an_ordinary_pull_request_keeps_the_classifier_rule`)이 직접 고정해요. 이번에 추가한 것은 스텁이 진짜 Python 처럼 입력을 읽는다는 성질의 고정이에요.

### 뮤턴트 (`target/c473/mutants.py`, 슬롯 한 번, 복사본으로 원복해 바이트 단위로 같음을 확인)

- `guard-mutant: the_json_reader_stand_in_reads_all_of_its_input | removed the stand-in's `cat > /dev/null; ` | exit 101` (죽었어요).
- `guard-mutant: an_open_file_keeps_its_entry_in_place | removed `use_real_process_table()` from the test | exit 101` (죽었어요: 빈 표에서는 `file_open` 을 못 봐요).
- `guard-mutant: worktrees_and_entries_in_use_stay_in_place | removed `use_real_process_table()` from the test | exit 101` (죽었어요).
- `guard-mutant: the_named_files_choose_their_table | removed `use_real_process_table()` from an_open_file_keeps_its_entry_in_place | exit 101` 와 같은 줄을 `worktrees_and_entries_in_use_stay_in_place` 에서도(`target/c473/mutants2.py`, 슬롯 한 번): 둘 다 죽었어요. 같은 뮤턴트가 `Command::new("sh")` 일 때는 `exit 0` 으로 살아남았고(스캔이 그 표기를 홀더로 안 봐요), 표기를 `/bin/sh` 로 고친 뒤 죽었어요.
- `src/` 의 제품 코드를 바꾼 뮤턴트는 돌리지 않았어요(범위 밖, not_run). 500개 한계의 단언은 이 슬라이스가 건드린 것이 프로세스 표뿐이고 단언 줄은 그대로예요.

### 한계 (not_run 포함)

- 순위는 부하 25~50 에서 병렬로 돈 24시간 로그 71개의 값이에요. 리더 전체 테스트 로그는 못 읽었고, 릴리스 빌드와 Linux·Windows 는 재지 않았어요.
- 전후 표는 `home_state_selective` 만이에요. `release_pr_three_os` 는 시간이 아니라 실패율로 쟀어요.
- 재현은 부하를 만든 환경에서만 됐고, 실제 전체 테스트(리더 `cargo test --locked`)에서의 실패율은 재지 않았어요.
- 스캔은 홀더를 문자열 표기(`HOLDERS` 목록)로만 알아봐서, 목록에 없는 표기(`Command::new("sh")` 등)로 홀더를 띄우면 스캔 밖이에요. 이 슬라이스는 그 목록을 넓히지 않았어요(범위 밖, 제품 스캔의 한계로 남겨요).
- 순위 상위 5개는 고치지 않았어요. 카드 기준 1(상위 20개 합계 절반)은 이 슬라이스에서 닿지 않아요.

### 기준

- AC-1: met — 위 순위 표와 방법·원자료.
- AC-2: met — `home_state_selective` 10개 통과(3.56초), 전후 표에서 두 홀더 테스트의 EOF 대기가 60초대에서 1.5~3.0초로 줄었고 고아 `sleep` 이 남지 않았어요.
- AC-3: met — `release_pr_three_os` 9개 통과, 같은 부하에서 실패 60/120 에서 0/120.
- AC-4: met — 뮤턴트가 모두 죽었고, 지운 단언 줄은 0개예요.
- AC-5: met — 스캔이 통과했고(5개 중 5개), 홀더 테스트 둘의 뮤턴트가 각각 죽었어요.
