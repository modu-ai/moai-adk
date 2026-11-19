---
id: SPEC-20261117-heavy-slot-status
title: "heavy-slot-status marks a dead recorded pid and names the processes that hold the slot"
status: draft
revision: 1
created: 2026-11-19
---

# heavy-slot-status marks a dead recorded pid and names the processes that hold the slot

## Intent

무거운 실행 래퍼(`moai heavy-slot-acquire`)는 슬롯 잠금 파일을 COMMAND 에 넘겨주고, COMMAND 이 남기고 간 자식·후손이 그 파일을 계속 열고 있으면 래퍼와 COMMAND 이 끝난 뒤에도 슬롯이 잡혀 있어요. 2026-11-15 이 기기에서의 관찰(읽기만, Issue #3033): `heavy-slot-status` 는 슬롯 1 보유자로 이미 끝난 pid 75578 을 보여 줬고, 실제 잠금은 그 명령이 남긴 후손(bun 실행)이 열린 파일 기술자로 쥐고 있었어요 — 상태 기록의 pid 는 래퍼 것이라 낡고, 누가 실제로 잡고 있는지 알 길이 없었어요. 원인은 설계예요: `flock` 은 열린 파일 기술자 단위라서 fd 를 물려받은 후손이 있으면 잠금이 남고, 후손이 다 끝나야 운영체제가 풀어요.

이 변경은 상태 읽기를 고쳐요: 각 held 항목이 기록 pid 의 생존(`alive`)을 보여 주고, 죽은 pid 슬롯에는 그 슬롯 파일을 연 프로세스(`holders`, `pid`·`comm`, argv 없음)와 근거 출처(`holders_source`: `proc` 또는 `lsof`, 둘 다 못 읽으면 `holders` null 과 `unavailable`)를 보여 줘요. 락의 의미는 바뀌지 않아요 — 자식이 슬롯을 이어받아 남는 것은 실제 부하를 세는 상한의 목적에 맞고, 문서로 못박아요.

이 SPEC 은 PR #3093(head 35ab1c5c2f, merge 76e8643098, 2026-11-16T01:19:02Z 병합)으로 이미 tip 에 있는 같은 일의 SDD 뒤붙이기예요(선례: SPEC-20261119-heavy-slot-pressure). 설계 결정은 카드 노트 1964(계획)와 리더 결정 노트 1965(L1–L5 전부 기본값)에 기록돼 있고, 이 문서가 그 계약을 남기고 팁에서 다시 검증해요.

## Non-goals

- 락 의미 바꾸기 — COMMAND 의 fd 상속 유지(close-on-exec 없음), 프로세스 그룹 대기·정리 없음(리더 결정 L1 (나)).
- 슬롯 회수 명령 만들기(리더 결정 L2) — 잠금을 쥔 프로세스를 끝내는 일은 그 프로세스의 주인 몫이에요.
- 신호를 자식 프로세스 그룹으로 보내기(리더 결정 L3).
- Python 쌍둥이 `tools/heavy_check_slot.py status` 에 `alive` 키 더하기(리더 결정 L4 — "Rust status 만").
- `alive`·`holders` 를 근거로 슬롯을 비우거나 프로세스를 끝내는 어떤 경로(리더 결정의 조건 — 읽기 전용 사실로만 쓰고, 시험으로 고정해요).
- `holders` 에 argv 넣기 — 명령행에 비밀이 섞일 수 있어요.

## Requirements

- REQ-1: While heavy-slot-status lists a held slot, it shall mark the record's pid as `alive` — true when a process of that pid exists, false when it has ended, null when the record names no process.
- REQ-2 (REQ-1): IF the recorded pid is dead, THEN the same held entry shall also list `holders`, the processes that have the slot file open as `pid` and `comm` pairs without their arguments, and `holders_source` naming where they were read (`proc` or `lsof`).
- REQ-3 (REQ-2): IF no source can name the processes that hold the slot file, THEN the entry shall show `holders` null with `holders_source` `unavailable`, which means they could not be looked up, never that nobody holds the slot.
- REQ-4: The wrapper shall keep handing the slot's open file to COMMAND, so a child or descendant that keeps it open holds the slot after the wrapper and COMMAND end, and the operating system frees it only when every one of them is gone — the lock meaning stays exactly as before this change.
- REQ-5: No command path shall empty a slot, edit or delete a slot file, or end a process from `alive` or `holders` alone; ending the process that holds a slot is for that process's owner.
- REQ-6: The Python twin `tools/heavy_check_slot.py status` shall keep none of the `alive`, `holders` or `holders_source` keys; the contract is the Rust status's alone.

## Acceptance

- AC-1 (REQ-1): 상태가 기록 pid 의 생존을 판정해요. 경계: 자기 pid 는 true, 갓 거둔 자식 pid 는 false, pid 1 은 true, 없는 pid·0·음수·숫자 아닌 문자열은 null, 2147483647 은 false, 2147483648 과 `{}` 는 null.
  verify: `cargo test --locked --offline --test heavy_slot_cli status_says_whether_the_recorded_pid_is_alive`
  verify: `cargo test --locked --offline --lib heavy_slot::alive_tests`
- AC-2 (REQ-2, REQ-3): 죽은 pid 슬롯은 실제로 파일을 연 프로세스를 이름으로 보여 줘요 — 래퍼가 남긴 배경 자식 사례에서 holders 가 pid·comm 을 알려 주고, 아무 소스도 답하지 못하면 `unavailable` 과 null(빈 목록이 아니라), 읽기만으로 상태 파일의 바이트와 mtime 이 그대로예요.
  verify: `cargo test --locked --offline --test heavy_slot_cli a_command_that_leaves_a_background_child_keeps_the_slot_and_status_shows_a_dead_pid`
  verify: `cargo test --locked --offline --test heavy_slot_cli status_reads_only`
- AC-3 (REQ-4): 락 보증이 그대로예요 — 래퍼를 죽여도 COMMAND 이 슬롯을 계속 쥐고(Rust·Python 쌍둥이), 배경 자식이 남으면 슬롯이 남는 것이 문서화된 계약이에요.
  verify: `cargo test --locked --offline --test heavy_slot_cli the_slot_stays_held_by_the_command_after_the_wrapper_is_killed`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'heavy_check_slot_test.py'`
- AC-4 (REQ-1…REQ-6): 이웃 검사 — `heavy_slot_cli` 전체, 형식, 잠금 경고, 저장소 단어 스캔과 소스 매니페스트.
  verify: `cargo test --locked --offline --test heavy_slot_cli`
  verify: `cargo fmt --all --check`
  verify: `cargo clippy --offline --locked --all-targets -- -D warnings`
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.repo_scans_test tests.source_manifest_test`

## Scope

- write: `.moai/specs/SPEC-20261117-heavy-slot-status/spec.md`
- read-only: `src/heavy_slot.rs`, `src/home_state.rs`, `tests/heavy_slot_cli.rs`, `tools/heavy_check_slot.py`, `tests/heavy_check_slot_test.py`, `docs/rust-product-path.md`, `skills/moai/workflows/verify.md` — 구현과 문서 계약은 #3093 으로 이미 tip 에 있고, 이 SPEC 은 재검증만 해요
- forbidden: 실제 HOME 과 원장, 네트워크, 실제 슬롯 파일의 편집·삭제, 키

## Plan

이 일은 이미 PR #3093 으로 병합됐어요; 이 문서는 그 계획을 남기고 팁에서 다시 검증해요.

1. 2026-11-15 관찰(이 기기, 읽기만): 슬롯 1 의 기록 pid 75578 은 끝났고, 실제 잠금은 bun 후손이 FD 로 쥐고 있었어요 — Issue #3033.
2. 계획(카드 노트 1964)과 리더 결정(노트 1965: L1–L5 전부 기본값)으로 설계 확정 — 락 의미 유지, 회수 명령 없음, 그룹 신호 없음, Python 키 없음, 키 이름 `alive`.
3. PR #3093 이 구현(`alive`·`holders`·`holders_source`, 문서 4곳, 매니페스트 동기)을 2026-11-16 에 병합했어요.
4. 이 라운드(2026-11-19): 이 SPEC 을 tip `924da079d5` 위에서 쓰고, Result 의 모든 검사를 이 워크트리(`/Users/goos/.moai/worktrees/moai-adk-39`, 가지 `spec/20261119-heavy-slot-status`)에서 다시 재어요 — 코드 차이는 없고 SPEC 파일 하나뿐이에요.

## Risks

- `holders` 의 `proc`·`lsof` 는 최선 노력이에요 — `unavailable` 은 조회 실패일 뿐 무보유가 아니고(REQ-3), 사람이 holders 를 보고 자기 프로세스를 정상 종료하는 길만 남아요.
- macOS 의 `lsof` 는 느리고 없을 수도 있어요 — `unavailable` 로 표시되고 status 는 여전히 읽기만 해요.
- `alive` null 은 기록이 숫자가 아닌 망가진 경우예요 — 그 슬롯은 holders 없이 그대로 보여요(단위 시험이 경계를 고정해요).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 락 의미를 바꾸지 않는다 — fd 상속 유지, 문서만 추가(리더 결정 L1 (나)) | 슬롯 상한 3 은 실제 부하를 세는 목적; 상한이 열리면 #2829 의 디스크·swap 사고가 되풀이될 수 있어요 | (가) close-on-exec + 그룹 정리 — 사용자 결정 몫 |
| D-2 | 회수 명령을 만들지 않는다(리더 결정 L2) | 다른 프로젝트의 프로세스를 구분할 방법이 없고, 사용자 지시가 정상 종료 대기만을 말해요 | 별도 카드로 회수 경로 설계 |
| D-3 | 프로세스 그룹 신호를 보내지 않는다(리더 결정 L3) | 터미널 입력 처리에 위험이 있어요 | 신호 전달 추가 |
| D-4 | Python status 에 `alive` 를 더하지 않는다(리더 결정 L4) | 문서가 "Rust status 만"으로 못박고, 쌍둥이는 최소 대칭을 유지해요 | Python 에 키 추가 |
| D-5 | `holders` 는 `pid`·`comm` 만, argv 없음 | 명령행에 비밀이 섞일 수 있어요 | argv 포함 — secret 유출 위험 |

## Result

검증 트리: tip `924da079d5` 위에 이 SPEC 파일 하나(코드 차이 없음), 워크트리 `/Users/goos/.moai/worktrees/moai-adk-39`, heavy-slot 안에서 측정(2026-11-19, agent-18).

- AC-1: met — `heavy_slot_cli` 21 passed 0 failed(빌드 25.5s, 시험 5.6s, slot 0; `status_says_whether_the_recorded_pid_is_alive` 포함), `alive_tests` 3 passed 0 failed(1450 filtered out; `the_recorded_pid_decides_alive` 포함).
- AC-2: met — 같은 21 passed 실행에서 `a_command_that_leaves_a_background_child_keeps_the_slot_and_status_shows_a_dead_pid` 와 `status_reads_only` 통과, `alive_tests` 의 `unreadable_holders_are_unavailable_never_an_empty_list`·`only_a_dead_pid_asks_for_holders` 통과.
- AC-3: met — 같은 실행에서 `the_slot_stays_held_by_the_command_after_the_wrapper_is_killed` 통과; Python 쌍둥이 `heavy_check_slot_test.py` 25 tests OK(15.5s, Python 3.13.1).
- AC-4: met — `heavy_slot_cli` 전체 21 passed 0 failed; `cargo fmt --all --check` exit 0; `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0(53.7s); `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.repo_scans_test tests.source_manifest_test` 68 tests OK(85.1s) — 이 네 검사는 SPEC 파일이 포함된 최종 트리에서 커밋 직전 다시 돌려 같은 결과를 확인했어요.
