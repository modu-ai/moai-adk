---
id: SPEC-20261122-remove-codex-shipped-text
title: "The shipped skill text, docs, forms and tests name Claude Code and GLM only"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-remove-codex-support
---

# The shipped skill text, docs, forms and tests name Claude Code and GLM only

## Intent

The user decided on 2026-11-21 that moai-adk supports Claude Code and GLM only. The launcher, the model settings, the project layer and the install texts no longer support Codex or OMP (the earlier slices of the parent SPEC). This slice takes the Codex and OMP words out of what MoAI ships and shows: the skill package, the current product docs, the issue forms and labels, the root docs, and the tests, scripts and fixtures that exist only for Codex or OMP. A word scan then keeps them out.

## Non-goals

- Lane, goal, leader and `CODEX_THREAD_ID` code and their operation docs: card C-313.
- The kept readers of the leader decisions D1–D7: the retired-host readers, the Go-era `.codex` classification, the cleanup detection, the worktree protection, the gate paths, `Host::Codex` and the asset slugs, and their tests.
- The bytes of `templates/bundle-packs` (D7), `install/moai-switch.sh` (pinned by `V2_SHA256`), the v1 shim snapshot and the recorded home-style renders.
- Dated history: `docs/design-review/**` except the current operation pages, `docs/evidence/**`, `docs/surveys/**`, release notes, CHANGELOG, SPEC and card files, `.moai-cli/reports/**`.
- A word scan over `src` (a later card), withdrawing the Codex-era cards (the leader), regenerating this repository's own profile, AGENTS and generation manifest (the step after this card), and Python tool ports (C-317).

## Requirements

- REQ-1: The skill package under `skills/moai/` shall hold no Codex or OMP entry spelling or host instruction outside the scan's allow-list.
- REQ-2: The interview question bank shall hold no row that asks which hosts the project uses.
- REQ-3: The README files and the Cargo description and the root policy files shall name Claude Code and GLM as the supported hosts.
- REQ-4: The repository issue forms and their product templates shall offer no Codex choice.
- REQ-5: The repository label list and its product template shall list `host:codex` under `not_used`.
- REQ-6: The current product docs named by the scan shall hold no Codex or OMP support claim.
- REQ-7: The repository shall carry no test or script or fixture that exists only for Codex or OMP.
- REQ-8: The repository shall track no file under its own `.codex` folder.
- REQ-9: When a text file named by the scan holds a retired host word outside the allow-list, the scan test shall fail with the file and the line. · failure: none — the failure is the required behavior
- REQ-10: When a file of the skill package changes, the change shall carry the source manifest that the manifest tool writes for it. · failure: none — a stale manifest fails the existing manifest check
- REQ-11: The error-kind table shall hold no message of the removed hook adapters.
- REQ-12: The change shall leave the kept readers of the leader decisions D1–D7 and their tests unchanged.

## Acceptance

- AC-1 (REQ-1, REQ-9): the scan reports no retired host word in the skill package outside the allow-list · verify: `cargo test --locked --offline --test retired_host_words skill_package_has_no_retired_host_words`
- AC-2 (REQ-2): the question bank has no `hosts` row, and a profile that lists codex still plans · verify: `cargo test --locked --offline --test question_bank_hosts`
- AC-3 (REQ-3): the root docs hold no retired host word, and the Cargo description equals the README opening sentence · verify: `cargo test --locked --offline --test retired_host_words root_docs_have_no_retired_host_words`
- AC-4 (REQ-4): the forms and their templates offer no Codex option · verify: `cargo test --locked --offline --test retired_host_words github_forms_have_no_retired_host_words`
- AC-5 (REQ-5): `host:codex` sits under `not_used` in both label lists and the label axes still parse · verify: `cargo test --locked --offline --test labels_check`
- AC-6 (REQ-6): the named product docs hold no retired host word · verify: `cargo test --locked --offline --test retired_host_words product_docs_have_no_retired_host_words`
- AC-7 (REQ-7): the Codex-only and OMP-only files are absent · verify: `cargo test --locked --offline --test retired_host_words codex_only_files_are_absent`
- AC-8 (REQ-8): the two tracked `.codex/agents` files are absent · verify: `cargo test --locked --offline --test retired_host_words the_repository_tracks_no_codex_folder`
- AC-9 (REQ-9): a planted word is reported with its file and line, and `compose`, `$MOAI_LANGUAGE` and `"$moai"` are not · verify: `cargo test --locked --offline --test retired_host_words a_planted_word_is_reported_with_file_and_line`
- AC-10 (REQ-10): the source manifest equals the tool's output · verify: `cargo test --locked --offline --test product_cli source_manifest_covers_every_package_file`
- AC-11 (REQ-11): the contract table holds no removed hook adapter message · verify: `cargo test --locked --offline --lib contract`
- AC-12 (REQ-12): the kept readers' tests pass unchanged · verify: `cargo test --locked --offline --test go_state_codex_cli --test update_cleanup_cli --test gate_check_cli --test retired_hosts_layer --test codex_home_update_removed --test switch_script_v2`

## Scope

- write (own reservation): this SPEC folder
- write (to add per slice, no holder now): `.github/ISSUE_TEMPLATE/moai-bugfix.yml`, `.github/ISSUE_TEMPLATE/moai-feature.yml`, `.github/moai/labels.yml`, `.github/moai/README.md`, `templates/github/**`, `src/labels_check.rs`, `CONTRIBUTING.md`, `SECURITY.md`, `SUPPORT.md`, the current product docs named in the plan, `skills/moai/references/internal/interview.md`, `git-safety.md`, `commit-trailers.md`, `tests/question_bank_hosts.rs`, `tests/update_always_asks.rs`, `tests/host_matrix_test.py`, `tests/e2e_receipt_not_run_test.py`, `tests/e2e_host_fixture.rs`, `tools/ci_scope.py`, `tests/ci_scope_test.py`, `src/contract.rs`, `.codex/agents/*.toml`, `scripts/benchmark.sh`, `experiments/rc-cli/src/main.rs`, the new `tests/retired_host_words.rs`
- write (held by other cards; after they land or on the leader's word): the skill files and mode tests of C-329, `skills/moai/workflows/update.md` and `tools/source_manifest.py` (C-328), `skills/moai/references/internal/questions.md` and `tests/codex_question_tool.rs` (C-253), `README.md` and `README.ko.md` (C-344), `Cargo.toml` (C-279), `tools/repo_scans.json` (C-327), `skills/moai/references/source-manifest.json` (by line lease), `install/install.ps1` (C-316)
- read-only: everything else

## Plan

1. S1 GitHub forms and labels: the two issue forms and their four templates drop the Codex choices; `host:codex` moves under `not_used` in both label lists; the forms' manifest digests follow; the label unit test uses `host:claude`.
2. S4a interview: the `hosts` row leaves the question bank; `tests/question_bank_hosts.rs` pins that the row is gone and that a profile with codex still plans; the source manifest is rewritten.
3. S6 and S7: the removed hook adapter messages leave the `src/contract.rs` table; the two tracked `.codex/agents` files are deleted.
4. S3 product docs: `docs/codex-adapter-contract.md` is deleted with its `ci_scope.py` entry and `host_matrix_test.py` rows; the Codex rows leave the support, e2e and live-session docs with their tests; the Codex-only e2e script and its verdict test and the OMP host survey test are deleted (after their holder gives them back).
5. S2 root docs (after C-344 and C-279): README, README.ko, the Cargo description in the same commit as the README opening, and the root policy files.
6. S4b and S4c skill text (after C-329, C-328 and C-253 give their paths back): the entry spellings, host instructions and adapter columns go, with the mode tests that pin them and a rewritten source manifest.
7. S5 leftovers: the review comments in `install/install.ps1`, the benchmark profile fixture and the experiment line.
8. S8 the word scan: the new `tests/retired_host_words.rs` reads named lists only (the source manifest's files, the GitHub template manifest and a named array of root, install and current-doc paths; no directory walk) with an allow-list, plus its `tools/repo_scans.json` entry.

## Risks

- The source manifest is written by six cards; each skill slice rewrites it under a line lease right before its merge.
- Docs that tests read (`host-support-matrix.md`, `support-matrix.md`, `e2e-matrix.md`, `protocol-v1.md`, `agent-model-effort-inheritance.md`, `llms.txt`) change only with those tests; the goal row of `host-support-matrix.md` stays until C-313 lands.
- A changed GitHub template shows an update diff in user projects with the managed forms.
- An open issue still labelled `host:codex` becomes a `not_used` finding.

## Decisions

- D-1: The bundle packs keep their bytes and go on the scan's allow-list (leader D7) (planning default).
- D-2: The interview `hosts` row is deleted; the profile still filters codex and omp on read (planning default).
- D-3: The skill files C-329 holds wait for C-329 to land and give them back; this card then edits them from the new tip (planning default).
- D-4: `skills/moai/references/rule-history.md` keeps its dated rows and is on the allow-list (planning default).
- D-5: "Codex review" provenance comments stay in src and tests; the three in `install/install.ps1` become "review P1/P2 on #2943" (planning default).
- D-6: The one migration line is a sentence in `skills/moai/workflows/update.md` that relays the preview's skipped notice (planning default).
- D-7: The codex rows of `src/audit.rs` `INDEPENDENCE` and `render_omp` in `src/agent_defs.rs` are left to the src card and C-329 (planning default).
- D-8: Python tools that name Codex are left to C-317, except the `tools/host_live_harness.py` rows that go with the live-session doc (planning default).
