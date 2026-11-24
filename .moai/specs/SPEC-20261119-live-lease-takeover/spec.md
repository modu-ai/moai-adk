---
id: SPEC-20261119-live-lease-takeover
title: "a card whose lease ended and whose holder went silent returns to ready, a live lane never does"
status: draft
revision: 2
created: 2026-11-19
---

# a card whose lease ended and whose holder went silent returns to ready, a live lane never does

## Interview

- Q1: 무엇이 바뀌었어요? -> A: revision 1 (2026-11-19) 은 "사용자 결정 노트를 가리키면 살아 있는 lease 도 회수" 를 골랐어요(선택지 A+D1, 사용자 확인 대기였어요). 2026-11-24 사용자 선택(리더 결정 노트 8562)이 방향을 바꿨어요: **하트비트가 끊기고 예약 기한이 지난 카드는 리더 결정 없이 회수해요. 살아 있는 레인은 절대 회수하지 않아요.** 카드 8571 은 이 카드를 병목 우선순위 1 로 올렸어요.
- Q2: 새로 만들 길이 있어요? -> A: 아니요. C-464 가 착지시킨 `moai card sweep` 과 `card take` 안의 sweep(유예 24시간, `SPEC-20261124-ended-lease-sweep`)이 이미 "끝난 lease 의 카드를 `ready` 로" 돌려요. 이 SPEC 은 그 길에 **두 번째 자격**을 하나 더해요: 유예 24시간이 안 지났어도, lease 가 끝났고 보유자의 하트비트가 두 시간 넘게 없으면 같은 길로 돌려요. 새 명령도 새 스키마도 없어요.
- Q3: 하트비트가 뭐예요, 어떻게 자동으로 갱신돼요? -> A: 보유자 세션이 원장에 남긴 가장 최근 흔적이에요: 예약의 `updated_at`(take·`card renew`·범위 확대·line-lease 가 갱신해요)와, 그 카드의 사건 가운데 `actor` 가 보유자 세션인 가장 늦은 사건(`card note`, 기준 표시, 거절 기록 …)의 `at` 중 늦은 쪽이에요. 레인이 일하며 원장에 쓰는 모든 쓰기가 곧 하트비트라서 레인은 따로 부를 명령이 없고, 쓰기 길도 하나 늘지 않아요. 읽기만 하는 `card check` 는 흔적을 남기지 않아요(Risks 2).
- Q4: 살아 있는 레인을 정말 안 건드려요? -> A: lease 가 끝나지 않은 카드는 하트비트가 얼마나 오래됐든 건드리지 않아요(REQ-3). lease 가 끝났어도 하트비트가 두 시간 안에 있으면 건드리지 않아요(REQ-4). 둘 다 어긋난 카드만 돌려요. 일하던 레인이 lease 를 못 늘리고 두 시간 넘게 원장에 아무것도 안 쓰면 회수될 수 있어요: 그때도 가지·워크트리·RESUME 노트는 그대로고, 그 레인의 다음 쓰기는 옛 세대라서 거절돼요(Risks 1).

## Intent

2026-11-19 에 여덟 레인이 사용 한도로 멈췄고, 2026-11-24 에는 38~111시간 전에 lease 가 끝난 카드 19장이 `active` 로 남아 병목 카드 6장의 배차를 막았어요. 지금은 lease 가 끝나고 24시간이 지나야 sweep 이 돌려요. 사용자는 "하트비트가 끊기고 예약 기한이 지났으면 리더 결정 없이 회수" 로 정했어요. 이 SPEC 은 sweep 의 자격에 하트비트 침묵(기본 두 시간)을 더해서, 멈춘 보유자의 카드가 24시간을 기다리지 않고 `card sweep` 이나 다음 레인의 `card take` 로 돌아오게 해요. 살아 있는 lease 와 최근 하트비트가 있는 카드는 그대로 둬요.

읽는 사람: 라운드마다 카드를 정리하는 리더, 멈춘 카드를 이어받는 레인, 이 규칙을 문서에 옮기는 카드.

## Non-goals

- lease 가 끝나지 않은 예약의 회수. 사용자 결정으로 살아 있는 lease 를 넘겨받는 `--user-decision` 길(revision 1 의 REQ-1~6)은 이 revision 에서 채택하지 않아요(Decisions D-1).
- 하트비트를 위한 새 쓰기 명령(`card heartbeat`)과 새 열·표. 하트비트는 원장에 이미 있는 값에서 읽어요(D-2).
- 도구 호출마다 lease 를 자동으로 늘리는 훅. 호스트마다 훅이 달라요(D-3).
- 읽기 전용 `card check` 가 하트비트를 남기게 하기: `src/card/work.rs` 는 이 카드의 범위 밖이에요. 후속 카드가 해요(Risks 2).
- `card next` 의 `stale[]` 행이 침묵한 카드에도 sweep 명령을 말하게 하기: 행의 사실을 읽는 `src/ledger/close.rs`·`src/card/work.rs` 는 범위 밖이라 범위 확대를 리더에게 요청해요(Risks 3).
- 하트비트 창을 프로젝트 설정으로 바꾸기. 두 시간 고정이에요(D-4). 필요하면 `lease_sweep_hours` 처럼 키를 후속으로 더해요.
- `review` 상태 카드, `held_for_review`·`expired_candidate` 예약, 가지·워크트리·RESUME 노트의 삭제, `src/leader_round.rs`(C-469 몫)와 라운드 배선.

## Requirements

- REQ-1: The ledger shall read the heartbeat of a held reservation as the later of the reservation's `updated_at` and the `at` of the latest event of the card whose `actor` is the holder · failure: none — the read has no write path.
- REQ-2 (REQ-1): When `moai card sweep --by leader` runs, the command shall release each `active` card whose current reservation is `held`, whose lease has ended and whose heartbeat is at least the silence window old, move the card to `ready` and write one `card.sweep` event, in one write transaction per card.
- REQ-3 (REQ-2): IF the lease of the reservation has not ended, THEN the sweep shall leave the card unchanged however old the heartbeat is.
- REQ-4 (REQ-2): IF the lease ended less than the grace period ago and the heartbeat is younger than the silence window, THEN the sweep shall leave the card unchanged and list it as skipped with the reason `within_grace`.
- REQ-5 (REQ-2): IF the holder renewed or wrote an event after the command listed the card, THEN the sweep shall leave the card unchanged, because the write transaction reads the heartbeat again.
- REQ-6 (REQ-2): When `moai card take` runs on an `active` card that meets REQ-2, the command shall sweep that card and take it in one write transaction · failure: none — a take that the scope check refuses after the sweep leaves the card `ready`, as the grace sweep does.
- REQ-7: The sweep shall treat a lease that ended at least the grace period ago as sweepable whatever the heartbeat says, and the silence window is two hours.
- REQ-8 (REQ-2): When a card is swept, the `card.sweep` event and the answer shall carry `via` (`heartbeat` when the grace period had not run out, else `grace`), `last_seen_at` and `silent_secs` · failure: none — the fields are written in the sweep's own transaction.
- REQ-9 (REQ-2): IF a swept holder writes with its old generation, THEN the ledger shall refuse the write as it refuses a stale generation today.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-8): in a temporary ledger an `active` card whose lease ended three hours ago and whose holder wrote nothing for three hours is swept to `ready` with its reservation `released`, one `card.sweep` event with `via: heartbeat`, `last_seen_at`, `silent_secs` and a note naming the silence, and a card whose lease ended 25 hours ago keeps `via: grace` · verify: `cargo nextest run --locked --offline --test card_heartbeat_cli silent_`
- AC-2 (REQ-3): a card whose lease has not ended and whose holder wrote nothing for ten hours is skipped by `card sweep`, by `card sweep --dry-run` and by a take, with every row as it was · verify: `cargo nextest run --locked --offline --test card_heartbeat_cli live_lease_`
- AC-3 (REQ-1, REQ-4, REQ-5): a lease ended three hours ago is kept when the holder renewed five minutes ago, when the holder wrote a note five minutes ago and when the holder's reservation was widened five minutes ago, while a note written by another session does not count, and a holder that writes after the listing but before the write transaction keeps the card · verify: `cargo nextest run --locked --offline --test card_heartbeat_cli heartbeat_ listed_then_`
- AC-4 (REQ-7): the silence window is pinned at one second before, at and one second after two hours, and the lease-end second counts as ended · verify: `cargo nextest run --locked --offline --lib heartbeat_window_`
- AC-5 (REQ-6): `card take` on a card that meets REQ-2 sweeps and takes it at the next generation with one `card.sweep` and one `reserve` event, and on a card with a live lease or a fresh heartbeat it refuses as before and writes nothing · verify: `cargo nextest run --locked --offline --test card_heartbeat_cli take_`
- AC-6 (REQ-9): after a heartbeat sweep the old holder's `renew`, `check` and `release` with the old generation are refused · verify: `cargo nextest run --locked --offline --test card_heartbeat_cli swept_holder_`
- AC-7 (REQ-2): a `review` card, a `held_for_review` reservation and an `expired_candidate` reservation with a silent holder are skipped as the grace sweep skips them · verify: `cargo nextest run --locked --offline --test card_heartbeat_cli not_swept_`
- AC-8 (REQ-1…REQ-9): the neighbours · verify: `cargo nextest run --locked --offline --test ledger_sweep`; `cargo nextest run --locked --offline --test card_sweep_cli`; `cargo nextest run --locked --offline --test card_stale_cli`

## Scope

- write: `.moai/specs/SPEC-20261119-live-lease-takeover/**`
- write: `src/ledger/reservation.rs`, `src/ledger/sweep.rs`, `src/card/sweep.rs`, `src/card/stale.rs`, `src/card/early_reclaim.rs`
- write: `tests/card_heartbeat_cli.rs`
- write (shared files, line-lease, only when needed): `src/ops/card.rs`, `src/profile.rs`, `docs/design-review/appendix/operations/card.md`
- read-only: `src/ledger/workflow.rs`, `src/ledger/close.rs`, `src/ledger/schema.rs`, `src/card/work.rs`, `tests/ledger_sweep.rs`, `tests/card_sweep_cli.rs`
- forbidden: `src/leader_round.rs` (C-469), the real home ledger (tests use a temporary HOME only), a schema change, a network call

## Plan

1. Tests first (RED on the unchanged code): `tests/card_heartbeat_cli.rs` for AC-1, AC-2, AC-3, AC-5, AC-6, AC-7, and the window edges as unit tests in `src/ledger/sweep.rs` for AC-4. Every test opens a temporary HOME. A test ages a reservation with SQL (`updated_at`, `lease_until`, an event's `at`) the way `tests/card_sweep_cli.rs` ages a lease.
2. `src/ledger/reservation.rs`: `heartbeat_at` reads REQ-1 from a connection.
3. `src/ledger/sweep.rs`: one `sweepable` rule used by the listing and by `sweep_in_tx`: the lease ended and (the grace ran out or the heartbeat is two hours old). `sweep_in_tx` reads the heartbeat inside its transaction (REQ-5), the event and the note carry the new fields (REQ-8). The `reserve_*sweeping` functions keep their signatures.
4. `src/card/sweep.rs`: the answer carries `via`, `last_seen_at`, `silent_secs`, `silent_minutes`.
5. `src/card/stale.rs` and `src/card/early_reclaim.rs`: read again after steps 1 to 4; they change only if a test shows they must.
6. Guard mutants for each new guard, `moai check affected` once at READY, then the independent audit.

## Risks

- A lane that is alive but silent loses its card: its lease ended and it wrote nothing to the ledger for two hours. The rule has no way to tell it from a stopped session (the `card check` of Risks 2 writes nothing). The loss is small: the branch, the worktree and the RESUME notes stay, the note names them, the lane's next write is refused as a stale generation (REQ-9), and `card renew` before a long run keeps the lease alive as the lane rules already say.
- `card check` is read-only (`src/card/work.rs`, out of scope), so the most frequent thing a working lane does leaves no heartbeat; only a write does (take, renew, scope, line-lease, note, criterion). A follow-up card can make a successful check refresh `updated_at` at most once a minute; this SPEC does not claim it.
- The `stale[]` row of `card next` for a silent card whose grace has not run out still says `withdraw` (`src/card/stale.rs` reads facts from `src/ledger/close.rs`). The sweep command works on that card anyway (`card sweep C-n --by leader`, and a take). Whether the row names it is the open scope request.
- A sweep and a heartbeat in the same second: the write transaction reads the heartbeat again (REQ-5), so a write that commits first wins.
- The next lane may take the card with a new worktree while the old worktree holds commits: the note names the old worktree and branch, as the grace sweep does.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The `--user-decision` takeover of a live lease of revision 1 (REQ-1 to REQ-6, AC-1 to AC-4) is dropped | user choice 2026-11-24 (card note 8562): a silent holder with an ended lease is swept without a leader decision, a live lane never | a follow-up SPEC |
| D-2 | The heartbeat is read from `reservation.updated_at` and the holder's own events, with no new write path and no schema change | the lane's ledger writes are the signal of work; revision 1 D-4 refused a new `card heartbeat` command for the same reason | add `card heartbeat` |
| D-3 | No automatic lease renewal by a hook | revision 1 D-2: hosts differ, and a failed renewal would take a working lane's card | a follow-up card |
| D-4 | The silence window is two hours, fixed | the default lease is four hours, the grace is 24 hours; two hours is long enough for a long test run between ledger writes, and a profile key can follow | a profile key like `lease_sweep_hours` |
| D-5 | The silence rule is an addition to the grace rule, not a replacement: past the grace a card is swept whatever the heartbeat | `SPEC-20261124-ended-lease-sweep` is landed and its tests stay as they are | drop the silence rule |

## Result

Not yet implemented: the SPEC is revised (revision 2) and waits for the leader's approval. The criteria below are `not_run` until the tests exist.

- AC-1: not_run
- AC-2: not_run
- AC-3: not_run
- AC-4: not_run
- AC-5: not_run
- AC-6: not_run
- AC-7: not_run
- AC-8: not_run
