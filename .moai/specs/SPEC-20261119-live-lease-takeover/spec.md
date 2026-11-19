---
id: SPEC-20261119-live-lease-takeover
title: "card reclaim of a live lease with a recorded user decision, and a visible holder heartbeat"
status: draft
revision: 1
created: 2026-11-19
---

# card reclaim of a live lease with a recorded user decision, and a visible holder heartbeat

## Intent

2026-11-19 에 여덟 레인이 5시간 사용 한도로 멈췄고, 멈춘 세션이 카드 20장을 2~3시간 lease 로 쥐고 있었어요. 사용자는 "멈춘 세션의 예약은 회수해도 된다" 고 정했지만 `card reclaim` 은 **끝난 lease 만** 되찾아요(SPEC-20261119-card-reclaim 의 Non-goal 첫 줄), 멈춘 세션은 `card release` 도 못 돌려요. 그 결과 한 커밋짜리 급한 수리(C-103)가 lease 가 끝날 때까지 묶음 하나를 막았어요(같은 날 `/clear` 가 보유자 id 를 잃은 C-87 도 같은 모양).

이 SPEC 은 두 가지를 정해요. (1) 사용자 결정이 **카드 노트로 기록돼 있을 때만** 살아 있는 lease 를 되찾는 길: 결정이 없으면 거절은 지금과 똑같고, 있으면 사건이 그 결정을 이름해요. (2) 보유자가 멈췄다는 사실이 `card list` 에서 보이게 하기: 새 쓰기 길 없이 원장에 이미 있는 값으로 `last_seen_at`·`silent_secs` 를 계산해요. 조각 1(이 문서)은 계획과 SPEC 만이에요: 코드는 한 줄도 바꾸지 않아요. 구현은 조각 2(되찾기)와 조각 3(목록 표시)이에요.

## Non-goals

- 원장이 사용자인지 리더인지 가려내기: 원장은 세션 id 만 알아요. `--user-decision` 은 사람이 남긴 기록을 가리키는 값이지 신원 증명이 아니에요(`--by` 와 같은 한계, Risks 첫 줄).
- 기본 lease 를 줄이거나 도구 호출마다 자동 갱신하기: 선택지 B 로 노트에 적었고 이 SPEC 은 채택하지 않아요(Decisions D-2).
- 보유자에게 쓰기를 시키는 `card heartbeat` 명령 더하기: 선택지 D2 로 노트에 적었고 채택하지 않아요(D-4).
- 리더가 멈춘 보유자를 대신해 `card handoff` 를 쓰는 길: 선택지 C 로 노트에 적었고 채택하지 않아요(D-3).
- 스키마 바꾸기, 실제 HOME·실제 원장 쓰기, 세션 프로세스 생존 대조, "회수는 세 번까지" 한도.
- 에이전트 워크플로 문서(`mode-agent.md`·`mode-leader.md`) 고치기: 두 파일은 다른 카드 예약 범위예요. 조각 2 가 필요하면 리더가 범위를 줘요.

## Requirements

- REQ-1: When a session runs `card reclaim` with `--user-decision <ref>` for a card whose current reservation is `held` and whose lease has not ended, the command shall mark that reservation `reclaimed` and give the calling session a new `held` reservation at the next generation, in one transaction.
- REQ-2 (REQ-1): IF `--user-decision` is absent and the lease has not ended, THEN the command shall refuse with the existing `only an ended lease is reclaimed` conflict and change nothing.
- REQ-3 (REQ-1, REQ-4): IF the referenced note does not exist on the card, is not of kind `decision`, was written by the calling session, or was created before the reservation's `acquired_at`, THEN the command shall refuse with an `invalid_input` or `conflict` answer that names which test failed and change nothing.
- REQ-4: When `--user-decision` is given, the command shall check the note whether or not the lease has ended, so a wrong reference never passes silently.
- REQ-5: The `card.reclaim` event shall carry `lease_live`, `user_decision` (the note's ref) and `user_decision_author` in the same transaction as the takeover · failure: none — a refused call writes no event.
- REQ-6 (REQ-1): While a newer generation of the card exists, every fence-checked call shall refuse the taken-over generation, so a stopped holder that wakes cannot renew, release, hand off, criterion or check.
- REQ-7: When `card list` prints a card that has a `held` reservation, it shall add a `reservation` object with `holder`, `generation`, `lease_until`, `lease` (`live` or `ended`), `last_seen_at` and `silent_secs`.
- REQ-8: The value of `last_seen_at` shall be the latest of the reservation's `updated_at` and the `at` of the latest event whose `actor` is the holder.
- REQ-9: The value of `silent_secs` shall be the ledger clock minus `last_seen_at`.
- REQ-10: The `card list` command shall stay read-only: `writes` stays `false`, and no new column, table or event is added.
- REQ-11 (REQ-7): IF a card has no `held` reservation, THEN `card list` shall print the card exactly as before, with no `reservation` key.

## Acceptance

- AC-1 (REQ-1, REQ-5): 살아 있는 lease 의 `held` 예약이 결정 노트를 가리키는 `--user-decision` 으로 되찾아져요: 옛 세대는 `reclaimed`, 새 세대는 호출 세션의 `held`(범위·worktree·branch 그대로, 새 lease), 카드 세대 +1, 카드 상태 그대로, `card.reclaim` 사건에 `lease_live: true`·`user_decision`·`user_decision_author`.
  verify: `cargo test --locked --offline --test ledger_reservation reclaim_live_lease_takes_over_with_a_decision_note`
- AC-2 (REQ-2): `--user-decision` 없이 살아 있는 lease 를 되찾으면 지금과 같은 `conflict`(`only an ended lease is reclaimed`)로 거절되고 원장의 모든 행이 그대로예요.
  verify: `cargo test --locked --offline --test card_reclaim_cli live_lease_without_a_decision_is_refused_as_before`
- AC-3 (REQ-3, REQ-4): 거절 다섯 가지 — 없는 노트, 다른 카드의 노트, `decision` 이 아닌 종류, 호출 세션이 쓴 노트, 예약의 `acquired_at` 보다 먼저 쓴 노트 — 가 살아 있는 lease 에서도 끝난 lease 에서도 원장을 안 바꾸고, 끝 초(`created_at == acquired_at`)는 받아들이고 한 초 전은 거절해요.
  verify: `cargo test --locked --offline --test ledger_reservation reclaim_decision_note_is_checked_and_edges_hold`; `cargo test --locked --offline --test card_reclaim_cli reclaim_decision_refusals_change_nothing`
- AC-4 (REQ-6): 살아 있는 lease 를 되찾은 뒤 옛 세대의 `renew`·`release`·`handoff`·`criterion`·`check` 가 모두 `is stale` 의 `conflict` 로 거절되고, 새 세대의 `check` 는 허용돼요.
  verify: `cargo test --locked --offline --test card_reclaim_cli the_old_generation_is_stale_after_a_live_takeover`
- AC-5 (REQ-7, REQ-8, REQ-9, REQ-10, REQ-11): `card list` 가 `held` 카드에 `reservation` 객체를 붙이고(`silent_secs` 는 `updated_at` 만 있을 때와 보유자의 더 늦은 사건이 있을 때 둘 다), `held` 가 아닌 카드는 옛 모양 그대로이며, 실행 전후 원장 행이 같고 `writes: false` 예요.
  verify: `cargo test --locked --offline --test card_cli list_shows_the_holder_and_how_long_it_has_been_silent`
- AC-6 (REQ-1…REQ-11): 이웃 검사.
  verify: `cargo test --locked --offline --test card_reclaim_cli`; `cargo test --locked --offline --test ledger_reservation`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_cli`; `cargo test --locked --offline --test help_layout`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test spec_cli`; `cargo test --locked --offline --test ac_command_lint`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write(조각 1): `.moai/specs/SPEC-20261119-live-lease-takeover/spec.md`
- write(조각 2·3, 이 조각에서는 쓰지 않아요): `src/ledger/reservation.rs`, `src/card/dispatch.rs`, `src/card/command.rs`, `src/ops/card.rs`, `tests/ledger_reservation.rs`, `tests/card_reclaim_cli.rs`, `tests/card_cli.rs`, `docs/design-review/appendix/operations/card.md`
- read-only: `src/ledger/schema.rs`, `src/ledger/workflow.rs`, `src/session_current.rs`, `src/contract.rs`, `tests/common/**`
- forbidden: 실제 HOME과 원장, 네트워크, 실제 원격 저장소, 스키마 바꾸기, 다른 카드의 예약 범위(`skills/moai/workflows/**` 포함)

## Plan

1. (조각 1, 이 문서) 선택지와 권장안을 카드 노트로 남기고 이 SPEC 을 `spec check` 에 통과시켜 push 만 해요. PR 은 열지 않아요.
2. (조각 2) 시험 먼저: `reservation::reclaim` 이 `user_decision: Option<&str>` 을 받는 시험(임시 HOME 원장, lease 를 SQL 로 살려둔 상태, 노트 5가지 거절과 경계). 그 다음 `reclaim` 한 쓰기 거래 안에서 노트를 읽고 검사하고(`card_note` 행: `kind`, `author`, `created_at`, `ref`) 쓰도록 바꾸고, `LeaseLive` 거절은 결정이 없을 때만 남겨요. 명령 줄 `--user-decision` 과 사건 payload, 동작 문서 한 줄.
3. (조각 3) `card list` 가 `held` 예약만 `LEFT JOIN` 으로 읽고(읽기 전용 연결 그대로), `last_seen_at` 은 `event` 표의 `actor` 로 계산해요. 보유자 id 로 쓰는 사건이 실제로 있는지는 조각 3 의 첫 시험(RED)이 확인해요.

## Risks

- `--user-decision` 이 사용자의 결정인지 원장은 몰라요: 결정 노트의 `author` 는 세션 이름이라 두 세션이 짜면 서로의 노트를 결정으로 내세울 수 있어요. 막는 것은 "자기가 쓴 노트로는 자기를 허가하지 못함"(REQ-3)까지고, 나머지는 리더 검토와 사건 기록(`user_decision_author`)이 남아 있다는 것뿐이에요. 사용자 결정은 `guarded` 범주라서 사용자의 말이 노트에 적혀 있는 것을 사람이 확인하는 절차는 그대로 필요해요.
- 노트의 `ref` 는 `card note` 가 찍는 값이에요: 노트에 `ref` 를 안 붙인 결정은 가리킬 수 없어요. 그런 결정은 새 `decision` 노트로 한 번 더 적어야 해요.
- 되찾은 뒤 옛 보유자가 아직 도는 세션이면 두 세션이 같은 worktree 를 쓸 수 있어요(SPEC-20261119-card-reclaim 과 같은 한계): `--reason` 에 확인한 것을 적게 하고, 옛 세대는 모든 쓰기가 거절돼요(REQ-6).
- `last_seen_at` 은 활동의 근사예요: 보유자가 `card note`·`card check` 를 부를 때 사건이 생기는지는 조각 3 시험으로 확인해요. 사건이 없으면 `silent_secs` 는 `updated_at`(take·renew) 이후 시간으로만 계산돼서 4시간 lease 에서는 거칠어요.
- 이 SPEC 은 선택지 B(짧은 lease+자동 갱신)를 채택하지 않아요. 훅이 없는 호스트에서는 갱신이 안 되고, 긴 턴 하나가 lease 를 넘기면 일하는 세션이 빼앗길 수 있어요(Decisions D-2).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 살아 있는 lease 를 되찾는 길은 `card reclaim --user-decision <note ref>` 하나; 결정 노트는 같은 카드의 `decision` 종류, 호출 세션이 쓴 것이 아님, 예약 `acquired_at` 이후에 씀 | 카드 기준 1: 결정 없으면 거절은 그대로; 사건이 결정을 이름; 사용자 결정 2026-11-19 | 플래그를 빼면 이전 동작 |
| D-2 | 선택지 B(기본 lease 단축+자동 갱신)는 이번에 채택하지 않음 | 호스트마다 훅이 달라 갱신이 불확실; 갱신 실패가 바로 빼앗김; 쓰기 늘어남(원장 tier 2) | 별도 카드로 다시 평가 |
| D-3 | 선택지 C(리더가 멈춘 보유자 대신 `handoff`)는 이번에 채택하지 않음 | "멈춘 것으로 보임" 판정이 원장 밖이라 같은 신뢰 문제를 `handoff` 의 보유자 규칙(#3115 N1)에 풀어 놓음 | — |
| D-4 | 보유자 heartbeat 는 새 쓰기 명령이 아니라 `card list` 의 읽기 전용 계산(`last_seen_at`) | 새 쓰기 길 0, 스키마 변경 0(카드 기준 1 은 "보이는 것"만 요구) | `card heartbeat` 추가 |
| D-5 | 결정 노트는 `--user-decision` 이 가리키는 카드 안에서만 찾음(다른 카드의 노트는 거절) | 한 결정이 어느 카드에 적용되는지 노트의 위치가 곧 범위 | 카드 간 참조 허용 |

## Result

조각 1(계획과 SPEC)만 해당해요 — 구현 조각은 아직 시작 전이에요.

- AC-1: not_run — 조각 2 의 시험이에요.
- AC-2: not_run — 조각 2 의 시험이에요.
- AC-3: not_run — 조각 2 의 시험이에요.
- AC-4: not_run — 조각 2 의 시험이에요.
- AC-5: not_run — 조각 3 의 시험이에요.
- AC-6: not_run — 조각 2·3 구현 뒤에 돌려요.
