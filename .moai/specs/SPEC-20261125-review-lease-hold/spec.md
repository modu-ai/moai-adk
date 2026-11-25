---
id: SPEC-20261125-review-lease-hold
title: "A card in review keeps its lease until the verdict is recorded, and a reclaimed reservation brings back its scope"
status: draft
revision: 1
created: 2026-11-25
---

# A card in review keeps its lease until the verdict is recorded, and a reclaimed reservation brings back its scope

## Intent

A card that is submitted waits for an independent audit, and the lease of its reservation keeps running while it waits (default 14400 seconds from `card take`). When the lease ends the author lane cannot close the card: `release` is refused on a `review` card, `reopen` needs a live lease, and a different session can only `card reclaim` it, which counts against a lifetime allowance of three. A measurement in another repository (addendum 4 of the mo.ai.kr leader's improvement request, moai 4.2648.35) shows the result: one card in review lost its lease twice, reached generation 3 and 2 of 3 reclaims, so a FAIL verdict and a repair would have found the allowance gone; this repository had the same thing when a card stayed held until 05:36Z on 2026-11-25. The same measurement shows that every reclaim came back with an empty scope, so the lane restored it by hand with `card scope` each time. The cause is in this repository's code, not a bug of the other one: `card review` gives every path of the reservation back unless `--keep` names it, `card reclaim` copies the scope of the old reservation (`reservation_scope`) and so copies nothing, and `yield_overlaps` can take paths away from an ended lease besides. This SPEC makes the lease of a card in review last until its verdict, gives the lane a fresh lease at the verdict, lets the reclaim and the reopen of a repair bring the scope back, and does not count the first reclaim after a FAIL against the allowance.

## Interview

- Q1: A card waits in review for its audit while its lease ends and costs reclaims. Which way do you want it fixed? Options: (a) the lease of a card in review lasts until the verdict is recorded, then the lane gets a fresh lease; (b) the leader renews other sessions' reservations; (c) a takeover after an ended lease does not count against the allowance. → A: not asked yet. The leader's default is (a), and this SPEC recommends (a) with a 24 hour cap (D-1): a review that waits longer than a day is a stalled card the leader should see, and the lease checks keep working unchanged. Pick (a) pure (no cap) if you want the literal pause; it needs a pause marker in the ledger (D-1, option 2).
- Q2: After a reclaim the new reservation has no paths, so the lane writes them back by hand. Should a reclaim (and a reopen after a FAIL) bring the card's own scope back the way a take does? → A: not asked yet. Default: yes, minus paths of another card's live reservation and shared files, which are named in the answer (D-3).
- Q3: After a FAIL verdict the lane may find its three reclaims used up. Should the first reclaim after each FAIL verdict stay out of the count? → A: not asked yet. Default: yes, one per FAIL verdict (D-2). A general way out for a card that reached the limit belongs to `SPEC-20261125-card-take-records-worktree` and is not decided here.

## Non-goals

- A general exit for a card that reached the three-reclaim limit for any other reason (it belongs to `SPEC-20261125-card-take-records-worktree`).
- Which sessions may renew or take over another session's reservation (the ownership checks stay as they are: a live session's reservation is never taken).
- The work-in-progress limit and what it counts (`SPEC-20261125-wip-limit-live-leases`).
- Removing path-overlap refusals between cards (`SPEC-20261125-card-no-file-locks`); this SPEC only reads the overlap rule where it exists.
- The lease of an `active` card, the sweep of ended leases, the early reclaim of a proven-gone holder and `card adopt`.

## Requirements

- REQ-1: When a card moves to `review`, the review command shall extend the current reservation's lease to end no earlier than the ledger clock plus 86400 seconds.
- REQ-2 (REQ-1): When a verdict is recorded on a card in `review`, the verdict command shall extend the current reservation's lease to end no earlier than the ledger clock plus 14400 seconds. · failure: none — a lease that already ends later is left as it is.
- REQ-3 (REQ-1): IF a card stays in `review` for 86400 seconds without a verdict, THEN its lease shall end as every other lease does. · failure: none — the reclaim and the sweep rules apply from then on.
- REQ-4: While a card is in `review` and its lease has not ended, the reclaim command shall refuse with `lease_live` and name that the lease is held for the audit and when it ends. · failure: none — the message only names the reason.
- REQ-5: When a reclaim or a reopen leaves the new or the continued reservation with no paths, the command shall fill the scope from the card's own scope, as a take does. · failure: none — a reservation that has paths is not changed.
- REQ-6 (REQ-5): IF a path of the card's own scope overlaps a path of another card's reservation or a shared file of the profile, THEN the command shall leave that path out, name it in the answer and still succeed.
- REQ-7: When a reclaim follows a FAIL verdict of the card and no `card.reclaim` event of the card is newer than that verdict, the reclaim command shall leave that reclaim out of the three-reclaim allowance. · failure: none — every other reclaim counts as before.
- REQ-8 (REQ-7): The reclaim command shall write `counted` in its event and name in its answer whether the reclaim counted, the reclaims counted so far and the allowance.
- REQ-9: The lease rules shall apply as before this SPEC to a card that is not in `review`, to the sweep, to the scope yield and to the early reclaim. · failure: none — a lane in the active state keeps its four hours.

## Acceptance

- AC-1 (REQ-1, REQ-2): in a temporary ledger, `card review` on a card with a short lease leaves a lease that ends at least 86400 seconds after the review, and the first `card verdict` on it leaves a lease that ends at least 14400 seconds after the verdict; a lease that already ends later stays as it is · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac1_`
- AC-2 (REQ-3, REQ-4): a card in review whose lease is live is refused by `card reclaim` with `lease_live` and a message that names the audit and the end time; after the 86400 seconds (test clock) the reclaim succeeds and counts · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac2_`
- AC-3 (REQ-5, REQ-6): a card reviewed without `--keep`, whose lease then ended, is reclaimed with the card's own scope again; a path that another card's live reservation overlaps is left out and named, and the reclaim still succeeds; `card reopen` on a review card with an empty reservation fills it the same way · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac3_`
- AC-4 (REQ-7, REQ-8): a card with three counted reclaims and a FAIL verdict is reclaimed once more with `counted` false, a second reclaim after the same verdict is refused with `reclaim_limit`, and a card without a FAIL verdict is refused at the fourth reclaim as before; the answer names the reclaims counted and the allowance · verify: `cargo test --locked --offline --test card card_review_lease_hold_cli::ac4_`
- AC-5 (REQ-9): the lease of an `active` card, the sweep of an ended lease and `card adopt` give the same results as before · verify: `cargo test --locked --offline --test card card_reclaim_cli`
  verify: `cargo test --locked --offline --test card card_adopt_cli`
  verify: `cargo test --locked --offline --test card ledger_lease_yield`
  verify: `cargo test --locked --offline --test card ledger_sweep`
  verify: `cargo test --locked --offline --test card ledger_workflow`
  verify: `cargo test --locked --offline --test card ledger_scope_return`

## Scope

- write: `.moai/specs/SPEC-20261125-review-lease-hold/spec.md`
- write: `src/ledger/reservation.rs`, `src/ledger/workflow.rs`, `src/card/work.rs`, `src/card/dispatch.rs`
- write: `tests/card/card_review_lease_hold_cli.rs`, `tests/card/main.rs`, `tests/card/card_reclaim_cli.rs`
- read-only: `src/ledger/sweep.rs`, `src/ledger/adopt.rs`, `src/card/wip_limit.rs`, `src/card/early_reclaim.rs`
- forbidden: the real HOME, the installed moai binary, another project, the ledger schema (no new column or table), a card number inside any file other than the card's own notes

## Plan

1. Wait for the approval of this SPEC and for `SPEC-20261125-card-no-file-locks` to land: it removes the path-overlap refusal between cards, which changes what REQ-6 has to leave out; if it lands first the overlap clause of REQ-6 shrinks to the shared-file case.
2. Write the failing tests first (`tests/card/card_review_lease_hold_cli.rs`, AC-1 to AC-4) against a temporary ledger with a test clock.
3. In `src/ledger/workflow.rs` extend the lease inside the same transaction as the state change of `review` (REQ-1) and of the verdict (REQ-2); keep one helper for "lease no earlier than clock plus n seconds".
4. In `src/ledger/reservation.rs` word the `lease_live` answer for a card in `review` (REQ-4) and exclude a reclaim that follows a FAIL verdict from the count (REQ-7); the count stays derived from events, with the `counted: false` payload as the only exception.
5. In `src/card/dispatch.rs` and `src/card/work.rs` fill an empty scope after reclaim and reopen from the card's own scope with the take's overlap and shared-file rules (REQ-5, REQ-6) and add the three fields of REQ-8 to the reclaim answer.
6. Run the neighbour tests of AC-5, record the Result with one line per acceptance criterion, and name every run that was not made.

## Risks

- The lease of a review card is longer, so a path kept with `card review --keep` stays reserved for up to a day; the leader sees it as a review card without a verdict and can ask for the audit. If the audit never comes the cap of REQ-3 ends it.
- `SPEC-20261125-card-take-records-worktree` and `SPEC-20261125-card-no-file-locks` also change `src/ledger/reservation.rs` and `src/card/dispatch.rs`; this SPEC must be rebased on both and the reclaim count query must stay one place.
- The work-in-progress limit (`SPEC-20261125-wip-limit-live-leases`) counts cards with a live lease, so a review card now counts for up to a day; the measurement after the change should look at how many review cards the limit holds.
- A lane that is gone while its card is in review can neither renew nor reopen; after the cap the reclaim is the only way, and with REQ-7 only the first one after a FAIL is free. A card that fails twice goes to the leader as the flow says.
- Filling the scope on reclaim can reserve a path that a lane no longer needs; the answer names what was filled and `card scope` can still change it.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The lease of a card in review is extended to 24 hours at `review` and to the default 4 hours at the verdict (REQ-1 to REQ-3). Options: 1 this window (recommended: no schema change, every lease check keeps one rule); 2 a literal pause until the verdict (needs a pause marker in the ledger and a change in every lease check); 3 no change, the leader renews other sessions' reservations. | The leader's default is a pause until the verdict; this SPEC proposes the window with the same effect for any audit that arrives within a day. The user has not chosen yet. | Remove the two extensions and the reclaim wording (REQ-1, REQ-2, REQ-4). |
| D-2 | The first reclaim after a FAIL verdict is not counted (REQ-7). | A repair after a FAIL must not stop at the allowance; one per FAIL keeps the allowance bounded because a second FAIL goes to the leader. | Count every reclaim again (remove the event exception). |
| D-3 | A reclaim and a reopen fill an empty scope from the card's own scope and leave out overlapping and shared paths (REQ-5, REQ-6). | `card take` already fills an empty scope this way, and `card review` empties the scope on purpose (`card review` gives back every path without `--keep`). | Remove the fill; the lane widens the scope by hand as before. |
| D-4 | The 86400 second cap equals the existing maximum lease (`MAX_LEASE_SECS`). | No new number: a lease may already be at most one day. | Change the one constant. |
