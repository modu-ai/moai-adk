---
id: SPEC-20261124-card-next-skips-no-criteria
title: "card next 가 기준 없는 ready 카드를 건너뛰어요"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 7811 (2026-11-23T19:11:44Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7810 에 대한 답): (A) 판 1 승인, 날짜 id 로 옮긴 것과 D-3, 쓰기 범위에 src/card/dispatch.rs 포함; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요"
  digest: "sha256:1b64265497831c37d293db533935d63ac3558f750151f764bc826940a7805989"
---

# card next 가 기준 없는 ready 카드를 건너뛰어요

## Interview

- Q1: Revise SPEC-20261123-card-needs-criteria or open a small new SPEC? -> A: A small new SPEC. The earlier SPEC is approved and landed (its digest is recorded); this one changes only the read side (`card next`) and leaves its refusal in `card ready` and `card take` as it is.
- Q2: Which cards does `next` skip, and with what word? -> A: A `ready` card nobody holds whose criterion rows are all `withdrawn`, or that has none, goes to `skipped` with reason `no_criteria`; the counting rule is the one `card take` uses (`status <> 'withdrawn'`, D-1 of SPEC-20261123-card-needs-criteria).
- Q3: Does the untracked `.moai-cli/state/github/` folder in lane worktrees (AUDIT #419 N4) belong in this change? -> A: Only the finding does: the command or test that writes it is named in a note on the card. A fix is made here only when it lies inside this scope; otherwise the note names the owner (card criterion 2).

## Intent

`card take` refuses a `ready` card that has no criterion that is not withdrawn (SPEC-20261123-card-needs-criteria), but `card next` still offers such a card. That breaks the invariant that `next` never offers what `take` would refuse. On the real ledger a `ready` p2 card with no criterion is listed under `ready`, so a lane that picks by itself takes the first card `next` names and is refused again and again. This SPEC makes `next` skip that card and say why.

읽는 사람: `moai card next` 로 카드를 고르는 레인과 리더.

## Non-goals

- Changing `card ready`, `card take` or the counting rule of SPEC-20261123-card-needs-criteria.
- Closing or repairing the cards that already have no criterion; adding a criterion to a card is the leader's job (`moai card criterion-add`).
- A fix for the untracked `.moai-cli/state/github/` folder outside this write scope (Q3).

## Requirements

- REQ-1: When `moai card next` finds a `ready` card that nobody holds and that has at least one criterion row whose status is not `withdrawn`, the tool shall list it as it does today.
- REQ-2 (REQ-1): IF the `ready` card nobody holds has no criterion row whose status is not `withdrawn`, THEN the tool shall list it under `skipped` with reason `no_criteria` instead of under `ready` or in `next`.
- REQ-3: The tool shall count a criterion row for `card next` by the rule `card take` uses, so a card `take` refuses for want of a criterion is a card `next` skips.
- REQ-4: The tool shall give a held card, a `review` card and a card with an empty scope the reason it gave before this change.
- REQ-5: The tool shall write nothing to the ledger when `card next` skips a card.

## Acceptance

- AC-1 (REQ-2, REQ-3, REQ-5): `moai card next` on a ledger with a `ready` card of zero criterion rows and a `ready` card whose only criterion is `withdrawn` lists both under `skipped` with reason `no_criteria`, names neither in `ready` or `next`, and `card take` refuses both; the ledger bytes are the same before and after · verify: `cargo nextest run --locked --offline --test card_needs_criteria_cli`
- AC-2 (REQ-1, REQ-3): a `ready` card with exactly one `open` criterion is listed under `ready` and named in `next`, and `card take` takes it; a card with a `met` criterion likewise; adding one criterion to a skipped card moves it from `skipped` to `ready` · verify: `cargo nextest run --locked --offline --test card_needs_criteria_cli`
- AC-3 (REQ-4): a held card, a `review` card with no holder and a card with an empty scope keep their reasons `held`, `review_no_holder` and `empty_scope` whether or not the card has a criterion · verify: `cargo nextest run --locked --offline --test card_needs_criteria_cli --test card_work_cli --test ledger_workflow`
- AC-4 (REQ-2): the comment above `done_is_refused_at_every_edge_until_the_card_can_close` in `tests/card_work_cli.rs` says what the fixture really creates (one open criterion, because the helper adds at least one), and the test still passes · verify: `cargo nextest run --locked --offline --test card_work_cli`

## Scope

- write: `.moai/specs/SPEC-20261124-card-next-skips-no-criteria/spec.md`, `src/ledger/workflow.rs`, `src/card/dispatch.rs`, `src/card/work.rs`, `src/ops/card.rs`, `tests/card_needs_criteria_cli.rs`, `tests/card_work_cli.rs`, `tests/ledger_workflow.rs`, `docs/design-review/appendix/operations/card.md`
- the `src/card/dispatch.rs` path is a scope change to the earlier approval (leader note 7804): one call-through so `take` and `next` share one counting function (D-3); the rest of that file is read-only
- read-only: `src/ledger/reservation.rs`
- forbidden: `CLAUDE.md`, the ledger file, `src/ledger/migrate.rs`

## Plan

1. RED: in `tests/card_needs_criteria_cli.rs` the cases of AC-1 and AC-2 (0 criteria, all withdrawn, 1 open, 1 met, a criterion added later), each also asking `card take` so `next` and `take` cannot drift apart; AC-1 fails today because `next` lists the card.
2. `src/ledger/workflow.rs`: one function `counted_criteria(conn, card_id)` that counts the criterion rows whose status is not `withdrawn`; a new `SkipReason::NoCriteria` with code `no_criteria`; `next_filled` calls the function for a card that is `ready`, unheld and has no other skip reason (a held, `review` or empty-scope card keeps its reason, REQ-4), and skips the card when the count is 0. `src/card/dispatch.rs`: `no_criteria_refusal` calls the same function instead of its own query. `src/card/work.rs` renders the new reason like the ones without extra fields.
3. Documentation: the `skipped` reason lists in `src/ops/card.rs` help and `docs/design-review/appendix/operations/card.md`; the comment of AC-4.
4. N4: search the sources and tests for what writes `.moai-cli/state/github/` and record the answer in a note on the card.
5. Guard mutant, `moai bundle check`, then `moai check affected` once at READY.

## Risks

- `src/card/dispatch.rs` is the earlier card's file; the change there is one call replacing one query, and the existing `tests/card_needs_criteria_cli.rs` cases of `ready` and `take` pin its behavior. The AC-1 and AC-2 tests ask `next` and `take` about the same card as well.
- A new reason code reaches every reader of `skipped[]`; the tests and documents that list the reasons are the ones named in Scope.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A new small SPEC instead of a revision of SPEC-20261123-card-needs-criteria | the dispatch for this card allows either; the earlier SPEC is approved and landed, and this one changes only `card next` | fold it into a revision |
| D-2 | The reason is `no_criteria`, a new `SkipReason::NoCriteria` | card criterion 1 names the word | rename the code |
| D-3 | `take` and `next` count criteria with one function, `counted_criteria` in `src/ledger/workflow.rs`, and `src/card/dispatch.rs` calls it | leader note 7804: "if possible use the same function as take, so the two cannot drift apart"; this adds `src/card/dispatch.rs` to the write scope, which asks for the leader's approval again | keep two queries and the tests that compare them |
| D-4 | The SPEC id is dated 20261124, the date of the branch | the preflight hook reads the SPEC from the branch name `spec/card-next-skips-no-criteria-20261124`; `moai spec new` dated the first draft 20261123 (UTC clock) | rename branch and SPEC together |
