---
id: SPEC-20261120-skill-text-citations-verify-audit
title: "the verify, audit and leader workflow texts carry no card number, Issue number, decision citation or internal tooling name"
status: draft
revision: 1
created: 2026-11-20
---

# the verify, audit and leader workflow texts carry no card number, Issue number, decision citation or internal tooling name

## Intent

The earlier change `SPEC-20261120-skill-text-citations` cleaned most shipped skill files and left four on its pending list because their wording is read by checkers or pinned by tests: `skills/moai/workflows/verify.md`, `skills/moai/workflows/audit.md`, `agents/audit.md` and `skills/moai/workflows/mode-leader.md`. They still quote moai-adk's own history (card numbers, Issue numbers, "leader decision <date>", the names of its development scripts). A user of MoAI cannot follow those, so this change takes them off the pending list the same way: every rule stays, only the citation goes.

## Requirements

- REQ-1: The text of the four files named in the Intent shall carry no card number, no Issue or pull-request number, no citation of a user or leader decision, and no name of a moai-adk development script.
- REQ-2: The pending list of `tests/skill_text_citations.rs` shall no longer name these four files, so the scan keeps them clean.
- REQ-3: A checker or a test that reads a phrase of these files shall keep finding it: the markers `AUDIT #<pr> … head <40-hex> base <40-hex> — verdict: PASS|FAIL`, `## 독립 review`, `수리 확인 범위`, `새 지적 분류`, `moai-audit-claim` and `moai-claim-handoff` stay as written.
- REQ-4: The developer document `docs/skill-text-rationale.md` shall keep each removed reason, and the moai-adk-only commands the shipped text no longer names, under the path of the file they came from.
- REQ-5: A change to a shipped file shall carry the new hash of the file in the package manifest and the new wording of every test that pinned the old text, in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2): the scan passes with the four files off the pending list and fails when a citation is put back into one of them · verify: `cargo test --locked --offline --test skill_text_citations scan`
- AC-2 (REQ-3, REQ-4): `docs/skill-text-rationale.md` has a section titled with each of the four paths, and every phrase the checkers read is still in the files · verify: `cargo test --locked --offline --test skill_text_citations rationale`; `cargo test --locked --offline --test self_audit_cli --test pr_link_cli --test gate_cli`
- AC-3 (REQ-5): the manifest and the tests that read the shipped text still pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test`; `cargo test --locked --offline --test product_cli --test agent_defs_package --test mode_workflow_steps --test mode_cards_53`

## Scope

- write: `.moai/specs/SPEC-20261120-skill-text-citations-verify-audit/spec.md`
- write: `tests/skill_text_citations.rs`
- write: `docs/skill-text-rationale.md`
- write: `skills/moai/workflows/verify.md`
- write: `skills/moai/workflows/audit.md`
- write: `agents/audit.md`
- write: `skills/moai/workflows/mode-leader.md`
- write: `skills/moai/references/source-manifest.json`

## Plan

1. Tests first (RED): take the four files off `PENDING` and into `CLEANED`, add a rationale section for each, and see the scan fail on the four files.
2. Clean the four files, keeping each rule: drop the citation, the sentence that tells where a rule came from, and the moai-adk script names; write the removed reasons and the moai-adk-only commands into the rationale document; keep the markers named in REQ-3 as written.
3. Find the tests and checkers that read the changed text with a repository-wide search of the old wording, fix what pins it, and update the manifest hashes.
4. Run the new file, the tests that read shipped text, the scan checks of `tools/repo_scans.json` and the same-prefix rule, then audit and write the Result.

## Result

Not yet written.
