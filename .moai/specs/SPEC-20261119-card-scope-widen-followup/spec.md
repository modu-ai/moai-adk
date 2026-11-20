---
id: SPEC-20261119-card-scope-widen-followup
title: "card scope widening: line separators, case variants, a generation to widen at, a true spec status"
status: approved
revision: 2
created: 2026-11-19
approval: {by: user, at: 2026-11-19, basis: "사용자 선택 2026-11-19T21:22:41Z 리더 대화 (선택지 질문, 선택지: 둘 다 승인(선택) / 하나만 / 승인 안 함; note 3582 는 이 SPEC 만 이름을 불러요) — 독립 감사 PASS(C-129 노트 3492) 내용 그대로 승인", digest: "sha256:727bfdf9687f63949eec1efad7a1b3a4a673bc3a1c64580c0f50617abb5fd27e"}
---

# card scope widening: line separators, case variants, a generation to widen at, a true spec status

## Intent

Card C-129 (follow-up of the C-94 audit, PASS note 2859, head 15b333fe73; the card was made from the audit's non-blocking notes N1, N2, N3 and N6; N4 and N5 are covered by SPEC-20261119-card-scope-widen-tests). Four small gaps of `card scope` on a held card, measured at the local integration tip e84d7c46ce:

- N3: `src/card/dispatch.rs` `checked_reason` and `checked_scope` refuse control characters only (`char::is_control`, Unicode category Cc). U+2028 and U+2029 are not Cc, so a reason or a scope entry that holds one is stored. `str::trim` also strips them at the ends, so an edge U+2028 would silently vanish while an inner one stays. `--reason` is stored as typed, with its leading and trailing blanks (a scope entry is trimmed already by `input::scope_entry`).
- N2: `reservation::widen_in_tx` compares added patterns with the reservation's own patterns exactly. `Apps/Web/` next to `apps/web/` is added as a new pattern, although `src/scope.rs` folds case when it judges overlap (`to_lowercase`), so the two mean the same files to every other reader.
- N1: `card scope` on a held card takes no `--generation`. A lane that read the card at generation 2 and a leader that widens it after a reclaim (generation 3) act on different screens without knowing.
- N6: SPEC-20261119-card-scope-widen says `status: draft` while its Result is filled. Measured here: its Scope already lists `tests/card_dispatch_cli.rs` (audit repair b205de81ea), and the status cannot be raised by a lane: an `approved` SPEC needs `approval: {by: user, at, basis, digest}` and an interview record, and `spec check` refuses a bare `status: approved` (`approval_missing`, `interview_missing`, measured on a trial edit). The status stayed `draft` until the user approved this SPEC through the leader (D-5, Result N6).

This SPEC is built in two slices. Slice A (N2, N3, N6) touched `src/card/dispatch.rs`, `src/ledger/reservation.rs`, tests and documents only. Slice B (N1) adds an optional `--generation` option, a change of `src/ops/card.rs` (the clap argument, the help and the call) and of `src/card/dispatch.rs` (the check inside the write transaction). It was built when `src/ops/card.rs` was free again (card C-99 gave it back, C-129 note 3432). Revision 2 adds its requirements REQ-4 and REQ-5, the acceptance lines AC-5 and AC-6, and the tests of audit notes N-1 and N-2 of slice A (a non-ASCII case variant; the reason at 500 and 501 characters with blanks around it).

## Interview

- Q1: 이 SPEC 을 승인하나요(둘 다 승인 / 하나만 / 승인 안 함)? → A: 승인해요(사용자 선택, 리더 대화 2026-11-19T21:22:41Z, 리더 결정 note 3582 of card C-129). 근거는 독립 감사 PASS(note 3492)예요.
- Q2: 같은 선택이 SPEC-20261119-card-scope-widen(C-94)도 승인하나요? → A: 이 기록은 그쪽을 승인하지 않아요. note 3582 는 이 SPEC 하나만 이름을 불러요(선택지 글에는 "둘 다 승인(선택)" 이 있어 사용자가 둘 다 골랐는지는 그 글만으로 알 수 없어요). 그쪽 SPEC 의 상태는 이 커밋이 바꾸지 않아요. 둘 다 승인이 맞다면 그쪽 승인은 리더가 따로 기록해요.
- Q3: 승인한 내용은 무엇인가요? → A: 독립 감사 PASS(note 3492) 때의 요구·수용 기준과, 이 커밋이 바로잡은 범위예요. 범위(Scope)는 사용자가 고른 뒤에 고쳤어요(C-94 SPEC 파일 추가, `src/card/command.rs` 를 forbidden 으로 옮김; 근거: 여섯 커밋의 `git show --name-only`). 그 때문에 digest 는 고친 범위로 다시 읽었고, 사용자가 이 고친 범위를 다시 확인하지는 않았어요.

## Non-goals

- Who may widen (the leader, by policy; SPEC-20261119-card-scope-widen D-2): `--by` stays free text.
- Changing which characters `card add` accepts in titles or criteria (`src/card/input.rs` is reserved by card C-122): only scope entries and the `--reason` are judged here, in `dispatch.rs`, before `input::validate_scope` is called.
- Any change of what `find_overlap` judges between two cards (that is SPEC-20261119-card-generated-scope and `src/scope.rs`).
- Narrowing a held card or deleting a pattern: unchanged, refused.
- Making `--generation` required: a caller that has no number to give (a plain widening by the leader) keeps working; the option is an expectation the caller may state (D-6).
- A check inside `reservation::widen_in_tx` (`src/ledger/reservation.rs` is reserved by card C-127): the generation is compared in `dispatch.rs`, in the same transaction, before `widen_in_tx` is called (D-7).

## Requirements

- REQ-1: IF a `--scope` entry or the `--reason` of `card scope`, `card add`, `card take` or `card reclaim` holds a line separator (U+2028 or U+2029), THEN the command shall refuse it as `invalid_input`, before any trimming, and write nothing.
- REQ-2: When `--reason` is accepted, the command shall keep it without its leading and trailing blanks, in the answer and in the event. · failure: none
- REQ-3: IF an added pattern of a widening differs from a pattern the reservation has, or from another added pattern, only by letter case, THEN `card scope` shall refuse it as `invalid_input`, name both spellings, and write nothing.

- REQ-4: When `--generation <n>` is given to `card scope` for a held card, the command shall widen the reservation only if the card's live reservation is at generation `n`.
- REQ-5 (REQ-4): IF `--generation <n>` is given and the card has no live reservation at generation `n`, THEN `card scope` shall refuse it as `conflict`, name the generation the card has (or that it is not held), and write nothing.

## Acceptance

- AC-1 (REQ-1): a reason or a scope entry with U+2028 or U+2029 inside it or at either end is `invalid_input` for `card scope` on a held card (reason, added pattern), `card add --scope`, `card take --scope` and `card reclaim --reason`, and every table row and event is unchanged.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli line_separators_are_refused_in_a_reason_and_a_scope_entry`
- AC-2 (REQ-2): `card scope --reason "  why  "` on a held card answers and records `why`.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli a_reason_is_kept_without_its_surrounding_blanks`
- AC-3 (REQ-3): `Apps/Web/` next to a held `apps/web/`, two case variants in one call, and a variant of a pattern added earlier are `invalid_input` naming both spellings with nothing written; a different path with the same letters in another folder is still added.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli a_case_variant_is_not_a_new_pattern`
- AC-4 (REQ-1…REQ-5): neighbours.
  verify: `cargo test --locked --offline --test card_scope_widen_cli`; `cargo test --locked --offline --test card_dispatch_cli`; `cargo test --locked --offline --test card_reclaim_cli`; `cargo test --locked --offline --test ledger_reservation`

- AC-5 (REQ-4, REQ-5): a card released and taken again is at generation 2; `card scope --generation 1` and `--generation 3` are `conflict` naming `generation 2` and the given number with every row unchanged; `--generation 2` widens (the answer's `generation` is 2); no `--generation` widens as before; a card nobody holds with `--generation 1` is `conflict` (`not held`) with nothing written; a `--generation` that is not a positive number is refused.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli a_stale_generation_is_refused`; `cargo test --locked --offline --test card_scope_widen_followup_cli a_generation_that_is_not_a_positive_number_is_a_usage_error`
- AC-6 (REQ-2, REQ-3): a non-ASCII case variant (`café/` and `CAFÉ/`) is `invalid_input` naming both; a reason of 500 characters with blanks around it is kept as 500 and one of 501 characters with blanks around it is `invalid_input` with nothing written.
  verify: `cargo test --locked --offline --test card_scope_widen_followup_cli a_non_ascii_case_variant_is_not_a_new_pattern`; `cargo test --locked --offline --test card_scope_widen_followup_cli a_reason_of_500_characters_with_blanks_is_kept_and_501_is_refused`

## Scope

- write: `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `tests/card_scope_widen_followup_cli.rs`, `docs/design-review/appendix/operations/card.md` (the card scope row only), `.moai/specs/SPEC-20261119-card-scope-widen-followup/**`, and `.moai/specs/SPEC-20261119-card-scope-widen/spec.md` (the C-94 SPEC: its `status:` line only, set to approved by commit b1bd9ea3d4 and back to draft by commit f837d0a906; the file is `status: draft` before and after, no other line changed); slice B adds `src/ops/card.rs` (the generation argument, its help and the call only) and `tests/card_scope_widen_cli.rs` only if a neighbour needs the new argument
- read-only: `src/card/input.rs`, `src/scope.rs`, `tests/card_scope_widen_cli.rs`, `tests/common/**`
- forbidden: the real HOME and ledger, the network, a change of table definitions or triggers, `skills/moai/**`, `src/card/command.rs` (card C-99 holds it)

## Plan

1. Slice A, tests first: `tests/card_scope_widen_followup_cli.rs` with AC-1 to AC-3; they fail before the code (U+2028 is stored, the variant is added, the reason keeps its blanks).
2. `dispatch.rs`: a line-separator check on the raw entries and the raw reason before trimming; `checked_reason` returns the trimmed text. `reservation.rs`: `widen_in_tx` folds added patterns and refuses a variant.
3. The `card scope` row of the operations document names the three rules.
4. Mutants for each rule; the SPEC Result.
5. Slice B (revision 2): tests first (`a_stale_generation_is_refused` and the notes N-1, N-2), then the option and help in `src/ops/card.rs`, the comparison in `dispatch.rs`, the `card scope` row of the operations document.

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
| D-6 | `--generation` is optional | the audit note N1 asked for an expected value that refuses a stale screen; requiring it would break every plain widening and the tests of C-94 | Make it required for a held card |
| D-7 | The generation is compared in `dispatch.rs` | `src/ledger/reservation.rs` is reserved by C-127; the comparison reads the live reservation in the same write transaction, so a retake between read and write cannot slip through | Move it into `widen_in_tx` |
| D-5 | A lane does not set `status: approved` | the approval is the user's: `spec check` wants `approval.by`, a digest and an interview | The user approved this SPEC (note 3582); the C-94 SPEC is approved separately, if at all, and the leader writes that block |

## Result

Slice A (N2, N3; N6 is a decision for the user, D-5), built from the local integration tip e84d7c46ce. RED first: commit e93897b421 (4 tests, all failing for the intended reason: a separator was accepted, a variant was added, the reason kept its blanks, reclaim reached the lease check). Green: commit b1bd9ea3d4. Slice B (`--generation`, N1) is built in revision 2 from the local integration tip 05cbe0ac67 once card C-99 gave `src/ops/card.rs` back (C-129 note 3432). RED first: commit b1b38109a6 (`a_stale_generation_is_refused` and `a_generation_that_is_not_a_positive_number_is_a_usage_error` fail: `unexpected argument '--generation'`; the tests of audit notes N-1 and N-2 pass at once, they pin what slice A already did).

- AC-1: met - `line_separators_are_refused_in_a_reason_and_a_scope_entry` (U+2028 and U+2029, inner, leading and trailing, in a reason and in an added pattern of a held card: `invalid_input`; `card add --scope` and `card take --scope` too; the reservation rows are as before and no `card.scope` event exists) and `a_reclaim_reason_with_a_line_separator_is_invalid_input`. Mutants M1 (scope check off) and M2 (reason check off) killed.
- AC-2: met - `a_reason_is_kept_without_its_surrounding_blanks` (spaces, U+3000 and NBSP around the reason: the answer and the event hold the bare text). Mutant M3 (stored untrimmed) killed.
- AC-3: met - `a_case_variant_is_not_a_new_pattern` (`Apps/Web/` next to `apps/web/`, `libs/Util/` with `libs/util/` in one call, and `Other/Apps/Web/` after `other/apps/web/` was added: `invalid_input` naming both spellings, every row unchanged; `other/apps/web/` is added). Mutants M4 (no folding) and M5 (no check among added patterns) killed.
- AC-4: met - `card_scope_widen_followup_cli` 4, `card_scope_widen_cli` 9, `card_dispatch_cli` 25, `card_reclaim_cli` 16, `card_work_cli` 25, `ledger_reservation` 28 and the library tests of `reservation` 14 all passed.

Not changed: `src/card/input.rs` (reserved by C-122; the separator check sits in `dispatch.rs`, D-2 and D-3), `src/ops/card.rs`, `src/card/command.rs`. The help text of `card scope` does not name the three rules (it lives in `src/ops/card.rs`); the operations document does.

- AC-5: met - `a_stale_generation_is_refused` (generation 1 and 3 on a card at generation 2: `conflict`, `generation 2` and the given number named, every row unchanged and no `card.scope` event; `--generation 2` widens and answers generation 2, the released reservation of generation 1 keeps its pattern; no `--generation` widens as before; a ready card with `--generation 1` is `conflict` `not held`) and `a_generation_that_is_not_a_positive_number_is_a_usage_error` (`0`, `-1`, `x` and an empty value: clap exit 2, kind `usage`, rows unchanged).
- AC-6: met - `a_non_ascii_case_variant_is_not_a_new_pattern` (`café/` and `CAFÉ/`: `invalid_input` naming both) and `a_reason_of_500_characters_with_blanks_is_kept_and_501_is_refused` (500 kept as 500 with blanks around, 501 refused, nothing written).
- Slice B not changed: `src/ledger/reservation.rs` (card C-127 holds it; the comparison sits in `dispatch.rs`, D-7), `src/card/command.rs` (card C-99 holds it). N6 (the SPEC status) was the user's decision (D-5): see the next line.
- N6: met - the user approved this SPEC in a session selection on 2026-11-19 (leader note 3582 of card C-129; the options were both SPECs, one, or none); the header carries `status: approved` and the `approval` block in one commit, its digest read from `spec show` after the Scope was corrected (the approval commit added the C-94 SPEC file to the write list and moved the `src/card/command.rs` note to forbidden; requirements and acceptance are unchanged), and `spec check` reports 0 errors and 0 warnings. An `## Interview` section was added because an approved SPEC created on 2026-11-19 or later needs one. The Scope lists every file the six C-129 commits touched (`git show --name-only` of f837d0a906, 7edc675d2e, b1bd9ea3d4, e93897b421, b1b38109a6, 8c88b03ef0: `src/card/dispatch.rs`, `src/ledger/reservation.rs`, `src/ops/card.rs`, `tests/card_scope_widen_followup_cli.rs`, `docs/design-review/appendix/operations/card.md`, this SPEC folder and `.moai/specs/SPEC-20261119-card-scope-widen/spec.md`, which the approval commit added to the Scope). SPEC-20261119-card-scope-widen (the C-94 SPEC) keeps its own status: note 3582 names this SPEC alone, so this record does not approve it.
