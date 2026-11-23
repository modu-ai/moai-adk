---
id: SPEC-20261123-update-adopt-unowned
title: "moai update adopts an unowned skill folder whose files match the package or an earlier generation, and finds the version folders under either state name"
status: draft
revision: 1
created: 2026-11-23
---

# moai update adopts an unowned skill folder whose files match the package or an earlier generation, and finds the version folders under either state name

## Intent

On 2026-11-23 the update preview of build 20 skipped every part of the `claude` host: "unmanaged existing package: 43 file(s) present in the projection root that install receipt … does not own". The folder `~/.claude/skills/moai` had been put there by hand, and the leader replaced it by hand after a backup.

Two causes were measured:
- A file the receipt does not own blocks the host even when its bytes equal the package's file. Nothing in the plan separates "a copy of ours" from "a file of the user".
- Build 20 sits in `~/.moai/versions/<v>/bin/`, while the HOME's state folder is `~/.moai-cli` (the Go generation's `~/.moai` has no Rust marker). `bin-names-plan` then says "not in the version folder layout ~/.moai-cli/versions", and the update plan does not read the receipts of the other version folders. So every new version folder starts without a baseline and every file it finds is unowned.

This SPEC adds an adoption to the update plan and makes the version folder layout one thing.

Adoption:
- A file the receipt does not own is adoptable when its sha256 equals the package's file of that path, or the same path's file in an earlier generation: the package of another installed version folder next to the running build's own.
- The preview lists the adopted files. Only `moai update --confirm <plan_digest>` (or the yes at the terminal question) applies them.
- The apply copies the adopted files and a manifest into a new backup folder, then writes the receipt that owns them. The files themselves are not rewritten by the adoption; a file of an earlier generation is then updated by the usual three-way path, with its known-good copy.
- A file that matches nothing is a file of the user. The host stays blocked as before, and the existing backup-then-replace remains the way out.

Readers: the person who runs `moai update` on a HOME with a hand-copied or foreign-receipt skill folder; the leader who installs builds by hand.

## Non-goals

- Adopting a file whose bytes match nothing, or deleting any file. Mixed folders stay blocked.
- A new command or a new flag. The confirmation is the existing `--confirm <plan_digest>`.
- Moving `~/.moai-cli` to `~/.moai`. `home-state-plan` and `home-state-apply` own that move.
- Changing where the receipts and backups are written: they stay under the HOME's state folder.
- Agents, routing and style parts of `moai update`. They follow the skill copy as before.

## Requirements

- REQ-1: When an unowned file of the skill folder has the sha256 of the package's file of that path, the update plan shall list it as adopted.
- REQ-2: When an unowned file has the sha256 of the same path in another installed version folder's package of a build inside the layout, the update plan shall list it as adopted with the generation it matched.
- REQ-3 (REQ-1, REQ-2): IF an unowned file that the package carries matches neither the package nor an earlier generation, THEN the update plan shall keep the host blocked with nothing adopted.
- REQ-4: While no confirmation is given, the update command shall write nothing for an adoption.
- REQ-5: When a plan with adopted files is confirmed, the update command shall copy the adopted files and a manifest into a new backup folder before it writes any receipt.
- REQ-6 (REQ-5): When the backup is complete, the update command shall write the receipt that owns the adopted files at their present sha256 and then apply the other actions of the plan.
- REQ-7 (REQ-5, REQ-6): IF an adopted file differs from the previewed plan when the confirmed run starts, THEN the update command shall refuse and write nothing.
- REQ-8 (REQ-6): When an adoption has been applied, the next update preview shall list no adoption for that host. · failure: none — the host is then planned by the usual three-way path.
- REQ-9: When the running binary sits in `<home>/.moai/versions/<v>/bin/` or `<home>/.moai-cli/versions/<v>/bin/`, the command-name plan and the update plan shall treat it as inside the version folder layout.
- REQ-10 (REQ-9): When the binary is inside the layout, the update plan shall read the host receipts of the sibling version folders of the same `versions` folder as the baseline. · failure: none — a folder or receipt that cannot be read lends nothing and the plan is as before.
- REQ-11 (REQ-9): IF the binary is in neither folder, THEN the command-name plan shall say it is outside the layout and name both folders.

## Acceptance

- AC-1 (REQ-1, REQ-4): in a temporary HOME whose skill folder is a byte copy of the package and has no receipt, the preview shows the host not blocked with every file adopted from the package, and the skill folder, the receipts and the backups are byte for byte as before. · verify: `cargo nextest run --locked --offline --test update_adopt_cli`
- AC-2 (REQ-2, REQ-3): a skill folder copied from an earlier version folder's package shows each file adopted with that generation and the rest of the plan as `update`; one file edited by the user keeps the host blocked with nothing adopted and names the file; a build outside the layout and a linked version folder lend no generation, so the same folder stays blocked. · verify: `cargo nextest run --locked --offline --test update_adopt_cli`
- AC-3 (REQ-5, REQ-6, REQ-7): `moai update --confirm <plan_digest>` creates one backup folder holding the manifest and a copy of each adopted file, then the receipt owns them, and the skill folder equals the package. A file changed after the preview is refused by the confirmed run and nothing is written. · verify: `cargo nextest run --locked --offline --test update_adopt_cli`
- AC-4 (REQ-8): after the confirmed run, a second preview lists no adoption, does not block the host and has nothing to apply. · verify: `cargo nextest run --locked --offline --test update_adopt_cli`
- AC-5 (REQ-9, REQ-10, REQ-11): with the binary in the HOME's `.moai/versions/a/bin/` folder on a HOME whose state folder is `.moai-cli`, `bin-names-plan` is not `outside_layout`, and the update plan reads the receipt of a sibling version folder as its baseline; a binary elsewhere is `outside_layout` and the reason names both folders. · verify: `cargo nextest run --locked --offline --test update_adopt_cli --test bin_names_cli`

## Scope

- write: `.moai/specs/SPEC-20261123-update-adopt-unowned/spec.md`
- write: `src/update.rs`
- write: `src/bin_names.rs`
- write: `tests/update_adopt_cli.rs` (new)
- read-only: `src/home_root.rs`, `src/apply.rs`, `src/recovery.rs`, `src/install.rs`
- forbidden: the real HOME and every file outside this checkout; every test runs in a temporary HOME

## Plan

1. This SPEC.
2. Tests first (RED): `tests/update_adopt_cli.rs`, AC-1 to AC-5 in temporary HOMEs.
3. `src/bin_names.rs`: one function that names the `versions` folder of a binary under either state name. `src/update.rs`: use it in the layout check; add the adoption to the plan and the apply.
4. Light checks, one affected run, READY with head, base and the temporary-HOME run of preview, confirm and the next update.

## Risks

- A hash match is evidence that a file is ours, not that nobody meant to keep it. The adoption writes no skill byte; the later update of an earlier-generation file keeps a known-good copy, and the backup folder holds every adopted file.
- Reading many version folders and receipts costs time. The reads are bounded by the same limits as the projection scan, and a folder or receipt that cannot be read lends nothing.
- Accepting `~/.moai/versions` on a HOME with a Go-generation `~/.moai` only takes the folder the running binary really sits in; no other path under it is read for layout.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The confirmation of an adoption is the existing `--confirm <plan_digest>`; no new flag. | The card asks for the preview by default and a write only with the confirm token; the digest binds the adoption list. | Add a token of its own in a later card. |
| D-2 | The adoption is all or nothing per host. | A folder with a file of the user is not a copy of ours; the existing blocked path and backup-then-replace stay its way out. | Allow a partial adoption in a later card. |
| D-3 | The adoption is not an entry of the receipt, and the receipt carries no `backup`. | Recovery reads `backup` as a backup-then-replace and would move a folder back; a failed apply keeps the pre-apply owned list, so the adoption simply does not happen. | None needed. |
| D-4 | Earlier generations are the packages of sibling version folders of an in-layout build, not the `owned` lists of receipts. | The existing rule of #2521 and its tests: a build outside the layout and a linked folder or receipt directory lend nothing; a receipt list outside that rule would lend through them. | Add the receipt lists with the same layout and link rules in a later card. |

## Result

- RED first: `cargo nextest run --locked --offline --no-fail-fast --test update_adopt_cli` ran before the code: 10 run, 8 failed, 2 passed (the two that already held: a file of the user stays blocked, a changed plan is refused).
- After the code and the repairs below: `update_adopt_cli`, `update_moved_receipts`, `state_folder_literal_scan` and `update_version_folders` ran together: 37 run, 37 passed.
- The wider selection (`update_*`, `bin_names_cli`, `codex_home_update_removed`, `install_no_dead_ends`, `recovery_cli`, `home_update_notice_follow_ups`, `retired_hosts_layer`, `skill_name_collision`, `state_folder_literal_scan` and the library tests) first failed in 6 of 2035 tests, then in 3 of 2042; the 3 are repaired below and were run again in the line above.
- Repairs the runs forced: the state-folder constant scan of `home_root` (the helper names no constant now); `update_version_folders` (the sources are the packages of sibling version folders of a build inside the layout, as the receipts already were, D-4); `update_moved_receipts` (two helpers now tell a baseline lent by a receipt from an adoption: `all_unchanged` requires no `adopt`, `blocked_unmanaged` accepts an adoption that lent no `receipt_from`); `state_folder_literal_scan` (one allow row for the new test file, which spells each name once through two constants). The last two files lie outside the card's first scope; it was widened with `card scope` and the reason on the card.
- `cargo fmt --check` exit 0, `git diff --check` exit 0, `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0 (it first found one `cmp_owned`, repaired).
- Run by hand in a temporary HOME with the debug binary of this head (state folder the earlier name, the binary in the new-name `versions` folder, a skill folder copied from an earlier version folder's package): the preview left the HOME tree byte for byte as it was (same tree digest before and after) and showed 46 files adopted (matched `package` or `generation:version folder old`) with 45 `unchanged` and 1 `update`; `moai update --confirm <plan_digest>` made one `<stamp>-claude-adopt` backup folder with `MANIFEST.json` and 46 copies, wrote the committed receipt and left the skill folder equal to the package; the next `moai update` showed `preflight` ok, no `adopt`, 46 `unchanged`.
- AC-1 met: `a_byte_copy_of_the_package_is_adopted_in_the_preview_and_nothing_is_written`.
- AC-2 met: `a_copy_of_an_earlier_version_folder_is_adopted_with_the_generation_it_matched`, `one_file_of_the_user_keeps_the_host_blocked_with_nothing_adopted`, `a_build_outside_the_layout_lends_nothing_and_stays_blocked`, `a_linked_version_folder_lends_nothing_and_stays_blocked`.
- AC-3 met: `the_confirmed_run_backs_up_then_owns_the_adopted_files`, `a_file_changed_after_the_preview_is_refused_and_nothing_is_written`.
- AC-4 met: `the_next_update_lists_no_adoption_and_does_not_block`.
- AC-5 met: `a_binary_in_the_moai_versions_folder_is_inside_the_layout`, `a_binary_elsewhere_is_outside_the_layout_and_the_reason_names_both_folders`, `the_receipt_of_a_sibling_folder_under_moai_versions_is_the_baseline`.
- Not tested: the second check of an adopted file just before the receipt commits (a race inside one apply; the stale-plan refusal at the start is tested). The real HOME was not read or written by any test or run.
