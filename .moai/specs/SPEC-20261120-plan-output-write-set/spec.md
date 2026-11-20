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
- REQ-4: When the project has a card ledger, the moai-plan agent shall name under `Current holders` the card that holds each file of the write set and of the shared spots.
- REQ-5: IF the project has no card ledger, THEN the moai-plan agent shall write `no ledger` under `Current holders`.
- REQ-6: The check-completeness reference shall say that a port or rewrite keeps the old tool's verdicts by running both over the same real inputs and reporting every input whose verdict differs.
- REQ-7: The chains index shall carry a row for a test that fails on one run and passes on a rerun of the same head and a row for a repair after a FAIL verdict.
- REQ-8: The chains index shall say that the re-audit after a FAIL goes to the first auditor, who judges the repair again and every changed criterion.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): the `Output` section of `agents/plan.md` lists `Write set`, `Shared spots`, `Tests that read changed text` and `Current holders` in that order after `Steps` and before `Files written` · verify: `cargo test --locked --offline --test plan_output_sections output`
- AC-2 (REQ-4, REQ-5): the `Verification` section of `agents/plan.md` asks for the current holder of each listed file from the card ledger and names the answer when there is none · verify: `cargo test --locked --offline --test plan_output_sections holders`
- AC-3 (REQ-6): `references/internal/check-completeness.md` has the parity sentence in its seventh item · verify: `cargo test --locked --offline --test plan_output_sections parity`
- AC-4 (REQ-7, REQ-8): the chain table has ten rows that name real routes, agents and references, the two new rows are there, and the re-audit rule is stated · verify: `cargo test --locked --offline --test plan_output_sections chains --test agent_defs_package chains`
- AC-5 (REQ-1, REQ-6, REQ-7): the manifest carries the new hashes and the tests that read these files pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test`; `cargo test --locked --offline --test agent_defs_package --test skill_text_citations --test product_cli`

## Scope

- write: `.moai/specs/SPEC-20261120-plan-output-write-set/spec.md`
- write: `tests/plan_output_sections.rs`
- write: `agents/plan.md`
- write: `skills/moai/references/internal/check-completeness.md`
- write: `skills/moai/references/chains/index.md`
- write: `skills/moai/references/source-manifest.json`

## Decisions

- D-1: The write set and the shared spots are listed by the plan, not enforced by it: the plan names them and the lane reserves them, the way a plan already names the check of each criterion. Reversible: a later change may make `card take` read them.
- D-2: The holders are read from the card ledger when there is one, because that is the only record of a reservation; a project without a ledger writes `no ledger` and the lane asks the leader.
- D-3: The re-audit rule is stated in the chains index, the one place that orders the agents of a request; the audit workflow already says the same auditor re-judges a repair.

## Plan

1. Tests first (RED): `tests/plan_output_sections.rs` with the output, holders, parity and chains cases; the row count pinned by `chains_name_real_routes_agents_and_references` in `tests/agent_defs_package.rs` (held by another card; asked to change the one number from 8 to 10) moves with the table; see the new file fail on the unchanged text.
2. `agents/plan.md`: the three output sections and the holders section, one Verification bullet; `check-completeness.md` item 7: the parity sentence; `chains/index.md`: two rows and two rules.
3. Rewrite the manifest hashes, then run the new file, the tests that read these files, the python manifest tests, clippy and rustfmt, then audit and write the Result.

## Result

Not yet written.
