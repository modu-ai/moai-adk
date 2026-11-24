---
id: SPEC-20261125-test-switch-release-scan
title: "테스트용 스위치는 릴리스 빌드에서 읽히지 않아요"
status: draft
revision: 1
created: 2026-11-25
---

# 테스트용 스위치는 릴리스 빌드에서 읽히지 않아요

## Intent

`MOAI_TEST_*` 스위치는 테스트가 제품 안의 한 지점을 흔드는 환경 변수라서 릴리스 빌드가 읽으면 안 돼요. 어젯밤 감사 두 건에서 같은 종류가 나왔어요: 한 스위치가 `cfg(debug_assertions)` 없이 모든 빌드에서 읽혔고, 다른 스위치는 디버그 전용이지만 그것을 지키는 스캔이 그 이름 하나만 봤어요. `Cargo.toml` 의 release 나 dist 프로파일이 `debug-assertions` 를 켜면 모든 스위치가 릴리스에서 읽혀요. 그래서 먼저 `src` 의 이름마다 읽는 곳과 컴파일되는 빌드를 재고, 이름 목록이 아니라 접두사 `MOAI_TEST` 전체를 보는 스캔 하나와 프로파일 검사를 더해요. 예외 목록은 두지 않아요.

## Requirements

- REQ-1: While a file under `src/` spells the prefix `MOAI_TEST` in code, the scan shall require that line to sit under a `cfg` attribute that only a debug build or a test build compiles.
- REQ-2: While a constant of `src/` holds a spelling of the prefix `MOAI_TEST`, the scan shall apply REQ-1 to every other line of `src/` that uses the constant.
- REQ-3: While the scan reads `src/`, it shall find at least ten distinct switch names so that a wrong root or a changed layout cannot make it pass on nothing.
- REQ-4: While `Cargo.toml` or `.cargo/config.toml` turns debug assertions on for the `release` or `dist` profile, the scan shall fail.
- REQ-5: While the `dist` profile does not inherit `release`, the scan shall fail.
- REQ-6: While the table `tools/repo_scans.json` lists the scan, its one entry shall hold `src/`, `Cargo.toml`, `.cargo/config.toml` and the scan's own file in its scope.
- REQ-7: While the Result holds the measurement, it shall list every `MOAI_TEST_` name of `src/` with its read site, the `cfg` that compiles the site and the build that holds the name, with the command and its output.
- REQ-8: IF a switch read in a release build serves a product behavior, THEN the lane shall leave that switch where it is and ask for the decision on the card.
- REQ-9: While the table `tools/repo_scans.json` lists the scan, every other entry of the table shall stay as it was.

## Acceptance

- AC-1 (REQ-7, REQ-8): the Result gives the measurement table with the commands and the answer to the product-behavior question · verify: `moai spec check SPEC-20261125-test-switch-release-scan --project .`
- AC-2 (REQ-1, REQ-2, REQ-3): the source scan passes on the tree and its reader tells a gated read from an ungated one on fixtures · verify: `cargo test --locked --offline --test test_switch_release_scan -- every_test_switch_read_in_src_is_debug_or_test_only the_gate_reader_tells_a_gated_read_from_an_ungated_one`
- AC-3 (REQ-4, REQ-5): the profile scan passes on the manifest and its reader tells an on switch from an off one on fixtures · verify: `cargo test --locked --offline --test test_switch_release_scan -- the_release_and_dist_profiles_keep_debug_assertions_off the_profile_reader_tells_an_on_switch_from_an_off_one`
- AC-4 (REQ-6): the table lists the scan and the table test passes · verify: `cargo test --locked --offline --test repo_scans_table`
- AC-6 (REQ-9): the table changed by additions only · verify: `test "$(git diff --numstat NeXT...HEAD -- tools/repo_scans.json | cut -f2)" = 0`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): each new guard is killed by a mutant (the Result lists each `guard-mutant:` line) · verify: `test "$(grep -c '^- guard-mutant:' .moai/specs/SPEC-20261125-test-switch-release-scan/spec.md)" -ge 6`

## Scope

- write: `.moai/specs/SPEC-20261125-test-switch-release-scan/**`
- write: `tests/test_switch_release_scan.rs`
- write (line lease): `tools/repo_scans.json`, one new entry at the end of the list; the other entries and `tests/repo_scans_table.rs` stay unchanged
- read-only: everything else, `src/` and `Cargo.toml` included; a switch that has to move waits for the decision on the card

## Plan

1. `src` 의 `MOAI_TEST_` 이름마다 읽는 곳과 컴파일되는 빌드를 잰다: 소스의 `cfg` 줄을 읽고, 릴리스와 디버그 바이너리의 문자열을 대조한다.
2. 접두사 `MOAI_TEST` 가 보이는 모든 줄이 디버그 전용(`cfg(debug_assertions)`, `cfg(test)`) 코드 안에 있는지 보는 스캔을 새 통합 테스트 파일에 둔다. 상수가 이름을 쥐면 그 상수를 쓰는 줄도 같은 규칙으로 본다.
3. `Cargo.toml` 과 `.cargo/config.toml` 에서 release 와 dist 프로파일이 `debug-assertions` 를 켜지 않고 dist 가 release 를 물려받는지 보는 스캔을 같은 파일에 둔다.
4. 두 스캔을 `tools/repo_scans.json` 에 항목 하나로 등록한다(줄 임대, 더하기만).
5. 가드마다 뮤턴트로 죽는지 본다.
6. 묶음 착지의 확인 프로그램 스위치를 고치는 후속 수리는 이 카드가 건드리지 않고 스캔의 예외로도 두지 않는다. 그 수리가 `cfg(debug_assertions)` 아래에서 스위치를 읽으면 이 스캔을 통과하고, 그러지 않으면 이 스캔이 막는다. 착지 순서는 어느 쪽이 먼저여도 되고 나중에 오는 쪽이 앞선 쪽의 결과를 따른다.

## Result

머리: 이 가지의 코드 커밋은 `f42f8b9d9b` 하나이고 새 통합 테스트 `tests/test_switch_release_scan.rs` 와 `tools/repo_scans.json` 의 항목 하나를 더해요. `src/` 와 `Cargo.toml` 은 바꾸지 않았어요. 측정은 2026-11-25 에 macOS arm64 에서 NeXT 끝 `0babbe3566` 의 소스로 했고, 도구와 원자료는 레인 워크트리의 `target/c484/` 에 있어요(카드가 닫히면 사라져요: `ctx.py`, `show.py`, `release-build.sh`, `debug-build.sh`, `release.strings`, `debug.strings`, `mutants.py`, `mutants.jsonl`, `probe-other.sh`).

### 측정: 이름마다 읽는 곳과 컴파일되는 빌드

명령과 출력:

- `grep -rno "MOAI_TEST_[A-Z0-9_]*" src | sort | uniq -c` 는 30줄에 서로 다른 이름 15개를 내요(`ctx.py` 가 줄마다 둘러싼 함수와 `cfg` 줄을 읽어요).
- `cargo build --release --locked --bin moai-cli` 는 종료 코드 0 이고(1분 14초), `strings -a target/release/moai-cli | grep -o "MOAI_TEST[A-Z0-9_]*"` 는 아무것도 내지 않아요(0개, 문자열 67,026줄).
- 대조군: `cargo build --locked --bin moai-cli`(디버그)의 `strings -a target/debug/moai-cli` 는 15개 이름을 모두 내요(한 줄에 붙어 나온 두 쌍은 `strings` 가 이어 읽은 것).

| 이름 | 읽는 곳 | 읽는 줄을 컴파일하는 조건 | 디버그 문자열 | 릴리스 문자열 |
|---|---|---|---|---|
| `MOAI_TEST_BACKUP_STAMP` | `src/update.rs:3079` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_CLEANUP_DEVICE` | `src/update_cleanup.rs:837` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_FAULT` | `src/update.rs:478`, `:491`, `:504`, `src/go_state.rs:4069` | 문장 위 `#[cfg(debug_assertions)]` 넷(`src/update/extra_skill.rs:953` 은 주석) | 있음 | 없음 |
| `MOAI_TEST_INIT_FAIL_AFTER_PROFILE` | 상수 `src/launcher_init.rs:34`, 읽기 `:382` | 읽기 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_INIT_HOLD_MS` | 상수 `src/launcher_init.rs:31`, 읽기 `:375` | 읽기 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_INSTALL_FAULT` | 상수 `src/home_install_apply.rs:48`, 읽기 `:62` | 함수 위 `#[cfg(debug_assertions)]`(짝 함수는 `None`) | 있음 | 없음 |
| `MOAI_TEST_LSOF_TIMEOUT_MS` | `src/worktree_gc.rs:397` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_MACHINE` | `src/heavy_slot.rs:350` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_MIRROR_FILE_LIMIT` | `src/go_state.rs:118` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_PROBE_TIMEOUT_MS` | `src/install.rs:44` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_PROCESS_CHECK` | `src/home_state.rs:468` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_PROCESS_TABLE` | `src/worktree_gc.rs:365` | 함수 위 `#[cfg(debug_assertions)]`(짝 함수는 `None`) | 있음 | 없음 |
| `MOAI_TEST_SESSION_REGISTRY_HOLD_MS` | 상수 `src/lane_registry.rs:44`, 읽기 `:859` | 읽기 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |
| `MOAI_TEST_STATE_RETIRE_FAULT` | 상수 `src/home_state_retire.rs:42`, 읽기 `:503` | 함수 위 `#[cfg(debug_assertions)]`(짝 함수는 `None`) | 있음 | 없음 |
| `MOAI_TEST_UNRECORDED_LIMIT` | `src/go_state.rs:146` | 문장 위 `#[cfg(debug_assertions)]` | 있음 | 없음 |

첫 결과: 릴리스 바이너리에서 읽히는 이름은 0개예요. 15개 모두 읽는 줄이 디버그 전용 `cfg` 아래에 있고 릴리스 바이너리에 문자열이 없어요. 그래서 옮긴 스위치도 없고, 제품 동작에 쓰이는 스위치가 나오지 않아서 결정을 묻지 않았어요. 릴리스 쪽 요인도 읽었어요: `Cargo.toml` 에는 `[profile.dev]`, `[profile.test]`, `[profile.dev.package]` 와 `dist` 가 `inherits = "release"` 로 `lto` 만 더하는 것뿐이고 `debug-assertions` 는 어디에도 없어요. `.cargo/config.toml` 은 `[env]` 와 `[build] incremental = false` 뿐이에요. 둘째 사례(`MOAI_TEST_PUBLISH_SYNC`)는 NeXT 끝의 `src` 에 아직 없어요(그것을 더하는 가지가 착지 대기예요). 그 가지의 머리(`827a599b91`)의 `src`, `Cargo.toml`, `.cargo/config.toml` 에 같은 스캔 코드를 따로 컴파일해 돌리니(아래 순서 절과 같은 방법) 4개 모두 통과했어요: 그 가지에서는 이름이 16개가 되고, 읽는 줄은 `#[cfg(debug_assertions)]` 함수 안이며, 이름을 조각으로 이어 쓰는 줄은 `cfg(test)` 단위 테스트 모듈 안이에요.

### 스캔이 하는 것

- 소스 스캔(`every_test_switch_read_in_src_is_debug_or_test_only`): 접두사 `MOAI_TEST` 가 보이는 코드 줄마다 그 줄, 둘러싼 블록, 둘러싼 항목 중 하나의 바로 위에 `#[cfg(debug_assertions)]`, `#[cfg(test)]` 또는 둘을 `any` 로 묶은 줄이 있어야 해요. 부정(`not`)과 그 밖의 `cfg` 는 통과하지 못해요. 이름이 상수에 들어 있으면(여러 줄 배열 포함) 그 상수를 쓰는 다른 모든 줄에 같은 규칙을 걸고, `use` 줄은 쓰임이 아니에요. 이름 목록도 예외 목록도 없어요. `src` 가 50개 이상 파일과 10개 이상 이름을 못 찾으면 실패해요(아무것도 안 보고 통과하는 길을 막아요).
- 프로파일 스캔(`the_release_and_dist_profiles_keep_debug_assertions_off`): `Cargo.toml` 과 `.cargo/config.toml` 에서 release·dist 프로파일(하위 표, 점 표기, 인라인 표 포함)이 `debug-assertions` 를 끄지 않은 값으로 두는 줄, `rustflags` 의 켠 값, `CARGO_PROFILE_{RELEASE,DIST}_DEBUG_ASSERTIONS` 를 잡고, `dist` 가 `release` 를 물려받지 않으면 실패해요.
- 두 읽는 쪽은 고정 텍스트 조각으로 따로 시험해요(`the_gate_reader_tells_a_gated_read_from_an_ungated_one` 22꼴, `the_profile_reader_tells_an_on_switch_from_an_off_one` 15꼴).

### 묶음 착지 수리와의 순서

`MOAI_TEST_LAND_BATCH_CHECK` 는 NeXT 끝의 `src` 에 없어요(`grep` 0줄). 그 수리는 이 카드가 건드리지 않고 스캔의 예외로도 두지 않았어요. 수리 가지(`spec/card-select-batch-land-followup-20261125`, 커밋 `690837446d`)의 `git archive` 소스와 수리 전 소스(`690837446d^`)에 같은 스캔 코드를 `rustc --test` 로 따로 컴파일해 돌렸어요(`probe-other.sh`): 수리 가지는 4개 모두 통과했고, 수리 전은 소스 스캔이 `src/ops/card.rs:1933: spells MOAI_TEST outside debug-only code: let check_cmd = env("MOAI_TEST_LAND_BATCH_CHECK")...` 로 실패했어요(감사가 찾은 바로 그 줄). 그래서 착지는 어느 쪽이 먼저여도 되고, 이 스캔이 먼저 들어간 뒤 수리가 `cfg` 없이 읽으면 막혀요.

### 가드 뮤턴트 (22개, 모두 KILLED)

한 번의 슬롯 호출로 돌렸어요(`mutants.py`, 판정기는 저장된 실패 로그 꼴로 먼저 시험했어요). 소스 텍스트 뮤턴트(M)는 먼저 빌드한 테스트 실행 파일을 다시 돌려요(스캔이 실행 때 텍스트를 읽어서 다시 빌드가 없어요). 스캐너 뮤턴트(G)는 스캔 파일을 바꿔 다시 빌드했어요. 편집은 모두 정확히 한 번이고 파일은 바이트 단위로 되돌려 비교했으며, 끝의 복원 실행은 4개 통과예요(`restored run exit 0`). 표시한 곳 말고는 기대한 테스트 하나만 실패했어요. 되돌리고 난 뒤 `git status --short` 는 비어 있어요.

- guard-mutant: every_test_switch_read_in_src_is_debug_or_test_only | the `#[cfg(debug_assertions)]` line above the `MOAI_TEST_PROCESS_CHECK` read in `src/home_state.rs` | exit 101
- guard-mutant: every_test_switch_read_in_src_is_debug_or_test_only | the `#[cfg(debug_assertions)]` line above `fault_knob` in `src/home_install_apply.rs` (the read goes through a constant) | exit 101
- guard-mutant: every_test_switch_read_in_src_is_debug_or_test_only | `cfg(debug_assertions)` made `cfg(not(debug_assertions))` above the `MOAI_TEST_PROBE_TIMEOUT_MS` read in `src/install.rs` | exit 101
- guard-mutant: every_test_switch_read_in_src_is_debug_or_test_only | a new `format!("MOAI_TEST_{x}")` function with no `cfg` appended to `src/go_state.rs` | exit 101
- guard-mutant: every_test_switch_read_in_src_is_debug_or_test_only | the name search of the scan changed so that it finds no names (the floor of ten names) | exit 101
- guard-mutant: the_release_and_dist_profiles_keep_debug_assertions_off | `debug-assertions = true` added to `[profile.dist]` of `Cargo.toml` | exit 101
- guard-mutant: the_release_and_dist_profiles_keep_debug_assertions_off | `inherits = "release"` of the dist profile changed to `"dev"` | exit 101
- guard-mutant: the_release_and_dist_profiles_keep_debug_assertions_off | a `[profile.release]` section with `debug-assertions = true` appended to `Cargo.toml` | exit 101
- guard-mutant: the_release_and_dist_profiles_keep_debug_assertions_off | a `rustflags` line with `debug-assertions=on` added to `.cargo/config.toml` | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | `#[cfg(not(debug_assertions))]` added to the accepted attribute list | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | the climb to the enclosing blocks removed (`while false`) | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | the second pass over constants emptied | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | the `use` line exemption of the constant pass removed | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | any `#[cfg(` attribute accepted instead of the debug list | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | comment lines counted as code | exit 101
- guard-mutant: the_gate_reader_tells_a_gated_read_from_an_ungated_one | the `:` type check removed so that a `const fn` counts as a constant definition | exit 101
- guard-mutant: the_profile_reader_tells_an_on_switch_from_an_off_one | the `profile.dist` section dropped from the release side | exit 101
- guard-mutant: the_profile_reader_tells_an_on_switch_from_an_off_one | the dist inherits value no longer compared with `release` | exit 101
- guard-mutant: the_profile_reader_tells_an_on_switch_from_an_off_one | the `CARGO_PROFILE_RELEASE_DEBUG_ASSERTIONS` spelling no longer matched | exit 101
- guard-mutant: the_profile_reader_tells_an_on_switch_from_an_off_one | the `rustflags` test removed from the check | exit 101
- guard-mutant: the_profile_reader_tells_an_on_switch_from_an_off_one | trailing `#` comments no longer cut off | exit 101
- guard-mutant: the_profile_reader_tells_an_on_switch_from_an_off_one | `false` removed from the off values | exit 101

두 뮤턴트는 기대한 테스트 말고도 소스 스캔을 실패시켰어요(그 뮤턴트가 실제 트리에서도 잘못 읽게 만들기 때문이에요): 올라가기를 없앤 것과 주석을 코드로 센 것. 이름 찾기를 막은 뮤턴트도 같은 두 테스트가 실패했는데, 이름이 0개면 소스 스캔에서 실패할 수 있는 것은 바닥 단언뿐이에요(메시지는 읽지 않고 그렇게 추론했어요).

### 남은 위험과 한계

- 스캔이 따라가지 못하는 것은 테스트 파일 머리 주석에 적었어요: `concat!` 처럼 접두사를 한 번도 통째로 쓰지 않는 이름, 다른 파일의 `#[cfg(test)] mod x;` 가 통째로 막은 파일, `use ... as` 로 만든 상수 별칭, rustfmt 가 두지 않을 모양의 코드(들여쓰기를 읽어요), 줄 끝 주석 안의 이름(코드로 세므로 닫힌 쪽으로 실패해요). 이런 모양이 생기면 스캔이 실패하거나, 스캔을 고쳐야 해요.
- 문자열 대조는 `moai-cli` 바이너리만 했어요. `moai` 바이너리는 같은 `main.rs` 와 같은 라이브러리를 쓰지만 빌드하지 않았어요(`not_run`). CI 의 `dist` 프로파일 빌드도 하지 않았고(`not_run`), `.github/workflows/*`, `scripts/*`, `build.rs` 에 `debug-assertions`, `debug_assertions`, `RUSTFLAGS`, `rustflags` 가 없다는 것은 `grep -ln` 으로 읽었고(결과 없음), `dist-workspace.toml` 은 전체를 읽었지만 스캔은 그 파일들을 보지 않아요. 저장소 전체를 한 번에 훑는 `grep -r` 는 셸 글롭 오류로 돌지 않았어요(`not_run`).
- 새 스위치를 이름만 `MOAI_TEST` 접두사 없이 만들면 이 스캔 밖이에요(범위 밖: 이름 바꾸기와 새 스위치).
- 기존 스캔 `the_test_seams_are_read_only_in_debug_builds`(이름 3개)는 그대로 두었어요. 이 스캔이 그것을 포함하지만 지우는 일은 이 카드의 범위가 아니에요.

### 기준 (AC)

- AC-1: met. 위 측정표와 명령, 제품 동작 스위치가 없다는 답이 이 Result 에 있어요.
- AC-2: met. `cargo test --locked --offline --test test_switch_release_scan` 4개 통과(소스 스캔과 고정 조각 22꼴 포함).
- AC-3: met. 같은 실행의 프로파일 스캔과 고정 조각 15꼴 통과.
- AC-4: met. `cargo test --locked --offline --test repo_scans_table` 33개 통과, `tools/repo_scans.json` 에 항목 `test-switch-prefix-read-only-in-debug-builds` 가 `src/`(.rs), `Cargo.toml`, `.cargo/config.toml`, 스캔 파일을 범위로 해요.
- AC-5: met. 22개 뮤턴트가 모두 죽었어요(위 줄).
- AC-6: met. `git diff --numstat NeXT...HEAD -- tools/repo_scans.json` 은 `34	0` 이라서 지운 줄이 없어요.
