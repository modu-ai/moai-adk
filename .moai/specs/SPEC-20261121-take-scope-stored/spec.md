---
id: SPEC-20261121-take-scope-stored
title: "card take keeps the scope it was given on the card and fills an empty scope from the SPEC or the last reservation"
status: draft
revision: 1
created: 2026-11-21
---

# card take keeps the scope it was given on the card and fills an empty scope from the SPEC or the last reservation

## Intent

`moai card take C-n --scope P…` 로 준 scope 는 예약에만 저장되고 카드 자체의 `write_scope` 에는 남지 않아요. 그래서 한 번 놓은(release) 카드는 scope 가 비어 있고, `card next` 는 그 카드를 `empty_scope` 로 건너뛰고, 다음 사람은 `card scope` 를 먼저 쳐야 해요. 2026-11-21 조사에서 ready 카드 76장 중 69장이 이 때문에 막혀 있었어요(리더 조사 보고서 `card-flows-2026-11-21.md`). 또 거절 문구 `the write scope is empty; name at least one path` 는 `--scope` 도 `card scope` 도 알려 주지 않아요.

이 SPEC 은 세 가지를 고쳐요. 첫째, `take --scope` 가 카드의 scope 가 비어 있을 때 그 scope 를 카드에도 남겨요. 둘째, `--scope` 없이 take 했는데 카드 scope 가 비어 있으면 연결된 SPEC 의 `## Scope` `write:` 목록으로, 그것도 없으면 이 카드의 가장 최근 예약 scope 로 채워요. 셋째, 아무것도 없으면 어떻게 하라고 말해 주고 거절해요. `card next` 도 같은 순서로 판단해서 채울 수 있는 카드를 건너뛰지 않아요.

## Non-goals

- 이미 scope 가 있는 카드의 scope 를 take 가 바꾸는 일(예약만 다르게 잡을 수 있어요).
- 끝난 리스가 파일을 막는 문제, 예약 scope 줄이기, `/clear` 뒤 이어받기(다른 카드의 일).
- `card scope` 명령의 동작.
- SPEC 을 쓰는 규칙 자체와 `spec check`.

## Requirements

- REQ-1: When `card take` is given `--scope` and the card's own write scope is empty, the take shall store the given scope as the card's write scope in the reserving transaction and record one `card.scope` event whose actor is the taker. · failure: none — a refused take writes nothing, REQ-7 covers it.
- REQ-2: When `card take` is given `--scope` and the card's own write scope is not empty, the take shall leave the card's write scope unchanged. · failure: none — only the reservation takes the given scope, and REQ-7 covers a refusal.
- REQ-3: When `card take` is given no `--scope` and the card's own write scope is empty, the take shall use the `write:` list of the SPEC the card names and store it as the card's write scope. · failure: none — REQ-6 covers a card with no source.
- REQ-4: When the card names no SPEC or its SPEC has no readable `write:` list, the take shall use the scope of the card's most recent reservation and store it as the card's write scope. · failure: none — REQ-6 covers a card with no source.
- REQ-5: The answer of a take shall carry `scope_source` naming where the scope came from: `given`, `card`, `spec` or `last_reservation`.
- REQ-6: If no source gives a scope, then the take shall be refused with a message that names `--scope` and `card scope` and writes nothing.
- REQ-7: The take shall store a filled or given scope on the card in the same write transaction as the overlap check, so a scope that overlaps another card's active reservation is refused and leaves the card as it was.
- REQ-8: If a SPEC `write:` list holds more than 64 paths, then the take shall be refused with the 64-path limit named and nothing written.
- REQ-9: `card next` shall list a `ready` card whose empty scope a take would fill from its SPEC or its last reservation, showing the filled scope and its `scope_source`, instead of skipping it as `empty_scope`.

## Acceptance

- AC-1 (REQ-1, REQ-5): a card added without a scope is taken with `--scope a --scope b` and released; `card show` then reads the scope `[a, b]`, the answer of the take named `scope_source` `given`, and `card next` lists the card.
  verify: `cargo test --locked --offline --test card_take_scope_cli a_given_scope_is_kept_on_the_card`
- AC-2 (REQ-2): a card that has the scope `[a]` is taken with `--scope b`; the reservation has `[b]` and the card still has `[a]`.
  verify: `cargo test --locked --offline --test card_take_scope_cli an_own_scope_is_never_changed_by_take`
- AC-3 (REQ-3, REQ-5): an empty-scope card that names a SPEC with a `write:` list is taken without `--scope`; the reservation and the card get that list and `scope_source` reads `spec`.
  verify: `cargo test --locked --offline --test card_take_scope_cli an_empty_card_is_filled_from_its_spec`
- AC-4 (REQ-4, REQ-5): an empty-scope card with a SPEC without a list, or no SPEC, that was taken and released earlier is taken again without `--scope`; it gets the last reservation's scope and `scope_source` reads `last_reservation`.
  verify: `cargo test --locked --offline --test card_take_scope_cli an_empty_card_is_filled_from_its_last_reservation`
- AC-5 (REQ-6): an empty-scope card with no SPEC and no earlier reservation is refused; the message holds `--scope` and `card scope`, and the ledger is unchanged.
  verify: `cargo test --locked --offline --test card_take_scope_cli no_source_names_the_two_ways_out`
- AC-6 (REQ-7): a fill whose paths overlap another card's active reservation is refused with the overlap named, and the card's scope stays empty with no event.
  verify: `cargo test --locked --offline --test card_take_scope_cli a_fill_keeps_the_overlap_check`
- AC-7 (REQ-8): a SPEC `write:` list of 64 paths fills the scope and a list of 65 is refused with the limit named, nothing written.
  verify: `cargo test --locked --offline --test card_take_scope_cli the_spec_list_has_the_same_limit`
- AC-8 (REQ-9): `card next` lists an empty-scope ready card that has a SPEC list or an earlier reservation, with its filled scope and `scope_source`, and still skips one with no source as `empty_scope`.
  verify: `cargo test --locked --offline --test card_take_scope_cli next_lists_what_take_can_fill`
- AC-9 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the existing take, scope and next tests of the ledger still pass.
  verify: `cargo test --locked --offline --test ledger_reservation --test card_scope_cli --test card_next_cli`
- AC-10 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261121-take-scope-stored`

## Scope

- write: `.moai/specs/SPEC-20261121-take-scope-stored/**`, `tests/card_take_scope_cli.rs` (new), `src/card/take_scope.rs` (new: the SPEC `write:` list read), `src/card/mod.rs` (one module line), `src/card/dispatch.rs` (`take`), `src/card/work.rs` (`next`: the fill of a card it would skip), `src/ledger/reservation.rs` (the reserving transaction and the last-reservation read)
- read-only: every other file, the real HOME
- tier 2: an independent audit of the exact head (a ledger rule)
- depends on: `src/ledger/reservation.rs` is held in turn by the early-reclaim change and the renewal change, and `src/card/work.rs` by the shared-line-lease change; this change takes them after they land. The SPEC and the red tests go first

## Plan

1. SPEC first (this file).
2. RED: `tests/card_take_scope_cli.rs`, one test per acceptance line, each failing for the stated reason.
3. `take_scope.rs`: read the `write:` list of a SPEC folder (bounded, no symlink out of the project) and answer a list or none.
4. `reservation::reserve_filled`: the reserving transaction with an optional SPEC list as fallback; it stores the scope on the card with the `card.scope` event and answers the source; `reserve` and `reserve_own_scope` keep their results.
5. `dispatch::take`: pass the SPEC list, add `scope_source` to the answer, refusal text with the two ways out.
6. `card next`: the same order for the cards it would skip as `empty_scope`.
7. The unnarrowed `check affected` set, then an independent audit (tier 2).

## Risks

- A SPEC list wider than the work wants: the taker sees the scope in the answer and can widen or narrow it as before; the card keeps what was used, and `card scope` replaces it on a card nobody holds.
- A path in an old reservation that another card holds now: the overlap check runs on the filled scope in the same transaction, so the take is refused and nothing is stored.
- The SPEC folder is read from the project the command is run for; a SPEC that lives only on a lane branch is not seen and the next source is used.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the order is the given scope, the card's own, the SPEC list, the last reservation | the report names this order; the card's own scope stays the authority | change the order |
| D-2 | a SPEC list over 64 paths refuses instead of falling through to the last reservation | a SPEC that names the work in full should not be silently replaced by an older scope | fall through instead |
| D-3 | the fill is stored on the card with a `card.scope` event | the stored scope is what makes the card visible to `card next` after the release | keep the fill on the reservation only |

## Result

(written at SYNC)
