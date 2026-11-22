---
id: SPEC-20261122-home-update-notice-follow-ups
title: "the home update notice says how to remove the old files and the apply screen names the host"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261122-codex-home-update-removed
---

# the home update notice says how to remove the old files and the apply screen names the host

## Intent

The independent audit of the home update slice passed it and left three follow-ups that do not block (N2, N3 and N4 of its audit note). The "no longer managed" notice tells a person that old Codex and OMP files stay, but not that removing them is the person's own step, and the Korean screens show the English reason. The apply screen prints a skipped part without its host, while the preview screen prints both. One test lost its check of the confirm result, and two doc comments still describe hosts that the update no longer writes. This change closes the three without changing what the update writes.

## Requirements

- REQ-1: The unmanaged notice shall end its reason with the sentence that the person removes those files themselves if they do not use them, for the Codex notice and the OMP notice alike.
- REQ-2: When the update screen runs in Korean, it shall print the unmanaged notice row in Korean with the same two statements, that moai no longer manages the files and that the person removes them if unused. · failure: none
- REQ-3: The apply screen shall print the host beside the part name of every skipped row, as the preview screen already does.
- REQ-4: The notice shall stay outside the plan digest and keep one entry per host, so that no confirmation changes.
- REQ-5: The test of a blocked skill copy shall check that the confirmed update reports `applied`.
- REQ-6: The doc comments of the home routing module and the home style module shall name only the hosts the update writes.

## Acceptance

- AC-1 (REQ-1, REQ-4): the notice of each host in the preview ends with the removal sentence, and the plan digest is the same with and without the folders · verify: `cargo test --locked --offline --test home_update_notice_follow_ups the_notice_says_to_remove_the_old_files_yourself`
- AC-2 (REQ-2): the Korean preview and apply rows for an unmanaged entry hold the Korean words for "no longer managed" and "remove them yourself", and the English rows hold the English sentence · verify: `cargo test --locked --offline --test home_update_notice_follow_ups the_unmanaged_notice_reads_in_both_languages_and_names_the_host`
- AC-3 (REQ-3): the apply screen row of an unmanaged entry carries `unmanaged codex` and the apply row of a Claude skipped part carries its part and `claude` · verify: `cargo test --locked --offline --test home_update_notice_follow_ups the_unmanaged_notice_reads_in_both_languages_and_names_the_host`
- AC-4 (REQ-5): the blocked skill copy test reads the confirm result and asserts `applied` is true · verify: `cargo test --locked --offline --test update_user_layer_cli a_blocked_skill_copy_gets_no_agents_and_no_section`
- AC-5 (REQ-6): the two doc comments no longer name the Codex folder as a written target · verify: `git grep -n "codex" -- src/home_routing.rs src/home_style.rs` shows only the lines that say the update no longer writes it

## Scope

- write: only the named lines of `src/update.rs` (the two notice reasons), `src/update_tux.rs` (the two skipped-row renderers), `src/home_routing.rs` and `src/home_style.rs` (one doc comment each) and `tests/update_user_layer_cli.rs` (the confirm result check), and the new file `tests/home_update_notice_follow_ups.rs` and this SPEC; those files are held by other cards and the leader granted these lines only
- read-only: everything else, in particular `tests/codex_home_update_removed.rs`, `tests/update_tux_golden.rs` and the golden files under `tests/fixtures` (no golden holds a skipped row)
- forbidden: any file under the user's Codex, shared agents or OMP folders on this machine; the real HOME (every test runs in a temporary home)

## Plan

1. RED: add the removal sentence to the notice cases, add the screen case for both languages and both screens, and restore the confirm result check; the first two fail on the tip.
2. GREEN: extend the two notice reasons, add the Korean form and the host label in the two skipped-row renderers, and fix the two comments.
3. Run the new test file, the notice test file, the golden test file, the user layer test file, clippy, rustfmt and `moai spec check`.

## Result

RED afb0ce8825, GREEN 39c3850f71, scan row b8ced3e49e. The leader granted the named lines of the held files (LEADER DECISION on LANE-BLOCKED 6282, extended on 6441 for the one EXEMPT row in `tests/repo_scans_test.py`).

- AC-1 met: `the_notice_says_to_remove_the_old_files_yourself` failed first (the Codex reason ended at "no longer manages them") and passes now; `home_update_notice_follow_ups` 2 passed, `codex_home_update_removed` 12 passed (green run, `target/green.log`).
- AC-2 met: `the_unmanaged_notice_reads_in_both_languages_and_names_the_host` failed first (no row `unmanaged codex` on the apply screen) and passes now in the same run; `update_tux_golden` 14 passed, no golden changed.
- AC-3 met: the same test checks `unmanaged codex`, `unmanaged omp` and `agents claude` on the preview and the apply screens in both languages.
- AC-4 met: `a_blocked_skill_copy_gets_no_agents_and_no_section` reads the confirm result again and passes; `update_user_layer_cli` 11 passed.
- AC-5 met: `git grep -n -i codex -- src/home_routing.rs src/home_style.rs` prints five lines, each saying the Codex file is no longer written.
- Checks: clippy `--all-targets -D warnings` exit 0; rustfmt check clean; `moai spec check` 0 errors. `moai check affected fd56877af5..39c3850f71` (70 targets) failed on the walking-file scan, fixed by b8ced3e49e (`tests/repo_scans_test.py` 6 passed by hand), and on `update_route` 1 test and `update_end_to_end_cli` 2 tests, which this diff does not touch; the leader checks those on the integration tip.
- Not run: check affected after b8ced3e49e, the whole suite, and the independent audit.
