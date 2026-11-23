---
id: SPEC-20261124-card-number-bare-prefix
title: "Pin a line that ends with a bare card prefix"
status: approved
revision: 1
created: 2026-11-24
approval: {by: leader, at: 2026-11-23, basis: "leader decision note 7923 of the card (2026-11-23T21:47:02Z, delegation of the user 2026-11-23T13:57Z, the leader's judgement and not a selection by the user) answering request note 7922", digest: "sha256:408200a687bd6f9bd7161b7d70d7f7491c69a862c2bac992b4d8ab148ce16fc1"}
---

# Pin a line that ends with a bare card prefix

## Interview

- none: 리더 배차가 요구사항을 전달했어요 — 카드 본문(독립 감사의 비차단 N1)이 범위를 정했고, 별도 사용자 인터뷰는 없었어요. Related: SPEC-20261123-card-number-observe.

## Intent

SPEC-20261123-card-number-observe added the card-number observe notice. Its scan, `card_numbers_in` in `src/spec_doc.rs`, loops while `i + 2 < b.len()` and then reads `b[i + 2]`. The independent audit of that change recorded one non-blocking finding (N1): the loop-bound mutant `i + 2` → `i + 1` survives the tests. With that mutant, a line that ends with the bare prefix (`C-` or `c-` and no digit) reads index `len` and panics, while the real code returns 0 and gives no notice. This SPEC pins that line with a test; it changes no behavior.

Grounded in the code at the integration tip `b77d1efe75`: the loop bound is `src/spec_doc.rs:717`, the existing line-end sample is in `only_a_card_number_at_a_word_start_counts` (`tests/spec_card_number_observe.rs`), and it ends in a digit, so it does not reach the bare prefix.

## Non-goals

- Any change under `src/`: the behavior is already right; only the test is missing. Another card holds `src/spec_doc.rs`.
- The notice text, the date boundary, the word-start rule and rule 10 of `sdd.md`: all stay as SPEC-20261123-card-number-observe left them.

## Requirements

- REQ-1: When a line of a new-form SPEC ends with the bare card prefix (`C-` or `c-` with no digit after it), the card-number observe scan shall give no notice for that line.
- REQ-2 (REQ-1): IF the bare prefix is the last two bytes of the whole text with no line break after it, THEN the card-number observe scan shall return without reading past the end of the text.

## Acceptance

- AC-1 (REQ-1, REQ-2): a text whose Intent line ends with a bare prefix, alone or after a word, in both letter cases, and a text whose last two bytes are a bare prefix, give no notice and no panic, and a line that holds a real number and then ends with a bare prefix counts one number · verify: `cargo test --locked --offline --test spec_card_number_observe a_line_ending_with_a_bare_card_prefix_gives_no_notice_and_no_panic`

## Scope

- write: `.moai/specs/SPEC-20261124-card-number-bare-prefix/spec.md`, `tests/spec_card_number_observe.rs`
- read-only: `src/spec_doc.rs`, `.moai/specs/SPEC-20261123-card-number-observe/spec.md`
- forbidden: `src/`, `Cargo.toml`, `tools/repo_scans.json`

## Plan

1. This SPEC, then one new test in `tests/spec_card_number_observe.rs` that passes on the current code (a characterization: the behavior is already right).
2. The guard mutant, `i + 2` → `i + 1` at `src/spec_doc.rs:717`, made only in an exported copy of the head in a scratch folder with its own build folder, never in this worktree: the new test must fail there.
3. One affected run at the READY head, a Result line, the READY note.

## Risks

- A text that ends with a bare prefix and no line break is built by hand in the test, because `moai spec` files normally end with a line break; the case is still reachable when a text is parsed from a string.
- The mutant is made in a copy, so the worktree never holds a changed `src/spec_doc.rs`; the READY note says where the copy ran.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A new SPEC instead of a Result line in the old one | The branch name needs a SPEC folder of its own (leader decision note 7920) | none needed |
| D-2 | The mutant runs in an exported copy | `src/spec_doc.rs` is reserved by another card (leader decision note 7918) | Run it in the worktree after that card lands |

## Result

Not yet run. Revision 1 is a draft.
