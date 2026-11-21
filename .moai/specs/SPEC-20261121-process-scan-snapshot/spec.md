---
id: SPEC-20261121-process-scan-snapshot
title: "One process snapshot serves both use checks, tests inject it from a table, and a stuck lsof is given up on"
status: draft
revision: 1
created: 2026-11-21
---

# One process snapshot serves both use checks, tests inject it from a table, and a stuck lsof is given up on

## Intent

SPEC-20261121-cheap-process-scan 이 `lsof` 를 싸게 했지만(기계 전체 6 초대 → 0.66 초) 한 명령이 아직 `lsof` 를 두 번 불러요: 작업 폴더(`-a -d cwd`)와 열린 파일 스캔이 따로 돌아요. HOME 명령 하나가 이 기계에서 2.9~4.3 초(부하 31~38)이고, 시험 바이너리 `home_install_apply_cli` 는 81.84 초(목표 ≤ 60 초)예요. 남은 비용을 줄이는 방법은 스캔 횟수예요.

이 SPEC 은 (1) 한 번의 `lsof -F fn` 으로 작업 폴더와 열린 파일을 함께 읽고, (2) 1 초 안의 다음 질문은 직전 스냅샷으로 답하며(두 검사가 연달아 오는 한 명령 안), 그보다 오래되면 다시 읽고(삭제·이동 직전 재확인은 스캔 자체가 걸리는 시간 정도의 틈만 남아요), (3) 시험이 고정된 프로세스 표를 파일로 주입할 수 있게 하고, (4) 응답하지 않는 `lsof` 를 시간 한도 뒤 포기해 "알 수 없음"(닫힌 쪽)으로 처리해요. 경로 판단은 지금도 스냅샷의 경로 접두어 비교예요: 같은 규칙을 한 곳으로 모아 12 가지 상황에서 옛 두 스캔과 같은 답을 내는지 시험해요.

## Non-goals

- 사용 중 판단 규칙과 "알 수 없으면 거절"을 바꾸지 않아요.
- 삭제·이동 직전의 재확인을 없애지 않아요(1 초가 넘으면 새로 읽어요).
- 릴리스 빌드에는 주입 끄개와 시간 한도 조절이 없어요(시간 한도 자체는 릴리스에도 있어요).
- 명령줄 읽기(`ps`)는 바꾸지 않아요(0.15 초).

## Requirements

- REQ-1: The process scan shall run `lsof` once for the working folders and the open files together.
- REQ-2: When a scan is asked for within one second of the last one, the scan shall answer from the last snapshot without running `lsof`.
- REQ-3 (REQ-2): IF the last snapshot is older than one second, THEN the scan shall run `lsof` again.
- REQ-4: The scan shall report a path as used when a working folder or an open file lies inside it by path prefix.
- REQ-5: The scan shall not report a sibling path whose name only starts the same as a used path.
- REQ-6: While a debug build runs with `MOAI_CLI_TEST_PROCESS_TABLE` naming a file, the scan shall read the lines `cwd <path>` and `open <path>` of that file as the table without running `lsof`.
- REQ-7: The scan shall stop an `lsof` that has not answered within its time limit.
- REQ-8: The scan shall report the table as unavailable when its `lsof` was stopped.
- REQ-9: The scan shall report the same working folders and open files as the two earlier scans for the same `lsof` output.

## Acceptance

- AC-1 (REQ-1): 가짜 `lsof` 를 PATH 앞에 두고 HOME 계획을 돌리면 호출이 한 번이고 `-w -n -P -b` 와 `-F fn` 을 가져요.
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli one_lsof_call_serves_both_checks`
- AC-2 (REQ-2, REQ-3): 스냅샷 시험이 1 초 안의 두 번째 질문은 스캔을 더하지 않고, 한도를 0 으로 준 질문은 스캔을 더하는 것을 세어 보여요.
  verify: `cargo test --locked --offline --lib worktree_gc::tests::a_second_question_within_the_age_reuses_the_snapshot`
- AC-3 (REQ-4, REQ-5, REQ-9): 12 가지 상황(폴더 안 cwd, 안의 열린 파일, 둘 다, 둘 다 없음, 공백 있는 경로, 심볼릭 링크 경로, 이름만 같은 형제, 부모가 cwd, cwd 가 `/`, 소켓·`->` 줄, 한 파일을 여러 프로세스가 염, 빈 출력·잘린 출력)에서 `-F fn` 한 번의 출력이 옛 두 출력(`-a -d cwd -Fn`, `-Fn`)과 같은 작업 폴더·열린 파일·판단을 내요.
  verify: `cargo test --locked --offline --lib worktree_gc::tests::one_listing_gives_the_answers_of_the_two_earlier_scans_in_twelve_situations`
- AC-4 (REQ-6): 주입 표 파일이 `cwd` 줄만 가지면 HOME 계획이 `in_use` 로 막히고, `open` 줄만이면 `file_open_in_use` 만 막히고, 밖의 경로만이면 막힘이 없고, 이 셋 모두 가짜 `lsof` 가 불리지 않아요.
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli an_injected_table_file_drives_the_decisions_without_lsof`
- AC-5 (REQ-7, REQ-8): 응답하지 않는 `lsof` 는 한도 뒤 끝내고 HOME 계획이 `process_check_unavailable` 로 막혀요(막히는 데 걸린 시간에 상한).
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli a_stuck_lsof_is_stopped_and_the_check_reports_unavailable`
- AC-6: 이웃 시험과 이전 조각의 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test process_scan_cost_cli --test home_state_plan --test home_state_selective --test worktree_gc --test worktree_gc_apply --test hygiene_apply_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-process-scan-snapshot/spec.md`, `src/worktree_gc.rs`(스냅샷, `process_cwds`·`open_file_paths` 가 스냅샷에서 답함, 주입 표, 시간 한도, 단위 시험), `tests/process_scan_snapshot_cli.rs`(새 파일)
- read-only: `src/home_state.rs`, `src/hygiene.rs`(다른 카드가 쥐고 있어요), `src/optimize.rs`, `src/worktree_move.rs`
- forbidden: 판단 규칙, 삭제·이동 직전 재확인 없애기, `heavy_slot.rs`

## Plan

1. SPEC 을 커밋해요. `spec check` 를 돌려요.
2. RED: AC-1·AC-4·AC-5 시험이 지금 코드에서 실패하는 것을 봐요(AC-2·AC-3 는 단위 시험이라 구현과 함께 써요).
3. GREEN: `process_cwds()`·`open_file_paths()` 는 같은 이름으로 두고 둘 다 한 스냅샷에서 답해요. 스냅샷은 `lsof -w -n -P -b -F fn` 한 번(리눅스는 `/proc` 한 번)으로 읽고 1 초 동안 재사용해요. 시간 한도는 60 초(디버그 빌드는 `MOAI_CLI_TEST_LSOF_TIMEOUT_MS` 로 줄여요).
4. 확인: 새 시험, 이웃 시험, fmt, clippy, 변이 5 개, `home_install_apply_cli` 와 HOME 명령 하나의 전후 시간.

## Risks

- 1 초 재사용은 그 안에 새로 일을 시작한 프로세스를 못 봐요. 스캔 한 번이 이미 0.66~6 초 걸리고 판단 시점과 삭제 시점 사이에도 같은 틈이 있어서 노출은 같은 정도예요. 삭제·이동 직전 재확인은 1 초가 지났으면 새로 읽어요.
- `-F fn` 은 `-Fn` 보다 출력이 조금 길어요(`f` 줄). 파서가 `f` 줄이 없는 옛 형식도 읽어서 두 출력을 시험으로 대조해요.
- 시간 한도는 응답이 아주 늦은 기계에서 "알 수 없음"으로 거절할 수 있어요: 닫힌 쪽이라 안전하고 한도(60 초)는 측정값(6 초대)의 열 배예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 스냅샷 재사용은 1 초 시간 한도(명령 전체 캐시가 아님) | 삭제·이동 직전 재확인은 신선해야 해요. 연달아 오는 두 검사만 합치면 그 명령의 `lsof` 호출이 절반이 돼요 | 한도를 0 으로 |
| D-2 | 주입은 파일(`cwd`·`open` 줄)이고 `empty` 도 그대로 쓴다 | 고정 상황 12 가지를 시험에서 같은 모양으로 만들 수 있어요 | 변수 빼기 |
| D-3 | 시간 한도 60 초, 넘으면 `lsof` 를 끝내고 알 수 없음 | 멈춘 `lsof` 가 삭제 명령을 영원히 붙잡지 않아요. 알 수 없음은 거절이에요 | 한도 늘리기 |

## Result

(구현 뒤에 채워요.)
