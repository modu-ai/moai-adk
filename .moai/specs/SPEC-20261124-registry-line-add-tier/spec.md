---
id: SPEC-20261124-registry-line-add-tier
title: "Registry additions keep the audit floor at tier 1"
status: draft
revision: 1
created: 2026-11-24
---

# Registry additions keep the audit floor at tier 1

## Intent

`tools/repo_scans.json` is a tier 2 prefix (`gate::TIER2_PREFIXES`), and the lane rules ask every new test file that reads the checked-in tree to register in it. So a plain registration (a rule file and a `--test` option) is a full tier 2 audit, and on 2026-11-24 31 of the 54 audit notes read were tier 2. The path rule judges names only; it cannot tell a registration from an edit of a scan rule.

Fix: when a change of the registry only grows it, its path does not set the tier 2 floor, so the floor comes from what else the change touches (tier 1 for the registry itself). A change that edits or removes anything the registry already said stays tier 2.

Readers: the lander and the auditors (`moai bundle check` prints the floor), and every lane that registers a test file.

## Non-goals

- The audit tier of any other path, and the names-only judgment `gate::tier_of` makes for the pull request gate (`audit-gate`) and for the landing check: they still read paths alone and so keep tier 2 for the registry (the stricter side).
- `tests/areas.toml` and `tests/fixtures/state_folder_literals/allow.tsv`: both sit under `tests/`, so they are tier 0 for any change today; they get no rule (D-2).
- Judging whether a new scan entry is a good scan: a scan added is growth and is audited at tier 1 as any code change.

## Requirements

- REQ-1: When a change of `tools/repo_scans.json` only grows the registry (a rule file or a test name added to a scan, a `--test <name>` option added to a cargo line of its checks, a scan added), the audit floor shall not count that path as a tier 2 path.
- REQ-2: When a change of `tools/repo_scans.json` edits or removes anything the old text had (a text field, a list member, a check line, a scan), or adds a key, a check line or a check word of any other kind, the audit floor shall stay tier 2 · failure: none — the rule only decides whether the path prefix applies, it reads and writes nothing.
- REQ-3 (REQ-1): IF the old text, the new text or either side's JSON is missing or unreadable, or a repeated key overrides an old value, THEN the audit floor shall stay tier 2.
- REQ-4 (REQ-1): While another path of the same change is a tier 2 path, the audit floor shall stay tier 2.
- REQ-5: When a changed path is not exactly `tools/repo_scans.json`, the audit floor shall apply every tier 2 prefix to that path as before · failure: none — a path rule has no failure path.
- REQ-6: When `moai bundle check` judges a head and when `moai bundle body` writes its `audit-tier:` line, they shall read the floor with the same rule · failure: none — a head with no merge base is still the existing `no_merge_base` finding.

## Acceptance

- AC-1 (REQ-1): a registration (rule file and `--test` added; test name, second `--test` and a scan added; no change) gives floor 1 · verify: `cargo nextest run --locked --offline --test gate_registry_tier a_registration_keeps_the_floor_at_tier_1`
- AC-2 (REQ-2): an edited text field, a removed or replaced rule file or test name, a removed scan, a removed or replaced `--test`, any other added word, a check line added or removed, a key added each give floor 2 · verify: `cargo nextest run --locked --offline --test gate_registry_tier an_edited_or_removed_line_is_tier_2 a_check_line_changed_in_any_other_way_is_tier_2 a_registry_of_another_shape_is_tier_2`
- AC-3 (REQ-3): a missing side, text that is not JSON, and one added line that repeats a key to override an old value each give floor 2 · verify: `cargo nextest run --locked --offline --test gate_registry_tier the_names_alone_still_give_tier_2 a_line_that_only_adds_text_but_overrides_an_old_value_is_tier_2`
- AC-4 (REQ-4, REQ-5): a registration next to a tier 2 path gives 2 and next to a tier 0 or tier 1 path gives 1; the same texts under another path that starts with a tier 2 prefix give 2; `tests/areas.toml` and the allow list stay tier 0 · verify: `cargo nextest run --locked --offline --test gate_registry_tier another_tier_2_path_in_the_change_keeps_the_floor_at_tier_2 only_the_registry_path_is_judged_by_its_text the_two_other_registries_are_tier_0_for_any_change`
- AC-5 (REQ-6): in a temporary repository `moai bundle check` answers `audit_floor` 1 for a registration, 2 for an edit, a removal and a registration next to `src/gate.rs`, 0 for a document, and judges a head behind the tip by its own change · verify: `cargo nextest run --locked --offline --test gate_registry_tier bundle_check_reads_a_registration_as_tier_1_and_an_edit_as_tier_2 bundle_check_judges_a_head_behind_the_tip_by_its_own_registry_change`
- AC-6 (REQ-1): the 50 most recent audited ranges are recomputed with the new rule and the number of ranges whose floor changes is in the Result · verify: `moai spec check SPEC-20261124-registry-line-add-tier --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-registry-line-add-tier/**`, `src/gate.rs`, `src/bundle.rs` (the two floor calls, added by the leader's decision on the card, D-4), `tests/gate_registry_tier.rs`
- read-only: `src/audit_gate.rs`, `src/card/land/check.rs` (they keep `gate::tier_of`), `tools/repo_scans.json`, `tests/areas.toml`, `tests/fixtures/state_folder_literals/allow.tsv`, `docs/lane-rules.md`
- forbidden: any other change of `tools/repo_scans.json`; `CLAUDE.md`

## Plan

1. RED then GREEN in `tests/gate_registry_tier.rs`: the function cases first, then the `bundle check` scenes.
2. `src/gate.rs`: `ChangedFile`, `tier_of_changes` and `registry_only_grew`; `tier_of` keeps its signature and its names-only behaviour through the one shared loop.
3. `src/bundle.rs`: one helper reads the old and new text of the registry path and gives both call sites (`bundle check`, `bundle body`) the floor.
4. Guard mutants: one per new guard, on the committed head, restored by git.
5. Recompute the 50 most recent audited ranges with the old and the new rule.

## Risks

- A registration that is judged tier 1 is audited by one session that did not write it, without the full audit of tier 2. The exemption is narrow (growth only, one path, nothing else tier 2 in the change) and a scan added is the one open end: a new scan's check line is a command the check step runs, so the auditor of a tier 1 change reads it.
- The texts of the registry on both sides come from `git show`; a head with no merge base or a path that is gone on one side reads as no growth, tier 2.
- `gate::tier_of` stays names-only, so the pull request gate and the landing check can still say tier 2 for a head `moai bundle check` says tier 1; both only require a non-lander reviewer from tier 1 up, so a landing is not blocked by the difference.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | "Only adds" is read on the parsed registry, not on lines: a rule file, a test name, a `--test` option or a scan may be added; nothing the old text had may change | the card's rule says pure added lines, but a real registration changes one existing line (the check line gets `--test <name>`, a line removed and written again) and a pure line add can override an old value with a repeated key; measured on the 50 ranges: lines alone change the floor of 2, the parsed rule of 11 (Result table) | restore `registry_only_grew` to a line-only comparison, or drop the exemption (`tier_by` ignores the flag) |
| D-2 | `tests/areas.toml` and the allow list get no rule | both are under `tests/`: `tier0_safe` already makes any change of them tier 0, and raising an edit of them to tier 2 would add audit load, not remove it | none needed |
| D-3 | `gate::tier_of` keeps its signature; the pull request gate and the landing check keep judging names | they have no old and new text, and judging them tier 2 is the stricter side | give those two callers the texts in a later change |
| D-4 | `src/bundle.rs` is in the write scope | leader decision on the card (2026-11-24T14:05:26Z, delegated by the user), requested because the floor calls are there | drop `src/bundle.rs` and the floor of `moai bundle check` stays names-only |

## Result

(written after the checks)
