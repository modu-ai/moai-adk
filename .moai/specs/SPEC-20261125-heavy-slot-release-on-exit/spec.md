---
id: SPEC-20261125-heavy-slot-release-on-exit
title: "heavy-slot-acquire gives the slot back when its command ends, and heavy-slot-status warns about a slot a leftover process holds"
status: draft
revision: 1
created: 2026-11-25
branch: spec/heavy-slot-release-on-exit-20261125
---

# heavy-slot-acquire gives the slot back when its command ends, and heavy-slot-status warns about a slot a leftover process holds

## Intent

끝난 검사가 무거운 실행 자리(heavy slot)를 계속 쥐어서 이 기계의 모든 검사가 한 줄로 서는 일이 또 생기지 않게 해요. 2026-11-25 실측(이 일을 맡긴 카드 본문과 검사 속도 카드의 노트): 슬롯 4개 중 3개가 `alive=false` 인데 held 였고(하나는 1403분), 쥔 쪽은 `launcher_auto_init` 의 대기 테스트가 남긴 가짜 호스트 스크립트(`/bin/sh …/.tmpXXXX/claude --name lane-N`, 부모 pid 1)였어요. 사용자가 "왜 레인들이 멈춰져있지?"라고 물었고 "이런 문제 없도록 죽음검사가 또 발생하지 않도록 개선을 하자"고 지시했어요.

읽은 사실(NeXT 팁 `12dab143c9`, 이 트리):

1. 슬롯은 `slot-<i>` 파일의 `flock` 이고, 잠금은 열린 파일 기술자(open file description) 것이에요(`src/heavy_slot.rs` 머리말 7~18줄). 래퍼는 잠근 파일을 COMMAND 에 넘기려고 close-on-exec 를 일부러 풀어요(`keep_for_children`, 1265줄). 그래서 COMMAND 가 남긴 자식·후손이 그 기술자를 물려받으면 래퍼와 COMMAND 가 끝난 뒤에도 슬롯이 잡혀 있어요. 운영체제는 기술자를 가진 프로세스가 모두 끝나야 잠금을 풀어요.
2. 이 동작은 설계로 적혀 있고 시험으로 못박혀 있어요: `SPEC-20261117-heavy-slot-status`(Issue #3033)는 "원인은 설계예요"라고 쓰고 보여주기만 하며, `tests/heavy_slot_cli.rs:358` 의 `a_command_that_leaves_a_background_child_keeps_the_slot_and_status_shows_a_dead_pid` 가 "오늘의 동작을 고정"해요. 이 SPEC 은 그 고정을 뒤집어요(Decisions D-1).
3. 이미 있는 것: `heavy-slot-status` 는 기록 pid 가 죽었으면 `alive: false` 와 `holders`(pid·프로세스 이름, argv 없음)를 줘요(`held_entry`, 1577줄). 없는 것: 사람이 한눈에 읽는 경고 줄과 `warnings` 모음. 지금은 JSON 항목 안쪽에 묻혀 있어요.
4. `flock(fd, LOCK_UN)` 은 같은 열린 파일 기술자를 나눠 가진 모든 기술자의 잠금을 한꺼번에 풀어요(기술자를 닫는 것과 달라요). 그래서 래퍼가 끝날 때 한 번 풀면, 후손이 기술자를 쥐고 있어도 슬롯이 비어요. 슬롯 파일을 지우거나 프로세스를 끝내는 일은 필요 없어요.
5. 후손이 물려받은 낡은 기술자로 다시 슬롯을 요청해도 안전해요: `inherited()` 는 남이 슬롯을 쥔 지금 `flock(LOCK_EX|LOCK_NB)` 가 실패하고(그 기술자의 잠금은 풀렸고 다른 열린 파일 기술자가 쥐고 있어서), 아무도 안 쥐면 3단계에서 `None` 이라서 줄을 서요(`src/heavy_slot.rs:1249-1255`).
6. 대기 테스트의 가짜 호스트(`tests/launcher_auto_init.rs:42-49`)는 `MOAI_TEST_HOLD_UNTIL` 파일이 생길 때까지 도는 루프예요. 시험이 끝나면서 그 파일이 든 임시 폴더가 지워지거나, 런처가 먼저 죽으면 영원히 돌아요. 그 자식이 슬롯 기술자까지 물려받아 1번의 문제를 키웠어요.

이 SPEC 은 셋을 해요. (a) 슬롯을 쥔 쪽(래퍼, `check affected`, `audit-prep`)이 끝날 때 잠금을 한 번 풀어요. (b) `heavy-slot-status` 가 죽은 pid 의 슬롯을 경고 줄과 `warnings` 로 올려요. (c) 가짜 호스트가 부모가 끝나거나 시간이 지나면 스스로 끝나요.

## Interview

- Q1: 슬롯을 쥔 쪽이 끝날 때 잠금을 풀어서, 그 뒤에도 사는 후손이 슬롯을 더는 쥐지 못하게 할까요(D-1 선택지 a)? → A: (사용자 답 대기)
- Q2: 이 SPEC 의 판 1을 승인할까요? → A: (사용자 답 대기)

## Non-goals

- 프로세스를 끝내거나 슬롯 파일을 지우거나 고치는 명령. 끝내는 일은 소유자 몫이고 슬롯 파일은 손으로 지우지 않아요(`src/heavy_slot.rs` 머리말). 회수는 잠금을 푸는 것으로만 해요.
- 래퍼가 SIGKILL 로 죽을 때의 회수. 그때는 잠금을 풀 코드가 못 돌아서 후손이 다 끝날 때까지 슬롯이 잡혀 있어요(바뀌지 않음). 경고(REQ-3)가 그 슬롯을 보여줄 뿐이에요.
- 중첩 호출이 바깥 슬롯을 다시 쓰는 규칙(`SPEC-20261122-check-affected-one-slot` REQ-5)과 줄 서기 순서. 그대로예요.
- 이 저장소나 사용자 홈의 실제 슬롯 폴더를 건드리는 일. 시험은 임시 폴더에서만 해요.
- 지금 쥔 슬롯을 치우는 일(리더와 사용자 몫).

## Requirements

- REQ-1: When a process that holds a heavy slot (`heavy-slot-acquire`, `check affected`, `audit-prep`) ends its own hold, the holder shall unlock the slot on the open file description it took before it lets the descriptor go, so a descendant that still runs and inherited the descriptor no longer keeps the slot · failure: none — a holder killed by SIGKILL unlocks nothing and the slot stays held until every process with the file is gone (unchanged), and a caller that reuses an inherited slot never unlocks the outer holder's slot.
- REQ-2: When a caller asks for a slot through an inherited descriptor whose lock was released, the caller shall queue like any other caller · failure: none — it never takes a slot another caller holds and never reuses a slot nobody holds.
- REQ-3: When `heavy-slot-status` reads a held slot whose recorded pid is not alive, the status shall carry one warning line for that slot, naming the slot, the dead pid and every process that has the slot file open by pid and name, and a top-level `warnings` array of all such lines · failure: none — a slot of a live pid and a free slot add no warning, the status stays read-only, holders that cannot be read say `unavailable`, and no argv is shown.
- REQ-4: When the stub host of `tests/launcher_auto_init.rs` waits for its release file, the stub shall end as soon as the process that started it has ended and no later than 300 seconds after it started · failure: none — a stub released in time ends at once as before.
- REQ-5: When the release rule or the warning changes what a person reads, the module documentation of `src/heavy_slot.rs` shall state the release at holder exit, the cases it cannot cover and the `warnings` array · failure: none — documentation only.

## Acceptance

- AC-1 (REQ-1): a command that starts a background child and ends leaves no held slot in `heavy-slot-status` while the child is alive, and a second caller runs at once without waiting for the child (the dangerous path itself, in a temporary slot folder with a real `sleep` child that outlives its wrapper) · verify: `cargo test --locked --offline --test heavy_slot_cli release_ac1_`
- AC-2 (REQ-1): a nested caller that reuses the slot its outer holder passed down leaves the outer slot held when it ends, and a wrapper killed by SIGKILL while its command still runs keeps the slot held · verify: `cargo test --locked --offline --test heavy_slot_cli release_ac2_`
- AC-3 (REQ-2): a leftover child that asks for a slot with the descriptor and `MOAI_HEAVY_SLOT_FD` it inherited waits in the queue while another caller holds the slot, and runs after that caller ends · verify: `cargo test --locked --offline --test heavy_slot_cli release_ac3_`
- AC-4 (REQ-3): after a wrapper is killed while its command still runs, the status lists the slot with `alive: false`, a warning line with the slot, the dead pid and the holders' pid and name, and the same line in `warnings`; a slot of a live pid adds no warning · verify: `cargo test --locked --offline --test heavy_slot_cli release_ac4_`
- AC-5 (REQ-4): a launcher started by the test is killed with SIGKILL before it releases its stub host, and the stub host is gone within 10 seconds, with the release file never written · verify: `cargo test --locked --offline --test launcher_auto_init stub_host_ends_`
- AC-6 (REQ-5): the module documentation names the release, the SIGKILL case and `warnings` · verify: `cargo test --locked --offline --test heavy_slot_cli release_ac6_`
- AC-7 (REQ-1, REQ-2, REQ-3): the existing slot tests still pass, and the test that pinned the old keep-the-slot behaviour is replaced by AC-1 and AC-4, none by a weaker assertion · verify: `cargo test --locked --offline --test heavy_slot_cli --test heavy_slot_fifo_cli --test heavy_slot_jobs_cleanup_cli --test check_affected_one_slot --test launcher_auto_init`

## Scope

- write: `.moai/specs/SPEC-20261125-heavy-slot-release-on-exit/spec.md`
- write: `src/heavy_slot.rs`
- write: `tests/heavy_slot_cli.rs`, `tests/launcher_auto_init.rs`
- read-only: `src/check_affected.rs`, `src/audit_prep.rs`, `src/home_state.rs`
- forbidden: the real slot folder of the user and every process of the user, `.moai/project.yaml`, `CLAUDE.md`

## Plan

1. 테스트 먼저(RED): `tests/heavy_slot_cli.rs` 에 `release_ac1_`~`release_ac4_`·`release_ac6_`, `tests/launcher_auto_init.rs` 에 `stub_host_ends_`. 변경 전 트리에서 실패하는 것을 읽어요. 못박힌 기존 시험(`a_command_that_leaves_a_background_child_keeps_the_slot…`)은 AC-1·AC-4 로 대체해요(SIGKILL 로 죽은 래퍼의 낡은 슬롯 시험은 AC-4 가 이어받아요).
2. `src/heavy_slot.rs`: `Held` 에 `Drop` 을 더해 `file` 이 있을 때만 `flock(fd, LOCK_UN)` 을 불러요(`file` 이 없는 물려받은 슬롯은 바깥 소유자 것이라 건드리지 않아요). `status` 는 죽은 pid 의 슬롯마다 경고 한 줄을 만들어 항목의 `warning` 과 최상위 `warnings` 에 넣어요. 머리말 문서를 새 동작으로 고쳐요.
3. `tests/launcher_auto_init.rs`: 가짜 호스트 루프에 부모(`$PPID`)가 살아 있는지와 최대 시간을 더해요.
4. 가벼운 검사(`cargo fmt --all`, clippy), 가드 뮤턴트(Result 에 줄로), 임시 폴더에서 실제로 남는 자식을 만든 실행(READY 에 인용), `check affected --card` 한 번, READY 노트.

## Risks

- 슬롯을 쥔 명령이 일부러 남긴 후손은 이제 슬롯 밖에서 돌아요(자원 한도를 넘길 수 있어요). 이 저장소의 호출은 후손을 기다리거나 남기지 않아요. 남기는 호출은 이 SPEC 이전에도 슬롯이 후손이 끝나길 기다리게 하는 부작용(이번 사고)만 만들었어요.
- 래퍼가 SIGKILL 로 죽으면 풀 기회가 없어요. 그 잔여 위험은 REQ-3 의 경고로 눈에 띄게 하고, 회수 명령은 만들지 않아요(Non-goals).
- 낡은 기술자로 줄 서지 않고 슬롯을 다시 쥐는 경쟁(`inherited()` 3·4단계 사이)은 이미 있던 것이고 바뀌지 않아요(사실 5).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | (a) 슬롯을 쥔 쪽이 끝날 때 `flock(LOCK_UN)` 으로 잠금을 풀어요. 후손이 기술자를 쥐고 있어도 슬롯이 비어요. 대안 (b) 기술자를 후손에게 안 넘기기: 중첩 호출이 바깥 슬롯을 다시 쓰는 규칙(REQ-5 of check-affected-one-slot)이 깨져서 조상 확인을 새로 만들어야 해요. (c) 죽은 pid 의 슬롯 파일을 바꿔 치기: 슬롯 파일을 지우는 일이라 리더가 하지 않기로 했어요 | 사용자 지시(2026-11-25 "죽음검사가 또 발생하지 않도록")와 이 일을 맡긴 카드의 "프로세스를 끝내거나 슬롯 파일을 지우는 명령은 만들지 않아요". (a) 가 가장 작고 두 제약을 지켜요 | `Drop` 한 곳을 되돌려요 |
| D-2 | `heavy-slot-status` 는 읽기만 해요. 경고와 `warnings` 만 더하고 슬롯을 회수하거나 프로세스를 끝내지 않아요 | 카드 본문. 끝내는 일은 소유자 몫이에요 | `warnings` 필드를 빼요 |
| D-3 | 가짜 호스트는 부모가 끝나거나 300초 뒤에 끝나요 | 카드 본문 (3) | 루프를 되돌려요 |

## Result

2026-11-25(UTC) macOS, 워크트리 `.claude/worktrees/heavy-slot-release-on-exit`, 코드 head `f427b7ea3d53e429f4ceb1d9e50e8bd2e1cf7557`, base `12dab143c9191080f04b2daef4120d6f8107fc80`(워크트리를 만들 때의 `NeXT` 팁, 따라가지 않았어요).

- AC-1 met: `release_ac1_a_command_that_leaves_a_background_child_gives_the_slot_back`(`heavy_slot_cli`)가 실제 `sleep 600` 자식을 남기고 래퍼가 끝난 뒤 `status` 의 `held` 가 비고 두 번째 호출이 기다리지 않고 도는 것을 읽어요. 손으로도 한 번(`target/debug/moai-cli`, 임시 슬롯 폴더): 자식 살아 있음, `held: []`, 두 번째 호출이 바로 실행됨.
- AC-2 met: `release_ac2_a_nested_caller_leaves_the_outer_slot_held`(중첩 호출이 끝나도 바깥 슬롯이 held)와 `release_ac2_a_wrapper_killed_by_sigkill_keeps_the_slot_while_the_command_runs`(SIGKILL 한계는 그대로).
- AC-3 met: `release_ac3_a_leftover_childs_inherited_descriptor_queues_like_any_caller`가 낡은 fd 로 요청한 후손이 줄을 서고 다른 호출이 끝난 뒤 도는 것을 읽어요.
- AC-4 met: `release_ac4_status_warns_about_a_slot_a_leftover_process_holds`가 `alive: false`, 슬롯·죽은 pid·holders 가 든 `warning` 줄, 최상위 `warnings`, 살아 있는 pid 의 슬롯은 경고 없음을 읽어요.
- AC-5 met: `stub_host_ends_when_its_launcher_is_killed`(`launcher_auto_init`)가 런처를 SIGKILL 로 죽이면 가짜 호스트가 10초 안에 사라지고 release 파일이 쓰이지 않는 것을 읽어요.
- AC-6 met: `release_ac6_the_module_documentation_names_the_release_the_sigkill_case_and_warnings`.
- AC-7 met: 기존 슬롯 시험(`heavy_slot_cli` 38, `heavy_slot_fifo_cli` 10, `heavy_slot_jobs_cleanup_cli` 6, `check_affected_one_slot`)과 `launcher_auto_init` 22개가 통과하고, 오늘의 동작을 못박던 시험은 AC-1·AC-4 로 대체했어요. 약한 단언은 없어요.
- 영수증: `check-affected base=12dab143c919 head=f427b7ea3d53 targets=84 scans=listed:70 test-threads=5 diff_check=0 fmt=0 clippy=0 test=0 result=PASS`(3253개 통과), 묶지 않고 `CLAUDE_CODE_SESSION_ID` 를 풀고 한 번, 로그 폴더 `target/check-affected-f427b7ea3d53/`. 첫 영수증(head `853590dc70`)은 `include_str!` 읽기를 `tools/repo_scans.json` 에 등록하지 않아 FAIL 이었고 항목 하나로 고쳤어요.
- 가드 뮤턴트 5개(각각 한 줄을 고치고 대상 시험 후 복원·`cmp`): `Held::drop` 의 `file.is_some()` 조건을 `false`/`true` 로, `warnings` 를 빈 배열로, `inherited()` 의 `flock` 검사를 `&& false` 로, 가짜 호스트의 부모 확인 줄을 `true` 로. 모두 해당 시험이 exit 101 로 FAIL.
- Gap: 래퍼가 SIGKILL 로 죽으면 잠금을 풀지 못해요(Non-goals). 경고로만 보여요.
- Gap: `check affected`·`audit-prep` 의 `hold` 경로 위에서 남는 자식을 만든 실행은 따로 안 했어요(같은 `Held` 의 `Drop` 이 덮어요).
- Gap: 이미 잡힌 슬롯은 이 수정이 풀지 않아요. 새 바이너리가 설치된 뒤의 래퍼부터 적용돼요.
- Gap: 가짜 호스트 시험의 RED 는 못 읽었어요(호스트 루프를 시험 전에 고쳤어요). 호스트 줄을 되돌린 뮤턴트가 FAIL 하는 것으로 대신 증명했어요.
- Gap: `tools/repo_scans.json` 항목은 줄 임대로 썼고 SPEC 쓰기 목록에 없는 공유 파일이에요.
