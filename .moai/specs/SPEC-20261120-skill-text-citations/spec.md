---
id: SPEC-20261120-skill-text-citations
title: "shipped skill text carries no card number, Issue number, decision citation or internal tooling name"
status: draft
revision: 1
created: 2026-11-20
---

# shipped skill text carries no card number, Issue number, decision citation or internal tooling name

## Intent

The files that ship to users (`skills/`, `agents/`, `templates/`, `install/`) quote moai-adk's own history: card numbers, Issue numbers, "user decision <date>" and the names of its development scripts. A user of MoAI cannot follow those and should not need to. This first change adds a scan that keeps such text out, with a list of the files that still carry it that can only shrink, cleans the shipped files that no other work holds, and moves the reasons to a developer document.

## Requirements

- REQ-1: The text of every shipped file except those on the pending list shall carry no card number, no Issue or pull-request number, no citation of a user or leader decision, and no name of a moai-adk development script.
- REQ-2: The pending list shall only shrink: a file on it shall still carry such text, so a file that was cleaned leaves the list in the same change.
- REQ-3: A cleaned file shall keep every rule it states: only the citation, the sentence that tells where a rule came from, or the development script name goes, and the developer document `docs/skill-text-rationale.md` keeps each removed reason under the file's path.
- REQ-4: A change to a shipped file shall carry the new hash of the file in the package manifest and the new wording of every test that pinned the old text, in the same commit.

## Acceptance

- AC-1 (REQ-1, REQ-2): the scan passes, flags a card number, an Issue number, a decision citation and a development script name in sample text, ignores the machine markers `LANE-READY-INTEGRATION` and `LANE-BLOCKED`, and fails for a shipped file outside the pending list and for a pending file that has gone clean · verify: `cargo test --locked --offline --test skill_text_citations scan`
- AC-2 (REQ-3): `docs/skill-text-rationale.md` has one section per cleaned file, titled with its path, and each cleaned file has no citation left · verify: `cargo test --locked --offline --test skill_text_citations rationale`
- AC-3 (REQ-4): the manifest and the tests that read the shipped text still pass · verify: `python3 -m unittest tests.source_manifest_test tests.skill_references_test`; `cargo test --locked --offline --test product_cli --test agent_defs_package --test mode_workflow_steps --test mode_cards_53`

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
- write: `skills/moai/references/internal/self-check.md`
- write: `skills/moai/references/internal/review.md`
- write: `skills/moai/references/internal/commit-trailers.md`

## Plan

1. Tests first (RED): `tests/skill_text_citations.rs` reads the shipped files of this checkout (no HOME, nothing is run). It holds the scan, a pending list of the files that still carry citations after this change, and the check of the rationale document; run it and see the thirteen files fail.
2. Clean the thirteen files one by one, keeping each rule and writing the removed reasons into `docs/skill-text-rationale.md` as they go; fix the tests that pinned the old wording in the same commit; update the manifest hashes.
3. The pending list holds what other work still owns or a later change cleans: the project and update workflows and the question files (held by another card), `verify.md`, `audit.md` and the two mode workflows (their markers are read by checkers and tests), `rule-history.md` (moves to developer documents), the manifest's own reason texts, `install/` and `templates/`.
4. Run the new file, the tests that read shipped text, the scan checks of `tools/repo_scans.json` and the same-prefix rule, then audit and write the Result.
