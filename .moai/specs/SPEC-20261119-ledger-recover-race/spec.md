---
id: SPEC-20261119-ledger-recover-race
title: "A recovery that loses the race to a finished one answers unchanged, not target_healthy"
status: draft
revision: 1
created: 2026-11-19
---

# A recovery that loses the race to a finished one answers unchanged, not target_healthy

## Intent

카드 C-109예요. `tests/ledger_recover.rs`의 `two_processes_recovering_at_once_leave_one_winner_and_a_sound_ledger`가 macOS CI에서 가끔 실패해요(2026-11-19 두 번: PR #3219 run 44438864038, PR #3216 run 44452712179). 두 번 모두 같은 단언이 같은 값으로 실패했어요: 진 쪽의 결과가 `target_healthy`예요(run 44452712179 round 6, run 44438864038 round 11: `["applied", "target_healthy"]`). 시험은 진 쪽이 `recovery_in_progress`(락에서 쫓겨남) 또는 `unchanged`(이긴 쪽이 끝난 뒤에 시작함)라고 약속해요.

원인은 `src/ledger/recover.rs`의 `apply_with_suffix`에 있어요. 락 없이 하는 1차 읽기는 대상이 아직 손상 상태라고 보고 지나가요. 그다음 `prepare`와 `check_plan`이 계획을 다시 계산하는데, 그 사이에 이긴 쪽이 설치를 끝내면 계획의 대상이 멀쩡해져서 `check_plan`이 `target_healthy`로 거절해요. 락 아래의 2차 확인(이미 이 계획이 적용됐으면 `unchanged`)은 `check_plan` 뒤에 있어서 거기까지 가지 못해요. 이긴 쪽이 끝난 뒤에 시작한 경우와 다른 점은 한 가지, 1차 읽기가 설치 직전이었다는 것뿐이라, 같은 `unchanged`가 맞아요. 그 사이의 틈이 이 SPEC의 원인이에요.

이 SPEC은 그 틈을 막는 코드와 시험 한 개를 더해요. 시간 제한을 늘리거나 재시도를 넣지 않아요.

## Non-goals

- 시험의 라운드 수·대기 시간을 바꾸는 일.
- `recovery_in_progress`가 나오는 경로(락) 바꾸기.
- `src/ledger/workflow.rs` 등 복구 밖의 파일.

## Requirements

- REQ-1: WHEN another process finishes the same recovery after this process read the ledger as damaged and before it recomputed the plan, THEN the apply shall answer `unchanged`.
- REQ-2 (REQ-1): IF the ledger is usable but holds no recovery event for the plan's digest, THEN the apply shall refuse with `target_healthy` as before.
- REQ-3: The apply shall make the same `unchanged` decision on the first read, after a refused plan, and under the lock.
- REQ-4: The apply shall add no retry and no wait.

## Acceptance

- AC-1 (REQ-1, REQ-3): 이긴 쪽이 1차 읽기와 계획 재계산 사이에 끝나는 순서를 시험이 정확히 만들어서(코드 사이에 끼는 시험용 자리), 진 쪽이 `unchanged`를 받아요.
  verify: `cargo test --locked --offline --test ledger_recover a_recovery_finished_between_the_first_read_and_the_plan_is_unchanged`
- AC-2 (REQ-2): 멀쩡하지만 이 계획의 결과가 아닌 원장은 그대로 `target_healthy`로 거절돼요.
  verify: `cargo test --locked --offline --test ledger_recover a_usable_ledger_that_is_not_this_plans_result_is_still_refused`
- AC-3 (REQ-4): 재시도와 대기가 없어요: 고친 함수에 `sleep`·`loop`·`retry`가 없고, 원래의 경주 시험이 부하 아래에서 통과해요.
  verify: `cargo test --locked --offline --test ledger_recover two_processes_recovering_at_once_leave_one_winner_and_a_sound_ledger`

## Scope

- write: `.moai/specs/SPEC-20261119-ledger-recover-race/spec.md`, `src/ledger/recover.rs`, `tests/ledger_recover.rs`
- forbidden: 실제 HOME 쓰기, 위 목록 밖의 `src/**`(특히 `src/ledger/workflow.rs`는 C-68 예약)

## Plan

1. 틈을 정확히 만드는 시험을 먼저 더해요(RED).
2. 1차 읽기의 판정을 함수 하나로 묶고, 계획이 거절될 때 같은 판정을 한 번 더 해요.
3. 부하 아래에서 원래 시험을 반복해 실패 0을 확인해요(고치기 전 같은 조건의 실패 수도 적어요).

## Risks

- 시험용 자리가 공개 API로 새요 → `#[doc(hidden)]`로 두고 일반 호출 경로는 빈 함수를 넘겨요.
- `unchanged`를 너무 넓게 돌려 진짜 거절을 가려요 → `already_applied`(이 계획의 이벤트가 원장에 있다)가 참일 때만이라 REQ-2 시험이 지켜요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 락을 먼저 잡지 않고, 거절될 때 판정을 한 번 더 한다 | 1차 읽기의 목적은 거절이 폴더를 건드리지 않게 하는 것(락 파일도 안 만듦)이라 순서를 바꾸면 그 약속이 깨져요 | 함수 변경을 되돌리면 돼요 |
| D-2 | 틈을 만드는 시험용 자리를 `apply_with_suffix_hooked`로 둔다 | 이 틈은 프로세스 둘의 우연으로만 열려서, 시험이 순서를 직접 만들어야 해요 | 자리와 시험을 지우면 돼요 |

## Result

- AC-1: met — `cargo test --locked --offline --test ledger_recover a_recovery_finished_between_the_first_read_and_the_plan_is_unchanged`: 통과(진 쪽 `unchanged`, 복구 이벤트 1개, 보관 파일 1개). 수정 전 동작은 변이 M1(둘째 확인을 뺌)로 만들어 시험을 돌렸고, 시험이 `called Result::unwrap() on an Err value: Constraint("target_healthy: the ledger at the real name is usable…")`로 실패했어요: CI에서 본 값과 같은 `target_healthy`예요(RED 증거).
- AC-2: met — `cargo test --locked --offline --test ledger_recover a_usable_ledger_that_is_not_this_plans_result_is_still_refused`: 통과(다른 계획이 만든 멀쩡한 원장은 `target_healthy`로 거절, 이벤트는 그 계획의 것 1개). 변이 M2(둘째 확인이 계획 digest를 무시)와 M3(거절을 항상 `unchanged`로) 둘 다 이 시험이 잡아요.
- AC-3: met — 원래 경주 시험을 부하 아래에서 반복했어요. 방법: 임시 HOME(`mktemp`)에서 시험 실행 파일 하나를 `--exact`로 반복하고, 같은 때 CPU 부하 프로세스 32개(16코어)를 돌려요(스크립트는 이 SPEC의 증거로만 쓰고 저장소에 넣지 않았어요). 수정 전(HEAD `bdb4d3db8c`) 150회 중 4번 실패: `round 1`, `round 2`, `round 10`, `round 2` 모두 `["applied", "target_healthy"]`(CI 로그와 같은 문구). 수정 뒤 300회 중 0번 실패. 수정 전 실패율(4/150)이 그대로면 300회 중 0번이 나올 확률은 약 0.03%예요. 고친 함수에 `sleep`·`loop`·재시도·시간 제한이 없어요(변경은 `already_finished` 함수와 거절 뒤 둘째 확인뿐).

원인(SPEC Intent에 적은 것): CI 두 건(run 44438864038 round 11, run 44452712179 round 6)과 로컬 재현 4건이 같은 값 `target_healthy`예요. 락 없는 1차 읽기(손상) → 이긴 쪽이 설치를 끝냄 → `check_plan`이 계획의 대상을 멀쩡하다고 읽고 거절 → 락 아래의 `unchanged` 확인에 닿지 못함.

변이(`--test ledger_recover`의 세 시험): M1 거절 뒤 둘째 확인 삭제 → KILLED(`a_recovery_finished_between_…`); M2 둘째 확인이 digest를 무시 → KILLED(`a_usable_ledger_that_is_not_…`); M3 거절을 항상 `unchanged`로 → KILLED(같은 시험). 변이마다 `git checkout --` 뒤 작업 트리 clean.

검사: `cargo test --locked --offline --lib` 1451 passed; `--test ledger_recover` 51 passed; `--test ledger_open` 26; `--test spec_cli` 17; `cargo fmt --check` 0; `cargo clippy --locked --offline --all-targets -- -D warnings` 0; `spec check` errors 0 warnings 0.

검사하지 않은 것: macOS CI 자체의 재실행(이 환경은 macOS arm64지만 CI 러너와 같다고는 말하지 않아요; 부하 재현은 CI의 느린 러너를 흉내 낸 것이에요). 통합 시험 전체(`tests/*`)와 파이썬 핀은 돌리지 않았어요(바뀐 곳이 복구 함수 하나뿐). 이 카드에서 `recovery_in_progress` 경로(락)는 바꾸지 않았고 그 경로의 시험은 기존 그대로 통과해요.
