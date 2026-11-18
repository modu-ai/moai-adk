---
id: SPEC-20261119-card-reclaim
title: "moai card reclaim — take over a card whose reservation has ended"
status: draft
revision: 1
created: 2026-11-19
---

# moai card reclaim — take over a card whose reservation has ended

## Intent

원장 카드가 `review` 에 멈추는 일이 있어요: 보유자 세션이 끝났는데 PR 은 이미 합쳐졌고, lease 가 끝나서 `card renew`·`card criterion`·`card done`·`card release`
가 모두 거절돼요(모두의AI 원장 C-1·C-45, 이 저장소 C-62·C-67). 사용자 지시로 원장을 직접 고치거나 강제로 인수하는 길은 막혀 있어서, 카드 위의 겹치는 쓰기 범위
(`apps/bo` 등)를 쓰는 다른 카드 일곱 장이 `take` 를 못 해요. 이 조각은 **끝난 예약만** 되찾는 명령 `moai card reclaim` 하나를 더해요: 옛 세대는 `reclaimed` 로
남고, 이 명령을 부른 세션이 다음 세대의 보유자가 돼서 `card criterion`·`card pr`·`card done`·`card release` 로 카드를 닫을 수 있어요. 한 번의 거래가 감사 기록
(`card.reclaim` 사건)을 남겨요.

## Non-goals

- 살아 있는 예약을 되찾는 일(사용자의 "언제든" 회수, 결정 L7의 앞부분): 이 조각은 끝난 lease만 되찾아요. 살아 있는 lease는 거절돼요.
- 보유자 세션의 프로세스 생존, 미커밋·미push, 외부 작업(push·PR·댓글) 대조와 "회수는 세 번까지" 한도: 후속 조각이에요. 이 조각에서는 회수자가 확인한 것을 `--reason` 에 적어 남겨요.
- 회수자가 리더인지 사용자인지 확인하기: 원장은 세션을 알 뿐 역할을 모르므로 `--by` 는 감사용 이름이에요.
- 스키마 바꾸기(상태 값 `reclaimed` 는 1판에 이미 있어요), 손상 원장 복구(`held_for_review`), 실제 HOME·실제 원장 쓰기.

## Requirements

- REQ-1: When a session runs `card reclaim` for a card whose current reservation is `held` and whose lease has ended, the command shall mark that reservation `reclaimed` and give the calling session a new `held` reservation at the next generation, in one transaction.
- REQ-2 (REQ-1): IF the lease has not ended (its end is later than the ledger clock), THEN the command shall refuse and change nothing.
- REQ-3 (REQ-1): IF the offered generation is not the card's current generation, THEN the command shall refuse and change nothing.
- REQ-4 (REQ-1): IF the card does not exist, is `done`, or has no current `held` reservation, THEN the command shall refuse and change nothing.
- REQ-5 (REQ-1): IF `--by` or `--reason` is empty, longer than its limit or has a line break or control character, or the session has no id, THEN the command shall refuse and change nothing.
- REQ-6: The new reservation shall keep the reclaimed one's write scope, worktree and branch, get a lease of `--lease-secs` seconds (default four hours) from the ledger clock, and leave the card's state, notes, criteria and links as they were.
- REQ-7: The reclaim shall write one `card.reclaim` event that names the previous holder, generation and lease end, the new holder and generation, who reclaimed it and why, in the same transaction · failure: none — a refused call writes no event, so a refusal leaves no half-written record.
- REQ-8: While a newer generation of the card exists, every fence-checked call shall refuse the reclaimed generation.
- REQ-9: While the new reservation is `held`, its holder shall be able to close the card with `card criterion`, `card pr` and `card done`, or give it back with `card release`.

## Acceptance

- AC-1 (REQ-1, REQ-6, REQ-7): 끝난 lease 의 `held` 예약이 되찾아져요: 옛 세대는 `reclaimed`, 새 세대는 호출 세션의 `held`(범위·worktree·branch 그대로, 새 lease), 카드 세대 +1, 카드 상태 그대로, `card.reclaim` 사건 한 줄.
  verify: `cargo test --locked --offline --test ledger_reservation reclaim_hands_an_ended_lease_to_the_reclaimer`
- AC-2 (REQ-2, REQ-3, REQ-4): 살아 있는 lease, 세대 불일치(낮은·한 번도 안 낸), 없는 카드, `done` 카드, 풀린·이미 되찾은 예약, 복구용 `held_for_review` 는 거절되고 원장의 모든 행이 그대로예요; lease 끝 초가 경계예요(끝 초 자체는 되찾음, 한 초 전은 거절).
  verify: `cargo test --locked --offline --test ledger_reservation reclaim_refuses_and_writes_nothing`; `cargo test --locked --offline --test ledger_reservation reclaim_lease_end_second_is_the_edge`
- AC-3 (REQ-1, REQ-6, REQ-9): 모두의AI C-1 모양(끝난 lease 의 `review` 카드, 열린 기준, 합쳐진 PR)을 다른 세션이 되찾고 `criterion`·`pr`·`done` 으로 닫아요; 카드 상태 `review`, 노트·기준 글·링크 그대로.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_lets_the_new_holder_close_a_review_card`
- AC-4 (REQ-2, REQ-3, REQ-4, REQ-5): 명령 줄에서도 같은 거절이 계약 분류(`conflict`·`not_found`·`invalid_input`)로 나가고, 임시 HOME 아래 원장 행이 그대로예요; `--by`·`--reason` 경계 N/N+1.
  verify: `cargo test --locked --offline --test card_reclaim_cli reclaim_refusals_change_nothing`
- AC-5 (REQ-7, REQ-8): `card.reclaim` 사건의 payload 와 옛 세대의 모든 울타리 호출(`renew`·`release`·`handoff`·`criterion`·`check`) 거절.
  verify: `cargo test --locked --offline --test card_reclaim_cli the_reclaim_event_names_everything_and_the_old_generation_is_stale`
- AC-6 (REQ-1…REQ-9): 이웃 검사.
  verify: `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_work_cli`; `cargo test --locked --offline --test ledger_reservation`;
  `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`;
  `cargo test --locked --offline --test git_env_scrub`; `cargo test --locked --offline --test ac_command_lint`; `cargo fmt --check`;
  `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/ledger/reservation.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/card/work.rs`, `src/ops/card.rs`, `tests/ledger_reservation.rs`,
  `tests/card_reclaim_cli.rs`, `tests/card_dispatch_cli.rs`, `tests/card_work_cli.rs`, `docs/design-review/appendix/operations/card.md`,
  `.moai/specs/SPEC-20261119-card-reclaim/spec.md`
- read-only: `src/ledger/workflow.rs`, `src/ledger/schema.rs`, `src/session_current.rs`, `src/contract.rs`, `tests/common/**`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소, 스키마 바꾸기

## Plan

1. 시험 먼저: `reservation::reclaim` 의 시험(임시 HOME 원장, lease 를 SQL 로 끝낸 상태), 그 다음 명령 줄 시험(가짜 `gh` 로 합쳐진 PR).
2. `src/ledger/reservation.rs`: `reclaim` — 한 쓰기 거래 안에서 읽고 검사하고 씀(`reserve`·`handoff` 와 같은 모양); 거절은 아무것도 안 씀.
3. `src/card/dispatch.rs`: `card reclaim` 의 입력 검사와 답; `src/ops/card.rs`: 인자(`--generation`·`--by`·`--reason`·`--lease-secs`); 동작 문서 한 줄.

## Risks

- 끝난 lease 만으로는 보유자가 정말 떠났는지 알 수 없어요: 아직 도는 세션이 있으면 회수 뒤 두 세션이 같은 worktree 를 쓸 수 있어요. 이 조각은 확인을 하지 않고 `--reason` 에 적게만 해요(후속: 프로세스·미커밋·미push 대조, 세 번 한도).
- `--by` 는 입력이라 위조할 수 있어요(다른 명령의 `--by` 와 같은 한계): 사건에 세션 id(새 보유자)가 따로 남아서 누가 불렀는지는 지워지지 않아요.
- 새 보유자가 옛 worktree·branch 를 그대로 물려받아요(`handoff` 와 같이): 그 폴더가 없거나 다른 사람 것이어도 원장은 확인하지 않아요.
- `expired_candidate` 를 만드는 코드가 아직 없어서 이 조각은 `held` 만 되찾아요; 그 상태가 생기면 별도 조각이 규칙을 정해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 되찾는 것은 `held` 이면서 lease 가 끝난 예약뿐; `held_for_review`(복구)·`expired_candidate`(만드는 코드 없음)·`released`·`reclaimed` 는 거절 | 리더 지침 1: 끝난 예약만; 복구는 별도 절차 | 상태 목록 넓히기 |
| D-2 | "끝남" 은 울타리의 `LeaseExpired` 와 같은 시험: `lease_until <= 원장 시계` | 한 가지 낱말 한 가지 판정; 경계 시험이 묶음 | 부등호 바꾸기 |
| D-3 | 새 보유자는 이 명령을 부른 세션(호스트의 세션 id); `--by` 는 감사용 이름 | `take`·`handoff` 와 같은 "보유자는 입력이 아니다" 규칙(#3115 N1) | `--to` 로 받기 |
| D-4 | `--generation G` 를 보고 있던 현재 세대로 요구하고 다르면 거절 | 리더 지침 3 "세대 불일치" 거절; 두 회수자가 겹쳐도 하나만 | 세대 없이 부르기 |
| D-5 | 거절은 원장을 안 바꿈(`fence_reject` 사건도 없음) | 거절 시험이 "행이 그대로"를 묶음이 가장 단순; 시도 기록은 후속 | 거절 사건 추가 |
| D-6 | 새 lease 는 `--lease-secs`(60–86400, 기본 14400) 를 원장 시계에서 | `take` 와 같은 한도 | 옛 lease 끝 물려받기(이미 끝났으므로 불가) |

## Result

- AC-1: met — `reclaim_hands_an_ended_lease_to_the_reclaimer` (RED 먼저: `reservation::reclaim`·`Reclaimed` 없음 → E0425/E0432 컴파일 오류, 명령 줄 시험은 `unrecognized subcommand 'reclaim'`); 옛 세대 `reclaimed`, 새 세대 `held`(범위·worktree·branch 그대로, 원장 시계 + 3600초 안), 카드 `review` 그대로·세대 2, `card.reclaim` 사건 한 줄(actor = `--by`).
- AC-2: met — `reclaim_refuses_and_writes_nothing`: 거절 9가지(살아 있는 lease, 풀린 예약, 예약 없는 카드, 끝난 예약이 남은 `done` 카드, 없는 카드, `held_for_review`, `reclaimed`, 한 번도 안 낸 세대, 세대 0)와 진짜 회수 뒤의 낮은 세대 모두 7개 표의 모든 행이 그대로; `reclaim_lease_end_second_is_the_edge`(끝 초 자체는 되찾음, 30초 남음은 거절, 보유자 자신의 끝난 lease 도 되찾음); `reclaim_input_limits_accept_n_and_refuse_n_plus_one`(`by` 80/81, 이유 500/501, lease 60/59·86400/86401, 줄바꿈·제어 문자·빈 값). 변이 M1–M6 모두 KILLED(M6 은 끝 초 `>=` 변이: 끝 초 시험이 잡음).
- AC-3: met — `reclaim_lets_the_new_holder_close_a_review_card`: 모두의AI C-1 모양(끝난 lease 의 `review`, 기준 셋 중 둘 충족, 열린 PR 링크)에서 옛 보유자의 `renew`(`the lease ended`)·`release`(`is review`)는 막힌 채, 다른 세션이 되찾고(세대 2) 기준 3 충족 → `pr` → `done`; 노트·기준·링크 표 전후 같음; 예약은 `[1 reclaimed, 2 released]`. 활성 카드는 되찾은 뒤 `release` 로 `ready` 가 됨(`reclaim_can_hand_an_active_card_back_to_the_ready_queue`).
- AC-4: met — `reclaim_refusals_change_nothing`: 명령 줄에서 살아 있는 lease·한 번도 안 낸 세대·`#7`·세션 id 없음·`--by`/`--reason` 6가지·`--lease-secs` 59/86401·옵션 빠짐·낮은 세대·`done` 카드가 `conflict`·`not_found`·`invalid_input`·사용법 오류로 나가고 원장의 7개 표가 그대로; 80자 `--by`·500자 이유·60초 lease 는 받아들임. 변이 M13(CLI 이유 검사 끄기) KILLED.
- AC-5: met — `the_reclaim_event_names_everything_and_the_old_generation_is_stale`: 사건 payload 의 `from_holder`·`from_generation`·`from_lease_until`(되찾기 전 값)·`holder`·`generation`·`by`·`reason`·`card_state`; 옛 세대의 `renew`·`release`·`handoff`·`criterion`·`check` 가 모두 `is stale` 의 `conflict`; 새 세대의 `check` 는 허용. 변이 M7(범위 복사 끔)·M8(actor)·M9(옛 lease)·M10(카드 상태)·M11(`released`)·M2–M5 도 KILLED(변이 12개 중 12개; 14개 후보에서 겹치는 둘은 상한 12 때문에 뺌).
- AC-6: met — card_reclaim_cli 5, ledger_reservation 27, card_dispatch_cli 25, card_work_cli 24, card_cli 14, card_spec_cli 11, card_import_cli 9, card_export_cli 17, card_summary_cli 16, card_times_cli 7, help_layout 10, product_cli 79, home_isolation_scan 18, git_env_scrub 3, ac_command_lint 5, hot_file_ops_merge 3, mode_cards_53 9, `--lib` 1434, `--bin moai` 28, clippy 0, `cargo fmt --check` 0, `python3 -m unittest tests.ci_scope_test` OK. 옛 문구 시험 하나(`release` 가 거절한 `review` 카드의 안내)는 새 안내(`card reclaim …`)로 고쳤어요. 전체 `cargo test` 는 로컬에서 돌리지 않고 CI 에서 인용해요(`not_run`).
