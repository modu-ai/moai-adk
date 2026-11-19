---
id: SPEC-20261119-heavy-slot-pressure
title: "heavy-slot judges macOS memory pressure, not swap use"
status: draft
revision: 1
created: 2026-11-19
---

# heavy-slot judges macOS memory pressure, not swap use

## Intent

빌드 슬롯 래퍼(`moai heavy-slot-acquire`, 저장소 도구 `tools/heavy_check_slot.py`)가 macOS 에서 스왑 사용률로 메모리를 판단해서, 압력은 이미 풀렸는데도 모든 무거운 검사를 거부했어요. 2026-11-17 측정: 스왑 49.8/50.2 GiB(99 %) 사용인데 커널 압력 수준은 2(warn), `memory_pressure` 여유는 37 % — macOS 는 압력이 내려가도 스왑 페이지를 오래 남겨서, 80 % 스왑 상한이 1.5 시간 동안 매 빌드를 거부(exit 75)했어요. 이 변경은 macOS 를 커널 메모리 압력 수준(`kern.memorystatus_vm_pressure_level`)과 시스템 전체 메모리 여유 백분율(`kern.memorystatus_level`)로 판단하게 하고, 스왑 사용률은 거부 문구에 참고로만 보여요. Linux 는 스왑 규칙을 그대로 둬요.

이 SPEC 은 PR #3151(커밋 c4eda5a8c3, a46481ff2e)로 병합된 같은 일의 SDD 뒤붙이기예요: 병합은 SDD 관문 전 전환기 예외로 이뤄져 SPEC 과 `Moai-Spec` 트레일러가 없었고, 이 문서가 그 설계·계약을 남기고 팁에서 다시 검증해요.

## Non-goals

- Linux 의 스왑 규칙, 디스크 여유 바닥(`min_free_gb`), 슬롯 수·대기열 동작 바꾸기.
- 두 sysctl 외의 새 판독(페이지 상태 세부 내역 등)과 `memory_pressure` 하위 프로세스 실행(Rust 는 `sysctlbyname` 직접 호출).
- 스왑 비우기·시스템 튜닝, 그리고 슬롯 파일·폴더 관리 규칙.

## Requirements

- REQ-1: Where the machine reports memory pressure readings (macOS), `heavy-slot-acquire` and `heavy_check_slot.py` shall judge whether a heavy run may start on the kernel memory pressure level and the system-wide memory free percentage, treating swap use as information only.
- REQ-2 (REQ-1): IF the pressure level (`kern.memorystatus_vm_pressure_level`) is 4 (critical) or higher, THEN the wrapper shall refuse the run with exit 75 and the measured readings in the message.
- REQ-3 (REQ-1): IF the memory free percentage (`kern.memorystatus_level`) is below `min_mem_free_pct`, THEN the wrapper shall refuse the run with exit 75 and the measured readings in the message.
- REQ-4: While a machine with memory readings is over a limit, the wrapper shall show the swap use in the refusal marked as not a limit on macOS.
- REQ-5: Where the machine reports no memory readings (Linux), the wrapper shall judge a run on swap use against `max_swap_pct` exactly as before this change.
- REQ-6: Where the machine reports no memory readings (Linux), the wrapper shall leave `min_mem_free_pct` unused.
- REQ-7: While the limits file sets `min_mem_free_pct`, the wrapper shall use that value in place of the default 10, looser or stricter.
- REQ-8: While a caller passes `--min-mem-free-pct`, the wrapper shall use it only when it is stricter than the limits file's value or the default 10.
- REQ-9: While the system does not give one of the two readings, the wrapper shall not refuse a run for that missing value alone.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): macOS 판단이 압력·여유 % 로 바뀌어요. 스텁 경계: (스왑 99 %, 압력 2, 여유 37 %)·(스왑 99 %, 압력 3, 여유 37 %)·(압력 2, 여유 10)은 시작(exit 0), (압력 4)·(여유 9)는 거부(exit 75) — 스왑 99 % 는 거부 이유가 아니고, 거부 문구에 "shown only; not a limit on macOS" 로 표시돼요. 두 쌍둥이(Rust·Python)가 같은 입력을 같게 판단해요.
  verify: `cargo test --locked --offline --test heavy_slot_cli macos_is_judged_on_memory_pressure_not_swap`
- AC-2 (REQ-7, REQ-8): `min_mem_free_pct` 한계는 다른 한도와 같은 규칙으로 움직여요. `limits.json` 은 기본 10 을 양쪽으로 대체(20 으로 올리면 여유 19 거부·20 통과, 0 으로 내리면 여유 0 도 통과); 플래그는 엄격할 때만 적용(9 는 기본 10 에 눌리고 각주 출력, 15 는 여유 14 거부·15 통과; `limits.json` 20 앞의 플래그 5 도 눌림).
  verify: `cargo test --locked --offline --lib heavy_slot::limits_tests::the_memory_free_flag_only_tightens`
- AC-3 (REQ-5, REQ-6, REQ-9): Linux 동작은 그대로예요. 판독이 없는 기기(스텁에 `pressure`/`mem_free_pct` 없음)는 스왑 80 통과·81 거부의 옛 규칙을 지키고, `--min-mem-free-pct` 를 내주어도 메모리 한계로 거부하지 않아요. 두 판독 중 하나를 읽지 못하면 그 값 때문에 거부하지 않아요.
  verify: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -p 'heavy_check_slot_test.py'`
- AC-4 (REQ-1…REQ-9): 이웃 검사 — 단위 시험 전체(`--lib heavy_slot::`), 파이썬 파일 전체(패리티 시험 포함), `cargo fmt --check`, `cargo clippy --offline --locked --all-targets -- -D warnings`.
  verify: `cargo test --locked --offline --lib heavy_slot::`
  verify: `cargo fmt --check`
  verify: `cargo clippy --offline --locked --all-targets -- -D warnings`

## Scope

- write: `src/heavy_slot.rs`, `src/main.rs`, `tools/heavy_check_slot.py`, `tests/heavy_check_slot_test.py`, `tests/heavy_slot_cli.rs`, `docs/rust-product-path.md`, `skills/moai/workflows/verify.md`, `skills/moai/references/source-manifest.json`, `.moai/specs/SPEC-20261119-heavy-slot-pressure/spec.md`
- read-only: `src/lane_env.rs` (빌드 폴더 규칙 참조), `.moai/specs/SPEC-20261118-spec-close/spec.md` (SPEC 형식 참조)
- forbidden: 실제 HOME 과 원장, 네트워크, 실제 원격 저장소, 키

## Plan

이 일은 이미 PR #3151 로 병합됐어요; 이 문서는 그 계획을 남기고 팁에서 다시 검증해요.

1. 시험 먼저(RED): 스텁 판독으로 경계(압력 3/4, 여유 10/9, Linux 스왑 80/81)와 Rust·Python 패리티를 검사하는 시험을 먼저 실패시켰어요.
2. Rust `src/heavy_slot.rs`: `Memory` 판독(`sysctlbyname` 두 번, 하위 프로세스 없음), `Limits`·`DEFAULT_LIMITS`·`limits.json` 에 `min_mem_free_pct`(기본 10), `over_limits` 의 macOS 분기(압력 수준 4 이상·여유 % 미만 거부, 스왑은 표시), `--min-mem-free-pct` 플래그는 `src/main.rs` 에.
3. Python `tools/heavy_check_slot.py`: 같은 판단을 `/usr/sbin/sysctl` 절대 경로·5 초 제한·고정 환경으로; 문서(`docs/rust-product-path.md`, `skills/moai/workflows/verify.md`)와 소스 매니페스트 동기.
4. 감사 반영: a46481ff2e(압력 수준 3 경계 시험 B1, 판독 설명 N2).
5. SDD 뒤붙이기 라운드(2026-11-19): 이 SPEC 을 쓰고, REQ-9(판독 하나만 읽히는 경계)의 스텁 사례를 세 시험 표(`tests/heavy_slot_cli.rs` 의 macOS 판단 표와 `tests/heavy_check_slot_test.py` 의 macOS 판단 표·패리티 표)에 추가해요 — 제품 코드는 그대로예요.

## Risks

- `kern.memorystatus_level` 은 문서화되지 않은 값이에요 — Apple 이 바꾸면 판독이 `None` 이 되고, 그 값 때문의 거부는 사라져요(디스크 바닥은 남아요). 실패 방향이 열려 있는 대신, 잘못 막히던 원래 문제(모든 빌드 거부)보다 위험이 작아요.
- Python 은 sysctl 을 하위 프로세스로 돌려요(5 초 제한·고정 환경) — 답이 없으면 `None`, 그 값으로 거부하지 않아요.
- 거부 문구의 "swap use … shown only" 줄을 Linux 사용자가 한계로 오읽할 수 있어요 — 문구가 "not a limit on macOS" 라고 못 박아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 압력 수준 4(critical)부터 거부, 2(warn)는 통과 | 2026-11-17 측정: warn 에서 1.5 시간 모든 빌드 거부가 바로 이 결함 | 압력 2부터 거부 |
| D-2 | 여유 % 한계는 `kern.memorystatus_level`, 기본 10 | `memory_pressure` 가 출력하는 같은 값(측정일 37 %) | 다른 기본값 |
| D-3 | macOS 에서 스왑 % 는 표시 전용 | 압력 해소 후에도 스왑이 남아 거짓 양성의 원인(측정) | 스왑을 한계로 유지 |
| D-4 | `--min-mem-free-pct`·`limits.json` 갱신 규칙은 다른 한도와 같음(엄격한 쪽만) | #2930 R-2 의 한 규칙 | 이 키만 느슨해지게 허용 |

## Result

- AC-1: met — 팁 192f6fa7e2(2026-11-19, 이 worktree)에서 `cargo test --locked --offline --test heavy_slot_cli` 21 passed 0 failed(`macos_is_judged_on_memory_pressure_not_swap` 포함), Python 25 tests OK(`test_c88_macos_is_judged_on_memory_pressure_not_swap` 포함); 실측(같은 날 이 Mac): 압력 수준 2·여유 37 %·스왑 97.7 % 에서 `heavy-slot-acquire -- /usr/bin/true` exit 0.
- AC-2: met — `cargo test --locked --offline --lib heavy_slot::` 21 passed 0 failed(`limits_tests::the_memory_free_flag_only_tightens` 포함); Python 의 macOS 판단 표가 같은 경계(9→10 각주, `limits.json` 20>5·20/19, 0 으로 하향, 15 대 14/15)를 통과.
- AC-3: met — Python 25 tests OK(Linux 스왑 80 통과·81 거부, `--min-mem-free-pct 90` 내주어도 메모리 거부 없음, 부분 판독 사례 포함); `heavy_slot_cli` 21 passed(Linux 사례 포함).
- AC-4: met — `cargo fmt --check` exit 0, `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0(2026-11-19, 같은 worktree); 병합 시점 전수 기록은 PR #3151 본문.
