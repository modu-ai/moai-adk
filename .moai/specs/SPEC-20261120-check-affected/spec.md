---
id: SPEC-20261120-check-affected
title: "check affected — pick and run the tests a change can break"
status: draft
revision: 1
created: 2026-11-20
---

# check affected — pick and run the tests a change can break

## Intent

`moai check affected <base>..<head>` 는 바뀐 경로에서 그 변경이 깨뜨릴 수 있는 시험을 스스로 고르고, 검사를 돌리고, 한 줄로 결과를 알려요. 레인이 READY 를 쓰기 전과 리더가 병합한 head 를 확인할 때 같은 명령을 써요.

지금은 사람이 시험 범위를 손으로 좁혀요. 같은 영역 이름을 가진 시험(`src/update_models.rs` 를 바꿨는데 `update_command_cli` 를 안 돌림), `tools/repo_scans.json` 이 묶어 둔 저장소 전체 검사(`home_isolation_scan`), 영역별 파이썬 시험(`update_tux_tty_test`), 이 호스트에서만 도는 `cfg(target_os)` 시험, `--bins` 의 도움말 시험을 놓쳤고, 한 릴리즈 PR 에서 CI 가 두 번 실패하고 통합이 세 번 다시 돌았어요. 고르는 규칙을 한 곳에 코드로 두면 레인마다 다른 목록을 쓰는 일이 없어요.

이 명령은 어떤 MoAI 프로젝트에서도 같은 규칙으로 돌아야 해요. 통합 가지 이름이나 카드 번호를 코드에 두지 않아요.

## Non-goals

- 시험의 통과 여부를 판정하거나 PASS 를 선언하지 않아요: 독립 감사와 CI 는 그대로예요. 이 명령은 돌린 것과 종료 코드를 알릴 뿐이에요.
- `bundle verify` 의 확인 목록을 바꾸지 않아요: 그쪽은 묶음 가지를 한 번에 확인하는 별개의 기록(`run.json`)이고, 나중에 이 선택 규칙을 부를지는 따로 정해요.
- 시험을 고치거나 시험 도구(cargo, rustfmt, python)를 설치하지 않아요.
- 한 시험 안의 어떤 케이스가 바뀐 줄에 닿는지 가리지 않아요: 시험 대상(파일) 단위로만 골라요.

## Requirements

- REQ-1: When a range changes a path under `src/`, the selector shall select every Rust test target of `tests/` whose name starts with the path's area or shares the area's first word · failure: none — the selection is a pure function of the changed paths and the file lists.
- REQ-2: When a range changes a file `tests/<name>.rs`, `tests/<name>_test.py`, a file under `tests/fixtures/<dir>/` or a shared test helper, the selector shall select that target, the targets named like `<dir>` or every target that declares the helper · failure: none — the selection is a pure function of the changed paths and the file lists.
- REQ-3: When the project has `tools/repo_scans.json`, the selector shall select every Rust target and Python module that file names whatever the range changes · failure: none — an unreadable file makes the selector name it and select every `tests/` target, and a project without the file reports `scans=none`.
- REQ-4: When the selection is made, the selector shall add the library tests and the binary tests of the crate · failure: none — these two steps are fixed.
- REQ-5: When a range changes a path under `src/`, the selector shall select the Python modules `tests/<first word>_*_test.py` of its area · failure: none — the selection is a pure function of the changed paths and the file lists.
- REQ-6: When a test file names `cfg(target_os = "<os>")` for the host's operating system, the selector shall select that Rust target · failure: none — the host name comes from the standard library.
- REQ-7: When the selection is made, the runner shall also plan `git diff --check`, `cargo fmt --all -- --check`, `rustfmt --edition 2021 --check` of each changed `src/ops/*.rs`, `bash -n` of each changed shell file and `cargo clippy --all-targets -- -D warnings` · failure: none — these steps are fixed by the changed paths.
- REQ-8: When the steps run, the runner shall run each cargo step through one `heavy-slot-acquire` hold and never stop at the first failure.
- REQ-9 (REQ-8): IF a step's program cannot start, THEN the runner shall record that step with exit status 127 and the result FAIL.
- REQ-10 (REQ-8): IF the checkout is not at the range's head, a tracked file is modified, the base cannot be read or no heavy slot is obtained, THEN the runner shall refuse and run no step.
- REQ-11: When the run ends, the runner shall print one summary line of `key=value` words with the counts and the exit status of every step and `result=PASS` only when every exit status is 0 · failure: none — a step that did not run is counted and makes the result FAIL.
- REQ-12: When `--list` is given, the command shall print the selection and the planned steps and run and write nothing · failure: none — the list is read from git and the files only.

## Acceptance

- AC-1 (REQ-1, REQ-3, REQ-5): 오늘 놓친 사례를 고정물로 둬요: `src/update_models.rs` 를 바꾼 범위는 `update_command_cli`(같은 첫 낱말), `update_tux_tty_test`(영역 첫 낱말의 파이썬), `home_isolation_scan`(저장소 검사 대상)을 모두 고르고 `bundle_*` 처럼 닿지 않는 시험은 고르지 않아요 · verify: `cargo test --locked --offline --test check_affected_cli the_misses_of_the_integration_day_are_selected`
- AC-2 (REQ-1, REQ-2): 영역은 `src/<x>.rs` 면 x, `src/<dir>/<y>.rs` 면 dir, `src/ops/<y>.rs` 면 y 예요; 바뀐 시험 파일은 자기 자신을, `tests/fixtures/<dir>/` 는 그 이름의 시험을, 공용 도우미는 그것을 쓰는 모든 대상을 골라요 · verify: `cargo test --locked --offline --test check_affected_cli areas_changed_tests_fixtures_and_the_helper_select_their_targets`
- AC-3 (REQ-4, REQ-6, REQ-7): 어떤 변경이든 `--lib` 와 `--bins` 가 목록에 있고, 이 호스트의 `cfg(target_os)` 시험 대상이 들어가고, 바뀐 `src/ops/*.rs` 마다 `rustfmt --edition 2021 --check` 가 있어요 · verify: `cargo test --locked --offline --test check_affected_cli the_fixed_steps_and_the_host_targets_are_always_listed`
- AC-4 (REQ-8, REQ-9, REQ-11): 가짜 도구(PATH 의 짧은 스크립트)로 돌리면 한 단계가 실패해도 나머지가 모두 돌고, 단계마다 종료 코드가 기록되고, 요약 한 줄이 정해진 키 순서로 나오고 `result=FAIL`; 모두 0 이면 `result=PASS`; 시작할 수 없는 프로그램은 127 이에요 · verify: `cargo test --locked --offline --test check_affected_cli a_failed_step_does_not_stop_the_rest_and_the_summary_says_fail`
- AC-5 (REQ-10): 체크아웃이 head 가 아니거나 추적 파일이 고쳐졌거나 base 를 읽을 수 없으면 아무 단계도 돌지 않아요(가짜 도구가 호출 흔적을 남기지 않아요) · verify: `cargo test --locked --offline --test check_affected_cli a_checkout_that_is_not_the_head_runs_nothing`
- AC-6 (REQ-12): `--list` 는 저장소 상태(ref, 작업 트리, 파일)를 바꾸지 않고 선택과 단계를 보여 줘요 · verify: `cargo test --locked --offline --test check_affected_cli list_reads_only_and_shows_the_selection`
- AC-7 (REQ-3): `tools/repo_scans.json` 이 없는 프로젝트에서는 `scans=none` 이고 저장소 검사 대상을 고르지 않아요 · verify: `cargo test --locked --offline --test check_affected_cli a_project_without_a_scan_list_says_scans_none`

## Scope

- write: `.moai/specs/SPEC-20261120-check-affected/spec.md`, `src/check_affected.rs`(새), `src/ops/check.rs`(새), `src/lib.rs`(모듈 한 줄), `src/help.rs` 와 `tests/fixtures/help/all*.txt`(표·예시 줄), `docs/design-review/appendix/operations/check.md`(새), `tests/check_affected_cli.rs`(새). 조각 4 에서: `skills/moai/workflows/mode-agent.md`·`mode-leader.md` 의 안내와 `skills/moai/references/source-manifest.json`.
- read-only: `tools/repo_scans.json`, `src/heavy_slot.rs`(`heavy-slot-acquire` 의 옵션), `src/bundle/verify.rs`(한 슬롯 아래에서 명령을 돌리는 방식), 기존 `tests/*.rs`·`tests/*_test.py` 이름
- forbidden: 카드 번호를 SPEC id·가지·커밋·PR 에 쓰는 것, 코드에 통합 가지 이름 박기, 시험을 대신 통과로 쓰는 것, 실제 HOME, 다른 레인의 파일

## Plan

1. 조각 1(이번): SPEC 과 PLAN 노트.
2. 조각 2(고르는 규칙, 읽기만): `src/check_affected.rs` 의 순수 함수(바뀐 경로·파일 목록·스캔 목록 글·호스트 이름 → 선택과 단계 목록), 연산 파일과 문서 조각, 도움말 줄, `--list`. 시험은 오늘 놓친 사례를 고정물로(AC-1·2·3·6·7).
3. 조각 3(돌리는 쪽): 단계를 heavy-slot 아래에서 하나씩 돌리고, 종료 코드를 모으고, 요약 한 줄과 거절을 만들어요. 가짜 cargo·python 으로 시험해요(AC-4·5).
4. 조각 4(마감): `mode-agent`·`mode-leader` 안내(READY 에 이 줄을 적고 리더가 같은 명령을 병합 head 에서 돌려요)와 매니페스트, 변이 시험, 이 저장소에서 실제로 돌려 본 기록.

## Risks

- **첫 낱말 규칙은 많이 골라요.** `src/card/…` 를 바꾸면 `card_*` 시험이 모두 돌아요: 느려도 놓치는 것보다 낫다는 선택이에요(D-2). 필요하면 나중에 `--skip` 같은 좁히기를 사람이 명시해서 쓰게 해요(기본은 넓게).
- **공용 시험 도우미가 바뀌면 거의 모든 대상이 골라져요**(지금 351개 중 299개가 `mod common;` 을 써요). 정확하지만 길어서, 목록 출력에 개수를 보여 줘서 알아보게 해요(D-3).
- `tools/repo_scans.json` 모양에 기대요: `checks` 의 `--test <이름>`, `python3 -m unittest tests.<이름>` 과 `tests` 의 `tests/<이름>.rs` 만 읽어요. 모양이 바뀌면 조용히 덜 고를 수 있어서, 읽지 못한 항목은 이름을 알리고 `tests/` 전체를 고르는 쪽으로 실패해요(REQ-3).
- 파이썬 시험은 빌드된 `moai-cli` 가 있어야 하는 것이 많아요: 파이썬 단계 앞에 `cargo build --bin moai-cli` 단계를 두고 그 종료 코드도 요약에 넣어요(조각 3).
- `cfg(unix)`·`cfg(windows)` 는 고르지 않아요(카드가 `target_os` 만 말해요). 이 호스트가 macOS 면 macOS 이름을 쓴 파일만 골라요: 리눅스 CI 는 그 시험을 돌리지 않으므로 이쪽이 유일한 확인이에요.
- 여러 명령을 한 슬롯에 묶지 않고 cargo 단계마다 슬롯을 잡아요(D-4): 단계 사이에 다른 레인이 끼어들 수 있어요.
- 이 변경 자체가 `src/lib.rs` 한 줄과 `src/help.rs` 줄을 건드려요: 잠깐만 예약하고 바로 반환해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 고르는 일은 순수 함수(경로, 파일 목록, 스캔 목록 글, 호스트 이름 → 선택)이고 git·cargo 를 부르지 않아요. 부르는 쪽이 목록을 만들어 넘겨요 | 배차의 요구와 시험 쉬움: 고정물만으로 오늘 놓친 사례를 못 박을 수 있어요 | 함수에 입출력을 더하면 돼요 |
| D-2 | 영역: `src/<x>.rs` 면 x, `src/<dir>/<y>.rs` 면 dir, 단 `src/ops/<y>.rs` 면 y. 첫 낱말 = 영역의 첫 `_` 앞. 시험 대상 이름이 영역으로 시작하거나 첫 낱말이 같으면 골라요 | 리더의 통합 확인 스크립트 규칙을 그대로 가져오되 `src/ops/` 만 고쳐요(거기서는 영역이 `ops` 하나로 뭉개져 모든 연산이 같은 시험을 골라요) | 규칙 한 곳을 고치면 돼요 |
| D-3 | 공용 시험 도우미(`tests/common*`)가 바뀌면 그것을 `mod common` 으로 쓰는 대상을 모두 골라요 | 고정 목록보다 정확해요; 비용은 출력의 개수로 보여요 | 도우미 이름 목록을 좁혀요 |
| D-4 | 범위의 변경 파일은 `git diff --name-only <base>...<head>`(세 점, merge-base 기준)이고 cargo 단계마다 `heavy-slot-acquire --max 4 --min-free-gb 50` 를 따로 잡아요 | 레인은 기준이 옛 tip 이어도 자기 변경만 골라야 하고, 병합 head 는 base 가 조상이라 같은 답이에요; 한 슬롯 안에 오래 묶지 않아요 | 한 슬롯 스크립트(`bundle verify` 방식)로 바꿔요 |
| D-5 | 체크아웃이 head 가 아니거나 추적 파일이 고쳐졌으면 거절해요 | 결과는 그 head 를 말해야 해요: 고친 파일이 섞이면 READY 의 증거가 아니에요 | 거절을 경고로 낮춰요 |
| D-6 | 기본은 돌리기, `--list` 는 읽기만 | 카드가 "골라서 돌리고 한 줄 요약"을 말해요; 읽기 전용 쓰임은 시험과 사람이 쓰는 미리 보기예요 | 기본을 `--list` 로 뒤집어요 |
| D-7 | 요약 한 줄은 `check-affected base=<12> head=<12> targets=<n> py=<n> lib=<exit> bins=<exit> …  result=PASS|FAIL` 꼴의 `key=value` 낱말이고 키 순서가 고정이에요 | 레인은 READY 노트에, 리더는 병합 head 에서 같은 줄을 읽고 기계가 비교해요 | 키를 더해도 앞 키 순서는 지켜요 |
| D-8 | 통합 가지 이름·카드 번호를 코드에 두지 않아요; base 는 사람이 인자로 줘요 | 이 도구는 모든 MoAI 프로젝트의 것이고, 카드 번호는 밖으로 나가는 글에 안 써요 | — |

## Result

(구현 뒤에 채워요.)
