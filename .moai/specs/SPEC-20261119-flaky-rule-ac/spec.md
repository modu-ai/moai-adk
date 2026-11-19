---
id: SPEC-20261119-flaky-rule-ac
title: "mode-leader flaky rule: the card add example carries a criterion"
status: draft
revision: 1
created: 2026-11-19
---

# mode-leader flaky rule: the card add example carries a criterion

## Intent

Card C-108 criterion 3, slice 2 of `SPEC-20261119-card-criterion-add` (slice 1 is the `card criterion-add` action, in PR 3230). The flaky rule of the merge-train step in `skills/moai/workflows/mode-leader.md` (card C-99, `SPEC-20261119-flaky-policy`) tells the leader to open the flaky card with `moai card add --kind task --title "flaky: <test>"`. That command has no `--ac`, so the card it makes has no acceptance criterion, and `card done` refuses a card without a met criterion ("the card has no acceptance criterion that is met"): the flaky card could never be closed. The audit of #3225 (finding N2) found it; C-108 criterion 3 names the fix: "the flaky rule's card add example in mode-leader.md carries --ac".

Measured at tip 747738af63: the rule is 501 bytes between `Flaky rule (card C-99): ` and `touches that test.` (budget 700, `tests/workflow_size_budget.rs`), and the example is `moai card add --kind task --title "flaky: <test>"` inside the rule. `moai card add --ac` exists at the tip and takes one criterion per flag; the five needles that `flaky_rule_names_the_card_and_what_it_records` pins (title, `moai card add --kind task --title`, `one card per test`, `--kind report`, the four facts) stay as they are.

This slice adds `--ac "<criterion>"` to the example and says what the criterion is: the flake is fixed or the test is retired with its reason. It is a text change in one package file plus the manifest entry that fixes the file's sha256.

## Non-goals

- Making `card add` refuse a card without a criterion (the other way criterion 3 allows): a change to a command many callers use, and a refusal at creation hides the legitimate draft that gets its criterion later with `card criterion-add`.
- The `card criterion-add` action itself (slice 1, PR 3230) and its follow-up C-125.
- Any other sentence of `mode-leader.md`, or another workflow.

## Requirements

- REQ-1: The flaky rule of the merge-train step shall give the `moai card add` example with an `--ac` criterion next to `--kind task` and `--title`.
- REQ-2 (REQ-1): The criterion in the example shall say what closes the flaky card, namely that the flake is fixed or the test is retired with its reason.
- REQ-3: The flaky rule shall stay within its 700-byte budget and keep every statement that `flaky_rule_names_the_card_and_what_it_records` pins.
- REQ-4: The sha256 entry of `workflows/mode-leader.md` in `skills/moai/references/source-manifest.json` shall be the file's hash after the change.

## Acceptance

- AC-1 (REQ-1, REQ-2): the rule contains `moai card add --kind task --title "flaky: <test>" --ac` followed by a criterion that names the fix or the retired test with its reason; the old example without `--ac` is gone.
  verify: `cargo test --locked --offline --test workflow_size_budget flaky_rule_card_add_example_carries_a_criterion`
- AC-2 (REQ-3): the rule is at most 700 bytes and the earlier pins still pass.
  verify: `cargo test --locked --offline --test workflow_size_budget`
- AC-3 (REQ-4): the manifest hash is current.
  verify: `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`
- AC-4 (REQ-1…REQ-4): neighbours.
  verify: `python3 -m unittest tests.repo_scans_test tests.source_manifest_test tests.ci_scope_test`; `cargo test --locked --offline --test spec_cli`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `skills/moai/workflows/mode-leader.md`, `skills/moai/references/source-manifest.json` (the one entry), `tests/workflow_size_budget.rs` (one new test), `.moai/specs/SPEC-20261119-flaky-rule-ac/spec.md`
- read-only: `.moai/specs/SPEC-20261119-flaky-policy/spec.md`, `src/card/input.rs` (the `--ac` rules)
- forbidden: the real HOME and ledger, the network, other lines of `mode-leader.md`

## Plan

1. Test first (RED): `flaky_rule_card_add_example_carries_a_criterion` fails on the tip text.
2. Edit the example in the rule; read the rule's byte count back.
3. Update the manifest entry (sha256 of the new file).
4. Run the example against a temporary ledger to show `card add` takes it, then `card criterion` and the neighbours.

## Risks

- The rule has 199 bytes of room; the added words must leave the earlier pins and the end words (`touches that test.`) alone, or the rule cut in the test breaks.
- A criterion text fixed in the rule makes every flaky card start with the same criterion; the leader may replace it on the command, and `card criterion-add` can add more.
- `skills/moai/references/source-manifest.json` is a shared file: the edit is one entry, and a merge against another lane's entry change is a text merge on different lines.

## Decisions

- D-1: fix the text (the way card C-108 criterion 3 names first), not `card add`: the command stays as it is for every other caller (see Non-goals).
- D-2: the example's criterion is one short English phrase, as the rest of the workflow is English, and it names the closing condition so the leader does not invent one per card.

## Result

RED first: commit e9b95c5f6a (SPEC and one test; the test failed on the tip text with "the example has no ` --ac "…"` criterion: moai card add --kind task --title "flaky: <test>""). GREEN: commit 040c8b49bd. Base tip 747738af63. Gates ran through the heavy-slot wrapper in this worktree's own build folder.

- AC-1: met — `flaky_rule_card_add_example_carries_a_criterion` passes: the rule now reads `moai card add --kind task --title "flaky: <test>" --ac "the flake is fixed, or the test is retired with its reason"`, and the old example without `--ac` is gone. Mutants M1 (`retired` replaced by `dropped`) and M2 (`--ac` replaced by `--by`) were both killed (`test result: FAILED`), and the file was restored and `git status` was clean after each. The example was also run against a temporary ledger (project `c122-scratch` under the scratchpad, never the real HOME): `card add --kind task --title "flaky: example_test" --ac "…" --scope x --priority p3` answered `ok`, `criteria` 1.
- AC-2: met — `workflow_size_budget` 6 passed, 0 failed: the rule is 567 bytes (501 before, budget 700; read with a script that cuts between `Flaky rule (card C-99): ` and `touches that test.`), and the five earlier pins pass.
- AC-3: met — `product_cli source_manifest_covers_every_package_file` 1 passed; the manifest entry of `workflows/mode-leader.md` is now the file's `shasum -a 256` (7e336803…48e3, was 02c83e52…).
- AC-4: met — `python3 -m unittest tests.repo_scans_test tests.source_manifest_test tests.ci_scope_test` 108 tests OK (with `CARGO_TARGET_DIR` set to the build folder), `spec_cli` 18 passed, `cargo fmt --check` exit 0 (the first run found one line wrap in the new test, fixed before the GREEN commit), `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0, `git diff --check` empty.

Not run here: the full `--lib` and the other integration test files (the change is one sentence in a package file, one manifest hash and one test); Windows and Linux.

Not done: card C-108 criteria 1 and 2 belong to slice 1 (PR 3230, not merged when this was written).
