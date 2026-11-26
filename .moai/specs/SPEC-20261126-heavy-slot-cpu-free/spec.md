---
id: SPEC-20261126-heavy-slot-cpu-free
title: "Heavy slot waits on CPU headroom, not on the load average"
status: draft
revision: 2
created: 2026-11-26
---

# Heavy slot waits on CPU headroom, not on the load average

## Intent

무거운 실행 자리(heavy slot)가 부하 평균(load average) 대신 CPU 여유로 판단할 수 있게 해요. 2026-11-25T03:25Z 실측(`~/.moai/reports/moai-adk/2026-11-25-build-test-load-rootcause.md` 2절): 부하 평균이 111.67 인 순간에 실행 가능 프로세스 107개 중 76개가 브라우저 보조 프로세스(CPU 합 6%)였고 CPU 는 사용 88%·여유 12% 였어요. 리더가 부하 평균(50, 81)으로 슬롯 수를 낮춰 온 건 잘못된 신호였어요.

읽은 사실(NeXT 팁 `882cba7497`):

1. 슬롯 래퍼는 줄을 선 뒤 슬롯을 잡고, 디스크·메모리가 모자라면 종료 75 로 시작하지 않아요(`src/heavy_slot.rs` `over_limits`). CPU 는 아무 데서도 보지 않고 부하 평균은 대기 줄과 종료 줄에 찍기만 해요(`machine_line`).
2. 기계 한도는 슬롯 폴더의 `limits.json` 에 있어요(`max`, `min_free_gb`, `max_swap_pct`, `min_mem_free_pct`). 모르는 키는 거절하고, 호출자의 플래그는 더 엄격할 때만 이겨요(`effective`).
3. 시험은 `MOAI_TEST_MACHINE`(디버그 빌드 전용)으로 디스크·swap·메모리 값을 넣어요. `tests/test_switch_release_scan.rs` 가 이 변수를 디버그 `cfg` 아래에서만 읽는지 지켜요.
4. CPU 틱(tick)은 부팅 뒤 누적값이라 한 번 읽으면 평균일 뿐이에요. 두 번 읽은 차이로만 지금의 여유를 알 수 있어요.

이 SPEC 은 여섯을 해요. (a) `heavy-slot-status` 가 CPU 여유, 코어 수, 기준, 실행 중인 무거운 실행 수를 말해요. (b) 기준이 켜져 있고 CPU 여유가 그보다 낮으면 `heavy-slot-acquire` 가 슬롯을 잡기 전에 기다려요. (c) 기다린 초를 슬롯 줄과 종료 줄에 써요. (d) 기다림에 위쪽 한도(300초)를 둬요. (e) 기준값 `min_cpu_free_pct` 를 `limits.json` 에서 읽어요(기본 0 = 끔). (f) 슬롯 줄에 잰 CPU 여유를 남겨 켜기 전에 관찰해요.

## Interview

- Q1: 카드 기준 3 의 "프로필"을 프로젝트 프로필로 읽을까요, 기계 한도 파일로 읽을까요? → A: 기계 한도 파일 `limits.json` 의 `min_cpu_free_pct` 로 읽어요(리더 결정 노트 9510; 슬롯은 기계 전체라 프로젝트마다 다른 숫자로 판단하면 안 돼요)
- Q2: 기본으로 켤까요? → A: 기본값은 0(끔, 기다리지 않음)이고 사용자가 값을 넣으면 켜져요. 새 기다림은 관찰부터 시작해요. 켤지 고를 수 있게 슬롯 줄 끝에 잰 값을 남겨요(리더 결정 노트 10057, 10047 을 바꿈)

## Non-goals

- 프로젝트 프로필(`.moai/project.yaml`)에 새 키를 더하는 일.
- `--min-cpu-free-pct` 호출자 플래그. 더 엄격하게만 이기는 플래그는 `src/main.rs` 의 인자와 `effective` 호출 네 곳을 건드려서 뒤로 미뤄요(D-4).
- 종료 줄(`machine_line`)에 CPU 를 넣는 일. 관찰 값은 슬롯 줄에 한 번만 남겨요.
- 슬롯 개수(`max`)를 CPU 여유로 자동 조절하는 일.
- 이어받은 슬롯(바깥 래퍼가 넘긴 슬롯)을 다시 판단하는 일.
- 이 저장소나 사용자 홈의 실제 슬롯 폴더와 실제 `limits.json` 을 건드리는 일. 시험은 임시 폴더에서만 해요.

## Requirements

- REQ-1: When `heavy-slot-status` reads the machine, the status shall report `cpu_free_pct`, `cpu_cores`, `min_cpu_free_pct` and `heavy_runs` · failure: none — a reading the machine cannot give is `null`, `heavy_runs` counts every held slot including one held by a dead pid, and the status stays read-only.
- REQ-2: When the CPU free percentage is below `min_cpu_free_pct`, `heavy-slot-acquire` shall wait before it takes a slot · failure: none — a reading that cannot be taken never makes a caller wait, a later look in which no new tick passed keeps the last percentage, a slot inherited from an outer holder is never judged again, and a limit of 0 switches the rule off.
- REQ-3: When `heavy-slot-acquire` has waited for CPU headroom, it shall write the waited seconds on its slot line and on its exit line · failure: none — a caller that did not wait gets no waiting text and no waited seconds.
- REQ-4: When the wait for CPU headroom reaches 300 seconds, `heavy-slot-acquire` shall start COMMAND anyway · failure: none — the slot line says that the cap ended the wait and how many seconds were waited.
- REQ-5: The slot tools shall read `min_cpu_free_pct` from `limits.json` beside the slots and use 0 when the file has no such key · failure: none — a caller that passes no limits file gets the default and never waits.
- REQ-6 (REQ-5): IF `min_cpu_free_pct` is above 100, THEN the limits reader shall refuse the file with exit code 125 and run nothing.
- REQ-7: When `check affected` ends a run whose slot waited for CPU headroom, it shall add `slot_wait=<seconds>` to the summary line · failure: none — a run that did not wait adds nothing, so every summary line read today stays as it is.
- REQ-8: When the CPU rule changes what a person reads, the module documentation of `src/heavy_slot.rs` shall state the rule, the limit key, the cap and the waited seconds · failure: none — documentation only.
- REQ-9: When `heavy-slot-acquire` takes a slot and the machine gives a CPU reading, the slot line shall name the measured CPU free percentage and the limit it was held against, or `rule off` · failure: none — a machine that gives no reading adds nothing, and an inherited slot names none.

## Acceptance

- AC-1 (REQ-1): with injected readings (`cpu_free_pct`, `cpu_cores`) the status carries the four keys, `null` for a reading that is not given, and a held slot counts in `heavy_runs` · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac1_`
- AC-2 (REQ-2, REQ-3): with a limit of 20 and an injected CPU free of 5, a caller prints a waiting line with the reading and starts nothing; after the reading is raised to 90 it runs COMMAND, exits with COMMAND's code and its exit line names the waited seconds (the dangerous path itself, in a temporary slot folder with a real `touch` command); on the real reading with a limit of 100 the caller still waits after more than a second of looks, ending it starts nothing, and the next caller clears its waiting file · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac2_`
- AC-3 (REQ-5, REQ-6): the default is 0 and a CPU free of 1 runs at once, a file value 30 wins, with a limit of 20 a CPU free of 19 waits and 20 runs, 0 switches the rule off, 101 is refused with exit 125 and nothing runs · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac3_`
- AC-4 (REQ-4): the decision after a wait is a pure function: below the limit and under 300 seconds it waits, at 300 seconds it starts, and a wrapping tick difference is read right · verify: `cargo test --locked --offline --lib heavy_slot::cpu_tests`
- AC-5 (REQ-7): the summary line carries `slot_wait=<seconds>` only when the slot waited · verify: `cargo test --locked --offline --test check_affected_one_slot cpu_ac5_`
- AC-6 (REQ-8): the module documentation names the CPU rule, `min_cpu_free_pct`, the 300-second cap and the waited seconds · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac6_`
- AC-7 (REQ-1, REQ-2, REQ-3, REQ-5): the existing slot tests still pass, none replaced by a weaker assertion · verify: `cargo test --locked --offline --test heavy_slot_cli --test heavy_slot_fifo_cli --test heavy_slot_jobs_cleanup_cli --test check_affected_one_slot`
- AC-8 (REQ-9): the slot line ends with `cpu free N% (rule off)` when there is no limit and names the limit when there is one, on injected readings and on the real one, and says nothing about CPU without a reading · verify: `cargo test --locked --offline --test heavy_slot_cli cpu_ac8_`

## Scope

- write: `.moai/specs/SPEC-20261126-heavy-slot-cpu-free/spec.md`
- write: `src/heavy_slot.rs`
- write: `tests/heavy_slot_cli.rs`
- write (when its holder gives it back, REQ-7): `src/check_affected.rs`, `src/ops/check.rs`, `tests/check_affected_one_slot.rs`
- read-only: `src/audit_prep.rs`, `src/check_fast.rs`, `src/main.rs`, `tests/state_folder_literal_scan.rs`, `tests/test_switch_release_scan.rs`
- forbidden: the real slot folder and the real `limits.json` of the user, `.moai/project.yaml`, `CLAUDE.md`, `tools/repo_scans.json`

## Plan

1. Red tests `cpu_ac1_`, `cpu_ac2_`, `cpu_ac3_`, `cpu_ac8_` in `tests/heavy_slot_cli.rs`: they compile and fail by their assertions.
2. The limit key (type, range 0..=100, default 0, refusal of 101), the injected keys `cpu_free_pct` and `cpu_cores` on the existing `MOAI_TEST_MACHINE`, the CPU sampler, the new `heavy-slot-status` keys.
3. The wait in the queue loop, the waited seconds, the 300-second cap, the text of the slot line and the exit line, the measured CPU free on the slot line.
4. `slot_wait=` in the `check affected` summary once `src/check_affected.rs` is free.
5. The dangerous path as integration tests in temporary folders: a caller held back that starts after the reading rises, and a caller on the real reading that is ended while it waits.

## Risks

- With a limit set, a busy machine can keep CPU free under it for good (measured 2026-11-26 on the 16-core Mac at load 44 to 80: 0 to 15% free). The cap of 300 seconds is the way out; the slot line says the cap ended the wait, so a slow day shows up in the lines and not as a hang.
- Every taken slot reads the CPU once, the rule off included: usually 100 ms, and at most about 2 seconds while the system hands out unchanged counters.
- An earlier binary refuses a `limits.json` that holds the new key (the installed build 54 could not read such a file), and `heavy-slot-acquire` then runs nothing. The key is added only after a build with this change is installed, and going back to an earlier build needs the key removed.
- A new variable would be a gated seam in `tests/state_folder_literal_scan.rs` and an edit of the shared `tools/repo_scans.json`. The injected keys reuse `MOAI_TEST_MACHINE`, so no registry file changes.
- macOS `mach_host_self` is marked deprecated in the locked `libc`; clippy runs with `-D warnings`. The one call carries an `#[allow(deprecated)]` with the reason, and no new dependency is added.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The threshold is `min_cpu_free_pct` in the machine limits file `<slot folder>/limits.json`, not a key of the project profile. The card's word "profile" is read as this file. | Leader decision note 9510 on this SPEC's card: the slots are one machine-wide resource, so two projects must not judge the same slots by different numbers; `limits.json` already holds the machine limits with a strict key list. | Move the key into the profile and read it in `effective`. |
| D-2 | CPU free is `idle ticks / all ticks` over two readings at least 100 ms apart (`host_statistics` with `HOST_CPU_LOAD_INFO` on macOS, the first line of `/proc/stat` on Linux); tick differences wrap on 32 bits. A reading equal to the one before it is not a sample: the earlier reading stays the base and the last percentage holds; the first look reads again for at most 2 seconds. | The ticks are cumulative since boot, so one reading says nothing about now. macOS answered the same counters to repeated calls for up to about a second (measured 2026-11-26 on a 16-core Mac), and treating that as unreadable ended a real wait after 1 second. | Read the load average again. |
| D-3 | Low CPU makes a caller wait; it is not a refusal like low disk or memory (exit 75). The wait ends after 300 seconds and COMMAND starts. | The card says "wait". An unbounded wait could hang every check on a day another application keeps CPU busy. 300 seconds is an estimate to tune with the first measured waits. | Exit 75 after the cap, or no cap. |
| D-4 | No `--min-cpu-free-pct` flag in this revision. | The flag needs `src/main.rs` and four call sites of `effective`; `src/main.rs` is held by another card. A later card can add it. | Add the flag with the "only tightens" rule. |
| D-5 | `MOAI_TEST_MACHINE` gets the keys `cpu_free_pct` and `cpu_cores`; without them an injected machine gives no CPU reading. | A new variable needs a registry entry in a shared tier-2-floor file; the old tests stay as they are. | Add a gated seam and its registry entry. |
| D-6 | `slot_wait=<seconds>` is added to the `check affected` summary line only when the slot waited. | Every reader of the summary line reads it word by word, and tests that compare the whole line stay green when nothing waited. | Put the number in the answer JSON only. |
| D-7 | The default of `min_cpu_free_pct` is 0: the rule is off until a person writes a value. | Leader decision note 10057 (replacing 10047): a new wait starts as observation, and no one writes the `limits.json` under HOME without the user. | Make the default a number above 0. |
| D-8 | The slot line ends the taken or reuse text with `; cpu free N% (limit M%)` or `; cpu free N% (rule off)`, before any wait text. | Leader decision note 10057: the measured numbers must stay so the user can choose whether to switch the rule on. One line per taken slot; a line that waited still ends with the waited seconds. | Drop the text, or move it to the exit line. |

## Result

2026-11-26 기준이에요. REQ-1 부터 REQ-9 까지 구현했어요. NeXT 팁 `75b6920066` 를 브랜치에 합친(`bab95d2cc9`) 뒤에 REQ-7 을 넣었어요.

바뀐 파일: `src/heavy_slot.rs`, `src/check_affected.rs`, `src/ops/check.rs`, `tests/heavy_slot_cli.rs`, `tests/check_affected_one_slot.rs`.

코드 head `f5af3ae71c8bb8c0fca3a16d82b8a3c58900ee30` 에서 돌린 확인이에요(macOS, `--max 2` 슬롯, 로그 `target/zoe/green-ac5.log`, 종료 0). AC 의 이름 필터는 따로 돌리지 않고 그 필터가 들어 있는 대상 전체를 돌렸어요.

| 확인 | 결과 | 덮는 AC |
| --- | --- | --- |
| `cargo clippy --locked --offline --lib --test heavy_slot_cli --test check_affected_one_slot -- -D warnings` | 통과 | — |
| `cargo test --locked --offline --lib heavy_slot` | 32 통과 | AC-4 |
| `--test heavy_slot_cli` | 54 통과 | AC-1, AC-2, AC-3, AC-6, AC-8 |
| `--test check_affected_one_slot` | 9 통과 (`cpu_ac5_` 2개 포함) | AC-5 |
| `--test heavy_slot_fifo_cli`, `--test heavy_slot_jobs_cleanup_cli` | 10, 6 통과 | AC-7 |
| `--test test_switch_release_scan`, `--test cli_text_citations` | 4, 5 통과 | — |

AC-5 의 실패 확인: `bab95d2cc9` 에 테스트만 더해(`6b79bebf2b`) `cargo test --locked --offline --test check_affected_one_slot cpu_ac5_` 를 돌리면 종료 101 이에요. 기다린 실행의 테스트가 `no slot_wait word` 단언에서 실패하고, 기다리지 않은 실행의 테스트는 통과해요.

구현 중에 찾아 고친 것: macOS 는 짧은 간격의 두 CPU 틱 읽기에 같은 값을 돌려줘서, 처음 코드는 실제 기다림을 1초 만에 끝냈어요. 같은 값은 표본으로 치지 않고 앞의 비율을 유지해요(`88e69902c0`, D-2).

관찰: 16코어 Mac 에서 부하 44~80 일 때 CPU 여유는 0~15% 였어요. 기본값은 0(끔)으로 두고 슬롯 줄의 수치를 모아요(사용자 결정, 카드 노트 10124).

남은 한계:

- Linux 와 Windows 에서는 돌리지 않았어요. 릴리스 빌드도 만들지 않았어요.
- 설치된 이전 빌드는 새 키가 있는 `limits.json` 을 거절해요. 키는 새 빌드를 설치한 뒤에만 넣고, 되돌리기 전에 지워요.
- 이 카드의 영수증은 설치된 빌드의 `check affected` 가 만들어요. 그 줄에는 `slot_wait` 가 나올 수 없어요.
