---
id: SPEC-20261121-card-lease-yield
title: "an ended lease no longer blocks a new take or scope widening, and the holder of an ended lease can renew it"
status: draft
revision: 1
created: 2026-11-21
---

# an ended lease no longer blocks a new take or scope widening, and the holder of an ended lease can renew it

## Intent

A reservation whose lease has ended still blocks every path of its scope, and its holder cannot renew it; the only way out is a leader who reclaims and releases it. Of the held reservations in a recent count of the ledger, 25 had an ended lease, and a lane that waited on one could do nothing. Three changes remove the wait without a new schema. A take or a scope widening on a path that only an ended lease holds succeeds and removes that path from the old reservation, with one `card.yield` event that names who took it and why. A live lease still refuses, and its refusal names the holder, the card and the lease end. The holder of an ended lease that nobody reclaimed renews it at the same generation instead of reclaiming itself.

## Requirements

- REQ-1: When a card is taken and a path of its scope overlaps a path of another card's held reservation whose lease has ended, the take shall remove the overlapping paths from that reservation and continue.
- REQ-2 (REQ-1): IF an overlapping held reservation has a lease that has not ended, THEN the take shall refuse, name the holder, the card and the lease end, and write nothing.
- REQ-3: When a held card widens its scope with a path that overlaps another card's held reservation whose lease has ended, the widening shall remove the overlapping paths from that reservation and continue.
- REQ-4 (REQ-3): IF an overlapping held reservation has a lease that has not ended, THEN the widening shall refuse, name the holder, the card and the lease end, and write nothing.
- REQ-5: When paths are removed from an ended reservation, the ledger shall write one `card.yield` event on the old card that names the paths, the old and the new scope, the old holder, its generation, its lease end, who took the paths and why. · failure: none — the event is written in the same transaction as the removal, so a refused take or widening writes neither, and the refusals are REQ-2 and REQ-4.
- REQ-6: While a held reservation has an ended lease and nobody reclaimed it, the renew command shall extend the lease from the database clock at the same generation.
- REQ-7 (REQ-6): IF the generation is not the current one or the caller is not the holder, THEN the renew command shall refuse as it did before.
- REQ-8: The next-card listing shall not skip a ready card because a held reservation with an ended lease overlaps its scope.
- REQ-9 (REQ-1): IF the overlapping reservation is in review or is an expired candidate, THEN the take and the widening shall refuse as for a live lease.

## Acceptance

- AC-1 (REQ-1, REQ-2): a take on a path that only an ended lease holds succeeds and the old reservation no longer lists that path, a take on a path of a live lease is refused with the card, the holder and the lease end named and no row changed, and the lease edge is pinned one second before, at and one second after the lease end · verify: `cargo test --locked --offline --test ledger_lease_yield take_`
- AC-2 (REQ-3, REQ-4): a scope widening behaves as a take does, for an ended and for a live lease · verify: `cargo test --locked --offline --test ledger_lease_yield widen_`
- AC-3 (REQ-5): one `card.yield` event names the paths, both scopes, the old holder, the generation, the lease end, the taker and the reason; the old reservation keeps its generation, holder and state, and nothing is written when a take is refused · verify: `cargo test --locked --offline --test ledger_lease_yield event_`
- AC-4 (REQ-6, REQ-7): renew extends an ended lease of the same session at the same generation, one second before and at the lease end, and a stale generation or another session is refused · verify: `cargo test --locked --offline --test ledger_lease_yield renew_`
- AC-5 (REQ-8, REQ-9): the next-card listing offers a ready card whose scope only an ended lease overlaps and still skips one that a live lease, a review or an expired candidate overlaps · verify: `cargo test --locked --offline --test ledger_lease_yield next_`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9): the existing card and ledger tests that read a reservation still pass, the one test that pinned a refused renew of an ended lease now pins the new rule · verify: `cargo test --locked --offline --no-fail-fast --test ledger_reservation --test ledger_workflow --test ledger_scope_return --test card_work_cli --test card_scope_widen_cli --test card_scope_widen_followup_cli --test card_dispatch_cli --test card_reclaim_cli --test card_line_lease_cli --test prop_reservation`

## Scope

- write: `.moai/specs/SPEC-20261121-card-lease-yield/spec.md`
- write: `src/ledger/reservation.rs`, `src/ledger/workflow.rs`
- write: `tests/ledger_lease_yield.rs`, `tests/ledger_reservation.rs`, `tests/ledger_workflow.rs`

## Decisions

- D-1: A path is yielded by deleting the overlapping `reservation_scope` rows of the old reservation, as `card give-back` does; the old card keeps its notes, branch, generation, holder and state, and the history stays in the `card.yield` event. The schema does not change.
- D-2: Only a `held` reservation yields. A reservation in review or an expired candidate keeps blocking, because an audit or a merge may still need its files; a lane that wants them asks the holder or the leader as before.
- D-3: The whole overlapping pattern of the old reservation is removed, not only the part the taker needs: the scope is a list of patterns, not a set of files, and the old holder widens again, with a reason, when its lease is renewed and the paths are free.
- D-4: The renew of an ended lease needs no new overlap check. A path that was yielded is no longer in the old scope, so the renewed reservation overlaps nothing, and `card check --path` refuses the yielded path for the old holder because it is not in its scope.
- D-5: The refusal names the lease end in Unix seconds with the UTC time next to it, so a reader can compare it with the clock; the first words of the refusal stay as they were, so the error class does not change.
- D-6: Not in this card: a separate `card yield` command that removes paths on demand. The automatic yield covers the take and the widening that lanes wait on; the command would also need `src/ops/card.rs` and `src/card/work.rs`, which another card holds until it lands.

## Plan

1. Tests first (RED): `tests/ledger_lease_yield.rs` with the take, widen, event, renew and next cases of AC-1 to AC-5, and the pure unit test of the lease edge; see them fail on the unchanged code.
2. `find_overlap` takes the database clock, ignores a held reservation whose lease has ended and answers the holder and the lease end of the one that blocks; a `yield_overlaps` function removes the overlapping rows of ended reservations and writes the event, called by the take and the widening after the overlap check passes.
3. `renew` stops needing a live lease; `next` passes its clock to the overlap check.
4. Update the one test that pinned the old renew, run the card and ledger tests, the repository scans, `--lib`, `--bins`, clippy, then audit and write the Result.

## Result

- AC-1: met — `cargo test --locked --offline --test ledger_lease_yield take_` passed (6 tests; they were red before the change except the ones that pin unchanged refusals): a take on a path that only an ended lease holds succeeds and the old reservation keeps its other paths, its holder, generation and state; a live lease is refused with the card, the holder and the lease end named and nothing written; a review reservation and an expired candidate still refuse when their lease ended; a mixed live and ended overlap refuses and yields nothing; `blocks_new_claims` is pinned one second before, at and one second after the lease end.
- AC-2: met — `... --test ledger_lease_yield widen_` passed (3 tests): the same three cases for a scope widening.
- AC-3: met — `... --test ledger_lease_yield event_` passed (3 tests): one `card.yield` event on the old card with the paths, both scopes, the old holder, generation, lease end, the taker and the reason; one event per ended reservation; a widening names the widening card.
- AC-4: met — `... --test ledger_lease_yield renew_` passed (4 tests): an ended lease of the holder is extended at the same generation, a live lease is still extended and never shortened, a wrong generation and another session are refused, and a renew after a yield keeps the rest of the scope and not the yielded path.
- AC-5: met — `... --test ledger_lease_yield next_` passed (3 tests): the listing offers a card that only an ended lease overlaps and skips one that a live lease, a review or an expired candidate overlaps.
- AC-6: not yet — every listed file passed with `--no-fail-fast` except `prop_reservation`, which passed its 1000 cases and then failed its own 30 second bound (see below); the other files are `ledger_reservation`, `ledger_workflow`, `ledger_scope_return`, `card_work_cli`, `card_scope_widen_cli`, `card_scope_widen_followup_cli`, `card_dispatch_cli`, `card_reclaim_cli` and `card_line_lease_cli`, plus `ledger_schema`, `ledger_recover`, `card_cli`, `card_generated_scope`, `card_precheck_cli`, `claim_renew_doc_3042` and `mode_workflow_steps`. `ledger_reservation` changed in two places (the error variants gained fields; the test that pinned a refused renew of an ended lease now pins the fence and handoff refusals and the accepted renew) and `card_reclaim_cli` lost the one renew refusal it pinned for the stuck-card case. AC-6 is met when `prop_reservation` passes on a quiet machine.
- Whole-area run on this head: `--lib` 1619 passed, `--bins` 28 and 28 passed, `cargo clippy --all-targets -D warnings` exit 0, rustfmt of the four Rust files ok.
- Mutants (five, each restored): an always-blocking rule (killed by at least 7 tests), a review that stops blocking (5 tests), a yield that removes every pattern of the old reservation (4 tests), a renew that needs a live lease again (2 tests) and `next` ignoring the clock (1 test).
- Not green here, not caused by the change: `prop_reservation random_interleavings_keep_the_two_invariants` passed its 1000 cases and then failed its own 30 second bound (78 seconds) while the machine's load average was above 60; it is to be re-run on a quiet machine and on the unchanged base for comparison.
- Not done here, and needed in the same release: step 5 of `skills/moai/workflows/mode-agent.md` and the two tests that pin it (`tests/claim_renew_doc_3042.rs`, `tests/mode_workflow_steps.rs`) still say `renew` needs a lease that has not ended and that an ended lease is the leader's to settle; since this change a holder renews its own ended lease while nobody reclaimed it, and a lane that wants another card's ended paths just takes them. A separate `card yield` command (D-6), a test at the command-line level for the new refusal text, the python area modules, the audit and a test for macOS are not done.
- closure_status: the audit and the whole-area run after the merge onto the train are still to come.
