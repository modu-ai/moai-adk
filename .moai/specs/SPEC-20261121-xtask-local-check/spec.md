---
id: SPEC-20261121-xtask-local-check
title: "A Rust xtask crate for the repository's own tools, first tool: the local check set, with a parity test against the shell script"
status: draft
revision: 1
created: 2026-11-21
---

# A Rust xtask crate for the repository's own tools

## Intent

사용자 결정 2026-11-20(선택 "Rust xtask 로 단계 이전"): 이 저장소의 개발 도구(`tools/` 의 파이썬 22개·셸 6개)를 `cargo xtask <이름>` 으로 부르는 Rust 크레이트로 한 번에 하나씩 옮겨요. 까닭: 언어 하나, 형 검사, 제품 코드(트레일러·SPEC·manifest 파싱) 재사용, 파이썬 환경 함정(pyenv·Xcode python3) 없음, 시험도 Rust.

이 SPEC 은 그 **조각 1** 이에요:

1. 작업공간 구성원 `xtask/`(표준 라이브러리만 쓰는 이진 크레이트)와 `.cargo/config.toml` 의 별칭 `xtask`. 배포 이진과 스킬 꾸러미에는 들어가지 않아요.
2. 첫 도구 `local-check`: `tools/local_check.sh` 와 같은 줄·같은 종료 코드를 내요. 같은 저장소·같은 환경·같은 인자에서 두 도구를 돌려 비교하는 대조 시험이 함께 가요.

옮기는 순서는 도구마다 (1) 같은 인자의 하위 명령 → (2) 바꾸기 전까지 대조 시험 → (3) 부르는 곳을 한 번에 바꾸기 → (4) 파이썬·셸 파일과 그 시험 지우기예요. 이 SPEC 은 (1)(2) 만 해요. (3)(4) 는 다음 조각이에요: 게이트가 증거 줄의 `cmd=tools/local_check.sh` 를 읽고(`src/gate.rs`), CI 가 그 파일을 부르고, 보호 경로 목록이 그 이름을 가리켜서 한 번에 바꿔야 해요.

## Non-goals

- `tools/local_check.sh`, `tests/local_check_test.py`, CI 워크플로, `src/gate.rs` 의 파서, 보호 경로 목록을 바꾸지 않아요(부르는 곳 바꾸기와 지우기는 다음 조각).
- 다른 도구(`ci_scope.py`, `repo_scans.py` 등)를 옮기지 않아요.
- 배포되는 스킬·에이전트 글이 `xtask` 를 말하게 하지 않아요(관리자 도구예요).
- `xtask` 시험을 CI 의 기본 시험 단계에 넣지 않아요: 그 단계는 루트 꾸러미만 돌려요. 다음 조각이 부르는 곳을 바꿀 때 `cargo test -p xtask` 를 CI 에 넣어요.
- 신호(INT·TERM)와 후보 작업 트리 정리 외에 셸의 동작을 고치지 않아요(같은 동작을 그대로 옮겨요).

## Requirements

- REQ-1: The repository shall define the workspace member `xtask`, a binary crate without dependencies, that runs as `cargo xtask <tool>` through the alias `xtask` in `.cargo/config.toml`.
- REQ-2: The workspace shall keep `xtask` out of every shipped build: the root package is the only default member, `xtask` is neither published nor marked for distribution, and `moai-cli` does not depend on it.
- REQ-3: The shipped skills and agents shall not name `xtask`.
- REQ-4: When `cargo xtask local-check` runs, it shall print the same lines and exit with the same code as `tools/local_check.sh` does for the same repository, environment and arguments, apart from the time, run id, script hash, candidate commit and temporary folder in those lines.
- REQ-5 (REQ-4): IF a mode, leg, branch name, fetch, merge, step or diff check fails, THEN `cargo xtask local-check` shall print the same message and exit with the same code as the shell script, with the final line as a fail where the script prints one.
- REQ-6: IF `cargo xtask` is given no tool name or a name it does not know, THEN it shall list its tools on stderr and exit 2.
- REQ-7: When `cargo xtask local-check` receives an interrupt or terminate signal, it shall print its final line as a fail and exit 1 after the running step ends, as the shell script does.
- REQ-8 (REQ-7): IF the signal arrives in candidate mode, THEN `cargo xtask local-check` shall remove the candidate worktree and its temporary folder before it exits.

## Acceptance

- AC-1 (REQ-1): 작업공간에 구성원 `xtask` 가 있고 의존이 없으며, 별칭 `xtask` 가 `run --quiet --package xtask --` 예요.
  verify: `cargo test --locked --offline --test xtask_workspace the_workspace_has_the_xtask_member_and_the_alias`
- AC-2 (REQ-2): `cargo metadata` 가 기본 구성원은 `moai-cli` 하나, `xtask` 는 `publish = false`·`dist = false`, `moai-cli` 의 의존과 이진 목록에 `xtask` 가 없다고 답해요.
  verify: `cargo test --locked --offline --test xtask_workspace the_shipped_build_does_not_contain_xtask`
- AC-3 (REQ-3): `skills/` 와 `agents/` 아래 어느 파일도 `xtask` 를 말하지 않아요.
  verify: `cargo test --locked --offline --test xtask_workspace no_shipped_skill_or_agent_names_xtask`
- AC-4 (REQ-4, REQ-5): 제자리(in-place) 방식의 실행들(마른 실행, 통과, 첫 실패에서 멈춤, 계속 진행, 더러운 트리, diff 검사 실패, 단계 요약 파일)이 셸과 같은 줄·같은 종료 코드·같은 단계 호출을 내요.
  verify: `cargo test --locked --offline -p xtask --test local_check_parity in_place_runs_match_the_script`
- AC-5 (REQ-4, REQ-5): 방식·다리(leg) 값의 모든 경계(빈 값, `nextest`, `rest`, 모르는 값, 후보 방식과 CI 밖 제자리 방식에서의 거절, 모르는 방식)가 셸과 같은 메시지·종료 코드예요.
  verify: `cargo test --locked --offline -p xtask --test local_check_parity modes_and_legs_match_the_script`
- AC-6 (REQ-4, REQ-5): 후보 방식(밖의 origin 에서 받아 합친 트리에서 확인)의 통과·충돌·받기 실패·더러운 트리·마른 실행·분기 이름 경계·`rustc` 낱말 경계가 셸과 같은 줄·종료 코드예요.
  verify: `cargo test --locked --offline -p xtask --test local_check_parity candidate_runs_match_the_script`
- AC-7 (REQ-4): 무거운·파이썬 슬롯(`moai heavy-slot-acquire`)의 쓰임(기본 폴더·한도·이름·거절 종료 코드·슬롯 없음)이 셸과 같은 호출이에요.
  verify: `cargo test --locked --offline -p xtask --test local_check_parity slot_runs_match_the_script`
- AC-8 (REQ-6): 도구 이름이 없거나 모르면 목록을 stderr 에 내고 종료 코드 2 예요.
  verify: `cargo test --locked --offline -p xtask --test local_check_parity an_unknown_tool_lists_the_tools_and_exits_2`
- AC-9 (REQ-7, REQ-8): TERM 을 받으면 실패 줄을 내고 1 로 끝나며 후보 방식에서는 작업 트리와 임시 폴더가 지워져요.
  verify: `cargo test --locked --offline -p xtask --test local_check_parity a_terminate_signal_ends_with_a_fail_line_and_cleans_up`

## Scope

- write: `.moai/specs/SPEC-20261121-xtask-local-check/spec.md`, `Cargo.toml`(작업공간 표와 구성원 한 줄), `Cargo.lock`(`xtask` 항목 하나), `.cargo/config.toml`(별칭), `xtask/**`, `tests/xtask_workspace.rs`
- read-only: `tools/local_check.sh`, `tests/local_check_test.py`, `src/gate.rs`, `.github/workflows/*`, `tools/protected_checks.json`
- forbidden: `tools/` 아래 파일 바꾸기·지우기, 워크플로, `src/gate.rs`, 보호 경로 목록, 배포되는 스킬·에이전트 글

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: `tests/xtask_workspace.rs` 셋과 `xtask/tests/local_check_parity.rs` 를 먼저 써서 실패하는 것을 봐요(대조 시험은 셸과 xtask 를 같은 가짜 저장소에서 차례로 돌려 줄을 비교해요).
3. GREEN: 작업공간 표·별칭·`xtask` 크레이트(`main.rs` 가 도구를 고르고 `local_check.rs` 가 셸의 흐름을 옮겨요), `Cargo.lock` 갱신.
4. 확인: 두 시험 묶음, fmt, clippy `-p xtask --all-targets`, 루트 clippy, 변이(대조가 못 잡는 곳이 있는지), `moai check affected`.

## Risks

- 루트 `Cargo.toml` 이 작업공간을 선언하면 루트에서 도는 모든 `cargo` 명령은 기본 구성원(루트 꾸러미)만 대상이에요. `xtask` 는 clippy·fmt·시험·릴리스 빌드에 들어가지 않아요: 그래서 이 조각의 시험은 `-p xtask` 로 돌리고, 루트의 검사 단계는 그대로예요(Non-goals). 다음 조각이 CI 에 `-p xtask` 단계를 넣기 전까지 `xtask` 의 변경은 CI 에서 검사되지 않아요.
- 증거 줄의 `cmd=tools/local_check.sh` 와 메시지의 `tools/local_check.sh:` 접두어를 그대로 둬요(게이트 파서가 읽어요). 이름 바꾸기는 부르는 곳을 바꾸는 조각이 파서와 함께 해요.
- `script=` 필드는 셸이 자기 파일의 `git hash-object` 를 쓰는 자리예요: `xtask` 는 `xtask/src/local_check.rs` 의 해시를 써요(어느 코드가 돌았는지 말하는 자리라서). 대조 시험은 이 필드를 가리고 40자리 소문자 16진수인지만 봐요.
- 셸의 `trap` 을 신호 처리로 옮기려면 `signal` 시스템 호출이 필요해요: 의존 없이 `extern "C"` 선언 하나로 하고 유닉스에서만 켜요(스크립트가 `bash` 를 부르므로 윈도는 이미 대상이 아니에요).
- 대조 시험은 `cargo`·`python3`·`rustfmt`·`rustc` 를 가짜 실행 파일로 바꿔 돌려서 실제 빌드는 하지 않아요: 실제 도구와의 어긋남은 이 시험이 못 잡아요. 같은 한계를 파이썬 시험이 이미 가져요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `xtask` 는 표준 라이브러리만 쓴다 | 리더 지시: CI 에서 본 빌드보다 먼저 도는 도구가 몇 초에 컴파일돼야 해요. `Cargo.lock` 에 `xtask` 항목 하나만 더해져요 | 의존 더하기 |
| D-2 | 이 조각의 출력은 셸과 글자까지 같다(이름 `tools/local_check.sh` 포함) | 게이트가 `cmd=tools/local_check.sh` 를 읽어요. 대조 시험이 같은 글을 요구해야 바꾸기 전까지 두 도구가 어긋나지 않아요 | 부르는 곳 바꾸는 조각이 이름을 바꿔요 |
| D-3 | `script=` 는 `xtask/src/local_check.rs` 의 `git hash-object` | 셸은 자기 파일을 해시해요. 같은 자리에 돌아간 코드의 해시를 넣어요 | 다른 파일 가리키기 |
| D-4 | 저장소 뿌리는 현재 폴더의 `git rev-parse --show-toplevel` | 셸은 파일 위치로 뿌리를 알지만 `cargo xtask` 는 어느 하위 폴더에서도 불려요 | 환경 변수로 지정 |
| D-5 | 루트 `Cargo.toml` 의 기본 구성원은 루트 꾸러미 하나로 둔다 | 배포 빌드(`cargo build --release --locked`)와 CI 가 `xtask` 이진을 만들지 않아요 | `default-members` 에 더하기 |
| D-6 | 신호는 의존 없이 `extern "C" signal` 로 받는다 | 셸의 `trap` 동작(실패 줄을 내고 1 로 끝남)을 같게 하려면 필요해요 | 신호 처리 빼기(후보 작업 트리가 남아요) |

## Result

구현 끝(2026-11-21, 가지 `spec/20261121-xtask-local-check`, 조각 1). 먼저 시험을 쓰고(`tests/xtask_workspace.rs`, `xtask/tests/local_check_parity.rs`) 뼈대(실행하면 70 으로 끝남)에서 대조 시험 여섯이 모두 실패하는 것을 봤어요. 그다음 `local-check` 를 옮겼고, 변이 열둘(`target/c/mutx.py`, 결과 `target/c/mutx.sum`)을 모두 시험이 잡았어요. 각 변이 뒤 `git checkout -q -- xtask` 로 되돌렸어요. `tools/local_check.sh`·`tests/local_check_test.py`·워크플로·게이트 파서는 바꾸지 않았어요(Non-goals).

- AC-1: met — `cargo test --locked --offline --test xtask_workspace the_workspace_has_the_xtask_member_and_the_alias` 통과(`cargo metadata` 의 작업공간 구성원이 둘, `xtask` 의 의존 0, 이진 대상 `xtask`, 별칭 `run --quiet --package xtask --`). `Cargo.lock` 은 `xtask` 항목 하나만 늘었어요.
- AC-2: met — `cargo test --locked --offline --test xtask_workspace the_shipped_build_does_not_contain_xtask` 통과(기본 구성원은 `moai-cli` 하나, `xtask` 는 `publish` 없음·`dist = false`, `moai-cli` 의 의존에 `xtask` 없음, 배포 이진은 `moai`·`moai-cli`). 이 시험에 대한 변이는 돌리지 않았어요.
- AC-3: met — `cargo test --locked --offline --test xtask_workspace no_shipped_skill_or_agent_names_xtask` 통과(`skills/`·`agents/` 아래 파일 중 `xtask` 를 말하는 것 0). 이 시험에 대한 변이는 돌리지 않았어요.
- AC-4: met — `cargo test --locked --offline -p xtask --test local_check_parity in_place_runs_match_the_script` 통과(마른 실행·통과·첫 실패·계속 진행·더러운 트리 셋·diff 검사 실패·단계 요약 여덟 줄·기준 지정, 열한 장면에서 셸과 stdout·stderr·종료 코드·도구 호출·요약 파일이 같음). 변이 "`fmt` 단계 인자 바꾸기", "계속 진행 무시", "`dirty=1` 표시 빼기"에 실패했어요.
- AC-5: met — `cargo test --locked --offline -p xtask --test local_check_parity modes_and_legs_match_the_script` 통과(모르는 방식, 다리 세 값, 모르는 다리, CI 밖과 후보 방식의 다리 거절, 다른 다리의 실패 무시, 다리의 첫 실패와 계속 진행). 변이 "계속 진행 무시"와 "후보 방식의 다리 거절 빼기"에 실패했어요.
- AC-6: met — `cargo test --locked --offline -p xtask --test local_check_parity candidate_runs_match_the_script` 통과(깨끗한 합침, 충돌, 관계없는 역사, 받기 실패, 더러운 폴더, 마른 실행, 분기 이름 다섯, `rustc` 출력 여섯, 이전 후보 폴더, 단계 실패, diff 검사, 옛 덮어쓰기 값 무시, 기준이 더한 규칙). 변이 "분기 이름의 `..` 검사 빼기", "`rustc` 낱말 길이 33", "임시 폴더 지우기 건너뛰기", "diff 범위 뒤집기", "충돌 설명의 첫 줄 건너뛰기", "merge-tree 실패 검사 빼기"에 실패했어요.
- AC-7: met — `cargo test --locked --offline -p xtask --test local_check_parity slot_runs_match_the_script` 통과(기본 슬롯 호출 넷, `XDG_CACHE_HOME`, 폴더·한도·이름 지정, 빈 설정, HOME 없음에서 슬롯 없이 실행, 거절 종료 코드 75). 변이 "파이썬 슬롯 이름 접미어 바꾸기"에 실패했어요.
- AC-8: met — `cargo test --locked --offline -p xtask --test local_check_parity an_unknown_tool_lists_the_tools_and_exits_2` 통과(이름 없음과 모르는 이름이 종료 코드 2 와 stderr 의 목록). 이 시험에 대한 변이는 돌리지 않았어요.
- AC-9: met — `cargo test --locked --offline -p xtask --test local_check_parity a_terminate_signal_ends_with_a_fail_line_and_cleans_up` 통과(제자리와 후보 방식 모두 TERM 뒤 실패 줄과 종료 코드 1, 이어지는 단계 없음, 후보 작업 트리와 임시 폴더가 지워지고 셸과 같은 출력). 변이 "임시 폴더 지우기 건너뛰기"와 "단계 뒤 신호 검사 빼기"에 실패했어요.
- 돌리지 않은 것: 실제 `cargo`·`python3` 로 도는 전체 검사(대조 시험은 가짜 실행 파일을 써요), 윈도(스크립트가 `bash` 를 써서 이미 대상이 아니에요), CI 에서 `-p xtask` 시험(다음 조각이 부르는 곳을 바꿀 때 넣어요). 전체 `moai check affected` 는 병합 준비 노트에 줄째로 적어요.
