---
id: SPEC-20261126-card-next-speed
title: "card next and card list answer a real ledger in 0.1 seconds"
status: draft
revision: 2
created: 2026-11-26
---

## Interview

- none: the speed target came from the leader's measured stage bottleneck of 2026-11-25; no new question went to the user.

# card next and card list answer a real ledger in 0.1 seconds

## Intent

On the real ledger (about 500 cards, 15,000 events) `moai card next` took 6.5 s and `moai card list` 1.7 s, so every lane round and every leader check waited on them. The measured cause is the `event` table: per-card lookups by `(card_id, kind)` have no index and scan a 7 MB table, the read connection's 2 MB page cache is smaller than that table, and `card next` spawns one git process per card head. The answers stay as they are; only the way they are found changes.

## Requirements

- REQ-1: When a ledger is created or opened for writing, the ledger shall hold the index `event_card_kind` on `event(card_id, kind, event_id)` without changing `user_version`.
- REQ-2 (REQ-1): IF a ledger file made by an earlier build has no `event_card_kind`, THEN the ledger shall create the index at the next write open and leave every event row unchanged.
- REQ-9 (REQ-2): IF a read-only open finds no `event_card_kind`, THEN the ledger shall write nothing and answer through the slower path.
- REQ-3: When a read connection is opened, the ledger shall give it a page cache large enough to hold the whole database of a ledger with 15,000 events.
- REQ-4: When `card next` checks which card heads the integration branch already contains, the command shall ask git once for all distinct heads instead of once per head.
- REQ-5 (REQ-4): IF git cannot answer the batched question, THEN the command shall keep answering with `landing.checked` false and the reason, as for a per-head failure.
- REQ-6 (REQ-3): IF the page cache is smaller than the database, THEN the ledger shall still answer correctly and only more slowly.
- REQ-7: When `card next` and `card list` run on the same ledger before and after this change, their answers shall be equal except for the time-dependent fields.
- REQ-8 (REQ-7): IF the ledger has no `event_card_kind`, THEN the command shall give the same answer as with the index.
- REQ-10: When `card next` runs, the command shall ask git for the repository identity of one path and HOME once and reuse an `ok` answer until it returns.
- REQ-11 (REQ-10): IF the identity answer is refused or not `ok`, THEN the command shall keep nothing and ask again, so that every refusal and failure answers as before.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9): a new ledger has `event_card_kind`; an earlier-shaped ledger gets it at write open with `user_version` and all event rows unchanged, a read-only open writes nothing, and the lookup plan uses the index · verify: `cargo test --lib ledger::migrate`
- AC-2 (REQ-3, REQ-4): on a temporary ledger of 500 cards, 15,000 events and 9,000 notes, `card next` and `card list` each answer within 0.3 s and the number of git processes of `card next` does not grow with the number of cards · verify: `cargo test --test card card_read_speed`
- AC-3 (REQ-4, REQ-5): the batched head check gives the same verdict as the per-head check for a head inside the tip, outside it, one that is no commit and a repeated head · verify: `cargo test --test card card_read_speed`
- AC-4 (REQ-6, REQ-7, REQ-8): the answer of `card next` and `card list` on a frozen copy of the real ledger is equal before and after, with `now` and lease fields removed · verify: the golden comparison recorded in the card's READY note
- AC-5 (REQ-10): one `card next` run asks git for the repository identity once · verify: `cargo test --test card card_next_asks_for_the_repository_identity_once`
- AC-6 (REQ-10, REQ-11): an `ok` answer is reused only for the same path and HOME and only until the reuse scope ends; a refusal and an answer that is not `ok` are asked again · verify: `cargo test --lib identity::tests::reuse`

## Scope

- write: `src/ledger/**`
- write: `src/card/landed.rs`
- write: `src/card/wip_limit.rs` (one project name instead of two identity reads)
- write: `src/card/work.rs` (the `card next` entry reuses one identity answer)
- write: `src/identity.rs` (reuse of an `ok` answer for the length of one command)
- write: `tests/card/card_read_speed.rs`
- write: `tests/card/main.rs` (one `mod` line, by line lease)

## Plan

1. Measure which part is slow (event scans, cache, git processes) on a frozen copy of the real ledger.
2. Add the `event_card_kind` index to the new-ledger DDL and to the write open, with the migration test.
3. Raise the read connection's page cache.
4. Check the heads of all cards with two git commands.
5. Time the new build on the same copy and compare the answers.
6. Ask for the repository identity once in `card next` and time it again on the same copy.
