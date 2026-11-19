---
id: SPEC-20261119-card-scope-widen-followup
title: "card scope widening: line separators, case variants, a generation to widen at, a true spec status"
status: draft
revision: 1
created: 2026-11-19
---

# card scope widening: line separators, case variants, a generation to widen at, a true spec status

## Intent

Card C-129 (follow-up of the C-94 audit, PASS note 2859, head 15b333fe73; the card was made from the audit's non-blocking notes N1, N2, N3 and N6; N4 and N5 are covered by SPEC-20261119-card-scope-widen-tests). Four small gaps of `card scope` on a held card, measured at the local integration tip e84d7c46ce:

- N3: `src/card/dispatch.rs` `checked_reason` and `checked_scope` refuse control characters only (`char::is_control`, Unicode category Cc). U+2028 and U+2029 are not Cc, so a reason or a scope entry that holds one is stored. `str::trim` also strips them at the ends, so an edge U+2028 would silently vanish while an inner one stays. `--reason` is stored as typed, with its leading and trailing blanks (a scope entry is trimmed already by `input::scope_entry`).
- N2: `reservation::widen_in_tx` compares added patterns with the reservation's own patterns exactly. `Apps/Web/` next to `apps/web/` is added as a new pattern, although `src/scope.rs` folds case when it judges overlap (`to_lowercase`), so the two mean the same files to every other reader.
- N1: `card scope` on a held card takes no `--generation`. A lane that read the card at generation 2 and a leader that widens it after a reclaim (generation 3) act on different screens without knowing.
- N6: SPEC-20261119-card-scope-widen says `status: draft` while its Result is filled.

This SPEC is built in two slices. Slice A (N2, N3, N6) touches `src/card/dispatch.rs`, `src/ledger/reservation.rs`, tests and documents only. Slice B (N1) adds a `--generation` option, which is a change of `src/ops/card.rs` (the clap arguments and the help); that file is reserved by card C-125, so slice B waits until it is free or the leader allows the exact lines. Its requirement and acceptance line (REQ: `card scope --generation <n>` widens only a reservation at generation `n`, otherwise `conflict` naming the generation it has; AC with its test `a_stale_generation_is_refused`) are added in revision 2 together with the code, because a recorded verify command that names a test that does not exist yet fails `ac_command_lint`.

## Non-goals

- Who may widen (the leader, by policy; SPEC-20261119-card-scope-widen D-2): `--by` stays free text.
- Changing which characters `card add` accepts in titles or criteria (`src/card/input.rs` is reserved by card C-122): only scope entries and the `--reason` are judged here, in `dispatch.rs`, before `input::validate_scope` is called.
- Any change of what `find_overlap` judges between two cards (that is SPEC-20261119-card-generated-scope and `src/scope.rs`).
- Narrowing a held card or deleting a pattern: unchanged, refused.

## Requirements

- REQ-1: IF a `--scope` entry or the `--reason` of `card scope`, `card add`, `card take` or `card reclaim` holds a line separator (U+2028 or U+2029), THEN the command shall refuse it as `invalid_input`, before any trimming, and write nothing.
- REQ-2: When `--reason` is accepted, the command shall keep it without its leading and trailing blanks, in the answer and in the event.
- REQ-3: IF an added pattern of a widening differs from a pattern the reservation has, or from another added pattern, only by letter case, THEN `card scope` shall refuse it as `invalid_input`, name both spellings, and write nothing.
- REQ-4: The status of SPEC-20261119-card-scope-widen shall be `approved` and its Scope shall list `tests/card_dispatch_cli.rs`.

## Acceptance

- AC-1 (REQ-1): a reason or a scope entry with U+2028 or U+2029 inside it or at either end is `invalid_input` for `card scope` on a held card (reason, added pattern), `card add --scope`, `card take --scope` and `card reclaim --reason`, and every table row and event is unchanged.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli line_separators_are_refused_in_a_reason_and_a_scope_entry`
- AC-2 (REQ-2): `card scope --reason "  why  "` on a held card answers and records `why`.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli a_reason_is_kept_without_its_surrounding_blanks`
- AC-3 (REQ-3): `Apps/Web/` next to a held `apps/web/`, two case variants in one call, and a variant of a pattern added earlier are `invalid_input` naming both spellings with nothing written; a different path with the same letters in another folder is still added.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli a_case_variant_is_not_a_new_pattern`
- AC-4 (REQ-4): the SPEC file says `status: approved` and its Scope names `tests/card_dispatch_cli.rs`.
  verify: `cargo test --locked --offline --test spec_cli the_repository_specs_pass_check`
- AC-5 (REQ-1 to REQ-3): neighbours.
  verify: `cargo test --locked --offline --test card_scope_widen_cli`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_reclaim_cli`; `cargo test --locked --offline --test ledger_reservation`

## Scope

- write: `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `tests/card_scope_widen_followup_cli.rs`, `docs/design-review/appendix/operations/card.md` (the `card scope` row only), `.moai/specs/SPEC-20261119-card-scope-widen-followup/**`, `.moai/specs/SPEC-20261119-card-scope-widen/spec.md` (status line and Scope line only); slice B adds `src/ops/card.rs` (the `--generation` argument and the help only) and `src/card/command.rs` if the parser needs it
- read-only: `src/card/input.rs`, `src/scope.rs`, `tests/card_scope_widen_cli.rs`, `tests/common/**`
- forbidden: the real HOME and ledger, the network, a change of table definitions or triggers, `skills/moai/**`

## Plan

1. Slice A, tests first: `tests/card_scope_widen_followup_cli.rs` with AC-1 to AC-3; they fail before the code (U+2028 is stored, the variant is added, the reason keeps its blanks).
2. `dispatch.rs`: a line-separator check on the raw entries and the raw reason before trimming; `checked_reason` returns the trimmed text. `reservation.rs`: `widen_in_tx` folds added patterns and refuses a variant.
3. The `card scope` row of the operations document names the three rules; the C-94 SPEC gets its status and Scope line.
4. Mutants for each rule; the SPEC Result.
5. Slice B, when `src/ops/card.rs` is free: revision 2 with the requirement and AC for `--generation`, its test, the option, the help and the document.

## Risks

- The reason is trimmed before its length is counted, so a reason of 500 characters plus blanks is accepted; the stored text is still at most 500 characters.
- Folding case is `str::to_lowercase`, the same fold as `src/scope.rs`; a reader that folds differently (Turkish dotless i) is not covered, as there.
- A scope entry that had an edge U+2028 and used to be accepted (trimmed to a clean entry) is refused now. Nothing in the repository or in the ledger tests writes one.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A variant is refused, not dropped silently | a dropped entry would hide that the caller meant a different spelling; the refusal names both | Drop it and say so in `added` |
| D-2 | Line separators are checked on the raw text, before `trim` | `trim` strips U+2028 and U+2029 at the ends, so a check after it sees only the inner ones | Check after trimming |
| D-3 | The check lives in `dispatch.rs`, not in `input.rs` | `src/card/input.rs` is reserved by C-122; `checked_scope` runs for `add`, `scope` and `take` | Move it into `input::scope_entry` |
| D-4 | Slice B (`--generation`) waits for `src/ops/card.rs` | C-125 holds the file; narrow slices land sooner | Take both files now with the leader's allowance |

## Result

Slice A (N2, N3, N6), built from the local integration tip e84d7c46ce. RED first: commit e93897b421 (4 tests, all failing for the intended reason: a separator was accepted, a variant was added, the reason kept its blanks, reclaim reached the lease check). Green: commit b1bd9ea3d4. Slice B (`--generation`, N1) is not built: `src/ops/card.rs` is reserved by C-125; revision 2 will add its requirement, acceptance line and test.

- AC-1: met - `line_separators_are_refused_in_a_reason_and_a_scope_entry` (U+2028 and U+2029, inner, leading and trailing, in a reason and in an added pattern of a held card: `invalid_input`; `card add --scope` and `card take --scope` too; the reservation rows are as before and no `card.scope` event exists) and `a_reclaim_reason_with_a_line_separator_is_invalid_input`. Mutants M1 (scope check off) and M2 (reason check off) killed.
- AC-2: met - `a_reason_is_kept_without_its_surrounding_blanks` (spaces, U+3000 and NBSP around the reason: the answer and the event hold the bare text). Mutant M3 (stored untrimmed) killed.
- AC-3: met - `a_case_variant_is_not_a_new_pattern` (`Apps/Web/` next to `apps/web/`, `libs/Util/` with `libs/util/` in one call, and `Other/Apps/Web/` after `other/apps/web/` was added: `invalid_input` naming both spellings, every row unchanged; `other/apps/web/` is added). Mutants M4 (no folding) and M5 (no check among added patterns) killed.
- AC-4: met - `spec_cli the_repository_specs_pass_check` (see the checks below); `.moai/specs/SPEC-20261119-card-scope-widen/spec.md` says `status: approved` and its Scope already names `tests/card_dispatch_cli.rs` (added by the audit repair b205de81ea).
- AC-5: met - `card_scope_widen_followup_cli` 4, `card_scope_widen_cli` 9, `card_dispatch_cli` 25, `card_reclaim_cli` 16, `card_work_cli` 25, `ledger_reservation` 28 and the library tests of `reservation` 14 all passed.

Not changed: `src/card/input.rs` (reserved by C-122; the separator check sits in `dispatch.rs`, D-2 and D-3), `src/ops/card.rs`, `src/card/command.rs`. The help text of `card scope` does not name the three rules (it lives in `src/ops/card.rs`); the operations document does.
