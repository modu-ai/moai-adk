---
id: SPEC-20261120-skill-text-citations
title: "shipped skill text carries no card number, Issue number, decision citation or internal tooling name"
status: draft
revision: 1
created: 2026-11-20
---

# shipped skill text carries no card number, Issue number, decision citation or internal tooling name

## Intent

The files that ship to users (`skills/`, `agents/`, `templates/`, `install/`) quote moai-adk's own history: card numbers, Issue numbers, "user decision <date>" and the names of its development scripts. A user of MoAI cannot follow those and should not need to. This first change adds a scan that keeps such text out, with a list of the files that still carry it that can only shrink, cleans the shipped citations that no test held by another card pins, and moves the reasons to a developer document.

## Requirements

- REQ-1: The text of every shipped file except those on the pending list shall carry no card number, no Issue or pull-request number, no citation of a user or leader decision, and no name of a moai-adk development script.
- REQ-2: The pending list shall name only files that still carry such text, so a file that was cleaned leaves the list in the same change.
- REQ-3: A cleaned file shall keep every rule it states, losing only a citation, a sentence that tells where a rule came from, or a development script name.
- REQ-4: The developer document `docs/skill-text-rationale.md` shall keep each removed reason under the path of the file it came from.
- REQ-5: A change to a shipped file shall carry the new hash of the file in the package manifest and the new wording of every test that pinned the old text, in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2): the scan passes, flags a card number, an Issue number, a decision citation and a development script name in sample text, ignores the machine markers `LANE-READY-INTEGRATION` and `LANE-BLOCKED`, and fails for a shipped file outside the pending list and for a pending file that has gone clean · verify: `cargo test --locked --offline --test skill_text_citations scan`
- AC-2 (REQ-3, REQ-4): `docs/skill-text-rationale.md` has one section per cleaned file, titled with its path, and each cleaned file has no citation left · verify: `cargo test --locked --offline --test skill_text_citations rationale`
- AC-3 (REQ-5): the manifest and the tests that read the shipped text still pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test`; `cargo test --locked --offline --test product_cli --test agent_defs_package --test mode_workflow_steps --test mode_cards_53`

## Scope

- write: `.moai/specs/SPEC-20261120-skill-text-citations/spec.md`
- write: `tests/skill_text_citations.rs`
- write: `docs/skill-text-rationale.md`
- write: `skills/moai/SKILL.md`
- write: `skills/moai/references/source-manifest.json`
- write: `skills/moai/workflows/setup.md`
- write: `skills/moai/workflows/sync.md`
- write: `skills/moai/workflows/optimize.md`
- write: `skills/moai/workflows/read.md`
- write: `skills/moai/workflows/run.md`
- write: `skills/moai/references/adapters/orca.md`
- write: `skills/moai/references/adapters/aside.md`
- write: `skills/moai/references/internal/build-folder.md`
- write: `skills/moai/references/internal/git-safety.md`
- write: `skills/moai/references/internal/review.md`
- write: `tests/product_cli.rs`
- write: `skills/moai/references/internal/commit-trailers.md`
- write: `tools/ci_scope.py`
- write: `tools/repo_scans.json`

## Plan

1. Tests first (RED): `tests/skill_text_citations.rs` reads the shipped files of this checkout (no HOME, nothing is run). It holds the scan, a pending list of the files that still carry citations after this change, and the check of the rationale document; run it and see the cleaned files fail.
2. Clean the shipped files one by one, keeping each rule and writing the removed reasons into `docs/skill-text-rationale.md` as they go; fix the tests that pinned the old wording in the same commit; update the manifest hashes.
3. A citation that a test held by another card pins (`tests/update_route.rs`, `tests/orca_optional_3005.rs`, `tests/agent_defs_package.rs`) stays for now, so `SKILL.md`, `orca.md`, `run.md` and `self-check.md` stay on the pending list until those tests change. The pending list also holds what other work still owns or a later change cleans: the project and update workflows and the question files (held by another card), `verify.md`, `audit.md` and the two mode workflows (their markers are read by checkers and tests), `rule-history.md` (moves to developer documents), the manifest's own reason texts, `install/` and `templates/`.
4. Run the new file, the tests that read shipped text, the scan checks of `tools/repo_scans.json` and the same-prefix rule, then audit and write the Result.

## Result

Done for the part no other card's test pins; the rest stays on `PENDING`. RED `8bce5dda4b` (3 of 5 new tests failed: the cleaned files still carried citations and the rationale document was missing), GREEN `aa09df1693` (nine shipped files clean, three partly clean, 12 manifest hashes, the document, the `product_cli` needle), `9f168bbbd9` (the document denied in `tools/ci_scope.py`), RED `fbc7fe8f54` and GREEN `9b6f8effa2` (the audit's forms the scan missed: `Issue 2619`, `PR 3234`, `Issue#2619`, `Decision D7`, `user's decision`, four digit cards, a character whose lower case is longer), `4b0e818c8b` (the scan registered in `tools/repo_scans.json`). Mutations (each reverted with `git checkout`, tree clean after): a card number, an Issue number, a decision citation, a decision id and a development script name put back into a cleaned file, and a rationale heading renamed: six of six killed by `skill_text_citations`; the auditor's own three plus a six-form sample were caught too.

- AC-1: met — `skill_text_citations` 5 passed: the scan flags every kind (card, number, decision, tool) and the audit's added forms, leaves `LANE-READY-INTEGRATION`, `LANE-BLOCKED`, `ABC-12`, `CI-3`, `&#8212;`, `#123456`, `C-<number>` and plain "user decision" sentences alone, passes for every shipped file outside `PENDING`, and fails for a pending file that went clean; the independent count of the auditor (own regular expression over `skills/`, `agents/`, `templates/`, `install/`) found the same 22 files (its two extra hits were its own false positives).
- AC-2: met — `rationale_has_a_section_per_cleaned_file` and `rationale_cleaned_files_carry_no_citation` pass: nine cleaned files have a section titled with their path in `docs/skill-text-rationale.md`, plus sections for the three partly cleaned ones; the auditor compared every removed line of 12 files with the document and found nothing missing.
- AC-3: met — final head `4b0e818c8b`, each `test result` read: the 25 targets that read shipped text 306 passed, 0 failed (`agent_defs_package`, `mode_workflow_steps`, `mode_cards_53`, `core_rules_2876`, `sdd_guidance`, `orca_optional_3005`, `update_route`, `product_cli`, `workflow_size_budget` and the others), the same-prefix run (16 targets) 160 passed, `--lib` 1595 passed, `--bins` 56 passed; python `repo_scans_test`, `source_manifest_test`, `local_check_test` 123 OK, `skill_references_test` 5 OK, `ci_scope_test` 40 OK; `tools/source_manifest.py check` exit 0; all 40 checks of `tools/repo_scans.json` exit 0 (39 old ones and the new one).

Not done (stays on `PENDING`, each for a stated reason): `skills/moai/SKILL.md` (one phrase pinned by `tests/update_route.rs`), `references/adapters/orca.md` (two table cells and one list sentence pinned by `tests/orca_optional_3005.rs`), `workflows/run.md` (the cleanup rule pinned by `tests/agent_defs_package.rs`), `references/internal/self-check.md` (several needles of `tests/agent_defs_package.rs`), the project and update workflows and the question files (rewritten by another card), `verify.md`, `audit.md`, `mode-agent.md`, `mode-leader.md`, `rule-history.md` and the manifest's own reason texts (later pieces), `install/`, `templates/`, `agents/audit.md`. Known small residues the auditor named, not fixed: the possessive form only matches a straight apostrophe; item 8 of `build-folder.md` leaves a picked folder that no longer shows `trim` alone, where the removed repository command refused all picked rows. Not run: Windows, a host session reading the shortened text.
