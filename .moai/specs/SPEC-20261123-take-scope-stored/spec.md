---
id: SPEC-20261123-take-scope-stored
title: "card take keeps the scope it was given on the card and fills an empty scope from the SPEC or the last reservation"
status: draft
revision: 1
created: 2026-11-23
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

- REQ-1: When `card take` is given `--scope` and the card's own write scope is empty, the take shall store the given scope as the card's write scope as soon as the reservation is made and record one `card.scope` event whose actor is the taker. · failure: none — a refused take writes nothing, REQ-7 covers it.
- REQ-2: When `card take` is given `--scope` and the card's own write scope is not empty, the take shall leave the card's write scope unchanged. · failure: none — only the reservation takes the given scope, and REQ-7 covers a refusal.
- REQ-3: When `card take` is given no `--scope` and the card's own write scope is empty, the take shall use the `write:` list of the SPEC the card names and store it as the card's write scope. · failure: none — REQ-6 covers a card with no source.
- REQ-4: When the card names no SPEC or its SPEC has no readable `write:` list, the take shall use the scope of the card's most recent reservation and store it as the card's write scope. · failure: none — REQ-6 covers a card with no source.
- REQ-5: The answer of a take shall carry `scope_source` naming where the scope came from: `given`, `card`, `spec` or `last_reservation`.
- REQ-6: If no source gives a scope, then the take shall be refused with a message that names `--scope` and `card scope` and writes nothing.
- REQ-7: The take shall store a filled or given scope on the card only after the reservation, whose own overlap check runs first, has been made, so a scope that overlaps another card's active reservation is refused and leaves the card as it was.
- REQ-8: If a SPEC `write:` list holds more than 64 paths, then the take shall be refused with the 64-path limit named and nothing written.
- REQ-9: `card next` shall list a `ready` card whose empty scope a take would fill from its SPEC or its last reservation, showing the filled scope and its `scope_source`, instead of skipping it as `empty_scope`.
- REQ-10: If an entry of the SPEC `write:` list is not shaped like a path (it has white space or a comma, starts with `-`, or has no `/`, glob character or file extension), then the take shall refuse, fill nothing and name the SPEC file, the entry, `card take <C-n> --scope <path>` and `card scope <C-n> --scope <path>`. · failure: the card stays empty and `card next` skips it as `empty_scope`.
- REQ-11: When the SPEC `write:` list names a path more than once, the take shall reserve and store the path once and name the merged paths in the `warnings` of its answer. · failure: none — a merged list that still breaks REQ-8 or REQ-10 is refused by them.
- REQ-12: If a `--scope`, `--path` or `--keep` value holds a comma, then the command shall refuse it and say that one such option, named as the caller typed it, names one path. · failure: nothing is written, for `take`, `card scope`, `give-back`, `line-lease` and `review`.

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
- AC-6 (REQ-7): a fill whose paths overlap another card's active reservation is refused with the write scope in conflict and the card that holds it named, and the card's scope stays empty with no event.
  verify: `cargo test --locked --offline --test card_take_scope_cli a_fill_keeps_the_overlap_check`
- AC-7 (REQ-8): a SPEC `write:` list of 64 paths fills the scope and a list of 65 is refused with the limit named, nothing written.
  verify: `cargo test --locked --offline --test card_take_scope_cli the_spec_list_has_the_same_limit`
- AC-8 (REQ-9): `card next` lists an empty-scope ready card that has a SPEC list or an earlier reservation, with its filled scope and `scope_source`, and still skips one with no source as `empty_scope`.
  verify: `cargo test --locked --offline --test card_take_scope_cli next_lists_what_take_can_fill`
- AC-9 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the existing take, scope and next tests of the ledger still pass.
  verify: `cargo test --locked --offline --test ledger_reservation --test card_scope_widen_cli`
- AC-11 (REQ-8, REQ-9): `card next` leaves a card with a SPEC list of 65 paths skipped as `empty_scope`, so it never offers what a take would refuse.
  verify: `cargo test --locked --offline --test card_take_scope_cli next_skips_a_fill_a_take_would_refuse`
- AC-12 (REQ-10): on `## Scope` sections copied from three real SPECs (a code span `pub mod build;`, a prose line whose words were read as paths, an option span `--reason`), a take is refused naming the SPEC file, the entry and both ways out, the ledger is unchanged, the card keeps no scope and `card next` skips it as `empty_scope`; a clean real list still fills.
  verify: `cargo nextest run --locked --offline --test card_take_scope_cli -E 'test(a_spec_list_with_a_word_is_refused_not_stored) | test(a_clean_real_spec_list_still_fills)'`
- AC-13 (REQ-11): a real untagged list whose items name `src/home_workdir.rs` twice is taken with the path once and a `merged` warning.
  verify: `cargo nextest run --locked --offline --test card_take_scope_cli -E 'test(a_spec_list_that_repeats_a_path_is_merged)'`
- AC-14 (REQ-12): `take --scope a,b` and `card scope --scope a,b` are refused naming the comma and `one --scope per path`, with nothing reserved and the card scope unchanged; `give-back`, `line-lease` and `review` name `--path`, `--path` and `--keep`.
  verify: `cargo nextest run --locked --offline --test card_take_scope_cli -E 'test(a_comma_in_a_scope_value_is_refused)'`
- AC-15 (REQ-4): a card with two earlier reservations gets the newest one's scope.
  verify: `cargo nextest run --locked --offline --test card_take_scope_cli -E 'test(the_last_reservation_is_the_newest_one)'`
- AC-16 (REQ-7, REQ-9): `card next` skips a card whose fill overlaps another card's active reservation as `scope_overlap` and does not offer it.
  verify: `cargo nextest run --locked --offline --test card_take_scope_cli -E 'test(next_skips_a_fill_that_overlaps_a_held_scope)'`
- AC-10 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11, REQ-12): the SPEC passes the checker.
  verify: `moai spec check SPEC-20261123-take-scope-stored`

## Scope

- write: `.moai/specs/SPEC-20261123-take-scope-stored/**`, `tests/card_take_scope_cli.rs` (new), `src/card/take_scope.rs` (new: the SPEC `write:` list read, the shared fill choice, the store), `src/card/mod.rs` (one module line), `src/card/dispatch.rs` (`take`), `src/card/work.rs` (`next`: passes the fill choice), `src/ledger/workflow.rs` (`next_filled`: the ledger read takes a fill function and runs the overlap check on the filled scope; `next` keeps its signature and answer), `src/ops/card.rs` and `docs/design-review/appendix/operations/card.md` (the help and the page say what a take fills and refuses; repair of the first audit)
- read-only: every other file, the real HOME; `src/ledger/reservation.rs` is not edited: the reservation's own overlap check runs first and the scope is kept on the card afterwards
- tier 2: an independent audit of the exact head (a ledger rule)
- history: the SPEC, the red tests and the first piece were written on an earlier branch (`spec/take-scope-stored-20261121`, never landed); this branch starts again from the integration tip and takes them over, so the second piece no longer waits for other cards

## Plan

1. SPEC first (this file).
2. RED: `tests/card_take_scope_cli.rs`, one test per acceptance line, each failing for the stated reason.
3. `take_scope.rs`: read the `write:` list of a SPEC folder (bounded, no symlink out of the project) and answer a list or none.
4. `take_scope::store`: after a successful reservation, keep the scope on a card whose own scope is empty, with one `card.scope` event; `reserve` and `reserve_own_scope` keep their results.
5. `dispatch::take`: choose the scope by the order of D-1, add `scope_source` to the answer, refusal text with the two ways out (first piece: REQ-1 to REQ-8).
6. `card next` uses the same choice (`take_scope::choose`, shared with `take`) for the cards the ledger read would skip as `empty_scope`: `workflow::next_filled` takes it as a fill function, so a card that a take could fill and that meets no overlap is offered as `ready` with its filled scope and `scope_source` (REQ-9); nothing is written.
7. The unnarrowed `check affected` set, then an independent audit (tier 2).

## Risks

- A SPEC list wider than the work wants: the taker sees the scope in the answer and can widen or narrow it as before; the card keeps what was used, and `card scope` replaces it on a card nobody holds.
- A path in an old reservation that another card holds now: the reservation's overlap check runs on the filled scope before anything is stored, so the take is refused and nothing is stored.
- The SPEC folder is read from the project the command is run for; a SPEC that lives only on a lane branch is not seen and the next source is used.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | the order is the given scope, the card's own, the SPEC list, the last reservation | the report names this order; the card's own scope stays the authority | change the order |
| D-2 | a SPEC list over 64 paths refuses instead of falling through to the last reservation | a SPEC that names the work in full should not be silently replaced by an older scope | fall through instead |
| D-3 | the fill is stored on the card with a `card.scope` event | the stored scope is what makes the card visible to `card next` after the release | keep the fill on the reservation only |
| D-4 | a path the SPEC list names twice is merged into one, with a `merged` warning, instead of refused | the first audit found 10 of 321 real SPECs refused for a repeat with no way out; the repeat is a note, not a mistake (leader decision 2026-11-23, the lane chose) | refuse a repeat as `--scope` does |
| D-5 | every entry of a SPEC list must be shaped like a path (no white space or comma, no leading `-`, a `/`, a glob character or a file extension), else the whole list is refused | the first audit stored 61 words (`the`, `--reason`, `pub mod build;`) as paths from 311 real SPECs; a bare file name without an extension (`HEADVER`) is refused too and is named with `--scope` (leader decision 2026-11-23) | accept words and let the taker narrow |
| D-6 | the comma check is in the shared `checked_scope`, so `add`, `scope`, `take`, `review`, `give-back` and `line-lease` all refuse a comma in a value | one place; the glob has no braces, so no path of a scope holds a comma | check in `take` only |

## Result

- AC-1: met — `card_take_scope_cli a_given_scope_is_kept_on_the_card` (nextest, pass)
- AC-2: met — `an_own_scope_is_never_changed_by_take` (pass; the guard is two layers, the take condition and the SQL condition, and a mutant that removes both fails this test, one layer alone is covered by the other)
- AC-3: met — `an_empty_card_is_filled_from_its_spec` (pass)
- AC-4: met — `an_empty_card_is_filled_from_its_last_reservation` (pass)
- AC-5: met — `no_source_names_the_two_ways_out` (pass; a mutant that renames `card scope` in the message fails it)
- AC-6: met — `a_fill_keeps_the_overlap_check` (pass; a mutant that stores the scope before the reservation fails it)
- AC-7: met — `the_spec_list_has_the_same_limit` (pass)
- AC-8: met — `next_lists_what_take_can_fill` (pass)
- AC-9: met — `ledger_reservation`, `card_scope_widen_cli`, `card_cli`, `card_dispatch_cli`, `card_work_cli`, `ledger_workflow` pass inside the 3133 tests of `check affected` on the head
- AC-10: met — `spec check SPEC-20261123-take-scope-stored` 0 errors 0 warnings
- AC-11: met — `next_skips_a_fill_a_take_would_refuse` (pass; a mutant that removes the limit check in `take_scope::choose` fails it, which no test caught before it was added)
- AC-12: met — `a_spec_list_with_a_word_is_refused_not_stored` over five `## Scope` sections copied from real SPECs (code span, prose line, option span, spaced span, call span) plus a `--out/dir.rs` entry that only its leading `-` makes no path, and `a_clean_real_spec_list_still_fills` (pass; mutants that turn the shape check off, drop its white-space rule, drop its leading-`-` rule, accept any extension, or drop `card scope` from the refusal each fail it)
- AC-13: met — `a_spec_list_that_repeats_a_path_is_merged` (pass; mutants that turn the merge off or drop the `merged` warning fail it)
- AC-14: met — `a_comma_in_a_scope_value_is_refused` and `the_comma_refusal_names_the_option_of_the_caller` (pass; a mutant that removes the comma check fails the first, one that names `--scope` for `review` fails the second)
- AC-15: met — `the_last_reservation_is_the_newest_one` (pass; a mutant that orders the reservations oldest first fails it)
- AC-16: met — `next_skips_a_fill_that_overlaps_a_held_scope` (pass; a mutant that skips the overlap check for a filled scope fails it)
- Repair of the delta audit (B4 the leading-`-` rule had no killing fixture, N6 the comma refusal said `--scope` for `--path` and `--keep`): AC-12 and AC-14 above.
- Repair of the first audit (B1 words stored as paths, B2 and B3 without a killing test, N1 stale help, N2 comma): D-4, D-5 and D-6 record the choices. The `check affected` receipt of the head is in the READY note of that head (a number in this file would name an earlier head).
