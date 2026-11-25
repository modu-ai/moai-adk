---
id: SPEC-20261126-heavy-slot-cpu-free
title: "Heavy slot waits on CPU headroom, not on the load average"
status: draft
revision: 1
created: 2026-11-26
---

# Heavy slot waits on CPU headroom, not on the load average

## Intent

무거운 실행 자리(heavy slot)가 부하 평균(load average) 대신 CPU 여유로 판단하게 해요. 2026-11-25T03:25Z 실측(`~/.moai/reports/moai-adk/2026-11-25-build-test-load-rootcause.md` 2절): 부하 평균이 111.67 인 순간에 실행 가능 프로세스 107개 중 76개가 브라우저 보조 프로세스(CPU 합 6%)였고 CPU 는 사용 88%·여유 12% 였어요. 리더가 부하 평균(50, 81)으로 슬롯 수를 낮춰 온 건 잘못된 신호였어요.

읽은 사실(NeXT 팁 `882cba7497`):

1. 슬롯 래퍼는 줄을 선 뒤 슬롯을 잡고, 디스크·메모리가 모자라면 종료 75 로 시작하지 않아요(`src/heavy_slot.rs` `over_limits`). CPU 는 아무 데서도 보지 않고 부하 평균은 대기 줄과 종료 줄에 찍기만 해요(`machine_line`).
2. 기계 한도는 슬롯 폴더의 `limits.json` 에 있어요(`max`, `min_free_gb`, `max_swap_pct`, `min_mem_free_pct`). 모르는 키는 거절하고, 호출자의 플래그는 더 엄격할 때만 이겨요(`effective`).
3. 시험은 `MOAI_TEST_MACHINE`(디버그 빌드 전용)으로 디스크·swap·메모리 값을 넣어요. `tests/test_switch_release_scan.rs` 가 이 변수를 디버그 `cfg` 아래에서만 읽는지 지켜요.
4. CPU 틱(tick)은 부팅 뒤 누적값이라 한 번 읽으면 평균일 뿐이에요. 두 번 읽은 차이로만 지금의 여유를 알 수 있어요.

이 SPEC 은 다섯을 해요. (a) `heavy-slot-status` 가 CPU 여유, 코어 수, 기준, 실행 중인 무거운 실행 수를 말해요. (b) CPU 여유가 기준보다 낮으면 `heavy-slot-acquire` 가 슬롯을 잡기 전에 기다려요. (c) 기다린 초를 슬롯 줄과 종료 줄에 써요. (d) 기다림에 위쪽 한도(300초)를 둬요. (e) 기준값 `min_cpu_free_pct` 를 `limits.json` 에서 읽어요(기본 20).

## Interview

- Q1: 카드 기준 3 의 "프로필"을 프로젝트 프로필로 읽을까요, 기계 한도 파일로 읽을까요? → A: 기계 한도 파일 `limits.json` 의 `min_cpu_free_pct`(기본 20)로 읽어요(리더 결정 노트 9510; 슬롯은 기계 전체라 프로젝트마다 다른 숫자로 판단하면 안 돼요)

## Non-goals

- 프로젝트 프로필(`.moai/project.yaml`)에 새 키를 더하는 일.
- `--min-cpu-free-pct` 호출자 플래그. 더 엄격하게만 이기는 플래그는 `src/main.rs` 의 인자와 `effective` 호출 네 곳을 건드려서 뒤로 미뤄요(D-4).
- 종료 줄(`machine_line`)에 CPU 를 넣는 일. 매 줄이 100 ms 늦어져서 하지 않아요.
- 슬롯 개수(`max`)를 CPU 여유로 자동 조절하는 일. 이 SPEC 은 기다림만 더해요.
- 이어받은 슬롯(바깥 래퍼가 넘긴 슬롯)을 다시 판단하는 일.
- 이 저장소나 사용자 홈의 실제 슬롯 폴더와 실제 `limits.json` 을 건드리는 일. 시험은 임시 폴더에서만 해요.

## Requirements

- REQ-1: When `heavy-slot-status` reads the machine, the status shall report `cpu_free_pct`, `cpu_cores`, `min_cpu_free_pct` and `heavy_runs` · failure: none — a reading the machine cannot give is `null`, `heavy_runs` counts every held slot including one held by a dead pid, and the status stays read-only.
- REQ-2: When the CPU free percentage is below `min_cpu_free_pct`, `heavy-slot-acquire` shall wait before it takes a slot · failure: none — a reading that cannot be taken never makes a caller wait, a slot inherited from an outer holder is never judged again, and a limit of 0 switches the rule off.
- REQ-3: When `heavy-slot-acquire` has waited for CPU headroom, it shall write the waited seconds on its slot line and on its exit line · failure: none — a caller that did not wait gets no extra text.
- REQ-4: When the wait for CPU headroom reaches 300 seconds, `heavy-slot-acquire` shall start COMMAND anyway · failure: none — the slot line says that the cap ended the wait and how many seconds were waited.
- REQ-5: The slot tools shall read `min_cpu_free_pct` from `limits.json` beside the slots and use 20 when the file has no such key · failure: none — a caller that passes no limits file gets the default.
- REQ-6 (REQ-5): IF `min_cpu_free_pct` is above 100, THEN the limits reader shall refuse the file with exit code 125 and run nothing.
- REQ-7: When `check affected` ends a run whose slot waited for CPU headroom, it shall add `slot_wait=<seconds>` to the summary line · failure: none — a run that did not wait adds nothing, so every summary line read today stays as it is.
- REQ-8: When the CPU rule changes what a person reads, the module documentation of `src/heavy_slot.rs` shall state the rule, the limit key, the cap and the waited seconds · failure: none — documentation only.

## Acceptance

- AC-1 (REQ-1): with injected readings (`cpu_free_pct`, `cpu_cores`) the status carries the four keys, `null` for a reading that is not given, and a held slot counts in `heavy_runs` · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac1_`
- AC-2 (REQ-2, REQ-3): with a limit of 20 and an injected CPU free of 5, a caller prints a waiting line with the reading and starts nothing; after the reading is raised to 90 it runs COMMAND, exits with COMMAND's code and its exit line names the waited seconds (the dangerous path itself, in a temporary slot folder with a real `touch` command) · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac2_`
- AC-3 (REQ-5, REQ-6): the default is 20, a file value 30 wins, a CPU free of 19 waits and 20 runs, 0 switches the rule off, 101 is refused with exit 125 and nothing runs · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac3_`
- AC-4 (REQ-4): the decision after a wait is a pure function: below the limit and under 300 seconds it waits, at 300 seconds it starts, and a wrapping tick difference is read right · verify: `cargo test --locked --offline --lib heavy_slot::cpu_tests`
- AC-5 (REQ-7): the summary line carries `slot_wait=<seconds>` only when the slot waited · verify: `cargo test --locked --offline --test check_affected_one_slot cpu_ac5_`
- AC-6 (REQ-8): the module documentation names the CPU rule, `min_cpu_free_pct`, the 300-second cap and the waited seconds · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac6_`
- AC-7 (REQ-1, REQ-2, REQ-3, REQ-5): the existing slot tests still pass, none replaced by a weaker assertion · verify: `cargo test --locked --offline --test heavy_slot_cli --test heavy_slot_fifo_cli --test heavy_slot_jobs_cleanup_cli --test check_affected_one_slot`

## Scope

- write: `.moai/specs/SPEC-20261126-heavy-slot-cpu-free/spec.md`
- write: `src/heavy_slot.rs`
- write: `tests/heavy_slot_cli.rs`
- write (when its holder gives it back, REQ-7): `src/check_affected.rs`, `src/ops/check.rs`, `tests/check_affected_one_slot.rs`
- read-only: `src/audit_prep.rs`, `src/check_fast.rs`, `src/main.rs`, `tests/state_folder_literal_scan.rs`, `tests/test_switch_release_scan.rs`
- forbidden: the real slot folder and the real `limits.json` of the user, `.moai/project.yaml`, `CLAUDE.md`, `tools/repo_scans.json`

## Plan

1. Red tests `cpu_ac1_`, `cpu_ac2_`, `cpu_ac3_` in `tests/heavy_slot_cli.rs`: they compile and fail by their assertions.
2. The limit key (type, range 0..=100, default 20, refusal of 101), the injected keys `cpu_free_pct` and `cpu_cores` on the existing `MOAI_TEST_MACHINE`, the CPU sampler, the new `heavy-slot-status` keys.
3. The wait in the queue loop, the waited seconds, the 300-second cap, the text of the slot line and the exit line.
4. `slot_wait=` in the `check affected` summary once `src/check_affected.rs` is free.
5. The dangerous path in a throwaway: a limits file with `min_cpu_free_pct` 100 so a real caller waits, and then a lowered reading so it starts.

## Risks

- A busy machine keeps CPU free under the limit for good (the measured case itself: 12% free with a limit of 20). The cap of 300 seconds is the way out; the exit line says the cap ended the wait, so a slow day shows up in the lines and not as a hang.
- The first look in an empty queue has no second sample. The sampler takes a baseline at the start and a second reading at least 100 ms later, so an empty queue costs 100 ms once; a limit of 0 and injected readings skip it.
- A new variable would be a gated seam in `tests/state_folder_literal_scan.rs` and an edit of the shared `tools/repo_scans.json`. The injected keys reuse `MOAI_TEST_MACHINE`, so no registry file changes.
- macOS `mach_host_self` is marked deprecated in the locked `libc`; clippy runs with `-D warnings`. The one call carries an `#[allow(deprecated)]` with the reason, and no new dependency is added.
- Tests that use a `limits.json` and the real machine reading would wait on a loaded machine. Those tests state `min_cpu_free_pct` 0 (D-5).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The threshold is `min_cpu_free_pct` in the machine limits file `<slot folder>/limits.json` (default 20), not a key of the project profile. The card's word "profile" is read as this file. | Leader decision note 9510 on this SPEC's card: the slots are one machine-wide resource, so two projects must not judge the same slots by different numbers; `limits.json` already holds the machine limits with a strict key list. | Move the key into the profile and read it in `effective`. |
| D-2 | CPU free is `idle ticks / all ticks` over two readings at least 100 ms apart (`host_statistics` with `HOST_CPU_LOAD_INFO` on macOS, the first line of `/proc/stat` on Linux); tick differences wrap on 32 bits. | The ticks are cumulative since boot, so one reading says nothing about now. | Read the load average again. |
| D-3 | Low CPU makes a caller wait; it is not a refusal like low disk or memory (exit 75). The wait ends after 300 seconds and COMMAND starts. | The card says "wait". An unbounded wait could hang every check on a day another application keeps CPU busy. 300 seconds is an estimate to tune with the first measured waits. | Exit 75 after the cap, or no cap. |
| D-4 | No `--min-cpu-free-pct` flag in this revision. | The flag needs `src/main.rs` and four call sites of `effective`; `src/main.rs` is held by another card. A later card can add it. | Add the flag with the "only tightens" rule. |
| D-5 | `MOAI_TEST_MACHINE` gets the keys `cpu_free_pct` and `cpu_cores`; without them no CPU rule applies, and tests that write a limits file say `min_cpu_free_pct` 0. | A new variable needs a registry entry in a shared tier-2-floor file; the old tests stay as they are. | Add a gated seam and its registry entry. |
| D-6 | `slot_wait=<seconds>` is added to the `check affected` summary line only when the slot waited. | Every reader of the summary line reads it word by word, and tests that compare the whole line stay green when nothing waited. | Put the number in the answer JSON only. |
