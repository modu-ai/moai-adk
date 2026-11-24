---
id: SPEC-20261124-card-next-skip-reason
title: "card next 가 이미 착지한 기준 없는 카드를 already_landed 로 말해요"
status: approved
revision: 1
created: 2026-11-24
approval:
  by: leader
  at: 2026-11-23
  basis: "리더 결정 노트 (note:430:7926febfe9cfd528db74755ddc7ae960, 2026-11-23T21:42:33Z, 사용자 위임 2026-11-23T13:57Z, 리더 판단, 사용자 직접 선택 아님, 노트 7915 에 대한 답): D-1 선택지 A (already_landed 가 이김), 판 1 승인, is_landable_skip 이름 바꾸기 좋음, 힌트 문장에 criterion-add → met → finish 순서; 리더가 노트에 적었고 레인은 그 노트를 옮겨 적었어요"
  digest: "sha256:ef41fce4b08e1c85ede4fcf47c0174ed4d6de2050d67b62917403d7673027882"
---

# card next 가 이미 착지한 기준 없는 카드를 already_landed 로 말해요

## Interview

- Q1: A `ready` card has no counted criterion and its head is already on the integration branch: which skip reason does `card next` give? -> A: the leader's default is `already_landed` (dispatch of C-430; to be confirmed by the leader's note, not by this lane). The proof (`via`, `head` or `pr`) tells the leader the work is in; `no_criteria` says only what `take` would refuse, which is no news for a card nobody will take.
- Q2: Does `card finish` close such a card from the landing hint? -> A: No. `finish` refuses a card with no met criterion (`src/ledger/workflow.rs:1488`), so the hint means "landed: `card criterion-add`, mark it, then `card finish`", not "finish now". The hint is still the more useful word, because it names the head.
- Q3: Where is the choice written? -> A: here. `docs/design-review/appendix/operations/card.md:22` is held by C-428, so the sentence for it is given under Scope and is added after C-428 lands (criterion 2 of the card).

## Intent

Since SPEC-20261124-card-next-skips-no-criteria, `next_filled` (`src/ledger/workflow.rs`) skips a `ready` card with no counted criterion as `no_criteria` before `src/card/landed.rs` asks whether the card landed. `landed::candidates` hands the proof lookup only the cards still `ready` and the cards skipped for a scope reason (`is_scope_skip`: `empty_scope`, `scope_overlap`), so a criteria-less card whose head is already on the integration branch gets `no_criteria` and no landing hint. This SPEC decides that the landing proof wins and says it in one place.

읽는 사람: `moai card next` 로 카드를 고르는 레인과 리더, 이 규칙을 문서에 옮기는 C-428.

## Non-goals

- Changing `card take`, `card ready`, `card finish` or the counting rule `counted_criteria`.
- Closing the card or adding a criterion to it; the landing hint only reports.
- Changing the reason of a held card, a `review` card or a card that is not `ready`.

## Requirements

- REQ-1: When `moai card next` finds a `ready` card nobody holds that has no counted criterion and whose landing is proven, the tool shall list it under `skipped` with reason `already_landed` and the proof, and not under `ready` or with reason `no_criteria`.
- REQ-2 (REQ-1): IF the `ready` card has no counted criterion and no landing proof, THEN the tool shall list it under `skipped` with reason `no_criteria` as before.
- REQ-3: The tool shall give a held card, a `review` card with no holder, a card with an empty scope and a card whose scope overlaps another reservation the reason it gave before this change, and a card with a counted criterion the answer it gave before.
- REQ-4: The tool shall write nothing to the ledger when `card next` skips a card.

## Acceptance

- AC-1 (REQ-1, REQ-4): on a temporary ledger and git project, a `ready` card with a landing event for a head inside the integration branch and all criteria withdrawn is listed under `skipped` with reason `already_landed`, `via` and `head`, is named neither in `ready` nor in `next`, and the ledger bytes are the same before and after · verify: `cargo nextest run --locked --offline --test card_next_skip_reason_cli`
- AC-2 (REQ-2, REQ-3): a criteria-less `ready` card with no landing keeps reason `no_criteria`, and one whose recorded head is outside the integration branch keeps it too; a card with one open criterion and a landed head is `already_landed` as before; a held card with no criterion keeps `held` · verify: `cargo nextest run --locked --offline --test card_next_skip_reason_cli --test card_finish_cli --test card_needs_criteria_cli`

## Scope

- write: `.moai/specs/SPEC-20261124-card-next-skip-reason/spec.md`, `src/card/landed.rs`, `tests/card_next_skip_reason_cli.rs`
- read-only: `src/ledger/workflow.rs`, `src/card/work.rs`, `src/ops/card.rs`, `tests/card_finish_cli.rs`, `tests/card_needs_criteria_cli.rs`
- held by C-428, not written here: `docs/design-review/appendix/operations/card.md` (the row of `card next` at line 22). The sentence for it: "A `ready` card with no counted criterion is skipped as `no_criteria`, except that a landing proof makes it `already_landed` (the proof wins; `card finish` still needs one criterion marked met)."
- forbidden: `CLAUDE.md`, the ledger file

## Plan

1. RED: `tests/card_next_skip_reason_cli.rs` builds a temporary HOME with a git project (an integration branch one commit past the card's head), makes a card with one criterion, records its landing and withdraws the criterion; `card next` must say `already_landed` and today says `no_criteria`.
2. `src/card/landed.rs`: the filter that picks the skips a landing proof can remove (`is_scope_skip`, used by `candidates` and `take_landed`) also accepts `SkipReason::NoCriteria`; the function is renamed `is_landable_skip` so its name says what it asks. `src/ledger/workflow.rs` stays as it is.
3. Guard mutant: take `NoCriteria` out of the filter again; the new test must fail with exit 100.
4. `moai check affected` once at READY.

## Risks

- Both `candidates` and `take_landed` share one filter, so a card is looked up and taken in one place; a change to only one of them would leave the card in `skipped` twice or never. The test asserts the card appears exactly once.
- `already_landed` on a criteria-less card does not mean `card finish` works; Q2 and the sentence for the operations doc say so.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `already_landed` wins over `no_criteria` | leader's default in the dispatch of C-430: the landing hint is more useful; needs the leader's approval note (LANE-BLOCKED kind=decision), never recorded by the lane | make `is_landable_skip` ignore `NoCriteria` again and flip the test |
| D-2 | The filter is renamed `is_landable_skip` | the old name `is_scope_skip` would lie once `no_criteria` is in it; both callers are in `src/card/landed.rs` | rename back |
| D-3 | The operations document sentence waits for C-428 | C-428 holds `docs/design-review/appendix/operations/card.md`; this card's write scope does not name it | add the path to this card's scope after C-428 lands |

## Result

Measured 2026-11-24 (UTC) on macOS in the worktree `.claude/worktrees/card-next-skip-reason`; code head `4beafe0989773d695f53646764e3e8b865c6c90c` (the change itself is `f788d141b7`, the test repair is the last commit) on base `fb97c6af9a` (the READY head adds this Result only). Commands ran as `RUSTC_WRAPPER=sccache moai heavy-slot-acquire --name card-next-skip-reason --max 3 -- cargo nextest run --locked --offline --no-fail-fast ...`.
- AC-1 met: `a_landed_card_with_no_counted_criterion_is_already_landed_and_not_no_criteria` (the card's only criterion is withdrawn and its landed head is on `integ`; `skipped` lists it once as `already_landed` with `via` and `head`, `ready` and `next` do not name it, the home snapshot is the same before and after `card next`). RED on the first valid run (commit `d3d655e08f`): the entry said `no_criteria`, exit 100. After the change it passes.
- AC-2 met: `a_card_with_no_counted_criterion_and_no_landing_keeps_no_criteria` (no landing, and a head on a branch outside `integ`), `the_other_skip_reasons_stay_as_they_were` (one open criterion with a landed head is `already_landed`; a held card whose criterion is withdrawn is `held`). Run `--test card_next_skip_reason_cli --test card_finish_cli --test card_needs_criteria_cli`: 58 run, 58 passed.
- A first RED run failed at the snapshot assertion before it reached the reason, because the helper wrote the profile file again after the snapshot; the helper no longer does, and the assertion uses `assert!` so a failure does not print the ledger bytes. The RED commit is the second run.
- guard mutant, run on the committed head and reverted with `git checkout` (`git status --short` empty after): `SkipReason::NoCriteria` taken out of `is_landable_skip`, exit 100 (1 failed: the AC-1 test; the two control tests still pass).
- Repair: the first `moai check affected fb97c6af9a..c0f177168e` ended `result=FAIL` with `test=100` and one failing test of 4228, `state_folder_literal_scan the_repository_tests_stay_within_the_state_folder_baseline`: the new test file spelled a state folder name (one `.moai-cli` for the profile path, one `.moai` in the header comment). The path now comes from `moai_cli::state_root::path` and the comment says "the real home state"; `--test card_next_skip_reason_cli --test state_folder_literal_scan` ran 9, 9 passed, and `cargo fmt --check` exit 0. The receipt is the second run on the final head.
- The change is one filter in `src/card/landed.rs` (`is_scope_skip` renamed `is_landable_skip`, accepts `NoCriteria`); `candidates` and `take_landed` share it, so the card is looked up and taken in one place. `src/ledger/workflow.rs` is untouched.
- Gap (D-3): the sentence for `docs/design-review/appendix/operations/card.md:22` is not written; C-428 holds that file. Card criterion 2 stays open until C-428 lands. The `card next` help text in `src/ops/card.rs` says `no_criteria` is "given only when no other reason applies", which still holds, so it is left as it is.
- `moai check affected` runs once on the READY head and its receipt is in the READY note.
