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
