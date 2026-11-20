---
id: SPEC-20261120-skill-text-citations-verify-audit
title: "the verify, audit, leader and lane workflow texts carry no card number, Issue number, decision citation or internal tooling name"
status: draft
revision: 1
created: 2026-11-20
---

# the verify, audit, leader and lane workflow texts carry no card number, Issue number, decision citation or internal tooling name

## Intent

The earlier change `SPEC-20261120-skill-text-citations` cleaned most shipped skill files and left five on its pending list because their wording is read by checkers or pinned by tests: `skills/moai/workflows/verify.md`, `skills/moai/workflows/audit.md`, `agents/audit.md`, `skills/moai/workflows/mode-leader.md` and `skills/moai/workflows/mode-agent.md`. They still quote moai-adk's own history (card numbers, Issue numbers, "leader decision <date>", the names of its development scripts). A user of MoAI cannot follow those, so this change takes them off the pending list the same way: every rule stays, only the citation goes.

## Requirements

- REQ-1: The text of the five files named in the Intent shall carry no card number, no Issue or pull-request number, no citation of a user or leader decision, and no name of a moai-adk development script.
- REQ-2: The pending list of `tests/skill_text_citations.rs` shall no longer name these five files, so the scan keeps them clean.
- REQ-3: A checker or a test that reads a phrase of these files shall keep finding it: the markers `AUDIT #<pr> … head <40-hex> base <40-hex> — verdict: PASS|FAIL`, `## 독립 review`, `수리 확인 범위`, `새 지적 분류`, `moai-audit-claim` and `moai-claim-handoff` stay as written.
- REQ-4: The developer document `docs/skill-text-rationale.md` shall keep each removed reason, and the moai-adk-only commands the shipped text no longer names, under the path of the file they came from.
- REQ-5: A change to a shipped file shall carry the new hash of the file in the package manifest and the new wording of every test that pinned the old text, in the same commit.
- REQ-6: The two lane workflows and the verify workflow shall state the local finish line once each and name the full CI of the release pull request.

## Acceptance

- AC-1 (REQ-1, REQ-2): the scan passes with the five files off the pending list and fails when a citation is put back into one of them · verify: `cargo test --locked --offline --test skill_text_citations scan`
- AC-2 (REQ-3, REQ-4): `docs/skill-text-rationale.md` has a section titled with each of the five paths, and every phrase the checkers read is still in the files · verify: `cargo test --locked --offline --test skill_text_citations rationale`; `cargo test --locked --offline --test self_audit_cli --test gate_collect --test audit_gate_cli`
- AC-3 (REQ-5): the manifest and the tests that read the shipped text still pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test tests.check_pr_self_audit_test`; `cargo test --locked --offline --test product_cli --test agent_defs_package --test mode_workflow_steps --test mode_cards_53 --test flaky_followups --test gate_collect --test lane_rules_2879 --test lane_rules_2836 --test close_step_2903 --test test_isolation_rules --test workflow_size_budget`
- AC-4 (REQ-6): each lane workflow states the local finish line once and verify names the release pull request's CI · verify: `cargo test --locked --offline --test skill_text_citations finish_line`

## Scope

- write: `.moai/specs/SPEC-20261120-skill-text-citations-verify-audit/spec.md`
- write: `tests/skill_text_citations.rs`
- write: `docs/skill-text-rationale.md`
- write: `skills/moai/workflows/verify.md`
- write: `skills/moai/workflows/audit.md`
- write: `agents/audit.md`
- write: `skills/moai/workflows/mode-leader.md`
- write: `skills/moai/workflows/mode-agent.md`
- write: `skills/moai/references/source-manifest.json`
- write: `tests/flaky_followups.rs`
- write: `tests/gate_collect.rs`
- write: `tests/lane_rules_2879.rs`
- write: `tests/mode_cards_53.rs`
- write: `tests/self_audit_cli.rs`
- write: `tests/test_isolation_rules.rs`
- write: `tests/workflow_size_budget.rs`
- write: `tests/mode_workflow_steps.rs`
- write: `tests/product_cli.rs`
- write: `tests/agent_defs_package.rs`
- write: `tests/close_step_2903.rs`
- write: `tests/lane_rules_2836.rs`
- write: `tests/lane_next_rules.rs`
- write: `tests/self_check_test.py`
- write: `tests/smoke_harness_test.py`
- write: `tests/check_pr_self_audit_test.py`

## Decisions

- D-1: Closing a card at the local finish line is not solved here. `card done` accepts only a card with a linked merged pull request, and a card at the local finish line has none. The lane text says so and tells the holder to ask the leader; a rule that closes such a card needs a change of `card done` and its tests, which is outside this SPEC's scope and is left to a follow-up card the leader adds. Reversible: yes (text only).
- D-2: The finish-line section of `verify.md` is the section the paired change wrote for every file that tells how a change ends, and it carries two sentences that are ahead of the code: item 4 says the status line shows commits ahead (↑) and behind (↓), while the status line draws only the behind count on purpose (`src/statusline/render.rs`), and item 6 says a card that landed locally is closed with `card done` on the local landing evidence, while `card done` still asks for a linked merged pull request (`src/ops/card.rs`). The user's decision of the finish line comes first and the code follows in the paired change; the lane text of `mode-agent.md` and `mode-leader.md` keeps the honest wording until then (D-1). The scan does not catch the plural "user decisions" followed by a date; the paired text keeps that marker word for word because its test requires it (an exception to REQ-1 recorded here, the scan stays as it is).

## Plan

1. Tests first (RED): take the five files off `PENDING` and into `CLEANED`, add a rationale section for each, and see the scan fail on the four files.
2. Clean the five files, keeping each rule: drop the citation, the sentence that tells where a rule came from, and the moai-adk script names; write the removed reasons and the moai-adk-only commands into the rationale document; keep the markers named in REQ-3 as written.
3. Find the tests and checkers that read the changed text with a repository-wide search of the old wording, fix what pins it, and update the manifest hashes.
4. Run the new file, the tests that read shipped text, the scan checks of `tools/repo_scans.json` and the same-prefix rule, then audit and write the Result.

## Result

- AC-1: met — `cargo test --locked --offline --test skill_text_citations` passed with `mode-agent.md`, `verify.md`, `audit.md`, `agents/audit.md` and `mode-leader.md` off the pending list (PENDING 17, CLEANED 14), inside one run of `--lib`, `--bins` and 129 test targets: 132 `test result: ok`, 0 failed.
- AC-2: met — `docs/skill-text-rationale.md` has a `## ` section for each of the five paths (`rationale_has_a_section_per_cleaned_file` passes); the checker markers stay as written (`self_audit_cli`, `gate_collect`, `agent_defs_package` pass in the same run). The AC-2 command now names real targets (`self_audit_cli`, `gate_collect`, `audit_gate_cli`: 5, 48 and 5 passed).
- AC-3: met — `python3 -m unittest` over 26 python modules (`source_manifest_test`, `skill_references_test`, `check_pr_self_audit_test`, `smoke_harness_test`, `self_check_test`, `ci_scope_test`, `repo_scans_test` and nineteen more that read workflow or agent text) ran 585 tests; the one failure of that run was a manifest test racing a manifest rewrite I made while it ran, and `tests.source_manifest_test` alone then ran 45 tests, OK; `python3 -I tools/source_manifest.py check` answered `ok: true`. The pinned wording was moved in thirteen Rust test files and three python tests (`check_pr_self_audit_test` after the first audit named two of its pins, `self_check_test`, `smoke_harness_test`); the smoke paragraph, which names a moai-adk-only script, stands word for word in the rationale document that `smoke_harness_test` reads. The Rust readers of the five files (22 targets, including `mode_workflow_steps`, `skill_text_citations`, `self_audit_cli`) ran 279 tests, 0 failed. `self_audit_cli` no longer counts mentions of repository scripts, which the cleaning removes on purpose; it now fails when its scan does not reach `mode-agent.md`, `mode-leader.md`, `verify.md` or `audit.md`.
- AC-4: met — `finish_line_rule_is_stated_once_in_each_lane_workflow` passes: `mode-agent.md` and `mode-leader.md` state the local finish line once each, and `verify.md` names the full CI of the release pull request.
- Also run: `cargo clippy --locked --offline --all-targets -- -D warnings` finished with exit 0 and no warning line (read in `target/c151c/clippy2.log`); `rustfmt --edition 2021 --check` on `tests/self_audit_cli.rs` and `tests/skill_text_citations.rs` is clean; `git diff --check` is clean; `moai spec check` is run again before the commit.
- Not done here: the same cleaning for the files still on the pending list (the SKILL.md and workflow files of other cards, `install/`, `templates/`); the held file `docs/rust-product-path.md` still names an old reference; closing a card that has no merged pull request (D-1); the card inbox, `moai bundle` and the status line still read the old `LANE-READY-INTEGRATION` marker and nothing reads `MERGE-READY` yet; `skills/moai/workflows/sync.md` still tells every change to open a pull request, and the finish line is not yet tied to the project's chosen workflow kind (a project without a local integration branch is told only by the words "keeps no local integration branch"); the gate reads only a `local-check` line that names the repository's own script, so a user project's line is not taken as evidence; the repository's own check scripts are named only in the rationale document, which no lane is told to read; "this repository" and "SPEC §" wording stays in a few sentences. The auditor named these as non-blocking. Also named by the last audit and left: the finish line (4) says the status line shows only how far the branch is behind its upstream, because the status line hides the ahead count on purpose; the text says three slots in `verify.md` because the code default is three, and the paired card that raises the default to five changes both numbers there with its code; the commands, the `refs/moai/lane-merge` ref and the lock path that finish-line items (2) and (3) name are not yet in `src` or `tools`; a primary-checkout guard with `deny_commits_on` refuses the leader's revert commit (`src/hook_guards.rs`); and the sentence about the recorded open-pull-request limit does not say where it is recorded.
- Pair change after the audit (leader decisions on the pair request of C-204): the shipped-text scan no longer treats `M-Flow` as a development word (it is the product flow named by the `work_flow.flow` value of a profile; `DEV_WORDS` 9 to 8, the fixture line moved to the clean lines) and `verify.md` carries the finish-line section with the text of `sync.md` at the pair's newest commit (checked equal by a script that reads both; its marker once, its ten sentences present). The manifest hash of `verify.md` is rewritten. Run on this head: `skill_text_citations` 6, `agent_defs_package` 19, `flaky_followups` 3, `sdd_guidance` 6, `test_isolation_rules` 3, `workflow_size_budget` 6, `product_cli` 79, `mode_workflow_steps` 43, `core_rules_2876` 6, `home_routing_cli` 11 passed, 0 failed; python `source_manifest_test` 45, `skill_references_test` 5, `repo_scans_test` 25, `ci_scope_test` 40, `check_pr_self_audit_test` 39 OK in an empty temporary HOME. Not run for this change: `--lib`, `--bins`, clippy, and the pair's own `mflow_finish_line` test (it lives on its branch); the first auditor read this delta twice (heads 9beb1d4dff and ab83e1ea73) and passed it. The old-wording test `finish_line_rule_is_stated_once_in_each_lane_workflow` still pins `Local finish line:` of `mode-agent.md` and `mode-leader.md`; it moves with the pair's rewrite of those two files when the pair lands.
- Not done here, added after the delta audit: the two sentences of `verify.md` named in D-2 (status line ↑ and `card done` on local evidence) are ahead of the code until the paired change and its code cards land; the scan does not flag the plural "user decisions" with a date; the six development words other than the product flow name have no fixture line each, so deleting one is not caught (not caused by this change).
