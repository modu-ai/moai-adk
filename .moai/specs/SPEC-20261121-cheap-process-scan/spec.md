---
id: SPEC-20261121-cheap-process-scan
title: "The home, worktree and hygiene commands read the process table cheaply, and tests can inject it"
status: draft
revision: 1
created: 2026-11-21
---

# The home, worktree and hygiene commands read the process table cheaply, and tests can inject it

## Intent

HOME 이동·정리, 워크트리 정리, 위생 정리 명령은 지우거나 옮기기 전에 "어떤 프로세스가 이 경로를 쓰나"를 `lsof` 로 읽어요(`src/worktree_gc.rs` 의 `process_cwds` 와 `open_file_paths`). 기계 전체의 열린 파일을 읽는 `lsof -w -Fn` 은 이 기계에서 한 번에 5.8~7.4 초, 출력 4.9 MB 예요. 같은 명령을 시험이 수백 번 돌려서 시험 실행 시간의 큰 몫이 돼요(보고서의 추정 25~30%, 이 SPEC 의 확인으로 정해요).

느린 이유를 가려 재 봤어요(2026-11-21, 부하 평균 25~60 인 같은 기계): 전체 `lsof -w -Fn` 5.8~7.4 초, `-n -P`(이름 풀이 끔) 2.8 초, `-n -P -b`(차단될 수 있는 커널 호출 끔) 0.66 초이고 파일 이름 줄 수는 같아요. 열어 둔 파일과 cwd 를 가진 프로세스를 띄워 두 방식이 모두 보는 것도 확인했어요. 경로를 인자로 주는 방식(`lsof -- <폴더>`)은 폴더 자체만 맞고 안의 파일은 못 보며, `+D` 는 큰 폴더에서 10 분 넘게 걸려서(`~/.moai`) 쓸 수 없어요. 그래서 판단은 지금과 같게 두고 읽는 방법만 싸게 해요. 시험에는 프로세스 표를 고정해서 주입하는 끄개를 더해요(디버그 빌드만).

## Non-goals

- 사용 중 판단(cwd, 열린 파일, 명령줄)과 "알 수 없으면 거절"을 바꾸지 않아요.
- 삭제·이동 직전에 다시 읽는 일을 줄이지 않아요(캐시로 재확인을 없애지 않아요).
- 릴리스 빌드에는 주입 끄개가 없어요.
- `heavy-slot` 의 `lsof`(슬롯 파일 하나를 묻는 것)는 바꾸지 않아요.

## Requirements

- REQ-1: The working-folder scan shall run `lsof` with `-w -n -P -b`.
- REQ-2: The open-file scan shall run `lsof` with `-w -n -P -b`.
- REQ-3: The process scans shall report a working folder and a file held open that another process has, as before.
- REQ-4: While a debug build runs with `MOAI_CLI_TEST_PROCESS_TABLE` set to `empty`, the process scans shall report no working folders and no open files without running `lsof`.

## Acceptance

- AC-1 (REQ-1, REQ-2): 가짜 `lsof` 를 PATH 앞에 두고 HOME 계획을 돌리면 기록된 호출 두 개가 모두 `-w -n -P -b` 를 가져요.
  verify: `cargo test --locked --offline --test process_scan_cost_cli the_process_scans_run_lsof_without_name_lookups_or_blocking_calls`
- AC-2 (REQ-3): 파일을 열어 두고 폴더에서 일하는 프로세스를 두고 돌리면 열린 파일 검사와 cwd 검사가 둘 다 그것을 이름으로 대요(진짜 `lsof`).
  verify: `cargo test --locked --offline --lib worktree_gc::tests::the_cheap_scans_still_see_a_held_file_and_a_working_folder`
- AC-3 (REQ-4): 끄개를 `empty` 로 주면 가짜 `lsof` 가 한 번도 불리지 않고 HOME 계획이 사용 중이라는 막힘 없이 나와요.
  verify: `cargo test --locked --offline --test process_scan_cost_cli the_injected_empty_table_skips_lsof_and_blocks_nothing`
- AC-4: 이웃 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test home_state_plan --test home_state_selective --test worktree_gc --test worktree_gc_apply --test hygiene_apply_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-cheap-process-scan/spec.md`, `src/worktree_gc.rs`(`process_cwds`, `open_file_paths`, 주입 끄개, 단위 시험), `tests/process_scan_cost_cli.rs`(새 파일)
- read-only: `src/home_state.rs`, `src/hygiene.rs`
- forbidden: 판단 규칙, 삭제 직전 재확인 줄이기, `heavy_slot.rs`

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: `tests/process_scan_cost_cli.rs` 의 AC-1·AC-3 시험이 지금 코드에서 실패하는 것을 봐요.
3. GREEN: 두 스캔에 `-w -n -P -b` 를 한 곳(상수)에서 주고, 디버그 빌드에서 `MOAI_CLI_TEST_PROCESS_TABLE=empty` 이면 빈 표를 돌려줘요. 단위 시험으로 AC-2 를 확인해요.
4. 확인: 새 시험, 이웃 시험, fmt, clippy, 변이 5 개, 고치기 전후 한 시험 바이너리의 실행 시간(끄개 없음/`-n -P -b`/끄개 있음)과 `moai check affected`.

## Risks

- `-b` 는 차단될 수 있는 커널 호출(`stat`, `readlink`)을 건너뛰어요. 이름은 파일 기술자 정보에서 오므로 열린 파일·cwd 이름은 그대로예요(AC-2 가 지켜요). 일부 원격 파일시스템에서는 이름이 덜 풀릴 수 있어요: 그 경우도 "못 본다"가 아니라 이름이 짧아지는 쪽이고, 삭제 대상은 로컬 폴더예요.
- 주입 끄개가 잘못 켜진 채 디버그 빌드가 실제 삭제를 하면 사용 중 검사를 건너뛰어요: 디버그 빌드 한정이고, 값은 정확히 `empty` 일 때만 켜져요(기존 `unavailable` 끄개와 같은 방식).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 기계 전체 스캔은 두고 `-n -P -b` 로 싸게 한다 | 경로별 질의는 안쪽 파일을 못 보고 `+D` 는 큰 폴더에서 매우 느려요. 전체 스캔이 0.66 초면 판단 규칙을 바꿀 이유가 없어요 | 옵션 빼기 |
| D-2 | 주입은 환경 변수 한 개(`MOAI_CLI_TEST_PROCESS_TABLE=empty`)와 디버그 빌드 한정 | 기존 `MOAI_CLI_TEST_PROCESS_CHECK=unavailable` 과 같은 모양이에요 | 변수 빼기 |

## Result

(구현 뒤에 채워요.)
