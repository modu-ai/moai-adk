---
id: SPEC-20261124-registry-line-add-tier
title: "Registry additions keep the audit floor at tier 1"
status: draft
revision: 3
created: 2026-11-24
branch: spec/registry-line-add-tier-followup-20261125
---

# Registry additions keep the audit floor at tier 1

## Intent

`tools/repo_scans.json` is a tier 2 prefix (`gate::TIER2_PREFIXES`), and the lane rules ask every new test file that reads the checked-in tree to register in it. So a plain registration (a rule file and a `--test` option) is a full tier 2 audit, and on 2026-11-24 31 of the 54 audit notes read were tier 2. The path rule judges names only; it cannot tell a registration from an edit of a scan rule.

Fix: when a change of the registry only grows it, its path does not set the tier 2 floor, so the floor comes from what else the change touches (tier 1 for the registry itself). A change that edits or removes anything the registry already said stays tier 2.

Readers: the lander and the auditors (`moai bundle check` prints the floor), and every lane that registers a test file.

## Non-goals

- The audit tier of any other path.
- `tests/areas.toml` and `tests/fixtures/state_folder_literals/allow.tsv`: both sit under `tests/`, so they are tier 0 for any change today; they get no rule (D-2).
- Judging whether a new scan entry is a good scan: a scan added is growth and is audited at tier 1 as any code change.

## Requirements

- REQ-1: When a change of `tools/repo_scans.json` only grows the registry (a rule file or a test name added to a scan, a `--test <name>` option added to a cargo line of its checks before the `--` that starts libtest's arguments, whose name is a target name `[a-z0-9_]+`, a scan added), the audit floor shall not count that path as a tier 2 path.
- REQ-2: When a change of `tools/repo_scans.json` edits or removes anything the old text had (a text field, a list member, a check line, a scan), or adds a key, a check line or a check word of any other kind, the audit floor shall stay tier 2 · failure: none — the rule only decides whether the path prefix applies, it reads and writes nothing.
- REQ-3 (REQ-1): IF the old text, the new text or either side's JSON is missing or unreadable, or a repeated key overrides an old value, THEN the audit floor shall stay tier 2.
- REQ-4 (REQ-1): While another path of the same change is a tier 2 path, the audit floor shall stay tier 2.
- REQ-5: When a changed path is not exactly `tools/repo_scans.json`, the audit floor shall apply every tier 2 prefix to that path as before · failure: none — a path rule has no failure path.
- REQ-6: When `moai bundle check` judges a head and when `moai bundle body` writes its `audit-tier:` line, they shall read the floor with the same rule · failure: none — a head with no merge base is still the existing `no_merge_base` finding.
- REQ-7: When the landing check judges the tier of a verdict's head, it shall read the registry text at the merge base and at the head and judge the floor with the growth rule of `moai bundle check` · failure: none — the rule only decides whether the path prefix applies.
- REQ-8: When the `audit-gate` job decides a pull request whose files include `tools/repo_scans.json`, it shall read the registry text at the merge base and at the head and judge the floor with the growth rule of `moai bundle check`.
- REQ-9 (REQ-8): IF the merge base, the answer for either side or its text cannot be read, THEN `audit-gate` shall judge the registry by its name, tier 2.
- REQ-10: When `moai bundle check`, `moai bundle body` or the landing check list the paths a head changes, they shall list a renamed file by its old and its new name · failure: none — a path list has no failure path.
- REQ-11: When `gate::registry_only_grew` reads the value of a `--test` option, it shall use `repo_scans::is_target_name`, the one function that spells the target-name rule · failure: none — a pure function.

## Acceptance

- AC-1 (REQ-1): a registration (rule file and `--test` added; test name, second `--test` and a scan added; no change) gives floor 1 · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_registration_keeps_the_floor_at_tier_1`
- AC-2 (REQ-2): an edited text field, a removed or replaced rule file or test name, a removed scan, a removed or replaced `--test`, any other added word, a check line added or removed, a key added each give floor 2 · verify: `cargo nextest run --locked --offline --test gate_registry_tier an_edited_or_removed_line_is_tier_2 a_check_line_changed_in_any_other_way_is_tier_2 a_registry_of_another_shape_is_tier_2`
- AC-3 (REQ-3): a missing side, text that is not JSON, and one added line that repeats a key to override an old value each give floor 2 · verify: `cargo nextest run --locked --offline --test gate_registry_tier the_names_alone_still_give_tier_2 a_line_that_only_adds_text_but_overrides_an_old_value_is_tier_2`
- AC-4 (REQ-4, REQ-5): a registration next to a tier 2 path gives 2 and next to a tier 0 or tier 1 path gives 1; the same texts under another path that starts with a tier 2 prefix give 2; `tests/areas.toml` and the allow list stay tier 0 · verify: `cargo nextest run --locked --offline --test gate_registry_tier another_tier_2_path_in_the_change_keeps_the_floor_at_tier_2 only_the_registry_path_is_judged_by_its_text the_two_other_registries_are_tier_0_for_any_change`
- AC-5 (REQ-6): in a temporary repository `moai bundle check` answers `audit_floor` 1 for a registration, 2 for an edit, a removal and a registration next to `src/gate.rs`, 0 for a document, and judges a head behind the tip by its own change · verify: `cargo nextest run --locked --offline --test gate_registry_tier bundle_check_reads_a_registration_as_tier_1_and_an_edit_as_tier_2 bundle_check_judges_a_head_behind_the_tip_by_its_own_registry_change`
- AC-6 (REQ-1): the 50 most recent audited ranges are recomputed with the new rule and the number of ranges whose floor changes is in the Result · verify: `moai spec check SPEC-20261124-registry-line-add-tier --project .`
- AC-7 (REQ-2): a `--test` pair whose value is an option, a filter or a shell fragment (`--list`, `--exact`, `--skip`, `--ignored`, `x||true`, `x;true`), a pair whose value is not `[a-z0-9_]+`, and a `--test` pair after the first `--` of the line (with a valid name or not) each give floor 2, while a target-named pair before the `--` gives 1, and the name rule agrees with `repo_scans::checks_of` on ten sample names · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_test_option_whose_value_is_no_target_name_is_tier_2 a_test_option_after_the_libtest_separator_is_tier_2 a_test_option_value_is_read_as_repo_scans_reads_a_target_name`
- AC-8 (REQ-8, REQ-9): `audit_gate::decide` on an input that carries the registry texts gives tier 1 for a registration and 2 for an edit, a removal, a registration next to `src/gate.rs`, and for a text that is missing on either side; an input that carries no texts and a body that declares tier 1 still gives 2 for the registry path · verify: `cargo nextest run --locked --offline --test gate_registry_tier audit_gate_reads_a_registration_as_tier_1_and_an_edit_as_tier_2 audit_gate_without_registry_texts_keeps_tier_2`
- AC-9 (REQ-8, REQ-9): `audit_gate::collect` reads the merge base from the compare answer and the registry text at both commits (base64 content), only when the files list names the registry, and an unreadable side leaves the texts out · verify: `cargo nextest run --locked --offline --test gate_registry_tier collect_reads_the_registry_at_the_merge_base_and_the_head collect_reads_no_registry_text_for_other_files collect_leaves_the_texts_out_when_a_side_cannot_be_read`
- AC-10 (REQ-7): in a temporary repository the landing floor is 1 for a registration, 2 for an edit, a removal and a registration next to `src/gate.rs`, 0 for a document, and a head behind the tip is judged by its own change · verify: `cargo nextest run --locked --offline --test gate_registry_tier the_landing_floor_reads_a_registration_as_tier_1_and_an_edit_as_tier_2`
- AC-11 (REQ-6, REQ-7, REQ-8): for the same heads `moai bundle check`, the landing floor and `audit_gate::decide` (with the texts `git show` gives) answer the same tier · verify: `cargo nextest run --locked --offline --test gate_registry_tier the_three_judgments_give_one_tier_for_the_same_head`
- AC-12 (REQ-6, REQ-10): `moai bundle body` writes `audit-tier: 1 …` for a bundle whose head only registers a test, `audit-tier: 2 …` for one whose head edits the registry and for one whose head renames it · verify: `cargo nextest run --locked --offline --test bundle_body_cli the_body_reads_a_registration_as_tier_1_and_an_edit_as_tier_2 the_body_reads_a_renamed_registry_as_tier_2`
- AC-13 (REQ-10): a head that renames the registry file gives floor 2 in `moai bundle check` and in the landing floor · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_renamed_registry_is_tier_2_in_bundle_check_and_the_landing_floor`
- AC-14 (REQ-11): `repo_scans::is_target_name` is public and the growth rule answers as it does for the ten sample names; `src/gate.rs` has no `fn is_target_name` · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_test_option_value_is_read_as_repo_scans_reads_a_target_name`
  manual: `grep -c "fn is_target_name" src/gate.rs` prints 0.

## Scope

- write: `.moai/specs/SPEC-20261124-registry-line-add-tier/**`, `src/gate.rs`, `src/bundle.rs` (the two floor calls, added by the leader's decision on the card, D-4; the path list, revision 3), `src/audit_gate.rs`, `src/card/land/check.rs`, `src/repo_scans.rs` (`is_target_name` public), `tests/gate_registry_tier.rs`, `tests/bundle_body_cli.rs` (the body test, whose world helper lives there)
- read-only: `tools/repo_scans.json`, `tests/areas.toml`, `tests/fixtures/state_folder_literals/allow.tsv`, `docs/lane-rules.md`
- forbidden: any other change of `tools/repo_scans.json`; `CLAUDE.md`

## Plan

1. RED then GREEN in `tests/gate_registry_tier.rs`: the function cases first, then the `bundle check` scenes.
2. `src/gate.rs`: `ChangedFile`, `tier_of_changes` and `registry_only_grew`; `tier_of` keeps its signature and its names-only behaviour through the one shared loop.
3. `src/bundle.rs`: one helper reads the old and new text of the registry path and gives both call sites (`bundle check`, `bundle body`) the floor.
4. Guard mutants: one per new guard, on the committed head, restored by git.
5. Recompute the 50 most recent audited ranges with the old and the new rule.
6. Revision 3 (follow-up criteria 4 to 6 of the card): `gate::changed_files_with_text` builds the changed files of a head once for the git callers; `moai bundle check`, `moai bundle body` and the landing check (`card::land::check::floor_tier`) call it on a path list made without rename detection; `audit_gate::Input` carries the registry texts and `audit_gate::collect` reads them through the compare and contents API only when the registry is a changed file; `repo_scans::is_target_name` becomes the one name rule. RED first in `tests/gate_registry_tier.rs` and `tests/bundle_body_cli.rs`.

## Risks

- A registration that is judged tier 1 is audited by one session that did not write it, without the full audit of tier 2. The exemption is narrow (growth only, one path, nothing else tier 2 in the change) and a scan added is the one open end: a new scan's check line is a command the check step runs, so the auditor of a tier 1 change reads it.
- The texts of the registry on both sides come from `git show`; a head with no merge base or a path that is gone on one side reads as no growth, tier 2.
- `gate::tier_of` stays names-only for a caller that has no texts; after revision 3 the pull request gate, the landing check and `moai bundle check` read one growth rule from one function, and a registry whose text cannot be read still counts tier 2 (the stricter side).
- `audit-gate` reads the registry through the GitHub compare and contents API: one extra call for the merge base and two for the texts, made only when the registry is a changed file. A file over the contents API size limit, a missing field and a base64 text that does not decode read as an unreadable side, tier 2.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "Only adds" is read on the parsed registry, not on lines: a rule file, a test name, a `--test` option or a scan may be added; nothing the old text had may change | the card's rule says pure added lines, but a real registration changes one existing line (the check line gets `--test <name>`, a line removed and written again) and a pure line add can override an old value with a repeated key; measured on the 50 ranges: lines alone change the floor of 2, the parsed rule of 11 (Result table) | restore `registry_only_grew` to a line-only comparison, or drop the exemption (`tier_by` ignores the flag) |
| D-2 | `tests/areas.toml` and the allow list get no rule | both are under `tests/`: `tier0_safe` already makes any change of them tier 0, and raising an edit of them to tier 2 would add audit load, not remove it | none needed |
| D-3 | `gate::tier_of` keeps its signature; the pull request gate and the landing check keep judging names | they have no old and new text, and judging them tier 2 is the stricter side | give those two callers the texts in a later change |
| D-4 | `src/bundle.rs` is in the write scope | leader decision on the card (2026-11-24T14:05:26Z, delegated by the user), requested because the floor calls are there | drop `src/bundle.rs` and the floor of `moai bundle check` stays names-only |
| D-5 | Revision 2: a `--test <value>` pair is growth only when its value is a target name `[a-z0-9_]+` and it stands before the first `--` of the line. The name rule is spelled in `src/gate.rs` (`is_target_name`) because `repo_scans::is_target_name` is private and `src/repo_scans.rs` is outside this card's write scope; a test pins that the two agree. | Audit of head `a3ac976d6f` (note 8625), blocking finding B1, leader repair order of 2026-11-24: `--test --list`, `--test --exact`, `--test --skip`, `--test --ignored` after a `--` turn a red check green, and `x\|\|true` and `x;true` were counted as growth. | Count every `--test` pair again (floor 1 for the six probes, which is the defect). Sharing the one function is a follow-up in the card that owns `src/repo_scans.rs`. |
| D-6 | Revision 3: D-3 is replaced. `audit-gate` and the landing check read the registry texts and call `gate::tier_of_changes`, so the three judgments of one head give one tier; `gate::tier_of` stays for a caller that has no texts. The path lists are made without rename detection, so a renamed registry is a removed and an added path. | Card C-475 criteria 4 to 6 (follow-ups of D-3, audit note 8625 N1 and the minor finding of note 8640); leader dispatch of 2026-11-25. | Give `audit_gate::decide` and `floor_tier` the names-only call again (`gate::tier_of`); `--no-renames` back to the default list. |
| D-7 | `audit_gate::Input` gets an optional `registry` field (`old`, `new`); old input files without it still parse. The texts come from the GitHub contents API (base64) at the merge base the compare answer names. | `audit-gate` decides from GitHub data alone (no git, no clone), and the schema stays version 1 because an input without the field keeps its meaning (tier 2 for the registry path). | Drop the field and the reads; `decide` falls back to names. |

## Result

Measured 2026-11-24 on macOS in the worktree `.claude/worktrees/c475-registry-tier`, code head `ac9fccef9c81b778f5b6f2c12df4f2e04ea1c0f3`.
- AC-1 met: `a_registration_keeps_the_floor_at_tier_1` passes (rule file plus `--test`; test name, second `--test` and a scan added; an unchanged registry), and `cargo nextest run --locked --offline --test gate_registry_tier` ran 11 tests, 11 passed.
- AC-2 met: `an_edited_or_removed_line_is_tier_2`, `a_check_line_changed_in_any_other_way_is_tier_2` and `a_registry_of_another_shape_is_tier_2` pass: each edit, removal, replaced `--test`, extra check word (`--release`, `-- --ignored`, `--lib`, `; true`, a bare `true`), dangling `--test`, added or removed check line, added key, repeated scan id and changed `about` gives floor 2.
- AC-3 met: `the_names_alone_still_give_tier_2` (a missing side, either side) and `a_line_that_only_adds_text_but_overrides_an_old_value_is_tier_2` (one added line that repeats `rules` or `about`, nothing removed) give floor 2; text that is not JSON gives 2 in `a_registry_of_another_shape_is_tier_2`.
- AC-4 met: `another_tier_2_path_in_the_change_keeps_the_floor_at_tier_2` (next to `src/gate.rs`, a hook, `.claude/`, `CLAUDE.md`, `tools/check_guard.py`: 2; next to a test file or code: 1; body 7 or none: 2), `only_the_registry_path_is_judged_by_its_text` (six other tier 2 paths with the registration's texts: 2) and `the_two_other_registries_are_tier_0_for_any_change` (0 for an added and an edited line).
- AC-5 met: `bundle_check_reads_a_registration_as_tier_1_and_an_edit_as_tier_2` (a registration 1; an edit 2; a removed rule file 2; a registration next to `src/gate.rs` 2; a document 0) and `bundle_check_judges_a_head_behind_the_tip_by_its_own_registry_change` run `moai bundle check` in a temporary repository. `moai bundle body` takes the floor from the same helper (`Git::audit_floor`); no test writes a body for a registry change (gap).
- AC-6 met: the 50 most recent distinct audited ranges (`head` and `base` of the `AUDIT` verdict lines of the card ledger, read from `moai card export` of 2026-11-24, index sha256 `5a5baca9fafbdfcf189b3cab8c2fcaf315834317b952254c6fecd04ae67a084a`; 358 distinct verdict ranges, newest note 8578) were recomputed twice, by a script that mirrors the rule and by a probe linked to the head's library calling `gate::tier_of` (names only, the old rule) and `gate::tier_of_changes` (new rule) with `git show` texts of the two sides; both gave the same answer. **Tier 2 ranges: 27 of 50 before, 16 of 50 after; 11 ranges change from 2 to 1, 0 change the other way, 31 ranges do not touch the registry.** A line-only reading (the card's first wording, no edited lines) would change 2.

  | range head | the registry change | other tier 2 path | floor before → after |
  | --- | --- | --- | --- |
  | 1489d1645de7 | rule files and `--test` added | none | 2 → 1 |
  | 9de55e2c2c59 | rule files and `--test` added | none | 2 → 1 |
  | 3874a3c91604 | a scan added | none | 2 → 1 |
  | c3ac0f393650 | rule files and `--test` added | none | 2 → 1 |
  | b8fb5096cbd7 | rule files and `--test` added | none | 2 → 1 |
  | 1ce2663acc8e | rule files and `--test` added | none | 2 → 1 |
  | 5fc44126bb96 | rule files and `--test` added | none | 2 → 1 |
  | 89e264afc258 | rule files and `--test` added | none | 2 → 1 |
  | 3f4b15bad45c | rule files and `--test` added | none | 2 → 1 |
  | 548822bf0a19 | rule files and `--test` added | none | 2 → 1 |
  | fd8876baddf8 | a scan added | none | 2 → 1 |
  | 8213a5a3f790 | rule files and `--test` added | `src/agent_install.rs`, `src/update.rs` | 2 → 2 |
  | e80b345fa6d0 | a scan added | `src/go_state.rs`, `src/state_root.rs` | 2 → 2 |
  | c228b6f3eb3a | a scan added | `src/go_state.rs`, `src/state_root.rs` | 2 → 2 |
  | 6e876acb7a4b | rule files and `--test` added | `.github/workflows/ci.yml` and another | 2 → 2 |
  | a4656e1e3ead | a scope entry added next to the registration | none | 2 → 2 |
  | eab4fccb1e55 | a scope entry, a `--lib` check and `rejects` text edited | none | 2 → 2 |
  | d6b28a54110c | a scan scope and text edited | none | 2 → 2 |
  | 9cee440a7e44 | a scan edited | `src/hook_bash_boundary.rs` and another | 2 → 2 |

- guard mutants, each run on the committed head by `target/c475/mutants.py` in one slot (baseline exit 0, 11 passed; the tree was clean after the last): M1 exemption removed from `tier_by`, exit 100 (1 failed); M2 text rule applied to any path, exit 100 (1 failed); M3 the registry path always counts as growth, exit 100 (4 failed); M4 top-level fields not compared, exit 100 (2 failed); M5 list members may be dropped, exit 100 (1 failed); M6 check line count not compared, exit 100 (1 failed); M7 any added check word accepted, exit 100 (1 failed); M8 old check words may be dropped, exit 100 (1 failed); M9 repeated scan id allowed, exit 100 (1 failed); M10 top-level key set not compared, exit 100 (1 failed); M11 scan key set not compared, exit 100 (1 failed); M12 check lines not read, exit 100 (1 failed); M13 text fields not compared, exit 100 (2 failed); M14 a missing side counts as growth, exit 100 (3 failed); M15 text never judged, exit 100 (4 failed); M16 `bundle check` back to names only, exit 100 (2 failed); M17 registry text never read, exit 100 (2 failed); M18 old side read at the head, exit 100 (1 failed); M19 old side read at the tip, exit 100 (1 failed).
- Other runs: `cargo nextest run --locked --offline --test gate_registry_tier --test bundle_check_cli --lib` ran 1929 tests, 1929 passed (before the commit of the code); `moai spec check SPEC-20261124-registry-line-add-tier --project .` ends with 0 errors and 0 warnings.
- Gaps: the pull request gate (`audit-gate`) and the landing check still judge names only (D-3), so for a registration they say tier 2 where `moai bundle check` says 1; `moai bundle body` is not tested for a registry change; the `moai check affected` receipt is in the READY note.

Revision 2 (repair after the audit of head `a3ac976d6f`, finding B1 of note 8625; leader repair order of 2026-11-24). Measured 2026-11-24T15:22Z on macOS in the same worktree. RED first (`db9a6a5949`): of 14 tests 3 failed on the old code (`--test --list` after `--` and between the old words read as floor 1, and `A`, `a-b`, `a.b`, `x;true`, `--list` read as growth) and 11 passed. After the repair (`982cdbf321`, `src/gate.rs` only) 14 of 14 passed, and `--lib gate::` with the integration tests ran 21 tests, 21 passed.

- AC-7 met (REQ-2): `a_test_option_whose_value_is_no_target_name_is_tier_2` (the audit's P1 to P6 values `--list`, `--exact`, `--skip`, `--ignored`, `x||true`, `x;true`, plus `x&&true`, `$(true)`, `X`, `a-b`, `a.b`, `a/b`, `--`, and one between the old words: floor 2; `beta_2` as the control: floor 1), `a_test_option_after_the_libtest_separator_is_tier_2` (the same values and a valid name after an old `--`, and one inserted after it: floor 2; the same pair before the `--`: floor 1) and `a_test_option_value_is_read_as_repo_scans_reads_a_target_name` (ten sample names: the growth rule and `repo_scans::checks_of` accept the same ones).
- AC-1 to AC-5 unchanged and still met: the 14 tests of `gate_registry_tier` pass, among them the registration, the other-shape and the `bundle check` tests.
- Guard mutants of revision 2, one per new guard, each applied once to `982cdbf321` by `target/c475-repair/mutants.py` in one slot, run on the one named test and restored (baseline 14 passed; the tree was clean after the last; each kill is a FAIL of that test, no build error): A the value is no longer checked as a target name, exit 100; B the separator is no longer looked at, exit 100; C the name rule also lets a hyphen through, exit 100.
- Not done and not claimed: N1 of the audit (`--no-renames` for `changed_paths` in `src/bundle.rs`) stays an open follow-up criterion of the card; `repo_scans::is_target_name` is still private and the rule is spelled twice (D-5); the `moai check affected` receipt of this repair is in the REPAIRED note.
