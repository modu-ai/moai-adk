---
id: SPEC-20261125-slow-test-top2
title: "느린 테스트 순위 1·2위를 빠르게"
status: draft
revision: 1
created: 2026-11-25
---

# 느린 테스트 순위 1·2위를 빠르게

## Intent

`SPEC-20261125-slow-test-top` 의 순위 표(24시간, 로그 71개)에서 1위 `go_state_cli::a_receipt_above_256_kib_reverts_completely` 와 2위 `worktree_gc::tests::the_recheck_before_a_removal_refuses_files_and_an_install_folder` 를 같은 단언으로 빠르게 해요. 원인은 카드의 후보일 뿐이라서 먼저 `SPEC-20261124-test-cost-profile-git-path` 의 방법(번갈아 5회, 중앙값, 부하 기록)으로 재고, 잰 원인만 고쳐요.

## Requirements

- REQ-1: While the Result holds the cost of a ranked test, each row shall give the median of five alternating runs of the baseline binary and the repaired binary, with the load of each run.
- REQ-2: While `the_recheck_before_a_removal_refuses_files_and_an_install_folder` judges files and install folders and no process, its re-checks shall read an empty process table that belongs to the thread running the test.
- REQ-3: While one test thread holds an empty process table, a check on another thread of the same process shall read the table of its own thread, the real one unless that thread set one.
- REQ-4: While a debug build publishes a file with `MOAI_TEST_PUBLISH_SYNC` set to `skip`, the publish shall leave out the file sync and the directory sync and keep the write order, the temporary file and the rename.
- REQ-5: While a release build publishes a file, the publish shall keep both syncs whatever `MOAI_TEST_PUBLISH_SYNC` holds.
- REQ-6: While `a_receipt_above_256_kib_reverts_completely` starts the two commands that save the receipt, the test shall give only those two commands `MOAI_TEST_PUBLISH_SYNC`.
- REQ-7: While a test file or a unit test in the write scope is changed, it shall keep every assertion it had.

## Acceptance

- AC-1 (REQ-1): the Result gives the cost split of both tests as tables with the data they came from · verify: `moai spec check SPEC-20261125-slow-test-top2 --project .`
- AC-2 (REQ-2, REQ-3, REQ-7): the `worktree_gc` unit tests pass with the same names, and the guard tests of the thread table pass · verify: `cargo test --locked --offline --lib worktree_gc::tests`
- AC-3 (REQ-4, REQ-5, REQ-6, REQ-7): `go_state_cli` passes with the same test names and the 256 KiB test runs in seconds · verify: `cargo test --locked --offline --test go_state_cli`
- AC-6 (REQ-4, REQ-5, REQ-6): the two source scans of the switch pass in the unit tests of `src/apply.rs` · verify: `cargo test --locked --offline --lib -- apply::tests::the_publish_sync_switch_is_read_only_by_a_debug_build_function apply::tests::only_the_receipt_size_test_gives_the_commands_the_switch`
- AC-4 (REQ-7): no assertion line was removed from the changed test files and the unit-test module · verify: `test -z "$(git diff -U0 NeXT...HEAD -- tests/go_state_cli.rs src/worktree_gc.rs | grep -E '^-[^-].*assert')"`
- AC-5 (REQ-2, REQ-3, REQ-5, REQ-6): each new guard is killed by a mutant that removes the guarded line (the Result lists each `guard-mutant:` line) · verify: `test "$(grep -c '^- guard-mutant:' .moai/specs/SPEC-20261125-slow-test-top2/spec.md)" -ge 2`

## Scope

- write: `.moai/specs/SPEC-20261125-slow-test-top2/**`
- write: `tests/go_state_cli.rs`, `src/worktree_gc.rs`
- write (line lease, leader decision of the scope request on the card): `src/apply.rs`, the switch function, the two sync calls of `stage` and `sync_parent`, and the two source scans in its unit tests (a test file under `tests/` that names the manifest folder joins `tools/repo_scans.json`, which is not in this scope); `tests/common/**` stays unchanged
- write (line lease, second leader decision of the scope request on the card): `tools/repo_scans.json`, one new entry at the end of the list that registers the two source scans; the other entries and `tests/repo_scans_table.rs` stay unchanged
- read-only: everything else; a further path waits for a SCOPE-REQUEST note on the card

## Plan

1. 두 테스트의 원인을 번갈아 5회(1위는 단계별로 시간을 찍은 사본으로)로 잰다.
2. 2위: 단위 테스트 모듈에 스레드별 빈 프로세스 표를 두고 그 테스트만 쓴다.
3. 1위: 리더의 결정(카드 노트)에 따라 디버그 전용 스위치로 영수증 저장의 동기화를 건너뛴다. 스위치는 디버그 빌드 함수에서만 읽고, 이 테스트의 apply·revert 호출에만 싣고, 두 가지를 소스 텍스트를 읽는 테스트로 고정한다.
4. 같은 단언을 둔 채 가드는 뮤턴트로 확인하고, 번갈아 5회로 전후를 비교한다.

## Result

머리: 이 가지의 코드 커밋은 `30c7affe9e` 하나이고 `src/apply.rs`(스위치 함수와 sync 두 곳, 리더가 `card line-lease` 로 승인), `src/worktree_gc.rs`, `tests/go_state_cli.rs` 를 바꿔요. 측정은 2026-11-25 에 macOS arm64 디버그 빌드에서 했고, 도구와 원자료는 레인 워크트리의 `target/c482/` 에 있어요(카드가 닫히면 사라져요: `ab.py`, `ab-both.sh`, `phases.py`, `fsync_cost.py`, `lsof_cost.py`, `mutants.py`, `ab-first.jsonl`, `ab-second.jsonl`, `phases-base.jsonl`). 순위의 출처는 `SPEC-20261125-slow-test-top` 의 Result 예요(24시간, 로그 71개, 1위 중앙값 81.76초, 2위 38.87초는 부하 25~50 에서 병렬로 돈 값).

### 원인 (카드의 후보를 재서 가린 결과)

1위 `a_receipt_above_256_kib_reverts_completely`: 시간은 제품 바이너리 안에 있고 테스트 쪽에는 없어요. 단계에 시간을 찍은 사본(`tests/go_state_cli.rs` 에 `PHASE` 줄만 더한 것, 되돌렸어요)을 다섯 번 돌린 중앙값이에요(`tools/test_home_guard.sh` 아래, 부하 8.8~12.3).

| 단계 | 중앙값(초) | 최소~최대 |
|---|---|---|
| 픽스처(`many_specs_fixture(1000)`) | 0.10 | 0.09~0.25 |
| 트리 읽기(전) | 0.05 | 0.04~0.06 |
| `go-state-plan` | 0.24 | 0.22~0.32 |
| `go-state-apply` | 21.74 | 20.52~33.33 |
| `go-state-revert-plan` | 0.12 | 0.11~0.17 |
| `go-state-revert` | 11.73 | 11.38~15.38 |
| 트리 비교(후) | 0.09 | 0.08~0.10 |
| 테스트 전체 | 34.32 | 32.92~49.68 |

- 카드의 후보 "더 적고 긴 경로로 같은 크기에 닿기"는 기각했어요: 테스트 쪽 단계는 합쳐 0.6초이고, 테스트가 `restored.len() > 1000` 을 단언해서 항목을 줄이면 단언을 고쳐야 해요(카드: 단언 그대로).
- 확인된 원인: 영수증을 단계마다 다시 저장하고(`go_state.rs` 의 `save_receipt` → `apply.rs` 의 `publish_replace_limited`), 저장 한 번이 `stage` 의 `file.sync_all()` 과 `sync_parent` 의 `dir.sync_all()` 이에요. macOS 에서 러스트의 `sync_all` 은 `fcntl(F_FULLFSYNC)` 예요. 262 KB 파일 저장 1회를 200번 잰 값(`fsync_cost.py`, 부하 같은 시간대):

| 저장 방식 | 1회(ms) |
|---|---|
| 동기화 없음 | 0.13 |
| 일반 `fsync` | 0.22 |
| `F_FULLFSYNC` | 8.88 |

  항목이 1000개 넘으니 apply·revert 가 저장만으로 약 20초를 써요.

2위 `worktree_gc::tests::the_recheck_before_a_removal_refuses_files_and_an_install_folder`: 시간은 프로세스 표 읽기에 있어요. 테스트가 `busy_now`/`recheck_entry` 를 13번 부르고(소스에서 센 값: `probe` 호출 10번, `recheck_entry` 3번), 호출마다 `read_cwds` 와 `read_open` 이 기계 전체를 `lsof` 로 읽어요. 제품과 같은 인자(`LSOF_CHEAP` 포함)로 다섯 번씩 잰 `lsof` 한 번(`lsof_cost.py`, 부하 15.9):

| `lsof` 호출 | 중앙값(초) | 출력 크기 |
|---|---|---|
| 작업 폴더(`-a -d cwd -F fn`) | 0.37 | 43,704 바이트 |
| 열린 파일(`-F fn`) | 0.43 | 2,710,940 바이트 |

0.80초 × 13번 = 10.4초이고, 같은 기계에서 이 테스트를 단독으로 한 번 잰 값 11.3초(부하 9.3)의 약 92%예요. 카드의 후보(프로세스 표 주입)가 맞아요. 이 테스트는 파일과 설치 폴더를 판정할 뿐 프로세스를 판정하지 않아요.

### 고친 것

- 2위: `src/worktree_gc.rs` 에 스레드별 빈 프로세스 표를 두었어요(`use_empty_table`, 테스트 빌드에서만, 값은 그 스레드 안에서만 보여요). 환경변수 `MOAI_TEST_PROCESS_TABLE` 을 단위 테스트에서 바꾸면 `cargo test` 의 다른 스레드가 같이 영향을 받아서 쓰지 않았어요. 테스트는 맨 앞에서 빈 표를 잡고 `read_cwds() == Some([])` 를 단언해요(추가한 단언이에요). 새 테스트 `an_empty_table_belongs_to_the_thread_that_holds_it` 가 다른 스레드는 자기 표를 읽는다는 것과 표가 drop 에서 돌아간다는 것을 고정해요.
- 1위: `src/apply.rs` 에 디버그 빌드 전용 스위치 `MOAI_TEST_PUBLISH_SYNC=skip` 를 더했어요(리더 결정 노트 8689, 조건 5개). 켜면 `stage` 의 `file.sync_all` 과 `sync_parent` 의 `dir.sync_all` 두 호출만 건너뛰고 쓰기 순서·임시 파일·`rename` 은 그대로예요. 읽는 함수는 `#[cfg(debug_assertions)]` 이고 릴리스 빌드의 짝 함수는 `false` 를 돌려줘요. 이 테스트의 `go-state-apply`·`go-state-revert` 호출만 `UNSYNCED` 로 스위치를 실어요(`tests/common` 에는 안 넣었어요). 소스를 읽는 테스트 둘이 고정해요(`src/apply.rs` 의 단위 테스트예요. 처음에는 `tests/go_state_cli.rs` 에 두었다가 `check affected` 가 FAIL 이라 옮겼어요: `CARGO_MANIFEST_DIR` 로 저장소를 걷는 `tests/*.rs` 파일은 `tools/repo_scans.json` 에 올라야 하고 그 목록은 이 카드의 범위 밖이에요): `the_publish_sync_switch_is_read_only_by_a_debug_build_function`(`src/` 어디서든 스위치 이름을 읽는 줄은 `#[cfg(debug_assertions)]` 함수 안에 있어야 해요), `only_the_receipt_size_test_gives_the_commands_the_switch`(그 테스트의 apply·revert 호출이 `UNSYNCED` 를 싣고 `tests/common/mod.rs` 에는 이름이 없어요).
- 지운 단언은 0개예요: 두 파일의 지운 줄 23개 중 `assert` 가 든 줄은 없어요(AC-4 의 verify 줄을 그대로 실행해 exit 0, 스크립트 `target/c482/ac4.sh`).

### 전후 표 (기준 2)

방법(`SPEC-20261124-test-cost-profile-git-path` 와 같아요): `tools/test_home_guard.sh <바이너리> --exact <이름> --nocapture` 로 테스트 하나씩, 기준선과 수정본을 번갈아 5회씩 돌려 중앙값을 냈어요(`ab.py`, 첫 실행 한 번은 재지 않아요). `t_exit` 는 프로세스가 끝난 시각이고 `t_eof` 는 출력 파이프가 EOF 에 닿은 시각인데 이 두 테스트는 둘이 같았어요. 모든 실행의 종료 코드가 0 이었어요.

| 테스트 | 전 중앙값 | 후 중앙값 | 배율 | 전 반복별 | 후 반복별 | 부하(load1) |
|---|---|---|---|---|---|---|
| 1위 `a_receipt_above_256_kib_reverts_completely` | 38.16 | 10.03 | 3.8배 | 35.3 37.0 38.8 38.2 38.2 | 8.1 10.0 8.4 10.4 12.3 | 11.0~14.5 |
| 2위 `the_recheck_before_a_removal_refuses_files_and_an_install_folder` | 17.96 | 0.67 | 27배 | 18.3 15.8 18.0 18.4 17.6 | 1.0 0.6 1.3 0.6 0.7 | 17.1~19.3 |

- 두 테스트 모두 중앙값이 절반 이하예요. 전과 후는 같은 시간대에 번갈아 돌려서 부하 변동이 두 쪽에 같이 걸려요.
- "전" 쪽 테스트 바이너리는 NeXT 끝(`0babbe3566`)에서 빌드한 것이고 1위 테스트가 부르는 `moai-cli` 는 두 쪽 모두 수정본이에요. 스위치를 안 실으면 동작이 같아서 "전"의 시간으로 읽어요(스위치 없는 호출은 수정 전과 같은 코드를 지나요).
- 1위의 남은 10초는 쪼개지 않았어요(not_run). 동기화를 뺀 뒤에 남는 영수증 직렬화·다시 읽기 몫이라고 짐작하지만 재지 않았어요.
- 2위의 전 값 18초는 이 기계가 부하 17~19 일 때예요(단독 한 번은 11.3초). 순위 표의 38.9초는 부하 25~50 이었어요.

### 뮤턴트 (`target/c482/mutants.py`, 슬롯 한 번, 파일은 복사본으로 원복해 바이트 단위로 같음을 확인)

여섯 개 모두 `cargo test --lib` 로 돌렸어요(스캔 둘이 `src/apply.rs` 의 단위 테스트이고 실행 때 소스 텍스트를 읽어서, `tests/` 파일을 바꾸는 G1~G3 은 다시 빌드하지 않아요). 이름은 `apply::tests::` 와 `worktree_gc::tests::` 아래이고, 여섯 개 모두 컴파일됐어요(로그에서 의도한 단언 메시지를 읽었어요).

- guard-mutant: only_the_receipt_size_test_gives_the_commands_the_switch | the apply call of the receipt-size test loses UNSYNCED | exit 101
- guard-mutant: only_the_receipt_size_test_gives_the_commands_the_switch | the revert call of the receipt-size test loses UNSYNCED | exit 101
- guard-mutant: only_the_receipt_size_test_gives_the_commands_the_switch | tests/common/mod.rs names the switch | exit 101
- guard-mutant: the_publish_sync_switch_is_read_only_by_a_debug_build_function | the switch is read in a function without #[cfg(debug_assertions)] | exit 101
- guard-mutant: the_recheck_before_a_removal_refuses_files_and_an_install_folder | removed `let _table = use_empty_table();` | exit 101
- guard-mutant: an_empty_table_belongs_to_the_thread_that_holds_it | the empty table is one flag of the whole process | exit 101

### 제품 후속 후보 (결정은 사용자 몫, 이 카드에서 하지 않았어요)

측정이 제품 쪽 사실을 보여 줬어요: 항목 1000개짜리 `go-state-apply` 는 영수증을 단계마다 `F_FULLFSYNC` 로 저장해 중앙값 21.74초, `go-state-revert` 는 11.73초가 걸려요(이 기계, 디버그 빌드). 릴리스 빌드의 같은 시간은 재지 않았어요(not_run: 같은 호출이라 비슷할 것이라고만 읽어요). 중간 저장의 내구성을 줄일지(마지막 저장만 동기화하기)는 정전 때 영수증이 한두 단계 뒤처질 수 있는 제품 결정이라 리더 결정 노트 8689 대로 아침 보고에 올려요.

### 한계 (not_run 포함)

- 이 슬라이스는 순위 1·2위만 다뤄요. 3~5위(`update_hooks_cli`, `update_user_layer_cli`, `update_home`)는 하지 않았어요.
- 측정은 이 기계 하나, 디버그 빌드예요. `F_FULLFSYNC` 는 macOS 의 `sync_all` 이라 Linux·Windows 의 이득은 재지 않았어요(Linux 의 `sync_all` 은 `fsync`).
- 스위치는 시간만 줄여요. 스위치를 안 싣는 변경은 느릴 뿐 통과하므로 가드는 소스 텍스트를 읽는 스캔이에요(리더 결정). 스캔이 `src/apply.rs` 의 단위 테스트라서 `tests/go_state_cli.rs` 나 `tests/common/mod.rs` 만 바뀐 범위를 `check affected` 가 어떻게 고르는지는 읽지 않았어요(not_run). 대신 `tools/repo_scans.json` 의 새 항목이 두 파일을 범위로 가지고 있어서 `moai repo-scans` 가 머지마다 스캔 둘을 골라요(리더 결정 8704). 건너뛴 동기화가 실제로 일어나지 않는지는 시간으로만 보였어요.
- 2위의 가드 단언 `read_cwds() == Some([])` 는 실제 표를 읽는 기계에서는 실패하는 값이라 뮤턴트가 죽어요. 표를 읽을 수 없는 기계(`lsof` 없음)에서도 `Some([])` 가 아니라서 마찬가지로 죽어요.
- `cargo test`(한 프로세스의 여러 스레드) 아래에서 이 파일의 단위 테스트 전체를 한 번 돌려 보는 일은 하지 않았어요(수정한 두 테스트와 새 테스트를 직접 돌렸어요). 전체는 `check affected` 의 nextest 가 돌려요.

### 기준

- AC-1: met — 위 표들과 방법·원자료.
- AC-2: met — `worktree_gc` 단위 테스트가 같은 이름으로 통과하고(`the_recheck_...`, `an_empty_table_belongs_...`), 뮤턴트 둘이 죽어요.
- 등록: 두 번째 `check affected` 도 FAIL 이었어요(헤드 `fda32b3531`, `repo_scans_table every_walking_test_file_is_listed_or_exempt` 한 개, `src/apply.rs` 가 `CARGO_MANIFEST_DIR` 와 `read_dir(` 를 함께 가져 저장소를 걷는 파일로 잡혔어요). 리더 결정 8704 에 따라 `tools/repo_scans.json` 에 항목 `publish-sync-switch-read-only-in-debug-builds` 하나를 끝에 더했어요(`always`, 규칙 파일 `src/apply.rs`, 범위 `src/**/*.rs`·`tests/go_state_cli.rs`·`tests/common/mod.rs`, 71개에서 72개, 기존 항목과 `tests/repo_scans_table.rs` 는 그대로). 더하기만임은 스크립트가 옛 JSON 의 모든 항목이 같음을 비교해 확인했어요. 등록표 테스트 전체(`--test repo_scans_table`)가 33개 통과했고 항목의 확인 명령 2개도 통과했어요.
- AC-6: met — 두 스캔이 `apply::tests` 에서 통과해요(기준선 실행 4개 통과, 로그 `target/c482/baseline2-lib.log`).
- AC-3: met — `go_state_cli` 의 1위 테스트가 통과하고(9.09초) 1위 테스트는 38.16초에서 10.03초가 됐어요(파일 전체는 `check affected` 가 돌려요).
- AC-4: met — 지운 단언 줄 0개(verify 줄 exit 0).
- AC-5: met — 위 `guard-mutant:` 줄 여섯 개가 모두 exit 101 이에요.
