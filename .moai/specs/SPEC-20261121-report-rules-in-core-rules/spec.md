---
id: SPEC-20261121-report-rules-in-core-rules
title: "The completion-report and report-number rules live in the core rules, and the skill entry points to them"
status: draft
revision: 1
created: 2026-11-21
---

# The completion-report and report-number rules live in the core rules, and the skill entry points to them

## Intent

The skill entry `skills/moai/SKILL.md` is meant to be the entry and the router: it says which route to read and when. Today 3219 of its 6632 bytes are rules (the decision order, the HOLD categories, the decision record fields, the completion report and the report-number rule). The decision order and the HOLD categories are defined once there on purpose: `references/internal/decide.md` and the workflows point to `SKILL.md` for them, and a test pins their words. The two sentences about reports are different: the completion-report rule and the rule that every number, state and time in a report comes from a named command overlap rules 10 to 18 of `references/internal/core-rules.md`, which every route reads once per session, and nothing points to them in `SKILL.md`. This SPEC moves those two rules, word for word, into the core rules as rules 23 and 24, leaves one line in the skill entry that says where they are, and keeps the decision order, the HOLD categories and the decision record in the skill entry. The skill entry shrinks by about 700 bytes; the bytes are not lost, they move to a file that is read at the start of every route anyway. It changes skill text and tests only.

## Non-goals

- Moving the decision order, the HOLD categories or the decision record fields out of `SKILL.md` (a narrower range than the leader's first wording; see D-1).
- Changing the words of the moved rules: only their place changes.
- Changing frontmatter, the route list, the mode flags or the host-tool paragraph of the skill entry.
- The stage order and the `/loop` rule, which are other slices of the same card.

## Requirements

- REQ-1: The skill entry shall say in one line that the report rules are rules 23 and 24 of the core rules, named by the route table's `core_rules` entry and not by a package path.
- REQ-2: The core rules shall carry the completion-report rule and the report-number rule as rules 23 and 24 with the words the skill entry carried.
- REQ-3: The skill entry shall keep the decision order, the HOLD categories and the decision record fields.
- REQ-4: The size cap of the core rules shall rise by no more than the bytes of the two moved rules and their section line.
- REQ-5: The skill's source manifest shall match the files after the change.
- REQ-6: The live inventory of the core rules in `references/rule-history.md` shall list rules 22 to 24.

## Acceptance

- AC-1 (REQ-1, REQ-3): the skill entry has the pointer line and still has the decision order, the HOLD categories and the record fields, and no longer has the two moved sentences.
  verify: `cargo test --locked --offline --test report_rules_core the_skill_entry_points_to_the_core_rules_and_keeps_the_decision_order`
- AC-2 (REQ-2): the core rules carry both moved rules word for word as rules 23 and 24.
  verify: `cargo test --locked --offline --test report_rules_core the_core_rules_carry_the_moved_report_rules`
- AC-3 (REQ-2, REQ-3): the existing decision test reads the report sentences in the core rules and the decision sentences in the skill entry.
  verify: `cargo test --locked --offline --test product_cli open_questions_are_sorted_and_holds_name_five_categories`
- AC-4 (REQ-4): the core rules stay numbered without gaps, host-neutral and under the new cap.
  verify: `cargo test --locked --offline --test core_rules_2876 core_rules_stay_small_numbered_and_host_neutral`
- AC-5 (REQ-5): the manifest matches the package.
  verify: `python3 -I tools/source_manifest.py check`
- AC-6 (REQ-6): the rule history guard passes with the three new rows.
  verify: `cargo test --locked --offline --test rule_history_guard`

## Scope

- write: `.moai/specs/SPEC-20261121-report-rules-in-core-rules/**`, `tests/report_rules_core.rs`, `tests/core_rules_2876.rs`, `skills/moai/SKILL.md`, `skills/moai/references/internal/core-rules.md`, `skills/moai/references/rule-history.md`, `skills/moai/references/source-manifest.json` (regenerated last, after its other holders release it), and the lines of `tests/product_cli.rs` that read the moved sentences in `skill_package` (leased for those lines only)
- read-only: `skills/moai/references/internal/decide.md`, `skills/moai/workflows/mode-leader.md`, `tests/rule_history_guard.rs`
- forbidden: a card, Issue or pull request number in the sentences this change adds, and any edit of a file another card holds before that card has released it or agreed in a pair line

## Plan

1. Write the test first: `tests/report_rules_core.rs` asserts the pointer line and the kept decision text in the skill entry and the two rules in the core rules; it fails on the tip because the rules are still in the skill entry.
2. Move the two rules, add the pointer line, add the three inventory rows, raise the cap.
3. Move the four sentence checks of `open_questions_are_sorted_and_holds_name_five_categories` from the skill entry to the core rules (a line lease on `tests/product_cli.rs`).
4. Regenerate the manifest rows; run the targets named above and every test that reads these files; ask for an independent audit.

## Risks

- `tests/product_cli.rs` and the source manifest are held by other cards while this is written: the test lines are changed under a line lease with a pair note naming the exact lines, and the manifest rows are written last.
- Inference, not read from the documents: the official skill documentation says the skill entry is re-attached after a compaction, while a file read with the file tool is not. The decision order and the HOLD categories are safety rules, so they stay in the skill entry; the two moved rules guard report wording, not an action.
- The core-rules cap exists to keep loading cheap: the moved bytes leave the skill entry, so the two files together load as many bytes as before plus one pointer line.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Only the completion-report and report-number sentences move; the decision record fields stay | `decide.md` and `mode-leader.md` say "every field `SKILL.md` names", and a test pins the field sentence in the skill entry | Move the field sentence too and change those pointers |
| D-2 | The moved rules keep their words | The decision test pins them; a reworded rule would need a new decision | Reword in a later SPEC |
| D-3 | The core-rules cap rises from 4608 to 5120 bytes | The same bytes leave the skill entry; the cap was already raised once for added rules | Lower it after the rules move back |
| D-4 | The rules go in a section of their own, `## Reports`, after rule 22 | Rule numbers are fixed and the numbering test needs no gaps | Move them into another section |
| D-5 | The pointer line names the route table's `core_rules` entry, not the file path | A skill package test lets the skill entry name only the route table as a package path (one hop) | Name the path and change that test with a decision |

## Result

Filled in by the implementation lane on 2026-11-21; the audit checks it again. The SPEC is `1c56cfc447`; red `9b3036d00d`, green `51dbf6c90a`, the pointer fix `84697d1eb1` (the first green pointer named a package path, which `skill_package_frontmatter_is_host_neutral` refused; it now names the route table's `core_rules` entry), the lines of `tests/product_cli.rs` `78921658c5`, the three manifest rows `2509e156a4`; the local tip `42f5af0800` is merged in as `03fc31a63f` (the earlier tip as `9999b47943`, no conflict, `git merge-tree` exit 0). Evidence files are in this worktree's `target/notes/`.

Red (`red1.out`, before the text): `the_core_rules_carry_the_moved_report_rules` and `the_skill_entry_points_to_the_core_rules_and_keeps_the_decision_order` FAILED on the tip text, 0 passed and 2 failed in `tests/report_rules_core.rs`.

- AC-1: met — `cargo test --locked --offline --test report_rules_core` (`green4.out`, no name filter; again in `affected/06-test_targets.log`): 2 passed, 0 failed; the first test is `the_skill_entry_points_to_the_core_rules_and_keeps_the_decision_order`.
- AC-2: met — same run, `the_core_rules_carry_the_moved_report_rules` (rules 23 and 24 word for word, `## Reports` after rule 22).
- AC-3: met — `cargo test --locked --offline --test product_cli` in `affected/06-test_targets.log`: 79 passed, 0 failed; `open_questions_are_sorted_and_holds_name_five_categories` reads the four report sentences in the core rules and the decision sentences in the skill entry.
- AC-4: met — `cargo test --locked --offline --test core_rules_2876` (`green4.out`, `affected/06-test_targets.log`): 6 passed; the cap is 5120.
- AC-5: met — `python3 -I tools/source_manifest.py check` prints `{"conflicted": false, "changed": [], "dropped": [], "ok": true}` on the final head.
- AC-6: met — `cargo test --locked --offline --test rule_history_guard` (`green4.out`): 6 passed, 0 failed.

### Checks

`moai check affected 42f5af0800..03fc31a63f` (the branch's own build of the binary, `--max 2`), summary line as printed: `check-affected base=42f5af0800fb head=03fc31a63f6d targets=43 py=8 scans=listed:42 diff_check=0 fmt=0 clippy=0 test_lib=0 test_bins=0 test_targets=0 build_cli=0 python=0 result=PASS`.

Shared files: the lines of `tests/product_cli.rs` were committed on this branch only, after the holder's note that its own hunks do not overlap (the holder's hunks are in other functions) and the leader's pair approval; the card check refuses that path, so this is a pair commit and not an allowed write. The same holds for the three manifest rows (leader approval, decision note of 2026-11-21T13:42:14Z).

Mutants (five, on the committed files, each restored, `restored: clean = True`; `mutate.out`): the completion sentence back in the skill entry, the time sentence of rule 24 dropped, rule 24 numbered 25, the decision record sentence dropped from the skill entry, the pointer naming a package path — each killed by a named test of `report_rules_core` or `core_rules_2876`.

### Not run

- The pointer text read by a host in a live session after a compaction: not run (the SPEC marks that premise as inference from the documents, not tested).
- A second run of the same tests on the same tree: not repeated (each result above is the one run's output).
