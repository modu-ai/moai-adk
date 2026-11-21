id: SPEC-20261121-process-scan-snapshot
title: "cwd questions read the cwd list, open files once per command, and the deletion re-check reads fresh"
status: draft
revision: 2
created: 2026-11-21
---

# cwd questions read the cwd list, open files once per command, and the deletion re-check reads fresh

## Intent

조각 1(SPEC-20261121-cheap-process-scan)이 기계 전체 `lsof` 를 싸게 했지만, 조각 2 의 "한 스냅샷 + 1 초 창"은 독립 감사(B1~B3, 리더 결정 2026-11-21T08:3xZ)에서 뒤집혔어요: 작업 폴더만 묻는 호출도 열린 파일 목록(약 3 만 경로)을 읽고 canonicalize 했고(`worktree-move-plan` 2.6~9.2 초 → cwd 전용 1 초 안), 1 초 창이 만료되면 같은 명령이 `lsof` 를 또 돌렸고, 재확인이 새로 읽는다는 시험이 없었어요.

이 SPEC 의 rev 2 는 그 결정대로 바꿔요: (1) 작업 폴더 질문은 cwd 목록(`-a -d cwd`)만 읽고 열린 파일을 풀지 않고, (2) 열린 파일 데이터는 caller 가 열린 파일을 물을 때만 명령당 한 번 읽어 명령 전체에서 재사용하고(1 초 창이 아니에요), (3) 삭제·이동 직전 재확인은 캐시와 무관하게 늘 새로 읽고, (4) 주입 표와 시간 한도(조각 2 에서 만든 것)는 그대로 쓰며, (5) 시간 한도로 끊긴 `lsof` 는 알 수 없음(거절)이라는 것을 시험이 고정해요. 경로 판단은 조각 1 과 같은 접두어 비교예요: 같은 규칙을 12 가지 상황에서 옛 두 스캔과 같은 답을 내는지 시험해요.

## Non-goals

- 사용 중 판단 규칙과 "알 수 없으면 거절"을 바꾸지 않아요.
- 삭제·이동 직전의 재확인을 없애지 않아요(오히려 늘 새로 읽게 강화해요).
- 릴리스 빌드에는 주입 끄개와 시간 한도 조절이 없어요(시간 한도 자체는 릴리스에도 있어요).
- 명령줄 읽기(`ps`)는 바꾸지 않아요(0.15 초).
- `src/home_state.rs`, `src/hygiene.rs`, `src/optimize.rs`, `src/worktree_move.rs` 는 건드리지 않아요(호출은 함수 이름 그대로 새 답을 받아요).

## Requirements

- REQ-1: The working-folder scan shall ask `lsof` only for the cwd descriptors of the machine.
- REQ-2: The working-folder scan shall answer without reading the machine-wide open files.
- REQ-3: The open-file scan shall run at most once per command.
- REQ-4: The re-check right before a deletion or a move shall read the process table again.
- REQ-5: The scan shall report a path as used when a working folder or an open file lies inside it by path prefix.
- REQ-6: The scan shall not report a sibling path whose name only starts the same as a used path.
- REQ-7: While a debug build runs with `MOAI_CLI_TEST_PROCESS_TABLE` naming a file, the scan shall read the lines `cwd <path>` and `open <path>` of that file as the table without running `lsof`.
- REQ-8: The scan shall stop an `lsof` that has not answered within its time limit.
- REQ-9: The scan shall report the table as unavailable when its `lsof` was stopped.
- REQ-10: The scan shall report the same working folders and open files as the two earlier scans for the same `lsof` output.

## Acceptance

- AC-1 (REQ-1, REQ-3): 가짜 `lsof` 를 PATH 앞에 두고 HOME 계획을 돌리면 호출이 둘이고(명령이 작업 폴더와 열린 파일을 둘 다 물어요), 첫 호출이 `-a -d cwd` 와 `-w -n -P -b` 를, 둘째 호출이 `-F fn` 을 가져요.
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli the_cwd_check_asks_only_cwd_descriptors_and_the_open_check_the_machine_listing`
- AC-2 (REQ-3): 같은 명령 안의 두 번째 열린 파일 질문은 스캔을 더하지 않고 답해요(가짜 `lsof` 의 호출 기록으로 세어요).
  verify: `cargo test --locked --offline --lib worktree_gc::tests::a_second_question_answers_from_the_command_cache`
- AC-3 (REQ-4): 삭제 후보가 둘인 `worktree-gc-apply` 에서 두 번째 스캔부터 프로세스가 `b` 안에서 일하면 `b` 는 거절돼요(재확인이 자기 스캔을 돌려요) — cached snapshot 으로는 지워져요.
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli the_deletion_recheck_runs_its_own_scan`
- AC-4 (REQ-8, REQ-9): `lsof` 가 시간 한도 안에 답하지 않으면 삭제 재확인이 `process_check_unavailable` 로 거절하고 워크트리는 제자리에 남아요.
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli an_lsof_cut_off_by_the_time_limit_refuses_the_removal`
- AC-5 (REQ-5, REQ-6, REQ-10): 12 가지 상황(폴더 안 cwd, 안의 열린 파일, 둘 다, 둘 다 없음, 공백 있는 경로, 심볼릭 링크 경로, 이름만 같은 형제, 부모가 cwd, cwd 가 `/`, 소켓·`->` 줄, 한 파일을 여러 프로세스가 염, 빈 출력·잘린 출력)에서 새 파서가 옛 두 파서와 같은 작업 폴더·열린 파일·판단을 내요.
  verify: `cargo test --locked --offline --lib worktree_gc::tests::one_listing_gives_the_answers_of_the_two_earlier_scans_in_twelve_situations`
- AC-6 (REQ-7): 주입 표 파일이 `cwd` 줄만 가지면 HOME 계획이 `in_use` 로 막히고, `open` 줄만이면 `file_open_in_use` 만 막히고, 밖의 경로만이면 막힘이 없고, 이 셋 모두 가짜 `lsof` 가 불리지 않아요.
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli an_injected_table_file_drives_the_decisions_without_lsof`
- AC-7 (REQ-8, REQ-9): 응답하지 않는 `lsof` 는 한도 뒤 끝나고 HOME 계획이 `process_check_unavailable` 로 막혀요(막히는 데 걸린 시간에 상한).
  verify: `cargo test --locked --offline --test process_scan_snapshot_cli a_stuck_lsof_is_stopped_and_the_check_reports_unavailable`
- AC-8: 이웃 시험과 이전 조각의 시험이 그대로 통과해요.
  verify: `cargo test --locked --offline --test process_scan_cost_cli --test home_state_plan --test home_state_selective --test worktree_gc --test worktree_gc_apply --test hygiene_apply_cli`
- AC-9: 전후 측정이 기록돼요(조각 2 head 대 수리 head; `home_install_apply_cli`, HOME 명령 하나).

## Scope

- write: `.moai/specs/SPEC-20261121-process-scan-snapshot/spec.md`, `src/worktree_gc.rs`(cwd·open 스캔 분리, 명령 캐시, 재확인 새로 읽기, 주입 표, 시간 한도, 단위 시험), `tests/process_scan_snapshot_cli.rs`(REQ-1·REQ-3·REQ-4·REQ-8 시험), `tests/process_scan_cost_cli.rs`(호출 기록 어설션)
- read-only: `src/home_state.rs`, `src/hygiene.rs`, `src/optimize.rs`, `src/worktree_move.rs`, `src/ops/hygiene_plan.rs`, `src/ops/hygiene_apply.rs`
- forbidden: 판단 규칙, 삭제·이동 직전 재확인 없애기, `heavy_slot.rs`

## Plan

1. RED: AC-1·AC-3·AC-4 시험이 조각 2 코드에서 실패하는 것을 봐요(호출 한 번, 재확인이 캐시를 신뢰, 한도 거절 미고정).
2. GREEN: `read_cwds`(cwd 전용 스캔: 맥 `lsof -w -n -P -b -a -d cwd -F fn`, 리눅스 `/proc` cwd pass)와 `read_open`(맥 `lsof -w -n -P -b -F fn`, 리눅스 `/proc` fd pass)으로 분리해요. 각 명령 캐시는 프로세스 수명 동안 한 번 읽어요(실패한 읽기는 남기지 않아요). `busy_now` 는 캐시를 우회해 새로 읽어요.
3. 확인: 새 시험, 이전 조각 시험의 갱신(REQ-1 기록), 이웃 시험, fmt, clippy, 변이, 전후 측정.

## Risks

- HOME 계획은 작업 폴더와 열린 파일을 둘 다 물어서(호출 순서 cwd 먼저) 맥에서 `lsof` 를 두 번 돌려요. cwd 스캔은 cwd 기술자만 읽어 전체 목록보다 훨씬 가벼워요 — AC-9 측정에서 합을 확인해요.
- 열린 파일 canonicalize(약 3 만 경로)는 열린 파일을 묻는 명령에만 남아요(안전한 접두어 비교에 필요해요). 작업 폴더 질문에서는 사라져요.
- 재확인이 항목마다 새로 읽으니 항목이 많은 apply 는 스캔을 항목 수만큼 돌려요. 안전이 성능보다 앞서요(리더 결정) — 항목은 보통 소수예요.
- 시간 한도는 응답이 아주 늦은 기계에서 "알 수 없음"으로 거절할 수 있어요: 닫힌 쪽이라 안전하고 한도(60 초)는 측정값(6 초대)의 열 배예요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | ~~스냅샷 재사용은 1 초 시간 한도~~ 폐기(리더 결정 2026-11-21T08:3xZ): 명령 전체 캐시 + 삭제 직전 항상 재읽기 | 1 초 창은 재확인을 stale 로 두고 만료 시 스캔을 더 돌려요 | — |
| D-2 | 주입은 파일(`cwd`·`open` 줄)이고 `empty` 도 그대로 쓴다 | 고정 상황 12 가지를 시험에서 같은 모양으로 만들 수 있어요 | 변수 빼기 |
| D-3 | 시간 한도 60 초, 넘으면 `lsof` 를 끝내고 알 수 없음 | 멈춘 `lsof` 가 삭제 명령을 영원히 붙잡지 않아요. 알 수 없음은 거절이에요 | 한도 늘리기 |
| D-4 | cwd 질문은 cwd 목록만 읽고 열린 파일을 canonicalize 하지 않는다; 열린 파일 데이터는 물을 때만 명령당 한 번 | 리더 결정(독립 감사 B1) — cwd 질문이 열린 파일 비용을 치르지 않아요 | cwd 스캔을 전체 목록으로 |

## Result

- 조각 2(rev 1, head 92841d239f87ac8c2af3f48e2c67415cb5ccf92b)는 독립 감사에서 **FAIL**(리더 기록 2026-11-21T08:3xZ): B1 — cwd 질문이 기계 전체 `lsof -F fn` + 열린 파일 canonicalize 로 돌아가 성능 역행, B2 — 삭제 직전 재확인이 새로 읽는다는 시험 없음, B3 — 시간 한도로 끊긴 `lsof` 가 거절로 처리된다는 시험 없음. rev 2 는 그 수리예요.
