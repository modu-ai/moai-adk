---
id: SPEC-20261120-plan-output-write-set
title: "the plan names the files a change writes, the shared one-line spots, the tests that read the changed text and their current holders; a port starts with a verdict parity table"
status: draft
revision: 1
created: 2026-11-20
---

# the plan names the files a change writes, the shared one-line spots, the tests that read the changed text and their current holders; a port starts with a verdict parity table

## Intent

Lanes were stopped again and again by files they had not known they needed: the help table, the operation map or the module list of a new command, and the fixed tests that read a sentence a change rewrites, each held by another card's reservation. The planning agent finds them before the work starts: its output lists the write set, the shared one-line spots, the tests that read the changed text and who holds each file now. A port or rewrite of a tool starts with a verdict parity table against the old tool, because unit tests alone let a port differ from the old tool on real inputs. The chain table gains the two paths that were missing: a flaky test, and a repair after a FAIL that goes to the same auditor.

## Requirements

- REQ-1: The output of the moai-plan agent shall list under `Write set` every file the plan changes, by exact path.
- REQ-2: The output of the moai-plan agent shall list under `Shared spots` every file that several changes add one line to (a module list, a help table, an operation map, a golden file, a manifest) and that the plan also needs, with the line it needs.
- REQ-3: The output of the moai-plan agent shall list under `Tests that read changed text` the tests that read a sentence of a file in the write set, found by searching the tests for the path.
- REQ-4: When the project has a card ledger, the moai-plan agent shall name under `Current holders` the card that holds each file of the write set and of the shared spots, or `free` when none does.
- REQ-5 (REQ-4): IF the project has no card ledger, THEN the moai-plan agent shall write `no ledger` under `Current holders`.
- REQ-6: The check-completeness reference shall say that a port or rewrite keeps the old tool's verdicts by running both over the same real inputs and reporting every input whose verdict differs.
- REQ-7: The chains index shall carry a row for a test that fails on one run and passes on a rerun of the same head and a row for a repair after a FAIL verdict.
- REQ-8: The chains index shall say that the re-audit after a FAIL goes to the first auditor, who judges the repair again and every changed criterion.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): the `Output` section of `agents/plan.md` lists `Write set`, `Shared spots`, `Tests that read changed text` and `Current holders` in that order after `Steps` and before `Files written` · verify: `cargo test --locked --offline --test plan_output_sections output`
- AC-2 (REQ-4, REQ-5): the `Verification` section of `agents/plan.md` asks for the current holder of each listed file from the card ledger and names the answer when there is none · verify: `cargo test --locked --offline --test plan_output_sections holders_are`
- AC-3 (REQ-6): `references/internal/check-completeness.md` has the parity sentence in its seventh item · verify: `cargo test --locked --offline --test plan_output_sections parity`
- AC-4 (REQ-7, REQ-8): the chain table has ten rows that name real routes, agents and references, the two new rows are there, and the re-audit rule is stated · verify: `cargo test --locked --offline --test plan_output_sections chains`; `cargo test --locked --offline --test agent_defs_package chains`
- AC-5 (REQ-1, REQ-6, REQ-7): the manifest carries the new hashes and the tests that read these files pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test`; `cargo test --locked --offline --test agent_defs_package --test product_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-plan-output-write-set/spec.md`
- write: `tests/plan_output_sections.rs`
- write: `agents/plan.md`
- write: `skills/moai/references/internal/check-completeness.md`
- write: `skills/moai/references/chains/index.md`
- write: `skills/moai/references/source-manifest.json`
- write: `tests/agent_defs_package.rs` (one number, the row count of the chain table, 8 to 10)

## Decisions

- D-1: The write set and the shared spots are listed by the plan, not enforced by it: the plan names them and the lane reserves them, the way a plan already names the check of each criterion. Reversible: a later change may make `card take` read them.
- D-2: The holders are read from the card ledger when there is one, because that is the only record of a reservation; a project without a ledger writes `no ledger` and the lane asks the leader.
- D-3: The re-audit rule is stated in the chains index, the one place that orders the agents of a request; no other file says it yet (`internal/review` only says a re-audit checks the listed repairs and sorts what is new), so the chains index is where the rule is first written.

## Plan

1. Tests first (RED): `tests/plan_output_sections.rs` with the output, holders, parity and chains cases; the row count pinned by `chains_name_real_routes_agents_and_references` in `tests/agent_defs_package.rs` (held by another card; asked to change the one number from 8 to 10) moves with the table; see the new file fail on the unchanged text.
2. `agents/plan.md`: the three output sections and the holders section, one Verification bullet; `check-completeness.md` item 7: the parity sentence; `chains/index.md`: two rows and two rules.
3. Rewrite the manifest hashes, then run the new file, the tests that read these files, the python manifest tests, clippy and rustfmt, then audit and write the Result.

## Result

- AC-1: met — `cargo test --locked --offline --test plan_output_sections output` passed: the `Output` section of `agents/plan.md` lists `Steps`, `Write set`, `Shared spots`, `Tests that read changed text`, `Current holders` and `Files written` in that order, with the exact-path, shared-spot and search-the-tests phrases. The four cases of the file failed on the unchanged text first (0 passed, 4 failed, each for the missing text) and pass now (4 passed).
- AC-2: met — `--test plan_output_sections holders_are` passed (1 test): `Verification` reads the current holder of each listed file from the card ledger and writes `free` for a file no card holds and `no ledger` when there is none; it names `moai card list` and `moai card show` as the two commands that only read the ledger.
- AC-3: met — `--test plan_output_sections parity` passed: item 7 of `references/internal/check-completeness.md` says a port or rewrite keeps the old tool's verdicts over the same real inputs and reports every input whose verdict differs unless the SPEC names it as intended.
- AC-4: met — `--test plan_output_sections chains` passed (the table has ten rows, the flaky-test row names VERIFY, the repair row names RUN, VERIFY, SYNC, AUDIT, `moai-impl` and `moai-audit`, the first-auditor rule is there); `tests/agent_defs_package.rs` pins ten rows now and its 19 tests pass.
- AC-5: met — `python3 -m unittest` of `source_manifest_test` (45), `skill_references_test` (5), `prop_manifest_test` (2), `skill_web_pages_test` (4), `repo_scans_test` (25) and `ci_scope_test` (47) is OK in an empty temporary HOME; the Rust readers (`agent_defs_package`, `product_cli`, `workflow_size_budget`, `mode_workflow_steps` 43, `core_rules_2876`, `sdd_guidance`, `home_routing_cli`, `mode_cards_53`, `lane_next_rules`, `self_audit_cli`, `update_tux_golden`) and `--lib` (1641 passed) and `--bins` are 15 `test result: ok` lines with 0 failed; `cargo clippy --locked --offline --all-targets -- -D warnings` exit 0 with no warning; `rustfmt --edition 2021 --check` of the two Rust test files is clean; `spec check` 0 errors 0 warnings; the manifest check is `ok`.
- Not done here: the audit's three sentence mutations that no test caught (the `The lane reserves these and no others.` sentence, the `no ledger` alternative in the `Output` list and the give-back sentence of `Shared spots`) stay as meaning that a reviewer reads, not pinned by a test; `skill_text_citations` is not in this branch's base, so the shipped-text scan did not run on these files (it runs in the merge check of the area); no card ledger is needed to read the plan, but nothing yet checks that a plan's `Current holders` is true.
- Repair after the first audit (B1 to B3, N1, N2, N4): the AC-4 verify is two commands, AC-5 no longer names `skill_text_citations`, the Scope lists `tests/agent_defs_package.rs`, D-3 says the chains index is the first place that writes the re-audit rule, `Write set` includes the shared spots and tests the plan edits, `Current holders` has `free`, the ledger commands are named, and the repair row runs VERIFY and SYNC before AUDIT. Run again after it: 12 reader targets ok (agent_defs_package 19, core_rules_2876 6, home_routing_cli 11, lane_next_rules 3, mode_cards_53 9, mode_workflow_steps 43, plan_output_sections 4, product_cli 79, sdd_guidance 7, self_audit_cli 5, update_tux_golden 14, workflow_size_budget 6), python source_manifest_test, skill_references_test, repo_scans_test and ci_scope_test 122 tests OK in an empty temporary HOME, manifest check ok.
