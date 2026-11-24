---
id: SPEC-20261124-ended-lease-sweep
title: "a lease that ended long ago returns its card to ready without a human hand"
status: draft
revision: 1
created: 2026-11-24
---

# a lease that ended long ago returns its card to ready without a human hand

## Interview

- Q1: 무엇이 재발했어요? -> A: 2026-11-24 에 리더가 손으로 두 번 정리했어요. 한 번은 리스가 며칠 전에 끝난 C-216·C-253 이 공유 파일을 막았을 때, 한 번은 active 27장 가운데 19장이 38~111시간 전에 리스가 끝난 카드였을 때예요. 리더는 카드마다 `card reclaim` 뒤 `card release` 를 불렀어요 (카드 C-464 본문).
- Q2: 끝난 리스는 이미 경로를 내놓지 않아요? -> A: 내놓아요. `SPEC-20261121-card-lease-yield` 가 끝난 `held` 예약의 경로를 take·scope 확대·`next` 에서 양보하게 했어요(`src/ledger/reservation.rs` 의 `blocks_new_claims`, `yield_overlaps`). 그래서 이 SPEC 은 경로가 아니라 카드 상태를 다뤄요: 예약이 끝난 카드는 `active` 로 남고, 그 수가 상태줄 진행 수를 부풀려요(C-452). `card next` 의 `stale[]` 는 이 카드를 `lease_ended` 행으로 보여 주면서 `card withdraw` 를 권하는데(`src/card/stale.rs`), withdraw 는 카드를 `hold` 로 보내서 일이 사라져요. 경로가 실제로 막힌 원인(검토 중 상태, `held_for_review`, 오래된 바이너리)은 읽기만으로 정하지 않았고, 구현 첫 단계에서 임시 원장으로 재현해 확인해요(Plan 1).
- Q3: reclaim 뒤 release 를 부르는 길을 그대로 자동화하면 안 돼요? -> A: 안 돼요. `card reclaim` 은 새 holder 와 새 리스를 만들고, 카드마다 세 번만 허용하며(`reclaim >= 3` 거절), 기록된 워크트리를 검사해요. 좀비 카드마다 그 허용량을 쓰고 낭비해요. 새 동작은 reclaim 없이 예약만 `released` 로 끝내요(D-2).
- Q4: 일하던 레인의 흔적은요? -> A: 가지·워크트리·RESUME 노트는 건드리지 않아요. 노트 한 줄이 옛 holder·워크트리·가지를 적어서, 다음에 가져가는 레인이 노트를 먼저 읽어요(리더 결정 노트 8468 과 같은 문장).

## Intent

리스가 끝난 지 오래된 예약이 카드를 `active` 로 붙잡고 있으면, 상태줄 진행 수가 부풀고 일이 아무에게도 배차되지 않아요. 지금은 리더가 카드마다 reclaim 과 release 를 손으로 불러 풀어요. 이 SPEC 은 프로젝트 설정으로 정한 시간(유예)이 리스 끝 뒤로 지난 `active` 카드를 한 명령이 한 번에 `ready` 로 돌리고, 카드마다 노트를 남기게 해요. 새 스키마와 새 카드 상태는 없어요.

읽는 사람: 라운드마다 카드를 정리하는 리더, `card next` 를 읽는 레인, 이 규칙을 문서에 옮기는 카드.

## Non-goals

- 일하는 레인의 카드를 빼앗는 일. 유예 안에 리스가 갱신되면 건드리지 않아요 (살아 있는 리스의 회수는 `SPEC-20261120-reclaim-dead-holder` 와 C-117 의 몫이에요).
- `review` 상태 카드. 그 카드를 닫을 사람이 아무도 없게 되므로 `stale[]` 에 남겨 둬요 (`SPEC-20261124-card-fold-stale`).
- 가지·워크트리·RESUME 노트의 삭제나 이동. 워크트리는 리더의 `worktree-gc-plan` 몫이에요.
- `held_for_review` 와 `expired_candidate` 예약. 누구 것인지 확인되지 않아서 계속 막아요 (`SPEC-20261121-card-lease-yield` D-2).
- 라운드에 이 명령을 넣는 배선. `src/leader_round.rs` 는 다른 카드가 쥐고 있어서 후속 카드가 해요.
- 상태줄 숫자 계산. 카드 상태가 바뀌면 숫자가 바로잡혀요 (C-452 의 파일은 건드리지 않아요).

## Requirements

- REQ-1: When the leader runs `moai card sweep --by leader`, the command shall release each `active` card whose current reservation is `held` and whose lease ended at least the grace period ago, move the card to `ready` and write one `card.sweep` event, in one write transaction per card.
- REQ-2 (REQ-1): IF the lease ended less than the grace period ago, the card is in `review`, the reservation is not `held`, or the lease was renewed after the command read it, THEN the command shall leave the card unchanged and list it as skipped with the reason.
- REQ-3: When a card is swept, the command shall add one note to the card that names the old holder, generation, lease end, worktree and branch and tells the next lane to read the notes first. · failure: none — the note is written in the same transaction as the sweep, so a skipped card gets no note.
- REQ-4: When a card is swept, the ledger shall keep the card's generation, branch, worktree, notes and criteria as they were. · failure: none — the sweep only ends the reservation and moves the state, it writes no other column.
- REQ-5: When a card is swept, the ledger shall not count the sweep against the three-reclaim allowance of the card. · failure: none — the event kind is `card.sweep`, and the allowance counts `card.reclaim` events only.
- REQ-6: While the project profile sets `lease_sweep_hours`, the grace period shall be that many hours; without it, 24 hours.
- REQ-7 (REQ-6): IF `lease_sweep_hours` is 0, THEN the profile check shall refuse the profile as the other grace keys are refused.
- REQ-8: When `moai card sweep --dry-run` runs, the command shall list the cards it would sweep and the cards it would skip and write nothing. · failure: none — a dry run opens the ledger read-only.
- REQ-9 (REQ-1): IF `moai card sweep` runs without `--by leader`, THEN the command shall refuse and write nothing.
- REQ-10 (REQ-1): IF a swept holder writes with its old generation, THEN the ledger shall refuse the write as it refuses a stale generation today.
- REQ-11: When `moai card next` lists the `lease_ended` row of a card past the grace period, the row's command shall be `moai card sweep C-n --by leader`. · failure: none — the row is read-only text, and a card inside the grace period keeps the row it has.
- REQ-12: When `moai card take` is run on an `active` card whose lease ended at least the grace period ago, the command shall sweep that card and take it in one write transaction.
- REQ-13 (REQ-12): IF the lease ended less than the grace period ago, THEN `moai card take` shall refuse as before and write nothing.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary ledger, an `active` card with a lease ended 25 hours ago is swept to `ready` with its reservation `released`, and one ended 23 hours ago, one in `review`, one `held_for_review` and one renewed between the read and the write are skipped with their reasons and every row as it was · verify: `cargo nextest run --locked --offline --test ledger_sweep sweep_`
- AC-2 (REQ-3, REQ-4, REQ-5): the note names the five facts, the card's generation, branch, worktree, other notes and criteria are unchanged, and a card with three earlier `card.reclaim` events is still swept and can still be reclaimed once more by the rules of that command · verify: `cargo nextest run --locked --offline --test ledger_sweep kept_`
- AC-3 (REQ-6, REQ-7): a profile with `lease_sweep_hours: 48` sweeps at 49 hours and skips at 47, no key means 24, and 0 is refused · verify: `cargo nextest run --locked --offline --test ledger_sweep grace_`
- AC-4 (REQ-8, REQ-9): `--dry-run` leaves every ledger row as it was and a call without `--by leader` is refused with nothing written · verify: `cargo nextest run --locked --offline --test card_sweep_cli dry_run_ refuses_`
- AC-5 (REQ-10): `card renew`, `card check` and `card release` with the swept generation are refused as stale or not held, and a second lane takes the card at the next generation · verify: `cargo nextest run --locked --offline --test card_sweep_cli swept_holder_`
- AC-6 (REQ-11): the `lease_ended` row of a past-grace card carries the sweep command and the row of one inside the grace period keeps its row as `SPEC-20261124-card-fold-stale` pinned it · verify: `cargo nextest run --locked --offline --test card_sweep_cli stale_row_`
- AC-7 (REQ-12): `card take` on a past-grace `active` card yields a `ready`→`active` card at the next generation with one `card.sweep` and one `reserve` event, and on one inside the grace period it refuses as before · verify: `cargo nextest run --locked --offline --test card_sweep_cli take_`
- AC-8 (REQ-1): a line lease and a scope take on a path held by a card whose lease ended three days ago succeed, with the old reservation losing that path (the `SPEC-20261121-card-lease-yield` rule, pinned here for the past-grace case) · verify: `cargo nextest run --locked --offline --test card_sweep_cli ended_lease_paths_`

## Scope

- write: `.moai/specs/SPEC-20261124-ended-lease-sweep/spec.md`
- write: `src/ledger/sweep.rs`, `src/ledger/mod.rs`, `src/ledger/reservation.rs`, `src/card/sweep.rs`, `src/card/mod.rs`, `src/card/stale.rs`, `src/card/work.rs`, `src/card/dispatch.rs`
- write: `tests/ledger_sweep.rs`, `tests/card_sweep_cli.rs`, `tests/card_stale_cli.rs`
- write: `src/ops/card.rs`, `src/profile.rs`, `tests/areas.toml`, `docs/design-review/appendix/operations/card.md` (shared files of the project profile: written with the line-lease command, not held in the card scope)
- read-only: `src/ledger/workflow.rs`, `src/ledger/close.rs`, `src/card/take_scope.rs`
- forbidden: `src/leader_round.rs`, `src/statusline/state.rs` (held by other cards), the real home ledger (tests use a temporary HOME only)

## Plan

1. Reproduce first (RED, temporary ledger): a take and a line lease on a path held by a card whose lease ended three days ago, for a `held`, a `review` and a `held_for_review` reservation. A case that fails is an in-scope defect and gets its own test (AC-8); the reading of the code says `held` and `review` pass.
2. Tests first for AC-1 to AC-7 and see them fail on the unchanged code.
3. `src/ledger/sweep.rs`: one function that, inside one write transaction per card, re-reads the reservation and the DB clock, checks the grace, sets the reservation `released`, the card `ready`, writes `card.sweep` and the note.
4. `moai card sweep [--dry-run] --by leader` in the dispatch, the `lease_sweep_hours` profile key and its check, the `stale[]` command change, and the take-time sweep: `reservation::reserve` and `reserve_own_scope` keep their signatures and call new variants that carry the grace period and run the sweep inside the take transaction; the hold and criteria gates of `take` read a `ready` card, and a card that was `active` passed them when it was first taken.
5. Mutants for each guard, whole-area test run through `moai check affected`, then the independent audit.

## Risks

- A holder whose session paused for more than the grace period loses its card; its work stays in its worktree and branch, the note says where, and its next write is refused as stale (REQ-10). The default of 24 hours is long for a 4-hour lease and the profile key shortens or lengthens it.
- A sweep and a renew at the same second: the sweep re-reads the lease inside its own write transaction (REQ-2), so a renewal that commits first wins.
- The next lane may take the card with a new worktree while the old worktree still holds commits. The note names the old worktree and branch; the existing takeover gate of `card reclaim` is not part of a plain take, and this SPEC does not add one (Non-goals).
- A test that would touch the real ledger: every test opens a temporary HOME as `tests/ledger_lease_yield.rs` does.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The grace period is the profile key `lease_sweep_hours`, default 24 | the card says "정한 시간(프로젝트 설정)"; the zombies of 2026-11-24 were 38 to 111 hours past their lease; the pattern follows `stale_draft_days` | change the key or the default in a follow-up SPEC |
| D-2 | A sweep ends the reservation as `released` without a reclaim: no new holder, no new lease, no reclaim allowance used, no worktree gate | Q3; a reclaim would use one of the three lifetime reclaims per zombie | a follow-up SPEC can switch to reclaim, no data depends on it |
| D-3 | `--by leader` is required, as for `card withdraw` and `card land` | `SPEC-20261124-card-fold-stale` Q4 | drop the flag check |
| D-4 | A past-grace card is also taken directly by `card take`, in the same transaction that sweeps it (REQ-12, REQ-13); `card next` stays read-only, names the sweep command (REQ-11) and does not list the card as a ready candidate | leader decision, note 8505 (2026-11-24, option 1 of 3: option 2 leaves lanes waiting for a leader round, option 3 hides writes inside commands that did not say so) | drop REQ-12 and REQ-13 |
| D-5 | The round wiring is a follow-up card after the card that holds `src/leader_round.rs` lands | forbidden list of Scope | none needed |

## Result

Not started: the leader approved revision 1 with D-4 = option 1 (note 8505); the scope list and the wording of D-4 changed after that approval, so the new digest goes to the card for the leader to record.
