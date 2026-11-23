---
id: SPEC-20261123-card-needs-criteria
title: "기준 행이 없는 카드는 ready·take 로 못 가게"
status: draft
revision: 1
created: 2026-11-23
---

# 기준 행이 없는 카드는 ready·take 로 못 가게

## Intent

기준(card_ac) 행이 0개인 카드는 착지해도 닫을 수 없어요. 원장 트리거 `card_done_guard` 가 "충족된 기준이 하나도 없으면 done 으로 못 바꿔요"라고 막고 `moai card finish` 도 같은 이유로 거절해요. 그래서 착지했거나 철회된 카드 다섯 장이 열린 채 남았고, 같은 종류가 세 번 반복됐어요(재발 규칙, 리더 결정 2026-11-23T17:42Z 의 (a)). 이 SPEC 은 입구를 막아요: `moai card ready` 와 `moai card take` 가 기준 행이 0개인 카드를 거절하고, 거절문이 `moai card criterion-add` 를 알려 줘요.

읽는 사람: 카드를 열고 ready 로 올리는 리더, 카드를 잡는 레인.

## Non-goals

- 원장 파일, 트리거(`card_done_guard`), 마이그레이션(`src/ledger/migrate.rs`) 변경. 이미 있는 기준 0개 카드(착지했지만 열린 다섯 장)를 닫는 길은 이 SPEC 이 아니에요. 그건 guarded 라서 사용자 선택을 기다려요.
- `moai card add` 가 기준 없이 만들어지는 것을 막는 일. draft 카드는 기준을 나중에 더할 수 있어요.
- 원장 라이브러리(`src/ledger/**`)의 `reserve` 같은 함수 자체를 거절시키는 일. 막는 곳은 `moai card ready` 와 `moai card take` 명령 둘이에요.
- 모두 철회된 기준만 남은 카드를 따로 다루는 일(D-1).

## Requirements

- REQ-1: When `moai card ready` names a `draft` or `hold` card that has at least one criterion row, the tool shall make the card `ready` as it does today.
- REQ-2 (REQ-1): IF the card has no criterion row, THEN the tool shall refuse with kind `invalid_input` and write nothing to the ledger.
- REQ-3 (REQ-2): The refusal message shall name `moai card criterion-add <card> --ac <text>`.
- REQ-4: When `moai card take` names a `ready` card that has at least one criterion row, the tool shall reserve the card as it does today.
- REQ-5 (REQ-4): IF the `ready` card has no criterion row, THEN the tool shall refuse as REQ-2 and REQ-3 say, and reserve nothing.
- REQ-6: The tool shall give `moai card ready` and `moai card take` on a missing card, a card in a state they refuse or a held card the refusal it gave before this change, whether or not the card has criterion rows.
- REQ-7: The tool shall count a criterion row of any status (`open`, `met` or `withdrawn`) as a criterion row.

## Acceptance

- AC-1 (REQ-2, REQ-3, REQ-7): `moai card ready` on a `draft` card with no criterion row exits 1 with kind `invalid_input`, the message names `criterion-add`, and the card's state, `rev` and event count are the same before and after; after one `criterion-add` the same command succeeds · verify: `cargo nextest run --locked --offline --test card_needs_criteria_cli`
- AC-2 (REQ-5, REQ-3): `moai card take` on a `ready` card with no criterion row exits 1 with kind `invalid_input`, names `criterion-add`, and leaves no reservation and the card `ready`; after one `criterion-add` the same command takes the card · verify: `cargo nextest run --locked --offline --test card_needs_criteria_cli`
- AC-3 (REQ-1, REQ-4, REQ-6): a card with at least one criterion keeps its `ready` and `take` behavior and every earlier refusal (missing card, state, live reservation, scope conflict): the existing tests of the broken fixtures pass once their cards carry a criterion · verify: `cargo nextest run --locked --offline --test branch_gc_cli --test card_criterion_add_cli --test card_dispatch_cli --test card_done_mflow_cli --test card_generated_scope --test card_reclaim_cli --test card_reclaim_early_cli --test card_roles_cli --test card_spec_cli --test card_take_scope_cli --test card_verdict_cli --test card_withdraw_cli --test card_work_cli --test ledger_readonly_cli --test worktree_cleanup_triggers --test worktree_gc_slots` and `cargo nextest run --locked --offline --lib the_recheck_reads_the_claims_of_the_given_home`

## Scope

- write: `.moai/specs/SPEC-20261123-card-needs-criteria/spec.md`, `src/card/dispatch.rs`, `tests/card_needs_criteria_cli.rs`
- write (added by decision D-2, only after the leader approves): `src/ops/card.rs` (one help sentence for `ready` and `take`), `docs/claude-md-rule-provenance.md` (one cause line), the sixteen fixture files named in AC-3 (`tests/<name>.rs`, only the lines that create a card without `--ac`) and the test module of `src/worktree_gc.rs` (the one test that readies and takes a card)
- read-only: `src/ledger/**` (the trigger and the migration stay as they are), `src/card/command.rs`
- forbidden: `CLAUDE.md`, the ledger file, `src/ledger/migrate.rs`

## Plan

1. RED: `tests/card_needs_criteria_cli.rs` pins AC-1 and AC-2 in a temporary HOME; both fail today because `ready` and `take` accept a card without criteria.
2. `src/card/dispatch.rs`: one helper reads the criterion rows of the card; `ready` calls it inside its write transaction after the state and reservation checks, `take` calls it before it reserves a `ready` card. The rows are never deleted (trigger `card_ac_no_delete`), so a card that has a row when it is read cannot lose it before the reservation.
3. Fixtures: each broken fixture card gets `--ac` where the test creates it (AC-3).
4. Help sentences, the rule-provenance cause line, `moai bundle check`, then the guard mutants and `moai check affected` once at READY.

## Risks

- Measured on 2026-11-23 with the check applied on the head: 169 tests in 16 files under `tests/` and 1 test in `src/lib` (`worktree_gc::tests::the_recheck_reads_the_claims_of_the_given_home`) fail only because their cards have no criterion; the same 16 files ran 230 tests, 230 passed, without the check (`cargo nextest run --locked --offline --no-fail-fast`, 69 test files that spell `"ready"` or `"take"` plus the whole `--lib`: 1181 tests run, 169 failed; `--lib` alone 1771 run, 1 failed). The 16 files hold about 72 `"add"` call sites; most are one fixture helper per file. A test file that creates and readies cards through a spelling these greps miss would show up when the whole suite runs.
- `src/worktree_gc.rs` is another module's file; only its `#[cfg(test)]` fixture changes, and D-2 asks for it.
- A caller that uses the ledger library directly (`reserve`, a SQL insert in a test) is not guarded; the doc says so in D-3.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A card whose criteria are all `withdrawn` still has criterion rows and passes (any row counts) | the recurring case is a card with no row at all; a card with only withdrawn rows was never seen (rule: no design for a rare case) | count only non-withdrawn rows in the helper |
| D-2 | The write scope also names the sixteen fixture files, the test module of `src/worktree_gc.rs`, `src/ops/card.rs` (help) and `docs/claude-md-rule-provenance.md` (cause line) | the card allows adding criteria to fixtures that break; the recurring-problem rule asks for the cause in the rule provenance; needs the leader's approval as a scope change | drop the paths and their edits; the check then breaks 169 tests |
| D-3 | The refusal lives in the `card ready` and `card take` commands (`src/card/dispatch.rs`), not in `ledger::reservation::reserve` | the card names the two commands; a check in `reserve` would also break `tests/ledger_reservation.rs` and the other library fixtures, which build cards by SQL | move the call into `reserve` |
