---
id: SPEC-20261123-spec-check-glued-items
title: "spec check 가 한 줄에 붙은 REQ·AC·D 목록 항목을 오류로 잡아요"
status: draft
revision: 1
created: 2026-11-23
---

# spec check 가 한 줄에 붙은 REQ·AC·D 목록 항목을 오류로 잡아요

## Intent

SPEC 의 `- AC-9 … - AC-10 …` 처럼 목록 항목 둘이 한 줄에 붙으면 파서는 줄 하나에서 항목 하나만 읽어서 `AC-10` 을 잃어요. 그런데 `moai spec check` 는 이때 오류 0 이라고 답해요(C-354 조각 4 감사 7456 에서 발견). 이 SPEC 은 그런 줄을 줄 번호와 붙은 id 와 함께 오류로 잡게 해서, 항목이 조용히 사라진 SPEC 이 검사를 통과하지 못하게 해요.

## Requirements

- REQ-1: When a line of `## Requirements`, `## Acceptance`, `## Decisions` or `## Result` starts a `- REQ-n`, `- AC-n` or `- D-n` list item and a second such item starts later on the same line, the checker shall report the error `item_glued` with the line number and the id of the second item. · failure: none — a line that holds one item is the normal case and REQ-2 keeps it clean.
- REQ-2: The checker shall read a ` - ` that sits inside a backtick span, inside a fenced code block or right after an id as text and never as a second item.
- REQ-3: The checker shall report `item_glued` as an error in a new-form SPEC and as a warning in a Go-generation SPEC, as it does for the other format codes.

## Acceptance

- AC-1 (REQ-1): a `## Acceptance` line `- AC-9 … - AC-10 …` gives one `item_glued` error that names that line number and `AC-10`, and a line with three items gives one error for each later item.
  verify: `cargo nextest run --locked --offline --test spec_doc -E 'test(a_glued_item_is_an_error_with_its_line_and_id)'`
- AC-2 (REQ-1): glued `REQ-n` items in `## Requirements`, `D-n` items in `## Decisions` and `AC-n` items in `## Result` give the same error, and a mixed pair such as `- REQ-1 … - AC-2 …` does too.
  verify: `cargo nextest run --locked --offline --test spec_doc -E 'test(glued_items_are_found_in_every_item_section)'`
- AC-3 (REQ-2): a ` - AC-n` inside a backtick span, inside a fenced block, after an id such as `AC-3 - AC-5` and on a line that starts no item gives no `item_glued` error.
  verify: `cargo nextest run --locked --offline --test spec_doc -E 'test(text_that_only_looks_glued_is_not_an_error)'`
- AC-4 (REQ-2): a SPEC whose items each start their own line stays free of `item_glued` and the existing parser tests still pass.
  verify: `cargo nextest run --locked --offline --test spec_doc`
- AC-5 (REQ-3): the same glued line is an error in a new-form SPEC and a warning in a Go-generation one.
  verify: `cargo nextest run --locked --offline --test spec_doc -E 'test(a_glued_item_is_a_warning_in_a_go_generation_spec)'`
- AC-6 (REQ-1, REQ-2, REQ-3): this SPEC passes the checker, and the number of `item_glued` errors over every SPEC in the repository is counted and written in the Result.
  verify: `moai spec check SPEC-20261123-spec-check-glued-items`

## Scope

- write: `src/spec_doc.rs`
- write: `tests/spec_doc.rs`
- write: `.moai/specs/SPEC-20261123-spec-check-glued-items/**`

## Plan

1. Count the `item_glued` errors the new check gives over every SPEC in the repository before it is switched on for them, and record the number. SPECs that show errors are not edited; only their number goes to the leader.
2. Write the tests of AC-1 to AC-5 first and watch them fail.
3. Add a `glued_item` helper to `src/spec_doc.rs` and call it from `body()` for the four item sections, skipping fenced lines, backtick spans and a ` - ` that follows an id.
4. Run the `spec_doc` tests and `moai spec check --all`, then write the Result with each AC and the count.

## Result

- AC-1: met — `spec_doc a_glued_item_is_an_error_with_its_line_and_id` (nextest, pass; RED first, 4 failed with 0 `item_glued` findings, exit 100)
- AC-2: met — `glued_items_are_found_in_every_item_section` (pass; a mutant that drops `Decisions` and `Result` from the checked sections fails it)
- AC-3: met — `text_that_only_looks_glued_is_not_an_error` (pass; mutants that drop the after-an-id rule, the backtick rule or the fence rule each fail it)
- AC-4: met — all 31 tests of `--test spec_doc` pass, the 27 older ones unchanged
- AC-5: met — `a_glued_item_is_a_warning_in_a_go_generation_spec` (pass)
- AC-6: met — `moai spec check SPEC-20261123-spec-check-glued-items` 0 errors 0 warnings; the build of this branch run as `moai-cli spec check --all` over 338 SPEC files gives 0 `item_glued` findings (0 errors, 25 warnings, none of them `item_glued`); no SPEC needed an edit
- Gap: a first item written without a list marker (`AC-1 …`, which the parser reads as an item) is not checked, because REQ-1 names list items; a decision for a later revision.
- Repair after audit 402 (head 0dad411093, FAIL), same SPEC scope, no new approval: B1 the next item right after the closing backtick of a span (`` `a`- AC-10 ``, the shape of `SPEC-20261122-test-defects-session-reclaim` spec.md:47 at 267614fd4f) is read as glued — `a_glued_item_right_after_a_closing_backtick_is_an_error`; N1 a list line that starts no item is not read — `a_list_line_that_starts_no_item_is_not_read`; N2 a last backtick with no partner is a letter and the check goes on after it — `a_stray_last_backtick_is_text`; N4 a dash after an id is a range only when the id starts a word (`CD-1 - AC-2` is glued, `(AC-3 - AC-5)` is a range) — `a_dash_after_a_word_that_ends_in_an_id_is_not_a_range`. RED first: 3 of the 4 new tests failed before the change (the N1 test already passed and is held by its mutant); all 35 `--test spec_doc` tests pass after it; `moai-cli spec check --all` over 338 SPEC files still gives 0 `item_glued` findings (0 errors, 25 warnings).
- Gap (audit N3): a glued line inside a fenced block is read by the parser as a normal line but not checked, because REQ-2 reads a fenced block as text; widening it is a decision for a later revision.
- Gap (audit N5): not found are `-AC-10` with no space, `.- AC-10` after a full stop, an en or em dash, and a `*` or `+` marker in the middle of a line; REQ-1 names the `- ` marker after a space or a closing backtick only, and an em dash is ignored on purpose.
