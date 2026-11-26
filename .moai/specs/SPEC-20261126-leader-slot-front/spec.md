---
id: SPEC-20261126-leader-slot-front
title: "heavy slot: 확인된 리더의 넓은 검사는 대기 줄 맨 앞에 서요"
status: draft
revision: 1
created: 2026-11-26
---

# heavy slot: 확인된 리더의 넓은 검사는 대기 줄 맨 앞에 서요

## Intent

착지 대기가 단계별 시간에서 가장 긴 단계예요. 2026-11-26 01:46Z 에는 카드 4장의 착지를 여는 리더의 넓은 검사가 heavy slot 대기 12개 중 11번째로 레인 영수증 뒤에 섰어요. 등록된 리더 세션이 돌리는 `--full` 검사(`card land-batch` 가 안에서 부르는 검사 포함)가 대기 줄 맨 앞에 서게 해요. 순서만 바꾸고, 자리 수·한계·거절·실행 중인 작업은 그대로예요.

`SPEC-20261122-check-affected-one-slot` 의 REQ-4(도착 순서)에 이 한 경우의 예외를 두는 변경이에요. slot 순서는 착지 관문에 닿아서 tier 2 로 보고, 승인 전에는 코드를 쓰지 않아요. 구현은 CPU 여유 규칙(`SPEC-20261126-heavy-slot-cpu-free`, `hold` 의 대기 반복을 바꿈)이 착지한 뒤 그 위에서 해요(리더 결정 2026-11-26).

## Interview

- Q1: 무엇으로 리더를 확인하나요? → A: (기본값, 승인에서 확정) `card land-batch` 가 쓰는 확인(`confirm_leader`)이에요. 이 호출의 세션 id 가 살아 있는 `leader` 항목(mode leader)의 `session` 과 같아야 해요. 카드 문구의 "인스턴스"는 이 뜻으로 읽어요(D-2).
- Q2: `card land-sync` 의 넓은 검사도 앞에 서나요? → A: (기본값) 예. 같은 `--full` 길이에요(D-3).
- Q3: 리더가 손으로 돌리는 `--full` 은요? → A: (기본값) 세션 id 가 있으면 앞에 서요. 세션 id 를 지우고 돌리면 도착 순서예요. 지금 습관대로 지우고 돌리면 실제로 앞에 서는 것은 `card land-batch`·`land-sync` 뿐이에요(D-4).

## Non-goals

- 자리 수, `limits.json` 의 한계, 디스크·메모리·CPU 규칙, 거절 코드, 실행 중인 작업은 바꾸지 않아요.
- 리더 등록과 `confirm_leader` 의 판단은 바꾸지 않아요(알려진 한계는 Risks).
- `heavy-slot-acquire`, `check fast`, `audit-prep` 의 순서는 바꾸지 않아요.
- 앞자리 상한, 나이 먹임(aging), `heavy-slot-status` 의 새 필드는 만들지 않아요.

## Requirements

- REQ-1: When a `check affected` run with `--full` is made by the session that the live registry entry `leader` of mode leader names, the run shall take a place in the heavy slot waiting queue ahead of every waiter that is not such a run.
- REQ-2 (REQ-1): While more than one such run waits, the waiting queue shall order those runs by their arrival.
- REQ-3 (REQ-1): The waiting file of such a run shall keep the three-part name `<20 digits>-<pid>-<max>` with the arrival time in microseconds, so that a build without this change orders it ahead of every nanosecond name.
- REQ-4 (REQ-1): IF the caller of a `--full` run cannot be confirmed as that session (no session id, no live leader entry, another session id, or an unreadable registry), THEN `check affected` shall queue the run in arrival order without refusing it.
- REQ-5 (REQ-1): Every other heavy slot caller shall queue in arrival order whatever session name, `--name` value, flag or environment variable it supplies other than the session id that REQ-1 confirms.
- REQ-6 (REQ-1): The front place shall change only the order of waiters, leaving the slot count, the `limits.json` limits, the machine limit refusals, the CPU headroom rule and every held slot as they are.
- REQ-7 (REQ-1): When a `--full` run enters the queue, `check affected` shall write one line to `heavy-slot.log` that says whether the run took the front place or arrival order, with the reason for arrival order · failure: none — a log file that cannot be opened keeps the current behaviour (the lines are dropped) and the order does not depend on it.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): with a temporary slot folder of one slot, the registered leader session's `--full` run that arrives after two waiting lane runs starts before both, two leader runs keep their arrival order ahead of the lanes, and a run in a linked worktree reads its repository's leader entry · verify: `cargo test --locked --offline --test heavy_slot_front_cli front_ac1_`
- AC-2 (REQ-6): the running command finishes untouched before the leader run starts, `--max 4` does not raise a file `max` of 1 and `limits.json` keeps its bytes, a busy machine answers the leader run with exit 75, and a low CPU reading holds the leader run too · verify: `cargo test --locked --offline --test heavy_slot_front_cli front_ac2_`
- AC-3 (REQ-4, REQ-5): another session, `heavy-slot-acquire --name leader`, an entry without leader mode or past its lease, a call without a session id and a leader run without `--full` all queue in arrival order without a refusal, and a second `leader` registration over a live entry is refused · verify: `cargo test --locked --offline --test heavy_slot_front_cli front_ac3_`
- AC-4 (REQ-7): `heavy-slot.log` of a `--full` run names the front place, or arrival order with its reason · verify: `cargo test --locked --offline --test heavy_slot_front_cli front_ac4_`
- AC-5 (REQ-5, REQ-6): the existing arrival order tests are unchanged and pass · verify: `cargo test --locked --offline --test heavy_slot_fifo_cli`
- AC-6: light checks · verify: `git diff --check`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`; `moai spec check SPEC-20261126-leader-slot-front`; `moai trailer-check NeXT..HEAD`

## Scope

- write: `src/heavy_slot.rs`, `src/check_affected.rs`, `tests/heavy_slot_front_cli.rs`(새 파일), `docs/design-review/appendix/operations/check.md`, `docs/rust-product-path.md`, `.moai/specs/SPEC-20261126-leader-slot-front/**`, `.moai/specs/SPEC-20261122-check-affected-one-slot/spec.md`(Result 에 REQ-4 예외 한 줄만)
- write(필요할 때만): `tests/fixtures/state_folder_literals/allow.tsv`(새 테스트가 상태 폴더 이름을 글자로 쓰면)
- read-only: `src/card/land/batch.rs`, `src/card/land/sync.rs`, `src/card/dispatch.rs`, `src/lane_registry.rs`, `src/ops/check.rs`
- forbidden: 기계의 실제 slot 폴더와 `limits.json`, 실제 프로젝트의 레지스트리 파일

## Plan

1. 승인과 CPU 여유 규칙의 착지를 기다려요. 레인은 옮긴 팁을 합치지 않으므로, 그 착지가 든 통합 팁에서 구현 브랜치를 새로 만들고 SPEC 커밋만 옮겨요(리더 확인). `hold` 를 다시 읽고 쓰기 범위를 넓혀요. `src/check_affected.rs` 는 다른 카드의 점유를 확인해요.
2. RED: `tests/heavy_slot_front_cli.rs`(`#![cfg(unix)]`). 임시 HOME·임시 git 저장소·임시 slot 폴더, `limits.json` 은 `max` 1. 자식 프로세스에서는 `CLAUDE_CODE_SESSION_ID`, `CODEX_THREAD_ID`, `MOAI_HEAVY_SLOT_FD`, `XDG_CACHE_HOME`, `CARGO_TARGET_DIR` 을 지우고 경우마다 필요한 것만 넣어요. 리더 등록은 `session-register` 로만 해요. 앞 대기 파일이 생긴 것을 본 뒤에 다음 대기를 시작해요.
3. `src/heavy_slot.rs`: 순수 함수 `waiting_stamp(front, ns)`(앞자리면 `ns / 1_000`)와 `hold_ordered(.., front)`. `hold` 의 모양은 그대로이고 `false` 를 넘겨요. non-unix 쪽 함수와 단위 테스트, 도착 순서만 약속하는 설명을 맞춰요.
4. `src/check_affected.rs`: `hold_slot` 이 저장소 폴더를 받고, `--full` 일 때만 세션 id 를 읽어 `confirm_leader` 로 확인한 뒤 `hold_ordered` 를 불러요. 로그 한 줄은 `heavy-slot: slot ` 로 시작하지 않고, `confirm_leader` 의 거절 문구를 그대로 옮기지 않아요.
5. 문서 두 곳의 순서 설명을 고치고, 옛 SPEC 의 Result 에 예외 한 줄을 적어요.
6. 검증: AC 의 명령, `heavy_slot_cli`, `check_affected_one_slot`, `card` 바이너리의 `card_land_batch_cli::`, 문서를 읽는 테스트. 가드 뮤턴트(`/ 1_000` 제거 → AC-1 실패, 확인을 항상 참으로 → AC-3 실패, 새 가드 테스트마다 한 줄). 위험한 길: 버리는 저장소와 slot 폴더에서 레인의 흉내를 실제로 실행해요. `check affected --card` 뒤 READY.

## Risks

- **세션 id 를 믿어요(이 SPEC 이 막지 않음).** 확인은 환경 변수의 세션 id(`CLAUDE_CODE_SESSION_ID`, 없으면 `CODEX_THREAD_ID`)를 항목의 `session` 과 비교해요. `moai session-register --show` 가 그 값을 알려 주므로, 값을 베껴 넣은 레인은 앞에 설 수 있어요. `instance` 도 같은 명령이 보여 줘서 비밀이 아니에요. 이 변경이 막는 것은 **이름·역할 말만 흉내 내는 것**이에요. 등록 해제·`--continue` 길로 리더 항목을 바꿀 수 있어 보인다는 코드 읽기 보고도 있어요(실행 안 함, 미확인).
- **손으로 만든 대기 파일:** 더 낮은 이름의 잠긴 파일은 지금도 줄을 앞지를 수 있어요.
- **저장소와 기계:** 레지스트리는 클론마다, slot 은 기계에 하나예요. 어느 클론의 확인된 리더든 다른 프로젝트의 레인 앞에 서요.
- **굶김:** 리더가 연달아 돌리면 그때마다 레인보다 먼저예요. 동시에 도는 리더 실행은 자리 수만큼 자리를 다 가질 수 있어요.
- **CPU 규칙과의 만남(그 변경의 초안 `081b34def7` 읽기, 실행 안 함):** 빈 자리가 둘 이상이면 CPU 대기가 오래 쌓인 레인이 상한에 닿아 먼저 시작할 수 있어요.
- **경합:** 줄을 이미 읽은 레인은 그 한 번의 보기에서 자리를 잡을 수 있어요. 자리 수와 배타성은 그대로예요.
- **섞인 빌드:** 옛 빌드의 레인도 앞자리 이름을 앞으로 정렬하지만, 옛 빌드의 리더는 앞에 서지 못해요. 착지 뒤 설치가 필요해요.
- **보이는 것:** `heavy-slot-status` 는 살아 있는 대기를 나열하지 않고, 레인은 자기 순번이 뒤로 가는 것을 볼 수 있어요.
- **감싼 실행:** 이미 자리를 물려받은 실행은 줄을 서지 않아서 앞자리가 아무 일도 하지 않아요.
- **한 번 확인:** lease 는 줄에 들어갈 때 한 번 확인해요. 리더의 하위 에이전트와 백그라운드 실행은 같은 세션 id 를 써요.
- non-unix 타입 확인과 Windows·Linux 실행은 하지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The front place is a microsecond stamp inside the same three-part waiting name; no fourth part and no second folder. | At the base and at the draft `081b34def7`, the name is parsed only by `waiter_pid` (digits check, reads part two) and `waiter_max` (drops part one), and `now_ns()` is used only to build the name (`git grep` over `src/heavy_slot.rs` and the heavy slot tests); CPU waiting uses `Instant`. Older builds sort by string, so they honour it. A fourth part or a second folder would be ignored by older builds and cleanup. Must be read again after the CPU rule lands. | Use plain `now_ns()` again. |
| D-2 | "Confirmed leader instance" means the existing `confirm_leader` session check. | It is the gate `card land-batch` itself passes; `instance` is not passed by any check command and is documented as uniqueness, not secrecy. | Require a further proof in a later card. |
| D-3 | `card land-sync` gets the front place too. | It runs the same wide check through the same call with `full` set, and it moves the integration branch. | Exclude it in `batch.rs` and `sync.rs`. |
| D-4 | Any `--full` run of the confirmed session counts, with `--card` or `--lane` too; a bare run does not. | The card names `check affected --full`. | Count only the full selection. |
| D-5 | No cap on front runs, no status field, the CPU waiting count unchanged, the per-clone registry recorded as a limit, and no end-to-end `land-batch` queue test (the linked-worktree case and the existing test of the full option cover the call). | Smallest change that meets the card. | Add each in a later card. |

## Result

(아직 없음 — 승인 전 초안)
