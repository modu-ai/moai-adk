---
id: SPEC-20261124-installed-version-by-name
title: "moai update reads the installed version when the executable name is moai or moai-cli"
status: draft
revision: 2
created: 2026-11-24
---

# moai update reads the installed version when the executable name is moai or moai-cli

## Intent

An installation id is the hash of the binary path, and a build can run under two names: `moai` and `moai-cli` (hard-linked in the version folder). The earlier version folders were read only under the running build's name, so a receipt written under the other name was never found, the update screen said `installed none`, and the whole skill folder was adopted again with every new build. The update must read the installed version whichever of the two names wrote the receipt.

## Requirements

- REQ-1: When an earlier version folder holds a finished receipt under the `moai` name or the `moai-cli` name, the update baseline reader shall read that receipt as the installed version.
- REQ-2 (REQ-1): IF a receipt of an earlier version folder cannot be read or names another installation, THEN the update baseline reader shall block the host and adopt nothing.
- REQ-3: When both names of one earlier version folder hold a finished receipt, the update baseline reader shall take the newer one for that folder · failure: none — one folder never ties with itself, and two different folders finishing in one second stay blocked as before.
- REQ-4: When the baseline reader finds an earlier receipt, the update screen header shall read `installed <earlier> → package <new>` · failure: none — without a receipt the header keeps `none` as before.
- REQ-5: When the extra-skill baseline reader looks at an earlier version folder, it shall read the receipts under both executable names as REQ-1 and REQ-3 say · failure: none — an unreadable receipt blocks it as REQ-2 says for the host reader.
- REQ-6: When an id candidate of an earlier folder equals the running build's own install id, the update baseline reader shall skip that candidate · failure: none — the own receipt is read first.

## Acceptance

- AC-1 (REQ-1, REQ-6): in a temporary HOME `versions/A/bin/moai` and `versions/A/bin/moai-cli` hold different bytes and the receipt sits under the `moai` name id; after `versions/B/bin/moai` is hard-linked to `versions/B/bin/moai-cli` the plan from B names A's version as `source_version`, carries `receipt_from`, and has no `adopt` · verify: `cargo test --locked --test update_installed_by_name`
- AC-2 (REQ-4): the update screen of that plan starts with `installed <A version> → package <B version>` and never `none` · verify: `cargo test --locked --test update_installed_by_name`
- AC-3 (REQ-3): when both names of folder A hold a finished receipt the newer one is the baseline and the plan is not blocked as ambiguous · verify: `cargo test --locked --test update_installed_by_name`
- AC-4 (REQ-2): an unreadable receipt under the `moai` name of folder A blocks the host and adopts nothing · verify: `cargo test --locked --test update_installed_by_name`
- AC-5 (REQ-5): a build that runs as `bin/moai` alone reads the html-report receipt its earlier folder wrote under the `moai-cli` name, carries `receipt_from` and leaves every file unchanged · verify: `cargo test --locked --test update_user_layer_cli`

## Scope

- write: `.moai/specs/SPEC-20261124-installed-version-by-name/**`
- write: `src/update.rs` (the earlier-folder id list and one new helper beside `earlier_install_ids`)
- write: `src/update/extra_skill.rs` (the same id list)
- write: `tests/update_installed_by_name.rs`
- write: `tests/update_user_layer_cli.rs` (one test beside the existing html-report receipt test)
- write: `tests/fixtures/state_folder_literals/allow.tsv` (one row for the new test file)

## Plan

1. Add a helper that gives the install id of an earlier folder under the other executable name (`moai-cli` ↔ `moai`, `.exe` kept).
2. Read both ids of each earlier folder in `resolve_host_receipt` and in the extra-skill reader; the newest finished receipt stands for the folder.
3. Write the temporary-HOME tests that run copies of the built binary as child processes; the real `~/.moai` and `~/.claude` are never touched.
