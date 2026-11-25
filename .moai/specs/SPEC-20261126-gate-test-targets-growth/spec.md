---
id: SPEC-20261126-gate-test-targets-growth
title: "the gate reads a target name added to test_targets of the scan registry as growth"
status: draft
revision: 1
created: 2026-11-26
---

# the gate reads a target name added to test_targets of the scan registry as growth

## Intent

`registry_only_grew` and `scan_only_grew` in `src/gate.rs` (SPEC-20261124-registry-line-add-tier) let a change of `tools/repo_scans.json` keep the low audit floor when it only grows the table: a rule file or a test name added, a `--test <name>` added to a `checks` line, a scan added. The other keys, `test_targets` among them, fall into `_ => grown == value`, so adding one target name to `test_targets` is no growth and sets the tier 2 floor by the file's path. The conversion SPEC `SPEC-20261125-repo-scans-convert-test-targets` (the finding of its lane) turns the table into `test_targets` per scan; after that conversion a card that only registers a new test would lose the low floor it has today and get a full audit each time.

This SPEC teaches the gate the new form first and leaves the conversion to the conversion SPEC, which lands after it (leader decision A). The gate is a tier 2 floor path, so the rule is strict: the only new growth is a target name inserted at its sorted place in a non-empty `test_targets` list. Nothing that was not growth becomes growth, and no check of the old `checks` form changes.

## Non-goals

- Converting `tools/repo_scans.json` or editing it: that is the conversion SPEC.
- Changing how `repo_scans` reads the table (`checks_of`) or what it accepts.
- Reading a scan that gains or loses the `test_targets` key as growth: the key set of a scan must stay the same, as today, so the conversion itself stays a tier 2 change by its path.
- Judging the order of scans or any field other than `test_targets`.

## Requirements

- REQ-1: When a scan has a `test_targets` list at the base and at the head, the head's list is the base's list with one or more target names added, and the head's list is sorted in ascending byte order with no name twice, `registry_only_grew` shall treat that list as growth · failure: none — an answer of no growth is tier 2 as before.
- REQ-2 (REQ-1): IF the head's `test_targets` differs from the base's in any other way (a name removed or replaced, the order broken, a name twice, a value that is no target name `[a-z0-9_]+`, a value that is not a list, or a list that is empty at the base or at the head), THEN `registry_only_grew` shall answer that the table did not only grow · failure: the answer is no growth, so the floor is tier 2.
- REQ-3: When the `test_targets` of a scan is the same at the base and at the head, whatever it holds, `registry_only_grew` shall judge the scan by its other fields only · failure: none — an unchanged field is no difference.
- REQ-4: When a name of `test_targets` is judged, `registry_only_grew` shall use `repo_scans::is_target_name`, the function `checks_of` reads the same names with · failure: none — a test compares the two on a list of names.
- REQ-5: When a table in the old form (`checks` lines, no `test_targets`) changes, `registry_only_grew` shall answer exactly as it answered before this change · failure: none — the existing tests of `tests/gate_registry_tier.rs` pass unchanged.
- REQ-6: When the key set of a scan differs between the base and the head (a `test_targets` key added or removed), `registry_only_grew` shall answer that the table did not only grow · failure: none — the answer is already the safe one, tier 2.

## Acceptance

- AC-1 (REQ-1): a head that adds one name, several names, a name at the start, in the middle and at the end of a sorted `test_targets`, or adds names to the lists of two scans, is tier 1 by `tier_of_changes`, and a scan added with `test_targets` is tier 1 · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_target_name_added_to_test_targets_keeps_the_floor_at_tier_1`
- AC-2 (REQ-2): each of a name removed, a name replaced, an unsorted insert, a name twice, an uppercase name, a name with `-`, `.`, `/`, `;` or a leading `--`, a number, null, a string instead of a list, an empty list at the base, an empty list at the head, and a non-empty list made empty, is tier 2 · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_test_targets_change_that_is_no_growth_is_tier_2`
- AC-3 (REQ-3, REQ-6): an unchanged list of any content next to a grown `rules` list stays tier 1; a `test_targets` key added to or removed from a scan, and a list that holds a bad name at both sides but is otherwise unchanged and gets one more good name, are tier 2 · verify: `cargo nextest run --locked --offline --test gate_registry_tier an_unchanged_test_targets_is_no_difference_and_a_changed_key_set_is_tier_2`
- AC-4 (REQ-4): for a list of names, the answer "growth" of a one-name addition equals whether `checks_of` accepts that name · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_test_targets_name_is_read_as_repo_scans_reads_a_target_name`
- AC-5 (REQ-5): the tests of `tests/gate_registry_tier.rs` that exist at the base commit of this SPEC pass without an edit · verify: `cargo nextest run --locked --offline --test gate_registry_tier`
- AC-6 (REQ-1, REQ-2): the landing floor, `bundle check` and the audit gate give the same tier for a `test_targets` registration and for an edit of it, in a temporary repository · verify: `cargo nextest run --locked --offline --test gate_registry_tier the_three_judgments_give_one_tier_for_a_test_targets_registration`

## Scope

- write: `src/gate.rs`
- write: `tests/gate_registry_tier.rs`
- write: `.moai/specs/SPEC-20261126-gate-test-targets-growth/**`
- read-only: `src/repo_scans.rs`, `tools/repo_scans.json`
- forbidden: `tools/repo_scans.json` (the conversion SPEC converts it), `src/repo_scans.rs`, every other path

## Plan

1. Commit this draft, report the decision to the leader with the approval digest, and write no code before the delegated approval is on the card.
2. RED: write the five new tests of AC-1 to AC-4 and AC-6 and run them; they fail on the base (tier 2 for a registration).
3. GREEN: add a `"test_targets"` arm to `scan_only_grew` and one function, `target_names_only_grew`, beside `strings_only_grew`; update the doc comment of `registry_only_grew`.
4. Run AC-5 and the receipt, then run the guard mutants (one per new guard test) and write the READY note.

## Risks

- The gate is a tier 2 path, so a wrong arm can widen the low floor. The tests of AC-2 and AC-3 fix every non-growth shape, and the guard mutants remove the sorted check, the empty check and the same-key-set check one by one.
- Another open card holds `src/gate.rs`. The edits are disjoint in the file (a match arm and one new function next to `strings_only_grew`); the lander merges on the tip and runs the registry test once more after a merge.
- The conversion SPEC must land after this change: landed first, a registration after the conversion has the full floor until this change is on the tip. Noted in the landing order of both cards.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Growth is exactly: names inserted into a non-empty, sorted, duplicate-free `test_targets` list; the base's names stay. | Leader decision A for the conversion SPEC, the recommended default | Remove the arm and the function; the gate returns to `grown == value` |
| D-2 | An empty list at the base or the head is no growth. | The gate fails safe; `checks_of` treats an empty list as no command, so it carries no registration | Allow it in the function and the test |
| D-3 | A changed key set is no growth, so the conversion keeps the tier 2 floor by its path. | The conversion rewrites the table; it is a change of a different kind than a registration | Not planned |
| D-4 | An unchanged `test_targets` is no difference, whatever it holds. | Same as today's `grown == value`; the old shapes keep their answers | Judge the content of an unchanged list too |
| D-5 | No change of `repo_scans`; the name rule is the existing `is_target_name`. | One function for both readers (SPEC-20261124-registry-line-add-tier, REQ-11) | Not planned |
| D-6 | Landing order: this change first, then the conversion. | Leader decision A for the conversion SPEC | A leader note that reorders both cards |
